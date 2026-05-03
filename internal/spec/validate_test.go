package spec_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/hdickson/specdeck/internal/spec"
)

func TestValidate_Valid(t *testing.T) {
	facts := []spec.Fact{
		{Name: "is-logged-in", Type: spec.FactTypeBoolean},
	}
	states := []spec.State{
		{Name: "default", Facts: map[string]interface{}{"is-logged-in": "false"}},
	}
	containers := []spec.Container{
		{Path: "app/card.yml", Default: "default", States: []spec.StateSpec{{Ref: "default"}}},
	}
	if errs := spec.Validate(facts, states, containers); len(errs) != 0 {
		t.Errorf("expected no errors, got %v", errs)
	}
}

func TestValidate_UnknownFactInState(t *testing.T) {
	facts := []spec.Fact{
		{Name: "is-logged-in", Type: spec.FactTypeBoolean},
	}
	states := []spec.State{
		{Name: "default", Facts: map[string]interface{}{"is-logged-in": "false", "unknown-fact": "x"}},
	}
	errs := spec.Validate(facts, states, nil)
	if len(errs) == 0 {
		t.Fatal("expected error for unknown fact in state")
	}
}

func TestValidate_UnknownStateRefInContainer(t *testing.T) {
	states := []spec.State{{Name: "default"}}
	containers := []spec.Container{
		{Path: "app/card.yml", Default: "default", States: []spec.StateSpec{
			{Ref: "default"},
			{Ref: "nonexistent"},
		}},
	}
	errs := spec.Validate(nil, states, containers)
	if len(errs) == 0 {
		t.Fatal("expected error for unknown state ref in container")
	}
	if errs[0].File != "app/card.yml" {
		t.Errorf("expected file 'app/card.yml', got %q", errs[0].File)
	}
}

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

func TestValidate_MultipleErrors(t *testing.T) {
	states := []spec.State{
		{Name: "default", Facts: map[string]interface{}{"bad-fact": "x"}},
	}
	containers := []spec.Container{
		{Path: "app/card.yml", Default: "default", States: []spec.StateSpec{{Ref: "bad-state"}}},
	}
	errs := spec.Validate(nil, states, containers)
	if len(errs) < 2 {
		t.Errorf("expected at least 2 errors, got %d: %v", len(errs), errs)
	}
}
