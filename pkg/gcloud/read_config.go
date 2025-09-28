// Package gcloud provides Google Cloud Platform integration for the validate-pr-review-app.
// It handles Cloud Functions execution, HTTP triggers, and Google Cloud Secret Manager integration.
package gcloud

import (
	"errors"
	"fmt"
	"os"

	"github.com/suzuki-shunsuke/validate-pr-review-app/pkg/config"
	"gopkg.in/yaml.v3"
)

// readConfig reads and parses the application configuration from the CONFIG environment variable.
// The configuration is expected to be a YAML-formatted string containing the application settings.
// It unmarshals the YAML into the provided config struct and initializes it.
func readConfig(cfg *config.Config) error {
	cfgstr := os.Getenv("CONFIG")
	if cfgstr == "" {
		return errors.New("CONFIG environment variable is required")
	}
	if err := yaml.Unmarshal([]byte(cfgstr), cfg); err != nil {
		return fmt.Errorf("failed to parse CONFIG environment variable: %w", err)
	}
	if err := cfg.Init(); err != nil {
		return fmt.Errorf("initialize config: %w", err)
	}
	return nil
}
