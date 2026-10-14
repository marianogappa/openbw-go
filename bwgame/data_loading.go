package bwgame

import (
	"errors"
	"fmt"
	"unsafe"
)

type DataLoadingError struct {
	Message string
}

func (e *DataLoadingError) Error() string {
	return e.Message
}

func throwDataLoadingError(format string, args ...any) {
	panic(&DataLoadingError{Message: fmt.Sprintf(format, args...)})
}

func recoverDataLoadingError(err *error) {
	recovered := recover()
	if recovered == nil {
		return
	}
	var loadingError *DataLoadingError
	if asError, ok := recovered.(error); ok && errors.As(asError, &loadingError) {
		*err = loadingError
		return
	}
	panic(recovered)
}

type DataReader struct {
	data []byte
	ptr  uint
}

func NewDataReaderLe(data []byte) DataReader {
	return DataReader{data: data}
}

func (r *DataReader) GetN(n uint) []byte {
	if r.Left() < n {
		throwDataLoadingError("data_reader: attempt to read past end")
	}
	result := r.data[r.ptr : r.ptr+n]
	r.ptr += n
	return result
}

func (r *DataReader) GetU8() uint8 {
	return r.GetN(1)[0]
}

func (r *DataReader) GetI8() int8 {
	return int8(r.GetU8())
}

func (r *DataReader) GetU16() uint16 {
	return valueAtLe[uint16](r.GetN(2))
}

func (r *DataReader) GetI16() int16 {
	return valueAtLe[int16](r.GetN(2))
}

func (r *DataReader) GetU32() uint32 {
	return valueAtLe[uint32](r.GetN(4))
}

func (r *DataReader) GetI32() int32 {
	return valueAtLe[int32](r.GetN(4))
}

func DataReaderGetVec[T loadInteger](r *DataReader, n uint) []T {
	size := uint(unsafe.Sizeof(T(0)))
	data := r.GetN(n * size)
	values := make([]T, n)
	for i := range values {
		values[i] = valueAtLe[T](data[uint(i)*size:])
	}
	return values
}

func (r *DataReader) GetBytes(dst []byte) {
	copy(dst, r.GetN(uint(len(dst))))
}

func (r *DataReader) Skip(n uint) {
	if r.Left() < n {
		throwDataLoadingError("data_reader: attempt to seek past end")
	}
	r.ptr += n
}

func (r *DataReader) Seek(offset uint) {
	if offset > r.Size() {
		throwDataLoadingError("data_reader: attempt to seek past end")
	}
	r.ptr = offset
}

func (r *DataReader) Left() uint {
	return r.Size() - r.ptr
}

func (r *DataReader) Size() uint {
	return uint(len(r.data))
}

func (r *DataReader) Tell() uint {
	return r.ptr
}

type loadInteger interface {
	~uint8 | ~int8 | ~uint16 | ~int16 | ~uint32 | ~int32
}

func valueAtLe[T loadInteger](bytes []byte) T {
	var raw uint64
	for i := uintptr(0); i < unsafe.Sizeof(T(0)); i++ {
		raw |= uint64(bytes[i]) << (i * 8)
	}
	return T(raw)
}

func readArray[L loadInteger](r *DataReader, num int, set func(index int, value L)) {
	readArrayN(r, num, 1, func(index int, values []L) { set(index, values[0]) })
}

func readArrayN[L loadInteger](r *DataReader, num int, readN int, set func(index int, values []L)) {
	loadSize := int(unsafe.Sizeof(L(0)))
	data := r.GetN(uint(loadSize * readN * num))
	values := make([]L, readN)
	for i := 0; i < num; i++ {
		element := data[i*loadSize*readN:]
		for j := range values {
			values[j] = valueAtLe[L](element[j*loadSize:])
		}
		set(i, values)
	}
}

func readXyArray[L loadInteger](r *DataReader, num int, set func(index int, value Xy)) {
	readArrayN(r, num, 2, func(index int, v []L) {
		set(index, Xy{X: int32(v[0]), Y: int32(v[1])})
	})
}

func readRectArray[L loadInteger](r *DataReader, num int, set func(index int, value Rect)) {
	readArrayN(r, num, 4, func(index int, v []L) {
		set(index, Rect{From: Xy{X: int32(v[0]), Y: int32(v[1])}, To: Xy{X: int32(v[2]), Y: int32(v[3])}})
	})
}

func dataTypeCastBool[L loadInteger](v L) bool {
	if v != 0 && v != 1 {
		throwDataLoadingError("value 0x%x is not a boolean", v)
	}
	return v != 0
}

func checkNothingLeft(r *DataReader, filename string) {
	if r.Left() != 0 {
		throwDataLoadingError("%s: %d bytes left (incorrect version?)\n", filename, r.Left())
	}
}

func newTypeContainer[T any](count int, initialize func(index int, v *T)) TypeContainer[T] {
	container := TypeContainer[T]{Vec: make([]T, count)}
	for i := range container.Vec {
		initialize(i, &container.Vec[i])
	}
	return container
}

func LoadUnitsDat(data []byte) (unitTypes UnitTypesT, err error) {
	defer recoverDataLoadingError(&err)
	const totalCount = 228
	const unitsCount = 106
	const buildingsCount = 96

	unitTypes = newTypeContainer(totalCount, func(i int, v *UnitTypeT) { v.ID = UnitTypes(i) })
	r := NewDataReaderLe(data)
	arr := unitTypes.Vec

	readArray(&r, totalCount, func(i int, v uint8) { arr[i].Flingy.ID = FlingyTypes(v) })
	readArray(&r, totalCount, func(i int, v uint16) { arr[i].TurretUnitType.ID = UnitTypes(v) })
	readArray(&r, totalCount, func(i int, v uint16) { arr[i].Subunit2.ID = UnitTypes(v) })
	readArray(&r, buildingsCount, func(i int, v uint16) { arr[unitsCount+i].InfestationUnit.ID = UnitTypes(v) })
	readArray(&r, totalCount, func(i int, v uint32) { arr[i].ConstructionAnimation.ID = ImageTypes(int32(v)) })
	readArray(&r, totalCount, func(i int, v uint8) { arr[i].UnitDirection = int32(v) })
	readArray(&r, totalCount, func(i int, v uint8) { arr[i].HasShield = dataTypeCastBool(v) })
	readArray(&r, totalCount, func(i int, v uint16) { arr[i].ShieldPoints = int32(v) })
	readArray(&r, totalCount, func(i int, v int32) { arr[i].Hitpoints = Fp8FromRaw(v) })
	readArray(&r, totalCount, func(i int, v uint8) { arr[i].ElevationLevel = int32(v) })
	readArray(&r, totalCount, func(i int, v uint8) { arr[i].Unknown1 = int32(v) })
	readArray(&r, totalCount, func(i int, v uint8) { arr[i].Sublabel = int32(v) })
	readArray(&r, totalCount, func(i int, v uint8) { arr[i].ComputerAiIdle.ID = Orders(v) })
	readArray(&r, totalCount, func(i int, v uint8) { arr[i].HumanAiIdle.ID = Orders(v) })
	readArray(&r, totalCount, func(i int, v uint8) { arr[i].ReturnToIdle.ID = Orders(v) })
	readArray(&r, totalCount, func(i int, v uint8) { arr[i].AttackUnit.ID = Orders(v) })
	readArray(&r, totalCount, func(i int, v uint8) { arr[i].AttackMove.ID = Orders(v) })
	readArray(&r, totalCount, func(i int, v uint8) { arr[i].GroundWeapon.ID = WeaponTypes(v) })
	readArray(&r, totalCount, func(i int, v uint8) { arr[i].MaxGroundHits = int32(v) })
	readArray(&r, totalCount, func(i int, v uint8) { arr[i].AirWeapon.ID = WeaponTypes(v) })
	readArray(&r, totalCount, func(i int, v uint8) { arr[i].MaxAirHits = int32(v) })
	readArray(&r, totalCount, func(i int, v uint8) { arr[i].AiInternal = int32(v) })
	readArray(&r, totalCount, func(i int, v uint32) { arr[i].Flags = UnitTypeFlagsT(v) })
	readArray(&r, totalCount, func(i int, v uint8) { arr[i].TargetAcquisitionRange = int32(v) })
	readArray(&r, totalCount, func(i int, v uint8) { arr[i].SightRange = int32(v) })
	readArray(&r, totalCount, func(i int, v uint8) { arr[i].ArmorUpgrade.ID = UpgradeTypes(v) })
	readArray(&r, totalCount, func(i int, v uint8) { arr[i].UnitSize = int32(v) })
	readArray(&r, totalCount, func(i int, v uint8) { arr[i].Armor = int32(v) })
	readArray(&r, totalCount, func(i int, v uint8) { arr[i].RightClickAction = uint(v) })
	readArray(&r, unitsCount, func(i int, v uint16) { arr[i].ReadySound = int32(v) })
	readArray(&r, totalCount, func(i int, v uint16) { arr[i].FirstWhatSound = int32(v) })
	readArray(&r, totalCount, func(i int, v uint16) { arr[i].LastWhatSound = int32(v) })
	readArray(&r, unitsCount, func(i int, v uint16) { arr[i].FirstPissedSound = int32(v) })
	readArray(&r, unitsCount, func(i int, v uint16) { arr[i].LastPissedSound = int32(v) })
	readArray(&r, unitsCount, func(i int, v uint16) { arr[i].FirstYesSound = int32(v) })
	readArray(&r, unitsCount, func(i int, v uint16) { arr[i].LastYesSound = int32(v) })
	readXyArray[int16](&r, totalCount, func(i int, v Xy) { arr[i].PlacementSize = v })
	readXyArray[int16](&r, buildingsCount, func(i int, v Xy) { arr[unitsCount+i].AddonPosition = v })
	readRectArray[int16](&r, totalCount, func(i int, v Rect) { arr[i].Dimensions = v })
	readArray(&r, totalCount, func(i int, v uint16) { arr[i].Portrait = int32(v) })
	readArray(&r, totalCount, func(i int, v uint16) { arr[i].MineralCost = int32(v) })
	readArray(&r, totalCount, func(i int, v uint16) { arr[i].GasCost = int32(v) })
	readArray(&r, totalCount, func(i int, v uint16) { arr[i].BuildTime = int32(v) })
	readArray(&r, totalCount, func(i int, v uint16) { arr[i].Unknown2 = int32(v) })
	readArray(&r, totalCount, func(i int, v uint8) { arr[i].GroupFlags = int32(v) })
	readArray(&r, totalCount, func(i int, v uint8) { arr[i].SupplyProvided = Fp1FromRaw(int32(v)) })
	readArray(&r, totalCount, func(i int, v uint8) { arr[i].SupplyRequired = Fp1FromRaw(int32(v)) })
	readArray(&r, totalCount, func(i int, v uint8) { arr[i].SpaceRequired = uint(v) })
	readArray(&r, totalCount, func(i int, v uint8) { arr[i].SpaceProvided = uint(v) })
	readArray(&r, totalCount, func(i int, v uint16) { arr[i].BuildScore = int32(v) })
	readArray(&r, totalCount, func(i int, v uint16) { arr[i].DestroyScore = int32(v) })
	readArray(&r, totalCount, func(i int, v uint16) { arr[i].UnitMapStringIndex = uint(v) })
	readArray(&r, totalCount, func(i int, v uint8) { arr[i].IsBroodwar = dataTypeCastBool(v) })
	readArray(&r, totalCount, func(i int, v uint16) { arr[i].StareditAvailabilityFlags = int32(v) })

	checkNothingLeft(&r, "units.dat")
	return unitTypes, nil
}

func LoadWeaponsDat(data []byte) (weaponTypes WeaponTypesT, err error) {
	defer recoverDataLoadingError(&err)
	const count = 130

	weaponTypes = newTypeContainer(count, func(i int, v *WeaponTypeT) { v.ID = WeaponTypes(i) })
	r := NewDataReaderLe(data)
	arr := weaponTypes.Vec

	readArray(&r, count, func(i int, v uint16) { arr[i].Label = int32(v) })
	readArray(&r, count, func(i int, v uint32) { arr[i].Flingy.ID = FlingyTypes(int32(v)) })
	readArray(&r, count, func(i int, v uint8) { arr[i].Unused = int32(v) })
	readArray(&r, count, func(i int, v uint16) { arr[i].TargetFlags = int32(v) })
	readArray(&r, count, func(i int, v uint32) { arr[i].MinRange = int32(v) })
	readArray(&r, count, func(i int, v uint32) { arr[i].MaxRange = int32(v) })
	readArray(&r, count, func(i int, v uint8) { arr[i].DamageUpgrade.ID = UpgradeTypes(v) })
	readArray(&r, count, func(i int, v uint8) { arr[i].DamageType = int32(v) })
	readArray(&r, count, func(i int, v uint8) { arr[i].BulletType = int32(v) })
	readArray(&r, count, func(i int, v uint8) { arr[i].Lifetime = int32(v) })
	readArray(&r, count, func(i int, v uint8) { arr[i].HitType = int32(v) })
	readArray(&r, count, func(i int, v uint16) { arr[i].InnerSplashRadius = int32(v) })
	readArray(&r, count, func(i int, v uint16) { arr[i].MediumSplashRadius = int32(v) })
	readArray(&r, count, func(i int, v uint16) { arr[i].OuterSplashRadius = int32(v) })
	readArray(&r, count, func(i int, v uint16) { arr[i].DamageAmount = int32(v) })
	readArray(&r, count, func(i int, v uint16) { arr[i].DamageBonus = int32(v) })
	readArray(&r, count, func(i int, v uint8) { arr[i].Cooldown = int32(v) })
	readArray(&r, count, func(i int, v uint8) { arr[i].BulletCount = int32(v) })
	readArray(&r, count, func(i int, v uint8) { arr[i].AttackAngle = Fp8FromRaw(int32(v)) })
	readArray(&r, count, func(i int, v int8) { arr[i].BulletHeadingOffset = DirectionTFromRaw(v) })
	readArray(&r, count, func(i int, v int8) { arr[i].ForwardOffset = int32(v) })
	readArray(&r, count, func(i int, v int8) { arr[i].UpwardOffset = int32(v) })
	readArray(&r, count, func(i int, v uint16) { arr[i].TargetErrorMessage = int32(v) })
	readArray(&r, count, func(i int, v uint16) { arr[i].Icon = int32(v) })

	checkNothingLeft(&r, "weapons.dat")
	return weaponTypes, nil
}

func LoadUpgradesDat(data []byte) (upgradeTypes UpgradeTypesT, err error) {
	defer recoverDataLoadingError(&err)
	const count = 61

	upgradeTypes = newTypeContainer(count, func(i int, v *UpgradeTypeT) { v.ID = UpgradeTypes(i) })
	r := NewDataReaderLe(data)
	arr := upgradeTypes.Vec

	readArray(&r, count, func(i int, v uint16) { arr[i].MineralCostBase = int32(v) })
	readArray(&r, count, func(i int, v uint16) { arr[i].MineralCostFactor = int32(v) })
	readArray(&r, count, func(i int, v uint16) { arr[i].GasCostBase = int32(v) })
	readArray(&r, count, func(i int, v uint16) { arr[i].GasCostFactor = int32(v) })
	readArray(&r, count, func(i int, v uint16) { arr[i].TimeCostBase = int32(v) })
	readArray(&r, count, func(i int, v uint16) { arr[i].TimeCostFactor = int32(v) })
	readArray(&r, count, func(i int, v uint16) { arr[i].Unknown = int32(v) })
	readArray(&r, count, func(i int, v uint16) { arr[i].Icon = int32(v) })
	readArray(&r, count, func(i int, v uint16) { arr[i].Label = int32(v) })
	readArray(&r, count, func(i int, v uint8) { arr[i].Race = int32(v) })
	readArray(&r, count, func(i int, v uint8) { arr[i].MaxLevel = int32(v) })
	readArray(&r, count, func(i int, v uint8) { arr[i].IsBroodwar = dataTypeCastBool(v) })

	checkNothingLeft(&r, "upgrades.dat")
	return upgradeTypes, nil
}

func LoadTechdataDat(data []byte) (techTypes TechTypesT, err error) {
	defer recoverDataLoadingError(&err)
	const count = 44

	techTypes = newTypeContainer(count, func(i int, v *TechTypeT) { v.ID = TechTypes(i) })
	r := NewDataReaderLe(data)
	arr := techTypes.Vec

	readArray(&r, count, func(i int, v uint16) { arr[i].MineralCost = int32(v) })
	readArray(&r, count, func(i int, v uint16) { arr[i].GasCost = int32(v) })
	readArray(&r, count, func(i int, v uint16) { arr[i].ResearchTime = int32(v) })
	readArray(&r, count, func(i int, v uint16) { arr[i].EnergyCost = int32(v) })
	readArray(&r, count, func(i int, v uint32) { arr[i].Unknown = int32(v) })
	readArray(&r, count, func(i int, v uint16) { arr[i].Icon = int32(v) })
	readArray(&r, count, func(i int, v uint16) { arr[i].Label = int32(v) })
	readArray(&r, count, func(i int, v uint8) { arr[i].Race = int32(v) })
	readArray(&r, count, func(i int, v uint16) { arr[i].Flags = int32(v) })

	checkNothingLeft(&r, "techdata.dat")
	return techTypes, nil
}

func LoadFlingyDat(data []byte) (flingyTypes FlingyTypesT, err error) {
	defer recoverDataLoadingError(&err)
	const count = 209

	flingyTypes = newTypeContainer(count, func(i int, v *FlingyTypeT) { v.ID = FlingyTypes(i) })
	r := NewDataReaderLe(data)
	arr := flingyTypes.Vec

	readArray(&r, count, func(i int, v uint16) { arr[i].Sprite.ID = SpriteTypes(v) })
	readArray(&r, count, func(i int, v int32) { arr[i].TopSpeed = Fp8FromRaw(v) })
	readArray(&r, count, func(i int, v int16) { arr[i].Acceleration = Fp8FromRaw(int32(v)) })
	readArray(&r, count, func(i int, v int32) { arr[i].HaltDistance = Fp8FromRaw(v) })
	readArray(&r, count, func(i int, v uint8) { arr[i].TurnRate = Fp8FromRaw(int32(v)) })
	readArray(&r, count, func(i int, v uint8) { arr[i].Unused = int32(v) })
	readArray(&r, count, func(i int, v uint8) { arr[i].MovementType = int32(v) })

	checkNothingLeft(&r, "flingy.dat")
	return flingyTypes, nil
}

func LoadSpritesDat(data []byte) (spriteTypes SpriteTypesT, err error) {
	defer recoverDataLoadingError(&err)
	const count = 517
	const selectableCount = 387
	const nonSelectableCount = 130

	spriteTypes = newTypeContainer(count, func(i int, v *SpriteTypeT) { v.ID = SpriteTypes(i) })
	r := NewDataReaderLe(data)
	arr := spriteTypes.Vec

	readArray(&r, count, func(i int, v uint16) { arr[i].Image.ID = ImageTypes(v) })
	readArray(&r, selectableCount, func(i int, v uint8) { arr[nonSelectableCount+i].HealthBarSize = int32(v) })
	readArray(&r, count, func(i int, v uint8) { arr[i].Unk0 = int32(v) })
	readArray(&r, count, func(i int, v uint8) { arr[i].Visible = dataTypeCastBool(v) })
	readArray(&r, selectableCount, func(i int, v uint8) { arr[nonSelectableCount+i].SelectionCircle = int32(v) })
	readArray(&r, selectableCount, func(i int, v uint8) { arr[nonSelectableCount+i].SelectionCircleVpos = int32(v) })

	checkNothingLeft(&r, "sprites.dat")
	return spriteTypes, nil
}

func LoadImagesDat(data []byte) (imageTypes ImageTypesT, err error) {
	defer recoverDataLoadingError(&err)
	const count = 999

	imageTypes = newTypeContainer(count, func(i int, v *ImageTypeT) { v.ID = ImageTypes(i) })
	r := NewDataReaderLe(data)
	arr := imageTypes.Vec

	readArray(&r, count, func(i int, v uint32) { arr[i].GrpFilenameIndex = int32(v) })
	readArray(&r, count, func(i int, v uint8) { arr[i].HasDirectionalFrames = dataTypeCastBool(v) })
	readArray(&r, count, func(i int, v uint8) { arr[i].IsClickable = dataTypeCastBool(v) })
	readArray(&r, count, func(i int, v uint8) { arr[i].HasIscriptAnimations = dataTypeCastBool(v) })
	readArray(&r, count, func(i int, v uint8) { arr[i].AlwaysVisible = dataTypeCastBool(v) })
	readArray(&r, count, func(i int, v uint8) { arr[i].Modifier = int32(v) })
	readArray(&r, count, func(i int, v uint8) { arr[i].ColorShift = int32(v) })
	readArray(&r, count, func(i int, v uint32) { arr[i].IscriptID = int32(v) })
	readArray(&r, count, func(i int, v uint32) { arr[i].ShieldFilenameIndex = int32(v) })
	readArray(&r, count, func(i int, v uint32) { arr[i].AttackFilenameIndex = int32(v) })
	readArray(&r, count, func(i int, v uint32) { arr[i].DamageFilenameIndex = int32(v) })
	readArray(&r, count, func(i int, v uint32) { arr[i].SpecialFilenameIndex = int32(v) })
	readArray(&r, count, func(i int, v uint32) { arr[i].LandingDustFilenameIndex = int32(v) })
	readArray(&r, count, func(i int, v uint32) { arr[i].LiftOffFilenameIndex = int32(v) })

	checkNothingLeft(&r, "images.dat")
	return imageTypes, nil
}

func LoadOrdersDat(data []byte) (orderTypes OrderTypesT, err error) {
	defer recoverDataLoadingError(&err)
	const count = 189

	orderTypes = newTypeContainer(count, func(i int, v *OrderTypeT) { v.ID = Orders(i) })
	r := NewDataReaderLe(data)
	arr := orderTypes.Vec

	readArray(&r, count, func(i int, v uint16) { arr[i].Label = int32(v) })
	readArray(&r, count, func(i int, v uint8) { arr[i].TargetsEnemies = dataTypeCastBool(v) })
	readArray(&r, count, func(i int, v uint8) { arr[i].Background = int32(v) })
	readArray(&r, count, func(i int, v uint8) { arr[i].Unused3 = int32(v) })
	readArray(&r, count, func(i int, v uint8) { arr[i].ValidForTurret = dataTypeCastBool(v) })
	readArray(&r, count, func(i int, v uint8) { arr[i].Unused5 = int32(v) })
	readArray(&r, count, func(i int, v uint8) { arr[i].CanBeInterrupted = dataTypeCastBool(v) })
	readArray(&r, count, func(i int, v uint8) { arr[i].Unk7 = dataTypeCastBool(v) })
	readArray(&r, count, func(i int, v uint8) { arr[i].CanBeQueued = dataTypeCastBool(v) })
	readArray(&r, count, func(i int, v uint8) { arr[i].Unk9 = int32(v) })
	readArray(&r, count, func(i int, v uint8) { arr[i].CanBeObstructed = dataTypeCastBool(v) })
	readArray(&r, count, func(i int, v uint8) { arr[i].Unk11 = int32(v) })
	readArray(&r, count, func(i int, v uint8) { arr[i].Unused12 = int32(v) })
	readArray(&r, count, func(i int, v uint8) { arr[i].Weapon = WeaponTypes(v) })
	readArray(&r, count, func(i int, v uint8) { arr[i].TechType = TechTypes(v) })
	readArray(&r, count, func(i int, v uint8) { arr[i].Animation = int32(v) })
	readArray(&r, count, func(i int, v int16) { arr[i].Highlight = int32(v) })
	readArray(&r, count, func(i int, v uint16) { arr[i].DepIndex = int32(v) })
	readArray(&r, count, func(i int, v uint8) { arr[i].TargetOrder = Orders(v) })

	checkNothingLeft(&r, "orders.dat")
	return orderTypes, nil
}
