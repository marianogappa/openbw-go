package bwgame

type Hooks struct {
	PlaySound          func(id int32, position Xy, sourceUnit *Unit, addRaceIndex bool)
	OnUnitDeselect     func(u *Unit)
	OnUnitDestroy      func(u *Unit)
	OnKillUnit         func(u *Unit)
	OnPlayerEliminated func(owner int32)
	OnVictoryState     func(owner int32, state int32)
	OnAction           func(owner int32, action int32)
}

func (f *StateFunctions) playSoundAt(id int32, position Xy, sourceUnit *Unit, addRaceIndex bool) {
	if f.Hooks.PlaySound != nil {
		f.Hooks.PlaySound(id, position, sourceUnit, addRaceIndex)
	}
}

func (f *StateFunctions) onUnitDeselect(u *Unit) {
	if f.deselectUnit != nil {
		f.deselectUnit(u)
	}
	if f.Hooks.OnUnitDeselect != nil {
		f.Hooks.OnUnitDeselect(u)
	}
}

func (f *StateFunctions) onUnitDestroy(u *Unit) {
	if f.Hooks.OnUnitDestroy != nil {
		f.Hooks.OnUnitDestroy(u)
	}
}

func (f *StateFunctions) onKillUnit(u *Unit) {
	if f.Hooks.OnKillUnit != nil {
		f.Hooks.OnKillUnit(u)
	}
}

func (f *StateFunctions) onPlayerEliminated(owner int32) {
	if f.Hooks.OnPlayerEliminated != nil {
		f.Hooks.OnPlayerEliminated(owner)
	}
}

func (f *StateFunctions) onVictoryState(owner int32, state int32) {
	if f.Hooks.OnVictoryState != nil {
		f.Hooks.OnVictoryState(owner, state)
	}
}

func (f *ActionFunctions) onAction(owner int32, action int32) {
	if f.Hooks.OnAction != nil {
		f.Hooks.OnAction(owner, action)
	}
}
