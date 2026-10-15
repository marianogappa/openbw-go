package openbw

import (
	"slices"
	"sync"
	"testing"

	"github.com/marianogappa/openbw-go/bwgame"
)

type factCounts struct {
	archonsCompleted  int
	refineriesLinked  int
	recallsChecked    int
	interceptorKills  int
	reaverKills       int
	cancelledResearch int
	cancelledUnits    int
}

func (c *factCounts) add(other *factCounts) {
	c.archonsCompleted += other.archonsCompleted
	c.refineriesLinked += other.refineriesLinked
	c.recallsChecked += other.recallsChecked
	c.interceptorKills += other.interceptorKills
	c.reaverKills += other.reaverKills
	c.cancelledResearch += other.cancelledResearch
	c.cancelledUnits += other.cancelledUnits
}

func isArchon(t UnitType) bool {
	return t == UnitType(bwgame.UnitTypesProtossArchon) || t == UnitType(bwgame.UnitTypesProtossDarkArchon)
}

func isGasBuilding(t UnitType) bool {
	switch t {
	case UnitType(bwgame.UnitTypesTerranRefinery), UnitType(bwgame.UnitTypesProtossAssimilator), UnitType(bwgame.UnitTypesZergExtractor):
		return true
	}
	return false
}

type factChecker struct {
	t             *testing.T
	counts        factCounts
	archons       map[UnitID]bool
	refineryStart map[UnitID]int32
	lastSelection map[int8][]UnitID
	selectionSet  map[int8]int32
}

func (c *factChecker) checkKill(e *UnitDestroyed) {
	if e.Reason == ReasonOther {
		c.t.Errorf("frame %d: %v died with ReasonOther", e.Frame, e.Type)
	}
	if e.Killer == 0 && e.KillerType != UnitTypeNone {
		c.t.Errorf("frame %d: no killer but KillerType %v", e.Frame, e.KillerType)
	}
	if e.KillerType == UnitType(bwgame.UnitTypesProtossScarab) || e.KillerType == UnitType(bwgame.UnitTypesTerranVultureSpiderMine) {
		c.t.Errorf("frame %d: killer type %v", e.Frame, e.KillerType)
	}
	if (e.KillerParent != 0) != (e.KillerType == UnitType(bwgame.UnitTypesProtossInterceptor)) {
		c.t.Errorf("frame %d: killer %v with parent %d", e.Frame, e.KillerType, e.KillerParent)
	}
	switch e.KillerType {
	case UnitType(bwgame.UnitTypesProtossInterceptor):
		c.counts.interceptorKills++
	case UnitType(bwgame.UnitTypesProtossReaver):
		c.counts.reaverKills++
	}
}

func (c *factChecker) checkCancel(e *ProductionCancelled) {
	research := e.What == ProductionResearch || e.What == ProductionUpgrade
	if research != e.Research.Valid() || research != (e.Type == UnitTypeNone) {
		c.t.Errorf("frame %d: cancelled %v with type %v and research %v", e.Frame, e.What, e.Type, e.Research)
	}
	if research {
		c.counts.cancelledResearch++
	} else {
		c.counts.cancelledUnits++
	}
}

func (c *factChecker) checkCommand(r *CommandResult) {
	switch r.ActionID {
	case actionHotkey, actionSelect, actionShiftSelect, actionDeselect:
		c.lastSelection[r.Owner] = r.Selection
		c.selectionSet[r.Owner] = r.Frame
		return
	}
	if frame, ok := c.selectionSet[r.Owner]; ok && frame == r.Frame {
		if !slices.Equal(c.lastSelection[r.Owner], r.Selection) {
			c.t.Errorf("frame %d: player %d command %#x acts on %v after selection %v", r.Frame, r.Owner, r.ActionID, r.Selection, c.lastSelection[r.Owner])
		}
		c.counts.recallsChecked++
	}
}

func (c *factChecker) check(e Event) {
	switch e := e.(type) {
	case *UnitMorphed:
		if isArchon(e.To) {
			c.archons[e.Unit] = true
		}
		if isGasBuilding(e.To) && e.From == UnitType(bwgame.UnitTypesResourceVespeneGeyser) {
			c.refineryStart[e.Unit] = e.Frame
		}
	case *MorphStarted:
		if isGasBuilding(e.Into) {
			if frame, ok := c.refineryStart[e.Unit]; !ok || frame != e.Frame || e.Cmd < 0 {
				c.t.Errorf("frame %d: %v MorphStarted with Cmd %d", e.Frame, e.Into, e.Cmd)
			}
			delete(c.refineryStart, e.Unit)
			c.counts.refineriesLinked++
		}
	case *UnitCompleted:
		if isArchon(e.Type) {
			if !c.archons[e.Unit] {
				c.t.Errorf("frame %d: %v completed without a morph", e.Frame, e.Type)
			}
			delete(c.archons, e.Unit)
			c.counts.archonsCompleted++
		}
	case *UnitDestroyed:
		delete(c.archons, e.Unit)
		delete(c.refineryStart, e.Unit)
		c.checkKill(e)
	case *ProductionCancelled:
		c.checkCancel(e)
	case *CommandResult:
		c.checkCommand(e)
	}
}

func TestEventFacts(t *testing.T) {
	data := testData(t)
	var total factCounts
	var mu sync.Mutex
	t.Run("replays", func(t *testing.T) {
		forEachReplay(t, func(t *testing.T, replay []byte) {
			g, err := Open(replay, Options{Data: data, Kinds: AllKinds &^ KindTick})
			if err != nil {
				t.Fatal(err)
			}
			c := &factChecker{
				t: t, archons: map[UnitID]bool{}, refineryStart: map[UnitID]int32{},
				lastSelection: map[int8][]UnitID{}, selectionSet: map[int8]int32{},
			}
			for e := range g.Events() {
				c.check(e)
			}
			if err := g.Err(); err != nil {
				t.Fatal(err)
			}
			if len(c.refineryStart) > 0 {
				t.Errorf("refineries started without MorphStarted: %v", c.refineryStart)
			}
			mu.Lock()
			total.add(&c.counts)
			mu.Unlock()
		})
	})
	if total.archonsCompleted == 0 || total.refineriesLinked == 0 || total.recallsChecked == 0 ||
		total.interceptorKills+total.reaverKills == 0 || total.cancelledResearch+total.cancelledUnits == 0 {
		t.Errorf("too few facts checked: %+v", total)
	}
	t.Logf("%+v", total)
}

func TestHotkeyRecallSelection(t *testing.T) {
	data := testData(t)
	g, err := Open(readReplay(t, testReplays[0]), Options{Data: data, Kinds: KindCommandResult, MaxFrame: 24 * 60 * 2})
	if err != nil {
		t.Fatal(err)
	}
	recalls := 0
	g.funcs.AfterAction = func(owner, actionID int32, ok bool) {
		r := g.cmd.current
		g.afterAction(owner, actionID, ok)
		if r != nil && actionID == actionHotkey && ok {
			recalls++
			if !slices.Equal(r.Selection, g.selection(owner)) {
				t.Errorf("frame %d: hotkey selection %v, engine %v", r.Frame, r.Selection, g.selection(owner))
			}
		}
	}
	for range g.Events() {
	}
	if err := g.Err(); err != nil {
		t.Fatal(err)
	}
	if recalls == 0 {
		t.Fatal("no hotkey commands")
	}
}
