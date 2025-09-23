package github

type PullRequestCommit struct {
	Commit *Commit `json:"commit"`
}

type UntrustedCommit struct {
	Login                  string
	SHA                    string
	IsUntrustedMachineUser bool
	IsUntrustedApp         bool
	InvalidSign            *Signature
	NotLinkedToUser        bool
}

// ValidateUntrusted checks if the commit is untrusted.
// It returns nil if the commit is trusted, otherwise returns the reason why it is untrusted.
// A commit is untrusted if it is not linked to any GitHub user, if its sign is invalid the commiter is untrusted app or untrusted machine user.
func (c *PullRequestCommit) ValidateUntrusted(trustedApps, trustedMachineUsers, untrustedMachineUsers map[string]struct{}) *UntrustedCommit {
	commit := c.Commit
	user := commit.User()
	login := user.Login
	sha := commit.OID
	if !commit.Linked() {
		return &UntrustedCommit{
			NotLinkedToUser: true,
			SHA:             sha,
		}
	}
	sig := commit.Signature
	if sig == nil || !sig.IsValid {
		return &UntrustedCommit{
			Login:       login,
			SHA:         sha,
			InvalidSign: sig,
		}
	}
	if user.IsApp() {
		if _, ok := trustedApps[login]; ok {
			return nil
		}
		return &UntrustedCommit{
			Login:          login,
			SHA:            sha,
			IsUntrustedApp: true,
		}
	}
	if user.IsTrustedUser(trustedMachineUsers, untrustedMachineUsers) {
		return nil
	}
	return &UntrustedCommit{
		Login:                  login,
		SHA:                    sha,
		IsUntrustedMachineUser: true,
	}
}

type Commit struct {
	OID       string     `json:"oid"`
	Committer *Committer `json:"committer"`
	Author    *Committer `json:"author"`
	Signature *Signature `json:"signature"`
}

type Signature struct {
	IsValid bool   `json:"isValid"`
	State   string `json:"state"`
}

func (c *Commit) User() *User {
	if c.Committer.User != nil {
		return c.Committer.User
	}
	return c.Author.User
}

func (c *Commit) Linked() bool {
	return c.User().Login != ""
}
