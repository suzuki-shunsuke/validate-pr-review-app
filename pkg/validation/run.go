package validation

import (
	"log/slog"
	"maps"
	"slices"
	"sort"

	"github.com/suzuki-shunsuke/enforce-pr-review-app/pkg/config"
	"github.com/suzuki-shunsuke/enforce-pr-review-app/pkg/github"
)

// Run enforces pull request reviews.
// It gets pull request reviews and committers via GitHub GraphQL API, and checks if people other than committers approve the PR.
// If the PR isn't approved by people other than committers, it returns an error.
func (c *Controller) Run(_ *slog.Logger, input *Input) *config.Result { //nolint:cyclop
	pr := input.PR
	result := &config.Result{
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
		result.State = config.StateApproved
		return result
	}

	ignoredApproversSlice := slices.Collect(maps.Values(ignoredApprovers))
	sort.Slice(ignoredApproversSlice, func(i, j int) bool {
		return ignoredApproversSlice[i].Login < ignoredApproversSlice[j].Login
	})
	result.IgnoredApprovers = ignoredApproversSlice
	if len(approvers) == 0 {
		// Approval is required
		result.State = config.StateApprovalIsRequired
		return result
	}

	// One approval

	// Require two approvals if the PR author is untrusted
	for _, commit := range pr.Commits.Nodes {
		if untrustedCommit := commit.ValidateUntrusted(input.Config.UniqueTrustedApps, input.Config.UniqueTrustedMachineUsers, input.Config.UniqueUntrustedMachineUsers); untrustedCommit != nil {
			// Two approvals are required as there is an untrusted commit, but one approval is given
			result.UntrustedCommits = append(result.UntrustedCommits, untrustedCommit)
		}
		committer := commit.Commit.User()
		login := committer.Login
		if _, ok := approvers[login]; ok {
			// Only one approval is given, but it's a self approval
			result.SelfApprover = login
		}
	}
	if result.SelfApprover != "" || len(result.UntrustedCommits) > 0 {
		result.State = config.StateTwoApprovalsAreRequired
		return result
	}
	// One approval is sufficient
	// author and commits are trusted
	result.Approvers = slices.Sorted(maps.Keys(approvers))
	result.State = config.StateApproved
	return result
}
