package spec

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type SpecValue struct {
	Value       any
	Description string
}

type Import struct {
	Ref       string
	Line      int
	Overrides map[string]SpecValue
}

type StateSpec struct {
	Ref    string
	Line   int
	Specs  map[string]SpecValue
	Events map[string]interface{}
}

// Action holds the payload fields for a single action type.
type Action map[string]interface{}

// Event is a top-level interaction on a container: a trigger name, optional
// description, and one or more typed actions.
type Event struct {
	Title       string
	Description string
	Actions     map[string]Action
}

type Container struct {
	// Path is relative to the containers/ root directory.
	Path        string
	Title       string
	Description string
	Default     string
	States      []StateSpec
	Imports     []Import
	Events      []Event
}

// containerFile is the raw YAML structure for a container file.
type containerFile struct {
	Title       string         `yaml:"title"`
	Description string         `yaml:"description"`
	Default     string         `yaml:"default"`
	Containers  []containerRef `yaml:"containers"`
	Events      []eventFile    `yaml:"events"`
}

type eventFile struct {
	Title       string            `yaml:"title"`
	Description string            `yaml:"description"`
	Actions     map[string]Action `yaml:"actions"`
}

type containerRef struct {
	Ref string `yaml:"$ref"`
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

	var raw containerFile
	if err := doc.Decode(&raw); err != nil {
		return Container{}, fmt.Errorf("parsing container file: %w", err)
	}

	defaultRef := "default"
	if raw.Default != "" {
		defaultRef = raw.Default
	}

	title := raw.Title
	if title == "" {
		title = strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	}

	c := Container{Default: defaultRef, Title: title, Description: raw.Description}

	// Top-level specs fold into the default state.
	if specsNode := findMappingValue(doc, "specs"); specsNode != nil {
		specs, err := parseSpecsNode(specsNode)
		if err != nil {
			return Container{}, err
		}
		if len(specs) > 0 {
			c.States = append(c.States, StateSpec{Ref: defaultRef, Specs: specs})
		}
	}

	statesSeq := findSequenceNode(doc, "states")
	for _, item := range statesSeq {
		ss, err := parseStateSpecNode(item)
		if err != nil {
			return Container{}, err
		}
		if ss.Ref != "" {
			c.States = append(c.States, ss)
		}
	}

	imports, err := parseContainerImports(doc)
	if err != nil {
		return Container{}, err
	}
	c.Imports = imports

	for _, ef := range raw.Events {
		c.Events = append(c.Events, Event{
			Title:       ef.Title,
			Description: ef.Description,
			Actions:     ef.Actions,
		})
	}

	return c, nil
}

// findMappingValue returns the value node for key within a YAML mapping node,
// or nil if the key is not present.
func findMappingValue(doc *yaml.Node, key string) *yaml.Node {
	if doc.Kind != yaml.MappingNode {
		return nil
	}
	for i := 0; i+1 < len(doc.Content); i += 2 {
		if doc.Content[i].Value == key {
			return doc.Content[i+1]
		}
	}
	return nil
}

// findSequenceNode returns the content slice of a sequence node identified by key
// within a YAML mapping node, or nil if not found.
func findSequenceNode(doc *yaml.Node, key string) []*yaml.Node {
	if doc.Kind != yaml.MappingNode {
		return nil
	}
	for i := 0; i+1 < len(doc.Content); i += 2 {
		if doc.Content[i].Value == key && doc.Content[i+1].Kind == yaml.SequenceNode {
			return doc.Content[i+1].Content
		}
	}
	return nil
}

// parseStateSpecNode parses a state entry from a YAML mapping node.
// Supports two formats:
//   - ref format:      `ref: state-name`
//   - name-as-key:    `state-name:` (first key that is not a known metadata key)
func parseStateSpecNode(node *yaml.Node) (StateSpec, error) {
	if node.Kind != yaml.MappingNode {
		return StateSpec{}, fmt.Errorf("expected state entry to be a mapping")
	}

	var ref string
	var line int
	var events map[string]interface{}

	// Check for explicit `ref` key first.
	for i := 0; i+1 < len(node.Content); i += 2 {
		if node.Content[i].Value == "ref" {
			ref = node.Content[i+1].Value
			line = node.Content[i+1].Line
			break
		}
	}

	// Name-as-key format: first key that isn't a known metadata key is the state name.
	if ref == "" && len(node.Content) >= 2 {
		known := map[string]bool{"specs": true, "events": true, "behavior": true}
		for i := 0; i+1 < len(node.Content); i += 2 {
			if !known[node.Content[i].Value] {
				ref = node.Content[i].Value
				line = node.Content[i].Line
				break
			}
		}
	}

	var specsNode *yaml.Node
	for i := 0; i+1 < len(node.Content); i += 2 {
		switch node.Content[i].Value {
		case "specs":
			specsNode = node.Content[i+1]
		case "events":
			if err := node.Content[i+1].Decode(&events); err != nil {
				return StateSpec{}, fmt.Errorf("parsing events: %w", err)
			}
		}
	}

	var specs map[string]SpecValue
	if specsNode != nil {
		var err error
		specs, err = parseSpecsNode(specsNode)
		if err != nil {
			return StateSpec{}, err
		}
	}

	return StateSpec{Ref: ref, Line: line, Specs: specs, Events: events}, nil
}

// parseContainerImports extracts Import entries from the "containers" sequence
// of a YAML mapping node. Keys other than "$ref" in each entry are treated as
// spec overrides applied when the import is resolved.
func parseContainerImports(doc *yaml.Node) ([]Import, error) {
	if doc.Kind != yaml.MappingNode {
		return nil, nil
	}
	for i := 0; i+1 < len(doc.Content); i += 2 {
		if doc.Content[i].Value != "containers" {
			continue
		}
		seq := doc.Content[i+1]
		if seq.Kind != yaml.SequenceNode {
			break
		}
		var imports []Import
		for _, item := range seq.Content {
			if item.Kind != yaml.MappingNode {
				continue
			}
			var imp Import
			var overrideNodes map[string]yaml.Node
			for j := 0; j+1 < len(item.Content); j += 2 {
				k := item.Content[j].Value
				v := item.Content[j+1]
				if k == "$ref" {
					imp.Ref = v.Value
					imp.Line = v.Line
				} else {
					if overrideNodes == nil {
						overrideNodes = make(map[string]yaml.Node)
					}
					overrideNodes[k] = *v
				}
			}
			if imp.Ref == "" {
				continue
			}
			if len(overrideNodes) > 0 {
				specs, err := parseSpecMap(overrideNodes)
				if err != nil {
					return nil, fmt.Errorf("parsing overrides for %s: %w", imp.Ref, err)
				}
				imp.Overrides = specs
			}
			imports = append(imports, imp)
		}
		return imports, nil
	}
	return nil, nil
}

// parseSpecsNode parses a specs node that may be either a YAML mapping or a
// sequence of single-key mappings (e.g. "- key: value").
func parseSpecsNode(node *yaml.Node) (map[string]SpecValue, error) {
	n := node
	if n.Kind == yaml.AliasNode {
		n = n.Alias
	}
	switch n.Kind {
	case yaml.MappingNode:
		var raw map[string]yaml.Node
		if err := n.Decode(&raw); err != nil {
			return nil, fmt.Errorf("parsing specs: %w", err)
		}
		return parseSpecMap(raw)
	case yaml.SequenceNode:
		result := make(map[string]SpecValue)
		for _, item := range n.Content {
			if item.Kind != yaml.MappingNode {
				continue
			}
			for i := 0; i+1 < len(item.Content); i += 2 {
				key := item.Content[i].Value
				sv, err := parseSpecValue(*item.Content[i+1])
				if err != nil {
					return nil, fmt.Errorf("spec %q: %w", key, err)
				}
				result[key] = sv
			}
		}
		return result, nil
	default:
		return nil, fmt.Errorf("specs: expected mapping or sequence, got node kind %v", n.Kind)
	}
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
