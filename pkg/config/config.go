package config

import (
	_ "embed"
	"fmt"
	"html/template"
	"path"
	"strings"
)

type Config struct {
	AppID          int64  `yaml:"app_id"`
	InstallationID int64  `yaml:"installation_id"`
	AWS            *AWS   `yaml:"aws"`
	CheckName      string `yaml:"check_name"`
	Trust          *Trust `yaml:"trust"`
	Templates      map[string]string             `yaml:"templates"`
	BuiltTemplates map[string]*template.Template `yaml:"-"`
	LogLevel       string                        `yaml:"log_level"`
	// Repositories   []*Repository                 `yaml:"repositories"`
}

type Trust struct {
	TrustedMachineUsers         []string            `yaml:"trusted_machine_users"`
	UntrustedMachineUsers       []string            `yaml:"untrusted_machine_users"`
	TrustedApps                 []string            `yaml:"trusted_apps"`
	UniqueTrustedMachineUsers   map[string]struct{} `yaml:"-"`
	UniqueUntrustedMachineUsers map[string]struct{} `yaml:"-"`
	UniqueTrustedApps           map[string]struct{} `yaml:"-"`
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
	if c.Trust == nil {
		c.Trust = &Trust{}
	}
	if c.Trust.TrustedApps == nil {
		c.Trust.TrustedApps = []string{
			"dependabot[bot]",
			"renovate[bot]",
		}
	} else {
		for i, v := range c.Trust.TrustedApps {
			// Append [bot] suffix if not exists
			if !strings.HasSuffix(v, "[bot]") {
				c.Trust.TrustedApps[i] = v + "[bot]"
			}
		}
	}
	c.Trust.UniqueTrustedApps = make(map[string]struct{}, len(c.Trust.TrustedApps))
	for _, app := range c.Trust.TrustedApps {
		// TODO validate the app name
		if app == "" {
			continue
		}
		c.Trust.UniqueTrustedApps[app] = struct{}{}
	}
	c.Trust.UniqueTrustedMachineUsers = make(map[string]struct{}, len(c.Trust.TrustedMachineUsers))
	for _, user := range c.Trust.TrustedMachineUsers {
		// TODO validate the user name
		if user == "" {
			continue
		}
		c.Trust.UniqueTrustedMachineUsers[user] = struct{}{}
	}
	c.Trust.UniqueUntrustedMachineUsers = make(map[string]struct{}, len(c.Trust.UntrustedMachineUsers))
	for _, user := range c.Trust.UntrustedMachineUsers {
		// TODO validate the user name
		if user == "" {
			continue
		}
		c.Trust.UniqueUntrustedMachineUsers[user] = struct{}{}
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
	for pattern := range c.Trust.UniqueUntrustedMachineUsers {
		if _, err := path.Match(pattern, "foo"); err != nil {
			return fmt.Errorf("invalid untrusted machine user pattern %q: %w", pattern, err)
		}
	}
	return nil
}
