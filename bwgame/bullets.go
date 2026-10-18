package bwgame

func (f *StateFunctions) bulletHit(b *Bullet) {
	switch b.WeaponType.HitType {
	case WeaponTypeTHitTypeNone:
	case WeaponTypeTHitTypeNormal:
		if b.BulletTarget != nil && ^b.HitFlags&1 != 0 {
			f.bulletDealDamageDivisor(b, b.BulletTarget, bounceDamageDivisor(b))
		}
	case WeaponTypeTHitTypeRadialSplash, WeaponTypeTHitTypeEnemySplash, WeaponTypeTHitTypeNuclearMissile:
		if b.WeaponType.ID == WeaponTypesSubterraneanSpines {
			f.subterraneanSpinesHit(b)
		} else {
			f.bulletDealSplashDamage(b, false)
		}
	case WeaponTypeTHitTypeLockdown:
		if b.BulletTarget != nil && !f.unitDying(b.BulletTarget) {
			f.lockdownUnit(b.BulletTarget)
		}
	case WeaponTypeTHitTypeParasite:
		if b.BulletTarget != nil && !f.unitDying(b.BulletTarget) {
			f.playSoundUnit(921, b.BulletTarget, false)
			f.parasiteUnit(b.BulletTarget, b.Owner)
		}
	case WeaponTypeTHitTypeBroodlings:
		if b.BulletTarget != nil && b.BulletOwnerUnit != nil && !f.unitDying(b.BulletTarget) {
			f.spawnBroodlings(b.BulletTarget, b.BulletOwnerUnit)
		}
	case WeaponTypeTHitTypeEmpShockwave:
		f.empShockwave(b.Sprite.Position, b.BulletOwnerUnit)
	case WeaponTypeTHitTypeIrradiate:
		if b.BulletTarget != nil && !f.unitDying(b.BulletTarget) {
			f.playSoundUnit(351, b.BulletTarget, false)
			f.irradiateUnit(b.BulletTarget, b.BulletOwnerUnit, b.Owner)
		}
	case WeaponTypeTHitTypeEnsnare:
		f.ensnare(b.Sprite.Position, b.BulletOwnerUnit)
	case WeaponTypeTHitTypePlague:
		f.plague(b.BulletTargetPos, b.BulletOwnerUnit)
	case WeaponTypeTHitTypeStasisField:
		if b.BulletTargetPos != (Xy{}) {
			f.stasisField(b.BulletTargetPos, b.BulletOwnerUnit)
		}
	case WeaponTypeTHitTypeDarkSwarm:
		f.darkSwarm(b.BulletTargetPos, b.Owner)
	case WeaponTypeTHitTypeYamatoGun:
		if b.BulletTarget != nil {
			f.bulletDealDamage(b, b.BulletTarget)
		}
	case WeaponTypeTHitTypeRestoration:
		if b.BulletTarget != nil {
			f.playSoundUnit(998, b.BulletTarget, false)
			f.restoreUnit(b.BulletTarget)
		}
	case WeaponTypeTHitTypeOpticalFlare:
		if b.BulletTarget != nil && !f.unitDying(b.BulletTarget) {
			f.blindUnit(b.BulletTarget, b.Owner)
		}
	case WeaponTypeTHitTypeAirSplash:
		f.bulletDealSplashDamage(b, true)
	case WeaponTypeTHitTypeConsume:
		if b.BulletTarget != nil && b.BulletOwnerUnit != nil && !f.unitDying(b.BulletTarget) {
			f.consumeUnit(b.BulletTarget, b.BulletOwnerUnit)
		}
	case WeaponTypeTHitTypeDisruptionWeb:
		f.disruptionWeb(b.BulletTargetPos, b.Owner)
	case WeaponTypeTHitTypeCorrosiveAcid:
		if b.BulletTarget != nil && ^b.HitFlags&1 != 0 {
			f.bulletDealDamage(b, b.BulletTarget)
			if ^b.HitFlags&2 != 0 {
				f.addAcidSporeAt(b.Sprite.Position, b.Owner)
			}
		}
	case WeaponTypeTHitTypeMaelstrom:
		if b.BulletTargetPos != (Xy{}) {
			f.maelstrom(b.BulletTargetPos, b.BulletOwnerUnit)
		}
	default:
		throwError("unknown bullet hit type %d", b.WeaponType.HitType)
	}
}

func bounceDamageDivisor(b *Bullet) int32 {
	div := int32(1)
	if b.WeaponType.BulletType == WeaponTypeTBulletTypeBounce {
		bounces := 2 - b.RemainingBounces
		for i := int32(0); i < bounces; i++ {
			div *= 3
		}
	}
	return div
}

func (f *StateFunctions) subterraneanSpinesHit(b *Bullet) {
	for target := range f.findUnits(f.squareAt(b.Sprite.Position, b.WeaponType.OuterSplashRadius)) {
		if target == b.BulletOwnerUnit {
			continue
		}
		if target.Owner == b.Owner {
			continue
		}
		if !f.weaponCanTargetUnit(b.WeaponType, target) {
			continue
		}
		if f.unitDistanceTo(target, b.Sprite.Position) > b.WeaponType.InnerSplashRadius {
			continue
		}
		if b.BulletOwnerUnit != nil {
			if f.recentLurkerHitsContain(b.BulletOwnerUnit.Index, target.Index) {
				continue
			}
			arr := &f.st.RecentLurkerHits[f.st.RecentLurkerHitCurrentIndex]
			if arr.Size() != 16 {
				arr.PushBack(Pair[uint, uint]{First: b.BulletOwnerUnit.Index, Second: target.Index})
			}
		}
		f.bulletDealDamage(b, target)
	}
}

func (f *StateFunctions) recentLurkerHitsContain(ownerIndex uint, targetIndex uint) bool {
	hit := Pair[uint, uint]{First: ownerIndex, Second: targetIndex}
	for i := range f.st.RecentLurkerHits {
		for _, v := range f.st.RecentLurkerHits[i].Values() {
			if v == hit {
				return true
			}
		}
	}
	return false
}

func (f *StateFunctions) bulletKill(b *Bullet) {
	b.BulletState = BulletStateDying
	f.spriteRunAnim(b.Sprite, IscriptAnimsDeath)
}

func (f *StateFunctions) bulletMove(b *Bullet, ems *ExecuteMovementStruct) {
	f.updateUnitMovementValues(&b.Flingy, ems)
	f.finishFlingyMovement(&b.Flingy, ems)
	if !f.isPosInBounds(b.Position, f.mapBounds()) {
		b.RemainingTime = 0
	}
	pos := f.restrictPosToMapBounds(b.Position)
	f.moveSprite(b.Sprite, pos)
	if b.Position != pos {
		b.Position = pos
		b.ExactPosition = f.toXyFp8(pos)
	}
}

func (f *StateFunctions) setBulletMoveTarget(b *Bullet, target Xy) {
	f.setNextTargetWaypoint(&b.Flingy, target)
	f.setFlingyMoveTarget(&b.Flingy, target, nil)
}

func (f *StateFunctions) bulletStateInit(b *Bullet, ems *ExecuteMovementStruct) bool {
	if ^b.OrderSignal&1 != 0 {
		return false
	}
	b.OrderSignal &= ^int32(1)
	switch b.WeaponType.BulletType {
	case WeaponTypeTBulletTypeFly, WeaponTypeTBulletTypeExtendToMaxRange:
		b.BulletState = BulletStateMove
		f.spriteRunAnim(b.Sprite, IscriptAnimsGndAttkInit)
	case WeaponTypeTBulletTypeAppearAtTargetUnit,
		WeaponTypeTBulletTypeAppearAtTargetPos,
		WeaponTypeTBulletTypeAppearAtSourceUnit,
		WeaponTypeTBulletTypeSelfDestruct:
		f.bulletKill(b)
	case WeaponTypeTBulletTypeFollowTarget:
		b.BulletState = BulletStateFollow
		f.spriteRunAnim(b.Sprite, IscriptAnimsGndAttkInit)
	case WeaponTypeTBulletTypePersistAtTargetPos:
		b.BulletState = BulletStateDamageOverTime
		f.spriteRunAnim(b.Sprite, IscriptAnimsSpecialState2)
	case WeaponTypeTBulletTypeBounce:
		b.BulletState = BulletStateBounce
		f.spriteRunAnim(b.Sprite, IscriptAnimsGndAttkInit)
	case WeaponTypeTBulletTypeAttackTargetPos:
		b.BulletState = BulletStateHitNearTarget
		f.spriteRunAnim(b.Sprite, IscriptAnimsGndAttkInit)
	default:
		throwError("unknown bullet type %d", b.WeaponType.BulletType)
	}
	return true
}

func (f *StateFunctions) bulletStateDying(b *Bullet, ems *ExecuteMovementStruct) bool {
	if b.Sprite != nil {
		return false
	}
	f.st.ActiveBulletsSize--
	f.st.ActiveBullets.Remove(b)
	f.st.BulletsContainer.Push(b)
	return false
}

func (f *StateFunctions) bulletStateMove(b *Bullet, ems *ExecuteMovementStruct) bool {
	f.bulletMove(b, ems)
	remainingTime := b.RemainingTime
	b.RemainingTime--
	if remainingTime == 0 || b.Position == b.MoveTarget.Pos {
		f.bulletKill(b)
	}
	return false
}

func (f *StateFunctions) bulletStateFollow(b *Bullet, ems *ExecuteMovementStruct) bool {
	target := b.BulletTarget
	if target != nil {
		if ^b.HitFlags&1 != 0 {
			f.setBulletMoveTarget(b, target.Sprite.Position)
		} else {
			b.BulletState = BulletStateMove
		}
	} else {
		b.BulletState = BulletStateMove
		b.BulletTarget = nil
	}
	return f.bulletStateMove(b, ems)
}

func (f *StateFunctions) bulletStateHitNearTarget(b *Bullet, ems *ExecuteMovementStruct) bool {
	target := b.BulletTarget
	if target != nil {
		if ^b.HitFlags&1 != 0 {
			f.setFlingyMoveTarget(&b.Flingy, f.restrictPosToMapBounds(XyAdd(target.Sprite.Position, hitNearTargetPositions[b.HitNearTargetPositionIndex])), nil)
			f.setNextTargetWaypoint(&b.Flingy, b.MoveTarget.Pos)
		} else {
			b.BulletState = BulletStateMove
		}
	} else {
		b.BulletState = BulletStateMove
		b.BulletTarget = nil
	}
	return f.bulletStateMove(b, ems)
}

func (f *StateFunctions) bulletStateBounce(b *Bullet, ems *ExecuteMovementStruct) bool {
	target := b.BulletTarget
	if target != nil && ^b.HitFlags&1 != 0 {
		f.setFlingyMoveTarget(&b.Flingy, target.Sprite.Position, nil)
		f.setNextTargetWaypoint(&b.Flingy, b.MoveTarget.Pos)
	}
	f.bulletMove(b, ems)
	if !f.unitIsAtMoveTarget(&b.Flingy) {
		return false
	}
	b.RemainingBounces--
	var newTarget *Unit
	if b.RemainingBounces != 0 {
		if b.BulletOwnerUnit != nil {
			newTarget = f.findUnitNoexpand(f.squareAt(b.Sprite.Position, 96), func(u *Unit) bool {
				if !f.unitCanAttackTarget(b.BulletOwnerUnit, u) {
					return false
				}
				if !f.unitTargetIsEnemy(b.BulletOwnerUnit, u) {
					return false
				}
				if u == target || u == b.PrevBounceUnit {
					return false
				}
				return true
			})
		}
		b.PrevBounceUnit = target
	}
	if newTarget != nil {
		f.spriteRunAnim(b.Sprite, IscriptAnimsSpecialState1)
		b.BulletTarget = newTarget
	} else {
		f.bulletKill(b)
	}
	return false
}

func (f *StateFunctions) bulletStateDamageOverTime(b *Bullet, ems *ExecuteMovementStruct) bool {
	remainingTime := b.RemainingTime
	b.RemainingTime--
	if remainingTime == 0 {
		f.bulletKill(b)
	} else if b.RemainingTime%7 == 0 {
		f.bulletHit(b)
	}
	return false
}

func (f *StateFunctions) bulletExecute(b *Bullet) {
	var ems ExecuteMovementStruct
	for f.bulletExecuteState(b, &ems) {
	}
}

func (f *StateFunctions) bulletExecuteState(b *Bullet, ems *ExecuteMovementStruct) bool {
	switch b.BulletState {
	case BulletStateInit:
		return f.bulletStateInit(b, ems)
	case BulletStateMove:
		return f.bulletStateMove(b, ems)
	case BulletStateFollow:
		return f.bulletStateFollow(b, ems)
	case BulletStateBounce:
		return f.bulletStateBounce(b, ems)
	case BulletStateDamageOverTime:
		return f.bulletStateDamageOverTime(b, ems)
	case BulletStateDying:
		return f.bulletStateDying(b, ems)
	case BulletStateHitNearTarget:
		return f.bulletStateHitNearTarget(b, ems)
	}
	throwError("unknown bullet state %d", b.BulletState)
	return false
}

func (f *StateFunctions) unitIsUnderDarkSwarm(u *Unit) bool {
	if f.utBuilding(u.UnitType) {
		return false
	}
	if f.st.CompletedUnitCounts[11][UnitTypesSpellDarkSwarm] == 0 {
		return false
	}
	return f.findUnit(f.unitSpriteInnerBoundingBox(u), func(n *Unit) bool {
		return f.unitIs(n.UnitType, UnitTypesSpellDarkSwarm)
	}) != nil
}

func (f *StateFunctions) unitDodgeChance(u *Unit) Fp8 {
	if f.uFlying(u) {
		return Fp8FromRaw(0)
	}
	if f.unitIsUnderDarkSwarm(u) {
		return Fp8FromRaw(255)
	}
	if f.st.Tiles[f.tileIndex(u.Sprite.Position)].Flags&TileFlagProvidesCover != 0 {
		return Fp8FromRaw(119)
	}
	return Fp8FromRaw(0)
}

func (f *StateFunctions) unitTargetMissChance(u *Unit, target *Unit) Fp8 {
	r := f.unitDodgeChance(target)
	if !f.uFlying(u) && !f.uFlying(target) {
		if f.getGroundHeightAt(target.Sprite.Position) > f.getGroundHeightAt(u.Sprite.Position) {
			if r.Less(Fp8FromRaw(119)) {
				r = Fp8FromRaw(119)
			}
		}
	}
	return r
}

func (f *StateFunctions) bulletMisses(bulletOwnerUnit *Unit, targetUnit *Unit) bool {
	return targetUnit != nil && bulletOwnerUnit != nil && Fp8FromRaw(f.lcgRand(1)&0xff).LessEqual(f.unitTargetMissChance(bulletOwnerUnit, targetUnit))
}

func (f *StateFunctions) initializeBullet(b *Bullet, weaponType *WeaponTypeT, sourceUnit *Unit, pos Xy, owner int32, heading DirectionT) bool {
	flingyType := weaponType.Flingy.Get()
	if flingyType == nil {
		throwError("attempt to create bullet with null flingy")
	}
	if !f.initializeFlingy(&b.Flingy, flingyType, pos, owner, heading) {
		return false
	}

	b.MovementFlags |= 8
	b.BulletState = 0
	b.BulletTarget = nil
	b.OrderSignal = 0
	b.WeaponType = weaponType
	b.RemainingTime = weaponType.Lifetime
	b.HitFlags = 0
	b.RemainingBounces = 0
	b.Owner = owner

	if weaponType.BulletHeadingOffset.RawValue != 0 {
		f.applyBulletHeadingOffset(b, weaponType, sourceUnit)
	}

	bulletOwnerUnit := sourceUnit
	if f.utTurret(sourceUnit.UnitType) {
		bulletOwnerUnit = sourceUnit.Subunit
	}
	if f.unitIs(sourceUnit.UnitType, UnitTypesProtossScarab) {
		bulletOwnerUnit = sourceUnit.Fighter.Parent
	}
	b.BulletOwnerUnit = bulletOwnerUnit

	if f.uHallucination(sourceUnit) {
		b.HitFlags |= 2
	}
	b.PrevBounceUnit = nil

	targetUnit := sourceUnit.OrderTarget.Unit
	var targetPos Xy
	if targetUnit != nil {
		b.Sprite.ElevationLevel = targetUnit.Sprite.ElevationLevel + 1
		b.BulletTarget = targetUnit
		targetPos = targetUnit.Sprite.Position
	} else {
		b.Sprite.ElevationLevel = sourceUnit.Sprite.ElevationLevel + 1
		b.BulletTarget = nil
		targetPos = sourceUnit.OrderTarget.Pos
	}
	switch weaponType.BulletType {
	case WeaponTypeTBulletTypeFly, WeaponTypeTBulletTypeFollowTarget, WeaponTypeTBulletTypeBounce:
		if weaponType.BulletType == WeaponTypeTBulletTypeBounce {
			b.RemainingBounces = 3
		}
		if f.bulletMisses(bulletOwnerUnit, targetUnit) {
			b.HitFlags |= 1
			targetPos = XySub(targetPos, f.toXy(f.directionXYInt(heading, 30)))
		}
		f.setFlingyMoveTarget(&b.Flingy, targetPos, nil)
	case WeaponTypeTBulletTypeAppearAtTargetUnit, WeaponTypeTBulletTypeAppearAtTargetPos:
		if f.bulletMisses(bulletOwnerUnit, targetUnit) {
			b.HitFlags |= 1
			missPos := XySub(b.Sprite.Position, f.toXy(f.directionXYInt(b.Heading, 30)))
			b.ExactPosition = f.toXyFp8(missPos)
			b.Position = missPos
			f.moveSprite(b.Sprite, missPos)
		}
	case WeaponTypeTBulletTypePersistAtTargetPos:
		b.ExactPosition = f.toXyFp8(targetPos)
		b.Position = targetPos
		f.moveSprite(b.Sprite, targetPos)
	case WeaponTypeTBulletTypeAppearAtSourceUnit:
	case WeaponTypeTBulletTypeSelfDestruct:
		if b.BulletOwnerUnit != nil {
			f.uSetStatusFlag(b.BulletOwnerUnit, UnitStatusFlagLifetimeExpired)
			b.BulletOwnerUnit.UserActionFlags |= 4
			f.killUnitWith(b.BulletOwnerUnit, noKiller(ReasonDetonated))
		}
	case WeaponTypeTBulletTypeAttackTargetPos:
		b.HitNearTargetPositionIndex = uint(sourceUnit.NextHitNearTargetPositionIndex)
		if sourceUnit.NextHitNearTargetPositionIndex == 13 {
			sourceUnit.NextHitNearTargetPositionIndex = 0
		} else {
			sourceUnit.NextHitNearTargetPositionIndex++
		}
		f.setFlingyMoveTarget(&b.Flingy, XyAdd(targetPos, hitNearTargetPositions[b.HitNearTargetPositionIndex]), nil)
	case WeaponTypeTBulletTypeExtendToMaxRange:
		targetPos = XyAdd(sourceUnit.Sprite.Position, f.toXy(f.directionXYInt(b.Heading, b.WeaponType.MaxRange+20)))
		f.setFlingyMoveTarget(&b.Flingy, targetPos, nil)
	default:
		throwError("unknown bullet_type %d", weaponType.BulletType)
	}
	b.BulletTargetPos = targetPos
	return true
}

func (f *StateFunctions) applyBulletHeadingOffset(b *Bullet, weaponType *WeaponTypeT, sourceUnit *Unit) {
	var clockwise bool
	if sourceUnit == f.st.PrevBulletSourceUnit {
		clockwise = !f.st.PrevBulletHeadingOffsetClockwise
	} else {
		clockwise = f.lcgRand(0)&1 != 0
	}
	headingOffset := weaponType.BulletHeadingOffset
	if !clockwise {
		headingOffset = headingOffset.Neg()
	}
	b.NextVelocityDirection = b.NextVelocityDirection.Add(headingOffset)
	b.Heading = b.NextVelocityDirection
	f.st.PrevBulletSourceUnit = sourceUnit
	f.st.PrevBulletHeadingOffsetClockwise = clockwise
}

func (f *StateFunctions) createBullet(weaponType *WeaponTypeT, sourceUnit *Unit, pos Xy, owner int32, heading DirectionT) *Bullet {
	if weaponType.ID == WeaponTypesHaloRockets && f.st.ActiveBulletsSize >= f.gameSt.Rules.HaloRocketsBulletCap() {
		return nil
	}
	if f.uCannotAttack(sourceUnit) && !f.isSpell(weaponType) {
		return nil
	}
	b := f.st.BulletsContainer.Top()
	if b == nil {
		return nil
	}
	if !f.initializeBullet(b, weaponType, sourceUnit, pos, owner, heading) {
		return nil
	}
	f.st.BulletsContainer.Pop()
	f.st.ActiveBulletsSize++
	bwInsertList(&f.st.ActiveBullets, b)
	return b
}

func (f *StateFunctions) getBulletAppearAtTargetPos(u *Unit, target *Unit) Xy {
	targetBB := f.unitSpriteInnerBoundingBox(u.OrderTarget.Unit)
	marginW := (targetBB.To.X - targetBB.From.X) / 4
	marginH := (targetBB.To.Y - targetBB.From.Y) / 4
	bb := RectAdd(targetBB, Rect{From: Xy{X: marginW, Y: marginH}, To: Xy{X: -marginW, Y: -marginH}})
	a := XyAdd(bb.From, Xy{X: (bb.To.X - bb.From.X) / 2, Y: (bb.To.Y - bb.From.Y) / 2})
	b := u.Sprite.Position
	if f.getUniqueSidedPositionsWithinBounds(&a, &b, bb) {
		return b
	}
	return a
}

func (f *StateFunctions) fireWeapon(u *Unit, weaponType *WeaponTypeT, forwardOffset int32) {
	if !weaponType.Flingy.Valid() {
		return
	}
	var pos Xy
	switch weaponType.BulletType {
	case WeaponTypeTBulletTypeAppearAtTargetUnit:
		if u.OrderTarget.Unit == nil {
			return
		}
		pos = f.getBulletAppearAtTargetPos(u, u.OrderTarget.Unit)
	case WeaponTypeTBulletTypeAppearAtTargetPos:
		pos = u.OrderTarget.Pos
	default:
		if forwardOffset == -1 {
			forwardOffset = weaponType.ForwardOffset
		}
		pos = XyAdd(u.Sprite.Position, f.toXy(f.directionXYInt(u.Heading, forwardOffset)))
		pos.Y -= weaponType.UpwardOffset
	}
	f.createBullet(weaponType, u, pos, u.Owner, u.Heading)
}
