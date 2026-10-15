package main

import (
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	openbw "github.com/marianogappa/openbw-go"
	"github.com/marianogappa/openbw-go/downgrade"
)

type filter struct {
	mapName   string
	race      string
	minFrames int
	count     int
	skip      int
}

type replay struct {
	path    string
	mapName string
	matchup string
	frames  int
}

func main() {
	replays := flag.String("replays", "", "directory with SC:R replays (searched recursively)")
	out := flag.String("out", "", "directory for the downgraded .116.rep files")
	dataDir := flag.String("data", "", "game data directory (default $OPENBW_DATA)")
	var f filter
	flag.StringVar(&f.mapName, "map", "", "only maps whose name contains this text")
	flag.StringVar(&f.race, "race", "", "only matchups that contain this race letter (T, Z or P)")
	flag.IntVar(&f.minFrames, "min-frames", 7145, "only games with at least this many frames")
	flag.IntVar(&f.count, "n", 60, "number of replays to collect")
	flag.IntVar(&f.skip, "skip", 0, "matching replays to skip first, in path order")
	flag.Parse()
	if *replays == "" || *out == "" {
		flag.Usage()
		os.Exit(2)
	}
	data, err := openbw.LoadData(*dataDir)
	if err != nil {
		fmt.Fprintln(os.Stderr, "collect:", err)
		os.Exit(1)
	}
	if err := run(data, *replays, *out, f); err != nil {
		fmt.Fprintln(os.Stderr, "collect:", err)
		os.Exit(1)
	}
}

func run(data *openbw.Data, replayDir, outDir string, f filter) error {
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return err
	}
	paths, err := replayPaths(replayDir)
	if err != nil {
		return err
	}
	skipped, collected := 0, 0
	for _, path := range paths {
		if collected == f.count {
			break
		}
		r, ok := readReplay(data, path)
		if !ok || !f.matches(r) {
			continue
		}
		if skipped < f.skip {
			skipped++
			continue
		}
		if err := downgradeTo(path, outDir); err != nil {
			fmt.Fprintf(os.Stderr, "%s: %v\n", path, err)
			continue
		}
		collected++
		fmt.Printf("%s\t%s\t%s\t%d\n", strings.TrimSuffix(filepath.Base(path), ".rep"), r.mapName, r.matchup, r.frames)
	}
	return nil
}

func replayPaths(dir string) ([]string, error) {
	var paths []string
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || filepath.Ext(path) != ".rep" {
			return err
		}
		paths = append(paths, path)
		return nil
	})
	sort.Strings(paths)
	return paths, err
}

var raceLetters = map[uint8]string{0: "Z", 1: "T", 2: "P"}

func readReplay(data *openbw.Data, path string) (replay, bool) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return replay{}, false
	}
	g, err := openbw.Open(raw, openbw.Options{Data: data, Kinds: openbw.NoKinds})
	if err != nil {
		return replay{}, false
	}
	slots := g.PlayingSlots()
	if len(slots) != 2 {
		return replay{}, false
	}
	h := g.Header()
	var races []string
	for _, slot := range slots {
		races = append(races, raceLetters[h.Slots[slot].Race])
	}
	return replay{path: path, mapName: plainMapName(h.MapName), matchup: strings.Join(races, "v"), frames: int(h.EndFrame)}, true
}

func plainMapName(name string) string {
	return strings.TrimSpace(strings.Map(func(r rune) rune {
		if r < ' ' {
			return -1
		}
		return r
	}, name))
}

func (f filter) matches(r replay) bool {
	if r.frames < f.minFrames {
		return false
	}
	if f.mapName != "" && !strings.Contains(r.mapName, f.mapName) {
		return false
	}
	return f.race == "" || strings.Contains(r.matchup, f.race)
}

func downgradeTo(path, outDir string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	legacy, _, err := downgrade.Downgrade(data)
	if err != nil {
		return err
	}
	name := strings.TrimSuffix(filepath.Base(path), ".rep") + ".116.rep"
	return os.WriteFile(filepath.Join(outDir, name), legacy, 0o644)
}
