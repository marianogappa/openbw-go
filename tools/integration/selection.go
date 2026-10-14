package main

import (
	"io/fs"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/icza/screp/repparser"
)

const (
	framesPerMinute       = 1429
	minimumGameFrames     = 5 * framesPerMinute
	maxPicksPerFailingMap = 1
	maximumLongGameFrames = 60 * framesPerMinute
)

var knownFailingMaps = map[string]bool{"Pole Star 1.1": true}

type candidate struct {
	Path     string
	Map      string
	Date     time.Time
	Frames   int
	Matchup  string
	Limits   []uint32
	BugFixes string
	Build    uint32
}

func scanReplays(dir string) ([]candidate, error) {
	var candidates []candidate
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || filepath.Ext(path) != ".rep" {
			return err
		}
		if c, ok := readCandidate(path); ok {
			candidates = append(candidates, c)
		}
		return nil
	})
	sort.Slice(candidates, func(i, j int) bool { return candidates[i].Path < candidates[j].Path })
	return candidates, err
}

func readCandidate(path string) (candidate, bool) {
	r, err := repparser.ParseFileConfig(path, repparser.Config{})
	if err != nil {
		return candidate{}, false
	}
	h := r.Header
	if len(h.Players) != 2 || int(h.Frames) < minimumGameFrames {
		return candidate{}, false
	}
	extra, _ := readExtraSections(path)
	return candidate{Path: path, Map: plainMapName(h.Map), Date: h.StartTime, Frames: int(h.Frames), Matchup: h.Matchup(),
		Limits: extra.Limits, BugFixes: extra.BugFixes, Build: extra.Build}, true
}

func plainMapName(name string) string {
	return strings.TrimSpace(strings.Map(func(r rune) rune {
		if r < ' ' {
			return -1
		}
		return r
	}, name))
}

func pickReplays(candidates []candidate, count, longCount int) []candidate {
	picked := pickLongest(candidates, longCount)
	chosen := map[string]bool{}
	for _, c := range picked {
		chosen[c.Path] = true
	}
	byMap := groupByMapSortedByDate(candidates)
	maps := mapsByPopularity(byMap)
	picksPerMap := map[string]int{}
	for round := 0; len(picked) < count && round < len(dateQuantiles); round++ {
		for _, m := range maps {
			if len(picked) >= count {
				break
			}
			if knownFailingMaps[m] && picksPerMap[m] >= maxPicksPerFailingMap {
				continue
			}
			c, ok := pickAtQuantile(byMap[m], dateQuantiles[round], chosen)
			if !ok {
				continue
			}
			chosen[c.Path] = true
			picksPerMap[m]++
			picked = append(picked, c)
		}
	}
	sort.Slice(picked, func(i, j int) bool { return picked[i].Date.Before(picked[j].Date) })
	return picked
}

var dateQuantiles = []float64{0.5, 0.05, 0.95, 0.25, 0.75, 0.15, 0.85}

func pickLongest(candidates []candidate, n int) []candidate {
	sorted := append([]candidate(nil), candidates...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].Frames > sorted[j].Frames })
	var picked []candidate
	usedMaps := map[string]bool{}
	for _, c := range sorted {
		if len(picked) >= n {
			break
		}
		if usedMaps[c.Map] || knownFailingMaps[c.Map] || c.Frames > maximumLongGameFrames {
			continue
		}
		usedMaps[c.Map] = true
		picked = append(picked, c)
	}
	return picked
}

func groupByMapSortedByDate(candidates []candidate) map[string][]candidate {
	byMap := map[string][]candidate{}
	for _, c := range candidates {
		byMap[c.Map] = append(byMap[c.Map], c)
	}
	for _, list := range byMap {
		sort.SliceStable(list, func(i, j int) bool { return list[i].Date.Before(list[j].Date) })
	}
	return byMap
}

func mapsByPopularity(byMap map[string][]candidate) []string {
	var maps []string
	for m := range byMap {
		maps = append(maps, m)
	}
	sort.Slice(maps, func(i, j int) bool {
		if len(byMap[maps[i]]) != len(byMap[maps[j]]) {
			return len(byMap[maps[i]]) > len(byMap[maps[j]])
		}
		return maps[i] < maps[j]
	})
	return maps
}

func pickAtQuantile(list []candidate, quantile float64, chosen map[string]bool) (candidate, bool) {
	start := int(quantile * float64(len(list)-1))
	for offset := 0; offset < len(list); offset++ {
		c := list[(start+offset)%len(list)]
		if !chosen[c.Path] {
			return c, true
		}
	}
	return candidate{}, false
}

func filterCandidates(candidates []candidate, cfg config) []candidate {
	var kept []candidate
	for _, c := range candidates {
		date := c.Date.Format("2006-01-02")
		switch {
		case cfg.from != "" && date < cfg.from:
		case cfg.to != "" && date >= cfg.to:
		case cfg.race != "" && !strings.Contains(c.Matchup, cfg.race):
		default:
			kept = append(kept, c)
		}
	}
	return kept
}
