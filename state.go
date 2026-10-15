package openbw

import (
	"iter"

	"github.com/marianogappa/openbw-go/bwgame"
)

const (
	playerSlots   = 12
	maxSupply     = 200
	fixedOneShift = 8
	raceCount     = 3
)

// Unit is a copy of one unit's live state.
type Unit struct {
	ID             UnitID
	Type           UnitType
	Owner          int8
	X, Y           int32
	HP, Shields    int32
	Energy         int32
	Order          uint8
	RemainingBuild int32
	Container      UnitID
	Completed      bool
	Burrowed       bool
	Cloaked        bool
	Loaded         bool
	Dying          bool
	Hallucination  bool
}

// Player is a copy of one player's live state.
type Player struct {
	Owner            int8
	Race             uint8
	Minerals, Gas    int32
	GatheredMinerals int32
	GatheredGas      int32
	SupplyUsed       float64
	SupplyMax        float64
	VictoryState     int32
	Left             bool
}

// UnitTypeInfo is this game's units.dat values for a type (a map can change them).
type UnitTypeInfo struct {
	Minerals, Gas, BuildTime int32
	Supply                   float64
	Building, Worker, Flyer  bool
}

func unitDead(u *bwgame.Unit) bool {
	return u.OrderType == nil || u.OrderType.ID == bwgame.OrdersDie && u.OrderState == 1
}

func containerOf(u *bwgame.Unit) *bwgame.Unit {
	if u.StatusFlags&bwgame.UnitStatusFlagLoaded != 0 && u.ConnectedUnit != nil {
		return u.ConnectedUnit
	}
	switch u.UnitType.ID {
	case bwgame.UnitTypesProtossInterceptor, bwgame.UnitTypesProtossScarab:
		if u.Fighter.Parent != nil && !u.Fighter.IsOutside {
			return u.Fighter.Parent
		}
	}
	return nil
}

func unitValue(u *bwgame.Unit) Unit {
	v := Unit{
		ID:             unitIDOf(u),
		Type:           unitTypeOf(u.UnitType),
		Owner:          int8(u.Owner),
		X:              u.Position.X,
		Y:              u.Position.Y,
		HP:             u.Hp.RawValue >> fixedOneShift,
		Shields:        u.ShieldPoints.RawValue >> fixedOneShift,
		Energy:         u.Energy.RawValue >> fixedOneShift,
		RemainingBuild: u.RemainingBuildTime,
		Container:      unitIDOf(containerOf(u)),
		Completed:      u.StatusFlags&bwgame.UnitStatusFlagCompleted != 0,
		Burrowed:       u.StatusFlags&bwgame.UnitStatusFlagBurrowed != 0,
		Cloaked:        u.StatusFlags&bwgame.UnitStatusFlagCloaked != 0,
		Loaded:         u.StatusFlags&bwgame.UnitStatusFlagLoaded != 0,
		Hallucination:  u.StatusFlags&bwgame.UnitStatusFlagHallucination != 0,
	}
	if u.OrderType != nil {
		v.Order = uint8(u.OrderType.ID)
		v.Dying = u.OrderType.ID == bwgame.OrdersDie
	}
	return v
}

// Unit reads a unit's live state. It is false if the unit was removed.
func (g *Game) Unit(id UnitID) (Unit, bool) {
	if id.Index() == 0 {
		return Unit{}, false
	}
	u := g.st.UnitsContainer.TryGet(uint(id.Index() - 1))
	if u == nil || unitDead(u) || unitIDOf(u) != id {
		return Unit{}, false
	}
	return unitValue(u), true
}

// Units yields the units not yet removed (dying ones too), owner 0..11, in engine list order.
// Turret subunits are left out.
func (g *Game) Units() iter.Seq[Unit] {
	return func(yield func(Unit) bool) {
		for owner := range g.st.PlayerUnits {
			for u := range g.st.PlayerUnits[owner].All() {
				if isTurret(u) {
					continue
				}
				if !yield(unitValue(u)) {
					return
				}
			}
		}
	}
}

func supplyValue(v bwgame.Fp1) float64 { return float64(v.RawValue) / 2 }

// Player reads a player's live state. Owner is 0..11.
func (g *Game) Player(owner int) Player {
	st := g.st
	p := Player{Owner: int8(owner)}
	if owner < 0 || owner >= playerSlots {
		return p
	}
	race := st.Players[owner].Race
	p.Race = uint8(race)
	p.Minerals, p.Gas = st.CurrentMinerals[owner], st.CurrentGas[owner]
	p.GatheredMinerals, p.GatheredGas = st.TotalMineralsGathered[owner], st.TotalGasGathered[owner]
	if race >= 0 && race < raceCount {
		p.SupplyUsed = supplyValue(st.SupplyUsed[owner][race])
		p.SupplyMax = min(supplyValue(st.SupplyAvailable[owner][race]), maxSupply)
	}
	p.VictoryState = st.Players[owner].VictoryState
	p.Left = st.Players[owner].Controller == bwgame.PlayerControllerUserLeft
	return p
}

// UnitTypeInfo gives this game's units.dat values for t.
func (g *Game) UnitTypeInfo(t UnitType) UnitTypeInfo {
	types := g.st.Game.UnitTypes.Vec
	if int(t) >= len(types) {
		return UnitTypeInfo{}
	}
	ut := &types[t]
	return UnitTypeInfo{
		Minerals:  ut.MineralCost,
		Gas:       ut.GasCost,
		BuildTime: ut.BuildTime,
		Supply:    supplyValue(ut.SupplyRequired),
		Building:  ut.Flags&bwgame.UnitTypeTFlagBuilding != 0,
		Worker:    ut.Flags&bwgame.UnitTypeTFlagWorker != 0,
		Flyer:     ut.Flags&bwgame.UnitTypeTFlagFlyer != 0,
	}
}

func isPlayingController(controller int32) bool {
	switch controller {
	case bwgame.PlayerControllerOccupied, bwgame.PlayerControllerComputerGame, bwgame.PlayerControllerComputer,
		bwgame.PlayerControllerUserLeft, bwgame.PlayerControllerComputerDefeated:
		return true
	}
	return false
}

func playingSlots(st *bwgame.State) []int {
	var slots []int
	for i := range playerSlots {
		if isPlayingController(st.Players[i].Controller) {
			slots = append(slots, i)
		}
	}
	return slots
}

// PlayingSlots are the owners of the players in the game (occupied, computer or left at load).
func (g *Game) PlayingSlots() []int { return append([]int(nil), g.playing...) }
