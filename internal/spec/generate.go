package spec

import "strings"

// GenerateStates produces states from facts based on their scope.
// Cross facts are fully cartesian-producted together.
// Isolated facts each contribute one state per non-default value.
// Manual facts are ignored.
func GenerateStates(facts []Fact) []State {
	var cross, isolated []Fact
	for _, f := range facts {
		switch f.Scope {
		case FactScopeCross:
			cross = append(cross, f)
		case FactScopeIsolated:
			isolated = append(isolated, f)
		}
	}

	var states []State

	if len(cross) > 0 {
		for _, combo := range cartesianProduct(cross) {
			states = append(states, stateFromCombo(combo, cross))
		}
	}

	for _, f := range isolated {
		defaultVal := defaultValue(f)
		// Build the default cross values to hold constant.
		crossDefaults := make(map[string]string, len(cross))
		for _, cf := range cross {
			crossDefaults[cf.Name] = defaultValue(cf)
		}

		for _, val := range f.Values {
			if val == defaultVal {
				continue
			}
			facts := make(map[string]interface{}, len(crossDefaults)+1)
			for k, v := range crossDefaults {
				facts[k] = v
			}
			facts[f.Name] = val

			name := valueLabel(f, val)
			summary := "Default state with " + f.Name + "=" + val
			states = append(states, State{Name: name, Summary: summary, Facts: facts})
		}
	}

	// If there are no cross facts but there are isolated facts, ensure a default state exists.
	if len(cross) == 0 && len(isolated) > 0 {
		states = append([]State{defaultState(nil)}, states...)
	}

	return states
}

// cartesianProduct returns all combinations of values for a set of cross facts.
func cartesianProduct(facts []Fact) []map[string]string {
	result := []map[string]string{{}}
	for _, f := range facts {
		var next []map[string]string
		for _, combo := range result {
			for _, val := range factValues(f) {
				m := make(map[string]string, len(combo)+1)
				for k, v := range combo {
					m[k] = v
				}
				m[f.Name] = val
				next = append(next, m)
			}
		}
		result = next
	}
	return result
}

// factValues returns the values to iterate over for a fact.
// For booleans: ["false", "true"]; for enums: the values list.
func factValues(f Fact) []string {
	if f.Type == FactTypeBoolean {
		return []string{"false", "true"}
	}
	return f.Values
}

// defaultValue returns the default value for a fact.
// Boolean defaults to false; enum defaults to the first value.
func defaultValue(f Fact) string {
	if f.Type == FactTypeBoolean {
		return "false"
	}
	return f.Values[0]
}

// stateFromCombo builds a State from a fact-value combination.
func stateFromCombo(combo map[string]string, orderedFacts []Fact) State {
	name := stateName(combo, orderedFacts)
	facts := make(map[string]interface{}, len(combo))
	for k, v := range combo {
		facts[k] = v
	}

	summary := "Generated state: " + name
	if name == "default" {
		summary = "All facts at their default values"
	}

	return State{Name: name, Summary: summary, Facts: facts}
}

// stateName derives a human-readable state name from a fact-value combination.
// If all values are defaults the name is "default".
func stateName(combo map[string]string, orderedFacts []Fact) string {
	allDefault := true
	for _, f := range orderedFacts {
		if combo[f.Name] != defaultValue(f) {
			allDefault = false
			break
		}
	}
	if allDefault {
		return "default"
	}

	var parts []string
	for _, f := range orderedFacts {
		val := combo[f.Name]
		if val == defaultValue(f) {
			continue
		}
		parts = append(parts, valueLabel(f, val))
	}
	return strings.Join(parts, "-")
}

// valueLabel converts a fact+value pair to its name segment.
// Boolean true → fact name (stripping leading "is-"); false → "not-"+name.
// Enum → the value itself.
func valueLabel(f Fact, val string) string {
	if f.Type == FactTypeBoolean {
		base := strings.TrimPrefix(f.Name, "is-")
		if val == "true" {
			return base
		}
		return "not-" + base
	}
	return val
}

// defaultState returns a state representing all facts at default values.
func defaultState(crossFacts []Fact) State {
	facts := make(map[string]interface{}, len(crossFacts))
	for _, f := range crossFacts {
		facts[f.Name] = defaultValue(f)
	}
	return State{Name: "default", Summary: "All facts at their default values", Facts: facts}
}
