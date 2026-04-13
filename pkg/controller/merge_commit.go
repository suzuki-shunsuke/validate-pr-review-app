package controller

import (
	"context"
	"log/slog"

	"github.com/suzuki-shunsuke/validate-pr-review-app/pkg/github"
)

// checkMergeCommits checks if commits by approvers are clean merge commits
// that don't change the PR diff (e.g., "Update branch" on GitHub).
// If so, it marks them as IsAllowedMergeCommit so the validator can skip
// the self-approval check for those commits.
// Only commits where the committer is an approver are checked.
func (c *Controller) checkMergeCommits(ctx context.Context, logger *slog.Logger, ev *Event, pr *github.PullRequest) {
	for _, commit := range pr.Commits {
		if commit.Committer == nil {
			continue
		}
		login := commit.Committer.Login
		if _, ok := pr.Approvers[login]; !ok {
			continue
		}
		allowed := c.isCleanMergeCommit(ctx, logger, ev, commit)
		commit.IsAllowedMergeCommit = allowed
		if !allowed {
			// Early termination: if any approver commit is not a clean merge,
			// two approvals will be required regardless, so stop checking.
			return
		}
	}
}

const maxCompareFiles = 300

// isCleanMergeCommit checks whether a commit is a merge commit whose parents'
// diffs to the merge commit have no overlapping files (i.e., no conflict resolution).
func (c *Controller) isCleanMergeCommit(ctx context.Context, logger *slog.Logger, ev *Event, commit *github.Commit) bool {
	if len(commit.Parents) < 2 { //nolint:mnd
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

	return true
}
