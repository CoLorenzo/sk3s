package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var installedCmd = &cobra.Command{
	Use:   "installed",
	Short: "List installed packages in the current project",
	RunE: func(cmd *cobra.Command, args []string) error {
		type dirType struct {
			dir     string
			pkgType string
			exclude []string
		}
		dirs := []dirType{
			{dstPlaybooks, "playbook", []string{"deploy_all.yaml"}},
			{dstManifests, "manifest", nil},
			{dstCharts, "chart", nil},
		}

		var entries []packageEntry
		for _, dt := range dirs {
			files, err := os.ReadDir(dt.dir)
			if err != nil {
				if os.IsNotExist(err) {
					continue
				}
				return err
			}
			for _, f := range files {
				if f.IsDir() {
					continue
				}
				name := f.Name()
				if isExcluded(name, dt.exclude) {
					continue
				}
				stem := strings.TrimSuffix(name, filepath.Ext(name))
				entries = append(entries, packageEntry{
					name:     stem,
					pkgType:  dt.pkgType,
					fullPath: filepath.Join(dt.dir, name),
				})
			}
		}

		if len(entries) == 0 {
			fmt.Println("No packages installed.")
			return nil
		}
		printTable(entries)
		return nil
	},
}

func isExcluded(name string, list []string) bool {
	for _, ex := range list {
		if name == ex {
			return true
		}
	}
	return false
}
