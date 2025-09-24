package github

import v4 "github.com/suzuki-shunsuke/enforce-pr-review-app/pkg/github/v4"

type UntrustedCommit struct {
	Login                  string
	SHA                    string
	IsUntrustedMachineUser bool
	IsUntrustedApp         bool
	InvalidSign            *Signature
	NotLinkedToUser        bool
}

func (c *UntrustedCommit) Message() string {
	if c == nil {
		return ""
	}
	if c.NotLinkedToUser {
		return "The commit is not linked to any GitHub user."
	}
	if c.IsUntrustedApp {
		return "The committer is an untrusted app."
	}
	if c.IsUntrustedMachineUser {
		return "The committer is an untrusted machine user."
	}
	if c.InvalidSign == nil {
		return "The commit isn't signed."
	}
	if !c.InvalidSign.IsValid {
		return "The commit sign is invalid. " + c.InvalidSign.State
	}
	return ""
}

type Commit struct {
	SHA       string        `json:"oid"`
	Committer *User         `json:"committer"`
	Signature *v4.Signature `json:"signature"`
}

func (c *Commit) Linked() bool {
	return c.Committer.Login != ""
}
