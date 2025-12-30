package config

import (
	"errors"
	"fmt"
)

type AWS struct {
	SecretID             string `json:"secret_id" yaml:"secret_id"`
	UseLambdaFunctionURL bool   `json:"use_lambda_function_url,omitempty" yaml:"use_lambda_function_url"`
}

func (a *AWS) Validate() error {
	if a.SecretID == "" {
		return errors.New("secret_id is required")
	}
	return nil
}

type GoogleCloud struct {
	SecretName string `json:"secret_name" yaml:"secret_name"`
}

func (g *GoogleCloud) Validate() error {
	if g.SecretName == "" {
		return errors.New("secret_name is required")
	}
	return nil
}

func (c *Config) validatePlatform() error {
	if c.AWS != nil {
		if err := c.AWS.Validate(); err != nil {
			return fmt.Errorf("validate aws config: %w", err)
		}
	}
	if c.GoogleCloud != nil {
		if err := c.GoogleCloud.Validate(); err != nil {
			return fmt.Errorf("validate google_cloud config: %w", err)
		}
	}
	if c.AWS != nil && c.GoogleCloud != nil {
		return errors.New("only one of aws or google_cloud configuration can be set")
	}
	return nil
}
