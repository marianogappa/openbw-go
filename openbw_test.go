package openbw

import (
	"bytes"
	"encoding/json"
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
	cwalReplays  = "/Users/mariano.gappa/Code/go/src/github.com/marianogappa/cwalggdl/replays-cwal-dl/"
	otherReplays = "/Users/mariano.gappa/Code/go/src/github.com/marianogappa/scmapanalyzer/replays/"
)

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
	replay, err := os.ReadFile(path)
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
	st := bwgame.NewState(data.Global(), bwgame.NewGameState())
	st.Game.Rules = chooseRules(StreamOptions{}, prepared)
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
		g, err := Open(replay, StreamOptions{
			Data: data, Kinds: AllKinds, Classes: AllClasses, TickEvery: 24, MaxFrame: testMaxFrame,
			Hooks: bwgame.Hooks{OnUnitDamaged: func(u, source *bwgame.Unit, owner int32, damage bwgame.Fp8) { damaged++ }},
		})
		if err != nil {
			t.Fatal(err)
		}
		c := &lifeChecker{t: t, created: map[UnitID]bool{}, destroyed: map[UnitID]bool{}}
		counts := map[Kind]int{}
		var lastAction int32 = -1
		results, linked := 0, 0
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
				if e.ScrepIndex >= 0 {
					linked++
					checkScrepLink(t, g, e)
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
		if g.prepared.Report.SourceFormat != "legacy" && linked < results*99/100 {
			t.Errorf("%d of %d command results linked to screp", linked, results)
		}
		if got, want := bwgame.StateHash(g.st), plainEngineHash(t, data, replay, g.Frame()); got != want {
			t.Errorf("state hash %x with the stream, %x without", got, want)
		}
		t.Logf("frame %d, results %d (linked %d), counts %v", g.Frame(), results, linked, counts)
	})
}

func checkScrepLink(t *testing.T, g *Game, r *CommandResult) {
	t.Helper()
	cmds := g.replay.Commands.Cmds
	base := cmds[r.ScrepIndex].BaseCmd()
	var source *downgrade.SourceCommand
	for i := range g.prepared.Commands {
		if g.prepared.Commands[i].Action == r.Action {
			source = &g.prepared.Commands[i]
			break
		}
	}
	if source == nil || int32(base.Frame) != r.Frame || base.PlayerID != r.PlayerID || !screpIDMatches(source.ID, base.Type.ID) {
		t.Fatalf("action %d at frame %d player %d links to screp command %d at frame %d player %d type %#x",
			r.Action, r.Frame, r.PlayerID, r.ScrepIndex, base.Frame, base.PlayerID, base.Type.ID)
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
	opts := StreamOptions{Data: data, Kinds: AllKinds, TickEvery: 100, MaxFrame: 24 * 60 * 3}
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
	g, err := Open(readReplay(t, testReplays[0]), StreamOptions{Data: data, MaxFrame: 10})
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

func TestRunJSONIsDeterministicAndRoundTrips(t *testing.T) {
	data := testData(t)
	replay := readReplay(t, testReplays[0])
	opts := Options{
		StreamOptions:  StreamOptions{Data: data, MaxFrame: 24 * 60 * 5},
		CommandResults: ResultsAll,
		UnitSnapshots:  &UnitSnapshotOptions{Every: 240, Fields: Position | HP | State | Order},
	}
	first, err := Run(replay, opts)
	if err != nil {
		t.Fatal(err)
	}
	second, err := Run(replay, opts)
	if err != nil {
		t.Fatal(err)
	}
	a, _ := json.Marshal(first)
	b, _ := json.Marshal(second)
	if !bytes.Equal(a, b) {
		t.Fatal("two runs give different JSON")
	}
	var decoded Result
	if err := json.Unmarshal(a, &decoded); err != nil {
		t.Fatal(err)
	}
	c, err := json.Marshal(&decoded)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(a, c) {
		t.Fatalf("JSON does not round-trip: %d bytes, %d after decoding", len(a), len(c))
	}
	if len(decoded.Events) == 0 || len(decoded.CommandResults) == 0 || len(decoded.Stats) == 0 || len(decoded.UnitSnapshots) == 0 {
		t.Fatal("empty parts in the result")
	}
}

func TestRunDefaults(t *testing.T) {
	data := testData(t)
	res, err := Run(readReplay(t, testReplays[0]), Options{StreamOptions: StreamOptions{Data: data}})
	if err != nil {
		t.Fatal(err)
	}
	if !res.Engine.ReachedEnd || res.Engine.DataHash == "" || res.Header == nil || res.Computed == nil {
		t.Fatalf("engine info %+v", res.Engine)
	}
	produced := 0
	for _, r := range res.CommandResults {
		if isSelectAction(r.ActionID) && r.MissingTags+r.ForeignTags == 0 {
			t.Fatalf("select result kept by default: %+v", r)
		}
		produced += len(r.Produced)
	}
	for _, e := range res.Events {
		if e.Kind() == KindCommandResult || e.Kind() == KindTick {
			t.Fatalf("%T in Events", e)
		}
		if u, ok := e.(*UnitCreated); ok && (u.Class == ClassFighter || u.Class == ClassSpell) {
			t.Fatalf("class %v in default events", u.Class)
		}
	}
	if produced == 0 || len(res.Stats) == 0 || len(res.Trust.Production) != 2 || res.Trust.NotExecuted != 0 {
		t.Fatalf("produced %d, stats %d, trust %+v", produced, len(res.Stats), res.Trust)
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

func TestDesyncSignalMatchesTagTest(t *testing.T) {
	data := testData(t)
	forEachReplay(t, func(t *testing.T, replay []byte) {
		res, err := Run(replay, Options{StreamOptions: StreamOptions{Data: data, MaxFrame: testMaxFrame}, StatsEvery: -1})
		if err != nil {
			t.Fatal(err)
		}
		prepared, err := downgrade.Prepare(replay)
		if err != nil {
			t.Fatal(err)
		}
		st := bwgame.NewState(data.Global(), bwgame.NewGameState())
		st.Game.Rules = chooseRules(StreamOptions{}, prepared)
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
		got := res.Trust.Desync
		if got.InconsistentSelects != c.inconsistent || got.FirstFrame != c.first {
			t.Fatalf("desync signal %+v, tag test %d first %d", got, c.inconsistent, c.first)
		}
	})
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
			_, errs[i] = Run(replay, Options{StreamOptions: StreamOptions{Data: data, MaxFrame: 24 * 60 * 3}})
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
	if !id.MatchesSCRTag(1705|37<<13) || id.MatchesSCRTag(1706|37<<13) || id.MatchesSCRTag(5|37<<13) {
		t.Fatal("SC:R tag match")
	}
}

func BenchmarkPlainEngine(b *testing.B) {
	data := testData(b)
	replay := readReplay(b, testReplays[0])
	prepared, err := downgrade.Prepare(replay)
	if err != nil {
		b.Fatal(err)
	}
	for b.Loop() {
		st := bwgame.NewState(data.Global(), bwgame.NewGameState())
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
	}
}

func BenchmarkStreamAllKinds(b *testing.B) {
	data := testData(b)
	replay := readReplay(b, testReplays[0])
	for b.Loop() {
		g, err := Open(replay, StreamOptions{Data: data, Kinds: AllKinds, Classes: AllClasses})
		if err != nil {
			b.Fatal(err)
		}
		for range g.Events() {
		}
	}
}

func BenchmarkRun(b *testing.B) {
	data := testData(b)
	replay := readReplay(b, testReplays[0])
	for b.Loop() {
		if _, err := Run(replay, Options{StreamOptions: StreamOptions{Data: data}}); err != nil {
			b.Fatal(err)
		}
	}
}
