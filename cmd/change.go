package cmd

import (
	"fmt"
	"os"

	"github.com/hdickson/specdeck/internal/scaffold"
	"github.com/spf13/cobra"
)

var changeCmd = &cobra.Command{
	Use:   "change",
	Short: "Manage change records",
}

var changeNewCmd = &cobra.Command{
	Use:   "new <title>",
	Short: "Create a new change record in changes/",
	Args:  cobra.ExactArgs(1),
	RunE:  runChangeNew,
}

func init() {
	changeCmd.AddCommand(changeNewCmd)
	rootCmd.AddCommand(changeCmd)
}

func runChangeNew(cmd *cobra.Command, args []string) error {
	dir, err := os.Getwd()
	if err != nil {
		return err
	}

	path, err := scaffold.NewChange(dir, args[0])
	if err != nil {
		return err
	}

	fmt.Println(path)
	return nil
}
