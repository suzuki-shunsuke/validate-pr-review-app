package config

import (
	"fmt"
	"path"
	"strings"
)

type Trust struct {
	UntrustedMachineUsers []string            `json:"untrusted_machine_users,omitempty" yaml:"untrusted_machine_users"`
	TrustedApps           []string            `json:"trusted_apps,omitempty" yaml:"trusted_apps"`
	UniqueTrustedApps     map[string]struct{} `json:"-" yaml:"-"`
}

func (t *Trust) Validate() error {
	if err := validateLoginNames(t.TrustedApps, "trusted_apps"); err != nil {
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
}

func (c *Config) testUntrustedMachineUsers() error {
	for _, pattern := range c.Trust.UntrustedMachineUsers {
		p := strings.TrimPrefix(pattern, "!")
		if _, err := path.Match(p, "foo"); err != nil {
			return fmt.Errorf("invalid untrusted machine user pattern %q: %w", pattern, err)
		}
	}
	return nil
}
