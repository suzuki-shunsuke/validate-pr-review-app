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
	if !c.Commit.Linked() {
		return &UntrustedCommit{
			NotLinkedToUser: true,
			SHA:             c.Commit.OID,
		}
	}
	if c.Commit.Signature == nil || !c.Commit.Signature.IsValid {
		return &UntrustedCommit{
			Login:       c.Commit.Login(),
			SHA:         c.Commit.OID,
			InvalidSign: c.Commit.Signature,
		}
	}
	if c.Commit.User().IsApp() {
		if _, ok := trustedApps[c.Commit.Login()]; ok {
			return nil
		}
		return &UntrustedCommit{
			Login:          c.Commit.Login(),
			SHA:            c.Commit.OID,
			IsUntrustedApp: true,
		}
	}
	if c.Commit.User().IsTrustedUser(trustedMachineUsers, untrustedMachineUsers) {
		return nil
	}
	return &UntrustedCommit{
		Login:                  c.Commit.Login(),
		SHA:                    c.Commit.OID,
		IsUntrustedMachineUser: true,
	}
}

func (c *PullRequestCommit) Signature() *Signature {
	return c.Commit.Signature
}

func (c *PullRequestCommit) User() *User {
	return c.Commit.User()
}

func (c *PullRequestCommit) SHA() string {
	return c.Commit.OID
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

func (c *Commit) Login() string {
	return c.User().Login
}

func (c *Commit) Linked() bool {
	return c.Login() != ""
}
