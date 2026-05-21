package link

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type config struct {
	SpecsRepo string `yaml:"specs_repo"`
}

const specifySkill = `Read ` + "`specdeck.yml`" + ` in the project root to find the ` + "`specs_repo`" + ` path.

Look at $ARGUMENTS in the context of the current codebase — read relevant files, understand the feature's scope, boundaries, and how it fits with existing behaviour.

Then draft a spec document and write it to the specs repo under a kebab-case filename matching the feature (e.g. ` + "`user-login.md`" + `).

The spec should cover:
- What the feature does (behaviour, not implementation)
- Who uses it and when
- Key states and transitions
- Edge cases and constraints

Do not invent details that are not evident from the codebase or the feature description. Ask if anything is ambiguous.
`

// Link writes specdeck.yml and the Claude Code skills into dir.
// specsPath is the local path to the specs repo; it may be empty.
func Link(dir, specsPath string) error {
	if err := writeConfig(dir, specsPath); err != nil {
		return err
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
	commandsDir := filepath.Join(dir, ".claude", "commands")
	if err := os.MkdirAll(commandsDir, 0755); err != nil {
		return fmt.Errorf("creating .claude/commands: %w", err)
	}

	skillPath := filepath.Join(commandsDir, "specify.md")
	if err := os.WriteFile(skillPath, []byte(specifySkill), 0644); err != nil {
		return fmt.Errorf("writing specify skill: %w", err)
	}

	return nil
}
