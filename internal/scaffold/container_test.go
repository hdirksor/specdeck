package scaffold_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/hdickson/specdeck/internal/scaffold"
)

func setupProject(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".git"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := scaffold.New(dir, "testproject"); err != nil {
		t.Fatal(err)
	}
	return dir
}

func writeStateFile(t *testing.T, dir, filename, content string) {
	t.Helper()
	path := filepath.Join(dir, "states", filename)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}
