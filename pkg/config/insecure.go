package config

import "errors"

type Insecure struct {
	AllowUnsignedCommits       *bool    `json:"allow_unsigned_commits,omitempty" yaml:"allow_unsigned_commits"`
	UnsignedCommitApps         []string `json:"unsigned_commit_apps,omitempty" yaml:"unsigned_commit_apps"`
	UnsignedCommitMachineUsers []string `json:"unsigned_commit_machine_users,omitempty" yaml:"unsigned_commit_machine_users"`
}

func (i *Insecure) Validate() error {
	if i.AllowUnsignedCommits != nil && *i.AllowUnsignedCommits {
		if len(i.UnsignedCommitApps) > 0 || len(i.UnsignedCommitMachineUsers) > 0 {
			return errors.New("allow_unsigned_commits cannot be used together with unsigned_commit_apps or unsigned_commit_machine_users")
		}
	}
	if err := validateLoginNames(i.UnsignedCommitApps, "unsigned_commit_apps"); err != nil {
		return err
	}
	if err := validateLoginNames(i.UnsignedCommitMachineUsers, "unsigned_commit_machine_users"); err != nil {
		return err
	}
	return nil
}
