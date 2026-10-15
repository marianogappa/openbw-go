package bwgame

import "iter"

func (f *StateFunctions) setRemoveTimer(u *Unit, timer int32) {
	if u.RemoveTimer == 0 || timer < u.RemoveTimer {
		u.RemoveTimer = timer
	}
}

func (f *StateFunctions) defaultRemoveTimer(u *Unit) int32 {
	if f.uHallucination(u) {
		return 1350
	}
	if f.unitIs(u.UnitType, UnitTypesZergBroodling) {
		return 1800
	}
	return 0
}

func (f *StateFunctions) setDefaultRemoveTimer(u *Unit) {
	f.setRemoveTimer(u, f.defaultRemoveTimer(u))
}

func (f *StateFunctions) makeUnitHallucination(u *Unit) {
	f.incrementUnitCounts(u, -1)
	if f.uCompleted(u) {
		f.addCompletedUnit(u, -1, true)
	}
	f.uSetStatusFlag(u, UnitStatusFlagHallucination)
	f.uSetStatusFlag(u, UnitStatusFlagCompleted)
	f.incrementUnitCounts(u, 1)
	f.addCompletedUnit(u, 1, true)
	if f.utTurret(u.UnitType) {
		return
	}
	f.setDefaultRemoveTimer(u)
	u.AirStrength = f.getUnitStrength(u, false)
	u.GroundStrength = f.getUnitStrength(u, true)
	f.setSecondaryOrder(u, f.getOrderType(OrdersHallucination2))
	if u.Subunit != nil {
		f.incrementUnitCounts(u.Subunit, -1)
		f.addCompletedUnit(u.Subunit, -1, true)
		f.uSetStatusFlag(u.Subunit, UnitStatusFlagHallucination)
		f.incrementUnitCounts(u.Subunit, 1)
		f.addCompletedUnit(u.Subunit, 1, true)
	}
	if f.unitIsCarrier(u.UnitType) {
		f.killInterceptors(u)
	} else if f.unitIsReaver(u.UnitType) {
		for n := range u.Reaver().InsideUnits.All() {
			f.makeUnitHallucination(n)
		}
	}
}

func (f *StateFunctions) unitCanSeeOrderTarget(u *Unit) bool {
	if u.OrderTarget.Unit != nil {
		return f.unitCanSeeTarget(u, u.OrderTarget.Unit)
	}
	return f.playerPositionIsVisible(u.Owner, u.OrderTarget.Pos)
}

func (f *StateFunctions) medicCanHealTarget(u *Unit, target *Unit) bool {
	if target == nil || target == u {
		return false
	}
	if !f.utOrganic(target.UnitType) || f.utBuilding(target.UnitType) {
		return false
	}
	if f.uFlying(target) || f.uHallucination(target) {
		return false
	}
	if target.Hp.GreaterEqual(target.UnitType.Hitpoints) {
		return false
	}
	if f.unitTargetIsEnemy(u, target) {
		return false
	}
	if u.Owner >= 8 || target.Owner >= 8 {
		return false
	}
	if f.unitIsDisabled(target) {
		return false
	}
	return true
}

func (f *StateFunctions) findMedicTarget(u *Unit) *Unit {
	if !f.unitCanUseTech(u, f.getTechType(TechTypesHealing)) {
		return nil
	}
	return f.findNearestUnitToUnit(u, f.squareAt(u.Sprite.Position, 160), func(target *Unit) bool {
		if target.IsBeingHealed {
			return false
		}
		return f.medicCanHealTarget(u, target)
	})
}

func (f *StateFunctions) medicTryHeal(u *Unit) int32 {
	if !f.unitCanUseTech(u, f.getTechType(TechTypesHealing)) {
		return 0
	}
	target := u.OrderTarget.Unit
	if target == nil || !f.medicCanHealTarget(u, target) {
		return 0
	}
	if target.IsBeingHealed {
		target = f.findMedicTarget(u)
		if target == nil {
			return 0
		}
		u.OrderTarget.Unit = target
		u.OrderTarget.Pos = target.Sprite.Position
		f.moveToTarget(u, target)
	}
	if !f.unitTargetInRange(u, target, 30) {
		return 3
	}
	healAmount := target.UnitType.Hitpoints.Sub(target.Hp)
	if healAmount.Greater(Fp8FromRaw(200)) {
		healAmount = Fp8FromRaw(200)
	}
	energyRequired := healAmount.DivInt(2)
	if u.Energy.Less(energyRequired) {
		return 2
	}
	u.Energy = u.Energy.Sub(energyRequired)
	f.setUnitHp(target, target.Hp.Add(healAmount))
	target.IsBeingHealed = true
	return 1
}

func (f *StateFunctions) orderSelfDestructing(u *Unit) {
	f.killUnitWith(u, noKiller(ReasonDetonated))
}

func (f *StateFunctions) copyStatusEffects(to *Unit, from *Unit) {
	to.RemoveTimer = from.RemoveTimer
	to.DefensiveMatrixHp = from.DefensiveMatrixHp
	to.DefensiveMatrixTimer = from.DefensiveMatrixTimer
	to.StimTimer = from.StimTimer
	to.EnsnareTimer = from.EnsnareTimer
	to.LockdownTimer = from.LockdownTimer
	to.IrradiateTimer = from.IrradiateTimer
	to.StasisTimer = from.StasisTimer
	to.PlagueTimer = from.PlagueTimer
	to.StormTimer = from.StormTimer
	to.IrradiatedBy = from.IrradiatedBy
	to.IrradiateOwner = from.IrradiateOwner
	to.ParasiteFlags = from.ParasiteFlags
	to.CycleCounter = from.CycleCounter
	to.BlindedBy = from.BlindedBy
	to.MaelstromTimer = from.MaelstromTimer
	to.AcidSporeCount = from.AcidSporeCount
	to.AcidSporeTime = from.AcidSporeTime
	f.applyUnitEffects(to)
	f.updateUnitSpeed(to)
}

func maxStatusEffectFp8(a, b Fp8) Fp8 {
	if a.Less(b) {
		return b
	}
	return a
}

func (f *StateFunctions) mergeStatusEffects(to *Unit, from *Unit) {
	to.DefensiveMatrixHp = maxStatusEffectFp8(to.DefensiveMatrixHp, from.DefensiveMatrixHp)
	to.DefensiveMatrixTimer = max(to.DefensiveMatrixTimer, from.DefensiveMatrixTimer)
	to.StimTimer = max(to.StimTimer, from.StimTimer)
	to.EnsnareTimer = max(to.EnsnareTimer, from.EnsnareTimer)
	to.LockdownTimer = max(to.LockdownTimer, from.LockdownTimer)
	to.StasisTimer = max(to.StasisTimer, from.StasisTimer)
	to.PlagueTimer = max(to.PlagueTimer, from.PlagueTimer)
	if from.IrradiateTimer > to.IrradiateTimer {
		to.IrradiateTimer = from.IrradiateTimer
		to.IrradiatedBy = from.IrradiatedBy
		to.IrradiateOwner = from.IrradiateOwner
	}
	to.ParasiteFlags |= from.ParasiteFlags
	to.BlindedBy |= from.BlindedBy
	to.MaelstromTimer = max(from.MaelstromTimer, to.MaelstromTimer)
	to.AcidSporeCount = from.AcidSporeCount
	to.AcidSporeTime = from.AcidSporeTime
}

func (f *StateFunctions) refreshUnitPosition(u *Unit) {
	turret := f.unitTurret(u)
	if turret != nil {
		turret.ExactPosition = u.ExactPosition
		turret.Position = f.toXy(turret.ExactPosition)
		f.moveSprite(turret.Sprite, turret.Position)
	}
	if f.usHidden(&u.Thingy) {
		u.Sprite.Flags &^= SpriteFlagHidden
		if turret != nil {
			turret.Sprite.Flags &^= SpriteFlagHidden
		}
		f.st.HiddenUnits.Remove(u)
		bwInsertList(&f.st.VisibleUnits, u)
	}
	f.refreshUnitVision(u)
	f.updateUnitSprite(u)
	f.unitFinderInsert(u)
	if f.uGroundedBuilding(u) {
		f.setUnitTilesOccupied(u.UnitType, u.Sprite.Position)
	}
	f.checkUnitCollision(u)
	if f.uFlying(u) {
		f.incrementRepulseField(u)
	}
	f.resetMovementState(u)
	if turret != nil {
		f.resetMovementState(turret)
	}
}

func (f *StateFunctions) unitsShareUnions(a *UnitTypeT, b *UnitTypeT) bool {
	if f.unitIs(a, UnitTypesProtossInterceptor) != f.unitIs(b, UnitTypesProtossInterceptor) {
		return false
	}
	if f.unitIs(a, UnitTypesProtossScarab) != f.unitIs(b, UnitTypesProtossScarab) {
		return false
	}
	if f.unitIsCarrier(a) != f.unitIsCarrier(b) {
		return false
	}
	if f.unitIsReaver(a) != f.unitIsReaver(b) {
		return false
	}
	if f.unitIsGhost(a) != f.unitIsGhost(b) {
		return false
	}
	if f.utResource(a) != f.utResource(b) {
		return false
	}
	if f.unitIsNydus(a) != f.unitIsNydus(b) {
		return false
	}
	if f.unitIs(a, UnitTypesProtossPylon) != f.unitIs(b, UnitTypesProtossPylon) {
		return false
	}
	if f.unitIs(a, UnitTypesTerranNuclearSilo) != f.unitIs(b, UnitTypesTerranNuclearSilo) {
		return false
	}
	if f.utPowerup(a) != f.utPowerup(b) {
		return false
	}
	if f.unitIsHatchery(a) != f.unitIsHatchery(b) {
		return false
	}
	return true
}

func (f *StateFunctions) interceptorsAttack(u *Unit) {
	if !f.unitIsCarrier(u.UnitType) {
		return
	}
	for n := range u.Carrier.OutsideUnits.All() {
		n.Sprite.ElevationLevel = u.Sprite.ElevationLevel - 1
		f.setUnitOrderUnit(n, n.UnitType.AttackUnit.Get(), u.OrderTarget.Unit)
	}
}

func (f *StateFunctions) setCarrierOrReaverOrderType(u *Unit, reaverOrder Orders, carrierOrder Orders) {
	if f.unitIsReaver(u.UnitType) {
		u.OrderType = f.getOrderType(reaverOrder)
	} else {
		u.OrderType = f.getOrderType(carrierOrder)
	}
}

func (f *StateFunctions) carrierReaverAttack(u *Unit, acquireRange int32, maxGroundDistance int32) bool {
	canMove := u.OrderType.ID != OrdersCarrierHoldPosition && u.OrderType.ID != OrdersReaverHoldPosition
	if !canMove && f.uMovementFlag(&u.Flingy, 2) {
		f.stopUnit(u)
	}
	target := u.OrderTarget.Unit
	canAttackTarget := func() bool {
		if f.unitCanSeeTarget(u, target) && f.unitTargetInRange(u, target, acquireRange+f.unitTargetMovementRange(&u.Flingy, &target.Flingy)) {
			if maxGroundDistance == 0 || f.unitLongPathDistance(u, u.Sprite.Position, u.OrderTarget.Pos) < maxGroundDistance {
				return true
			}
		}
		return false
	}
	if target != nil {
		if f.uInvincible(target) || !f.unitCanAttackTarget(u, target) {
			if u.AutoTargetUnit == target {
				u.AutoTargetUnit = nil
			}
			u.OrderTarget.Unit = nil
		} else {
			u.OrderTarget.Pos = target.Sprite.Position
			if canAttackTarget() {
				f.stopUnit(u)
			} else {
				if !canMove || f.uReadyToAttack(u) {
					if u.AutoTargetUnit == target {
						u.AutoTargetUnit = nil
					}
					u.OrderTarget.Unit = nil
					return false
				}
				f.setUnitMoveTarget(u, u.OrderTarget.Pos)
				f.setNextTargetWaypoint(&u.Flingy, u.OrderTarget.Pos)
			}
		}
	}
	if f.unitInterceptorCount(u) == 0 && (!f.unitIsReaver(u.UnitType) || u.Reaver().InsideCount+u.Reaver().OutsideCount == 0) {
		f.stopUnit(u)
		f.orderDone(u, OrderTarget{})
		return false
	}
	if u.MainOrderTimer != 0 {
		return false
	}
	target = u.OrderTarget.Unit
	if target == nil {
		if !u.OrderQueue.Empty() {
			f.activateNextOrder(u)
			return false
		}
	}
	if target == nil || f.uReadyToAttack(u) {
		f.attackUnitReacquireTarget(u)
		if f.unitIsCarrier(u.UnitType) && u.OrderTarget.Unit != target {
			f.interceptorsAttack(u)
		}
		target = u.OrderTarget.Unit
	}
	if target != nil {
		if canMove {
			f.setCarrierOrReaverOrderType(u, OrdersReaverFight, OrdersCarrierFight)
		}
		return canAttackTarget()
	}
	newTarget := f.findAcquireRandomTarget(u)
	if newTarget == nil {
		if canMove && u.OrderType.ID != OrdersCarrierIgnore2 {
			f.setCarrierOrReaverOrderType(u, OrdersReaver, OrdersCarrier)
		}
		return false
	}
	if canMove {
		f.setCarrierOrReaverOrderType(u, OrdersReaverFight, OrdersCarrierFight)
	}
	u.OrderTarget.Pos = newTarget.Sprite.Position
	f.uSetStatusFlag(u, UnitStatusFlagReadyToAttack)
	if u.Subunit != nil {
		f.uSetStatusFlag(u.Subunit, UnitStatusFlagReadyToAttack)
	}
	if maxGroundDistance != 0 && f.unitLongPathDistance(u, u.Sprite.Position, u.OrderTarget.Pos) >= maxGroundDistance {
		return false
	}
	u.OrderTarget.Unit = newTarget
	return true
}

func (f *StateFunctions) releaseFighter(u *Unit) *Unit {
	if !f.unitIsCarrier(u.UnitType) && !f.unitIsReaver(u.UnitType) {
		return nil
	}
	insideUnits := &u.Carrier.InsideUnits
	if f.unitIsReaver(u.UnitType) {
		insideUnits = &u.Reaver().InsideUnits
	}
	var r *Unit
	var bestHp Fp8
	for n := range insideUnits.All() {
		if !f.uCompleted(n) {
			continue
		}
		if r == nil || n.Hp.Greater(bestHp) {
			if r == nil {
				bestHp = n.Hp
			}
			r = n
		}
	}
	if r == nil {
		return nil
	}
	if f.unitIs(r.UnitType, UnitTypesProtossInterceptor) && r.Hp.Less(r.UnitType.Hitpoints.DivInt(2)) {
		return nil
	}
	pos := u.Position
	if f.unitIs(r.UnitType, UnitTypesProtossScarab) {
		pos = XyAdd(pos, f.toXy(f.directionXYInt(u.Heading, 25)))
		f.setUnitHeading(r, u.Heading)
	}
	insideUnits.Remove(r)
	if f.unitIsReaver(u.UnitType) {
		u.Reaver().InsideCount--
		u.Reaver().OutsideUnits.PushFront(r)
		u.Reaver().OutsideCount++
	} else {
		u.Carrier.InsideCount--
		u.Carrier.OutsideUnits.PushFront(r)
		u.Carrier.OutsideCount++
	}
	f.moveUnit(r, pos)
	f.showUnit(r)
	r.Fighter.IsOutside = true
	return r
}

func (f *StateFunctions) createHallucination(sourceUnit *Unit, owner int32, parent *Unit) *Unit {
	u := f.createUnit(sourceUnit.UnitType, sourceUnit.Sprite.Position, owner)
	if u == nil {
		f.displayLastErrorForPlayer(owner)
		return nil
	}
	f.onUnitCreated(u, SourceHallucination, parent)
	u.Hp = u.UnitType.Hitpoints
	f.makeUnitHallucination(u)
	f.setUnitOrder(u, u.UnitType.HumanAiIdle.Get())
	if f.unitIsReaver(sourceUnit.UnitType) && f.unitIsReaver(u.UnitType) {
		for sourceN := range sourceUnit.Reaver().InsideUnits.All() {
			f.addHallucinatedFighter(u, sourceN, owner)
		}
		for sourceN := range sourceUnit.Reaver().OutsideUnits.All() {
			f.addHallucinatedFighter(u, sourceN, owner)
		}
	}
	return u
}

func (f *StateFunctions) addHallucinatedFighter(u *Unit, sourceN *Unit, owner int32) {
	n := f.createHallucination(sourceN, owner, u)
	if n == nil {
		return
	}
	f.hideUnit(n, true)
	if f.unitIsFighter(n.UnitType) {
		n.Fighter.Parent = u
		n.Fighter.IsOutside = false
		u.Reaver().InsideUnits.PushFront(n)
		u.Reaver().InsideCount++
	}
}

func (f *StateFunctions) haltAtMoveTarget(u *Unit) {
	f.stopUnit(u)
	f.setNextTargetWaypoint(&u.Flingy, u.MoveTarget.Pos)
}

func (f *StateFunctions) orderStop(u *Unit) {
	f.stopUnit(u)
	f.setNextTargetWaypoint(&u.Flingy, u.MoveTarget.Pos)
	f.iscriptRunToIdle(u)
	f.orderDone(u, OrderTarget{})
}

func (f *StateFunctions) orderGuard(u *Unit) {
	u.MainOrderTimer = f.lcgRandRange(29, 0, 15)
	u.OrderType = f.getOrderType(OrdersPlayerGuard)
}

func (f *StateFunctions) orderPlayerGuard(u *Unit) {
	if f.unitAutoattack(u) {
		return
	}
	if u.MainOrderTimer != 0 {
		return
	}
	u.MainOrderTimer = 15
	if f.utTurret(u.UnitType) {
		f.setNextTargetWaypoint(&u.Flingy, u.Subunit.NextTargetWaypoint)
	}
	if f.unitTargetAcquisitionRange(u) != 0 {
		target := f.findAcquireTarget(u)
		if target != nil {
			f.autoAttackTarget(u, target)
		}
	}
}

func (f *StateFunctions) orderTurretGuard(u *Unit) {
	f.setNextTargetWaypoint(&u.Flingy, u.Subunit.NextTargetWaypoint)
	f.orderPlayerGuard(u)
}

func (f *StateFunctions) orderMove(u *Unit) {
	if u.OrderState == 0 {
		if u.OrderTarget.Unit != nil {
			f.moveToTargetReset(u, u.OrderTarget.Unit)
		} else {
			f.setUnitMoveTarget(u, u.OrderTarget.Pos)
			f.setNextTargetWaypoint(&u.Flingy, u.OrderTarget.Pos)
		}
		u.OrderState = 1
	} else {
		if f.unitIsAtMoveTarget(&u.Flingy) {
			f.orderDone(u, OrderTarget{})
		}
	}
}

func (f *StateFunctions) orderNothing(u *Unit) {
	if !u.OrderQueue.Empty() {
		f.activateNextOrder(u)
	}
}

func (f *StateFunctions) setWorkerResourceTarget(u *Unit, target *Unit) {
	f.moveToTargetReset(u, target)
	u.Worker.TargetResourcePosition = target.Sprite.Position
	u.Worker.TargetResourceUnit = target
	u.OrderTarget.Unit = target
	u.OrderTarget.Pos = target.Sprite.Position
}

func (f *StateFunctions) orderMoveToMinerals(u *Unit) {
	if u.CarryingFlags != 0 && (u.CarryingFlags&3) == 0 {
		f.haltAtMoveTarget(u)
		f.orderDone(u, OrderTarget{})
		return
	}
	f.setUnitGathering(u)

	findMineralField := func(pos Xy, rangeValue int32, requireNotBeingGathered bool) *Unit {
		area := Rect{From: XySub(pos, Xy{X: rangeValue, Y: rangeValue}), To: XyAdd(pos, Xy{X: rangeValue, Y: rangeValue})}
		groundHeight := f.getGroundHeightAt(pos)
		return f.findNearestUnit(pos, area, func(target *Unit) bool {
			if !f.unitIsMineralField(target.UnitType) {
				return false
			}
			if !f.anyNeighborTileUnoccupied(target) {
				return false
			}
			if !f.isReachable(u.Sprite.Position, target.Sprite.Position) {
				return false
			}
			if f.getGroundHeightAt(target.Sprite.Position) != groundHeight {
				return false
			}
			if !f.unitPositionIsExplored(u, target.Sprite.Position) {
				return false
			}
			if f.unitLongPathDistance(target, target.Sprite.Position, u.Sprite.Position) > 2*rangeValue {
				return false
			}
			if requireNotBeingGathered && target.Building.Resource.IsBeingGathered {
				return false
			}
			return true
		})
	}

	target := u.OrderTarget.Unit
	if target == nil || !f.unitIsMineralField(target.UnitType) {
		searchPos := u.Sprite.Position
		if u.Worker.TargetResourcePosition != (Xy{}) {
			searchPos = u.Worker.TargetResourcePosition
			u.Worker.TargetResourcePosition = Xy{}
		}
		target = findMineralField(searchPos, 32*12, false)
		if target == nil {
			f.haltAtMoveTarget(u)
			f.orderDone(u, OrderTarget{})
			return
		}
		u.OrderTarget.Pos = target.Sprite.Position
		u.OrderTarget.Unit = target
		u.OrderState = 0
	}
	if !f.utResource(target.UnitType) {
		throwError("MoveToMinerals: target is not a resource")
	}
	if u.OrderState == 0 {
		f.moveToTargetReset(u, u.OrderTarget.Unit)
		u.Worker.TargetResourcePosition = u.OrderTarget.Pos
		u.Worker.TargetResourceUnit = u.OrderTarget.Unit
		u.OrderState = 1
		return
	}
	if u.OrderState == 1 {
		if !f.unitTargetInRange(u, target, 10) && !f.unitIsAtMoveTarget(&u.Flingy) {
			return
		}
		if target.Building.Resource.IsBeingGathered {
			newTarget := findMineralField(u.Sprite.Position, 32*8, true)
			if newTarget != nil {
				f.setWorkerResourceTarget(u, newTarget)
				return
			}
		}
		u.OrderState = 3
	}
	if u.OrderState == 3 {
		if !f.unitIsAtMoveTarget(&u.Flingy) {
			return
		}
		if f.uImmovable(u) {
			target = findMineralField(u.Sprite.Position, 32*12, false)
			if target != nil {
				f.setWorkerResourceTarget(u, target)
			} else {
				f.orderDone(u, OrderTarget{})
			}
		} else if u.CarryingFlags&2 != 0 {
			f.setUnitOrder(u, f.getOrderType(OrdersReturnMinerals))
		} else {
			f.setNextTargetWaypoint(&u.Flingy, u.OrderTarget.Pos)
			orderTarget := OrderTarget{Position: target.Sprite.Position, Unit: target}
			f.queueOrderFront(u, f.getOrderType(OrdersWaitForMinerals), orderTarget)
			f.orderDone(u, OrderTarget{})
		}
	}
}

func (f *StateFunctions) orderWaitForMinerals(u *Unit) {
	if u.OrderTarget.Unit != nil && f.unitIsMineralField(u.OrderTarget.Unit.UnitType) {
		if u.OrderState == 0 {
			f.waitForResource(u, u.OrderTarget.Unit)
		}
	} else {
		f.setUnitOrder(u, f.getOrderType(OrdersMoveToMinerals))
	}
}

func (f *StateFunctions) orderMiningMinerals(u *Unit) {
	if !f.utWorker(u.UnitType) {
		throwError("order_MiningMinerals: unit is not a worker")
	}
	target := u.OrderTarget.Unit
	if target == nil || !f.unitIsMineralField(target.UnitType) {
		if u.Worker.IsGathering {
			f.gatherQueueNext(u, u.Worker.GatherTarget)
		}
		f.removeOneOrder(u, f.getOrderType(OrdersGatheringInterrupted))
		f.setUnitOrder(u, f.getOrderType(OrdersMoveToMinerals))
		return
	}
	f.setUnitGathering(u)
	if u.OrderState == 0 {
		if f.isFacingNextTargetWaypoint(u, Fp8FromRaw(0)) {
			u.OrderTarget.Pos = XyAdd(u.Sprite.Position, f.toXy(f.directionXY(u.Heading, Fp8Integer(20))))
			f.spriteRunAnim(u.Sprite, IscriptAnimsAlmostBuilt)
			u.MainOrderTimer = 75
			u.OrderState = 5
		}
	} else if u.OrderState == 5 {
		if u.MainOrderTimer == 0 {
			f.spriteRunAnim(u.Sprite, IscriptAnimsGndAttkToIdle)
			f.unitGatherResourcesFrom(u, target)
			f.gatherQueueNext(u, target)
			f.removeOneOrder(u, f.getOrderType(OrdersGatheringInterrupted))
			if u.CarryingFlags&2 != 0 {
				f.setUnitOrder(u, f.getOrderType(OrdersReturnMinerals))
			} else {
				f.setUnitOrder(u, f.getOrderType(OrdersMoveToMinerals))
			}
		}
	}
}

func isHarvestOrder(id Orders) bool {
	switch id {
	case OrdersHarvest1, OrdersMoveToGas, OrdersWaitForGas, OrdersHarvestGas, OrdersMoveToMinerals,
		OrdersWaitForMinerals, OrdersMiningMinerals, OrdersGatheringInterrupted, OrdersGatherWaitInterrupted:
		return true
	}
	return false
}

func (f *StateFunctions) orderResetHarvestCollision(u *Unit) {
	if !u.OrderQueue.Empty() && isHarvestOrder(u.OrderQueue.Front().OrderType.ID) {
		for u.OrderQueue.Back().OrderType.ID == OrdersResetHarvestCollision {
			f.removeQueuedOrder(u, u.OrderQueue.Back())
		}
		f.queueOrderBack(u, f.getOrderType(OrdersResetHarvestCollision), OrderTarget{})
		f.uUnsetStatusFlag(u, UnitStatusFlagOrderNotInterruptible)
		f.orderDone(u, OrderTarget{})
		return
	}
	f.uUnsetStatusFlag(u, UnitStatusFlagGathering)
	f.uUnsetStatusFlag(u, UnitStatusFlagNoCollide)
	f.checkUnitCollision(u)
	if u.OrderQueue.Empty() {
		f.setQueuedOrder(u, false, u.UnitType.ReturnToIdle.Get(), OrderTarget{})
	}
	f.resetMovementState(u)
	f.uUnsetStatusFlag(u, UnitStatusFlagOrderNotInterruptible)
	f.orderDone(u, OrderTarget{})
}

func (f *StateFunctions) orderReturnGas(u *Unit) {
	f.orderReturnMinerals(u)
}

func (f *StateFunctions) orderReturnMinerals(u *Unit) {
	if (u.CarryingFlags & 3) == 0 {
		f.haltAtMoveTarget(u)
		f.orderDone(u, OrderTarget{})
		return
	}
	f.setUnitGathering(u)
	target := u.OrderTarget.Unit
	if target == nil || !f.unitIsActiveResourceDepot(target) {
		target = nil
		u.OrderState = 0
	}
	if u.OrderState == 0 {
		if target == nil {
			target = f.findNearestActiveResourceDepot(u)
		}
		if target != nil {
			u.OrderTarget.Unit = target
			f.moveToTargetReset(u, target)
			u.OrderState = 1
		} else {
			u.OrderProcessTimer = 75
		}
		return
	}
	if u.OrderState != 1 || !f.unitIsAtMoveTarget(&u.Flingy) {
		return
	}
	if f.uImmovable(u) {
		u.OrderTarget.Unit = nil
		u.OrderState = 0
		return
	}
	f.returnResourcesToDepot(u, target)
}

func (f *StateFunctions) returnResourcesToDepot(u *Unit, target *Unit) {
	var nextOrderType *OrderTypeT
	if u.CarryingFlags&1 != 0 {
		nextOrderType = f.getOrderType(OrdersMoveToGas)
	} else {
		nextOrderType = f.getOrderType(OrdersMoveToMinerals)
	}
	if u.Worker.TargetResourceUnit != nil {
		f.queueOrderFront(u, nextOrderType, OrderTarget{Unit: u.Worker.TargetResourceUnit})
	} else {
		f.queueOrderFront(u, nextOrderType, OrderTarget{})
	}
	f.orderDone(u, OrderTarget{})
	f.destroyCarryingImages(u)
	if u.CarryingFlags&2 != 0 {
		f.st.CurrentMinerals[u.Owner] += u.Worker.ResourcesCarried
		f.st.TotalMineralsGathered[u.Owner] += u.Worker.ResourcesCarried
	} else if u.CarryingFlags&1 != 0 {
		f.st.CurrentGas[u.Owner] += u.Worker.ResourcesCarried
		f.st.TotalGasGathered[u.Owner] += u.Worker.ResourcesCarried
	}
	u.CarryingFlags = 0
	u.Worker.ResourcesCarried = 0

	if f.unitIsHatchery(target.UnitType) {
		f.updateLarvaSpawnSideValues(u, target)
	}
}

func (f *StateFunctions) updateLarvaSpawnSideValues(u *Unit, hatchery *Unit) {
	values := &hatchery.Building.Hatchery.LarvaSpawnSideValues
	for i := range values {
		if values[i] != 0 {
			values[i]--
		}
	}
	bb := f.unitSpriteInnerBoundingBox(hatchery)
	var val *int32
	if u.Sprite.Position.Y < bb.From.Y {
		val = &values[1]
	} else if u.Sprite.Position.Y > bb.To.Y {
		val = &values[3]
	} else if u.Sprite.Position.X < bb.From.X {
		val = &values[0]
	} else {
		val = &values[2]
	}
	if *val < 100 {
		*val += 25
	}
}

func (f *StateFunctions) orderDie(u *Unit) {
	if u.UserActionFlags&4 != 0 {
		f.hideUnit(u, true)
	}
	if u.Subunit != nil {
		mainImage := u.Subunit.Sprite.MainImage
		if mainImage != nil {
			u.Subunit.Sprite.Images.Remove(mainImage)
			u.Sprite.Images.PushFront(mainImage)
			mainImage.Sprite = u.Sprite
			f.iUnsetFlag(mainImage, ImageFlagHasDirectionalFrames)
		}
		f.destroyUnit(u.Subunit)
		u.Subunit = nil
	}
	u.OrderState = 1
	if !f.usHidden(&u.Thingy) && (!f.uHallucination(u) || f.uLifetimeExpired(u)) {
		f.spriteRunAnim(u.Sprite, IscriptAnimsDeath)
		f.destroyUnitImpl(u)
		return
	}
	if f.uHallucination(u) {
		f.playSoundUnit(619, u, false)
		t := f.createThingy(f.getSpriteType(SpriteTypesSPRITEIDHallucinationDeath1), u.Sprite.Position, 0)
		if t != nil {
			t.Sprite.ElevationLevel = u.Sprite.ElevationLevel + 1
			if !f.usHidden(t) {
				f.setSpriteVisibility(t.Sprite, int32(f.tileVisibility(t.Sprite.Position)))
			}
		}
	}
	f.hideUnit(u, true)
	f.destroyUnit(u)
}

func (f *StateFunctions) orderGatherWaitInterrupted(u *Unit) {
	if u.Worker.GatherTarget != nil {
		u.Worker.GatherTarget.Building.Resource.GatherQueue.Remove(u)
		u.Worker.GatherTarget = nil
	}
	f.uUnsetStatusFlag(u, UnitStatusFlagOrderNotInterruptible)
	f.orderDone(u, OrderTarget{})
}

func (f *StateFunctions) orderGatheringInterrupted(u *Unit) {
	if u.Worker.IsGathering {
		f.gatherQueueNext(u, u.Worker.GatherTarget)
	}
	f.uUnsetStatusFlag(u, UnitStatusFlagOrderNotInterruptible)
	f.orderDone(u, OrderTarget{})
}

func (f *StateFunctions) orderTargetPosDistance(u *Unit) int32 {
	return f.xyLengthFP8(XyFp8Sub(f.toXyFp8(u.OrderTarget.Pos), u.ExactPosition)).IntegerPart()
}

func (f *StateFunctions) orderPlaceBuilding(u *Unit) {
	if u.OrderState == 0 {
		if u.OrderTarget.Unit != nil {
			u.OrderTarget.Pos = u.OrderTarget.Unit.Sprite.Position
		}
		f.setUnitMoveTarget(u, u.OrderTarget.Pos)
		f.setNextTargetWaypoint(&u.Flingy, u.OrderTarget.Pos)
		u.OrderState = 1
		return
	}
	if u.OrderState != 1 || !f.unitIsAtMoveTarget(&u.Flingy) {
		return
	}
	if u.BuildQueue.Empty() {
		throwError("order_PlaceBuilding: empty build queue")
	}
	unitType := u.BuildQueue.Front()
	canBuild := func() bool {
		if f.uImmovable(u) {
			return false
		}
		if f.orderTargetPosDistance(u) > 128 {
			return false
		}
		if !f.playerHasSupplyAndResourcesFor(u.Owner, unitType, true) {
			return false
		}
		if !f.canPlaceBuilding(u, u.Owner, unitType, u.OrderTarget.Pos, true, false) {
			return false
		}
		return true
	}
	if !canBuild() {
		f.orderDone(u, OrderTarget{})
		return
	}
	f.startPlacedBuilding(u, unitType)
}

func (f *StateFunctions) startPlacedBuilding(u *Unit, unitType *UnitTypeT) {
	f.st.CurrentMinerals[u.Owner] -= unitType.MineralCost
	f.st.CurrentGas[u.Owner] -= unitType.GasCost
	f.uUnsetStatusFlag(u, UnitStatusFlagGroundUnit)
	u.Sprite.ElevationLevel = u.UnitType.ElevationLevel + 1
	if u.OrderQueue.Empty() {
		f.queueOrderBack(u, f.getOrderType(OrdersResetCollision), OrderTarget{})
		f.queueOrderBack(u, u.UnitType.ReturnToIdle.Get(), OrderTarget{})
	} else {
		f.queueOrderFront(u, f.getOrderType(OrdersResetCollision), OrderTarget{})
	}
	var buildUnit *Unit
	if unitType.ID == UnitTypesTerranRefinery {
		buildUnit = f.buildRefinery(u, unitType)
	} else {
		buildUnit = f.createUnit(unitType, u.OrderTarget.Pos, u.Owner)
		if buildUnit != nil {
			f.onUnitCreated(buildUnit, SourceBuilt, u)
		}
	}
	f.buildQueueClearSingle(u)
	if buildUnit != nil {
		buildUnit.ConnectedUnit = u
		u.OrderType = f.getOrderType(OrdersConstructingBuilding)
		u.OrderState = 3
		u.OrderTarget.Unit = buildUnit
		f.setUnitOrder(buildUnit, f.getOrderType(OrdersIncompleteBuilding))
	} else {
		f.displayLastErrorForPlayer(u.Owner)
		f.setUnitOrderTarget(u, u.UnitType.ReturnToIdle.Get(), OrderTarget{})
	}
}

func (f *StateFunctions) orderIncompleteBuilding(u *Unit) {
	maxHp := u.UnitType.Hitpoints
	switch u.OrderState {
	case 0:
		if u.Hp.Greater(maxHp.DivInt(5)) {
			u.OrderState++
		}
	case 1:
		f.spriteRunAnim(u.Sprite, IscriptAnimsSpecialState1)
		u.OrderState++
	case 2:
		if u.Hp.Greater(maxHp.MulInt(2).DivInt(5)) {
			u.OrderState++
		}
	case 3:
		f.spriteRunAnim(u.Sprite, IscriptAnimsSpecialState2)
		u.OrderState++
	case 4:
		if u.Hp.Greater(maxHp.MulInt(3).DivInt(5)) {
			u.OrderState++
		}
	case 5:
		f.setConstructionGraphic(u, false)
		f.spriteRunAnim(u.Sprite, IscriptAnimsAlmostBuilt)
		u.OrderState++
	case 6:
		if u.Hp.Greater(maxHp.MulInt(4).DivInt(5)) {
			f.setUnitOrder(u, f.getOrderType(OrdersNothing))
		}
	}
}

func (f *StateFunctions) orderConstructingBuilding(u *Unit) {
	target := u.OrderTarget.Unit
	done := func() {
		f.spriteRunAnim(u.Sprite, IscriptAnimsWalkingToIdle)
		if target != nil {
			target.ConnectedUnit = nil
			if f.unitCanGatherGas(u, target) {
				f.setUnitOrderUnit(u, f.getOrderType(OrdersHarvest1), target)
			} else {
				f.orderDone(u, OrderTarget{})
			}
		} else {
			f.orderDone(u, OrderTarget{})
		}
	}
	if target == nil || f.uCompleted(target) {
		done()
		return
	}

	build := u.OrderState >= 4

	switch u.OrderState {
	case 0, 2:
		f.constructingBuildingApproach(u, target)
	case 3:
		if f.unitIsAtMoveTarget(&u.Flingy) {
			u.OrderState = 4
		}
	case 4:
		f.constructingBuildingPickSpot(u, target)
	case 5:
		u.OrderState = 6
	case 6:
		if f.orderTargetPosDistance(u) <= 20 {
			f.stopUnit(u)
			f.setNextTargetWaypoint(&u.Flingy, target.Sprite.Position)
			u.OrderTarget.Pos = XyAdd(u.Sprite.Position, f.toXy(f.directionXYInt(f.xyDirection(XySub(u.NextTargetWaypoint, u.Sprite.Position)), 20)))
			u.OrderState = 7
		}
	case 7:
		if f.isAtNextTargetWaypointOrWithinAttackAngle(u) {
			u.MainOrderTimer = 30 + (f.lcgRand(10) & 63)
			f.spriteRunAnim(u.Sprite, IscriptAnimsAlmostBuilt)
			u.OrderState = 8
		}
	case 8:
		if u.MainOrderTimer == 0 {
			f.spriteRunAnim(u.Sprite, IscriptAnimsGndAttkToIdle)
			u.OrderState = 4
		}
	}

	if build {
		f.resumeBuildingUnit(target, false)
		if f.uCompleted(target) {
			done()
		}
	}
}

func (f *StateFunctions) constructingBuildingApproach(u *Unit, target *Unit) {
	if u.OrderState == 0 {
		f.moveToTargetReset(u, target)
		u.OrderState = 2
	}
	if !f.unitIsAtMoveTarget(&u.Flingy) {
		return
	}
	if f.uImmovable(u) || (target.ConnectedUnit != nil && target.ConnectedUnit != u && target.ConnectedUnit.OrderTarget.Unit == target) {
		f.orderDone(u, OrderTarget{})
		return
	}
	target.ConnectedUnit = u
	f.uUnsetStatusFlag(u, UnitStatusFlagGroundUnit)
	u.Sprite.ElevationLevel = u.UnitType.ElevationLevel + 1
	f.setQueuedOrder(u, false, f.getOrderType(OrdersResetCollision), OrderTarget{})
	f.setQueuedOrder(u, false, u.UnitType.ReturnToIdle.Get(), OrderTarget{})
	f.setUnitMoveTarget(u, target.Sprite.Position)
	f.setNextTargetWaypoint(&u.Flingy, target.Sprite.Position)
	u.OrderState = 3
}

func (f *StateFunctions) constructingBuildingPickSpot(u *Unit, target *Unit) {
	var spriteRect Rect
	spriteRect.From = XySub(target.Sprite.Position, Xy{X: int32(target.Sprite.Width / 2), Y: int32(target.Sprite.Height / 2)})
	spriteRect.To = XyAdd(spriteRect.From, Xy{X: int32(target.Sprite.Width) - 1, Y: int32(target.Sprite.Height) - 1})
	var placementRect Rect
	placementRect.From = XySub(target.Sprite.Position, XyDivScalar(target.UnitType.PlacementSize, 2))
	placementRect.To = XySub(XyAdd(placementRect.From, target.UnitType.PlacementSize), Xy{X: 1, Y: 1})
	var insideRect Rect
	insideRect.From = XyAdd(XyAdd(placementRect.From, u.UnitType.Dimensions.From), Xy{X: 1, Y: 1})
	insideRect.To = XySub(XySub(placementRect.To, u.UnitType.Dimensions.To), Xy{X: 1, Y: 1})
	if insideRect.From.X < spriteRect.From.X {
		insideRect.From.X = spriteRect.From.X
	}
	if insideRect.From.Y < spriteRect.From.Y {
		insideRect.From.Y = spriteRect.From.Y
	}
	if insideRect.To.X > spriteRect.To.X {
		insideRect.To.X = spriteRect.To.X
	}
	if insideRect.To.Y > spriteRect.To.Y {
		insideRect.To.Y = spriteRect.To.Y
	}
	var pos Xy
	pos.X = f.lcgRandRange(9, spriteRect.From.X, spriteRect.To.X)
	pos.Y = f.lcgRandRange(9, spriteRect.From.Y, spriteRect.To.Y)
	pos = f.restrictMoveTargetToValidBounds(u.UnitType, pos)
	if f.checkUnitMovementTerrainCollision(u, XySub(pos, u.Sprite.Position)) {
		pos.X = f.lcgRandRange(9, insideRect.From.X, insideRect.To.X)
		pos.Y = f.lcgRandRange(9, insideRect.From.Y, insideRect.To.Y)
		pos = f.restrictMoveTargetToValidBounds(u.UnitType, pos)
	}
	f.setUnitMoveTarget(u, pos)
	f.setNextTargetWaypoint(&u.Flingy, pos)
	u.OrderTarget.Pos = pos
	u.OrderState = 5
}

func (f *StateFunctions) orderResetCollision(u *Unit) {
	u.Sprite.ElevationLevel = u.UnitType.ElevationLevel
	f.uUnsetStatusFlag(u, UnitStatusFlagGathering)
	f.uUnsetStatusFlag(u, UnitStatusFlagNoCollide)
	f.uSetStatusFlag(u, UnitStatusFlagGroundUnit)
	f.checkUnitCollision(u)
	f.resetMovementState(u)
	if u.OrderQueue.Empty() {
		f.setQueuedOrder(u, false, u.UnitType.ReturnToIdle.Get(), OrderTarget{})
	}
	f.uUnsetStatusFlag(u, UnitStatusFlagOrderNotInterruptible)
	f.orderDone(u, OrderTarget{})
}

func (f *StateFunctions) orderFollow(u *Unit) {
	target := u.OrderTarget.Unit
	if target == nil {
		f.orderDone(u, OrderTarget{})
		return
	}
	if u.OrderState == 0 {
		f.moveToTargetReset(u, target)
		if f.unitIsNydus(target.UnitType) {
			if f.unitCanEnterNydus(u, target) {
				f.setUnitOrderUnit(u, f.getOrderType(OrdersEnterNydusCanal), target)
				return
			}
		} else if f.unitIs(target.UnitType, UnitTypesProtossShieldBattery) {
			if f.unitCanUseShieldBattery(u, target) {
				f.setUnitOrderUnit(u, f.getOrderType(OrdersRechargeShieldsUnit), target)
				return
			}
		}
		u.OrderState = 1
	}
	if u.MainOrderTimer != 0 {
		return
	}
	u.MainOrderTimer = 7
	if f.unitTargetIsEnemy(u, target) && f.unitCanAttackTarget(u, target) {
		f.setUnitOrder(u, f.getOrderType(OrdersGuard))
		return
	}
	if !f.uFlying(u) {
		if !f.isReachableUnitToUnit(u, target) {
			f.haltAtMoveTarget(u)
			return
		}
	}
	if f.unitIsRescuable(target) {
		if !f.unitTargetInRange(u, target, 1) {
			f.moveToTargetReset(u, target)
			return
		}
	}
	if !f.unitTargetInRange(u, target, 32*3) {
		f.tryFollowUnit(u, target)
		return
	}
	if !f.uGroundedBuilding(target) && u.OrderQueue.Empty() {
		f.haltAtMoveTarget(u)
		return
	}
	f.orderDone(u, OrderTarget{})
}

func (f *StateFunctions) orderAttackMove(u *Unit) {
	if !f.attackMoveMovement(u, true) {
		return
	}
	if u.MainOrderTimer != 0 {
		return
	}
	u.MainOrderTimer = 15
	if !f.unitIsAtMoveTarget(&u.Flingy) {
		f.attackMoveAcquireTarget(u, true)
	} else {
		f.orderDone(u, OrderTarget{})
	}
}

func (f *StateFunctions) orderAttackUnit(u *Unit) {
	f.attackUnitReacquireTarget(u)
	if f.attackUnitMoveInRange(u) {
		f.attackUnitFireWeapon(u)
	}
}

func (f *StateFunctions) orderMoveToGas(u *Unit) {
	if u.CarryingFlags != 0 && (u.CarryingFlags&3) == 0 {
		f.haltAtMoveTarget(u)
		f.orderDone(u, OrderTarget{})
		return
	}
	f.setUnitGathering(u)
	target := u.OrderTarget.Unit
	if u.OrderState == 0 {
		if target == nil {
			f.haltAtMoveTarget(u)
			f.orderDone(u, OrderTarget{})
		} else if f.unitIsMineralField(target.UnitType) {
			f.queueOrderFront(u, f.getOrderType(OrdersMoveToMinerals), OrderTarget{Position: u.OrderTarget.Pos, Unit: target})
			f.orderDone(u, OrderTarget{})
		} else {
			f.moveToTargetReset(u, target)
			u.Worker.TargetResourcePosition = u.OrderTarget.Pos
			u.Worker.TargetResourceUnit = target
			u.OrderState = 3
		}
		return
	}
	if u.OrderState != 3 || !f.unitIsAtMoveTarget(&u.Flingy) {
		return
	}
	if target == nil || !f.unitCanGatherGas(u, target) {
		f.haltAtMoveTarget(u)
		f.orderDone(u, OrderTarget{})
	} else if f.uImmovable(u) {
		u.OrderState = 0
	} else if u.CarryingFlags&1 != 0 {
		f.setUnitOrder(u, f.getOrderType(OrdersReturnGas))
	} else {
		f.queueOrderFront(u, f.getOrderType(OrdersWaitForGas), OrderTarget{Position: target.Sprite.Position, Unit: target})
		f.orderDone(u, OrderTarget{})
	}
}

func (f *StateFunctions) canHarvestGasFrom(u *Unit, target *Unit) bool {
	return target != nil && f.unitCanGatherGas(u, target) && (u.CarryingFlags == 0 || u.CarryingFlags&3 != 0)
}

func (f *StateFunctions) orderWaitForGas(u *Unit) {
	target := u.OrderTarget.Unit
	if !f.canHarvestGasFrom(u, target) {
		f.haltAtMoveTarget(u)
		f.orderDone(u, OrderTarget{})
		return
	}
	if u.OrderState == 0 {
		f.waitForResource(u, target)
	}
}

func (f *StateFunctions) orderHarvestGas(u *Unit) {
	if !f.utWorker(u.UnitType) {
		throwError("order_HarvestGas: unit is not a worker")
	}
	target := u.OrderTarget.Unit
	if !f.usHidden(&u.Thingy) && !f.canHarvestGasFrom(u, target) {
		f.stopUnit(u)
		f.uUnsetStatusFlag(u, UnitStatusFlagOrderNotInterruptible)
		f.orderDone(u, OrderTarget{})
		if u.Worker.GatherTarget != nil {
			f.gatherQueueNext(u, u.Worker.GatherTarget)
		}
		return
	}
	f.setUnitGathering(u)
	if u.OrderState == 0 {
		f.hideUnit(u, false)
		u.OrderState = 5
		u.MainOrderTimer = 37
	} else if u.OrderState == 5 {
		if u.MainOrderTimer == 0 {
			f.exitGas(u, target)
		}
	}
}

func (f *StateFunctions) exitGas(u *Unit, target *Unit) {
	var pos Xy
	heading := u.Heading
	if !f.findGasExitPosition(u, target, &pos, &heading) {
		u.MainOrderTimer = 18
		return
	}
	f.setUnitHeading(u, heading)
	f.moveUnit(u, pos)
	f.showUnit(u)
	if target != nil {
		f.unitGatherResourcesFrom(u, target)
	}
	if u.Worker.GatherTarget != nil {
		f.gatherQueueNext(u, u.Worker.GatherTarget)
	}
	f.removeOneOrder(u, f.getOrderType(OrdersResetHarvestCollision))
	if u.OrderQueue.Empty() {
		if u.CarryingFlags&1 != 0 {
			f.setQueuedOrder(u, false, f.getOrderType(OrdersReturnGas), OrderTarget{})
		} else {
			f.setQueuedOrder(u, false, u.UnitType.ReturnToIdle.Get(), OrderTarget{})
		}
	}
	f.queueOrderFront(u, f.getOrderType(OrdersResetHarvestCollision), OrderTarget{})
	f.uUnsetStatusFlag(u, UnitStatusFlagOrderNotInterruptible)
	f.orderDone(u, OrderTarget{})
}

type repairCosts struct {
	mineral int32
	gas     int32
	time    int32
}

func (f *StateFunctions) calcRepairCosts(target *Unit) repairCosts {
	targetMineralCost := target.UnitType.MineralCost
	targetGasCost := target.UnitType.GasCost
	lowestCost := targetMineralCost
	if targetGasCost != 0 {
		lowestCost = min(targetMineralCost, targetGasCost)
	}
	if lowestCost == 0 {
		return repairCosts{mineral: 0, gas: 0, time: 0xffff}
	}
	targetMaxHp := target.UnitType.Hitpoints
	hpConstructionRate := target.HpConstructionRate
	if f.st.CheatOperationCwal {
		hpConstructionRate = hpConstructionRate.MulInt(16)
	}
	var c repairCosts
	c.time = targetMaxHp.RawValue * 3 / (lowestCost * hpConstructionRate.RawValue)
	if c.time != 0 {
		if lowestCost == targetMineralCost {
			c.mineral = 1
			c.gas = targetGasCost / targetMineralCost
		} else {
			c.mineral = targetMineralCost / targetGasCost
			c.gas = 1
		}
		return c
	}
	c.mineral = repairCostPerStep(targetMineralCost, hpConstructionRate, targetMaxHp)
	c.gas = repairCostPerStep(targetGasCost, hpConstructionRate, targetMaxHp)
	c.time = 1
	return c
}

func repairCostPerStep(targetCost int32, hpConstructionRate Fp8, targetMaxHp Fp8) int32 {
	if targetCost == 0 {
		return 0
	}
	cost := targetCost * hpConstructionRate.RawValue / (targetMaxHp.RawValue * 3)
	if cost == 0 {
		cost = 1
	}
	return cost
}

func (f *StateFunctions) repairStep(u *Unit, target *Unit) {
	if u.Worker.RepairTimer == 0 {
		costs := f.calcRepairCosts(target)
		if f.st.CurrentMinerals[u.Owner] >= costs.mineral && f.st.CurrentGas[u.Owner] >= costs.gas {
			u.Worker.RepairTimer = costs.time
			f.st.CurrentMinerals[u.Owner] -= costs.mineral
			f.st.CurrentGas[u.Owner] -= costs.gas
		} else {
			f.spriteRunAnim(u.Sprite, IscriptAnimsWalkingToIdle)
			f.orderDone(u, OrderTarget{})
			return
		}
	} else {
		u.Worker.RepairTimer--
	}
	if f.st.CheatOperationCwal {
		f.setUnitHp(target, target.Hp.Add(target.HpConstructionRate.MulInt(16)))
	} else {
		f.setUnitHp(target, target.Hp.Add(target.HpConstructionRate))
	}
	if target.Hp.GreaterEqual(target.UnitType.Hitpoints) {
		f.spriteRunAnim(u.Sprite, IscriptAnimsWalkingToIdle)
		f.orderDone(u, OrderTarget{})
		if target != nil && !f.uLoaded(target) {
			target.ConnectedUnit = nil
		}
	}
}

func (f *StateFunctions) repairTargetInvalid(target *Unit) bool {
	return target == nil || target.Hp.GreaterEqual(target.UnitType.Hitpoints) || target.StasisTimer != 0 || f.uLoaded(target) || !f.utMechanical(target.UnitType) || !f.uCompleted(target)
}

func (f *StateFunctions) faceRepairTarget(u *Unit, target *Unit) {
	f.setNextTargetWaypoint(&u.Flingy, target.Sprite.Position)
	u.OrderTarget.Pos = target.Sprite.Position
}

func (f *StateFunctions) orderRepair(u *Unit) {
	if !f.utWorker(u.UnitType) {
		throwError("order_Repair: unit is not a worker")
	}
	target := u.OrderTarget.Unit
	if f.repairTargetInvalid(target) {
		f.spriteRunAnim(u.Sprite, IscriptAnimsWalkingToIdle)
		f.orderDone(u, OrderTarget{})
		if target != nil && !f.uLoaded(target) {
			target.ConnectedUnit = nil
		}
		return
	}
	if f.unitRace(target.UnitType) != RaceTerran {
		f.spriteRunAnim(u.Sprite, IscriptAnimsWalkingToIdle)
		f.orderDone(u, OrderTarget{})
		return
	}

	switch u.OrderState {
	case 0, 1:
		if u.OrderState == 0 {
			u.Worker.RepairTimer = 0
			u.OrderState = 1
		}
		if f.uCompleted(target) {
			f.moveToTargetReset(u, target)
			u.OrderTarget.Pos = target.Sprite.Position
			u.OrderState = 6
		} else {
			f.setUnitOrderUnit(u, f.getOrderType(OrdersConstructingBuilding), target)
		}
	case 6:
		if !f.unitIsAtMoveTarget(&u.Flingy) {
			f.tryFollowUnit(u, target)
		} else if f.unitTargetInRange(u, target, 5) {
			f.faceRepairTarget(u, target)
			u.OrderState = 7
			f.repairStep(u, target)
		} else {
			u.OrderState = 1
		}
	case 7:
		if f.unitTargetInRange(u, target, 5) {
			f.faceRepairTarget(u, target)
			if f.isAtNextTargetWaypointOrWithinAttackAngle(u) {
				f.spriteRunAnim(u.Sprite, IscriptAnimsAlmostBuilt)
				u.OrderState = 8
			}
			f.repairStep(u, target)
		} else {
			u.OrderState = 1
		}
	case 8:
		f.faceRepairTarget(u, target)
		if !f.unitTargetInRange(u, target, 5) || target.OrderType.ID == OrdersLiftingOff {
			f.spriteRunAnim(u.Sprite, IscriptAnimsGndAttkToIdle)
			u.OrderState = 1
		}
		f.repairStep(u, target)
	default:
		f.repairStep(u, target)
	}
}

func (f *StateFunctions) orderEnterTransport(u *Unit) {
	target := u.OrderTarget.Unit
	if target == nil || f.uHallucination(target) || !f.unitProvidesSpace(target) || !f.unitCanLoadTarget(target, u) {
		f.orderDone(u, OrderTarget{})
		return
	}
	if u.OrderState == 0 {
		if target.OrderType.ID != OrdersPickupTransport {
			if f.uCanMove(target) {
				f.queueOrderFront(target, f.getOrderType(OrdersPickupTransport), OrderTarget{Position: u.Sprite.Position, Unit: u})
				f.activateNextOrder(target)
			}
			u.OrderState = 1
		}
	}
	f.setNextTargetWaypoint(&u.Flingy, target.Sprite.Position)
	f.tryFollowUnit(u, target)
	if f.unitTargetInRange(u, target, 1) {
		f.unitLoadTarget(target, u)
	}
}

func firstLoadedUnitIsLast(units iter.Seq[*Unit]) (first *Unit, isLast bool) {
	count := 0
	for n := range units {
		count++
		if count == 1 {
			first = n
			continue
		}
		break
	}
	return first, count == 1
}

func (f *StateFunctions) orderUnload(u *Unit) {
	if f.uGroundedBuilding(u) {
		for n := range f.loadedUnits(u) {
			f.unitUnload(n)
		}
		f.orderDone(u, OrderTarget{})
		return
	}
	if u.MainOrderTimer != 0 {
		return
	}
	first, isLast := firstLoadedUnitIsLast(f.loadedUnits(u))
	if first == nil {
		f.orderDone(u, OrderTarget{})
		return
	}
	if !f.unitUnload(first) || isLast {
		f.orderDone(u, OrderTarget{})
	}
}

func (f *StateFunctions) freezeLastImageY(u *Unit) {
	last := u.Sprite.Images.Back()
	if last.Modifier == 10 {
		last.FrozenYValue = f.getImageMapPosition(last).Y
		last.Flags |= ImageFlagYFrozen
	}
}

func (f *StateFunctions) orderBuildingLiftoff(u *Unit) {
	if !f.uGroundedBuilding(u) || !f.utFlyingBuilding(u.UnitType) {
		f.uUnsetStatusFlag(u, UnitStatusFlagOrderNotInterruptible)
		f.orderDone(u, OrderTarget{})
		return
	}
	if u.Building.Addon != nil {
		f.buildingAbandonAddon(u)
	}
	if f.unitIsFactory(u) {
		u.Building.Rally.Unit = u
	}
	f.setUnitTilesUnoccupied(u.UnitType, u.Sprite.Position)
	f.uUnsetStatusFlag(u, UnitStatusFlagGroundedBuilding)
	f.uSetStatusFlag(u, UnitStatusFlagCanTurn)
	f.uSetStatusFlagValue(u, UnitStatusFlagCanMove, !f.utTurret(u.UnitType))
	f.uSetStatusFlag(u, UnitStatusFlagFlying)
	u.Sprite.ElevationLevel = 12
	f.resetMovementState(u)
	f.checkUnitCollision(u)
	f.unitFinderRemove(u)
	f.unitFinderInsert(u)
	f.freezeLastImageY(u)
	f.setNextTargetWaypoint(&u.Flingy, XySub(u.Sprite.Position, Xy{X: 0, Y: 42}))
	if f.unitIs(u.UnitType, UnitTypesZergInfestedCommandCenter) {
		f.removeUnitCreepProvider(u)
	}
	u.OrderType = f.getOrderType(OrdersLiftingOff)
}

func (f *StateFunctions) liftOffHeight(u *Unit) int32 {
	if f.utWorker(u.UnitType) {
		return 7
	}
	return 42
}

func (f *StateFunctions) isFacingTargetWaypointForLanding(u *Unit) bool {
	return u.Position == u.NextTargetWaypoint || u.NextVelocityDirection == f.xyDirection(XySub(u.NextTargetWaypoint, u.Sprite.Position))
}

func (f *StateFunctions) createBuildingDust(u *Unit, loIndex uint) {
	f.createDustSprite(u.Sprite, SpriteTypesSPRITEIDBuildingLandingDustType1, loIndex, 0, 8, false)
	f.createDustSprite(u.Sprite, SpriteTypesSPRITEIDBuildingLandingDustType1, loIndex, 16, 24, true)
}

func (f *StateFunctions) orderLiftingOff(u *Unit) {
	switch u.OrderState {
	case 0, 1:
		if u.OrderState == 0 {
			if !f.isFacingTargetWaypointForLanding(u) {
				return
			}
			u.OrderState = 1
		}
		pos := XyAdd(u.Sprite.Position, Xy{X: 0, Y: u.Sprite.Images.Back().Offset.Y - f.liftOffHeight(u)})
		f.setUnitMoveTarget(u, pos)
		f.setNextTargetWaypoint(&u.Flingy, pos)
		u.OrderState = 2
	case 2, 3:
		if u.OrderState == 2 {
			f.spriteRunAnim(u.Sprite, IscriptAnimsLiftOff)
			if u.Sprite.MainImage.ImageType.LiftOffFilenameIndex != 0 {
				f.createBuildingDust(u, 4)
			}
			u.FlingyTopSpeed = Fp8Integer(1)
			f.setNextSpeed(&u.Flingy, Fp8Integer(1))
			u.OrderState = 3
		}
		if f.unitIsAtMoveTarget(&u.Flingy) && u.OrderSignal&0x10 != 0 {
			f.finishLiftingOff(u)
		}
	}
}

func (f *StateFunctions) finishLiftingOff(u *Unit) {
	u.OrderSignal &^= 0x10
	if !f.utBuilding(u.UnitType) {
		u.FlingyTopSpeed = f.getModifiedUnitSpeed(u, u.FlingyType.TopSpeed)
	}
	last := u.Sprite.Images.Back()
	if last.Modifier == 10 {
		last.Offset.Y = f.liftOffHeight(u)
		last.Flags &^= ImageFlagYFrozen
	}
	f.uUnsetStatusFlag(u, UnitStatusFlagOrderNotInterruptible)
	f.orderDone(u, OrderTarget{})
}

func (f *StateFunctions) orderPlaceAddon(u *Unit) {
	addonType := u.Building.AddonBuildType
	if addonType == nil || !f.utAddon(addonType) {
		return
	}
	if u.OrderState == 0 {
		if f.uGroundedBuilding(u) {
			if f.orderTargetPosDistance(u) == 0 {
				f.setUnitMoveTarget(u, u.OrderTarget.Pos)
				f.setNextTargetWaypoint(&u.Flingy, u.OrderTarget.Pos)
				u.OrderState = 1
			} else {
				f.setUnitOrderPos(u, f.getOrderType(OrdersBuildingLiftoff), u.OrderTarget.Pos)
				f.setQueuedOrder(u, false, f.getOrderType(OrdersPlaceAddon), OrderTarget{Position: u.OrderTarget.Pos})
			}
		} else {
			f.setUnitOrderPos(u, f.getOrderType(OrdersBuildingLand), u.OrderTarget.Pos)
			f.setQueuedOrder(u, false, f.getOrderType(OrdersPlaceAddon), OrderTarget{Position: u.OrderTarget.Pos})
		}
	}
	if u.OrderState == 1 && f.unitIsAtMoveTarget(&u.Flingy) {
		f.placeAddonAtTarget(u, addonType)
	}
}

func (f *StateFunctions) placeAddonAtTarget(u *Unit, addonType *UnitTypeT) {
	u.OrderTarget.Pos = XySub(u.OrderTarget.Pos, XyDivScalar(u.UnitType.PlacementSize, 2))
	u.OrderTarget.Pos = XyAdd(u.OrderTarget.Pos, addonType.AddonPosition)
	u.OrderTarget.Pos = XyAdd(u.OrderTarget.Pos, XyDivScalar(addonType.PlacementSize, 2))
	if !f.playerHasSupplyAndResourcesFor(u.Owner, addonType, true) {
		f.orderDone(u, OrderTarget{})
		return
	}
	if !f.canPlaceBuilding(u, u.Owner, addonType, u.OrderTarget.Pos, true, false) {
		f.orderDone(u, OrderTarget{})
		return
	}
	addon := f.createUnit(addonType, u.OrderTarget.Pos, u.Owner)
	if addon == nil {
		return
	}
	f.onUnitCreated(addon, SourceBuilt, u)
	f.st.CurrentMinerals[u.Owner] -= addonType.MineralCost
	f.st.CurrentGas[u.Owner] -= addonType.GasCost
	f.setUnitOrder(u, u.UnitType.ReturnToIdle.Get())
	f.setSecondaryOrder(u, f.getOrderType(OrdersBuildAddon))
	u.CurrentBuildUnit = addon
	f.setUnitOrder(addon, f.getOrderType(OrdersIncompleteBuilding))
	addon.ConnectedUnit = u
}

func (f *StateFunctions) orderBuildingLand(u *Unit) {
	if f.uGroundedBuilding(u) {
		f.setUnitOrderPos(u, f.getOrderType(OrdersBuildingLiftoff), u.OrderTarget.Pos)
		f.setQueuedOrder(u, false, f.getOrderType(OrdersBuildingLand), OrderTarget{Position: u.OrderTarget.Pos})
		return
	}
	switch u.OrderState {
	case 0:
		pos := XySub(u.OrderTarget.Pos, Xy{X: 0, Y: u.Sprite.Images.Back().Offset.Y})
		f.setUnitMoveTarget(u, pos)
		f.setNextTargetWaypoint(&u.Flingy, pos)
		u.OrderState = 1
	case 1:
		if f.unitIsAtMoveTarget(&u.Flingy) {
			f.startLanding(u)
		}
	case 2:
		if f.isFacingTargetWaypointForLanding(u) {
			f.spriteRunAnim(u.Sprite, IscriptAnimsLanding)
			u.OrderState = 3
		}
	}
	if u.OrderState == 3 {
		if u.OrderSignal&0x10 != 0 && f.unitIsAtMoveTarget(&u.Flingy) {
			f.finishLanding(u)
		}
	}
}

func (f *StateFunctions) startLanding(u *Unit) {
	if !f.canPlaceBuilding(u, u.Owner, u.UnitType, u.OrderTarget.Pos, true, false) {
		if !u.OrderQueue.Empty() && u.OrderQueue.Front().OrderType.ID == OrdersPlaceAddon {
			f.removeQueuedOrder(u, u.OrderQueue.Front())
		}
		f.orderDone(u, OrderTarget{})
		return
	}
	f.setUnitTilesOccupied(u.UnitType, u.OrderTarget.Pos)
	u.Building.IsLanding = true
	f.setUnitMoveTarget(u, u.OrderTarget.Pos)
	f.setNextTargetWaypoint(&u.Flingy, u.OrderTarget.Pos)
	f.freezeLastImageY(u)
	f.uSetStatusFlag(u, UnitStatusFlagOrderNotInterruptible)
	u.FlingyTopSpeed = Fp8Integer(1)
	f.setNextSpeed(&u.Flingy, Fp8Integer(1))
	u.OrderState = 2
}

func (f *StateFunctions) finishLanding(u *Unit) {
	u.OrderSignal &^= 0x10
	f.unitFinderRemove(u)
	if f.uFlying(u) {
		f.decrementRepulseField(u)
	}
	f.uUnsetStatusFlag(u, UnitStatusFlagFlying)
	f.uUnsetStatusFlag(u, UnitStatusFlagCanTurn)
	f.uUnsetStatusFlag(u, UnitStatusFlagCanMove)
	f.uSetStatusFlag(u, UnitStatusFlagGroundedBuilding)
	u.Sprite.ElevationLevel = 4
	f.resetMovementState(u)
	f.moveUnit(u, u.OrderTarget.Pos)
	f.unitFinderInsert(u)
	f.checkUnitCollision(u)
	last := u.Sprite.Images.Back()
	last.Offset.Y = 0
	if last.Modifier == 10 {
		last.Flags &^= ImageFlagYFrozen
	}
	if u.Sprite.MainImage.ImageType.LandingDustFilenameIndex != 0 {
		f.createBuildingDust(u, 3)
	}
	u.Building.IsLanding = false
	for n := range f.findUnitsNoexpand(f.unitSpriteInnerBoundingBox(u)) {
		if n == u {
			continue
		}
		if f.unitTargetIsEnemy(n, u) {
			continue
		}
		if (!f.unitCanMove(n) && f.unitCanAttack(n)) || f.uBurrowed(n) {
			f.killUnit(n)
		}
	}
	if !u.OrderQueue.Empty() {
		o := u.OrderQueue.Front()
		if o.OrderType.ID == OrdersMove || o.OrderType.ID == OrdersFollow {
			f.queueOrderFront(u, f.getOrderType(OrdersBuildingLiftoff), OrderTarget{})
		} else {
			for !u.OrderQueue.Empty() && u.OrderQueue.Back().OrderType.ID != OrdersPlaceAddon {
				f.removeQueuedOrder(u, u.OrderQueue.Back())
			}
			if u.OrderQueue.Empty() {
				f.setUnitOrder(u, u.UnitType.ReturnToIdle.Get())
			}
		}
	}
	for n := range f.findUnitsNoexpand(f.unitSpriteInnerBoundingBox(u)) {
		if n == u {
			continue
		}
		if f.uGroundedBuilding(n) || !f.unitCanMove(n) {
			if f.unitCanAttack(n) && n.OrderType.ID != OrdersDie {
				f.setUnitOrder(u, f.getOrderType(OrdersBuildingLiftoff))
				f.setQueuedOrder(u, false, u.UnitType.ReturnToIdle.Get(), OrderTarget{})
				break
			}
		}
	}
	f.uUnsetStatusFlag(u, UnitStatusFlagOrderNotInterruptible)
	f.orderDone(u, OrderTarget{})
	f.findAndConnectAddon(u)
}

func decrementResearchTime(v *int32) int32 {
	old := *v
	*v--
	return old
}

func (f *StateFunctions) orderResearchTech(u *Unit) {
	tech := u.Building.ResearchingType
	done := func() {
		u.Building.ResearchingType = nil
		if tech != nil {
			f.st.TechResearching[u.Owner][tech.ID] = false
		}
		f.spriteRunAnim(u.Sprite, IscriptAnimsWorkingToIdle)
		f.orderDone(u, OrderTarget{})
	}
	if tech == nil {
		done()
		return
	}
	if decrementResearchTime(&u.Building.UpgradeResearchTime) == 0 || f.playerHasResearched(u.Owner, tech.ID) || f.st.CheatOperationCwal {
		wasResearched := f.st.TechResearched[u.Owner][tech.ID]
		f.st.TechResearched[u.Owner][tech.ID] = true
		if !wasResearched {
			f.onResearch(u, Research{ID: uint8(tech.ID)}, 1, true)
		}
		done()
	}
}

func (f *StateFunctions) orderUpgrade(u *Unit) {
	upgrade := u.Building.UpgradingType
	done := func() {
		u.Building.UpgradingType = nil
		u.Building.UpgradingLevel = 0
		if upgrade != nil {
			f.st.UpgradeUpgrading[u.Owner][upgrade.ID] = false
		}
		f.spriteRunAnim(u.Sprite, IscriptAnimsWorkingToIdle)
		f.orderDone(u, OrderTarget{})
	}
	if !f.uGroundedBuilding(u) || upgrade == nil {
		done()
		return
	}
	alreadyUpgraded := f.playerUpgradeLevel(u.Owner, upgrade.ID) >= u.Building.UpgradingLevel
	if decrementResearchTime(&u.Building.UpgradeResearchTime) == 0 || alreadyUpgraded || f.st.CheatOperationCwal {
		if !alreadyUpgraded && f.playerMaxUpgradeLevel(u.Owner, upgrade.ID) >= u.Building.UpgradingLevel {
			f.st.UpgradeLevels[u.Owner][upgrade.ID] = u.Building.UpgradingLevel
			f.onResearch(u, Research{Upgrade: true, ID: uint8(upgrade.ID)}, u.Building.UpgradingLevel, true)
			f.applyUpgradesToPlayerUnits(u.Owner)
		}
		done()
	}
}

func (f *StateFunctions) orderStayInRange(u *Unit) {
	target := u.OrderTarget.Unit
	if target == nil {
		f.haltAtMoveTarget(u)
		f.orderDone(u, OrderTarget{})
		return
	}
	u.OrderTarget.Pos = target.Sprite.Position
	if !f.unitCanAttackTarget(u, target) {
		if !f.uReadyToAttack(u) {
			f.queueOrderFront(u, f.getOrderType(OrdersMove), OrderTarget{Position: u.OrderTarget.Pos})
		}
		f.orderDone(u, OrderTarget{})
		return
	}
	if u.OrderState == 0 {
		if !f.unitTargetInWeaponMovementRange(u, target) {
			f.moveToTarget(u, target)
		}
		u.OrderState = 1
	} else {
		if !f.unitTargetInWeaponMovementRange(u, target) {
			if !f.tryFollowUnit(u, target) {
				return
			}
		} else {
			f.haltAtMoveTarget(u)
		}
	}
	if f.unitIsGoliath(u.UnitType) && !f.uMovementFlag(&u.Flingy, 2) {
		headingError := Fp8Extend(f.xyDirection(XySub(u.OrderTarget.Pos, u.Sprite.Position)).Sub(u.Heading)).Abs()
		if headingError.GreaterEqual(Fp8FromRaw(32)) {
			f.setNextTargetWaypoint(&u.Flingy, u.OrderTarget.Pos)
		}
	}
}

func (f *StateFunctions) orderTurretAttack(u *Unit) {
	f.attackUnitReacquireTarget(u)
	target := u.OrderTarget.Unit
	if target != nil {
		u.OrderTarget.Pos = target.Sprite.Position
	}
	if !f.unitCanAttackTarget(u, target) || !f.unitCanSeeTarget(u, target) {
		f.setUnitOrder(u, f.getOrderType(OrdersTurretGuard))
		return
	}
	attack := false
	if f.unitCanSeeTarget(u, target) {
		if f.unitTargetInWeaponMovementRange(u, target) {
			attack = true
		}
	}
	if !attack && (u.Subunit.OrderType != u.Subunit.UnitType.AttackUnit.Get() || f.unitIsImmovableAttacker(u.Subunit)) {
		f.setUnitOrder(u, f.getOrderType(OrdersTurretGuard))
	} else {
		attack = true
	}
	if attack {
		f.setNextTargetWaypoint(&u.Flingy, u.OrderTarget.Pos)
		if !f.uMovementFlag(&u.Subunit.Flingy, 2) {
			f.attackUnitFireWeapon(u)
		}
	}
}

func (f *StateFunctions) unitDirectionWaypoint(u *Unit) Xy {
	dir := DirectionTFromRaw(8).MulInt(u.UnitType.UnitDirection)
	return f.toXy(XyFp8Add(u.ExactPosition, f.directionXYInt(dir, 16)))
}

func (f *StateFunctions) runTurretAnim(turret *Unit, anim int32) {
	previousIscriptUnit := f.iscriptUnit
	f.iscriptUnit = turret
	f.spriteRunAnim(turret.Sprite, anim)
	f.iscriptUnit = previousIscriptUnit
}

func (f *StateFunctions) orderSieging(u *Unit) {
	if u.OrderState == 0 {
		if !f.unitIsUnsiegedTank(u.UnitType) {
			f.orderDone(u, OrderTarget{})
			return
		}
		if f.uMovementFlag(&u.Flingy, 2) {
			f.haltAtMoveTarget(u)
		}
		u.OrderState = 1
	}
	turret := f.unitTurret(u)
	if turret == nil {
		throwError("order_Sieging: null turret")
	}
	if u.OrderState == 1 {
		if !f.uMovementFlag(&u.Flingy, 2) && !f.uMovementFlag(&turret.Flingy, 2) && !f.uIscriptNobrk(turret) {
			if f.startSieging(u, turret) {
				return
			}
		}
	}
	if u.OrderState == 2 {
		if f.isFacingNextTargetWaypoint(u, Fp8FromRaw(0)) && f.isFacingNextTargetWaypoint(turret, Fp8FromRaw(0)) {
			if f.unitIs(u.UnitType, UnitTypesTerranSiegeTankTankMode) {
				f.morphUnit(u, f.getUnitType(UnitTypesTerranSiegeTankSiegeMode))
			} else if f.unitIs(u.UnitType, UnitTypesHeroEdmundDukeTankMode) {
				f.morphUnit(u, f.getUnitType(UnitTypesHeroEdmundDukeSiegeMode))
			}
			u.OrderState = 3
		}
	}
	if u.OrderState == 3 {
		if u.OrderSignal&1 != 0 {
			u.OrderSignal &^= 1
			if !u.OrderQueue.Empty() && u.OrderQueue.Front().OrderType.ID != OrdersWatchTarget {
				f.setUnitOrder(u, u.UnitType.ReturnToIdle.Get())
			}
			f.finishTankModeChange(u, turret)
		}
	}
}

func (f *StateFunctions) startSieging(u *Unit, turret *Unit) (killed bool) {
	f.setNextTargetWaypoint(&u.Flingy, f.unitDirectionWaypoint(u))
	f.setNextTargetWaypoint(&turret.Flingy, f.unitDirectionWaypoint(turret))
	f.setUnitOrderUnit(turret, f.getOrderType(OrdersNothingWait), u)
	previousIscriptUnit := f.iscriptUnit
	f.iscriptUnit = turret
	defer func() { f.iscriptUnit = previousIscriptUnit }()
	f.spriteRunAnim(turret.Sprite, IscriptAnimsSpecialState1)
	u.OrderState = 2
	for n := range f.findUnits(f.unitSpriteInnerBoundingBox(u)) {
		if n == u {
			continue
		}
		if !f.uGroundedBuilding(n) {
			continue
		}
		if f.uNoCollide(n) {
			continue
		}
		f.killUnit(u)
		return true
	}
	return false
}

func (f *StateFunctions) finishTankModeChange(u *Unit, turret *Unit) {
	f.uUnsetStatusFlag(u, UnitStatusFlagOrderNotInterruptible)
	f.orderDone(u, OrderTarget{})
	f.uUnsetStatusFlag(turret, UnitStatusFlagOrderNotInterruptible)
	f.orderDone(turret, OrderTarget{})
}

func (f *StateFunctions) orderUnsieging(u *Unit) {
	turret := f.unitTurret(u)
	if turret == nil {
		throwError("order_Unsieging: null turret")
	}
	if u.OrderState == 0 {
		if !f.unitIsSiegedTank(u.UnitType) {
			f.orderDone(u, OrderTarget{})
			return
		}
		if !f.uIscriptNobrk(turret) {
			f.setUnitOrderUnit(turret, f.getOrderType(OrdersNothingWait), u)
			f.setNextTargetWaypoint(&turret.Flingy, f.unitDirectionWaypoint(turret))
			f.runTurretAnim(turret, IscriptAnimsSpecialState1)
			u.OrderState = 2
		}
	}
	if u.OrderState == 2 {
		if f.isFacingNextTargetWaypoint(turret, Fp8FromRaw(0)) {
			f.spriteRunAnim(u.Sprite, IscriptAnimsSpecialState2)
			f.runTurretAnim(turret, IscriptAnimsSpecialState2)
			u.OrderSignal &^= 1
			u.OrderState = 3
		}
	}
	if u.OrderState == 3 {
		if u.OrderSignal&1 != 0 {
			u.OrderSignal &^= 1
			if f.unitIs(u.UnitType, UnitTypesTerranSiegeTankSiegeMode) {
				f.morphUnit(u, f.getUnitType(UnitTypesTerranSiegeTankTankMode))
			} else if f.unitIs(u.UnitType, UnitTypesHeroEdmundDukeSiegeMode) {
				f.morphUnit(u, f.getUnitType(UnitTypesHeroEdmundDukeTankMode))
			}
			f.finishTankModeChange(u, turret)
		}
	}
}

func (f *StateFunctions) orderWatchTarget(u *Unit) {
	target := u.OrderTarget.Unit
	if !f.unitCanAttackTarget(u, target) {
		f.orderDone(u, OrderTarget{Position: u.OrderTarget.Pos})
		return
	}
	if !f.unitCanSeeTarget(u, target) || !f.unitTargetInWeaponMovementRange(u, target) {
		if f.unitIsImmovableAttacker(u) {
			f.orderDone(u, OrderTarget{Position: u.OrderTarget.Pos})
			return
		}
	}
}

func (f *StateFunctions) orderPickupIdle(u *Unit) {
	target := u.OrderTarget.Unit
	if target != nil {
		f.queueOrderFront(u, f.getOrderType(OrdersPickupTransport), OrderTarget{Position: target.Sprite.Position, Unit: target})
		f.activateNextOrder(u)
	}
}

func (f *StateFunctions) orderPickupTransport(u *Unit) {
	target := u.OrderTarget.Unit
	if target == nil {
		f.orderDone(u, OrderTarget{})
		return
	}
	if !f.unitCanLoadTarget(u, target) {
		f.orderDone(u, OrderTarget{})
		return
	}
	if u.MainOrderTimer != 0 {
		return
	}
	u.MainOrderTimer = 7
	if !f.unitTargetInRange(u, target, 1) {
		f.moveToTargetReset(u, target)
		return
	}
	f.haltAtMoveTarget(u)
	if target.OrderType.ID != OrdersEnterTransport && target.OrderType.ID != OrdersPickup4 {
		f.setUnitOrderUnit(target, f.getOrderType(OrdersEnterTransport), u)
	}
	f.queueOrderFront(u, f.getOrderType(OrdersPickupIdle), OrderTarget{Position: target.Sprite.Position, Unit: target})
	f.activateNextOrder(u)
}

func (f *StateFunctions) orderMoveUnload(u *Unit) {
	if u.OrderState == 0 {
		if u.OrderTarget.Unit != nil {
			u.OrderTarget.Pos = u.OrderTarget.Unit.Sprite.Position
			u.OrderTarget.Unit = nil
			u.OrderState = 1
		}
	}
	if f.orderTargetPosDistance(u) <= 16 {
		f.queueOrderFront(u, f.getOrderType(OrdersUnload), OrderTarget{})
		f.activateNextOrder(u)
	} else {
		f.setUnitMoveTarget(u, u.OrderTarget.Pos)
		f.setNextTargetWaypoint(&u.Flingy, u.OrderTarget.Pos)
	}
}

func (f *StateFunctions) orderPlaceMine(u *Unit) {
	if u.OrderState == 0 {
		target := u.OrderTarget.Unit
		if target != nil {
			f.moveToTarget(u, target)
			f.setNextTargetWaypoint(&u.Flingy, target.Sprite.Position)
			u.OrderTarget.Pos = target.Sprite.Position
		} else {
			f.setUnitMoveTarget(u, u.OrderTarget.Pos)
			f.setNextTargetWaypoint(&u.Flingy, u.OrderTarget.Pos)
		}
		u.OrderState = 1
		return
	}
	if !f.unitCanUseTech(u, f.getTechType(TechTypesSpiderMines)) || (f.unitIsAtMoveTarget(&u.Flingy) && f.uImmovable(u)) {
		f.orderDone(u, OrderTarget{})
		return
	}
	if f.orderTargetPosDistance(u) > 20 {
		return
	}
	f.haltAtMoveTarget(u)
	pos := f.restrictMoveTargetToValidBounds(f.getUnitType(UnitTypesTerranVultureSpiderMine), u.OrderTarget.Pos)
	mine := f.createUnitByID(UnitTypesTerranVultureSpiderMine, pos, u.Owner)
	if mine != nil {
		f.onUnitCreated(mine, SourceSpiderMine, u)
		f.finishBuildingUnit(mine)
		f.completeUnit(mine)
		f.setUnitHeading(mine, u.Heading)
		u.Vulture.SpiderMineCount--
	} else {
		f.displayLastErrorForPlayer(u.Owner)
	}
	f.orderDone(u, OrderTarget{})
}

func (f *StateFunctions) spiderMineBurrow(u *Unit) {
	u.OrderSignal &^= 4
	if !f.uHallucination(u) {
		f.setUnitBurrowed(u)
		f.setUnitCloaked(u)
		f.setSpriteCloakModifier(u.Sprite, true, true, true, 0, 0)
		f.uSetStatusFlag(u, UnitStatusFlagCloaked)
		f.uSetStatusFlag(u, UnitStatusFlagRequiresDetector)
		u.DetectedFlags = 0x80000000
		u.SecondaryOrderTimer = 0
		f.setSecondaryOrder(u, f.getOrderType(OrdersCloak))
	}
	u.OrderState = 3
}

func (f *StateFunctions) spiderMineFindTarget(u *Unit) *Unit {
	rangeValue := 32 * f.unitTargetAcquisitionRange(u)
	return f.findNearestUnitToUnit(u, f.squareAt(u.Sprite.Position, rangeValue), func(target *Unit) bool {
		if target == u {
			return false
		}
		if !f.unitTargetIsEnemy(u, target) {
			return false
		}
		if f.uInvincible(target) {
			return false
		}
		if f.uGroundedBuilding(target) {
			return false
		}
		if f.uFlying(target) {
			return false
		}
		if target.UnitType.Unknown1 == 0xc1 {
			return false
		}
		return true
	})
}

func (f *StateFunctions) spiderMineChase(u *Unit, target *Unit) {
	f.resetMovementState(u)
	f.uUnsetStatusFlag(u, UnitStatusFlagPassivelyCloaked)
	f.uUnsetStatusFlag(u, UnitStatusFlagBurrowed)
	u.Sprite.SpriteType = u.UnitType.Flingy.Get().Sprite.Get()
	if target == nil || !f.unitTargetInRange(u, target, 576) {
		f.stopUnit(u)
		u.OrderState = 1
		return
	}
	f.crappyMoveToTarget(u, target)
	if f.unitTargetInRange(u, target, 30) {
		u.OrderTarget.Pos = u.Sprite.Position
		u.OrderTarget.Unit = nil
		f.spriteRunAnim(u.Sprite, IscriptAnimsSpecialState1)
		u.OrderState = 6
	} else if f.unitIsAtMoveTarget(&u.Flingy) && f.uImmovable(u) {
		f.stopUnit(u)
		u.OrderState = 1
	}
}

func (f *StateFunctions) orderSpiderMine(u *Unit) {
	if u.OrderState == 0 {
		u.GroundWeaponCooldown = 60
		u.OrderState = 1
	}
	if u.OrderState == 1 {
		if u.GroundWeaponCooldown == 0 {
			f.spriteRunAnim(u.Sprite, IscriptAnimsBurrow)
			f.uSetStatusFlag(u, UnitStatusFlagNoCollide)
			u.OrderState = 2
		}
	}
	if u.OrderState == 2 {
		if u.OrderSignal&4 != 0 {
			f.spiderMineBurrow(u)
		}
	}
	if u.OrderState == 3 {
		if !f.uCannotAttack(u) {
			target := f.spiderMineFindTarget(u)
			if target != nil {
				u.OrderTarget.Unit = target
				f.spriteRunAnim(u.Sprite, IscriptAnimsUnBurrow)
				u.Sprite.Flags &^= 0x40
				f.setSecondaryOrder(u, f.getOrderType(OrdersNothing))
				u.OrderState = 4
			}
		}
	}
	target := u.OrderTarget.Unit
	if u.OrderState == 4 {
		if u.OrderSignal&4 != 0 {
			u.OrderSignal &^= 4
			u.Sprite.ElevationLevel = u.UnitType.ElevationLevel
			f.uUnsetStatusFlag(u, UnitStatusFlagNoCollide)
			if target != nil {
				f.crappyMoveToTarget(u, target)
				u.OrderState = 5
			} else {
				u.OrderState = 1
			}
		}
	} else if u.OrderState == 5 {
		f.spiderMineChase(u, target)
	}
	if u.OrderState == 6 {
		if u.OrderSignal&1 != 0 {
			f.killUnitWith(u, noKiller(ReasonDetonated))
		}
	}
}

func (f *StateFunctions) orderCritter(u *Unit) {
	switch u.OrderState {
	case 0:
		dir := f.directionFromIndex(uint(f.lcgRandRange(48, 0, 255)))
		var distance int32 = 32
		if f.unitIs(u.UnitType, UnitTypesCritterKakaru) {
			distance = f.lcgRandRange(48, 160, 640)
		}
		targetPos := XyAdd(u.Sprite.Position, f.toXy(f.directionXYInt(dir, distance)))
		targetPos = f.restrictMoveTargetToValidBounds(u.UnitType, targetPos)
		u.OrderTarget.Pos = targetPos
		f.setUnitMoveTarget(u, targetPos)
		u.OrderState = 1
	case 1:
		if !f.unitIsAtMoveTarget(&u.Flingy) {
			return
		}
		if f.unitIs(u.UnitType, UnitTypesCritterKakaru) {
			u.OrderState = 0
		} else if f.uImmovable(u) {
			u.MainOrderTimer = 75
			u.OrderState = 2
		} else {
			u.MainOrderTimer = f.lcgRandRange(47, 0, 75)
			if u.MainOrderTimer <= 15 {
				u.MainOrderTimer = 0
			}
			u.OrderState = 2
		}
	case 2:
		if u.MainOrderTimer == 0 {
			u.OrderState = 0
		}
	}
}

func (f *StateFunctions) orderHoldPosition(u *Unit) {
	if u.OrderState == 0 {
		f.haltAtMoveTarget(u)
		f.uSetStatusFlag(u, UnitStatusFlagReadyToAttack)
		if u.Subunit != nil {
			f.uSetStatusFlag(u.Subunit, UnitStatusFlagReadyToAttack)
		}
		u.OrderState = 1
	}
	if f.tryAttackSomething(u) {
		turret := f.unitTurret(u)
		if turret == nil || f.unitIsGoliath(u.UnitType) {
			f.setNextTargetWaypoint(&u.Flingy, u.OrderTarget.Pos)
		}
		return
	}
	if u.MainOrderTimer == 0 {
		u.MainOrderTimer = 15
		target := f.findAcquireRandomTarget(u)
		u.OrderTarget.Unit = target
		if target != nil {
			u.OrderProcessTimer = 0
		}
	}
}
