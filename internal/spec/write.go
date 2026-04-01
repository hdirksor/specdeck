package spec

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type containerOutput struct {
	Default string              `yaml:"default,omitempty"`
	States  []stateSpecOutput   `yaml:"states"`
}

type stateSpecOutput struct {
	Ref    string                     `yaml:"ref"`
	Specs  map[string]specValueOutput `yaml:"specs"`
	Events map[string]interface{}     `yaml:"events"`
}

type specValueOutput struct {
	Value       interface{}
	Description string
}

func (s specValueOutput) MarshalYAML() (interface{}, error) {
	if s.Description == "" {
		return s.Value, nil
	}
	return struct {
		Value       interface{} `yaml:"value"`
		Description string      `yaml:"description"`
	}{s.Value, s.Description}, nil
}

func WriteStates(path string, states []State) error {
	type stateOut struct {
		Name    string                 `yaml:"name"`
		Summary string                 `yaml:"summary"`
		Facts   map[string]interface{} `yaml:"facts"`
	}
	out := make([]stateOut, len(states))
	for i, s := range states {
		out[i] = stateOut{Name: s.Name, Summary: s.Summary, Facts: s.Facts}
	}

	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("creating states file: %w", err)
	}
	defer f.Close()

	enc := yaml.NewEncoder(f)
	enc.SetIndent(2)
	return enc.Encode(out)
}

func WriteContainer(path string, c Container) error {
	out := containerOutput{Default: c.Default}
	for _, ss := range c.States {
		specs := make(map[string]specValueOutput, len(ss.Specs))
		for k, v := range ss.Specs {
			specs[k] = specValueOutput{Value: v.Value, Description: v.Description}
		}
		out.States = append(out.States, stateSpecOutput{Ref: ss.Ref, Specs: specs, Events: ss.Events})
	}

	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("creating container file: %w", err)
	}
	defer f.Close()

	enc := yaml.NewEncoder(f)
	enc.SetIndent(2)
	return enc.Encode(out)
}
