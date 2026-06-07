package build

// Specs is a flat key-value map of property names to string values.
type Specs map[string]string

// Container is a resolved UI container with all $refs inlined.
// Ref and Overrides are input-only fields used during loading; they are
// always zero on containers returned by Load.
type Container struct {
	Path        string      // relative to containers/ root, set by Load
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
	Actions     []Action `yaml:"actions,omitempty"`
}

type Action struct {
	Title       string `yaml:"title"`
	Description string `yaml:"description,omitempty"`
}
