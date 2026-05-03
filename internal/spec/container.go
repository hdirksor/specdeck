package spec

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type SpecValue struct {
	Value       any
	Description string
}

type Import struct {
	Ref  string
	Line int
}

type StateSpec struct {
	Ref    string
	Line   int
	Specs  map[string]SpecValue
	Events map[string]interface{}
}

type Container struct {
	// Path is relative to the containers/ root directory.
	Path    string
	Default string
	States  []StateSpec
	Imports []Import
}

// containerFile is the raw YAML structure for a container file.
type containerFile struct {
	Default    string               `yaml:"default"`
	Specs      map[string]yaml.Node `yaml:"specs"`
	States     []containerStateRaw  `yaml:"states"`
	Containers []containerRef       `yaml:"containers"`
}

type containerRef struct {
	Ref string `yaml:"$ref"`
}

type containerStateRaw struct {
	Ref    string                 `yaml:"ref"`
	Specs  map[string]yaml.Node   `yaml:"specs"`
	Events map[string]interface{} `yaml:"events"`
}

func ParseContainer(path string) (Container, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Container{}, fmt.Errorf("reading container file: %w", err)
	}

	var root yaml.Node
	if err := yaml.Unmarshal(data, &root); err != nil {
		return Container{}, fmt.Errorf("parsing container file: %w", err)
	}

	if root.Kind != yaml.DocumentNode || len(root.Content) == 0 {
		return Container{}, nil
	}

	doc := root.Content[0]
	stateLines := sequenceKeyLines(doc, "states", "ref")
	importLines := sequenceKeyLines(doc, "containers", "$ref")

	var raw containerFile
	if err := doc.Decode(&raw); err != nil {
		return Container{}, fmt.Errorf("parsing container file: %w", err)
	}

	defaultRef := "default"
	if raw.Default != "" {
		defaultRef = raw.Default
	}

	c := Container{Default: defaultRef}

	// Top-level specs fold into the default state.
	if len(raw.Specs) > 0 {
		specs, err := parseSpecMap(raw.Specs)
		if err != nil {
			return Container{}, err
		}
		c.States = append(c.States, StateSpec{Ref: defaultRef, Specs: specs})
	}

	for _, s := range raw.States {
		specs, err := parseSpecMap(s.Specs)
		if err != nil {
			return Container{}, err
		}
		c.States = append(c.States, StateSpec{
			Ref:    s.Ref,
			Line:   stateLines[s.Ref],
			Specs:  specs,
			Events: s.Events,
		})
	}

	for _, cr := range raw.Containers {
		c.Imports = append(c.Imports, Import{
			Ref:  cr.Ref,
			Line: importLines[cr.Ref],
		})
	}

	return c, nil
}

// sequenceKeyLines returns a map of value → line number for keyField entries
// within the sequence identified by seqKey in a YAML mapping node.
func sequenceKeyLines(doc *yaml.Node, seqKey, keyField string) map[string]int {
	lines := map[string]int{}
	if doc.Kind != yaml.MappingNode {
		return lines
	}
	for i := 0; i+1 < len(doc.Content); i += 2 {
		if doc.Content[i].Value != seqKey {
			continue
		}
		seq := doc.Content[i+1]
		if seq.Kind != yaml.SequenceNode {
			break
		}
		for _, item := range seq.Content {
			if item.Kind != yaml.MappingNode {
				continue
			}
			for j := 0; j+1 < len(item.Content); j += 2 {
				k := item.Content[j]
				v := item.Content[j+1]
				if k.Value == keyField {
					lines[v.Value] = v.Line
				}
			}
		}
		break
	}
	return lines
}

// parseSpecMap handles both shorthand (scalar) and verbose (mapping) spec values.
func parseSpecMap(raw map[string]yaml.Node) (map[string]SpecValue, error) {
	specs := make(map[string]SpecValue, len(raw))
	for key, node := range raw {
		sv, err := parseSpecValue(node)
		if err != nil {
			return nil, fmt.Errorf("spec %q: %w", key, err)
		}
		specs[key] = sv
	}
	return specs, nil
}

func parseSpecValue(node yaml.Node) (SpecValue, error) {
	// Dereference alias nodes.
	n := &node
	if n.Kind == yaml.AliasNode {
		n = n.Alias
	}

	switch n.Kind {
	case yaml.ScalarNode:
		return SpecValue{Value: n.Value}, nil
	case yaml.MappingNode:
		var verbose struct {
			Value       interface{} `yaml:"value"`
			Description string      `yaml:"description"`
		}
		if err := n.Decode(&verbose); err != nil {
			return SpecValue{}, fmt.Errorf("decoding verbose spec: %w", err)
		}
		return SpecValue{Value: verbose.Value, Description: verbose.Description}, nil
	default:
		return SpecValue{}, fmt.Errorf("unexpected YAML node kind %v", n.Kind)
	}
}

// LoadContainers walks the containers root directory and parses every container file.
// Non-leaf containers are represented by index.yml files.
func LoadContainers(root string) ([]Container, error) {
	var containers []Container

	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		ext := filepath.Ext(d.Name())
		if ext != ".yml" {
			return nil
		}

		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}

		c, err := ParseContainer(path)
		if err != nil {
			return fmt.Errorf("parsing container %s: %w", rel, err)
		}

		c.Path = rel
		containers = append(containers, c)
		return nil
	})

	return containers, err
}
