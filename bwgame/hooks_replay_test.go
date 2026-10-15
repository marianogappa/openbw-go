package bwgame

import (
	"fmt"
	"os"
	"slices"
	"testing"

	"github.com/marianogappa/openbw-go/downgrade"
)

var hookTestReplays = []string{
	testReplayPath("cwalggdl/replays-cwal-dl/30-NEW-Horang2[._.]/MM-39C48C10-35DD-11F1-A3DD-FA167B5461B6.rep"),
	testReplayPath("cwalggdl/replays-cwal-dl/30-NEW-lillljilililili/MM-491926A2-11BC-11F1-A3DD-FA167B5461B6.rep"),
	testReplayPath("cwalggdl/replays-cwal-dl/30-NEW-lillljilililili/MM-CB157510-11C2-11F1-A3DD-FA167B5461B6.rep"),
	testReplayPath("cwalggdl/replays-cwal-dl/30-MORE-SST_JumJaJungJi/MM-AAA58B20-12F9-11F1-A3DD-FA167B5461B6.rep"),
	testReplayPath("scmapanalyzer/replays/2v2v2v2 BGH !! Favourite!!!.rep"),
	compressedLegacyReplayPath,
}

const hookTestMaxFrame = 24 * 60 * 16

type hookUnitKey struct {
	index      uint
	generation uint32
}

func hookKeyOf(u *Unit) hookUnitKey {
	return hookUnitKey{index: uint(u.Index), generation: u.UnitIDGeneration}
}

type hookUnitLife struct {
	created      bool
	destroyed    bool
	completedNow bool
}

type mirroredQueue struct {
	queue []*UnitTypeT
	limbo []*UnitTypeT
}

type hookRecorder struct {
	st     *State
	units  map[hookUnitKey]*hookUnitLife
	queues map[*Unit]*mirroredQueue

	events       int
	kills        map[DeathReason]int
	created      map[CreateSource]int
	cancels      int
	research     int
	rejections   int
	morphStarts  int
	ownerChanges int

	inAction       bool
	actionsStarted int
	actionsEnded   int
	actionCalls    []hookAction
	violations     []string
}

type hookAction struct {
	owner    int32
	actionID int32
}

func newHookRecorder() *hookRecorder {
	return &hookRecorder{
		units:   map[hookUnitKey]*hookUnitLife{},
		queues:  map[*Unit]*mirroredQueue{},
		kills:   map[DeathReason]int{},
		created: map[CreateSource]int{},
	}
}

func (r *hookRecorder) fail(format string, args ...any) {
	if len(r.violations) < 20 {
		r.violations = append(r.violations, fmt.Sprintf("frame %d: ", r.st.CurrentFrame)+fmt.Sprintf(format, args...))
	}
}

func (r *hookRecorder) life(u *Unit, event string) *hookUnitLife {
	l := r.units[hookKeyOf(u)]
	if l == nil || !l.created {
		r.fail("%s for unit %d (type %d) before UnitCreated", event, u.Index, u.UnitType.ID)
		return &hookUnitLife{}
	}
	if l.destroyed && !isRefineryType(u.UnitType) && !isGeyserType(u.UnitType) {
		r.fail("%s for unit %d (type %d) after UnitDestroyed", event, u.Index, u.UnitType.ID)
	}
	return l
}

func isRefineryType(t *UnitTypeT) bool {
	return t.ID == UnitTypesTerranRefinery || t.ID == UnitTypesZergExtractor || t.ID == UnitTypesProtossAssimilator
}

func isGeyserType(t *UnitTypeT) bool {
	return t.ID == UnitTypesResourceVespeneGeyser
}

func (r *hookRecorder) hooks() Hooks {
	return Hooks{
		OnUnitCreated: func(u *Unit, source CreateSource, parent *Unit) {
			r.events++
			r.created[source]++
			if (&StateFunctions{}).utTurret(u.UnitType) {
				r.fail("UnitCreated for turret %d", u.Index)
			}
			if l := r.units[hookKeyOf(u)]; l != nil && l.created {
				r.fail("UnitCreated twice for unit %d generation %d", u.Index, u.UnitIDGeneration)
			}
			r.units[hookKeyOf(u)] = &hookUnitLife{created: true}
		},
		OnUnitCompleted: func(u *Unit) {
			r.events++
			l := r.life(u, "UnitCompleted")
			if l.completedNow {
				r.fail("UnitCompleted twice for unit %d (type %d) without a morph in between", u.Index, u.UnitType.ID)
			}
			l.completedNow = true
		},
		OnUnitMorphed: func(u *Unit, from *UnitTypeT) {
			r.events++
			r.life(u, "UnitMorphed").completedNow = false
		},
		OnMorphStarted: func(u, morpher *Unit, from, into *UnitTypeT) {
			r.events++
			r.morphStarts++
			r.life(u, "MorphStarted").completedNow = false
			if morpher != u {
				r.life(morpher, "MorphStarted morpher")
			}
			if from == nil || into == nil {
				r.fail("MorphStarted without types")
			}
		},
		OnKillUnit: func(u *Unit, kill KillContext) {
			r.events++
			r.kills[kill.Reason]++
			l := r.life(u, "UnitDestroyed")
			if l.destroyed && !isRefineryType(u.UnitType) {
				r.fail("UnitDestroyed twice for unit %d (type %d)", u.Index, u.UnitType.ID)
			}
			if kill.Reason == ReasonContainer && kill.Container == nil {
				r.fail("container death without container")
			}
			if !isRefineryType(u.UnitType) {
				l.destroyed = true
			}
		},
		OnOwnerChanged: func(u *Unit, from int32, reason OwnerReason) {
			r.events++
			r.ownerChanges++
			if from == u.Owner {
				r.fail("UnitOwnerChanged without a change")
			}
			if l := r.units[hookKeyOf(u)]; l == nil || !l.created {
				r.fail("UnitOwnerChanged for unit %d before UnitCreated", u.Index)
			}
		},
		OnProductionCancel: func(c ProductionCancel) {
			r.events++
			r.cancels++
			if c.Why == CancelCommand && !r.inAction {
				r.fail("cancel with reason Command outside a command")
			}
			if (c.What == ProductionResearch || c.What == ProductionUpgrade) != (c.Type == nil) {
				r.fail("cancel %d with type %v", c.What, c.Type)
			}
		},
		OnResearch: func(u *Unit, res Research, level int32, completed bool) {
			r.events++
			r.research++
			if !completed && !r.inAction {
				r.fail("research started outside a command")
			}
			if level < 1 {
				r.fail("research level %d", level)
			}
		},
		OnPlayerLeft: func(owner int32, dropped bool) {
			r.events++
			if !r.inAction {
				r.fail("player left outside a command")
			}
		},
		OnActionRejected: func(u *Unit, reason RejectReason) {
			r.rejections++
			if !r.inAction {
				r.fail("rejection outside a command")
			}
		},
		OnBuildQueue: r.mirrorBuildQueue,
	}
}

func (r *hookRecorder) mirrorBuildQueue(u *Unit, op BuildQueueOp, index int) {
	m := r.queues[u]
	if m == nil {
		m = &mirroredQueue{}
		r.queues[u] = m
	}
	switch op {
	case BuildQueuePush:
		if index != len(m.queue) {
			r.fail("push at %d into a mirror of %d", index, len(m.queue))
			return
		}
		m.queue = append(m.queue, u.BuildQueue.At(index))
		m.queue = append(m.queue, m.limbo...)
		m.limbo = nil
	case BuildQueueClearSingle:
		m.limbo = slices.Clone(m.queue[1:])
		m.queue = nil
	case BuildQueueErase:
		m.queue = slices.Delete(m.queue, index, index+1)
	case BuildQueueReset:
		m.queue, m.limbo = nil, nil
	}
}

func (r *hookRecorder) checkBuildQueues() {
	for owner := range r.st.PlayerUnits {
		for u := range r.st.PlayerUnits[owner].All() {
			m := r.queues[u]
			if m == nil {
				m = &mirroredQueue{}
			}
			if !slices.Equal(m.queue, u.BuildQueue.Values()) || !slices.Equal(m.limbo, u.BuildQueueLimbo.Values()) {
				r.fail("build queue mirror of unit %d (type %d) is %d+%d, engine %d+%d", u.Index, u.UnitType.ID, len(m.queue), len(m.limbo), u.BuildQueue.Size(), u.BuildQueueLimbo.Size())
			}
		}
	}
}

type hookTestGame struct {
	st    *State
	funcs *ReplayFunctions
}

func newHookTestGame(t *testing.T, global *GlobalState, loadable []byte, hooks Hooks, configure func(*ReplayFunctions)) *hookTestGame {
	t.Helper()
	st := NewState(global, NewGameState())
	funcs := NewReplayFunctions(st, &ActionState{}, &ReplayState{})
	funcs.Hooks = hooks
	if configure != nil {
		configure(funcs)
	}
	if err := funcs.LoadReplayData(loadable); err != nil {
		t.Fatal(err)
	}
	return &hookTestGame{st: st, funcs: funcs}
}

func TestHooksDoNotChangeTheGameAndKeepInvariants(t *testing.T) {
	global := testGlobalState(t)
	for _, path := range hookTestReplays {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Logf("skip %s: %v", path, err)
			continue
		}
		t.Run(shortReplayName(path), func(t *testing.T) {
			t.Parallel()
			checkHookedReplay(t, global, data)
		})
	}
}

func shortReplayName(path string) string {
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' {
			return path[i+1:]
		}
	}
	return path
}

func checkHookedReplay(t *testing.T, global *GlobalState, data []byte) {
	prepared, err := downgrade.Prepare(data)
	if err != nil {
		t.Fatal(err)
	}
	plain := newHookTestGame(t, global, prepared.Loadable, Hooks{}, nil)
	rec := newHookRecorder()
	hooked := newHookTestGame(t, global, prepared.Loadable, rec.hooks(), func(f *ReplayFunctions) {
		rec.st = f.State()
		f.BeforeAction = func(owner, actionID int32, payload []byte) {
			rec.inAction = true
			rec.actionsStarted++
		}
		f.AfterAction = func(owner, actionID int32, ok bool) {
			rec.inAction = false
			rec.actionsEnded++
			rec.actionCalls = append(rec.actionCalls, hookAction{owner, actionID})
		}
	})
	if got, want := StateHash(hooked.st), StateHash(plain.st); got != want {
		t.Fatalf("state hash after load %x, want %x", got, want)
	}
	rec.checkBuildQueues()
	for !plain.funcs.IsDone() && plain.st.CurrentFrame < hookTestMaxFrame {
		plainErr := plain.funcs.NextFrame()
		hookedErr := hooked.funcs.NextFrame()
		if (plainErr == nil) != (hookedErr == nil) {
			t.Fatalf("frame %d: error %v with hooks, %v without", plain.st.CurrentFrame, hookedErr, plainErr)
		}
		if got, want := StateHash(hooked.st), StateHash(plain.st); got != want {
			t.Fatalf("frame %d: state hash %x with hooks, %x without", plain.st.CurrentFrame, got, want)
		}
		if plainErr != nil {
			t.Logf("engine error at frame %d: %v", plain.st.CurrentFrame, plainErr)
			break
		}
		rec.checkBuildQueues()
		if len(rec.violations) > 0 {
			break
		}
	}
	for _, v := range rec.violations {
		t.Error(v)
	}
	if rec.actionsStarted != rec.actionsEnded {
		t.Errorf("%d actions started, %d ended", rec.actionsStarted, rec.actionsEnded)
	}
	checkActionIndex(t, prepared, rec.actionCalls, hooked.funcs)
	if rec.created[SourceMap] == 0 || rec.kills[ReasonDamage] == 0 || rec.morphStarts+rec.created[SourceTrained] == 0 {
		t.Errorf("too few events: created %v, kills %v, morph starts %d", rec.created, rec.kills, rec.morphStarts)
	}
	t.Logf("frames %d, events %d, created %v, kills %v, owner changes %d, cancels %d, research %d, rejections %d",
		hooked.st.CurrentFrame, rec.events, rec.created, rec.kills, rec.ownerChanges, rec.cancels, rec.research, rec.rejections)
}

func checkActionIndex(t *testing.T, prepared *downgrade.Prepared, calls []hookAction, funcs *ReplayFunctions) {
	t.Helper()
	executed := 0
	for _, c := range prepared.Commands {
		if c.Action < 0 {
			continue
		}
		if int(c.Action) != executed && int(c.Action) < len(calls) {
			t.Fatalf("action index %d, want %d", c.Action, executed)
		}
		if int(c.Action) >= len(calls) {
			break
		}
		call := calls[c.Action]
		if want := legacyActionID(c.ID); call.actionID != int32(want) {
			t.Fatalf("action %d at frame %d: engine read action ID %#x, source command %#x", c.Action, c.Frame, call.actionID, c.ID)
		}
		if got := funcs.ActionState().PlayerID[call.owner]; got != int32(c.PlayerID) {
			t.Fatalf("action %d: engine owner %d has player ID %d, source command player ID %d", c.Action, call.owner, got, c.PlayerID)
		}
		executed++
	}
	if executed != len(calls) {
		t.Fatalf("%d actions matched, engine read %d", executed, len(calls))
	}
}

func legacyActionID(id byte) byte {
	if id >= 0x60 && id <= 0x65 {
		return [...]byte{0x14, 0x15, 0x29, 0x09, 0x0a, 0x0b}[id-0x60]
	}
	return id
}
