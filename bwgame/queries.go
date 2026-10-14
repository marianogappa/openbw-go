package bwgame

import (
	"iter"
	"sort"
)

func (f *StateFunctions) getOrderType(id Orders) *OrderTypeT {
	if uint(id) >= 189 {
		throwError("invalid order id %d", uint(id))
	}
	return &f.globalSt.OrderTypes.Vec[id]
}

func (f *StateFunctions) uSetStatusFlag(u *Unit, flag int32) {
	u.StatusFlags |= flag
}

func (f *StateFunctions) uUnsetStatusFlag(u *Unit, flag int32) {
	u.StatusFlags &^= flag
}

func (f *StateFunctions) uSetStatusFlagValue(u *Unit, flag int32, value bool) {
	if value {
		u.StatusFlags |= flag
	} else {
		u.StatusFlags &^= flag
	}
}

func (f *StateFunctions) uSetMovementFlag(u *Flingy, flag int32) {
	u.MovementFlags |= flag
}

func (f *StateFunctions) uUnsetMovementFlag(u *Flingy, flag int32) {
	u.MovementFlags &^= flag
}

func (f *StateFunctions) iSetFlag(i *Image, flag int32) {
	i.Flags |= flag
}

func (f *StateFunctions) iUnsetFlag(i *Image, flag int32) {
	i.Flags &^= flag
}

func (f *StateFunctions) iSetFlagValue(i *Image, flag int32, value bool) {
	if value {
		f.iSetFlag(i, flag)
	} else {
		f.iUnsetFlag(i, flag)
	}
}

func (f *StateFunctions) utFlag(ut *UnitTypeT, flag UnitTypeFlagsT) bool {
	return ut.Flags&flag != 0
}

func (f *StateFunctions) uStatusFlag(u *Unit, flag int32) bool {
	return u.StatusFlags&flag != 0
}

func (f *StateFunctions) usFlag(u *Thingy, flag int32) bool {
	return u.Sprite.Flags&flag != 0
}

func (f *StateFunctions) sFlag(s *Sprite, flag int32) bool {
	return s.Flags&flag != 0
}

func (f *StateFunctions) uMovementFlag(u *Flingy, flag int32) bool {
	return u.MovementFlags&flag != 0
}

func (f *StateFunctions) iFlag(i *Image, flag int32) bool {
	return i.Flags&flag != 0
}

func (f *StateFunctions) uCompleted(u *Unit) bool {
	return f.uStatusFlag(u, UnitStatusFlagCompleted)
}

func (f *StateFunctions) uInBunker(u *Unit) bool {
	return f.uStatusFlag(u, UnitStatusFlagInBunker)
}

func (f *StateFunctions) uLoaded(u *Unit) bool {
	return f.uStatusFlag(u, UnitStatusFlagLoaded)
}

func (f *StateFunctions) uImmovable(u *Unit) bool {
	return f.uStatusFlag(u, UnitStatusFlagImmovable)
}

func (f *StateFunctions) uDisabled(u *Unit) bool {
	return f.uStatusFlag(u, UnitStatusFlagDisabled)
}

func (f *StateFunctions) uBurrowed(u *Unit) bool {
	return f.uStatusFlag(u, UnitStatusFlagBurrowed)
}

func (f *StateFunctions) uCanTurn(u *Unit) bool {
	return f.uStatusFlag(u, UnitStatusFlagCanTurn)
}

func (f *StateFunctions) uCanMove(u *Unit) bool {
	return f.uStatusFlag(u, UnitStatusFlagCanMove)
}

func (f *StateFunctions) uGroundedBuilding(u *Unit) bool {
	return f.uStatusFlag(u, UnitStatusFlagGroundedBuilding)
}

func (f *StateFunctions) uHallucination(u *Unit) bool {
	return f.uStatusFlag(u, UnitStatusFlagHallucination)
}

func (f *StateFunctions) uLifetimeExpired(u *Unit) bool {
	return f.uStatusFlag(u, UnitStatusFlagLifetimeExpired)
}

func (f *StateFunctions) uFlying(u *Unit) bool {
	return f.uStatusFlag(u, UnitStatusFlagFlying)
}

func (f *StateFunctions) uSpeedUpgrade(u *Unit) bool {
	return f.uStatusFlag(u, UnitStatusFlagSpeedUpgrade)
}

func (f *StateFunctions) uCooldownUpgrade(u *Unit) bool {
	return f.uStatusFlag(u, UnitStatusFlagCooldownUpgrade)
}

func (f *StateFunctions) uGathering(u *Unit) bool {
	return f.uStatusFlag(u, UnitStatusFlagGathering)
}

func (f *StateFunctions) uRequiresDetector(u *Unit) bool {
	return f.uStatusFlag(u, UnitStatusFlagRequiresDetector)
}

func (f *StateFunctions) uCloaked(u *Unit) bool {
	return f.uStatusFlag(u, UnitStatusFlagCloaked)
}

func (f *StateFunctions) uCannotAttack(u *Unit) bool {
	return f.uStatusFlag(u, UnitStatusFlagCannotAttack)
}

func (f *StateFunctions) uOrderNotInterruptible(u *Unit) bool {
	return f.uStatusFlag(u, UnitStatusFlagOrderNotInterruptible)
}

func (f *StateFunctions) uIscriptNobrk(u *Unit) bool {
	return f.uStatusFlag(u, UnitStatusFlagIscriptNobrk)
}

func (f *StateFunctions) uCollision(u *Unit) bool {
	return f.uStatusFlag(u, UnitStatusFlagCollision)
}

func (f *StateFunctions) uGroundUnit(u *Unit) bool {
	return f.uStatusFlag(u, UnitStatusFlagGroundUnit)
}

func (f *StateFunctions) uNoCollide(u *Unit) bool {
	return f.uStatusFlag(u, UnitStatusFlagNoCollide)
}

func (f *StateFunctions) uInvincible(u *Unit) bool {
	return f.uStatusFlag(u, UnitStatusFlagInvincible)
}

func (f *StateFunctions) uReadyToAttack(u *Unit) bool {
	return f.uStatusFlag(u, UnitStatusFlagReadyToAttack)
}

func (f *StateFunctions) uPassivelyCloaked(u *Unit) bool {
	return f.uStatusFlag(u, UnitStatusFlagPassivelyCloaked)
}

func (f *StateFunctions) utTurret(ut *UnitTypeT) bool {
	return f.utFlag(ut, UnitTypeTFlagTurret)
}

func (f *StateFunctions) utWorker(ut *UnitTypeT) bool {
	return f.utFlag(ut, UnitTypeTFlagWorker)
}

func (f *StateFunctions) utHero(ut *UnitTypeT) bool {
	return f.utFlag(ut, UnitTypeTFlagHero)
}

func (f *StateFunctions) utBuilding(ut *UnitTypeT) bool {
	return f.utFlag(ut, UnitTypeTFlagBuilding)
}

func (f *StateFunctions) utAddon(ut *UnitTypeT) bool {
	return f.utFlag(ut, UnitTypeTFlagAddon)
}

func (f *StateFunctions) utFlyer(ut *UnitTypeT) bool {
	return f.utFlag(ut, UnitTypeTFlagFlyer)
}

func (f *StateFunctions) utCanTurn(ut *UnitTypeT) bool {
	return f.utFlag(ut, UnitTypeTFlagCanTurn)
}

func (f *StateFunctions) utCanMove(ut *UnitTypeT) bool {
	return f.utFlag(ut, UnitTypeTFlagCanMove)
}

func (f *StateFunctions) utInvincible(ut *UnitTypeT) bool {
	return f.utFlag(ut, UnitTypeTFlagInvincible)
}

func (f *StateFunctions) utTwoUnitsInOneEgg(ut *UnitTypeT) bool {
	return f.utFlag(ut, UnitTypeTFlagTwoUnitsInOneEgg)
}

func (f *StateFunctions) utRegensHp(ut *UnitTypeT) bool {
	return f.utFlag(ut, UnitTypeTFlagRegensHp)
}

func (f *StateFunctions) utFlyingBuilding(ut *UnitTypeT) bool {
	return f.utFlag(ut, UnitTypeTFlagFlyingBuilding)
}

func (f *StateFunctions) utRequiresPsionicMatrix(ut *UnitTypeT) bool {
	return f.utFlag(ut, UnitTypeTFlagRequiresPsionicMatrix)
}

func (f *StateFunctions) utCanBurrow(ut *UnitTypeT) bool {
	return f.utFlag(ut, UnitTypeTFlagCanBurrow)
}

func (f *StateFunctions) utHasEnergy(ut *UnitTypeT) bool {
	return f.utFlag(ut, UnitTypeTFlagHasEnergy)
}

func (f *StateFunctions) utResourceDepot(ut *UnitTypeT) bool {
	return f.utFlag(ut, UnitTypeTFlagResourceDepot)
}

func (f *StateFunctions) utResource(ut *UnitTypeT) bool {
	return f.utFlag(ut, UnitTypeTFlagResource)
}

func (f *StateFunctions) utInitiallyCloaked(ut *UnitTypeT) bool {
	return f.utFlag(ut, UnitTypeTFlagInitiallyCloaked)
}

func (f *StateFunctions) utDetector(ut *UnitTypeT) bool {
	return f.utFlag(ut, UnitTypeTFlagDetector)
}

func (f *StateFunctions) utRequiresCreep(ut *UnitTypeT) bool {
	return f.utFlag(ut, UnitTypeTFlagRequiresCreep)
}

func (f *StateFunctions) utMechanical(ut *UnitTypeT) bool {
	return f.utFlag(ut, UnitTypeTFlagMechanical)
}

func (f *StateFunctions) utOrganic(ut *UnitTypeT) bool {
	return f.utFlag(ut, UnitTypeTFlagOrganic)
}

func (f *StateFunctions) utRobotic(ut *UnitTypeT) bool {
	return f.utFlag(ut, UnitTypeTFlagRobotic)
}

func (f *StateFunctions) utPowerup(ut *UnitTypeT) bool {
	return f.utFlag(ut, UnitTypeTFlagPowerup)
}

func (f *StateFunctions) usHidden(u *Thingy) bool {
	return f.usFlag(u, SpriteFlagHidden)
}

func (f *StateFunctions) sHidden(u *Sprite) bool {
	return f.sFlag(u, SpriteFlagHidden)
}

func (f *StateFunctions) getUnitType(id UnitTypes) *UnitTypeT {
	if uint(id) >= 228 {
		throwError("invalid unit id %d", uint(id))
	}
	return &f.gameSt.UnitTypes.Vec[id]
}

func (f *StateFunctions) getImageType(id ImageTypes) *ImageTypeT {
	if uint(id) >= 999 {
		throwError("invalid image id %d", uint(id))
	}
	return &f.globalSt.ImageTypes.Vec[id]
}

func (f *StateFunctions) getWeaponType(id WeaponTypes) *WeaponTypeT {
	if uint(id) >= 130 {
		throwError("invalid weapon id %d", uint(id))
	}
	return &f.gameSt.WeaponTypes.Vec[id]
}

func (f *StateFunctions) getSpriteType(id SpriteTypes) *SpriteTypeT {
	if uint(id) >= 517 {
		throwError("invalid sprite id %d", uint(id))
	}
	return &f.globalSt.SpriteTypes.Vec[id]
}

func (f *StateFunctions) getUpgradeType(id UpgradeTypes) *UpgradeTypeT {
	if uint(id) >= 61 {
		throwError("invalid upgrade id %d", uint(id))
	}
	return &f.gameSt.UpgradeTypes.Vec[id]
}

func (f *StateFunctions) getTechType(id TechTypes) *TechTypeT {
	if uint(id) >= 44 {
		throwError("invalid tech id %d", uint(id))
	}
	return &f.gameSt.TechTypes.Vec[id]
}

func (f *StateFunctions) getFlingyType(id FlingyTypes) *FlingyTypeT {
	if uint(id) >= 209 {
		throwError("invalid flingy id %d", uint(id))
	}
	return &f.globalSt.FlingyTypes.Vec[id]
}

func (f *StateFunctions) playSoundUnit(id int32, sourceUnit *Unit, addRaceIndex bool) {
	var position Xy
	if sourceUnit != nil {
		position = sourceUnit.Sprite.Position
	}
	f.playSoundAt(id, position, sourceUnit, addRaceIndex)
}

func (f *StateFunctions) playSound(id int32, addRaceIndex bool) {
	f.playSoundAt(id, Xy{}, nil, addRaceIndex)
}

func (f *StateFunctions) getUnit(index uint) *Unit {
	u := f.st.UnitsContainer.TryGet(index)
	if u == nil {
		return nil
	}
	if f.unitDead(u) {
		return nil
	}
	return u
}

func (f *StateFunctions) getUnitByID(id UnitID) *Unit {
	idx := id.Index()
	if idx == 0 {
		return nil
	}
	u := f.getUnit(idx - 1)
	if u == nil {
		return nil
	}
	if u.UnitIDGeneration%(1<<5) != id.Generation() {
		return nil
	}
	return u
}

func (f *StateFunctions) getUnitID(u *Unit) UnitID {
	if u == nil {
		return UnitID{}
	}
	return NewUnitID(u.Index+1, u.UnitIDGeneration%(1<<5))
}

func (f *StateFunctions) getUnitID32(u *Unit) UnitID32 {
	if u == nil {
		return UnitID32{}
	}
	return NewUnitID32(u.Index+1, u.UnitIDGeneration%(1<<21))
}

func (f *StateFunctions) isUnitTypeInMapBounds(unitType *UnitTypeT, pos Xy) bool {
	if pos.X-unitType.Dimensions.From.X < 0 {
		return false
	}
	if pos.Y-unitType.Dimensions.From.Y < 0 {
		return false
	}
	if uint(pos.X+unitType.Dimensions.To.X) >= f.gameSt.MapWidth {
		return false
	}
	if uint(pos.Y+unitType.Dimensions.To.Y) >= f.gameSt.MapHeight {
		return false
	}
	return true
}

func (f *StateFunctions) isInMapBounds(area Rect) bool {
	if area.From.X < 0 {
		return false
	}
	if uint(area.To.X) > f.gameSt.MapWidth {
		return false
	}
	if area.From.Y < 0 {
		return false
	}
	if uint(area.To.Y) > f.gameSt.MapHeight {
		return false
	}
	return true
}

func (f *StateFunctions) isInInnerMapBounds(area Rect) bool {
	if area.From.X < 0 {
		return false
	}
	if uint(area.To.X) >= f.gameSt.MapWidth {
		return false
	}
	if area.From.Y < 0 {
		return false
	}
	if uint(area.To.Y) >= f.gameSt.MapHeight {
		return false
	}
	return true
}

func (f *StateFunctions) isPosInMapBounds(pos Xy) bool {
	return uint(pos.X) < f.gameSt.MapWidth && uint(pos.Y) < f.gameSt.MapHeight
}

func (f *StateFunctions) isInBounds(area Rect, bounds Rect) bool {
	if area.From.X < bounds.From.X {
		return false
	}
	if area.To.X > bounds.To.X {
		return false
	}
	if area.From.Y < bounds.From.Y {
		return false
	}
	if area.To.Y > bounds.To.Y {
		return false
	}
	return true
}

func (f *StateFunctions) isInInnerBounds(area Rect, bounds Rect) bool {
	if area.From.X < bounds.From.X {
		return false
	}
	if area.To.X >= bounds.To.X {
		return false
	}
	if area.From.Y < bounds.From.Y {
		return false
	}
	if area.To.Y >= bounds.To.Y {
		return false
	}
	return true
}

func (f *StateFunctions) isPosInBounds(pos Xy, bounds Rect) bool {
	return pos.X >= bounds.From.X && pos.Y >= bounds.From.Y && pos.X < bounds.To.X && pos.Y < bounds.To.Y
}

func (f *StateFunctions) isPosInInnerBounds(pos Xy, bounds Rect) bool {
	return pos.X >= bounds.From.X && pos.Y >= bounds.From.Y && pos.X <= bounds.To.X && pos.Y <= bounds.To.Y
}

func (f *StateFunctions) translateRect(src Rect, translation Xy) Rect {
	return Rect{From: XyAdd(src.From, translation), To: XyAdd(src.To, translation)}
}

func (f *StateFunctions) squareAt(pos Xy, halfWidth int32) Rect {
	half := Xy{X: halfWidth, Y: halfWidth}
	return Rect{From: XySub(pos, half), To: XyAdd(pos, half)}
}

func (f *StateFunctions) unitTypeBoundingBox(unitType *UnitTypeT) Rect {
	return Rect{From: XyNeg(unitType.Dimensions.From), To: XyAdd(unitType.Dimensions.To, Xy{X: 1, Y: 1})}
}

func (f *StateFunctions) unitTypeBoundingBoxAt(unitType *UnitTypeT, origin Xy) Rect {
	return Rect{From: XySub(origin, unitType.Dimensions.From), To: XyAdd(XyAdd(origin, unitType.Dimensions.To), Xy{X: 1, Y: 1})}
}

func (f *StateFunctions) unitBoundingBox(u *Unit, origin Xy) Rect {
	return f.unitTypeBoundingBoxAt(u.UnitType, origin)
}

func (f *StateFunctions) unitSpriteBoundingBox(u *Unit) Rect {
	pos := u.Sprite.Position
	return Rect{From: XySub(pos, u.UnitType.Dimensions.From), To: XyAdd(XyAdd(pos, u.UnitType.Dimensions.To), Xy{X: 1, Y: 1})}
}

func (f *StateFunctions) unitTypeInnerBoundingBox(unitType *UnitTypeT) Rect {
	return Rect{From: XyNeg(unitType.Dimensions.From), To: unitType.Dimensions.To}
}

func (f *StateFunctions) unitTypeInnerBoundingBoxAt(unitType *UnitTypeT, origin Xy) Rect {
	return Rect{From: XySub(origin, unitType.Dimensions.From), To: XyAdd(origin, unitType.Dimensions.To)}
}

func (f *StateFunctions) unitInnerBoundingBox(u *Unit, origin Xy) Rect {
	return f.unitTypeInnerBoundingBoxAt(u.UnitType, origin)
}

func (f *StateFunctions) unitSpriteInnerBoundingBox(u *Unit) Rect {
	pos := u.Sprite.Position
	return Rect{From: XySub(pos, u.UnitType.Dimensions.From), To: XyAdd(pos, u.UnitType.Dimensions.To)}
}

func (f *StateFunctions) mapBounds() Rect {
	return Rect{From: Xy{}, To: Xy{X: int32(f.gameSt.MapWidth), Y: int32(f.gameSt.MapHeight)}}
}

func (f *StateFunctions) restrictUnitPosToBounds(pos Xy, ut *UnitTypeT, bounds Rect) Xy {
	bb := f.unitTypeBoundingBoxAt(ut, pos)
	if bb.From.X < bounds.From.X {
		pos.X -= bb.From.X - bounds.From.X
	} else if bb.To.X >= bounds.To.X {
		pos.X -= bb.To.X - bounds.To.X
	}
	if bb.From.Y < bounds.From.Y {
		pos.Y -= bb.From.Y - bounds.From.Y
	} else if bb.To.Y >= bounds.To.Y {
		pos.Y -= bb.To.Y - bounds.To.Y
	}
	return pos
}

func (f *StateFunctions) restrictPosToMapBounds(pos Xy) Xy {
	if pos.X < 0 {
		pos.X = 0
	} else if pos.X >= int32(f.gameSt.MapWidth) {
		pos.X = int32(f.gameSt.MapWidth) - 1
	}
	if pos.Y < 0 {
		pos.Y = 0
	} else if pos.Y >= int32(f.gameSt.MapHeight) {
		pos.Y = int32(f.gameSt.MapHeight) - 1
	}
	return pos
}

func (f *StateFunctions) isInnerBbMoveTargetInValidBounds(bb Rect) bool {
	return f.isInInnerBounds(bb, RectAdd(f.mapBounds(), Rect{From: Xy{X: 0, Y: 0}, To: Xy{X: 0, Y: -32}}))
}

func (f *StateFunctions) restrictMoveTargetToValidBounds(ut *UnitTypeT, moveTarget Xy) Xy {
	return f.restrictUnitPosToBounds(moveTarget, ut, RectAdd(f.mapBounds(), Rect{From: Xy{X: 0, Y: 0}, To: Xy{X: 0, Y: -32}}))
}

func (f *StateFunctions) isWalkable(pos Xy) bool {
	index := f.tileIndex(pos)
	tile := &f.st.Tiles[index]
	if tile.Flags&TileFlagHasCreep != 0 {
		return true
	}
	if tile.Flags&TileFlagPartiallyWalkable != 0 {
		ux := uint(pos.X)
		uy := uint(pos.Y)
		megatileIndex := uint(f.st.TilesMegaTileIndex[index])
		flags := int32(f.gameSt.Vf4[megatileIndex].Flags[uy/8%4*4+ux/8%4])
		return flags&int32(Vf4EntryFlagWalkable) != 0
	}
	if tile.Flags&TileFlagWalkable != 0 {
		return true
	}
	return false
}

func (f *StateFunctions) tilesFlagsAnd(offsetX uint, offsetY uint, width uint, height uint, flags int32) {
	if offsetX >= f.gameSt.MapTileWidth {
		throwError("attempt to mask tile out of bounds")
	}
	if offsetY >= f.gameSt.MapTileHeight {
		throwError("attempt to mask tile out of bounds")
	}
	if width > f.gameSt.MapTileWidth || offsetX+width > f.gameSt.MapTileWidth {
		throwError("attempt to mask tile out of bounds")
	}
	if height > f.gameSt.MapTileHeight || offsetY+height > f.gameSt.MapTileHeight {
		throwError("attempt to mask tile out of bounds")
	}
	for y := offsetY; y != offsetY+height; y++ {
		for x := offsetX; x != offsetX+width; x++ {
			f.st.Tiles[x+y*f.gameSt.MapTileWidth].Flags &= uint16(flags)
		}
	}
}

func (f *StateFunctions) tilesFlagsOr(offsetX uint, offsetY uint, width uint, height uint, flags int32) {
	if offsetX >= f.gameSt.MapTileWidth {
		throwError("attempt to mask tile out of bounds")
	}
	if offsetY >= f.gameSt.MapTileHeight {
		throwError("attempt to mask tile out of bounds")
	}
	if width > f.gameSt.MapTileWidth || offsetX+width > f.gameSt.MapTileWidth {
		throwError("attempt to mask tile out of bounds")
	}
	if height > f.gameSt.MapTileHeight || offsetY+height > f.gameSt.MapTileHeight {
		throwError("attempt to mask tile out of bounds")
	}
	for y := offsetY; y != offsetY+height; y++ {
		for x := offsetX; x != offsetX+width; x++ {
			f.st.Tiles[x+y*f.gameSt.MapTileWidth].Flags |= uint16(flags)
		}
	}
}

func (f *StateFunctions) unitTypeSpreadsCreep(ut *UnitTypeT, unitIsCompleted bool) bool {
	if f.unitIs(ut, UnitTypesZergHatchery) && unitIsCompleted {
		return true
	}
	if f.unitIs(ut, UnitTypesZergLair) {
		return true
	}
	if f.unitIs(ut, UnitTypesZergHive) {
		return true
	}
	if f.unitIs(ut, UnitTypesZergCreepColony) && unitIsCompleted {
		return true
	}
	if f.unitIs(ut, UnitTypesZergSporeColony) {
		return true
	}
	if f.unitIs(ut, UnitTypesZergSunkenColony) {
		return true
	}
	return false
}

func (f *StateFunctions) visibleHpPlusShields(u *Unit) int32 {
	var r int32
	if u.UnitType.HasShield {
		r += u.ShieldPoints.IntegerPart()
	}
	r += u.Hp.Ceil().IntegerPart()
	return r
}

func (f *StateFunctions) maxVisibleHp(u *Unit) int32 {
	hp := u.UnitType.Hitpoints.IntegerPart()
	if hp == 0 {
		hp = u.Hp.Ceil().IntegerPart()
	}
	if hp == 0 {
		hp = 1
	}
	return hp
}

func (f *StateFunctions) maxVisibleHpPlusShields(u *Unit) int32 {
	var shields int32
	if u.UnitType.HasShield {
		shields += u.UnitType.ShieldPoints
	}
	return f.maxVisibleHp(u) + shields
}

func (f *StateFunctions) loadedUnits(u *Unit) iter.Seq[*Unit] {
	return func(yield func(*Unit) bool) {
		for i := range u.LoadedUnits {
			nu := f.getUnitByID(u.LoadedUnits[i])
			if nu == nil {
				continue
			}
			if !yield(nu) {
				return
			}
		}
	}
}

func (f *StateFunctions) unitSpaceOccupied(u *Unit) uint {
	var r uint
	for nu := range f.loadedUnits(u) {
		r += nu.UnitType.SpaceRequired
	}
	return r
}

func (f *StateFunctions) getUnitStrength(u *Unit, ground bool) int32 {
	if f.unitIs(u.UnitType, UnitTypesZergLarva) {
		return 0
	}
	if f.unitIsEgg(u.UnitType) {
		return 0
	}
	visHpShields := f.visibleHpPlusShields(u)
	maxVisHpShields := f.maxVisibleHpPlusShields(u)
	if f.uHallucination(u) {
		if visHpShields < maxVisHpShields {
			return 0
		}
	}

	r := f.gameSt.UnitAirStrength[u.UnitType.ID]
	if ground {
		r = f.gameSt.UnitGroundStrength[u.UnitType.ID]
	}
	if f.unitIs(u.UnitType, UnitTypesTerranBunker) {
		r = f.gameSt.UnitAirStrength[UnitTypesTerranMarine]
		if ground {
			r = f.gameSt.UnitGroundStrength[UnitTypesTerranMarine]
		}
		r *= int32(f.unitSpaceOccupied(u))
	}
	if f.utHasEnergy(u.UnitType) && !f.uHallucination(u) {
		r += u.Energy.IntegerPart() / 2
	}
	return r * visHpShields / maxVisHpShields
}

func (f *StateFunctions) updateUnitDamageOverlay(u *Unit) {
	if u.Sprite.MainImage.ImageType.DamageFilenameIndex == 0 {
		return
	}
	states := f.damageOverlayStates(u)
	maxHp := u.UnitType.Hitpoints
	twoThirdsMaxHp := maxHp.Sub(maxHp.DivInt(3))
	hpPerState := twoThirdsMaxHp.DivInt(states + 1)
	if hpPerState == Fp8FromRaw(0) {
		hpPerState = Fp8FromRaw(1)
	}
	damageState := (u.Hp.RawValue - twoThirdsMaxHp.RawValue%hpPerState.RawValue - maxHp.RawValue/3 - 1) / hpPerState.RawValue
	if damageState < 0 {
		damageState = 0
	}
	for ; u.DamageOverlayState < damageState; u.DamageOverlayState++ {
		largeFlames := f.findImage(u.Sprite, ImageTypesIMAGEIDFlames1Type1Large, ImageTypesIMAGEIDFlames8Type3Large)
		if largeFlames != nil {
			index := uint(largeFlames.ImageType.ID) - uint(ImageTypesIMAGEIDFlames1Type1Large)
			f.destroyImage(largeFlames)
			offset := f.getImageLoOffset(u.Sprite.MainImage, 1, index, false)
			imageType := f.getImageType(ImageTypes(uint(ImageTypesIMAGEIDFlames1Type1Small) + index))
			f.createImage(imageType, u.Sprite, offset, imageOrderTop, nil)
		} else {
			f.destroyImageFromTo(u.Sprite, ImageTypesIMAGEIDFlames1Type1Small, ImageTypesIMAGEIDFlames8Type3Small)
		}
	}
	for ; u.DamageOverlayState > damageState; u.DamageOverlayState-- {
		smallFlames := f.findImage(u.Sprite, ImageTypesIMAGEIDFlames1Type1Small, ImageTypesIMAGEIDFlames8Type3Small)
		if smallFlames != nil {
			index := uint(smallFlames.ImageType.ID) - uint(ImageTypesIMAGEIDFlames1Type1Small)
			f.destroyImage(smallFlames)
			offset := f.getImageLoOffset(u.Sprite.MainImage, 1, index, false)
			imageType := f.getImageType(ImageTypes(uint(ImageTypesIMAGEIDFlames1Type1Large) + index))
			f.createImage(imageType, u.Sprite, offset, imageOrderTop, nil)
			continue
		}
		index := uint(f.lcgRand(8)%(states/2) + 1)
		var n uint
		mainImage := u.Sprite.MainImage
		offsets := (*f.globalSt.ImageLoOffsets[mainImage.ImageType.ID][1])[mainImage.FrameIndex]
		for i := uint(0); i != uint(len(offsets)); {
			offset := offsets[i]
			if offset != (Xy{X: 127, Y: 127}) {
				n++
				imageID := ImageTypes(uint(ImageTypesIMAGEIDFlames1Type1Large) + i)
				image := f.findImage(u.Sprite, imageID, imageID)
				if image == nil && n >= index {
					newImageID := ImageTypes(uint(ImageTypesIMAGEIDFlames1Type1Small) + i)
					if f.iFlag(mainImage, ImageFlagHorizontallyFlipped) {
						offset.X = -offset.X
					}
					f.createImage(f.getImageType(newImageID), u.Sprite, offset, imageOrderTop, nil)
					break
				}
			}
			i++
			if i == uint(len(offsets)) {
				i = 0
				if n == 0 {
					throwError("missing damage overlay offset")
				}
			}
		}
	}
}

func (f *StateFunctions) setUnitHp(u *Unit, hitpoints Fp8) {
	u.Hp = hitpoints
	if u.UnitType.Hitpoints.Less(hitpoints) {
		u.Hp = u.UnitType.Hitpoints
	}
	if f.uCompleted(u) {
		f.updateUnitDamageOverlay(u)

		u.AirStrength = f.getUnitStrength(u, false)
		u.GroundStrength = f.getUnitStrength(u, true)
	}
}

func (f *StateFunctions) setUnitShieldPoints(u *Unit, shieldPoints Fp8) {
	maxShieldPoints := Fp8Integer(u.UnitType.ShieldPoints)
	u.ShieldPoints = shieldPoints
	if maxShieldPoints.Less(shieldPoints) {
		u.ShieldPoints = maxShieldPoints
	}
}

func (f *StateFunctions) setUnitEnergy(u *Unit, energy Fp8) {
	maxEnergy := f.unitMaxEnergy(u)
	u.Energy = energy
	if maxEnergy.Less(energy) {
		u.Energy = maxEnergy
	}
}

func (f *StateFunctions) unitIsMineralField(ut *UnitTypeT) bool {
	if ut.ID == UnitTypesResourceMineralField {
		return true
	}
	if ut.ID == UnitTypesResourceMineralFieldType2 {
		return true
	}
	if ut.ID == UnitTypesResourceMineralFieldType3 {
		return true
	}
	return false
}

func (f *StateFunctions) setUnitResources(u *Unit, resources int32) {
	if !f.utResource(u.UnitType) {
		return
	}
	u.Building.Resource.ResourceCount = resources
	if f.unitIsMineralField(u.UnitType) {
		anim := int32(IscriptAnimsWorkingToIdle)
		if resources < 250 {
			anim = IscriptAnimsSpecialState1
		} else if resources < 500 {
			anim = IscriptAnimsSpecialState2
		} else if resources < 750 {
			anim = IscriptAnimsAlmostBuilt
		}
		if u.Building.Resource.ResourceIscript != anim {
			u.Building.Resource.ResourceIscript = anim
			f.spriteRunAnim(u.Sprite, anim)
		}
	}
}

func (f *StateFunctions) unitIsDisabled(u *Unit) bool {
	if f.uDisabled(u) {
		return true
	}
	if u.LockdownTimer != 0 {
		return true
	}
	if u.StasisTimer != 0 {
		return true
	}
	if u.MaelstromTimer != 0 {
		return true
	}
	return false
}

func (f *StateFunctions) findImage(sprite *Sprite, firstID ImageTypes, lastID ImageTypes) *Image {
	for i := range sprite.Images.All() {
		if i.ImageType.ID >= firstID && i.ImageType.ID <= lastID {
			return i
		}
	}
	return nil
}

func (f *StateFunctions) findImageUnit(u *Unit, firstID ImageTypes, lastID ImageTypes) *Image {
	r := f.findImage(u.Sprite, firstID, lastID)
	if r == nil && u.Subunit != nil {
		r = f.findImage(u.Subunit.Sprite, firstID, lastID)
	}
	return r
}

func (f *StateFunctions) destroyImageFromTo(sprite *Sprite, firstID ImageTypes, lastID ImageTypes) {
	image := f.findImage(sprite, firstID, lastID)
	if image != nil {
		f.destroyImage(image)
	}
}

func (f *StateFunctions) destroyImageFromToUnit(u *Unit, firstID ImageTypes, lastID ImageTypes) {
	f.destroyImageFromTo(u.Sprite, firstID, lastID)
	if u.Subunit != nil {
		f.destroyImageFromTo(u.Subunit.Sprite, firstID, lastID)
	}
}

func (f *StateFunctions) disableEffectEnd(u *Unit, first ImageTypes, last ImageTypes) {
	stillDisabled := f.unitIsDisabled(u)
	if u.Subunit != nil && !stillDisabled {
		f.uUnsetStatusFlag(u.Subunit, UnitStatusFlagDisabled)
		f.setUnitOrder(u.Subunit, u.Subunit.UnitType.ReturnToIdle.Get())
	}
	image := f.findImage(u.Sprite, first, last)
	if image == nil && u.Subunit != nil {
		image = f.findImage(u.Subunit.Sprite, first, last)
	}
	if image != nil {
		f.iscriptRunAnim(image, IscriptAnimsDeath)
	}
	if f.utWorker(u.UnitType) && !stillDisabled {
		target := u.Worker.GatherTarget
		if target != nil && f.utResource(target.UnitType) {
			if u.Worker.IsGathering && target.Building.Resource.IsBeingGathered {
				if u.OrderType.ID == OrdersWaitForMinerals || u.OrderType.ID == OrdersWaitForGas {
					if u.OrderState == 2 {
						target.Building.Resource.GatherQueue.Remove(u)
						f.removeOneOrder(u, f.getOrderType(OrdersGatherWaitInterrupted))
						f.tryGatherResource(u, target)
					}
				}
			}
		}
	}
	u.OrderProcessTimer = 15
}

func (f *StateFunctions) removeStasis(u *Unit) {
	u.StasisTimer = 0
	f.uSetStatusFlagValue(u, UnitStatusFlagInvincible, f.utInvincible(u.UnitType))
	f.disableEffectEnd(u, ImageTypesIMAGEIDStasisFieldSmall, ImageTypesIMAGEIDStasisFieldLarge)
}

func (f *StateFunctions) dealIrradiateDamage(sourceUnit *Unit) {
	damage := func(target *Unit) {
		if !f.utOrganic(target.UnitType) {
			return
		}
		if f.utBuilding(target.UnitType) {
			return
		}
		if f.unitIs(target.UnitType, UnitTypesZergLarva) {
			return
		}
		if f.unitIs(target.UnitType, UnitTypesZergEgg) {
			return
		}
		if f.unitIs(target.UnitType, UnitTypesZergLurkerEgg) {
			return
		}
		if f.uBurrowed(target) && target != sourceUnit {
			return
		}
		if !f.uLoaded(target) {
			if !f.unitTargetInRange(sourceUnit, target, 32) {
				return
			}
		}
		w := f.getWeaponType(WeaponTypesIrradiate)
		f.weaponDealDamage(w, Fp8Integer(w.DamageAmount).DivInt(w.Cooldown), 1, target, DirectionTFromRaw(0), sourceUnit.IrradiatedBy, sourceUnit.IrradiateOwner)
	}
	if f.uBurrowed(sourceUnit) {
		damage(sourceUnit)
	} else if f.uLoaded(sourceUnit) {
		if sourceUnit.ConnectedUnit != nil {
			damage(sourceUnit.ConnectedUnit)
		}
	} else {
		for n := range f.findUnitsNoexpand(f.squareAt(sourceUnit.Sprite.Position, 160)) {
			damage(n)
		}
	}
}

func (f *StateFunctions) removeIrradiate(u *Unit) {
	u.IrradiateTimer = 0
	u.IrradiatedBy = nil
	u.IrradiateOwner = 8
	f.destroyImageFromToUnit(u, ImageTypesIMAGEIDIrradiateSmall, ImageTypesIMAGEIDIrradiateLarge)
}

func (f *StateFunctions) removeEnsnare(u *Unit) {
	u.EnsnareTimer = 0
	f.destroyImageFromToUnit(u, ImageTypesIMAGEIDEnsnareOverlaySmall, ImageTypesIMAGEIDEnsnareOverlayLarge)
	f.updateUnitSpeed(u)
}

func (f *StateFunctions) removeLockdown(u *Unit) {
	u.LockdownTimer = 0
	f.disableEffectEnd(u, ImageTypesIMAGEIDLockdownFieldSmall, ImageTypesIMAGEIDLockdownFieldLarge)
}

func (f *StateFunctions) removePlague(u *Unit) {
	u.PlagueTimer = 0
	f.destroyImageFromToUnit(u, ImageTypesIMAGEIDPlagueOverlaySmall, ImageTypesIMAGEIDPlagueOverlayLarge)
}

func (f *StateFunctions) removeMaelstrom(u *Unit) {
	u.MaelstromTimer = 0
	f.disableEffectEnd(u, ImageTypesIMAGEIDMaelstormOverlaySmall, ImageTypesIMAGEIDMaelstormOverlayLarge)
}

func (f *StateFunctions) removeAcidSpores(u *Unit) {
	u.AcidSporeCount = 0
	u.AcidSporeTime = [9]int32{}
	f.destroyImageFromToUnit(u, ImageTypesIMAGEIDAcidSpores1OverlaySmall, ImageTypesIMAGEIDAcidSpores6_9OverlayLarge)
}

func (f *StateFunctions) updateUnitStatusTimers(u *Unit) {
	if u.StasisTimer != 0 {
		u.StasisTimer--
		if u.StasisTimer == 0 {
			f.removeStasis(u)
		}
	}
	if u.StimTimer != 0 {
		u.StimTimer--
		if u.StimTimer == 0 {
			f.updateUnitSpeed(u)
		}
	}
	if u.EnsnareTimer != 0 {
		u.EnsnareTimer--
		if u.EnsnareTimer == 0 {
			f.removeEnsnare(u)
		}
	}
	if u.DefensiveMatrixTimer != 0 {
		u.DefensiveMatrixTimer--
		if u.DefensiveMatrixTimer == 0 {
			f.dealDefensiveMatrixDamage(u, u.DefensiveMatrixHp)
		}
	}
	if u.IrradiateTimer != 0 {
		u.IrradiateTimer--
		f.dealIrradiateDamage(u)
		if u.IrradiateTimer == 0 {
			f.removeIrradiate(u)
		}
	}
	if u.LockdownTimer != 0 {
		u.LockdownTimer--
		if u.LockdownTimer == 0 {
			f.removeLockdown(u)
		}
	}
	if u.MaelstromTimer != 0 {
		u.MaelstromTimer--
		if u.MaelstromTimer == 0 {
			f.removeMaelstrom(u)
		}
	}
	if u.PlagueTimer != 0 {
		u.PlagueTimer--
		if !f.uInvincible(u) {
			damage := Fp8Integer(f.getWeaponType(WeaponTypesPlague).DamageAmount).DivInt(76)
			if u.Hp.Greater(damage) {
				f.unitDealDamage(u, damage, nil, ^int32(0), true)
			}
		}
		if u.PlagueTimer == 0 {
			f.removePlague(u)
		}
	}
	if u.StormTimer != 0 {
		u.StormTimer--
	}
	prevAcidSporeCount := u.AcidSporeCount
	if prevAcidSporeCount != 0 {
		for i := range u.AcidSporeTime {
			v := &u.AcidSporeTime[i]
			if *v == 0 {
				continue
			}
			*v--
			if *v == 0 {
				u.AcidSporeCount--
			}
		}
		if u.AcidSporeCount != 0 {
			f.updateAcidSporeImage(u)
		} else {
			f.removeAcidSpores(u)
		}
	}
}

func (f *StateFunctions) unitCloakEnergyCost(u *Unit) Fp8 {
	switch u.UnitType.ID {
	case UnitTypesTerranGhost, UnitTypesHeroSarahKerrigan, UnitTypesHeroAlexeiStukov, UnitTypesHeroSamirDuran, UnitTypesHeroInfestedDuran, UnitTypesHeroInfestedKerrigan:
		return Fp8FromRaw(10)
	case UnitTypesTerranWraith, UnitTypesHeroTomKazansky:
		return Fp8FromRaw(13)
	default:
		return Fp8FromRaw(0)
	}
}

func (f *StateFunctions) setSecondaryOrder(u *Unit, orderType *OrderTypeT) {
	if u.SecondaryOrderType == orderType {
		return
	}
	u.SecondaryOrderType = orderType
	u.SecondaryOrderState = 0
	u.CurrentBuildUnit = nil
}

func (f *StateFunctions) updateUnitEnergy(u *Unit) {
	if !f.utHasEnergy(u.UnitType) {
		return
	}
	if f.uHallucination(u) {
		return
	}
	if !f.uCompleted(u) {
		return
	}
	if (f.uCloaked(u) || f.uRequiresDetector(u)) && !f.uPassivelyCloaked(u) {
		cost := f.unitCloakEnergyCost(u)
		if u.Energy.Less(cost) {
			if u.SecondaryOrderType.ID == OrdersCloak {
				f.setSecondaryOrder(u, f.getOrderType(OrdersNothing))
			}
		} else {
			u.Energy = u.Energy.Sub(cost)
		}
	} else {
		maxEnergy := f.unitMaxEnergy(u)
		if f.unitIs(u.UnitType, UnitTypesProtossDarkArchon) && u.OrderType.ID == OrdersCompletingArchonSummon && u.OrderState == 0 {
			maxEnergy = Fp8Integer(50)
		}
		energy := u.Energy.Add(Fp8FromRaw(8))
		u.Energy = energy
		if maxEnergy.Less(energy) {
			u.Energy = maxEnergy
		}
	}
}

func (f *StateFunctions) unitHpPercent(u *Unit) int32 {
	maxHp := f.maxVisibleHp(u)
	hp := u.Hp.Ceil().IntegerPart()
	return hp * 100 / maxHp
}

func (f *StateFunctions) updateUnitValues(u *Unit) {
	if u.MainOrderTimer != 0 {
		u.MainOrderTimer--
	}
	if u.GroundWeaponCooldown != 0 {
		u.GroundWeaponCooldown--
	}
	if u.AirWeaponCooldown != 0 {
		u.AirWeaponCooldown--
	}
	if u.SpellCooldown != 0 {
		u.SpellCooldown--
	}
	if u.UnitType.HasShield {
		maxShields := Fp8Integer(u.UnitType.ShieldPoints)
		if u.ShieldPoints != maxShields {
			u.ShieldPoints = u.ShieldPoints.Add(Fp8FromRaw(7))
			if u.ShieldPoints.Greater(maxShields) {
				u.ShieldPoints = maxShields
			}
		}
	}
	if f.unitIs(u.UnitType, UnitTypesZergZergling) || f.unitIs(u.UnitType, UnitTypesHeroDevouringOne) {
		if u.GroundWeaponCooldown == 0 {
			u.OrderProcessTimer = 0
		}
	}
	if u.IsBeingHealed {
		u.IsBeingHealed = false
	}
	if f.uCompleted(u) || !f.usHidden(&u.Thingy) {
		u.CycleCounter++
		if u.CycleCounter >= 8 {
			u.CycleCounter = 0
			f.updateUnitStatusTimers(u)
		}
	}
	if f.uCompleted(u) {
		if f.utRegensHp(u.UnitType) {
			if u.Hp.Greater(Fp8FromRaw(0)) && u.Hp != u.UnitType.Hitpoints {
				f.setUnitHp(u, u.Hp.Add(Fp8FromRaw(4)))
			}
		}
		f.updateUnitEnergy(u)
		if u.MoveTargetTimer != 0 {
			u.MoveTargetTimer--
		}
		removeTimerExpired := false
		if u.RemoveTimer != 0 {
			u.RemoveTimer--
			removeTimerExpired = u.RemoveTimer == 0
		}
		if removeTimerExpired {
			f.killUnit(u)
		} else {
			if f.unitRace(u.UnitType) == RaceTerran {
				if f.uGroundedBuilding(u) || f.utFlyingBuilding(u.UnitType) {
					if f.unitHpPercent(u) <= 33 {
						f.unitDealDamage(u, Fp8FromRaw(20), nil, u.LastAttackingPlayer, true)
					}
				}
			}
		}
	}
}

func (f *StateFunctions) unitTurret(u *Unit) *Unit {
	if u.Subunit == nil {
		return nil
	}
	if !f.utTurret(u.Subunit.UnitType) {
		return nil
	}
	return u.Subunit
}

func (f *StateFunctions) unitMainUnit(u *Unit) *Unit {
	if f.utTurret(u.UnitType) {
		return u.Subunit
	}
	return u
}

func (f *StateFunctions) unitAttackingUnit(u *Unit) *Unit {
	if u.Subunit != nil && f.utTurret(u.Subunit.UnitType) {
		return u.Subunit
	}
	return u
}

func (f *StateFunctions) unitGroundWeapon(u *Unit) *WeaponTypeT {
	if f.unitIs(u.UnitType, UnitTypesZergLurker) && !f.uBurrowed(u) {
		return nil
	}
	return u.UnitType.GroundWeapon.Get()
}

func (f *StateFunctions) unitAirWeapon(u *Unit) *WeaponTypeT {
	return u.UnitType.AirWeapon.Get()
}

func (f *StateFunctions) unitOrSubunitGroundWeapon(u *Unit) *WeaponTypeT {
	w := f.unitGroundWeapon(u)
	if w != nil || u.Subunit == nil {
		return w
	}
	return f.unitGroundWeapon(u.Subunit)
}

func (f *StateFunctions) unitOrSubunitAirWeapon(u *Unit) *WeaponTypeT {
	w := f.unitAirWeapon(u)
	if w != nil || u.Subunit == nil {
		return w
	}
	return f.unitAirWeapon(u.Subunit)
}

func (f *StateFunctions) unitTargetWeapon(u *Unit, target *Unit) *WeaponTypeT {
	if f.uFlying(target) {
		return f.unitAirWeapon(f.unitAttackingUnit(u))
	}
	return f.unitGroundWeapon(f.unitAttackingUnit(u))
}

func (f *StateFunctions) unitIsCarrier(ut *UnitTypeT) bool {
	return f.unitIs(ut, UnitTypesProtossCarrier) || f.unitIs(ut, UnitTypesHeroGantrithor)
}

func (f *StateFunctions) unitIsReaver(ut *UnitTypeT) bool {
	return f.unitIs(ut, UnitTypesProtossReaver) || f.unitIs(ut, UnitTypesHeroWarbringer)
}

func (f *StateFunctions) unitIsQueen(ut *UnitTypeT) bool {
	return f.unitIs(ut, UnitTypesZergQueen) || f.unitIs(ut, UnitTypesHeroMatriarch)
}

func (f *StateFunctions) unitIsHatchery(ut *UnitTypeT) bool {
	if f.unitIs(ut, UnitTypesZergHatchery) {
		return true
	}
	if f.unitIs(ut, UnitTypesZergLair) {
		return true
	}
	if f.unitIs(ut, UnitTypesZergHive) {
		return true
	}
	return false
}

func (f *StateFunctions) unitIsMarine(ut *UnitTypeT) bool {
	if f.unitIs(ut, UnitTypesTerranMarine) {
		return true
	}
	if f.unitIs(ut, UnitTypesHeroJimRaynorMarine) {
		return true
	}
	return false
}

func (f *StateFunctions) unitIsFirebat(ut *UnitTypeT) bool {
	if f.unitIs(ut, UnitTypesTerranFirebat) {
		return true
	}
	if f.unitIs(ut, UnitTypesHeroGuiMontag) {
		return true
	}
	return false
}

func (f *StateFunctions) unitIsGhost(ut *UnitTypeT) bool {
	if f.unitIs(ut, UnitTypesTerranGhost) {
		return true
	}
	if f.unitIs(ut, UnitTypesHeroSarahKerrigan) {
		return true
	}
	if f.unitIs(ut, UnitTypesHeroAlexeiStukov) {
		return true
	}
	if f.unitIs(ut, UnitTypesHeroSamirDuran) {
		return true
	}
	if f.unitIs(ut, UnitTypesHeroInfestedDuran) {
		return true
	}
	return false
}

func (f *StateFunctions) unitIsWraith(ut *UnitTypeT) bool {
	if f.unitIs(ut, UnitTypesTerranWraith) {
		return true
	}
	if f.unitIs(ut, UnitTypesHeroTomKazansky) {
		return true
	}
	return false
}

func (f *StateFunctions) unitIsMapRevealer(ut *UnitTypeT) bool {
	if f.unitIs(ut, UnitTypesSpellScannerSweep) {
		return true
	}
	if f.unitIs(ut, UnitTypesSpecialMapRevealer) {
		return true
	}
	return false
}

func (f *StateFunctions) unitIsRefinery(ut *UnitTypeT) bool {
	if f.unitIs(ut, UnitTypesTerranRefinery) {
		return true
	}
	if f.unitIs(ut, UnitTypesProtossAssimilator) {
		return true
	}
	if f.unitIs(ut, UnitTypesZergExtractor) {
		return true
	}
	return false
}

func (f *StateFunctions) unitIsNonFlagBeacon(ut *UnitTypeT) bool {
	return ut.ID >= UnitTypesSpecialZergBeacon && ut.ID <= UnitTypesSpecialProtossBeacon
}

func (f *StateFunctions) unitIsSpecialBeacon(ut *UnitTypeT) bool {
	return ut.ID >= UnitTypesSpecialZergBeacon && ut.ID <= UnitTypesSpecialProtossFlagBeacon
}

func (f *StateFunctions) unitIsScout(ut *UnitTypeT) bool {
	return f.unitIs(ut, UnitTypesProtossScout) || f.unitIs(ut, UnitTypesHeroMojo) || f.unitIs(ut, UnitTypesHeroArtanis)
}

func (f *StateFunctions) unitIsVulture(ut *UnitTypeT) bool {
	return f.unitIs(ut, UnitTypesTerranVulture) || f.unitIs(ut, UnitTypesHeroJimRaynorVulture)
}

func (f *StateFunctions) unitIsNydus(ut *UnitTypeT) bool {
	return f.unitIs(ut, UnitTypesZergNydusCanal)
}

func (f *StateFunctions) unitIsDefiler(ut *UnitTypeT) bool {
	if f.unitIs(ut, UnitTypesZergDefiler) {
		return true
	}
	if f.unitIs(ut, UnitTypesHeroUncleanOne) {
		return true
	}
	return false
}

func (f *StateFunctions) unitIsUltralisk(ut *UnitTypeT) bool {
	if f.unitIs(ut, UnitTypesZergUltralisk) {
		return true
	}
	if f.unitIs(ut, UnitTypesHeroTorrasque) {
		return true
	}
	return false
}

func (f *StateFunctions) unitIsArbiter(ut *UnitTypeT) bool {
	if f.unitIs(ut, UnitTypesProtossArbiter) {
		return true
	}
	if f.unitIs(ut, UnitTypesHeroDanimoth) {
		return true
	}
	return false
}

func (f *StateFunctions) unitIsEgg(ut *UnitTypeT) bool {
	if f.unitIs(ut, UnitTypesZergEgg) {
		return true
	}
	if f.unitIs(ut, UnitTypesZergCocoon) {
		return true
	}
	if f.unitIs(ut, UnitTypesZergLurkerEgg) {
		return true
	}
	return false
}

func (f *StateFunctions) unitIsGoliath(ut *UnitTypeT) bool {
	if f.unitIs(ut, UnitTypesTerranGoliath) {
		return true
	}
	if f.unitIs(ut, UnitTypesHeroAlanSchezar) {
		return true
	}
	return false
}

func (f *StateFunctions) unitIsCritter(ut *UnitTypeT) bool {
	if f.unitIs(ut, UnitTypesCritterRhynadon) {
		return true
	}
	if f.unitIs(ut, UnitTypesCritterBengalaas) {
		return true
	}
	if f.unitIs(ut, UnitTypesCritterRagnasaur) {
		return true
	}
	if f.unitIs(ut, UnitTypesCritterScantid) {
		return true
	}
	if f.unitIs(ut, UnitTypesCritterKakaru) {
		return true
	}
	if f.unitIs(ut, UnitTypesCritterUrsadon) {
		return true
	}
	return false
}

func (f *StateFunctions) unitIsSiegedTank(ut *UnitTypeT) bool {
	if f.unitIs(ut, UnitTypesTerranSiegeTankSiegeMode) {
		return true
	}
	if f.unitIs(ut, UnitTypesHeroEdmundDukeSiegeMode) {
		return true
	}
	return false
}

func (f *StateFunctions) unitIsUnsiegedTank(ut *UnitTypeT) bool {
	if f.unitIs(ut, UnitTypesTerranSiegeTankTankMode) {
		return true
	}
	if f.unitIs(ut, UnitTypesHeroEdmundDukeTankMode) {
		return true
	}
	return false
}

func (f *StateFunctions) unitIsTank(ut *UnitTypeT) bool {
	return f.unitIsSiegedTank(ut) || f.unitIsUnsiegedTank(ut)
}

func (f *StateFunctions) unitIsTrap(ut *UnitTypeT) bool {
	return ut.ID >= UnitTypesSpecialFloorMissileTrap && ut.ID <= UnitTypesSpecialRightWallFlameTrap
}

func (f *StateFunctions) unitIsZergBuilding(ut *UnitTypeT) bool {
	return ut.ID >= UnitTypesZergInfestedCommandCenter && ut.ID <= UnitTypesSpecialCerebrateDaggoth
}

func (f *StateFunctions) unitIsFighter(ut *UnitTypeT) bool {
	if f.unitIs(ut, UnitTypesProtossInterceptor) {
		return true
	}
	if f.unitIs(ut, UnitTypesProtossScarab) {
		return true
	}
	return false
}

func (f *StateFunctions) unitIsUndetected(u *Unit, owner int32) bool {
	if !f.uCloaked(u) && !f.uRequiresDetector(u) {
		return false
	}
	if u.DetectedFlags&(uint32(1)<<owner) != 0 {
		return false
	}
	return true
}

func (f *StateFunctions) unitTargetIsUndetected(u *Unit, target *Unit) bool {
	return f.unitIsUndetected(target, u.Owner)
}

func (f *StateFunctions) unitTargetIsVisible(u *Unit, target *Unit) bool {
	return target.Sprite.VisibilityFlags&(int32(1)<<u.Owner) != 0
}

func (f *StateFunctions) unitPositionIsVisible(u *Unit, position Xy) bool {
	return int32(f.st.Tiles[f.tileIndex(position)].Visible)&(int32(1)<<u.Owner) == 0
}

func (f *StateFunctions) unitPositionIsExplored(u *Unit, position Xy) bool {
	return int32(f.st.Tiles[f.tileIndex(position)].Explored)&(int32(1)<<u.Owner) == 0
}

func (f *StateFunctions) unitCanSeeTarget(u *Unit, target *Unit) bool {
	if f.unitTargetIsUndetected(u, target) {
		return false
	}
	return f.unitTargetIsVisible(u, target)
}

func (f *StateFunctions) unitOrderCanTargetSelf(u *Unit, order *OrderTypeT) bool {
	if f.uGroundedBuilding(u) {
		if order.ID == OrdersRallyPointUnit {
			return true
		}
		if order.ID == OrdersRallyPointTile {
			return true
		}
	}
	if f.unitIsDefiler(u.UnitType) {
		if order.ID == OrdersCastDarkSwarm {
			return true
		}
	}
	if f.unitProvidesSpace(u) {
		if order.ID == OrdersMoveUnload {
			return true
		}
	}
	return false
}

func (f *StateFunctions) playerPositionIsVisible(owner int32, position Xy) bool {
	return int32(f.tileVisibility(position))&(int32(1)<<owner) != 0
}

func (f *StateFunctions) playerPositionIsExplored(owner int32, position Xy) bool {
	return int32(f.tileExplored(position))&(int32(1)<<owner) != 0
}

func (f *StateFunctions) getRegionAt(pos Xy) *Region {
	index := f.gameSt.Regions.TileRegionIndex[uint(pos.Y)/32*256+uint(pos.X)/32]
	if index >= 0x2000 {
		maskIndex := (uint(pos.Y)/8&3)*4 + (uint(pos.X) / 8 & 3)
		split := &f.gameSt.Regions.SplitRegions[index-0x2000]
		if uint32(split.Mask)&(uint32(1)<<maskIndex) != 0 {
			return split.B
		}
		return split.A
	}
	return &f.gameSt.Regions.Regions[index]
}

func (f *StateFunctions) getRegionAtPreferWalkable(pos Xy) *Region {
	index := f.gameSt.Regions.TileRegionIndex[uint(pos.Y)/32*256+uint(pos.X)/32]
	if index >= 0x2000 {
		split := &f.gameSt.Regions.SplitRegions[index-0x2000]
		return split.A
	}
	return &f.gameSt.Regions.Regions[index]
}

func (f *StateFunctions) isReachable(from Xy, to Xy) bool {
	return f.getRegionAt(from).GroupIndex == f.getRegionAt(to).GroupIndex
}

func (f *StateFunctions) isReachableUnitToPos(u *Unit, to Xy) bool {
	if ^u.PathingFlags&1 != 0 {
		return true
	}
	return f.isReachable(u.Sprite.Position, to)
}

func (f *StateFunctions) isReachableUnitToUnit(fromU *Unit, toU *Unit) bool {
	return f.isReachable(fromU.Sprite.Position, toU.Sprite.Position)
}

func (f *StateFunctions) unitCanBeInfested(u *Unit) bool {
	if !f.unitIs(u.UnitType, UnitTypesTerranCommandCenter) {
		return false
	}
	if !f.uCompleted(u) {
		return false
	}
	return f.unitHpPercent(u) < 50
}

func (f *StateFunctions) unitCanAttackTarget(u *Unit, target *Unit) bool {
	if target == nil {
		return false
	}
	if f.unitIsDisabled(u) {
		return false
	}
	if f.uInvincible(target) {
		return false
	}
	if f.usHidden(&target.Thingy) {
		return false
	}
	if f.unitTargetIsUndetected(u, target) {
		return false
	}
	if f.unitIsCarrier(u.UnitType) {
		return true
	}
	if f.unitIsReaver(u.UnitType) {
		if f.uFlying(target) {
			return false
		}
		return f.isReachable(u.Sprite.Position, target.Sprite.Position)
	}
	if f.unitIsQueen(u.UnitType) {
		return f.unitCanBeInfested(target)
	}
	return f.unitTargetWeapon(u, target) != nil
}

func (f *StateFunctions) unitAutoattack(u *Unit) bool {
	if u.AutoTargetUnit == nil {
		return false
	}
	if f.unitTargetIsEnemy(u, u.AutoTargetUnit) && f.unitCanAttackTarget(u, u.AutoTargetUnit) {
		f.autoAttackTarget(u, u.AutoTargetUnit)
		f.uSetStatusFlag(u, UnitStatusFlag8)
		return true
	}
	u.AutoTargetUnit = nil
	return false
}

func (f *StateFunctions) unitCanCollideWith(u *Unit, target *Unit) bool {
	if target == u {
		return false
	}
	if ^target.PathingFlags&1 != 0 {
		return false
	}
	if f.uNoCollide(target) {
		return false
	}
	if f.uStatusFlag(target, UnitStatusFlag400000) {
		if f.unitTargetIsEnemy(u, target) {
			return false
		}
	}
	if f.uGathering(u) && !f.uGroundedBuilding(target) {
		if !f.uGathering(target) {
			return false
		}
		if u.OrderType.ID == OrdersReturnGas {
			return false
		}
		if target.OrderType.ID != OrdersWaitForGas {
			return false
		}
	}
	if f.unitIs(u.UnitType, UnitTypesZergLarva) {
		if !f.uGroundedBuilding(target) {
			return false
		}
	} else if f.unitIs(target.UnitType, UnitTypesZergLarva) {
		return false
	}
	return true
}

func (f *StateFunctions) rectDifference(a Rect, b Rect) Xy {
	var x, y int32
	if a.From.X > b.To.X {
		x = a.From.X - b.To.X
	} else if b.From.X > a.To.X {
		x = b.From.X - a.To.X
	} else {
		x = 0
	}
	if a.From.Y > b.To.Y {
		y = a.From.Y - b.To.Y
	} else if b.From.Y > a.To.Y {
		y = b.From.Y - a.To.Y
	} else {
		y = 0
	}
	return Xy{X: x, Y: y}
}

func (f *StateFunctions) nearestPosInRect(pos Xy, area Rect) Xy {
	if area.From.X > pos.X {
		pos.X = area.From.X
	} else if area.To.X < pos.X {
		pos.X = area.To.X
	}
	if area.From.Y > pos.Y {
		pos.Y = area.From.Y
	} else if area.To.Y < pos.Y {
		pos.Y = area.To.Y
	}
	return pos
}

func (f *StateFunctions) xyLength(vec Xy) int32 {
	x := uint32(vec.X)
	if vec.X < 0 {
		x = uint32(-vec.X)
	}
	y := uint32(vec.Y)
	if vec.Y < 0 {
		y = uint32(-vec.Y)
	}
	if x < y {
		x, y = y, x
	}
	if x/4 < y {
		x = x - x/16 + y*3/8 - x/64 + y*3/256
	}
	return int32(x)
}

func (f *StateFunctions) xyLengthFP8(vec XyFp8) Fp8 {
	return Fp8FromRaw(f.xyLength(Xy{X: vec.X.RawValue, Y: vec.Y.RawValue}))
}

func (f *StateFunctions) toXyFp8(position Xy) XyFp8 {
	return XyFp8{X: Fp8Integer(position.X), Y: Fp8Integer(position.Y)}
}

func (f *StateFunctions) toXy(position XyFp8) Xy {
	return Xy{X: position.X.IntegerPart(), Y: position.Y.IntegerPart()}
}

func (f *StateFunctions) unitsDistance(a *Unit, b *Unit) int32 {
	aRect := f.unitSpriteBoundingBox(a)
	bRect := f.unitSpriteBoundingBox(b)
	return f.xyLength(f.rectDifference(aRect, bRect))
}

func (f *StateFunctions) unitDistanceTo(u *Unit, pos Xy) int32 {
	return f.xyLength(XySub(pos, f.nearestPosInRect(pos, f.unitSpriteBoundingBox(u))))
}

func (f *StateFunctions) atanFP8(x Fp8) DirectionT {
	negative := x.Less(Fp8Zero())
	if negative {
		x = x.Neg()
	}
	uv := uint32(x.RawValue)
	r := sort.Search(len(tanTable), func(i int) bool { return tanTable[i] >= uv })
	d := DirectionTFromRaw(int8(r))
	if negative {
		return d.Neg()
	}
	return d
}

func (f *StateFunctions) xyDirectionFP8(pos XyFp8) DirectionT {
	if pos.X == Fp8FromRaw(0) {
		if pos.Y.LessEqual(Fp8FromRaw(0)) {
			return DirectionTFromRaw(0)
		}
		return DirectionTFromRaw(-128)
	}
	r := f.atanFP8(pos.Y.Div(pos.X))
	if pos.X.Greater(Fp8FromRaw(0)) {
		r = r.Add(DirectionTFromRaw(64))
	} else {
		r = DirectionTFromRaw(-64).Add(r)
	}
	return r
}

func (f *StateFunctions) xyDirection(pos Xy) DirectionT {
	if pos.X == 0 {
		if pos.Y <= 0 {
			return DirectionTFromRaw(0)
		}
		return DirectionTFromRaw(-128)
	}
	r := f.atanFP8(Fp8Integer(pos.Y).DivInt(pos.X))
	if pos.X > 0 {
		r = DirectionTFromRaw(64).Add(r)
	} else {
		r = DirectionTFromRaw(-64).Add(r)
	}
	return r
}

func (f *StateFunctions) directionXY(dir DirectionT, length Fp8) XyFp8 {
	return XyFp8Mul(directionTable[f.directionIndex(dir)], length)
}

func (f *StateFunctions) directionXYInt(dir DirectionT, length int32) XyFp8 {
	return XyFp8MulInt(directionTable[f.directionIndex(dir)], length)
}

func (f *StateFunctions) directionXYOne(dir DirectionT) XyFp8 {
	return directionTable[f.directionIndex(dir)]
}

func (f *StateFunctions) directionIndex(dir DirectionT) uint {
	v := int32(dir.FractionalPart())
	if v < 0 {
		return uint(256 + v)
	}
	return uint(v)
}

func (f *StateFunctions) directionFromIndex(index uint) DirectionT {
	v := int32(index)
	if v >= 128 {
		v = -(256 - v)
	}
	return DirectionTFromRaw(int8(v))
}

func (f *StateFunctions) unitsDirection(from *Unit, to *Unit) DirectionT {
	relpos := XySub(to.Sprite.Position, from.Sprite.Position)
	return f.xyDirection(relpos)
}

func (f *StateFunctions) unitTargetInAttackAngle(u *Unit, targetPos Xy, weapon *WeaponTypeT) bool {
	dir := f.xyDirection(XySub(targetPos, u.Sprite.Position))
	if f.unitIs(u.UnitType, UnitTypesZergLurker) {
		throwError("unit_target_in_attack_angle lurker: unreachable?")
		return true
	}
	return Fp8Extend(dir.Sub(u.Heading)).Abs().LessEqual(weapon.AttackAngle)
}

func (f *StateFunctions) weaponMaxRange(u *Unit, w *WeaponTypeT) int32 {
	rangeUpgradeBonus := func() int32 {
		switch u.UnitType.ID {
		case UnitTypesTerranMarine:
			if f.playerHasUpgrade(u.Owner, UpgradeTypesU238Shells) {
				return 32
			}
			return 0
		case UnitTypesZergHydralisk:
			if f.playerHasUpgrade(u.Owner, UpgradeTypesGroovedSpines) {
				return 32
			}
			return 0
		case UnitTypesProtossDragoon:
			if f.playerHasUpgrade(u.Owner, UpgradeTypesSingularityCharge) {
				return 64
			}
			return 0
		case UnitTypesHeroFenixDragoon:
			return 64
		case UnitTypesTerranGoliath, UnitTypesTerranGoliathTurret:
			if w.ID == WeaponTypesHellfireMissilePack && f.playerHasUpgrade(u.Owner, UpgradeTypesCharonBoosters) {
				return 96
			}
			return 0
		case UnitTypesHeroAlanSchezar, UnitTypesHeroAlanSchezarTurret:
			if w.ID == WeaponTypesHellfireMissilePackAlanSchezar {
				return 96
			}
			return 0
		default:
			return 0
		}
	}
	r := w.MaxRange
	if f.uInBunker(u) {
		r += 64
	}
	r += rangeUpgradeBonus()
	return r
}

func (f *StateFunctions) unitTargetMovementRange(u *Flingy, target *Flingy) int32 {
	if !f.uMovementFlag(u, 2) {
		return 0
	}
	if f.uMovementFlag(target, 2) {
		if Fp8Extend(target.NextVelocityDirection.Sub(u.NextVelocityDirection)).Abs().LessEqual(Fp8FromRaw(32)) {
			return 0
		}
	}
	return f.unitHaltDistance(u).IntegerPart()
}

func (f *StateFunctions) unitTargetInWeaponMovementRange(u *Unit, target *Unit) bool {
	if target == nil {
		target = u.OrderTarget.Unit
	}
	if target == nil {
		return true
	}
	if !f.unitCanSeeTarget(u, target) {
		return false
	}
	w := f.unitTargetWeapon(u, target)
	if w == nil {
		return false
	}
	d := f.unitsDistance(f.unitMainUnit(u), target)
	if w.MinRange != 0 && d <= w.MinRange {
		return false
	}
	maxRange := f.weaponMaxRange(u, w)
	maxRange += f.unitTargetMovementRange(&u.Flingy, &target.Flingy)
	return d <= maxRange
}

func (f *StateFunctions) unitTargetOutOfMaxRange(u *Unit, target *Unit) bool {
	if !f.unitCanSeeTarget(u, target) {
		return true
	}
	w := f.unitTargetWeapon(u, target)
	if w == nil {
		return false
	}
	d := f.unitsDistance(u, target)
	maxRange := f.weaponMaxRange(u, w)
	maxRange += f.unitTargetMovementRange(&u.Flingy, &target.Flingy)
	return d > maxRange
}

func (f *StateFunctions) unitFirstLoadedUnit(u *Unit) *Unit {
	for nu := range f.loadedUnits(u) {
		return nu
	}
	return nil
}

func (f *StateFunctions) unitRace(ut *UnitTypeT) Race {
	if ut.GroupFlags&GroupFlagsZerg != 0 {
		return RaceZerg
	}
	if ut.GroupFlags&GroupFlagsTerran != 0 {
		return RaceTerran
	}
	if ut.GroupFlags&GroupFlagsProtoss != 0 {
		return RaceProtoss
	}
	return RaceNone
}

func (f *StateFunctions) unitProvidesSpace(u *Unit) bool {
	if u.UnitType.SpaceProvided == 0 {
		return false
	}
	if f.unitIs(u.UnitType, UnitTypesZergOverlord) && !f.playerHasUpgrade(u.Owner, UpgradeTypesVentralSacs) {
		return false
	}
	if f.uHallucination(u) {
		return false
	}
	return true
}

func (f *StateFunctions) unitSpaceLeft(u *Unit) uint {
	r := u.UnitType.SpaceProvided
	for nu := range f.loadedUnits(u) {
		r -= nu.UnitType.SpaceRequired
	}
	return r
}

func (f *StateFunctions) unitCanLoadTarget(u *Unit, target *Unit) bool {
	if !f.uCompleted(u) {
		return false
	}
	if f.unitIsDisabled(u) {
		return false
	}
	if f.uBurrowed(target) {
		return false
	}
	if u.Owner != target.Owner {
		return false
	}
	if f.utBuilding(u.UnitType) {
		if f.unitRace(target.UnitType) != RaceTerran {
			return false
		}
		if target.UnitType.SpaceRequired > 1 {
			return false
		}
	}
	return f.unitSpaceLeft(u) >= target.UnitType.SpaceRequired
}

func (f *StateFunctions) unitInterceptorCount(u *Unit) uint {
	if !f.unitIsCarrier(u.UnitType) {
		return 0
	}
	return u.Carrier.InsideCount + u.Carrier.OutsideCount
}

func (f *StateFunctions) unitScarabCount(u *Unit) uint {
	if !f.unitIsReaver(u.UnitType) {
		return 0
	}
	return u.Reaver().InsideCount
}

func (f *StateFunctions) unitMaxInterceptorCount(u *Unit) uint {
	if !f.unitIsCarrier(u.UnitType) {
		return 0
	}
	if f.utHero(u.UnitType) {
		return 8
	}
	if f.playerHasUpgrade(u.Owner, UpgradeTypesCarrierCapacity) {
		return 8
	}
	return 4
}

func (f *StateFunctions) unitMaxScarabCount(u *Unit) uint {
	if !f.unitIsReaver(u.UnitType) {
		return 0
	}
	if f.utHero(u.UnitType) {
		return 10
	}
	if f.playerHasUpgrade(u.Owner, UpgradeTypesReaverCapacity) {
		return 10
	}
	return 5
}

func (f *StateFunctions) unitQueuedFighterUnits(u *Unit) uint {
	var r uint
	for _, ut := range u.BuildQueue.Values() {
		if f.unitIsFighter(ut) {
			r++
		}
	}
	return r
}

func (f *StateFunctions) unitSpiderMineCount(u *Unit) uint {
	if !f.unitIsVulture(u.UnitType) {
		return 0
	}
	return u.Vulture.SpiderMineCount
}

func (f *StateFunctions) unitCanAttack(u *Unit) bool {
	if f.unitOrSubunitGroundWeapon(u) != nil || f.unitOrSubunitAirWeapon(u) != nil {
		return true
	}
	if f.unitInterceptorCount(u) != 0 {
		return true
	}
	if f.unitScarabCount(u) != 0 {
		return true
	}
	return false
}

func (f *StateFunctions) unitTargetAttackPriority(u *Unit, target *Unit) int32 {
	isLoadedUnit := false
	if f.unitIs(target.UnitType, UnitTypesTerranBunker) {
		loadedUnit := f.unitFirstLoadedUnit(target)
		if loadedUnit != nil {
			target = loadedUnit
			isLoadedUnit = true
		}
	}
	if f.unitIs(target.UnitType, UnitTypesZergLarva) {
		return 5
	}
	if f.unitIsEgg(target.UnitType) {
		return 5
	}
	var r int32
	if f.utWorker(target.UnitType) {
		r += 2
	} else if !f.unitCanAttackTarget(target, u) {
		if f.unitCanAttack(target) {
			r += 2
		} else if f.uCanMove(target) {
			r += 3
		} else {
			r += 4
		}
	}
	if isLoadedUnit || !f.uCompleted(target) {
		r++
	}
	if r == 0 && f.uCannotAttack(target) {
		r++
	}
	return r
}

func (f *StateFunctions) unitTargetInRange(u *Unit, target *Unit, rangeValue int32) bool {
	return f.unitsDistance(f.unitMainUnit(u), target) <= rangeValue
}

func (f *StateFunctions) getDefaultPriorityTargets(u *Unit, minDistance int32, maxDistance int32) [6]StaticVector[*Unit, [16]*Unit] {
	pos := u.Sprite.Position
	bounds := Rect{
		From: Xy{X: pos.X - maxDistance - 64, Y: pos.Y - maxDistance - 64},
		To:   Xy{X: pos.X + maxDistance + 64, Y: pos.Y + maxDistance + 64},
	}
	mainUnit := f.unitMainUnit(u)
	attackingUnit := f.unitAttackingUnit(u)
	canTurn := f.uCanTurn(attackingUnit)
	var targets [6]StaticVector[*Unit, [16]*Unit]
	for target := range f.findUnitsNoexpand(bounds) {
		if target == u {
			continue
		}
		if !f.unitTargetIsEnemy(u, target) {
			continue
		}
		if !f.unitTargetIsVisible(u, target) {
			continue
		}
		if !f.unitCanAttackTarget(u, target) {
			continue
		}
		distance := f.unitsDistance(mainUnit, target)
		if distance > maxDistance {
			continue
		}
		if minDistance != 0 && distance <= minDistance {
			continue
		}
		if !canTurn {
			groundWeapon := attackingUnit.UnitType.GroundWeapon.Get()
			if groundWeapon == nil {
				throwError("get_default_priority_targets: null ground weapon")
			}
			if !f.unitTargetInAttackAngle(attackingUnit, target.Sprite.Position, groundWeapon) {
				continue
			}
		}
		prio := f.unitTargetAttackPriority(u, target)
		if targets[prio].Size() < 0x10 {
			targets[prio].PushBack(target)
		}
	}
	return targets
}

func (f *StateFunctions) findAcquireTarget(u *Unit) *Unit {
	acqRange := f.unitTargetAcquisitionRange(u)
	if f.uInBunker(u) {
		acqRange += 2
	}

	var minDistance int32
	maxDistance := acqRange * 32

	groundWeapon := f.unitOrSubunitGroundWeapon(u)
	airWeapon := f.unitOrSubunitAirWeapon(u)
	if groundWeapon != nil {
		if airWeapon == nil {
			minDistance = groundWeapon.MinRange
		} else {
			minDistance = min(groundWeapon.MinRange, airWeapon.MinRange)
		}
	} else if airWeapon != nil {
		minDistance = airWeapon.MinRange
	}
	targets := f.getDefaultPriorityTargets(u, minDistance, maxDistance)
	for i := range targets {
		v := targets[i].Values()
		if len(v) == 0 {
			continue
		}
		best := getBestScore(v, func(target *Unit) int32 {
			return f.xyLength(XySub(target.Sprite.Position, u.Sprite.Position))
		})
		return v[best]
	}
	return nil
}

func (f *StateFunctions) findAcquireRandomTarget(u *Unit) *Unit {
	acqRange := f.unitTargetAcquisitionRange(u)
	if f.uInBunker(u) {
		acqRange += 2
	}

	var minDistance int32
	maxDistance := acqRange * 32

	targets := f.getDefaultPriorityTargets(u, minDistance, maxDistance)
	for i := range targets {
		v := &targets[i]
		if v.Empty() {
			continue
		}
		if v.Size() == 1 {
			return v.Front()
		}
		index := uint(f.lcgRand(19)) % uint(v.Size())
		return v.At(int(index))
	}
	return nil
}

func (f *StateFunctions) unitTypeIsMorphingBuilding(ut *UnitTypeT) bool {
	if f.unitIs(ut, UnitTypesZergHive) {
		return true
	}
	if f.unitIs(ut, UnitTypesZergLair) {
		return true
	}
	if f.unitIs(ut, UnitTypesZergGreaterSpire) {
		return true
	}
	if f.unitIs(ut, UnitTypesZergSporeColony) {
		return true
	}
	if f.unitIs(ut, UnitTypesZergSunkenColony) {
		return true
	}
	return false
}

func (f *StateFunctions) unitIsMorphingBuilding(u *Unit) bool {
	if f.uCompleted(u) {
		return false
	}
	if u.BuildQueue.Empty() {
		return false
	}
	return f.unitTypeIsMorphingBuilding(u.BuildQueue.Front())
}

func (f *StateFunctions) unitSightRange(u *Unit, ignoreBlindness bool) int32 {
	if f.uGroundedBuilding(u) && !f.uCompleted(u) && !f.unitIsMorphingBuilding(u) {
		return 32 * 4
	}
	if !ignoreBlindness && u.BlindedBy != 0 {
		return 32 * 2
	}
	if f.unitIs(u.UnitType, UnitTypesTerranGhost) && f.playerHasUpgrade(u.Owner, UpgradeTypesOcularImplants) {
		return 32 * 11
	}
	if f.unitIs(u.UnitType, UnitTypesZergOverlord) && f.playerHasUpgrade(u.Owner, UpgradeTypesAntennae) {
		return 32 * 11
	}
	if f.unitIs(u.UnitType, UnitTypesProtossObserver) && f.playerHasUpgrade(u.Owner, UpgradeTypesSensorArray) {
		return 32 * 11
	}
	if f.unitIs(u.UnitType, UnitTypesProtossScout) && f.playerHasUpgrade(u.Owner, UpgradeTypesApialSensors) {
		return 32 * 11
	}
	return 32 * u.UnitType.SightRange
}

func (f *StateFunctions) unitSightRangeIgnoreBlindness(u *Unit) int32 {
	return f.unitSightRange(u, true)
}

func (f *StateFunctions) unitTargetAcquisitionRange(u *Unit) int32 {
	if (f.uCloaked(u) || f.uRequiresDetector(u)) && u.OrderType.ID != OrdersHoldPosition {
		if f.unitIsGhost(u.UnitType) {
			return 0
		}
	}
	var bonus int32
	if f.unitIs(u.UnitType, UnitTypesTerranMarine) && f.playerHasUpgrade(u.Owner, UpgradeTypesU238Shells) {
		bonus = 1
	}
	if f.unitIs(u.UnitType, UnitTypesZergHydralisk) && f.playerHasUpgrade(u.Owner, UpgradeTypesGroovedSpines) {
		bonus = 1
	}
	if f.unitIs(u.UnitType, UnitTypesProtossDragoon) && f.playerHasUpgrade(u.Owner, UpgradeTypesSingularityCharge) {
		bonus = 2
	}
	if f.unitIs(u.UnitType, UnitTypesHeroFenixDragoon) {
		bonus = 2
	}
	if f.unitIs(u.UnitType, UnitTypesTerranGoliath) && f.playerHasUpgrade(u.Owner, UpgradeTypesCharonBoosters) {
		bonus = 3
	}
	if f.unitIs(u.UnitType, UnitTypesTerranGoliathTurret) && f.playerHasUpgrade(u.Owner, UpgradeTypesCharonBoosters) {
		bonus = 3
	}
	if f.unitIs(u.UnitType, UnitTypesHeroAlanSchezar) {
		bonus = 3
	}
	if f.unitIs(u.UnitType, UnitTypesHeroAlanSchezarTurret) {
		bonus = 3
	}
	return u.UnitType.TargetAcquisitionRange + bonus
}

func (f *StateFunctions) unitMaxEnergy(u *Unit) Fp8 {
	if f.utHero(u.UnitType) {
		return Fp8Integer(250)
	}
	energyUpgrade := func() UpgradeTypes {
		switch u.UnitType.ID {
		case UnitTypesTerranGhost:
			return UpgradeTypesMoebiusReactor
		case UnitTypesTerranWraith:
			return UpgradeTypesApolloReactor
		case UnitTypesTerranScienceVessel:
			return UpgradeTypesTitanReactor
		case UnitTypesTerranBattlecruiser:
			return UpgradeTypesColossusReactor
		case UnitTypesTerranMedic:
			return UpgradeTypesCaduceusReactor
		case UnitTypesZergQueen:
			return UpgradeTypesGameteMeiosis
		case UnitTypesZergDefiler:
			return UpgradeTypesMetasynapticNode
		case UnitTypesProtossCorsair:
			return UpgradeTypesArgusJewel
		case UnitTypesProtossDarkArchon:
			return UpgradeTypesArgusTalisman
		case UnitTypesProtossHighTemplar:
			return UpgradeTypesKhaydarinAmulet
		case UnitTypesProtossArbiter:
			return UpgradeTypesKhaydarinCore
		default:
			return UpgradeTypesNone
		}
	}
	upg := energyUpgrade()
	if upg != UpgradeTypesNone && f.playerHasUpgrade(u.Owner, upg) {
		return Fp8Integer(250)
	}
	return Fp8Integer(200)
}

func (f *StateFunctions) visibleToEveryone(u *Unit) bool {
	if f.utWorker(u.UnitType) {
		if u.Worker.Powerup != nil && f.unitIs(u.Worker.Powerup.UnitType, UnitTypesPowerupFlag) {
			return true
		}
		return false
	}
	if !f.unitProvidesSpace(u) {
		return false
	}
	for lu := range f.loadedUnits(u) {
		if lu.Worker.Powerup != nil && f.unitIs(lu.Worker.Powerup.UnitType, UnitTypesPowerupFlag) {
			return true
		}
	}
	return false
}

func (f *StateFunctions) tileIndex(pos Xy) uint {
	ux := uint(pos.X) / 32
	uy := uint(pos.Y) / 32
	if ux >= f.gameSt.MapTileWidth || uy >= f.gameSt.MapTileHeight {
		throwError("attempt to get tile index for invalid position %d %d", pos.X, pos.Y)
	}
	return uy*f.gameSt.MapTileWidth + ux
}

func (f *StateFunctions) tileVisibility(pos Xy) uint8 {
	return ^f.st.Tiles[f.tileIndex(pos)].Visible
}

func (f *StateFunctions) tileExplored(pos Xy) uint8 {
	return ^f.st.Tiles[f.tileIndex(pos)].Explored
}

func (f *StateFunctions) getGroundHeightAt(pos Xy) int32 {
	index := f.tileIndex(pos)
	tileID := f.gameSt.GfxTiles[index]
	megatileIndex := uint(f.gameSt.Cv5[tileID.GroupIndex()].MegaTileIndex[tileID.SubtileIndex()])
	ux := uint(pos.X)
	uy := uint(pos.Y)
	flags := int32(f.gameSt.Vf4[megatileIndex].Flags[uy/8%4*4+ux/8%4])
	if flags&int32(Vf4EntryFlagHigh) != 0 {
		return 2
	}
	if flags&int32(Vf4EntryFlagMiddle) != 0 {
		return 1
	}
	return 0
}

func (f *StateFunctions) displayLastErrorForPlayer(player int32) {
}

func (f *StateFunctions) playerHasUnit(owner int32, unitTypeID UnitTypes) bool {
	return f.st.UnitCounts[owner][unitTypeID] != 0
}

func (f *StateFunctions) playerHasCompletedUnit(owner int32, unitTypeID UnitTypes) bool {
	return f.st.CompletedUnitCounts[owner][unitTypeID] != 0
}

func (f *StateFunctions) playerTechAvailable(owner int32, techID TechTypes) bool {
	return f.gameSt.TechAvailable[owner][techID]
}

func (f *StateFunctions) playerHasResearched(owner int32, techID TechTypes) bool {
	return f.st.TechResearched[owner][techID]
}

func (f *StateFunctions) playerIsResearching(owner int32, techID TechTypes) bool {
	return f.st.TechResearching[owner][techID]
}

func (f *StateFunctions) playerMaxUpgradeLevel(owner int32, upgradeID UpgradeTypes) int32 {
	return f.gameSt.MaxUpgradeLevels[owner][upgradeID]
}

func (f *StateFunctions) playerHasUpgrade(owner int32, upgradeID UpgradeTypes) bool {
	return f.st.UpgradeLevels[owner][upgradeID] != 0
}

func (f *StateFunctions) playerIsUpgrading(owner int32, upgradeID UpgradeTypes) bool {
	return f.st.UpgradeUpgrading[owner][upgradeID]
}

func (f *StateFunctions) playerUpgradeLevel(owner int32, upgradeID UpgradeTypes) int32 {
	return f.st.UpgradeLevels[owner][upgradeID]
}

func (f *StateFunctions) unitIs(ut *UnitTypeT, unitTypeID UnitTypes) bool {
	return ut.ID == unitTypeID
}

func (f *StateFunctions) unitAddon(u *Unit) *Unit {
	if f.utBuilding(u.UnitType) {
		return u.Building.Addon
	}
	return nil
}

func (f *StateFunctions) unitHasAddon(u *Unit, unitTypeID UnitTypes) bool {
	return f.utBuilding(u.UnitType) && u.Building.Addon != nil && f.unitIs(u.Building.Addon.UnitType, unitTypeID)
}

func (f *StateFunctions) unitIsBuildingAddon(u *Unit) bool {
	if u.SecondaryOrderType.ID == OrdersBuildAddon {
		if f.uGroundedBuilding(u) && u.CurrentBuildUnit != nil && !f.uCompleted(u.CurrentBuildUnit) {
			return true
		}
	}
	return false
}

func (f *StateFunctions) unitIsBuildingProtossThing(u *Unit) bool {
	if u.SecondaryOrderType.ID == OrdersPlaceProtossBuilding {
		if f.uGroundedBuilding(u) && u.OrderTarget.Unit != nil && !f.uCompleted(u.OrderTarget.Unit) {
			return true
		}
	}
	return false
}

func (f *StateFunctions) unitIsConstructing(u *Unit) bool {
	if !u.BuildQueue.Empty() {
		return true
	}
	if f.unitIsBuildingAddon(u) {
		return true
	}
	if f.unitIsBuildingProtossThing(u) {
		return true
	}
	return false
}

func (f *StateFunctions) unitIsResearching(u *Unit) bool {
	return u.Building.ResearchingType != nil
}

func (f *StateFunctions) unitIsUpgrading(u *Unit) bool {
	return u.Building.UpgradingType != nil
}

func (f *StateFunctions) unitCanMove(u *Unit) bool {
	if f.uGroundedBuilding(u) {
		return false
	}
	if u.UnitType.RightClickAction == 0 {
		return false
	}
	if u.UnitType.RightClickAction == 3 {
		return false
	}
	if f.unitIs(u.UnitType, UnitTypesZergLurker) && f.uBurrowed(u) {
		return false
	}
	return true
}

func (f *StateFunctions) unitCanHoldPosition(u *Unit) bool {
	if u.UnitType.RightClickAction != 0 {
		return true
	}
	if f.uGroundedBuilding(u) && f.unitIsFactory(u) {
		return true
	}
	return false
}

func (f *StateFunctions) unitIsImmovableAttacker(u *Unit) bool {
	if f.unitIs(u.UnitType, UnitTypesZergLurker) && f.uBurrowed(u) {
		return true
	}
	if u.UnitType.RightClickAction == 3 {
		return true
	}
	return false
}
