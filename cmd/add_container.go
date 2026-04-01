package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/hdickson/specdeck/internal/scaffold"
	"github.com/spf13/cobra"
)

var addContainerCmd = &cobra.Command{
	Use:   "container <path>",
	Short: "Scaffold a new leaf container at the given path",
	Long: `Creates a new container file at the given path within containers/, creating
intermediate directories as needed. All existing states are stubbed in with empty specs.

Example: specdeck add container app/home-tab/feed-screen/post-card`,
	Args: cobra.ExactArgs(1),
	RunE: runAddContainer,
}

var skipConfirm bool

func init() {
	addCmd.AddCommand(addContainerCmd)
	addContainerCmd.Flags().BoolVarP(&skipConfirm, "yes", "y", false, "Skip confirmation for new intermediate directories")
}

func runAddContainer(cmd *cobra.Command, args []string) error {
	projectRoot, err := os.Getwd()
	if err != nil {
		return err
	}

	containerPath := args[0]

	intermediates := scaffold.IntermediateDirectories(projectRoot, containerPath)
	if len(intermediates) > 0 && !skipConfirm {
		fmt.Println("The following directories will be created:")
		for _, d := range intermediates {
			fmt.Printf("  %s\n", d)
		}
		fmt.Print("Continue? [y/N] ")
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		answer := strings.TrimSpace(strings.ToLower(scanner.Text()))
		if answer != "y" && answer != "yes" {
			fmt.Println("Aborted.")
			return nil
		}
	}

	if err := scaffold.AddContainer(projectRoot, containerPath); err != nil {
		return err
	}

	fmt.Printf("Created containers/%s.yml\n", containerPath)
	return nil
}
