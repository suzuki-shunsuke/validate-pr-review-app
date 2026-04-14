package github

import v4 "github.com/suzuki-shunsuke/validate-pr-review-app/pkg/github/v4"

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
	SHA                     string        `json:"oid"`
	Committer               *User         `json:"committer"`
	Signature               *v4.Signature `json:"signature"`
	Parents                 []string      `json:"parents"`
	ChangedFilesIfAvailable *int          `json:"changed_files_if_available"`
	IsAllowedMergeCommit    bool          `json:"is_allowed_merge_commit"`
}

func (c *Commit) Linked() bool {
	return c.Committer != nil && c.Committer.Login != ""
}

func newCommit(pc *v4.PullRequestCommit) *Commit {
	var parents []string
	if pc.Commit.Parents != nil {
		parents = make([]string, len(pc.Commit.Parents.Nodes))
		for i, p := range pc.Commit.Parents.Nodes {
			parents[i] = p.OID
		}
	}
	return &Commit{
		SHA:                     pc.Commit.OID,
		Committer:               newUser(pc.Commit.User()),
		Signature:               pc.Commit.Signature,
		Parents:                 parents,
		ChangedFilesIfAvailable: pc.Commit.ChangedFilesIfAvailable,
	}
}
