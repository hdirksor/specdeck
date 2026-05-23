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
		t.Error("expected .claude/commands/specdeck directory to exist")
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


func TestLink_CreatesChangeSkill(t *testing.T) {
	dir := t.TempDir()

	if err := link.Link(dir, ""); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dir, ".claude", "commands", "specdeck", "change.md")); os.IsNotExist(err) {
		t.Error("expected .claude/commands/change.md to exist")
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


func TestSync_WritesSkills(t *testing.T) {
	dir := t.TempDir()

	if err := link.Link(dir, ""); err != nil {
		t.Fatalf("setup error: %v", err)
	}

	// Remove skills to simulate stale state.
	os.Remove(filepath.Join(dir, ".claude", "commands", "specdeck", "specify.md"))
	os.Remove(filepath.Join(dir, ".claude", "commands", "specdeck", "change.md"))

	if err := link.Sync(dir); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, skill := range []string{"specify.md", "change.md"} {
		if _, err := os.Stat(filepath.Join(dir, ".claude", "commands", "specdeck", skill)); os.IsNotExist(err) {
			t.Errorf("expected .claude/commands/specdeck/%s to exist after sync", skill)
		}
	}
}

func TestSync_ErrorsWithoutSpecdeckYML(t *testing.T) {
	dir := t.TempDir()

	if err := link.Sync(dir); err == nil {
		t.Error("expected error when specdeck.yml is missing")
	}
}
