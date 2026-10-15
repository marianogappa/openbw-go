package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

type config struct {
	replayDir   string
	outDir      string
	harnessPath string
	probePath   string
	dataDir     string
	count       int
	longCount   int
	workers     int
	from        string
	to          string
	race        string
}

func main() {
	cfg := parseFlags()
	if err := run(cfg); err != nil {
		fmt.Fprintln(os.Stderr, "integration:", err)
		os.Exit(1)
	}
}

func parseFlags() config {
	var cfg config
	flag.StringVar(&cfg.replayDir, "replays", "", "directory with SC:R replays (searched recursively)")
	flag.StringVar(&cfg.outDir, "out", "", "directory for downgraded replays, raw outputs and results.json")
	flag.StringVar(&cfg.harnessPath, "harness", "tools/harness/harness", "path to the tools/harness binary")
	flag.StringVar(&cfg.probePath, "probe", "tools/screp/integration/probe/probe", "path to the tools/screp/integration probe binary")
	flag.StringVar(&cfg.dataDir, "data", defaultDataDir(), "game data directory")
	flag.IntVar(&cfg.count, "n", 30, "number of replays to pick")
	flag.IntVar(&cfg.longCount, "long", 4, "number of the longest games to include")
	flag.IntVar(&cfg.workers, "workers", 8, "replays to run at the same time")
	flag.StringVar(&cfg.from, "from", "", "only replays started on or after this date (YYYY-MM-DD)")
	flag.StringVar(&cfg.to, "to", "", "only replays started before this date (YYYY-MM-DD)")
	flag.StringVar(&cfg.race, "race", "", "only matchups that contain this race letter (T, Z or P)")
	flag.Parse()
	if cfg.replayDir == "" || cfg.outDir == "" {
		flag.Usage()
		os.Exit(2)
	}
	return cfg
}

func defaultDataDir() string {
	if dir := os.Getenv("OPENBW_DATA"); dir != "" {
		return dir
	}
	return "testdata/bwdata"
}

func run(cfg config) error {
	if err := os.MkdirAll(cfg.outDir, 0o755); err != nil {
		return err
	}
	candidates, err := scanReplays(cfg.replayDir)
	if err != nil {
		return err
	}
	picked := pickReplays(filterCandidates(candidates, cfg), cfg.count, cfg.longCount)
	writeSurvey(os.Stdout, candidates)
	results := runAll(cfg, picked)
	for i := range results {
		classify(&results[i])
	}
	if err := writeJSON(filepath.Join(cfg.outDir, "results.json"), results); err != nil {
		return err
	}
	writeMarkdown(os.Stdout, results)
	return nil
}

func writeJSON(path string, v any) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}
