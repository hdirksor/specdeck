package scaffold_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/hdickson/specdeck/internal/scaffold"
	"github.com/hdickson/specdeck/internal/spec"
)

func TestPropagateState_AddsStateToLeafContainers(t *testing.T) {
	dir := setupProject(t)
	writeStateFile(t, dir, "app.yml", `
- name: dark-mode
  summary: Dark mode
  facts: {}
`)

	// Create a leaf container with default specs filled in.
	heroPath := filepath.Join(dir, "containers", "hero.yml")
	if err := spec.WriteContainer(heroPath, spec.Container{
		States: []spec.StateSpec{
			{Ref: "default", Specs: map[string]spec.SpecValue{
				"background-color": {Value: "#FFFFFF"},
			}},
		},
	}); err != nil {
		t.Fatal(err)
	}

	if err := scaffold.PropagateState(dir, "dark-mode"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	c, err := spec.ParseContainer(heroPath)
	if err != nil {
		t.Fatal(err)
	}

	var darkMode *spec.StateSpec
	for i := range c.States {
		if c.States[i].Ref == "dark-mode" {
			darkMode = &c.States[i]
		}
	}
	if darkMode == nil {
		t.Fatal("expected dark-mode state to be added")
	}
	if v := darkMode.Specs["background-color"].Value; v != "#FFFFFF" {
		t.Errorf("expected default specs copied, got %v", v)
	}
}

func TestPropagateState_SkipsContainersThatAlreadyHaveState(t *testing.T) {
	dir := setupProject(t)
	writeStateFile(t, dir, "app.yml", `
- name: dark-mode
  summary: Dark mode
  facts: {}
`)

	heroPath := filepath.Join(dir, "containers", "hero.yml")
	if err := spec.WriteContainer(heroPath, spec.Container{
		States: []spec.StateSpec{
			{Ref: "default", Specs: map[string]spec.SpecValue{
				"background-color": {Value: "#FFFFFF"},
			}},
			{Ref: "dark-mode", Specs: map[string]spec.SpecValue{
				"background-color": {Value: "#000000"},
			}},
		},
	}); err != nil {
		t.Fatal(err)
	}

	if err := scaffold.PropagateState(dir, "dark-mode"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	c, err := spec.ParseContainer(heroPath)
	if err != nil {
		t.Fatal(err)
	}
	for _, s := range c.States {
		if s.Ref == "dark-mode" {
			if v := s.Specs["background-color"].Value; v != "#000000" {
				t.Errorf("expected existing dark-mode value preserved, got %v", v)
			}
		}
	}
}

func TestPropagateState_SkipsIndexYml(t *testing.T) {
	dir := setupProject(t)
	writeStateFile(t, dir, "app.yml", `
- name: dark-mode
  summary: Dark mode
  facts: {}
`)

	// Create a non-leaf index.yml — should not be touched.
	appDir := filepath.Join(dir, "containers", "app")
	if err := os.MkdirAll(appDir, 0755); err != nil {
		t.Fatal(err)
	}
	indexPath := filepath.Join(appDir, "index.yml")
	if err := spec.WriteContainer(indexPath, spec.Container{
		States: []spec.StateSpec{
			{Ref: "default", Specs: map[string]spec.SpecValue{
				"font-family": {Value: "Inter"},
			}},
		},
	}); err != nil {
		t.Fatal(err)
	}

	// Create a leaf container in a subdirectory.
	heroPath := filepath.Join(appDir, "hero.yml")
	if err := spec.WriteContainer(heroPath, spec.Container{
		States: []spec.StateSpec{
			{Ref: "default", Specs: map[string]spec.SpecValue{
				"background-color": {Value: "#FFFFFF"},
			}},
		},
	}); err != nil {
		t.Fatal(err)
	}

	if err := scaffold.PropagateState(dir, "dark-mode"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// index.yml should not have dark-mode added.
	index, err := spec.ParseContainer(indexPath)
	if err != nil {
		t.Fatal(err)
	}
	for _, s := range index.States {
		if s.Ref == "dark-mode" {
			t.Error("expected index.yml to be skipped, but dark-mode was added")
		}
	}

	// hero.yml should have dark-mode added.
	hero, err := spec.ParseContainer(heroPath)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, s := range hero.States {
		if s.Ref == "dark-mode" {
			found = true
		}
	}
	if !found {
		t.Error("expected dark-mode to be added to leaf container hero.yml")
	}
}

func TestPropagateState_ErrorsIfStateNotDefined(t *testing.T) {
	dir := setupProject(t)
	if err := scaffold.PropagateState(dir, "nonexistent"); err == nil {
		t.Fatal("expected error for undefined state")
	}
}
