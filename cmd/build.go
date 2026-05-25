package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hdickson/specdeck/internal/spec"
	"github.com/spf13/cobra"
)

// dirTitle converts a directory name like "note-list" to "Note List".
func dirTitle(name string) string {
	words := strings.FieldsFunc(name, func(r rune) bool { return r == '-' || r == '_' })
	for i, w := range words {
		if len(w) > 0 {
			words[i] = strings.ToUpper(w[:1]) + w[1:]
		}
	}
	return strings.Join(words, " ")
}

var buildFormat string

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Resolve container refs and write flat specs to dist/",
	Args:  cobra.NoArgs,
	RunE:  runBuild,
}

func init() {
	rootCmd.AddCommand(buildCmd)
	buildCmd.Flags().StringVar(&buildFormat, "format", "markdown", "output format: markdown or yaml")
}

func runBuild(cmd *cobra.Command, args []string) error {
	root, err := os.Getwd()
	if err != nil {
		return err
	}

	containersRoot := filepath.Join(root, "containers")
	containers, err := spec.LoadContainers(containersRoot)
	if err != nil {
		return fmt.Errorf("\U0001F7E5  %w", err)
	}

	byPath := make(map[string]spec.Container, len(containers))
	for _, c := range containers {
		byPath[c.Path] = c
	}

	distRoot := filepath.Join(root, "dist")

	// For markdown builds: track dirs covered by index.yml → _index.md and dirs with only leaf files.
	sectionDirs := make(map[string]bool)
	leafDirs := make(map[string]string) // abs dir path → dir base name

	for _, c := range containers {
		ownStates, imports := spec.ResolveContainerSections(c, containersRoot, byPath)

		stem := strings.TrimSuffix(c.Path, filepath.Ext(c.Path))
		var outPath string
		var writeErr error

		if buildFormat == "yaml" {
			outPath = filepath.Join(distRoot, stem+".yml")
			writeErr = spec.WriteBuiltContainer(outPath, c, ownStates, imports)
		} else {
			relDir := filepath.Dir(stem)
			if filepath.Base(stem) == "index" {
				outPath = filepath.Join(distRoot, relDir, "_index.md")
				sectionDirs[filepath.Join(distRoot, relDir)] = true
			} else {
				outPath = filepath.Join(distRoot, stem+".md")
				if relDir != "." {
					absParent := filepath.Join(distRoot, relDir)
					if _, ok := leafDirs[absParent]; !ok {
						leafDirs[absParent] = filepath.Base(relDir)
					}
				}
			}
			writeErr = spec.WriteBuiltContainerMarkdown(outPath, c, ownStates, imports)
		}

		if writeErr != nil {
			return fmt.Errorf("\U0001F7E5  %s: %w", c.Path, writeErr)
		}
		relOut, _ := filepath.Rel(distRoot, outPath)
		fmt.Printf("\U0001F7E2  dist/%s\n", filepath.ToSlash(relOut))
	}

	// Emit _index.md stubs for dirs that have leaf files but no index.yml.
	if buildFormat == "markdown" {
		for dir, name := range leafDirs {
			if sectionDirs[dir] {
				continue
			}
			stubPath := filepath.Join(dir, "_index.md")
			if err := spec.WriteHugoSectionStub(stubPath, dirTitle(name)); err != nil {
				return fmt.Errorf("\U0001F7E5  writing section stub for %s: %w", name, err)
			}
			fmt.Printf("\U0001F7E2  dist/%s\n", filepath.ToSlash(filepath.Join(name, "_index.md")))
		}
	}

	return nil
}
