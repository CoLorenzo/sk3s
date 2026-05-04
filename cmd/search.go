package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var installableListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all available packages",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runSearch("")
	},
}

var installableSearchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search available packages by name",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runSearch(args[0])
	},
}

func runSearch(query string) error {
	entries, err := searchPackages(query)
	if err != nil {
		return err
	}

	if len(entries) == 0 {
		if query != "" {
			fmt.Printf("No matches found for: %s\n", query)
		} else {
			fmt.Println("No packages available.")
		}
		return nil
	}

	printTable(entries)
	return nil
}

func init() {
	installableCmd.AddCommand(installableListCmd)
	installableCmd.AddCommand(installableSearchCmd)
}
