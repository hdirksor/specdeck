// Package config reads and writes a code repo's specdeck.yml.
package config

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// FileName is the code repo's config file.
const FileName = "specdeck.yml"

// Config is the content of a code repo's specdeck.yml.
type Config struct {
	SpecsRepo string   `yaml:"specs_repo"`
	Agents    []string `yaml:"agents,omitempty"`
}

// Load reads specdeck.yml from dir. A missing file is an empty config.
func Load(dir string) (Config, error) {
	data, err := os.ReadFile(filepath.Join(dir, FileName))
	if os.IsNotExist(err) {
		return Config{}, nil
	}
	if err != nil {
		return Config{}, fmt.Errorf("reading %s: %w", FileName, err)
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("parsing %s: %w", FileName, err)
	}
	return cfg, nil
}

// Save writes cfg to specdeck.yml in dir.
func Save(dir string, cfg Config) error {
	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	if err := enc.Encode(cfg); err != nil {
		return fmt.Errorf("encoding %s: %w", FileName, err)
	}
	if err := os.WriteFile(filepath.Join(dir, FileName), buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("writing %s: %w", FileName, err)
	}
	return nil
}
