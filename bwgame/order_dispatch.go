package bwgame

func (f *StateFunctions) executeMainOrder(u *Unit) {
	switch u.OrderType.ID {
	case OrdersDie:
		f.orderDie(u)
		return
	case OrdersIncompleteWarping:
		f.orderIncompleteWarping(u)
		return
	case OrdersNukeTrack:
		f.orderNukeTrack(u)
		return
	case OrdersWarpIn:
		throwError("WarpIn")
		return
	}

	if f.unitIsDisabled(u) || (!f.uCanMove(u) && f.uCannotAttack(u)) {
		if u.MainOrderTimer == 0 {
			u.MainOrderTimer = 15
		}
		if f.unitIsDisabled(u) {
			return
		}
	}

	switch u.OrderType.ID {
	case OrdersTurretGuard:
		f.orderTurretGuard(u)
	case OrdersTurretAttack:
		f.orderTurretAttack(u)
	case OrdersDroneBuild:
		f.orderDroneBuild(u)
	case OrdersPlaceBuilding:
		f.orderPlaceBuilding(u)
	case OrdersPlaceProtossBuilding:
		f.orderPlaceProtossBuilding(u)
	case OrdersConstructingBuilding:
		f.orderConstructingBuilding(u)
	case OrdersRepair:
		f.orderRepair(u)
	case OrdersZergBirth:
		f.orderZergBirth(u)
	case OrdersZergUnitMorph:
		f.orderZergUnitMorph(u)
	case OrdersIncompleteBuilding:
		f.orderIncompleteBuilding(u)
	case OrdersIncompleteMorphing:
		f.orderIncompleteMorphing(u)
	case OrdersScarabAttack:
		f.orderScarabAttack(u)
	case OrdersRechargeShieldsUnit:
		f.orderRechargeShieldsUnit(u)
	case OrdersBuildingLand:
		f.orderBuildingLand(u)
	case OrdersBuildingLiftoff:
		f.orderBuildingLiftoff(u)
	case OrdersResearchTech:
		f.orderResearchTech(u)
	case OrdersUpgrade:
		f.orderUpgrade(u)
	case OrdersGatheringInterrupted:
		f.orderGatheringInterrupted(u)
	case OrdersGatherWaitInterrupted:
		f.orderGatherWaitInterrupted(u)
	case OrdersRechargeShieldsUnitRemoveOverlay:
		f.orderRechargeShieldsUnitRemoveOverlay(u)
	case OrdersSieging:
		f.orderSieging(u)
	case OrdersUnsieging:
		f.orderUnsieging(u)
	case OrdersArchonWarp:
		f.orderArchonWarp(u)
	case OrdersCompletingArchonSummon:
		f.orderCompletingArchonSummon(u)
	case OrdersNukeTrain:
		f.orderNukeTrain(u)
	case OrdersInitializeArbiter:
		f.orderInitializeArbiter(u)
	case OrdersResetCollision:
		f.orderResetCollision(u)
	case OrdersResetHarvestCollision:
		f.orderResetHarvestCollision(u)
	case OrdersCTFCOP2:
		throwError("CTFCOP2")
	case OrdersSelfDestructing:
		f.orderSelfDestructing(u)
	case OrdersCritter:
		f.orderCritter(u)
	case OrdersMedicHeal:
		f.orderMedicHeal(u)
	case OrdersHealMove:
		f.orderHealMove(u)
	case OrdersMedicHoldPosition:
		f.orderMedicHoldPosition(u)
	case OrdersMedicHealToIdle:
		f.orderMedicHealToIdle(u)
	case OrdersDarkArchonMeld:
		f.orderDarkArchonMeld(u)
	}
	if u.OrderProcessTimer != 0 {
		u.OrderProcessTimer--
		return
	}
	u.OrderProcessTimer = 8
	switch u.OrderType.ID {
	case OrdersStop:
		f.orderStop(u)
	case OrdersGuard:
		f.orderGuard(u)
	case OrdersPlayerGuard:
		f.orderPlayerGuard(u)
	case OrdersBunkerGuard:
		throwError("BunkerGuard")
	case OrdersMove:
		f.orderMove(u)
	case OrdersReaverStop:
		f.orderReaverStop(u)
	case OrdersMoveToAttack:
		f.orderMoveToTargetOrder(u)
	case OrdersAttackUnit:
		f.orderAttackUnit(u)
	case OrdersAttackFixedRange:
		f.orderAttackFixedRange(u)
	case OrdersHover:
		throwError("Hover")
	case OrdersAttackMove:
		f.orderAttackMove(u)
	case OrdersInfestedCommandCenter:
		f.orderInfestedCommandCenter(u)
	case OrdersUnusedNothing:
		throwError("UnusedNothing")
	case OrdersUnusedPowerup:
		throwError("UnusedPowerup")
	case OrdersTowerGuard:
		f.orderTowerGuard(u)
	case OrdersTowerAttack:
		f.orderTowerAttack(u)
	case OrdersSpiderMine:
		f.orderSpiderMine(u)
	case OrdersStayInRange:
		f.orderStayInRange(u)
	case OrdersNothing:
		f.orderNothing(u)
	case OrdersDroneStartBuild:
		f.orderDroneStartBuild(u)
	case OrdersCastInfestation:
		f.orderCastInfestation(u)
	case OrdersMoveToInfest:
		f.orderMoveToTargetOrder(u)
	case OrdersInfestingCommandCenter:
		f.orderInfestingCommandCenter(u)
	case OrdersMoveToRepair:
		f.orderMoveToTargetOrder(u)
	case OrdersPlaceAddon:
		f.orderPlaceAddon(u)
	case OrdersZergBuildingMorph:
		f.orderZergBuildingMorph(u)
	case OrdersBuildNydusExit:
		f.orderBuildNydusExit(u)
	case OrdersEnterNydusCanal:
		f.orderEnterNydusCanal(u)
	case OrdersIncompleteWarping:
		throwError("IncompleteWarping")
	case OrdersFollow:
		f.orderFollow(u)
	case OrdersCarrier:
		f.orderCarrier(u)
	case OrdersReaverCarrierMove:
		f.orderMove(u)
	case OrdersCarrierStop:
		f.orderCarrierStop(u)
	case OrdersCarrierAttack:
		f.orderCarrierAttack(u)
	case OrdersCarrierMoveToAttack:
		f.orderMoveToTargetOrder(u)
	case OrdersCarrierIgnore2:
		throwError("CarrierIgnore2")
	case OrdersCarrierFight:
		f.orderCarrier(u)
	case OrdersCarrierHoldPosition:
		f.orderCarrier(u)
	case OrdersReaver:
		f.orderReaver(u)
	case OrdersReaverAttack:
		f.orderReaverAttack(u)
	case OrdersReaverMoveToAttack:
		f.orderMoveToTargetOrder(u)
	case OrdersReaverFight:
		f.orderReaver(u)
	case OrdersReaverHoldPosition:
		f.orderReaver(u)
	case OrdersInterceptorAttack:
		f.orderInterceptorAttack(u)
	case OrdersRechargeShieldsBattery:
		f.orderRechargeShieldsBattery(u)
	case OrdersInterceptorReturn:
		f.orderInterceptorReturn(u)
	case OrdersDroneLand:
		f.orderDroneLand(u)
	case OrdersDroneLiftOff:
		throwError("DroneLiftOff")
	case OrdersLiftingOff:
		f.orderLiftingOff(u)
	case OrdersLarva:
		f.orderLarva(u)
	case OrdersHarvest1:
		f.orderMoveToGas(u)
	case OrdersHarvest2:
		f.orderMoveToTargetOrder(u)
	case OrdersMoveToGas:
		f.orderMoveToGas(u)
	case OrdersWaitForGas:
		f.orderWaitForGas(u)
	case OrdersHarvestGas:
		f.orderHarvestGas(u)
	case OrdersReturnGas:
		f.orderReturnGas(u)
	case OrdersMoveToMinerals:
		f.orderMoveToMinerals(u)
	case OrdersWaitForMinerals:
		f.orderWaitForMinerals(u)
	case OrdersMiningMinerals:
		f.orderMiningMinerals(u)
	case OrdersReturnMinerals:
		f.orderReturnMinerals(u)
	case OrdersEnterTransport:
		f.orderEnterTransport(u)
	case OrdersPickupIdle:
		f.orderPickupIdle(u)
	case OrdersPickupTransport:
		f.orderPickupTransport(u)
	case OrdersPickupBunker:
		f.orderPickupBunker(u)
	case OrdersPickup4:
		throwError("Pickup4")
	case OrdersWatchTarget:
		f.orderWatchTarget(u)
	case OrdersInitCreepGrowth:
		f.orderInitCreepGrowth(u)
	case OrdersHoldPosition:
		f.orderHoldPosition(u)
	case OrdersQueenHoldPosition:
		f.orderQueenHoldPosition(u)
	case OrdersUnload:
		f.orderUnload(u)
	case OrdersMoveUnload:
		f.orderMoveUnload(u)
	case OrdersFireYamatoGun:
		f.orderSpell(u)
	case OrdersMoveToFireYamatoGun:
		f.orderMoveToTargetOrder(u)
	case OrdersCastLockdown:
		f.orderSpell(u)
	case OrdersBurrowing:
		f.orderBurrowing(u)
	case OrdersBurrowed:
		f.orderBurrowed(u)
	case OrdersUnburrowing:
		f.orderUnburrowing(u)
	case OrdersCastDarkSwarm:
		f.orderSpell(u)
	case OrdersCastParasite:
		f.orderSpell(u)
	case OrdersCastSpawnBroodlings:
		f.orderSpell(u)
	case OrdersCastEMPShockwave:
		f.orderSpell(u)
	case OrdersNukeLaunch:
		f.orderNukeLaunch(u)
	case OrdersNukePaint:
		f.orderNukePaint(u)
	case OrdersNukeUnit:
		f.orderNukeUnit(u)
	case OrdersCastNuclearStrike:
		f.orderCastNuclearStrike(u)
	case OrdersPlaceMine:
		f.orderPlaceMine(u)
	case OrdersRightClickAction:
		f.orderRightClickAction(u)
	case OrdersSuicideUnit:
		f.orderSuicideUnit(u)
	case OrdersSuicideLocation:
		f.orderSuicideLocation(u)
	case OrdersSuicideHoldPosition:
		f.orderSuicideHoldPosition(u)
	case OrdersCastRecall:
		f.orderCastRecall(u)
	case OrdersTeleport:
		throwError("Teleport")
	case OrdersCastScannerSweep:
		f.orderCastScannerSweep(u)
	case OrdersScanner:
		f.orderScanner(u)
	case OrdersCastDefensiveMatrix:
		f.orderCastDefensiveMatrix(u)
	case OrdersCastPsionicStorm:
		f.orderSpell(u)
	case OrdersCastIrradiate:
		f.orderSpell(u)
	case OrdersCastPlague:
		f.orderSpell(u)
	case OrdersCastConsume:
		f.orderSpell(u)
	case OrdersCastEnsnare:
		f.orderSpell(u)
	case OrdersCastStasisField:
		f.orderSpell(u)
	case OrdersCastHallucination:
		f.orderCastHallucination(u)
	case OrdersPatrol:
		f.orderPatrol(u)
	case OrdersCTFCOPInit:
		throwError("CTFCOPInit")
	case OrdersCTFCOP2:
		throwError("CTFCOP2")
	case OrdersComputerAI:
		throwError("ComputerAI")
	case OrdersAtkMoveEP:
		throwError("AtkMoveEP")
	case OrdersHarassMove:
		throwError("HarassMove")
	case OrdersAIPatrol:
		throwError("AIPatrol")
	case OrdersGuardPost:
		throwError("GuardPost")
	case OrdersRescuePassive:
		throwError("RescuePassive")
	case OrdersNeutral:
		f.orderNeutral(u)
	case OrdersComputerReturn:
		throwError("ComputerReturn")
	case OrdersInitializePsiProvider:
		f.orderInitializePsiProvider(u)
	case OrdersHiddenGun:
		throwError("HiddenGun")
	case OrdersOpenDoor:
		throwError("OpenDoor")
	case OrdersCloseDoor:
		throwError("CloseDoor")
	case OrdersHideTrap:
		throwError("HideTrap")
	case OrdersRevealTrap:
		throwError("RevealTrap")
	case OrdersEnableDoodad:
		throwError("EnableDoodad")
	case OrdersWarpIn:
		throwError("WarpIn")
	case OrdersMedicIdle:
		f.orderMedicIdle(u)
	case OrdersCastRestoration:
		f.orderSpell(u)
	case OrdersCastDisruptionWeb:
		f.orderSpell(u)
	case OrdersCastMindControl:
		f.orderCastMindControl(u)
	case OrdersCastFeedback:
		f.orderFeedback(u)
	case OrdersCastOpticalFlare:
		f.orderSpell(u)
	case OrdersCastMaelstrom:
		f.orderSpell(u)
	}
}

func (f *StateFunctions) resumeBuildingUnit(u *Unit, placeWhenCompleted bool) bool {
	if f.st.CheatOperationCwal {
		f.setUnitHp(u, u.Hp.Add(u.HpConstructionRate.MulInt(16)))
		if u.RemainingBuildTime != 0 {
			if u.RemainingBuildTime > 16 {
				u.RemainingBuildTime -= 16
			} else {
				u.RemainingBuildTime = 0
			}
			return true
		}
	} else {
		f.setUnitHp(u, u.Hp.Add(u.HpConstructionRate))
		if u.RemainingBuildTime != 0 {
			u.RemainingBuildTime--
			return true
		}
	}
	f.finishBuildingUnit(u)
	if placeWhenCompleted {
		if !f.placeCompletedUnit(u) {
			return false
		}
	}
	f.completeUnit(u)
	if f.uGroundedBuilding(u) {
		if f.unitRace(u.UnitType) == RaceTerran {
			f.findAndConnectAddon(u)
		}
	}
	return true
}

func (f *StateFunctions) rallyUnit(u *Unit, factoryUnit *Unit) {
	if !f.unitIsFactory(factoryUnit) {
		return
	}
	target := factoryUnit.Building.Rally
	if target.Unit == factoryUnit {
		return
	}
	if target.Pos.X != 0 {
		if target.Unit != nil && target.Unit.Owner == u.Owner {
			f.setUnitOrderUnit(u, f.getOrderType(OrdersFollow), target.Unit)
		} else {
			f.setUnitOrderPos(u, f.getOrderType(OrdersMove), target.Pos)
		}
	}
}

func (f *StateFunctions) secondaryOrderTrain(u *Unit) {
	if f.unitIsDisabled(u) {
		return
	}
	if f.unitRace(u.UnitType) == RaceZerg && !f.unitIs(u.UnitType, UnitTypesZergInfestedCommandCenter) {
		return
	}
	if u.SecondaryOrderState == 0 || u.SecondaryOrderState == 1 {
		if u.BuildQueue.Empty() {
			f.setSecondaryOrder(u, f.getOrderType(OrdersNothing))
			f.spriteRunAnim(u.Sprite, IscriptAnimsWorkingToIdle)
		} else {
			var buildUnit *Unit
			buildUnitType := u.BuildQueue.Front()
			if f.uGroundedBuilding(u) || f.unitIsCarrier(u.UnitType) || f.unitIsReaver(u.UnitType) {
				if f.hasAvailableSupplyFor(u.Owner, buildUnitType, u.SecondaryOrderState == 0) {
					buildUnit = f.createUnit(buildUnitType, u.Sprite.Position, u.Owner)
					if buildUnit == nil {
						f.displayLastErrorForPlayer(u.Owner)
					}
				}
			}
			u.CurrentBuildUnit = buildUnit
			if buildUnit != nil {
				u.SecondaryOrderState = 2
				f.spriteRunAnim(u.Sprite, IscriptAnimsIsWorking)
			} else {
				u.SecondaryOrderState = 1
			}
		}
	} else if u.SecondaryOrderState == 2 {
		buildUnit := u.CurrentBuildUnit
		if buildUnit != nil {
			if f.resumeBuildingUnit(buildUnit, true) {
				if f.uCompleted(buildUnit) {
					if f.unitIs(buildUnit.UnitType, UnitTypesTerranNuclearMissile) {
						f.hideUnit(buildUnit, true)
					} else {
						f.rallyUnit(buildUnit, u)
					}
					if !u.BuildQueue.Empty() {
						u.BuildQueue.Erase(0)
					}
					u.CurrentBuildUnit = nil
					u.SecondaryOrderState = 0
				}
			} else {
				if !u.BuildQueue.Empty() {
					if u.CurrentBuildUnit != nil {
						f.cancelBuildingUnit(u.CurrentBuildUnit)
					} else if !f.utBuilding(u.BuildQueue.Front()) {
						f.refundUnitCosts(u.Owner, u.BuildQueue.Front())
					}
					u.BuildQueue.Erase(0)
				}
				u.CurrentBuildUnit = nil
				u.SecondaryOrderState = 0
			}
		} else {
			u.SecondaryOrderState = 0
		}
	}
}

func (f *StateFunctions) secondaryOrderBuildAddon(u *Unit) {
	if f.unitIsDisabled(u) {
		return
	}
	addon := u.CurrentBuildUnit
	if addon == nil {
		u.Building.AddonBuildType = nil
		f.setSecondaryOrder(u, f.getOrderType(OrdersNothing))
		return
	}
	if !f.uGroundedBuilding(u) && !f.uCompleted(addon) {
		f.cancelBuildingUnit(addon)
		u.Building.AddonBuildType = nil
		f.setSecondaryOrder(u, f.getOrderType(OrdersNothing))
		return
	}
	f.resumeBuildingUnit(addon, false)
	if f.uCompleted(addon) {
		f.connectAddon(u, addon)
		u.Building.AddonBuildType = nil
		f.setSecondaryOrder(u, f.getOrderType(OrdersNothing))
	}
}

func (f *StateFunctions) secondaryOrderCloak(u *Unit) {
	f.setUnitCloaked(u)
}

func (f *StateFunctions) secondaryOrderDecloak(u *Unit) {
	f.setSecondaryOrder(u, f.getOrderType(OrdersNothing))
}

func (f *StateFunctions) secondaryOrderSpawningLarva(u *Unit) {
	if u.OrderProcessTimer != 0 {
		return
	}
	if u.Owner == 11 {
		return
	}
	larvaCount := int32(0)
	for n := range f.findUnitsNoexpand(f.squareAt(u.Sprite.Position, 32*8)) {
		if !f.unitIs(n.UnitType, UnitTypesZergLarva) {
			continue
		}
		if n.ConnectedUnit != u {
			continue
		}
		if f.usHidden(&n.Thingy) {
			continue
		}
		larvaCount++
		if larvaCount >= 3 {
			return
		}
	}
	if u.Building.LarvaTimer != 0 {
		u.Building.LarvaTimer--
		return
	}
	if f.spawnLarva(u) != nil {
		u.Building.LarvaTimer = 37
	} else {
		u.Building.LarvaTimer = 3
	}
}

func (f *StateFunctions) secondaryOrderSpreadCreep(u *Unit) {
	if f.unitIsHatchery(u.UnitType) {
		f.secondaryOrderSpawningLarva(u)
	}
	if u.Building.CreepTimer != 0 {
		u.Building.CreepTimer--
		return
	}
	u.Building.CreepTimer = 15
	anyTilesOccupied := false
	if !f.spreadCreep(f.getUnitType(UnitTypesZergHive), u.Sprite.Position, &anyTilesOccupied) && !anyTilesOccupied {
		if f.unitIsHatchery(u.UnitType) {
			f.setSecondaryOrder(u, f.getOrderType(OrdersSpawningLarva))
		} else {
			f.setSecondaryOrder(u, f.getOrderType(OrdersNothing))
		}
	}
}

func (f *StateFunctions) secondaryOrderCloakNearbyUnits(u *Unit) {
	w := f.unitAirWeapon(u)
	if w == nil {
		throwError("secondary_order_CloakNearbyUnits: no air weapon")
	}
	rangeValue := f.weaponMaxRange(u, w)
	area := f.squareAt(u.Sprite.Position, rangeValue)
	if area.From.X < 0 {
		area.From.X = 0
	} else if area.To.X > int32(f.gameSt.MapWidth) {
		area.To.X = int32(f.gameSt.MapWidth)
	}
	if area.From.Y < 0 {
		area.From.Y = 0
	} else if area.To.Y > int32(f.gameSt.MapHeight) {
		area.To.Y = int32(f.gameSt.MapHeight)
	}
	for target := range f.findUnitsNoexpand(area) {
		if target.Owner != u.Owner {
			continue
		}
		if f.utBuilding(target.UnitType) {
			continue
		}
		if f.unitIsArbiter(target.UnitType) {
			continue
		}
		if f.uHallucination(target) {
			continue
		}
		if f.utPowerup(target.UnitType) {
			continue
		}
		if f.unitIs(target.UnitType, UnitTypesTerranNuclearMissile) {
			continue
		}
		if target.OrderType.ID == OrdersWarpIn {
			continue
		}
		if !f.unitTargetInRange(u, target, rangeValue) {
			continue
		}
		f.setUnitCloaked(target)
		if !f.uPassivelyCloaked(target) {
			f.uSetStatusFlag(target, UnitStatusFlagPassivelyCloaked)
		}
	}
}

func (f *StateFunctions) secondaryOrderTrainFighter(u *Unit) {
	if !f.unitIsCarrier(u.UnitType) && !f.unitIsReaver(u.UnitType) {
		return
	}
	if u.SecondaryOrderState == 0 || u.SecondaryOrderState == 1 {
		if u.BuildQueue.Empty() {
			u.SecondaryOrderState = 3
			u.CurrentBuildUnit = nil
		} else {
			var buildUnit *Unit
			buildUnitType := u.BuildQueue.Front()
			if f.uGroundedBuilding(u) || f.unitIsCarrier(u.UnitType) || f.unitIsReaver(u.UnitType) {
				if f.hasAvailableSupplyFor(u.Owner, buildUnitType, u.SecondaryOrderState == 0) {
					buildUnit = f.createUnit(buildUnitType, u.Sprite.Position, u.Owner)
					if buildUnit == nil {
						f.displayLastErrorForPlayer(u.Owner)
					}
				}
			}
			u.CurrentBuildUnit = buildUnit
			if buildUnit != nil {
				if f.unitIsFighter(buildUnit.UnitType) {
					buildUnit.Fighter.Parent = u
				}
				u.SecondaryOrderState = 2
			} else {
				u.SecondaryOrderState = 1
			}
		}
	} else if u.SecondaryOrderState == 2 {
		buildUnit := u.CurrentBuildUnit
		if buildUnit == nil {
			u.SecondaryOrderState = 0
		} else {
			f.resumeBuildingUnit(buildUnit, false)
			if f.uCompleted(buildUnit) {
				f.hideUnit(buildUnit, true)
				if f.unitIsFighter(buildUnit.UnitType) {
					if f.unitIsCarrier(u.UnitType) {
						u.Carrier.InsideUnits.PushFront(buildUnit)
						u.Carrier.InsideCount++
					} else if f.unitIsReaver(u.UnitType) {
						u.Reaver().InsideUnits.PushFront(buildUnit)
						u.Reaver().InsideCount++
					}
					buildUnit.Fighter.IsOutside = false
				}
				u.BuildQueue.Erase(0)
				u.CurrentBuildUnit = nil
				u.SecondaryOrderState = 0
			}
		}
	} else if u.SecondaryOrderState == 3 {
		if f.unitIsCarrier(u.UnitType) {
			for n := range u.Carrier.InsideUnits.All() {
				if n.Hp.GreaterEqual(n.UnitType.Hitpoints) {
					continue
				}
				f.setUnitHp(n, n.Hp.Add(Fp8FromRaw(128)))
				break
			}
		} else {
			u.SecondaryOrderState = 4
		}
	}
}

func (f *StateFunctions) secondaryOrderShieldBattery(u *Unit) {
	target := u.OrderTarget.Unit
	if u.SecondaryOrderState == 0 {
		if target != nil {
			f.spriteRunAnim(u.Sprite, IscriptAnimsIsWorking)
			u.SecondaryOrderState = 1
		}
	} else if u.SecondaryOrderState == 1 {
		if target != nil {
			if target.OrderType.ID != OrdersRechargeShieldsUnit {
				u.OrderTarget.Unit = nil
				u.SecondaryOrderState = 2
			}
		} else {
			u.SecondaryOrderState = 2
		}
	} else if u.SecondaryOrderState == 2 {
		if target != nil {
			u.SecondaryOrderState = 1
		} else {
			f.spriteRunAnim(u.Sprite, IscriptAnimsWorkingToIdle)
			u.SecondaryOrderState = 0
		}
	}
}

func (f *StateFunctions) executeSecondaryOrder(u *Unit) {
	if u.SecondaryOrderType.ID == OrdersHallucination2 {
		if u.DefensiveMatrixHp.RawValue != 0 || u.StimTimer != 0 || u.EnsnareTimer != 0 || u.LockdownTimer != 0 || u.IrradiateTimer != 0 || u.StasisTimer != 0 || u.ParasiteFlags != 0 || u.StormTimer != 0 || u.PlagueTimer != 0 || u.BlindedBy != 0 || u.MaelstromTimer != 0 {
			f.killUnit(u)
		}
		return
	}
	if f.unitIsDisabled(u) {
		return
	}
	switch u.SecondaryOrderType.ID {
	case OrdersBuildAddon:
		f.secondaryOrderBuildAddon(u)
	case OrdersTrain:
		f.secondaryOrderTrain(u)
	case OrdersTrainFighter:
		f.secondaryOrderTrainFighter(u)
	case OrdersShieldBattery:
		f.secondaryOrderShieldBattery(u)
	case OrdersSpawningLarva:
		f.secondaryOrderSpawningLarva(u)
	case OrdersSpreadCreep:
		f.secondaryOrderSpreadCreep(u)
	case OrdersCloak:
		f.secondaryOrderCloak(u)
	case OrdersDecloak:
		f.secondaryOrderDecloak(u)
	case OrdersCloakNearbyUnits:
		f.secondaryOrderCloakNearbyUnits(u)
	}
}

func (f *StateFunctions) updateUnit(u *Unit) {
	f.updateUnitValues(u)

	f.executeMainOrder(u)
	f.executeSecondaryOrder(u)

	if u.Subunit != nil && !f.utTurret(u.UnitType) {
		previousIscriptUnit := f.iscriptUnit
		f.iscriptUnit = u.Subunit
		f.updateUnit(u.Subunit)
		f.iscriptUnit = previousIscriptUnit
	}

	if u.Sprite != nil {
		if !f.iscriptExecuteSprite(u.Sprite) {
			u.Sprite = nil
		}
	}

	if u.Sprite == nil && !f.unitDead(u) {
		throwError("update_unit: unit has null sprite")
	}
}
