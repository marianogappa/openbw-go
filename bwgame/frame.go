package bwgame

import "iter"

const revealSightMaxWidth = 11*2 + 3

func (f *StateFunctions) revealSightAt(pos Xy, rangeValue int32, revealTo int32, inAir bool) {
	visibilityMask := ^revealTo
	heightMask := int32(0)
	if !inAir {
		height := f.getGroundHeightAt(pos)
		if height == 2 {
			heightMask = TileFlagVeryHigh
		} else if height == 1 {
			heightMask = TileFlagVeryHigh | TileFlagHigh
		} else {
			heightMask = TileFlagVeryHigh | TileFlagHigh | TileFlagMiddle
		}
	}
	visionPropagation := &f.visionPropagation
	requiredTileMask := uint32(heightMask)<<16 | uint32(uint8(^visibilityMask))<<8 | uint32(uint8(^visibilityMask))
	sightVals := &f.gameSt.SightValues[rangeValue]
	tileX := uint(pos.X) / 32
	tileY := uint(pos.Y) / 32
	baseTileIndex := int(tileX + tileY*f.gameSt.MapTileWidth)
	tileAt := func(relativeTileIndex int32) *Tile {
		return &f.st.Tiles[baseTileIndex+int(relativeTileIndex)]
	}
	isOutsideMap := func(cur *SightValuesMaskdatNode) bool {
		return tileX+uint(cur.X) >= f.gameSt.MapTileWidth || tileY+uint(cur.Y) >= f.gameSt.MapTileHeight
	}
	reveal := func(tile *Tile) {
		tile.Visible &= uint8(visibilityMask)
		tile.Explored &= uint8(visibilityMask)
	}
	if !inAir {
		index := uint(0)
		end := uint(sightVals.MinMaskSize)
		for ; index != end; index++ {
			cur := &sightVals.Maskdat[index]
			visionPropagation[index] = 0xff
			if isOutsideMap(cur) {
				continue
			}
			tile := tileAt(cur.RelativeTileIndex)
			reveal(tile)
			visionPropagation[index] = uint32(tile.Flags)<<16 | uint32(tile.Explored)<<8 | uint32(tile.Visible)
		}
		end += uint(sightVals.ExtMaskedCount)
		for ; index != end; index++ {
			cur := &sightVals.Maskdat[index]
			visionPropagation[index] = 0xff
			if isOutsideMap(cur) {
				continue
			}
			if visionPropagation[cur.Prev]&requiredTileMask != 0 {
				if cur.Prev2 == ^uint(0) || visionPropagation[cur.Prev2]&requiredTileMask != 0 {
					continue
				}
			}
			tile := tileAt(cur.RelativeTileIndex)
			reveal(tile)
			visionPropagation[index] = uint32(tile.Flags)<<16 | uint32(tile.Explored)<<8 | uint32(tile.Visible)
		}
	} else {
		for i := range uint(sightVals.ExtMaskedCount) {
			cur := &sightVals.Maskdat[i]
			if isOutsideMap(cur) {
				continue
			}
			reveal(tileAt(cur.RelativeTileIndex))
		}
	}
}

func (f *StateFunctions) refreshUnitVision(u *Unit) {
	if u.Owner >= 8 && u.ParasiteFlags == 0 {
		return
	}
	if f.unitIs(u.UnitType, UnitTypesTerranNuclearMissile) {
		return
	}
	visibleTo := int32(0)
	if f.visibleToEveryone(u) || (f.unitIs(u.UnitType, UnitTypesPowerupFlag) && u.OrderType.ID == OrdersUnusedPowerup) {
		visibleTo = 0xff
	} else {
		visibleTo = int32(f.st.SharedVision[u.Owner]) | u.ParasiteFlags
		if u.ParasiteFlags != 0 {
			visibleTo |= u.ParasiteFlags
			for i := range uint(12) {
				if ^u.ParasiteFlags&(1<<i) != 0 {
					continue
				}
				visibleTo |= int32(f.st.SharedVision[i])
			}
		}
	}
	f.revealSightAt(u.Sprite.Position, int32(uint32(f.unitSightRange(u, false))/32), visibleTo, f.uFlying(u))
}

const turretFollowsSubunitHeadingFlag int32 = 0x2000000

func (f *StateFunctions) turnTurret(u *Unit, turn DirectionT) {
	if u.OrderTarget.Unit != nil {
		f.uUnsetStatusFlag(u, turretFollowsSubunitHeadingFlag)
	} else {
		if u.Heading == u.Subunit.Heading {
			f.uSetStatusFlag(u, turretFollowsSubunitHeadingFlag)
		}
	}
	if f.uStatusFlag(u, turretFollowsSubunitHeadingFlag) {
		f.setUnitHeading(u, u.Subunit.Heading)
	} else {
		u.NextVelocityDirection = u.NextVelocityDirection.Add(turn)
		u.Heading = u.NextVelocityDirection
	}
	if f.unitIs(u.UnitType, UnitTypesTerranGoliathTurret) || f.unitIs(u.UnitType, UnitTypesHeroAlanSchezarTurret) {
		diff := u.Subunit.Heading.Sub(u.Heading)
		if diff.Greater(DirectionTFromRaw(32)) {
			u.Heading = u.Subunit.Heading.Sub(DirectionTFromRaw(32))
		} else if diff.Less(DirectionTFromRaw(-32)) {
			u.Heading = u.Subunit.Heading.Add(DirectionTFromRaw(32))
		}
	}
}

func (f *StateFunctions) updateUnitMovement(u *Unit) {
	prevVelocityDirection := u.NextVelocityDirection
	refreshVision := f.executeMovement(u)
	if refreshVision {
		f.refreshUnitVision(u)
	}

	if f.uCompleted(u) && u.Subunit != nil && !f.utTurret(u.UnitType) {
		subunit := u.Subunit
		f.turnTurret(subunit, u.NextVelocityDirection.Sub(prevVelocityDirection))
		subunit.ExactPosition = u.ExactPosition
		subunit.Position = f.toXy(u.ExactPosition)
		f.moveSprite(subunit.Sprite, subunit.Position)
		f.setImageOffset(subunit.Sprite.MainImage, f.getImageLoOffset(u.Sprite.MainImage, 2, 0, false))
		previousIscriptUnit := f.iscriptUnit
		f.iscriptUnit = subunit
		if !f.uMovementFlag(&u.Flingy, 2) {
			if f.uStatusFlag(subunit, UnitStatusFlagTurretWalking) {
				f.uUnsetStatusFlag(subunit, UnitStatusFlagTurretWalking)
				if f.uCanMove(u) && !f.uMovementFlag(&subunit.Flingy, 8) {
					f.spriteRunAnim(subunit.Sprite, IscriptAnimsWalkingToIdle)
				}
			}
		} else {
			if !f.uStatusFlag(subunit, UnitStatusFlagTurretWalking) {
				f.uSetStatusFlag(subunit, UnitStatusFlagTurretWalking)
				if f.uCanMove(u) && !f.uMovementFlag(&subunit.Flingy, 8) {
					f.spriteRunAnim(subunit.Sprite, IscriptAnimsWalking)
				}
			}
		}
		f.updateUnitMovement(subunit)
		f.iscriptUnit = previousIscriptUnit
	}
}

func (f *StateFunctions) updateThingyVisibility(t *Thingy, size Xy) bool {
	if t.Sprite == nil || t.Sprite.Flags&SpriteFlagHidden != 0 {
		return false
	}
	tileFromX := (t.Sprite.Position.X - size.X/2) / 32
	tileFromY := (t.Sprite.Position.Y - size.Y/2) / 32
	tileToX := uint(uint32(tileFromX) + uint32(size.X+31)/32)
	tileToY := uint(uint32(tileFromY) + uint32(size.Y+31)/32)
	if tileFromX < 0 {
		tileFromX = 0
	}
	if tileFromY < 0 {
		tileFromY = 0
	}
	if tileToX > f.gameSt.MapTileWidth {
		tileToX = f.gameSt.MapTileWidth
	}
	if tileToY > f.gameSt.MapTileHeight {
		tileToY = f.gameSt.MapTileHeight
	}

	if uint(tileFromX) == tileToX && uint(tileFromY) == tileToY {
		return false
	}

	visibilityFlags := uint8(0)
	for y := uint(tileFromY); y != tileToY; y++ {
		for x := uint(tileFromX); x != tileToX; x++ {
			visibilityFlags |= ^f.st.Tiles[y*f.gameSt.MapTileWidth+x].Visible
		}
	}
	if t.Sprite.VisibilityFlags != int32(visibilityFlags) {
		f.setSpriteVisibility(t.Sprite, int32(visibilityFlags))
		return true
	}
	return false
}

func (f *StateFunctions) updateUnitSprite(u *Unit) {
	if f.updateThingyVisibility(&u.Thingy, u.UnitType.PlacementSize) {
		if u.Subunit != nil && !f.usHidden(&u.Subunit.Thingy) {
			f.setSpriteVisibility(u.Subunit.Sprite, u.Sprite.VisibilityFlags)
		}
	}
}

func (f *StateFunctions) unitCanFireFromBunker(u *Unit) bool {
	if f.unitIsMarine(u.UnitType) {
		return true
	}
	if f.unitIsFirebat(u.UnitType) {
		return true
	}
	if f.unitIsGhost(u.UnitType) {
		return true
	}
	return false
}

func (f *StateFunctions) createBunkerFireAnimation(u *Unit) {
	var sprite *SpriteTypeT
	index := (f.directionIndex(u.Heading) + 16) / 32 & 7
	var heading DirectionT
	if f.unitIsFirebat(u.UnitType) {
		sprite = f.getSpriteType(SpriteTypesSPRITEIDFlameThrower)
		heading = f.directionFromIndex(16 * ((f.directionIndex(u.Heading) + 8) / 16 & 15))
	} else {
		sprite = f.getSpriteType(SpriteTypesSPRITEIDBunkerOverlay)
		heading = f.directionFromIndex(16 * 2 * index)
	}
	offset := f.getImageLoOffset(u.ConnectedUnit.Sprite.MainImage, 0, index, false)
	t := f.createThingy(sprite, XyAdd(u.Sprite.Position, offset), 0)
	if t != nil {
		t.Sprite.ElevationLevel = u.Sprite.ElevationLevel + 1
		if !f.usHidden(t) {
			f.setSpriteVisibility(t.Sprite, int32(f.tileVisibility(t.Sprite.Position)))
		}
		for i := range t.Sprite.Images.All() {
			f.setImageHeading(i, heading)
		}
	}
}

func (f *StateFunctions) tryAttackSomething(u *Unit) bool {
	if u.OrderTarget.Unit != nil {
		f.attackUnitReacquireTarget(u)
	}
	if u.OrderTarget.Unit == nil {
		return false
	}
	if !f.unitCanSeeTarget(u, u.OrderTarget.Unit) {
		return false
	}
	target := u.OrderTarget.Unit
	if !f.unitCanAttackTarget(u, target) {
		return false
	}
	u.OrderTarget.Pos = target.Sprite.Position
	if f.unitTurret(u) != nil {
		return true
	}

	var weapon *WeaponTypeT
	var cooldown int32
	var anim int32
	if target != nil && f.uFlying(target) {
		weapon = f.unitAirWeapon(u)
		cooldown = u.AirWeaponCooldown
		anim = IscriptAnimsAirAttkRpt
	} else {
		weapon = f.unitGroundWeapon(u)
		cooldown = u.GroundWeaponCooldown
		anim = IscriptAnimsGndAttkRpt
	}
	if weapon == nil {
		return false
	}
	if cooldown != 0 {
		if u.OrderProcessTimer > cooldown-1 {
			u.OrderProcessTimer = cooldown - 1
		}
		return true
	}

	if f.uMovementFlag(&u.Flingy, 8) {
		return true
	}
	distance := f.unitsDistance(f.unitMainUnit(u), target)
	if weapon.MinRange != 0 && distance < weapon.MinRange {
		return false
	}
	if distance > f.weaponMaxRange(u, weapon) {
		return false
	}

	headingError := Fp8Extend(u.Heading.Sub(f.xyDirection(XySub(target.Sprite.Position, u.Sprite.Position)))).Abs()
	if headingError.Greater(weapon.AttackAngle) {
		if f.uCanTurn(u) {
			u.OrderProcessTimer = 0
			return true
		}
		return false
	}

	if f.uCannotAttack(u) {
		return false
	}
	if f.uInBunker(u) {
		f.createBunkerFireAnimation(u)
	}

	f.st.PrevBulletSourceUnit = nil
	f.uSetMovementFlag(&u.Flingy, 8)
	cooldown = f.getModifiedWeaponCooldown(u, weapon) + (f.lcgRand(12) & 3) - 1
	u.GroundWeaponCooldown = cooldown
	u.AirWeaponCooldown = cooldown
	f.spriteRunAnim(u.Sprite, anim)
	return true
}

func (f *StateFunctions) orderHiddenBunkerGuard(u *Unit) {
	if !f.unitCanFireFromBunker(u) {
		return
	}
	f.uSetStatusFlag(u, UnitStatusFlagReadyToAttack)
	if u.Subunit != nil {
		f.uSetStatusFlag(u.Subunit, UnitStatusFlagReadyToAttack)
	}
	if f.tryAttackSomething(u) {
		f.setNextTargetWaypoint(&u.Flingy, u.OrderTarget.Pos)
	} else {
		if u.MainOrderTimer == 0 {
			u.MainOrderTimer = 15
			u.OrderTarget.Unit = f.findAcquireRandomTarget(u)
			if u.OrderTarget.Unit != nil {
				u.OrderProcessTimer = 0
			}
		}
	}
}

func (f *StateFunctions) executeHiddenUnitMainOrder(u *Unit) {
	switch u.OrderType.ID {
	case OrdersDie:
		f.orderDie(u)
		return
	case OrdersPlayerGuard, OrdersTurretGuard, OrdersTurretAttack, OrdersEnterTransport:
		if f.uInBunker(u) {
			f.setUnitOrder(u, f.getOrderType(OrdersBunkerGuard))
		} else {
			f.setUnitOrder(u, f.getOrderType(OrdersNothing))
		}
		return
	case OrdersUnusedPowerup:
		throwError("hidden UnusedPowerup")
		return
	case OrdersNothing:
		return
	case OrdersNothingWait:
		return
	case OrdersInfestingCommandCenter:
		f.orderInfestingCommandCenter(u)
		return
	case OrdersHarvestGas:
		f.orderHarvestGas(u)
		return
	case OrdersPowerupIdle:
		throwError("hidden PowerupIdle")
		return
	case OrdersNukeLaunch:
		f.orderNukeLaunch(u)
		return
	case OrdersResetCollision:
		f.orderResetCollision(u)
		return
	case OrdersResetHarvestCollision:
		f.orderResetHarvestCollision(u)
		return
	case OrdersNeutral:
		return
	case OrdersMedicIdle:
		return
	case OrdersMedicHeal:
		return
	}
	if u.OrderProcessTimer != 0 {
		u.OrderProcessTimer--
		return
	}
	u.OrderProcessTimer = 8
	switch u.OrderType.ID {
	case OrdersBunkerGuard:
		f.orderHiddenBunkerGuard(u)
	case OrdersEnterTransport:
		throwError("hidden EnterTransport")
	case OrdersComputerAI:
		throwError("hidden ComputerAI")
	case OrdersRescuePassive:
		throwError("hidden RescuePassive")
	}
}

func (f *StateFunctions) executeHiddenUnitSecondaryOrder(u *Unit) {
	switch u.SecondaryOrderType.ID {
	case OrdersTrainFighter:
		f.secondaryOrderTrainFighter(u)
	case OrdersCloak:
		f.secondaryOrderCloak(u)
	case OrdersDecloak:
		f.secondaryOrderDecloak(u)
	}
}

func (f *StateFunctions) updateHiddenUnit(u *Unit) {
	if u.Subunit != nil && !f.utTurret(u.UnitType) {
		previousIscriptUnit := f.iscriptUnit
		f.iscriptUnit = u.Subunit
		f.updateHiddenUnit(u.Subunit)
		f.iscriptUnit = previousIscriptUnit
	}
	f.executeMovement(u)
	f.updateUnitValues(u)

	f.executeHiddenUnitMainOrder(u)
	f.executeHiddenUnitSecondaryOrder(u)

	if u.Sprite != nil {
		if !f.iscriptExecuteSprite(u.Sprite) {
			u.Sprite = nil
		}
	}
}

func (f *StateFunctions) unitCanDetect(u *Unit) bool {
	if !f.utDetector(u.UnitType) {
		return false
	}
	if !f.uCompleted(u) {
		return false
	}
	if f.unitIsDisabled(u) {
		return false
	}
	if u.BlindedBy != 0 {
		return false
	}
	return true
}

func (f *StateFunctions) unitCalculateDetectedFlags(u *Unit) uint32 {
	if u.DefensiveMatrixHp.RawValue != 0 || u.LockdownTimer != 0 || u.MaelstromTimer != 0 || u.IrradiateTimer != 0 || u.EnsnareTimer != 0 || u.StasisTimer != 0 || u.PlagueTimer != 0 || u.AcidSporeCount != 0 {
		return 0xff
	} else if f.visibleToEveryone(u) {
		return 0xff
	}

	detectedFlags := uint32(0)

	detectedFlags |= uint32(int32(1) << u.Owner)
	detectedFlags |= f.st.SharedVision[u.Owner]
	detectedFlags |= uint32(u.ParasiteFlags)

	for i := range uint(len(f.st.SharedVision)) {
		if u.ParasiteFlags&(1<<i) != 0 {
			detectedFlags |= f.st.SharedVision[i]
		}
	}

	test := func(detector *Unit) {
		if !f.unitCanDetect(detector) {
			return
		}
		if detector == u {
			return
		}
		if f.uHallucination(detector) {
			return
		}
		if ^u.Sprite.VisibilityFlags&(1<<detector.Owner) != 0 {
			return
		}
		var rangeValue int32
		if f.uGroundedBuilding(detector) {
			rangeValue = 32 * 7
		} else {
			rangeValue = f.unitSightRange(detector, false)
		}
		if !f.unitTargetInRange(detector, u, rangeValue) {
			return
		}
		detectedFlags |= uint32(int32(1)<<detector.Owner) | f.st.SharedVision[detector.Owner] | uint32(detector.ParasiteFlags)
	}
	for nu := range f.findUnitsNoexpand(f.squareAt(u.Sprite.Position, 32*11)) {
		test(nu)
	}
	for nu := range f.st.MapRevealerUnits.All() {
		test(nu)
	}

	return detectedFlags
}

func (f *StateFunctions) removeTargetReferences(u *Unit, target *Unit) {
	test := func(ref **Unit) bool {
		if *ref == target {
			*ref = nil
			return true
		}
		return false
	}
	test(&u.CurrentBuildUnit)
	test(&u.OrderTarget.Unit)
	test(&u.MoveTarget.Unit)
	test(&u.AutoTargetUnit)
	test(&u.ConnectedUnit)
	test(&u.IrradiatedBy)
	if f.utWorker(u.UnitType) {
		test(&u.Worker.TargetResourceUnit)
		if u.Worker.IsGathering && test(&u.Worker.GatherTarget) {
			u.Worker.IsGathering = false
		}
	}
	if f.uGroundedBuilding(u) {
		test(&u.Building.Addon)
	}
	if f.unitIsFactory(u) {
		if u.Building.Rally.Unit == target {
			u.Building.Rally.Pos = target.Sprite.Position
		}
		test(&u.Building.Rally.Unit)
	}
	if turret := f.unitTurret(u); turret != nil {
		f.removeTargetReferences(turret, target)
	}
	for i := u.OrderQueue.Begin(); i != u.OrderQueue.End(); {
		o := i
		i = u.OrderQueue.Next(i)
		if o.Target.Unit == target {
			f.removeQueuedOrder(u, o)
		}
	}
}

func (f *StateFunctions) removeBulletTargetReferences(b *Bullet, target *Unit) {
	if b.BulletTarget == target {
		b.BulletTarget = nil
	}
	if b.BulletOwnerUnit == target {
		b.BulletOwnerUnit = nil
	}
}

func (f *StateFunctions) updateUnitDetectedFlags(u *Unit) {
	newFlags := f.unitCalculateDetectedFlags(u)
	if u.DetectedFlags == newFlags {
		return
	}
	oldFlags := u.DetectedFlags
	if oldFlags == 0x80000000 {
		oldFlags = ^newFlags & 0xff
	}
	removedFlags := oldFlags &^ newFlags
	u.DetectedFlags = newFlags
	for i := uint(0); removedFlags != 0; i++ {
		if removedFlags&(uint32(1)<<i) != 0 {
			removedFlags &^= uint32(1) << i
			for nu := range f.st.PlayerUnits[i].All() {
				f.removeTargetReferences(nu, u)
			}
		}
	}
}

func (f *StateFunctions) updatePsionicMatrix() {
	f.st.UpdatePsionicMatrix = false
	for u := range f.st.VisibleUnits.All() {
		if !f.uGroundedBuilding(u) {
			continue
		}
		if f.unitRace(u.UnitType) != RaceProtoss {
			continue
		}
		if f.st.Players[u.Owner].Controller == PlayerControllerRescuePassive {
			continue
		}
		if !f.utRequiresPsionicMatrix(u.UnitType) {
			continue
		}
		wasDisabled := f.uDisabled(u)
		if f.isInPsionicMatrix(u.Owner, u.Sprite.Position) {
			f.uUnsetStatusFlag(u, UnitStatusFlagDisabled)
			if f.uCompleted(u) && wasDisabled {
				f.spriteRunAnim(u.Sprite, IscriptAnimsEnable)
				if f.unitIsBuildingProtossThing(u) || f.unitIsUpgrading(u) {
					f.spriteRunAnim(u.Sprite, IscriptAnimsIsWorking)
				}
			}
		} else {
			f.uSetStatusFlag(u, UnitStatusFlagDisabled)
			f.iscriptRunToIdle(u)
			if f.uCompleted(u) && !wasDisabled {
				f.spriteRunAnim(u.Sprite, IscriptAnimsDisable)
			}
		}
	}
}

func (f *StateFunctions) updateDeadUnit(u *Unit) {
	if u.Sprite != nil {
		if f.usHidden(&u.Thingy) {
			f.destroySprite(u.Sprite)
			u.Sprite = nil
		} else if !f.iscriptExecuteSprite(u.Sprite) {
			u.Sprite = nil
		}
	}
	if f.uStatusFlag(u, UnitStatusFlag4000) {
		throwError("update_dead_unit: fixme some creep stuff?")
	}
	if u.Sprite != nil {
		if f.updateTiles && f.st.Players[u.Owner].Controller == PlayerControllerOccupied {
			f.revealSightAt(u.Sprite.Position, 1, int32(f.st.SharedVision[u.Owner]), f.uFlying(u))
		}
		f.updateUnitSprite(u)
	} else {
		for !u.OrderQueue.Empty() {
			f.removeQueuedOrder(u, u.OrderQueue.Front())
		}
		f.st.DeadUnits.Remove(u)
		f.st.UnitsContainer.Push(u)
	}
}

func (f *StateFunctions) updateDisruptionWeb() {
	st := f.st
	if st.DisruptionWebbedUnits != 0 {
		for u := range st.VisibleUnits.All() {
			if !f.uFlying(u) || f.uCannotAttack(u) {
				if f.uCannotAttack(u) {
					f.uUnsetStatusFlag(u, UnitStatusFlagCannotAttack)
					st.DisruptionWebbedUnits--
				}
				if f.unitProvidesSpace(u) {
					for nu := range f.loadedUnits(u) {
						if f.uCannotAttack(nu) {
							f.uUnsetStatusFlag(nu, UnitStatusFlagCannotAttack)
							st.DisruptionWebbedUnits--
						}
					}
				} else if u.Subunit != nil {
					if f.uCannotAttack(u.Subunit) {
						f.uUnsetStatusFlag(u.Subunit, UnitStatusFlagCannotAttack)
						st.DisruptionWebbedUnits--
					}
				}
			}
		}
	}
	if st.CompletedUnitCounts[11][UnitTypesSpellDisruptionWeb] != 0 {
		for u := range st.PlayerUnits[11].All() {
			if !f.unitIs(u.UnitType, UnitTypesSpellDisruptionWeb) {
				continue
			}
			for target := range f.findUnitsNoexpand(f.unitSpriteInnerBoundingBox(u)) {
				if f.uFlying(target) {
					continue
				}
				if !f.uCannotAttack(target) {
					f.uSetStatusFlag(target, UnitStatusFlagCannotAttack)
					st.DisruptionWebbedUnits++
				}
				if f.unitProvidesSpace(target) {
					for n := range f.loadedUnits(target) {
						if !f.uCannotAttack(n) {
							f.uSetStatusFlag(n, UnitStatusFlagCannotAttack)
							st.DisruptionWebbedUnits++
						}
					}
				} else if target.Subunit != nil {
					if !f.uCannotAttack(target.Subunit) {
						f.uSetStatusFlag(target.Subunit, UnitStatusFlagCannotAttack)
						st.DisruptionWebbedUnits++
					}
				}
			}
		}
	}
}

func (f *StateFunctions) updateUnits() {
	st := f.st
	st.OrderTimerCounter--
	if st.OrderTimerCounter == 0 {
		st.OrderTimerCounter = 150
		v := int32(0)
		for u := range st.VisibleUnits.All() {
			u.OrderProcessTimer = v
			v++
			if v == 8 {
				v = 0
			}
		}
	}
	st.SecondaryOrderTimerCounter--
	if st.SecondaryOrderTimerCounter == 0 {
		st.SecondaryOrderTimerCounter = 300
		v := int32(0)
		for u := range st.VisibleUnits.All() {
			u.SecondaryOrderTimer = v
			v++
			if v == 30 {
				v = 0
			}
		}
	}

	f.updateDisruptionWeb()

	for i := st.DeadUnits.Begin(); i != st.DeadUnits.End(); {
		u := i
		i = st.DeadUnits.Next(i)
		f.iscriptFlingy = &u.Flingy
		f.iscriptUnit = u
		f.updateDeadUnit(u)
	}

	for u := range st.VisibleUnits.All() {
		f.iscriptFlingy = &u.Flingy
		f.iscriptUnit = u
		f.updateUnitMovement(u)
	}

	if f.updateTiles {
		for u := range st.MapRevealerUnits.All() {
			f.refreshUnitVision(u)
		}
	}

	for u := range st.VisibleUnits.All() {
		f.updateUnitSprite(u)
		if f.uCloaked(u) || f.uRequiresDetector(u) {
			u.CloakCounter = 0
			if u.SecondaryOrderTimer != 0 {
				u.SecondaryOrderTimer--
			} else {
				f.updateUnitDetectedFlags(u)
				u.SecondaryOrderTimer = 30
			}
		}
	}

	for i := st.VisibleUnits.Begin(); i != st.VisibleUnits.End(); {
		u := i
		i = st.VisibleUnits.Next(i)
		f.iscriptFlingy = &u.Flingy
		f.iscriptUnit = u
		f.updateUnit(u)
	}

	for i := st.HiddenUnits.Begin(); i != st.HiddenUnits.End(); {
		u := i
		i = st.HiddenUnits.Next(i)
		if f.uCloaked(u) || f.uRequiresDetector(u) {
			u.CloakCounter = 0
		}
		f.iscriptFlingy = &u.Flingy
		f.iscriptUnit = u
		f.updateHiddenUnit(u)
	}

	for i := st.CloakedUnits.Begin(); i != st.CloakedUnits.End(); {
		u := i
		i = st.CloakedUnits.Next(i)
		if u.CloakCounter == 0 {
			st.CloakedUnits.Remove(u)
			u.CloakedUnitLink = Link[Unit]{}
			f.uUnsetStatusFlag(u, UnitStatusFlagPassivelyCloaked)
			f.decloakUnit(u)
		} else {
			if !f.uBurrowed(u) && u.SecondaryOrderType.ID == OrdersCloak && u.CloakCounter == 1 && f.uPassivelyCloaked(u) {
				f.uUnsetStatusFlag(u, UnitStatusFlagPassivelyCloaked)
			}
			if !f.uRequiresDetector(u) {
				f.cloakUnit(u)
			}
		}
	}

	if st.UpdatePsionicMatrix {
		f.updatePsionicMatrix()
	}

	st.RecentLurkerHitCurrentIndex = (st.RecentLurkerHitCurrentIndex + 1) % uint(len(st.RecentLurkerHits))
	st.RecentLurkerHits[st.RecentLurkerHitCurrentIndex].Clear()

	for i := st.MapRevealerUnits.Begin(); i != st.MapRevealerUnits.End(); {
		u := i
		i = st.MapRevealerUnits.Next(i)
		f.iscriptFlingy = &u.Flingy
		f.iscriptUnit = u
		f.updateUnit(u)
	}

	f.iscriptFlingy = nil
	f.iscriptUnit = nil
}

func (f *StateFunctions) updateBullets() {
	st := f.st
	for i := st.ActiveBullets.Begin(); i != st.ActiveBullets.End(); {
		b := i
		i = st.ActiveBullets.Next(i)
		f.iscriptBullet = b
		f.iscriptFlingy = &b.Flingy
		if b.Sprite != nil {
			if !f.usHidden(&b.Thingy) {
				f.setSpriteVisibility(b.Sprite, int32(f.tileVisibility(b.Sprite.Position)))
			}
			if !f.iscriptExecuteSprite(b.Sprite) {
				b.Sprite = nil
			}
		}
		if b.Sprite == nil && b.BulletState != BulletStateDying {
			throwError("non-dying bullet has null sprite")
		}
		f.bulletExecute(b)
	}
	f.iscriptBullet = nil
	f.iscriptFlingy = nil
}

func (f *StateFunctions) spriteIsDoodad(spriteType *SpriteTypeT) bool {
	if spriteType.ID <= SpriteTypesSPRITEID1_6Badlands11 {
		return true
	}
	if spriteType.ID >= SpriteTypesSPRITEID7_13Twilight && spriteType.ID <= SpriteTypesSPRITEID11_3Desert {
		return true
	}
	return false
}

func (f *StateFunctions) updateThingy(t *Thingy) {
	if f.spriteIsDoodad(t.Sprite.SpriteType) {
		f.setSpriteVisibility(t.Sprite, ^int32(0))
	} else if !f.usHidden(t) {
		f.setSpriteVisibility(t.Sprite, int32(f.tileVisibility(t.Sprite.Position)))
	}
	if !f.iscriptExecuteSprite(t.Sprite) {
		t.Sprite = nil
		f.st.ActiveThingiesSize--
		f.st.ActiveThingies.Remove(t)
		bwInsertList(&f.st.FreeThingies, t)
	}
}

func (f *StateFunctions) updateThingies() {
	st := f.st
	for i := st.ActiveThingies.Begin(); i != st.ActiveThingies.End(); {
		t := i
		i = st.ActiveThingies.Next(i)
		f.updateThingy(t)
	}
}

func (f *StateFunctions) activePlayers() iter.Seq[int32] {
	return func(yield func(int32) bool) {
		for _, n := range allPlayerSlots {
			if !f.playerSlotActive(n) {
				continue
			}
			if !yield(n) {
				return
			}
		}
	}
}

func (f *StateFunctions) playerLeft(owner int32) {
	p := &f.st.Players[owner]
	f.st.RunningTriggers[owner] = nil
	if p.VictoryState != 0 {
		return
	}
	p.VictoryState = 2
	f.onVictoryState(owner, 2)
	f.removePlayer(owner)
}

func (f *StateFunctions) playerDropped(owner int32) {
	p := &f.st.Players[owner]
	f.st.RunningTriggers[owner] = nil
	if p.VictoryState != 0 {
		return
	}
	p.VictoryState = 1
	f.onVictoryState(owner, 1)
	f.removePlayer(owner)
}

func (f *StateFunctions) removePlayer(owner int32) {
	st := f.st
	p := &st.Players[owner]
	if p.Controller == PlayerControllerOccupied {
		p.Controller = PlayerControllerUserLeft
	} else if p.Controller == PlayerControllerComputer {
		p.Controller = PlayerControllerComputerDefeated
	}
	for i := st.PlayerUnits[owner].Begin(); i != st.PlayerUnits[owner].End(); {
		u := i
		i = st.PlayerUnits[owner].Next(i)
		f.makeUnitNeutral(u)
	}
}

func (f *StateFunctions) processFrame() {
	f.recedeCreep()

	st := f.st
	if st.UpdateTilesCountdown == 0 {
		st.UpdateTilesCountdown = 100
	}
	st.UpdateTilesCountdown--
	f.updateTiles = st.UpdateTilesCountdown == 0

	if f.updateTiles {
		for i := range st.Tiles {
			st.Tiles[i].Visible = 0xff
		}
	}

	f.updateUnits()
	f.updateBullets()
	f.updateThingies()
}

func (f *StateFunctions) nextFrame() {
	f.st.CurrentFrame++
	f.processFrame()
	f.processTriggers()
}

func (f *StateFunctions) localUnitStatusError(u *Unit, err int32) {
}
