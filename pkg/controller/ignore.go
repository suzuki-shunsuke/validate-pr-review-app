package controller

import (
	"log/slog"
)

func ignore(logger *slog.Logger, ev *Event) bool {
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
