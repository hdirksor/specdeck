package spec_test

import (
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
