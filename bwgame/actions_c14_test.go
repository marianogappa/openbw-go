package bwgame

import (
	"encoding/binary"
	"errors"
	"testing"
)

func newTestActionFunctions() (*ActionFunctions, *State, *ActionState) {
	st := NewState(nil, nil)
	actionSt := &ActionState{}
	return NewActionFunctions(st, actionSt), st, actionSt
}

func TestEraseFromSelectionKeepsStaleTail_c14(t *testing.T) {
	a, b, c := &Unit{}, &Unit{}, &Unit{}
	var selection unitSelection
	selection.PushBack(a)
	selection.PushBack(b)
	selection.PushBack(c)

	end := selection.size
	var seen []*Unit
	for i := 0; i != end; i++ {
		u := selection.data[i]
		seen = append(seen, u)
		if u == a {
			eraseFromSelection(&selection, findInSelection(&selection, a))
		}
	}

	want := []*Unit{a, c, c}
	for i := range want {
		if seen[i] != want[i] {
			t.Fatalf("iteration %d: got unexpected unit", i)
		}
	}
	if selection.Size() != 2 {
		t.Fatalf("size: got %d, want 2", selection.Size())
	}
}

func TestReadActionSetAlliances_c14(t *testing.T) {
	f, st, _ := newTestActionFunctions()
	r := NewDataReaderLe([]byte{0xEF, 0xBE, 0xAD, 0xDE})
	if !f.readActionSetAlliances(2, &r) {
		t.Fatal("readActionSetAlliances returned false")
	}
	want := [12]int32{3, 3, 2, 3, 2, 3, 3, 2, 1, 3, 2, 2}
	if st.Alliances[2] != want {
		t.Fatalf("got %v, want %v", st.Alliances[2], want)
	}
}

func TestActionSetSharedVision_c14(t *testing.T) {
	f, st, _ := newTestActionFunctions()
	st.SharedVision[3] = 0xABCD
	f.actionSetSharedVision(3, 0x1234)
	if st.SharedVision[3] != 0xAB34 {
		t.Fatalf("got %#x, want 0xab34", st.SharedVision[3])
	}
}

func TestUnitPlacementContainsPos_c14(t *testing.T) {
	ut := &UnitTypeT{PlacementSize: Xy{X: 64, Y: 64}}
	n := &Unit{UnitType: ut}
	n.Sprite = &Sprite{Position: Xy{X: 100, Y: 100}}
	cases := []struct {
		pos  Xy
		want bool
	}{
		{Xy{X: -5, Y: 100}, false},
		{Xy{X: 100, Y: 100}, true},
		{Xy{X: 68, Y: 100}, false},
		{Xy{X: 69, Y: 69}, true},
		{Xy{X: 132, Y: 100}, true},
		{Xy{X: 133, Y: 100}, false},
		{Xy{X: 131, Y: 131}, true},
	}
	for _, c := range cases {
		if got := unitPlacementContainsPos(n, c.pos); got != c.want {
			t.Errorf("pos %v: got %v, want %v", c.pos, got, c.want)
		}
	}
}

func TestActionCheat_c14(t *testing.T) {
	f, st, _ := newTestActionFunctions()
	if f.actionCheat(0, 0x10) {
		t.Fatal("cheat accepted while cheats are disabled")
	}
	st.CheatsEnabled = true
	st.Players[1].Controller = PlayerControllerOccupied
	if !f.actionCheat(0, 0x12) {
		t.Fatal("cheat rejected")
	}
	if !st.CheatOperationCwal {
		t.Fatal("operation cwal not set")
	}
	if st.CurrentMinerals[1] != 10000 || st.CurrentGas[1] != 10000 || st.CurrentMinerals[0] != 0 {
		t.Fatalf("unexpected resources: %v %v", st.CurrentMinerals, st.CurrentGas)
	}
	assertPanicsWithException(t, func() { f.actionCheat(0, 1) })
}

func assertPanicsWithException(t *testing.T, fn func()) {
	t.Helper()
	defer func() {
		recovered := recover()
		err, ok := recovered.(error)
		var exception *Exception
		if !ok || !errors.As(err, &exception) {
			t.Fatalf("expected *Exception panic, got %v", recovered)
		}
	}()
	fn()
}

func actionRecord(frame int32, actions ...byte) []byte {
	record := binary.LittleEndian.AppendUint32(nil, uint32(frame))
	record = append(record, byte(len(actions)))
	return append(record, actions...)
}

func TestExecuteActions_c14(t *testing.T) {
	f, st, actionSt := newTestActionFunctions()
	actionSt.PlayerID[1] = 7
	type call struct{ owner, action int32 }
	var calls []call
	f.Hooks.OnAction = func(owner, action int32) { calls = append(calls, call{owner, action}) }

	chat := append([]byte{7, 92}, make([]byte, 81)...)
	frame5 := append(append([]byte{7, 88, 1, 0, 2, 0}, chat...), 7, 13, 0x34, 0x12)
	data := append(actionRecord(5, frame5...), actionRecord(9, 7, 88, 0, 0, 0, 0)...)

	f.executeActions(data)
	if actionSt.NextActionFrame != 5 || actionSt.ActionsDataPosition != 0 || len(calls) != 0 {
		t.Fatalf("frame 0: next=%d pos=%d calls=%v", actionSt.NextActionFrame, actionSt.ActionsDataPosition, calls)
	}

	st.CurrentFrame = 5
	f.executeActions(data)
	if actionSt.NextActionFrame != 9 || actionSt.ActionsDataPosition != uint(5+len(frame5)) {
		t.Fatalf("frame 5: next=%d pos=%d", actionSt.NextActionFrame, actionSt.ActionsDataPosition)
	}
	wantCalls := []call{{1, 88}, {1, 92}, {1, 13}}
	if len(calls) != len(wantCalls) {
		t.Fatalf("calls: got %v, want %v", calls, wantCalls)
	}
	for i := range wantCalls {
		if calls[i] != wantCalls[i] {
			t.Fatalf("calls: got %v, want %v", calls, wantCalls)
		}
	}
	if st.SharedVision[1] != 0x34 {
		t.Fatalf("shared vision: got %#x", st.SharedVision[1])
	}

	st.CurrentFrame = 9
	f.executeActions(data)
	if actionSt.ActionsDataPosition != uint(len(data)) {
		t.Fatalf("frame 9: pos=%d, want %d", actionSt.ActionsDataPosition, len(data))
	}
}

func TestReadActionUnknownIDPanics_c14(t *testing.T) {
	f, _, _ := newTestActionFunctions()
	assertPanicsWithException(t, func() { f.readActionOwnerBytes(0, []byte{200}) })
	assertPanicsWithException(t, func() { f.readActionBytes([]byte{3, 92}) })
}
