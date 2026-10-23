// Package link connects a code repo to its specs repo.
package link

import "github.com/hdirksor/specdeck/internal/config"

// Link records specsPath as the specs repo in dir's specdeck.yml, creating
// the file if needed and keeping the rest of its settings.
func Link(dir, specsPath string) error {
	cfg, err := config.Load(dir)
	if err != nil {
		return err
	}
	cfg.SpecsRepo = specsPath
	return config.Save(dir, cfg)
}
