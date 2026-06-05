package link

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const skillsVersion = 2

type config struct {
	SpecsRepo     string `yaml:"specs_repo"`
	SkillsVersion int    `yaml:"skills_version"`
}

//go:embed skills
var skillsFS embed.FS

// Link writes specdeck.yml and the Claude Code skills into dir.
// specsPath is the local path to the specs repo; it may be empty.
func Link(dir, specsPath string) error {
	if err := writeConfig(dir, specsPath); err != nil {
		return err
	}
	return writeSkills(dir)
}

// Sync re-writes all Claude Code skill files to the current version without
// changing the specs_repo path in specdeck.yml. Returns an error if
// specdeck.yml does not exist.
func Sync(dir string) error {
	cfg, err := readConfig(dir)
	if err != nil {
		return fmt.Errorf("reading specdeck.yml: %w — run `specdeck link` first", err)
	}
	if err := writeConfig(dir, cfg.SpecsRepo); err != nil {
		return err
	}
	return writeSkills(dir)
}

func readConfig(dir string) (config, error) {
	data, err := os.ReadFile(filepath.Join(dir, "specdeck.yml"))
	if err != nil {
		return config{}, err
	}
	var cfg config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return config{}, err
	}
	return cfg, nil
}

func writeConfig(dir, specsPath string) error {
	f, err := os.Create(filepath.Join(dir, "specdeck.yml"))
	if err != nil {
		return fmt.Errorf("creating specdeck.yml: %w", err)
	}
	defer f.Close()

	enc := yaml.NewEncoder(f)
	enc.SetIndent(2)
	return enc.Encode(config{SpecsRepo: specsPath, SkillsVersion: skillsVersion})
}

func writeSkills(dir string) error {
	commandsDir := filepath.Join(dir, ".claude", "commands")
	if err := os.MkdirAll(commandsDir, 0755); err != nil {
		return fmt.Errorf("creating .claude/commands: %w", err)
	}

	entries, err := skillsFS.ReadDir("skills")
	if err != nil {
		return fmt.Errorf("reading embedded skills: %w", err)
	}

	for _, entry := range entries {
		content, err := skillsFS.ReadFile("skills/" + entry.Name())
		if err != nil {
			return fmt.Errorf("reading embedded skill %s: %w", entry.Name(), err)
		}
		path := filepath.Join(commandsDir, entry.Name())
		if err := os.WriteFile(path, content, 0644); err != nil {
			return fmt.Errorf("writing %s: %w", entry.Name(), err)
		}
	}

	return nil
}
