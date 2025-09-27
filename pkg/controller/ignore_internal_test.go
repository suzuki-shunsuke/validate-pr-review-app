//nolint:funlen
package controller

import (
	"log/slog"
	"testing"
)

func Test_ignore(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		event    *Event
		expected bool
	}{
		{
			name: "ignore edited action",
			event: &Event{
				Action:      "edited",
				ReviewState: "approved",
			},
			expected: true,
		},
		{
			name: "ignore commented state",
			event: &Event{
				Action:      "submitted",
				ReviewState: "commented",
			},
			expected: true,
		},
		{
			name: "ignore pending state",
			event: &Event{
				Action:      "submitted",
				ReviewState: "pending",
			},
			expected: true,
		},
		{
			name: "do not ignore approved state",
			event: &Event{
				Action:      "submitted",
				ReviewState: "approved",
			},
			expected: false,
		},
		{
			name: "do not ignore changes_requested state",
			event: &Event{
				Action:      "submitted",
				ReviewState: "changes_requested",
			},
			expected: false,
		},
		{
			name: "do not ignore dismissed state",
			event: &Event{
				Action:      "dismissed",
				ReviewState: "dismissed",
			},
			expected: false,
		},
		{
			name: "handle empty action",
			event: &Event{
				Action:      "",
				ReviewState: "approved",
			},
			expected: false,
		},
		{
			name: "handle empty review state",
			event: &Event{
				Action:      "submitted",
				ReviewState: "",
			},
			expected: false,
		},
		{
			name: "handle empty fields",
			event: &Event{
				Action:      "",
				ReviewState: "",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			logger := slog.Default()
			result := ignore(logger, tt.event)

			if result != tt.expected {
				t.Errorf("ignore() = %v, want %v", result, tt.expected)
			}
		})
	}
}
