package main

import (
	"fmt"
	"io"
	"path/filepath"
	"sort"
	"strings"
)

func writeMarkdown(w io.Writer, results []result) {
	writeSummaryTable(w, results)
	fmt.Fprintln(w)
	writePlayerTable(w, results)
	fmt.Fprintln(w)
	writeEvidence(w, results)
}

func writeSummaryTable(w io.Writer, results []result) {
	fmt.Fprintln(w, "| # | Replay | Date | Map | Matchup | Frames | Loads | To end | Dropped | action_fail | Peak units / bullets / sprites / images / orders | First anomaly cluster (frame) | Verdict |")
	fmt.Fprintln(w, "|---|---|---|---|---|---|---|---|---|---|---|---|---|")
	for i, r := range results {
		fmt.Fprintf(w, "| %d | %s | %s | %s | %s | %d | %s | %s | %d | %d | %s | %s | %s |\n",
			i+1, shortName(r.Replay.Path), r.Replay.Date.Format("2006-01-02"), r.Replay.Map, r.Replay.Matchup,
			r.Replay.Frames, yesNo(r.Loaded), endCell(r), r.DroppedCommands, r.ActionFails, peaksCell(r), onsetCell(r), r.Verdict)
	}
}

func writePlayerTable(w io.Writer, results []result) {
	fmt.Fprintln(w, "| # | Player | Race | Minerals | Gas | Supply | Alive | Made units / orders | Made buildings / orders | Rejected: no unit / economy / cmds | Missing tags | No units at | Left at |")
	fmt.Fprintln(w, "|---|---|---|---|---|---|---|---|---|---|---|---|---|")
	for i, r := range results {
		for _, p := range r.Players {
			fmt.Fprintf(w, "| %d | %s | %s | %d | %d | %s | %d | %d / %d | %d / %d | %d / %d / %d | %d / %d | %s | %s |\n",
				i+1, p.Name, p.Race, p.Minerals, p.Gas, p.Supply, p.AliveUnits, p.MadeMen, p.OrderedUnits,
				p.MadeBuildings, p.OrderedBuildings, p.MismatchRejects, p.EconomyRejects, p.ProductionCmds,
				p.missingTags(), p.SelectionTags.Total+p.TargetTags.Total, frameCell(p.NoUnitsFrame), frameCell(p.LeaveFrame))
		}
	}
}

func writeEvidence(w io.Writer, results []result) {
	for i, r := range results {
		if len(r.Evidence) == 0 {
			continue
		}
		fmt.Fprintf(w, "- %d (%s): %s\n", i+1, r.Verdict, strings.Join(r.Evidence, "; "))
	}
}

func shortName(path string) string {
	return filepath.Base(filepath.Dir(path)) + "/" + strings.TrimSuffix(filepath.Base(path), ".rep")[:11]
}

func yesNo(b bool) string {
	if b {
		return "yes"
	}
	return "no"
}

func endCell(r result) string {
	if r.ReachedEnd {
		return "yes"
	}
	return fmt.Sprintf("no (%d)", r.EndFrame)
}

func frameCell(frame int) string {
	if frame == 0 {
		return "-"
	}
	return fmt.Sprint(frame)
}

func onsetCell(r result) string {
	if !r.Loaded || r.AnomalyOnset == noOnset {
		return "-"
	}
	return fmt.Sprintf("%d (min %d of %d)", r.AnomalyOnset, r.AnomalyOnset/framesPerMinute, r.Replay.Frames/framesPerMinute)
}

func writeSurvey(w io.Writer, candidates []candidate) {
	limits := map[string]int{}
	bugFixes := map[string]int{}
	builds := map[uint32]string{}
	buildCounts := map[uint32]int{}
	for _, c := range candidates {
		limits[fmt.Sprint(c.Limits)]++
		bugFixes[c.BugFixes]++
		buildCounts[c.Build]++
		builds[c.Build] = dateRange(builds[c.Build], c.Date.Format("2006-01-02"))
	}
	fmt.Fprintf(w, "Scanned %d replays (1v1, at least %d frames).\n\n", len(candidates), minimumGameFrames)
	fmt.Fprintf(w, "- LMTS (images, sprites, lone sprites, units, bullets, orders, fog sprites): %v\n", limits)
	fmt.Fprintf(w, "- BFIX: %v\n", bugFixes)
	for _, build := range sortedBuilds(buildCounts) {
		fmt.Fprintf(w, "- GCFG build %d: %d replays, %s\n", build, buildCounts[build], builds[build])
	}
	fmt.Fprintln(w)
}

var limitNames = []string{"units", "bullets", "sprites", "images", "orders"}

func peaksCell(r result) string {
	if len(r.EngineLimits) == 0 {
		return "-"
	}
	var parts []string
	for _, name := range limitNames {
		parts = append(parts, fmt.Sprint(r.EngineLimits[name].Peak))
	}
	return strings.Join(parts, " / ")
}

func sortedLimitNames(limits map[string]limitUsage) []string {
	var names []string
	for _, name := range limitNames {
		if _, ok := limits[name]; ok {
			names = append(names, name)
		}
	}
	return names
}

func dateRange(current, date string) string {
	if current == "" {
		return date + " to " + date
	}
	first, last, _ := strings.Cut(current, " to ")
	if date < first {
		first = date
	}
	if date > last {
		last = date
	}
	return first + " to " + last
}

func sortedBuilds(counts map[uint32]int) []uint32 {
	var builds []uint32
	for b := range counts {
		builds = append(builds, b)
	}
	sort.Slice(builds, func(i, j int) bool { return builds[i] < builds[j] })
	return builds
}
