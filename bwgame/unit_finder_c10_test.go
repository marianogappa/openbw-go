package bwgame

import (
	"fmt"
	"strings"
	"testing"
)

const unitFinderReferenceOutput = `
X 18:4 16:21 26:23 20:32 26:36 7:39 18:45 16:53 20:61 7:73 12:94 8:125 12:143 38:144 15:165 15:177 38:177 8:188 27:201 27:221 31:228 13:240 6:245 28:246 1:250 13:250 28:254 3:258 6:261 39:266 31:268 39:270 3:277 1:307 22:327 29:339 29:343 33:347 23:358 14:358 22:360 33:376 23:380 36:380 30:402 36:407 9:408 24:409 14:412 25:418 30:419 2:437 25:439 24:442 9:444 32:458 0:468 19:474 34:475 2:484 10:492 32:495 34:507 0:509 19:512 10:541
Y 6:5 19:19 6:31 23:52 23:78 19:82 3:87 12:88 20:97 38:103 36:111 3:119 30:130 14:130 12:130 32:131 36:131 16:143 15:146 38:146 26:155 30:159 20:159 15:160 32:162 16:166 14:189 26:208 31:239 2:246 2:251 29:253 29:260 34:264 9:265 31:286 34:290 7:301 9:302 7:312 10:333 10:339 28:350 1:356 22:375 1:376 28:379 33:384 25:391 8:394 8:405 33:410 22:416 13:418 13:426 25:438 24:453 18:470 39:471 18:474 24:475 27:486 0:488 39:506 0:510 27:528
S 8( 8 ) 13( 13 ) 28( 28 1 ) 1( 1 28 ) 22( 22 33 )
S 12 38 15 31 3
S 24 25
S 22 33 24 25 0
S 7
S
S 34 2 10
S
S
S
S 23( 23 ) 14( 14 36 30 ) 36( 36 30 14 ) 30( 30 36 14 ) 32( 32 ) 19( 19 )
S 8 27 13 39 22
S 29 14 36
S 12 38 3
S 3 14 36 30
S 9( 9 ) 25( 25 ) 2( 2 ) 34( 34 )
S 18
S 18 8
S 26 7 16
S
`

type c10TestRandom struct {
	state uint32
}

func (r *c10TestRandom) next(n int32) int32 {
	r.state = r.state*1103515245 + 12345
	return int32((r.state>>16)&0x7fff) % n
}

func (r *c10TestRandom) boundingBox() Rect {
	var bb Rect
	bb.From.X = r.next(500)
	bb.From.Y = r.next(500)
	bb.To.X = bb.From.X + 4 + r.next(60)
	bb.To.Y = bb.From.Y + 4 + r.next(60)
	return bb
}

func newC10TestFunctions() *StateFunctions {
	game := NewGameState()
	game.MaxUnitWidth = 64
	game.MaxUnitHeight = 64
	return NewStateFunctions(NewState(nil, game))
}

func unitFinderReferenceRun() string {
	f := newC10TestFunctions()
	r := &c10TestRandom{state: 12345}
	const n = 40
	units := make([]Unit, n)
	index := map[*Unit]int{}
	for i := range units {
		units[i].UnitFinderBoundingBox = Rect{From: Xy{X: -1, Y: -1}, To: Xy{X: -1, Y: -1}}
		index[&units[i]] = i
	}
	for i := range units {
		f.unitFinderInsertBB(&units[i], r.boundingBox())
	}
	for range 60 {
		u := &units[r.next(n)]
		f.unitFinderReinsertBB(u, r.boundingBox())
	}
	for range 10 {
		f.unitFinderRemove(&units[r.next(n)])
	}
	var out strings.Builder
	out.WriteString("X")
	for _, e := range f.st.UnitFinderX {
		fmt.Fprintf(&out, " %d:%d", index[e.U], e.Value)
	}
	out.WriteString("\nY")
	for _, e := range f.st.UnitFinderY {
		fmt.Fprintf(&out, " %d:%d", index[e.U], e.Value)
	}
	out.WriteString("\n")
	for i := range 20 {
		var a Rect
		a.From.X = r.next(520)
		a.From.Y = r.next(520)
		a.To.X = a.From.X + r.next(250)
		a.To.Y = a.From.Y + r.next(250)
		search := f.findUnitsNoexpand(a)
		if i%2 == 0 {
			search = f.findUnits(a)
		}
		out.WriteString("S")
		for u := range search {
			fmt.Fprintf(&out, " %d", index[u])
			if i%5 == 0 {
				out.WriteString("(")
				for nu := range f.findUnits(u.UnitFinderBoundingBox) {
					fmt.Fprintf(&out, " %d", index[nu])
				}
				out.WriteString(" )")
			}
		}
		out.WriteString("\n")
	}
	return out.String()
}

func TestUnitFinderMatchesReference_c10(t *testing.T) {
	got := unitFinderReferenceRun()
	want := strings.TrimPrefix(unitFinderReferenceOutput, "\n")
	if got != want {
		t.Errorf("unit finder output differs\ngot:\n%s\nwant:\n%s", got, want)
	}
}

func TestUnitFinderSearchClearsState_c10(t *testing.T) {
	f := newC10TestFunctions()
	u := &Unit{}
	u.UnitFinderBoundingBox = Rect{From: Xy{X: -1, Y: -1}, To: Xy{X: -1, Y: -1}}
	f.unitFinderInsertBB(u, Rect{From: Xy{X: 10, Y: 10}, To: Xy{X: 20, Y: 20}})
	for range f.findUnits(Rect{From: Xy{X: 0, Y: 0}, To: Xy{X: 30, Y: 30}}) {
		break
	}
	if f.unitFinderSearchIndex != 0 {
		t.Errorf("search index = %d, want 0", f.unitFinderSearchIndex)
	}
	if u.UnitFinderVisited != [4]bool{} {
		t.Errorf("visited marks = %v, want all false", u.UnitFinderVisited)
	}
}

func TestIsIntersecting_c10(t *testing.T) {
	f := newC10TestFunctions()
	a := Rect{From: Xy{X: 0, Y: 0}, To: Xy{X: 10, Y: 10}}
	cases := []struct {
		b    Rect
		want bool
	}{
		{Rect{From: Xy{X: 10, Y: 10}, To: Xy{X: 20, Y: 20}}, true},
		{Rect{From: Xy{X: 11, Y: 0}, To: Xy{X: 20, Y: 20}}, false},
		{Rect{From: Xy{X: -5, Y: 11}, To: Xy{X: 20, Y: 20}}, false},
		{Rect{From: Xy{X: -5, Y: -5}, To: Xy{X: -1, Y: 5}}, false},
		{Rect{From: Xy{X: -5, Y: -5}, To: Xy{X: 0, Y: 0}}, true},
	}
	for _, c := range cases {
		if got := f.isIntersecting(a, c.b); got != c.want {
			t.Errorf("isIntersecting(%v, %v) = %v, want %v", a, c.b, got, c.want)
		}
	}
}

func TestRepulseIndex_c10(t *testing.T) {
	f := newC10TestFunctions()
	f.gameSt.RepulseFieldWidth = 10
	if got := f.repulseIndex(Xy{X: 100, Y: 50}); got != 12 {
		t.Errorf("repulseIndex = %d, want 12", got)
	}
	if got := f.repulseIndex(Xy{X: -1, Y: 0}); got != ^uint(0)/48 {
		t.Errorf("repulseIndex of negative x = %d, want %d", got, ^uint(0)/48)
	}
}
