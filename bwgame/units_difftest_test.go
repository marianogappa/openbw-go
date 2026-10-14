package bwgame

import "testing"

func TestHpConstructionRateSubtractsOneRawUnit(t *testing.T) {
	creepColony := &UnitTypeT{Hitpoints: Fp8Integer(400), BuildTime: 300}
	u := &Unit{UnitType: creepColony}
	(&StateFunctions{}).initializeUnitConstructionRates(u)
	if u.HpConstructionRate.RawValue != 308 {
		t.Fatalf("hp construction rate raw %d, want 308", u.HpConstructionRate.RawValue)
	}
}

func TestHpConstructionRateMinimumIsOneRawUnit(t *testing.T) {
	u := &Unit{UnitType: &UnitTypeT{Hitpoints: Fp8FromRaw(0), BuildTime: 300}}
	(&StateFunctions{}).initializeUnitConstructionRates(u)
	if u.HpConstructionRate.RawValue != 1 {
		t.Fatalf("hp construction rate raw %d, want 1", u.HpConstructionRate.RawValue)
	}
}
