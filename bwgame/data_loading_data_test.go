package bwgame

import (
	"errors"
	"testing"
)

func TestDataReaderLittleEndian(t *testing.T) {
	r := NewDataReaderLe([]byte{0x01, 0xfe, 0xff, 0x34, 0x12, 0x78, 0x56, 0x34, 0x12, 0xff})
	if got := r.GetU8(); got != 1 {
		t.Errorf("GetU8 = %d, want 1", got)
	}
	if got := r.GetI16(); got != -2 {
		t.Errorf("GetI16 = %d, want -2", got)
	}
	if got := r.GetU16(); got != 0x1234 {
		t.Errorf("GetU16 = %#x, want 0x1234", got)
	}
	if got := r.GetU32(); got != 0x12345678 {
		t.Errorf("GetU32 = %#x, want 0x12345678", got)
	}
	if got := r.GetI8(); got != -1 {
		t.Errorf("GetI8 = %d, want -1", got)
	}
	if r.Left() != 0 || r.Tell() != 10 {
		t.Errorf("Left = %d, Tell = %d, want 0 and 10", r.Left(), r.Tell())
	}
}

func TestDataLoadersRejectWrongSize(t *testing.T) {
	_, err := LoadOrdersDat(make([]byte, 10))
	var loadingError *DataLoadingError
	if !errors.As(err, &loadingError) {
		t.Fatalf("err = %v, want a *DataLoadingError", err)
	}
	_, err = LoadTechdataDat(make([]byte, 837))
	if err == nil || err.Error() != "techdata.dat: 1 bytes left (incorrect version?)\n" {
		t.Fatalf("err = %v, want 1 byte left", err)
	}
}

func TestDataLoadersRejectInvalidBoolean(t *testing.T) {
	data := make([]byte, 1281)
	data[61*2*9+61*2] = 2
	_, err := LoadUpgradesDat(data)
	if err == nil || err.Error() != "value 0x2 is not a boolean" {
		t.Fatalf("err = %v, want a boolean error", err)
	}
}

func TestUnitsDat(t *testing.T) {
	st := testGlobalState(t)
	units, err := LoadUnitsDat(st.UnitsDat)
	if err != nil {
		t.Fatal(err)
	}
	marine := units.Vec[UnitTypesTerranMarine]
	if marine.Hitpoints != Fp8FromRaw(40<<8) {
		t.Errorf("marine hit points raw = %d, want %d", marine.Hitpoints.RawValue, 40<<8)
	}
	if marine.Flingy.ID != 78 || marine.GroundWeapon.ID != WeaponTypesGaussRifle || marine.AirWeapon.ID != WeaponTypesGaussRifle {
		t.Errorf("marine flingy %d, weapons %d/%d", marine.Flingy.ID, marine.GroundWeapon.ID, marine.AirWeapon.ID)
	}
	if marine.MineralCost != 50 || marine.BuildTime != 360 || marine.SupplyRequired != Fp1FromRaw(2) {
		t.Errorf("marine cost %d, build time %d, supply %d", marine.MineralCost, marine.BuildTime, marine.SupplyRequired.RawValue)
	}
	if marine.TurretUnitType.ID != UnitTypesNone || marine.Subunit2.ID != UnitTypesNone {
		t.Errorf("marine subunits %d/%d, want None", marine.TurretUnitType.ID, marine.Subunit2.ID)
	}
	if marine.Dimensions != (Rect{From: Xy{X: 8, Y: 9}, To: Xy{X: 8, Y: 10}}) {
		t.Errorf("marine dimensions = %+v", marine.Dimensions)
	}

	zergling := units.Vec[UnitTypesZergZergling]
	if zergling.Hitpoints != Fp8FromRaw(35<<8) || zergling.Flingy.ID != 15 || zergling.SupplyRequired != Fp1FromRaw(1) {
		t.Errorf("zergling hit points %d, flingy %d, supply %d", zergling.Hitpoints.RawValue, zergling.Flingy.ID, zergling.SupplyRequired.RawValue)
	}
	if zergling.Flags&UnitTypeTFlagOrganic == 0 || zergling.Flags&UnitTypeTFlagCanBurrow == 0 {
		t.Errorf("zergling flags = %#x, want organic and can burrow", zergling.Flags)
	}

	commandCenter := units.Vec[UnitTypesTerranCommandCenter]
	if commandCenter.Flags&UnitTypeTFlagBuilding == 0 || commandCenter.SupplyProvided != Fp1FromRaw(20) {
		t.Errorf("command center flags %#x, supply provided %d", commandCenter.Flags, commandCenter.SupplyProvided.RawValue)
	}
}

func TestFlingyDat(t *testing.T) {
	st := testGlobalState(t)
	zerglingFlingy := st.FlingyTypes.Vec[15]
	want := FlingyTypeT{ID: 15, TopSpeed: Fp8FromRaw(1), Acceleration: Fp8FromRaw(1), HaltDistance: Fp8FromRaw(1), TurnRate: Fp8FromRaw(27), Unused: 127, MovementType: 2}
	got := zerglingFlingy
	got.Sprite = TypeID[SpriteTypeT, SpriteTypes]{}
	if got != want || zerglingFlingy.Sprite.ID != 159 {
		t.Errorf("zergling flingy = %+v, want %+v with sprite 159", zerglingFlingy, want)
	}
	vulture := st.FlingyTypes.Vec[37]
	if vulture.TopSpeed != Fp8FromRaw(1280) || vulture.Acceleration != Fp8FromRaw(33) || vulture.HaltDistance != Fp8FromRaw(24824) || vulture.TurnRate != Fp8FromRaw(40) {
		t.Errorf("flingy 37 = %+v", vulture)
	}
}

func TestWeaponsDat(t *testing.T) {
	st := testGlobalState(t)
	weapons, err := LoadWeaponsDat(st.WeaponsDat)
	if err != nil {
		t.Fatal(err)
	}
	gauss := weapons.Vec[WeaponTypesGaussRifle]
	if gauss.DamageAmount != 6 || gauss.Cooldown != 15 || gauss.MaxRange != 128 || gauss.DamageType != WeaponTypeTDamageTypeNormal {
		t.Errorf("gauss rifle = %+v", gauss)
	}
}

func TestUpgradesAndTechdataDat(t *testing.T) {
	st := testGlobalState(t)
	upgrades, err := LoadUpgradesDat(st.UpgradesDat)
	if err != nil {
		t.Fatal(err)
	}
	if got := upgrades.Vec[UpgradeTypesTerranInfantryArmor]; got.MaxLevel != 3 || got.MineralCostBase != 100 {
		t.Errorf("infantry armor = %+v", got)
	}
	techs, err := LoadTechdataDat(st.TechdataDat)
	if err != nil {
		t.Fatal(err)
	}
	if got := techs.Vec[TechTypesStimPacks]; got.MineralCost != 100 || got.GasCost != 100 {
		t.Errorf("stim packs = %+v", got)
	}
}

func TestOrdersDat(t *testing.T) {
	st := testGlobalState(t)
	attackUnit := st.OrderTypes.Vec[OrdersAttackUnit]
	if !attackUnit.TargetsEnemies || !attackUnit.CanBeInterrupted {
		t.Errorf("AttackUnit order = %+v", attackUnit)
	}
	if len(st.OrderTypes.Vec) != 189 || len(st.ImageTypes.Vec) != 999 || len(st.SpriteTypes.Vec) != 517 || len(st.FlingyTypes.Vec) != 209 {
		t.Errorf("type counts: orders %d images %d sprites %d flingy %d", len(st.OrderTypes.Vec), len(st.ImageTypes.Vec), len(st.SpriteTypes.Vec), len(st.FlingyTypes.Vec))
	}
}

func TestDataReaderGetVecReadsLittleEndianValues(t *testing.T) {
	r := NewDataReaderLe([]byte{0x01, 0x00, 0xff, 0xff, 0x34, 0x12, 0x07})
	got := DataReaderGetVec[int16](&r, 3)
	want := []int16{1, -1, 0x1234}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("values = %v, want %v", got, want)
		}
	}
	if r.Tell() != 6 {
		t.Fatalf("Tell = %d, want 6", r.Tell())
	}
	_, err := func() (_ []uint16, err error) {
		defer recoverDataLoadingError(&err)
		return DataReaderGetVec[uint16](&r, 1), nil
	}()
	if err == nil || err.Error() != "data_reader: attempt to read past end" {
		t.Fatalf("err = %v, want a read past end error", err)
	}
}
