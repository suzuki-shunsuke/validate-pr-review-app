package config

import (
	_ "embed"
	"fmt"
	"text/template"
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
		cfg.UniqueTrustedApps[app] = struct{}{}
	}
	cfg.UniqueTrustedMachineUsers = make(map[string]struct{}, len(cfg.TrustedMachineUsers))
	for _, user := range cfg.TrustedMachineUsers {
		cfg.UniqueTrustedMachineUsers[user] = struct{}{}
	}
	cfg.UniqueUntrustedMachineUsers = make(map[string]struct{}, len(cfg.UntrustedMachineUsers))
	for _, user := range cfg.UntrustedMachineUsers {
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
		tplParsed, err := template.New(k).Parse(tpl)
		if err != nil {
			return fmt.Errorf("parse the template: %w", err)
		}
		templates[k] = tplParsed
	}
	cfg.BuiltTemplates = templates
	// TODO test templates
	return nil
}
