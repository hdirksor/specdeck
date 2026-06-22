package spec_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/hdickson/specdeck/internal/spec"
)

func TestValidateImports_Valid(t *testing.T) {
	dir := t.TempDir()
	jotDir := filepath.Join(dir, "jot")
	if err := os.MkdirAll(jotDir, 0755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(jotDir, "noteInput.yml"), `specs:
  label: Notes
`)
	containers := []spec.Container{
		{Path: "jot/index.yml", Imports: []spec.Import{{Ref: "./noteInput.yml"}}},
	}
	errs := spec.ValidateImports(dir, containers)
	if len(errs) != 0 {
		t.Errorf("expected no errors, got %v", errs)
	}
}

func TestValidateImports_MissingFile(t *testing.T) {
	dir := t.TempDir()
	containers := []spec.Container{
		{Path: "jot/index.yml", Imports: []spec.Import{{Ref: "./missing.yml"}}},
	}
	errs := spec.ValidateImports(dir, containers)
	if len(errs) == 0 {
		t.Fatal("expected error for missing import")
	}
	if errs[0].File != "jot/index.yml" {
		t.Errorf("expected file 'jot/index.yml', got %q", errs[0].File)
	}
}

func TestValidateImports_RelativeTraversal(t *testing.T) {
	dir := t.TempDir()
	sharedDir := filepath.Join(dir, "shared")
	if err := os.MkdirAll(sharedDir, 0755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(sharedDir, "hero.yml"), `specs:
  headerText: Hero
`)
	containers := []spec.Container{
		{Path: "jot/index.yml", Imports: []spec.Import{{Ref: "../shared/hero.yml"}}},
	}
	errs := spec.ValidateImports(dir, containers)
	if len(errs) != 0 {
		t.Errorf("expected no errors, got %v", errs)
	}
}
