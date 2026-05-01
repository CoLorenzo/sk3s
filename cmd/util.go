package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"
)

const (
	srcPlaybooks = "./.resources/installers/playbooks"
	srcManifests = "./.resources/installers/manifests"
	srcCharts    = "./.resources/installers/charts"

	dstPlaybooks = "./playbooks"
	dstManifests = "./playbooks/files/manifests"
	dstCharts    = "./playbooks/files/charts"

	deployAll = "./playbooks/deploy_all.yaml"
)

type packageEntry struct {
	name     string
	pkgType  string
	fullPath string
}

// fuzzyMatch reports whether all runes of query appear in s in order (case-insensitive).
func fuzzyMatch(s, query string) bool {
	if query == "" {
		return true
	}
	s = strings.ToLower(s)
	q := strings.ToLower(query)
	si := 0
	for _, qr := range q {
		found := false
		for si < len(s) {
			if rune(s[si]) == qr {
				si++
				found = true
				break
			}
			si++
		}
		if !found {
			return false
		}
	}
	return true
}

// listDir returns filenames (without extension) in dir, filtering by fuzzy query.
func listDir(dir, query string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		stem := strings.TrimSuffix(e.Name(), filepath.Ext(e.Name()))
		if fuzzyMatch(stem, query) {
			names = append(names, stem)
		}
	}
	return names, nil
}

// searchPackages returns all packages matching query across the three source dirs.
func searchPackages(query string) ([]packageEntry, error) {
	var results []packageEntry

	type dirType struct {
		dir     string
		pkgType string
	}
	dirs := []dirType{
		{srcPlaybooks, "playbook"},
		{srcManifests, "manifest"},
		{srcCharts, "chart"},
	}

	for _, dt := range dirs {
		names, err := listDir(dt.dir, query)
		if err != nil {
			return nil, fmt.Errorf("reading %s: %w", dt.dir, err)
		}
		for _, n := range names {
			pattern := filepath.Join(dt.dir, n+".*")
			matches, _ := filepath.Glob(pattern)
			full := ""
			if len(matches) > 0 {
				full = matches[0]
			}
			results = append(results, packageEntry{name: n, pkgType: dt.pkgType, fullPath: full})
		}
	}
	return results, nil
}

// printTable prints a NAME/TYPE table to stdout.
func printTable(entries []packageEntry) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', tabwriter.TabIndent)
	fmt.Fprintln(w, "NAME\tTYPE")
	fmt.Fprintln(w, "----\t----")
	for _, e := range entries {
		fmt.Fprintf(w, "%s\t%s\n", e.name, e.pkgType)
	}
	w.Flush()
}
