package main

import (
	"github.com/marianogappa/openbw-go/bwgame"
	"github.com/marianogappa/openbw-go/downgrade"
)

type plainAction struct {
	frame     int32
	owner     int32
	actionID  byte
	ok        bool
	selection []uint32
}

const (
	actionSelect      = 0x09
	actionShiftSelect = 0x0a
	actionDeselect    = 0x0b
	actionHotkey      = 0x13
	hotkeyRecall      = 1
	selectionOwners   = 8
	unitIndexBits     = 11
	generationLimit   = 1 << 21
)

func engineUnitID(u *bwgame.Unit) uint32 {
	return uint32(u.Index+1) | (u.UnitIDGeneration%generationLimit)<<unitIndexBits
}

func changesSelection(id byte, payload []byte) bool {
	switch id {
	case actionSelect, actionShiftSelect, actionDeselect:
		return true
	case actionHotkey:
		return len(payload) > 0 && payload[0] == hotkeyRecall
	}
	return false
}

func selectionOf(f *bwgame.ReplayFunctions, owner int32) []uint32 {
	if owner < 0 || owner >= selectionOwners {
		return nil
	}
	units := f.ActionState().Selection[owner].Values()
	if len(units) == 0 {
		return nil
	}
	ids := make([]uint32, len(units))
	for i, u := range units {
		ids[i] = engineUnitID(u)
	}
	return ids
}

type plainRun struct {
	actions []plainAction
	frame   int32
	hash    uint64
}

func runPlainEngine(global *bwgame.GlobalState, prepared *downgrade.Prepared, rules bwgame.Rules) (*plainRun, error) {
	st := bwgame.NewState(global, bwgame.NewGameState())
	st.Game.Rules = rules
	f := bwgame.NewReplayFunctions(st, &bwgame.ActionState{}, &bwgame.ReplayState{})
	run := &plainRun{}
	var pending *plainAction
	selectAfter := false
	f.BeforeAction = func(owner, actionID int32, payload []byte) {
		run.actions = append(run.actions, plainAction{frame: st.CurrentFrame, owner: owner, actionID: byte(actionID)})
		pending = &run.actions[len(run.actions)-1]
		selectAfter = changesSelection(byte(actionID), payload)
		if !selectAfter {
			pending.selection = selectionOf(f, owner)
		}
	}
	f.AfterAction = func(owner, actionID int32, ok bool) {
		if pending == nil {
			run.actions = append(run.actions, plainAction{frame: st.CurrentFrame, owner: owner, actionID: byte(actionID), ok: ok})
			return
		}
		pending.ok = ok
		if selectAfter {
			pending.selection = selectionOf(f, owner)
		}
		pending, selectAfter = nil, false
	}
	if err := f.LoadReplayData(prepared.Loadable); err != nil {
		return nil, err
	}
	for !f.IsDone() {
		if err := f.NextFrame(); err != nil {
			return nil, err
		}
	}
	run.frame = st.CurrentFrame
	run.hash = bwgame.StateHash(st)
	return run, nil
}
