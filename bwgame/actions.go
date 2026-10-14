package bwgame

import (
	"iter"
	"math"
)

type unitSelection = StaticVector[*Unit, [12]*Unit]

func eraseFromSelection(selection *unitSelection, pos int) {
	for i := pos; i+1 < selection.size; i++ {
		selection.data[i] = selection.data[i+1]
	}
	selection.size--
}

func findInSelection(selection *unitSelection, u *Unit) int {
	for i := 0; i < selection.size; i++ {
		if selection.data[i] == u {
			return i
		}
	}
	return selection.size
}

func (f *ActionFunctions) unitCanBeMultiSelected(u *Unit) bool {
	if f.utBuilding(u.UnitType) {
		return false
	}
	if f.utFlag(u.UnitType, UnitTypeFlagsT(0x800)) {
		return false
	}
	if f.unitIsDisabled(u) {
		return false
	}
	tid := u.UnitType.ID
	if tid >= UnitTypesSpecialFloorMissileTrap && tid <= UnitTypesSpecialRightWallFlameTrap {
		return false
	}
	switch tid {
	case UnitTypesTerranVultureSpiderMine,
		UnitTypesZergEgg,
		UnitTypesCritterRhynadon,
		UnitTypesCritterBengalaas,
		UnitTypesCritterScantid,
		UnitTypesCritterKakaru,
		UnitTypesCritterRagnasaur,
		UnitTypesCritterUrsadon,
		UnitTypesSpellDarkSwarm,
		UnitTypesSpellDisruptionWeb:
		return false
	}
	return true
}

func (f *ActionFunctions) selectedUnits(owner int32) iter.Seq[*Unit] {
	selection := &f.actionSt.Selection[owner]
	end := selection.size
	return func(yield func(*Unit) bool) {
		for i := 0; i != end; i++ {
			u := selection.data[i]
			if u == nil || f.unitDead(u) {
				continue
			}
			if !yield(u) {
				return
			}
		}
	}
}

func (f *ActionFunctions) controlGroup(owner int32, groupN uint) iter.Seq[*Unit] {
	group := &f.actionSt.ControlGroups[owner][groupN]
	end := group.size
	return func(yield func(*Unit) bool) {
		for i := 0; i != end; i++ {
			u := f.getUnitByID(group.data[i])
			if u == nil {
				continue
			}
			if !yield(u) {
				return
			}
		}
	}
}

func (f *ActionFunctions) getFirstSelectedUnit(owner int32) *Unit {
	for u := range f.selectedUnits(owner) {
		return u
	}
	return nil
}

func (f *ActionFunctions) getSingleSelectedUnit(owner int32) *Unit {
	var single *Unit
	count := 0
	for u := range f.selectedUnits(owner) {
		count++
		if count == 2 {
			return nil
		}
		single = u
	}
	return single
}

func (f *ActionFunctions) onUnitDeselect(u *Unit) {
	for i := range f.actionSt.Selection {
		selection := &f.actionSt.Selection[i]
		pos := findInSelection(selection, u)
		if pos != selection.size {
			eraseFromSelection(selection, pos)
		}
	}
}

type groupMoveSelectionSummary struct {
	nUnits                   int32
	anyCollisionEnabledUnits bool
	area                     Rect
	sumPos                   Xy
	unitsArea                uint32
}

func (f *ActionFunctions) summarizeSelectionForGroupMove(owner int32) groupMoveSelectionSummary {
	s := groupMoveSelectionSummary{area: Rect{From: Xy{X: math.MaxInt32, Y: math.MaxInt32}, To: Xy{}}}
	for u := range f.selectedUnits(owner) {
		s.nUnits++
		if u.PathingFlags&1 != 0 {
			s.anyCollisionEnabledUnits = true
		}
		pos := u.Sprite.Position
		if pos.X < s.area.From.X {
			s.area.From.X = pos.X
		}
		if pos.Y < s.area.From.Y {
			s.area.From.Y = pos.Y
		}
		if pos.X > s.area.To.X {
			s.area.To.X = pos.X
		}
		if pos.Y > s.area.To.Y {
			s.area.To.Y = pos.Y
		}
		s.sumPos = XyAdd(s.sumPos, pos)
		w := u.UnitType.Dimensions.From.X + u.UnitType.Dimensions.To.X + 1
		h := u.UnitType.Dimensions.From.Y + u.UnitType.Dimensions.To.Y + 1
		s.unitsArea += uint32(w * h)
	}
	return s
}

func (f *ActionFunctions) groupMoveFindTarget(bb Rect, sourcePos Xy, targetPos Xy) Xy {
	sourceRegion := f.getRegionAtPreferWalkable(sourcePos)
	targetRegion := f.getRegionAtPreferWalkable(targetPos)
	if !f.isWalkable(targetPos) || sourceRegion.GroupIndex != targetRegion.GroupIndex {
		var pf Pathfinder
		if f.pathfinderFindLongPathFromTo(&pf, sourcePos, targetPos) {
			targetRegion = pf.LongPath[len(pf.LongPath)-1]
		}
	}
	targetPos = f.pathfinderAdjustDestination(targetRegion, targetPos)
	_, targetPos = f.pathfinderAdjustTargetPos(bb, targetPos)
	return targetPos
}

func (f *ActionFunctions) calcGroupMove(g *GroupMove, owner int32, targetPos Xy, canBeObstructed bool) {
	g.OriginalTargetPos = targetPos
	g.HasMoveOffset = false
	g.TargetIsInUnitBounds = false

	s := f.summarizeSelectionForGroupMove(owner)
	if s.nUnits == 0 {
		return
	}

	if s.anyCollisionEnabledUnits && canBeObstructed {
		if s.nUnits == 1 {
			u := f.getFirstSelectedUnit(owner)
			if !f.unitTypeCanFitAt(u.UnitType, targetPos) {
				targetPos = f.groupMoveFindTarget(f.unitTypeInnerBoundingBox(u.UnitType), u.Sprite.Position, targetPos)
			}
		} else {
			unitsAreaSquareWidth := int32(isqrt(s.unitsArea * 2))
			if !f.rectangleCanFitAt(targetPos, unitsAreaSquareWidth, unitsAreaSquareWidth) {
				var bb Rect
				bb.From.X = -unitsAreaSquareWidth / 2
				bb.From.Y = -unitsAreaSquareWidth / 2
				bb.To.X = bb.From.X + unitsAreaSquareWidth
				bb.To.Y = bb.From.Y + unitsAreaSquareWidth
				targetPos = f.groupMoveFindTarget(bb, XyDivScalar(s.sumPos, s.nUnits), targetPos)
			}
		}
		targetPos = f.restrictPosToMapBounds(targetPos)
	}
	if s.nUnits >= 2 {
		if f.isPosInInnerBounds(g.OriginalTargetPos, s.area) {
			g.TargetIsInUnitBounds = true
			g.MoveOffset = Xy{}
		} else if s.anyCollisionEnabledUnits {
			if s.area.To.X-s.area.From.X <= 192 && s.area.To.Y-s.area.From.Y <= 192 {
				g.MoveOffset = XySub(g.OriginalTargetPos, XyDivScalar(s.sumPos, s.nUnits))
				g.HasMoveOffset = true
			}
		} else {
			if s.area.To.X-s.area.From.X <= 256 && s.area.To.Y-s.area.From.Y <= 256 {
				g.MoveOffset = XySub(g.OriginalTargetPos, XyDivScalar(s.sumPos, s.nUnits))
				g.HasMoveOffset = true
			}
		}
	}

	g.TargetPos = targetPos
}

func regionIsSameOrWalkableNeighbor(source *Region, r *Region) bool {
	if r == source {
		return true
	}
	for _, n := range source.WalkableNeighbors {
		if n == r {
			return true
		}
	}
	return false
}

func clampToTargetBox(v int32, target int32) int32 {
	if v <= target-32 {
		return target - 32
	} else if v >= target+32 {
		return target + 32
	}
	return v
}

func (f *ActionFunctions) getGroupMovePos(u *Unit, pos Xy, g *GroupMove) Xy {
	targetPos := g.OriginalTargetPos
	if u.PathingFlags&1 != 0 {
		targetPos = g.TargetPos
	}
	if g.TargetIsInUnitBounds {
		pos = u.Sprite.Position
		pos.X = clampToTargetBox(pos.X, targetPos.X)
		pos.Y = clampToTargetBox(pos.Y, targetPos.Y)
	} else if g.HasMoveOffset {
		pos.X = u.Sprite.Position.X + g.MoveOffset.X
		pos.Y = u.Sprite.Position.Y + g.MoveOffset.Y
	}

	pos = f.restrictMoveTargetToValidBounds(u.UnitType, pos)
	if ^u.PathingFlags&1 != 0 {
		return pos
	}
	sourceRegion := f.getRegionAt(targetPos)
	if f.unitTypeCanFitAt(u.UnitType, pos) {
		if regionIsSameOrWalkableNeighbor(sourceRegion, f.getRegionAt(pos)) {
			return pos
		}
	}
	distanceToTarget := f.xyLength(XySub(targetPos, pos))
	if distanceToTarget == 0 {
		distanceToTarget = 1
	}
	stepDistance := min(distanceToTarget/4, 16)
	if stepDistance < 2 {
		stepDistance = distanceToTarget
	}
	for distance := stepDistance; distance <= distanceToTarget; distance += stepDistance {
		npos := XyAdd(pos, XyDivScalar(XyMulScalar(XySub(targetPos, pos), distance), distanceToTarget))
		if f.unitTypeCanFitAt(u.UnitType, npos) {
			if regionIsSameOrWalkableNeighbor(sourceRegion, f.getRegionAt(npos)) {
				return npos
			}
		}
	}
	return targetPos
}

func (f *ActionFunctions) morphArchonImpl(owner int32, isDarkArchon bool) bool {
	first := f.getFirstSelectedUnit(owner)
	if first == nil {
		return false
	}
	techID, templarID, orderID := TechTypesArchonWarp, UnitTypesProtossHighTemplar, OrdersArchonWarp
	if isDarkArchon {
		techID, templarID, orderID = TechTypesDarkArchonMeld, UnitTypesProtossDarkTemplar, OrdersDarkArchonMeld
	}
	if !f.unitCanUseTechOwner(first, f.getTechType(techID), owner) {
		return false
	}
	retval := false
	var units []*Unit
	for u := range f.selectedUnits(owner) {
		if f.unitIs(u.UnitType, templarID) {
			units = append(units, u)
		}
	}
	if len(units) < 2 {
		return false
	}
	for index := 0; index != len(units); index++ {
		u := units[index]
		if u == nil {
			continue
		}
		nearestDistance := int32(math.MaxInt32)
		var nearestUnit *Unit
		for index2 := index + 1; index2 != len(units); index2++ {
			u2 := units[index2]
			if u2 == nil {
				continue
			}
			relpos := XySub(u.Sprite.Position, u2.Sprite.Position)
			d := relpos.X*relpos.X + relpos.Y*relpos.Y
			if d < nearestDistance {
				units[index2] = nearestUnit
				nearestDistance = d
				nearestUnit = u2
			}
		}
		if nearestUnit == nil {
			continue
		}
		f.setUnitOrderUnit(u, f.getOrderType(orderID), nearestUnit)
		f.setUnitOrderUnit(nearestUnit, f.getOrderType(orderID), u)
		retval = true
	}
	return retval
}

func (f *ActionFunctions) actionSelect(owner int32, units []*Unit) bool {
	selection := &f.actionSt.Selection[owner]
	selection.Clear()
	retval := false
	for _, u := range units {
		if u == nil || u.UnitType.ID == UnitTypesTerranNuclearMissile {
			continue
		}
		if findInSelection(selection, u) != selection.size {
			continue
		}
		if !f.usHidden(&u.Thingy) && (selection.Empty() || f.unitCanBeMultiSelected(u)) {
			if selection.Size() == 12 {
				throwError("action_select: attempt to select more than 12 units")
			}
			selection.PushBack(u)
			retval = true
		}
	}
	return retval
}

func (f *ActionFunctions) actionSelectUnit(owner int32, u *Unit) bool {
	return f.actionSelect(owner, []*Unit{u})
}

func (f *ActionFunctions) actionShiftSelect(owner int32, units []*Unit) bool {
	selection := &f.actionSt.Selection[owner]
	if selection.Size()+len(units) > 12 {
		return false
	}
	retval := false
	for _, u := range units {
		if u == nil {
			continue
		}
		if findInSelection(selection, u) != selection.size {
			continue
		}
		if !f.usHidden(&u.Thingy) && (selection.Empty() || f.unitCanBeMultiSelected(u)) {
			if selection.Size() == 12 {
				throwError("action_shift_select: attempt to select more than 12 units")
			}
			selection.PushBack(u)
			retval = true
		}
	}
	return retval
}

func (f *ActionFunctions) actionShiftSelectUnit(owner int32, u *Unit) bool {
	return f.actionShiftSelect(owner, []*Unit{u})
}

func (f *ActionFunctions) actionDeselect(owner int32, units []*Unit) bool {
	selection := &f.actionSt.Selection[owner]
	retval := false
	for _, u := range units {
		if u == nil {
			continue
		}
		pos := findInSelection(selection, u)
		if pos != selection.size {
			eraseFromSelection(selection, pos)
			retval = true
		}
	}
	return retval
}

func (f *ActionFunctions) actionDeselectUnit(owner int32, u *Unit) bool {
	return f.actionDeselect(owner, []*Unit{u})
}

func (f *ActionFunctions) actionTrain(owner int32, unitType *UnitTypeT) bool {
	u := f.getSingleSelectedUnit(owner)
	if u == nil {
		return false
	}
	if u.Owner != owner {
		return false
	}
	if !f.unitCanBuild(u, unitType) {
		return false
	}
	if unitType.ID > UnitTypesSpellDisruptionWeb {
		return false
	}
	if !f.buildQueuePush(u, unitType) {
		return false
	}
	f.setSecondaryOrder(u, f.getOrderType(OrdersTrain))
	return true
}

func buildingPosAtTile(tilePos XyT[uint], placementSize Xy) Xy {
	return Xy{X: int32(32*tilePos.X) + placementSize.X/2, Y: int32(32*tilePos.Y) + placementSize.Y/2}
}

func (f *ActionFunctions) actionBuild(owner int32, orderType *OrderTypeT, unitType *UnitTypeT, tilePos XyT[uint]) bool {
	u := f.getSingleSelectedUnit(owner)
	if u == nil {
		return false
	}
	if u.Owner != owner {
		return false
	}
	if !f.unitBuildOrderValid(u, orderType, unitType, owner) {
		return false
	}
	pos := buildingPosAtTile(tilePos, unitType.PlacementSize)
	if f.utAddon(unitType) {
		if f.canPlaceBuilding(u, owner, unitType, pos, false, false) {
			builderPos := buildingPosAtTile(tilePos, u.UnitType.PlacementSize)
			builderPos.X -= unitType.AddonPosition.X / 32 * 32
			builderPos.Y -= unitType.AddonPosition.Y / 32 * 32
			if f.canPlaceBuilding(u, owner, u.UnitType, builderPos, false, false) {
				f.placeBuilding(u, orderType, unitType, builderPos)
			}
		}
	} else {
		if f.canPlaceBuilding(u, owner, unitType, pos, false, false) {
			f.placeBuilding(u, orderType, unitType, pos)
		}
	}
	return true
}

func unitPlacementContainsPos(n *Unit, pos Xy) bool {
	size := n.UnitType.PlacementSize
	if uint32(n.Sprite.Position.X+size.X/2)-uint32(pos.X) >= uint32(size.X) {
		return false
	}
	if uint32(n.Sprite.Position.Y+size.Y/2)-uint32(pos.Y) >= uint32(size.Y) {
		return false
	}
	return true
}

func (f *ActionFunctions) setRallyToUnit(u *Unit, target *Unit) *Unit {
	if target == nil {
		target = u
	}
	u.Building.Rally.Unit = target
	u.Building.Rally.Pos = target.Sprite.Position
	return target
}

func (f *ActionFunctions) setRallyToPos(u *Unit, pos Xy) {
	u.Building.Rally.Unit = nil
	u.Building.Rally.Pos = pos
}

func (f *ActionFunctions) actionDefaultOrder(owner int32, pos Xy, target *Unit, targetUnitType *UnitTypeT, queue bool) bool {
	if !f.isPosInMapBounds(pos) {
		return false
	}
	if target != nil {
		targetUnitType = nil
	} else if targetUnitType != nil {
		if f.playerPositionIsVisible(owner, pos) {
			area := Rect{From: XySub(pos, Xy{X: 96, Y: 96}), To: XyAdd(pos, Xy{X: 96, Y: 96})}
			wantedType := targetUnitType
			target = f.findUnitNoexpand(area, func(n *Unit) bool {
				if n.UnitType != wantedType {
					return false
				}
				return unitPlacementContainsPos(n, pos)
			})
			targetUnitType = nil
		}
	}
	var g GroupMove
	hasCalculatedGroupMove := false
	retval := false
	for u := range f.selectedUnits(owner) {
		order := f.getDefaultOrder(f.defaultAction(u), u, pos, target, targetUnitType)
		if order == nil {
			continue
		}
		if order.ID == OrdersRallyPointUnit {
			if f.unitIsFactory(u) {
				retval = true
				target = f.setRallyToUnit(u, target)
			}
			continue
		}
		if target == u {
			continue
		}
		if order.ID == OrdersRallyPointTile {
			if f.unitIsFactory(u) {
				retval = true
				f.setRallyToPos(u, pos)
			}
			continue
		}
		if order.ID == OrdersAttackDefault {
			if u.Subunit != nil {
				f.issueOrder(u.Subunit, queue, u.Subunit.UnitType.AttackUnit.Get(), OrderTarget{Unit: target})
			}
			order = u.UnitType.AttackUnit.Get()
		}
		if !f.unitCanReceiveOrder(u, order, owner) {
			continue
		}
		retval = true
		if target != nil {
			f.issueOrder(u, queue, order, OrderTarget{Unit: target})
		} else {
			orderTarget := OrderTarget{Position: pos, UnitType: targetUnitType}
			if targetUnitType == nil {
				if !hasCalculatedGroupMove {
					hasCalculatedGroupMove = true
					f.calcGroupMove(&g, owner, pos, true)
				}
				orderTarget.Position = f.getGroupMovePos(u, pos, &g)
			}
			f.issueOrder(u, queue, order, orderTarget)
		}
		u.UserActionFlags |= 2
	}
	return retval
}

func actionOrderIsForbidden(id Orders) bool {
	switch id {
	case OrdersDie,
		OrdersInfestedCommandCenter,
		OrdersSpiderMine,
		OrdersDroneStartBuild,
		OrdersInfestingCommandCenter,
		OrdersPlaceBuilding,
		OrdersPlaceProtossBuilding,
		OrdersCreateProtossBuilding,
		OrdersConstructingBuilding,
		OrdersPlaceAddon,
		OrdersBuildNydusExit,
		OrdersBuildingLand,
		OrdersBuildingLiftoff,
		OrdersDroneLiftOff,
		OrdersHarvest2,
		OrdersMoveToGas,
		OrdersWaitForGas,
		OrdersHarvestGas,
		OrdersMoveToMinerals,
		OrdersWaitForMinerals,
		OrdersMiningMinerals,
		OrdersGatheringInterrupted,
		OrdersGatherWaitInterrupted,
		OrdersCTFCOP2:
		return true
	}
	return false
}

func (f *ActionFunctions) unitCanTakeActionOrder(u *Unit, order *OrderTypeT, target *Unit, owner int32) bool {
	if order.TechType == TechTypesNone {
		if !f.unitCanReceiveOrder(u, order, owner) {
			return false
		}
	} else {
		if !f.unitCanUseTechOwner(u, f.getTechType(order.TechType), owner) {
			return false
		}
	}
	if u == target {
		if !f.unitOrderCanTargetSelf(u, order) {
			return false
		}
	}
	if u.OrderType.ID == OrdersNukeLaunch {
		return false
	}
	if f.defaultAction(u) == 0 {
		oid := order.ID
		if oid != OrdersRallyPointUnit && oid != OrdersRallyPointTile && oid != OrdersRechargeShieldsBattery && oid != OrdersPickupBunker {
			return false
		}
	}
	return true
}

func (f *ActionFunctions) actionOrder(owner int32, inputOrder *OrderTypeT, pos Xy, target *Unit, targetUnitType *UnitTypeT, queue bool) bool {
	if !f.isPosInMapBounds(pos) {
		return false
	}
	if actionOrderIsForbidden(inputOrder.ID) {
		return false
	}
	canBeObstructed := inputOrder.CanBeObstructed
	var g GroupMove
	hasCalculatedGroupMove := false
	retval := false
	for u := range f.selectedUnits(owner) {
		order := inputOrder
		if !f.unitCanTakeActionOrder(u, order, target, owner) {
			continue
		}
		if f.unitIs(u.UnitType, UnitTypesTerranMedic) {
			if order.ID == OrdersAttackMove || order.ID == OrdersAttackDefault {
				order = f.getOrderType(OrdersHealMove)
			}
		}
		subunitOrder := f.getOrderType(OrdersNothing)
		if order.ID == OrdersAttackDefault {
			if target != nil {
				order = u.UnitType.AttackUnit.Get()
			} else {
				order = u.UnitType.AttackMove.Get()
			}
			if order.ID == OrdersNothing {
				continue
			}
			if u.Subunit != nil {
				if target != nil {
					subunitOrder = u.Subunit.UnitType.AttackUnit.Get()
				} else {
					subunitOrder = u.Subunit.UnitType.AttackMove.Get()
				}
			}
		}
		if order.ID == OrdersRallyPointUnit {
			if f.unitIsFactory(u) {
				retval = true
				target = f.setRallyToUnit(u, target)
			}
			u.UserActionFlags |= 2
		} else if order.ID == OrdersRallyPointTile {
			if f.unitIsFactory(u) {
				retval = true
				f.setRallyToPos(u, pos)
			}
			u.UserActionFlags |= 2
		} else {
			retval = true
			var orderTarget OrderTarget
			if target != nil {
				orderTarget.Unit = target
			} else {
				orderTarget.Position = pos
				orderTarget.UnitType = targetUnitType
				if targetUnitType == nil {
					if !hasCalculatedGroupMove {
						hasCalculatedGroupMove = true
						f.calcGroupMove(&g, owner, pos, canBeObstructed)
					}
					orderTarget.Position = f.getGroupMovePos(u, pos, &g)
				}
			}
			f.issueOrder(u, queue, order, orderTarget)
			if subunitOrder.ID != OrdersNothing {
				f.issueOrder(u.Subunit, queue, subunitOrder, orderTarget)
			}
			u.UserActionFlags |= 2
		}
	}
	return retval
}

func (f *ActionFunctions) actionStop(owner int32, queue bool) bool {
	retval := false
	for u := range f.selectedUnits(owner) {
		if f.unitCanReceiveOrder(u, f.getOrderType(OrdersStop), owner) {
			retval = true
			f.issueOrder(u, queue, f.getOrderType(OrdersStop), OrderTarget{})
		}
	}
	return retval
}

func (f *ActionFunctions) actionCancelBuildingUnit(owner int32) bool {
	u := f.getSingleSelectedUnit(owner)
	if u == nil || u.Owner != owner {
		return false
	}
	f.cancelBuildingUnit(u)
	return true
}

func (f *ActionFunctions) unitBuildQueueCannotBeCancelled(u *Unit) bool {
	for _, id := range []UnitTypes{
		UnitTypesZergLarva,
		UnitTypesZergMutalisk,
		UnitTypesZergHydralisk,
		UnitTypesZergHatchery,
		UnitTypesZergLair,
		UnitTypesZergSpire,
		UnitTypesZergCreepColony,
	} {
		if f.unitIs(u.UnitType, id) {
			return true
		}
	}
	return false
}

func (f *ActionFunctions) actionCancelBuildQueue(owner int32, slot uint) bool {
	u := f.getSingleSelectedUnit(owner)
	if u == nil || u.Owner != owner {
		return false
	}
	if !f.uCompleted(u) {
		return false
	}
	if f.unitBuildQueueCannotBeCancelled(u) {
		return false
	}
	if u.BuildQueue.Empty() {
		return false
	}
	if slot == 254 {
		slot = uint(u.BuildQueue.Size()) - 1
	}
	if slot >= uint(u.BuildQueue.Size()) {
		return false
	}
	f.cancelBuildQueueIndex(u, slot)
	return true
}

func (f *ActionFunctions) actionControlGroup(owner int32, groupN uint, subaction int32) bool {
	if groupN >= 10 {
		return false
	}
	switch subaction {
	case 1:
		return f.recallControlGroup(owner, groupN)
	case 0:
		return f.createControlGroup(owner, groupN)
	case 2:
		return f.addToControlGroup(owner, groupN)
	}
	throwError("action_control_group: unknown subaction %d", subaction)
	return false
}

func (f *ActionFunctions) recallControlGroup(owner int32, groupN uint) bool {
	retval := false
	group := &f.actionSt.ControlGroups[owner][groupN]
	if group.Empty() {
		return false
	}
	selection := &f.actionSt.Selection[owner]
	selection.Clear()
	for i := 0; i != group.Size(); i++ {
		u := f.getUnitByID(group.At(i))
		if u != nil && !f.usHidden(&u.Thingy) && u.Owner == owner && (group.Size() == 1 || f.unitCanBeMultiSelected(u)) {
			selection.PushBack(u)
			retval = true
		} else {
			last := group.Size() - 1
			if i != last {
				*group.Ptr(i), *group.Ptr(last) = group.At(last), group.At(i)
			}
			group.PopBack()
			i--
		}
	}
	return retval
}

func (f *ActionFunctions) createControlGroup(owner int32, groupN uint) bool {
	retval := false
	group := &f.actionSt.ControlGroups[owner][groupN]
	group.Clear()
	for u := range f.selectedUnits(owner) {
		if u.Owner != owner {
			break
		}
		uid := f.getUnitID(u)
		if group.Size() == 12 {
			throwError("attempt to control group more than 12 units")
		}
		group.PushBack(uid)
		retval = true
	}
	return retval
}

func controlGroupContains(group *StaticVector[UnitID, [12]UnitID], uid UnitID) bool {
	for _, v := range group.Values() {
		if v == uid {
			return true
		}
	}
	return false
}

func (f *ActionFunctions) addToControlGroup(owner int32, groupN uint) bool {
	retval := false
	group := &f.actionSt.ControlGroups[owner][groupN]
	if !group.Empty() {
		u := f.getUnitByID(group.At(0))
		if u != nil && !f.unitDead(u) && !f.unitCanBeMultiSelected(u) {
			return false
		}
	}
	for u := range f.selectedUnits(owner) {
		if u.Owner != owner {
			break
		}
		uid := f.getUnitID(u)
		if !(group.Empty() || f.unitCanBeMultiSelected(u)) {
			continue
		}
		if controlGroupContains(group, uid) {
			continue
		}
		if group.Size() == 12 {
			if groupN != 9 {
				nextGroup := &f.actionSt.ControlGroups[owner][groupN+1]
				if nextGroup.Empty() {
					nextGroup.PushBack(uid)
				} else {
					*nextGroup.Ptr(0) = uid
				}
			}
			break
		}
		group.PushBack(uid)
		retval = true
		if group.Size() == 12 {
			break
		}
	}
	return retval
}

func seqIsEmpty[T any](s iter.Seq[T]) bool {
	for range s {
		return false
	}
	return true
}

func (f *ActionFunctions) actionUnloadAll(owner int32, queue bool) bool {
	retval := false
	for u := range f.selectedUnits(owner) {
		if !f.unitCanReceiveOrder(u, f.getOrderType(OrdersUnload), owner) {
			continue
		}
		if seqIsEmpty(f.loadedUnits(u)) {
			continue
		}
		f.issueOrder(u, queue, f.getOrderType(OrdersUnload), OrderTarget{})
	}
	return retval
}

func (f *ActionFunctions) actionUnload(owner int32, target *Unit) bool {
	if target == nil {
		return false
	}
	if target.Owner != owner {
		return false
	}
	if target.ConnectedUnit == nil || !f.uLoaded(target) {
		return false
	}
	return f.unitUnload(target)
}

func (f *ActionFunctions) actionLiftoff(owner int32, pos Xy) bool {
	if !f.isPosInMapBounds(pos) {
		return false
	}
	u := f.getSingleSelectedUnit(owner)
	if u == nil {
		return false
	}
	if f.unitIsConstructing(u) {
		return false
	}
	if f.unitIsResearching(u) || f.unitIsUpgrading(u) {
		return false
	}
	if !f.unitCanReceiveOrder(u, f.getOrderType(OrdersBuildingLiftoff), owner) {
		return false
	}
	f.setUnitOrderPos(u, f.getOrderType(OrdersBuildingLiftoff), pos)
	f.setQueuedOrder(u, false, u.UnitType.ReturnToIdle.Get(), OrderTarget{Position: XySub(pos, Xy{X: 0, Y: 42})})
	return true
}

func (f *ActionFunctions) actionResearch(owner int32, tech *TechTypeT) bool {
	u := f.getSingleSelectedUnit(owner)
	if u == nil {
		return false
	}
	if !f.unitCanResearchOwner(u, tech, owner) {
		return false
	}
	if !f.hasAvailableResourcesForTech(u.Owner, tech, true) {
		return false
	}
	f.st.CurrentMinerals[u.Owner] -= tech.MineralCost
	f.st.CurrentGas[u.Owner] -= tech.GasCost
	u.Building.ResearchingType = tech
	u.Building.UpgradeResearchTime = tech.ResearchTime
	f.st.TechResearching[u.Owner][tech.ID] = true
	f.spriteRunAnim(u.Sprite, IscriptAnimsIsWorking)
	f.setUnitOrder(u, f.getOrderType(OrdersResearchTech))
	return true
}

func (f *ActionFunctions) actionCancelResearch(owner int32) bool {
	u := f.getSingleSelectedUnit(owner)
	if u == nil || u.Owner != owner {
		return false
	}
	if !f.utBuilding(u.UnitType) {
		return false
	}
	if !f.uCompleted(u) {
		return false
	}
	if !f.unitIsResearching(u) {
		return false
	}
	f.cancelResearch(u, false)
	return true
}

func (f *ActionFunctions) actionUpgrade(owner int32, upgrade *UpgradeTypeT) bool {
	u := f.getSingleSelectedUnit(owner)
	if u == nil {
		return false
	}
	if !f.unitCanUpgradeOwner(u, upgrade, owner) {
		return false
	}
	if !f.hasAvailableResourcesForUpgrade(u.Owner, upgrade, true) {
		return false
	}
	f.st.CurrentMinerals[u.Owner] -= f.upgradeMineralCost(owner, upgrade)
	f.st.CurrentGas[u.Owner] -= f.upgradeGasCost(owner, upgrade)
	u.Building.UpgradingType = upgrade
	u.Building.UpgradeResearchTime = f.upgradeTimeCost(owner, upgrade)
	u.Building.UpgradingLevel = f.playerUpgradeLevel(owner, upgrade.ID) + 1
	f.st.UpgradeUpgrading[u.Owner][upgrade.ID] = true
	f.spriteRunAnim(u.Sprite, IscriptAnimsIsWorking)
	f.setUnitOrder(u, f.getOrderType(OrdersUpgrade))
	return true
}

func (f *ActionFunctions) actionCancelUpgrade(owner int32) bool {
	u := f.getSingleSelectedUnit(owner)
	if u == nil || u.Owner != owner {
		return false
	}
	if !f.utBuilding(u.UnitType) {
		return false
	}
	if !f.uCompleted(u) {
		return false
	}
	if !f.unitIsUpgrading(u) {
		return false
	}
	f.cancelUpgrade(u, false)
	return true
}

func (f *ActionFunctions) actionUnsiege(owner int32, queue bool) bool {
	retval := false
	for u := range f.selectedUnits(owner) {
		if u.Owner != owner {
			continue
		}
		if !f.unitIsSiegedTank(u.UnitType) {
			continue
		}
		if u.OrderType.ID == OrdersUnsieging {
			continue
		}
		if !f.unitCanUseTech(u, f.getTechType(TechTypesTankSiegeMode)) {
			continue
		}
		f.issueOrder(u, queue, f.getOrderType(OrdersUnsieging), OrderTarget{})
		retval = true
	}
	return retval
}

func (f *ActionFunctions) actionSiege(owner int32, queue bool) bool {
	retval := false
	for u := range f.selectedUnits(owner) {
		if u.Owner != owner {
			continue
		}
		if !f.unitIsUnsiegedTank(u.UnitType) {
			continue
		}
		if u.OrderType.ID == OrdersSieging {
			continue
		}
		if !f.unitCanUseTech(u, f.getTechType(TechTypesTankSiegeMode)) {
			continue
		}
		f.issueOrder(u, queue, f.getOrderType(OrdersSieging), OrderTarget{})
		retval = true
	}
	return retval
}

func (f *ActionFunctions) returnCargoOrder(u *Unit) *OrderTypeT {
	if u.CarryingFlags&2 != 0 {
		return f.getOrderType(OrdersReturnMinerals)
	} else if u.CarryingFlags&1 != 0 {
		return f.getOrderType(OrdersReturnGas)
	}
	return nil
}

func (f *ActionFunctions) actionReturnCargo(owner int32, queue bool) bool {
	retval := false
	for u := range f.selectedUnits(owner) {
		order := f.returnCargoOrder(u)
		if order == nil {
			continue
		}
		if !f.unitCanReceiveOrder(u, order, owner) {
			continue
		}
		f.issueOrder(u, queue, order, OrderTarget{})
		retval = true
	}
	return retval
}

func (f *ActionFunctions) holdPositionOrder(u *Unit) *OrderTypeT {
	order := f.getOrderType(OrdersHoldPosition)
	if f.unitIsCarrier(u.UnitType) {
		order = f.getOrderType(OrdersCarrierHoldPosition)
	} else if f.unitIsReaver(u.UnitType) {
		order = f.getOrderType(OrdersReaverHoldPosition)
	} else if f.unitIsQueen(u.UnitType) {
		order = f.getOrderType(OrdersQueenHoldPosition)
	} else if f.unitIs(u.UnitType, UnitTypesZergScourge) || f.unitIs(u.UnitType, UnitTypesZergInfestedTerran) {
		order = f.getOrderType(OrdersSuicideHoldPosition)
	} else if f.unitIs(u.UnitType, UnitTypesTerranMedic) {
		order = f.getOrderType(OrdersMedicHoldPosition)
	}
	return order
}

func (f *ActionFunctions) actionHoldPosition(owner int32, queue bool) bool {
	retval := false
	for u := range f.selectedUnits(owner) {
		order := f.holdPositionOrder(u)
		if !f.unitCanReceiveOrder(u, order, owner) {
			continue
		}
		f.issueOrder(u, queue, order, OrderTarget{})
		retval = true
	}
	return retval
}

func (f *ActionFunctions) cloakTech(u *Unit) *TechTypeT {
	if f.unitIsGhost(u.UnitType) {
		return f.getTechType(TechTypesPersonnelCloaking)
	} else if f.unitIsWraith(u.UnitType) {
		return f.getTechType(TechTypesCloakingField)
	}
	return nil
}

func (f *ActionFunctions) actionCloak(owner int32) bool {
	retval := false
	for u := range f.selectedUnits(owner) {
		if u.Owner != owner {
			continue
		}
		tech := f.cloakTech(u)
		if tech == nil {
			continue
		}
		if !f.unitCanUseTech(u, tech) {
			continue
		}
		if f.uRequiresDetector(u) {
			continue
		}
		if u.Energy.Less(Fp8Integer(tech.EnergyCost)) {
			continue
		}
		u.Energy = u.Energy.Sub(Fp8Integer(tech.EnergyCost))
		f.setSecondaryOrder(u, f.getOrderType(OrdersCloak))
	}
	return retval
}

func (f *ActionFunctions) actionDecloak(owner int32) bool {
	retval := false
	for u := range f.selectedUnits(owner) {
		if u.Owner != owner {
			continue
		}
		tech := f.cloakTech(u)
		if tech == nil {
			continue
		}
		if !f.unitCanUseTech(u, tech) {
			continue
		}
		f.setSecondaryOrder(u, f.getOrderType(OrdersDecloak))
	}
	return retval
}

func (f *ActionFunctions) actionSetAlliances(owner int32, alliances [12]int32) bool {
	f.st.Alliances[owner] = alliances
	for u := range f.st.PlayerUnits[owner].All() {
		if u.OrderTarget.Unit != nil && u.OrderType.TargetsEnemies {
			if alliances[u.OrderTarget.Unit.Owner] != 0 && !f.unitTargetIsEnemy(u, u.OrderTarget.Unit) {
				u.OrderTarget.Unit = nil
			}
		}
	}
	return true
}

func (f *ActionFunctions) actionSetSharedVision(owner int32, flags uint32) bool {
	flags &= 0xff
	flags |= f.st.SharedVision[owner] & 0xff00
	f.st.SharedVision[owner] = flags
	return true
}

func (f *ActionFunctions) actionCancelAddon(owner int32) bool {
	u := f.getSingleSelectedUnit(owner)
	if u == nil || u.Owner != owner {
		return false
	}
	if !f.uCompleted(u) {
		return false
	}
	if !f.uGroundedBuilding(u) {
		return false
	}
	if u.SecondaryOrderType.ID != OrdersBuildAddon {
		return false
	}
	addon := u.CurrentBuildUnit
	if addon == nil {
		return false
	}
	if f.uCompleted(addon) {
		return false
	}
	f.cancelBuildingUnit(addon)
	return true
}

func (f *ActionFunctions) actionStimPack(owner int32) bool {
	retval := false
	for u := range f.selectedUnits(owner) {
		if u.Owner != owner {
			continue
		}
		if !f.unitCanUseTech(u, f.getTechType(TechTypesStimPacks)) {
			continue
		}
		if u.Hp.LessEqual(Fp8Integer(10)) {
			continue
		}
		f.playSoundUnit(f.lcgRandRange(31, 278, 279), u, false)
		f.unitDealDamage(u, Fp8Integer(10), nil, ^int32(0), true)
		if u.StimTimer < 37 {
			u.StimTimer = 37
			f.updateUnitSpeed(u)
		}
		retval = true
	}
	return retval
}

func (f *ActionFunctions) actionCancelNuke(owner int32) bool {
	retval := false
	for u := range f.selectedUnits(owner) {
		if !f.unitIsGhost(u.UnitType) {
			continue
		}
		if u.ConnectedUnit == nil || !f.unitIs(u.ConnectedUnit.UnitType, UnitTypesTerranNuclearMissile) {
			continue
		}
		u.ConnectedUnit.ConnectedUnit = nil
		u.ConnectedUnit = nil
		retval = true
	}
	return retval
}

func (f *ActionFunctions) unitCanMorphInto(u *Unit, unitType *UnitTypeT) bool {
	if f.unitIs(u.UnitType, UnitTypesZergHydralisk) && f.unitIs(unitType, UnitTypesZergLurker) {
		return f.unitCanUseTech(u, f.getTechType(TechTypesLurkerAspect))
	} else if f.unitIs(u.UnitType, UnitTypesZergLarva) || f.unitIs(u.UnitType, UnitTypesZergMutalisk) {
		return f.unitCanBuild(u, unitType)
	}
	return false
}

func (f *ActionFunctions) actionMorph(owner int32, unitType *UnitTypeT) bool {
	retval := false
	for u := range f.selectedUnits(owner) {
		if u.Owner != owner {
			continue
		}
		if !f.unitCanMorphInto(u, unitType) {
			continue
		}
		if u.OrderType.ID == OrdersZergUnitMorph {
			continue
		}
		if !f.buildQueuePush(u, unitType) {
			continue
		}
		f.setUnitOrder(u, f.getOrderType(OrdersZergUnitMorph))
		retval = true
	}
	return retval
}

func (f *ActionFunctions) actionMorphBuilding(owner int32, unitType *UnitTypeT) bool {
	if !f.unitIsZergBuilding(unitType) {
		return false
	}
	retval := false
	for u := range f.selectedUnits(owner) {
		if u.Owner != owner {
			continue
		}
		if !f.unitIsZergBuilding(u.UnitType) {
			continue
		}
		if !f.unitCanBuild(u, unitType) {
			continue
		}
		if !f.buildQueuePush(u, unitType) {
			continue
		}
		if u.OrderType.ID != OrdersZergBuildingMorph {
			f.setUnitOrder(u, f.getOrderType(OrdersZergBuildingMorph))
		}
		retval = true
	}
	return retval
}

func (f *ActionFunctions) actionBurrow(owner int32, queue bool) bool {
	retval := false
	for u := range f.selectedUnits(owner) {
		if !f.unitCanUseTechOwner(u, f.getTechType(TechTypesBurrowing), owner) {
			continue
		}
		if f.uBurrowed(u) {
			continue
		}
		if u.OrderType.ID == OrdersBurrowing {
			continue
		}
		f.issueOrder(u, queue, f.getOrderType(OrdersBurrowing), OrderTarget{})
		retval = true
	}
	return retval
}

func (f *ActionFunctions) actionUnburrow(owner int32) bool {
	retval := false
	for u := range f.selectedUnits(owner) {
		if !f.unitCanUseTechOwner(u, f.getTechType(TechTypesBurrowing), owner) {
			continue
		}
		if !f.utCanBurrow(u.UnitType) {
			continue
		}
		f.unburrowUnit(u)
		retval = true
	}
	return retval
}

func (f *ActionFunctions) actionCancelMorph(owner int32) bool {
	retval := false
	for u := range f.selectedUnits(owner) {
		if u == nil || u.Owner != owner {
			return false
		}
		if f.cancelBuildingUnit(u) {
			retval = true
		}
	}
	return retval
}

func (f *ActionFunctions) actionMorphArchon(owner int32) bool {
	return f.morphArchonImpl(owner, false)
}

func (f *ActionFunctions) actionCarrierStop(owner int32) bool {
	retval := false
	for u := range f.selectedUnits(owner) {
		if !f.unitCanReceiveOrder(u, f.getOrderType(OrdersCarrierStop), owner) {
			continue
		}
		f.setUnitOrder(u, f.getOrderType(OrdersCarrierStop))
		retval = true
	}
	return retval
}

func (f *ActionFunctions) fighterBuildType(u *Unit) *UnitTypeT {
	if f.unitIsCarrier(u.UnitType) {
		return f.getUnitType(UnitTypesProtossInterceptor)
	} else if f.unitIsReaver(u.UnitType) {
		return f.getUnitType(UnitTypesProtossScarab)
	}
	return nil
}

func (f *ActionFunctions) actionTrainFighter(owner int32) bool {
	retval := false
	for u := range f.selectedUnits(owner) {
		if u.Owner != owner {
			continue
		}
		buildType := f.fighterBuildType(u)
		if buildType == nil {
			continue
		}
		if !f.unitCanBuild(u, buildType) {
			continue
		}
		if !f.buildQueuePush(u, buildType) {
			continue
		}
		if u.SecondaryOrderState != 2 {
			u.SecondaryOrderType = nil
			f.setSecondaryOrder(u, f.getOrderType(OrdersTrainFighter))
		}
		retval = true
	}
	return retval
}

func (f *ActionFunctions) actionMorphDarkArchon(owner int32) bool {
	return f.morphArchonImpl(owner, true)
}

func (f *ActionFunctions) actionReaverStop(owner int32) bool {
	retval := false
	for u := range f.selectedUnits(owner) {
		if !f.unitCanReceiveOrder(u, f.getOrderType(OrdersReaverStop), owner) {
			continue
		}
		f.setUnitOrder(u, f.getOrderType(OrdersReaverStop))
		retval = true
	}
	return retval
}

func (f *ActionFunctions) actionPlayerLeave(owner int32, reason int32) bool {
	if reason == 6 {
		f.playerDropped(owner)
	} else {
		f.playerLeft(owner)
	}
	return true
}

func (f *ActionFunctions) actionCheat(owner int32, flags int32) bool {
	if !f.st.CheatsEnabled {
		return false
	}
	f.st.CheatOperationCwal = flags&2 != 0
	if flags&2 != 0 {
		flags &^= 2
	}
	if flags&0x10 != 0 {
		flags &^= 0x10
		for i := 0; i != 12; i++ {
			if f.st.Players[i].Controller == PlayerControllerOccupied {
				f.st.CurrentMinerals[i] += 10000
				f.st.CurrentGas[i] += 10000
			}
		}
	}
	if flags != 0 {
		throwError("unknown cheat flag(s) %#x", flags)
	}
	return true
}

func (f *ActionFunctions) readSelectionUnits(r *DataReader, errorFormat string) []*Unit {
	n := uint(r.GetU8())
	if n > 12 {
		throwError(errorFormat, n)
	}
	units := make([]*Unit, 0, 12)
	for i := uint(0); i != n; i++ {
		uid := UnitID{RawValue: r.GetU16()}
		units = append(units, f.getUnitByID(uid))
	}
	return units
}

func (f *ActionFunctions) readActionSelect(owner int32, r *DataReader) bool {
	return f.actionSelect(owner, f.readSelectionUnits(r, "invalid selection of %d units"))
}

func (f *ActionFunctions) readActionShiftSelect(owner int32, r *DataReader) bool {
	return f.actionShiftSelect(owner, f.readSelectionUnits(r, "invalid selection of %d units"))
}

func (f *ActionFunctions) readActionDeselect(owner int32, r *DataReader) bool {
	return f.actionDeselect(owner, f.readSelectionUnits(r, "invalid deselection of %d units"))
}

func readQueueFlag(r *DataReader) bool {
	return r.GetU8() != 0
}

func (f *ActionFunctions) readUnitByID(r *DataReader) *Unit {
	return f.getUnitByID(UnitID{RawValue: r.GetU16()})
}

func (f *ActionFunctions) readOptionalUnitType(r *DataReader) *UnitTypeT {
	id := UnitTypes(r.GetU16())
	if id == UnitTypesNone {
		return nil
	}
	return f.getUnitType(id)
}

func readPos(r *DataReader) Xy {
	x := int32(r.GetI16())
	y := int32(r.GetI16())
	return Xy{X: x, Y: y}
}

func (f *ActionFunctions) readActionTrain(owner int32, r *DataReader) bool {
	unitType := f.getUnitType(UnitTypes(r.GetU16()))
	return f.actionTrain(owner, unitType)
}

func (f *ActionFunctions) readActionDefaultOrder(owner int32, r *DataReader) bool {
	pos := readPos(r)
	target := f.readUnitByID(r)
	targetUnitType := f.readOptionalUnitType(r)
	queue := readQueueFlag(r)
	return f.actionDefaultOrder(owner, pos, target, targetUnitType, queue)
}

func (f *ActionFunctions) readActionPlayerLeave(owner int32, r *DataReader) bool {
	reason := int32(r.GetI8())
	return f.actionPlayerLeave(owner, reason)
}

func (f *ActionFunctions) readActionBuild(owner int32, r *DataReader) bool {
	orderType := f.getOrderType(Orders(r.GetU8()))
	tileX := uint(r.GetU16())
	tileY := uint(r.GetU16())
	unitType := f.getUnitType(UnitTypes(r.GetU16()))
	return f.actionBuild(owner, orderType, unitType, XyT[uint]{X: tileX, Y: tileY})
}

func (f *ActionFunctions) readActionStop(owner int32, r *DataReader) bool {
	return f.actionStop(owner, readQueueFlag(r))
}

func (f *ActionFunctions) readActionOrder(owner int32, r *DataReader) bool {
	pos := readPos(r)
	target := f.readUnitByID(r)
	targetUnitType := f.readOptionalUnitType(r)
	orderType := f.getOrderType(Orders(r.GetU8()))
	queue := readQueueFlag(r)
	return f.actionOrder(owner, orderType, pos, target, targetUnitType, queue)
}

func (f *ActionFunctions) readActionCancelBuildingUnit(owner int32, r *DataReader) bool {
	return f.actionCancelBuildingUnit(owner)
}

func (f *ActionFunctions) readActionCancelBuildQueue(owner int32, r *DataReader) bool {
	slot := uint(r.GetU16())
	return f.actionCancelBuildQueue(owner, slot)
}

func (f *ActionFunctions) readActionControlGroup(owner int32, r *DataReader) bool {
	subaction := int32(r.GetU8())
	groupN := uint(r.GetU8())
	if groupN > 10 {
		return false
	}
	return f.actionControlGroup(owner, groupN, subaction)
}

func (f *ActionFunctions) readActionUnloadAll(owner int32, r *DataReader) bool {
	return f.actionUnloadAll(owner, readQueueFlag(r))
}

func (f *ActionFunctions) readActionUnload(owner int32, r *DataReader) bool {
	return f.actionUnload(owner, f.readUnitByID(r))
}

func (f *ActionFunctions) readActionLiftoff(owner int32, r *DataReader) bool {
	return f.actionLiftoff(owner, readPos(r))
}

func (f *ActionFunctions) readActionResearch(owner int32, r *DataReader) bool {
	tech := f.getTechType(TechTypes(r.GetU8()))
	return f.actionResearch(owner, tech)
}

func (f *ActionFunctions) readActionCancelResearch(owner int32, r *DataReader) bool {
	return f.actionCancelResearch(owner)
}

func (f *ActionFunctions) readActionUpgrade(owner int32, r *DataReader) bool {
	upgrade := f.getUpgradeType(UpgradeTypes(r.GetU8()))
	return f.actionUpgrade(owner, upgrade)
}

func (f *ActionFunctions) readActionCancelUpgrade(owner int32, r *DataReader) bool {
	return f.actionCancelUpgrade(owner)
}

func (f *ActionFunctions) readActionUnsiege(owner int32, r *DataReader) bool {
	return f.actionUnsiege(owner, readQueueFlag(r))
}

func (f *ActionFunctions) readActionSiege(owner int32, r *DataReader) bool {
	return f.actionSiege(owner, readQueueFlag(r))
}

func (f *ActionFunctions) readActionChat(owner int32, r *DataReader) bool {
	r.GetN(81)
	return true
}

func (f *ActionFunctions) readActionReturnCargo(owner int32, r *DataReader) bool {
	return f.actionReturnCargo(owner, readQueueFlag(r))
}

func (f *ActionFunctions) readActionHoldPosition(owner int32, r *DataReader) bool {
	return f.actionHoldPosition(owner, readQueueFlag(r))
}

func (f *ActionFunctions) readActionCloak(owner int32, r *DataReader) bool {
	r.GetU8()
	return f.actionCloak(owner)
}

func (f *ActionFunctions) readActionDecloak(owner int32, r *DataReader) bool {
	r.GetU8()
	return f.actionDecloak(owner)
}

func (f *ActionFunctions) readActionSetAlliances(owner int32, r *DataReader) bool {
	var arr [12]int32
	vals := r.GetU32()
	for i := range arr {
		arr[i] = int32(vals & 3)
		vals >>= 2
	}
	return f.actionSetAlliances(owner, arr)
}

func (f *ActionFunctions) readActionSetSharedVision(owner int32, r *DataReader) bool {
	flags := uint32(r.GetU16())
	return f.actionSetSharedVision(owner, flags)
}

func (f *ActionFunctions) readActionCancelAddon(owner int32, r *DataReader) bool {
	return f.actionCancelAddon(owner)
}

func (f *ActionFunctions) readActionStimPack(owner int32, r *DataReader) bool {
	return f.actionStimPack(owner)
}

func (f *ActionFunctions) readActionCancelNuke(owner int32, r *DataReader) bool {
	return f.actionCancelNuke(owner)
}

func (f *ActionFunctions) readActionPingMinimap(owner int32, r *DataReader) bool {
	r.GetU16()
	r.GetU16()
	return true
}

func (f *ActionFunctions) readActionMorph(owner int32, r *DataReader) bool {
	unitType := f.getUnitType(UnitTypes(r.GetU16()))
	return f.actionMorph(owner, unitType)
}

func (f *ActionFunctions) readActionMorphBuilding(owner int32, r *DataReader) bool {
	unitType := f.getUnitType(UnitTypes(r.GetU16()))
	return f.actionMorphBuilding(owner, unitType)
}

func (f *ActionFunctions) readActionBurrow(owner int32, r *DataReader) bool {
	return f.actionBurrow(owner, readQueueFlag(r))
}

func (f *ActionFunctions) readActionUnburrow(owner int32, r *DataReader) bool {
	r.GetU8()
	return f.actionUnburrow(owner)
}

func (f *ActionFunctions) readActionCancelMorph(owner int32, r *DataReader) bool {
	return f.actionCancelMorph(owner)
}

func (f *ActionFunctions) readActionMorphArchon(owner int32, r *DataReader) bool {
	return f.actionMorphArchon(owner)
}

func (f *ActionFunctions) readActionCarrierStop(owner int32, r *DataReader) bool {
	return f.actionCarrierStop(owner)
}

func (f *ActionFunctions) readActionTrainFighter(owner int32, r *DataReader) bool {
	return f.actionTrainFighter(owner)
}

func (f *ActionFunctions) readActionMorphDarkArchon(owner int32, r *DataReader) bool {
	return f.actionMorphDarkArchon(owner)
}

func (f *ActionFunctions) readActionReaverStop(owner int32, r *DataReader) bool {
	return f.actionReaverStop(owner)
}

func (f *ActionFunctions) readActionCheat(owner int32, r *DataReader) bool {
	flags := int32(r.GetU32())
	return f.actionCheat(owner, flags)
}

func (f *ActionFunctions) readActionExtCheat(owner int32, r *DataReader) bool {
	cheatType := int32(r.GetU8())
	switch cheatType {
	case 0:
		f.readExtCheatUnit(r)
	case 1:
		f.readExtCheatPlayer(r)
	default:
		throwError("unknown ext cheat type %d", cheatType)
	}
	return true
}

func (f *ActionFunctions) readExtCheatUnit(r *DataReader) {
	subtype := int32(r.GetU8())
	target := f.readUnitByID(r)
	switch subtype {
	case 0:
		f.setUnitHp(target, Fp8Integer(r.GetI32()))
	case 1:
		f.setUnitShieldPoints(target, Fp8Integer(r.GetI32()))
	case 2:
		f.setUnitEnergy(target, Fp8Integer(r.GetI32()))
	default:
		throwError("unknown ext cheat unit subtype %d", subtype)
	}
}

func (f *ActionFunctions) readExtCheatPlayer(r *DataReader) {
	subtype := int32(r.GetU8())
	player := int32(r.GetU8())
	switch subtype {
	case 0:
		upgrade := f.getUpgradeType(UpgradeTypes(r.GetU8()))
		level := int32(r.GetI8())
		f.st.UpgradeLevels[player][upgrade.ID] = level
		f.applyUpgradesToPlayerUnits(player)
	case 1:
		tech := f.getTechType(TechTypes(r.GetU8()))
		researched := int32(r.GetI8())
		f.st.TechResearched[player][tech.ID] = researched != 0
	case 2:
		f.st.CurrentMinerals[player] = r.GetI32()
	case 3:
		f.st.CurrentGas[player] = r.GetI32()
	default:
		throwError("unknown ext cheat player subtype %d", subtype)
	}
}

func (f *ActionFunctions) readAction(r *DataReader) bool {
	playerID := int32(r.GetU8())
	owner := int32(-1)
	for i, id := range f.actionSt.PlayerID {
		if id == playerID {
			owner = int32(i)
			break
		}
	}
	if owner == -1 {
		throwError("execute_action: player id %d not found", playerID)
	}
	return f.readActionOwner(owner, r)
}

func (f *ActionFunctions) readActionOwner(owner int32, r *DataReader) bool {
	actionID := int32(r.GetU8())
	f.onAction(owner, actionID)
	switch actionID {
	case 9:
		return f.readActionSelect(owner, r)
	case 10:
		return f.readActionShiftSelect(owner, r)
	case 11:
		return f.readActionDeselect(owner, r)
	case 12:
		return f.readActionBuild(owner, r)
	case 13:
		return f.readActionSetSharedVision(owner, r)
	case 14:
		return f.readActionSetAlliances(owner, r)
	case 18:
		return f.readActionCheat(owner, r)
	case 19:
		return f.readActionControlGroup(owner, r)
	case 20:
		return f.readActionDefaultOrder(owner, r)
	case 21:
		return f.readActionOrder(owner, r)
	case 24:
		return f.readActionCancelBuildingUnit(owner, r)
	case 25:
		return f.readActionCancelMorph(owner, r)
	case 26:
		return f.readActionStop(owner, r)
	case 27:
		return f.readActionCarrierStop(owner, r)
	case 28:
		return f.readActionReaverStop(owner, r)
	case 30:
		return f.readActionReturnCargo(owner, r)
	case 31:
		return f.readActionTrain(owner, r)
	case 32:
		return f.readActionCancelBuildQueue(owner, r)
	case 33:
		return f.readActionCloak(owner, r)
	case 34:
		return f.readActionDecloak(owner, r)
	case 35:
		return f.readActionMorph(owner, r)
	case 37:
		return f.readActionUnsiege(owner, r)
	case 38:
		return f.readActionSiege(owner, r)
	case 39:
		return f.readActionTrainFighter(owner, r)
	case 40:
		return f.readActionUnloadAll(owner, r)
	case 41:
		return f.readActionUnload(owner, r)
	case 42:
		return f.readActionMorphArchon(owner, r)
	case 43:
		return f.readActionHoldPosition(owner, r)
	case 44:
		return f.readActionBurrow(owner, r)
	case 45:
		return f.readActionUnburrow(owner, r)
	case 46:
		return f.readActionCancelNuke(owner, r)
	case 47:
		return f.readActionLiftoff(owner, r)
	case 48:
		return f.readActionResearch(owner, r)
	case 49:
		return f.readActionCancelResearch(owner, r)
	case 50:
		return f.readActionUpgrade(owner, r)
	case 51:
		return f.readActionCancelUpgrade(owner, r)
	case 52:
		return f.readActionCancelAddon(owner, r)
	case 53:
		return f.readActionMorphBuilding(owner, r)
	case 54:
		return f.readActionStimPack(owner, r)
	case 87:
		return f.readActionPlayerLeave(owner, r)
	case 88:
		return f.readActionPingMinimap(owner, r)
	case 90:
		return f.readActionMorphDarkArchon(owner, r)
	case 92:
		return f.readActionChat(owner, r)
	case 210:
		return f.readActionExtCheat(owner, r)
	}
	throwError("execute_action: unknown action %d", actionID)
	return false
}

func (f *ActionFunctions) readActionBytes(data []byte) bool {
	r := NewDataReaderLe(data)
	return f.readAction(&r)
}

func (f *ActionFunctions) readActionOwnerBytes(owner int32, data []byte) bool {
	r := NewDataReaderLe(data)
	return f.readActionOwner(owner, &r)
}

func (f *ActionFunctions) executeActions(actionsData []byte) {
	if f.st.CurrentFrame != f.actionSt.NextActionFrame {
		return
	}
	for f.actionSt.ActionsDataPosition != uint(len(actionsData)) {
		r := NewDataReaderLe(actionsData[f.actionSt.ActionsDataPosition:])
		frame := r.GetI32()
		if frame != f.st.CurrentFrame {
			f.actionSt.NextActionFrame = frame
			return
		}
		actionsSize := uint(r.GetU8())
		r2 := NewDataReaderLe(r.GetN(actionsSize))
		for r2.Left() != 0 {
			f.readAction(&r2)
		}
		f.actionSt.ActionsDataPosition += r.Tell()
	}
}
