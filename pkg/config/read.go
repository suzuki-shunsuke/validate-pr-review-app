package config

import (
	"errors"
	"fmt"
	"os"

	"github.com/suzuki-shunsuke/slog-error/slogerr"
	"gopkg.in/yaml.v3"
)

func Read(cfg *Config) error {
	cfgBytes, err := readConfigBytes()
	if err != nil {
		return err
	}
	if err := yaml.Unmarshal(cfgBytes, cfg); err != nil {
		return fmt.Errorf("failed to parse CONFIG environment variable: %w", err)
	}
	if err := cfg.Init(); err != nil {
		return fmt.Errorf("initialize config: %w", err)
	}
	return nil
}

func readConfigBytes() ([]byte, error) {
	if cfgStr := os.Getenv("CONFIG"); cfgStr != "" {
		return []byte(cfgStr), nil
	}
	if cfgPath := os.Getenv("CONFIG_FILE"); cfgPath != "" {
		data, err := os.ReadFile(cfgPath) //nolint:gosec
		if err != nil {
			return nil, fmt.Errorf("read config file: %w", slogerr.With(err, "config_file", cfgPath))
		}
		return data, nil
	}
	return nil, errors.New("CONFIG or CONFIG_FILE environment variable is required")
}
