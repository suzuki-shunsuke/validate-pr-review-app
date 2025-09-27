package config

import (
	_ "embed"
	"errors"
	"fmt"
	"html/template"
	"path"
	"strings"
)

type Config struct {
	AppID          int64                         `yaml:"app_id"`
	InstallationID int64                         `yaml:"installation_id"`
	AWS            *AWS                          `yaml:"aws"`
	CheckName      string                        `yaml:"check_name"`
	Trust          *Trust                        `yaml:"trust"`
	Templates      map[string]string             `yaml:"templates"`
	BuiltTemplates map[string]*template.Template `yaml:"-"`
	LogLevel       string                        `yaml:"log_level"`
	Repositories   []*Repository                 `yaml:"repositories"`
}

type Repository struct {
	Repositories []string `yaml:"repositories"`
	Trust        *Trust   `yaml:"trust"`
	Ignored      bool     `yaml:"ignored"`
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

func (c *Config) GetRepo(repo string) *Repository {
	for _, r := range c.Repositories {
		if r.Match(repo) {
			return r
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
		// TODO validate the app name
		if app == "" {
			continue
		}
		t.UniqueTrustedApps[app] = struct{}{}
	}
	t.UniqueTrustedMachineUsers = make(map[string]struct{}, len(t.TrustedMachineUsers))
	for _, user := range t.TrustedMachineUsers {
		// TODO validate the user name
		if user == "" {
			continue
		}
		t.UniqueTrustedMachineUsers[user] = struct{}{}
	}
	t.UniqueUntrustedMachineUsers = make(map[string]struct{}, len(t.UntrustedMachineUsers))
	for _, user := range t.UntrustedMachineUsers {
		// TODO validate the user name
		if user == "" {
			continue
		}
		t.UniqueUntrustedMachineUsers[user] = struct{}{}
	}
}

func (c *Config) Init() error {
	if c.Trust == nil {
		c.Trust = &Trust{}
	}
	c.Trust.Init()
	if c.CheckName == "" {
		c.CheckName = "validate-review"
	}
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
