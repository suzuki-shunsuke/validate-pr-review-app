package validation

import (
	"log/slog"
	"maps"
	"slices"

	"github.com/suzuki-shunsuke/enforce-pr-review-app/pkg/github"
)

// Run enforces pull request reviews.
// It gets pull request reviews and committers via GitHub GraphQL API, and checks if people other than committers approve the PR.
// If the PR isn't approved by people other than committers, it returns an error.
func (c *Controller) Run(_ *slog.Logger, input *Input) *Result {
	// Get a pull request reviews and committers via GraphQL API
	return validatePR(input)
}

type State string

const (
	StateTwoApprovals            State = "two_approvals"
	StateApprovalIsRequired      State = "approval_is_required"
	StateTwoApprovalsAreRequired State = "two_approvals_are_required"
)

type Result struct {
	Error         string
	State         State
	Author        *User
	Approvers     []string
	SelfApprovers []string
	// app or untrusted machine user approvals
	IgnoredApprovers []string
	// app
	// untrusted machine user
	// not linked to any GitHub user
	// not signed commits
	UntrustedCommits []*Commit
	// settings
	TrustedApps           []string
	UntrustedMachineUsers []string
	TrustedMachineUsers   []string
}

type Commit struct {
	Login     string
	SHA       string
	Signature *github.Signature
}

type User struct {
	Login   string
	Trusted bool
}

func validatePR(input *Input) *Result { //nolint:cyclop,funlen
	pr := input.PR
	result := &Result{
		TrustedApps:           input.Config.TrustedApps,
		UntrustedMachineUsers: input.Config.UntrustedMachineUsers,
		TrustedMachineUsers:   input.Config.TrustedMachineUsers,
	}
	var ignoredApprovers []string
	approvers := make(map[string]struct{}, len(pr.Reviews.Nodes))
	for _, review := range pr.Reviews.Nodes {
		// Exclude reviews other than APPROVED and reviews for non head commits
		if !isLatestApproval(review, pr.HeadRefOID) {
			continue
		}
		// Exclude approvals from apps
		if review.Author.IsApp() {
			ignoredApprovers = append(ignoredApprovers, review.Author.GetLogin())
			continue
		}
		// Exclude approvals from untrusted machine users
		if _, ok := input.Config.UniqueUntrustedMachineUsers[review.Author.Login]; ok {
			ignoredApprovers = append(ignoredApprovers, review.Author.GetLogin())
			continue
		}
		approvers[review.Author.GetLogin()] = struct{}{}
	}
	// Convert map to sorted slice
	approversL := slices.Sorted(maps.Keys(approvers))

	if len(approvers) > 1 {
		// Allow multiple approvals
		result.Approvers = approversL
		result.State = StateTwoApprovals
		return result
	}

	if len(approvers) == 0 {
		// Approval is required
		result.State = StateApprovalIsRequired
		result.IgnoredApprovers = ignoredApprovers
		return result
	}

	// One approval

	// Check if the PR author is trusted
	requiredTwoApprovals := checkIfUserRequiresTwoApprovals(pr.Author, input)
	result.Author = &User{
		Login: pr.Author.GetLogin(),
	}
	if requiredTwoApprovals {
		result.State = StateTwoApprovalsAreRequired
		return result
	}
	oneApproval := false
	for _, commit := range pr.Commits.Nodes {
		committer := commit.User()
		login := committer.GetLogin()
		if checkIfUserRequiresTwoApprovals(committer, input) {
			requiredTwoApprovals = true
			commit := &Commit{
				Login:     login,
				SHA:       commit.SHA(),
				Signature: commit.Signature(),
			}
			result.UntrustedCommits = append(result.UntrustedCommits, commit)
			continue
		}
		// TODO check CODEOWNERS
		if _, ok := approvers[login]; ok {
			// self-approve
			result.SelfApprovers = append(result.SelfApprovers, login)
			continue
		}
		result.Approvers = append(result.Approvers, login)
		if !requiredTwoApprovals || oneApproval {
			return result
		}
		oneApproval = true
	}
	return result
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
		return !user.Trusted(input.Config.UniqueTrustedApps)
	}
	// Require two approvals for PRs created by untrusted machine users
	_, ok := input.Config.UniqueUntrustedMachineUsers[user.Login]
	return ok
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
}

type Review struct {
	Review               *github.Review
	Ignored              bool
	ApprovalFromApp      bool
	UntrustedMachineUser bool
	Message              string
}

func isLatestApproval(review *github.Review, headRefOID string) bool {
	return review.State == "APPROVED" && review.Commit.OID != headRefOID
}
