package scaffold

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type config struct {
	Name string `toml:"name"`
}

// New initialises a specdeck project in dir with the given project name.
// dir must be a git repository and must contain no files other than .git.
func New(dir, projectName string) error {
	if err := requireGitRepo(dir); err != nil {
		return err
	}
	if err := requireEmpty(dir); err != nil {
		return err
	}
	if err := writeConfig(dir, projectName); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join(dir, "states", "facts"), 0755); err != nil {
		return fmt.Errorf("creating states/facts directory: %w", err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "containers"), 0755); err != nil {
		return fmt.Errorf("creating containers directory: %w", err)
	}
	if err := writeStarterFiles(dir); err != nil {
		return err
	}
	return nil
}

func writeStarterFiles(dir string) error {
	starterFact := `name: is-logged-in
type: boolean
`
	if err := os.WriteFile(filepath.Join(dir, "states", "facts", "is-logged-in.yml"), []byte(starterFact), 0644); err != nil {
		return fmt.Errorf("writing starter fact: %w", err)
	}

	starterStates := `- name: default
  summary: Default state
  facts:
    is-logged-in: false
`
	if err := os.WriteFile(filepath.Join(dir, "states", "core.yml"), []byte(starterStates), 0644); err != nil {
		return fmt.Errorf("writing starter states: %w", err)
	}

	if err := os.WriteFile(filepath.Join(dir, "containers", ".gitkeep"), []byte{}, 0644); err != nil {
		return fmt.Errorf("writing containers/.gitkeep: %w", err)
	}

	return nil
}

func requireGitRepo(dir string) error {
	if _, err := os.Stat(filepath.Join(dir, ".git")); os.IsNotExist(err) {
		return fmt.Errorf("%s is not a git repository — run `git init` first", dir)
	}
	return nil
}

func requireEmpty(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("reading directory: %w", err)
	}
	for _, e := range entries {
		if e.Name() != ".git" {
			return fmt.Errorf("directory is not empty — specdeck new must be run in an empty directory")
		}
	}
	return nil
}

func writeConfig(dir, projectName string) error {
	f, err := os.Create(filepath.Join(dir, "specdeck.toml"))
	if err != nil {
		return fmt.Errorf("creating specdeck.toml: %w", err)
	}
	defer f.Close()
	return toml.NewEncoder(f).Encode(config{Name: projectName})
}
