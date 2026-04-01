package spec_test

import (
	"testing"

	"github.com/hdickson/specdeck/internal/spec"
)

func facts(defs ...spec.Fact) []spec.Fact { return defs }

func crossEnum(name string, values ...string) spec.Fact {
	return spec.Fact{Name: name, Type: spec.FactTypeEnum, Values: values, Scope: spec.FactScopeCross}
}

func crossBool(name string) spec.Fact {
	return spec.Fact{Name: name, Type: spec.FactTypeBoolean, Scope: spec.FactScopeCross}
}

func isolatedEnum(name string, values ...string) spec.Fact {
	return spec.Fact{Name: name, Type: spec.FactTypeEnum, Values: values, Scope: spec.FactScopeIsolated}
}

func manualBool(name string) spec.Fact {
	return spec.Fact{Name: name, Type: spec.FactTypeBoolean, Scope: spec.FactScopeManual}
}

func stateNames(states []spec.State) []string {
	names := make([]string, len(states))
	for i, s := range states {
		names[i] = s.Name
	}
	return names
}

func hasState(states []spec.State, name string) bool {
	for _, s := range states {
		if s.Name == name {
			return true
		}
	}
	return false
}

func TestGenerateStates_AllDefaultsNamedDefault(t *testing.T) {
	states := spec.GenerateStates(facts(
		crossEnum("subscription", "free", "premium"),
		crossBool("is-logged-in"),
	))
	if !hasState(states, "default") {
		t.Errorf("expected a 'default' state, got: %v", stateNames(states))
	}
}

func TestGenerateStates_CrossCartesianProduct(t *testing.T) {
	states := spec.GenerateStates(facts(
		crossEnum("subscription", "free", "premium"),
		crossBool("is-logged-in"),
	))
	// 2 subscription values × 2 boolean values = 4 states
	if len(states) != 4 {
		t.Errorf("expected 4 states, got %d: %v", len(states), stateNames(states))
	}
}

func TestGenerateStates_BooleanNaming(t *testing.T) {
	states := spec.GenerateStates(facts(
		crossBool("is-logged-in"),
	))
	names := stateNames(states)
	if !hasState(states, "default") {
		t.Errorf("expected 'default' (logged-in=false is default), got: %v", names)
	}
	if !hasState(states, "logged-in") {
		t.Errorf("expected 'logged-in', got: %v", names)
	}
}

func TestGenerateStates_IsolatedProducesNonDefaultValues(t *testing.T) {
	states := spec.GenerateStates(facts(
		isolatedEnum("language", "en", "es", "fr"),
	))
	// 'en' is default so no state for it; 'es' and 'fr' get states
	if len(states) != 3 {
		t.Errorf("expected 3 states (default + es + fr), got %d: %v", len(states), stateNames(states))
	}
	if !hasState(states, "es") {
		t.Errorf("expected 'es' state, got: %v", stateNames(states))
	}
	if !hasState(states, "fr") {
		t.Errorf("expected 'fr' state, got: %v", stateNames(states))
	}
}

func TestGenerateStates_ManualFactsIgnored(t *testing.T) {
	states := spec.GenerateStates(facts(
		manualBool("experiment-new-feed"),
	))
	// manual facts produce no states — not even a default
	if len(states) != 0 {
		t.Errorf("expected 0 states from manual-only facts, got %d: %v", len(states), stateNames(states))
	}
}

func TestGenerateStates_CrossAndIsolatedCombined(t *testing.T) {
	states := spec.GenerateStates(facts(
		crossEnum("subscription", "free", "premium"),
		crossBool("is-logged-in"),
		isolatedEnum("language", "en", "es", "fr"),
	))
	// cross: 2×2 = 4 states
	// isolated: es, fr (en is default, already covered by cross default state)
	if len(states) != 6 {
		t.Errorf("expected 6 states, got %d: %v", len(states), stateNames(states))
	}
}

func TestGenerateStates_EnumNaming(t *testing.T) {
	states := spec.GenerateStates(facts(
		crossEnum("subscription", "free", "premium"),
		crossEnum("theme", "light", "dark"),
	))
	if !hasState(states, "default") {
		t.Errorf("expected 'default' state")
	}
	// free is default subscription → omitted; only theme=dark varies
	if !hasState(states, "dark") {
		t.Errorf("expected 'dark' state, got: %v", stateNames(states))
	}
	// light is default theme → omitted; only subscription=premium varies
	if !hasState(states, "premium") {
		t.Errorf("expected 'premium' state, got: %v", stateNames(states))
	}
	if !hasState(states, "premium-dark") {
		t.Errorf("expected 'premium-dark' state, got: %v", stateNames(states))
	}
}

func TestGenerateStates_StateFacts(t *testing.T) {
	states := spec.GenerateStates(facts(
		crossEnum("subscription", "free", "premium"),
	))
	for _, s := range states {
		if _, ok := s.Facts["subscription"]; !ok {
			t.Errorf("state %q missing subscription fact", s.Name)
		}
	}
}
