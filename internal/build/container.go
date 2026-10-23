package build

import "gopkg.in/yaml.v3"

// Specs is a flat key-value map of property names to string values.
type Specs map[string]string

// Container is a resolved UI container with all $refs inlined.
// Ref and Overrides are input-only fields used during loading; they are
// always zero on containers returned by Load.
type Container struct {
	Path        string      // relative to containers/ root, set by Load
	Image       string      `yaml:"-"` // sibling image filename, set by Load; empty for nested/ref'd containers
	Title       string      `yaml:"title"`
	Description string      `yaml:"description,omitempty"`
	Behavior    []string    `yaml:"behavior,omitempty"`
	Specs       Specs       `yaml:"specs,omitempty"`
	States      []State     `yaml:"states,omitempty"`
	Events      []Event     `yaml:"events,omitempty"`
	Containers  []Container `yaml:"containers,omitempty"`
	Ref         string      `yaml:"$ref,omitempty"`
	Overrides   Specs       `yaml:"overrides,omitempty"`
}

type State struct {
	Title       string `yaml:"title"`
	Description string `yaml:"description,omitempty"`
	Specs       Specs  `yaml:"specs,omitempty"`
}

type Event struct {
	Title       string   `yaml:"title"`
	Description string   `yaml:"description,omitempty"`
	Effects     []Effect `yaml:"effects,omitempty"`
}

type Effect struct {
	Title       string `yaml:"title"`
	Description string `yaml:"description,omitempty"`
}

// UnmarshalYAML allows an Effect to be written as a plain string, which is
// taken as the Description with an empty Title.
func (e *Effect) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind == yaml.ScalarNode {
		e.Title = ""
		return value.Decode(&e.Description)
	}

	type effectAlias Effect
	var full effectAlias
	if err := value.Decode(&full); err != nil {
		return err
	}
	*e = Effect(full)
	return nil
}
