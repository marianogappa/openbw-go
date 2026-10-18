package bwgame

import "testing"

var scrObjectLimits = ObjectLimits{Images: 10000, Sprites: 5000, LoneSprites: 1000, Bullets: 400, Orders: 4000}

func TestZeroRulesUseOpenBWLimits(t *testing.T) {
	var rules Rules
	if rules.ObjectLimits() != Legacy116ObjectLimits || DefaultRules().ObjectLimits() != Legacy116ObjectLimits {
		t.Fatalf("limits %+v", rules.ObjectLimits())
	}
	if rules.OrderSupplyLowThreshold() != 1800 {
		t.Fatalf("threshold %d", rules.OrderSupplyLowThreshold())
	}
	if rules.HaloRocketsBulletCap() != 80 {
		t.Fatalf("halo rockets cap %d", rules.HaloRocketsBulletCap())
	}
}

func TestHaloRocketsBulletCapScalesWithBulletLimit(t *testing.T) {
	for _, c := range []struct{ bullets, want uint }{{200, 160}, {400, 320}} {
		rules := Rules{Limits: ObjectLimits{Bullets: c.bullets}}
		if got := rules.HaloRocketsBulletCap(); got != c.want {
			t.Errorf("bullets %d: cap %d, want %d", c.bullets, got, c.want)
		}
	}
}

func TestNewStateSizesContainersFromRules(t *testing.T) {
	game := NewGameState()
	game.Rules.Limits = scrObjectLimits
	st := NewState(nil, game)
	sizes := []uint{st.UnitsContainer.MaxSize, st.BulletsContainer.MaxSize, st.SpritesContainer.MaxSize, st.ImagesContainer.MaxSize, st.OrdersContainer.MaxSize}
	want := []uint{1700, 400, 5000, 10000, 4000}
	if !equalIndices(sizes, want) {
		t.Fatalf("sizes %v, want %v", sizes, want)
	}
	if game.Rules.OrderSupplyLowThreshold() != 3600 {
		t.Fatalf("threshold %d", game.Rules.OrderSupplyLowThreshold())
	}
}

func TestBulletContainerHandsOutIndicesUpToTheLimit(t *testing.T) {
	game := NewGameState()
	game.Rules.Limits = scrObjectLimits
	st := NewState(nil, game)
	count := 0
	for st.BulletsContainer.Top() != nil {
		st.BulletsContainer.Pop()
		count++
	}
	if count != 400 {
		t.Fatalf("allocated %d bullets", count)
	}
}

func newOrderAllocationTestFunctions(rules Rules) *StateFunctions {
	game := NewGameState()
	game.Rules = rules
	return NewStateFunctions(NewState(nil, game))
}

func TestCanAllocateOrder(t *testing.T) {
	scr := Rules{Limits: scrObjectLimits, OrderAllocationRule: true, FlyingSCVFix: true}
	resetCollision := &OrderTypeT{ID: OrdersResetCollision}
	move := &OrderTypeT{ID: OrdersMove}
	cases := []struct {
		name         string
		rules        Rules
		activeOrders uint
		queueCount   int32
		currentOrder Orders
		newOrder     *OrderTypeT
		want         bool
	}{
		{"default rules never drop", DefaultRules(), 3999, 8, OrdersMove, move, true},
		{"supply not low and short queue", scr, 3599, 7, OrdersMove, move, true},
		{"order supply low", scr, 3600, 0, OrdersMove, move, false},
		{"queue full", scr, 0, 8, OrdersMove, move, false},
		{"dying unit", scr, 3600, 8, OrdersDie, move, true},
		{"reset harvest collision", scr, 3600, 8, OrdersResetHarvestCollision, move, true},
		{"flying SCV fix", scr, 0, 8, OrdersConstructingBuilding, resetCollision, true},
		{"flying SCV fix needs reset collision", scr, 0, 8, OrdersConstructingBuilding, move, false},
		{"no flying SCV fix", Rules{Limits: scrObjectLimits, OrderAllocationRule: true}, 0, 8, OrdersConstructingBuilding, resetCollision, false},
	}
	for _, c := range cases {
		f := newOrderAllocationTestFunctions(c.rules)
		f.st.ActiveOrdersSize = c.activeOrders
		u := &Unit{OrderType: &OrderTypeT{ID: c.currentOrder}, OrderQueueCount: c.queueCount}
		if got := f.canAllocateOrder(u, c.newOrder); got != c.want {
			t.Errorf("%s: got %v, want %v", c.name, got, c.want)
		}
	}
}

func TestAllyForcesAlliesOccupiedTeammates(t *testing.T) {
	f := NewGameLoadFunctions(NewState(nil, NewGameState()))
	st := f.st
	forces := []int32{1, 2, 1, 2, 1, 2}
	for i, force := range forces {
		st.Players[i] = Player{Controller: PlayerControllerOccupied, Force: force}
		st.SharedVision[i] = 1 << i
	}
	st.Players[4].Controller = PlayerControllerOpen
	(&mapDataLoader{f: f}).allyForces()

	if st.Alliances[0][2] != 2 || st.Alliances[2][0] != 2 || st.Alliances[1][3] != 2 {
		t.Fatalf("teammates not allied: %v", st.Alliances)
	}
	if st.Alliances[0][1] != 0 || st.Alliances[0][4] != 0 || st.Alliances[4][0] != 0 {
		t.Fatalf("enemy or open slot allied: %v", st.Alliances)
	}
	if st.SharedVision[0] != 0b101 || st.SharedVision[1] != 0b101010 {
		t.Fatalf("shared vision %b %b", st.SharedVision[0], st.SharedVision[1])
	}
}

func TestAlliedVictoryWithinForcesCapsOtherTeams(t *testing.T) {
	game := NewGameState()
	game.AlliedVictoryWithinForces = true
	f := NewActionFunctions(NewState(nil, game), &ActionState{})
	for i, force := range []int32{1, 2, 1} {
		f.st.Players[i] = Player{Controller: PlayerControllerOccupied, Force: force}
	}
	f.actionSetAlliances(0, [12]int32{2, 2, 2, 0, 0, 0, 0, 0, 0, 0, 0, 2})
	if got := f.st.Alliances[0]; got[1] != 1 || got[2] != 2 || got[11] != 2 {
		t.Fatalf("alliances %v", got)
	}
}
