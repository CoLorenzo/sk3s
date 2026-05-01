package cmd

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var installCmd = &cobra.Command{
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

// findPackageFile searches all source dirs for the best fuzzy match.
func findPackageFile(pkg string) (*packageEntry, error) {
	entries, err := searchPackages(pkg)
	if err != nil {
		return nil, err
	}
	if len(entries) == 0 {
		return nil, nil
	}
	return &entries[0], nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}

func appendDeployAll(filename string) error {
	var entries []map[string]string

	data, err := os.ReadFile(deployAll)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	if len(data) > 0 {
		if err := yaml.Unmarshal(data, &entries); err != nil {
			return fmt.Errorf("parsing deploy_all.yaml: %w", err)
		}
	}

	entries = append(entries, map[string]string{"import_playbook": filename})

	out, err := yaml.Marshal(entries)
	if err != nil {
		return err
	}
	return os.WriteFile(deployAll, out, 0644)
}

func runApply(filename string) error {
	c := exec.Command("./apply.sh", filename)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	return c.Run()
}
