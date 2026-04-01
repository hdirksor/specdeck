package cmd

import (
	"os"
	"path/filepath"

	"github.com/hdickson/specdeck/internal/scaffold"
	"github.com/spf13/cobra"
)

var newCmd = &cobra.Command{
	Use:   "new [name]",
	Short: "Initialise a new specdeck project in the current directory",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runNew,
}

func init() {
	rootCmd.AddCommand(newCmd)
}

func runNew(cmd *cobra.Command, args []string) error {
	dir, err := os.Getwd()
	if err != nil {
		return err
	}

	name := filepath.Base(dir)
	if len(args) == 1 {
		name = args[0]
	}

	return scaffold.New(dir, name)
}
