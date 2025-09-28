package config

import (
	"html/template"
)

type Config struct {
	AppID          int64                         `json:"app_id" yaml:"app_id"`
	InstallationID int64                         `json:"installation_id" yaml:"installation_id"`
	AWS            *AWS                          `json:"aws,omitempty" yaml:"aws"`
	GoogleCloud    *GoogleCloud                  `json:"google_cloud,omitempty" yaml:"google_cloud"`
	CheckName      string                        `json:"check_name,omitempty" yaml:"check_name"`
	Trust          *Trust                        `json:"trust,omitempty" yaml:"trust"`
	Templates      map[string]string             `json:"templates,omitempty" yaml:"templates"`
	BuiltTemplates map[string]*template.Template `json:"-" yaml:"-"`
	LogLevel       string                        `json:"log_level,omitempty" yaml:"log_level"`
	Repositories   []*Repository                 `json:"repositories,omitempty" yaml:"repositories"`
}

func (c *Config) Init() error {
	if c.Trust == nil {
		c.Trust = &Trust{}
	}
	c.Trust.Init()
	if c.CheckName == "" {
		c.CheckName = "validate-review"
	}

	if err := c.validatePlatform(); err != nil {
		return err
	}

	if err := c.initRepos(); err != nil {
		return err
	}
	if err := c.initTemplates(); err != nil {
		return err
	}
	if err := c.testUntrustedMachineUsers(); err != nil {
		return err
	}
	return c.testTemplate()
}
