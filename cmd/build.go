package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hdickson/specdeck/internal/build"
	"github.com/spf13/cobra"
)

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Resolve container refs and write specs to dist/",
	Args:  cobra.NoArgs,
	RunE:  runBuild,
}

func init() {
	rootCmd.AddCommand(buildCmd)
}

func runBuild(cmd *cobra.Command, args []string) error {
	root, err := os.Getwd()
	if err != nil {
		return err
	}

	containersRoot := filepath.Join(root, "containers")
	containers, err := build.Load(containersRoot)
	if err != nil {
		return fmt.Errorf("\U0001F7E5  %w", err)
	}

	if err := build.Validate(containers); err != nil {
		return fmt.Errorf("\U0001F7E5  %w", err)
	}

	distRoot := filepath.Join(root, "dist")
	sectionDirs := make(map[string]bool)
	leafDirs := make(map[string]string)

	for _, c := range containers {
		stem := strings.TrimSuffix(c.Path, filepath.Ext(c.Path))
		relDir := filepath.Dir(stem)

		var outPath string
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

		if err := build.Write(c, outPath); err != nil {
			return fmt.Errorf("\U0001F7E5  %s: %w", c.Path, err)
		}
		relOut, _ := filepath.Rel(distRoot, outPath)
		fmt.Printf("\033[32m↳\033[0m  dist/%s\n", filepath.ToSlash(relOut))
	}

	for dir, name := range leafDirs {
		if sectionDirs[dir] {
			continue
		}
		stubPath := filepath.Join(dir, "_index.md")
		if err := writeHugoSectionStub(stubPath, dirTitle(name)); err != nil {
			return fmt.Errorf("\U0001F7E5  writing section stub for %s: %w", name, err)
		}
		fmt.Printf("\033[32m↳\033[0m  dist/%s\n", filepath.ToSlash(filepath.Join(name, "_index.md")))
	}

	return nil
}

func writeHugoSectionStub(path, title string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("creating directory: %w", err)
	}
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("creating file: %w", err)
	}
	defer f.Close()
	fmt.Fprintf(f, "---\ntitle: %s\n---\n", title)
	return nil
}

func dirTitle(name string) string {
	words := strings.FieldsFunc(name, func(r rune) bool { return r == '-' || r == '_' })
	for i, w := range words {
		if len(w) > 0 {
			words[i] = strings.ToUpper(w[:1]) + w[1:]
		}
	}
	return strings.Join(words, " ")
}
