package spec

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Section is one imported sub-container within a built container output.
type Section struct {
	Title  string
	States map[string]map[string]SpecValue
	Events []Event
}

// resolveStateSpecs computes per-state specs for a container with inheritance:
// each state = default specs merged with that state's own overrides (state wins).
func resolveStateSpecs(c Container) map[string]map[string]SpecValue {
	if len(c.States) == 0 {
		return nil
	}

	defaultSpecs := make(map[string]SpecValue)
	for _, ss := range c.States {
		if ss.Ref == c.Default {
			defaultSpecs = ss.Specs
			break
		}
	}

	result := make(map[string]map[string]SpecValue, len(c.States))
	for _, ss := range c.States {
		resolved := make(map[string]SpecValue, len(defaultSpecs))
		for k, v := range defaultSpecs {
			resolved[k] = v
		}
		for k, v := range ss.Specs {
			resolved[k] = v
		}
		result[ss.Ref] = resolved
	}
	return result
}

// ResolveContainerSections returns the container's own per-state specs and
// one Section per imported container.
func ResolveContainerSections(c Container, containersRoot string, byPath map[string]Container) (map[string]map[string]SpecValue, []Section) {
	ownStates := resolveStateSpecs(c)

	var imports []Section
	for _, imp := range c.Imports {
		absDir := filepath.Dir(filepath.Join(containersRoot, c.Path))
		absImport := filepath.Join(absDir, imp.Ref)
		rel, err := filepath.Rel(containersRoot, absImport)
		if err != nil {
			continue
		}
		imported, ok := byPath[rel]
		if !ok {
			continue
		}
		states := resolveStateSpecs(imported)
		// Import-site overrides apply across all states.
		if len(imp.Overrides) > 0 && len(states) > 0 {
			for stateName, specs := range states {
				merged := make(map[string]SpecValue, len(specs))
				for k, v := range specs {
					merged[k] = v
				}
				for k, v := range imp.Overrides {
					merged[k] = v
				}
				states[stateName] = merged
			}
		}
		imports = append(imports, Section{Title: imported.Title, States: states, Events: imported.Events})
	}

	return ownStates, imports
}

type builtContainerOutput struct {
	Title       string                 `yaml:"title"`
	Description string                 `yaml:"description,omitempty"`
	States      map[string]stateOutput `yaml:"states,omitempty"`
	Sections    []sectionOutput        `yaml:"containers,omitempty"`
	Events      []eventOutput          `yaml:"events,omitempty"`
}

type stateOutput struct {
	Specs map[string]specValueOutput `yaml:"specs,omitempty"`
}

type sectionOutput struct {
	Title  string                 `yaml:"title"`
	States map[string]stateOutput `yaml:"states,omitempty"`
	Events []eventOutput          `yaml:"events,omitempty"`
}

type eventOutput struct {
	Title       string            `yaml:"title"`
	Description string            `yaml:"description,omitempty"`
	Actions     map[string]Action `yaml:"actions,omitempty"`
}

func eventToOutput(ev Event) eventOutput {
	return eventOutput{Title: ev.Title, Description: ev.Description, Actions: ev.Actions}
}

func specsToOutput(specs map[string]SpecValue) map[string]specValueOutput {
	if len(specs) == 0 {
		return nil
	}
	out := make(map[string]specValueOutput, len(specs))
	for k, v := range specs {
		out[k] = specValueOutput{Value: v.Value, Description: v.Description}
	}
	return out
}

func statesToOutput(states map[string]map[string]SpecValue) map[string]stateOutput {
	if len(states) == 0 {
		return nil
	}
	out := make(map[string]stateOutput, len(states))
	for name, specs := range states {
		out[name] = stateOutput{Specs: specsToOutput(specs)}
	}
	return out
}

func WriteBuiltContainer(path string, c Container, ownStates map[string]map[string]SpecValue, imports []Section) error {
	out := builtContainerOutput{
		Title:       c.Title,
		Description: c.Description,
		States:      statesToOutput(ownStates),
	}

	for _, s := range imports {
		sec := sectionOutput{
			Title:  s.Title,
			States: statesToOutput(s.States),
		}
		for _, ev := range s.Events {
			sec.Events = append(sec.Events, eventToOutput(ev))
		}
		out.Sections = append(out.Sections, sec)
	}

	for _, ev := range c.Events {
		out.Events = append(out.Events, eventToOutput(ev))
	}

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("creating directory: %w", err)
	}
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("creating file: %w", err)
	}
	defer f.Close()

	enc := yaml.NewEncoder(f)
	enc.SetIndent(2)
	return enc.Encode(out)
}
