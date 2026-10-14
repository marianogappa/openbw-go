package main

import (
	"flag"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/icza/screp/repparser/repdecoder"
	"github.com/marianogappa/openbw-go/downgrade"
)

type config struct {
	harness   engine
	obwhash   engine
	outDir    string
	maxFrames int
	limit     int
	maxLines  int
	noActions bool
}

func main() {
	cfg, replays, err := parseConfig()
	if err != nil {
		fmt.Fprintln(os.Stderr, "difftest:", err)
		os.Exit(2)
	}
	differing := 0
	for _, replay := range replays {
		if !compareReplay(cfg, replay) {
			differing++
		}
	}
	fmt.Printf("summary: %d replays, %d differ\n", len(replays), differing)
	if differing > 0 {
		os.Exit(1)
	}
}

func parseConfig() (config, []string, error) {
	var cfg config
	harnessPath := flag.String("harness", "tools/harness/harness", "C++ harness binary")
	obwhashPath := flag.String("obwhash", "", "obwhash binary (default: build ./cmd/obwhash)")
	data := flag.String("data", "", "data directory passed to both engines")
	flag.StringVar(&cfg.outDir, "out", filepath.Join(os.TempDir(), "openbw-difftest"), "directory for downgraded replays")
	flag.IntVar(&cfg.maxFrames, "max-frames", -1, "stop both engines at this frame")
	flag.IntVar(&cfg.limit, "n", 0, "compare at most this many replays (0: all)")
	flag.IntVar(&cfg.maxLines, "max-lines", 100, "print at most this many differences per frame (0: all)")
	flag.BoolVar(&cfg.noActions, "no-actions", false, "run both engines without replay actions")
	listPath := flag.String("list", "", "file with one replay path per line ('#' starts a comment line)")
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "usage: difftest [flags] [-list replays.txt] [replay.rep|dir ...]")
		flag.PrintDefaults()
	}
	flag.Parse()
	if flag.NArg() == 0 && *listPath == "" {
		flag.Usage()
		os.Exit(2)
	}
	if err := os.MkdirAll(cfg.outDir, 0o755); err != nil {
		return cfg, nil, err
	}
	if *obwhashPath == "" {
		built, err := buildObwhash(cfg.outDir)
		if err != nil {
			return cfg, nil, err
		}
		*obwhashPath = built
	}
	cfg.harness = engine{name: "harness", path: *harnessPath, data: *data}
	cfg.obwhash = engine{name: "obwhash", path: *obwhashPath, data: *data}
	args := flag.Args()
	if *listPath != "" {
		listed, err := readReplayList(*listPath)
		if err != nil {
			return cfg, nil, err
		}
		args = append(listed, args...)
	}
	replays, err := collectReplays(args, cfg.limit)
	return cfg, replays, err
}

func readReplayList(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var replays []string
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "#") {
			replays = append(replays, line)
		}
	}
	return replays, nil
}

func buildObwhash(outDir string) (string, error) {
	path := filepath.Join(outDir, "obwhash")
	cmd := exec.Command("go", "build", "-o", path, "github.com/marianogappa/openbw-go/cmd/obwhash")
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("build obwhash: %w", err)
	}
	return path, nil
}

func collectReplays(args []string, limit int) ([]string, error) {
	var replays []string
	for _, arg := range args {
		err := filepath.WalkDir(arg, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if !d.IsDir() && strings.EqualFold(filepath.Ext(path), ".rep") {
				replays = append(replays, path)
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	if limit > 0 && len(replays) > limit {
		replays = replays[:limit]
	}
	return replays, nil
}

func isLegacyReplay(data []byte) bool {
	dec := repdecoder.New(data)
	defer dec.Close()
	return dec.RepFormat() == repdecoder.RepFormatLegacy
}

func legacyReplayPath(cfg config, replay string) (string, error) {
	data, err := os.ReadFile(replay)
	if err != nil {
		return "", err
	}
	if isLegacyReplay(data) {
		return replay, nil
	}
	legacy, report, err := downgrade.Downgrade(data)
	if err != nil {
		return "", fmt.Errorf("downgrade: %w", err)
	}
	if dropped := report.DroppedCount(); dropped > 0 {
		fmt.Printf("  downgrade dropped %d commands\n", dropped)
	}
	path := filepath.Join(cfg.outDir, strings.TrimSuffix(filepath.Base(replay), filepath.Ext(replay))+".116.rep")
	return path, os.WriteFile(path, legacy, 0o644)
}

func compareReplay(cfg config, replay string) bool {
	fmt.Printf("== %s\n", replay)
	legacy, err := legacyReplayPath(cfg, replay)
	if err != nil {
		fmt.Printf("  skipped: %v\n", err)
		return false
	}
	a := cfg.harness.run(legacy, cfg.engineArgs(frameArgs(cfg.maxFrames))...)
	b := cfg.obwhash.run(legacy, cfg.engineArgs(frameArgs(cfg.maxFrames))...)
	frame, differs := firstDifferentFrame(a, b)
	if !differs {
		return reportMatch(cfg, a, b)
	}
	reportDifference(cfg, legacy, frame, a, b)
	return false
}

func reportMatch(cfg config, a, b *runOutput) bool {
	if a.lastFrame < 0 {
		if a.errorLine != "" && a.errorLine == b.errorLine {
			fmt.Printf("  match: both engines fail with %q\n", a.errorLine)
			return true
		}
		fmt.Printf("  %s produced no frames: %s\n", cfg.harness.name, a.describeFailure())
		return false
	}
	lines := playerDiffLines(cfg.harness.name, cfg.obwhash.name, a, b)
	if a.endLine != b.endLine || a.errorLine != b.errorLine || len(lines) > 0 {
		fmt.Printf("  frame hashes match to frame %d, but the end differs\n", a.lastFrame)
		fmt.Printf("    %s: %s %s\n", cfg.harness.name, a.endLine, a.errorLine)
		fmt.Printf("    %s: %s %s\n", cfg.obwhash.name, b.endLine, b.errorLine)
		for _, line := range lines {
			fmt.Printf("    %s\n", strings.TrimSpace(line))
		}
		return false
	}
	fmt.Printf("  match: %d frames, %s\n", a.lastFrame, a.endLine)
	return true
}

func reportDifference(cfg config, legacy string, frame int, a, b *runOutput) {
	fmt.Printf("  first different frame: %d\n", frame)
	fmt.Printf("    %s: %s\n", cfg.harness.name, hashOrFailure(a, frame))
	fmt.Printf("    %s: %s\n", cfg.obwhash.name, hashOrFailure(b, frame))
	if frame > 0 {
		printSnapshotAt(cfg, legacy, frame-1)
	}
	printSnapshotAt(cfg, legacy, frame)
}

func (cfg config) engineArgs(args []string) []string {
	if cfg.noActions {
		return append([]string{"-no-actions"}, args...)
	}
	return args
}

func hashOrFailure(out *runOutput, frame int) string {
	if hash, ok := out.frameHashes[frame]; ok {
		return hash
	}
	return "no hash (" + out.describeFailure() + ")"
}

func printSnapshotAt(cfg config, legacy string, frame int) {
	args := []string{"-every", "1000000000", "-max-frames", fmt.Sprint(frame), "-dump-units", fmt.Sprint(frame)}
	a := cfg.harness.run(legacy, cfg.engineArgs(args)...)
	b := cfg.obwhash.run(legacy, cfg.engineArgs(args)...)
	printSnapshotDiff(os.Stdout, frame, cfg.harness.name, cfg.obwhash.name, a, b, cfg.maxLines)
}
