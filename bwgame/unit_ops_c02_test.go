package bwgame

import "testing"

func newUnitOpsTestFunctions() *StateFunctions {
	return NewStateFunctions(NewState(nil, NewGameState()))
}

func TestIsInPsionicMatrixRangeMatchesCpp(t *testing.T) {
	f := newUnitOpsTestFunctions()
	var hash uint64
	count := 0
	for y := int32(-170); y <= 170; y += 3 {
		for x := int32(-260); x <= 260; x += 3 {
			r := f.isInPsionicMatrixRange(Xy{X: x, Y: y})
			hash *= 31
			if r {
				hash++
				count++
			}
		}
	}
	if hash != 17924447535075114112 || count != 14536 {
		t.Fatalf("grid: got hash %d count %d", hash, count)
	}

	samples := []struct {
		rel  Xy
		want bool
	}{
		{Xy{X: 0, Y: 0}, true},
		{Xy{X: -1, Y: -1}, true},
		{Xy{X: -32, Y: 0}, true},
		{Xy{X: 31, Y: -31}, true},
		{Xy{X: -255, Y: 159}, false},
		{Xy{X: 255, Y: -159}, false},
		{Xy{X: -256, Y: 0}, false},
		{Xy{X: 0, Y: -160}, false},
		{Xy{X: -33, Y: -33}, true},
		{Xy{X: 100, Y: -100}, true},
	}
	for _, s := range samples {
		if got := f.isInPsionicMatrixRange(s.rel); got != s.want {
			t.Errorf("isInPsionicMatrixRange(%v) = %v, want %v", s.rel, got, s.want)
		}
	}
}

func TestPartiallyRefundUnitCostsMultipliesBeforeDividing(t *testing.T) {
	f := newUnitOpsTestFunctions()
	f.st.CurrentMinerals[1] = 10
	f.st.CurrentGas[1] = 20
	f.partiallyRefundUnitCosts(1, &UnitTypeT{MineralCost: 75, GasCost: 25})
	if f.st.CurrentMinerals[1] != 10+56 || f.st.CurrentGas[1] != 20+18 {
		t.Fatalf("got minerals %d gas %d", f.st.CurrentMinerals[1], f.st.CurrentGas[1])
	}
	f.refundUnitCosts(1, &UnitTypeT{MineralCost: 75, GasCost: 25})
	if f.st.CurrentMinerals[1] != 141 || f.st.CurrentGas[1] != 63 {
		t.Fatalf("got minerals %d gas %d", f.st.CurrentMinerals[1], f.st.CurrentGas[1])
	}
}

func TestHasAvailableResourcesFor(t *testing.T) {
	f := newUnitOpsTestFunctions()
	f.st.CurrentMinerals[0] = 100
	f.st.CurrentGas[0] = 50
	cases := []struct {
		minerals, gas int32
		want          bool
	}{
		{100, 50, true},
		{101, 0, false},
		{0, 51, false},
	}
	for _, c := range cases {
		if got := f.hasAvailableResourcesFor(0, &UnitTypeT{MineralCost: c.minerals, GasCost: c.gas}, false); got != c.want {
			t.Errorf("unit %d/%d: got %v", c.minerals, c.gas, got)
		}
		if got := f.hasAvailableResourcesForTech(0, &TechTypeT{MineralCost: c.minerals, GasCost: c.gas}, false); got != c.want {
			t.Errorf("tech %d/%d: got %v", c.minerals, c.gas, got)
		}
	}
}

func TestScaledHpKeepsRawOneForZero(t *testing.T) {
	if got := scaledHp(Fp8Integer(1), 100, 50); got != Fp8FromRaw(1) {
		t.Fatalf("got %v", got)
	}
	if got := scaledHp(Fp8Integer(30), 0, 50); got != Fp8Integer(1) {
		t.Fatalf("got %v", got)
	}
	if got := scaledHp(Fp8Integer(30), 60, 200); got != Fp8Integer(100) {
		t.Fatalf("got %v", got)
	}
}

func TestBuildQueueClearSingleMovesRestToLimbo(t *testing.T) {
	f := newUnitOpsTestFunctions()
	a, b, c := &UnitTypeT{ID: 1}, &UnitTypeT{ID: 2}, &UnitTypeT{ID: 3}
	u := &Unit{}
	u.BuildQueue.PushBack(a)
	u.BuildQueue.PushBack(b)
	u.BuildQueue.PushBack(c)
	f.buildQueueClearSingle(u)
	if !u.BuildQueue.Empty() || u.BuildQueueLimbo.Size() != 2 || u.BuildQueueLimbo.At(0) != b || u.BuildQueueLimbo.At(1) != c {
		t.Fatalf("queue %v limbo %v", u.BuildQueue.Values(), u.BuildQueueLimbo.Values())
	}
}
