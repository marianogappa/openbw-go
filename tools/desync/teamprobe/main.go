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
	useSCRRules := flag.Bool("scr-rules", false, "run the engine with the replay's SC:R rules")
	from := flag.Int("from", 0, "first frame to trace")
	to := flag.Int("to", 3000, "last frame to trace")
	stats := flag.Int("stats", 0, "print object pool sizes every n frames")
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
	if *useSCRRules {
		st.Game.Rules = scrrules.FromSettings(settings)
	}
	f := bwgame.NewReplayFunctions(st, &bwgame.ActionState{}, &bwgame.ReplayState{})
	inWindow := func() bool { return int(st.CurrentFrame) >= *from && int(st.CurrentFrame) <= *to }
	f.BeforeAction = func(owner, actionID int32, payload []byte) {
		if !inWindow() {
			return
		}
		line := fmt.Sprintf("f=%d owner=%d act=%#x", st.CurrentFrame, owner, actionID)
		if (actionID == 0x09 || actionID == 0x0a || actionID == 0x0b) && len(payload) > 0 {
			n := int(payload[0])
			for i := 0; i < n && 2+2*i < len(payload); i++ {
				tag := uint16(payload[1+2*i]) | uint16(payload[2+2*i])<<8
				u := f.UnitByID(bwgame.UnitID{RawValue: tag})
				if u == nil {
					line += fmt.Sprintf(" [%#x:MISSING]", tag)
				} else {
					line += fmt.Sprintf(" [%d:p%d %v]", tag&0x7ff, u.Owner, u.UnitType.ID)
				}
			}
		} else if (actionID == 0x14 || actionID == 0x15) && len(payload) >= 6 {
			x := int(payload[0]) | int(payload[1])<<8
			y := int(payload[2]) | int(payload[3])<<8
			tag := uint16(payload[4]) | uint16(payload[5])<<8
			line += fmt.Sprintf(" pos=(%d,%d)", x, y)
			if tag != 0 {
				if u := f.UnitByID(bwgame.UnitID{RawValue: tag}); u == nil {
					line += fmt.Sprintf(" target=[%#x:MISSING]", tag)
				} else {
					line += fmt.Sprintf(" target=[%d:p%d %v %v]", tag&0x7ff, u.Owner, u.UnitType.ID, u.Position)
				}
			}
		} else {
			line += fmt.Sprintf(" % x", payload[:min(len(payload), 12)])
		}
		fmt.Println(line)
	}
	f.OnActionFail = func(frame, owner, actionID int32) {
		if inWindow() {
			fmt.Printf("  FAIL f=%d owner=%d act=%#x err=%d\n", frame, owner, actionID, st.LastError)
		}
	}
	f.Hooks.OnUnitCreated = func(u *bwgame.Unit, _ bwgame.CreateSource, _ *bwgame.Unit) {
		if inWindow() {
			fmt.Printf("  new f=%d idx=%d gen=%d owner=%d type=%v pos=%v\n", st.CurrentFrame, u.Index+1, u.UnitIDGeneration%32, u.Owner, u.UnitType.ID, u.Position)
		}
	}
	f.Hooks.OnUnitDestroy = func(u *bwgame.Unit) {
		if inWindow() {
			fmt.Printf("  del f=%d idx=%d owner=%d type=%v pos=%v\n", st.CurrentFrame, u.Index+1, u.Owner, u.UnitType.ID, u.Position)
		}
	}
	f.Hooks.OnUnitMorphed = func(u *bwgame.Unit, from *bwgame.UnitTypeT) {
		if inWindow() {
			fmt.Printf("  morph f=%d idx=%d owner=%d %v->%v\n", st.CurrentFrame, u.Index+1, u.Owner, from.ID, u.UnitType.ID)
		}
	}
	f.Hooks.OnVictoryState = func(owner, state int32) {
		fmt.Printf("  victory f=%d owner=%d state=%d alliances=%v\n", st.CurrentFrame, owner, state, st.Alliances[owner][:8])
	}
	f.Hooks.OnPlayerLeft = func(owner int32, dropped bool) {
		fmt.Printf("  left f=%d owner=%d dropped=%v\n", st.CurrentFrame, owner, dropped)
	}
	if err := f.LoadReplayData(loadable); err != nil {
		panic(err)
	}
	rs := f.ReplayState()
	as := f.ActionState()
	fmt.Println("gametype", rs.GameType, "allies", rs.AlliesEnabled, "teams", rs.TeamsEnabled)
	for i := range 12 {
		p := st.Players[i]
		fmt.Printf("player %d name=%q ctrl=%d race=%d force=%d playerID=%d start=%v color=%d\n", i, rs.PlayerName[i], p.Controller, p.Race, p.Force, as.PlayerID[i], st.Game.StartLocations[i], p.Color)
	}
	for u := range st.VisibleUnits.All() {
		if u.Owner < 8 {
			fmt.Printf("unit idx=%d owner=%d type=%v pos=%v\n", u.Index+1, u.Owner, u.UnitType.ID, u.Position)
		}
	}
	for !f.IsDone() && int(st.CurrentFrame) <= *to {
		if *stats > 0 && int(st.CurrentFrame)%*stats == 0 {
			fmt.Printf("stats f=%d units=%d bullets=%d sprites=%d images=%d orders=%d activeOrders=%d thingies=%d\n", st.CurrentFrame,
				st.UnitsContainer.Size, st.BulletsContainer.Size, st.SpritesContainer.Size, st.ImagesContainer.Size, st.OrdersContainer.Size, st.ActiveOrdersSize, st.ActiveThingiesSize)
		}
		if err := f.NextFrame(); err != nil {
			panic(err)
		}
	}
}
