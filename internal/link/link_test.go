package link_test

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/hdirksor/specdeck/internal/config"
	"github.com/hdirksor/specdeck/internal/link"
)

func loadConfig(t *testing.T, dir string) config.Config {
	t.Helper()
	cfg, err := config.Load(dir)
	if err != nil {
		t.Fatalf("loading config: %v", err)
	}
	return cfg
}

func TestLink_CreatesSpecdeckYMLWithPath(t *testing.T) {
	dir := t.TempDir()

	if err := link.Link(dir, "/path/to/specs"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got := loadConfig(t, dir).SpecsRepo; got != "/path/to/specs" {
		t.Errorf("specs_repo = %q, want /path/to/specs", got)
	}
}

func TestLink_UpdatesSpecsRepoAndKeepsAgents(t *testing.T) {
	dir := t.TempDir()
	if err := config.Save(dir, config.Config{SpecsRepo: "/original/specs", Agents: []string{"claude"}}); err != nil {
		t.Fatal(err)
	}

	if err := link.Link(dir, "/new/specs"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := config.Config{SpecsRepo: "/new/specs", Agents: []string{"claude"}}
	if got := loadConfig(t, dir); !reflect.DeepEqual(got, want) {
		t.Errorf("got %+v, want %+v", got, want)
	}
}

func TestLink_DoesNotInstallAgentCommands(t *testing.T) {
	dir := t.TempDir()

	if err := link.Link(dir, "/path/to/specs"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dir, ".claude")); !os.IsNotExist(err) {
		t.Error("expected link not to create .claude")
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

	if got := loadConfig(t, dir).SpecsRepo; got != "/new/specs" {
		t.Errorf("specs_repo = %q, want /new/specs", got)
	}
}

func TestLink_ErrorsOnMalformedSpecdeckYML(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "specdeck.yml"), []byte("specs_repo: [unclosed"), 0644); err != nil {
		t.Fatal(err)
	}

	err := link.Link(dir, "/new/specs")
	if err == nil || !strings.Contains(err.Error(), "specdeck.yml") {
		t.Errorf("expected specdeck.yml error, got %v", err)
	}
}
