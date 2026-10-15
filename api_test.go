package openbw

import (
	"slices"
	"testing"

	"github.com/marianogappa/openbw-go/bwgame"
)

func TestNoKindsAndAdvanceTo(t *testing.T) {
	data := testData(t)
	replay := readReplay(t, testReplays[0])
	g, err := Open(replay, Options{Data: data, Kinds: NoKinds})
	if err != nil {
		t.Fatal(err)
	}
	if g.kinds != 0 || g.cmd != nil || g.funcs.Hooks.OnUnitCreated != nil {
		t.Fatalf("NoKinds subscribed kinds %v", g.kinds)
	}
	const frame = 24 * 60
	if err := g.AdvanceTo(frame); err != nil {
		t.Fatal(err)
	}
	if g.Frame() != frame || len(g.buf) != 0 {
		t.Fatalf("advanced to frame %d with %d buffered events", g.Frame(), len(g.buf))
	}
	if got, want := bwgame.StateHash(g.st), plainEngineHash(t, data, replay, frame); got != want {
		t.Fatalf("state hash %x after AdvanceTo, %x without", got, want)
	}
	for e := range g.Events() {
		t.Fatalf("NoKinds yielded %T", e)
	}
	if !g.Done() || g.Frame() != g.EndFrame() {
		t.Fatalf("stopped at frame %d of %d", g.Frame(), g.EndFrame())
	}

	opts := Options{Data: data, Kinds: AllKinds, TickEvery: 24, MaxFrame: 2 * frame}
	pulled, err := Open(replay, opts)
	if err != nil {
		t.Fatal(err)
	}
	advanced, err := Open(replay, opts)
	if err != nil {
		t.Fatal(err)
	}
	if err := advanced.AdvanceTo(frame); err != nil {
		t.Fatal(err)
	}
	a, b := collectEvents(t, pulled, 0), collectEvents(t, advanced, 0)
	if len(a) == 0 || !slices.Equal(a, b) {
		t.Fatalf("%d events pulled, %d after AdvanceTo", len(a), len(b))
	}
}

func TestOpenOptions(t *testing.T) {
	data := testData(t)
	replay := readReplay(t, testReplays[0])
	for _, opts := range []Options{{MaxFrame: -5}, {TickEvery: -5}} {
		opts.Data = data
		if _, err := Open(replay, opts); err == nil {
			t.Errorf("Open with %+v: no error", opts)
		}
	}
	g, err := Open(replay, Options{Data: data})
	if err != nil {
		t.Fatal(err)
	}
	if g.kinds != DefaultKinds || g.sources != nil {
		t.Fatalf("default game: kinds %v, %d sources", g.kinds, len(g.sources))
	}
	h := g.Header()
	if h.MapName == "" || h.EndFrame != g.EndFrame() {
		t.Errorf("header %+v, end frame %d", h, g.EndFrame())
	}
	for _, owner := range g.PlayingSlots() {
		if s := h.Slots[owner]; s.Owner != owner || s.PlayerID < 0 || s.Name == "" {
			t.Errorf("slot %d: %+v", owner, s)
		}
	}
	if p := g.Player(99); p != (Player{Owner: 99}) {
		t.Errorf("player 99: %+v", p)
	}
}
