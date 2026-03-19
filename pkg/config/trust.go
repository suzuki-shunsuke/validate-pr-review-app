package config

import (
	"fmt"
	"path"
	"strings"
)

type Trust struct {
	TrustedMachineUsers         []string            `json:"trusted_machine_users,omitempty" yaml:"trusted_machine_users"`
	UntrustedMachineUsers       []string            `json:"untrusted_machine_users,omitempty" yaml:"untrusted_machine_users"`
	TrustedApps                 []string            `json:"trusted_apps,omitempty" yaml:"trusted_apps"`
	UniqueTrustedMachineUsers   map[string]struct{} `json:"-" yaml:"-"`
	UniqueUntrustedMachineUsers map[string]struct{} `json:"-" yaml:"-"`
	UniqueTrustedApps           map[string]struct{} `json:"-" yaml:"-"`
}

func (t *Trust) Validate() error {
	if err := validateLoginNames(t.TrustedApps, "trusted_apps"); err != nil {
		return err
	}
	if err := validateLoginNames(t.TrustedMachineUsers, "trusted_machine_users"); err != nil {
		return err
	}
	return nil
}

func (t *Trust) Init() {
	if t.TrustedApps == nil {
		t.TrustedApps = []string{
			"dependabot[bot]",
			"renovate[bot]",
		}
	} else {
		for i, v := range t.TrustedApps {
			// Append [bot] suffix if not exists
			if !strings.HasSuffix(v, "[bot]") {
				t.TrustedApps[i] = v + "[bot]"
			}
		}
	}
	t.UniqueTrustedApps = make(map[string]struct{}, len(t.TrustedApps))
	for _, app := range t.TrustedApps {
		if app == "" {
			continue
		}
		t.UniqueTrustedApps[app] = struct{}{}
	}
	t.UniqueTrustedMachineUsers = make(map[string]struct{}, len(t.TrustedMachineUsers))
	for _, user := range t.TrustedMachineUsers {
		if user == "" {
			continue
		}
		t.UniqueTrustedMachineUsers[user] = struct{}{}
	}
	t.UniqueUntrustedMachineUsers = make(map[string]struct{}, len(t.UntrustedMachineUsers))
	for _, user := range t.UntrustedMachineUsers {
		if user == "" {
			continue
		}
		t.UniqueUntrustedMachineUsers[user] = struct{}{}
	}
}

func (c *Config) testUntrustedMachineUsers() error {
	for pattern := range c.Trust.UniqueUntrustedMachineUsers {
		if _, err := path.Match(pattern, "foo"); err != nil {
			return fmt.Errorf("invalid untrusted machine user pattern %q: %w", pattern, err)
		}
	}
	return nil
}
