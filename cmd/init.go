package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init [project-name]",
	Short: "Initialize a new sk3s project by cloning the Ansible template",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		projectName := "sk3s-ansible"
		if len(args) > 0 {
			projectName = args[0]
		}
		fmt.Printf("Initializing project: %s\n", projectName)

		c := exec.Command("git", "clone", "https://github.com/CoLorenzo/Sk3s-ansible", projectName)
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr

		if err := c.Run(); err != nil {
			return fmt.Errorf("clone failed: %w", err)
		}

		fmt.Printf("\nProject '%s' ready.\n", projectName)
		return nil
	},
}
