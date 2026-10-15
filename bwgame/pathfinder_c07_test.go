package bwgame

import (
	"fmt"
	"strings"
	"testing"
)

func TestBinaryHeapMatchesCpp(t *testing.T) {
	var open []int32
	seed := uint32(12345)
	rnd := func() int32 {
		seed = seed*22695477 + 1
		return int32((seed >> 16) % 50)
	}
	less := func(a, b int32) bool { return a < b }
	var out strings.Builder
	for step := 0; step < 300; step++ {
		r := rnd()
		if r < 30 || len(open) == 0 {
			open = append(open, rnd())
			binaryHeapUp(open, len(open)-1, less)
		} else if r < 40 {
			v := open[0]
			open[0], open[len(open)-1] = open[len(open)-1], open[0]
			open = open[:len(open)-1]
			binaryHeapDown(open, 0, less)
			fmt.Fprintf(&out, "pop %d\n", v)
		} else {
			idx := int(rnd()) % len(open)
			open[idx] -= 7
			binaryHeapUp(open, idx, less)
		}
	}
	out.WriteString("final")
	for _, v := range open {
		fmt.Fprintf(&out, " %d", v)
	}
	out.WriteString("\n")
	sorted := []int32{1, 3, 3, 3, 5, 8, 8, 13}
	for v := int32(0); v <= 14; v++ {
		lower := pfPartitionPoint(len(sorted), func(i int) bool { return sorted[i] < v })
		upper := pfPartitionPoint(len(sorted), func(i int) bool { return !(v < sorted[i]) })
		fmt.Fprintf(&out, "bounds %d %d %d\n", v, lower, upper)
	}
	if out.String() != strings.TrimPrefix(binaryHeapCppOutput, "\n") {
		t.Errorf("output differs from C++:\n%s", out.String())
	}
}

const binaryHeapCppOutput = `
pop 14
pop 27
pop 1
pop 10
pop 14
pop 19
pop 3
pop 9
pop 16
pop 10
pop 11
pop 18
pop 3
pop 17
pop -5
pop 3
pop 7
pop 7
pop 9
pop 0
pop 10
pop 13
pop 0
pop 8
pop 3
pop 1
pop 11
pop 13
pop 14
pop 0
pop 8
pop 9
pop 6
pop 10
pop 5
pop 6
pop 3
pop 3
pop 1
pop 2
pop -3
pop 0
pop 0
pop 1
pop 2
pop 0
pop 2
pop 3
pop 4
pop 2
pop 6
pop 0
pop 1
pop 1
pop 2
pop 2
pop 3
pop 5
final 4 5 6 6 8 11 8 8 12 9 9 13 12 16 10 12 20 13 19 19 17 15 25 16 14 33 14 22 21 18 11 17 14 28 37 17 23 29 23 32 29 24 20 30 24 34 31 18 22 36 25 36 35 15 24 27 24 27 34 21 29 21 35 21 18 17 16 47 29 46 37 33 30 46 41 38 45 36 44 48 38 39 48 37 24 44 22 39 48 40 40 49 36 42 42 20 19 34 42 44 37 29 42 43 42 36 42 35 32 45 38 29 43 43 32 42 37 40 35 46 23 37 44 43 38 42 47 37 24 21 31 27 30 41
bounds 0 0 0
bounds 1 0 1
bounds 2 1 1
bounds 3 1 4
bounds 4 4 4
bounds 5 4 5
bounds 6 5 5
bounds 7 5 5
bounds 8 5 7
bounds 9 7 7
bounds 10 7 7
bounds 11 7 7
bounds 12 7 7
bounds 13 7 8
bounds 14 8 8
`

func newPathTestFunctions(regionCount int) *StateFunctions {
	game := NewGameState()
	game.Regions.Regions = make([]Region, regionCount)
	return NewStateFunctions(NewState(&GlobalState{}, game))
}

func TestNewPathReusesLastFreedPathFirst(t *testing.T) {
	f := newPathTestFunctions(0)
	a := f.newPath()
	b := f.newPath()
	c := f.newPath()
	if len(f.st.Paths) != 3 {
		t.Fatalf("paths = %d, want 3", len(f.st.Paths))
	}
	f.freePath(a)
	f.freePath(c)
	f.freePath(b)
	for i, want := range []*Path{b, c, a} {
		if got := f.newPath(); got != want {
			t.Errorf("reuse %d: got another path", i)
		}
	}
	if len(f.st.Paths) != 3 {
		t.Errorf("paths = %d, want 3", len(f.st.Paths))
	}
}

func TestNewPathLimit(t *testing.T) {
	f := newPathTestFunctions(0)
	for i := 0; i < 1024; i++ {
		if f.newPath() == nil {
			t.Fatalf("path %d is nil", i)
		}
	}
	if f.newPath() != nil {
		t.Errorf("path 1025 is not nil")
	}
}

func TestFreeUnitPath(t *testing.T) {
	f := newPathTestFunctions(0)
	u := &Unit{}
	f.freeUnitPath(u)
	u.Path = f.newPath()
	path := u.Path
	f.freeUnitPath(u)
	if u.Path != nil {
		t.Errorf("unit path is not nil")
	}
	if f.st.FreePaths.Front() != path {
		t.Errorf("freed path is not at the front of the free list")
	}
}

func TestCreateSingleStepPathOnNewPath(t *testing.T) {
	f := newPathTestFunctions(4)
	path := f.createSingleStepPath(Xy{X: 1, Y: 2}, Xy{X: 30, Y: 40})
	if len(path.LongPath) != 1 || path.LongPath[0] != nil {
		t.Errorf("long path = %v, want [nil]", path.LongPath)
	}
	if len(path.ShortPath) != 1 || path.ShortPath[0] != (Xy{X: 30, Y: 40}) {
		t.Errorf("short path = %v", path.ShortPath)
	}
	if path.Source != (Xy{X: 1, Y: 2}) || path.Destination != (Xy{X: 30, Y: 40}) || path.Next != (Xy{X: 30, Y: 40}) {
		t.Errorf("source, destination or next is wrong: %+v", path)
	}
	if path.FullLongPathSize != 1 || path.CurrentLongPathIndex != 0 || path.CurrentShortPathIndex != 0 {
		t.Errorf("indices are wrong: %+v", path)
	}
}

func TestCreateSingleStepPathReadsStaleShortPath(t *testing.T) {
	f := newPathTestFunctions(4)
	reused := f.newPath()
	reused.ShortPath = []Xy{{X: 0, Y: 0}, {X: 2, Y: 0}}
	f.freePath(reused)
	path := f.createSingleStepPath(Xy{}, Xy{X: 5, Y: 5})
	if path != reused {
		t.Fatalf("path was not reused")
	}
	if len(path.LongPath) != 1 || path.LongPath[0] != &f.gameSt.Regions.Regions[2] {
		t.Errorf("long path does not point to region 2")
	}

	reused.ShortPath = []Xy{{X: 0, Y: 0}, {X: 9, Y: 0}}
	f.freePath(reused)
	path = f.createSingleStepPath(Xy{}, Xy{X: 5, Y: 5})
	if len(path.LongPath) != 1 || path.LongPath[0] != nil {
		t.Errorf("long path = %v, want [nil]", path.LongPath)
	}

	keptRegion := &f.gameSt.Regions.Regions[3]
	reused.ShortPath = []Xy{{X: 0, Y: 0}}
	reused.LongPath = []*Region{keptRegion, nil, nil}
	f.freePath(reused)
	path = f.createSingleStepPath(Xy{}, Xy{X: 5, Y: 5})
	if len(path.LongPath) != 1 || path.LongPath[0] != keptRegion {
		t.Errorf("long path did not keep its first region")
	}
}

func TestResizeSlice(t *testing.T) {
	s := resizeSlice([]int32{1, 2, 3}, 1)
	if len(s) != 1 || s[0] != 1 {
		t.Errorf("shrink = %v", s)
	}
	s = resizeSlice(s, 3)
	if len(s) != 3 || s[0] != 1 || s[1] != 0 || s[2] != 0 {
		t.Errorf("grow = %v", s)
	}
}

func pfPartitionPoint(n int, isBefore func(i int) bool) int {
	first := 0
	length := n
	for length > 0 {
		half := length / 2
		mid := first + half
		if isBefore(mid) {
			first = mid + 1
			length -= half + 1
		} else {
			length = half
		}
	}
	return first
}
