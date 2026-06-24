package link_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hdickson/specdeck/internal/link"
)

func TestLink_CreatesSpecdeckYML_WithPath(t *testing.T) {
	dir := t.TempDir()

	if err := link.Link(dir, "/path/to/specs"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "specdeck.yml"))
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(data), "/path/to/specs") {
		t.Errorf("specdeck.yml missing specs_repo path, got:\n%s", string(data))
	}
}

func TestLink_CreatesSpecdeckYML_WithEmptyPath(t *testing.T) {
	dir := t.TempDir()

	if err := link.Link(dir, ""); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dir, "specdeck.yml")); os.IsNotExist(err) {
		t.Error("expected specdeck.yml to exist")
	}
}

func TestLink_CreatesClaudeCommandsDir(t *testing.T) {
	dir := t.TempDir()

	if err := link.Link(dir, ""); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dir, ".claude", "commands", "specdeck")); os.IsNotExist(err) {
		t.Error("expected .claude/commands directory to exist")
	}
}

func TestLink_CreatesSpecifySkill(t *testing.T) {
	dir := t.TempDir()

	if err := link.Link(dir, ""); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dir, ".claude", "commands", "specdeck", "specify.md")); os.IsNotExist(err) {
		t.Error("expected .claude/commands/specify.md to exist")
	}
}

func TestLink_SpecifySkillReferencesSpecsRepo(t *testing.T) {
	dir := t.TempDir()

	if err := link.Link(dir, ""); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, ".claude", "commands", "specdeck", "specify.md"))
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(data), "specdeck.yml") {
		t.Errorf("specify.md should reference specdeck.yml, got:\n%s", string(data))
	}
}

func TestLink_WritesSkillsVersion(t *testing.T) {
	dir := t.TempDir()

	if err := link.Link(dir, "/some/path"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "specdeck.yml"))
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(data), "skills_version:") {
		t.Errorf("specdeck.yml missing skills_version, got:\n%s", string(data))
	}
}


func TestLink_SpecifySkillContainsVersionCheck(t *testing.T) {
	dir := t.TempDir()

	if err := link.Link(dir, ""); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, ".claude", "commands", "specdeck", "specify.md"))
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(data), "skills_version") {
		t.Errorf("specify.md should contain a skills_version check, got:\n%s", string(data))
	}
	if !strings.Contains(string(data), "specdeck link") {
		t.Errorf("specify.md should reference `specdeck link`, got:\n%s", string(data))
	}
}

func TestLink_PreservesExistingConfig(t *testing.T) {
	dir := t.TempDir()

	if err := link.Link(dir, "/original/specs"); err != nil {
		t.Fatalf("setup error: %v", err)
	}

	if err := link.Link(dir, "/different/specs"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "specdeck.yml"))
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(data), "/original/specs") {
		t.Errorf("Link should not overwrite existing config, got:\n%s", string(data))
	}
}

func TestLink_PreservesExistingToml(t *testing.T) {
	dir := t.TempDir()

	if err := os.WriteFile(filepath.Join(dir, "specdeck.toml"), []byte(`specs_repo = "/toml/specs"`), 0644); err != nil {
		t.Fatal(err)
	}

	if err := link.Link(dir, "/other/specs"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dir, "specdeck.yml")); err == nil {
		t.Error("Link should not create specdeck.yml when specdeck.toml exists")
	}
}

func TestLink_UpdatesSkillsOnRerun(t *testing.T) {
	dir := t.TempDir()

	if err := link.Link(dir, ""); err != nil {
		t.Fatalf("setup error: %v", err)
	}

	// Remove skills to simulate stale state.
	os.Remove(filepath.Join(dir, ".claude", "commands", "specdeck", "specify.md"))
	if err := link.Link(dir, ""); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, skill := range []string{"specify.md"} {
		if _, err := os.Stat(filepath.Join(dir, ".claude", "commands", "specdeck", skill)); os.IsNotExist(err) {
			t.Errorf("expected .claude/commands/specdeck/%s to exist after re-link", skill)
		}
	}
}
