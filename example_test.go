package openbw_test

import (
	"fmt"
	"log"
	"os"

	openbw "github.com/marianogappa/openbw-go"
)

var exampleReplay = "/Users/mariano.gappa/Code/go/src/github.com/marianogappa/cwalggdl/replays-cwal-dl/30-NEW-lillljilililili/MM-CB157510-11C2-11F1-A3DD-FA167B5461B6.rep"

func gameTime(frame int32) string {
	seconds := frame * 42 / 1000
	return fmt.Sprintf("%2d:%02d", seconds/60, seconds%60)
}

// Stream the build order of the first 8 minutes. The engine stops at MaxFrame.
func Example_buildOrder() {
	replay, err := os.ReadFile(exampleReplay)
	if err != nil {
		log.Fatal(err)
	}
	g, err := openbw.Open(replay, openbw.StreamOptions{
		Kinds:    openbw.KindUnitCreated | openbw.KindMorphStarted | openbw.KindResearchStarted,
		MaxFrame: 8 * 60 * 24,
	})
	if err != nil {
		log.Fatal(err)
	}
	for e := range g.Events() {
		switch e := e.(type) {
		case *openbw.UnitCreated:
			if e.Source == openbw.SourceBuilt || e.Source == openbw.SourceTrained {
				fmt.Println(gameTime(e.Frame), e.Owner, e.Type)
			}
		case *openbw.MorphStarted:
			fmt.Println(gameTime(e.Frame), e.Owner, e.Into)
		case *openbw.ResearchStarted:
			fmt.Println(gameTime(e.Frame), e.Owner, e.Research)
		}
	}
	if err := g.Err(); err != nil {
		log.Fatal(err)
	}
}

// Count the units each player killed, leaving out hallucinations and spider mines.
func Example_killsPerPlayer() {
	replay, err := os.ReadFile(exampleReplay)
	if err != nil {
		log.Fatal(err)
	}
	g, err := openbw.Open(replay, openbw.StreamOptions{
		Kinds:   openbw.KindUnitDestroyed,
		Classes: openbw.Classes(openbw.ClassNormal, openbw.ClassFighter),
	})
	if err != nil {
		log.Fatal(err)
	}
	kills := map[int8]int{}
	for e := range g.Events() {
		d := e.(*openbw.UnitDestroyed)
		if d.Reason == openbw.ReasonDamage && d.KillerOwner >= 0 && d.KillerOwner != d.Owner && d.Owner != 11 {
			kills[d.KillerOwner]++
		}
	}
	if err := g.Err(); err != nil {
		log.Fatal(err)
	}
	for _, owner := range g.PlayingSlots() {
		fmt.Printf("player %d killed %d units\n", owner, kills[int8(owner)])
	}
}

// Run the whole replay with the default options and print a summary of the Result.
func ExampleRun() {
	replay, err := os.ReadFile(exampleReplay)
	if err != nil {
		log.Fatal(err)
	}
	res, err := openbw.Run(replay, openbw.Options{})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("reached end:", res.Engine.ReachedEnd)
	fmt.Println("likely desync:", res.Trust.Desync.Likely)
	last := map[int8]openbw.PlayerStats{}
	for _, s := range res.Stats {
		if s.Frame <= 10*60*24 {
			last[s.Owner] = s
		}
	}
	for _, p := range res.Trust.Production {
		s := last[p.Owner]
		fmt.Printf("player %d at 10:00: %d workers, army value %d/%d\n", p.Owner, s.Workers, s.ArmyMinerals, s.ArmyGas)
	}
}
