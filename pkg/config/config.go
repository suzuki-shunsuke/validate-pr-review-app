package config

import (
	_ "embed"
	"fmt"
	"io"
	"path"
	"strings"
	"text/template"

	"github.com/suzuki-shunsuke/validate-pr-review-app/pkg/github"
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

func (c *Config) initTemplates() error {
	defaultTemplates := map[string]string{
		"footer":                string(templateFooter),
		"settings":              string(templateSettings),
		"approved":              string(templateApproved),
		"no_approval":           string(templateNoApproval),
		"require_two_approvals": string(templateRequireTwoApprovals),
		"error":                 string(templateError),
	}
	if c.Templates == nil {
		c.Templates = map[string]string{}
	}
	for name, tpl := range defaultTemplates {
		if _, ok := c.Templates[name]; !ok {
			c.Templates[name] = tpl
		}
	}
	var define string
	for k, v := range c.Templates {
		define += `{{define "` + k + `"}}` + v + "{{end}}"
	}
	keys := []string{
		"no_approval",
		"approved",
		"require_two_approvals",
		"error",
	}
	templates := make(map[string]*template.Template, len(keys))
	for _, k := range keys {
		tpl := c.Templates[k] + define
		tplParsed, err := template.New("_").Parse(tpl)
		if err != nil {
			return fmt.Errorf("parse the template %s: %w", k, err)
		}
		templates[k] = tplParsed
	}
	c.BuiltTemplates = templates
	return nil
}

func (c *Config) testUntrustedMachineUsers() error {
	for pattern := range c.UniqueUntrustedMachineUsers {
		if _, err := path.Match(pattern, "foo"); err != nil {
			return fmt.Errorf("invalid untrusted machine user pattern %q: %w", pattern, err)
		}
	}
	return nil
}

func (c *Config) testTemplate() error {
	// TODO add test cases
	result := &Result{}
	for key, tpl := range c.BuiltTemplates {
		if err := tpl.Execute(io.Discard, result); err != nil {
			return fmt.Errorf("test template %s: %w", key, err)
		}
	}
	return nil
}

type Result struct {
	Error        string
	State        State
	Approvers    []string
	SelfApprover string
	// app or untrusted machine user approvals
	IgnoredApprovers []*github.IgnoredApproval
	// app
	// untrusted machine user
	// not linked to any GitHub user
	// not signed commits
	UntrustedCommits []*github.UntrustedCommit
	// settings
	TrustedApps           []string
	UntrustedMachineUsers []string
	TrustedMachineUsers   []string
	Version               string
}

type State string

const (
	StateApproved                State = "approved"
	StateApprovalIsRequired      State = "no_approval"
	StateTwoApprovalsAreRequired State = "require_two_approvals"
)
