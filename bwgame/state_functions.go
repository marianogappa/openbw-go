package bwgame

type StateFunctions struct {
	Hooks Hooks

	st       *State
	globalSt *GlobalState
	gameSt   *GameState

	updateTiles           bool
	iscriptFlingy         *Flingy
	iscriptBullet         *Bullet
	iscriptUnit           *Unit
	unitFinderSearchIndex uint

	deselectUnit func(u *Unit)
}

func NewStateFunctions(st *State) *StateFunctions {
	return &StateFunctions{st: st, globalSt: st.Global, gameSt: st.Game}
}

type GameLoadFunctions struct {
	*StateFunctions

	setupInfo SetupInfo
}

func NewGameLoadFunctions(st *State) *GameLoadFunctions {
	return &GameLoadFunctions{StateFunctions: NewStateFunctions(st), setupInfo: NewSetupInfo()}
}

type ActionFunctions struct {
	*StateFunctions

	actionSt *ActionState
}

func NewActionFunctions(st *State, actionSt *ActionState) *ActionFunctions {
	f := &ActionFunctions{StateFunctions: NewStateFunctions(st), actionSt: actionSt}
	f.deselectUnit = f.onUnitDeselect
	return f
}
