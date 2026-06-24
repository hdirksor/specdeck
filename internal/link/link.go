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

// Link installs Claude Code skills into dir. If specdeck.yml or specdeck.toml
// already exists the config file is left untouched; otherwise specdeck.yml is
// created with specsPath as the specs_repo value.
func Link(dir, specsPath string) error {
	if !configExists(dir) {
		if err := writeConfig(dir, specsPath); err != nil {
			return err
		}
	}
	return writeSkills(dir)
}

func configExists(dir string) bool {
	for _, name := range []string{"specdeck.yml", "specdeck.toml"} {
		if _, err := os.Stat(filepath.Join(dir, name)); err == nil {
			return true
		}
	}
	return false
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
	commandsDir := filepath.Join(dir, ".claude", "commands", "specdeck")
	if err := os.MkdirAll(commandsDir, 0755); err != nil {
		return fmt.Errorf("creating .claude/commands/specdeck: %w", err)
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
