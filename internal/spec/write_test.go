package spec_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/hdickson/specdeck/internal/spec"
)

func TestWriteContainer_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.yml")

	original := spec.Container{
		States: []spec.StateSpec{
			{
				Ref: "default",
				Specs: map[string]spec.SpecValue{
					"background-color": {Value: "#FFFFFF"},
					"title": {Value: "Hello", Description: "Main heading"},
				},
			},
			{
				Ref:   "dark-mode",
				Specs: map[string]spec.SpecValue{},
			},
		},
	}

	if err := spec.WriteContainer(path, original); err != nil {
		t.Fatalf("unexpected error writing: %v", err)
	}

	parsed, err := spec.ParseContainer(path)
	if err != nil {
		t.Fatalf("unexpected error parsing: %v", err)
	}

	if len(parsed.States) != 2 {
		t.Fatalf("expected 2 states, got %d", len(parsed.States))
	}
	if parsed.States[0].Specs["background-color"].Value != "#FFFFFF" {
		t.Errorf("expected background-color '#FFFFFF', got %v", parsed.States[0].Specs["background-color"].Value)
	}
	if parsed.States[0].Specs["title"].Description != "Main heading" {
		t.Errorf("expected description 'Main heading', got %q", parsed.States[0].Specs["title"].Description)
	}
}

func TestWriteContainer_NamedDefault(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.yml")

	c := spec.Container{
		Default: "baseline",
		States: []spec.StateSpec{
			{Ref: "baseline", Specs: map[string]spec.SpecValue{
				"color": {Value: "red"},
			}},
		},
	}

	if err := spec.WriteContainer(path, c); err != nil {
		t.Fatalf("unexpected write error: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !containsStr(string(data), "default: baseline") {
		t.Errorf("expected 'default: baseline' in output, got:\n%s", data)
	}
}

func TestWriteContainer_EventsRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.yml")

	c := spec.Container{
		States: []spec.StateSpec{
			{
				Ref:   "default",
				Specs: map[string]spec.SpecValue{"background-color": {Value: "#FFFFFF"}},
				Events: map[string]interface{}{
					"on-tap":        map[string]interface{}{"action": "navigate", "destination": "detail"},
					"on-long-press": "show-menu",
				},
			},
			{
				Ref:    "dark-mode",
				Specs:  map[string]spec.SpecValue{},
				Events: map[string]interface{}{},
			},
		},
	}

	if err := spec.WriteContainer(path, c); err != nil {
		t.Fatalf("unexpected write error: %v", err)
	}

	parsed, err := spec.ParseContainer(path)
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	if len(parsed.States[0].Events) != 2 {
		t.Errorf("expected 2 events after round-trip, got %d", len(parsed.States[0].Events))
	}
	if parsed.States[0].Events["on-long-press"] != "show-menu" {
		t.Errorf("expected on-long-press 'show-menu', got %v", parsed.States[0].Events["on-long-press"])
	}
}

func TestWriteContainer_EmptySpecsWritesEmptyMap(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.yml")

	c := spec.Container{
		States: []spec.StateSpec{
			{Ref: "default", Specs: map[string]spec.SpecValue{}},
		},
	}

	if err := spec.WriteContainer(path, c); err != nil {
		t.Fatalf("unexpected write error: %v", err)
	}

	parsed, err := spec.ParseContainer(path)
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	if len(parsed.States[0].Specs) != 0 {
		t.Errorf("expected empty specs, got %v", parsed.States[0].Specs)
	}
}

func containsStr(s, substr string) bool {
	return len(s) >= len(substr) && func() bool {
		for i := 0; i <= len(s)-len(substr); i++ {
			if s[i:i+len(substr)] == substr {
				return true
			}
		}
		return false
	}()
}
