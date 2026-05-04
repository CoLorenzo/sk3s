package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var operationCmd = &cobra.Command{
	Use:   "operation",
	Short: "Manage cluster operations",
}

var operationListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all available operations",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runOperationSearch("")
	},
}

var operationSearchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search available operations by name",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runOperationSearch(args[0])
	},
}

var operationActiveCmd = &cobra.Command{
	Use:   "active",
	Short: "Show active operations from deploy_all.yaml",
	RunE: func(cmd *cobra.Command, args []string) error {
		data, err := os.ReadFile(deployAll)
		if err != nil {
			if os.IsNotExist(err) {
				fmt.Println("No active operations (deploy_all.yaml not found).")
				return nil
			}
			return err
		}
		fmt.Print(string(data))
		return nil
	},
}

var operationInstallCmd = &cobra.Command{
	Use:   "install <operation>",
	Short: "Install an operation into the current project",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		pkg := args[0]

		entries, err := searchOperations(pkg)
		if err != nil {
			return err
		}
		if len(entries) == 0 {
			return fmt.Errorf("no operation found for '%s'", pkg)
		}

		match := entries[0]
		filename := match.name + ".yaml"
		dst := filepath.Join(dstPlaybooks, filename)

		fmt.Printf("Installing operation: %s\n", filename)
		if err := copyFile(match.fullPath, dst); err != nil {
			return err
		}
		if err := appendDeployAll(filename); err != nil {
			return fmt.Errorf("updating deploy_all.yaml: %w", err)
		}
		fmt.Println("Updated ./playbooks/deploy_all.yaml")

		noApply, _ := cmd.Flags().GetBool("no-apply")
		if noApply {
			return nil
		}
		fmt.Println("Starting operation...")
		return runApply(filename)
	},
}

func runOperationSearch(query string) error {
	entries, err := searchOperations(query)
	if err != nil {
		return err
	}

	if len(entries) == 0 {
		if query != "" {
			fmt.Printf("No operations found for: %s\n", query)
		} else {
			fmt.Println("No operations available.")
		}
		return nil
	}

	printTable(entries)
	return nil
}

func init() {
	operationInstallCmd.Flags().Bool("no-apply", false, "copy files without running apply.sh")
	operationCmd.AddCommand(operationListCmd)
	operationCmd.AddCommand(operationSearchCmd)
	operationCmd.AddCommand(operationActiveCmd)
	operationCmd.AddCommand(operationInstallCmd)
}
