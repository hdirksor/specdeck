package cmd

import (
	"fmt"
	"os"

	"github.com/hdirksor/specdeck/internal/config"
	"github.com/hdirksor/specdeck/internal/link"
	"github.com/spf13/cobra"
)

var linkCmd = &cobra.Command{
	Use:   "link <specs-repo-path>",
	Short: "Link a code repository to a specdeck specs repo",
	Long:  `Records the specs repo path in specdeck.yml in the current directory, creating the file if needed. Re-run to point at a different specs repo.`,
	Args:  cobra.ExactArgs(1),
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

	if err := link.Link(dir, args[0]); err != nil {
		return err
	}
	fmt.Fprintf(cmd.OutOrStdout(), "Linked to %s.\n", args[0])

	cfg, err := config.Load(dir)
	if err != nil {
		return err
	}
	if len(cfg.Agents) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "Using an AI coding agent? Run `specdeck agents claude` to install specdeck's commands.")
	}
	return nil
}
