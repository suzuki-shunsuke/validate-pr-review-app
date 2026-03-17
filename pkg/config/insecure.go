package config

type Insecure struct {
	AllowUnsignedCommits       bool     `json:"allow_unsigned_commits,omitempty" yaml:"allow_unsigned_commits"`
	UnsignedCommitApps         []string `json:"unsigned_commit_apps,omitempty" yaml:"unsigned_commit_apps"`
	UnsignedCommitMachineUsers []string `json:"unsigned_commit_machine_users,omitempty" yaml:"unsigned_commit_machine_users"`
}
