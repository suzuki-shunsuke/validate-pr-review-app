package controller

import (
	"context"
	"log/slog"

	"github.com/suzuki-shunsuke/validate-pr-review-app/pkg/github"
)

// checkApproverCommits checks if commits by approvers are harmless
// (empty commits or clean merge commits) and marks them as IsAllowedMergeCommit
// so the validator can skip the self-approval check for those commits.
// Only commits where the committer is an approver are checked.
func (c *Controller) checkApproverCommits(ctx context.Context, logger *slog.Logger, ev *Event, pr *github.PullRequest) {
	prCommitSHAs := buildPRCommitSHAs(pr)
	for _, commit := range pr.Commits {
		if commit.Committer == nil {
			continue
		}
		login := commit.Committer.Login
		if _, ok := pr.Approvers[login]; !ok {
			continue
		}
		// Empty commits (0 changed files) cannot introduce malicious changes.
		if commit.ChangedFilesIfAvailable != nil && *commit.ChangedFilesIfAvailable == 0 {
			commit.IsAllowedMergeCommit = true
			continue
		}
		allowed := c.isCleanMergeCommit(ctx, logger, ev, commit, prCommitSHAs, pr.BaseSHA)
		commit.IsAllowedMergeCommit = allowed
		if !allowed {
			// Early termination: if any approver commit is not a clean merge,
			// two approvals will be required regardless, so stop checking.
			return
		}
	}
}

func buildPRCommitSHAs(pr *github.PullRequest) map[string]struct{} {
	m := make(map[string]struct{}, len(pr.Commits))
	for _, c := range pr.Commits {
		m[c.SHA] = struct{}{}
	}
	return m
}

const maxCompareFiles = 300

// isCleanMergeCommit checks whether a commit is a merge commit whose parents'
// diffs to the merge commit have no overlapping files (i.e., no conflict resolution)
// and whose non-PR parents are ancestors of the base branch.
func (c *Controller) isCleanMergeCommit(ctx context.Context, logger *slog.Logger, ev *Event, commit *github.Commit, prCommitSHAs map[string]struct{}, prBaseSHA string) bool { //nolint:cyclop
	if len(commit.Parents) != 2 { //nolint:mnd
		return false
	}

	allFiles := make(map[string]struct{})
	for _, parentSHA := range commit.Parents {
		files, err := c.gh.CompareCommits(ctx, ev.RepoOwner, ev.RepoName, parentSHA, commit.SHA)
		if err != nil {
			// Fail closed: treat API failure as requiring two approvals.
			logger.Warn("compare commits API failed, requiring two approvals",
				"error", err, "base", parentSHA, "head", commit.SHA)
			return false
		}
		if len(files) >= maxCompareFiles {
			// Compare Two Commits API limitation: cannot guarantee all changed files are returned.
			logger.Info("too many changed files, requiring two approvals",
				"file_count", len(files), "base", parentSHA, "head", commit.SHA)
			return false
		}
		for _, f := range files {
			if _, ok := allFiles[f]; ok {
				logger.Info("overlapping file found between parent diffs, requiring two approvals",
					"file", f, "commit", commit.SHA)
				return false
			}
			allFiles[f] = struct{}{}
		}
	}

	// Verify that non-PR parents are ancestors of the base branch.
	// This ensures the merge is with the base branch (e.g., "Update branch"),
	// not an arbitrary branch that could introduce unreviewed code.
	for _, parentSHA := range commit.Parents {
		if _, ok := prCommitSHAs[parentSHA]; ok {
			continue
		}
		ancestor, err := c.gh.IsAncestor(ctx, ev.RepoOwner, ev.RepoName, parentSHA, prBaseSHA)
		if err != nil {
			logger.Warn("ancestor check API failed, requiring two approvals",
				"error", err, "parent", parentSHA, "base", prBaseSHA)
			return false
		}
		if !ancestor {
			logger.Info("merge parent is not an ancestor of base branch, requiring two approvals",
				"parent", parentSHA, "base", prBaseSHA, "commit", commit.SHA)
			return false
		}
	}

	return true
}
