package main

import (
	"bufio"
	"bytes"
	"sort"
	"strconv"
	"strings"
)

func lines(data []byte) [][]string {
	var out [][]string
	scanner := bufio.NewScanner(bytes.NewReader(data))
	scanner.Buffer(make([]byte, 1<<20), 1<<20)
	for scanner.Scan() {
		if fields := strings.Fields(scanner.Text()); len(fields) > 0 {
			out = append(out, fields)
		}
	}
	return out
}

func atoi(s string) int {
	n, _ := strconv.Atoi(s)
	return n
}

func keyValues(fields []string) map[string]string {
	kv := map[string]string{}
	for _, f := range fields {
		if k, v, ok := strings.Cut(f, "="); ok {
			kv[k] = v
		}
	}
	return kv
}

func typeCounts(fields []string) map[int]int {
	counts := map[int]int{}
	for k, v := range keyValues(fields) {
		counts[atoi(k)] = atoi(v)
	}
	return counts
}

func parseHarnessOutput(out []byte, res *result, players map[int]*playerResult) {
	for _, f := range lines(out) {
		switch {
		case f[0] == "frame" && len(f) >= 2 && f[1] == "0":
			res.Loaded = true
		case f[0] == "end" && len(f) >= 2:
			res.EndFrame = atoi(f[1])
		case f[0] == "error":
			res.EngineError = strings.Join(f[1:], " ")
			if len(f) >= 2 {
				res.EndFrame = atoi(f[1])
			}
		case f[0] == "action_fails" && len(f) >= 2:
			res.ActionFails = atoi(f[1])
		case f[0] == "player" && len(f) >= 2:
			applyPlayerLine(players[atoi(f[1])], keyValues(f[2:]))
		case f[0] == "units" && len(f) >= 2:
			applyAliveUnits(players[atoi(f[1])], typeCounts(f[2:]))
		}
	}
}

func applyPlayerLine(p *playerResult, kv map[string]string) {
	if p == nil {
		return
	}
	p.Minerals = atoi(kv["minerals"])
	p.Gas = atoi(kv["gas"])
	p.Supply = kv[supplyKey(p.Race)]
}

func supplyKey(race string) string {
	return "supply_" + strings.ToLower(race)
}

func applyAliveUnits(p *playerResult, counts map[int]int) {
	if p == nil {
		return
	}
	p.AliveByType = counts
	for _, n := range counts {
		p.AliveUnits += n
	}
}

func parseHarnessFails(errOut []byte, players map[int]*playerResult) {
	for _, f := range lines(errOut) {
		if f[0] != "action_fail" || len(f) < 4 {
			continue
		}
		if p := players[atoi(f[2])]; p != nil {
			p.FailsByAction[atoi(f[3])]++
		}
	}
}

func parseProbeOutput(out []byte, res *result, players map[int]*playerResult) {
	for _, f := range lines(out) {
		if f[0] == "limit" && len(f) >= 2 {
			res.EngineLimits[f[1]] = parseLimitUsage(keyValues(f[2:]))
			continue
		}
		p := players[probeLineSlot(f)]
		if p == nil {
			continue
		}
		switch f[0] {
		case "made":
			p.MadeByType = typeCounts(f[2:])
		case "made_totals":
			kv := keyValues(f[2:])
			p.MadeMen = atoi(kv["men"])
			p.MadeBuildings = atoi(kv["buildings"])
		case "no_units":
			if len(f) >= 3 {
				p.NoUnitsFrame = atoi(f[2])
			}
		case "reject":
			if len(f) >= 5 {
				p.addReject(atoi(f[2]), f[3], atoi(f[4]))
			}
		case "selection_tags":
			p.SelectionTags.add(addTagMinute(p, f))
		case "target_tags":
			p.TargetTags.add(addTagMinute(p, f))
		case "timeline":
			p.GatheredMinerals = atoi(keyValues(f[3:])["gathered_minerals"])
		}
	}
}

func probeLineSlot(f []string) int {
	switch {
	case f[0] == "timeline" && len(f) >= 3:
		return atoi(f[2])
	case f[0] != "timeline" && len(f) >= 2:
		return atoi(f[1])
	}
	return -1
}

func sortedKeys(m map[int]int) []int {
	keys := make([]int, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	return keys
}

func sortedSlots(players map[int]*playerResult) []int {
	slots := make([]int, 0, len(players))
	for k := range players {
		slots = append(slots, k)
	}
	sort.Ints(slots)
	return slots
}

func addTagMinute(p *playerResult, f []string) tagCounts {
	if len(f) < 3 {
		return tagCounts{}
	}
	kv := keyValues(f[3:])
	c := tagCounts{Total: atoi(kv["total"]), Missing: atoi(kv["missing"])}
	minute := atoi(f[2])
	sum := p.tagsByMinute[minute]
	sum.add(c)
	p.tagsByMinute[minute] = sum
	return c
}

func parseLimitUsage(kv map[string]string) limitUsage {
	return limitUsage{Peak: atoi(kv["peak"]), Max: atoi(kv["max"]), FirstFullFrame: atoi(kv["first_full"])}
}

var anomalyRejectReasons = map[string]bool{"reason=no_unit": true, "reason=cannot_build": true}

func parseAnomalyFrames(errOut []byte) []int {
	var frames []int
	for _, f := range lines(errOut) {
		if len(f) < 2 {
			continue
		}
		isMissingTag := f[0] == "missing_tag"
		isMismatch := f[0] == "reject" && len(f) >= 5 && anomalyRejectReasons[f[4]]
		if isMissingTag || isMismatch {
			frames = append(frames, atoi(strings.TrimPrefix(f[1], "frame=")))
		}
	}
	return frames
}
