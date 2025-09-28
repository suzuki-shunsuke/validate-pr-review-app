package config

import (
	"errors"
	"fmt"
	"path"
)

func (c *Config) GetRepo(repo string) *Repository {
	for _, r := range c.Repositories {
		if r.Match(repo) {
			return r
		}
	}
	return nil
}

func (c *Config) initRepos() error {
	for _, repo := range c.Repositories {
		if err := repo.Validate(); err != nil {
			return fmt.Errorf("validate a repository config: %w", err)
		}
		if repo.Trust.TrustedApps == nil {
			repo.Trust.TrustedApps = c.Trust.TrustedApps
		}
		if repo.Trust.UntrustedMachineUsers == nil {
			repo.Trust.UntrustedMachineUsers = c.Trust.UntrustedMachineUsers
		}
		if repo.Trust.TrustedMachineUsers == nil {
			repo.Trust.TrustedMachineUsers = c.Trust.TrustedMachineUsers
		}
		repo.Trust.Init()
	}
	return nil
}

type Repository struct {
	Repositories []string `json:"repositories" yaml:"repositories"`
	Trust        *Trust   `json:"trust" yaml:"trust"`
	Ignored      bool     `json:"ignored,omitempty" yaml:"ignored"`
}

func (r *Repository) Validate() error {
	if len(r.Repositories) == 0 {
		return errors.New("repositories is required")
	}
	if r.Trust == nil {
		return errors.New("trust is required")
	}
	for _, pattern := range r.Repositories {
		if _, err := path.Match(pattern, "suzuki-shunsuke/validate-pr-review-app"); err != nil {
			return fmt.Errorf("invalid repository pattern %q: %w", pattern, err)
		}
	}
	return nil
}

func (r *Repository) Match(repo string) bool {
	for _, pattern := range r.Repositories {
		matched, err := path.Match(pattern, repo)
		if err != nil {
			continue
		}
		if matched {
			return true
		}
	}
	return false
}
