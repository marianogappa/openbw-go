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
