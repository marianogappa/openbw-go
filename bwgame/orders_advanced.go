package bwgame

func directionLiteral(value uint8) DirectionT {
	return DirectionT{RawValue: int8(value)}
}

func unitIsPlacedAt(n *Unit, unitType *UnitTypeT, pos Xy) bool {
	if n.UnitType != unitType {
		return false
	}
	placementSize := n.UnitType.PlacementSize
	if uint32(n.Sprite.Position.X+placementSize.X/2)-uint32(pos.X) >= uint32(placementSize.X) {
		return false
	}
	if uint32(n.Sprite.Position.Y+placementSize.Y/2)-uint32(pos.Y) >= uint32(placementSize.Y) {
		return false
	}
	return true
}

func searchAreaAround(pos Xy, halfSize int32) Rect {
	return Rect{From: XySub(pos, Xy{X: halfSize, Y: halfSize}), To: XyAdd(pos, Xy{X: halfSize, Y: halfSize})}
}

func (f *StateFunctions) setReadyToAttack(u *Unit) {
	f.uSetStatusFlag(u, UnitStatusFlagReadyToAttack)
	if u.Subunit != nil {
		f.uSetStatusFlag(u.Subunit, UnitStatusFlagReadyToAttack)
	}
}

func (f *StateFunctions) countDownRemainingBuildTime(u *Unit) {
	if f.st.CheatOperationCwal {
		if u.RemainingBuildTime > 16 {
			u.RemainingBuildTime -= 16
		} else {
			u.RemainingBuildTime = 0
		}
	} else {
		u.RemainingBuildTime--
	}
}

func (f *StateFunctions) createVisibleThingyAbove(spriteType *SpriteTypeT, pos Xy, owner int32, elevationLevel int32) *Thingy {
	t := f.createThingy(spriteType, pos, owner)
	if t != nil {
		t.Sprite.ElevationLevel = elevationLevel + 1
		if !f.usHidden(t) {
			f.setSpriteVisibility(t.Sprite, int32(f.tileVisibility(t.Sprite.Position)))
		}
	}
	return t
}

func (f *StateFunctions) lengthFromUnitToPos(u *Unit, pos Xy) int32 {
	return f.xyLengthFP8(XyFp8Sub(f.toXyFp8(pos), u.ExactPosition)).IntegerPart()
}

func (f *StateFunctions) orderRightClickAction(u *Unit) {
	if u.OrderState == 0 {
		f.setUnitMoveTarget(u, u.OrderTarget.Pos)
		f.setNextTargetWaypoint(&u.Flingy, u.OrderTarget.Pos)
		u.OrderState = 1
	} else if u.OrderState == 1 {
		if f.unitPositionIsVisible(u, u.OrderTarget.Pos) {
			if u.OrderUnitType == nil {
				throwError("order_RightClickAction: null unit type")
			}
			targetUnitType := u.OrderUnitType
			pos := u.OrderTarget.Pos
			target := f.findUnitNoexpand(searchAreaAround(pos, 96), func(n *Unit) bool {
				return unitIsPlacedAt(n, targetUnitType, pos)
			})
			order := f.getDefaultOrder(f.defaultAction(u), u, pos, target, nil)
			if order == nil {
				f.stopUnit(u)
				f.setNextTargetWaypoint(&u.Flingy, u.MoveTarget.Pos)
				f.orderDone(u, OrderTarget{})
			} else {
				if order.ID == OrdersAttackDefault {
					order = u.UnitType.AttackUnit.Get()
				}
				if f.unitCanReceiveOrder(u, order, u.Owner) {
					f.rightClickIssueOrder(u, order, pos, target)
				}
			}
		}
	}
}

func (f *StateFunctions) rightClickIssueOrder(u *Unit, order *OrderTypeT, pos Xy, target *Unit) {
	if order.ID == OrdersRallyPointUnit {
		if f.unitIsFactory(u) {
			if target == nil {
				target = u
			}
			u.Building.Rally.Unit = target
			u.Building.Rally.Pos = target.Sprite.Position
		}
	} else if order.ID == OrdersRallyPointTile {
		if f.unitIsFactory(u) {
			u.Building.Rally.Unit = nil
			u.Building.Rally.Pos = pos
		}
	} else if target != nil {
		f.queueOrderFront(u, order, OrderTarget{Position: target.Sprite.Position, Unit: target})
		f.orderDone(u, OrderTarget{})
	} else {
		f.queueOrderFront(u, order, OrderTarget{Position: pos})
		f.orderDone(u, OrderTarget{})
	}
}

func (f *StateFunctions) orderPatrol(u *Unit) {
	if u.OrderState == 0 {
		if u.OrderQueue.Empty() || u.OrderQueue.Front().OrderType.ID != OrdersPatrol {
			f.setQueuedOrder(u, true, f.getOrderType(OrdersPatrol), OrderTarget{Position: u.OrderTarget.Pos, Unit: u.OrderTarget.Unit})
			f.activateNextOrder(u)
			f.queueOrderBack(u, f.getOrderType(OrdersPatrol), OrderTarget{Position: u.Sprite.Position})
			return
		}
	}
	if !f.attackMoveMovement(u, false) {
		return
	}
	if u.MainOrderTimer != 0 {
		return
	}
	u.MainOrderTimer = 15
	if f.unitIs(u.UnitType, UnitTypesTerranMedic) && u.OrderTarget.Unit == nil {
		target := f.findMedicTarget(u)
		if target != nil {
			f.queueOrderFront(u, f.getOrderType(OrdersPatrol), OrderTarget{Position: u.OrderTarget.Pos})
			f.queueOrderFront(u, f.getOrderType(OrdersMedicHeal), OrderTarget{Position: target.Sprite.Position, Unit: target})
			u.AutoTargetUnit = nil
			f.orderDone(u, OrderTarget{})
			f.setReadyToAttack(u)
			return
		}
	}
	if !f.unitIsAtMoveTarget(&u.Flingy) {
		f.attackMoveAcquireTarget(u, false)
	} else {
		if f.uImmovable(u) {
			f.queueOrderBack(u, f.getOrderType(OrdersPatrol), OrderTarget{Position: u.OrderTarget.Pos})
		} else {
			f.queueOrderBack(u, f.getOrderType(OrdersPatrol), OrderTarget{Position: u.Sprite.Position})
		}
		f.uUnsetStatusFlag(u, UnitStatusFlagOrderNotInterruptible)
		f.orderDone(u, OrderTarget{})
	}
}

func (f *StateFunctions) orderCastScannerSweep(u *Unit) {
	energyCost := Fp8Integer(f.getTechType(TechTypesScannerSweep).EnergyCost)
	if u.Energy.Less(energyCost) {
		f.orderDone(u, OrderTarget{})
		return
	}
	scanType := f.getUnitType(UnitTypesSpellScannerSweep)
	pos := f.restrictMoveTargetToValidBounds(scanType, u.OrderTarget.Pos)
	scan := f.createUnit(scanType, pos, u.Owner)
	if scan != nil {
		f.finishBuildingUnit(scan)
		f.completeUnit(scan)
		f.orderDone(u, OrderTarget{})
		u.Energy = u.Energy.Sub(energyCost)
		f.playSoundUnit(388, u, false)
	} else {
		f.displayLastErrorForPlayer(u.Owner)
		f.orderDone(u, OrderTarget{})
	}
}

func (f *StateFunctions) orderScanner(u *Unit) {
	if u.OrderSignal&4 != 0 {
		u.OrderSignal &^= 4
		f.killUnit(u)
	}
}

func (f *StateFunctions) orderTowerGuard(u *Unit) {
	if f.unitIsDisabled(u) {
		return
	}
	if u.MainOrderTimer == 0 {
		u.MainOrderTimer = 15
		target := f.findAcquireRandomTarget(u)
		u.OrderTarget.Unit = target
		if target != nil {
			f.setUnitOrderUnit(u, f.getOrderType(OrdersTowerAttack), target)
			f.setReadyToAttack(u)
		}
	}
}

func (f *StateFunctions) orderTowerAttack(u *Unit) {
	if f.unitIsDisabled(u) || !f.tryAttackSomething(u) {
		f.orderDone(u, OrderTarget{})
	} else {
		f.setUnitMoveTarget(u, u.Sprite.Position)
		f.setNextTargetWaypoint(&u.Flingy, u.OrderTarget.Pos)
	}
}

func (f *StateFunctions) orderCastDefensiveMatrix(u *Unit) {
	tech := f.getTechType(TechTypesDefensiveMatrix)
	if f.spellCastTargetMovement(u, tech, f.unitSightRange(u, true)) {
		u.Energy = u.Energy.Sub(Fp8Integer(tech.EnergyCost))
		target := u.OrderTarget.Unit
		target.DefensiveMatrixHp = Fp8Integer(250)
		target.DefensiveMatrixTimer = 168
		f.createDefensiveMatrixImage(target)
		f.playSoundUnit(349, u, false)
		f.createImage(f.getImageType(ImageTypesIMAGEIDScienceVesselOverlayPart2), u.Sprite, Xy{}, imageOrderAbove, nil)
		f.orderDone(u, OrderTarget{})
	}
}

func (f *StateFunctions) orderSpell(u *Unit) {
	target := u.OrderTarget.Unit
	if u.OrderState == 2 {
		if u.OrderSignal&2 != 0 {
			u.OrderSignal &^= 2
			f.orderDone(u, OrderTarget{})
		} else if target != nil {
			u.OrderTarget.Pos = target.Sprite.Position
			f.setNextTargetWaypoint(&u.Flingy, target.Sprite.Position)
		}
		return
	}
	if !f.spellOrderValid(u) {
		f.stopUnit(u)
		f.orderDone(u, OrderTarget{})
		return
	}
	var tech *TechTypeT
	if u.OrderType.TechType != TechTypesNone {
		tech = f.getTechType(u.OrderType.TechType)
	}
	if tech != nil && u.Energy.Less(Fp8Integer(tech.EnergyCost)) {
		f.orderDone(u, OrderTarget{})
		return
	}
	weapon := f.getWeaponType(u.OrderType.Weapon)
	if target != nil {
		u.OrderTarget.Pos = target.Sprite.Position
	}
	f.setNextTargetWaypoint(&u.Flingy, u.OrderTarget.Pos)
	if f.spellTargetIsInRange(u, target, weapon) {
		if !f.uMovementFlag(&u.Flingy, 4) {
			f.stopUnit(u)
		}
		if u.SpellCooldown == 0 && f.unitCanFireWeapon(u, weapon) && !f.uMovementFlag(&u.Flingy, 2) && f.unitIsAtMoveTarget(&u.Flingy) {
			if tech != nil {
				u.Energy = u.Energy.Sub(Fp8Integer(tech.EnergyCost))
			}
			u.SpellCooldown = f.getModifiedWeaponCooldown(u, weapon) + (f.lcgRand(49) & 3) - 1
			f.uSetMovementFlag(&u.Flingy, 8)
			u.OrderSignal &^= 2
			f.spriteRunAnim(u.Sprite, u.OrderType.Animation)
			u.OrderState = 2
		}
	} else {
		if target != nil {
			if u.OrderState == 0 {
				f.moveToTarget(u, target)
			} else {
				f.tryFollowUnit(u, target)
			}
		} else if u.MoveTargetTimer == 0 {
			f.setUnitMoveTarget(u, u.OrderTarget.Pos)
		}
		f.setNextTargetWaypoint(&u.Flingy, u.OrderTarget.Pos)
		if f.unitIsAtMoveTarget(&u.Flingy) && f.uImmovable(u) {
			f.orderDone(u, OrderTarget{})
			return
		}
	}
	if u.OrderState == 0 {
		u.OrderState = 1
	}
}

func (f *StateFunctions) spellTargetIsInRange(u *Unit, target *Unit, weapon *WeaponTypeT) bool {
	if !f.unitCanSeeOrderTarget(u) {
		return false
	}
	if target != nil {
		rangeValue := f.weaponMaxRange(u, weapon)
		rangeValue += f.unitTargetMovementRange(&u.Flingy, &target.Flingy)
		return f.unitTargetInRange(u, target, rangeValue)
	}
	rangeValue := f.weaponMaxRange(u, weapon)
	rangeValue += f.unitHaltDistance(&u.Flingy).IntegerPart()
	return f.lengthFromUnitToPos(u, u.OrderTarget.Pos) <= rangeValue
}

func (f *StateFunctions) orderMoveToTargetOrder(u *Unit) {
	if u.OrderState == 0 {
		if f.uCanMove(u) {
			f.setUnitMoveTarget(u, u.OrderTarget.Pos)
			f.setNextTargetWaypoint(&u.Flingy, u.OrderTarget.Pos)
		}
		u.OrderState = 1
		return
	}
	if !f.playerPositionIsVisible(u.Owner, u.OrderTarget.Pos) {
		return
	}
	pos := u.OrderTarget.Pos
	target := f.findUnitNoexpand(searchAreaAround(pos, 96), func(n *Unit) bool {
		return unitIsPlacedAt(n, u.OrderUnitType, pos)
	})
	if target != nil {
		order := f.getOrderType(u.OrderType.TargetOrder)
		if order.ID == OrdersAttackDefault {
			order = u.UnitType.AttackUnit.Get()
			turret := f.unitTurret(u)
			if turret != nil {
				f.queueOrderFront(turret, turret.UnitType.AttackUnit.Get(), OrderTarget{Position: target.Sprite.Position, Unit: target})
				f.orderDone(turret, OrderTarget{})
			}
		}
		f.queueOrderFront(u, order, OrderTarget{Position: target.Sprite.Position, Unit: target})
	}
	f.orderDone(u, OrderTarget{})
}

func (f *StateFunctions) orderPickupBunker(u *Unit) {
	target := u.OrderTarget.Unit
	if target != nil && f.unitCanLoadTarget(u, target) {
		f.setUnitOrderUnit(target, f.getOrderType(OrdersEnterTransport), u)
	}
	f.orderDone(u, OrderTarget{})
}

func (f *StateFunctions) orderHealMove(u *Unit) {
	target := u.OrderTarget.Unit
	if target != nil {
		if f.medicCanHealTarget(u, target) {
			f.setUnitOrderUnit(u, f.getOrderType(OrdersMedicHeal), target)
			f.moveToTarget(u, target)
			return
		}
		u.OrderTarget.Pos = target.Sprite.Position
	}
	if u.MoveTargetTimer == 0 {
		f.setUnitMoveTarget(u, u.OrderTarget.Pos)
	}
	if f.unitIsAtMoveTarget(&u.Flingy) {
		f.orderDone(u, OrderTarget{})
		return
	}
	if u.MainOrderTimer == 0 {
		u.MainOrderTimer = 15
		healTarget := f.findMedicTarget(u)
		if healTarget != nil {
			prevTargetPos := u.OrderTarget.Pos
			f.setUnitOrderUnit(u, f.getOrderType(OrdersMedicHeal), healTarget)
			f.moveToTarget(u, healTarget)
			if target != nil {
				f.queueOrderBack(u, f.getOrderType(OrdersHealMove), OrderTarget{Position: target.Sprite.Position, Unit: target})
			} else {
				f.queueOrderBack(u, f.getOrderType(OrdersHealMove), OrderTarget{Position: prevTargetPos})
			}
		}
	}
}

func (f *StateFunctions) medicStartHealAnimation(u *Unit) {
	if ^u.OrderSignal&1 != 0 {
		f.spriteRunAnim(u.Sprite, IscriptAnimsSpecialState1)
		if u.OrderQueue.Empty() || u.OrderQueue.Front().OrderType.ID != OrdersMedicHealToIdle {
			f.queueOrderFront(u, f.getOrderType(OrdersMedicHealToIdle), OrderTarget{})
		}
	}
}

func (f *StateFunctions) orderMedicHeal(u *Unit) {
	r := f.medicTryHeal(u)
	if r == 0 {
		f.stopUnit(u)
		f.orderDone(u, OrderTarget{})
		return
	}
	target := u.OrderTarget.Unit
	if u.OrderState == 0 {
		if target != nil && f.tryFollowUnit(u, target) && (!f.unitIsAtMoveTarget(&u.Flingy) || !f.uImmovable(u)) {
			if f.unitTargetInRange(u, target, 30) {
				f.stopUnit(u)
				u.OrderState = 1
			}
		} else {
			f.orderDone(u, OrderTarget{})
		}
	} else {
		if r == 1 {
			f.medicStartHealAnimation(u)
			f.setNextTargetWaypoint(&u.Flingy, target.Sprite.Position)
		} else {
			u.OrderState = 0
		}
	}
}

func (f *StateFunctions) orderMedicHealToIdle(u *Unit) {
	if ^u.MovementFlags&2 != 0 {
		f.spriteRunAnim(u.Sprite, IscriptAnimsWalkingToIdle)
	}
	f.uUnsetStatusFlag(u, UnitStatusFlagOrderNotInterruptible)
	f.orderDone(u, OrderTarget{})
}

func (f *StateFunctions) orderMedicIdle(u *Unit) {
	target := f.findMedicTarget(u)
	if target != nil {
		f.setUnitOrderUnit(u, f.getOrderType(OrdersMedicHeal), target)
		f.moveToTarget(u, target)
	}
}

func (f *StateFunctions) orderMedicHoldPosition(u *Unit) {
	if u.OrderState == 0 {
		f.stopUnit(u)
		f.setNextTargetWaypoint(&u.Flingy, u.MoveTarget.Pos)
		f.setReadyToAttack(u)
		u.OrderState = 1
		f.spriteRunAnim(u.Sprite, IscriptAnimsWalkingToIdle)
	}
	r := f.medicTryHeal(u)
	if r == 1 {
		f.medicStartHealAnimation(u)
		f.setNextTargetWaypoint(&u.Flingy, u.OrderTarget.Unit.Sprite.Position)
		return
	}
	if r == 3 {
		u.OrderTarget.Unit = nil
	}
	f.spriteRunAnim(u.Sprite, IscriptAnimsWalkingToIdle)
	if u.MainOrderTimer == 0 {
		u.MainOrderTimer = 15
		target := f.findMedicTarget(u)
		u.OrderTarget.Unit = target
		if target != nil {
			u.OrderTarget.Pos = target.Sprite.Position
		}
	}
}

func (f *StateFunctions) orderNukeTrain(u *Unit) {
	if !f.unitIs(u.UnitType, UnitTypesTerranNuclearSilo) {
		throwError("order_NukeTrain: unit is not a nuclear silo")
	}
	if u.OrderState == 0 {
		if u.CurrentBuildUnit != nil {
			u.Building.Silo.Nuke = u.CurrentBuildUnit
			u.CurrentBuildUnit.ConnectedUnit = u
			u.OrderState = 1
		}
		return
	}
	if u.Building.Silo.Nuke == nil {
		u.OrderState = 0
		return
	}
	if f.uCompleted(u.Building.Silo.Nuke) {
		u.Building.Silo.Ready = true
		u.OrderState = 0
	}
}

func (f *StateFunctions) orderNukePaint(u *Unit) {
	if u.OrderTarget.Unit != nil {
		f.setUnitOrderUnit(u, f.getOrderType(OrdersNukeUnit), u)
	} else {
		f.setUnitOrderPos(u, f.getOrderType(OrdersCastNuclearStrike), u.OrderTarget.Pos)
	}
}

func (f *StateFunctions) orderNukeUnit(u *Unit) {
	if u.OrderTarget.Unit == nil {
		f.stopUnit(u)
		f.orderDone(u, OrderTarget{})
		return
	}
	u.OrderTarget.Pos = u.OrderTarget.Unit.Sprite.Position
	f.orderCastNuclearStrike(u)
}

func (f *StateFunctions) orderCastNuclearStrike(u *Unit) {
	f.setNextTargetWaypoint(&u.Flingy, u.OrderTarget.Pos)
	if f.lengthFromPosToUnit(u, u.OrderTarget.Pos) > f.unitSightRange(u, false) {
		if u.MoveTargetTimer == 0 {
			f.setUnitMoveTarget(u, u.OrderTarget.Pos)
		}
		return
	}
	f.stopUnit(u)
	if !f.isFacingNextTargetWaypoint(u, Fp8Integer(1)) {
		return
	}
	silo := f.findReadyNuclearSilo(u.Owner)
	if silo == nil {
		f.orderDone(u, OrderTarget{})
		return
	}
	f.playSoundUnit(239, u, false)
	nuke := silo.Building.Silo.Nuke
	silo.Building.Silo.Nuke = nil
	silo.Building.Silo.Ready = false
	f.setUnitOrderPos(nuke, f.getOrderType(OrdersNukeLaunch), u.OrderTarget.Pos)
	nuke.ConnectedUnit = u
	f.setSpriteImagesHeadingByIndex(nuke.Sprite, 0)
	f.showUnit(nuke)
	u.ConnectedUnit = nuke
	f.setUnitOrderPos(u, f.getOrderType(OrdersNukeTrack), u.Sprite.Position)
}

func (f *StateFunctions) findReadyNuclearSilo(owner int32) *Unit {
	for n := range f.st.PlayerUnits[owner].All() {
		if f.unitIs(n.UnitType, UnitTypesTerranNuclearSilo) && n.Building.Silo.Nuke != nil && n.Building.Silo.Ready {
			return n
		}
	}
	return nil
}

func (f *StateFunctions) orderNukeLaunch(u *Unit) {
	if u.ConnectedUnit == nil && u.OrderState < 5 {
		f.killUnit(u)
		return
	}
	switch u.OrderState {
	case 0:
		f.playSoundUnit(84, u, false)
		targetPos := Xy{X: u.Sprite.Position.X, Y: u.UnitType.Dimensions.From.Y}
		f.setUnitMoveTarget(u, targetPos)
		f.setNextTargetWaypoint(&u.Flingy, targetPos)
		u.MainOrderTimer = 90
		u.OrderState = 1
	case 1:
		if u.MainOrderTimer <= 45 || f.unitIsAtMoveTarget(&u.Flingy) {
			f.playSound(127, true)
			u.OrderState = 2
		}
	case 2:
		if u.MainOrderTimer == 0 || f.unitIsAtMoveTarget(&u.Flingy) {
			f.hideUnit(u, true)
			u.ConnectedUnit.ConnectedUnit = u
			f.stopUnit(u)
			u.OrderState = 3
		}
	case 3:
		if ^u.MovementFlags&2 != 0 {
			f.spriteRunAnim(u.Sprite, IscriptAnimsWarpIn)
			u.OrderState = 4
		}
	case 4:
		if u.OrderSignal&2 != 0 {
			u.OrderSignal &^= 2
			var appearAtPos Xy
			appearAtPos.X = max(u.UnitType.Dimensions.From.X, u.OrderTarget.Pos.X)
			appearAtPos.Y = max(u.UnitType.Dimensions.From.Y, u.OrderTarget.Pos.Y-320)
			f.moveUnit(u, appearAtPos)
			f.setUnitHeading(u, directionLiteral(128))
			f.setUnitMoveTarget(u, u.OrderTarget.Pos)
			f.setNextTargetWaypoint(&u.Flingy, u.OrderTarget.Pos)
			f.showUnit(u)
			u.OrderState = 5
		}
	case 5:
		if f.lengthFromPosToUnit(u, u.MoveTarget.Pos) <= 10 {
			u.OrderTarget.Unit = u
			f.spriteRunAnim(u.Sprite, IscriptAnimsSpecialState1)
			u.OrderState = 6
		}
	case 6:
		if u.OrderSignal&1 != 0 {
			u.OrderSignal &^= 1
			u.UserActionFlags |= 4
			f.killUnit(u)
		}
	}
}

func (f *StateFunctions) lengthFromPosToUnit(u *Unit, pos Xy) int32 {
	return f.xyLengthFP8(XyFp8Sub(u.ExactPosition, f.toXyFp8(pos))).IntegerPart()
}

func (f *StateFunctions) orderNukeTrack(u *Unit) {
	if !f.unitIsGhost(u.UnitType) {
		throwError("order_NukeTrack: unit is not a ghost")
	}
	if u.OrderState == 0 {
		f.spriteRunAnim(u.Sprite, IscriptAnimsSpecialState1)
		var pos Xy
		if u.ConnectedUnit != nil {
			pos = u.ConnectedUnit.OrderTarget.Pos
		}
		u.Ghost.NukeDot = f.createVisibleThingyAbove(f.getSpriteType(SpriteTypesSPRITEIDNukeTargetDot), pos, u.Owner, u.Sprite.ElevationLevel)
		u.OrderState = 6
	} else if u.OrderState == 6 {
		if u.ConnectedUnit == nil || u.ConnectedUnit.OrderState == 5 {
			if u.OrderQueue.Empty() {
				f.setQueuedOrder(u, false, u.UnitType.ReturnToIdle.Get(), OrderTarget{})
			}
			f.uUnsetStatusFlag(u, UnitStatusFlagOrderNotInterruptible)
			f.orderDone(u, OrderTarget{})
			f.spriteRunAnim(u.Sprite, IscriptAnimsWalkingToIdle)
			if u.Ghost.NukeDot != nil {
				f.spriteRunAnim(u.Ghost.NukeDot.Sprite, IscriptAnimsDeath)
				u.Ghost.NukeDot = nil
			}
		}
	}
}

func (f *StateFunctions) orderLarva(u *Unit) {
	hatchery := u.ConnectedUnit
	if hatchery != nil && !f.unitTargetInRange(u, hatchery, 10) {
		f.moveToTarget(u, hatchery)
		f.setNextTargetWaypoint(&u.Flingy, hatchery.Sprite.Position)
	}
	if !(f.unitIsAtMoveTarget(&u.Flingy) || (u.MoveTarget.Unit != nil && u.MoveTarget.Pos != u.MoveTarget.Unit.Sprite.Position && !f.tileHasCreep(u.MoveTarget.Pos))) {
		return
	}
	if (hatchery == nil || u.MoveTarget.Unit != hatchery) && !f.tileHasCreep(u.Sprite.Position) {
		f.killUnit(u)
		return
	}
	rv := f.lcgRand(20)
	targetPos := u.Sprite.Position
	if rv&8 != 0 {
		targetPos.X += 10
	} else {
		targetPos.X -= 10
	}
	if rv&0x80 != 0 {
		targetPos.Y += 10
	} else {
		targetPos.Y -= 10
	}
	if !f.isPosInMapBounds(targetPos) || !f.tileHasCreep(targetPos) {
		return
	}
	if hatchery != nil {
		bb := f.unitSpriteInnerBoundingBox(hatchery)
		isOnCorrectSide := func(pos Xy) bool {
			switch u.OrderState {
			case 0:
				return pos.X <= bb.From.X
			case 1:
				return pos.Y <= bb.From.Y
			case 2:
				return pos.X >= bb.To.X
			default:
				return pos.Y >= bb.To.Y
			}
		}
		if !isOnCorrectSide(targetPos) {
			if isOnCorrectSide(u.Sprite.Position) {
				return
			}
			switch u.OrderState {
			case 0:
				targetPos = Xy{X: bb.From.X - 10, Y: hatchery.Sprite.Position.Y}
			case 1:
				targetPos = Xy{X: hatchery.Sprite.Position.X, Y: bb.From.Y - 10}
			case 2:
				targetPos = Xy{X: bb.To.X + 10, Y: hatchery.Sprite.Position.Y}
			default:
				targetPos = Xy{X: hatchery.Sprite.Position.X, Y: bb.To.Y + 10}
			}
		}
	}
	f.setUnitMoveTarget(u, targetPos)
	f.setNextTargetWaypoint(&u.Flingy, targetPos)
}

func (f *StateFunctions) orderInitCreepGrowth(u *Unit) {
	f.setSecondaryOrder(u, f.getOrderType(OrdersSpreadCreep))
	f.setUnitOrder(u, u.UnitType.ReturnToIdle.Get())
}

func (f *StateFunctions) zergEggType(u *Unit) *UnitTypeT {
	if f.unitIs(u.UnitType, UnitTypesZergLarva) {
		return f.getUnitType(UnitTypesZergEgg)
	}
	if f.unitIs(u.UnitType, UnitTypesZergHydralisk) {
		return f.getUnitType(UnitTypesZergLurkerEgg)
	}
	if f.unitIs(u.UnitType, UnitTypesZergMutalisk) {
		return f.getUnitType(UnitTypesZergCocoon)
	}
	return nil
}

func (f *StateFunctions) orderZergUnitMorph(u *Unit) {
	if u.BuildQueue.Empty() {
		throwError("order_ZergUnitMorph: empty build queue")
	}
	unitType := u.BuildQueue.Front()
	switch u.OrderState {
	case 0:
		eggType := f.zergEggType(u)
		if eggType == nil {
			throwError("order_ZergUnitMorph: no egg type")
		}
		if !f.hasAvailableSupplyFor(u.Owner, unitType, true) {
			f.cancelBuildQueue(u)
			f.setUnitOrder(u, u.UnitType.ReturnToIdle.Get())
			return
		}
		f.addCompletedUnit(u, -1, false)
		f.uUnsetStatusFlag(u, UnitStatusFlagCompleted)
		f.morphUnit(u, eggType)
		u.RemainingBuildTime = unitType.BuildTime
		u.OrderState = 1
	case 1:
		if u.RemainingBuildTime != 0 {
			f.countDownRemainingBuildTime(u)
		} else {
			f.spriteRunAnim(u.Sprite, IscriptAnimsSpecialState1)
			u.OrderState = 2
		}
	case 2:
		if u.OrderSignal&4 != 0 {
			u.OrderSignal &^= 4
			f.morphUnit(u, unitType)
			f.buildQueueClearSingle(u)
			if u.UnitType.ConstructionAnimation.ID != 0 {
				f.setConstructionGraphic(u, true)
				f.spriteRunAnim(u.Sprite, IscriptAnimsSpecialState1)
				f.setUnitOrder(u, f.getOrderType(OrdersZergBirth))
			} else {
				f.finishBuildingUnit(u)
				f.completeUnit(u)
				if f.unitIs(u.UnitType, UnitTypesZergEgg) && u.ConnectedUnit != nil {
					f.rallyUnit(u, u.ConnectedUnit)
				}
			}
		}
	}
}

func (f *StateFunctions) orderZergBirth(u *Unit) {
	if ^u.OrderSignal&4 != 0 {
		return
	}
	u.OrderSignal &^= 4
	var u2 *Unit
	if f.utTwoUnitsInOneEgg(u.UnitType) {
		u2 = f.spawnSecondEggUnit(u)
	}
	f.setConstructionGraphic(u, false)
	f.finishBuildingUnit(u)
	f.completeUnit(u)
	pos := u.Sprite.Position
	if !f.utTwoUnitsInOneEgg(u.UnitType) && u.PreviousUnitType != f.getUnitType(UnitTypesZergCocoon) {
		if u.UnitType.Unknown1 == 0xc1 {
			pos.Y -= 7
		} else if f.uFlying(u) {
			pos.Y -= 42
		}
	}
	pos = f.restrictMoveTargetToValidBounds(u.UnitType, pos)
	if pos != u.Sprite.Position {
		f.moveUnit(u, pos)
	}
	if u.PreviousUnitType != f.getUnitType(UnitTypesZergCocoon) && u.PreviousUnitType != f.getUnitType(UnitTypesZergLurkerEgg) {
		if u.ConnectedUnit != nil {
			f.rallyUnit(u, u.ConnectedUnit)
			if u2 != nil {
				f.rallyUnit(u2, u.ConnectedUnit)
			}
		}
	}
	f.uUnsetStatusFlag(u, UnitStatusFlagOrderNotInterruptible)
	f.orderDone(u, OrderTarget{})
}

func (f *StateFunctions) spawnSecondEggUnit(u *Unit) *Unit {
	a := XyAdd(u.Sprite.Position, f.getImageLoOffset(u.Sprite.MainImage, 3, 0, false))
	b := XyAdd(u.Sprite.Position, f.getImageLoOffset(u.Sprite.MainImage, 3, 1, false))
	a = f.restrictMoveTargetToValidBounds(u.UnitType, a)
	b = f.restrictMoveTargetToValidBounds(u.UnitType, b)
	f.moveUnit(u, a)
	u2 := f.createUnit(u.UnitType, b, u.Owner)
	if u2 == nil {
		f.displayLastErrorForPlayer(u.Owner)
		return nil
	}
	if f.unitIs(u2.UnitType, UnitTypesZergZergling) {
		f.setUnitHeading(u2, directionLiteral(144))
	} else if f.unitIs(u2.UnitType, UnitTypesZergScourge) {
		f.setUnitHeading(u2, directionLiteral(16))
	} else {
		f.setUnitHeading(u, directionLiteral(0))
	}
	f.finishBuildingUnit(u2)
	f.completeUnit(u2)
	f.copyStatusEffects(u2, u)
	return u2
}

func (f *StateFunctions) orderDroneStartBuild(u *Unit) {
	f.setUnitOrderPos(u, f.getOrderType(OrdersDroneLand), u.OrderTarget.Pos)
	f.queueOrderBack(u, f.getOrderType(OrdersDroneBuild), OrderTarget{Position: u.OrderTarget.Pos})
}

func (f *StateFunctions) setLastImageOffsetYIfModifier10(u *Unit, y int32) {
	if u.Sprite.Images.Back().Modifier == 10 {
		u.Sprite.Images.Back().Offset.Y = y
	}
}

func (f *StateFunctions) orderDroneLand(u *Unit) {
	if u.BuildQueue.Empty() {
		throwError("order_DroneLand: empty build queue")
	}
	buildType := u.BuildQueue.Front()
	switch u.OrderState {
	case 0:
		pos := XySub(u.OrderTarget.Pos, Xy{X: 0, Y: 7})
		f.setUnitMoveTarget(u, pos)
		f.setNextTargetWaypoint(&u.Flingy, pos)
		u.OrderState = 1
	case 1:
		if f.unitIsAtMoveTarget(&u.Flingy) {
			f.droneLandAtTarget(u, buildType)
		}
	case 2:
		if f.unitIsAtMoveTarget(&u.Flingy) {
			f.droneLandFinish(u)
		}
	}
}

func (f *StateFunctions) droneLandAtTarget(u *Unit, buildType *UnitTypeT) {
	success := false
	if !f.uImmovable(u) {
		success = f.canPlaceBuilding(u, u.Owner, buildType, u.OrderTarget.Pos, true, true)
		if success && f.unitIs(buildType, UnitTypesZergExtractor) {
			gas := f.getBuildingAtCenterPosition(u.OrderTarget.Pos, f.getUnitType(UnitTypesResourceVespeneGeyser))
			if gas != nil && f.unitTargetInRange(u, gas, 32) {
				f.uUnsetStatusFlag(u, UnitStatusFlagGroundUnit)
				u.Sprite.ElevationLevel = u.UnitType.ElevationLevel + 1
				f.setQueuedOrder(u, false, f.getOrderType(OrdersResetCollision), OrderTarget{})
				f.setQueuedOrder(u, false, u.UnitType.ReturnToIdle.Get(), OrderTarget{})
				f.setUnitMoveTarget(u, u.OrderTarget.Pos)
				f.setNextTargetWaypoint(&u.Flingy, u.OrderTarget.Pos)
				f.uSetStatusFlag(u, UnitStatusFlagOrderNotInterruptible)
				u.OrderState = 2
				return
			}
		}
	}
	if success {
		if f.lengthFromUnitToPos(u, u.OrderTarget.Pos) <= 128 {
			if u.Sprite.Images.Back().Modifier == 10 {
				u.Sprite.Images.Back().FrozenYValue = f.getImageMapPosition(u.Sprite.Images.Back()).Y
				u.Sprite.Images.Back().Flags |= ImageFlagYFrozen
			}
			f.setUnitMoveTarget(u, u.OrderTarget.Pos)
			f.setNextTargetWaypoint(&u.Flingy, u.OrderTarget.Pos)
			f.uSetStatusFlag(u, UnitStatusFlagOrderNotInterruptible)
			u.OrderState = 2
		} else {
			success = false
		}
	}
	if !success {
		f.removeOneOrder(u, f.getOrderType(OrdersDroneBuild))
		f.setUnitOrder(u, u.UnitType.ReturnToIdle.Get())
	}
}

func (f *StateFunctions) droneLandFinish(u *Unit) {
	if u.Sprite.Images.Back().Modifier == 10 {
		u.Sprite.Images.Back().Offset.Y = 0
		u.Sprite.Images.Back().Flags &^= ImageFlagYFrozen
	}
	if !f.uImmovable(u) {
		if !u.OrderQueue.Empty() && u.OrderQueue.Front().OrderType.ID == OrdersDroneBuild {
			f.moveUnit(u, u.OrderTarget.Pos)
		} else {
			f.setLastImageOffsetYIfModifier10(u, 7)
			f.queueOrderFront(u, f.getOrderType(OrdersResetCollision), OrderTarget{})
		}
	} else {
		f.removeOneOrder(u, f.getOrderType(OrdersDroneBuild))
		f.setLastImageOffsetYIfModifier10(u, 7)
		f.queueOrderFront(u, f.getOrderType(OrdersResetCollision), OrderTarget{})
	}
	f.uUnsetStatusFlag(u, UnitStatusFlagOrderNotInterruptible)
	f.orderDone(u, OrderTarget{})
}

func (f *StateFunctions) orderDroneBuild(u *Unit) {
	if u.BuildQueue.Empty() {
		throwError("order_DroneBuild: empty build queue")
	}
	buildType := u.BuildQueue.Front()
	done := func() {
		f.setLastImageOffsetYIfModifier10(u, 7)
		f.queueOrderFront(u, f.getOrderType(OrdersResetCollision), OrderTarget{})
		f.uUnsetStatusFlag(u, UnitStatusFlagOrderNotInterruptible)
		f.orderDone(u, OrderTarget{})
	}
	if u.Sprite.Position != u.OrderTarget.Pos {
		done()
		return
	}
	if !f.canPlaceBuilding(u, u.Owner, buildType, u.Sprite.Position, true, true) {
		done()
		return
	}
	if !f.hasAvailableSupplyFor(u.Owner, buildType, true) || !f.hasAvailableResourcesFor(u.Owner, buildType, true) {
		done()
		return
	}
	f.st.CurrentMinerals[u.Owner] -= buildType.MineralCost
	f.st.CurrentGas[u.Owner] -= buildType.GasCost

	var powerup *Unit
	if f.utWorker(u.UnitType) {
		powerup = u.Worker.Powerup
	}
	if powerup != nil {
		f.dropCarriedItems(u)
	}
	if f.unitIs(buildType, UnitTypesZergExtractor) {
		buildUnit := f.buildRefinery(u, buildType)
		if buildUnit != nil {
			u.UserActionFlags |= 4
			f.killUnit(u)
			f.zergBuildingStartConstruction(buildUnit)
			f.createImage(f.getImageType(ImageTypesIMAGEIDVespeneGeyser2), buildUnit.Sprite, Xy{}, imageOrderBelow, nil)
		} else {
			done()
		}
	} else {
		f.addCompletedUnit(u, -1, false)
		f.uUnsetStatusFlag(u, UnitStatusFlagCompleted)
		f.morphUnit(u, buildType)
		f.zergBuildingStartConstruction(u)
	}
	if powerup != nil && f.utPowerup(powerup.UnitType) {
		f.moveUnit(powerup, powerup.Building.Powerup.Origin)
	}
}

func (f *StateFunctions) orderIncompleteMorphing(u *Unit) {
	buildType := u.UnitType
	if !u.BuildQueue.Empty() {
		queuedType := u.BuildQueue.Front()
		if f.unitTypeIsMorphingBuilding(queuedType) {
			buildType = queuedType
		}
	}
	switch u.OrderState {
	case 0:
		if u.RemainingBuildTime < buildType.BuildTime*3/4 {
			u.OrderState++
		}
	case 1:
		if f.unitIs(u.UnitType, UnitTypesZergExtractor) {
			f.destroyImageFromToUnit(u, ImageTypesIMAGEIDVespeneGeyser2, ImageTypesIMAGEIDVespeneGeyser2)
		}
		f.spriteRunAnim(u.Sprite, IscriptAnimsSpecialState1)
		u.OrderState++
	case 2:
		if u.RemainingBuildTime < buildType.BuildTime/2 {
			u.OrderState++
		}
	case 3:
		f.spriteRunAnim(u.Sprite, IscriptAnimsSpecialState2)
		u.OrderState++
	case 4:
		if u.RemainingBuildTime == 0 {
			u.OrderState++
		}
	case 5:
		f.spriteRunAnim(u.Sprite, IscriptAnimsAlmostBuilt)
		u.OrderState++
	case 6:
		if u.OrderSignal&4 != 0 {
			u.OrderSignal &^= 4
			f.completeMorphingBuilding(u, buildType)
		}
		return
	}
	if u.RemainingBuildTime != 0 {
		f.countDownRemainingBuildTime(u)
	}
	if !f.unitTypeIsMorphingBuilding(buildType) {
		if f.st.CheatOperationCwal {
			f.setUnitHp(u, u.Hp.Add(u.HpConstructionRate.MulInt(16)))
		} else {
			f.setUnitHp(u, u.Hp.Add(u.HpConstructionRate))
		}
	}
}

func (f *StateFunctions) completeMorphingBuilding(u *Unit, buildType *UnitTypeT) {
	if f.unitTypeIsMorphingBuilding(buildType) {
		f.uSetStatusFlag(u, UnitStatusFlagCompleted)
		f.addCompletedUnit(u, -1, false)
		f.uUnsetStatusFlag(u, UnitStatusFlagCompleted)
		hp := u.Hp
		f.morphUnit(u, buildType)
		hp = hp.Add(buildType.Hitpoints)
		if u.PreviousUnitType != nil {
			hp = hp.Sub(u.PreviousUnitType.Hitpoints)
		}
		if hp.Less(Fp8Integer(1)) {
			hp = Fp8Integer(1)
		}
		f.setUnitHp(u, hp)
		u.RemainingBuildTime = 0
	}
	f.finishBuildingUnit(u)
	f.completeUnit(u)
	f.addCreepProvider(u)
	f.spriteRunAnim(u.Sprite, IscriptAnimsAlmostBuilt)

	f.setCreepBuildingTiles(buildType, u.Sprite.Position)
}

func (f *StateFunctions) orderZergBuildingMorph(u *Unit) {
	if u.BuildQueue.Empty() {
		throwError("order_ZergBuildingMorph: empty build queue")
	}
	buildType := u.BuildQueue.Front()
	if !f.hasAvailableSupplyFor(u.Owner, buildType, true) || !f.hasAvailableResourcesFor(u.Owner, buildType, true) {
		f.setQueuedOrder(u, false, u.UnitType.ReturnToIdle.Get(), OrderTarget{})
		return
	}
	f.st.CurrentMinerals[u.Owner] -= buildType.MineralCost
	f.st.CurrentGas[u.Owner] -= buildType.GasCost
	f.uUnsetStatusFlag(u, UnitStatusFlagCompleted)
	f.setConstructionGraphic(u, true)
	u.RemainingBuildTime = buildType.BuildTime
	f.spriteRunAnim(u.Sprite, IscriptAnimsSpecialState1)
	f.setUnitOrder(u, f.getOrderType(OrdersIncompleteMorphing))
}

func (f *StateFunctions) findCarryImage(u *Unit) *Image {
	carryImage := f.findImage(u.Sprite, ImageTypesIMAGEIDFlag, ImageTypesIMAGEIDTerranGasTankType2)
	if carryImage == nil {
		carryImage = f.findImage(u.Sprite, ImageTypesIMAGEIDUraj, ImageTypesIMAGEIDKhalis)
	}
	return carryImage
}

func (f *StateFunctions) orderBurrowing(u *Unit) {
	if u.OrderState == 0 {
		f.stopUnit(u)
		u.OrderState = 1
	}
	if u.OrderState == 1 {
		if f.uMovementFlag(&u.Flingy, 2) {
			if u.NextSpeed == Fp8Zero() {
				f.setUnitImmovable(u)
			}
		} else {
			f.spriteRunAnim(u.Sprite, IscriptAnimsWalkingToIdle)
			var dir int32 = 12
			if f.unitIsDefiler(u.UnitType) {
				dir = 11
			}
			if f.unitIs(u.UnitType, UnitTypesZergDrone) {
				dir = 13
			}
			f.setNextTargetWaypoint(&u.Flingy, f.toXy(XyFp8Add(u.ExactPosition, f.directionXYInt(directionLiteral(8).MulInt(dir), 16))))
			u.OrderState = 2
		}
	}
	if u.OrderState == 2 {
		if f.isFacingNextTargetWaypoint(u, Fp8Zero()) {
			f.spriteRunAnim(u.Sprite, IscriptAnimsBurrow)
			u.OrderState = 3
			f.playSoundUnit(16, u, false)
		}
	}
	if u.OrderState == 3 {
		if u.OrderSignal&4 != 0 {
			u.OrderSignal &^= 4
			f.completeBurrowing(u)
		}
	}
}

func (f *StateFunctions) completeBurrowing(u *Unit) {
	for image := range u.Sprite.Images.All() {
		if image.ImageType.AlwaysVisible {
			continue
		}
		f.hideImage(image)
	}
	carryImage := f.findCarryImage(u)
	if carryImage != nil {
		f.hideImage(carryImage)
	}
	f.uSetStatusFlag(u, UnitStatusFlagNoCollide)
	f.setUnitBurrowed(u)
	f.setSecondaryOrder(u, f.getOrderType(OrdersCloak))
	u.DetectedFlags = 0x80000000
	if u.DefensiveMatrixHp != Fp8Zero() {
		hp := u.DefensiveMatrixHp
		time := u.DefensiveMatrixTimer
		f.dealDefensiveMatrixDamage(u, hp)
		u.DefensiveMatrixHp = hp
		u.DefensiveMatrixTimer = time
	}
	if u.IrradiateTimer != 0 {
		f.destroyImageFromToUnit(u, ImageTypesIMAGEIDIrradiateSmall, ImageTypesIMAGEIDIrradiateLarge)
	}
	if u.EnsnareTimer != 0 {
		f.destroyImageFromToUnit(u, ImageTypesIMAGEIDEnsnareOverlaySmall, ImageTypesIMAGEIDEnsnareOverlayLarge)
	}
	if u.PlagueTimer != 0 {
		f.destroyImageFromToUnit(u, ImageTypesIMAGEIDPlagueOverlaySmall, ImageTypesIMAGEIDPlagueOverlayLarge)
	}
	f.setUnitOrder(u, f.getOrderType(OrdersBurrowed))
	f.uUnsetStatusFlag(u, UnitStatusFlagOrderNotInterruptible)
	f.orderDone(u, OrderTarget{})
}

func (f *StateFunctions) orderBurrowed(u *Unit) {
	if f.unitIs(u.UnitType, UnitTypesZergLurker) {
		f.uUnsetStatusFlag(u, UnitStatusFlagOrderNotInterruptible)
		f.orderDone(u, OrderTarget{})
	} else if f.unitIs(u.UnitType, UnitTypesZergHydralisk) && !u.OrderQueue.Empty() && u.OrderQueue.Front().OrderType.ID == OrdersZergUnitMorph {
		f.setUnitOrder(u, f.getOrderType(OrdersUnburrowing))
		f.queueOrder(u.OrderQueue.Next(u.OrderQueue.Begin()), u, f.getOrderType(OrdersZergUnitMorph), OrderTarget{})
		f.uUnsetStatusFlag(u, UnitStatusFlagOrderNotInterruptible)
		f.orderDone(u, OrderTarget{})
	}
}

func (f *StateFunctions) orderAttackFixedRange(u *Unit) {
	f.attackUnitReacquireTarget(u)
	target := u.OrderTarget.Unit
	if target != nil {
		u.OrderTarget.Pos = target.Sprite.Position
	}
	if !f.unitCanAttackTarget(u, target) || !f.unitTargetInWeaponMovementRange(u, target) {
		f.orderDone(u, OrderTarget{})
		return
	}
	f.setNextTargetWaypoint(&u.Flingy, u.OrderTarget.Pos)
	f.attackUnitFireWeapon(u)
}

func (f *StateFunctions) orderUnburrowing(u *Unit) {
	if u.OrderState == 0 {
		u.Sprite.ElevationLevel = u.UnitType.ElevationLevel
		f.uUnsetStatusFlag(u, UnitStatusFlagNoCollide)
		f.checkUnitCollision(u)
		u.Sprite.Flags &= SpriteFlagHidden
		f.setSecondaryOrder(u, f.getOrderType(OrdersNothing))
		f.spriteRunAnim(u.Sprite, IscriptAnimsUnBurrow)
		f.playSoundUnit(17, u, false)
		u.OrderState = 1
	} else if u.OrderState == 1 {
		u.DetectedFlags = 0x80000000
		u.SecondaryOrderTimer = 0
		u.OrderState = 2
	}
	if u.OrderState == 2 {
		if u.OrderSignal&4 != 0 {
			u.OrderSignal &^= 4
			f.completeUnburrowing(u)
		}
	} else if u.OrderState == 3 {
		if u.OrderQueue.Empty() {
			f.setQueuedOrder(u, false, u.UnitType.ReturnToIdle.Get(), OrderTarget{})
		}
		f.uUnsetStatusFlag(u, UnitStatusFlagOrderNotInterruptible)
		f.orderDone(u, OrderTarget{})
	}
}

func (f *StateFunctions) completeUnburrowing(u *Unit) {
	f.resetMovementState(u)
	f.uUnsetStatusFlag(u, UnitStatusFlagPassivelyCloaked)
	f.uUnsetStatusFlag(u, UnitStatusFlagBurrowed)
	u.Sprite.SpriteType = u.UnitType.Flingy.Get().Sprite.Get()
	f.applyUnitEffects(u)
	carryImage := f.findCarryImage(u)
	if carryImage != nil {
		f.showImage(carryImage)
	}
	for image := range u.Sprite.Images.All() {
		f.setImageHeading(image, u.Heading)
	}
	u.OrderState = 3
}

func (f *StateFunctions) orderBuildNydusExit(u *Unit) {
	if !f.canPlaceBuilding(u, u.Owner, f.getUnitType(UnitTypesZergNydusCanal), u.OrderTarget.Pos, true, false) {
		f.orderDone(u, OrderTarget{})
		return
	}
	exit := f.createUnitByID(UnitTypesZergNydusCanal, u.OrderTarget.Pos, u.Owner)
	if exit == nil {
		f.displayLastErrorForPlayer(u.Owner)
		f.orderDone(u, OrderTarget{})
		return
	}
	f.setConstructionGraphic(exit, true)
	exit.Hp = exit.UnitType.Hitpoints.DivInt(10)
	f.setUnitOrder(exit, f.getOrderType(OrdersIncompleteMorphing))
	f.addCreepProvider(exit)
	if f.unitIsNydus(u.UnitType) {
		u.Building.Nydus.Exit = exit
	}
	if f.unitIsNydus(exit.UnitType) {
		exit.Building.Nydus.Exit = u
	}
	f.orderDone(u, OrderTarget{})
}

func (f *StateFunctions) orderEnterNydusCanal(u *Unit) {
	target := u.OrderTarget.Unit
	if target == nil || !f.unitCanEnterNydus(u, target) {
		f.orderDone(u, OrderTarget{})
		return
	}
	var exit *Unit
	if f.unitIsNydus(target.UnitType) {
		exit = target.Building.Nydus.Exit
	}
	if exit == nil {
		exit = target
	}
	if u.OrderState == 0 {
		f.moveToTargetReset(u, target)
		u.OrderState = 1
	} else if u.OrderState == 1 {
		if f.unitIsAtMoveTarget(&u.Flingy) {
			if f.uImmovable(u) {
				f.orderDone(u, OrderTarget{})
			} else {
				u.OrderState = 2
			}
		}
	}
	if u.OrderState == 2 {
		f.hideUnit(u, false)
		prevPos := u.Sprite.Position
		f.moveUnit(u, exit.Sprite.Position)
		res, pos := f.findUnitPlacement(u, exit.Sprite.Position, false)
		if res {
			f.playSoundUnit(19, target, false)
			f.moveUnit(u, pos)
			f.refreshUnitPosition(u)
			f.orderDone(u, OrderTarget{})
			f.playSoundUnit(19, exit, false)
		} else {
			f.moveUnit(u, prevPos)
			f.refreshUnitPosition(u)
		}
	}
}

func (f *StateFunctions) orderCastInfestation(u *Unit) {
	target := u.OrderTarget.Unit
	if u.OrderState == 0 {
		if target == nil {
			f.orderDone(u, OrderTarget{})
			return
		}
		if !f.unitCanBeInfested(target) {
			f.orderDone(u, OrderTarget{})
			return
		}
		f.moveToTargetReset(u, target)
		u.OrderState = 1
	} else if u.OrderState == 1 {
		if f.unitIsAtMoveTarget(&u.Flingy) {
			if target != nil && f.unitCanBeInfested(target) && !f.uImmovable(u) {
				f.setUnitOrderUnit(u, f.getOrderType(OrdersInfestingCommandCenter), target)
			} else {
				f.orderDone(u, OrderTarget{})
			}
		} else if target != nil {
			f.tryFollowUnit(u, target)
		}
	}
}

func (f *StateFunctions) finishHiddenOrder(u *Unit) {
	if u.OrderQueue.Empty() {
		f.setQueuedOrder(u, false, u.UnitType.ReturnToIdle.Get(), OrderTarget{})
	}
	f.uUnsetStatusFlag(u, UnitStatusFlagOrderNotInterruptible)
	f.orderDone(u, OrderTarget{})
}

func (f *StateFunctions) orderInfestingCommandCenter(u *Unit) {
	target := u.OrderTarget.Unit
	if u.OrderState == 0 {
		if target != nil && f.unitCanBeInfested(target) && target.OrderType.ID != OrdersInfestedCommandCenter {
			f.hideUnit(u, true)
			f.setUnitOrderUnit(target, f.getOrderType(OrdersInfestedCommandCenter), u)
			u.RemainingBuildTime = 60
			u.OrderState = 3
		} else {
			f.stopUnit(u)
			f.setNextTargetWaypoint(&u.Flingy, u.MoveTarget.Pos)
			f.finishHiddenOrder(u)
		}
	} else if u.OrderState == 3 {
		if u.RemainingBuildTime != 0 {
			u.RemainingBuildTime--
			return
		}
		res, pos := f.findUnitPlacement(u, u.Sprite.Position, false)
		if res {
			f.moveUnit(u, pos)
			f.stopUnit(u)
			f.setNextTargetWaypoint(&u.Flingy, u.MoveTarget.Pos)
			f.showUnit(u)
			f.finishHiddenOrder(u)
		}
	}
}

func (f *StateFunctions) orderInfestedCommandCenter(u *Unit) {
	queen := u.OrderTarget.Unit
	if queen == nil || !f.unitCanBeInfested(u) {
		f.uUnsetStatusFlag(u, UnitStatusFlagOrderNotInterruptible)
		f.orderDone(u, OrderTarget{})
		return
	}
	if u.OrderState == 0 {
		u.RemainingBuildTime = 3
		u.OrderState = 2
		return
	}
	if u.RemainingBuildTime != 0 {
		u.RemainingBuildTime--
		return
	}
	u.RemainingBuildTime = 0xffff
	if f.unitAddon(u) != nil {
		f.buildingAbandonAddon(u)
	}
	if f.unitIsFactory(u) {
		u.Building.Rally.Unit = u
	}
	newType := u.UnitType.InfestationUnit.Get()
	if !f.unitsShareUnions(u.UnitType, newType) {
		throwError("order_InfestedCommandCenter: can't change unit type due to unions mismatch")
	}
	f.incrementUnitCounts(u, -1)
	if f.uCompleted(u) {
		f.addCompletedUnit(u, -1, false)
	}
	u.UnitType = newType
	f.incrementUnitCounts(u, 1)
	if f.uCompleted(u) {
		f.addCompletedUnit(u, 1, false)
	}
	f.setUnitOwner(u, queen.Owner, true)
	f.setSpriteOwner(u, queen.Owner)
	u.SecondaryOrderType = nil
	f.setSecondaryOrder(u, f.getOrderType(OrdersNothing))
	f.cancelBuildQueue(u)
	for !u.OrderQueue.Empty() {
		f.removeQueuedOrder(u, u.OrderQueue.Front())
	}
	if f.uGroundedBuilding(u) {
		f.spriteRunAnim(u.Sprite, IscriptAnimsWorkingToIdle)
	}
	if u.UnitType.ConstructionAnimation.Valid() {
		f.createImage(u.UnitType.ConstructionAnimation.Get(), u.Sprite, Xy{}, imageOrderAbove, nil)
	}
	f.setUnitHp(u, u.UnitType.Hitpoints)
	f.uUnsetStatusFlag(u, UnitStatusFlagOrderNotInterruptible)
	f.orderDone(u, OrderTarget{})
}

func (f *StateFunctions) startSuicide(u *Unit) {
	u.OrderTarget.Unit = u
	f.spriteRunAnim(u.Sprite, IscriptAnimsSpecialState1)
	f.uSetStatusFlag(u, UnitStatusFlagOrderNotInterruptible)
	u.OrderState = 2
}

func (f *StateFunctions) explodeOnSignal(u *Unit) {
	if u.OrderSignal&1 != 0 {
		f.uSetStatusFlag(u, UnitStatusFlagLifetimeExpired)
		f.killUnit(u)
	}
}

func (f *StateFunctions) chaseTarget(u *Unit, target *Unit) {
	if f.unitTargetInRange(u, target, 256) && f.uMovementFlag(&target.Flingy, 2) {
		f.crappyMoveToTarget(u, target)
	} else {
		f.tryFollowUnit(u, target)
	}
}

func (f *StateFunctions) orderSuicideUnit(u *Unit) {
	target := u.OrderTarget.Unit
	if target == nil {
		f.orderDone(u, OrderTarget{})
		return
	}
	if !f.unitCanAttackTarget(u, target) {
		f.setUnitOrderPos(u, f.getOrderType(OrdersMove), target.Sprite.Position)
		return
	}
	if u.OrderState == 0 {
		f.moveToTarget(u, target)
		u.OrderState = 1
	}
	if u.OrderState == 1 {
		if (f.unitIsAtMoveTarget(&u.Flingy) && f.uImmovable(u)) || f.unitTargetInRange(u, target, 4) {
			f.stopUnit(u)
			if !f.unitTargetInRange(u, target, f.getWeaponType(WeaponTypesSuicideInfestedTerran).OuterSplashRadius) {
				f.orderDone(u, OrderTarget{})
				return
			}
			f.startSuicide(u)
		} else {
			f.chaseTarget(u, target)
		}
	}
	if u.OrderState == 2 {
		f.explodeOnSignal(u)
	}
}

func (f *StateFunctions) orderSuicideLocation(u *Unit) {
	if u.OrderState == 0 {
		f.setUnitMoveTarget(u, u.OrderTarget.Pos)
		f.setNextTargetWaypoint(&u.Flingy, u.OrderTarget.Pos)
		u.OrderState = 1
	}
	if u.OrderState == 1 {
		if f.unitIsAtMoveTarget(&u.Flingy) {
			if f.uImmovable(u) && f.lengthFromUnitToPos(u, u.OrderTarget.Pos) > f.getWeaponType(WeaponTypesSuicideInfestedTerran).OuterSplashRadius {
				f.orderDone(u, OrderTarget{})
				return
			}
			f.startSuicide(u)
		}
	}
	if u.OrderState == 2 {
		f.explodeOnSignal(u)
	}
}

func (f *StateFunctions) orderSuicideHoldPosition(u *Unit) {
	if u.OrderState == 0 {
		f.stopUnit(u)
		f.setNextTargetWaypoint(&u.Flingy, u.MoveTarget.Pos)
		u.OrderState = 1
	}
	if !u.OrderQueue.Empty() {
		f.activateNextOrder(u)
	}
}

func (f *StateFunctions) orderQueenHoldPosition(u *Unit) {
	f.orderSuicideHoldPosition(u)
}

func (f *StateFunctions) orderPlaceProtossBuilding(u *Unit) {
	if u.OrderState == 0 {
		f.setUnitMoveTarget(u, u.OrderTarget.Pos)
		f.setNextTargetWaypoint(&u.Flingy, u.OrderTarget.Pos)
		u.OrderState = 1
	} else if u.OrderState == 1 || u.OrderState == 2 {
		if u.BuildQueue.Empty() {
			throwError("order_PlaceProtossBuilding: empty build queue")
		}
		buildType := u.BuildQueue.Front()
		if u.OrderState == 2 && f.uMovementFlag(&u.Flingy, 2) {
			return
		}
		var d int32 = 70
		if f.unitIs(buildType, UnitTypesProtossAssimilator) {
			d += 18
		}
		if f.lengthFromUnitToPos(u, u.OrderTarget.Pos) <= d {
			if u.OrderState == 1 {
				f.stopUnit(u)
				f.setNextTargetWaypoint(&u.Flingy, u.MoveTarget.Pos)
				u.OrderState = 2
			} else {
				f.warpInProtossBuilding(u, buildType)
			}
		} else if f.unitIsAtMoveTarget(&u.Flingy) {
			f.setUnitMoveTarget(u, u.OrderTarget.Pos)
			f.setNextTargetWaypoint(&u.Flingy, u.OrderTarget.Pos)
		}
	} else if u.OrderState == 3 {
		if u.OrderTarget.Unit != nil && f.unitIs(u.OrderTarget.Unit.UnitType, UnitTypesProtossAssimilator) && u.OrderQueue.Empty() {
			f.setUnitOrderUnit(u, f.getOrderType(OrdersWaitForGas), u.OrderTarget.Unit)
		} else {
			f.orderDone(u, OrderTarget{})
		}
		f.spriteRunAnim(u.Sprite, IscriptAnimsWalkingToIdle)
	}
}

func (f *StateFunctions) canWarpInProtossBuilding(u *Unit, buildType *UnitTypeT) bool {
	if !f.playerHasSupplyAndResourcesFor(u.Owner, buildType, true) {
		return false
	}
	if !f.canPlaceBuilding(u, u.Owner, buildType, u.OrderTarget.Pos, true, false) {
		return false
	}
	return true
}

func (f *StateFunctions) warpInProtossBuilding(u *Unit, buildType *UnitTypeT) {
	if !f.canWarpInProtossBuilding(u, buildType) {
		f.orderDone(u, OrderTarget{})
		return
	}
	f.st.CurrentMinerals[u.Owner] -= buildType.MineralCost
	f.st.CurrentGas[u.Owner] -= buildType.GasCost
	var buildUnit *Unit
	if buildType.ID == UnitTypesProtossAssimilator {
		buildUnit = f.buildRefinery(u, buildType)
	} else {
		buildUnit = f.createUnit(buildType, u.OrderTarget.Pos, u.Owner)
	}
	f.buildQueueClearSingle(u)
	if buildUnit == nil {
		f.displayLastErrorForPlayer(u.Owner)
		f.orderDone(u, OrderTarget{})
		return
	}
	f.replaceSpriteImages(buildUnit.Sprite, f.getImageType(ImageTypesIMAGEIDWarpAnchor), directionLiteral(0))
	if f.unitIs(buildType, UnitTypesProtossAssimilator) {
		f.createImage(f.getImageType(ImageTypesIMAGEIDVespeneGeyser), buildUnit.Sprite, Xy{}, imageOrderBelow, nil)
	}
	f.setUnitOrder(buildUnit, f.getOrderType(OrdersIncompleteWarping))
	u.OrderTarget.Unit = buildUnit
	u.OrderState = 3
	f.playSoundUnit(528, u, false)
}

func (f *StateFunctions) orderIncompleteWarping(u *Unit) {
	switch u.OrderState {
	case 0:
		if u.RemainingBuildTime == 0 {
			f.spriteRunAnim(u.Sprite, IscriptAnimsSpecialState1)
			f.playSoundUnit(529, u, false)
			u.OrderState = 1
		}
	case 1:
		if u.OrderSignal&1 != 0 {
			u.OrderSignal &^= 1
			f.startWarpFlash(u)
			u.OrderState = 2
		}
		return
	case 2:
		if u.OrderSignal&1 != 0 {
			u.OrderSignal &^= 1
			f.replaceSpriteImages(u.Sprite, u.Sprite.SpriteType.Image.Get(), directionLiteral(0))
			f.spriteRunAnim(u.Sprite, IscriptAnimsWarpIn)
			u.OrderState = 3
		}
		return
	case 3:
		if ^u.OrderSignal&1 != 0 {
			return
		}
		u.OrderSignal &^= 1
		f.finishBuildingUnit(u)
		f.completeUnit(u)
		f.checkUnitCollision(u)
		if f.uDisabled(u) {
			f.spriteRunAnim(u.Sprite, IscriptAnimsDisable)
		}
	}
	if f.st.CheatOperationCwal {
		if u.RemainingBuildTime > 16 {
			u.RemainingBuildTime -= 16
		} else {
			u.RemainingBuildTime = 0
		}
		f.setUnitHp(u, u.Hp.Add(u.HpConstructionRate.MulInt(16)))
		f.setUnitShieldPoints(u, u.ShieldPoints.Add(u.ShieldConstructionRate.MulInt(16)))
	} else {
		if u.RemainingBuildTime != 0 {
			u.RemainingBuildTime--
		}
		f.setUnitHp(u, u.Hp.Add(u.HpConstructionRate))
		f.setUnitShieldPoints(u, u.ShieldPoints.Add(u.ShieldConstructionRate))
	}
}

func (f *StateFunctions) startWarpFlash(u *Unit) {
	f.replaceSpriteImages(u.Sprite, u.Sprite.SpriteType.Image.Get(), directionLiteral(0))

	f.iscriptSetScript(u.Sprite.MainImage, 193)
	f.iscriptRunAnim(u.Sprite.MainImage, IscriptAnimsInit)
	f.iscriptSetScript(u.Sprite.MainImage, u.Sprite.MainImage.ImageType.IscriptID)
	f.setImageModifier(u.Sprite.MainImage, 12)
	f.iscriptExecute(u.Sprite.MainImage, &u.Sprite.MainImage.IscriptState, false, nil, false)
}

func (f *StateFunctions) isInPsionicMatrixUnits(u *Unit) bool {
	return u.Building.Pylon.PsionicMatrixLink.Prev != nil || f.st.PsionicMatrixUnits.Front() == u
}

func (f *StateFunctions) orderInitializePsiProvider(u *Unit) {
	if !f.unitIs(u.UnitType, UnitTypesProtossPylon) {
		throwError("order_InitializePsiProvider: unit is not a pylon")
	}
	if !f.isInPsionicMatrixUnits(u) {
		f.st.PsionicMatrixUnits.PushFront(u)
		s := f.createSprite(f.getSpriteType(SpriteTypesSPRITEIDPsiFieldRightUpper), u.Sprite.Position, 0)
		u.Building.Pylon.PsiFieldSprite = s
		if s != nil {
			s.Flags |= SpriteFlagHidden
			f.setSpriteVisibility(s, 0)
		}
		f.st.UpdatePsionicMatrix = true
	}
	f.setUnitOrder(u, u.UnitType.ReturnToIdle.Get())
	f.uUnsetStatusFlag(u, UnitStatusFlagOrderNotInterruptible)
	f.orderDone(u, OrderTarget{})
}

func (f *StateFunctions) orderInitializeArbiter(u *Unit) {
	if u.OrderQueue.Empty() {
		f.setUnitOrder(u, f.getUnitType(UnitTypesProtossArbiter).ReturnToIdle.Get())
	}
	if !f.uHallucination(u) {
		f.setSecondaryOrder(u, f.getOrderType(OrdersCloakNearbyUnits))
	}
	f.uUnsetStatusFlag(u, UnitStatusFlagOrderNotInterruptible)
	f.orderDone(u, OrderTarget{})
}

func (f *StateFunctions) stopAndFinishOrder(u *Unit) {
	f.stopUnit(u)
	f.setNextTargetWaypoint(&u.Flingy, u.MoveTarget.Pos)
	f.orderDone(u, OrderTarget{})
}

func archonMergeDistance(isDarkArchon bool) int32 {
	if isDarkArchon {
		return 20
	}
	return 2
}

func archonApproachDistance(isDarkArchon bool) int32 {
	if isDarkArchon {
		return 19
	}
	return 10
}

func (f *StateFunctions) orderArchonWarp(u *Unit) {
	isDarkArchon := u.OrderType.ID == OrdersDarkArchonMeld
	target := u.OrderTarget.Unit
	if target == nil || target.Owner != u.Owner || target.OrderType != u.OrderType || target.OrderTarget.Unit != u {
		f.stopAndFinishOrder(u)
		return
	}
	if f.uGroundUnit(u) {
		if f.unitTargetInRange(u, target, u.NextSpeed.MulInt(2).IntegerPart()) {
			f.uUnsetStatusFlag(u, UnitStatusFlagGroundUnit)
			f.queueOrderFront(u, f.getOrderType(OrdersResetCollision), OrderTarget{})
		}
	}
	d := f.xyLength(XySub(u.Sprite.Position, target.Sprite.Position))
	if d <= archonMergeDistance(isDarkArchon) {
		f.mergeArchon(u, target, isDarkArchon)
		return
	}
	if u.OrderState != 0 && f.unitIsAtMoveTarget(&u.Flingy) && f.uImmovable(u) {
		f.stopAndFinishOrder(u)
		return
	}
	u.OrderState = 1
	if u.MoveTargetTimer == 0 {
		if d <= archonApproachDistance(isDarkArchon) {
			f.setUnitMoveTarget(u, XyDivScalar(XyAdd(u.Sprite.Position, target.Sprite.Position), 2))
		} else {
			f.setUnitMoveTarget(u, target.Sprite.Position)
		}
	}
}

func (f *StateFunctions) mergeArchon(u *Unit, target *Unit, isDarkArchon bool) {
	f.mergeStatusEffects(u, target)
	archonType := UnitTypesProtossArchon
	if isDarkArchon {
		archonType = UnitTypesProtossDarkArchon
	}
	f.morphUnit(u, f.getUnitType(archonType))
	if isDarkArchon {
		f.decloakUnit(u)
	}
	u.Sprite.Flags &^= SpriteFlagIscriptNobrk
	f.spriteRunAnim(u.Sprite, IscriptAnimsSpecialState1)
	u.KillCount += target.KillCount
	if isDarkArchon {
		u.Energy = Fp8Integer(50)
	}
	target.UserActionFlags |= 4
	f.killUnit(target)
	f.uUnsetStatusFlag(u, UnitStatusFlagCanMove)
	f.resetMovementState(u)
	f.setUnitOrder(u, f.getOrderType(OrdersCompletingArchonSummon))
}

func (f *StateFunctions) orderCompletingArchonSummon(u *Unit) {
	if u.OrderState == 0 {
		if u.RemainingBuildTime != 0 {
			f.countDownRemainingBuildTime(u)
		} else {
			u.Sprite.Flags &^= SpriteFlagIscriptNobrk
			f.spriteRunAnim(u.Sprite, IscriptAnimsInit)
			u.OrderState = 1
		}
	} else if u.OrderState == 1 {
		if u.OrderSignal&4 != 0 {
			u.OrderSignal &^= 4
			f.uSetStatusFlag(u, UnitStatusFlagCanMove)
			f.resetMovementState(u)
			f.setQueuedOrder(u, false, u.UnitType.HumanAiIdle.Get(), OrderTarget{})
			f.uUnsetStatusFlag(u, UnitStatusFlagOrderNotInterruptible)
			f.orderDone(u, OrderTarget{})
		}
	}
}

func (f *StateFunctions) orderCastRecall(u *Unit) {
	if u.OrderState == 0 {
		f.startRecall(u)
	} else if u.OrderState == 1 && u.MainOrderTimer == 0 {
		nRecalled := 0
		for _, target := range f.recallTargets(u) {
			if f.recallUnit(u, target) {
				nRecalled++
			}
		}
		if nRecalled != 0 {
			f.playSoundUnit(552+f.lcgRand(18)%2, u, false)
		}
		f.orderDone(u, OrderTarget{})
	}
}

func (f *StateFunctions) startRecall(u *Unit) {
	if u.Energy.Less(Fp8Integer(f.getTechType(TechTypesRecall).EnergyCost)) {
		f.orderDone(u, OrderTarget{})
		return
	}
	u.Energy = u.Energy.Sub(Fp8Integer(f.getTechType(TechTypesRecall).EnergyCost))
	if u.OrderTarget.Unit != nil {
		u.OrderTarget.Pos = u.OrderTarget.Unit.Sprite.Position
	}
	f.createVisibleThingyAbove(f.getSpriteType(SpriteTypesSPRITEIDRecallField), u.OrderTarget.Pos, 0, u.Sprite.ElevationLevel)
	f.playSoundAt(550+f.lcgRand(17)%2, u.OrderTarget.Pos, nil, false)
	u.MainOrderTimer = 22
	u.OrderState = 1
}

func (f *StateFunctions) isRecallable(u *Unit, target *Unit) bool {
	if target == u {
		return false
	}
	if target.Owner != u.Owner {
		return false
	}
	if f.usHidden(&target.Thingy) {
		return false
	}
	if f.utBuilding(target.UnitType) {
		return false
	}
	if f.uInvincible(target) {
		return false
	}
	if f.uHallucination(target) {
		return false
	}
	if f.uBurrowed(target) {
		return false
	}
	if f.unitIs(target.UnitType, UnitTypesZergLarva) {
		return false
	}
	if f.unitIs(target.UnitType, UnitTypesZergEgg) {
		return false
	}
	if f.unitIs(target.UnitType, UnitTypesZergLurkerEgg) {
		return false
	}
	return true
}

func (f *StateFunctions) recallTargets(u *Unit) []*Unit {
	var targets []*Unit
	for target := range f.findUnitsNoexpand(f.squareAt(u.OrderTarget.Pos, 64)) {
		if f.isRecallable(u, target) {
			targets = append(targets, target)
		}
	}
	return targets
}

func (f *StateFunctions) recallUnit(u *Unit, target *Unit) bool {
	targetPos := target.Sprite.Position
	f.moveUnit(target, u.Sprite.Position)
	placed, placement := f.findUnitPlacement(target, u.Sprite.Position, false)
	if !placed {
		f.moveUnit(target, targetPos)
		return false
	}

	for n := range f.st.VisibleUnits.All() {
		f.removeTargetReferences(n, target)
	}
	for n := range f.st.ActiveBullets.All() {
		f.removeBulletTargetReferences(n, target)
	}
	f.unitFinderRemove(target)
	if f.uGroundedBuilding(target) {
		f.setUnitTilesUnoccupied(target.UnitType, target.Sprite.Position)
	}
	if f.uFlying(target) {
		f.decrementRepulseField(target)
	}
	f.resetMovementState(target)

	f.moveUnit(target, placement)
	f.refreshUnitPosition(target)
	if !f.unitIs(target.UnitType, UnitTypesZergCocoon) {
		f.setUnitOrder(target, target.UnitType.ReturnToIdle.Get())
	}
	f.createVisibleThingyAbove(f.getSpriteType(SpriteTypesSPRITEIDRecallField), placement, 0, target.Sprite.ElevationLevel)
	if f.unitIsGhost(target.UnitType) && target.ConnectedUnit != nil && f.unitIs(target.ConnectedUnit.UnitType, UnitTypesTerranNuclearMissile) {
		target.ConnectedUnit.ConnectedUnit = nil
		target.ConnectedUnit = nil
	}
	return true
}

func (f *StateFunctions) orderCarrier(u *Unit) {
	if f.carrierReaverAttack(u, 32*f.unitTargetAcquisitionRange(u), 0) {
		fighter := f.releaseFighter(u)
		if fighter != nil {
			fighter.ShieldPoints = Fp8Integer(fighter.UnitType.ShieldPoints)
			fighter.Sprite.ElevationLevel = u.Sprite.ElevationLevel - 1
			f.setUnitOrderUnit(fighter, fighter.UnitType.AttackUnit.Get(), u.OrderTarget.Unit)
			u.MainOrderTimer = 7
		}
	}
}

func (f *StateFunctions) orderCarrierStop(u *Unit) {
	f.stopUnit(u)
	f.setNextTargetWaypoint(&u.Flingy, u.MoveTarget.Pos)
	f.returnInterceptors(u)
	f.setUnitOrderUnit(u, f.getOrderType(OrdersCarrier), u.OrderTarget.Unit)
}

func (f *StateFunctions) orderCarrierAttack(u *Unit) {
	target := u.OrderTarget.Unit
	if target == nil {
		f.orderDone(u, OrderTarget{})
		return
	}
	if !f.unitCanAttackTarget(u, target) {
		f.setUnitOrderPos(u, f.getOrderType(OrdersMove), u.OrderTarget.Pos)
		return
	}
	readyToAttack := f.uReadyToAttack(u)
	flag8 := f.uStatusFlag(u, UnitStatusFlag8)
	f.interceptorsAttack(u)
	f.queueOrderFront(u, f.getOrderType(OrdersCarrierFight), OrderTarget{Unit: target})
	f.activateNextOrder(u)
	f.uSetStatusFlagValue(u, UnitStatusFlagReadyToAttack, readyToAttack)
	f.uSetStatusFlagValue(u, UnitStatusFlag8, flag8)
}

func (f *StateFunctions) setInterceptorMoveTarget(u *Unit, mult int32) {
	x := f.lcgRandRange(11, -127, 128) * mult
	y := f.lcgRandRange(11, -127, 128) * mult
	pos := XyAdd(u.OrderTarget.Pos, Xy{X: x, Y: y})
	dims := u.UnitType.Dimensions
	if pos.X < dims.From.X || pos.X >= int32(f.gameSt.MapWidth)-dims.To.X {
		pos.X = u.OrderTarget.Pos.X - x
	}
	if pos.Y < dims.From.Y || pos.Y >= int32(f.gameSt.MapHeight)-dims.To.Y {
		pos.Y = u.OrderTarget.Pos.Y - y
	}
	pos = f.restrictMoveTargetToValidBounds(u.UnitType, pos)
	f.setUnitMoveTarget(u, pos)
	f.setNextTargetWaypoint(&u.Flingy, pos)
}

func (f *StateFunctions) orderInterceptorAttack(u *Unit) {
	if !f.unitIsFighter(u.UnitType) {
		throwError("order_InterceptorAttack: unit is not a fighter")
	}
	parent := u.Fighter.Parent
	if parent != nil && u.ShieldPoints.Less(Fp8Integer(u.UnitType.ShieldPoints).DivInt(4)) {
		f.setUnitOrder(u, f.getOrderType(OrdersInterceptorReturn))
		return
	}
	target := u.OrderTarget.Unit
	if target == nil || f.uInvincible(target) {
		if !f.unitAutoattack(u) {
			f.setUnitOrder(u, f.getOrderType(OrdersInterceptorReturn))
			return
		}
		target = u.OrderTarget.Unit
	}
	u.OrderTarget.Pos = target.Sprite.Position
	if parent != nil {
		rangeValue := 32 * (f.unitTargetAcquisitionRange(parent) + 2)
		if !f.unitTargetInRange(parent, target, rangeValue+f.unitTargetMovementRange(&parent.Flingy, &target.Flingy)) {
			f.setUnitOrder(u, f.getOrderType(OrdersInterceptorReturn))
			return
		}
	}

	if u.OrderState == 0 {
		if parent == nil {
			f.killUnit(u)
			return
		}
		f.playSoundUnit(616, u, false)
		f.setInterceptorMoveTarget(u, 3)
		u.MainOrderTimer = 15
		u.OrderState = 1
	} else if u.OrderState == 1 {
		if u.MainOrderTimer == 0 {
			u.Sprite.ElevationLevel += 2
			u.OrderState = 3
		}
	}
	switch u.OrderState {
	case 3:
		if f.unitTargetInWeaponMovementRange(u, target) {
			f.attackUnitFireWeapon(u)
		}
		if f.unitTargetInRange(u, target, 50+f.unitTargetMovementRange(&u.Flingy, &target.Flingy)) {
			f.setInterceptorMoveTarget(u, 1)
			u.OrderState = 4
		} else {
			f.setUnitMoveTarget(u, u.OrderTarget.Pos)
			f.setNextTargetWaypoint(&u.Flingy, u.OrderTarget.Pos)
		}
	case 4:
		if f.lengthFromUnitToPos(u, u.MoveTarget.Pos) <= 50 {
			f.setInterceptorMoveTarget(u, 2)
			u.OrderState = 5
		}
	case 5:
		if f.lengthFromUnitToPos(u, u.MoveTarget.Pos) <= 50 {
			f.setUnitMoveTarget(u, u.OrderTarget.Pos)
			f.setNextTargetWaypoint(&u.Flingy, u.OrderTarget.Pos)
			u.OrderState = 3
		}
	}
}

func (f *StateFunctions) orderInterceptorReturn(u *Unit) {
	if !f.unitIsFighter(u.UnitType) {
		throwError("order_InterceptorReturn: unit is not a fighter")
	}
	parent := u.Fighter.Parent
	if parent == nil {
		f.killUnit(u)
		return
	}
	if u.OrderState == 0 {
		if f.lengthFromUnitToPos(u, parent.Sprite.Position) <= 60 {
			u.OrderState = 1
			if f.unitIs(u.UnitType, UnitTypesProtossInterceptor) {
				u.Sprite.ElevationLevel = parent.Sprite.ElevationLevel - 2
			}
		}
	}
	if f.lengthFromUnitToPos(u, parent.Sprite.Position) <= 10 {
		f.dockFighter(parent, u)
	} else {
		f.setUnitMoveTarget(u, parent.Sprite.Position)
		f.setNextTargetWaypoint(&u.Flingy, parent.Sprite.Position)
	}
}

func (f *StateFunctions) dockFighter(parent *Unit, u *Unit) {
	if f.unitIsCarrier(parent.UnitType) {
		parent.Carrier.OutsideUnits.Remove(u)
		parent.Carrier.OutsideCount--
		parent.Carrier.InsideUnits.PushFront(u)
		parent.Carrier.InsideCount++
	} else if f.unitIsReaver(parent.UnitType) {
		parent.Reaver().OutsideUnits.Remove(u)
		parent.Reaver().OutsideCount--
		parent.Reaver().InsideUnits.PushFront(u)
		parent.Reaver().InsideCount++
	}
	u.Fighter.IsOutside = false
	f.hideUnit(u, true)
	f.setUnitOrder(u, f.getOrderType(OrdersNothing))
}

func (f *StateFunctions) orderDarkArchonMeld(u *Unit) {
	f.orderArchonWarp(u)
}

func (f *StateFunctions) orderCastMindControl(u *Unit) {
	target := u.OrderTarget.Unit
	if target == nil || target.Owner == u.Owner {
		f.orderDone(u, OrderTarget{})
		return
	}
	tech := f.getTechType(TechTypesMindControl)
	if !f.spellCastTargetMovement(u, tech, f.unitSightRange(u, true)) {
		return
	}
	if f.uHallucination(target) {
		f.killUnit(target)
	} else {
		f.createSizedImage(target, ImageTypesIMAGEIDMindControlHitSmall, true, imageOrderTop)
		f.triggerGiveUnitTo(target, u.Owner)
		if f.unitIs(target.UnitType, UnitTypesProtossDarkArchon) {
			target.Energy = Fp8Zero()
		}
		f.orderDone(target, OrderTarget{})
	}
	u.Energy = u.Energy.Sub(Fp8Integer(tech.EnergyCost))
	u.ShieldPoints = Fp8Zero()
	f.orderDone(u, OrderTarget{})
	f.playSoundUnit(1062, target, false)
}

func (f *StateFunctions) orderReaver(u *Unit) {
	if !f.unitIsReaver(u.UnitType) {
		throwError("order_Reaver: unit is not a reaver")
	}
	rangeValue := 32 * f.unitTargetAcquisitionRange(u)
	if !f.carrierReaverAttack(u, rangeValue, rangeValue*2) {
		return
	}
	f.setNextTargetWaypoint(&u.Flingy, u.OrderTarget.Pos)
	if f.directionIndex(u.Heading)-f.directionIndex(f.xyDirection(XySub(u.NextTargetWaypoint, u.Sprite.Position))) <= 1 {
		if !f.uCannotAttack(u) && u.Reaver().OutsideCount == 0 {
			fighter := f.releaseFighter(u)
			if fighter != nil {
				f.setUnitOrderUnit(fighter, fighter.UnitType.AttackUnit.Get(), u.OrderTarget.Unit)
				u.MainOrderTimer = 60
			}
		}
	} else {
		u.OrderProcessTimer = 0
	}
}

func (f *StateFunctions) orderReaverAttack(u *Unit) {
	target := u.OrderTarget.Unit
	if target == nil {
		f.orderDone(u, OrderTarget{})
		return
	}
	if !f.unitCanAttackTarget(u, target) {
		f.setUnitOrderPos(u, f.getOrderType(OrdersMove), u.OrderTarget.Pos)
		return
	}
	readyToAttack := f.uReadyToAttack(u)
	flag8 := f.uStatusFlag(u, UnitStatusFlag8)
	f.queueOrderFront(u, f.getOrderType(OrdersReaverFight), OrderTarget{Unit: target})
	f.activateNextOrder(u)
	f.uSetStatusFlagValue(u, UnitStatusFlagReadyToAttack, readyToAttack)
	f.uSetStatusFlagValue(u, UnitStatusFlag8, flag8)
}

func (f *StateFunctions) orderScarabAttack(u *Unit) {
	target := u.OrderTarget.Unit
	if u.OrderState == 0 {
		f.setRemoveTimer(u, 90)
		f.uSetStatusFlag(u, UnitStatusFlagNoCollide)
		u.MainOrderTimer = 7
		if target != nil {
			f.moveToTarget(u, target)
		} else {
			f.setUnitMoveTarget(u, u.OrderTarget.Pos)
			f.setNextTargetWaypoint(&u.Flingy, u.OrderTarget.Pos)
		}
		u.OrderState = 2
	} else if u.OrderState == 2 || u.OrderState == 3 {
		if u.OrderState == 2 {
			if u.MainOrderTimer == 0 {
				f.uUnsetStatusFlag(u, UnitStatusFlagNoCollide)
				u.OrderState = 3
			}
		}
		if f.scarabIsAtTarget(u, target) {
			u.OrderTarget.Pos = u.Sprite.Position
			f.spriteRunAnim(u.Sprite, IscriptAnimsSpecialState1)
			u.OrderState = 6
		}
	} else if u.OrderState == 6 {
		f.explodeOnSignal(u)
	}
}

func (f *StateFunctions) scarabIsAtTarget(u *Unit, target *Unit) bool {
	if target == nil {
		return f.unitIsAtMoveTarget(&u.Flingy)
	}
	f.chaseTarget(u, target)
	w := u.UnitType.GroundWeapon.Get()
	if w == nil {
		throwError("order_ScarabAttack: no ground weapon")
	}
	return f.unitTargetInRange(u, target, w.InnerSplashRadius/2)
}

func (f *StateFunctions) orderReaverStop(u *Unit) {
	if !f.unitIsReaver(u.UnitType) {
		throwError("order_ReaverStop: unit is not a reaver")
	}
	f.stopUnit(u)
	f.setNextTargetWaypoint(&u.Flingy, u.MoveTarget.Pos)
	for n := range u.Reaver().OutsideUnits.All() {
		f.setUnitOrder(n, f.getOrderType(OrdersSelfDestructing))
	}
	f.setUnitOrderUnit(u, f.getOrderType(OrdersReaver), u.OrderTarget.Unit)
}

func (f *StateFunctions) orderRechargeShieldsUnit(u *Unit) {
	target := u.OrderTarget.Unit
	if target == nil || target.Energy == Fp8Zero() || f.unitIsDisabled(target) {
		f.orderDone(u, OrderTarget{})
		return
	}
	if u.OrderState == 0 {
		f.moveToTargetReset(u, target)
		u.OrderState = 1
	} else if u.OrderState == 1 {
		if f.unitIsAtMoveTarget(&u.Flingy) {
			if f.uImmovable(u) {
				f.orderDone(u, OrderTarget{})
				return
			}
			f.stopUnit(u)
			f.setNextTargetWaypoint(&u.Flingy, u.MoveTarget.Pos)
			u.OrderState = 2
		} else if f.unitTargetInRange(u, target, 128) {
			f.stopUnit(u)
			f.setNextTargetWaypoint(&u.Flingy, u.MoveTarget.Pos)
			u.OrderState = 2
		}
	}
	if u.OrderState == 2 {
		f.setSecondaryOrder(target, f.getOrderType(OrdersShieldBattery))
		f.createSizedImage(u, ImageTypesIMAGEIDRechargeShieldsSmall, false, imageOrderTop)
		f.queueOrderFront(u, f.getOrderType(OrdersRechargeShieldsUnitRemoveOverlay), OrderTarget{})
		u.OrderState = 3
	}
	if u.OrderState == 3 {
		f.rechargeShields(u, target)
	}
}

func (f *StateFunctions) rechargeShields(u *Unit, battery *Unit) {
	if battery.OrderTarget.Unit == nil {
		battery.OrderTarget.Unit = u
	}

	shieldRecharge := Fp8Integer(u.UnitType.ShieldPoints).Sub(u.ShieldPoints)
	if shieldRecharge.Greater(Fp8Integer(5)) {
		shieldRecharge = Fp8Integer(5)
	}
	energyCost := shieldRecharge.DivInt(2)
	if energyCost.Greater(battery.Energy) {
		energyCost = battery.Energy
		shieldRecharge = energyCost.MulInt(2)
	}
	f.setUnitShieldPoints(u, u.ShieldPoints.Add(shieldRecharge))
	battery.Energy = battery.Energy.Sub(energyCost)

	if u.ShieldPoints.GreaterEqual(Fp8Integer(u.UnitType.ShieldPoints)) || battery.Energy == Fp8Zero() {
		if battery.OrderTarget.Unit == u {
			battery.OrderTarget.Unit = nil
		}
		f.orderDone(u, OrderTarget{})
	}
}

func (f *StateFunctions) orderRechargeShieldsUnitRemoveOverlay(u *Unit) {
	f.destroyImageFromToUnit(u, ImageTypesIMAGEIDRechargeShieldsSmall, ImageTypesIMAGEIDRechargeShieldsLarge)
	f.finishHiddenOrder(u)
}

func (f *StateFunctions) orderRechargeShieldsBattery(u *Unit) {
	target := u.OrderTarget.Unit
	f.setSecondaryOrder(u, f.getOrderType(OrdersShieldBattery))
	if target != nil && f.unitCanUseShieldBattery(target, u) {
		f.setUnitOrderUnit(target, f.getOrderType(OrdersFollow), u)
	} else {
		for n := range f.findUnitsNoexpand(f.squareAt(u.OrderTarget.Pos, 128)) {
			if f.unitCanUseShieldBattery(n, u) {
				f.setUnitOrderUnit(n, f.getOrderType(OrdersFollow), u)
			}
		}
	}
	f.orderDone(u, OrderTarget{})
}

func (f *StateFunctions) orderFeedback(u *Unit) {
	target := u.OrderTarget.Unit
	if target == nil || !f.utHasEnergy(target.UnitType) || f.uHallucination(target) {
		f.orderDone(u, OrderTarget{})
		return
	}
	tech := f.getTechType(TechTypesFeedback)
	if !f.spellCastTargetMovement(u, tech, f.unitSightRange(u, true)) {
		return
	}
	if target.Energy != Fp8Zero() {
		f.weaponDealDamage(f.getWeaponType(WeaponTypesFeedback), target.Energy, 1, target, directionLiteral(1), u, u.Owner)
		target.Energy = Fp8Zero()
		u.Energy = u.Energy.Sub(Fp8Integer(tech.EnergyCost))
		f.playSoundUnit(1061, target, false)
		if f.unitDying(target) {
			spriteID := SpriteTypes(int(SpriteTypesSPRITEIDFeedbackHitSmall) + int(f.unitSpriteSize(target.UnitType)))
			f.createVisibleThingyAbove(f.getSpriteType(spriteID), u.Sprite.Position, 0, u.Sprite.ElevationLevel)
		} else {
			f.createSizedImage(u, ImageTypesIMAGEIDFeedbackSmall, true, imageOrderTop)
		}
	}
	f.orderDone(u, OrderTarget{})
}

func (f *StateFunctions) orderCastHallucination(u *Unit) {
	tech := f.getTechType(TechTypesHallucination)
	if !f.spellCastTargetMovement(u, tech, f.unitSightRange(u, true)) {
		return
	}
	target := u.OrderTarget.Unit
	for i := 0; i != 2; i++ {
		if !f.placeHallucination(target, u.Owner) {
			break
		}
	}
	u.Energy = u.Energy.Sub(Fp8Integer(tech.EnergyCost))
	f.playSoundUnit(618, u.OrderTarget.Unit, false)
	hitTarget := target
	if target.Subunit != nil {
		hitTarget = target.Subunit
	}
	f.createImage(f.getImageType(ImageTypesIMAGEIDHallucinationHit), hitTarget.Sprite, Xy{}, imageOrderTop, nil)
	f.orderDone(u, OrderTarget{})
}

func (f *StateFunctions) placeHallucination(target *Unit, owner int32) bool {
	n := f.createHallucination(target, owner)
	if n == nil {
		return false
	}
	placed, placement := f.findUnitPlacement(n, n.Sprite.Position, false)
	if !placed {
		f.destroyUnit(n)
		return false
	}
	f.moveUnit(n, placement)
	if f.uCanTurn(n) {
		index := uint(f.lcgRandRange(32, 0, 32))
		f.setSpriteImagesHeadingByIndex(n.Sprite, index)
		f.setUnitHeading(n, f.directionFromIndex(8*index))
	}
	f.showUnit(n)
	return true
}

func (f *StateFunctions) orderNeutral(u *Unit) {
	target := u.AutoTargetUnit
	if target == nil {
		return
	}
	u.AutoTargetUnit = nil
	if !f.usHidden(&target.Thingy) && !f.unitTargetIsUndetected(u, target) && target.Owner != u.Owner {
		f.st.Players[u.Owner].Controller = PlayerControllerComputerGame
		f.st.Alliances[u.Owner][target.Owner] = 0
		f.st.Alliances[target.Owner][u.Owner] = 0
		throwError("a neutral unit attacked something. this triggers computer player behavior which is not supported yet :(")
	}
}
