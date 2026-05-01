package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var uninstallCmd = &cobra.Command{
	Use:   "uninstall <package>",
	Short: "Remove an installed package from the current project",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		pkg := args[0]

		type installedDir struct {
			dir     string
			pkgType string
		}
		dirs := []installedDir{
			{dstPlaybooks, "playbook"},
			{dstManifests, "manifest"},
			{dstCharts, "chart"},
		}

		for _, d := range dirs {
			pattern := filepath.Join(d.dir, pkg+"*")
			matches, err := filepath.Glob(pattern)
			if err != nil {
				return err
			}
			for _, m := range matches {
				filename := filepath.Base(m)
				stem := strings.TrimSuffix(filename, filepath.Ext(filename))

				// skip non-package files in playbooks dir
				if filename == "deploy_all.yaml" {
					continue
				}
				if stem != pkg {
					continue
				}

				fmt.Printf("Uninstalling %s: %s\n", d.pkgType, filename)

				if d.pkgType == "playbook" {
					if err := removeFromDeployAll(filename); err != nil {
						return fmt.Errorf("updating deploy_all.yaml: %w", err)
					}
					fmt.Println("Updated ./playbooks/deploy_all.yaml")
				}

				if err := os.Remove(m); err != nil {
					return fmt.Errorf("removing %s: %w", m, err)
				}
				fmt.Printf("Removed %s\n", m)
				return nil
			}
		}

		return fmt.Errorf("package '%s' is not installed", pkg)
	},
}

func removeFromDeployAll(filename string) error {
	data, err := os.ReadFile(deployAll)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	var entries []map[string]string
	if err := yaml.Unmarshal(data, &entries); err != nil {
		return err
	}

	filtered := entries[:0]
	for _, e := range entries {
		if e["import_playbook"] != filename {
			filtered = append(filtered, e)
		}
	}

	out, err := yaml.Marshal(filtered)
	if err != nil {
		return err
	}
	return os.WriteFile(deployAll, out, 0644)
}
