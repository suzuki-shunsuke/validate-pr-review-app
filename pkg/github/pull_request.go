package github

type PullRequest struct {
	HeadSHA   string              `json:"sha"`
	Approvers map[string]*User `json:"approvers"`
	Commits   []*Commit           `json:"commits"`
}

type Author struct {
	Login                string
	UntrustedMachineUser bool
	UntrustedApp         bool
}
