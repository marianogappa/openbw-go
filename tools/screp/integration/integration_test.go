package main

import (
	"testing"
	"time"
)

func TestPlainMapNameRemovesColorCodes(t *testing.T) {
	got := plainMapName("\x07P\x06ole \x07S\x06tar \x031\x04.1 ")
	if got != "Pole Star 1.1" {
		t.Fatalf("plainMapName = %q", got)
	}
}

func TestAnomalyOnsetNeedsAClusterWithinAMinute(t *testing.T) {
	spread := []int{100, 2000, 4000, 6000, 8000}
	if got := anomalyOnset(spread); got != noOnset {
		t.Fatalf("spread anomalies: onset = %d, want none", got)
	}
	clustered := []int{100, 5000, 5100, 5200, 5300, 5400}
	if got := anomalyOnset(clustered); got != 5000 {
		t.Fatalf("clustered anomalies: onset = %d, want 5000", got)
	}
}

func TestMissingTagOnsetUsesTheEarliestBadMinute(t *testing.T) {
	players := []*playerResult{
		{tagsByMinute: map[int]tagCounts{3: {Total: 100, Missing: 1}, 7: {Total: 50, Missing: 5}, 9: {Total: 10, Missing: 9}}},
		{tagsByMinute: map[int]tagCounts{5: {Total: 1000, Missing: 6}}},
	}
	if got := missingTagOnset(players); got != 7 {
		t.Fatalf("onset = %d, want 7", got)
	}
}

func TestPickReplaysCapsKnownFailingMapAndSpreadsMaps(t *testing.T) {
	var candidates []candidate
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	for i := range 20 {
		date := start.AddDate(0, 0, i)
		candidates = append(candidates,
			candidate{Path: "a" + date.Format("0102"), Map: "Pole Star 1.1", Date: date, Frames: 10000},
			candidate{Path: "b" + date.Format("0102"), Map: "Octagon 1.0", Date: date, Frames: 10000 + i},
		)
	}
	picked := pickReplays(candidates, 6, 1)
	if len(picked) != 6 {
		t.Fatalf("picked %d replays, want 6", len(picked))
	}
	poleStar := 0
	for _, c := range picked {
		if c.Map == "Pole Star 1.1" {
			poleStar++
		}
	}
	if poleStar != maxPicksPerFailingMap {
		t.Fatalf("picked %d Pole Star replays, want %d", poleStar, maxPicksPerFailingMap)
	}
}

func TestClassifyFlagsMissingTagOnsetAsDesynced(t *testing.T) {
	res := result{Loaded: true, ReachedEnd: true, MissingTagOnset: 12, Replay: candidate{Frames: 30 * framesPerMinute}}
	classify(&res)
	if res.Verdict != verdictDesynced {
		t.Fatalf("verdict = %s, want %s", res.Verdict, verdictDesynced)
	}
}

func TestClassifyIgnoresEconomyRejects(t *testing.T) {
	p := &playerResult{EconomyRejects: 300}
	p.OrderedUnits = 400
	p.MadeMen = 100
	res := result{Loaded: true, ReachedEnd: true, MissingTagOnset: noOnset, Players: []*playerResult{p}}
	classify(&res)
	if res.Verdict != verdictClean {
		t.Fatalf("verdict = %s, want %s (evidence %v)", res.Verdict, verdictClean, res.Evidence)
	}
}
