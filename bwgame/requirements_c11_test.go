package bwgame

import "testing"

func TestOwnerMismatchRejectsWithoutFurtherChecks(t *testing.T) {
	f := &StateFunctions{}
	u := &Unit{}
	u.Owner = 1
	if f.unitCanReceiveOrder(u, &OrderTypeT{ID: OrdersStop}, 2) {
		t.Fatal("unitCanReceiveOrder accepted another owner")
	}
	if f.unitCanResearchOwner(u, &TechTypeT{ID: TechTypesStimPacks}, 2) {
		t.Fatal("unitCanResearchOwner accepted another owner")
	}
	if f.unitCanUpgradeOwner(u, &UpgradeTypeT{ID: UpgradeTypesTerranInfantryArmor}, 2) {
		t.Fatal("unitCanUpgradeOwner accepted another owner")
	}
	if f.unitCanUseTechOwner(u, &TechTypeT{ID: TechTypesStimPacks}, 2) {
		t.Fatal("unitCanUseTechOwner accepted another owner")
	}
}

func TestUnitBuildOrderValidRejectsNonBuildOrders(t *testing.T) {
	f := &StateFunctions{}
	if f.unitBuildOrderValid(&Unit{}, &OrderTypeT{ID: OrdersMove}, &UnitTypeT{}, 0) {
		t.Fatal("Move accepted as a build order")
	}
}

func TestUnitBuildOrderValidRejectsFlagBeaconAfterFrame600(t *testing.T) {
	st := &State{}
	st.CurrentFrame = 601
	f := &StateFunctions{st: st}
	if f.unitBuildOrderValid(&Unit{}, &OrderTypeT{ID: OrdersCTFCOP2}, &UnitTypeT{}, 0) {
		t.Fatal("CTFCOP2 accepted after frame 600")
	}
}

func TestUnitOrderAllowedUnconditionalOrders(t *testing.T) {
	f := &StateFunctions{}
	for _, id := range []Orders{OrdersDie, OrdersStop, OrdersNothing, OrdersEnterTransport, OrdersWatchTarget, OrdersRightClickAction} {
		if !f.unitOrderAllowed(&Unit{}, &OrderTypeT{ID: id}) {
			t.Fatalf("order %d rejected", id)
		}
	}
	if f.unitOrderAllowed(&Unit{}, &OrderTypeT{ID: OrdersCTFCOPInit}) {
		t.Fatal("unlisted order accepted")
	}
}
