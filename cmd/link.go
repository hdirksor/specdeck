package cmd

import (
	"os"

	"github.com/hdickson/specdeck/internal/link"
	"github.com/spf13/cobra"
)

var linkCmd = &cobra.Command{
	Use:   "link [specs-repo-path]",
	Short: "Link a code repository to a specdeck specs repo",
	Long:  `Creates specdeck.yml and installs Claude Code skills in the current directory.`,
	Args:  cobra.MaximumNArgs(1),
	RunE:  runLink,
}

func init() {
	rootCmd.AddCommand(linkCmd)
}

func runLink(cmd *cobra.Command, args []string) error {
	dir, err := os.Getwd()
	if err != nil {
		return err
	}

	specsPath := ""
	if len(args) == 1 {
		specsPath = args[0]
	}

	return link.Link(dir, specsPath)
}
