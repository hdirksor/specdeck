package build

import (
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"
)

// open/close inventory (inline {{if}}...{{end}} on one line not counted):
//  opens:  range(outer) if(desc) if(behavior) range(behavior) if(specs) range(specs)
//          range(states) if(desc) range(specs)
//          if(events) range(events) range(effects)
//  closes: end×12 — one per open above
const markdownTmpl = `{{- range .}}
{{.Heading}} {{.Title}}
{{- if .Description}}

{{.Description}}
{{- end}}
{{- if .Behavior}}

{{- range .Behavior}}
- {{.}}
{{- end}}

{{- end}}
{{- if .Specs}}

| Spec | Value |
|------|-------|
{{- range .Specs}}
| {{.Key}} | {{.Value}} |
{{- end}}

{{- end}}
{{- range .States}}

{{.Heading}} {{.Title}}
{{- if .Description}}

{{.Description}}
{{- end}}
{{- if .Specs}}

| Spec | Value |
|------|-------|
{{- range .Specs}}
| {{.Key}} | {{.Value}} |
{{- end}}

{{- end}}
{{- end}}
{{- if .Events}}

{{.EventsHeading}} Events

{{- range .Events}}
**{{.Title}}**{{if .Description}} — {{.Description}}{{end}}
{{- range .Effects}}
- {{.Title}}{{if .Description}}: {{.Description}}{{end}}
{{- end}}

{{- end}}
{{- end}}

{{- end}}`

var containerTemplate = template.Must(template.New("container").Parse(markdownTmpl))

type specRow struct {
	Key   string
	Value string
}

type stateSection struct {
	Heading     string
	Title       string
	Description string
	Specs       []specRow
}

type renderNode struct {
	Heading       string
	Title         string
	Description   string
	Behavior      []string
	Specs         []specRow
	States        []stateSection
	EventsHeading string
	Events        []Event
}

// WriteTmpl renders container c as markdown using a text/template and writes it to path.
func Write(c Container, path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("creating directory: %w", err)
	}
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("creating file: %w", err)
	}
	defer f.Close()
	return containerTemplate.Execute(f, flattenContainer(c, 1))
}

func flattenContainer(c Container, level int) []renderNode {
	node := renderNode{
		Heading:       strings.Repeat("#", level),
		Title:         c.Title,
		Description:   c.Description,
		Behavior:      c.Behavior,
		EventsHeading: strings.Repeat("#", level+1),
	}

	if len(c.States) == 0 {
		node.Specs = toSpecRows(c.Specs)
	} else {
		subHeading := strings.Repeat("#", level+1)
		for _, s := range c.States {
			title := s.Title
			if title == "" {
				title = "State"
			}
			node.States = append(node.States, stateSection{
				Heading:     subHeading,
				Title:       title,
				Description: s.Description,
				Specs:       toSpecRows(mergeSpecMaps(c.Specs, s.Specs)),
			})
		}
	}

	node.Events = c.Events

	nodes := []renderNode{node}
	for _, sub := range c.Containers {
		nodes = append(nodes, flattenContainer(sub, level+1)...)
	}
	return nodes
}

func toSpecRows(specs Specs) []specRow {
	if len(specs) == 0 {
		return nil
	}
	keys := make([]string, 0, len(specs))
	for k := range specs {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	rows := make([]specRow, len(keys))
	for i, k := range keys {
		rows[i] = specRow{Key: k, Value: specs[k]}
	}
	return rows
}

func mergeSpecMaps(base, overrides Specs) Specs {
	if len(base) == 0 && len(overrides) == 0 {
		return nil
	}
	merged := make(Specs, len(base)+len(overrides))
	maps.Copy(merged, base)
	maps.Copy(merged, overrides)
	return merged
}
