package config

import (
	_ "embed"
	"fmt"
	"io"
	"path"
	"text/template"

	"github.com/suzuki-shunsuke/enforce-pr-review-app/pkg/github"
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
	SecretID string `yaml:"secret_id"`
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
	//go:embed templates/two_approvals.md
	templateTwoApprovals []byte
)

func (cfg *Config) Init() error {
	cfg.UniqueTrustedApps = make(map[string]struct{}, len(cfg.TrustedApps))
	for _, app := range cfg.TrustedApps {
		// TODO validate the app name
		if app == "" {
			continue
		}
		cfg.UniqueTrustedApps[app] = struct{}{}
	}
	cfg.UniqueTrustedMachineUsers = make(map[string]struct{}, len(cfg.TrustedMachineUsers))
	for _, user := range cfg.TrustedMachineUsers {
		// TODO validate the user name
		if user == "" {
			continue
		}
		cfg.UniqueTrustedMachineUsers[user] = struct{}{}
	}
	cfg.UniqueUntrustedMachineUsers = make(map[string]struct{}, len(cfg.UntrustedMachineUsers))
	for _, user := range cfg.UntrustedMachineUsers {
		// TODO validate the user name
		if user == "" {
			continue
		}
		cfg.UniqueUntrustedMachineUsers[user] = struct{}{}
	}
	if cfg.CheckName == "" {
		cfg.CheckName = "check-approval"
	}
	defaultTemplates := map[string]string{
		"footer":                string(templateFooter),
		"settings":              string(templateSettings),
		"two_approvals":         string(templateTwoApprovals),
		"no_approval":           string(templateNoApproval),
		"require_two_approvals": string(templateRequireTwoApprovals),
	}
	if cfg.Templates == nil {
		cfg.Templates = map[string]string{}
	}
	for name, tpl := range defaultTemplates {
		if _, ok := cfg.Templates[name]; !ok {
			cfg.Templates[name] = tpl
		}
	}
	var define string
	for k, v := range cfg.Templates {
		define += `{{define "` + k + `"}}` + v + "{{end}}"
	}
	keys := []string{
		"no_approval", "require_two_approvals", "two_approvals",
	}
	templates := make(map[string]*template.Template, len(keys))
	for _, k := range keys {
		tpl := cfg.Templates[k] + define
		tplParsed, err := template.New("_").Parse(tpl)
		if err != nil {
			return fmt.Errorf("parse the template %s: %w", k, err)
		}
		templates[k] = tplParsed
	}
	cfg.BuiltTemplates = templates
	for pattern := range cfg.UniqueUntrustedMachineUsers {
		if _, err := path.Match(pattern, "foo"); err != nil {
			return fmt.Errorf("invalid untrusted machine user pattern %q: %w", pattern, err)
		}
	}
	// TODO add test cases
	result := &Result{}
	for key, tpl := range cfg.BuiltTemplates {
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
}

type State string

const (
	// OK - Two approvals
	//   approvers
	StateTwoApprovals State = "two_approvals"
	// NG - approvals are required but actually no approval
	//   ignored approvers
	StateApprovalIsRequired State = "approval_is_required"
	// NG - two approvals are required but actually one approval
	//   why two approvals are required
	//     self approval
	//     untrusted author
	//     untrusted commit
	//   approvers
	//   self approvers
	//   ignored approvers
	StateTwoApprovalsAreRequired State = "two_approvals_are_required"
	// OK - one approval is sufficient
	//   approvers
	StateOneApproval State = "one_approval"
)
