package main

import (
	"bufio"
	"flag"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"sync"
	"time"

	openbw "github.com/marianogappa/openbw-go"
)

func main() {
	mode := flag.String("mode", "validate", "select, validate, cost or trace")
	list := flag.String("list", "", "file with one replay path per line, relative to $OPENBW_REPLAYS ('#' starts a comment line)")
	dataDir := flag.String("data", "testdata/bwdata", "game data directory")
	workers := flag.Int("j", runtime.NumCPU(), "parallel games")
	scrDir := flag.String("scr", "", "select: directory of SC:R 1v1 replays")
	scrCount := flag.Int("scr-n", 170, "select: SC:R replays to take, round robin over maps")
	otherDir := flag.String("other", "", "select: directory of team and older replays")
	otherCount := flag.Int("other-n", 40, "select: other replays to take")
	costGames := flag.Int("cost-n", 20, "cost: replays for the serial measurements")
	traceReplay := flag.String("replay", "", "trace: replay file")
	traceID := flag.Uint("unit", 0, "trace: unit ID")
	flag.Parse()
	plainData.dir = *dataDir

	var err error
	switch *mode {
	case "select":
		var data *openbw.Data
		if data, err = openbw.LoadData(*dataDir); err == nil {
			err = selectReplays(data, *scrDir, *scrCount, *otherDir, *otherCount, os.Stdout, os.Stderr)
		}
	case "trace":
		var data *openbw.Data
		if data, err = openbw.LoadData(*dataDir); err == nil {
			err = traceUnit(data, *traceReplay, openbw.UnitID(*traceID))
		}
	case "validate", "cost":
		var paths []string
		var data *openbw.Data
		if paths, err = readList(*list); err != nil {
			break
		}
		if data, err = openbw.LoadData(*dataDir); err != nil {
			break
		}
		if *mode == "validate" {
			err = validate(data, paths, *workers)
		} else {
			err = measureCost(data, paths, *workers, *costGames)
		}
	default:
		err = fmt.Errorf("unknown mode %q", *mode)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "apivalidate:", err)
		os.Exit(1)
	}
}

func readList(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var paths []string
	s := bufio.NewScanner(f)
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		if line != "" && !strings.HasPrefix(line, "#") {
			paths = append(paths, filepath.Join(replaysDir(), line))
		}
	}
	return paths, s.Err()
}

func replaysDir() string {
	if dir := os.Getenv("OPENBW_REPLAYS"); dir != "" {
		return dir
	}
	return "testdata/replays"
}

func parallel[T any](n, workers int, run func(int) T) []T {
	out := make([]T, n)
	next := make(chan int)
	var wg sync.WaitGroup
	for range workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range next {
				out[i] = run(i)
			}
		}()
	}
	for i := range n {
		next <- i
	}
	close(next)
	wg.Wait()
	return out
}

func validate(data *openbw.Data, paths []string, workers int) error {
	start := time.Now()
	results := parallel(len(paths), workers, func(i int) *result { return validateReplay(data, paths[i]) })
	fmt.Printf("parallel validation: %d replays in %v\n", len(paths), time.Since(start).Round(time.Second))
	start = time.Now()
	for _, r := range results {
		if r.err != nil {
			continue
		}
		replay, err := os.ReadFile(r.path)
		if err != nil {
			return err
		}
		if r.serialDigest, err = streamDigest(data, replay, openbw.Options{}, newDigester(false)); err != nil {
			r.fail("determinism", "serial run failed: %v", err)
		} else if r.serialDigest != r.digest {
			r.fail("determinism", "parallel and serial streams differ")
		}
	}
	fmt.Printf("serial rerun: %v\n", time.Since(start).Round(time.Second))
	report(results)
	return nil
}

func report(results []*result) {
	var errs, scr, legacy, scrSame, noLoad, events, unitEvents, checkpoints, units, commands, dropped, selections int
	var frames int64
	checks := map[string]int{}
	byMap, noLoadMaps := map[string]int{}, map[string]int{}
	failedReplays := map[string]int{}
	var diverge []string
	for _, r := range results {
		if r.err != nil {
			errs++
			fmt.Printf("ERROR %s: %v\n", r.path, r.err)
			continue
		}
		if r.rulesSCR {
			scr++
			switch {
			case r.openbwLoadErr != nil:
				noLoad++
				noLoadMaps[r.mapName]++
			case r.openbwDigest == r.digest:
				scrSame++
			default:
				diverge = append(diverge, fmt.Sprintf("%s: first difference at frame %d of %d, explained by %v", r.path, r.openbwDiverge, r.frames, r.rulesCause))
			}
		} else {
			legacy++
		}
		byMap[r.mapName]++
		events += r.events
		unitEvents += r.unitEvents
		checkpoints += r.checkpoints
		units += r.unitsCompared
		commands += r.commands
		dropped += r.dropped
		selections += r.selections
		frames += int64(r.frames)
		for check, n := range r.failureByCheck {
			checks[check] += n
			failedReplays[check]++
		}
		if r.failureCount > 0 {
			fmt.Printf("FAIL %s (%d failures)\n", r.path, r.failureCount)
			for _, f := range r.failures {
				fmt.Printf("  %s\n", f)
			}
		}
	}
	fmt.Printf("\nreplays %d (SC:R %d, legacy %d, errors %d), frames %d\n", len(results), scr, legacy, errs, frames)
	fmt.Printf("events %d, unit events %d, checkpoints %d, units compared %d\n", events, unitEvents, checkpoints, units)
	fmt.Printf("command results %d, selections compared %d, dropped by downgrade %d\n", commands, selections, dropped)
	fmt.Printf("SC:R replays with the same stream under OpenBW rules: %d of %d; %d do not load without SC:R rules (disabled thingy units)\n", scrSame, scr, noLoad)
	fmt.Printf("maps: %v\n", byMap)
	fmt.Printf("maps that do not load without SC:R rules: %v\n", noLoadMaps)
	for _, d := range diverge {
		fmt.Printf("  rules differ: %s\n", d)
	}
	for _, check := range slices.Sorted(maps.Keys(checks)) {
		fmt.Printf("check %-24s %d failures in %d replays\n", check, checks[check], failedReplays[check])
	}
	if len(checks) == 0 {
		fmt.Println("all checks passed")
	}
}
