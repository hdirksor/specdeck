package link

import (
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

const specifySkill = `Before doing anything else: read ` + "`specdeck.yml`" + `. If the ` + "`skills_version`" + ` field is missing or less than 2, stop and tell the user: "Your specdeck skills are out of date — run ` + "`specdeck sync`" + ` to update them, then retry."

Read ` + "`specdeck.yml`" + ` in the project root to find the ` + "`specs_repo`" + ` path.

Look at $ARGUMENTS in the context of the current codebase — read relevant files, understand the feature's scope, boundaries, and how it fits with existing behaviour.

Then draft a spec document and write it to the specs repo under a kebab-case filename matching the feature (e.g. ` + "`user-login.md`" + `).

The spec should cover:
- What the feature does (behaviour, not implementation)
- Who uses it and when
- Key states and transitions
- Edge cases and constraints

Do not invent details that are not evident from the codebase or the feature description. Ask if anything is ambiguous.
`

const changeSkill = `Run ` + "`specdeck change new \"$ARGUMENTS\"`" + ` to create a new change record.

Then open the generated file and fill in the sections based on the current branch and any available context (commit messages, PR description, linked issues):
- **What changed**: summarise the spec changes this work introduces
- **Why**: extract motivation from commit messages, PR description, or ask the user
- **References**: include any linked tickets, PRs, or Slack threads

Print the path to the created file when done.
`

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

	skills := map[string]string{
		"specify.md": specifySkill,
		"change.md":  changeSkill,
	}

	for filename, content := range skills {
		path := filepath.Join(commandsDir, filename)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			return fmt.Errorf("writing %s: %w", filename, err)
		}
	}

	return nil
}
