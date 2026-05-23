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

	"gopkg.in/yaml.v3"
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

// Prepare writes YAML into siteDir/data/containers/ and creates minimal content stubs.
func Prepare(distRoot, siteDir string, skip ...string) error {
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
		base := strings.TrimSuffix(rel, ".yml")
		dataKey := filepath.ToSlash(base)

		raw, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("reading %s: %w", path, err)
		}

		dataPath := filepath.Join(siteDir, "data", "containers", rel)
		if err := os.MkdirAll(filepath.Dir(dataPath), 0755); err != nil {
			return err
		}
		if err := os.WriteFile(dataPath, raw, 0644); err != nil {
			return err
		}

		var doc struct {
			Title string `yaml:"title"`
		}
		yaml.Unmarshal(raw, &doc) //nolint:errcheck

		stubPath := filepath.Join(siteDir, "content", "containers", base+".md")
		if err := os.MkdirAll(filepath.Dir(stubPath), 0755); err != nil {
			return err
		}
		stub := fmt.Sprintf("---\ntitle: %s\ndata_path: %s\n---\n", doc.Title, dataKey)
		return os.WriteFile(stubPath, []byte(stub), 0644)
	})
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
	if _, err := os.Stat(filepath.Join(siteDir, "go.mod")); os.IsNotExist(err) {
		cmd := exec.Command("hugo", "mod", "init", "specdeck-site")
		cmd.Dir = siteDir
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("hugo mod init: %w", err)
		}
	}
	cmd := exec.Command("hugo", "--source", siteDir, "--destination", siteOut)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("hugo: %w", err)
	}

	return nil
}
