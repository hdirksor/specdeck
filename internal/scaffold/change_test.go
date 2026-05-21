package scaffold_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/hdickson/specdeck/internal/scaffold"
)

func TestNewChange_CreatesFileInChangesDir(t *testing.T) {
	dir := setupProject(t)

	path, err := scaffold.NewChange(dir, "rename the submit button")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dir, path)); os.IsNotExist(err) {
		t.Errorf("expected file %s to exist", path)
	}
}

func TestNewChange_FileNameIsDatePrefixedSlug(t *testing.T) {
	dir := setupProject(t)

	path, err := scaffold.NewChange(dir, "Rename the Submit Button!")
	if err != nil {
		t.Fatal(err)
	}

	today := time.Now().Format("2006-01-02")
	expected := filepath.Join("changes", today+"-rename-the-submit-button.md")
	if path != expected {
		t.Errorf("expected path %q, got %q", expected, path)
	}
}

func TestNewChange_FileContainsTitleAndTemplateSections(t *testing.T) {
	dir := setupProject(t)

	path, err := scaffold.NewChange(dir, "add dark mode")
	if err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(filepath.Join(dir, path))
	if err != nil {
		t.Fatal(err)
	}

	content := string(data)
	if !strings.Contains(content, "add dark mode") {
		t.Error("file should contain the change title")
	}
	for _, section := range []string{"## What changed", "## Why", "## References"} {
		if !strings.Contains(content, section) {
			t.Errorf("file missing section %q", section)
		}
	}
}

func TestNewChange_ErrorsIfChangesDirectoryMissing(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".git"), 0755); err != nil {
		t.Fatal(err)
	}

	_, err := scaffold.NewChange(dir, "some change")
	if err == nil {
		t.Fatal("expected error when changes/ directory does not exist")
	}
}
