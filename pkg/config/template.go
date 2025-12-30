package config

import (
	_ "embed"
	"fmt"
	"html/template"
	"io"
	"strings"

	"github.com/suzuki-shunsuke/validate-pr-review-app/pkg/validation"
)

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
	var defineBuilder strings.Builder
	for k, v := range c.Templates {
		defineBuilder.WriteString(`{{define "`)
		defineBuilder.WriteString(k)
		defineBuilder.WriteString(`"}}`)
		defineBuilder.WriteString(v)
		defineBuilder.WriteString("{{end}}")
	}
	define := defineBuilder.String()
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

func (c *Config) testTemplate() error {
	// TODO add test cases
	result := &validation.Result{}
	for key, tpl := range c.BuiltTemplates {
		if err := tpl.Execute(io.Discard, result); err != nil {
			return fmt.Errorf("test template %s: %w", key, err)
		}
	}
	return nil
}
