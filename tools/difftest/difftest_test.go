package main

import (
	"bytes"
	"strings"
	"testing"
)

const harnessOutputA = `frame 0 aaaa
frame 1 bbbb
units_at_frame 1
index	type	owner	x	y	hp	shields	energy	order	completed
5	7	0	10	20	40	0	0	3	1
6	7	0	11	21	40	0	0	3	1
end_units
end 1 bbbb
player 0 name=p0 race=1 minerals=50 gas=0 supply_zerg=0/0 supply_terran=4/10 supply_protoss=0/0
units 0 7=2
action_fails 0
`

const harnessOutputB = `frame 0 aaaa
frame 1 cccc
units_at_frame 1
index	type	owner	x	y	hp	shields	energy	order	completed
5	7	0	10	22	40	0	0	3	1
7	7	0	11	21	40	0	0	3	1
end_units
end 1 cccc
player 0 name=p0 race=1 minerals=42 gas=0 supply_zerg=0/0 supply_terran=4/10 supply_protoss=0/0
units 0 7=2
action_fails 0
`

func TestFirstDifferentFrame(t *testing.T) {
	a := parseRunOutput([]byte(harnessOutputA))
	b := parseRunOutput([]byte(harnessOutputB))
	frame, differs := firstDifferentFrame(a, b)
	if !differs || frame != 1 {
		t.Fatalf("frame %d, differs %v", frame, differs)
	}
	if _, differs := firstDifferentFrame(a, a); differs {
		t.Fatal("same output differs")
	}
}

func TestMissingFrameIsADifference(t *testing.T) {
	a := parseRunOutput([]byte(harnessOutputA))
	b := parseRunOutput([]byte("frame 0 aaaa\nerror 0 boom\n"))
	if frame, differs := firstDifferentFrame(a, b); !differs || frame != 1 {
		t.Fatalf("frame %d, differs %v", frame, differs)
	}
	if b.errorLine != "error 0 boom" {
		t.Fatalf("error line %q", b.errorLine)
	}
}

func TestSnapshotDiffListsFieldAndUnitDifferences(t *testing.T) {
	a := parseRunOutput([]byte(harnessOutputA))
	b := parseRunOutput([]byte(harnessOutputB))
	var w bytes.Buffer
	printSnapshotDiff(&w, 1, "harness", "obwhash", a, b, 0)
	out := w.String()
	for _, want := range []string{
		"unit 5 (type 7) y: harness=20 obwhash=22",
		"unit 6 only in harness",
		"unit 7 only in obwhash",
		"player 0 minerals: harness=50 obwhash=42",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in:\n%s", want, out)
		}
	}
	if strings.Contains(out, "supply_terran") {
		t.Errorf("equal field reported:\n%s", out)
	}
}
