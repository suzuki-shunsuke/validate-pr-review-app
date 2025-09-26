package controller

import (
	"log/slog"

	"github.com/suzuki-shunsuke/validate-pr-review-app/pkg/github"
)

func ignore(logger *slog.Logger, ev *github.PullRequestReviewEvent) bool {
	if ev.GetAction() == "edited" {
		logger.Info("ignore the event because the action is 'edited'")
		return true
	}
	state := ev.GetReview().GetState()
	if state == "commented" || state == "pending" {
		logger.Info("ignore the event because the state is '" + state + "'")
		return true
	}
	return false
}
