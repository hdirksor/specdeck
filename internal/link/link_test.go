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

func TestLink_CreatesChangeSkill(t *testing.T) {
	dir := t.TempDir()

	if err := link.Link(dir, ""); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dir, ".claude", "commands", "change.md")); os.IsNotExist(err) {
		t.Error("expected .claude/commands/change.md to exist")
	}
}

func TestLink_SpecifySkillContainsVersionCheck(t *testing.T) {
	dir := t.TempDir()

	if err := link.Link(dir, ""); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, ".claude", "commands", "specify.md"))
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(data), "skills_version") {
		t.Errorf("specify.md should contain a skills_version check, got:\n%s", string(data))
	}
	if !strings.Contains(string(data), "specdeck sync") {
		t.Errorf("specify.md should reference `specdeck sync`, got:\n%s", string(data))
	}
}

func TestSync_PreservesSpecsRepo(t *testing.T) {
	dir := t.TempDir()

	if err := link.Link(dir, "/original/specs"); err != nil {
		t.Fatalf("setup error: %v", err)
	}

	if err := link.Sync(dir); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "specdeck.yml"))
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(data), "/original/specs") {
		t.Errorf("Sync should preserve specs_repo, got:\n%s", string(data))
	}
}

func TestSync_UpdatesSkillsVersion(t *testing.T) {
	dir := t.TempDir()

	if err := link.Link(dir, ""); err != nil {
		t.Fatalf("setup error: %v", err)
	}

	if err := link.Sync(dir); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "specdeck.yml"))
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(data), "skills_version:") {
		t.Errorf("specdeck.yml missing skills_version after sync, got:\n%s", string(data))
	}
}

func TestSync_WritesSkills(t *testing.T) {
	dir := t.TempDir()

	if err := link.Link(dir, ""); err != nil {
		t.Fatalf("setup error: %v", err)
	}

	// Remove skills to simulate stale state.
	os.Remove(filepath.Join(dir, ".claude", "commands", "specify.md"))
	os.Remove(filepath.Join(dir, ".claude", "commands", "change.md"))

	if err := link.Sync(dir); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, skill := range []string{"specify.md", "change.md"} {
		if _, err := os.Stat(filepath.Join(dir, ".claude", "commands", skill)); os.IsNotExist(err) {
			t.Errorf("expected .claude/commands/%s to exist after sync", skill)
		}
	}
}

func TestSync_ErrorsWithoutSpecdeckYML(t *testing.T) {
	dir := t.TempDir()

	if err := link.Sync(dir); err == nil {
		t.Error("expected error when specdeck.yml is missing")
	}
}
