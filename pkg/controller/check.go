package controller

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"log/slog"

	"github.com/shurcooL/githubv4"
	"github.com/suzuki-shunsuke/slog-error/slogerr"
	"github.com/suzuki-shunsuke/validate-pr-review-app/pkg/github"
	"github.com/suzuki-shunsuke/validate-pr-review-app/pkg/validation"
)

func (c *Controller) newCheckRunInput(logger *slog.Logger, ev *github.PullRequestReviewEvent, result *validation.Result) githubv4.CreateCheckRunInput {
	result.Version = c.input.Version
	var conclusion githubv4.CheckConclusionState
	var title githubv4.String
	switch result.State {
	case validation.StateApproved:
		conclusion = githubv4.CheckConclusionStateSuccess
		title = githubv4.String("Approved")
	case validation.StateApprovalIsRequired:
		conclusion = githubv4.CheckConclusionStateFailure
		title = githubv4.String("Approvals are required")
	case validation.StateTwoApprovalsAreRequired:
		conclusion = githubv4.CheckConclusionStateFailure
		title = githubv4.String("Two approvals are required")
	}
	if result.Error != "" {
		conclusion = githubv4.CheckConclusionStateFailure
		title = githubv4.String("Internal Error")
	}
	result.TrustedApps = c.input.Config.TrustedApps
	result.TrustedMachineUsers = c.input.Config.TrustedMachineUsers
	result.UntrustedMachineUsers = c.input.Config.UntrustedMachineUsers
	s, err := summarize(result, c.input.Config.BuiltTemplates)
	if err != nil {
		slogerr.WithError(logger, err).Error("summarize the result")
		// TODO use the default template
		conclusion = githubv4.CheckConclusionStateFailure
		title = githubv4.String("Internal Error")
	}

	// Create final check run with conclusion
	completedStatus := githubv4.RequestableCheckStatusStateCompleted
	return githubv4.CreateCheckRunInput{
		RepositoryID: githubv4.String(ev.GetRepo().GetNodeID()),
		HeadSha:      githubv4.GitObjectID(ev.GetPullRequest().GetHead().GetSHA()),
		Name:         githubv4.String(c.input.Config.CheckName),
		Status:       &completedStatus,
		Conclusion:   &conclusion,
		Output: &githubv4.CheckRunOutput{
			Title:   title,
			Summary: githubv4.String(s),
		},
	}
}

func summarize(result *validation.Result, templates map[string]*template.Template) (string, error) {
	var key string
	if result.Error != "" {
		key = "error"
	} else {
		key = string(result.State)
	}
	tpl, ok := templates[key]
	if !ok {
		return "", errors.New("summary template is not found")
	}
	var buf bytes.Buffer
	if err := tpl.Execute(&buf, result); err != nil {
		return "", fmt.Errorf("execute summary template: %w", err)
	}
	return buf.String(), nil
}
