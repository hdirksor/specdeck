package spec

import (
	"fmt"
	"os"
	"path/filepath"
)

// Validate checks cross-references between facts, states, and containers.
// Returns one ValidationError per violation; an empty slice means the project is valid.
func Validate(facts []Fact, states []State, containers []Container) []ValidationError {
	factNames := make(map[string]bool, len(facts))
	for _, f := range facts {
		factNames[f.Name] = true
	}

	stateNames := make(map[string]bool, len(states))
	for _, s := range states {
		stateNames[s.Name] = true
	}

	var errs []ValidationError

	for _, s := range states {
		for factName := range s.Facts {
			if !factNames[factName] {
				errs = append(errs, ValidationError{
					Message: fmt.Sprintf("state %q references unknown fact %q", s.Name, factName),
				})
			}
		}
	}

	for _, c := range containers {
		for _, ss := range c.States {
			if !stateNames[ss.Ref] {
				errs = append(errs, ValidationError{
					File:    c.Path,
					Message: fmt.Sprintf("references unknown state %q", ss.Ref),
				})
			}
		}
	}

	return errs
}

// ValidateFactFiles parses every .yml file in dir and returns one ValidationError per invalid file.
func ValidateFactFiles(dir string) (validFiles []string, errs []ValidationError) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, []ValidationError{{Message: err.Error()}}
	}
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".yml" {
			continue
		}
		if _, parseErr := ParseFacts(filepath.Join(dir, e.Name())); parseErr != nil {
			errs = append(errs, ValidationError{File: e.Name(), Message: parseErr.Error()})
		} else {
			validFiles = append(validFiles, e.Name())
		}
	}
	return validFiles, errs
}

// ValidateStateFiles parses every .yml file in dir and returns one ValidationError per invalid file.
func ValidateStateFiles(dir string) (validFiles []string, errs []ValidationError) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, []ValidationError{{Message: err.Error()}}
	}
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".yml" {
			continue
		}
		if _, err := ParseStates(filepath.Join(dir, e.Name())); err != nil {
			errs = append(errs, ValidationError{File: e.Name(), Message: err.Error()})
		} else {
			validFiles = append(validFiles, e.Name())
		}
	}
	return validFiles, errs
}
