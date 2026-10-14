package bwgame

func (f *StateFunctions) isSpell(weaponType *WeaponTypeT) bool {
	switch weaponType.ID {
	case WeaponTypesSpiderMines,
		WeaponTypesLockdown,
		WeaponTypesEMPShockwave,
		WeaponTypesIrradiate,
		WeaponTypesUnk50,
		WeaponTypesUnk51,
		WeaponTypesSuicideInfestedTerran,
		WeaponTypesParasite,
		WeaponTypesSpawnBroodlings,
		WeaponTypesEnsnare,
		WeaponTypesDarkSwarm,
		WeaponTypesPlague,
		WeaponTypesConsume,
		WeaponTypesUnk68,
		WeaponTypesPsiAssault,
		WeaponTypesScarab,
		WeaponTypesStasisField,
		WeaponTypesPsionicStorm,
		WeaponTypesRestoration,
		WeaponTypesMindControl,
		WeaponTypesFeedback,
		WeaponTypesOpticalFlare,
		WeaponTypesMaelstrom:
		return true
	}
	return false
}

func (f *StateFunctions) weaponDamageAmount(w *WeaponTypeT, owner int32) Fp8 {
	return Fp8Integer(w.DamageAmount + w.DamageBonus*f.playerUpgradeLevel(owner, w.DamageUpgrade.ID))
}

func (f *StateFunctions) bulletDamageAmount(b *Bullet, target *Unit) Fp8 {
	if b.WeaponType.HitType == WeaponTypeTHitTypeNuclearMissile {
		damage := f.maxVisibleHpPlusShields(target) * 2 / 3
		if damage < 500 {
			damage = 500
		}
		return Fp8Integer(damage)
	}
	return f.weaponDamageAmount(b.WeaponType, b.Owner)
}

func (f *StateFunctions) unitSpriteSize(ut *UnitTypeT) int32 {
	if f.utFlag(ut, UnitTypeTFlagSpriteSizeMedium) {
		return 1
	}
	if f.utFlag(ut, UnitTypeTFlagSpriteSizeLarge) {
		return 2
	}
	return 0
}

func (f *StateFunctions) createSizedImage(u *Unit, smallImage ImageTypes, onSubunit bool, imageOrder int32) *Image {
	image := ImageTypes(int(smallImage) + int(f.unitSpriteSize(u.UnitType)))
	spriteOwner := u
	if onSubunit && u.Subunit != nil {
		spriteOwner = u.Subunit
	}
	return f.createImage(f.getImageType(image), spriteOwner.Sprite, Xy{}, imageOrder, nil)
}

func (f *StateFunctions) createDefensiveMatrixImage(u *Unit) {
	if u.DefensiveMatrixTimer != 0 && !f.uBurrowed(u) {
		f.createSizedImage(u, ImageTypesIMAGEIDDefensiveMatrixFrontSmall, true, imageOrderTop)
		f.createSizedImage(u, ImageTypesIMAGEIDDefensiveMatrixBackSmall, false, imageOrderBelow)
	}
}

func (f *StateFunctions) dealDefensiveMatrixDamage(u *Unit, damage Fp8) {
	if damage.Less(u.DefensiveMatrixHp) {
		u.DefensiveMatrixHp = u.DefensiveMatrixHp.Sub(damage)
	} else {
		u.DefensiveMatrixHp = Fp8FromRaw(0)
		u.DefensiveMatrixTimer = 0
		f.destroyImageFromToUnit(u, ImageTypesIMAGEIDDefensiveMatrixFrontSmall, ImageTypesIMAGEIDDefensiveMatrixFrontLarge)
		f.destroyImageFromTo(u.Sprite, ImageTypesIMAGEIDDefensiveMatrixBackSmall, ImageTypesIMAGEIDDefensiveMatrixBackLarge)
	}
	if u.DefensiveMatrixTimer != 0 && !f.uBurrowed(u) {
		f.createSizedImage(u, ImageTypesIMAGEIDDefensiveMatrixHitSmall, true, imageOrderTop)
	}
}

func (f *StateFunctions) unitArmor(u *Unit) Fp8 {
	r := Fp8Integer(u.UnitType.Armor)
	if f.unitIsUltralisk(u.UnitType) && f.playerHasUpgrade(u.Owner, UpgradeTypesChitinousPlating) {
		r = r.Add(Fp8Integer(2))
	}
	r = r.Add(Fp8Integer(f.playerUpgradeLevel(u.Owner, u.UnitType.ArmorUpgrade.ID)))
	return r
}

func (f *StateFunctions) onHitShouldChangeTarget(u *Unit, oldTarget *Unit, newTarget *Unit) bool {
	if oldTarget == newTarget {
		return false
	}
	if !f.unitTargetIsEnemy(u, newTarget) {
		return false
	}
	if oldTarget == nil {
		return true
	}
	if !f.unitCanAttackTarget(u, oldTarget) {
		return true
	}
	if !f.unitTargetIsEnemy(u, oldTarget) {
		return true
	}
	if f.unitTargetInWeaponMovementRange(u, newTarget) && !f.unitTargetInWeaponMovementRange(u, oldTarget) {
		return true
	}
	return false
}

func (f *StateFunctions) unburrowUnit(u *Unit) {
	if u.OrderType.ID != OrdersBurrowing && u.OrderType.ID != OrdersBurrowed && !f.unitIs(u.UnitType, UnitTypesZergLurker) {
		return
	}
	if f.unitIs(u.UnitType, UnitTypesZergLurker) {
		f.uSetStatusFlag(u, UnitStatusFlagOrderNotInterruptible)
	}
	f.setUnitOrder(u, f.getOrderType(OrdersUnburrowing))
	f.uUnsetStatusFlag(u, UnitStatusFlagOrderNotInterruptible)
	f.orderDone(u, OrderTarget{})
}

func (f *StateFunctions) unitCanMoveFromToRegion(owner int32, from *Region, to *Region) bool {
	if from == to {
		return true
	}
	if owner >= 8 {
		return true
	}
	if from.GroupIndex != to.GroupIndex {
		return false
	}
	for _, r := range to.WalkableNeighbors {
		if r == from {
			return true
		}
		if r.GroupIndex == from.GroupIndex {
			return true
		}
	}
	for _, r := range to.NonWalkableNeighbors {
		if r == &f.gameSt.Regions.Regions[0] {
			continue
		}
		if r == to {
			return true
		}
	}
	return false
}

func (f *StateFunctions) setAutoTargetWithTurret(u *Unit, target *Unit) {
	u.AutoTargetUnit = target
	turret := f.unitTurret(u)
	if turret != nil {
		turret.LastAttackingPlayer = u.LastAttackingPlayer
		turret.AutoTargetUnit = target
	}
}

func (f *StateFunctions) onHitChangeTarget(u *Unit, target *Unit) {
	if f.unitIs(u.UnitType, UnitTypesZergLarva) {
		return
	}
	if f.uBurrowed(u) && !f.unitIs(u.UnitType, UnitTypesZergLurker) && u.IrradiateTimer == 0 {
		if f.utCanBurrow(u.UnitType) {
			f.unburrowUnit(u)
		}
		return
	}
	if f.utBuilding(u.UnitType) || (!f.utWorker(u.UnitType) && f.unitCanAttackTarget(u, target)) {
		f.setAutoTargetWithTurret(u, target)
		return
	}
	if f.uBurrowed(u) && f.unitIs(u.UnitType, UnitTypesZergLurker) {
		return
	}
	u = f.unitMainUnit(u)
	if u.OrderType.Unk11 != 0 && f.uCanMove(u) {
		if f.unitDodgeChance(u).Less(Fp8FromRaw(255)) {
			f.fleeFromAttacker(u, target)
		}
	}
}

func (f *StateFunctions) fleeFromAttacker(u *Unit, target *Unit) {
	fleeDir := f.directionXYOne(f.xyDirection(XySub(u.Sprite.Position, target.Sprite.Position)))
	fleePos := u.Sprite.Position
	lengths := [5]int32{192, 128, 64, 32, 16}
	r := f.getRegionAt(u.Sprite.Position)
	for _, v := range lengths {
		pos := f.restrictPosToMapBounds(XyAdd(u.Sprite.Position, f.toXy(XyFp8MulInt(fleeDir, v))))
		if ^u.PathingFlags&1 != 0 || f.unitCanMoveFromToRegion(u.Owner, r, f.getRegionAt(pos)) {
			fleePos = pos
			break
		}
	}
	f.setUnitOrderPos(u, f.getOrderType(OrdersMove), fleePos)
}

func (f *StateFunctions) onHitAoeChangeTarget(u *Unit, target *Unit) {
	if f.unitIsArbiter(u.UnitType) {
		return
	}
	if f.utWorker(u.UnitType) {
		return
	}
	if f.uBurrowed(u) {
		return
	}
	if !f.unitCanAttack(u) {
		return
	}
	if f.onHitShouldChangeTarget(u, u.AutoTargetUnit, target) {
		f.setAutoTargetWithTurret(u, target)
	}
}

func (f *StateFunctions) onUnitDamage(u *Unit, sourceUnit *Unit, revealSource bool) {
	if sourceUnit == nil || sourceUnit.Owner == u.Owner {
		return
	}
	u.LastAttackingPlayer = sourceUnit.Owner
	if revealSource && !f.unitTargetIsUndetected(u, sourceUnit) && u.Owner < 8 {
		f.revealSightAt(sourceUnit.Sprite.Position, 1, 1<<u.Owner, f.uFlying(sourceUnit))
	}
	if f.uInBunker(sourceUnit) {
		sourceUnit = sourceUnit.ConnectedUnit
	}
	if f.onHitShouldChangeTarget(u, u.AutoTargetUnit, sourceUnit) {
		f.onHitChangeTarget(u, sourceUnit)
	}
	bb := Rect{From: XySub(u.Sprite.Position, Xy{X: 96, Y: 96}), To: XyAdd(u.Sprite.Position, Xy{X: 96, Y: 96})}
	for n := range f.findUnitsNoexpand(bb) {
		if n != u && n.Owner == u.Owner {
			f.onHitAoeChangeTarget(n, sourceUnit)
		}
	}
}

func (f *StateFunctions) unitDealDamage(u *Unit, damage Fp8, sourceUnit *Unit, sourceOwner int32, revealSource bool) {
	if u.Hp.RawValue == 0 {
		return
	}
	f.onUnitDamage(u, sourceUnit, revealSource)
	if damage.Less(u.Hp) {
		u.Hp = u.Hp.Sub(damage)
		u.AirStrength = f.getUnitStrength(u, false)
		u.GroundStrength = f.getUnitStrength(u, true)
		if f.uCompleted(u) {
			f.updateUnitDamageOverlay(u)
		}
		return
	}
	if f.unitProvidesSpace(u) && !f.utBuilding(u.UnitType) {
		for n := range f.loadedUnits(u) {
			f.killUnit(n)
		}
	}
	u.Hp = Fp8FromRaw(0)
	f.killUnit(u)
}

func (f *StateFunctions) createShieldDamageEffect(target *Unit, heading DirectionT) {
	index := (f.directionIndex(heading) - 124) / 8 % 32
	offset := f.getImageLoOffset(target.Sprite.MainImage, 5, index, true)
	image := f.createImage(f.getImageType(ImageTypesIMAGEIDShieldOverlay), target.Sprite, offset, imageOrderAbove, nil)
	if image != nil {
		f.setImageHeadingByIndex(image, index)
	}
}

func (f *StateFunctions) weaponDealDamage(weapon *WeaponTypeT, damage Fp8, damageDivisor int32, target *Unit, heading DirectionT, sourceUnit *Unit, sourceOwner int32) {
	if target.Hp.RawValue == 0 {
		return
	}
	if f.uInvincible(target) {
		return
	}
	if f.uHallucination(target) {
		damage = damage.MulInt(2)
	}
	damage = damage.DivInt(damageDivisor)
	damage = damage.Add(Fp8Integer(target.AcidSporeCount))
	if damage.Less(Fp8FromRaw(128)) {
		damage = Fp8FromRaw(128)
	}
	if target.DefensiveMatrixHp.RawValue != 0 {
		defensiveMatrixDamage := damage
		if damage.Greater(target.DefensiveMatrixHp) {
			defensiveMatrixDamage = target.DefensiveMatrixHp
		}
		f.dealDefensiveMatrixDamage(target, defensiveMatrixDamage)
		damage = damage.Sub(defensiveMatrixDamage)
	}
	shieldDamage := Fp8FromRaw(0)
	if target.UnitType.HasShield {
		if target.ShieldPoints.GreaterEqual(Fp8FromRaw(256)) {
			if weapon.DamageType != WeaponTypeTDamageTypeIgnoreArmor {
				shieldArmor := Fp8Integer(f.playerUpgradeLevel(target.Owner, UpgradeTypesProtossPlasmaShields))
				if damage.Greater(shieldArmor) {
					damage = damage.Sub(shieldArmor)
				} else {
					damage = Fp8FromRaw(128)
				}
			}
			shieldDamage = damage
			if shieldDamage.Greater(target.ShieldPoints) {
				shieldDamage = target.ShieldPoints
			}
			damage = damage.Sub(shieldDamage)
		}
	}
	if weapon.DamageType != WeaponTypeTDamageTypeIgnoreArmor {
		armor := f.unitArmor(target)
		if damage.Greater(armor) {
			damage = damage.Sub(armor)
		} else {
			damage = Fp8FromRaw(0)
		}
	}
	damage = weaponDamageTypeModifier(weapon.DamageType, target.UnitType.UnitSize, damage)
	if shieldDamage.RawValue == 0 && damage.Less(Fp8FromRaw(128)) {
		damage = Fp8FromRaw(128)
	}
	f.unitDealDamage(target, damage, sourceUnit, sourceOwner, weapon.ID != WeaponTypesIrradiate)
	if shieldDamage.RawValue != 0 {
		target.ShieldPoints = target.ShieldPoints.Sub(shieldDamage)
		if weapon.DamageType != WeaponTypeTDamageTypeNone && target.ShieldPoints.RawValue != 0 {
			f.createShieldDamageEffect(target, heading)
		}
	}
	target.AirStrength = f.getUnitStrength(target, false)
	target.GroundStrength = f.getUnitStrength(target, true)
}

func weaponDamageTypeModifier(damageType int32, unitSize int32, damage Fp8) Fp8 {
	switch damageType {
	case WeaponTypeTDamageTypeNone:
		return Fp8FromRaw(0)
	case WeaponTypeTDamageTypeExplosive:
		switch unitSize {
		case 0:
			return Fp8FromRaw(0)
		case 1:
			return damage.Mul(Fp8FromRaw(128))
		case 2:
			return damage.Mul(Fp8FromRaw(192))
		}
	case WeaponTypeTDamageTypeConcussive:
		switch unitSize {
		case 0:
			return Fp8FromRaw(0)
		case 2:
			return damage.Mul(Fp8FromRaw(128))
		case 3:
			return damage.Mul(Fp8FromRaw(64))
		}
	}
	return damage
}

func (f *StateFunctions) hallucinatedWeaponHit(weapon *WeaponTypeT, target *Unit, heading DirectionT, sourceUnit *Unit) {
	if target.Hp.RawValue == 0 || f.uInvincible(target) {
		return
	}
	if sourceUnit == nil {
		return
	}
	f.onUnitDamage(target, sourceUnit, weapon.ID != WeaponTypesIrradiate)
	if weapon.DamageType != WeaponTypeTDamageTypeNone {
		if target.UnitType.HasShield && target.ShieldPoints.GreaterEqual(Fp8Integer(1)) {
			f.createShieldDamageEffect(target, heading)
		}
	}
}

func (f *StateFunctions) bulletDealDamageDivisor(b *Bullet, target *Unit, damageDivisor int32) {
	if b.HitFlags&2 != 0 {
		f.hallucinatedWeaponHit(b.WeaponType, target, b.Heading, b.BulletOwnerUnit)
	} else {
		f.weaponDealDamage(b.WeaponType, f.bulletDamageAmount(b, target), damageDivisor, target, b.Heading, b.BulletOwnerUnit, b.Owner)
	}
}

func (f *StateFunctions) bulletDealDamage(b *Bullet, target *Unit) {
	f.bulletDealDamageDivisor(b, target, 1)
}

func (f *StateFunctions) meleeDealDamage(u *Unit) {
	target := u.OrderTarget.Unit
	if target == nil {
		return
	}
	w := f.unitGroundWeapon(u)
	if f.uHallucination(u) {
		f.hallucinatedWeaponHit(w, target, u.Heading, u)
	} else {
		f.weaponDealDamage(w, f.weaponDamageAmount(w, u.Owner), 1, target, u.Heading, u, u.Owner)
	}
}

func (f *StateFunctions) weaponCanTargetUnit(weapon *WeaponTypeT, target *Unit) bool {
	if target == nil {
		return weapon.TargetFlags&0x40 != 0
	}
	if f.uInvincible(target) {
		return false
	}
	movementFlag := int32(2)
	if f.uFlying(target) {
		movementFlag = 1
	}
	if ^weapon.TargetFlags&movementFlag != 0 {
		return false
	}
	if weapon.TargetFlags&4 != 0 && !f.utMechanical(target.UnitType) {
		return false
	}
	if weapon.TargetFlags&8 != 0 && !f.utOrganic(target.UnitType) {
		return false
	}
	if weapon.TargetFlags&0x10 != 0 && f.utBuilding(target.UnitType) {
		return false
	}
	if weapon.TargetFlags&0x20 != 0 && f.utRobotic(target.UnitType) {
		return false
	}
	if weapon.TargetFlags&0x80 != 0 && !f.utMechanical(target.UnitType) && !f.utOrganic(target.UnitType) {
		return false
	}
	return true
}

func (f *StateFunctions) weaponCanTargetUnitFrom(weapon *WeaponTypeT, target *Unit, sourceUnit *Unit) bool {
	if !f.weaponCanTargetUnit(weapon, target) {
		return false
	}
	if target != nil && weapon.TargetFlags&0x100 != 0 && target.Owner != sourceUnit.Owner {
		return false
	}
	return true
}

func (f *StateFunctions) bulletDealSplashDamage(b *Bullet, isAirSplash bool) {
	bb := f.squareAt(b.Sprite.Position, b.WeaponType.OuterSplashRadius)
	if isAirSplash {
		f.dealAirSplashDamageToOneTarget(b, bb)
	}
	damagesAllies := !isAirSplash && b.WeaponType.HitType != WeaponTypeTHitTypeEnemySplash
	for target := range f.findUnits(bb) {
		if target == b.BulletOwnerUnit && b.WeaponType.ID != WeaponTypesPsionicStorm {
			continue
		}
		if !damagesAllies && target.Owner == b.Owner && target != b.BulletTarget {
			continue
		}
		if !f.weaponCanTargetUnit(b.WeaponType, target) {
			continue
		}
		distance := f.unitDistanceTo(target, b.Sprite.Position)
		if distance > b.WeaponType.OuterSplashRadius {
			continue
		}
		if b.WeaponType.ID == WeaponTypesPsionicStorm {
			if target.StormTimer != 0 {
				continue
			}
			target.StormTimer = 1
		}
		if isAirSplash {
			if target == b.BulletTarget {
				continue
			}
			if f.unitIs(target.UnitType, UnitTypesProtossInterceptor) {
				continue
			}
		}
		if !isAirSplash && distance <= b.WeaponType.InnerSplashRadius {
			f.bulletDealDamageDivisor(b, target, 1)
		} else if !f.uBurrowed(target) {
			if distance <= b.WeaponType.MediumSplashRadius {
				f.bulletDealDamageDivisor(b, target, 2)
			} else {
				f.bulletDealDamageDivisor(b, target, 4)
			}
		}
	}
}

func (f *StateFunctions) dealAirSplashDamageToOneTarget(b *Bullet, bb Rect) {
	var targets []*Unit
	for target := range f.findUnits(bb) {
		if target == b.BulletOwnerUnit {
			continue
		}
		if target.Owner == b.Owner && target != b.BulletTarget {
			continue
		}
		if !f.weaponCanTargetUnit(b.WeaponType, target) {
			continue
		}
		distance := f.unitDistanceTo(target, b.Sprite.Position)
		if distance > b.WeaponType.OuterSplashRadius {
			continue
		}
		if target == b.BulletTarget && distance <= b.WeaponType.InnerSplashRadius {
			targets = []*Unit{target}
			break
		}
		targets = append(targets, target)
	}
	if len(targets) == 0 {
		return
	}
	randomIndex := uint(0)
	if len(targets) != 1 {
		randomIndex = uint(f.lcgRand(58)) % uint(len(targets))
	}
	f.bulletDealDamage(b, targets[randomIndex])
}

func (f *StateFunctions) irradiateUnit(u *Unit, sourceUnit *Unit, sourceOwner int32) {
	if u.IrradiateTimer == 0 && !f.uBurrowed(u) {
		f.createSizedImage(u, ImageTypesIMAGEIDIrradiateSmall, true, imageOrderTop)
	}
	u.IrradiateTimer = f.getWeaponType(WeaponTypesIrradiate).Cooldown
	u.IrradiatedBy = sourceUnit
	u.IrradiateOwner = sourceOwner
}

func (f *StateFunctions) returnInterceptors(u *Unit) {
	if !f.unitIsCarrier(u.UnitType) {
		return
	}
	for n := range u.Carrier.OutsideUnits.All() {
		f.setUnitOrder(n, f.getOrderType(OrdersInterceptorReturn))
	}
}

func (f *StateFunctions) setUnitDisabled(u *Unit) {
	if u.Subunit != nil {
		f.uSetStatusFlag(u.Subunit, UnitStatusFlagDisabled)
	}
	f.stopUnit(u)
	if f.unitIsGhost(u.UnitType) && u.ConnectedUnit != nil && f.unitIs(u.ConnectedUnit.UnitType, UnitTypesTerranNuclearMissile) {
		u.ConnectedUnit.ConnectedUnit = nil
		u.ConnectedUnit = nil
	}
	if f.unitIsCarrier(u.UnitType) {
		f.returnInterceptors(u)
	}
	if u.OrderType.ID == OrdersRepair || u.OrderType.ID == OrdersDroneLand {
		u.OrderState = 0
	}
	if u.OrderType.Unk9 == 0 {
		u.OrderTarget.Unit = nil
	}
	f.iscriptRunToIdle(u)
}

func (f *StateFunctions) lockdownUnit(u *Unit) {
	if u.LockdownTimer == 0 {
		f.createSizedImage(u, ImageTypesIMAGEIDLockdownFieldSmall, true, imageOrderTop)
	}
	if u.LockdownTimer < 131 {
		u.LockdownTimer = 131
	}
	f.setUnitDisabled(u)
}

func (f *StateFunctions) blindUnit(u *Unit, sourceOwner int32) {
	if f.uHallucination(u) {
		f.killUnit(u)
		return
	}
	u.BlindedBy |= 1 << sourceOwner
	f.playSoundUnit(1019, u, false)
	f.createSizedImage(u, ImageTypesIMAGEIDOpticalFlareHitSmall, true, imageOrderTop)
	f.st.UpdateTilesCountdown = 1
}

func (f *StateFunctions) restoreUnit(u *Unit) {
	if f.uHallucination(u) {
		f.killUnit(u)
		return
	}
	f.createSizedImage(u, ImageTypesIMAGEIDRestorationHitSmall, true, imageOrderTop)
	u.ParasiteFlags = 0
	u.BlindedBy = 0
	if u.EnsnareTimer != 0 {
		f.removeEnsnare(u)
	}
	if u.PlagueTimer != 0 {
		f.removePlague(u)
	}
	if u.IrradiateTimer != 0 {
		f.removeIrradiate(u)
	}
	if u.LockdownTimer != 0 {
		f.removeLockdown(u)
	}
	if u.MaelstromTimer != 0 {
		f.removeMaelstrom(u)
	}
	if u.AcidSporeCount != 0 {
		f.removeAcidSpores(u)
	}
}

func (f *StateFunctions) empShockwave(position Xy, sourceUnit *Unit) {
	rangeValue := f.getWeaponType(WeaponTypesEMPShockwave).InnerSplashRadius
	for target := range f.findUnitsNoexpand(f.squareAt(position, rangeValue)) {
		if target == sourceUnit {
			continue
		}
		if sourceUnit != nil && target == sourceUnit.Subunit {
			continue
		}
		if f.uHallucination(target) {
			f.killUnit(target)
			continue
		}
		if target.StasisTimer != 0 {
			continue
		}
		target.Energy = Fp8FromRaw(0)
		target.ShieldPoints = Fp8FromRaw(0)
	}
}

func (f *StateFunctions) spawnBroodlings(target *Unit, sourceUnit *Unit) {
	if !f.uHallucination(target) {
		broodlingType := f.getUnitType(UnitTypesZergBroodling)
		spawn := func(pos Xy) {
			if !f.unitTypeCanFitAt(broodlingType, target.Sprite.Position) {
				bounds := Rect{From: XySub(pos, Xy{X: 32, Y: 32}), To: XyAdd(pos, Xy{X: 32, Y: 32})}
				if found, _ := f.findUnitPlacementInBounds(sourceUnit, pos, bounds, false); !found {
					return
				}
			}
			pos = f.restrictMoveTargetToValidBounds(broodlingType, pos)
			broodling := f.createUnit(broodlingType, pos, sourceUnit.Owner)
			if broodling != nil {
				f.finishBuildingUnit(broodling)
				f.completeUnit(broodling)
				f.setDefaultRemoveTimer(broodling)
			} else {
				f.displayLastErrorForPlayer(sourceUnit.Owner)
			}
		}
		spawn(XySub(target.Sprite.Position, Xy{X: 2, Y: 2}))
		spawn(XyAdd(target.Sprite.Position, Xy{X: 2, Y: 2}))
	}
	f.killUnit(target)
}

func (f *StateFunctions) ensnareUnit(u *Unit) {
	if u.EnsnareTimer == 0 && !f.uBurrowed(u) {
		f.createSizedImage(u, ImageTypesIMAGEIDEnsnareOverlaySmall, true, imageOrderTop)
	}
	u.EnsnareTimer = 75
	f.updateUnitSpeed(u)
}

func (f *StateFunctions) ensnare(pos Xy, sourceUnit *Unit) {
	t := f.createThingy(f.getSpriteType(SpriteTypesSPRITEIDEnsnare), pos, 0)
	t.Sprite.ElevationLevel = 19
	if !f.usHidden(t) {
		f.setSpriteVisibility(t.Sprite, int32(f.tileVisibility(t.Sprite.Position)))
	}
	for target := range f.findUnitsNoexpand(f.squareAt(pos, 64)) {
		if target == sourceUnit {
			continue
		}
		if f.utBuilding(target.UnitType) {
			continue
		}
		if f.uInvincible(target) {
			continue
		}
		if f.uBurrowed(target) {
			continue
		}
		f.ensnareUnit(target)
	}
}

func (f *StateFunctions) consumeUnit(target *Unit, sourceUnit *Unit) {
	if target == nil || f.uInvincible(target) {
		return
	}
	if !f.unitTechTargetValid(sourceUnit, f.getTechType(TechTypesConsume), target) {
		return
	}
	f.killUnit(target)
	if !f.uHallucination(target) {
		energy := sourceUnit.Energy.Add(Fp8Integer(50))
		maxEnergy := f.unitMaxEnergy(sourceUnit)
		if maxEnergy.Less(energy) {
			energy = maxEnergy
		}
		sourceUnit.Energy = energy
	}
}

func (f *StateFunctions) createSpellUnit(unitTypeID UnitTypes, pos Xy, owner int32, removeTimer int32) {
	unitType := f.getUnitType(unitTypeID)
	pos = f.restrictMoveTargetToValidBounds(unitType, pos)
	u := f.createUnit(unitType, pos, 11)
	if u == nil {
		f.displayLastErrorForPlayer(owner)
		return
	}
	f.uSetStatusFlag(u, UnitStatusFlagNoCollide)
	u.Sprite.ElevationLevel = 11
	f.finishBuildingUnit(u)
	f.completeUnit(u)
	f.setRemoveTimer(u, removeTimer)
}

func (f *StateFunctions) darkSwarm(pos Xy, owner int32) {
	f.createSpellUnit(UnitTypesSpellDarkSwarm, pos, owner, 900)
}

func (f *StateFunctions) plagueUnit(u *Unit) {
	if u.PlagueTimer == 0 && !f.uBurrowed(u) {
		f.createSizedImage(u, ImageTypesIMAGEIDPlagueOverlaySmall, true, imageOrderTop)
	}
	u.PlagueTimer = 75
}

func (f *StateFunctions) plague(pos Xy, sourceUnit *Unit) {
	for target := range f.findUnitsNoexpand(f.squareAt(pos, 64)) {
		if target == sourceUnit {
			continue
		}
		if f.uInvincible(target) {
			continue
		}
		if f.uBurrowed(target) {
			continue
		}
		f.plagueUnit(target)
		if sourceUnit != nil {
			f.onUnitDamage(target, sourceUnit, true)
		}
	}
}

func (f *StateFunctions) parasiteUnit(u *Unit, owner int32) {
	u.ParasiteFlags |= 1 << owner
}

func (f *StateFunctions) stasisUnit(u *Unit) {
	if u.StasisTimer == 0 {
		f.createSizedImage(u, ImageTypesIMAGEIDStasisFieldSmall, true, imageOrderTop)
		f.uSetStatusFlag(u, UnitStatusFlagInvincible)
	}
	if u.StasisTimer < 131 {
		u.StasisTimer = 131
	}
	f.setUnitDisabled(u)
}

func (f *StateFunctions) createSpellHitThingy(spriteTypeID SpriteTypes, pos Xy) {
	t := f.createThingy(f.getSpriteType(spriteTypeID), pos, 0)
	if t != nil {
		t.Sprite.ElevationLevel = 17
		if !f.usHidden(t) {
			f.setSpriteVisibility(t.Sprite, int32(f.tileVisibility(t.Sprite.Position)))
		}
	}
}

func (f *StateFunctions) stasisField(pos Xy, sourceUnit *Unit) {
	f.createSpellHitThingy(SpriteTypesSPRITEIDStasisFieldHit, pos)
	for target := range f.findUnitsNoexpand(f.squareAt(pos, 48)) {
		if target == sourceUnit {
			continue
		}
		if f.uInvincible(target) {
			continue
		}
		if f.uBurrowed(target) {
			continue
		}
		if f.utBuilding(target.UnitType) {
			continue
		}
		f.stasisUnit(target)
	}
}

func (f *StateFunctions) maelstromUnit(u *Unit) {
	if u.MaelstromTimer == 0 {
		f.createSizedImage(u, ImageTypesIMAGEIDMaelstormOverlaySmall, true, imageOrderTop)
	}
	if u.MaelstromTimer < 22 {
		u.MaelstromTimer = 22
	}
	f.setUnitDisabled(u)
}

func (f *StateFunctions) maelstrom(pos Xy, sourceUnit *Unit) {
	f.createSpellHitThingy(SpriteTypesSPRITEIDMaelstromHit, pos)
	f.playSoundUnit(1064, sourceUnit, false)
	for target := range f.findUnitsNoexpand(f.squareAt(pos, 48)) {
		if f.uHallucination(target) {
			f.killUnit(target)
			continue
		}
		if target == sourceUnit {
			continue
		}
		if f.uInvincible(target) {
			continue
		}
		if f.uBurrowed(target) {
			continue
		}
		if f.utBuilding(target.UnitType) {
			continue
		}
		if !f.utOrganic(target.UnitType) {
			continue
		}
		f.maelstromUnit(target)
	}
}

func (f *StateFunctions) disruptionWeb(pos Xy, owner int32) {
	f.createSpellUnit(UnitTypesSpellDisruptionWeb, pos, owner, 360)
}

func (f *StateFunctions) acidSporeImage(u *Unit) ImageTypes {
	n := min(uint(u.AcidSporeCount/2), uint(3))
	return ImageTypes(int(ImageTypesIMAGEIDAcidSpores1OverlaySmall) + 4*int(f.unitSpriteSize(u.UnitType)) + int(n))
}

func (f *StateFunctions) updateAcidSporeImage(u *Unit) {
	imageID := f.acidSporeImage(u)
	if f.findImageUnit(u, imageID, imageID) != nil {
		return
	}
	f.destroyImageFromToUnit(u, ImageTypesIMAGEIDAcidSpores1OverlaySmall, ImageTypesIMAGEIDAcidSpores6_9OverlayLarge)
	spriteOwner := u
	if u.Subunit != nil {
		spriteOwner = u.Subunit
	}
	f.createImage(f.getImageType(imageID), spriteOwner.Sprite, Xy{}, imageOrderTop, nil)
}

func (f *StateFunctions) addAcidSpore(u *Unit) {
	if u.AcidSporeCount < 9 {
		u.AcidSporeCount++
	}
	f.updateAcidSporeImage(u)
	u.AcidSporeTime[getBestScore(u.AcidSporeTime[:], func(v int32) int32 { return v })] = 150
}

func (f *StateFunctions) addAcidSporeAt(pos Xy, owner int32) {
	for target := range f.findUnitsNoexpand(f.squareAt(pos, 64)) {
		if target.Owner == owner {
			continue
		}
		if f.utBuilding(target.UnitType) {
			continue
		}
		if !f.uFlying(target) {
			continue
		}
		if f.uInvincible(target) {
			continue
		}
		if f.unitIsUndetected(target, owner) {
			continue
		}
		f.addAcidSpore(target)
	}
}

func (f *StateFunctions) unitTechTargetValid(u *Unit, tech *TechTypeT, target *Unit) bool {
	if target.StasisTimer != 0 {
		return false
	}
	switch tech.ID {
	case TechTypesFeedback:
		return !f.utBuilding(target.UnitType) && f.utHasEnergy(target.UnitType) && !f.uHallucination(target)
	case TechTypesMindControl:
		return f.mindControlTargetValid(u, target)
	case TechTypesHallucination:
		return !f.utBuilding(target.UnitType) && !f.unitIs(target.UnitType, UnitTypesProtossInterceptor)
	case TechTypesDefensiveMatrix, TechTypesIrradiate, TechTypesRestoration, TechTypesOpticalFlare:
		return !f.utBuilding(target.UnitType)
	case TechTypesConsume:
		if f.utBuilding(target.UnitType) {
			return false
		}
		if target.Owner != u.Owner {
			return false
		}
		if f.unitRace(target.UnitType) == RaceTerran {
			return false
		}
		return !f.unitIs(target.UnitType, UnitTypesZergLarva)
	}
	return false
}

func (f *StateFunctions) mindControlTargetValid(u *Unit, target *Unit) bool {
	if target.Owner == u.Owner {
		return false
	}
	if f.utBuilding(target.UnitType) {
		return false
	}
	if f.unitIs(target.UnitType, UnitTypesTerranVultureSpiderMine) {
		return false
	}
	if f.unitIs(target.UnitType, UnitTypesZergLarva) {
		return false
	}
	if f.unitIsEgg(target.UnitType) {
		return false
	}
	if f.unitIs(target.UnitType, UnitTypesProtossInterceptor) {
		return false
	}
	return !f.unitIs(target.UnitType, UnitTypesProtossScarab)
}

func (f *StateFunctions) spellOrderValid(u *Unit) bool {
	if u.OrderType.TargetsEnemies {
		if f.iscriptUnit.OrderType.Weapon == WeaponTypesNone {
			return false
		}
		return f.weaponCanTargetUnitFrom(f.getWeaponType(f.iscriptUnit.OrderType.Weapon), u.OrderTarget.Unit, u)
	}
	switch u.OrderType.TechType {
	case TechTypesNone:
		return true
	case TechTypesSpiderMines:
		return f.st.Tiles[f.tileIndex(u.OrderTarget.Pos)].Flags&(TileFlagWalkable|TileFlagHasCreep) != 0
	case TechTypesEMPShockwave,
		TechTypesScannerSweep,
		TechTypesDarkSwarm,
		TechTypesPlague,
		TechTypesEnsnare,
		TechTypesPsionicStorm,
		TechTypesRecall,
		TechTypesStasisField,
		TechTypesDisruptionWeb,
		TechTypesMaelstrom:
		return true
	}
	if u.OrderTarget.Unit == nil || f.uInvincible(u.OrderTarget.Unit) {
		return false
	}
	return f.unitTechTargetValid(u, f.getTechType(u.OrderType.TechType), u.OrderTarget.Unit)
}
