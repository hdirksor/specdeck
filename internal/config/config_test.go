package config_test

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/hdirksor/specdeck/internal/config"
)

func TestLoad_MissingFileIsEmpty(t *testing.T) {
	cfg, err := config.Load(t.TempDir())
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if !reflect.DeepEqual(cfg, config.Config{}) {
		t.Errorf("got %+v, want empty config", cfg)
	}
}

func TestSaveAndLoad_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	want := config.Config{SpecsRepo: "../specs", Agents: []string{"claude"}}

	if err := config.Save(dir, want); err != nil {
		t.Fatalf("Save: %v", err)
	}
	got, err := config.Load(dir)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %+v, want %+v", got, want)
	}
}

func TestSave_WritesReadableYAML(t *testing.T) {
	dir := t.TempDir()

	if err := config.Save(dir, config.Config{SpecsRepo: "../specs", Agents: []string{"claude"}}); err != nil {
		t.Fatalf("Save: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "specdeck.yml"))
	if err != nil {
		t.Fatal(err)
	}
	want := "specs_repo: ../specs\nagents:\n  - claude\n"
	if string(data) != want {
		t.Errorf("got:\n%s\nwant:\n%s", data, want)
	}
}

func TestSave_OmitsEmptyAgents(t *testing.T) {
	dir := t.TempDir()

	if err := config.Save(dir, config.Config{SpecsRepo: "../specs"}); err != nil {
		t.Fatalf("Save: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "specdeck.yml"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), "agents") {
		t.Errorf("expected no agents key, got:\n%s", data)
	}
}

func TestLoad_IgnoresUnknownFields(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "specdeck.yml"), []byte("specs_repo: ../specs\nskills_version: 2\n"), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := config.Load(dir)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.SpecsRepo != "../specs" {
		t.Errorf("specs_repo = %q, want ../specs", cfg.SpecsRepo)
	}
}

func TestLoad_ErrorsOnMalformedFile(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "specdeck.yml"), []byte("specs_repo: [unclosed"), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := config.Load(dir)
	if err == nil || !strings.Contains(err.Error(), "specdeck.yml") {
		t.Errorf("expected error naming specdeck.yml, got %v", err)
	}
}

func TestLoad_ErrorsWhenUnreadable(t *testing.T) {
	// A directory named specdeck.yml exists but cannot be read as a file.
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, "specdeck.yml"), 0755); err != nil {
		t.Fatal(err)
	}

	_, err := config.Load(dir)
	if err == nil || !strings.Contains(err.Error(), "reading specdeck.yml") {
		t.Errorf("expected read error, got %v", err)
	}
}

func TestSave_ErrorsWhenUnwritable(t *testing.T) {
	file := filepath.Join(t.TempDir(), "not-a-dir")
	if err := os.WriteFile(file, nil, 0644); err != nil {
		t.Fatal(err)
	}

	err := config.Save(file, config.Config{})
	if err == nil || !strings.Contains(err.Error(), "writing specdeck.yml") {
		t.Errorf("expected write error, got %v", err)
	}
}
