package config

import (
	_ "embed"
	"fmt"
	"html/template"
	"path"
	"strings"
)

type Config struct {
	AppID                       int64                         `yaml:"app_id"`
	InstallationID              int64                         `yaml:"installation_id"`
	AWS                         *AWS                          `yaml:"aws"`
	CheckName                   string                        `yaml:"check_name"`
	TrustedApps                 []string                      `yaml:"trusted_apps"`
	TrustedMachineUsers         []string                      `yaml:"trusted_machine_users"`
	UntrustedMachineUsers       []string                      `yaml:"untrusted_machine_users"`
	UniqueTrustedApps           map[string]struct{}           `yaml:"-"`
	UniqueTrustedMachineUsers   map[string]struct{}           `yaml:"-"`
	UniqueUntrustedMachineUsers map[string]struct{}           `yaml:"-"`
	Templates                   map[string]string             `yaml:"templates"`
	BuiltTemplates              map[string]*template.Template `yaml:"-"`
	LogLevel                    string                        `yaml:"log_level"`
}

type AWS struct {
	SecretID             string `yaml:"secret_id"`
	UseLambdaFunctionURL bool   `yaml:"use_lambda_function_url"`
}

var (
	//go:embed templates/footer.md
	templateFooter []byte
	//go:embed templates/no_approval.md
	templateNoApproval []byte
	//go:embed templates/require_two_approvals.md
	templateRequireTwoApprovals []byte
	//go:embed templates/settings.md
	templateSettings []byte
	//go:embed templates/approved.md
	templateApproved []byte
	//go:embed templates/error.md
	templateError []byte
)

func (c *Config) Init() error { //nolint:cyclop
	if c.TrustedApps == nil {
		c.TrustedApps = []string{
			"dependabot[bot]",
			"renovate[bot]",
		}
	} else {
		for i, v := range c.TrustedApps {
			// Append [bot] suffix if not exists
			if !strings.HasSuffix(v, "[bot]") {
				c.TrustedApps[i] = v + "[bot]"
			}
		}
	}
	c.UniqueTrustedApps = make(map[string]struct{}, len(c.TrustedApps))
	for _, app := range c.TrustedApps {
		// TODO validate the app name
		if app == "" {
			continue
		}
		c.UniqueTrustedApps[app] = struct{}{}
	}
	c.UniqueTrustedMachineUsers = make(map[string]struct{}, len(c.TrustedMachineUsers))
	for _, user := range c.TrustedMachineUsers {
		// TODO validate the user name
		if user == "" {
			continue
		}
		c.UniqueTrustedMachineUsers[user] = struct{}{}
	}
	c.UniqueUntrustedMachineUsers = make(map[string]struct{}, len(c.UntrustedMachineUsers))
	for _, user := range c.UntrustedMachineUsers {
		// TODO validate the user name
		if user == "" {
			continue
		}
		c.UniqueUntrustedMachineUsers[user] = struct{}{}
	}
	if c.CheckName == "" {
		c.CheckName = "validate-review"
	}
	if err := c.initTemplates(); err != nil {
		return err
	}
	if err := c.testUntrustedMachineUsers(); err != nil {
		return err
	}
	return c.testTemplate()
}

func (c *Config) testUntrustedMachineUsers() error {
	for pattern := range c.UniqueUntrustedMachineUsers {
		if _, err := path.Match(pattern, "foo"); err != nil {
			return fmt.Errorf("invalid untrusted machine user pattern %q: %w", pattern, err)
		}
	}
	return nil
}
