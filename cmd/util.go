package cmd

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/tabwriter"

	"gopkg.in/yaml.v3"
)

const (
	srcPlaybooks  = "./.resources/installers/playbooks"
	srcManifests  = "./.resources/installers/manifests"
	srcCharts     = "./.resources/installers/charts"
	srcOperations = "./.resources/operations/playbooks"

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

func searchOperations(query string) ([]packageEntry, error) {
	names, err := listDir(srcOperations, query)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", srcOperations, err)
	}
	var results []packageEntry
	for _, n := range names {
		pattern := filepath.Join(srcOperations, n+".*")
		matches, _ := filepath.Glob(pattern)
		full := ""
		if len(matches) > 0 {
			full = matches[0]
		}
		results = append(results, packageEntry{name: n, pkgType: "operation", fullPath: full})
	}
	return results, nil
}
