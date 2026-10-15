package openbw_test

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"

	openbw "github.com/marianogappa/openbw-go"
)

var exampleReplay = replayPath("cwalggdl/replays-cwal-dl/30-NEW-lillljilililili/MM-CB157510-11C2-11F1-A3DD-FA167B5461B6.rep")

var exampleReplays = []string{
	exampleReplay,
	replayPath("cwalggdl/replays-cwal-dl/30-Esccccccc/MM-4862526E-25E0-11F1-A3DD-FA167B5461B6.rep"),
}

func replayPath(name string) string {
	dir := os.Getenv("OPENBW_REPLAYS")
	if dir == "" {
		dir = "testdata/replays"
	}
	return filepath.Join(dir, name)
}

// Open a replay, range over its events and check the error after the loop.
func Example() {
	replay, err := os.ReadFile(exampleReplay)
	if err != nil {
		log.Fatal(err)
	}
	g, err := openbw.Open(replay, openbw.Options{})
	if err != nil {
		log.Fatal(err)
	}
	for e := range g.Events() {
		if d, ok := e.(*openbw.UnitDestroyed); ok && d.Killer != 0 && d.Frame < 24*60*5 {
			fmt.Println(d.Frame, d.Type, "of player", d.Owner, "killed by", d.KillerType, "of player", d.KillerOwner)
		}
	}
	if err := g.Err(); err != nil {
		log.Fatal(err)
	}
	fmt.Println("end frame", g.Frame())
}

// Subscribe only the kinds you need. Unsubscribed kinds cost nothing.
// Classes leaves out hallucinations, spider mines, Interceptors, Scarabs and spell units.
func Example_filterKinds() {
	replay, err := os.ReadFile(exampleReplay)
	if err != nil {
		log.Fatal(err)
	}
	g, err := openbw.Open(replay, openbw.Options{
		Kinds:    openbw.KindUnitCreated | openbw.KindMorphStarted,
		Classes:  openbw.Classes(openbw.ClassNormal),
		MaxFrame: 24 * 60 * 2,
	})
	if err != nil {
		log.Fatal(err)
	}
	for e := range g.Events() {
		switch e := e.(type) {
		case *openbw.UnitCreated:
			if e.Source == openbw.SourceTrained || e.Source == openbw.SourceBuilt {
				fmt.Println(e.Frame, e.Owner, e.Source, e.Type)
			}
		case *openbw.MorphStarted:
			fmt.Println(e.Frame, e.Owner, "Morph", e.Into)
		}
	}
	if err := g.Err(); err != nil {
		log.Fatal(err)
	}
}

// Read the engine state between events: on a Tick every minute, print each player's resources and supply.
func Example_readState() {
	replay, err := os.ReadFile(exampleReplay)
	if err != nil {
		log.Fatal(err)
	}
	g, err := openbw.Open(replay, openbw.Options{Kinds: openbw.KindTick, TickEvery: 24 * 60, MaxFrame: 24 * 60 * 3})
	if err != nil {
		log.Fatal(err)
	}
	for e := range g.Events() {
		for _, owner := range g.PlayingSlots() {
			p := g.Player(owner)
			fmt.Printf("frame %d player %d: %d minerals, %d gas, supply %g/%g\n",
				e.Base().Frame, owner, p.Minerals, p.Gas, p.SupplyUsed, p.SupplyMax)
		}
	}
	if err := g.Err(); err != nil {
		log.Fatal(err)
	}
}

// Stop at the first rejected command. The engine runs only as far as the loop reads.
// A new call to Events continues where the loop stopped.
func Example_earlyStop() {
	replay, err := os.ReadFile(exampleReplay)
	if err != nil {
		log.Fatal(err)
	}
	g, err := openbw.Open(replay, openbw.Options{Kinds: openbw.KindCommandResult})
	if err != nil {
		log.Fatal(err)
	}
	for e := range g.Events() {
		r := e.(*openbw.CommandResult)
		if r.Outcome != openbw.Rejected {
			continue
		}
		fmt.Printf("frame %d: player %d command %#x rejected: %v\n", r.Frame, r.Owner, r.ActionID, r.Rejections)
		for _, id := range r.Selection {
			if u, ok := g.Unit(id); ok {
				fmt.Printf("selected %v at %d,%d\n", u.Type, u.X, u.Y)
			}
		}
		break
	}
	if err := g.Err(); err != nil {
		log.Fatal(err)
	}
	fmt.Println("stopped at frame", g.Frame(), "of", g.EndFrame())
}

// Run many games in parallel. One Data is loaded once and shared; each Game belongs to one goroutine.
func Example_parallel() {
	data, err := openbw.LoadData(os.Getenv("OPENBW_DATA"))
	if err != nil {
		log.Fatal(err)
	}
	deaths := make([]int, len(exampleReplays))
	errs := make([]error, len(exampleReplays))
	var wg sync.WaitGroup
	for i, path := range exampleReplays {
		wg.Go(func() {
			replay, err := os.ReadFile(path)
			if err != nil {
				errs[i] = err
				return
			}
			g, err := openbw.Open(replay, openbw.Options{Data: data, Kinds: openbw.KindUnitDestroyed})
			if err != nil {
				errs[i] = err
				return
			}
			for range g.Events() {
				deaths[i]++
			}
			errs[i] = g.Err()
		})
	}
	wg.Wait()
	for i := range exampleReplays {
		if errs[i] != nil {
			log.Fatal(errs[i])
		}
		fmt.Printf("game %d: %d units destroyed\n", i, deaths[i])
	}
}
