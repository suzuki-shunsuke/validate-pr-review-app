package controller

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/suzuki-shunsuke/validate-pr-review-app/pkg/config"
	"github.com/suzuki-shunsuke/validate-pr-review-app/pkg/github"
	"github.com/suzuki-shunsuke/validate-pr-review-app/pkg/validation"
)

// carryForwardCheck handles pull_request.synchronize events.
// When new commits are pushed that are all empty or clean merge commits,
// carry forward the validation result from the most recent reviewed commit.
func (c *Controller) carryForwardCheck(ctx context.Context, logger *slog.Logger, ev *Event, trust *config.Trust, insecure *config.Insecure) *validation.Result {
	pr, err := c.gh.GetPR(ctx, ev.RepoOwner, ev.RepoName, ev.PRNumber)
	if err != nil {
		return &validation.Result{Error: fmt.Errorf("get a pull request: %w", err).Error()}
	}
	logger.Info("fetched a pull request for carry-forward check", "pull_request", pr)

	// Guard against stale webhook redeliveries.
	if ev.HeadSHA != pr.HeadSHA {
		logger.Info("ignoring stale webhook: event SHA does not match current PR HEAD",
			"event_sha", ev.HeadSHA, "head_sha", pr.HeadSHA)
		return nil
	}

	approvers := c.findCarryForwardApprovers(ctx, logger, ev, pr)
	if approvers == nil {
		return nil
	}

	// Use the carried-forward approvers for validation.
	pr.Approvers = approvers
	c.checkApproverCommits(ctx, logger, ev, pr)

	input := &validation.Input{
		PR: pr,
		Trust: &validation.Trust{
			TrustedApps:           trust.UniqueTrustedApps,
			UntrustedMachineUsers: trust.UntrustedMachineUsers,
		},
	}
	if insecure != nil {
		input.Insecure = &validation.Insecure{
			AllowUnsignedCommits:       insecure.AllowUnsignedCommits != nil && *insecure.AllowUnsignedCommits,
			UnsignedCommitApps:         toSet(insecure.UnsignedCommitApps),
			UnsignedCommitMachineUsers: toSet(insecure.UnsignedCommitMachineUsers),
		}
	}
	result := c.validator.Run(logger, input)
	result.CarriedForward = true
	return result
}

// findCarryForwardApprovers walks PR commits from HEAD (newest) to oldest.
// For each commit, it checks whether the commit is "harmless" (empty or clean merge).
// When a commit with reviews is found, its approvers are returned.
// Returns nil if carry-forward is not applicable.
func (c *Controller) findCarryForwardApprovers(ctx context.Context, logger *slog.Logger, ev *Event, pr *github.PullRequest) map[string]*github.User {
	prCommitSHAs := buildPRCommitSHAs(pr)
	// Walk commits from newest to oldest.
	for i := len(pr.Commits) - 1; i >= 0; i-- {
		commit := pr.Commits[i]

		// Check if this commit has reviews.
		if approvers, ok := pr.ApproversByCommit[commit.SHA]; ok && len(approvers) > 0 {
			logger.Info("found reviewed commit for carry-forward",
				"commit", commit.SHA, "approver_count", len(approvers))
			return approvers
		}

		// Check if this commit is harmless (empty or clean merge).
		if commit.ChangedFilesIfAvailable != nil && *commit.ChangedFilesIfAvailable == 0 {
			logger.Info("commit is empty, continuing carry-forward walk", "commit", commit.SHA)
			continue
		}
		if c.isCleanMergeCommit(ctx, logger, ev, commit, prCommitSHAs, pr.BaseSHA) {
			logger.Info("commit is a clean merge, continuing carry-forward walk", "commit", commit.SHA)
			continue
		}

		// Commit is not harmless — can't carry forward.
		logger.Info("commit is not empty or clean merge, carry-forward not applicable",
			"commit", commit.SHA)
		return nil
	}

	// No commit with reviews found.
	logger.Info("no reviewed commit found, carry-forward not applicable")
	return nil
}
