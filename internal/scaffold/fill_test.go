package scaffold_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/hdickson/specdeck/internal/scaffold"
	"github.com/hdickson/specdeck/internal/spec"
)

func TestFill_CopiesDefaultSpecsIntoEmptyStates(t *testing.T) {
	dir := setupProject(t)
	writeStateFile(t, dir, "app.yml", `
- name: dark-mode
  summary: Dark mode
  facts: {}
`)
	if err := scaffold.AddContainer(dir, "hero"); err != nil {
		t.Fatal(err)
	}

	// Manually write default specs into the container.
	heroPath := filepath.Join(dir, "containers", "hero.yml")
	if err := spec.WriteContainer(heroPath, spec.Container{
		States: []spec.StateSpec{
			{Ref: "default", Specs: map[string]spec.SpecValue{
				"background-color": {Value: "#FFFFFF"},
				"title":            {Value: "Hello", Description: "Main heading"},
			}},
			{Ref: "dark-mode", Specs: map[string]spec.SpecValue{}},
		},
	}); err != nil {
		t.Fatal(err)
	}

	if err := scaffold.Fill(dir, "hero"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	c, err := spec.ParseContainer(heroPath)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	var darkMode *spec.StateSpec
	for i := range c.States {
		if c.States[i].Ref == "dark-mode" {
			darkMode = &c.States[i]
		}
	}
	if darkMode == nil {
		t.Fatal("dark-mode state not found")
	}
	if v := darkMode.Specs["background-color"].Value; v != "#FFFFFF" {
		t.Errorf("expected background-color '#FFFFFF', got %v", v)
	}
	if darkMode.Specs["title"].Description != "Main heading" {
		t.Errorf("expected description carried over, got %q", darkMode.Specs["title"].Description)
	}
}

func TestFill_CopiesDefaultEventsIntoEmptyStates(t *testing.T) {
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
			}, Events: map[string]interface{}{
				"on-tap": map[string]interface{}{"action": "navigate", "destination": "detail"},
			}},
			{Ref: "dark-mode", Specs: map[string]spec.SpecValue{}, Events: map[string]interface{}{}},
		},
	}); err != nil {
		t.Fatal(err)
	}

	if err := scaffold.Fill(dir, "hero"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	c, err := spec.ParseContainer(heroPath)
	if err != nil {
		t.Fatal(err)
	}
	for _, s := range c.States {
		if s.Ref == "dark-mode" {
			if len(s.Events) == 0 {
				t.Error("expected default events copied into dark-mode")
			}
		}
	}
}

func TestFill_DoesNotOverwriteExistingSpecs(t *testing.T) {
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

	if err := scaffold.Fill(dir, "hero"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	c, err := spec.ParseContainer(heroPath)
	if err != nil {
		t.Fatal(err)
	}
	for _, s := range c.States {
		if s.Ref == "dark-mode" {
			if v := s.Specs["background-color"].Value; v != "#000000" {
				t.Errorf("expected dark-mode value to be preserved '#000000', got %v", v)
			}
		}
	}
}

func TestFillAll_FillsAllLeafContainers(t *testing.T) {
	dir := setupProject(t)
	writeStateFile(t, dir, "app.yml", `
- name: dark-mode
  summary: Dark mode
  facts: {}
`)

	for _, path := range []string{"app/screen/hero", "app/screen/button"} {
		if err := scaffold.AddContainer(dir, path); err != nil {
			t.Fatal(err)
		}
	}

	// Fill in default specs for both containers.
	for _, name := range []string{"hero", "button"} {
		p := filepath.Join(dir, "containers", "app", "screen", name+".yml")
		if err := spec.WriteContainer(p, spec.Container{
			States: []spec.StateSpec{
				{Ref: "default", Specs: map[string]spec.SpecValue{
					"background-color": {Value: "#FFFFFF"},
				}},
				{Ref: "dark-mode", Specs: map[string]spec.SpecValue{}},
			},
		}); err != nil {
			t.Fatal(err)
		}
	}

	filled, skipped, err := scaffold.FillAll(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if filled != 2 {
		t.Errorf("expected 2 filled, got %d", filled)
	}
	if skipped != 0 {
		t.Errorf("expected 0 skipped, got %d", skipped)
	}

	// Verify both containers were filled.
	for _, name := range []string{"hero", "button"} {
		p := filepath.Join(dir, "containers", "app", "screen", name+".yml")
		c, err := spec.ParseContainer(p)
		if err != nil {
			t.Fatal(err)
		}
		for _, s := range c.States {
			if s.Ref == "dark-mode" && len(s.Specs) == 0 {
				t.Errorf("container %s dark-mode state was not filled", name)
			}
		}
	}
}

func TestFillAll_SkipsContainersWithNoDefaultState(t *testing.T) {
	dir := setupProject(t)
	// Override core.yml's default state with one that has a different name
	// so the container has no default to fill from.
	if err := os.WriteFile(filepath.Join(dir, "states", "core.yml"), []byte(`
- name: other
  summary: Other
  facts: {}
`), 0644); err != nil {
		t.Fatal(err)
	}

	heroPath := filepath.Join(dir, "containers", "hero.yml")
	if err := spec.WriteContainer(heroPath, spec.Container{
		States: []spec.StateSpec{
			{Ref: "dark-mode", Specs: map[string]spec.SpecValue{}},
		},
	}); err != nil {
		t.Fatal(err)
	}

	_, skipped, err := scaffold.FillAll(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if skipped != 1 {
		t.Errorf("expected 1 skipped, got %d", skipped)
	}
}

func TestFill_ErrorsIfNoDefaultState(t *testing.T) {
	dir := setupProject(t)
	heroPath := filepath.Join(dir, "containers", "hero.yml")
	if err := spec.WriteContainer(heroPath, spec.Container{
		States: []spec.StateSpec{
			{Ref: "dark-mode", Specs: map[string]spec.SpecValue{}},
		},
	}); err != nil {
		t.Fatal(err)
	}

	if err := scaffold.Fill(dir, "hero"); err == nil {
		t.Fatal("expected error when no default state found")
	}
}
