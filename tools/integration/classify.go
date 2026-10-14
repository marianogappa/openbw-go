package main

import "fmt"

const (
	verdictClean      = "clean"
	verdictSuspicious = "suspicious"
	verdictDesynced   = "desynced"
	verdictFailed     = "failed"
)

const (
	desyncCmdsAfterNoUnits  = 20
	desyncMissingTypes      = 2
	desyncMismatches        = 10
	suspiciousMismatches    = 3
	suspiciousMissingTags   = 5
	suspiciousMissingRate   = 0.0025
	minimumUnitOrders       = 20
	lowUnitsRatio           = 0.30
	unspentMineralsToReport = 3000
)

type signal struct {
	severity string
	text     string
}

func classify(res *result) {
	switch {
	case res.DowngradeError != "":
		res.Verdict = verdictFailed
		res.Evidence = []string{"downgrade: " + res.DowngradeError}
		return
	case !res.Loaded || !res.ReachedEnd:
		res.Verdict = verdictFailed
		res.Evidence = []string{fmt.Sprintf("engine stopped at frame %d of %d: %s", res.EndFrame, res.Replay.Frames, res.EngineError)}
		return
	}
	res.Verdict = verdictClean
	for _, name := range sortedLimitNames(res.EngineLimits) {
		if usage := res.EngineLimits[name]; usage.FirstFullFrame > 0 {
			res.Verdict = verdictSuspicious
			res.Evidence = append(res.Evidence, fmt.Sprintf("OpenBW %s limit %d reached at frame %d (SC:R limit is higher)", name, usage.Max, usage.FirstFullFrame))
		}
	}
	if res.MissingTagOnset != noOnset {
		res.Verdict = verdictDesynced
		res.Evidence = append(res.Evidence, fmt.Sprintf("selected or targeted units stop existing in the engine from minute %d of %d",
			res.MissingTagOnset, res.Replay.Frames/framesPerMinute))
	}
	for _, p := range res.Players {
		for _, s := range playerSignals(p) {
			res.Evidence = append(res.Evidence, fmt.Sprintf("%s: %s", p.Name, s.text))
			res.Verdict = worse(res.Verdict, s.severity)
		}
	}
}

func worse(a, b string) string {
	rank := map[string]int{verdictClean: 0, verdictSuspicious: 1, verdictDesynced: 2}
	if rank[b] > rank[a] {
		return b
	}
	return a
}

func playerSignals(p *playerResult) []signal {
	var signals []signal
	add := func(severity, format string, args ...any) {
		signals = append(signals, signal{severity, fmt.Sprintf(format, args...)})
	}
	if p.NoUnitsFrame > 0 && p.CmdsAfterNoUnits >= desyncCmdsAfterNoUnits {
		add(verdictDesynced, "no units at frame %d, but %d commands after it", p.NoUnitsFrame, p.CmdsAfterNoUnits)
	}
	switch n := len(p.MissingTypes); {
	case n >= desyncMissingTypes:
		add(verdictDesynced, "never made %v", p.MissingTypes)
	case n == 1:
		add(verdictSuspicious, "never made %v", p.MissingTypes)
	}
	switch {
	case p.MismatchRejects >= desyncMismatches:
		add(verdictDesynced, "%d production commands had no selected unit that could do them", p.MismatchRejects)
	case p.MismatchRejects >= suspiciousMismatches:
		add(verdictSuspicious, "%d production commands had no selected unit that could do them", p.MismatchRejects)
	}
	if missing, total := p.missingTags(), p.SelectionTags.Total+p.TargetTags.Total; missing >= suspiciousMissingTags && float64(missing) >= suspiciousMissingRate*float64(total) {
		add(verdictSuspicious, "%d of %d selected or targeted units do not exist in the engine", missing, total)
	}
	if p.OrderedUnits >= minimumUnitOrders && float64(p.MadeMen) < lowUnitsRatio*float64(p.OrderedUnits) {
		add(verdictClean, "made %d units for %d train orders (%d rejected for money, supply or queue)", p.MadeMen, p.OrderedUnits, p.EconomyRejects)
	}
	if p.Minerals >= unspentMineralsToReport {
		add(verdictClean, "%d unspent minerals at the end", p.Minerals)
	}
	return signals
}

func (p *playerResult) missingTags() int {
	return p.SelectionTags.Missing + p.TargetTags.Missing
}
