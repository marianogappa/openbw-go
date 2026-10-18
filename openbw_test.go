package openbw

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/marianogappa/openbw-go/bwgame"
	"github.com/marianogappa/openbw-go/downgrade"
	"github.com/marianogappa/openbw-go/scrrules"
)

const (
	cwalReplays  = "cwalggdl/replays-cwal-dl/"
	otherReplays = "scmapanalyzer/replays/"
)

func replaysDir() string {
	if dir := os.Getenv("OPENBW_REPLAYS"); dir != "" {
		return dir
	}
	return "testdata/replays"
}

var testReplays = []string{
	cwalReplays + "30-NEW-lillljilililili/MM-CB157510-11C2-11F1-A3DD-FA167B5461B6.rep",
	cwalReplays + "30-MORE-SST_JumJaJungJi/MM-AAA58B20-12F9-11F1-A3DD-FA167B5461B6.rep",
	cwalReplays + "30-Esccccccc/MM-4862526E-25E0-11F1-A3DD-FA167B5461B6.rep",
	cwalReplays + "30-NEW-Horang2[._.]/MM-39C48C10-35DD-11F1-A3DD-FA167B5461B6.rep",
	otherReplays + "2v2v2v2 BGH !! Favourite!!!.rep",
	otherReplays + "2486 LimperT qeywro78Z.rep",
}

var testDataOnce = sync.OnceValues(func() (*Data, error) {
	dir := os.Getenv("OPENBW_DATA")
	if dir == "" {
		dir = "testdata/bwdata"
	}
	if _, err := os.Stat(filepath.Join(dir, "arr", "units.dat")); err != nil {
		return nil, err
	}
	return LoadData(dir)
})

func testData(t testing.TB) *Data {
	t.Helper()
	data, err := testDataOnce()
	if err != nil {
		t.Skipf("no game data: %v", err)
	}
	return data
}

func readReplay(t testing.TB, path string) []byte {
	t.Helper()
	replay, err := os.ReadFile(filepath.Join(replaysDir(), path))
	if err != nil {
		t.Skipf("no replay: %v", err)
	}
	return replay
}

func forEachReplay(t *testing.T, test func(t *testing.T, replay []byte)) {
	for _, path := range testReplays {
		t.Run(filepath.Base(path), func(t *testing.T) {
			t.Parallel()
			test(t, readReplay(t, path))
		})
	}
}

const testMaxFrame = 24 * 60 * 16

type lifeChecker struct {
	t         *testing.T
	created   map[UnitID]bool
	destroyed map[UnitID]bool
	lastSeq   int64
	lastFrame int32
	errors    int
}

func (c *lifeChecker) fail(format string, args ...any) {
	if c.errors < 20 {
		c.t.Errorf(format, args...)
	}
	c.errors++
}

func (c *lifeChecker) alive(id UnitID, e Event) {
	if !c.created[id] {
		c.fail("%T at frame %d for unit %d before UnitCreated", e, e.Base().Frame, id)
	}
	if c.destroyed[id] {
		c.fail("%T at frame %d for unit %d after UnitDestroyed", e, e.Base().Frame, id)
	}
}

func (c *lifeChecker) check(e Event) {
	b := e.Base()
	if b.Seq <= c.lastSeq || b.Frame < c.lastFrame {
		c.fail("event %T seq %d frame %d after seq %d frame %d", e, b.Seq, b.Frame, c.lastSeq, c.lastFrame)
	}
	c.lastSeq, c.lastFrame = b.Seq, b.Frame
	switch e := e.(type) {
	case *UnitCreated:
		if c.created[e.Unit] {
			c.fail("UnitCreated twice for %d", e.Unit)
		}
		c.created[e.Unit] = true
	case *UnitCompleted:
		c.alive(e.Unit, e)
	case *UnitMorphed:
		c.alive(e.Unit, e)
	case *MorphStarted:
		c.alive(e.Unit, e)
	case *UnitOwnerChanged:
		c.alive(e.Unit, e)
	case *UnitDestroyed:
		c.alive(e.Unit, e)
		if !e.Reverted {
			if c.destroyed[e.Unit] {
				c.fail("UnitDestroyed twice for %d", e.Unit)
			}
			c.destroyed[e.Unit] = true
		}
	}
}

func checkShadowQueues(t *testing.T, g *Game) {
	t.Helper()
	for owner := range g.st.PlayerUnits {
		for u := range g.st.PlayerUnits[owner].All() {
			m := g.cmd.queues[u.Index]
			if len(m.queue) != u.BuildQueue.Size() || len(m.limbo) != u.BuildQueueLimbo.Size() {
				t.Fatalf("frame %d: shadow queue of unit %d is %d+%d, engine %d+%d", g.Frame(), u.Index,
					len(m.queue), len(m.limbo), u.BuildQueue.Size(), u.BuildQueueLimbo.Size())
			}
		}
	}
}

func plainEngineHash(t *testing.T, data *Data, replay []byte, frame int32) uint64 {
	t.Helper()
	prepared, err := downgrade.Prepare(replay)
	if err != nil {
		t.Fatal(err)
	}
	st := bwgame.NewState(data.global, bwgame.NewGameState())
	st.Game.Rules, _ = chooseRules(Options{}, prepared)
	funcs := bwgame.NewReplayFunctions(st, &bwgame.ActionState{}, &bwgame.ReplayState{})
	if err := funcs.LoadReplayData(prepared.Loadable); err != nil {
		t.Fatal(err)
	}
	for st.CurrentFrame < frame {
		if err := funcs.NextFrame(); err != nil {
			t.Fatal(err)
		}
	}
	return bwgame.StateHash(st)
}

func TestStreamInvariants(t *testing.T) {
	data := testData(t)
	forEachReplay(t, func(t *testing.T, replay []byte) {
		damaged := 0
		g, err := Open(replay, Options{
			Data: data, Kinds: AllKinds, Classes: AllClasses, TickEvery: 24, MaxFrame: testMaxFrame,
			Engine: &EngineOptions{Hooks: bwgame.Hooks{OnUnitDamaged: func(u, source *bwgame.Unit, owner int32, damage bwgame.Fp8) { damaged++ }}},
		})
		if err != nil {
			t.Fatal(err)
		}
		sources := g.SourceCommands()
		c := &lifeChecker{t: t, created: map[UnitID]bool{}, destroyed: map[UnitID]bool{}}
		counts := map[Kind]int{}
		var lastAction int32 = -1
		results, sourced := 0, 0
		for e := range g.Events() {
			c.check(e)
			counts[e.Kind()]++
			switch e := e.(type) {
			case *CommandResult:
				if e.Cmd != e.Action || e.Action != lastAction+1 {
					t.Fatalf("command result %d with Cmd %d after action %d", e.Action, e.Cmd, lastAction)
				}
				lastAction = e.Action
				results++
				if e.Source >= 0 {
					sourced++
					checkSource(t, g, sources, e)
				}
			case *Tick:
				checkShadowQueues(t, g)
			default:
				if e.Base().Cmd > lastAction+1 {
					t.Fatalf("%T linked to future command %d", e, e.Base().Cmd)
				}
			}
		}
		if err := g.Err(); err != nil {
			t.Fatal(err)
		}
		if g.Frame() != min(g.EndFrame(), testMaxFrame) {
			t.Errorf("stopped at frame %d, end %d", g.Frame(), g.EndFrame())
		}
		if counts[KindUnitCreated] == 0 || counts[KindUnitDestroyed] == 0 || counts[KindTick] == 0 || damaged == 0 {
			t.Errorf("too few events: %v, damaged %d", counts, damaged)
		}
		legacy := g.prepared.Report.SourceFormat == legacyFormat
		if info := g.Info(); info.Rules.SCR == legacy || info.Rules.Custom || info.DataHash == "" {
			t.Errorf("engine info %+v for source format %s", info, g.prepared.Report.SourceFormat)
		}
		if sourced != results {
			t.Errorf("%d of %d command results have a source command", sourced, results)
		}
		if got, want := bwgame.StateHash(g.st), plainEngineHash(t, data, replay, g.Frame()); got != want {
			t.Errorf("state hash %x with the stream, %x without", got, want)
		}
		t.Logf("frame %d, results %d, counts %v", g.Frame(), results, counts)
	})
}

func checkSource(t *testing.T, g *Game, sources []SourceCommand, r *CommandResult) {
	t.Helper()
	c := sources[r.Source]
	slot := g.Header().Slots[r.Owner]
	if c.Action != r.Action || c.Frame != r.Frame || int32(c.PlayerID) != slot.PlayerID || c.Dropped != "" {
		t.Fatalf("action %d at frame %d slot %d (player %d) has source %d: %+v", r.Action, r.Frame, r.Owner, slot.PlayerID, r.Source, c)
	}
}

func collectEvents(t *testing.T, g *Game, breakEvery int) []string {
	t.Helper()
	var out []string
	for !g.Done() {
		n := 0
		for e := range g.Events() {
			out = append(out, fmt.Sprintf("%d %d %T", e.Base().Seq, e.Base().Frame, e))
			n++
			if breakEvery > 0 && n == breakEvery {
				break
			}
		}
	}
	return out
}

func TestStreamResumesAfterBreak(t *testing.T) {
	data := testData(t)
	replay := readReplay(t, testReplays[0])
	opts := Options{Data: data, Kinds: AllKinds, TickEvery: 100, MaxFrame: 24 * 60 * 3}
	whole, err := Open(replay, opts)
	if err != nil {
		t.Fatal(err)
	}
	pieces, err := Open(replay, opts)
	if err != nil {
		t.Fatal(err)
	}
	a, b := collectEvents(t, whole, 0), collectEvents(t, pieces, 7)
	if len(a) == 0 || fmt.Sprint(a) != fmt.Sprint(b) {
		t.Fatalf("%d events in one loop, %d with breaks", len(a), len(b))
	}
	if whole.Frame() != opts.MaxFrame || pieces.Frame() != opts.MaxFrame {
		t.Fatalf("stopped at frames %d and %d, want %d", whole.Frame(), pieces.Frame(), opts.MaxFrame)
	}
}

func TestEventsInsideLoopPanics(t *testing.T) {
	data := testData(t)
	g, err := Open(readReplay(t, testReplays[0]), Options{Data: data, MaxFrame: 10})
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if recover() == nil {
			t.Fatal("no panic")
		}
	}()
	for range g.Events() {
		for range g.Events() {
		}
	}
}

type tagChecker struct {
	funcs        *bwgame.ReplayFunctions
	inconsistent int
	first        int32
}

func (c *tagChecker) beforeAction(owner, actionID int32, payload []byte) {
	if actionID != actionSelect && actionID != actionShiftSelect || len(payload) == 0 {
		return
	}
	n := int(payload[0])
	bad := 0
	for i := range n {
		id := bwgame.UnitID{RawValue: uint16(payload[1+2*i]) | uint16(payload[2+2*i])<<8}
		if id.Index() == 0 {
			continue
		}
		u := c.funcs.UnitByID(id)
		if u == nil || n > 1 && u.Owner != owner {
			bad++
		}
	}
	if bad > 0 {
		c.inconsistent++
		if c.first == 0 {
			c.first = c.funcs.State().CurrentFrame
		}
	}
}

func TestTagCountsMatchTagTest(t *testing.T) {
	data := testData(t)
	forEachReplay(t, func(t *testing.T, replay []byte) {
		g, err := Open(replay, Options{Data: data, Kinds: KindCommandResult, MaxFrame: testMaxFrame})
		if err != nil {
			t.Fatal(err)
		}
		inconsistent := 0
		var first int32
		for e := range g.Events() {
			r := e.(*CommandResult)
			if r.MissingTags+r.ForeignTags > 0 {
				inconsistent++
				if first == 0 {
					first = r.Frame
				}
			}
		}
		if err := g.Err(); err != nil {
			t.Fatal(err)
		}
		prepared, err := downgrade.Prepare(replay)
		if err != nil {
			t.Fatal(err)
		}
		st := bwgame.NewState(data.global, bwgame.NewGameState())
		st.Game.Rules, _ = chooseRules(Options{}, prepared)
		c := &tagChecker{funcs: bwgame.NewReplayFunctions(st, &bwgame.ActionState{}, &bwgame.ReplayState{})}
		c.funcs.BeforeAction = c.beforeAction
		if err := c.funcs.LoadReplayData(prepared.Loadable); err != nil {
			t.Fatal(err)
		}
		for !c.funcs.IsDone() && st.CurrentFrame < testMaxFrame {
			if err := c.funcs.NextFrame(); err != nil {
				t.Fatal(err)
			}
		}
		if inconsistent != c.inconsistent || first != c.first {
			t.Fatalf("%d results with bad tags, first at %d; tag test %d, first at %d", inconsistent, first, c.inconsistent, c.first)
		}
	})
}

func drain(replay []byte, opts Options) error {
	g, err := Open(replay, opts)
	if err != nil {
		return err
	}
	for range g.Events() {
	}
	return g.Err()
}

func TestGamesShareData(t *testing.T) {
	data := testData(t)
	var wg sync.WaitGroup
	errs := make([]error, 4)
	for i := range errs {
		replay := readReplay(t, testReplays[i])
		wg.Add(1)
		go func() {
			defer wg.Done()
			errs[i] = drain(replay, Options{Data: data, Kinds: AllKinds, MaxFrame: 24 * 60 * 3})
		}()
	}
	wg.Wait()
	for _, err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
}

func TestUnitIDTags(t *testing.T) {
	id := UnitID(5 | 37<<11)
	if id.Index() != 5 || id.Generation() != 37 || id.LegacyTag() != 5|5<<11 {
		t.Fatalf("index %d generation %d tag %#x", id.Index(), id.Generation(), id.LegacyTag())
	}
	scr := &Game{prepared: &downgrade.Prepared{Settings: downgrade.SCRSettings{Limits: downgrade.ObjectLimits{Units: 3400}}}}
	if !scr.MatchesSCRTag(id, 1705|37<<13) || scr.MatchesSCRTag(id, 1706|37<<13) || scr.MatchesSCRTag(id, 5|37<<13) {
		t.Fatal("SC:R tag match")
	}
	noLMTS := &Game{prepared: &downgrade.Prepared{Settings: downgrade.LegacySettings()}}
	early := UnitID(1448 | 1<<11)
	if !noLMTS.MatchesSCRTag(early, 0x0DA8) || noLMTS.MatchesSCRTag(early, 0x0DA9) || scr.MatchesSCRTag(early, 0x0DA8) {
		t.Fatal("SC:R tag match without LMTS")
	}
}

func BenchmarkPlainEngine(b *testing.B) {
	data := testData(b)
	replay := readReplay(b, testReplays[0])
	b.ReportAllocs()
	var frames int32
	for b.Loop() {
		prepared, err := downgrade.Prepare(replay)
		if err != nil {
			b.Fatal(err)
		}
		st := bwgame.NewState(data.global, bwgame.NewGameState())
		st.Game.Rules = scrrules.FromSettings(prepared.Settings)
		funcs := bwgame.NewReplayFunctions(st, &bwgame.ActionState{}, &bwgame.ReplayState{})
		if err := funcs.LoadReplayData(prepared.Loadable); err != nil {
			b.Fatal(err)
		}
		for !funcs.IsDone() {
			if err := funcs.NextFrame(); err != nil {
				b.Fatal(err)
			}
		}
		frames = st.CurrentFrame
	}
	b.ReportMetric(float64(frames), "frames/op")
}

func benchmarkStream(b *testing.B, kinds Kind) {
	data := testData(b)
	replay := readReplay(b, testReplays[0])
	b.ReportAllocs()
	var frames int32
	events := 0
	for b.Loop() {
		g, err := Open(replay, Options{Data: data, Kinds: kinds, Classes: AllClasses})
		if err != nil {
			b.Fatal(err)
		}
		events = 0
		for range g.Events() {
			events++
		}
		if err := g.Err(); err != nil {
			b.Fatal(err)
		}
		frames = g.Frame()
	}
	b.ReportMetric(float64(frames), "frames/op")
	b.ReportMetric(float64(events), "events/op")
}

func BenchmarkStreamNoKinds(b *testing.B) { benchmarkStream(b, NoKinds) }

func BenchmarkStreamDefaultKinds(b *testing.B) { benchmarkStream(b, DefaultKinds) }

func BenchmarkStreamAllKinds(b *testing.B) { benchmarkStream(b, AllKinds) }

func BenchmarkOpen(b *testing.B) {
	data := testData(b)
	replay := readReplay(b, testReplays[0])
	b.ReportAllocs()
	for b.Loop() {
		if _, err := Open(replay, Options{Data: data}); err != nil {
			b.Fatal(err)
		}
	}
}
