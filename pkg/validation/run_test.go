//nolint:funlen,maintidx
package validation_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/suzuki-shunsuke/enforce-pr-review-app/pkg/validation"
)

func TestController_Run(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		input    *validation.Input
		expected *validation.Result
	}{}

	ctrl := &validation.Controller{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := ctrl.Run(nil, tt.input)

			if diff := cmp.Diff(tt.expected, result); diff != "" {
				t.Errorf("validation result mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
