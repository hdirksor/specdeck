package spec_test

import (
	"path/filepath"
	"testing"

	"github.com/hdickson/specdeck/internal/spec"
)

func TestParseStates_Basic(t *testing.T) {
	f := writeTempFile(t, `
- name: default
  summary: Standard logged-in English user
  facts:
    language: en
    is-logged-in: true

- name: spanish-user
  summary: Spanish-speaking user
  facts:
    language: es
    is-logged-in: true
`)
	states, err := spec.ParseStates(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(states) != 2 {
		t.Fatalf("expected 2 states, got %d", len(states))
	}
	if states[0].Name != "default" {
		t.Errorf("expected first state name 'default', got %q", states[0].Name)
	}
	if states[1].Facts["language"] != "es" {
		t.Errorf("expected language 'es', got %v", states[1].Facts["language"])
	}
}

func TestParseStates_ErrorOnDuplicateNames(t *testing.T) {
	f := writeTempFile(t, `
- name: default
  summary: First
  facts: {}

- name: default
  summary: Duplicate
  facts: {}
`)
	_, err := spec.ParseStates(f)
	if err == nil {
		t.Fatal("expected error for duplicate state names")
	}
}

func TestLoadStates(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "checkout.yml"), `
- name: default
  summary: Standard checkout
  facts:
    is-logged-in: true
`)
	writeFile(t, filepath.Join(dir, "onboarding.yml"), `
- name: new-user
  summary: First time user
  facts:
    is-logged-in: false
`)

	states, err := spec.LoadStates(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(states) != 2 {
		t.Errorf("expected 2 states, got %d", len(states))
	}
}

func TestLoadStates_ErrorOnDuplicateAcrossFiles(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "a.yml"), `
- name: default
  summary: In file a
  facts: {}
`)
	writeFile(t, filepath.Join(dir, "b.yml"), `
- name: default
  summary: Also in file b
  facts: {}
`)
	_, err := spec.LoadStates(dir)
	if err == nil {
		t.Fatal("expected error for duplicate state names across files")
	}
}
