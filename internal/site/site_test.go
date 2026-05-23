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

func TestPrepare_CreatesContentStub(t *testing.T) {
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

	stubPath := filepath.Join(siteDir, "content", "containers", "button.md")
	data, err := os.ReadFile(stubPath)
	if err != nil {
		t.Fatalf("content stub not created: %v", err)
	}

	body := string(data)
	if !strings.Contains(body, "title: Button") {
		t.Error("stub does not contain title")
	}
	if !strings.HasPrefix(body, "---\n") {
		t.Error("stub does not start with YAML frontmatter delimiter")
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

	must := []string{
		"hugo.toml",
		filepath.Join("layouts", "_default", "baseof.html"),
		filepath.Join("layouts", "containers", "single.html"),
		filepath.Join("layouts", "index.html"),
		filepath.Join("static", "style.css"),
	}
	for _, rel := range must {
		if _, err := os.Stat(filepath.Join(siteDir, rel)); err != nil {
			t.Errorf("expected %s to exist: %v", rel, err)
		}
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
