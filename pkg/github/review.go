package github

import "github.com/shurcooL/githubv4"

type Review struct {
	Author    *User             `json:"author"`
	State     string            `json:"state"`
	Commit    *ReviewCommit     `json:"commit"`
	CreatedAt githubv4.DateTime `json:"createdAt"`
}

type ReviewCommit struct {
	OID string `json:"oid"`
}

// Ignored returns true if the review should be ignored.
// A review is ignored if it is not an approval or if it is not for the latest commit.
func (r *Review) Ignored(latestSHA string) bool {
	return r.State != "APPROVED" || r.Commit.OID != latestSHA
}

// IgnoredApproval represents an approval that is ignored.
// It contains the login of the approver and the reason why the approval is ignored.
type IgnoredApproval struct {
	Login                  string
	IsApp                  bool
	IsUntrustedMachineUser bool
}

// ValidateIgnored checks if the approval should be ignored.
// It returns nil if the approval is valid, otherwise returns the reason why it is ignored.
// An approval is ignored if the approver is an app or an untrusted machine user.
func (r *Review) ValidateIgnored(trustedMachineUsers, untrustedMachineUsers map[string]struct{}) *IgnoredApproval {
	if r.Author.IsApp() {
		return &IgnoredApproval{
			Login: r.Author.Login,
			IsApp: true,
		}
	}
	if r.Author.IsTrustedUser(trustedMachineUsers, untrustedMachineUsers) {
		return nil
	}
	return &IgnoredApproval{
		Login:                  r.Author.Login,
		IsUntrustedMachineUser: true,
	}
}
