package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var installableCmd = &cobra.Command{
	Use:   "installable",
	Short: "Manage installable packages",
}

var installableInstallCmd = &cobra.Command{
	Use:   "install <package>",
	Short: "Install a package into the current project",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		pkg := args[0]

		match, err := findPackageFile(pkg)
		if err != nil {
			return err
		}
		if match == nil {
			return fmt.Errorf("no package found for '%s'", pkg)
		}

		filename := filepath.Base(match.fullPath)

		noApply, _ := cmd.Flags().GetBool("no-apply")

		switch match.pkgType {
		case "playbook":
			dst := filepath.Join(dstPlaybooks, filename)
			fmt.Printf("Installing playbook: %s\n", filename)
			if err := copyFile(match.fullPath, dst); err != nil {
				return err
			}
			if err := appendDeployAll(filename); err != nil {
				return fmt.Errorf("updating deploy_all.yaml: %w", err)
			}
			fmt.Println("Updated ./playbooks/deploy_all.yaml")
			if noApply {
				return nil
			}
			fmt.Println("Starting installation...")
			return runApply(filename)

		case "manifest":
			dst := filepath.Join(dstManifests, filename)
			fmt.Printf("Installing manifest: %s\n", filename)
			if err := os.MkdirAll(dstManifests, 0755); err != nil {
				return err
			}
			if err := copyFile(match.fullPath, dst); err != nil {
				return err
			}
			fmt.Printf("Copied to %s\n", dstManifests)

		case "chart":
			dst := filepath.Join(dstCharts, filename)
			fmt.Printf("Installing chart: %s\n", filename)
			if err := os.MkdirAll(dstCharts, 0755); err != nil {
				return err
			}
			if err := copyFile(match.fullPath, dst); err != nil {
				return err
			}
			fmt.Printf("Copied to %s\n", dstCharts)
		}

		return nil
	},
}

func init() {
	installableInstallCmd.Flags().Bool("no-apply", false, "copy files without running apply.sh")
	installableCmd.AddCommand(installableInstallCmd)
}
