package main

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"unicode"

	openbw "github.com/marianogappa/openbw-go"
)

type mapReplays struct {
	name  string
	paths []string
}

func findReplays(dir string) ([]string, error) {
	var paths []string
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.EqualFold(filepath.Ext(path), ".rep") {
			paths = append(paths, path)
		}
		return nil
	})
	slices.Sort(paths)
	return paths, err
}

func cleanMapName(name string) string {
	name = strings.Map(func(r rune) rune {
		if unicode.IsControl(r) {
			return -1
		}
		return r
	}, name)
	return strings.ToLower(strings.TrimSpace(name))
}

func mapName(data *openbw.Data, path string) (string, error) {
	replay, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	g, err := openbw.Open(replay, openbw.Options{Data: data, Kinds: openbw.NoKinds})
	if err != nil {
		return "", err
	}
	return cleanMapName(g.Header().MapName), nil
}

func groupByMap(data *openbw.Data, paths []string, log io.Writer) []mapReplays {
	index := map[string]int{}
	var groups []mapReplays
	for _, path := range paths {
		name, err := mapName(data, path)
		if err != nil {
			fmt.Fprintf(log, "skip %s: %v\n", path, err)
			continue
		}
		i, ok := index[name]
		if !ok {
			i = len(groups)
			index[name] = i
			groups = append(groups, mapReplays{name: name})
		}
		groups[i].paths = append(groups[i].paths, path)
	}
	slices.SortFunc(groups, func(a, b mapReplays) int { return strings.Compare(a.name, b.name) })
	return groups
}

func roundRobin(groups []mapReplays, n int) []string {
	var out []string
	for round := 0; len(out) < n; round++ {
		added := false
		for _, g := range groups {
			if round < len(g.paths) && len(out) < n {
				out = append(out, g.paths[round])
				added = true
			}
		}
		if !added {
			break
		}
	}
	return out
}

func spread(paths []string, n int) []string {
	if len(paths) <= n {
		return paths
	}
	out := make([]string, n)
	for i := range n {
		out[i] = paths[i*len(paths)/n]
	}
	return out
}

func selectReplays(data *openbw.Data, scrDir string, scrCount int, otherDir string, otherCount int, out, log io.Writer) error {
	scr, err := findReplays(scrDir)
	if err != nil {
		return err
	}
	groups := groupByMap(data, scr, log)
	for _, g := range groups {
		fmt.Fprintf(log, "map %q: %d replays\n", g.name, len(g.paths))
	}
	fmt.Fprintln(out, "# SC:R 1v1 replays, round robin over maps")
	for _, path := range roundRobin(groups, scrCount) {
		fmt.Fprintln(out, listEntry(path))
	}
	other, err := findReplays(otherDir)
	if err != nil {
		return err
	}
	fmt.Fprintln(out, "# team and older replays, evenly spread")
	for _, path := range spread(other, otherCount) {
		fmt.Fprintln(out, listEntry(path))
	}
	return nil
}

func listEntry(path string) string {
	root, err := filepath.Abs(replaysDir())
	if err != nil {
		return path
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return path
	}
	if rel, err := filepath.Rel(root, abs); err == nil && !strings.HasPrefix(rel, "..") {
		return rel
	}
	return path
}
