package cmd

import (
	"fmt"
	"os"

	"github.com/hdickson/specdeck/internal/scaffold"
	"github.com/spf13/cobra"
)

var fillCmd = &cobra.Command{
	Use:   "fill [path]",
	Short: "Copy default specs into empty states",
	Long: `Copies the default state's specs into any states with empty specs.
States that already have specs are left untouched.

With a path, fills a single container:
  specdeck fill app/home-tab/feed-screen/post-card

Without a path, fills all leaf containers in the project:
  specdeck fill`,
	Args: cobra.MaximumNArgs(1),
	RunE: runFill,
}

func init() {
	rootCmd.AddCommand(fillCmd)
}

func runFill(cmd *cobra.Command, args []string) error {
	projectRoot, err := os.Getwd()
	if err != nil {
		return err
	}

	if len(args) == 1 {
		if err := scaffold.Fill(projectRoot, args[0]); err != nil {
			return err
		}
		fmt.Printf("Filled empty states in containers/%s from default.\n", args[0])
		return nil
	}

	filled, skipped, err := scaffold.FillAll(projectRoot)
	if err != nil {
		return err
	}
	fmt.Printf("Filled %d containers.", filled)
	if skipped > 0 {
		fmt.Printf(" Skipped %d (no default state defined).", skipped)
	}
	fmt.Println()
	return nil
}
