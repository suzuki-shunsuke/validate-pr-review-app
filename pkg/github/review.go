package github

import v4 "github.com/suzuki-shunsuke/enforce-pr-review-app/pkg/github/v4"

type Review struct {
	Author *User  `json:"author"`
	State  string `json:"state"`
}

func newReview(v *v4.Review) *Review {
	return &Review{
		Author: newUser(v.Author),
		State:  v.State,
	}
}

// IgnoredApproval represents an approval that is ignored.
// It contains the login of the approver and the reason why the approval is ignored.
type IgnoredApproval struct {
	Login                  string
	IsApp                  bool
	IsUntrustedMachineUser bool
}
