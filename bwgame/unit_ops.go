package bwgame

import "math"

func (f *StateFunctions) stopFlingy(u *Flingy) {
	if f.uMovementFlag(u, 4) {
		return
	}
	f.uUnsetMovementFlag(u, 0x10)
	f.uUnsetMovementFlag(u, 0x20)
	f.uUnsetMovementFlag(u, 0x40)
	u.MoveTarget.Pos = u.Position
	u.MoveTarget.Unit = nil
	u.NextMovementWaypoint = u.Position
	if f.uMovementFlag(u, 2) && u.FlingyMovementType == 0 {
		d := f.unitHaltDistance(u)
		u.MoveTarget.Pos = f.toXy(XyFp8Add(u.ExactPosition, f.directionXY(u.NextVelocityDirection, d)))
		u.NextMovementWaypoint = u.MoveTarget.Pos
		if !f.unitIsAtMoveTarget(u) {
			f.uSetMovementFlag(u, 4)
		}
	}
}

func (f *StateFunctions) stopUnit(u *Unit) {
	if u.PathingFlags&2 != 0 {
		u.PathingFlags |= 4
	}
	f.stopFlingy(&u.Flingy)
	moveTarget := f.restrictMoveTargetToValidBounds(u.UnitType, u.MoveTarget.Pos)
	if u.MoveTarget.Pos != moveTarget {
		u.MoveTarget.Pos = moveTarget
		u.NextMovementWaypoint = moveTarget
	}
	u.MoveTargetTimer = 15
}

func (f *StateFunctions) orderDone(u *Unit, target OrderTarget) {
	if !u.OrderQueue.Empty() {
		u.UserActionFlags |= 1
		f.activateNextOrder(u)
	} else {
		f.setUnitOrderTarget(u, u.UnitType.ReturnToIdle.Get(), target)
	}
}

func (f *StateFunctions) setNextTargetWaypoint(u *Flingy, pos Xy) {
	if u.NextTargetWaypoint == pos {
		return
	}
	u.NextTargetWaypoint = pos
}

func (f *StateFunctions) setUnitGathering(u *Unit) {
	if f.uGathering(u) {
		return
	}
	f.uSetStatusFlag(u, UnitStatusFlagGathering)
	f.queueOrderFront(u, f.getOrderType(OrdersResetHarvestCollision), OrderTarget{})
}

func tileCoordinateBelow(v int32) uint {
	return uint(uint32(v)/32 - 1)
}

func tileCoordinateAbove(v int32) uint {
	return uint(uint32(v)/32 + 1)
}

func (f *StateFunctions) tileIsUnoccupied(index uint) bool {
	return ^int32(f.st.Tiles[index].Flags)&TileFlagOccupied != 0
}

func (f *StateFunctions) anyNeighborTileUnoccupied(u *Unit) bool {
	unitBB := f.unitSpriteInnerBoundingBox(u)
	var tileBB RectT[XyT[uint]]
	tileBB.From.X = tileCoordinateBelow(unitBB.From.X)
	tileBB.From.Y = tileCoordinateBelow(unitBB.From.Y)
	tileBB.To.X = tileCoordinateAbove(unitBB.To.X)
	tileBB.To.Y = tileCoordinateAbove(unitBB.To.Y)
	mapTileWidth := f.gameSt.MapTileWidth
	mapTileHeight := f.gameSt.MapTileHeight
	if tileBB.From.X >= mapTileWidth {
		tileBB.From.X = 0
	}
	if tileBB.From.Y >= mapTileHeight {
		tileBB.From.Y = 0
	}
	if tileBB.To.X >= mapTileWidth {
		tileBB.To.X = mapTileWidth - 1
	}
	if tileBB.To.Y >= mapTileHeight {
		tileBB.To.Y = mapTileHeight - 1
	}
	dimensions := u.UnitType.Dimensions
	tileWidth := uint(uint32(dimensions.From.X+dimensions.To.X+1) / 32)
	tileHeight := uint(uint32(dimensions.From.Y+dimensions.To.Y+1) / 32)
	topIndex := tileBB.From.Y*mapTileWidth + (tileBB.From.X + 1)
	bottomIndex := tileBB.To.Y*mapTileWidth + (tileBB.From.X + 1)
	for i := uint(0); i != tileWidth; i++ {
		if f.tileIsUnoccupied(topIndex) {
			return true
		}
		if f.tileIsUnoccupied(bottomIndex) {
			return true
		}
		topIndex++
		bottomIndex++
	}
	leftIndex := (tileBB.From.Y+1)*mapTileWidth + tileBB.From.X
	rightIndex := (tileBB.From.Y+1)*mapTileWidth + tileBB.To.X
	for i := uint(0); i != tileHeight; i++ {
		if f.tileIsUnoccupied(leftIndex) {
			return true
		}
		if f.tileIsUnoccupied(rightIndex) {
			return true
		}
		leftIndex += mapTileWidth
		rightIndex += mapTileWidth
	}
	return false
}

func (f *StateFunctions) unitPosIsBorderingTarget(u *Unit, pos Xy, target *Unit) bool {
	bb := f.unitSpriteInnerBoundingBox(target)
	bb.From = XySub(bb.From, XyAdd(u.UnitType.Dimensions.To, Xy{X: 1, Y: 1}))
	bb.To.X += u.UnitType.Dimensions.From.X + 1
	bb.To.X += u.UnitType.Dimensions.From.Y + 1
	if pos.X == bb.From.X || pos.X == bb.To.X {
		if pos.Y >= bb.From.Y && pos.Y <= bb.To.Y {
			return true
		}
	} else if pos.Y == bb.From.Y || pos.Y == bb.To.Y {
		if pos.X >= bb.From.X && pos.X <= bb.To.X {
			return true
		}
	}
	return false
}

func (f *StateFunctions) moveToTarget(u *Unit, target *Unit) {
	if u.MoveTarget.Unit == target && (u.MoveTarget.Pos == target.Sprite.Position || f.unitPosIsBorderingTarget(u, u.MoveTarget.Pos, target)) {
		f.uUnsetStatusFlag(u, UnitStatusFlagImmovable)
	} else {
		f.setUnitMoveTargetUnit(u, target)
	}
}

func (f *StateFunctions) moveToTargetReset(u *Unit, target *Unit) {
	f.moveToTarget(u, target)
	f.setNextTargetWaypoint(&u.Flingy, target.Sprite.Position)
}

func (f *StateFunctions) tryGatherResource(u *Unit, target *Unit) bool {
	if target.Building.Resource.IsBeingGathered {
		return false
	}
	target.Building.Resource.IsBeingGathered = true
	u.Worker.GatherTarget = target
	u.Worker.IsGathering = true
	if u.OrderType.ID == OrdersWaitForGas {
		f.queueOrderFront(u, f.getOrderType(OrdersHarvestGas), OrderTarget{Position: target.Sprite.Position, Unit: target})
	} else {
		f.queueOrderFront(u, f.getOrderType(OrdersGatheringInterrupted), OrderTarget{})
		f.queueOrderFront(u, f.getOrderType(OrdersMiningMinerals), OrderTarget{Position: target.Sprite.Position, Unit: target})
	}
	f.orderDone(u, OrderTarget{})
	return true
}

func (f *StateFunctions) waitForResource(u *Unit, target *Unit) {
	u.OrderProcessTimer = 0
	u.OrderState = 2
	if !f.tryGatherResource(u, target) {
		if u.Worker.GatherTarget == nil {
			u.Worker.GatherTarget = target
			target.Building.Resource.GatherQueue.PushFront(u)
		}
		f.queueOrderFront(u, f.getOrderType(OrdersGatherWaitInterrupted), OrderTarget{})
	}
}

func (f *StateFunctions) headingError(heading DirectionT, u *Unit) Fp8 {
	return Fp8Extend(heading.Sub(f.xyDirection(XySub(u.NextTargetWaypoint, u.Sprite.Position)))).Abs()
}

func (f *StateFunctions) isFacingNextTargetWaypoint(u *Unit, angle Fp8) bool {
	if u.Position == u.NextTargetWaypoint {
		return true
	}
	if f.headingError(u.Heading, u).LessEqual(angle) {
		return true
	}
	return false
}

func (f *StateFunctions) longPathDistance(from Xy, to Xy) int32 {
	if !f.isReachable(from, to) {
		return 0x7fff
	}
	var pf Pathfinder
	if !f.pathfinderFindLongPathFromTo(&pf, from, to) {
		return 0x7ffe
	}
	if len(pf.LongPath) <= 2 {
		return f.xyLength(XySub(to, from))
	}
	pos := f.toXy(pf.LongPath[0].Center)
	var r int32
	for _, region := range pf.LongPath[1:] {
		nextPos := f.toXy(region.Center)
		r += f.xyLength(XySub(nextPos, pos))
		pos = nextPos
	}
	return r
}

func (f *StateFunctions) unitLongPathDistance(u *Unit, from Xy, to Xy) int32 {
	if ^u.PathingFlags&1 != 0 {
		return f.xyLength(XySub(to, from))
	}
	return f.longPathDistance(from, to)
}

func (f *StateFunctions) destroyCarryingImages(u *Unit) {
	f.destroyImageFromTo(u.Sprite, ImageTypesIMAGEIDMineralChunkShadow, ImageTypesIMAGEIDPsiEmitterShadowCarried)
	f.destroyImageFromTo(u.Sprite, ImageTypesIMAGEIDFlag, ImageTypesIMAGEIDTerranGasTankType2)
	f.destroyImageFromTo(u.Sprite, ImageTypesIMAGEIDUraj, ImageTypesIMAGEIDKhalis)
}

func (f *StateFunctions) carriedResourceImageType(resource *Unit) (imageType *ImageTypeT, isMinerals bool) {
	switch resource.UnitType.ID {
	case UnitTypesTerranRefinery:
		return f.getImageType(ImageTypesIMAGEIDTerranGasTankType1), false
	case UnitTypesProtossAssimilator:
		return f.getImageType(ImageTypesIMAGEIDProtossGasOrbType1), false
	case UnitTypesZergExtractor:
		return f.getImageType(ImageTypesIMAGEIDZergGasSacType1), false
	case UnitTypesResourceMineralField, UnitTypesResourceMineralFieldType2, UnitTypesResourceMineralFieldType3:
		return f.getImageType(ImageTypesIMAGEIDMineralChunkType1), true
	}
	return nil, false
}

func (f *StateFunctions) unitGatherResourcesFrom(u *Unit, resource *Unit) {
	imageType, isMinerals := f.carriedResourceImageType(resource)
	if imageType == nil {
		return
	}
	var gathered int32
	if resource.Building.Resource.ResourceCount < 8 {
		imageType = f.getImageType(imageType.ID + 1)
		if isMinerals {
			gathered = resource.Building.Resource.ResourceCount
			f.killUnitWith(resource, noKiller(ReasonDepleted))
		} else {
			resource.Building.Resource.ResourceCount = 0
			gathered = 2
		}
	} else {
		f.setUnitResources(resource, resource.Building.Resource.ResourceCount-8)
		gathered = 8
		if isMinerals {
			if resource.Building.Resource.ResourceCount == 0 {
				f.killUnitWith(resource, noKiller(ReasonDepleted))
			}
		}
	}
	if gathered == 0 {
		return
	}
	if u.CarryingFlags&3 != 0 {
		f.destroyCarryingImages(u)
		u.CarryingFlags = 0
	}
	if u.CarryingFlags != 0 {
		return
	}
	if isMinerals {
		u.CarryingFlags = 2
	} else {
		u.CarryingFlags = 1
	}
	u.Worker.ResourcesCarried = gathered
	image := f.createImage(imageType, u.Sprite, Xy{}, imageOrderAbove, nil)
	if image != nil && !f.iFlag(image, ImageFlagUsesSpecialOffset) {
		f.iSetFlag(image, ImageFlagUsesSpecialOffset)
		f.updateImageSpecialOffset(image)
	}
}

func (f *StateFunctions) gatherQueueNext(u *Unit, resource *Unit) {
	u.Worker.IsGathering = false
	u.Worker.GatherTarget = nil
	if resource == nil {
		return
	}
	resource.Building.Resource.IsBeingGathered = false
	var nextUnit *Unit
	for queuedUnit := range resource.Building.Resource.GatherQueue.Backward() {
		if !f.unitIsDisabled(queuedUnit) {
			if queuedUnit.OrderType.ID == OrdersWaitForGas || queuedUnit.OrderType.ID == OrdersWaitForMinerals {
				nextUnit = queuedUnit
				break
			}
		}
	}
	if nextUnit != nil {
		nextUnit.Worker.GatherTarget = nil
		resource.Building.Resource.GatherQueue.Remove(nextUnit)
		f.removeOneOrder(nextUnit, f.getOrderType(OrdersGatherWaitInterrupted))
		f.tryGatherResource(nextUnit, resource)
	}
}

func (f *StateFunctions) unitIsActiveResourceDepot(u *Unit) bool {
	if !f.uGroundedBuilding(u) {
		return false
	}
	if !f.utResourceDepot(u.UnitType) {
		return false
	}
	if f.uCompleted(u) {
		return true
	}
	if f.unitIs(u.UnitType, UnitTypesZergHive) {
		return true
	}
	if f.unitIs(u.UnitType, UnitTypesZergLair) {
		return true
	}
	if f.unitIs(u.UnitType, UnitTypesZergHatchery) && f.unitIsMorphingBuilding(u) {
		return true
	}
	return false
}

func (f *StateFunctions) findNearestActiveResourceDepot(u *Unit) *Unit {
	return f.findNearestUnitToUnit(u, f.mapBounds(), func(target *Unit) bool {
		if target == u {
			return false
		}
		if !f.unitIsActiveResourceDepot(target) {
			return false
		}
		if target.Owner != u.Owner {
			return false
		}
		if ^u.PathingFlags&1 != 0 || !f.isReachable(u.Sprite.Position, target.Sprite.Position) {
			return false
		}
		return true
	})
}

func (f *StateFunctions) dropCarriedItems(u *Unit) {
	if u.CarryingFlags&^3 == 0 {
		return
	}
	throwError("drop_carried_items: fixme")
}

func (f *StateFunctions) hasAvailableSupplyFor(owner int32, unitType *UnitTypeT, showError bool) bool {
	if f.utFlag(unitType, 0x1000000) {
		return true
	}
	supplyRequired := unitType.SupplyRequired
	if f.utTwoUnitsInOneEgg(unitType) {
		supplyRequired = supplyRequired.MulInt(2)
	}
	if supplyRequired == Fp1Zero() {
		return true
	}
	if unitType.ID == UnitTypesZergLurker {
		supplyRequired = supplyRequired.Sub(f.getUnitType(UnitTypesZergHydralisk).SupplyRequired)
	}
	race := f.unitRace(unitType)
	if race == RaceNone {
		return false
	}
	index := uint(race)
	if f.st.SupplyUsed[owner][index].Add(supplyRequired).Greater(Fp1Integer(200)) {
		return false
	}
	if f.st.SupplyUsed[owner][index].Add(supplyRequired).Greater(f.st.SupplyAvailable[owner][index]) {
		return false
	}
	return true
}

func (f *StateFunctions) playerHasResources(owner int32, minerals int32, gas int32) bool {
	if f.st.CurrentMinerals[owner] < minerals {
		return false
	}
	if f.st.CurrentGas[owner] < gas {
		return false
	}
	return true
}

func (f *StateFunctions) hasAvailableResourcesFor(owner int32, unitType *UnitTypeT, showError bool) bool {
	return f.playerHasResources(owner, unitType.MineralCost, unitType.GasCost)
}

func (f *StateFunctions) hasAvailableResourcesForTech(owner int32, techType *TechTypeT, showError bool) bool {
	return f.playerHasResources(owner, techType.MineralCost, techType.GasCost)
}

func (f *StateFunctions) upgradeMineralCost(owner int32, upgradeType *UpgradeTypeT) int32 {
	return upgradeType.MineralCostBase + upgradeType.MineralCostFactor*f.playerUpgradeLevel(owner, upgradeType.ID)
}

func (f *StateFunctions) upgradeGasCost(owner int32, upgradeType *UpgradeTypeT) int32 {
	return upgradeType.GasCostBase + upgradeType.GasCostFactor*f.playerUpgradeLevel(owner, upgradeType.ID)
}

func (f *StateFunctions) upgradeTimeCost(owner int32, upgradeType *UpgradeTypeT) int32 {
	return upgradeType.TimeCostBase + upgradeType.TimeCostFactor*f.playerUpgradeLevel(owner, upgradeType.ID)
}

func (f *StateFunctions) hasAvailableResourcesForUpgrade(owner int32, upgradeType *UpgradeTypeT, showError bool) bool {
	if f.st.CurrentMinerals[owner] < f.upgradeMineralCost(owner, upgradeType) {
		return false
	}
	if f.st.CurrentGas[owner] < f.upgradeGasCost(owner, upgradeType) {
		return false
	}
	return true
}

func (f *StateFunctions) buildQueuePush(u *Unit, unitType *UnitTypeT) bool {
	if u.BuildQueue.Size() >= 5 {
		f.onActionRejected(u, RejectQueueFull)
		return false
	}
	if u.BuildQueue.Empty() && !f.hasAvailableSupplyFor(u.Owner, unitType, true) {
		f.onActionRejected(u, RejectSupply)
		return false
	}
	if !f.hasAvailableResourcesFor(u.Owner, unitType, true) {
		f.onActionRejectedForCost(u, u.Owner, unitType.MineralCost)
		return false
	}
	index := u.BuildQueue.Size()
	u.BuildQueue.PushBack(unitType)
	for _, t := range u.BuildQueueLimbo.All() {
		u.BuildQueue.PushBack(t)
	}
	u.BuildQueueLimbo.Clear()
	f.onBuildQueue(u, BuildQueuePush, index)
	if !f.utBuilding(unitType) {
		f.st.CurrentMinerals[u.Owner] -= unitType.MineralCost
		f.st.CurrentGas[u.Owner] -= unitType.GasCost
	}
	return true
}

func (f *StateFunctions) buildQueueClearSingle(u *Unit) {
	if u.BuildQueue.Empty() {
		return
	}
	u.BuildQueue.Erase(0)
	u.BuildQueueLimbo = u.BuildQueue
	u.BuildQueue.Clear()
	f.onBuildQueue(u, BuildQueueClearSingle, 0)
}

func (f *StateFunctions) trainUnit(u *Unit, unitType *UnitTypeT) bool {
	if f.buildQueuePush(u, unitType) {
		f.setSecondaryOrder(u, f.getOrderType(OrdersTrain))
		return true
	}
	return false
}

func (f *StateFunctions) refundUnitCosts(owner int32, unitType *UnitTypeT) {
	f.st.CurrentMinerals[owner] += unitType.MineralCost
	f.st.CurrentGas[owner] += unitType.GasCost
}

func threeQuarters(cost int32) int32 {
	return cost * 3 / 4
}

func (f *StateFunctions) partiallyRefundUnitCosts(owner int32, unitType *UnitTypeT) {
	f.st.CurrentMinerals[owner] += threeQuarters(unitType.MineralCost)
	f.st.CurrentGas[owner] += threeQuarters(unitType.GasCost)
}

func (f *StateFunctions) cancelMorphingBuilding(u *Unit) {
	if f.unitIsMorphingBuilding(u) {
		f.cancelBuildingMorph(u)
	} else {
		f.cancelZergBuildingConstruction(u)
	}
}

func (f *StateFunctions) cancelBuildingMorph(u *Unit) {
	if u.BuildQueue.Empty() {
		throwError("cancel_morphing_building: build queue is empty")
	}
	buildType := u.BuildQueue.Front()
	f.buildQueueClearSingle(u)
	f.partiallyRefundUnitCosts(u.Owner, buildType)
	f.uSetStatusFlag(u, UnitStatusFlagCompleted)
	f.addCompletedUnit(u, -1, false)
	f.uUnsetStatusFlag(u, UnitStatusFlagCompleted)
	hp := u.Hp
	f.finishBuildingUnit(u)
	f.completeUnit(u)
	f.setUnitHp(u, hp)
	f.spriteRunAnim(u.Sprite, IscriptAnimsAlmostBuilt)
	f.playSoundUnit(4, u, false)
}

func (f *StateFunctions) cancelZergBuildingConstruction(u *Unit) {
	f.partiallyRefundUnitCosts(u.Owner, u.UnitType)
	f.st.UnitScore[u.Owner] -= u.UnitType.BuildScore
	f.st.TotalNonBuildingsEverCompleted[u.Owner]--
	prevHp := Fp8Integer(u.PreviousHp)
	if f.unitIs(u.UnitType, UnitTypesZergExtractor) {
		drone := f.createUnitByID(UnitTypesZergDrone, u.Sprite.Position, u.Owner)
		if drone != nil {
			f.onUnitCreated(drone, SourceDroneReturned, u)
			f.finishBuildingUnit(drone)
			f.completeUnit(drone)
			f.setUnitHp(drone, prevHp)
		} else {
			f.displayLastErrorForPlayer(u.Owner)
		}
		f.killUnitWith(u, noKiller(ReasonCancelled))
		return
	}
	t := f.createThingy(f.getSpriteType(SpriteTypesSPRITEIDZergBuildingSpawnSmall), u.Sprite.Position, 0)
	if t != nil {
		t.Sprite.ElevationLevel = u.Sprite.ElevationLevel + 1
		if !f.usHidden(t) {
			f.setSpriteVisibility(t.Sprite, int32(f.tileVisibility(t.Sprite.Position)))
		}
	}
	buildType := u.UnitType
	f.morphUnit(u, f.getUnitType(UnitTypesZergDrone))
	f.finishBuildingUnit(u)
	f.completeUnit(u)
	if u.Sprite.Images.Back().Modifier == 10 {
		u.Sprite.Images.Back().Offset.Y = 7
	}
	f.setUnitOrder(u, f.getOrderType(OrdersResetCollision))
	f.setQueuedOrder(u, false, u.UnitType.ReturnToIdle.Get(), OrderTarget{})
	f.setUnitHp(u, prevHp)
	f.removeCreepProvider(buildType, u.Sprite.Position, false)
	f.playSoundUnit(4, u, false)
}

func (f *StateFunctions) unitIsUncancellableMorph(u *Unit) bool {
	return f.unitIs(u.UnitType, UnitTypesZergGuardian) ||
		f.unitIs(u.UnitType, UnitTypesZergLurker) ||
		f.unitIs(u.UnitType, UnitTypesZergDevourer) ||
		f.unitIs(u.UnitType, UnitTypesZergMutalisk) ||
		f.unitIs(u.UnitType, UnitTypesZergHydralisk)
}

func (f *StateFunctions) cancelBuildingUnit(u *Unit) bool {
	if f.unitDying(u) {
		return false
	}
	if f.uCompleted(u) {
		return false
	}
	if f.unitIsUncancellableMorph(u) {
		return false
	}
	if f.unitIsNydus(u.UnitType) && u.Building.Nydus.Exit == nil {
		return false
	}
	f.onBuildingUnitCancelled(u)
	if f.uGroundedBuilding(u) {
		if f.unitRace(u.UnitType) == RaceZerg {
			f.cancelMorphingBuilding(u)
			return true
		}
		f.partiallyRefundUnitCosts(u.Owner, u.UnitType)
	} else {
		refundType := u.UnitType
		if f.unitIsEgg(u.UnitType) && !u.BuildQueue.Empty() {
			refundType = u.BuildQueue.Front()
		}
		f.refundUnitCosts(u.Owner, refundType)
	}
	var morphTo *UnitTypeT
	if f.unitIs(u.UnitType, UnitTypesZergCocoon) {
		morphTo = f.getUnitType(UnitTypesZergMutalisk)
	} else if f.unitIs(u.UnitType, UnitTypesZergLurkerEgg) {
		morphTo = f.getUnitType(UnitTypesZergHydralisk)
	}
	if morphTo != nil {
		f.revertEggMorph(u, morphTo)
	} else {
		if f.unitIs(u.UnitType, UnitTypesTerranNuclearMissile) && u.ConnectedUnit != nil {
			if f.unitIsGhost(u.ConnectedUnit.UnitType) {
				u.ConnectedUnit.Ghost.NukeDot = nil
			}
			u.ConnectedUnit.OrderState = 0
		}
		f.killUnitWith(u, noKiller(ReasonCancelled))
	}
	return true
}

func (f *StateFunctions) revertEggMorph(u *Unit, morphTo *UnitTypeT) {
	f.morphUnit(u, morphTo)
	if !u.BuildQueue.Empty() {
		f.buildQueueClearSingle(u)
	}
	u.RemainingBuildTime = 0
	if u.PreviousUnitType == nil {
		throwError("cancel_building_unit: previous_unit_type is null")
	}
	f.replaceSpriteImages(u.Sprite, u.PreviousUnitType.Flingy.Get().Sprite.Get().Image.Get(), DirectionT{})
	u.OrderSignal &^= 4
	f.spriteRunAnim(u.Sprite, IscriptAnimsSpecialState2)
	f.setUnitOrder(u, f.getOrderType(OrdersZergBirth))
}

func (f *StateFunctions) cancelBuildQueueIndex(u *Unit, index uint) {
	if index >= uint(u.BuildQueue.Size()) {
		return
	}
	unitType := u.BuildQueue.At(int(index))
	u.BuildQueue.Erase(int(index))
	f.onBuildQueue(u, BuildQueueErase, int(index))
	if index != 0 || u.CurrentBuildUnit == nil {
		f.onQueuedProductionCancelled(u, unitType)
		if !f.utBuilding(unitType) {
			f.refundUnitCosts(u.Owner, unitType)
		}
	} else {
		f.cancelBuildingUnitOf(u, u.CurrentBuildUnit)
	}
}

func (f *StateFunctions) cancelBuildQueue(u *Unit) {
	for !u.BuildQueue.Empty() {
		f.cancelBuildQueueIndex(u, uint(u.BuildQueue.Size()-1))
	}
}

func (f *StateFunctions) placeBuilding(u *Unit, orderType *OrderTypeT, unitType *UnitTypeT, position Xy) {
	if orderType.ID == OrdersBuildingLand {
		f.setUnitOrderPos(u, orderType, position)
		return
	}
	if f.unitIs(u.UnitType, UnitTypesZergDrone) {
		if u.OrderType.ID == OrdersDroneBuild {
			f.onActionRejected(u, RejectBusy)
			return
		}
		if u.OrderType.ID == OrdersDroneLand && (f.uOrderNotInterruptible(u) || f.uIscriptNobrk(u)) {
			f.onActionRejected(u, RejectBusy)
			return
		}
	}
	if f.unitIs(u.UnitType, UnitTypesZergNydusCanal) && orderType.ID == OrdersBuildNydusExit {
		f.setUnitOrderPos(u, orderType, position)
		return
	}
	if orderType.ID == OrdersCTFCOP2 {
		f.setUnitOrderPos(u, orderType, position)
		return
	}
	if !f.hasAvailableSupplyFor(u.Owner, unitType, true) {
		f.onActionRejected(u, RejectSupply)
		return
	}
	if !f.hasAvailableResourcesFor(u.Owner, unitType, true) {
		f.onActionRejectedForCost(u, u.Owner, unitType.MineralCost)
		return
	}
	f.setUnitOrderPos(u, orderType, position)
	f.cancelBuildQueue(u)
	if f.utBuilding(u.UnitType) && f.utBuilding(unitType) {
		u.Building.AddonBuildType = unitType
		f.onAddonBuildType(u, true)
	} else {
		f.buildQueuePush(u, unitType)
	}
}

func (f *StateFunctions) playerHasSupplyAndResourcesFor(owner int32, unitType *UnitTypeT, showError bool) bool {
	if !f.hasAvailableSupplyFor(owner, unitType, showError) {
		return false
	}
	if !f.hasAvailableResourcesFor(owner, unitType, showError) {
		return false
	}
	return true
}

func (f *StateFunctions) getBuildingAtCenterPosition(pos Xy, unitType *UnitTypeT) *Unit {
	area := Rect{From: pos, To: XyAdd(pos, Xy{X: f.gameSt.MaxUnitWidth, Y: f.gameSt.MaxUnitHeight})}
	for n := range f.findUnitsNoexpand(area) {
		if n.UnitType == unitType && XyDivScalar(n.Sprite.Position, 32) == XyDivScalar(pos, 32) {
			return n
		}
	}
	return nil
}

func absUnsigned(v int32) uint32 {
	if v < 0 {
		return uint32(-v)
	}
	return uint32(v)
}

func (f *StateFunctions) isInPsionicMatrixRange(rel Xy) bool {
	x := absUnsigned(rel.X)
	y := absUnsigned(rel.Y)
	if x >= 256 {
		return false
	}
	if y >= 160 {
		return false
	}
	if rel.X < 0 {
		x--
	}
	if rel.Y < 0 {
		y--
	}
	return psiFieldMask[y/32][x/32]
}

func (f *StateFunctions) isInPsionicMatrix(owner int32, pos Xy) bool {
	for u := range f.st.PsionicMatrixUnits.All() {
		if u.Owner != owner {
			continue
		}
		if f.isInPsionicMatrixRange(XySub(pos, u.Sprite.Position)) {
			return true
		}
	}
	return false
}

type buildingTileArea struct {
	pos  XyT[uint]
	size XyT[uint]
}

func (a buildingTileArea) all() func(yield func(x, y uint) bool) {
	return func(yield func(x, y uint) bool) {
		for y := a.pos.Y; y != a.pos.Y+a.size.Y; y++ {
			for x := a.pos.X; x != a.pos.X+a.size.X; x++ {
				if !yield(x, y) {
					return
				}
			}
		}
	}
}

func (a buildingTileArea) origin() Xy {
	return Xy{X: int32(a.pos.X * 32), Y: int32(a.pos.Y * 32)}
}

func (f *StateFunctions) tileAt(x, y uint) *Tile {
	return &f.st.Tiles[y*f.gameSt.MapTileWidth+x]
}

func tileBoundsExclude(bb Rect, x, y uint) bool {
	if bb.From.X >= int32(x*32+32) {
		return true
	}
	if bb.To.X <= int32(x*32) {
		return true
	}
	if bb.From.Y >= int32(y*32+32) {
		return true
	}
	if bb.To.Y <= int32(y*32) {
		return true
	}
	return false
}

func (f *StateFunctions) canPlaceBuilding(u *Unit, owner int32, unitType *UnitTypeT, pos Xy, checkUndetectedUnits bool, checkInvisibleTiles bool) bool {
	var area buildingTileArea
	area.pos.X = uint(uint32(pos.X-unitType.PlacementSize.X/2) / 32)
	area.pos.Y = uint(uint32(pos.Y-unitType.PlacementSize.Y/2) / 32)
	area.size.X = uint(uint32(unitType.PlacementSize.X) / 32)
	area.size.Y = uint(uint32(unitType.PlacementSize.Y) / 32)
	if area.pos.X >= f.gameSt.MapTileWidth {
		return false
	}
	if area.pos.Y >= f.gameSt.MapTileHeight {
		return false
	}
	if area.pos.X+area.size.X > f.gameSt.MapTileWidth {
		return false
	}
	if area.pos.Y+area.size.Y > f.gameSt.MapTileHeight {
		return false
	}
	if u != nil && !f.uFlying(u) && !f.uGroundedBuilding(u) && !f.isReachableUnitToPos(u, area.origin()) {
		return false
	}

	visibilityMask := int32(1) << owner

	if f.unitIsRefinery(unitType) {
		return f.canPlaceRefinery(area, visibilityMask)
	}
	if f.utRequiresPsionicMatrix(unitType) {
		if !f.isInPsionicMatrix(owner, pos) {
			return false
		}
	}
	isNydusExit := u != nil && f.unitIsNydus(u.UnitType)
	if f.utRequiresCreep(unitType) || isNydusExit {
		if !f.buildingTilesHaveCreep(area, visibilityMask, checkInvisibleTiles || isNydusExit) {
			return false
		}
	} else if !f.buildingTilesAreBuildable(area, unitType, visibilityMask) {
		return false
	}

	if !f.buildingTilesAreUnoccupied(u, area, visibilityMask, checkInvisibleTiles) {
		return false
	}

	return !f.buildingAreaHasBlockingUnit(u, unitType, area, visibilityMask, checkUndetectedUnits, checkInvisibleTiles)
}

func (f *StateFunctions) canPlaceRefinery(area buildingTileArea, visibilityMask int32) bool {
	gasPlacementSize := f.getUnitType(UnitTypesResourceVespeneGeyser).PlacementSize
	area.size.X = uint(uint32(gasPlacementSize.X) / 32)
	area.size.Y = uint(uint32(gasPlacementSize.Y) / 32)

	anyVisible := false
	for x, y := range area.all() {
		tile := f.tileAt(x, y)
		if int32(tile.Explored)&visibilityMask != 0 {
			return false
		}
		if !anyVisible && int32(tile.Visible)&visibilityMask == 0 {
			anyVisible = true
		}
	}
	if !anyVisible {
		return true
	}

	gasPos := XyAdd(area.origin(), XyDivScalar(gasPlacementSize, 2))
	return f.getBuildingAtCenterPosition(gasPos, f.getUnitType(UnitTypesResourceVespeneGeyser)) != nil
}

func (f *StateFunctions) buildingTilesHaveCreep(area buildingTileArea, visibilityMask int32, checkAllTiles bool) bool {
	for x, y := range area.all() {
		tile := f.tileAt(x, y)
		if int32(tile.Explored)&visibilityMask != 0 {
			return false
		}
		if checkAllTiles || ^int32(tile.Visible)&visibilityMask != 0 {
			if ^int32(tile.Flags)&TileFlagHasCreep != 0 {
				return false
			}
		}
	}
	return true
}

func (f *StateFunctions) buildingTilesAreBuildable(area buildingTileArea, unitType *UnitTypeT, visibilityMask int32) bool {
	checkUnk4 := !f.unitIsNonFlagBeacon(unitType)
	isResourceDepot := f.utResourceDepot(unitType)
	flagsMask := int32(TileFlagPartiallyWalkable | TileFlagUnbuildable)
	if f.unitRace(unitType) != RaceZerg {
		flagsMask |= TileFlagHasCreep
	}

	for x, y := range area.all() {
		tile := f.tileAt(x, y)
		if int32(tile.Explored)&visibilityMask != 0 {
			return false
		}
		if int32(tile.Visible)&visibilityMask != 0 || int32(tile.Flags)&flagsMask == 0 || (checkUnk4 && tile.Flags&TileFlagUnk4 != 0) {
			if isResourceDepot && f.resourceIsNearTile(x, y) {
				return false
			}
		} else {
			return false
		}
	}
	return true
}

func (f *StateFunctions) resourceIsNearTile(x, y uint) bool {
	bb := Rect{From: Xy{X: 32 * (int32(x) - 3), Y: 32 * (int32(y) - 3)}, To: Xy{X: 32 * (int32(x) + 4), Y: 32 * (int32(y) + 4)}}
	for n := range f.findUnitsNoexpand(bb) {
		if f.utResource(n.UnitType) {
			return true
		}
	}
	return false
}

func (f *StateFunctions) buildingTilesAreUnoccupied(u *Unit, area buildingTileArea, visibilityMask int32, checkInvisibleTiles bool) bool {
	for x, y := range area.all() {
		tile := f.tileAt(x, y)
		if checkInvisibleTiles || int32(tile.Visible)&visibilityMask == 0 {
			if tile.Flags&TileFlagOccupied != 0 && !f.unitCanReplaceOccupiedTile(u, x, y) {
				return false
			}
		}
	}
	return true
}

func (f *StateFunctions) unitCanReplaceOccupiedTile(u *Unit, x, y uint) bool {
	if u == nil {
		return false
	}
	if !f.uGroundedBuilding(u) {
		return false
	}
	if f.unitIsNydus(u.UnitType) {
		return false
	}
	if f.usHidden(&u.Thingy) {
		return false
	}
	return !tileBoundsExclude(u.UnitFinderBoundingBox, x, y)
}

func (f *StateFunctions) buildingAreaHasBlockingUnit(u *Unit, unitType *UnitTypeT, area buildingTileArea, visibilityMask int32, checkUndetectedUnits bool, checkInvisibleTiles bool) bool {
	isAddon := f.utAddon(unitType)

	searchBB := f.unitTypeInnerBoundingBoxAt(unitType, XyAdd(area.origin(), XyDivScalar(unitType.PlacementSize, 2)))
	for n := range f.findUnitsNoexpand(searchBB) {
		if n == u {
			continue
		}
		if isAddon && f.uCanMove(n) {
			continue
		}
		if f.uFlying(n) {
			continue
		}
		if f.uGroundedBuilding(n) {
			continue
		}
		if f.unitIs(n.UnitType, UnitTypesSpellDarkSwarm) {
			continue
		}
		if f.unitIs(n.UnitType, UnitTypesSpellDisruptionWeb) {
			continue
		}
		if u != nil && !checkUndetectedUnits && f.unitTargetIsUndetected(u, n) {
			continue
		}

		for x, y := range area.all() {
			tile := f.tileAt(x, y)
			if checkInvisibleTiles || int32(tile.Visible)&visibilityMask == 0 {
				if tileBoundsExclude(n.UnitFinderBoundingBox, x, y) {
					continue
				}
				return true
			}
		}
	}
	return false
}

func (f *StateFunctions) killUnit(u *Unit) {
	wasDying := f.Hooks.OnKillUnit != nil && u.OrderType != nil && u.OrderType.ID == OrdersDie
	f.dropCarriedItems(u)
	for !u.OrderQueue.Empty() {
		f.removeQueuedOrder(u, u.OrderQueue.Front())
	}
	f.setUnitOrderPos(u, f.getOrderType(OrdersDie), u.OrderTarget.Pos)
	if f.Hooks.OnKillUnit != nil && !wasDying {
		f.Hooks.OnKillUnit(u, f.killContext)
	}
}

func (f *StateFunctions) unitCanEnterNydus(u *Unit, target *Unit) bool {
	if !f.unitIsNydus(target.UnitType) {
		return false
	}
	if !f.uCompleted(target) {
		return false
	}
	if target.Building.Nydus.Exit == nil || !f.uCompleted(target.Building.Nydus.Exit) {
		return false
	}
	if u.Owner != target.Owner {
		return false
	}
	if f.uFlying(u) {
		return false
	}
	if f.unitRace(u.UnitType) != RaceZerg {
		return false
	}
	return true
}

func (f *StateFunctions) unitCanUseShieldBattery(u *Unit, target *Unit) bool {
	if !f.unitIs(target.UnitType, UnitTypesProtossShieldBattery) {
		return false
	}
	if !f.uCompleted(target) {
		return false
	}
	if !u.UnitType.HasShield {
		return false
	}
	if !f.uCompleted(u) {
		return false
	}
	if !f.uCanMove(u) {
		return false
	}
	if f.uGroundedBuilding(u) {
		return false
	}
	if u.Owner != target.Owner {
		return false
	}
	if f.unitRace(u.UnitType) != RaceProtoss {
		return false
	}
	if u.ShieldPoints.GreaterEqual(Fp8Integer(u.UnitType.ShieldPoints)) {
		return false
	}
	if target.Energy == Fp8Zero() {
		return false
	}
	if f.unitIsDisabled(target) {
		return false
	}
	if f.unitIs(u.UnitType, UnitTypesProtossInterceptor) || f.unitIs(u.UnitType, UnitTypesProtossScarab) {
		return false
	}
	return true
}

func (f *StateFunctions) unitIsRescuable(u *Unit) bool {
	if f.st.Players[u.Owner].Controller == PlayerControllerRescuePassive {
		return true
	}
	if f.st.Players[u.Owner].Controller == PlayerControllerUnusedRescueActive {
		return true
	}
	return false
}

func (f *StateFunctions) tryFollowUnit(u *Unit, target *Unit) bool {
	if u.MoveTargetTimer != 0 {
		return true
	}
	if !f.unitTargetIsVisible(u, target) {
		if !f.unitTargetInRange(u, target, 32*14) {
			if f.unitIsAtMoveTarget(&u.Flingy) && !f.uImmovable(u) {
				f.orderDone(u, OrderTarget{})
				return false
			}
			return true
		}
	}
	f.moveToTarget(u, target)
	return true
}

func (f *StateFunctions) followAheadOfUnit(u *Unit, target *Unit) {
	if u.MoveTargetTimer != 0 {
		return
	}
	canSeeTarget := f.unitCanSeeTarget(u, target)
	targetIsMoving := f.uMovementFlag(&target.Flingy, 2)
	if canSeeTarget && targetIsMoving {
		moveTarget := target.Sprite.Position
		haltDistance := f.unitHaltDistance(&u.Flingy)
		pos := XyAdd(target.Sprite.Position, f.toXy(f.directionXY(target.NextVelocityDirection, haltDistance.MulInt(3))))
		if f.isPosInMapBounds(pos) {
			moveTarget = pos
		}
		f.setUnitMoveTarget(u, moveTarget)
	} else {
		f.tryFollowUnit(u, target)
	}
}

func (f *StateFunctions) autoAttackTarget(u *Unit, target *Unit) {
	if f.unitIs(target.UnitType, UnitTypesProtossInterceptor) {
		if target.Fighter.Parent != nil {
			target = target.Fighter.Parent
		}
	}
	f.setUnitOrderUnit(u, u.UnitType.AttackUnit.Get(), target)
	f.uSetStatusFlag(u, UnitStatusFlagReadyToAttack)
	if u.Subunit != nil {
		if f.utTurret(u.Subunit.UnitType) {
			f.setUnitOrderUnit(u.Subunit, u.Subunit.UnitType.AttackUnit.Get(), target)
		}
		f.uSetStatusFlag(u.Subunit, UnitStatusFlagReadyToAttack)
	}
}

func (f *StateFunctions) interruptAttackMoveToAttack(u *Unit, target *Unit, stopBeforeAttacking bool) {
	if stopBeforeAttacking {
		f.stopUnit(u)
	}
	f.queueOrderFront(u, u.OrderType, OrderTarget{Position: u.OrderTarget.Pos, Unit: u.OrderTarget.Unit})
	f.queueOrderFront(u, u.UnitType.AttackUnit.Get(), OrderTarget{Position: target.Sprite.Position, Unit: target})
	f.uUnsetStatusFlag(u, UnitStatusFlagOrderNotInterruptible)
	f.orderDone(u, OrderTarget{})
	f.uSetStatusFlag(u, UnitStatusFlagReadyToAttack)
	if u.Subunit != nil {
		f.uSetStatusFlag(u.Subunit, UnitStatusFlagReadyToAttack)
	}
}

func (f *StateFunctions) attackMoveMovement(u *Unit, stopBeforeAttacking bool) bool {
	if u.OrderState == 0 {
		f.setUnitMoveTarget(u, u.OrderTarget.Pos)
		f.setNextTargetWaypoint(&u.Flingy, u.OrderTarget.Pos)
		u.OrderState = 1
	}
	target := u.AutoTargetUnit
	if target != nil && f.unitTargetIsEnemy(u, target) && !f.usHidden(&target.Thingy) && !f.uInvincible(target) && !f.unitTargetIsUndetected(u, target) {
		u.AutoTargetUnit = nil
		f.interruptAttackMoveToAttack(u, target, stopBeforeAttacking)
		return false
	}
	return true
}

func (f *StateFunctions) attackMoveAcquireTarget(u *Unit, stopBeforeAttacking bool) {
	if f.unitTargetAcquisitionRange(u) == 0 {
		return
	}
	target := f.findAcquireTarget(u)
	if target != nil {
		f.interruptAttackMoveToAttack(u, target, stopBeforeAttacking)
	}
}

func (f *StateFunctions) attackUnitReacquireTarget(u *Unit) {
	if !f.uReadyToAttack(u) {
		return
	}
	target := u.OrderTarget.Unit
	var targetPriority int32 = math.MaxInt32
	if target != nil {
		if f.unitCanAttackTarget(u, target) {
			targetPriority = f.unitTargetAttackPriority(u, target)
			if targetPriority == 0 {
				if f.unitTargetInWeaponMovementRange(u, target) {
					return
				}
				targetPriority = 1
			}
		} else {
			target = nil
		}
	}
	autoTarget := u.AutoTargetUnit
	if autoTarget != nil {
		if f.unitCanAttackTarget(u, autoTarget) && f.unitTargetIsEnemy(u, autoTarget) {
			autoTargetPriority := f.unitTargetAttackPriority(u, autoTarget)
			if autoTargetPriority == 0 {
				if !f.unitTargetInWeaponMovementRange(u, autoTarget) {
					autoTargetPriority = 1
				}
			}
			if autoTargetPriority < targetPriority {
				targetPriority = autoTargetPriority
				target = autoTarget
			}
		} else {
			u.AutoTargetUnit = nil
		}
	}
	if target == nil || targetPriority != 0 {
		newTarget := f.findAcquireTarget(u)
		if newTarget != nil && f.unitTargetAttackPriority(u, newTarget) < targetPriority {
			target = newTarget
		}
	}
	if target == nil || f.unitCanAttackTarget(u, target) {
		u.OrderTarget.Unit = target
		if u.Subunit != nil {
			if u.Subunit.OrderType == u.Subunit.UnitType.AttackUnit.Get() || u.Subunit.OrderType.ID == OrdersHoldPosition {
				u.Subunit.OrderTarget.Unit = target
			}
		}
	}
}

func (f *StateFunctions) attackUnitMoveInRange(u *Unit) bool {
	target := u.OrderTarget.Unit
	if !f.unitCanAttackTarget(u, target) {
		if target == nil || f.uReadyToAttack(u) {
			f.stopUnit(u)
			f.orderDone(u, OrderTarget{})
		} else {
			f.queueOrderFront(u, f.getOrderType(OrdersMove), OrderTarget{Position: target.Sprite.Position})
			f.orderDone(u, OrderTarget{})
		}
		return false
	}
	u.OrderTarget.Pos = target.Sprite.Position
	f.setNextTargetWaypoint(&u.Flingy, target.Sprite.Position)
	if u.OrderState == 0 {
		if target != nil && !f.unitTargetInWeaponMovementRange(u, target) {
			f.moveToTarget(u, target)
		}
		u.OrderState = 1
		return true
	}
	if f.unitTargetOutOfMaxRange(u, target) {
		if f.attackTargetIsUnreachable(u) {
			f.stopUnit(u)
			f.setNextTargetWaypoint(&u.Flingy, u.MoveTarget.Pos)
			f.orderDone(u, OrderTarget{})
			return false
		}
		if f.unitIs(u.UnitType, UnitTypesZergScourge) || f.groundWeaponHasNoFlingy(u.UnitType) {
			f.followAheadOfUnit(u, target)
		} else {
			f.tryFollowUnit(u, target)
		}
		return false
	}
	if f.unitIs(u.UnitType, UnitTypesZergScourge) {
		f.followAheadOfUnit(u, target)
	} else {
		f.stopUnit(u)
	}
	return true
}

func (f *StateFunctions) attackTargetIsUnreachable(u *Unit) bool {
	if f.unitIsAtMoveTarget(&u.Flingy) && f.uImmovable(u) {
		return true
	}
	return f.uReadyToAttack(u) && !f.uStatusFlag(u, UnitStatusFlag8) && !f.uFlying(u) && f.uFlying(u.OrderTarget.Unit)
}

func (f *StateFunctions) groundWeaponHasNoFlingy(unitType *UnitTypeT) bool {
	groundWeapon := unitType.GroundWeapon.Get()
	return groundWeapon != nil && groundWeapon.Flingy.Get().ID == 0
}

func (f *StateFunctions) unitCanFireWeapon(u *Unit, weapon *WeaponTypeT) bool {
	if f.uMovementFlag(&u.Flingy, 8) {
		return false
	}
	target := u.OrderTarget.Unit
	var distance int32
	if target != nil {
		distance = f.unitsDistance(f.unitMainUnit(u), target)
	} else {
		distance = f.unitDistanceTo(f.unitMainUnit(u), u.OrderTarget.Pos)
	}
	if weapon.MinRange != 0 && distance < weapon.MinRange {
		return false
	}
	if distance > f.weaponMaxRange(u, weapon) {
		return false
	}
	if u.PathingFlags&2 != 0 {
		return false
	}
	if u.Position != u.NextTargetWaypoint {
		if f.unitIs(u.UnitType, UnitTypesZergLurker) {
			u.Heading = f.xyDirection(XySub(u.NextTargetWaypoint, u.Sprite.Position))
		} else if Fp8Extend(f.xyDirection(XySub(u.NextTargetWaypoint, u.Sprite.Position)).Sub(u.Heading)).Abs().Greater(weapon.AttackAngle) {
			u.OrderProcessTimer = 0
			return false
		}
	}
	return true
}

func (f *StateFunctions) attackUnitFireWeapon(u *Unit) {
	f.st.PrevBulletSourceUnit = nil
	if f.uCannotAttack(u) {
		return
	}
	target := u.OrderTarget.Unit
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
		return
	}
	if cooldown != 0 {
		if u.OrderProcessTimer > cooldown-1 {
			u.OrderProcessTimer = cooldown - 1
		}
		return
	}
	if !f.unitCanFireWeapon(u, weapon) {
		return
	}
	f.uSetMovementFlag(&u.Flingy, 8)
	cooldown = f.getModifiedWeaponCooldown(u, weapon) + (f.lcgRand(12) & 3) - 1
	u.GroundWeaponCooldown = cooldown
	u.AirWeaponCooldown = cooldown
	f.spriteRunAnim(u.Sprite, anim)
}

func (f *StateFunctions) clearNonEnemyOrderTarget(n *Unit, u *Unit) {
	n = f.unitAttackingUnit(n)
	if n.OrderTarget.Unit == u && n.OrderType.TargetsEnemies && !f.unitTargetIsEnemy(n, u) {
		n.OrderTarget.Unit = nil
	}
}

func (f *StateFunctions) setUnitOwner(u *Unit, owner int32, incrementScore bool) {
	if u.Owner == owner {
		return
	}
	from := u.Owner
	isMorphing := f.unitIsMorphingBuilding(u)
	if isMorphing {
		f.uSetStatusFlag(u, UnitStatusFlagCompleted)
	}
	f.incrementUnitCounts(u, -1)
	if f.uCompleted(u) {
		f.addCompletedUnit(u, -1, false)
	}
	f.st.PlayerUnits[u.Owner].Remove(u)
	u.Owner = owner
	f.st.PlayerUnits[owner].PushFront(u)
	f.incrementUnitCounts(u, 1)
	if f.uCompleted(u) {
		f.addCompletedUnit(u, 1, incrementScore)
	}
	if isMorphing {
		f.uUnsetStatusFlag(u, UnitStatusFlagCompleted)
	}

	if u.OrderTarget.Unit != nil && u.OrderType.TargetsEnemies && !f.unitTargetIsEnemy(u, u.OrderTarget.Unit) {
		u.OrderTarget.Unit = nil
	}
	for n := range f.st.VisibleUnits.All() {
		f.clearNonEnemyOrderTarget(n, u)
	}
	for n := range f.st.HiddenUnits.All() {
		f.clearNonEnemyOrderTarget(n, u)
	}
	if f.unitRace(u.UnitType) == RaceProtoss && f.uGroundedBuilding(u) {
		f.st.UpdatePsionicMatrix = true
	}
	turret := f.unitTurret(u)
	if turret != nil {
		f.setUnitOwner(u.Subunit, owner, incrementScore)
	}
	f.updateUnitSpeedUpgrades(u)
	if f.utWorker(u.UnitType) && u.Worker.GatherTarget != nil && f.unitIs(u.Worker.GatherTarget.UnitType, UnitTypesPowerupFlag) {
		if owner >= 8 || owner == u.Worker.GatherTarget.Owner {
			f.dropCarriedItems(u)
		}
	}
	f.onOwnerChanged(u, from)
}

func (f *StateFunctions) setSpriteOwner(u *Unit, owner int32) {
	if u.Sprite.Owner != owner {
		u.Sprite.Owner = owner
	}
	if u.Subunit != nil && u.Subunit.Sprite.Owner != owner {
		u.Subunit.Sprite.Owner = owner
	}
}

func (f *StateFunctions) leaveGatherQueue(u *Unit) {
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

func scaledHp(prevHp Fp8, prevMaxHp int32, newMaxHp int32) Fp8 {
	hp := int32(1)
	if prevMaxHp != 0 {
		hp = prevHp.IntegerPart() * newMaxHp / prevMaxHp
	}
	if hp == 0 {
		return Fp8FromRaw(1)
	}
	return Fp8Integer(hp)
}

func (f *StateFunctions) reinitializeUnitType(u *Unit, unitType *UnitTypeT) {
	f.leaveGatherQueue(u)
	position := u.Sprite.Position
	prevUnitType := u.UnitType
	prevHp := u.Hp
	prevOrderType := u.OrderType
	prevOrderTarget := u.OrderTarget
	prevOrderState := u.OrderState
	prevSpriteOwner := u.Sprite.Owner
	var prevResources int32
	if f.utResource(u.UnitType) {
		prevResources = u.Building.Resource.ResourceCount
	}
	var prevLarvaSpawnSideValues [4]int32
	wasHatchery := f.unitIsHatchery(u.UnitType)
	if wasHatchery {
		prevLarvaSpawnSideValues = u.Building.Hatchery.LarvaSpawnSideValues
	}
	u.PreviousUnitType = prevUnitType
	u.PreviousHp = max(prevHp.IntegerPart(), 1)
	u.OrderType = f.getOrderType(OrdersDie)
	u.OrderState = 1
	f.destroySprite(u.Sprite)
	u.Sprite = nil
	f.freeUnitPath(u)
	if !f.initializeUnitType(u, unitType, position, u.Owner) {
		throwError("reinitialize_unit_type: initialize_unit_type failed")
	}
	u.Hp = scaledHp(prevHp, prevUnitType.Hitpoints.IntegerPart(), u.UnitType.Hitpoints.IntegerPart())
	u.AirStrength = f.getUnitStrength(u, false)
	u.GroundStrength = f.getUnitStrength(u, true)
	u.Sprite.Owner = prevSpriteOwner
	f.uUnsetStatusFlag(u, UnitStatusFlagDisabled)
	f.uUnsetStatusFlag(u, UnitStatusFlagIscriptNobrk)
	f.uSetStatusFlagValue(u, UnitStatusFlagOrderNotInterruptible, !prevOrderType.CanBeInterrupted)
	u.OrderType = prevOrderType
	u.OrderTarget = prevOrderTarget
	u.OrderState = prevOrderState
	if f.uGroundedBuilding(u) && !u.BuildQueue.Empty() {
		f.buildQueueClearSingle(u)
	}
	if f.utResource(u.UnitType) {
		f.setUnitResources(u, prevResources)
	}
	if wasHatchery && f.unitIsHatchery(u.UnitType) {
		u.Building.Hatchery.LarvaSpawnSideValues = prevLarvaSpawnSideValues
	}
}

func (f *StateFunctions) morphUnit(u *Unit, unitType *UnitTypeT) {
	from := u.UnitType
	visibility := u.Sprite.VisibilityFlags
	if !f.usHidden(&u.Thingy) {
		f.unitFinderRemove(u)
		if f.uGroundedBuilding(u) {
			f.setUnitTilesUnoccupied(u.UnitType, u.Sprite.Position)
		}
		if f.uFlying(u) {
			f.decrementRepulseField(u)
		}
	}
	requiresDetector := f.uRequiresDetector(u)
	cloaked := f.uCloaked(u)
	var data1, data2 int32
	if requiresDetector || cloaked {
		data1 = u.Sprite.MainImage.ModifierData1
		data2 = u.Sprite.MainImage.ModifierData2
	}
	f.incrementUnitCounts(u, -1)
	if f.uCompleted(u) {
		f.addCompletedUnit(u, -1, true)
	}
	f.reinitializeUnitType(u, unitType)
	f.incrementUnitCounts(u, 1)
	if f.uCompleted(u) {
		f.addCompletedUnit(u, 1, true)
	}
	if requiresDetector || cloaked {
		f.setSpriteCloakModifier(u.Sprite, requiresDetector, cloaked, f.uBurrowed(u), data1, data2)
	}
	f.setSpriteVisibility(u.Sprite, visibility)
	if f.utBuilding(u.UnitType) && (requiresDetector || cloaked) {
		f.setSecondaryOrder(u, f.getOrderType(OrdersNothing))
		f.uUnsetStatusFlag(u, UnitStatusFlagRequiresDetector)
		f.uUnsetStatusFlag(u, UnitStatusFlagCloaked)
	}
	newPos := f.restrictMoveTargetToValidBounds(u.UnitType, u.Sprite.Position)
	if newPos != u.Sprite.Position {
		f.moveUnit(u, newPos)
	}
	if !f.usHidden(&u.Thingy) {
		f.unitFinderInsert(u)
		if f.uGroundedBuilding(u) {
			f.setUnitTilesOccupied(u.UnitType, u.Sprite.Position)
		}
		f.checkUnitCollision(u)
		if f.uFlying(u) {
			f.incrementRepulseField(u)
		}
	}
	if u.Subunit != nil {
		f.morphTurret(u, unitType)
	}
	f.applyUnitEffects(u)
	f.onUnitMorphed(u, from)
}

func (f *StateFunctions) morphTurret(u *Unit, unitType *UnitTypeT) {
	if f.utTurret(unitType) {
		throwError("unit type %d has a turret but is also flagged as a turret", int32(unitType.ID))
	}
	if !f.utTurret(unitType.TurretUnitType.Get()) {
		throwError("unit type %d was created as a turret but is not flagged as one", int32(unitType.TurretUnitType.Get().ID))
	}

	requiresDetector := f.uRequiresDetector(u.Subunit)
	cloaked := f.uCloaked(u.Subunit)
	var data1, data2 int32
	if requiresDetector || cloaked {
		data1 = u.Subunit.Sprite.MainImage.ModifierData1
		data2 = u.Subunit.Sprite.MainImage.ModifierData2
	}
	prevIscriptUnit := f.iscriptUnit
	f.iscriptUnit = u.Subunit
	defer func() { f.iscriptUnit = prevIscriptUnit }()
	if unitType.TurretUnitType.Get() == nil {
		throwError("unit_type->turret_unit_type is null")
	}
	f.reinitializeUnitType(u.Subunit, unitType.TurretUnitType.Get())
	u.Subunit.Sprite.Flags |= SpriteFlagTurret
	if requiresDetector || cloaked {
		f.setSpriteCloakModifier(u.Subunit.Sprite, requiresDetector, cloaked, f.uBurrowed(u), data1, data2)
	}
	f.setImageOffset(u.Subunit.Sprite.MainImage, f.getImageLoOffset(u.Sprite.MainImage, 2, 0, false))
	f.setSpriteVisibility(u.Subunit.Sprite, u.Sprite.VisibilityFlags)
}

func (f *StateFunctions) buildRefinery(u *Unit, unitType *UnitTypeT) *Unit {
	target := u.OrderTarget.Unit
	if target == nil || !f.unitIs(target.UnitType, UnitTypesResourceVespeneGeyser) {
		target = f.getBuildingAtCenterPosition(u.OrderTarget.Pos, f.getUnitType(UnitTypesResourceVespeneGeyser))
		if target == nil {
			return nil
		}
	}
	f.addCompletedUnit(target, -1, true)
	f.uUnsetStatusFlag(target, UnitStatusFlagCompleted)
	f.setUnitOwnerWith(target, u.Owner, false, OwnerRefinery)
	f.setSpriteOwner(target, u.Owner)
	f.morphUnit(target, unitType)
	f.createImage(f.getImageType(ImageTypesIMAGEIDVespeneGeyser2), target.Sprite, Xy{}, imageOrderBelow, nil)
	target.Hp = unitType.Hitpoints.DivInt(10)
	return target
}

func (f *StateFunctions) unitCanGatherGas(u *Unit, target *Unit) bool {
	if !f.utWorker(u.UnitType) {
		return false
	}
	if !f.unitIsRefinery(target.UnitType) {
		return false
	}
	if !f.uCompleted(target) {
		return false
	}
	if target.Owner != u.Owner {
		return false
	}
	return true
}

func (f *StateFunctions) findGasExitPosition(u *Unit, gas *Unit, pos *Xy, heading *DirectionT) bool {
	resourceDepot := f.findNearestActiveResourceDepot(u)
	if resourceDepot == nil || gas == nil {
		found, placement := f.findUnitPlacement(u, u.Sprite.Position, true)
		*pos = placement
		*heading = u.Heading
		return found
	}
	*heading = f.xyDirection(XySub(resourceDepot.Sprite.Position, gas.Sprite.Position))
	offset := f.toXy(f.directionXYInt(heading.Sub(DirectionTFromRaw(64)), 32))
	fromPos := XyAdd(gas.Sprite.Position, offset)
	toPos := XyAdd(resourceDepot.Sprite.Position, offset)

	bb := f.unitSpriteInnerBoundingBox(gas)
	bb.From = XySub(bb.From, XyAdd(u.UnitType.Dimensions.To, Xy{X: 1, Y: 1}))
	bb.To = XyAdd(bb.To, XyAdd(u.UnitType.Dimensions.From, Xy{X: 1, Y: 1}))
	f.getUniqueSidedPositionsWithinBounds(&fromPos, &toPos, bb)
	found, placement := f.findUnitPlacementInBounds(u, toPos, bb, true)
	if !found {
		found, placement = f.findUnitPlacement(u, toPos, true)
	}
	*pos = placement
	return found
}

func (f *StateFunctions) isAtNextTargetWaypointOrWithinAttackAngle(u *Unit) bool {
	if u.Position == u.NextTargetWaypoint {
		return true
	}
	weapon := f.unitGroundWeapon(u)
	if weapon != nil {
		if f.headingError(u.Heading, u).LessEqual(weapon.AttackAngle) {
			return true
		}
	}
	return false
}

func (f *StateFunctions) unitLoadTarget(u *Unit, target *Unit) {
	index := ^uint(0)
	for i := uint(0); i != u.UnitType.SpaceProvided; i++ {
		n := f.getUnitByID(u.LoadedUnits[i])
		if n == nil {
			index = i
			break
		}
	}
	f.playSoundUnit(40+int32(f.unitRace(u.UnitType)), u, false)
	u.LoadedUnits[index] = f.getUnitID(target)
	target.ConnectedUnit = u
	f.uSetStatusFlag(target, UnitStatusFlagLoaded)
	f.hideUnit(target, true)
	f.spriteRunAnim(target.Sprite, IscriptAnimsWalkingToIdle)
	if f.uGroundedBuilding(u) {
		f.uUnsetStatusFlag(target, UnitStatusFlagCanMove)
		f.uSetStatusFlag(target, UnitStatusFlagInBunker)
		f.resetMovementState(target)
		turret := f.unitTurret(target)
		if turret != nil {
			f.moveUnit(turret, u.Sprite.Position)
			f.uUnsetStatusFlag(turret, UnitStatusFlagCanMove)
			f.uSetStatusFlag(turret, UnitStatusFlagInBunker)
			f.resetMovementState(turret)
		} else {
			f.moveUnit(target, u.Sprite.Position)
		}
	}
}

func (f *StateFunctions) unitUnloadImpl(u *Unit, containerDestroyed bool) {
	if !f.uLoaded(u) {
		return
	}
	container := u.ConnectedUnit
	if container != nil {
		uid := f.getUnitID(u)
		index := ^uint(0)
		for i := uint(0); i != container.UnitType.SpaceProvided; i++ {
			if container.LoadedUnits[i] == uid {
				index = i
				break
			}
		}
		container.LoadedUnits[index] = UnitID{}
		u.ConnectedUnit = nil
		f.uUnsetStatusFlag(u, UnitStatusFlagLoaded)

		if containerDestroyed {
			f.moveUnit(u, container.Sprite.Position)
		} else {
			f.showUnit(u)
			f.playSoundUnit(43+int32(f.unitRace(container.UnitType)), container, false)
		}
	}
	f.uUnsetStatusFlag(u, UnitStatusFlagInBunker)
	f.uSetStatusFlagValue(u, UnitStatusFlagCanMove, f.utCanMove(u.UnitType))
	f.resetMovementState(u)
	turret := f.unitTurret(u)
	if turret != nil {
		f.uUnsetStatusFlag(turret, UnitStatusFlagInBunker)
		f.uSetStatusFlagValue(turret, UnitStatusFlagCanMove, f.utCanMove(turret.UnitType))
		f.resetMovementState(turret)
	}
}

func (f *StateFunctions) unitUnload(u *Unit) bool {
	container := u.ConnectedUnit
	if container == nil || f.unitIsDisabled(container) {
		return false
	}
	if !f.uGroundedBuilding(container) && container.MainOrderTimer != 0 {
		return false
	}
	pos := container.Sprite.Position
	f.moveUnit(u, pos)
	r, pos := f.findUnitPlacement(u, pos, false)
	if !r {
		return false
	}
	container.MainOrderTimer = 15
	f.moveUnit(u, pos)
	if u.Subunit != nil {
		f.moveUnit(u.Subunit, pos)
	}
	f.iscriptRunToIdle(u)
	f.unitUnloadImpl(u, false)
	f.setUnitOrder(u, u.UnitType.ReturnToIdle.Get())
	if !f.uGroundedBuilding(container) {
		f.setUnloadedUnitCooldowns(u)
	}
	return true
}

func (f *StateFunctions) setUnloadedUnitCooldowns(u *Unit) {
	if f.unitIs(u.UnitType, UnitTypesProtossReaver) {
		u.MainOrderTimer = 30
		return
	}
	groundWeapon := f.unitGroundWeapon(u)
	airWeapon := f.unitAirWeapon(u)
	if groundWeapon != nil {
		u.GroundWeaponCooldown = f.getModifiedWeaponCooldown(u, groundWeapon)
	}
	if airWeapon != nil {
		u.AirWeaponCooldown = f.getModifiedWeaponCooldown(u, airWeapon)
	}
	u.SpellCooldown = 30
}

func (f *StateFunctions) triggerGiveUnitTo(u *Unit, newOwner int32) bool {
	if f.uHallucination(u) {
		return false
	}
	if f.unitIsMineralField(u.UnitType) {
		return false
	}
	if f.unitIs(u.UnitType, UnitTypesResourceVespeneGeyser) {
		return false
	}
	if f.unitIsFighter(u.UnitType) {
		return false
	}
	if f.utAddon(u.UnitType) {
		return false
	}
	if u.Owner == newOwner {
		return false
	}
	if f.uCompleted(u) && f.unitIsRefinery(u.UnitType) {
		f.giveGasHarvesterTo(u, newOwner)
	}
	f.giveUnitTo(u, newOwner)
	if f.unitProvidesSpace(u) {
		for n := range f.loadedUnits(u) {
			f.giveUnitTo(n, newOwner)
		}
	}
	if f.unitIsCarrier(u.UnitType) {
		f.giveFightersTo(&u.Carrier, newOwner)
	} else if f.unitIsReaver(u.UnitType) {
		f.giveFightersTo(u.Reaver(), newOwner)
	}
	if f.uGroundedBuilding(u) {
		if u.Building.Addon != nil {
			f.giveUnitTo(u.Building.Addon, newOwner)
		}
		if f.unitIsBuildingAddon(u) {
			f.giveUnitTo(u.CurrentBuildUnit, newOwner)
		}
		if f.unitIsNydus(u.UnitType) && u.Building.Nydus.Exit != nil {
			f.giveUnitTo(u.Building.Nydus.Exit, newOwner)
		}
	}
	return true
}

func (f *StateFunctions) giveGasHarvesterTo(refinery *Unit, newOwner int32) {
	for n := range f.st.PlayerUnits[refinery.Owner].All() {
		if !f.utWorker(n.UnitType) {
			continue
		}
		if !f.usHidden(&n.Thingy) {
			continue
		}
		if n.OrderType.ID != OrdersHarvestGas {
			continue
		}
		if n.OrderTarget.Unit != refinery {
			continue
		}
		f.giveUnitTo(n, newOwner)
		break
	}
}

func (f *StateFunctions) giveFightersTo(fighters *UnitCarrier, newOwner int32) {
	for n := range fighters.InsideUnits.All() {
		f.giveUnitTo(n, newOwner)
	}
	for n := range fighters.OutsideUnits.All() {
		f.giveUnitTo(n, newOwner)
	}
}

type unitTypeGrants struct {
	upgrades []UpgradeTypes
	techs    []TechTypes
}

var giveUnitToGrants = map[UnitTypes]unitTypeGrants{
	UnitTypesTerranMarine:             {upgrades: []UpgradeTypes{UpgradeTypesU238Shells}, techs: []TechTypes{TechTypesStimPacks}},
	UnitTypesTerranGhost:              {upgrades: []UpgradeTypes{UpgradeTypesOcularImplants, UpgradeTypesMoebiusReactor}, techs: []TechTypes{TechTypesLockdown, TechTypesPersonnelCloaking}},
	UnitTypesTerranVulture:            {upgrades: []UpgradeTypes{UpgradeTypesIonThrusters}, techs: []TechTypes{TechTypesSpiderMines}},
	UnitTypesTerranGoliath:            {upgrades: []UpgradeTypes{UpgradeTypesCharonBoosters}},
	UnitTypesTerranSiegeTankTankMode:  {techs: []TechTypes{TechTypesTankSiegeMode}},
	UnitTypesTerranWraith:             {upgrades: []UpgradeTypes{UpgradeTypesApolloReactor}, techs: []TechTypes{TechTypesCloakingField}},
	UnitTypesTerranScienceVessel:      {upgrades: []UpgradeTypes{UpgradeTypesTitanReactor}, techs: []TechTypes{TechTypesDefensiveMatrix, TechTypesIrradiate, TechTypesEMPShockwave}},
	UnitTypesTerranBattlecruiser:      {upgrades: []UpgradeTypes{UpgradeTypesColossusReactor}, techs: []TechTypes{TechTypesYamatoGun}},
	UnitTypesTerranSiegeTankSiegeMode: {techs: []TechTypes{TechTypesTankSiegeMode}},
	UnitTypesTerranFirebat:            {techs: []TechTypes{TechTypesStimPacks}},
	UnitTypesTerranMedic:              {upgrades: []UpgradeTypes{UpgradeTypesCaduceusReactor}, techs: []TechTypes{TechTypesHealing, TechTypesRestoration, TechTypesOpticalFlare}},
	UnitTypesZergZergling:             {upgrades: []UpgradeTypes{UpgradeTypesMetabolicBoost, UpgradeTypesAdrenalGlands}, techs: []TechTypes{TechTypesBurrowing}},
	UnitTypesZergHydralisk:            {upgrades: []UpgradeTypes{UpgradeTypesMuscularAugments, UpgradeTypesGroovedSpines}, techs: []TechTypes{TechTypesBurrowing, TechTypesLurkerAspect}},
	UnitTypesZergUltralisk:            {upgrades: []UpgradeTypes{UpgradeTypesAnabolicSynthesis, UpgradeTypesChitinousPlating}},
	UnitTypesZergDrone:                {techs: []TechTypes{TechTypesBurrowing}},
	UnitTypesZergOverlord:             {upgrades: []UpgradeTypes{UpgradeTypesVentralSacs, UpgradeTypesAntennae, UpgradeTypesPneumatizedCarapace}},
	UnitTypesZergQueen:                {upgrades: []UpgradeTypes{UpgradeTypesGameteMeiosis}, techs: []TechTypes{TechTypesInfestation, TechTypesParasite, TechTypesSpawnBroodlings, TechTypesEnsnare}},
	UnitTypesZergDefiler:              {upgrades: []UpgradeTypes{UpgradeTypesMetasynapticNode}, techs: []TechTypes{TechTypesBurrowing, TechTypesDarkSwarm, TechTypesPlague, TechTypesConsume}},
	UnitTypesZergInfestedTerran:       {techs: []TechTypes{TechTypesBurrowing}},
	UnitTypesProtossCorsair:           {upgrades: []UpgradeTypes{UpgradeTypesArgusJewel}, techs: []TechTypes{TechTypesDisruptionWeb}},
	UnitTypesProtossDarkTemplar:       {techs: []TechTypes{TechTypesDarkArchonMeld}},
	UnitTypesProtossDarkArchon:        {upgrades: []UpgradeTypes{UpgradeTypesArgusTalisman}, techs: []TechTypes{TechTypesMindControl, TechTypesFeedback, TechTypesMaelstrom}},
	UnitTypesProtossZealot:            {upgrades: []UpgradeTypes{UpgradeTypesLegEnhancements}},
	UnitTypesProtossDragoon:           {upgrades: []UpgradeTypes{UpgradeTypesSingularityCharge}},
	UnitTypesProtossHighTemplar:       {upgrades: []UpgradeTypes{UpgradeTypesKhaydarinAmulet}, techs: []TechTypes{TechTypesArchonWarp, TechTypesPsionicStorm, TechTypesHallucination}},
	UnitTypesProtossShuttle:           {upgrades: []UpgradeTypes{UpgradeTypesGraviticDrive}},
	UnitTypesProtossScout:             {upgrades: []UpgradeTypes{UpgradeTypesApialSensors, UpgradeTypesGraviticThrusters}},
	UnitTypesProtossArbiter:           {upgrades: []UpgradeTypes{UpgradeTypesKhaydarinCore}, techs: []TechTypes{TechTypesRecall, TechTypesStasisField}},
	UnitTypesProtossCarrier:           {upgrades: []UpgradeTypes{UpgradeTypesCarrierCapacity}},
	UnitTypesProtossReaver:            {upgrades: []UpgradeTypes{UpgradeTypesScarabDamage, UpgradeTypesReaverCapacity}},
	UnitTypesProtossObserver:          {upgrades: []UpgradeTypes{UpgradeTypesSensorArray, UpgradeTypesGraviticBoosters}},
	UnitTypesZergLurker:               {techs: []TechTypes{TechTypesLurkerAspect}},
}

func (f *StateFunctions) grantUnitTypeUpgradesAndTechs(u *Unit, newOwner int32) (anyUpgradesGranted bool) {
	grants := giveUnitToGrants[u.UnitType.ID]
	for _, id := range grants.upgrades {
		level := f.playerUpgradeLevel(u.Owner, id)
		if f.playerUpgradeLevel(newOwner, id) < level {
			anyUpgradesGranted = true
			f.st.UpgradeLevels[newOwner][id] = level
		}
	}
	for _, id := range grants.techs {
		if !f.playerHasResearched(newOwner, id) {
			f.st.TechResearched[newOwner][id] = true
		}
	}
	return anyUpgradesGranted
}

func (f *StateFunctions) giveUnitTo(u *Unit, newOwner int32) {
	anyUpgradesGranted := f.grantUnitTypeUpgradesAndTechs(u, newOwner)
	f.onUnitDeselect(u)
	prevCancelReason := f.swapCancelReason(CancelOwnerChanged)
	if f.uGroundedBuilding(u) {
		if !u.BuildQueue.Empty() && u.BuildQueue.Front().ID <= UnitTypesSpellDisruptionWeb {
			f.cancelBuildQueue(u)
			f.spriteRunAnim(u.Sprite, IscriptAnimsWorkingToIdle)
		}
		if f.unitIsResearching(u) {
			f.cancelResearch(u, false)
		}
		if f.unitIsUpgrading(u) {
			f.cancelUpgrade(u, false)
		}
	}
	if f.unitIsReaver(u.UnitType) || f.unitIsCarrier(u.UnitType) {
		f.cancelBuildQueue(u)
		u.CurrentBuildUnit = nil
		u.SecondaryOrderState = 0
	}
	f.cancelReason = prevCancelReason
	f.setUnitOwner(u, newOwner, true)
	if newOwner < 8 {
		f.setSpriteOwner(u, newOwner)
	}
	if anyUpgradesGranted {
		f.applyUpgradesToPlayerUnits(newOwner)
	}
	if f.unitIsIdleAfterOwnerChange(u) {
		f.setIdleOrderForController(u, newOwner)
	}
}

func (f *StateFunctions) unitIsIdleAfterOwnerChange(u *Unit) bool {
	idle := f.uCompleted(u)
	if f.uLoaded(u) {
		idle = false
	} else if f.uGroundedBuilding(u) && u.SecondaryOrderType.ID == OrdersBuildAddon && u.CurrentBuildUnit != nil && !f.uCompleted(u.CurrentBuildUnit) {
		idle = false
	} else if f.utPowerup(u.UnitType) {
		idle = false
	} else if f.unitIs(u.UnitType, UnitTypesProtossInterceptor) {
		idle = false
	} else if f.unitIs(u.UnitType, UnitTypesProtossScarab) {
		idle = false
	} else if f.unitIs(u.UnitType, UnitTypesTerranNuclearMissile) {
		idle = false
	}
	return idle
}

func (f *StateFunctions) setIdleOrderForController(u *Unit, owner int32) {
	switch f.st.Players[owner].Controller {
	case PlayerControllerRescuePassive:
		f.setUnitOrder(u, f.getOrderType(OrdersRescuePassive))
	case PlayerControllerNeutral:
		f.setUnitOrder(u, f.getOrderType(OrdersNeutral))
	default:
		f.setUnitOrder(u, u.UnitType.HumanAiIdle.Get())
	}
}

func (f *StateFunctions) makeUnitNeutral(u *Unit) {
	if u.SecondaryOrderType.ID == OrdersBuildAddon && f.uGroundedBuilding(u) && u.CurrentBuildUnit != nil {
		if !f.uCompleted(u.CurrentBuildUnit) {
			prevCancelReason := f.swapCancelReason(CancelOwnerChanged)
			f.cancelBuildingUnit(u.CurrentBuildUnit)
			f.cancelReason = prevCancelReason
		}
	}
	f.giveUnitTo(u, 11)
	if !f.uHallucination(u) {
		f.setSecondaryOrder(u, f.getOrderType(OrdersNothing))
	}
}

func (f *StateFunctions) buildingAbandonAddon(u *Unit) {
	addon := u.Building.Addon
	if addon == nil {
		return
	}
	prevOwnerReason := f.swapOwnerReason(OwnerAddon)
	f.makeUnitNeutral(addon)
	f.ownerReason = prevOwnerReason
	u.Building.Addon = nil
	prevIscriptUnit := f.iscriptUnit
	f.iscriptUnit = addon
	defer func() { f.iscriptUnit = prevIscriptUnit }()
	f.spriteRunAnim(addon.Sprite, IscriptAnimsLiftOff)
}

func (f *StateFunctions) createDustSprite(sprite *Sprite, spriteID SpriteTypes, loIndex uint, offsetFrom uint, offsetTo uint, flipped bool) {
	for i := offsetFrom; i != offsetTo+1; i++ {
		offset := f.getImageLoOffset(sprite.MainImage, loIndex, i, false)
		if offset.X != 127 && offset.Y != 127 {
			f.createDustThingy(sprite, spriteID, offset, flipped)
		}
		spriteID++
	}
}

func (f *StateFunctions) createDustThingy(sprite *Sprite, spriteID SpriteTypes, offset Xy, flipped bool) {
	t := f.createThingy(f.getSpriteType(spriteID), XyAdd(sprite.Position, offset), 0)
	if t == nil {
		return
	}
	t.Sprite.ElevationLevel = sprite.ElevationLevel + 1
	if !f.usHidden(t) {
		f.setSpriteVisibility(t.Sprite, int32(f.tileVisibility(t.Sprite.Position)))
	}
	if flipped {
		for image := range t.Sprite.Images.All() {
			f.setImageFrameIndexOffset(image, image.FrameIndexOffset, true)
		}
	}
}

func (f *StateFunctions) unitCanAttachAddon(u *Unit, addon *Unit) bool {
	if f.unitIs(addon.UnitType, UnitTypesTerranComsatStation) {
		if f.unitIs(u.UnitType, UnitTypesTerranCommandCenter) {
			return true
		}
	} else if f.unitIs(addon.UnitType, UnitTypesTerranNuclearSilo) {
		if f.unitIs(u.UnitType, UnitTypesTerranCommandCenter) {
			return true
		}
	} else if f.unitIs(addon.UnitType, UnitTypesTerranControlTower) {
		if f.unitIs(u.UnitType, UnitTypesTerranStarport) {
			return true
		}
	} else if f.unitIs(addon.UnitType, UnitTypesTerranCovertOps) {
		if f.unitIs(u.UnitType, UnitTypesTerranScienceFacility) {
			return true
		}
	} else if f.unitIs(addon.UnitType, UnitTypesTerranPhysicsLab) {
		if f.unitIs(u.UnitType, UnitTypesTerranScienceFacility) {
			return true
		}
	} else if f.unitIs(addon.UnitType, UnitTypesTerranMachineShop) {
		if f.unitIs(u.UnitType, UnitTypesTerranFactory) {
			return true
		}
	}
	return false
}

func (f *StateFunctions) findConnectingAddon(u *Unit) *Unit {
	halfPlacementSize := XyDivScalar(u.UnitType.PlacementSize, 2)
	var bb Rect
	bb.From = XySub(XySub(u.Sprite.Position, halfPlacementSize), Xy{X: 32, Y: 32})
	bb.To = XyAdd(XyAdd(bb.From, u.UnitType.PlacementSize), Xy{X: 64, Y: 64})
	pos := XySub(u.Sprite.Position, halfPlacementSize)
	return f.findUnitNoexpand(bb, func(n *Unit) bool {
		if !f.utAddon(n.UnitType) {
			return false
		}
		if n.Owner != 11 {
			return false
		}
		if !f.unitCanAttachAddon(u, n) {
			return false
		}
		npos := XySub(XySub(n.Sprite.Position, XyDivScalar(n.UnitType.PlacementSize, 2)), n.UnitType.AddonPosition)
		return XyDivScalar(pos, 32) == XyDivScalar(npos, 32)
	})
}

func (f *StateFunctions) connectAddon(u *Unit, addon *Unit) {
	u.Building.Addon = addon
	f.setUnitOwnerWith(addon, u.Owner, false, OwnerAddon)
	f.setSpriteOwner(addon, u.Owner)
	prevIscriptUnit := f.iscriptUnit
	f.iscriptUnit = addon
	defer func() { f.iscriptUnit = prevIscriptUnit }()
	f.spriteRunAnim(addon.Sprite, IscriptAnimsLanding)
	if f.unitIs(addon.UnitType, UnitTypesTerranNuclearSilo) {
		if addon.Building.Silo.Nuke != nil {
			f.setUnitOwnerWith(addon.Building.Silo.Nuke, u.Owner, false, OwnerAddon)
		}
	}
}

func (f *StateFunctions) findAndConnectAddon(u *Unit) {
	addon := f.findConnectingAddon(u)
	if addon != nil {
		f.connectAddon(u, addon)
	}
}

func (f *StateFunctions) refundCancelledWork(u *Unit, minerals int32, gas int32, unitDestroyed bool) {
	if unitDestroyed {
		f.st.CurrentMinerals[u.Owner] += threeQuarters(minerals)
		f.st.CurrentGas[u.Owner] += threeQuarters(gas)
	} else {
		f.st.CurrentMinerals[u.Owner] += minerals
		f.st.CurrentGas[u.Owner] += gas
		f.spriteRunAnim(u.Sprite, IscriptAnimsWorkingToIdle)
	}
}

func (f *StateFunctions) cancelResearch(u *Unit, unitDestroyed bool) {
	if u.Building.ResearchingType == nil {
		return
	}
	techType := u.Building.ResearchingType
	f.onResearchCancelled(u, Research{ID: uint8(techType.ID)}, unitDestroyed)
	u.Building.ResearchingType = nil
	u.Building.UpgradeResearchTime = 0
	f.st.TechResearching[u.Owner][techType.ID] = false
	f.refundCancelledWork(u, techType.MineralCost, techType.GasCost, unitDestroyed)
}

func (f *StateFunctions) cancelUpgrade(u *Unit, unitDestroyed bool) {
	if u.Building.UpgradingType == nil {
		return
	}
	upgradeType := u.Building.UpgradingType
	f.onResearchCancelled(u, Research{Upgrade: true, ID: uint8(upgradeType.ID)}, unitDestroyed)
	u.Building.UpgradingType = nil
	u.Building.UpgradeResearchTime = 0
	u.Building.UpgradingLevel = 0
	f.st.UpgradeUpgrading[u.Owner][upgradeType.ID] = false
	f.refundCancelledWork(u, f.upgradeMineralCost(u.Owner, upgradeType), f.upgradeGasCost(u.Owner, upgradeType), unitDestroyed)
}

func (f *StateFunctions) applyUpgradesToPlayerUnits(owner int32) {
	for u := range f.st.PlayerUnits[owner].All() {
		f.updateUnitSpeedUpgrades(u)
	}
}

func (f *StateFunctions) setUnitBurrowed(u *Unit) {
	f.resetMovementState(u)
	f.uSetStatusFlag(u, UnitStatusFlagBurrowed)
	f.uSetStatusFlag(u, UnitStatusFlagPassivelyCloaked)
	u.Sprite.Flags |= SpriteFlagBurrowed
	u.Sprite.ElevationLevel = 1
	if f.unitIs(u.UnitType, UnitTypesZergLurker) {
		u.Sprite.SpriteType = f.getSpriteType(SpriteTypesSPRITEIDWhiteCircle2)
	} else {
		u.Sprite.SpriteType = f.getSpriteType(SpriteTypesSPRITEIDWhiteCircleInvisible)
	}
}

func (f *StateFunctions) unitHasCloakedUnitLink(u *Unit) bool {
	return u.CloakedUnitLink.Prev != nil || f.st.CloakedUnits.Front() == u
}

func (f *StateFunctions) setUnitCloaked(u *Unit) {
	u.CloakCounter++
	if u.CloakCounter == 1 && !f.uRequiresDetector(u) && !f.unitHasCloakedUnitLink(u) {
		f.st.CloakedUnits.PushFront(u)
	}
}

func approximateDistance(vec Xy) uint32 {
	x := absUnsigned(vec.X)
	y := absUnsigned(vec.Y)
	if x < y {
		x, y = y, x
	}
	if x/4 < y {
		return x + x/32 + y*3/8 - x/32 + y*3/64 - x/128 + y*3/256
	}
	return x + x/32 + y*3/256
}

func (f *StateFunctions) crappyMoveToTarget(u *Unit, target *Unit) {
	targetPos := target.Sprite.Position
	speed := u.FlingyTopSpeed.MulInt(2).IntegerPart()
	if target.Position != target.MoveTarget.Pos && speed != 0 {
		if f.uMovementFlag(&target.Flingy, 4) {
			targetPos = target.MoveTarget.Pos
		} else {
			targetPos = f.predictTargetPosition(u, target, speed)
		}
	}
	f.setFlingyMoveTarget(&u.Flingy, targetPos, nil)
	f.setNextTargetWaypoint(&u.Flingy, targetPos)
}

func (f *StateFunctions) predictTargetPosition(u *Unit, target *Unit, speed int32) Xy {
	t := int32(approximateDistance(XySub(target.Sprite.Position, u.Sprite.Position)) / uint32(speed))
	var predSpeed Fp8
	if u.FlingyMovementType == 2 {
		predSpeed = target.FlingyTopSpeed
	} else {
		predSpeed = target.NextSpeed.Add(target.FlingyAcceleration.MulInt(t).DivInt(2))
		if target.FlingyTopSpeed.Less(predSpeed) {
			predSpeed = target.FlingyTopSpeed
		}
	}
	predDistance := predSpeed.MulInt(t)
	targetPos := XyAdd(target.Sprite.Position, f.toXy(f.directionXY(target.NextVelocityDirection, predDistance)))
	return f.restrictPosToMapBounds(targetPos)
}

func (f *StateFunctions) getDefaultGatherOrder(u *Unit, target *Unit) *OrderTypeT {
	if f.utFlag(target.UnitType, 0x800) {
		return f.getOrderType(OrdersMove)
	}
	if f.unitIsMineralField(target.UnitType) {
		if u.CarryingFlags&^3 == 0 {
			return f.getOrderType(OrdersMoveToMinerals)
		}
		return f.getOrderType(OrdersMove)
	} else if f.unitCanGatherGas(u, target) || f.unitIs(target.UnitType, UnitTypesResourceVespeneGeyser) {
		if u.CarryingFlags&^3 == 0 {
			return f.getOrderType(OrdersMoveToGas)
		}
		return f.getOrderType(OrdersMove)
	} else if f.unitIsActiveResourceDepot(target) && target.Owner == u.Owner {
		if u.CarryingFlags&2 != 0 {
			return f.getOrderType(OrdersReturnMinerals)
		} else if u.CarryingFlags&1 != 0 {
			return f.getOrderType(OrdersReturnGas)
		}
		return nil
	}
	return nil
}

func (f *StateFunctions) getDefaultOrder(action uint, u *Unit, pos Xy, target *Unit, targetUnitType *UnitTypeT) *OrderTypeT {
	switch action {
	case 0:
		return nil
	case 1:
		return f.getDefaultMoveOrder(u, target, targetUnitType)
	case 2:
		return f.getDefaultTransportOrder(u, target, targetUnitType)
	case 3:
		if target == nil || f.unitIsSpecialBeacon(target.UnitType) || f.utPowerup(target.UnitType) {
			return u.UnitType.ReturnToIdle.Get()
		} else if f.unitTargetIsEnemy(u, target) {
			return f.getOrderType(OrdersAttackDefault)
		}
		return u.UnitType.ReturnToIdle.Get()
	case 4:
		var o *OrderTypeT
		if target != nil {
			o = f.getDefaultGatherOrder(u, target)
		}
		if o != nil {
			return o
		}
		return f.getDefaultOrder(1, u, pos, target, targetUnitType)
	case 5:
		return f.getDefaultWorkerOrder(u, pos, target, targetUnitType)
	case 6:
		return nil
	default:
		throwError("get_default_order: unknown action %d", action)
	}
	return nil
}

func (f *StateFunctions) getDefaultMoveOrder(u *Unit, target *Unit, targetUnitType *UnitTypeT) *OrderTypeT {
	if f.uGroundedBuilding(u) {
		return nil
	}
	if target != nil {
		if f.unitIsSpecialBeacon(target.UnitType) || f.utPowerup(target.UnitType) {
			return f.getOrderType(OrdersMove)
		} else if f.unitTargetIsEnemy(u, target) {
			return f.getOrderType(OrdersAttackDefault)
		} else if f.unitProvidesSpace(target) && f.unitCanLoadTarget(target, u) {
			return f.getOrderType(OrdersEnterTransport)
		} else if f.uBurrowed(target) {
			return f.getOrderType(OrdersMove)
		}
		return f.getOrderType(OrdersFollow)
	}
	if targetUnitType != nil {
		return f.getOrderType(OrdersRightClickAction)
	}
	return f.getOrderType(OrdersMove)
}

func (f *StateFunctions) getDefaultTransportOrder(u *Unit, target *Unit, targetUnitType *UnitTypeT) *OrderTypeT {
	if f.uGroundedBuilding(u) {
		if f.unitIsFactory(u) {
			if target != nil {
				return f.getOrderType(OrdersRallyPointUnit)
			}
			return f.getOrderType(OrdersRallyPointTile)
		}
		return nil
	}
	if target != nil {
		if f.unitProvidesSpace(u) && f.unitCanLoadTarget(u, target) {
			return f.getOrderType(OrdersPickupTransport)
		} else if f.unitProvidesSpace(target) && f.unitCanLoadTarget(target, u) {
			return f.getOrderType(OrdersEnterTransport)
		} else if f.uBurrowed(target) {
			return f.getOrderType(OrdersMove)
		} else if f.unitIsQueen(u.UnitType) && !f.uInvincible(target) && f.unitCanBeInfested(target) {
			return f.getOrderType(OrdersCastInfestation)
		} else if f.unitIs(u.UnitType, UnitTypesTerranMedic) {
			return f.getOrderType(OrdersHealMove)
		}
		return f.getOrderType(OrdersFollow)
	}
	if f.unitIsQueen(u.UnitType) && targetUnitType != nil {
		return f.getOrderType(OrdersRightClickAction)
	}
	return f.getOrderType(OrdersMove)
}

func (f *StateFunctions) getDefaultWorkerOrder(u *Unit, pos Xy, target *Unit, targetUnitType *UnitTypeT) *OrderTypeT {
	if target == nil {
		return f.getDefaultOrder(1, u, pos, target, targetUnitType)
	}
	order := f.getDefaultGatherOrder(u, target)
	if order != nil {
		return order
	}
	if f.uGroundedBuilding(target) && !f.uCompleted(target) {
		if target.Owner == u.Owner && f.unitRace(target.UnitType) == RaceTerran {
			if !f.utAddon(target.UnitType) {
				if target.ConnectedUnit == nil || target.ConnectedUnit.OrderTarget.Unit != target {
					return f.getOrderType(OrdersConstructingBuilding)
				}
			}
		}
	} else if f.st.Alliances[u.Owner][target.Owner] != 0 {
		canEnter := f.unitProvidesSpace(target) && f.unitCanLoadTarget(target, u)
		if (f.utBuilding(target.UnitType) || !canEnter) && f.unitRace(target.UnitType) == RaceTerran && f.utMechanical(target.UnitType) && f.uCompleted(target) && target.Hp.Less(target.UnitType.Hitpoints) {
			return f.getOrderType(OrdersRepair)
		} else if canEnter {
			return f.getOrderType(OrdersEnterTransport)
		}
	}
	return f.getDefaultOrder(1, u, pos, target, targetUnitType)
}

func (f *StateFunctions) defaultAction(u *Unit) uint {
	var action uint
	if u.UnitType.ID == UnitTypesZergLurker && f.uBurrowed(u) {
		action = 3
	} else {
		action = u.UnitType.RightClickAction
		if action == 0 && f.uGroundedBuilding(u) && f.unitIsFactory(u) {
			action = 2
		}
	}
	if f.uHallucination(u) {
		if action == 4 || action == 5 {
			action = 1
		}
	}
	return action
}

func (f *StateFunctions) spellCastTargetMovement(u *Unit, tech *TechTypeT, rangeValue int32) bool {
	target := u.OrderTarget.Unit
	if target == nil || !f.spellOrderValid(u) {
		f.orderDone(u, OrderTarget{})
		return false
	}
	if u.Energy.Less(Fp8Integer(tech.EnergyCost)) {
		f.orderDone(u, OrderTarget{})
		return false
	}
	u.OrderTarget.Pos = target.Sprite.Position
	f.setNextTargetWaypoint(&u.Flingy, target.Sprite.Position)
	if f.unitCanSeeTarget(u, target) {
		if f.uMovementFlag(&u.Flingy, 2) {
			if !f.uMovementFlag(&target.Flingy, 2) || Fp8Extend(target.NextVelocityDirection.Sub(u.NextVelocityDirection)).Abs().Greater(Fp8FromRaw(32)) {
				rangeValue += f.unitHaltDistance(&u.Flingy).IntegerPart()
			}
		}
		if f.unitTargetInRange(u, target, rangeValue) {
			f.stopUnit(u)
			return f.isFacingNextTargetWaypoint(u, Fp8FromRaw(5))
		}
	}
	if f.unitIsAtMoveTarget(&u.Flingy) && f.uImmovable(u) {
		f.stopUnit(u)
		f.orderDone(u, OrderTarget{})
		return false
	}
	f.tryFollowUnit(u, target)
	return false
}
