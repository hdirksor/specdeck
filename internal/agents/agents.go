// Package agents installs specdeck's commands for AI coding agents.
package agents

import (
	"bytes"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/hdirksor/specdeck/internal/config"
)

// commandFiles holds one directory of command files per supported agent.
//
//go:embed claude
var commandFiles embed.FS

// ErrNoAgents is returned when no agent is named and none are recorded.
var ErrNoAgents = errors.New("no agents set up — run `specdeck agents claude` to add one")

// Installed describes an agent whose commands were written.
type Installed struct {
	Name string
	// Dir is the slash-separated folder, relative to the repo, that specdeck
	// owns for this agent.
	Dir string
}

type agent struct {
	name string
	dir  string
}

var supported = []agent{
	{name: "claude", dir: ".claude/commands/specdeck"},
}

type commandFile struct {
	name    string
	content []byte
}

// Update installs or updates the commands for names, or for every agent
// recorded in specdeck.yml when names is empty, and records them there.
// Files in an agent's folder that specdeck no longer ships are removed.
func Update(dir string, names []string) ([]Installed, error) {
	cfg, err := config.Load(dir)
	if err != nil {
		return nil, err
	}
	if len(names) == 0 {
		names = cfg.Agents
	}
	if len(names) == 0 {
		return nil, ErrNoAgents
	}
	selected, err := lookup(names)
	if err != nil {
		return nil, err
	}

	var installed []Installed
	for _, a := range selected {
		if err := a.install(dir); err != nil {
			return nil, err
		}
		installed = append(installed, Installed{Name: a.name, Dir: a.dir})
	}

	for _, a := range selected {
		if !slices.Contains(cfg.Agents, a.name) {
			cfg.Agents = append(cfg.Agents, a.name)
		}
	}
	slices.Sort(cfg.Agents)
	if err := config.Save(dir, cfg); err != nil {
		return nil, err
	}
	return installed, nil
}

// Stale returns the files that Update would change for the agents recorded
// in specdeck.yml, as slash-separated paths relative to dir.
func Stale(dir string) ([]string, error) {
	cfg, err := config.Load(dir)
	if err != nil {
		return nil, err
	}
	selected, err := lookup(cfg.Agents)
	if err != nil {
		return nil, err
	}
	var stale []string
	for _, a := range selected {
		paths, err := a.stale(dir)
		if err != nil {
			return nil, err
		}
		stale = append(stale, paths...)
	}
	return stale, nil
}

func lookup(names []string) ([]agent, error) {
	var selected []agent
	for _, name := range names {
		i := slices.IndexFunc(supported, func(a agent) bool { return a.name == name })
		if i < 0 {
			return nil, fmt.Errorf("unknown agent %q (supported: %s)", name, supportedNames())
		}
		if !slices.Contains(selected, supported[i]) {
			selected = append(selected, supported[i])
		}
	}
	return selected, nil
}

func supportedNames() string {
	names := make([]string, len(supported))
	for i, a := range supported {
		names[i] = a.name
	}
	return strings.Join(names, ", ")
}

func (a agent) target(dir string) string {
	return filepath.Join(dir, filepath.FromSlash(a.dir))
}

func (a agent) files() ([]commandFile, error) {
	entries, err := fs.ReadDir(commandFiles, a.name)
	if err != nil {
		return nil, fmt.Errorf("reading embedded %s commands: %w", a.name, err)
	}
	var files []commandFile
	for _, entry := range entries {
		content, err := fs.ReadFile(commandFiles, a.name+"/"+entry.Name())
		if err != nil {
			return nil, fmt.Errorf("reading embedded %s command %s: %w", a.name, entry.Name(), err)
		}
		files = append(files, commandFile{name: entry.Name(), content: content})
	}
	return files, nil
}

func (a agent) install(dir string) error {
	files, err := a.files()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(a.target(dir), 0755); err != nil {
		return fmt.Errorf("creating %s: %w", a.dir, err)
	}
	for _, f := range files {
		if err := os.WriteFile(filepath.Join(a.target(dir), f.name), f.content, 0644); err != nil {
			return fmt.Errorf("writing %s/%s: %w", a.dir, f.name, err)
		}
	}
	extras, err := a.extraFiles(dir, files)
	if err != nil {
		return err
	}
	for _, name := range extras {
		if err := os.Remove(filepath.Join(a.target(dir), name)); err != nil {
			return fmt.Errorf("removing %s/%s: %w", a.dir, name, err)
		}
	}
	return nil
}

func (a agent) stale(dir string) ([]string, error) {
	files, err := a.files()
	if err != nil {
		return nil, err
	}
	var stale []string
	for _, f := range files {
		installed, err := os.ReadFile(filepath.Join(a.target(dir), f.name))
		if os.IsNotExist(err) || (err == nil && !bytes.Equal(installed, f.content)) {
			stale = append(stale, a.dir+"/"+f.name)
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("reading %s/%s: %w", a.dir, f.name, err)
		}
	}
	extras, err := a.extraFiles(dir, files)
	if err != nil {
		return nil, err
	}
	for _, name := range extras {
		stale = append(stale, a.dir+"/"+name)
	}
	return stale, nil
}

// extraFiles lists files in the agent's folder that specdeck does not ship.
func (a agent) extraFiles(dir string, shipped []commandFile) ([]string, error) {
	entries, err := os.ReadDir(a.target(dir))
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", a.dir, err)
	}
	var extras []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if !slices.ContainsFunc(shipped, func(f commandFile) bool { return f.name == entry.Name() }) {
			extras = append(extras, entry.Name())
		}
	}
	return extras, nil
}
