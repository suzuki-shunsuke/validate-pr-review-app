package github

import (
	"path"
	"strings"
)

type User struct {
	Login        string `json:"login"`
	ResourcePath string `json:"resourcePath"`
}

type ApprovalMarkIgnoredReason string

const (
	ApprovalMarkIgnoredReasonOK                   ApprovalMarkIgnoredReason = "ok"
	ApprovalMarkIgnoredReasonApp                  ApprovalMarkIgnoredReason = "app"
	ApprovalMarkIgnoredReasonUntrustedMachineUser ApprovalMarkIgnoredReason = "untrusted_machine_user"
)

func (u *User) MarkIgnored(trustedMachineUsers, untrustedMachineUsers map[string]struct{}) ApprovalMarkIgnoredReason {
	if u.IsApp() {
		return ApprovalMarkIgnoredReasonApp
	}
	if u.IsTrustedUser(trustedMachineUsers, untrustedMachineUsers) {
		return ApprovalMarkIgnoredReasonOK
	}
	return ApprovalMarkIgnoredReasonUntrustedMachineUser
}

func (u *User) IsApp() bool {
	return strings.HasPrefix(u.ResourcePath, "/apps/") || strings.HasSuffix(u.Login, "[bot]")
}

func (u *User) IsTrustedUser(trustedMachineUsers, untrustedMachineUsers map[string]struct{}) bool {
	if _, ok := trustedMachineUsers[u.Login]; ok {
		return true
	}
	for pattern := range untrustedMachineUsers {
		matched, err := path.Match(pattern, u.Login)
		if err != nil { // TODO error handling
			continue
		}
		if matched {
			return false
		}
	}
	return true
}
