package validation

import (
	"log/slog"
	"maps"
	"slices"

	"github.com/suzuki-shunsuke/enforce-pr-review-app/pkg/github"
)

type State string

const (
	// OK - Two approvals
	//   approvers
	StateTwoApprovals State = "two_approvals"
	// NG - approvals are required but actually no approval
	//   ignored approvers
	StateApprovalIsRequired State = "approval_is_required"
	// NG - two approvals are required but actually one approval
	//   why two approvals are required
	//     self approval
	//     untrusted author
	//     untrusted commit
	//   approvers
	//   self approvers
	//   ignored approvers
	StateTwoApprovalsAreRequired State = "two_approvals_are_required"
	// OK - one approval is sufficient
	//   approvers
	StateOneApproval State = "one_approval"
)

// Run enforces pull request reviews.
// It gets pull request reviews and committers via GitHub GraphQL API, and checks if people other than committers approve the PR.
// If the PR isn't approved by people other than committers, it returns an error.
func (c *Controller) Run(_ *slog.Logger, input *Input) *Result { //nolint:cyclop,funlen
	// Approval
	//   ignored
	//     non approval
	//     non latest
	//   mark ignored
	//      app
	//      untrusted machine user
	//   accepted but requires two approvals
	//     self approval
	//   accepted
	// PR Author
	//   requires two approvals
	//     untrusted user
	//   one approval is sufficient
	// Commits
	//   untrusted commits require two approvals
	//     untrusted app
	//     untrusted machine user
	//     not linked to user
	//     not signed
	// User
	//   trusted
	//     trusted app
	//     normal user
	//   untrusted
	//     normal app
	//     untrusted machine user
	//   not linked
	pr := input.PR
	result := &Result{
		TrustedApps:           input.Config.TrustedApps,
		UntrustedMachineUsers: input.Config.UntrustedMachineUsers,
		TrustedMachineUsers:   input.Config.TrustedMachineUsers,
	}
	ignoredApprovers := make(map[string]*github.IgnoredApproval, len(pr.Reviews.Nodes))
	approvers := make(map[string]struct{}, len(pr.Reviews.Nodes))
	for _, review := range pr.Reviews.Nodes {
		// Exclude reviews other than APPROVED and reviews for non head commits
		if review.Ignored(pr.HeadRefOID) {
			continue
		}
		if approval := review.ValidateIgnored(input.Config.UniqueTrustedMachineUsers, input.Config.UniqueUntrustedMachineUsers); approval != nil {
			ignoredApprovers[review.Author.Login] = approval
		} else {
			approvers[review.Author.Login] = struct{}{}
		}
	}

	if len(approvers) > 1 {
		// Allow multiple approvals
		result.Approvers = slices.Sorted(maps.Keys(approvers))
		result.State = StateTwoApprovals
		return result
	}

	result.IgnoredApprovers = ignoredApprovers
	if len(approvers) == 0 {
		// Approval is required
		result.State = StateApprovalIsRequired
		return result
	}

	// One approval

	// Require two approvals if the PR author is untrusted
	result.Author = pr.ValidateAuthor(input.Config.UniqueTrustedApps, input.Config.UniqueTrustedMachineUsers, input.Config.UniqueUntrustedMachineUsers)
	if result.Author != nil {
		// Two approvals are required as the pr author is untrusted, but one approval is given
		result.State = StateTwoApprovalsAreRequired
		return result
	}
	for _, commit := range pr.Commits.Nodes {
		if untrustedCommit := commit.ValidateUntrusted(input.Config.UniqueTrustedApps, input.Config.UniqueTrustedMachineUsers, input.Config.UniqueUntrustedMachineUsers); untrustedCommit != nil {
			// Two approvals are required as there is an untrusted commit, but one approval is given
			result.UntrustedCommits = append(result.UntrustedCommits, untrustedCommit)
		}
		committer := commit.User()
		login := committer.Login
		if _, ok := approvers[login]; ok {
			// Only one approval is given, but it's a self approval
			result.SelfApprover = login
		}
	}
	if result.SelfApprover != "" || len(result.UntrustedCommits) > 0 {
		return result
	}
	// One approval is sufficient
	// author and commits are trusted
	result.Approvers = slices.Sorted(maps.Keys(approvers))
	result.State = StateOneApproval
	return result
}

type Result struct {
	Error        string
	State        State
	Author       *github.Author
	Approvers    []string
	SelfApprover string
	// app or untrusted machine user approvals
	IgnoredApprovers map[string]*github.IgnoredApproval
	// app
	// untrusted machine user
	// not linked to any GitHub user
	// not signed commits
	UntrustedCommits []*github.UntrustedCommit
	// settings
	TrustedApps           []string
	UntrustedMachineUsers []string
	TrustedMachineUsers   []string
}
