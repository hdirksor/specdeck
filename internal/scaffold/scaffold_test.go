package scaffold_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/hdickson/specdeck/internal/scaffold"
)

func TestNew_ErrorsIfNotGitRepo(t *testing.T) {
	dir := t.TempDir()
	err := scaffold.New(dir, "myproject")
	if err == nil {
		t.Fatal("expected error when directory is not a git repo")
	}
}

func TestNew_ErrorsIfDirectoryNotEmpty(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".git"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "somefile.txt"), []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}

	err := scaffold.New(dir, "myproject")
	if err == nil {
		t.Fatal("expected error when directory is not empty")
	}
}

func TestNew_CreatesExpectedStructure(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".git"), 0755); err != nil {
		t.Fatal(err)
	}

	if err := scaffold.New(dir, "myproject"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, entry := range []string{
		"specdeck.toml",
		"containers/index.yml",
		"changes/TEMPLATE.md",
		".claude/commands/change.md",
	} {
		if _, err := os.Stat(filepath.Join(dir, entry)); os.IsNotExist(err) {
			t.Errorf("expected %s to exist", entry)
		}
	}
}

func TestNew_DoesNotCreateExportsDir(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".git"), 0755); err != nil {
		t.Fatal(err)
	}

	if err := scaffold.New(dir, "myproject"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dir, "exports")); !os.IsNotExist(err) {
		t.Error("exports directory should not be created by new")
	}
}

func TestNew_ChangesTemplateHasExpectedSections(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".git"), 0755); err != nil {
		t.Fatal(err)
	}

	if err := scaffold.New(dir, "myproject"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "changes", "TEMPLATE.md"))
	if err != nil {
		t.Fatal(err)
	}

	content := string(data)
	for _, section := range []string{"## What changed", "## Why", "## References"} {
		if !contains(content, section) {
			t.Errorf("changes/TEMPLATE.md missing section %q", section)
		}
	}
}

func TestNew_TomlContainsProjectName(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".git"), 0755); err != nil {
		t.Fatal(err)
	}

	if err := scaffold.New(dir, "myproject"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "specdeck.toml"))
	if err != nil {
		t.Fatal(err)
	}

	content := string(data)
	if !contains(content, `name = "myproject"`) {
		t.Errorf("specdeck.toml missing expected name field, got:\n%s", content)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
