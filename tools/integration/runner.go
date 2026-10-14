package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/marianogappa/openbw-go/downgrade"
)

type result struct {
	Replay          candidate
	LegacyPath      string
	DowngradeError  string `json:",omitempty"`
	DroppedCommands int
	Changes         []string
	Loaded          bool
	ReachedEnd      bool
	EngineError     string `json:",omitempty"`
	EndFrame        int
	ActionFails     int
	Players         []*playerResult
	MissingTagOnset int
	AnomalyOnset    int
	EngineLimits    map[string]limitUsage
	Verdict         string
	Evidence        []string
}

type playerResult struct {
	commandStats
	Minerals         int
	Gas              int
	Supply           string
	AliveUnits       int
	AliveByType      map[int]int
	MadeMen          int
	MadeBuildings    int
	MadeByType       map[int]int
	GatheredMinerals int
	NoUnitsFrame     int
	CmdsAfterNoUnits int
	ProductionFails  int
	FailsByAction    map[int]int
	MismatchRejects  int
	EconomyRejects   int
	RejectsByReason  map[string]int
	MissingTypes     []string
	SelectionTags    tagCounts
	TargetTags       tagCounts
	tagsByMinute     map[int]tagCounts
}

type limitUsage struct {
	Peak           int
	Max            int
	FirstFullFrame int
}

type tagCounts struct {
	Total   int
	Missing int
}

func (c *tagCounts) add(o tagCounts) {
	c.Total += o.Total
	c.Missing += o.Missing
}

var mismatchReasons = map[string]bool{"no_unit": true, "cannot_build": true}

func (p *playerResult) addReject(actionID int, reason string, count int) {
	p.RejectsByReason[fmt.Sprintf("%#02x %s", actionID, reason)] += count
	if mismatchReasons[reason] {
		p.MismatchRejects += count
	} else {
		p.EconomyRejects += count
	}
}

func runAll(cfg config, picked []candidate) []result {
	results := make([]result, len(picked))
	jobs := make(chan int)
	var wg sync.WaitGroup
	for range cfg.workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range jobs {
				results[i] = runOne(cfg, picked[i])
				fmt.Fprintf(os.Stderr, "done %d/%d %s\n", i+1, len(picked), filepath.Base(picked[i].Path))
			}
		}()
	}
	for i := range picked {
		jobs <- i
	}
	close(jobs)
	wg.Wait()
	return results
}

func runOne(cfg config, c candidate) result {
	res := result{Replay: c, EngineLimits: map[string]limitUsage{}, LegacyPath: filepath.Join(cfg.outDir, strings.TrimSuffix(filepath.Base(c.Path), ".rep")+".116.rep")}
	if err := downgradeFile(c.Path, res.LegacyPath, &res); err != nil {
		res.DowngradeError = err.Error()
		return res
	}
	stats, err := readCommandStats(c.Path)
	if err != nil {
		res.DowngradeError = "screp: " + err.Error()
		return res
	}
	players := map[int]*playerResult{}
	for slot, s := range stats {
		players[slot] = &playerResult{commandStats: *s, AliveByType: map[int]int{}, MadeByType: map[int]int{}, FailsByAction: map[int]int{}, RejectsByReason: map[string]int{}, tagsByMinute: map[int]tagCounts{}}
	}
	harnessOut, harnessErr := runTool(cfg, cfg.harnessPath, "-every", "100000000", res.LegacyPath)
	saveRaw(cfg, res.LegacyPath, ".harness.out", harnessOut)
	saveRaw(cfg, res.LegacyPath, ".harness.err", harnessErr)
	parseHarnessOutput(harnessOut, &res, players)
	parseHarnessFails(harnessErr, players)
	probeOut, probeErr := runTool(cfg, cfg.probePath, res.LegacyPath)
	saveRaw(cfg, res.LegacyPath, ".probe.out", probeOut)
	saveRaw(cfg, res.LegacyPath, ".probe.err", probeErr)
	parseProbeOutput(probeOut, &res, players)
	res.AnomalyOnset = anomalyOnset(parseAnomalyFrames(probeErr))
	res.ReachedEnd = res.EngineError == "" && res.EndFrame == c.Frames
	for _, slot := range sortedSlots(players) {
		finishPlayer(players[slot])
		res.Players = append(res.Players, players[slot])
	}
	res.MissingTagOnset = missingTagOnset(res.Players)
	return res
}

func downgradeFile(inPath, outPath string, res *result) error {
	in, err := os.ReadFile(inPath)
	if err != nil {
		return err
	}
	out, report, err := downgrade.Downgrade(in)
	if err != nil {
		return err
	}
	res.DroppedCommands = report.DroppedCount()
	res.Changes = report.Changes
	return os.WriteFile(outPath, out, 0o644)
}

func runTool(cfg config, path string, args ...string) (stdout, stderr []byte) {
	cmd := exec.Command(path, args...)
	cmd.Env = append(os.Environ(), "OPENBW_DATA="+cfg.dataDir)
	var out, errOut bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errOut
	_ = cmd.Run()
	return out.Bytes(), errOut.Bytes()
}

func saveRaw(cfg config, legacyPath, suffix string, data []byte) {
	_ = os.WriteFile(strings.TrimSuffix(legacyPath, ".116.rep")+suffix, data, 0o644)
}

func finishPlayer(p *playerResult) {
	for _, id := range productionActionIDs {
		p.ProductionFails += p.FailsByAction[id]
	}
	if p.NoUnitsFrame > 0 {
		p.CmdsAfterNoUnits = p.cmdsAfter(p.NoUnitsFrame)
	}
	for _, id := range sortedKeys(p.OrdersByType) {
		if p.OrdersByType[id] >= minimumOrdersForMissingType && p.MadeByType[id] == 0 {
			p.MissingTypes = append(p.MissingTypes, fmt.Sprintf("%s (%d orders)", unitName(id), p.OrdersByType[id]))
		}
	}
}

var productionActionIDs = []int{0x0c, 0x1f, 0x23, 0x35}

const anomaliesForOnset = 5

func anomalyOnset(frames []int) int {
	for i, start := range frames {
		end := i
		for end < len(frames) && frames[end] < start+framesPerMinute {
			end++
		}
		if end-i >= anomaliesForOnset {
			return start
		}
	}
	return noOnset
}

const (
	onsetMinimumMissing = 5
	onsetMinimumRate    = 0.05
	noOnset             = -1
)

func missingTagOnset(players []*playerResult) int {
	byMinute := map[int]tagCounts{}
	for _, p := range players {
		for minute, c := range p.tagsByMinute {
			sum := byMinute[minute]
			sum.add(c)
			byMinute[minute] = sum
		}
	}
	onset := noOnset
	for minute, c := range byMinute {
		if c.Missing >= onsetMinimumMissing && float64(c.Missing) >= onsetMinimumRate*float64(c.Total) && (onset == noOnset || minute < onset) {
			onset = minute
		}
	}
	return onset
}
