package bwgame

import "testing"

func newC03TestFunctions() *StateFunctions {
	return NewStateFunctions(NewState(nil, nil))
}

func TestCalcRepairCostsMatchesCpp(t *testing.T) {
	cases := [][8]int32{
		{400, 0, 1500, 0, 384, 1, 0, 7},
		{150, 100, 500, 0, 128, 1, 1, 30},
		{100, 150, 500, 0, 128, 1, 1, 30},
		{50, 0, 40, 0, 11, 1, 0, 55},
		{0, 0, 100, 0, 26, 0, 0, 65535},
		{150, 100, 500, 1, 128, 1, 1, 1},
		{3, 7, 2000, 0, 512, 1, 2, 1000},
		{75, 25, 80, 1, 21, 3, 1, 7},
		{1000, 0, 1, 0, 1, 1, 0, 1},
		{1000, 500, 2, 0, 400, 260, 130, 1},
		{100, 400, 1, 1, 3, 6, 25, 1},
		{0, 300, 1, 0, 700, 0, 0, 65535},
	}
	for _, c := range cases {
		f := newC03TestFunctions()
		f.st.CheatOperationCwal = c[3] != 0
		target := &Unit{UnitType: &UnitTypeT{MineralCost: c[0], GasCost: c[1], Hitpoints: Fp8Integer(c[2])}, HpConstructionRate: Fp8FromRaw(c[4])}
		got := f.calcRepairCosts(target)
		want := repairCosts{mineral: c[5], gas: c[6], time: c[7]}
		if got != want {
			t.Errorf("case %v: got %+v, want %+v", c, got, want)
		}
	}
}

func TestSetRemoveTimerKeepsSmallerNonZero(t *testing.T) {
	f := newC03TestFunctions()
	u := &Unit{}
	f.setRemoveTimer(u, 1350)
	if u.RemoveTimer != 1350 {
		t.Fatalf("got %d, want 1350", u.RemoveTimer)
	}
	f.setRemoveTimer(u, 1800)
	if u.RemoveTimer != 1350 {
		t.Fatalf("got %d, want 1350", u.RemoveTimer)
	}
	f.setRemoveTimer(u, 0)
	if u.RemoveTimer != 0 {
		t.Fatalf("got %d, want 0", u.RemoveTimer)
	}
}

func TestMergeStatusEffects(t *testing.T) {
	f := newC03TestFunctions()
	irradiator := &Unit{}
	to := &Unit{DefensiveMatrixHp: Fp8FromRaw(10), StimTimer: 5, IrradiateTimer: 3, ParasiteFlags: 1, BlindedBy: 2, MaelstromTimer: 9, AcidSporeCount: 4}
	from := &Unit{DefensiveMatrixHp: Fp8FromRaw(20), StimTimer: 2, IrradiateTimer: 7, IrradiatedBy: irradiator, IrradiateOwner: 3, ParasiteFlags: 4, BlindedBy: 8, MaelstromTimer: 1, AcidSporeCount: 1}
	from.AcidSporeTime[0] = 11
	f.mergeStatusEffects(to, from)
	if to.DefensiveMatrixHp != Fp8FromRaw(20) || to.StimTimer != 5 || to.MaelstromTimer != 9 {
		t.Fatalf("max merge wrong: %+v", to)
	}
	if to.IrradiateTimer != 7 || to.IrradiatedBy != irradiator || to.IrradiateOwner != 3 {
		t.Fatalf("irradiate merge wrong: %+v", to)
	}
	if to.ParasiteFlags != 5 || to.BlindedBy != 10 {
		t.Fatalf("flag merge wrong: %+v", to)
	}
	if to.AcidSporeCount != 1 || to.AcidSporeTime[0] != 11 {
		t.Fatalf("acid spores not copied: %+v", to)
	}
}
