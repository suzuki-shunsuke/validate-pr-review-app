package github

type PullRequest struct {
	HeadSHA           string                      `json:"sha"`
	Approvers         map[string]*User            `json:"approvers"`
	ApproversByCommit map[string]map[string]*User `json:"approvers_by_commit"`
	Commits           []*Commit                   `json:"commits"`
}

type Author struct {
	Login                string
	UntrustedMachineUser bool
	UntrustedApp         bool
}
