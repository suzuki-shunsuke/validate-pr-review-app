package validation_test

import (
	"testing"

	"github.com/suzuki-shunsuke/enforce-pr-review-app/pkg/validation"
)

func TestNew(t *testing.T) {
	t.Parallel()
	controller := validation.New()
	if controller == nil {
		t.Error("New() returned nil")
	}
	// Check if it's the correct type
	if controller == (*validation.Controller)(nil) {
		t.Error("New() returned nil Controller")
	}
}
