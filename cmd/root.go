package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "sk3s",
	Short: "sk3s – k3s cluster management via Ansible",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(installedCmd)
	rootCmd.AddCommand(installLsCmd)
	rootCmd.AddCommand(installSearchCmd)
	rootCmd.AddCommand(installCmd)
	rootCmd.AddCommand(uninstallCmd)
	rootCmd.AddCommand(nodesCmd)
	rootCmd.AddCommand(nodeAddCmd)
	rootCmd.AddCommand(nodeRmCmd)
}
