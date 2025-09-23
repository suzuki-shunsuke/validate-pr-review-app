package validation_test

import (
	"testing"

	"github.com/suzuki-shunsuke/enforce-pr-review-app/pkg/validation"
)

func TestNew(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
	}{
		{
			name: "creates new controller",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := validation.New()
			if got == nil {
				t.Error("New() returned nil")
			}
		})
	}
}
