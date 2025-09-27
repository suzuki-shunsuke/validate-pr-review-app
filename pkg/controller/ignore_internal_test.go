//nolint:funlen
package controller

import (
	"log/slog"
	"testing"

	"github.com/google/go-github/v75/github"
)

func Test_ignore(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		event    *github.PullRequestReviewEvent
		expected bool
	}{
		{
			name: "ignore edited action",
			event: &github.PullRequestReviewEvent{
				Action: github.Ptr("edited"),
				Review: &github.PullRequestReview{
					State: github.Ptr("approved"),
				},
			},
			expected: true,
		},
		{
			name: "ignore commented state",
			event: &github.PullRequestReviewEvent{
				Action: github.Ptr("submitted"),
				Review: &github.PullRequestReview{
					State: github.Ptr("commented"),
				},
			},
			expected: true,
		},
		{
			name: "ignore pending state",
			event: &github.PullRequestReviewEvent{
				Action: github.Ptr("submitted"),
				Review: &github.PullRequestReview{
					State: github.Ptr("pending"),
				},
			},
			expected: true,
		},
		{
			name: "do not ignore approved state",
			event: &github.PullRequestReviewEvent{
				Action: github.Ptr("submitted"),
				Review: &github.PullRequestReview{
					State: github.Ptr("approved"),
				},
			},
			expected: false,
		},
		{
			name: "do not ignore changes_requested state",
			event: &github.PullRequestReviewEvent{
				Action: github.Ptr("submitted"),
				Review: &github.PullRequestReview{
					State: github.Ptr("changes_requested"),
				},
			},
			expected: false,
		},
		{
			name: "do not ignore dismissed state",
			event: &github.PullRequestReviewEvent{
				Action: github.Ptr("dismissed"),
				Review: &github.PullRequestReview{
					State: github.Ptr("dismissed"),
				},
			},
			expected: false,
		},
		{
			name: "handle nil action",
			event: &github.PullRequestReviewEvent{
				Action: nil,
				Review: &github.PullRequestReview{
					State: github.Ptr("approved"),
				},
			},
			expected: false,
		},
		{
			name: "handle nil review state",
			event: &github.PullRequestReviewEvent{
				Action: github.Ptr("submitted"),
				Review: &github.PullRequestReview{
					State: nil,
				},
			},
			expected: false,
		},
		{
			name: "handle nil review",
			event: &github.PullRequestReviewEvent{
				Action: github.Ptr("submitted"),
				Review: nil,
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
