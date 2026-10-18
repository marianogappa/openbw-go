package main

import (
	"cmp"
	"flag"
	"fmt"
	"os"

	"github.com/marianogappa/openbw-go/bwgame"
	"github.com/marianogappa/openbw-go/downgrade"
	"github.com/marianogappa/openbw-go/scrrules"
)

func main() {
	data := flag.String("data", cmp.Or(os.Getenv("OPENBW_DATA"), "testdata/bwdata"), "game data directory")
	from := flag.Int("from", 0, "first frame to print")
	to := flag.Int("to", 0, "last frame to run")
	index := flag.Int("idx", 0, "unit index (1-based, as in tags) to watch")
	every := flag.Int("every", 0, "print the watched unit and the receding creep pool every n frames")
	areaX := flag.Int("x", 0, "area centre x")
	areaY := flag.Int("y", 0, "area centre y")
	areaRadius := flag.Int("r", 0, "area half-size: also print damage and deaths of units in this square")
	flag.Parse()
	global, err := bwgame.GlobalInitFromDirectory(*data)
	if err != nil {
		panic(err)
	}
	raw, err := os.ReadFile(flag.Arg(0))
	if err != nil {
		panic(err)
	}
	loadable, settings, err := downgrade.PrepareForOpenBW(raw)
	if err != nil {
		panic(err)
	}
	st := bwgame.NewState(global, bwgame.NewGameState())
	st.Game.Rules = scrrules.FromSettings(settings)
	f := bwgame.NewReplayFunctions(st, &bwgame.ActionState{}, &bwgame.ReplayState{})
	inWindow := func() bool { return int(st.CurrentFrame) >= *from && int(st.CurrentFrame) <= *to }
	watched := func(u *bwgame.Unit) bool {
		if int(u.Index+1) == *index {
			return true
		}
		return *areaRadius > 0 && abs(int(u.Position.X)-*areaX) <= *areaRadius && abs(int(u.Position.Y)-*areaY) <= *areaRadius
	}
	f.Hooks.OnUnitDamaged = func(u, source *bwgame.Unit, sourceOwner int32, damage bwgame.Fp8) {
		if inWindow() && watched(u) {
			fmt.Printf("f=%d dmg=%d to %s from p%d %s\n", st.CurrentFrame, damage.RawValue, describe(u), sourceOwner, describe(source))
		}
	}
	f.Hooks.OnUnitDestroy = func(u *bwgame.Unit) {
		if inWindow() && watched(u) {
			fmt.Printf("f=%d destroy %s\n", st.CurrentFrame, describe(u))
		}
	}
	if err := f.LoadReplayData(loadable); err != nil {
		panic(err)
	}
	for !f.IsDone() && int(st.CurrentFrame) <= *to {
		if *every > 0 && inWindow() && int(st.CurrentFrame)%*every == 0 {
			fmt.Printf("f=%d creepFree=%d creepLists=%v\n", st.CurrentFrame, st.CreepLife.FreeListSize, st.CreepLife.ListsSize)
			for u := range st.VisibleUnits.All() {
				if int(u.Index+1) == *index {
					tile := int(u.Position.Y/32)*int(st.Game.MapTileWidth) + int(u.Position.X/32)
					fmt.Printf("f=%d %s tileFlags=%#x\n", st.CurrentFrame, describe(u), st.Tiles[tile].Flags)
				}
			}
		}
		if err := f.NextFrame(); err != nil {
			panic(err)
		}
	}
}

func describe(u *bwgame.Unit) string {
	if u == nil {
		return "nil"
	}
	return fmt.Sprintf("[%d p%d type=%v hp=%d shields=%d pos=%v order=%v]", u.Index+1, u.Owner, u.UnitType.ID, u.Hp.RawValue, u.ShieldPoints.RawValue, u.Position, u.OrderType.ID)
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}
