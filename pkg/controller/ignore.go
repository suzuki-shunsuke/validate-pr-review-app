package controller

import (
	"log/slog"
)

func ignore(logger *slog.Logger, ev *Event) bool {
	// For pull_request events, only process "synchronize" action.
	if ev.EventType == eventPullRequest {
		if ev.Action != "synchronize" {
			logger.Info("ignore the pull_request event because the action is not 'synchronize'", "action", ev.Action)
			return true
		}
		return false
	}
	if ev.Action == "edited" {
		logger.Info("ignore the event because the action is 'edited'")
		return true
	}
	state := ev.ReviewState
	if state == "commented" || state == "pending" {
		logger.Info("ignore the event because the state is '" + state + "'")
		return true
	}
	return false
}
