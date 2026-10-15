package main

import (
	"fmt"
	"io"
	"strings"

	"github.com/icza/screp/rep"
	openbw "github.com/marianogappa/openbw-go"
)

func printOverview(out io.Writer, res *openbw.Result) {
	h := res.Header
	fmt.Fprintf(out, "Engine  : %s %s   openbw %s (OpenBW %s), rules %s\n", h.Engine.ShortName, h.Version,
		res.Engine.Version, res.Engine.OpenBWCommit, rulesText(res.Engine))
	fmt.Fprintln(out, "Map     :", printable(h.Map))
	fmt.Fprintf(out, "Length  : %s (%d frames), %s\n", h.Frames, h.Frames, simulatedText(res.Engine))
	fmt.Fprintln(out, "Players :", playersText(h))
	fmt.Fprintln(out, "Trust   :", trustText(res.Trust))
	printFinal(out, res)
}

func printable(s string) string {
	return strings.Map(func(r rune) rune {
		if r < ' ' {
			return -1
		}
		return r
	}, s)
}

func rulesText(e openbw.EngineInfo) string {
	if e.Rules.Mode != "Auto" {
		return strings.ToLower(e.Rules.Mode)
	}
	if e.SourceFormat == "legacy" {
		return "auto (OpenBW)"
	}
	return "auto (SC:R)"
}

func simulatedText(e openbw.EngineInfo) string {
	if e.ReachedEnd {
		return "simulated to the end"
	}
	text := fmt.Sprintf("simulated to frame %d", e.FramesSimulated)
	if e.Error != "" {
		text += " (" + e.Error + ")"
	}
	return text
}

func raceLetter(p *rep.Player) string {
	if p.Race == nil {
		return "?"
	}
	return string(p.Race.Letter)
}

func playersText(h *rep.Header) string {
	var teams []string
	var current []string
	var prevTeam byte
	for _, p := range h.Players {
		if p.Observer {
			continue
		}
		if len(current) > 0 && p.Team != prevTeam {
			teams = append(teams, strings.Join(current, ", "))
			current = nil
		}
		current = append(current, fmt.Sprintf("%s (%s)", p.Name, raceLetter(p)))
		prevTeam = p.Team
	}
	if len(current) > 0 {
		teams = append(teams, strings.Join(current, ", "))
	}
	return strings.Join(teams, "  vs  ")
}

func trustText(t openbw.Trust) string {
	dropped := "no dropped commands"
	if t.DroppedCommands > 0 {
		dropped = fmt.Sprintf("%d dropped commands", t.DroppedCommands)
	}
	verdict := "clean"
	if t.Desync.Likely {
		verdict = "likely desync"
	}
	var rates []string
	for _, p := range t.Production {
		rate := 0.0
		if p.Issued > 0 {
			rate = 100 * float64(p.Rejected+p.NoEffect) / float64(p.Issued)
		}
		rates = append(rates, fmt.Sprintf("%.1f%%", rate))
	}
	return fmt.Sprintf("%s, %d inconsistent selections (%s), rejected production %s",
		dropped, t.Desync.InconsistentSelects, verdict, strings.Join(rates, " / "))
}

func printFinal(out io.Writer, res *openbw.Result) {
	last := map[int8]openbw.PlayerStats{}
	var owners []int8
	for _, s := range res.Stats {
		if _, seen := last[s.Owner]; !seen {
			owners = append(owners, s.Owner)
		}
		last[s.Owner] = s
	}
	label := "Final   :"
	for _, owner := range owners {
		s := last[owner]
		race := "?"
		if int(owner) < len(res.Header.Slots) {
			race = raceLetter(res.Header.Slots[owner])
		}
		fmt.Fprintf(out, "%-9s %s  supply %3g/%-3g  workers %3d  army value %d/%d\n",
			label, race, s.SupplyUsed, s.SupplyMax, s.Workers, s.ArmyMinerals, s.ArmyGas)
		label = ""
	}
}
