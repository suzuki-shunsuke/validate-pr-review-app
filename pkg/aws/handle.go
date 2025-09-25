package aws

import (
	"context"
	"log/slog"

	"github.com/shurcooL/githubv4"
	"github.com/suzuki-shunsuke/enforce-pr-review-app/pkg/config"
	"github.com/suzuki-shunsuke/slog-error/slogerr"
)

func (h *Handler) handle(ctx context.Context, logger *slog.Logger, req *Request) error { //nolint:cyclop,funlen
	logger.Info("Starting a request", "request", req)
	defer logger.Info("Ending a request", "request", req)

	// Validate the request
	ev, err := h.validateRequest(logger, req)
	if err != nil {
		slogerr.WithError(logger, err).Warn("validate request")
		return nil
	}
	if ev.GetAction() == "edited" {
		logger.Info("ignore the event because the action is 'edited'")
		return nil
	}
	state := ev.GetReview().GetState()
	if state == "commented" || state == "pending" {
		logger.Info("ignore the event because the state is '" + state + "'")
		return nil
	}

	checkName := githubv4.String(h.config.CheckName)

	// Get repository ID for GraphQL mutation
	repoID := githubv4.String(ev.GetRepo().GetNodeID())
	headSha := githubv4.GitObjectID(ev.GetPullRequest().GetHead().GetSHA())

	// Run validation
	var conclusion githubv4.CheckConclusionState
	var title githubv4.String
	result := h.validate(ctx, ev)
	switch result.State {
	case config.StateApproved:
		conclusion = githubv4.CheckConclusionStateSuccess
		title = githubv4.String("Approved")
	case config.StateApprovalIsRequired:
		conclusion = githubv4.CheckConclusionStateFailure
		title = githubv4.String("Approvals are required")
	case config.StateTwoApprovalsAreRequired:
		conclusion = githubv4.CheckConclusionStateFailure
		title = githubv4.String("Two approvals are required")
	}
	if result.Error != "" {
		conclusion = githubv4.CheckConclusionStateFailure
		title = githubv4.String("Internal Error")
	}
	result.TrustedApps = h.config.TrustedApps
	result.TrustedMachineUsers = h.config.TrustedMachineUsers
	result.UntrustedMachineUsers = h.config.UntrustedMachineUsers
	s, err := summarize(result, h.config.BuiltTemplates)
	if err != nil {
		slogerr.WithError(h.logger, err).Error("summarize the result")
		// TODO use the default template
		conclusion = githubv4.CheckConclusionStateFailure
		title = githubv4.String("Internal Error")
	}

	// Create final check run with conclusion
	completedStatus := githubv4.RequestableCheckStatusStateCompleted
	finalCheckRunInput := githubv4.CreateCheckRunInput{
		RepositoryID: repoID,
		HeadSha:      headSha,
		Name:         checkName,
		Status:       &completedStatus,
		Conclusion:   &conclusion,
		Output: &githubv4.CheckRunOutput{
			Title:   title,
			Summary: githubv4.String(s),
		},
	}

	if err := h.gh.CreateCheckRun(ctx, finalCheckRunInput); err != nil {
		slogerr.WithError(logger, err).Error("create final check run")
	}
	return nil
}
