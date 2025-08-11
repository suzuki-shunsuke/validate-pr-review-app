package validation

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"

	"github.com/suzuki-shunsuke/enforce-pr-review-app/pkg/github"
)

// Run enforces pull request reviews.
// It gets pull request reviews and committers via GitHub GraphQL API, and checks if people other than committers approve the PR.
// If the PR isn't approved by people other than committers, it returns an error.
func (c *Controller) Run(ctx context.Context, _ *slog.Logger, input *Input) error {
	// Get a pull request reviews and committers via GraphQL API
	pr, err := c.gh.GetPR(ctx, input.RepoOwner, input.RepoName, input.PR)
	if err != nil {
		return fmt.Errorf("get a pull request: %w", err)
	}
	if err := c.output(input, pr); err != nil {
		return err
	}
	return validatePR(input, pr)
}

func validatePR(input *Input, pr *github.PullRequest) error {
	reviews := ignoreUntrustedReviews(filterReviews(pr.Reviews.Nodes, pr.HeadRefOID), input.UntrustedMachineUsers)

	if len(reviews) > 1 {
		// Allow multiple approvals
		return nil
	}

	if len(reviews) == 0 {
		// Approval is required
		return errApproval
	}

	requiredTwoApprovals := checkIfTwoApprovalsRequired(pr, input)
	if requiredTwoApprovals {
		if len(reviews) == 1 {
			return errTwoApproval
		}
	}

	committers := getCommitters(convertCommits(pr.Commits.Nodes))
	// Checks if people other than committers approve the PR
	return validate(reviews, committers, requiredTwoApprovals)
}

func checkIfTwoApprovalsRequired(pr *github.PullRequest, input *Input) bool {
	if checkIfUserRequiresTwoApprovals(pr.Author, input) {
		return true
	}
	// If the pull request has commits from untrusted apps or machine users, require two approvals
	for _, commit := range pr.Commits.Nodes {
		user := commit.Commit.User()
		if checkIfUserRequiresTwoApprovals(user, input) {
			return true
		}
	}
	return false
}

// checkIfUserRequiresTwoApprovals checks if the user requires two approvals.
// It returns true if the user is an untrusted app or machine user.
func checkIfUserRequiresTwoApprovals(user *github.User, input *Input) bool {
	if user.GetLogin() == "" {
		// If the user is not linked to any GitHub user, require two approvals
		return true
	}
	if user.IsApp() {
		// Require two approvals for PRs created by trusted apps, excluding trusted apps
		return !user.Trusted(input.TrustedApps)
	}
	// Require two approvals for PRs created by untrusted machine users
	_, ok := input.UntrustedMachineUsers[user.Login]
	return ok
}

// convertCommits converts []*PullRequestCommit to []*Commit
func convertCommits(commits []*github.PullRequestCommit) []*github.Commit {
	arr := make([]*github.Commit, len(commits))
	for i, commit := range commits {
		arr[i] = commit.Commit
	}
	return arr
}

func (c *Controller) output(input *Input, pr *github.PullRequest) error {
	encoder := json.NewEncoder(c.stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(&Result{
		PullRequest: &PullRequest{
			Repo:       fmt.Sprintf("%s/%s", input.RepoOwner, input.RepoName),
			Number:     input.PR,
			HeadRefOID: pr.HeadRefOID,
			Reviews:    pr.Reviews,
			Commits:    pr.Commits,
		},
	}); err != nil {
		return fmt.Errorf("encode the pull request: %w", err)
	}
	return nil
}

type Result struct {
	PullRequest *PullRequest `json:"pull_request"`
}

type PullRequest struct {
	Repo       string          `json:"repo"`
	Number     int             `json:"number"`
	HeadRefOID string          `json:"headRefOid"`
	Reviews    *github.Reviews `json:"reviews" graphql:"reviews(first:30)"`
	Commits    *github.Commits `json:"commits" graphql:"commits(first:30)"`
}

type Approval struct {
	Login string
	ID    string
}

func getCommitters(commits []*github.Commit) map[string]struct{} {
	committers := make(map[string]struct{}, len(commits))
	for _, commit := range commits {
		login := commit.Login()
		if login == "" {
			continue
		}
		committers[login] = struct{}{}
	}
	return committers
}

func filterReviews(reviews []*github.Review, headRefOID string) []*github.Review {
	arr := make([]*github.Review, 0, len(reviews))
	for _, review := range reviews {
		if review.State != "APPROVED" || review.Commit.OID != headRefOID {
			// Ignore reviews other than APPROVED
			// Ignore reviews for non head commits
			continue
		}
		if review.Author.IsApp() {
			// Ignore approvals from bots
			continue
		}
		arr = append(arr, review)
	}
	return arr
}

func ignoreUntrustedReviews(reviews []*github.Review, untrustedUsers map[string]struct{}) []*github.Review {
	arr := make([]*github.Review, 0, len(reviews))
	for _, review := range reviews {
		if _, ok := untrustedUsers[review.Author.Login]; ok {
			// Ignore approvals from untrusted users
			continue
		}
		arr = append(arr, review)
	}
	return arr
}

var (
	errApproval    = errors.New("pull requests must be approved by people who don't push commits to them")
	errTwoApproval = errors.New("pull requests created by untrusted apps or machine users must be approved by two people")
)

// validate validates if committers approve the pull request themselves.
func validate(reviews []*github.Review, committers map[string]struct{}, requiredTwoApprovals bool) error {
	oneApproval := false
	for _, review := range reviews {
		// TODO check CODEOWNERS
		if _, ok := committers[review.Author.Login]; ok {
			// self-approve
			continue
		}
		if !requiredTwoApprovals || oneApproval {
			// Someone other than committers approved the PR, so this PR is not self-approved.
			return nil
		}
		oneApproval = true
	}
	if oneApproval {
		return errTwoApproval
	}
	return errApproval
}
