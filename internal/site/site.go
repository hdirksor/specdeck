package site

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
)

//go:embed all:scaffold
var scaffoldFS embed.FS

// Bootstrap copies the embedded Hugo scaffold into siteDir if no hugo.toml exists.
func Bootstrap(siteDir string) error {
	if _, err := os.Stat(filepath.Join(siteDir, "hugo.toml")); err == nil {
		return nil
	}
	return fs.WalkDir(scaffoldFS, "scaffold", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel("scaffold", path)
		dst := filepath.Join(siteDir, rel)
		if d.IsDir() {
			return os.MkdirAll(dst, 0755)
		}
		data, err := scaffoldFS.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(dst, data, 0644)
	})
}

// Prepare writes content stubs into siteDir/content/containers/ from YAML in distRoot.
func Prepare(distRoot, siteDir string, skip ...string) error {
	return writeContentStubs(distRoot, siteDir, skip...)
}

// Build generates a static site at siteOut using the Hugo project in siteDir.
// Requires hugo to be installed and on PATH.
func Build(distRoot, siteDir, siteOut string) error {
	if err := Bootstrap(siteDir); err != nil {
		return err
	}
	if err := Prepare(distRoot, siteDir, siteOut); err != nil {
		return err
	}

	cmd := exec.Command("hugo", "--source", siteDir, "--destination", siteOut)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("hugo: %w", err)
	}
	return nil
}

func writeContentStubs(distRoot, siteDir string, skip ...string) error {
	return filepath.WalkDir(distRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if slices.Contains(skip, path) {
				return filepath.SkipDir
			}
			return nil
		}
		if filepath.Ext(d.Name()) != ".yml" {
			return nil
		}

		rel, err := filepath.Rel(distRoot, path)
		if err != nil {
			return err
		}

		rel = strings.TrimSuffix(rel, ".yml")
		stubPath := filepath.Join(siteDir, "content", "containers", rel+".md")

		return writeContentStub(path, stubPath)
	})
}

func writeContentStub(yamlPath, stubPath string) error {
	yamlData, err := os.ReadFile(yamlPath)
	if err != nil {
		return fmt.Errorf("reading %s: %w", yamlPath, err)
	}

	if err := os.MkdirAll(filepath.Dir(stubPath), 0755); err != nil {
		return err
	}

	content := "---\n" + string(yamlData) + "---\n"
	return os.WriteFile(stubPath, []byte(content), 0644)
}
