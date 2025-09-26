package aws

import (
	"errors"
	"fmt"
	"os"

	"github.com/suzuki-shunsuke/validate-pr-review-app/pkg/config"
	"gopkg.in/yaml.v3"
)

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
