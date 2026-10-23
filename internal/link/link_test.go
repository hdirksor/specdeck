package link_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hdirksor/specdeck/internal/link"
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

	if _, err := os.Stat(filepath.Join(dir, ".claude", "commands")); os.IsNotExist(err) {
		t.Error("expected .claude/commands directory to exist")
	}
}

func TestLink_CreatesSpecifySkill(t *testing.T) {
	dir := t.TempDir()

	if err := link.Link(dir, ""); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dir, ".claude", "commands", "specify.md")); os.IsNotExist(err) {
		t.Error("expected .claude/commands/specify.md to exist")
	}
}

func TestLink_SpecifySkillReferencesSpecsRepo(t *testing.T) {
	dir := t.TempDir()

	if err := link.Link(dir, ""); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, ".claude", "commands", "specify.md"))
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(data), "specdeck.yml") {
		t.Errorf("specify.md should reference specdeck.yml, got:\n%s", string(data))
	}
}

func TestLink_DoesNotWriteSkillsVersion(t *testing.T) {
	dir := t.TempDir()

	if err := link.Link(dir, "/some/path"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "specdeck.yml"))
	if err != nil {
		t.Fatal(err)
	}

	if strings.Contains(string(data), "skills_version") {
		t.Errorf("specdeck.yml should not contain skills_version, got:\n%s", string(data))
	}
}

func TestLink_CreatesChangeSkill(t *testing.T) {
	dir := t.TempDir()

	if err := link.Link(dir, ""); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dir, ".claude", "commands", "change.md")); os.IsNotExist(err) {
		t.Error("expected .claude/commands/change.md to exist")
	}
}

func TestLink_SpecifySkillHasNoVersionCheck(t *testing.T) {
	dir := t.TempDir()

	if err := link.Link(dir, ""); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, ".claude", "commands", "specify.md"))
	if err != nil {
		t.Fatal(err)
	}

	if strings.Contains(string(data), "skills_version") {
		t.Errorf("specify.md should not contain a skills_version check, got:\n%s", string(data))
	}
}

func TestLink_PreservesExistingSpecdeckYML(t *testing.T) {
	dir := t.TempDir()
	existing := "specs_repo: /original/specs\n"
	if err := os.WriteFile(filepath.Join(dir, "specdeck.yml"), []byte(existing), 0644); err != nil {
		t.Fatal(err)
	}

	if err := link.Link(dir, "/new/specs"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "specdeck.yml"))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != existing {
		t.Errorf("specdeck.yml should be unchanged, got:\n%s", string(data))
	}
}

func TestLink_IgnoresSpecdeckTOML(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "specdeck.toml"), []byte("specs_repo = \"/original/specs\"\n"), 0644); err != nil {
		t.Fatal(err)
	}

	if err := link.Link(dir, "/new/specs"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "specdeck.yml"))
	if err != nil {
		t.Fatalf("expected specdeck.yml to be created even when specdeck.toml exists: %v", err)
	}
	if !strings.Contains(string(data), "/new/specs") {
		t.Errorf("specdeck.yml missing specs_repo path, got:\n%s", string(data))
	}
}

func TestLink_RewritesSkillsWhenConfigExists(t *testing.T) {
	dir := t.TempDir()
	if err := link.Link(dir, "/original/specs"); err != nil {
		t.Fatalf("setup error: %v", err)
	}

	// Remove skills to simulate stale state.
	os.Remove(filepath.Join(dir, ".claude", "commands", "specify.md"))
	os.Remove(filepath.Join(dir, ".claude", "commands", "change.md"))

	if err := link.Link(dir, ""); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, skill := range []string{"specify.md", "change.md"} {
		if _, err := os.Stat(filepath.Join(dir, ".claude", "commands", skill)); os.IsNotExist(err) {
			t.Errorf("expected .claude/commands/%s to exist after re-link", skill)
		}
	}
}

func TestLink_OverwritesStaleSkills(t *testing.T) {
	dir := t.TempDir()
	if err := link.Link(dir, "/original/specs"); err != nil {
		t.Fatalf("setup error: %v", err)
	}

	specifyPath := filepath.Join(dir, ".claude", "commands", "specify.md")
	if err := os.WriteFile(specifyPath, []byte("stale"), 0644); err != nil {
		t.Fatal(err)
	}

	if err := link.Link(dir, ""); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(specifyPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) == "stale" {
		t.Error("expected specify.md to be overwritten on re-link")
	}
}
