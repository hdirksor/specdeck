package site_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hdickson/specdeck/internal/site"
)

func writeFixture(t *testing.T, dir, rel, content string) {
	t.Helper()
	full := filepath.Join(dir, rel)
	if err := os.MkdirAll(filepath.Dir(full), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(full, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

func TestPrepare_WritesDataFile(t *testing.T) {
	distRoot := t.TempDir()
	siteDir := t.TempDir()

	writeFixture(t, distRoot, "button.yml", `title: Button
description: A clickable button
states:
  default:
    specs:
      height: 40px
`)

	if err := site.Prepare(distRoot, siteDir); err != nil {
		t.Fatalf("Prepare: %v", err)
	}

	dataPath := filepath.Join(siteDir, "data", "containers", "button.yml")
	data, err := os.ReadFile(dataPath)
	if err != nil {
		t.Fatalf("data file not created: %v", err)
	}
	if !strings.Contains(string(data), "height: 40px") {
		t.Error("data file missing spec value")
	}
}

func TestPrepare_CreatesContentStub(t *testing.T) {
	distRoot := t.TempDir()
	siteDir := t.TempDir()

	writeFixture(t, distRoot, "button.yml", `title: Button
description: A clickable button
`)

	if err := site.Prepare(distRoot, siteDir); err != nil {
		t.Fatalf("Prepare: %v", err)
	}

	stubPath := filepath.Join(siteDir, "content", "containers", "button.md")
	data, err := os.ReadFile(stubPath)
	if err != nil {
		t.Fatalf("content stub not created: %v", err)
	}

	body := string(data)
	if !strings.Contains(body, "title: Button") {
		t.Error("stub missing title")
	}
	if !strings.Contains(body, "data_path: button") {
		t.Error("stub missing data_path")
	}
	if strings.Contains(body, "description") {
		t.Error("stub should not contain description — data file handles that")
	}
}

func TestPrepare_NestedContainerStub(t *testing.T) {
	distRoot := t.TempDir()
	siteDir := t.TempDir()

	writeFixture(t, distRoot, "forms/input.yml", `title: Input
description: Text input field
`)

	if err := site.Prepare(distRoot, siteDir); err != nil {
		t.Fatalf("Prepare: %v", err)
	}

	stubPath := filepath.Join(siteDir, "content", "containers", "forms", "input.md")
	if _, err := os.Stat(stubPath); err != nil {
		t.Errorf("nested content stub not created: %v", err)
	}

	dataPath := filepath.Join(siteDir, "data", "containers", "forms", "input.yml")
	if _, err := os.Stat(dataPath); err != nil {
		t.Errorf("nested data file not created: %v", err)
	}

	data, err := os.ReadFile(stubPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "data_path: forms/input") {
		t.Errorf("nested stub has wrong data_path: %s", data)
	}
}

func TestPrepare_EmptyDistRoot(t *testing.T) {
	distRoot := t.TempDir()
	siteDir := t.TempDir()

	if err := site.Prepare(distRoot, siteDir); err != nil {
		t.Fatalf("Prepare with empty dist should succeed: %v", err)
	}
}

func TestBootstrap_CreatesScaffold(t *testing.T) {
	siteDir := t.TempDir()

	if err := site.Bootstrap(siteDir); err != nil {
		t.Fatalf("Bootstrap: %v", err)
	}

	if _, err := os.Stat(filepath.Join(siteDir, "hugo.toml")); err != nil {
		t.Errorf("expected hugo.toml to exist: %v", err)
	}
}

func TestBootstrap_SkipsIfHugoConfigExists(t *testing.T) {
	siteDir := t.TempDir()
	existing := "existing config"
	if err := os.WriteFile(filepath.Join(siteDir, "hugo.toml"), []byte(existing), 0644); err != nil {
		t.Fatal(err)
	}

	if err := site.Bootstrap(siteDir); err != nil {
		t.Fatalf("Bootstrap: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(siteDir, "hugo.toml"))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != existing {
		t.Error("Bootstrap overwrote existing hugo.toml")
	}
}
