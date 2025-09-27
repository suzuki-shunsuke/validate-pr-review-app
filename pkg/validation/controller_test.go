package validation_test

import (
	"testing"

	"github.com/suzuki-shunsuke/validate-pr-review-app/pkg/validation"
)

func TestNew(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		inputNew *validation.InputNew
	}{
		{
			name:     "creates new controller with empty input",
			inputNew: &validation.InputNew{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := validation.New(tt.inputNew)
			if got == nil {
				t.Error("New() returned nil")
			}
		})
	}
}
