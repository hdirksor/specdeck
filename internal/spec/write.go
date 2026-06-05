package spec

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type containerOutput struct {
	Default string            `yaml:"default,omitempty"`
	States  []stateSpecOutput `yaml:"states"`
}

type stateSpecOutput struct {
	Ref    string                 `yaml:"ref"`
	Specs  map[string]SpecValue   `yaml:"specs"`
	Events map[string]interface{} `yaml:"events"`
}

func WriteContainer(path string, c Container) error {
	out := containerOutput{Default: c.Default}
	for _, ss := range c.States {
		out.States = append(out.States, stateSpecOutput{Ref: ss.Ref, Specs: ss.Specs, Events: ss.Events})
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
