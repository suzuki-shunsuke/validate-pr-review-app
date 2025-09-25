package validation

import (
	"log/slog"
	"maps"
	"path"
	"slices"
	"sort"
	"strings"

	"github.com/suzuki-shunsuke/validate-pr-review-app/pkg/config"
	"github.com/suzuki-shunsuke/validate-pr-review-app/pkg/github"
)

// Run validates pull request reviews.
// It gets pull request reviews and committers via GitHub GraphQL API, and checks if people other than committers approve the PR.
// If the PR isn't approved by people other than committers, it returns an error.
func (c *Controller) Run(_ *slog.Logger, input *Input) *config.Result { //nolint:cyclop
	pr := input.PR
	result := &config.Result{}
	ignoredApprovers := make(map[string]*github.IgnoredApproval, len(pr.Approvers))
	approvers := make(map[string]struct{}, len(pr.Approvers))
	for approver := range pr.Approvers {
		if isApp(approver) {
			if !c.VerifyApp(approver) {
				// Ignore the approval from untrusted apps
				ignoredApprovers[approver] = &github.IgnoredApproval{
					Login: approver,
					IsApp: true,
				}
			}
			continue
		}
		if !c.VerifyUser(approver) {
			// Ignore the approval from untrusted machine users
			ignoredApprovers[approver] = &github.IgnoredApproval{
				Login:                  approver,
				IsUntrustedMachineUser: true,
			}
			continue
		}
		approvers[approver] = struct{}{}
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
	for _, commit := range pr.Commits {
		if untrustedCommit := c.VerifyCommit(commit); untrustedCommit != nil {
			// Two approvals are required as there is an untrusted commit, but one approval is given
			result.UntrustedCommits = append(result.UntrustedCommits, untrustedCommit)
		}
		committer := commit.Committer
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

func isApp(login string) bool {
	return strings.HasSuffix(login, "[bot]")
}

func (c *Controller) VerifyApp(login string) bool {
	if _, ok := c.input.TrustedApps[login]; ok {
		return true
	}
	return false
}

func (c *Controller) VerifyUser(login string) bool {
	if _, ok := c.input.TrustedMachineUsers[login]; ok {
		return true
	}
	for pattern := range c.input.UntrustedMachineUsers {
		matched, err := path.Match(pattern, login)
		if err != nil { // TODO error handling
			continue
		}
		if matched {
			return false
		}
	}
	return true
}

func (c *Controller) VerifyCommit(commit *github.Commit) *github.UntrustedCommit {
	user := commit.Committer
	login := user.Login
	sha := commit.SHA
	if !commit.Linked() {
		return &github.UntrustedCommit{
			NotLinkedToUser: true,
			SHA:             sha,
		}
	}
	sig := commit.Signature
	if sig == nil || !sig.IsValid {
		return &github.UntrustedCommit{
			Login:       login,
			SHA:         sha,
			InvalidSign: sig,
		}
	}
	if user.IsApp {
		if _, ok := c.input.TrustedApps[login]; ok {
			return nil
		}
		return &github.UntrustedCommit{
			Login:          login,
			SHA:            sha,
			IsUntrustedApp: true,
		}
	}
	if c.VerifyUser(login) {
		return nil
	}
	return &github.UntrustedCommit{
		Login:                  login,
		SHA:                    sha,
		IsUntrustedMachineUser: true,
	}
}
