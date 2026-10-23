package link

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type config struct {
	SpecsRepo string `yaml:"specs_repo"`
}

//go:embed skills
var skillsFS embed.FS

// Link writes specdeck.yml and the Claude Code skills into dir.
// specsPath is the local path to the specs repo; it may be empty.
// An existing specdeck.yml is left untouched so re-running link only
// refreshes the skills.
func Link(dir, specsPath string) error {
	_, err := os.Stat(filepath.Join(dir, "specdeck.yml"))
	if os.IsNotExist(err) {
		if err := writeConfig(dir, specsPath); err != nil {
			return err
		}
	} else if err != nil {
		return fmt.Errorf("checking specdeck.yml: %w", err)
	}
	return writeSkills(dir)
}

func writeConfig(dir, specsPath string) error {
	f, err := os.Create(filepath.Join(dir, "specdeck.yml"))
	if err != nil {
		return fmt.Errorf("creating specdeck.yml: %w", err)
	}
	defer f.Close()

	enc := yaml.NewEncoder(f)
	enc.SetIndent(2)
	return enc.Encode(config{SpecsRepo: specsPath})
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
