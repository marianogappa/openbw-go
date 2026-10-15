package bwgame

func isLinked[T any, A LinkAccessor[T]](l *IntrusiveList[T, A], v *T) bool {
	l.learnLinkOffset(v)
	return l.first == v || l.link(v).Prev != nil
}

func unitSpeedUpgrade(id UnitTypes) UpgradeTypes {
	switch id {
	case UnitTypesTerranVulture, UnitTypesHeroJimRaynorVulture:
		return UpgradeTypesIonThrusters
	case UnitTypesZergOverlord:
		return UpgradeTypesPneumatizedCarapace
	case UnitTypesZergZergling:
		return UpgradeTypesMetabolicBoost
	case UnitTypesZergHydralisk:
		return UpgradeTypesMuscularAugments
	case UnitTypesProtossZealot:
		return UpgradeTypesLegEnhancements
	case UnitTypesProtossScout:
		return UpgradeTypesGraviticThrusters
	case UnitTypesProtossShuttle:
		return UpgradeTypesGraviticDrive
	case UnitTypesProtossObserver:
		return UpgradeTypesGraviticBoosters
	case UnitTypesZergUltralisk:
		return UpgradeTypesAnabolicSynthesis
	default:
		return UpgradeTypesNone
	}
}

func (f *StateFunctions) updateUnitSpeedUpgrades(u *Unit) {
	cooldown := false
	if f.unitIs(u.UnitType, UnitTypesHeroDevouringOne) {
		cooldown = true
	}
	if f.unitIs(u.UnitType, UnitTypesZergZergling) && f.playerHasUpgrade(u.Owner, UpgradeTypesAdrenalGlands) {
		cooldown = true
	}
	speed := false
	speedUpg := unitSpeedUpgrade(u.UnitType.ID)
	if speedUpg != UpgradeTypesNone && f.playerHasUpgrade(u.Owner, speedUpg) {
		speed = true
	}
	for _, id := range []UnitTypes{
		UnitTypesHeroHunterKiller,
		UnitTypesHeroYggdrasill,
		UnitTypesHeroFenixZealot,
		UnitTypesHeroMojo,
		UnitTypesHeroArtanis,
		UnitTypesZergLurker,
	} {
		if f.unitIs(u.UnitType, id) {
			speed = true
		}
	}
	if cooldown != f.uCooldownUpgrade(u) || speed != f.uSpeedUpgrade(u) {
		if cooldown {
			u.StatusFlags |= UnitStatusFlagCooldownUpgrade
		}
		if speed {
			u.StatusFlags |= UnitStatusFlagSpeedUpgrade
		}
		f.updateUnitSpeed(u)
	}
}

func (f *StateFunctions) updateUnitSpeed(u *Unit) {
	if u.FlingyMovementType == 2 {
		image := u.Sprite.MainImage
		if image == nil {
			throwError("null image")
		}
		script := image.IscriptState.CurrentScript
		animsPc := script.AnimationPc
		anim := int32(IscriptAnimsWalking)
		if uint(anim) < uint(len(animsPc)) && animsPc[anim] != 0 {
			prevIscriptUnit := f.iscriptUnit
			f.iscriptUnit = u
			var st IscriptState
			st.CurrentScript = script
			st.Animation = anim
			st.ProgramCounter = animsPc[anim]
			st.ReturnAddress = 0
			st.Wait = 0
			var totalDistanceMoved Fp8
			for i := 0; i < 32; i++ {
				var distanceMoved Fp8
				if !f.iscriptExecute(image, &st, true, &distanceMoved, false) {
					throwError("update_unit_speed: image destroyed")
				}
				totalDistanceMoved = totalDistanceMoved.Add(f.getModifiedUnitAcceleration(u, distanceMoved))
			}
			u.FlingyTopSpeed = totalDistanceMoved.DivInt(32)
			f.iscriptUnit = prevIscriptUnit
		}
	} else {
		u.FlingyTopSpeed = f.getModifiedUnitSpeed(u, u.FlingyType.TopSpeed)
		u.FlingyAcceleration = f.getModifiedUnitAcceleration(u, u.FlingyType.Acceleration)
		u.FlingyTurnRate = f.getModifiedUnitTurnRate(u, u.FlingyType.TurnRate)
	}
}

func (f *StateFunctions) incrementUnitCounts(u *Unit, count int32) {
	if f.uHallucination(u) {
		return
	}
	if f.utTurret(u.UnitType) {
		return
	}
	st := f.st
	st.UnitCounts[u.Owner][u.UnitType.ID] += count
	supplyRequired := u.UnitType.SupplyRequired
	race := f.unitRace(u.UnitType)
	if race == RaceZerg {
		if f.unitIsEgg(u.UnitType) {
			if !u.BuildQueue.Empty() {
				ut := u.BuildQueue.Front()
				supplyRequired = ut.SupplyRequired
				if f.utTwoUnitsInOneEgg(ut) {
					supplyRequired = supplyRequired.MulInt(2)
				}
			}
		} else {
			if f.utTwoUnitsInOneEgg(u.UnitType) && !f.uCompleted(u) {
				supplyRequired = supplyRequired.MulInt(2)
			}
		}
		st.SupplyUsed[u.Owner][0] = st.SupplyUsed[u.Owner][0].Add(supplyRequired.MulInt(count))
	} else if race == RaceTerran {
		st.SupplyUsed[u.Owner][1] = st.SupplyUsed[u.Owner][1].Add(supplyRequired.MulInt(count))
	} else if race == RaceProtoss {
		st.SupplyUsed[u.Owner][2] = st.SupplyUsed[u.Owner][2].Add(supplyRequired.MulInt(count))
	}
	groupFlags := u.UnitType.GroupFlags
	if groupFlags&GroupFlagsFactory != 0 {
		st.FactoryCounts[u.Owner] += count
	}
	if groupFlags&GroupFlagsMen != 0 {
		st.NonBuildingCounts[u.Owner] += count
	} else if groupFlags&GroupFlagsBuilding != 0 {
		st.BuildingCounts[u.Owner] += count
	} else if f.unitIsEgg(u.UnitType) {
		st.NonBuildingCounts[u.Owner] += count
	}
	if st.UnitCounts[u.Owner][u.UnitType.ID] < 0 {
		st.UnitCounts[u.Owner][u.UnitType.ID] = 0
	}
}

func (f *StateFunctions) unitIsFactory(u *Unit) bool {
	switch {
	case f.unitIs(u.UnitType, UnitTypesTerranCommandCenter),
		f.unitIs(u.UnitType, UnitTypesTerranBarracks),
		f.unitIs(u.UnitType, UnitTypesTerranFactory),
		f.unitIs(u.UnitType, UnitTypesTerranStarport),
		f.unitIs(u.UnitType, UnitTypesZergInfestedCommandCenter),
		f.unitIs(u.UnitType, UnitTypesZergHatchery),
		f.unitIs(u.UnitType, UnitTypesZergLair),
		f.unitIs(u.UnitType, UnitTypesZergHive),
		f.unitIs(u.UnitType, UnitTypesProtossNexus),
		f.unitIs(u.UnitType, UnitTypesProtossGateway),
		f.unitIs(u.UnitType, UnitTypesProtossStargate),
		f.unitIs(u.UnitType, UnitTypesProtossRoboticsFacility):
		return true
	}
	return false
}

func unitTilesOrigin(ut *UnitTypeT, position Xy) (Xy, Xy) {
	size := Xy{X: int32(uint32(ut.PlacementSize.X) / 32), Y: int32(uint32(ut.PlacementSize.Y) / 32)}
	pos := Xy{X: int32(uint32(position.X) / 32), Y: int32(uint32(position.Y) / 32)}
	return Xy{X: pos.X - size.X/2, Y: pos.Y - size.Y/2}, size
}

func (f *StateFunctions) setUnitTilesOccupied(ut *UnitTypeT, position Xy) {
	origin, size := unitTilesOrigin(ut, position)
	f.tilesFlagsOr(uint(origin.X), uint(origin.Y), uint(size.X), uint(size.Y), TileFlagOccupied)
}

func (f *StateFunctions) setUnitTilesUnoccupied(ut *UnitTypeT, position Xy) {
	origin, size := unitTilesOrigin(ut, position)
	f.tilesFlagsAnd(uint(origin.X), uint(origin.Y), uint(size.X), uint(size.Y), ^int32(TileFlagOccupied))
}

func (f *StateFunctions) initializeUnitType(u *Unit, unitType *UnitTypeT, pos Xy, owner int32) bool {
	prevIscriptUnit := f.iscriptUnit
	f.iscriptUnit = u
	defer func() { f.iscriptUnit = prevIscriptUnit }()

	if !f.initializeFlingy(&u.Flingy, unitType.Flingy.Get(), pos, owner, DirectionT{}) {
		return false
	}

	u.Owner = owner
	u.OrderType = f.getOrderType(OrdersFatal)
	u.OrderState = 0
	u.OrderSignal = 0
	u.OrderUnitType = nil
	u.MainOrderTimer = 0
	u.GroundWeaponCooldown = 0
	u.AirWeaponCooldown = 0
	u.SpellCooldown = 0
	u.OrderTarget = Target{}
	u.UnitType = unitType
	u.CarryingFlags = 0
	u.SecondaryOrderTimer = 0

	if !f.iscriptExecuteSprite(u.Sprite) {
		u.Sprite = nil
		return false
	}
	u.LastAttackingPlayer = 8
	u.ShieldPoints = Fp8Integer(u.UnitType.ShieldPoints)
	if f.unitIs(u.UnitType, UnitTypesProtossShieldBattery) {
		u.Energy = Fp8Integer(100)
	} else {
		u.Energy = f.unitMaxEnergy(u).DivInt(4)
	}

	u.Sprite.ElevationLevel = unitType.ElevationLevel
	f.uSetStatusFlagValue(u, UnitStatusFlagGroundedBuilding, f.utBuilding(u.UnitType))
	f.uSetStatusFlagValue(u, UnitStatusFlagFlying, f.utFlyer(u.UnitType))
	f.uSetStatusFlagValue(u, UnitStatusFlagCanTurn, f.utCanTurn(u.UnitType))
	f.uSetStatusFlagValue(u, UnitStatusFlagCanMove, f.utCanMove(u.UnitType))
	f.uSetStatusFlagValue(u, UnitStatusFlagGroundUnit, !f.utFlyer(u.UnitType))
	if u.UnitType.ElevationLevel < 12 {
		u.PathingFlags |= 1
	} else {
		u.PathingFlags &^= 1
	}
	u.Building.Addon = nil
	u.Building.AddonBuildType = nil
	f.onAddonBuildType(u, false)
	u.Building.UpgradeResearchTime = 0
	u.Building.ResearchingType = nil
	u.Building.UpgradingType = nil
	u.Building.LarvaTimer = 0
	u.Building.IsLanding = false
	u.Building.CreepTimer = 0
	u.Building.UpgradingLevel = 0

	f.initializeUnitBuildingUnion(u, unitType)
	f.initializeUnitUnion(u)

	if f.utWorker(u.UnitType) {
		u.Worker.Powerup = nil
		u.Worker.TargetResourcePosition = Xy{}
		u.Worker.TargetResourceUnit = nil
		u.Worker.RepairTimer = 0
		u.Worker.IsGathering = false
		u.Worker.ResourcesCarried = 0
		u.Worker.GatherTarget = nil
	}

	u.Path = nil
	u.MovementState = MovementStatesUMInit
	u.MoveTargetTimer = 0
	f.uSetStatusFlagValue(u, UnitStatusFlagInvincible, f.utInvincible(u.UnitType))

	u.DamageOverlayState = f.damageOverlayStates(u)

	f.initializeUnitConstructionRates(u)
	f.updateUnitSpeedUpgrades(u)
	f.updateUnitSpeed(u)

	return true
}

func claimUnion(used *bool, name string) {
	if *used {
		throwError("%s union already used", name)
	}
	*used = true
}

func (f *StateFunctions) initializeUnitBuildingUnion(u *Unit, unitType *UnitTypeT) {
	used := false
	if f.utResource(u.UnitType) {
		claimUnion(&used, "building")
		u.Building.Resource.ResourceCount = 0
		u.Building.Resource.ResourceIscript = 0
		u.Building.Resource.IsBeingGathered = false
		u.Building.Resource.GatherQueue.Clear()
	} else if f.unitIsMineralField(unitType) {
		throwError("mineral field is not a resource")
	}
	if f.unitIs(u.UnitType, UnitTypesTerranNuclearSilo) {
		claimUnion(&used, "building")
		u.Building.Silo = UnitBuildingSilo{}
	}
	if f.unitIsHatchery(u.UnitType) {
		claimUnion(&used, "building")
		u.Building.Hatchery.LarvaSpawnSideValues = [4]int32{}
	}
	if f.unitIsNydus(u.UnitType) {
		claimUnion(&used, "building")
		u.Building.Nydus.Exit = nil
	}
	if f.unitIs(u.UnitType, UnitTypesProtossPylon) {
		claimUnion(&used, "building")
		u.Building.Pylon.PsionicMatrixLink = Link[Unit]{}
		u.Building.Pylon.PsiFieldSprite = nil
	}
}

func (f *StateFunctions) initializeUnitUnion(u *Unit) {
	used := false
	if f.unitIsGhost(u.UnitType) {
		claimUnion(&used, "unit")
		u.Ghost.NukeDot = nil
	}
	if f.unitIsVulture(u.UnitType) {
		claimUnion(&used, "unit")
		u.Vulture.SpiderMineCount = 0
	}
	if f.unitIsCarrier(u.UnitType) {
		claimUnion(&used, "unit")
		u.Carrier.InsideUnits.Clear()
		u.Carrier.OutsideUnits.Clear()
		u.Carrier.InsideCount = 0
		u.Carrier.OutsideCount = 0
	}
	if f.unitIsReaver(u.UnitType) {
		claimUnion(&used, "unit")
		u.Reaver().InsideUnits.Clear()
		u.Reaver().OutsideUnits.Clear()
		u.Reaver().InsideCount = 0
		u.Reaver().OutsideCount = 0
	}
	if f.unitIsFighter(u.UnitType) {
		claimUnion(&used, "unit")
		u.Fighter.Parent = nil
		u.Fighter.FighterLink = Link[Unit]{}
		u.Fighter.IsOutside = false
	}
}

func (f *StateFunctions) initializeUnitConstructionRates(u *Unit) {
	ut := u.UnitType
	if ut.BuildTime == 0 {
		u.RemainingBuildTime = 1
		u.HpConstructionRate = Fp8FromRaw(1)
	} else {
		u.RemainingBuildTime = ut.BuildTime
		u.HpConstructionRate = ut.Hitpoints.Sub(ut.Hitpoints.DivInt(10)).Add(Fp8FromRaw(ut.BuildTime)).Sub(Fp8FromRaw(1)).DivInt(ut.BuildTime)
		if u.HpConstructionRate == Fp8Zero() {
			u.HpConstructionRate = Fp8FromRaw(1)
		}
	}
	if ut.HasShield && f.uGroundedBuilding(u) {
		maxShields := Fp8Integer(ut.ShieldPoints)
		u.ShieldPoints = maxShields.DivInt(10)
		if ut.BuildTime == 0 {
			u.ShieldConstructionRate = Fp8Integer(1)
		} else {
			u.ShieldConstructionRate = maxShields.Sub(u.ShieldPoints).DivInt(ut.BuildTime)
			if u.ShieldConstructionRate == Fp8Zero() {
				u.ShieldConstructionRate = Fp8Integer(1)
			}
		}
	}
}

func (f *StateFunctions) killInterceptors(u *Unit) {
	if !f.unitIsCarrier(u.UnitType) {
		return
	}
	inside := &u.Carrier.InsideUnits
	for n := inside.Begin(); n != nil; {
		next := inside.Next(n)
		f.killUnitWith(n, noKiller(ReasonLinked))
		n = next
	}
	outside := &u.Carrier.OutsideUnits
	for n := outside.Begin(); n != nil; {
		next := outside.Next(n)
		n.Fighter.Parent = nil
		f.setRemoveTimer(n, f.lcgRandRange(39, 15, 45))
		outside.Remove(n)
		u.Carrier.OutsideCount--
		n.Fighter.FighterLink = Link[Unit]{}
		n = next
	}
}

func (f *StateFunctions) destroyReaverFighters(u *Unit) {
	reaver := u.Reaver()
	for n := reaver.InsideUnits.Begin(); n != nil; {
		next := reaver.InsideUnits.Next(n)
		f.killUnitWith(n, noKiller(ReasonLinked))
		n = next
	}
	for n := reaver.OutsideUnits.Begin(); n != nil; {
		next := reaver.OutsideUnits.Next(n)
		n.Fighter.Parent = nil
		reaver.OutsideUnits.Remove(n)
		reaver.OutsideCount--
		n.Fighter.FighterLink = Link[Unit]{}
		n = next
	}
}

func (f *StateFunctions) detachFighterFromParent(u *Unit) {
	if !f.uCompleted(u) {
		return
	}
	parent := u.Fighter.Parent
	if u.Fighter.Parent != nil {
		if u.Fighter.IsOutside {
			if f.unitIs(parent.UnitType, UnitTypesProtossCarrier) {
				parent.Carrier.OutsideUnits.Remove(u)
				parent.Carrier.OutsideCount--
			} else if f.unitIs(parent.UnitType, UnitTypesProtossReaver) {
				parent.Reaver().OutsideUnits.Remove(u)
				parent.Reaver().OutsideCount--
			}
		} else {
			if f.unitIs(parent.UnitType, UnitTypesProtossCarrier) {
				parent.Carrier.InsideUnits.Remove(u)
				parent.Carrier.InsideCount--
			} else if f.unitIs(parent.UnitType, UnitTypesProtossReaver) {
				parent.Reaver().InsideUnits.Remove(u)
				parent.Reaver().InsideCount--
			}
		}
	}
	u.Fighter.Parent = nil
	u.Fighter.FighterLink = Link[Unit]{}
}

func (f *StateFunctions) destroyUnitTypeSpecifics(u *Unit) {
	switch u.UnitType.ID {
	case UnitTypesProtossInterceptor, UnitTypesProtossScarab:
		f.detachFighterFromParent(u)
	case UnitTypesTerranNuclearSilo:
		if u.Building.Silo.Nuke != nil {
			f.killUnitWith(u.Building.Silo.Nuke, noKiller(ReasonLinked))
			u.Building.Silo.Nuke = nil
		}
	case UnitTypesTerranNuclearMissile:
		if u.ConnectedUnit != nil && f.unitIs(u.ConnectedUnit.UnitType, UnitTypesTerranNuclearSilo) {
			u.ConnectedUnit.Building.Silo = UnitBuildingSilo{}
			u.ConnectedUnit = nil
		}
	case UnitTypesProtossPylon:
		if isLinked(&f.st.PsionicMatrixUnits, u) {
			f.st.PsionicMatrixUnits.Remove(u)
			u.Building.Pylon.PsionicMatrixLink = Link[Unit]{}
		}
		f.st.UpdatePsionicMatrix = true
		if u.Building.Pylon.PsiFieldSprite != nil {
			f.destroySprite(u.Building.Pylon.PsiFieldSprite)
			u.Building.Pylon.PsiFieldSprite = nil
		}
	case UnitTypesZergNydusCanal:
		if u.Building.Nydus.Exit != nil {
			exit := u.Building.Nydus.Exit
			u.Building.Nydus.Exit = nil
			exit.Building.Nydus.Exit = nil
			f.killUnitWith(exit, noKiller(ReasonLinked))
		}
	}
}

func (f *StateFunctions) destroyUnitUnionData(u *Unit) {
	if f.unitIsCarrier(u.UnitType) {
		f.killInterceptors(u)
	} else if f.unitIsReaver(u.UnitType) {
		f.destroyReaverFighters(u)
	} else if f.unitIsGhost(u.UnitType) {
		if u.Ghost.NukeDot != nil {
			f.spriteRunAnim(u.Ghost.NukeDot.Sprite, IscriptAnimsDeath)
			u.Ghost.NukeDot = nil
		}
	} else if f.utResource(u.UnitType) {
		resource := &u.Building.Resource
		resource.IsBeingGathered = false
		for !resource.GatherQueue.Empty() {
			gatherer := resource.GatherQueue.Front()
			gatherer.Worker.GatherTarget = nil
			resource.GatherQueue.PopFront()
		}
	} else {
		f.destroyUnitTypeSpecifics(u)
	}
}

func (f *StateFunctions) destroyLoadedUnits(u *Unit) {
	if f.uLoaded(u) {
		f.unitUnloadImpl(u, true)
	}
	for n := range f.loadedUnits(u) {
		if f.unitIs(u.UnitType, UnitTypesTerranBunker) {
			if !f.unitUnload(n) {
				f.killUnitWith(n, killedInContainer(u))
			}
		} else {
			f.killUnitWith(n, killedInContainer(u))
		}
	}
	u.LoadedUnits = [8]UnitID{}
}

func (f *StateFunctions) removeAllTargetReferencesTo(u *Unit) {
	for n := range f.st.VisibleUnits.All() {
		f.removeTargetReferences(n, u)
	}
	for n := range f.st.HiddenUnits.All() {
		f.removeTargetReferences(n, u)
	}
	for n := range f.st.ActiveBullets.All() {
		f.removeBulletTargetReferences(n, u)
	}
}

func (f *StateFunctions) destroyWorkerGatherTarget(u *Unit) {
	if !f.utWorker(u.UnitType) || u.Worker.GatherTarget == nil {
		return
	}
	if u.Worker.IsGathering {
		f.gatherQueueNext(u, u.Worker.GatherTarget)
	} else {
		u.Worker.GatherTarget.Building.Resource.GatherQueue.Remove(u)
		u.Worker.GatherTarget = nil
	}
}

func (f *StateFunctions) destroyUnitImpl(u *Unit) {
	f.destroyUnitUnionData(u)
	f.destroyLoadedUnits(u)
	f.removeAllTargetReferencesTo(u)
	f.destroyWorkerGatherTarget(u)
	if f.utTurret(u.UnitType) {
		f.incrementUnitCounts(u, -1)
		u.Subunit = nil
		f.st.PlayerUnits[u.Owner].Remove(u)
		bwInsertList(&f.st.DeadUnits, u)
		return
	}
	tid := u.UnitType.ID
	if tid == UnitTypesTerranRefinery || tid == UnitTypesZergExtractor || tid == UnitTypesProtossAssimilator {
		f.revertRefineryToGeyser(u)
	} else {
		f.destroyRegularUnit(u)
	}
}

func (f *StateFunctions) revertRefineryToGeyser(u *Unit) {
	if u.UnitType.ID == UnitTypesZergExtractor {
		f.removeExtractorCreep(u.UnitType, u.Sprite.Position, true)
	}
	if f.uCompleted(u) {
		f.addCompletedUnit(u, -1, false)
	}
	f.uUnsetStatusFlag(u, UnitStatusFlagCompleted)
	f.morphUnit(u, f.getUnitType(UnitTypesResourceVespeneGeyser))
	u.OrderType = u.UnitType.HumanAiIdle.Get()
	u.OrderTarget = Target{}
	u.OrderState = 0
	u.OrderUnitType = nil
	u.Hp = u.UnitType.Hitpoints
	f.setUnitOwnerWith(u, 11, false, OwnerRefinery)
	f.setSpriteOwner(u, 11)
	f.uSetStatusFlag(u, UnitStatusFlagCompleted)
	f.addCompletedUnit(u, 1, true)
}

func (f *StateFunctions) destroyRegularUnitCloak(u *Unit) {
	if f.utBuilding(u.UnitType) {
		return
	}
	if u.CloakCounter != 0 {
		u.CloakCounter = 0
	}
	if isLinked(&f.st.CloakedUnits, u) {
		f.st.CloakedUnits.Remove(u)
		u.CloakedUnitLink = Link[Unit]{}
	}
	if f.uPassivelyCloaked(u) {
		f.uUnsetStatusFlag(u, UnitStatusFlagPassivelyCloaked)
	}
}

func (f *StateFunctions) destroyGroundedBuilding(u *Unit) {
	if !f.uGroundedBuilding(u) {
		return
	}
	prevCancelReason := f.swapCancelReason(CancelProducerDestroyed)
	f.cancelGroundedBuildingWork(u)
	f.cancelReason = prevCancelReason
}

func (f *StateFunctions) cancelGroundedBuildingWork(u *Unit) {
	if u.SecondaryOrderType.ID == OrdersBuildAddon {
		if u.CurrentBuildUnit != nil && !f.uCompleted(u.CurrentBuildUnit) {
			f.cancelBuildingUnit(u.CurrentBuildUnit)
		}
	}
	if u.Building.Addon != nil {
		f.buildingAbandonAddon(u)
	}
	if u.Building.ResearchingType != nil {
		f.cancelResearch(u, true)
	}
	if u.Building.UpgradingType != nil {
		f.cancelUpgrade(u, true)
	}
	if f.unitRace(u.UnitType) == RaceZerg {
		if f.unitIsMorphingBuilding(u) {
			f.uSetStatusFlag(u, UnitStatusFlagCompleted)
		}
		if !f.removeUnitCreepProvider(u) {
			f.uSetStatusFlag(u, UnitStatusFlag4000)
		}
	}
	if !u.BuildQueue.Empty() {
		f.cancelBuildQueue(u)
	}
}

func (f *StateFunctions) destroyRegularUnit(u *Unit) {
	f.stopUnit(u)
	if !f.usHidden(&u.Thingy) {
		f.unitFinderRemove(u)
		if f.uGroundedBuilding(u) {
			f.setUnitTilesUnoccupied(u.UnitType, u.Sprite.Position)
		}
		if f.uFlying(u) {
			f.decrementRepulseField(u)
		}
	}
	f.destroyRegularUnitCloak(u)
	f.destroyGroundedBuilding(u)
	if f.utFlyingBuilding(u.UnitType) && u.Building.IsLanding {
		f.setUnitTilesUnoccupied(u.UnitType, u.OrderTarget.Pos)
	}
	if u.Path != nil {
		f.freePath(u.Path)
		u.Path = nil
	}
	if u.CurrentBuildUnit != nil {
		if u.SecondaryOrderType.ID == OrdersTrain || u.SecondaryOrderType.ID == OrdersTrainFighter {
			f.killUnitWith(u.CurrentBuildUnit, noKiller(ReasonLinked))
		}
	}
	f.setSecondaryOrder(u, f.getOrderType(OrdersNothing))
	f.dropCarriedItems(u)
	f.onUnitDeselect(u)
	f.onUnitDestroy(u)
	f.incrementUnitCounts(u, -1)
	if f.uCompleted(u) {
		f.addCompletedUnit(u, -1, false)
	}
	f.st.PlayerUnits[u.Owner].Remove(u)
	if f.unitIsMapRevealer(u.UnitType) {
		f.st.MapRevealerUnits.Remove(u)
	} else if f.usHidden(&u.Thingy) {
		f.st.HiddenUnits.Remove(u)
	} else {
		f.st.VisibleUnits.Remove(u)
	}
	bwInsertList(&f.st.DeadUnits, u)
}

func (f *StateFunctions) destroyUnit(u *Unit) {
	turret := f.unitTurret(u)
	if turret != nil {
		f.destroyUnit(turret)
		u.Subunit = nil
	}
	f.destroyUnitImpl(u)
	u.OrderType = f.getOrderType(OrdersDie)
	u.OrderState = 1
	f.destroySprite(u.Sprite)
	u.Sprite = nil
}

func resetUnitTimersAndEffects(u *Unit) {
	u.RemoveTimer = 0
	u.DefensiveMatrixHp = Fp8Zero()
	u.DefensiveMatrixTimer = 0
	u.StimTimer = 0
	u.EnsnareTimer = 0
	u.LockdownTimer = 0
	u.IrradiateTimer = 0
	u.StasisTimer = 0
	u.PlagueTimer = 0
	u.StormTimer = 0
	u.IrradiatedBy = nil
	u.IrradiateOwner = 0
	u.ParasiteFlags = 0
	u.CycleCounter = 0
	u.BlindedBy = 0
	u.MaelstromTimer = 0
	u.AcidSporeCount = 0
	u.AcidSporeTime = [9]int32{}
}

func (f *StateFunctions) initializeUnit(u *Unit, unitType *UnitTypeT, pos Xy, owner int32) bool {
	u.OrderQueue.Clear()

	u.AutoTargetUnit = nil
	u.ConnectedUnit = nil

	u.OrderQueueCount = 0
	u.OrderProcessTimer = 0
	u.Unknown0x086 = 0
	u.AttackNotifyTimer = 0
	u.PreviousUnitType = nil
	u.LastEventTimer = 0
	u.LastEventColor = 0
	u.RankIncrease = 0
	u.KillCount = 0

	resetUnitTimersAndEffects(u)
	u.StatusFlags = 0
	u.UserActionFlags = 0
	u.PathingFlags = 0
	u.PreviousHp = 1

	u.Building.Rally = Target{}

	if !f.initializeUnitType(u, unitType, pos, owner) {
		return false
	}

	u.BuildQueue = StaticVector[*UnitTypeT, [5]*UnitTypeT]{}
	u.BuildQueueLimbo = StaticVector[*UnitTypeT, [5]*UnitTypeT]{}
	f.onBuildQueue(u, BuildQueueReset, 0)
	u.UnitIDGeneration++
	u.WireframeRandomizer = f.lcgRand(15) & 0xff
	if f.utTurret(u.UnitType) {
		u.Hp = Fp8FromRaw(1)
	} else {
		u.Hp = u.UnitType.Hitpoints.DivInt(10)
	}
	if f.uGroundedBuilding(u) {
		u.OrderType = u.UnitType.HumanAiIdle.Get()
	} else {
		u.OrderType = f.getOrderType(OrdersNothing)
	}
	f.setSecondaryOrder(u, f.getOrderType(OrdersNothing))
	u.UnitFinderBoundingBox = Rect{From: Xy{X: -1, Y: -1}, To: Xy{X: -1, Y: -1}}
	f.st.PlayerUnits[owner].PushFront(u)
	f.incrementUnitCounts(u, 1)

	if f.uGroundedBuilding(u) {
		f.unitFinderInsert(u)
		f.setUnitTilesOccupied(u.UnitType, u.Sprite.Position)
		f.checkUnitCollision(u)
		if f.uFlying(u) {
			f.incrementRepulseField(u)
		}
		f.setConstructionGraphic(u, true)
		f.setSpriteVisibility(u.Sprite, 0)
	} else {
		u.Sprite.Flags |= SpriteFlagHidden
		f.setSpriteVisibility(u.Sprite, 0)
	}
	u.DetectedFlags = 0xffffffff
	if f.utTurret(u.UnitType) {
		u.Sprite.Flags |= SpriteFlagTurret
	} else if !f.usHidden(&u.Thingy) {
		f.refreshUnitVision(u)
	}

	if f.unitIsVulture(u.UnitType) {
		u.Vulture.SpiderMineCount = 3
	}

	return true
}

func (f *StateFunctions) createNewUnit(unitType *UnitTypeT, pos Xy, owner int32) *Unit {
	u := f.st.UnitsContainer.Top()
	if u == nil {
		f.st.LastError = 61
		return nil
	}
	if !f.isUnitTypeInMapBounds(unitType, pos) {
		f.st.LastError = 0
		return nil
	}
	if !f.initializeUnit(u, unitType, pos, owner) {
		f.st.LastError = 62
		return nil
	}
	f.st.UnitsContainer.Pop()
	return u
}

func (f *StateFunctions) createUnit(unitType *UnitTypeT, pos Xy, owner int32) *Unit {
	if unitType == nil {
		throwError("attempt to create unit of null type")
	}

	f.lcgRand(14)
	u := f.createNewUnit(unitType, pos, owner)
	if u == nil {
		return nil
	}
	if f.uGroundedBuilding(u) {
		bwInsertList(&f.st.VisibleUnits, u)
	} else {
		bwInsertList(&f.st.HiddenUnits, u)
	}

	if unitType.ID < UnitTypesTerranCommandCenter && unitType.TurretUnitType.Valid() {
		su := f.createNewUnit(unitType.TurretUnitType.Get(), pos, owner)
		if su == nil {
			f.destroyUnit(u)
			return nil
		}
		u.Subunit = su
		su.Subunit = u
		f.setImageOffset(su.Sprite.MainImage, f.getImageLoOffset(u.Sprite.MainImage, 2, 0, false))
		if f.utTurret(u.UnitType) {
			throwError("unit type %d has a turret but is also flagged as a turret", int32(u.UnitType.ID))
		}
		if !f.utTurret(su.UnitType) {
			throwError("unit type %d was created as a turret but is not flagged as one", int32(su.UnitType.ID))
		}
	} else {
		u.Subunit = nil
	}
	return u
}

func (f *StateFunctions) createUnitByID(unitTypeID UnitTypes, pos Xy, owner int32) *Unit {
	if uint(unitTypeID) >= 228 {
		throwError("attempt to create unit with invalid id %d", unitTypeID)
	}
	return f.createUnit(f.getUnitType(unitTypeID), pos, owner)
}

func (f *StateFunctions) replaceSpriteImages(sprite *Sprite, newImageType *ImageTypeT, heading DirectionT) {
	for image := sprite.Images.Begin(); image != nil; {
		next := sprite.Images.Next(image)
		f.destroyImage(image)
		image = next
	}

	f.createImage(newImageType, sprite, Xy{}, imageOrderAbove, nil)

	for img := range sprite.Images.All() {
		f.setImageHeading(img, heading)
	}
}

func (f *StateFunctions) reapplyTimedEffect(timer *int32, apply func()) {
	if *timer == 0 {
		return
	}
	saved := *timer
	*timer = 0
	apply()
	*timer = saved
}

func (f *StateFunctions) applyUnitEffects(u *Unit) {
	if u.DefensiveMatrixTimer != 0 {
		if f.uInvincible(u) || !f.unitTechTargetValid(u, f.getTechType(TechTypesDefensiveMatrix), u) {
			u.DefensiveMatrixHp = Fp8Zero()
			u.DefensiveMatrixTimer = 0
		} else {
			f.createDefensiveMatrixImage(u)
		}
	}

	f.reapplyTimedEffect(&u.LockdownTimer, func() { f.lockdownUnit(u) })
	f.reapplyTimedEffect(&u.MaelstromTimer, func() { f.maelstromUnit(u) })
	if u.IrradiateTimer != 0 {
		if !f.weaponCanTargetUnit(f.getWeaponType(WeaponTypesIrradiate), u) {
			u.IrradiateTimer = 0
			u.IrradiatedBy = nil
		} else {
			f.reapplyTimedEffect(&u.IrradiateTimer, func() { f.irradiateUnit(u, u.IrradiatedBy, u.IrradiateOwner) })
		}
	}
	f.reapplyTimedEffect(&u.EnsnareTimer, func() { f.ensnareUnit(u) })
	f.reapplyTimedEffect(&u.StasisTimer, func() { f.stasisUnit(u) })
	f.reapplyTimedEffect(&u.PlagueTimer, func() { f.plagueUnit(u) })
	if u.AcidSporeCount != 0 {
		target := u
		if u.Subunit != nil {
			target = u.Subunit
		}
		f.createImage(f.getImageType(f.acidSporeImage(u)), target.Sprite, Xy{}, imageOrderTop, nil)
	}
}

func (f *StateFunctions) damageOverlayStates(u *Unit) int32 {
	imageType := u.Sprite.MainImage.ImageType
	if imageType.DamageFilenameIndex == 0 {
		return 0
	}
	var n int32
	for _, v := range (*f.globalSt.ImageLoOffsets[uint(imageType.ID)][1])[0] {
		if v != (Xy{X: 127, Y: 127}) {
			n++
		}
	}
	return 2 * n
}

func (f *StateFunctions) setConstructionGraphic(u *Unit, animated bool) {
	requiresDetectorOrCloaked := f.uRequiresDetector(u) || f.uCloaked(u)
	var modifierData1, modifierData2 int32
	if requiresDetectorOrCloaked {
		modifierData1 = u.Sprite.MainImage.ModifierData1
		modifierData2 = u.Sprite.MainImage.ModifierData2
	}
	prevIscriptUnit := f.iscriptUnit
	f.iscriptUnit = u
	defer func() { f.iscriptUnit = prevIscriptUnit }()

	constructionImage := u.UnitType.ConstructionAnimation.Get()
	if !animated || constructionImage == nil || int32(constructionImage.ID) == 0 {
		constructionImage = u.Sprite.SpriteType.Image.Get()
	}
	f.replaceSpriteImages(u.Sprite, constructionImage, u.Heading)

	if requiresDetectorOrCloaked {
		f.setSpriteCloakModifier(u.Sprite, f.uRequiresDetector(u), f.uCloaked(u), f.uBurrowed(u), modifierData1, modifierData2)
	}

	f.applyUnitEffects(u)
	u.DamageOverlayState = f.damageOverlayStates(u)
}

func (f *StateFunctions) setUnitHeading(u *Unit, heading DirectionT) {
	u.NextVelocityDirection = heading
	u.Heading = heading
	u.CurrentVelocityDirection = heading
	u.Velocity = f.directionXY(heading, u.CurrentSpeed)
	f.setNextTargetWaypoint(&u.Flingy, u.Sprite.Position)
	for img := range u.Sprite.Images.All() {
		f.setImageHeading(img, heading)
	}
}

func (f *StateFunctions) finishBuildingUnit(u *Unit) {
	if u.RemainingBuildTime != 0 {
		u.Hp = u.UnitType.Hitpoints
		u.ShieldPoints = Fp8Integer(u.UnitType.ShieldPoints)
		u.RemainingBuildTime = 0
	}
	if f.uGroundedBuilding(u) {
		u.ParasiteFlags = 0
		u.BlindedBy = 0
		f.setConstructionGraphic(u, false)
		f.spriteRunAnim(u.Sprite, IscriptAnimsBuilt)
		return
	}
	if f.uCanTurn(u) {
		dir := u.UnitType.UnitDirection
		if dir == 32 {
			dir = f.lcgRand(36) % 32
		}
		f.setUnitHeading(u, DirectionTFromRaw(8).MulInt(dir))
	}
	if f.unitIsTrap(u.UnitType) {
		f.showUnit(u)
	}
}

func (f *StateFunctions) moveUnit(u *Unit, pos Xy) {
	pos = f.restrictMoveTargetToValidBounds(u.UnitType, pos)
	u.Position = pos
	u.ExactPosition = f.toXyFp8(pos)
	f.moveSprite(u.Sprite, pos)
	if u.OrderType.ID != OrdersDie {
		f.setUnitMoveTarget(u, pos)
		f.setNextTargetWaypoint(&u.Flingy, pos)
	}
	if !f.utTurret(u.UnitType) {
		f.unitFinderReinsert(u)
		if !f.usHidden(&u.Thingy) {
			f.checkUnitCollision(u)
		}
	}
}

func (f *StateFunctions) placeCompletedUnit(u *Unit) bool {
	if !f.usHidden(&u.Thingy) {
		return true
	}
	res, pos := f.findUnitPlacement(u, u.Sprite.Position, false)
	if !res {
		f.displayLastErrorForPlayer(u.Owner)
		return false
	}
	f.moveUnit(u, pos)
	turret := f.unitTurret(u)
	if turret != nil {
		f.addCompletedUnit(turret, 1, false)
		f.uSetStatusFlag(turret, UnitStatusFlagCompleted)
		turret.Hp = turret.UnitType.Hitpoints
		f.moveUnit(turret, pos)
	}
	return true
}

func (f *StateFunctions) addCompletedUnit(u *Unit, count int32, incrementScore bool) {
	if f.uHallucination(u) {
		return
	}
	if f.utTurret(u.UnitType) {
		return
	}
	st := f.st
	ut := u.UnitType
	st.CompletedUnitCounts[u.Owner][ut.ID] += count
	race := f.unitRace(ut)
	if race != RaceNone {
		st.SupplyAvailable[u.Owner][uint(race)] = st.SupplyAvailable[u.Owner][uint(race)].Add(ut.SupplyProvided.MulInt(count))
	}

	if ut.GroupFlags&GroupFlagsFactory != 0 {
		st.CompletedFactoryCounts[u.Owner] += count
	}
	if ut.GroupFlags&GroupFlagsMen != 0 {
		st.CompletedNonBuildingCounts[u.Owner] += count
	} else if ut.GroupFlags&GroupFlagsBuilding != 0 {
		st.CompletedBuildingCounts[u.Owner] += count
	}

	if incrementScore && u.Owner != 11 {
		f.addCompletedUnitScore(u, count)
	}

	if st.CompletedUnitCounts[u.Owner][ut.ID] < 0 {
		st.CompletedUnitCounts[u.Owner][ut.ID] = 0
	}
}

func (f *StateFunctions) unitIsMorphedNonBuilding(u *Unit) bool {
	for _, id := range []UnitTypes{
		UnitTypesZergGuardian,
		UnitTypesZergDevourer,
		UnitTypesProtossDarkArchon,
		UnitTypesProtossArchon,
		UnitTypesZergLurker,
	} {
		if f.unitIs(u.UnitType, id) {
			return true
		}
	}
	return false
}

func (f *StateFunctions) addCompletedUnitScore(u *Unit, count int32) {
	st := f.st
	ut := u.UnitType
	if ut.GroupFlags&GroupFlagsMen != 0 {
		if !f.unitIsMorphedNonBuilding(u) {
			st.TotalNonBuildingsEverCompleted[u.Owner] += count
		}
		st.UnitScore[u.Owner] += ut.BuildScore * count
	} else if ut.GroupFlags&GroupFlagsBuilding != 0 {
		if !f.unitTypeIsMorphingBuilding(ut) {
			st.TotalBuildingsEverCompleted[u.Owner] += count
		}
		st.BuildingScore[u.Owner] += ut.BuildScore * count
	}
}

func (f *StateFunctions) repulseIndex(pos Xy) uint {
	ux := uint(pos.X)
	uy := uint(pos.Y)
	ux /= 48
	uy /= 48
	return uy*f.gameSt.RepulseFieldWidth + ux
}

func (f *StateFunctions) unitAffectsRepulseField(u *Unit) bool {
	if !f.uCanMove(u) {
		return false
	}
	if f.utBuilding(u.UnitType) {
		return false
	}
	if f.unitIs(u.UnitType, UnitTypesProtossInterceptor) {
		return false
	}
	return true
}

func (f *StateFunctions) incrementRepulseField(u *Unit) {
	if !f.unitAffectsRepulseField(u) {
		return
	}
	if u.RepulseFlags&0xf0 == 0 {
		v := uint32(f.lcgRand(37))
		u.RepulseDirection = f.directionFromIndex(uint(v & 0xff))
		v = (v >> 8) & 0xf0
		if v == 0 {
			v = 0xf0
		}
		u.RepulseFlags = int32(v)
	}
	u.RepulseFlags &^= 7
	u.RepulseIndex = f.repulseIndex(u.Sprite.Position)
	v := &f.st.RepulseField[u.RepulseIndex]
	if *v < 0xff {
		*v++
	}
}

func (f *StateFunctions) decrementRepulseField(u *Unit) {
	if !f.unitAffectsRepulseField(u) {
		return
	}
	v := &f.st.RepulseField[u.RepulseIndex]
	if *v > 0 {
		*v--
	}
}

func (f *StateFunctions) showUnit(u *Unit) {
	if !f.usHidden(&u.Thingy) {
		return
	}
	u.Sprite.Flags &^= SpriteFlagHidden
	turret := f.unitTurret(u)
	if turret != nil {
		turret.Sprite.Flags &^= SpriteFlagHidden
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
	f.st.HiddenUnits.Remove(u)
	bwInsertList(&f.st.VisibleUnits, u)
}

func (f *StateFunctions) hideUnit(u *Unit, deselect bool) {
	if f.usHidden(&u.Thingy) {
		return
	}
	for n := range f.st.VisibleUnits.All() {
		f.removeTargetReferences(n, u)
	}
	for n := range f.st.ActiveBullets.All() {
		f.removeBulletTargetReferences(n, u)
	}
	f.unitFinderRemove(u)
	if f.uGroundedBuilding(u) {
		f.setUnitTilesUnoccupied(u.UnitType, u.Sprite.Position)
	}
	if f.uFlying(u) {
		f.decrementRepulseField(u)
	}
	f.resetMovementState(u)
	f.st.VisibleUnits.Remove(u)
	bwInsertList(&f.st.HiddenUnits, u)

	u.Sprite.Flags |= SpriteFlagHidden
	f.setSpriteVisibility(u.Sprite, 0)
	turret := f.unitTurret(u)
	if turret != nil {
		turret.Sprite.Flags |= SpriteFlagHidden
		f.setSpriteVisibility(turret.Sprite, 0)
	}

	if deselect {
		f.onUnitDeselect(u)
	}
}

func (f *StateFunctions) setSpriteCloakModifier(sprite *Sprite, requiresDetector bool, cloaked bool, burrowed bool, data1 int32, data2 int32) {
	for image := range sprite.Images.All() {
		if image.ImageType.AlwaysVisible {
			continue
		}
		f.hideImage(image)
	}
	switch {
	case burrowed:
		for image := range sprite.Images.All() {
			if image.Modifier >= 2 && image.Modifier <= 7 {
				f.setImageModifier(image, 0)
			}
		}
	case requiresDetector && !cloaked:
		setUnmodifiedImagesModifier(sprite, 2, data1, data2)
	case !requiresDetector && cloaked:
		setUnmodifiedImagesModifier(sprite, 4, data1, data2)
	default:
		for image := range sprite.Images.All() {
			if image.Modifier == 0 || image.Modifier == 2 || image.Modifier == 4 {
				image.Modifier = 3
			}
		}
	}
}

func setUnmodifiedImagesModifier(sprite *Sprite, modifier int32, data1 int32, data2 int32) {
	for image := range sprite.Images.All() {
		if image.Modifier == 0 {
			image.Modifier = modifier
			image.ModifierData1 = data1
			image.ModifierData2 = data2
		}
	}
}

func (f *StateFunctions) cloakUnitImages(u *Unit) {
	f.uUnsetStatusFlag(u, UnitStatusFlagCloaked)
	f.uSetStatusFlag(u, UnitStatusFlagRequiresDetector)
	for image := range u.Sprite.Images.All() {
		if image.Modifier == 0 || image.Modifier == 4 || image.Modifier == 7 {
			image.Modifier = 2
			image.ModifierData1 = 0
			image.ModifierData2 = 3
		}
		if !image.ImageType.AlwaysVisible {
			f.hideImage(image)
		}
	}
	u.DetectedFlags = 0x80000000
}

func (f *StateFunctions) cloakUnit(u *Unit) {
	if f.uBurrowed(u) {
		if !f.uRequiresDetector(u) {
			f.setUnitCloaked(u)
			f.uSetStatusFlag(u, UnitStatusFlagRequiresDetector)
			f.uSetStatusFlag(u, UnitStatusFlagCloaked)
			f.setSpriteCloakModifier(u.Sprite, true, true, true, 0, 0)
			f.setSecondaryOrder(u, f.getOrderType(OrdersCloak))
			u.DetectedFlags = 0x80000000
		}
		f.updateUnitDetectedFlags(u)
		u.SecondaryOrderTimer = 30
		return
	}
	f.playSoundUnit(273, u, false)
	f.cloakUnitImages(u)
	if u.Subunit != nil {
		f.cloakUnitImages(u.Subunit)
	}
	f.updateUnitDetectedFlags(u)
	u.SecondaryOrderTimer = 30
}

func (f *StateFunctions) decloakUnitImages(u *Unit) {
	f.uUnsetStatusFlag(u, UnitStatusFlagRequiresDetector)
	f.uSetStatusFlag(u, UnitStatusFlagCloaked)
	for image := range u.Sprite.Images.All() {
		f.showImage(image)
		if image.Modifier == 2 || image.Modifier == 3 {
			f.setImageModifier(image, 4)
			image.ModifierData1 = 8
			image.ModifierData2 = 3
		} else if image.Modifier == 5 || image.Modifier == 6 {
			f.setImageModifier(image, 7)
			image.ModifierData1 = 8
			image.ModifierData2 = 3
		}
	}
}

func (f *StateFunctions) decloakUnit(u *Unit) {
	if f.uBurrowed(u) {
		f.uUnsetStatusFlag(u, UnitStatusFlagRequiresDetector)
		f.uUnsetStatusFlag(u, UnitStatusFlagCloaked)
		for image := range u.Sprite.Images.All() {
			if !image.ImageType.AlwaysVisible {
				f.showImage(image)
			}
			if image.Modifier >= 2 && image.Modifier <= 7 {
				f.setImageModifier(image, 0)
			}
		}
		return
	}
	f.playSoundUnit(274, u, false)
	f.decloakUnitImages(u)
	if u.Subunit != nil {
		f.decloakUnitImages(u.Subunit)
	}
}

func (f *StateFunctions) completeUnit(u *Unit) {
	if f.utTwoUnitsInOneEgg(u.UnitType) {
		f.incrementUnitCounts(u, -1)
		f.uSetStatusFlag(u, UnitStatusFlagCompleted)
		f.incrementUnitCounts(u, 1)
	} else {
		f.uSetStatusFlag(u, UnitStatusFlagCompleted)
	}
	f.addCompletedUnit(u, 1, true)
	if f.utInitiallyCloaked(u.UnitType) {
		f.cloakUnit(u)
	}
	if f.unitIsMapRevealer(u.UnitType) {
		f.st.HiddenUnits.Remove(u)
		bwInsertList(&f.st.MapRevealerUnits, u)
		f.refreshUnitVision(u)
	} else if f.usHidden(&u.Thingy) {
		if !f.unitIs(u.UnitType, UnitTypesProtossInterceptor) && !f.unitIs(u.UnitType, UnitTypesProtossScarab) {
			f.showUnit(u)
		}
	}
	if f.unitIsTrap(u.UnitType) {
		f.uSetStatusFlag(u, UnitStatusFlagCloaked)
		f.uSetStatusFlag(u, UnitStatusFlagRequiresDetector)
		u.DetectedFlags = 0x80000000
		u.SecondaryOrderTimer = 0
	}
	controller := f.st.Players[u.Owner].Controller
	switch controller {
	case PlayerControllerRescuePassive:
		throwError("fixme rescue passive")
	case PlayerControllerNeutral:
		f.setUnitOrder(u, f.getOrderType(OrdersNeutral))
	case PlayerControllerComputerGame:
		f.setUnitOrder(u, u.UnitType.ComputerAiIdle.Get())
	default:
		f.setUnitOrder(u, u.UnitType.HumanAiIdle.Get())
	}
	if f.utFlag(u.UnitType, UnitTypeFlagsT(0x800)) {
		throwError("fixme unknown flag")
	}
	u.AirStrength = f.getUnitStrength(u, false)
	u.GroundStrength = f.getUnitStrength(u, true)
	f.onUnitCompleted(u)
}

func (f *StateFunctions) createInitialUnit(unitType *UnitTypeT, pos Xy, owner int32) *Unit {
	u := f.createUnit(unitType, pos, owner)
	if u == nil {
		f.displayLastErrorForPlayer(owner)
		return nil
	}
	f.onUnitCreated(u, SourceMap, nil)
	if f.unitTypeSpreadsCreep(unitType, true) || f.utRequiresCreep(unitType) {
		f.spreadCreepCompletely(u.UnitType, u.Sprite.Position)
	}
	f.finishBuildingUnit(u)
	if !f.placeCompletedUnit(u) {
		throwError("place_completed_unit failed")
	}

	f.completeUnit(u)

	return u
}

func (f *StateFunctions) createCompletedUnit(unitType *UnitTypeT, pos Xy, owner int32) *Unit {
	u := f.createUnit(unitType, pos, owner)
	if u == nil {
		return nil
	}
	f.finishBuildingUnit(u)
	f.completeUnit(u)
	return u
}

func (f *StateFunctions) placementBlocksUnit(u *Unit, target *Unit) bool {
	if f.unitDead(target) {
		return false
	}
	if f.uNoCollide(u) || f.uNoCollide(target) {
		return false
	}
	if f.utPowerup(u.UnitType) {
		return f.uGroundedBuilding(target) || f.utPowerup(target.UnitType)
	}
	if f.unitIs(target.UnitType, UnitTypesTerranStarport) {
		if f.unitIsAnyOf(u, UnitTypesTerranWraith, UnitTypesTerranDropship, UnitTypesTerranScienceVessel, UnitTypesTerranBattlecruiser, UnitTypesTerranValkyrie) {
			return true
		}
	}
	if f.unitIs(target.UnitType, UnitTypesProtossRoboticsFacility) {
		if f.unitIsAnyOf(u, UnitTypesProtossShuttle, UnitTypesProtossObserver) {
			return true
		}
	}
	if f.unitIs(target.UnitType, UnitTypesProtossStargate) {
		if f.unitIsAnyOf(u, UnitTypesProtossScout, UnitTypesProtossCarrier, UnitTypesProtossArbiter, UnitTypesProtossCorsair) {
			return true
		}
	}
	return f.uFlying(u) == f.uFlying(target)
}

func (f *StateFunctions) unitIsAnyOf(u *Unit, ids ...UnitTypes) bool {
	for _, id := range ids {
		if f.unitIs(u.UnitType, id) {
			return true
		}
	}
	return false
}

func roundDownTo8(v int32) int32 {
	return int32(uint32(v) / 8 * 8)
}

func roundUpTo8(v int32) int32 {
	return int32((uint32(v) + 7) / 8 * 8)
}

type unitPlacementSearch struct {
	f          *StateFunctions
	u          *Unit
	bounds     Rect
	blocking   func(*Unit) bool
	findResult Xy
}

func (s *unitPlacementSearch) tryPosition(pos Xy) bool {
	f := s.f
	if f.isReachableUnitToPos(s.u, pos) {
		if f.uFlying(s.u) || f.unitTypeCanFitAt(s.u.UnitType, pos) {
			s.findResult = pos
			return true
		}
	}
	return false
}

func (s *unitPlacementSearch) unitWidthRoundedUp() int32 {
	d := s.u.UnitType.Dimensions
	return roundUpTo8(d.From.X + d.To.X + 1)
}

func (s *unitPlacementSearch) scanTopRightward(bb Rect) bool {
	f := s.f
	searchBB := f.unitInnerBoundingBox(s.u, Xy{X: bb.From.X, Y: bb.To.Y})
	for x := bb.From.X; x <= bb.To.X; {
		if f.isInnerBbMoveTargetInValidBounds(searchBB) {
			n := f.findUnit(searchBB, s.blocking)
			if n != nil {
				inc := n.Sprite.Position.X + n.UnitType.Dimensions.To.X + 1 - searchBB.From.X
				inc += (8 - ((x + inc) & 7)) & 7
				searchBB.From.X += inc
				searchBB.To.X += inc
				x += inc
				continue
			}
			if s.tryPosition(Xy{X: x, Y: bb.To.Y}) {
				return true
			}
		}
		searchBB.From.X += 8
		searchBB.To.X += 8
		x += 8
	}
	return false
}

func (s *unitPlacementSearch) scanRightUpward(bb Rect) bool {
	f := s.f
	searchBB := f.unitInnerBoundingBox(s.u, Xy{X: bb.To.X, Y: bb.To.Y})
	for y := bb.To.Y; y >= bb.From.Y; {
		if f.isInnerBbMoveTargetInValidBounds(searchBB) {
			n := f.findUnit(searchBB, s.blocking)
			if n != nil {
				dec := searchBB.To.Y - (n.Sprite.Position.Y - n.UnitType.Dimensions.From.Y - 1)
				dec += (8 - ((y - dec) & 7)) & 7
				searchBB.From.Y -= dec
				searchBB.To.Y -= dec
				y -= dec
				continue
			}
			if s.tryPosition(Xy{X: bb.To.X, Y: y}) {
				return true
			}
		}
		searchBB.From.Y -= 8
		searchBB.To.Y -= 8
		y -= 8
	}
	return false
}

func (s *unitPlacementSearch) scanBottomLeftward(bb Rect) bool {
	f := s.f
	searchBB := f.unitInnerBoundingBox(s.u, Xy{X: bb.To.X, Y: bb.From.Y})
	for x := bb.To.X; x >= bb.From.X; {
		if f.isInnerBbMoveTargetInValidBounds(searchBB) {
			n := f.findUnit(searchBB, s.blocking)
			if n != nil {
				dec := searchBB.To.X - (n.Sprite.Position.X - n.UnitType.Dimensions.From.X - 1)
				dec += (8 - ((x - dec) & 7)) & 7
				searchBB.From.X -= dec
				searchBB.To.X -= dec
				x -= dec
				continue
			}
			if s.tryPosition(Xy{X: x, Y: bb.From.Y}) {
				return true
			}
		}
		searchBB.From.X -= 8
		searchBB.To.X -= 8
		x -= 8
	}
	return false
}

func (s *unitPlacementSearch) scanLeftDownward(bb Rect) bool {
	f := s.f
	searchBB := f.unitInnerBoundingBox(s.u, Xy{X: bb.From.X, Y: bb.From.Y})
	for y := bb.From.Y; y <= bb.To.Y; {
		if f.isInnerBbMoveTargetInValidBounds(searchBB) {
			n := f.findUnit(searchBB, s.blocking)
			if n != nil {
				inc := n.Sprite.Position.Y + n.UnitType.Dimensions.To.Y + 1 - searchBB.From.Y
				inc += (8 - ((y + inc) & 7)) & 7
				searchBB.From.Y += inc
				searchBB.To.Y += inc
				y += inc
				continue
			}
			if s.tryPosition(Xy{X: bb.From.X, Y: y}) {
				return true
			}
		}
		searchBB.From.Y += 8
		searchBB.To.Y += 8
		y += 8
	}
	return false
}

func (s *unitPlacementSearch) find(pos Xy, width int32, height int32) bool {
	var bb Rect
	bb.From.X = roundDownTo8(pos.X)
	bb.From.Y = roundDownTo8(pos.Y)
	bb.To.X = roundUpTo8(pos.X + width)
	bb.To.Y = roundUpTo8(pos.Y + height)
	if bb.From.X < s.bounds.From.X {
		bb.From.X = s.bounds.From.X
	}
	if bb.From.Y < s.bounds.From.Y {
		bb.From.Y = s.bounds.From.Y
	}
	if bb.To.X > s.bounds.To.X {
		bb.To.X = s.bounds.To.X
	}
	if bb.To.Y > s.bounds.To.Y {
		bb.To.Y = s.bounds.To.Y
	}

	d := s.u.UnitType.Dimensions
	isWide := width > (d.From.X+d.To.X+1)*2
	if isWide {
		bb.From.X += s.unitWidthRoundedUp()
	}
	if s.scanTopRightward(bb) || s.scanRightUpward(bb) {
		return true
	}
	if isWide {
		bb.From.X -= s.unitWidthRoundedUp()
	}
	return s.scanBottomLeftward(bb) || s.scanLeftDownward(bb)
}

func (f *StateFunctions) findUnitPlacementInBounds(u *Unit, pos Xy, bounds Rect, terrainDisplacesUnit bool) (bool, Xy) {
	mapWidth := int32(f.gameSt.MapWidth)
	mapHeight := int32(f.gameSt.MapHeight)
	if bounds.From.X < 0 {
		bounds.From.X = 0
	}
	if bounds.From.Y < 0 {
		bounds.From.Y = 0
	}
	if bounds.To.X >= mapWidth {
		bounds.To.X = mapWidth - 1
	}
	if bounds.To.Y >= mapHeight-32 {
		bounds.To.Y = mapHeight - 32 - 1
	}
	s := &unitPlacementSearch{f: f, u: u, bounds: bounds}
	s.blocking = func(target *Unit) bool { return f.placementBlocksUnit(u, target) }

	blockingUnit := f.findUnit(f.unitInnerBoundingBox(u, pos), s.blocking)
	if blockingUnit == nil && f.isInBounds(f.unitInnerBoundingBox(u, pos), Rect{To: Xy{X: mapWidth, Y: mapHeight - 32}}) {
		if f.uFlying(u) || f.unitTypeCanFitAt(u.UnitType, pos) {
			return true, pos
		}
		if !terrainDisplacesUnit {
			f.st.LastError = 60
			return false, Xy{}
		}
	}

	width := int32(8)
	height := int32(8)
	if blockingUnit != nil {
		newWidth := u.UnitType.Dimensions.From.X + blockingUnit.UnitType.Dimensions.To.X + 2
		newHeight := u.UnitType.Dimensions.From.Y + blockingUnit.UnitType.Dimensions.To.Y + 2
		if newWidth > width {
			width = newWidth
		}
		if newHeight > height {
			height = newHeight
		}
	}
	for {
		npos := XySub(pos, Xy{X: width, Y: height})
		if npos.X < bounds.From.X && npos.Y < bounds.From.Y && pos.X+width > bounds.To.X && pos.Y+height > bounds.To.Y {
			break
		}
		if s.find(npos, width*2, height*2) {
			return true, s.findResult
		}
		width += 16
		height += 16
	}
	f.st.LastError = 60
	return false, Xy{}
}

func (f *StateFunctions) findUnitPlacement(u *Unit, pos Xy, terrainDisplacesUnit bool) (bool, Xy) {
	bounds := Rect{From: XySub(pos, Xy{X: 128, Y: 128}), To: XyAdd(pos, Xy{X: 127, Y: 127})}
	return f.findUnitPlacementInBounds(u, pos, bounds, terrainDisplacesUnit)
}

func (f *StateFunctions) removeQueuedOrder(u *Unit, o *Order) {
	if o.OrderType.Highlight != -1 {
		u.OrderQueueCount--
	}
	if u.OrderQueueCount == -1 {
		u.OrderQueueCount = 0
	}
	u.OrderQueue.Remove(o)
	f.st.OrdersContainer.Push(o)
	f.st.ActiveOrdersSize--
}

func (f *StateFunctions) createOrder(orderType *OrderTypeT, target OrderTarget) *Order {
	o := f.st.OrdersContainer.Top()
	if o == nil {
		return nil
	}
	f.st.OrdersContainer.Pop()
	f.st.ActiveOrdersSize++
	o.OrderType = orderType
	o.Target = target
	return o
}

func (f *StateFunctions) queueOrder(pos *Order, u *Unit, orderType *OrderTypeT, target OrderTarget) {
	o := f.createOrder(orderType, target)
	if o == nil {
		f.localUnitStatusError(u, 872)
		return
	}
	if o.OrderType.Highlight != -1 {
		u.OrderQueueCount++
	}
	u.OrderQueue.Insert(pos, o)
}

func (f *StateFunctions) queueOrderBack(u *Unit, orderType *OrderTypeT, target OrderTarget) {
	f.queueOrder(u.OrderQueue.End(), u, orderType, target)
}

func (f *StateFunctions) queueOrderFront(u *Unit, orderType *OrderTypeT, target OrderTarget) {
	f.queueOrder(u.OrderQueue.Begin(), u, orderType, target)
}

func (f *StateFunctions) removeOneOrder(u *Unit, orderType *OrderTypeT) {
	for o := range u.OrderQueue.All() {
		if o.OrderType == orderType {
			f.removeQueuedOrder(u, o)
			break
		}
	}
}

func (f *StateFunctions) setQueuedOrder(u *Unit, interrupt bool, orderType *OrderTypeT, target OrderTarget) {
	if u.OrderType.ID == OrdersDie {
		return
	}
	for !u.OrderQueue.Empty() {
		o := u.OrderQueue.Back()
		if o == nil {
			break
		}
		if (!interrupt || !o.OrderType.CanBeInterrupted) && o.OrderType != orderType {
			break
		}
		f.removeQueuedOrder(u, o)
	}
	if !f.canAllocateOrder(u, orderType) {
		return
	}
	if orderType.ID == OrdersCloak {
		throwError("cloak fixme")
	}
	f.queueOrderBack(u, orderType, target)
}

func orderPreventsQueueing(id Orders) bool {
	switch id {
	case OrdersGuard, OrdersPlayerGuard, OrdersNothing, OrdersPickupIdle, OrdersPatrol, OrdersMedicIdle:
		return true
	}
	return false
}

func (f *StateFunctions) issueOrder(u *Unit, queue bool, orderType *OrderTypeT, target OrderTarget) {
	if target.Unit != nil {
		target.Position = target.Unit.Sprite.Position
	}
	if queue && !orderType.CanBeQueued {
		queue = false
	}
	if queue && orderPreventsQueueing(u.OrderType.ID) {
		queue = false
	}
	if queue {
		if u.OrderQueueCount < 8 && f.st.ActiveOrdersSize < f.gameSt.Rules.OrderSupplyLowThreshold() {
			f.queueOrderBack(u, orderType, target)
		}
	} else {
		f.setUnitOrderTarget(u, orderType, target)
	}
}

func idleAnimationFor(animation int32, sid SpriteTypes) int32 {
	switch animation {
	case IscriptAnimsAirAttkInit, IscriptAnimsAirAttkRpt:
		return IscriptAnimsAirAttkToIdle
	case IscriptAnimsAlmostBuilt:
		if sid == SpriteTypesSPRITEIDSCV || sid == SpriteTypesSPRITEIDDrone || sid == SpriteTypesSPRITEIDProbe {
			return IscriptAnimsGndAttkToIdle
		}
	case IscriptAnimsGndAttkInit, IscriptAnimsGndAttkRpt:
		return IscriptAnimsGndAttkToIdle
	case IscriptAnimsSpecialState1:
		if sid == SpriteTypesSPRITEIDMedic {
			return IscriptAnimsWalkingToIdle
		}
	case IscriptAnimsCastSpell:
		return IscriptAnimsWalkingToIdle
	}
	return -1
}

func (f *StateFunctions) iscriptRunToIdle(u *Unit) {
	u.StatusFlags &^= UnitStatusFlagIscriptNobrk
	u.Sprite.Flags &^= SpriteFlagIscriptNobrk
	prevIscriptUnit := f.iscriptUnit
	prevIscriptFlingy := f.iscriptFlingy
	f.iscriptUnit = u
	f.iscriptFlingy = &u.Flingy
	defer func() {
		f.iscriptFlingy = prevIscriptFlingy
		f.iscriptUnit = prevIscriptUnit
	}()
	anim := idleAnimationFor(u.Sprite.MainImage.IscriptState.Animation, u.Sprite.SpriteType.ID)
	if anim != -1 {
		f.spriteRunAnim(u.Sprite, anim)
	}
	f.uUnsetMovementFlag(&u.Flingy, 8)
}

func (f *StateFunctions) turretOrderType(u *Unit, orderType *OrderTypeT) *OrderTypeT {
	switch {
	case orderType == u.UnitType.ReturnToIdle.Get():
		return u.Subunit.UnitType.ReturnToIdle.Get()
	case orderType == u.UnitType.AttackUnit.Get():
		return u.Subunit.UnitType.AttackUnit.Get()
	case orderType == u.UnitType.AttackMove.Get():
		return u.Subunit.UnitType.AttackMove.Get()
	case !orderType.ValidForTurret:
		return nil
	}
	return orderType
}

func (f *StateFunctions) activateNextOrder(u *Unit) {
	if u.OrderQueue.Empty() {
		return
	}
	front := u.OrderQueue.Front()
	if (f.uOrderNotInterruptible(u) || f.uIscriptNobrk(u)) && front.OrderType.ID != OrdersDie {
		return
	}
	orderType := front.OrderType
	target := front.Target
	f.removeQueuedOrder(u, front)

	u.UserActionFlags &^= 1
	f.uUnsetStatusFlag(u, UnitStatusFlag8)
	f.uUnsetStatusFlag(u, UnitStatusFlagReadyToAttack)
	f.uSetStatusFlagValue(u, UnitStatusFlagOrderNotInterruptible, !orderType.CanBeInterrupted)
	u.OrderProcessTimer = 0
	u.MoveTargetTimer = 0

	u.OrderType = orderType
	u.OrderState = 0

	if target.Unit != nil {
		if f.unitDead(target.Unit) || target.Unit.Sprite == nil {
			throwError("attempt to activate order with dead target")
		}
		u.OrderTarget.Unit = target.Unit
		u.OrderTarget.Pos = target.Unit.Sprite.Position
		u.OrderUnitType = nil
	} else {
		u.OrderTarget.Unit = nil
		u.OrderTarget.Pos = target.Position
		u.OrderUnitType = target.UnitType
	}
	u.AutoTargetUnit = nil
	f.iscriptRunToIdle(u)
	if !f.utTurret(u.UnitType) && u.Subunit != nil && f.utTurret(u.Subunit.UnitType) {
		turretOrderType := f.turretOrderType(u, orderType)
		if turretOrderType != nil {
			f.setUnitOrderTarget(u.Subunit, turretOrderType, target)
		}
	}
}

func (f *StateFunctions) setUnitOrder(u *Unit, orderType *OrderTypeT) {
	u.UserActionFlags |= 1
	f.setQueuedOrder(u, true, orderType, OrderTarget{})
	f.activateNextOrder(u)
}

func (f *StateFunctions) setUnitOrderTarget(u *Unit, orderType *OrderTypeT, target OrderTarget) {
	u.UserActionFlags |= 1
	f.setQueuedOrder(u, true, orderType, target)
	f.activateNextOrder(u)
}

func (f *StateFunctions) setUnitOrderUnit(u *Unit, orderType *OrderTypeT, targetUnit *Unit) {
	u.UserActionFlags |= 1
	var target OrderTarget
	target.Unit = targetUnit
	if targetUnit != nil {
		target.Position = targetUnit.Sprite.Position
	}
	f.setQueuedOrder(u, true, orderType, target)
	f.activateNextOrder(u)
}

func (f *StateFunctions) setUnitOrderPos(u *Unit, orderType *OrderTypeT, position Xy) {
	u.UserActionFlags |= 1
	var target OrderTarget
	target.Position = position
	f.setQueuedOrder(u, true, orderType, target)
	f.activateNextOrder(u)
}
