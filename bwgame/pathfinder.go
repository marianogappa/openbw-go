package bwgame

import "slices"

type longPathNode struct {
	prev                   *longPathNode
	pos                    XyFp8
	region                 *Region
	totalCost              Fp8
	estimatedRemainingCost Fp8
	estimatedFinalCost     Fp8
	visited                bool
}

func longPathNodeLess(a, b *longPathNode) bool {
	return a.estimatedFinalCost.Less(b.estimatedFinalCost)
}

func regionLongPathNode(r *Region) *longPathNode {
	n, _ := r.PathfinderNode.(*longPathNode)
	return n
}

func clearRegionLongPathNodes(allNodes []*longPathNode) {
	for _, v := range allNodes {
		v.region.PathfinderNode = nil
	}
}

func pfPartitionPoint(n int, isBefore func(i int) bool) int {
	first := 0
	length := n
	for length > 0 {
		half := length / 2
		mid := first + half
		if isBefore(mid) {
			first = mid + 1
			length -= half + 1
		} else {
			length = half
		}
	}
	return first
}

func prependSlice[T any](s []T, v T) []T {
	return append([]T{v}, s...)
}

func (f *StateFunctions) pathfinderFindLongPath(pf *Pathfinder) bool {
	if pf.SourceRegion == pf.DestinationRegion {
		return false
	}

	var open []*longPathNode
	var allNodes []*longPathNode
	var goalNode *longPathNode

	regionPos := func(r *Region) XyFp8 {
		if r == pf.SourceRegion {
			return f.toXyFp8(pf.Source)
		}
		if r == pf.DestinationRegion {
			return f.toXyFp8(pf.Destination)
		}
		return r.Center
	}

	find := func(fromRegion *Region, toRegion *Region) {
		toPos := regionPos(toRegion)

		allNodes = allNodes[:0]
		open = open[:0]
		goalNode = nil

		startNode := &longPathNode{}
		allNodes = append(allNodes, startNode)
		startNode.pos = regionPos(fromRegion)
		startNode.region = fromRegion
		startNode.estimatedRemainingCost = Fp8Integer(128 * 128)
		startNode.estimatedFinalCost = startNode.estimatedRemainingCost
		startNode.region.PathfinderNode = startNode

		open = append(open, startNode)
		binaryHeapUp(open, len(open)-1, longPathNodeLess)

		for len(open) != 0 {
			cur := open[0]
			open[0], open[len(open)-1] = open[len(open)-1], open[0]
			open = open[:len(open)-1]
			binaryHeapDown(open, 0, longPathNodeLess)

			if cur.region == toRegion {
				goalNode = cur
				break
			}
			cur.visited = true

			add := func(r *Region) {
				if len(open) == 125 {
					return
				}
				if len(allNodes) == 350 {
					return
				}
				pos := regionPos(r)
				cost := f.xyLengthFP8(XyFp8Sub(pos, cur.pos))
				if r.GroupIndex != pf.SourceRegion.GroupIndex {
					cost = cost.MulInt(2)
				}
				totalCost := cur.totalCost.Add(cost)
				n := regionLongPathNode(r)
				if n == nil {
					n = &longPathNode{}
					allNodes = append(allNodes, n)
					n.prev = cur
					n.pos = pos
					n.region = r
					n.totalCost = totalCost
					n.estimatedRemainingCost = f.xyLengthFP8(XyFp8Sub(toPos, pos))
					n.estimatedFinalCost = n.totalCost.Add(n.estimatedRemainingCost)
					n.visited = false
					r.PathfinderNode = n
					open = append(open, n)
					binaryHeapUp(open, len(open)-1, longPathNodeLess)
				} else if cur.prev != n {
					if totalCost.Less(n.totalCost) {
						n.prev = cur
						n.totalCost = totalCost
						estimatedFinalCost := n.totalCost.Add(n.estimatedRemainingCost)
						if n.visited {
							n.estimatedFinalCost = estimatedFinalCost
							n.visited = false
							open = append(open, n)
							binaryHeapUp(open, len(open)-1, longPathNodeLess)
						} else {
							if estimatedFinalCost.GreaterEqual(n.estimatedFinalCost) {
								throwError("unreachable; cost did not decrease")
							}
							n.estimatedFinalCost = estimatedFinalCost
							binaryHeapUp(open, slices.Index(open, n), longPathNodeLess)
						}
					}
				}
			}
			for _, n := range cur.region.WalkableNeighbors {
				add(n)
			}
			if cur.region.GroupIndex != pf.SourceRegion.GroupIndex {
				for _, n := range cur.region.NonWalkableNeighbors {
					add(n)
				}
			}
			if uint(len(open)) > pf.LongHighestOpenSize {
				pf.LongHighestOpenSize = uint(len(open))
			}
			if len(open) == 125 {
				break
			}
			if len(allNodes) == 350 {
				break
			}
		}
		if goalNode == nil && len(open) != 0 {
			goalNode = open[0]
		}

		if goalNode == nil {
			startNode.estimatedRemainingCost = f.xyLengthFP8(XyFp8Sub(toPos, startNode.region.Center))
			bestCost := startNode.estimatedRemainingCost
			bestNode := startNode
			for _, n := range allNodes[1:] {
				if n.estimatedRemainingCost.Less(bestCost) {
					bestCost = n.estimatedRemainingCost
					bestNode = n
				}
			}
			goalNode = bestNode
		}
	}

	var pathIsReversed bool
	if pf.SourceRegion.GroupIndex == pf.DestinationRegion.GroupIndex {
		find(pf.SourceRegion, pf.DestinationRegion)
		pathIsReversed = false
	} else {
		find(pf.DestinationRegion, pf.SourceRegion)
		pathIsReversed = true
		if goalNode.region != pf.SourceRegion {
			clearRegionLongPathNodes(allNodes)
			find(pf.SourceRegion, goalNode.region)
			pathIsReversed = false
		}
	}

	pf.LongAllNodesSize = uint(len(allNodes))

	pf.LongPath = pf.LongPath[:0]
	pf.FullLongPathSize = 0
	pf.CurrentLongPathIndex = ^uint(0)
	fullPathSize := uint(0)
	if pathIsReversed {
		goalGroupIndex := goalNode.region.GroupIndex
		for n := goalNode; n != nil; n = n.prev {
			if n.region.GroupIndex != goalGroupIndex {
				break
			}
			if len(pf.LongPath) != 50 {
				pf.LongPath = append(pf.LongPath, n.region)
			}
			fullPathSize++
		}
	} else {
		for n := goalNode; n != nil; n = n.prev {
			fullPathSize++
		}
		for i := fullPathSize; i > 50; i-- {
			goalNode = goalNode.prev
		}
		for n := goalNode; n != nil; n = n.prev {
			pf.LongPath = prependSlice(pf.LongPath, n.region)
		}
	}
	pf.FullLongPathSize = fullPathSize
	clearRegionLongPathNodes(allNodes)
	return len(pf.LongPath) != 0
}

func binaryHeapUp[T any](open []T, element int, compare func(a, b T) bool) {
	index := element
	for index != 0 {
		parentIndex := (index - 1) / 2
		if compare(open[parentIndex], open[index]) {
			break
		}
		open[parentIndex], open[index] = open[index], open[parentIndex]
		index = parentIndex
	}
}

func binaryHeapDown[T any](open []T, element int, compare func(a, b T) bool) {
	index := element
	endIndex := len(open)
	for {
		childIndex := index*2 + 1
		if childIndex >= endIndex {
			break
		}
		if childIndex+1 != endIndex {
			if compare(open[childIndex+1], open[childIndex]) {
				childIndex++
			}
		}
		if !compare(open[childIndex], open[index]) {
			break
		}
		open[index], open[childIndex] = open[childIndex], open[index]
		index = childIndex
	}
}

func (f *StateFunctions) pathfinderUnitCanCollideWith(u *Unit, target *Unit, considerCollisionWithUnit *Unit, considerCollisionWithMovingUnits bool) bool {
	if target != considerCollisionWithUnit && !considerCollisionWithMovingUnits {
		if !f.unitIsAtMoveTarget(&target.Flingy) {
			return false
		}
	}
	if f.unitIs(target.UnitType, UnitTypesSpecialUpperLevelDoor) {
		return false
	}
	if f.unitIs(target.UnitType, UnitTypesSpecialRightUpperLevelDoor) {
		return false
	}
	if f.unitIs(target.UnitType, UnitTypesSpecialPitDoor) {
		return false
	}
	if f.unitIs(target.UnitType, UnitTypesSpecialRightPitDoor) {
		return false
	}
	return f.unitCanCollideWith(u, target)
}

func (f *StateFunctions) pathfinderUnitCanCollideWithPF(pf *Pathfinder, target *Unit) bool {
	return f.pathfinderUnitCanCollideWith(pf.U, target, pf.ConsiderCollisionWithUnit, pf.ConsiderCollisionWithMovingUnits)
}

func (f *StateFunctions) pathfinderAdjustTargetPos(unitInnerBb Rect, target Xy) (bool, Xy) {
	for width := uint(1); width < 10; width += 2 {
		for dir := uint(0); dir != 4; dir++ {
			for n := uint(0); n != width; n++ {
				bb := f.translateRect(unitInnerBb, target)
				if f.isInInnerMapBounds(bb) {
					entirelyWalkable := true
					fromX := uint(uint32(bb.From.X) / 32)
					fromY := uint(uint32(bb.From.Y) / 32)
					toX := uint(uint32(bb.To.X) / 32)
					toY := uint(uint32(bb.To.Y) / 32)
					tileRegionIndex := f.gameSt.Regions.TileRegionIndex
					for y := fromY; entirelyWalkable && y <= toY; y++ {
						for x := fromX; x <= toX; x++ {
							if 256*y+x >= uint(len(tileRegionIndex)) {
								throwError("vector::at: index out of range")
							}
							index := tileRegionIndex[256*y+x]
							if index < 0x2000 {
								region := &f.gameSt.Regions.Regions[index]
								if !region.Walkable() {
									entirelyWalkable = false
									break
								}
							}
						}
					}
					if entirelyWalkable {
						return true, target
					}
				}
				target = XyAdd(target, XyMulScalar(cardinalDirectionXy[dir], 8))
			}
			target = XySub(target, XyMulScalar(cardinalDirectionXy[dir], 8))
		}
		target = XySub(target, Xy{X: 8, Y: 8})
	}
	return false, target
}

func (f *StateFunctions) pathfinderAdjustDestination(sourceRegion *Region, destination Xy) Xy {
	target := XyDivScalar(f.nearestPosInRect(destination, sourceRegion.Area), 32)
	for width := uint(1); width < 16; width += 2 {
		for dir := uint(0); dir != 4; dir++ {
			for n := uint(0); n != width; n++ {
				if uint(target.X) < f.gameSt.MapTileWidth && uint(target.Y) < f.gameSt.MapTileHeight {
					r := f.getRegionAtPreferWalkable(XyMulScalar(target, 32))
					if r == sourceRegion {
						return XyAdd(XyMulScalar(target, 32), Xy{X: 16, Y: 16})
					}
				}
				target = XyAdd(target, cardinalDirectionXy[dir])
			}
			target = XySub(target, cardinalDirectionXy[dir])
		}
		target = XySub(target, Xy{X: 1, Y: 1})
	}
	return f.toXy(sourceRegion.Center)
}

func (f *StateFunctions) pathfinderFindNextShortPath(pf *Pathfinder) bool {
	pf.CurrentLongPathIndex++
	if pf.CurrentLongPathIndex >= uint(len(pf.LongPath)) {
		return false
	}
	sourceRegion := pf.LongPath[pf.CurrentLongPathIndex]
	for sourceRegion != pf.SourceRegion {
		pf.CurrentLongPathIndex++
		if pf.CurrentLongPathIndex >= uint(len(pf.LongPath)) {
			return false
		}
		sourceRegion = pf.LongPath[pf.CurrentLongPathIndex]
	}
	target := pf.Destination
	isNearDestination := true
	if uint(len(pf.LongPath))-pf.CurrentLongPathIndex > 3 {
		target = f.toXy(pf.LongPath[pf.CurrentLongPathIndex+2].Center)
		isNearDestination = false
	}
	var targetRegion *Region
	if pf.CurrentLongPathIndex != uint(len(pf.LongPath))-1 {
		targetRegion = pf.LongPath[pf.CurrentLongPathIndex+1]
	} else {
		if sourceRegion != pf.DestinationRegion {
			pf.Destination = f.pathfinderAdjustDestination(sourceRegion, pf.Destination)
			pf.IsStuck = true
			isNearDestination = true
			targetRegion = nil
		}
	}
	if isNearDestination {
		pf.TargetUnit = pf.U.MoveTarget.Unit
		if pf.TargetUnit == nil {
			unitBB := f.unitBoundingBox(pf.U, pf.Destination)
			add := Xy{X: f.gameSt.MaxUnitWidth/2 + 1, Y: f.gameSt.MaxUnitHeight/2 + 1}
			searchBB := Rect{From: XySub(pf.Destination, add), To: XyAdd(pf.Destination, add)}
			for u := range f.findUnitsNoexpand(searchBB) {
				if f.isIntersecting(unitBB, f.unitSpriteBoundingBox(u)) {
					if f.pathfinderUnitCanCollideWithPF(pf, u) {
						pf.TargetUnit = u
						break
					}
				}
			}
		}
		if pf.TargetUnit != nil {
			if f.unitIsSpecialBeacon(pf.TargetUnit.UnitType) {
				pf.TargetUnit = nil
			}
		}
		if pf.TargetUnit != nil {
			pf.TargetUnitBB = f.unitSpriteInnerBoundingBox(pf.TargetUnit)
			pf.TargetUnitBB.From = XySub(pf.TargetUnitBB.From, XyAdd(pf.U.UnitType.Dimensions.To, Xy{X: 1, Y: 1}))
			pf.TargetUnitBB.To = XyAdd(pf.TargetUnitBB.To, XyAdd(pf.U.UnitType.Dimensions.From, Xy{X: 1, Y: 1}))
			isGoal := false
			if pf.Source.X == pf.TargetUnitBB.From.X || pf.Source.X == pf.TargetUnitBB.To.X {
				if pf.Source.Y >= pf.TargetUnitBB.From.Y && pf.Source.Y <= pf.TargetUnitBB.To.Y {
					isGoal = true
				}
			}
			if pf.Source.Y == pf.TargetUnitBB.From.Y || pf.Source.Y == pf.TargetUnitBB.To.Y {
				if pf.Source.X >= pf.TargetUnitBB.From.X && pf.Source.X <= pf.TargetUnitBB.To.X {
					isGoal = true
				}
			}
			if isGoal {
				pf.ShortPath = []Xy{pf.Source}
				pf.Destination = pf.Source
				pf.CurrentShortPathIndex = 0
				pf.DestinationReached = true
				pf.IsStuck = false
				return true
			}
		} else {
			pf.TargetUnitBB = Rect{From: Xy{X: -32000, Y: -32000}, To: Xy{X: -32000, Y: -32000}}
			adjusted, adjustedTarget := f.pathfinderAdjustTargetPos(pf.UnitBB, target)
			if adjusted && adjustedTarget != target {
				target = adjustedTarget
				pf.Destination = target
			}
		}
	}
	f.pathfinderFindShortPath(pf, target, targetRegion)
	if len(pf.ShortPath) == 0 {
		return false
	}
	lastPathPos := pf.ShortPath[len(pf.ShortPath)-1]
	if pf.DestinationReached {
		if isNearDestination || lastPathPos != target {
			pf.Destination = lastPathPos
		} else {
			pf.DestinationReached = false
		}
	} else {
		if targetRegion == nil {
			pf.Destination = lastPathPos
			pf.DestinationReached = true
			pf.IsStuck = true
		}
	}
	return true
}

func (f *StateFunctions) pathfinderFindLongPathFromTo(pf *Pathfinder, from Xy, to Xy) bool {
	pf.Source = from
	pf.Destination = to
	pf.SourceRegion = f.getRegionAt(pf.Source)
	pf.DestinationRegion = f.getRegionAt(pf.Destination)
	pf.LongAllNodesSize = 0
	pf.LongHighestOpenSize = 0
	pf.DestinationReached = false
	pf.IsStuck = false
	if pf.SourceRegion == pf.DestinationRegion {
		pf.LongPath = []*Region{pf.SourceRegion}
		pf.FullLongPathSize = 1
		pf.CurrentLongPathIndex = ^uint(0)
		return true
	}
	return f.pathfinderFindLongPath(pf)
}

func (f *StateFunctions) pathfinderFind(pf *Pathfinder, shortPathOnly bool) bool {
	pf.SourceRegion = f.getRegionAt(pf.Source)
	pf.DestinationRegion = f.getRegionAt(pf.Destination)
	pf.UnitBB = f.unitTypeInnerBoundingBox(pf.U.UnitType)
	pf.ShortHighestOpenSize = 0
	pf.ShortAllNodesSize = 0
	pf.LongAllNodesSize = 0
	pf.LongHighestOpenSize = 0
	pf.DestinationReached = false
	pf.IsStuck = false
	if shortPathOnly {
		return f.pathfinderFindNextShortPath(pf)
	}
	if pf.SourceRegion == pf.DestinationRegion {
		pf.LongPath = []*Region{pf.SourceRegion}
		pf.FullLongPathSize = 1
		pf.CurrentLongPathIndex = ^uint(0)
		return f.pathfinderFindNextShortPath(pf)
	}
	if !f.pathfinderFindLongPath(pf) {
		return false
	}
	return f.pathfinderFindNextShortPath(pf)
}

func (f *StateFunctions) newPath() *Path {
	if !f.st.FreePaths.Empty() {
		r := f.st.FreePaths.Front()
		f.st.FreePaths.PopFront()
		return r
	}
	if len(f.st.Paths) >= 1024 {
		return nil
	}
	path := &Path{}
	f.st.Paths = append(f.st.Paths, path)
	return path
}

func (f *StateFunctions) freeUnitPath(u *Unit) {
	if u.Path == nil {
		return
	}
	f.freePath(u.Path)
	u.Path = nil
}

func (f *StateFunctions) freePath(path *Path) {
	f.st.FreePaths.PushFront(path)
}

func (f *StateFunctions) pathProgress(u *Unit, to Xy, considerCollisionWithUnit *Unit, considerCollisionWithMovingUnits bool) bool {
	f.uUnsetMovementFlag(&u.Flingy, 0x40)
	f.uSetMovementFlag(&u.Flingy, 0x10)
	pf := &Pathfinder{}
	pf.ConsiderCollisionWithUnit = considerCollisionWithUnit
	pf.ConsiderCollisionWithMovingUnits = considerCollisionWithMovingUnits
	findNewPath := true
	if u.Path != nil {
		u.Path.CurrentShortPathIndex++
		if u.Path.CurrentShortPathIndex < uint(len(u.Path.ShortPath)) {
			findNewPath = false
		} else {
			pf.U = u
			pf.Source = u.Sprite.Position
			pf.Destination = u.Path.Destination
			pf.LongPath, u.Path.LongPath = u.Path.LongPath, pf.LongPath
			pf.FullLongPathSize = u.Path.FullLongPathSize
			pf.CurrentLongPathIndex = u.Path.CurrentLongPathIndex
			pf.ShortPath, u.Path.ShortPath = u.Path.ShortPath, pf.ShortPath
			pf.CurrentShortPathIndex = u.Path.CurrentShortPathIndex
			f.freePath(u.Path)
			u.Path = nil
			if pf.CurrentLongPathIndex+1 < uint(len(pf.LongPath)) {
				if pf.FullLongPathSize == uint(len(pf.LongPath)) || pf.CurrentLongPathIndex+4 < uint(len(pf.LongPath)) {
					if pf.Source != pf.Destination {
						if f.pathfinderFind(pf, true) {
							if len(pf.LongPath) != 0 && len(pf.ShortPath) != 0 {
								findNewPath = false
							}
						}
					}
				}
			}
		}
	}
	if findNewPath {
		pf.U = u
		pf.Source = u.Sprite.Position
		pf.Destination = to
		f.pathfinderFind(pf, false)
		if len(pf.LongPath) == 0 || len(pf.ShortPath) == 0 {
			return false
		}
	}
	path := u.Path
	if path == nil {
		if pf.DestinationReached {
			if to != pf.Destination {
				to = pf.Destination
				moveTarget := to
				moveTargetUnit := u.MoveTarget.Unit
				if moveTargetUnit != nil && f.uMovementFlag(&moveTargetUnit.Flingy, 2) {
					haltDistance := f.unitHaltDistance(&u.Flingy)
					pos := XyAdd(moveTargetUnit.Sprite.Position, f.toXy(f.directionXY(moveTargetUnit.NextVelocityDirection, haltDistance.MulInt(3))))
					if f.isPosInMapBounds(pos) {
						moveTarget = pos
					}
				}
				f.setFlingyMoveTarget(&u.Flingy, moveTarget, nil)
				u.MoveTarget.Unit = moveTargetUnit
				f.uSetStatusFlagValue(u, UnitStatusFlagImmovable, pf.IsStuck)
			}
		}

		if (uint(len(pf.LongPath))+uint(len(pf.ShortPath))*2+2-1)/2*2 > 48 {
			if len(pf.ShortPath) >= 24 {
				pf.ShortPath = resizeSlice(pf.ShortPath, 23)
				pf.LongPath = resizeSlice(pf.LongPath, 1)
			} else {
				pf.LongPath = resizeSlice(pf.LongPath, (24-len(pf.ShortPath))*2)
			}
		}
		path = f.newPath()
		if path == nil {
			return false
		}
		path.Delay = 0
		path.CreationFrame = f.st.CurrentFrame
		path.StateFlags = 0
		path.LongPath, pf.LongPath = pf.LongPath, path.LongPath
		path.FullLongPathSize = pf.FullLongPathSize
		path.CurrentLongPathIndex = min(pf.CurrentLongPathIndex, uint(len(path.LongPath)))
		path.ShortPath, pf.ShortPath = pf.ShortPath, path.ShortPath
		path.CurrentShortPathIndex = 0

		path.Source = u.Sprite.Position
		path.Destination = to

		u.Path = path
	}
	if path.CurrentShortPathIndex >= uint(len(path.ShortPath)) {
		throwError("circular_vector subscript out of range")
	}
	path.Next = path.ShortPath[path.CurrentShortPathIndex]
	if path.Next == u.MoveTarget.Pos {
		f.uUnsetMovementFlag(&u.Flingy, 0x10)
	} else if f.xyLength(XySub(path.Destination, path.Next)) < 16 {
		f.uSetMovementFlag(&u.Flingy, 0x40)
	}
	return true
}

func (f *StateFunctions) createSingleStepPath(source Xy, destination Xy) *Path {
	path := f.newPath()
	if path == nil {
		return nil
	}
	path.Delay = 0
	path.StateFlags = 0

	if len(path.ShortPath) <= 1 {
		path.LongPath = resizeSlice(path.LongPath, 1)
	} else {
		index := uint(path.ShortPath[1].X)
		if index < uint(len(f.gameSt.Regions.Regions)) {
			path.LongPath = []*Region{&f.gameSt.Regions.Regions[index]}
		} else {
			path.LongPath = []*Region{nil}
		}
	}

	path.FullLongPathSize = 1
	path.CurrentLongPathIndex = 0
	path.ShortPath = []Xy{destination}
	path.CurrentShortPathIndex = 0

	path.Source = source
	path.Destination = destination
	path.Next = destination

	return path
}

func (f *StateFunctions) unitPathTo(u *Unit, to Xy, considerCollisionWithUnit *Unit, considerCollisionWithMovingUnits bool) bool {
	if f.isMovingAlongPath(u) {
		return true
	}
	if !f.uGroundUnit(u) {
		f.setUnitMoveTarget(u, to)
		f.freeUnitPath(u)
		u.Path = f.createSingleStepPath(u.Sprite.Position, to)
		return u.Path != nil
	}
	if !f.pathProgress(u, to, considerCollisionWithUnit, considerCollisionWithMovingUnits) {
		return false
	}
	f.setNextTargetWaypoint(&u.Flingy, u.Path.Next)
	if u.NextMovementWaypoint != u.Path.Next {
		u.NextMovementWaypoint = u.Path.Next
		f.uSetMovementFlag(&u.Flingy, 1)
	}
	return true
}

func (f *StateFunctions) unitUpdatePathMovementState(u *Unit, allowNewPath bool) bool {
	if f.unitIsAtMoveTarget(&u.Flingy) || f.uMovementFlag(&u.Flingy, 4) {
		u.MovementState = MovementStatesUMAtMoveTarget
		return true
	}
	if f.uCollision(u) && f.uGroundUnit(u) {
		if u.Path != nil {
			f.freePath(u.Path)
			u.Path = nil
		}
		u.MovementState = MovementStatesUMCheckIllegal
		return true
	}
	nextMovementState := int32(MovementStatesUMAnotherPath)
	if !allowNewPath || u.Path != nil {
		if u.Path == nil || ^u.Path.StateFlags&1 != 0 {
			return false
		}
		nextMovementState = MovementStatesUMNewMoveTarget
	}
	if u.MovementState >= MovementStatesUMFixCollision && u.MovementState <= MovementStatesUMTerrainSlide {
		nextMovementState = MovementStatesUMTurnAndStart
	}
	u.MovementState = nextMovementState
	return true
}

type pfNeighbor struct {
	pos    Xy
	flags  int32
	isGoal bool
}

type pfSearch struct {
	u            *Unit
	targetUnit   *Unit
	inner        [4]int32
	outer        [4]int32
	targetUnitBB Rect
	target       Xy

	hasFoundGoal bool
	curPos       Xy
	curPosMax    Xy
	curPosMin    Xy

	localEdges [4][]Contour

	nearestEdge [4]*Contour

	neighbors StaticVector[pfNeighbor, [32]pfNeighbor]

	visitedAreas []Rect
}

func (w *pfSearch) innerFrom(c Contour) int32 {
	return c.V[1] + w.inner[c.Flags&3]
}

func (w *pfSearch) innerTo(c Contour) int32 {
	return c.V[2] + w.inner[(c.Flags>>2)&3]
}

func (w *pfSearch) expandInner(c Contour) Contour {
	c.V[0] += w.inner[c.Dir]
	c.V[1] += w.inner[c.Flags&3]
	c.V[2] += w.inner[(c.Flags>>2)&3]
	return c
}

func (w *pfSearch) expandOuter(c Contour) Contour {
	c.V[0] += w.outer[c.Dir]
	c.V[1] += w.outer[c.Flags&3]
	c.V[2] += w.outer[(c.Flags>>2)&3]
	return c
}

type shortPathNode struct {
	prev                   *shortPathNode
	pos                    Xy
	totalCost              Fp8
	estimatedRemainingCost Fp8
	estimatedFinalCost     Fp8
	isGoal                 bool
	depth                  int32
	directionalFlags       int32
	region                 *Region
	isTargetRegion         bool
	isNeighborRegion       bool
	visited                bool
}

func shortPathNodeLess(a, b *shortPathNode) bool {
	return a.estimatedFinalCost.Less(b.estimatedFinalCost)
}

type pfVisitedRange struct {
	first  int32
	second int32
}

type pfVisitedColumn struct {
	x int32
	y []pfVisitedRange
}

func firstRangeEndingAtOrAfter(ranges []pfVisitedRange, y int32) int {
	i := 0
	for ; i != len(ranges); i++ {
		if y <= ranges[i].second {
			break
		}
	}
	return i
}

func (f *StateFunctions) pathfinderFindShortPath(pf *Pathfinder, target Xy, targetRegion *Region) {
	targetIsDestination := target == pf.Destination
	targetRegionWalkable := targetRegion != nil && targetRegion.Walkable()

	sourceRegion := pf.SourceRegion
	destinationRegion := f.getRegionAt(pf.Destination)

	moveToRegion := destinationRegion
	if targetRegion != nil {
		moveToRegion = targetRegion
	}

	for _, nr := range moveToRegion.WalkableNeighbors {
		if nr == sourceRegion {
			continue
		}
		nr.PathfinderFlag = 1
	}

	w := &pfSearch{}

	w.u = pf.U
	w.targetUnit = pf.TargetUnit
	w.target = target
	dimensions := pf.U.UnitType.Dimensions
	w.inner[0] = dimensions.From.Y
	w.outer[0] = dimensions.From.Y + 1
	w.inner[1] = -dimensions.To.X
	w.outer[1] = -dimensions.To.X - 1
	w.inner[2] = -dimensions.To.Y
	w.outer[2] = -dimensions.To.Y - 1
	w.inner[3] = dimensions.From.X
	w.outer[3] = dimensions.From.X + 1
	w.targetUnitBB = pf.TargetUnitBB

	var open StaticVector[*shortPathNode, [150]*shortPathNode]

	allNodes := &StaticVector[shortPathNode, [250]shortPathNode]{}
	allNodes.PushBack(shortPathNode{})
	startNode := allNodes.Ptr(allNodes.Size() - 1)
	startNode.pos = pf.Source
	startNode.estimatedRemainingCost = Fp8Integer(128 * 128)
	startNode.estimatedFinalCost = startNode.estimatedRemainingCost
	startNode.directionalFlags = 0xff
	startNode.region = sourceRegion

	open.PushBack(startNode)
	binaryHeapUp(open.Values(), open.Size()-1, shortPathNodeLess)

	areaVisited := []pfVisitedColumn{{x: 0}, {x: int32(f.gameSt.MapWidth)}}

	columnAfter := func(x int32) int {
		return pfPartitionPoint(len(areaVisited), func(i int) bool { return !(x < areaVisited[i].x) })
	}

	pfRemoveVisitedFlags := func(pos Xy, flags int32) int32 {
		cur := columnAfter(pos.X) - 1

		curY := areaVisited[cur].y
		i := firstRangeEndingAtOrAfter(curY, pos.Y)
		if i == len(curY) {
			return flags
		}
		mask := int32(0)
		if pos.Y <= curY[i].first {
			mask |= 0x19
		}
		if pos.Y >= curY[i].second {
			mask |= 0x46
		}
		if pos.X == areaVisited[cur].x {
			mask |= 0x8c
			if cur != 0 {
				prevY := areaVisited[cur-1].y
				i = firstRangeEndingAtOrAfter(prevY, pos.Y)
				if i != len(prevY) {
					if pos.Y >= prevY[i].first && pos.Y <= prevY[i].second {
						mask &= 0x7f
						if pos.Y > prevY[i].first {
							mask &= 0xf7
						}
						if pos.Y < prevY[i].second {
							mask &= 0xfb
						}
					}
				}
			}
		}
		next := cur + 1
		if pos.X == areaVisited[next].x-1 {
			mask |= 0x23
			nextY := areaVisited[next].y
			i = firstRangeEndingAtOrAfter(nextY, pos.Y)
			if i != len(nextY) {
				if pos.Y >= nextY[i].first && pos.Y <= nextY[i].second {
					mask &= 0xdf
					if pos.Y > nextY[i].first {
						mask &= 0xfe
					}
					if pos.Y < nextY[i].second {
						mask &= 0xfd
					}
				}
			}
		}
		return flags & mask
	}

	mergeLocalEdge := func(i *Contour, c Contour) bool {
		if w.innerFrom(c) <= w.innerTo(*i) {
			if w.innerTo(c) >= w.innerFrom(*i)-1 {
				if c.V[1] < i.V[1] {
					i.V[1] = c.V[1]
				}
				if c.V[2] > i.V[2] {
					i.V[2] = c.V[2]
				}
				return true
			}
		}
		return false
	}

	pfAddLocalEdge := func(n int, c Contour) {
		localEdges := w.localEdges[n]
		i := contourLowerBound(localEdges, c.V[0])
		for i != len(localEdges) && localEdges[i].V[0] == c.V[0] {
			if mergeLocalEdge(&localEdges[i], c) {
				return
			}
			if w.innerFrom(c) <= w.innerTo(localEdges[i]) {
				break
			}
			i++
		}
		w.localEdges[n] = slices.Insert(localEdges, i, c)
	}

	pfPushBackLocalEdge := func(n int, c Contour) {
		localEdges := w.localEdges[n]
		if len(localEdges) != 0 && localEdges[len(localEdges)-1].V[0] == c.V[0] {
			last := len(localEdges) - 1
			if mergeLocalEdge(&localEdges[last], c) {
				return
			}
			if w.innerFrom(c) <= w.innerTo(localEdges[last]) {
				w.localEdges[n] = slices.Insert(localEdges, last, c)
			} else {
				w.localEdges[n] = append(localEdges, c)
			}
		} else {
			w.localEdges[n] = append(localEdges, c)
		}
	}

	pfAddLocalTerrain := func() {
		c0 := f.gameSt.Regions.Contours[0]
		for i := contourLowerBound(c0, w.curPosMin.Y-w.inner[0]-1); i != len(c0); i++ {
			if c0[i].V[0] > w.curPos.Y-w.inner[0]-1 {
				break
			}
			if w.innerFrom(c0[i]) <= w.curPosMax.X {
				if w.innerTo(c0[i]) >= w.curPosMin.X {
					pfPushBackLocalEdge(0, c0[i])
				}
			}
		}
		c1 := f.gameSt.Regions.Contours[1]
		for i := contourUpperBound(c1, w.curPos.X-w.inner[1]); i != len(c1); i++ {
			if c1[i].V[0] > w.curPosMax.X-w.inner[1]+1 {
				break
			}
			if w.innerFrom(c1[i]) <= w.curPosMax.Y {
				if w.innerTo(c1[i]) >= w.curPosMin.Y {
					pfPushBackLocalEdge(1, c1[i])
				}
			}
		}
		c2 := f.gameSt.Regions.Contours[2]
		for i := contourUpperBound(c2, w.curPos.Y-w.inner[2]); i != len(c2); i++ {
			if c2[i].V[0] > w.curPosMax.Y-w.inner[2]+1 {
				break
			}
			if w.innerFrom(c2[i]) <= w.curPosMax.X {
				if w.innerTo(c2[i]) >= w.curPosMin.X {
					pfPushBackLocalEdge(2, c2[i])
				}
			}
		}
		c3 := f.gameSt.Regions.Contours[3]
		for i := contourLowerBound(c3, w.curPosMin.X-w.inner[3]-1); i != len(c3); i++ {
			if c3[i].V[0] > w.curPos.X-w.inner[3]-1 {
				break
			}
			if w.innerFrom(c3[i]) <= w.curPosMax.Y {
				if w.innerTo(c3[i]) >= w.curPosMin.Y {
					pfPushBackLocalEdge(3, c3[i])
				}
			}
		}
	}

	addLocalUnitEdge := func(n int, u *Unit, c Contour) {
		if u == w.targetUnit || f.pathfinderUnitCanCollideWithPF(pf, u) {
			pfAddLocalEdge(n, c)
		}
	}

	pfAddLocalUnits := func() {
		finderY := f.st.UnitFinderY
		finderX := f.st.UnitFinderX
		for i := unitFinderLowerBound(finderY, w.curPosMin.Y-w.inner[0]-1); i != len(finderY); i++ {
			bb := finderY[i].U.UnitFinderBoundingBox
			if finderY[i].Value >= w.curPos.Y-w.inner[0] {
				break
			}
			if finderY[i].Value == bb.To.Y {
				c := Contour{V: [3]int32{bb.To.Y, bb.From.X, bb.To.X}, Dir: 0, Flags: 0x3d}
				if c.V[1]+w.inner[1] <= w.curPosMax.X && c.V[2]+w.inner[3] >= w.curPosMin.X {
					addLocalUnitEdge(0, finderY[i].U, c)
				}
			}
		}
		for i := unitFinderLowerBound(finderX, w.curPos.X-w.inner[1]); i != len(finderX); i++ {
			bb := finderX[i].U.UnitFinderBoundingBox
			if finderX[i].Value > w.curPosMax.X-w.inner[1]+1 {
				break
			}
			if finderX[i].Value == bb.From.X {
				c := Contour{V: [3]int32{bb.From.X, bb.From.Y, bb.To.Y}, Dir: 1, Flags: 0x32}
				if c.V[1]+w.inner[2] <= w.curPosMax.Y && c.V[2]+w.inner[0] >= w.curPosMin.Y {
					addLocalUnitEdge(1, finderX[i].U, c)
				}
			}
		}
		for i := unitFinderLowerBound(finderY, w.curPos.Y-w.inner[2]); i != len(finderY); i++ {
			bb := finderY[i].U.UnitFinderBoundingBox
			if finderY[i].Value > w.curPosMax.Y-w.inner[2]+1 {
				break
			}
			if finderY[i].Value == bb.From.Y {
				c := Contour{V: [3]int32{bb.From.Y, bb.From.X, bb.To.X}, Dir: 2, Flags: 0x3d}
				if c.V[1]+w.inner[1] <= w.curPosMax.X && c.V[2]+w.inner[3] >= w.curPosMin.X {
					addLocalUnitEdge(2, finderY[i].U, c)
				}
			}
		}
		for i := unitFinderLowerBound(finderX, w.curPosMin.X-w.inner[3]-1); i != len(finderX); i++ {
			bb := finderX[i].U.UnitFinderBoundingBox
			if finderX[i].Value >= w.curPos.X-w.inner[3] {
				break
			}
			if finderX[i].Value == bb.To.X {
				c := Contour{V: [3]int32{bb.To.X, bb.From.Y, bb.To.Y}, Dir: 3, Flags: 0x32}
				if c.V[1]+w.inner[2] <= w.curPosMax.Y && c.V[2]+w.inner[0] >= w.curPosMin.Y {
					addLocalUnitEdge(3, finderX[i].U, c)
				}
			}
		}
	}

	edgeOverlaps := func(c Contour, v0, v1 int32) bool {
		return max(w.innerFrom(c), v0) <= min(w.innerTo(c), v1)
	}

	pfLocalEdgesFind := func(dir int, v, v0, v1, v2 int32) int {
		edges := w.localEdges[dir]
		switch dir {
		case 0, 3:
			for i := contourUpperBound(edges, v-w.inner[dir]); i != 0; {
				i--
				if edges[i].V[0]+w.inner[dir] < v2 {
					break
				}
				if edgeOverlaps(edges[i], v0, v1) {
					return i
				}
			}
		case 1, 2:
			for i := contourLowerBound(edges, v-w.inner[dir]); i != len(edges); i++ {
				if edges[i].V[0]+w.inner[dir] > v2 {
					break
				}
				if edgeOverlaps(edges[i], v0, v1) {
					return i
				}
			}
		}
		return -1
	}

	pfAddNeighbor := func(pos Xy, flags int32) {
		if !f.isPosInMapBounds(pos) {
			return
		}
		if pos == w.curPos {
			return
		}
		isGoal := false
		if w.targetUnit != nil {
			if pos.X == w.targetUnitBB.From.X || pos.X == w.targetUnitBB.To.X {
				if pos.Y >= w.targetUnitBB.From.Y && pos.Y <= w.targetUnitBB.To.Y {
					isGoal = true
				}
			}
			if pos.Y == w.targetUnitBB.From.Y || pos.Y == w.targetUnitBB.To.Y {
				if pos.X >= w.targetUnitBB.From.X && pos.X <= w.targetUnitBB.To.X {
					isGoal = true
				}
			}
		}
		if pos == target {
			isGoal = true
		}
		if isGoal {
			flags = 0xff
			w.hasFoundGoal = true
		}
		if w.neighbors.Size() < 32 {
			if isGoal || w.neighbors.Empty() {
				w.neighbors.PushBack(pfNeighbor{pos: pos, flags: flags, isGoal: isGoal})
			} else {
				i := slices.IndexFunc(w.neighbors.Values(), func(v pfNeighbor) bool { return v.pos == pos })
				if i != -1 {
					w.neighbors.Ptr(i).flags &= flags
				} else {
					w.neighbors.PushBack(pfNeighbor{pos: pos, flags: flags, isGoal: isGoal})
				}
			}
		}
	}

	visitArea := func(fromX, fromY, toX, toY int32) {
		if toX < fromX {
			fromX, toX = toX, fromX
		}
		if toY < fromY {
			fromY, toY = toY, fromY
		}
		if w.target.X >= fromX && w.target.X <= toX {
			if w.target.Y >= fromY && w.target.Y <= toY {
				w.neighbors.PushBack(pfNeighbor{pos: w.target, flags: 0xff, isGoal: true})
				w.hasFoundGoal = true
			}
		}
		w.visitedAreas = append(w.visitedAreas, Rect{From: Xy{X: fromX, Y: fromY}, To: Xy{X: toX, Y: toY}})
	}

	addNeighbors := func(n int) bool {
		var c Contour
		pos := w.curPos
		posMax := w.curPosMax
		posMin := w.curPosMin
		at := func(x, y int32) Xy { return Xy{X: x, Y: y} }
		if n == 0 {
			if pos.X < posMax.X {
				xval := posMax.X
				yval := posMin.Y
				edgeY := posMin.Y
				if w.nearestEdge[1] != nil {
					edgeY = w.innerFrom(*w.nearestEdge[1]) - 1
					if edgeY > yval {
						yval = edgeY
					}
				}
				flags := int32(0x3b)
				edges := w.localEdges[0]
				i := pfLocalEdgesFind(0, pos.Y-1, pos.X, xval, yval-1)
				if i != -1 {
					c = w.expandOuter(edges[i])
					yval = c.V[0]
					if xval == c.V[2] {
						if c.Flags&0x20 != 0 {
							flags = 0x33
						} else {
							flags = 0
						}
					} else if xval < c.V[2] {
						flags = 0x22
					}
				}
				if w.nearestEdge[1] != nil {
					flags &= 0xfd
					if edgeY < yval {
						flags &= 0xde
					}
				}
				if yval >= pos.Y {
					flags |= 0x44
				} else {
					pfAddNeighbor(at(xval, pos.Y), 0xe7)
				}
				pfAddNeighbor(at(xval, yval), flags)
				if yval == posMin.Y {
					pfAddNeighbor(at(pos.X, yval), 0xdd)
				}
				if i != -1 {
					if xval > c.V[2] {
						pfAddNeighbor(at(c.V[2], c.V[0]), 0x11)
					}
					for {
						if c.V[1] < pos.X {
							posMax.X = pos.X
							break
						}
						posMax.X = c.V[1]
						pfAddNeighbor(at(c.V[1], c.V[0]), 0x18)
						if i == 0 {
							break
						}
						i--
						c = w.expandOuter(edges[i])
						if c.V[0] != yval {
							break
						}
						if c.V[2] < pos.X {
							break
						}
						if c.V[2] <= xval {
							pfAddNeighbor(at(c.V[2], c.V[0]), 0x11)
						}
					}
				}
				visitArea(pos.X, pos.Y, xval, yval)
				pos.Y = yval
			}
			if pos.Y > posMin.Y {
				xval := posMax.X
				yval := posMin.Y
				edgeX := posMin.Y
				if w.nearestEdge[0] != nil {
					edgeX = w.innerTo(*w.nearestEdge[0]) + 1
					if edgeX < xval {
						xval = edgeX
					}
				}
				flags := int32(0x3b)
				edges := w.localEdges[1]
				i := pfLocalEdgesFind(1, pos.X+1, yval, pos.Y, xval+1)
				if i != -1 {
					c = w.expandOuter(edges[i])
					xval = c.V[0]
					if yval == c.V[1] {
						if c.Flags&0x10 != 0 {
							flags = 0x39
						} else {
							flags = 0
						}
					} else if yval > c.V[1] {
						flags = 0x18
					}
				}
				if w.nearestEdge[0] != nil {
					flags &= 0xf7
					if edgeX > xval {
						flags &= 0xde
					}
				}
				if xval <= pos.X {
					flags |= 0x84
				} else {
					pfAddNeighbor(at(pos.X, yval), 0xdd)
				}
				pfAddNeighbor(at(xval, yval), flags)
				if xval < posMax.X {
					pfAddNeighbor(at(xval, pos.Y), 1)
				}
				if i != -1 {
					if yval < c.V[1] {
						pfAddNeighbor(at(c.V[0], c.V[1]), 0x21)
					}
					for c.V[2] <= pos.Y {
						pfAddNeighbor(at(c.V[0], c.V[2]), 0x22)
						if i == len(edges)-1 {
							break
						}
						i++
						c = w.expandOuter(edges[i])
						if c.V[0] != xval || c.V[1] > pos.Y {
							pfAddNeighbor(at(xval, pos.Y), 0x21)
							break
						}
						if c.V[1] >= yval {
							pfAddNeighbor(at(c.V[0], c.V[1]), 0x21)
						}
					}
				}
				visitArea(pos.X, pos.Y, xval, yval)
			}
			return false
		} else if n == 1 {
			if pos.X < posMax.X {
				xval := posMax.X
				yval := posMax.Y
				edgeY := posMax.Y
				if w.nearestEdge[1] != nil {
					edgeY = w.innerTo(*w.nearestEdge[1]) + 1
					if edgeY < yval {
						yval = edgeY
					}
				}
				flags := int32(0x67)
				edges := w.localEdges[2]
				i := pfLocalEdgesFind(2, pos.Y+1, pos.X, xval, yval+1)
				if i != -1 {
					c = w.expandOuter(edges[i])
					if pos.X < c.V[1] {
						pfAddNeighbor(at(c.V[1], c.V[0]), 0x44)
					}
					yval = c.V[0]
					posMax.X = c.V[1]
					for c.V[2] < xval {
						pfAddNeighbor(at(c.V[2], c.V[0]), 0x42)
						if i == len(edges)-1 {
							break
						}
						i++
						ic := w.expandInner(edges[i])
						if ic.V[0]-1 != yval {
							break
						}
						if ic.V[1]-1 >= xval {
							break
						}
						c = w.expandOuter(edges[i])
						if c.V[1] <= xval {
							pfAddNeighbor(at(c.V[1], c.V[0]), 0x44)
						}
					}
					if xval == c.V[2] {
						if c.Flags&0x20 != 0 {
							flags = 0x63
						} else {
							flags = 0
						}
					} else if xval < c.V[2] {
						flags = 0x21
					}
				}
				if w.nearestEdge[1] != nil {
					flags &= 0xfe
					if edgeY > yval {
						flags &= 0xdd
					}
				}
				if yval <= pos.Y {
					flags |= 0x18
				} else {
					pfAddNeighbor(at(xval, pos.Y), 0xbb)
				}
				pfAddNeighbor(at(xval, yval), flags)
				if yval == posMax.Y {
					pfAddNeighbor(at(pos.X, yval), 0xde)
				}
				visitArea(pos.X, pos.Y, xval, yval)
				pos.Y = yval
			}
			if pos.Y < posMax.Y {
				xval := posMax.X
				yval := posMax.Y
				edgeX := posMax.X
				if w.nearestEdge[2] != nil {
					edgeX = w.innerTo(*w.nearestEdge[2]) + 1
					if edgeX < xval {
						xval = edgeX
					}
				}
				flags := int32(0x67)
				edges := w.localEdges[1]
				i := pfLocalEdgesFind(1, pos.X+1, pos.Y, yval, xval+1)
				if i != -1 {
					c = w.expandOuter(edges[i])
					if c.V[0] < posMax.X {
						pfAddNeighbor(at(c.V[0], pos.Y), 2)
					}
					if c.V[1] > pos.Y {
						pfAddNeighbor(at(c.V[0], c.V[1]), 0x21)
					}
					xval = c.V[0]
					for c.V[2] < yval {
						pfAddNeighbor(at(c.V[0], c.V[2]), 0x22)
						if i == len(edges)-1 {
							break
						}
						i++
						ic := w.expandInner(edges[i])
						if ic.V[0]-1 != xval {
							break
						}
						if ic.V[1]-1 >= yval {
							break
						}
						c = w.expandOuter(edges[i])
						if pos.Y <= c.V[1] && c.V[1] <= posMax.Y {
							pfAddNeighbor(at(c.V[0], c.V[1]), 0x21)
						}
					}
					if yval == c.V[2] {
						if c.Flags&0x20 != 0 {
							flags = 0x66
						} else {
							flags = 0
						}
					} else if yval < c.V[2] {
						flags = 0x44
					}
				}
				if w.nearestEdge[2] != nil {
					flags &= 0xfb
					if edgeX > xval {
						flags &= 0xbd
					}
				}
				if xval <= pos.X {
					flags |= 0x88
				} else {
					pfAddNeighbor(at(pos.X, yval), 0xde)
				}
				pfAddNeighbor(at(xval, yval), flags)
				visitArea(pos.X, pos.Y, xval, yval)
			}
			return false
		} else if n == 2 {
			if pos.X > posMin.X {
				xval := posMin.X
				yval := posMax.Y
				edgeY := posMax.Y
				if w.nearestEdge[3] != nil {
					edgeY = w.innerTo(*w.nearestEdge[3]) + 1
					if edgeY < yval {
						yval = edgeY
					}
				}
				flags := int32(0xce)
				edges := w.localEdges[2]
				i := pfLocalEdgesFind(2, pos.Y+1, xval, pos.X, yval+1)
				if i != -1 {
					c = w.expandOuter(edges[i])
					yval = c.V[0]
					if xval == c.V[1] {
						if c.Flags&0x10 != 0 {
							flags = 0xcc
						} else {
							flags = 0
						}
					} else if xval > c.V[1] {
						flags = 0x88
					}
				}
				if w.nearestEdge[3] != nil {
					flags &= 0xf7
					if edgeY > yval {
						flags &= 0x7b
					}
				}
				if yval <= pos.Y {
					flags |= 0x11
				} else {
					pfAddNeighbor(at(xval, pos.Y), 0xbd)
				}
				pfAddNeighbor(at(xval, yval), flags)
				if yval == posMax.Y {
					pfAddNeighbor(at(pos.X, yval), 0x77)
				}
				if i != -1 {
					if xval < c.V[1] {
						pfAddNeighbor(at(c.V[1], c.V[0]), 0x44)
					}
					for {
						if c.V[2] > pos.X {
							pfAddNeighbor(at(pos.X, c.V[0]), 0x21)
							posMin.X = pos.X
							break
						}
						posMin.X = c.V[2]
						pfAddNeighbor(at(c.V[2], c.V[0]), 0x42)
						if i == len(edges)-1 {
							break
						}
						i++
						c = w.expandOuter(edges[i])
						if c.V[0] != yval {
							break
						}
						if c.V[1] > pos.X {
							break
						}
						pfAddNeighbor(at(c.V[1], c.V[0]), 0x44)
					}
				}
				visitArea(pos.X, pos.Y, xval, yval)
				pos.Y = yval
			}
			if pos.Y < posMax.Y {
				xval := posMin.X
				yval := posMax.Y
				edgeX := posMin.Y
				if w.nearestEdge[2] != nil {
					edgeX = w.innerFrom(*w.nearestEdge[2]) - 1
					if edgeX > xval {
						xval = edgeX
					}
				}
				flags := int32(0xce)
				edges := w.localEdges[3]
				i := pfLocalEdgesFind(3, pos.X-1, pos.Y, yval, xval-1)
				if i != -1 {
					c = w.expandOuter(edges[i])
					xval = c.V[0]
					if yval == c.V[2] {
						if c.Flags&0x20 != 0 {
							flags = 0xc6
						} else {
							flags = 0
						}
					} else if yval < c.V[2] {
						flags = 0x42
					}
				}
				if w.nearestEdge[2] != nil {
					flags &= 0xfd
					if edgeX < xval {
						flags &= 0xbb
					}
				}
				if xval >= pos.X {
					flags |= 0x22
				} else {
					pfAddNeighbor(at(pos.X, yval), 0x77)
				}
				pfAddNeighbor(at(xval, yval), flags)
				if xval > posMin.X {
					pfAddNeighbor(at(xval, pos.Y), 4)
				}
				if i != -1 {
					if yval > c.V[2] {
						pfAddNeighbor(at(c.V[0], c.V[2]), 0x84)
					}
					for c.V[1] >= pos.Y {
						pfAddNeighbor(at(c.V[0], c.V[1]), 0x88)
						if i == 0 {
							break
						}
						i--
						c = w.expandOuter(edges[i])
						if c.V[0] != xval || c.V[2] < pos.Y {
							pfAddNeighbor(at(xval, pos.Y), 0x84)
							break
						}
						if c.V[2] <= yval {
							pfAddNeighbor(at(c.V[0], c.V[2]), 0x84)
						}
					}
				}
				visitArea(pos.X, pos.Y, xval, yval)
			}
			return false
		}
		if pos.X > posMin.X {
			xval := posMin.X
			yval := posMin.Y
			edgeY := posMin.Y
			if w.nearestEdge[3] != nil {
				edgeY = w.innerFrom(*w.nearestEdge[3]) - 1
				if edgeY > yval {
					yval = edgeY
				}
			}
			flags := int32(0x9d)
			edges := w.localEdges[0]
			i := pfLocalEdgesFind(0, pos.Y-1, xval, pos.X, yval-1)
			if i != -1 {
				c = w.expandOuter(edges[i])
				if pos.X <= c.V[2] {
					pfAddNeighbor(at(pos.X, c.V[0]), 0x22)
				} else {
					pfAddNeighbor(at(c.V[2], c.V[0]), 0x11)
				}
				yval = c.V[0]
				posMin.X = c.V[2]
				for c.V[1] > xval {
					pfAddNeighbor(at(c.V[1], c.V[0]), 0x18)
					if i == 0 {
						break
					}
					i--
					ic := w.expandInner(edges[i])
					if ic.V[0]+1 != yval {
						break
					}
					if ic.V[2]+1 <= xval {
						break
					}
					c = w.expandOuter(edges[i])
					if c.V[2] >= xval {
						pfAddNeighbor(at(c.V[2], c.V[0]), 0x11)
					}
				}
				if xval == c.V[1] {
					if c.Flags&0x10 != 0 {
						flags = 0x9c
					} else {
						flags = 0
					}
				} else if xval > c.V[1] {
					flags = 0x84
				}
			}
			if w.nearestEdge[3] != nil {
				flags &= 0xfb
				if edgeY < yval {
					flags &= 0x77
				}
			}
			if yval >= pos.Y {
				flags |= 0x44
			} else {
				pfAddNeighbor(at(xval, pos.Y), 0xee)
			}
			pfAddNeighbor(at(xval, yval), flags)
			if yval == posMin.Y {
				pfAddNeighbor(at(pos.X, yval), 0x7b)
			}
			visitArea(pos.X, pos.Y, xval, yval)
			pos.Y = yval
		}
		if pos.Y > posMin.Y {
			xval := posMin.X
			yval := posMin.Y
			edgeX := posMin.X
			if w.nearestEdge[0] != nil {
				edgeX = w.innerFrom(*w.nearestEdge[0]) - 1
				if edgeX > xval {
					xval = edgeX
				}
			}
			flags := int32(0x9d)
			edges := w.localEdges[3]
			i := pfLocalEdgesFind(3, pos.X-1, yval, pos.Y, xval-1)
			if i != -1 {
				c = w.expandOuter(edges[i])
				if c.V[0] > posMin.X {
					pfAddNeighbor(at(c.V[0], pos.Y), 8)
				}
				if c.V[2] < pos.Y {
					pfAddNeighbor(at(c.V[0], c.V[2]), 0x84)
				}
				xval = c.V[0]
				for c.V[1] > yval {
					pfAddNeighbor(at(c.V[0], c.V[1]), 0x88)
					if i == 0 {
						break
					}
					i--
					ic := w.expandInner(edges[i])
					if ic.V[0]+1 != xval {
						break
					}
					if ic.V[2]+1 <= yval {
						break
					}
					c = w.expandOuter(edges[i])
					if c.V[2] >= yval {
						pfAddNeighbor(at(c.V[0], c.V[2]), 0x84)
					}
				}
				if yval == c.V[1] {
					if c.Flags&0x10 != 0 {
						flags = 0x99
					} else {
						flags = 0
					}
				} else if yval > c.V[1] {
					flags = 0x11
				}
			}
			if w.nearestEdge[0] != nil {
				flags &= 0xfe
				if edgeX < xval {
					flags &= 0xe7
				}
			}
			if xval >= pos.X {
				flags |= 0x21
			} else {
				pfAddNeighbor(at(pos.X, yval), 0x7b)
			}
			pfAddNeighbor(at(xval, yval), flags)
			visitArea(pos.X, pos.Y, xval, yval)
		}
		return false
	}

	mergeVisitedRanges := func(ranges []pfVisitedRange) []pfVisitedRange {
		t := 0
		for n := t + 1; n != len(ranges); n = t + 1 {
			if ranges[n].first-1 <= ranges[t].second {
				if ranges[n].second > ranges[t].second {
					ranges[t].second = ranges[n].second
				}
				ranges = slices.Delete(ranges, n, n+1)
			} else {
				t = n
			}
		}
		return ranges
	}

	pfMarkVisited := func(area Rect) {
		next := columnAfter(area.From.X)
		cur := next - 1
		if area.From.X > areaVisited[cur].x {
			if len(areaVisited) == 128+1 {
				return
			}
			column := pfVisitedColumn{x: area.From.X, y: slices.Clone(areaVisited[cur].y)}
			areaVisited = slices.Insert(areaVisited, next, column)
			cur = next
			next = cur + 1
		}
		if area.From.X > area.To.X {
			return
		}
		for {
			if areaVisited[next].x > area.To.X+1 {
				if len(areaVisited) == 128+1 {
					return
				}
				column := pfVisitedColumn{x: area.To.X + 1, y: slices.Clone(areaVisited[cur].y)}
				areaVisited = slices.Insert(areaVisited, next, column)
				cur = next - 1
			}
			ranges := areaVisited[cur].y
			i := 0
			for ; i != len(ranges); i++ {
				if area.From.Y <= ranges[i].second+1 {
					break
				}
			}
			newRange := pfVisitedRange{first: area.From.Y, second: area.To.Y}
			if i == len(ranges) {
				if len(ranges) == 10 {
					return
				}
				areaVisited[cur].y = slices.Insert(ranges, i, newRange)
			} else if area.To.Y < ranges[i].first-1 || area.From.Y > ranges[i].second+1 {
				if len(ranges) == 10 {
					return
				}
				areaVisited[cur].y = slices.Insert(ranges, i, newRange)
			} else {
				if ranges[i].first > area.From.Y {
					ranges[i].first = area.From.Y
				}
				if ranges[i].second < area.To.Y {
					ranges[i].second = area.To.Y
				}
				areaVisited[cur].y = mergeVisitedRanges(ranges)
			}
			cur++
			next++
			if !(areaVisited[cur].x <= area.To.X) {
				break
			}
		}
	}

	pfGenerateNeighbors := func(pos Xy, flags int32) {
		w.hasFoundGoal = false
		w.neighbors.Clear()
		w.visitedAreas = w.visitedAreas[:0]

		w.curPos = pos
		w.curPosMax = XyAdd(pos, Xy{X: 64, Y: 64})
		w.curPosMin = XySub(pos, Xy{X: 64, Y: 64})
		if pos.X < 64 {
			w.curPosMin.X = 0
		}
		if pos.Y < 64 {
			w.curPosMin.Y = 0
		}
		if uint(pos.X)+64 >= f.gameSt.MapWidth {
			w.curPosMax.X = int32(f.gameSt.MapWidth) - 1
		}
		if uint(pos.Y)+64 >= f.gameSt.MapHeight {
			w.curPosMax.Y = int32(f.gameSt.MapHeight) - 1
		}
		if flags&(0x10|(8|1)) == 0 {
			w.curPosMin.Y = pos.Y
		}
		if flags&(0x20|(1|2)) == 0 {
			w.curPosMax.X = pos.X
		}
		if flags&(0x40|(2|4)) == 0 {
			w.curPosMax.Y = pos.Y
		}
		if flags&(0x80|(4|8)) == 0 {
			w.curPosMin.X = pos.X
		}

		for i := range w.localEdges {
			w.localEdges[i] = w.localEdges[i][:0]
		}

		pfAddLocalTerrain()
		pfAddLocalUnits()

		nearestEdge := func(dir int, v, v0, v1, v2 int32) *Contour {
			i := pfLocalEdgesFind(dir, v, v0, v1, v2)
			if i == -1 {
				return nil
			}
			return &w.localEdges[dir][i]
		}

		w.nearestEdge[0] = nearestEdge(0, pos.Y, pos.X, pos.X, w.curPosMin.Y-1)
		if w.nearestEdge[0] != nil {
			w.curPosMin.Y = w.nearestEdge[0].V[0] + w.inner[0] + 1
		}
		w.nearestEdge[1] = nearestEdge(1, pos.X, pos.Y, pos.Y, w.curPosMax.X+1)
		if w.nearestEdge[1] != nil {
			w.curPosMax.X = w.nearestEdge[1].V[0] + w.inner[1] - 1
		}
		w.nearestEdge[2] = nearestEdge(2, pos.Y, pos.X, pos.X, w.curPosMax.Y+1)
		if w.nearestEdge[2] != nil {
			w.curPosMax.Y = w.nearestEdge[2].V[0] + w.inner[2] - 1
		}
		w.nearestEdge[3] = nearestEdge(3, pos.X, pos.Y, pos.Y, w.curPosMin.X-1)
		if w.nearestEdge[3] != nil {
			w.curPosMin.X = w.nearestEdge[3].V[0] + w.inner[3] + 1
		}

		if w.curPosMin.Y == pos.Y && w.curPosMax.X == pos.X {
			flags &^= 1
		}
		if w.curPosMax.Y == pos.Y {
			if w.curPosMax.X == pos.X {
				flags &^= 2
			}
			if w.curPosMin.X == pos.X {
				flags &^= 4
			}
		}
		if w.curPosMin.Y == pos.Y {
			if w.curPosMin.X == pos.X {
				flags &^= 8
			}
			if w.curPosMin.Y == pos.Y {
				flags &^= 0x10
			}
		}
		if w.curPosMax.X == pos.X {
			flags &^= 0x20
		}
		if w.curPosMax.Y == pos.Y {
			flags &^= 0x40
		}
		if w.curPosMin.X == pos.X {
			flags &^= 0x80
		}

		edgeFlags := func(edge *Contour, withEdge, withoutEdge int32) int32 {
			if edge != nil {
				return withEdge
			}
			return withoutEdge
		}
		if flags&0x10 != 0 {
			flags |= 9
			pfAddNeighbor(Xy{X: pos.X, Y: w.curPosMin.Y}, edgeFlags(w.nearestEdge[0], 0xa6, 0xbf))
		}
		if flags&0x20 != 0 {
			flags |= 3
			pfAddNeighbor(Xy{X: w.curPosMax.X, Y: pos.Y}, edgeFlags(w.nearestEdge[1], 0x5c, 0x7f))
		}
		if flags&0x40 != 0 {
			flags |= 6
			pfAddNeighbor(Xy{X: pos.X, Y: w.curPosMax.Y}, edgeFlags(w.nearestEdge[2], 0xa9, 0xef))
		}
		if flags&0x80 != 0 {
			flags |= 0xc
			pfAddNeighbor(Xy{X: w.curPosMin.X, Y: pos.Y}, edgeFlags(w.nearestEdge[3], 0x53, 0xdf))
		}

		for dir := 0; dir != 4; dir++ {
			if flags&(1<<dir) != 0 {
				if addNeighbors(dir) {
					break
				}
			}
		}

		for _, v := range w.visitedAreas {
			pfMarkVisited(v)
		}
	}

	hasFoundGoal := false
	isTired := false
	nOpenNodesInNeighborRegion := int32(0)
	nOpenNodesInTargetRegion := int32(0)

	pushOpen := func(n *shortPathNode) {
		open.PushBack(n)
		binaryHeapUp(open.Values(), open.Size()-1, shortPathNodeLess)
		if n.isTargetRegion {
			nOpenNodesInTargetRegion++
		}
		if n.isNeighborRegion {
			nOpenNodesInNeighborRegion++
		}
	}

	var goalNode *shortPathNode
	for !open.Empty() {
		openValues := open.Values()
		cur := openValues[0]
		openValues[0], openValues[len(openValues)-1] = openValues[len(openValues)-1], openValues[0]
		open.PopBack()
		binaryHeapDown(open.Values(), 0, shortPathNodeLess)
		if cur.isGoal {
			goalNode = cur
			pf.DestinationReached = true
			break
		}

		if cur.directionalFlags != 0 {
			cur.directionalFlags = pfRemoveVisitedFlags(cur.pos, cur.directionalFlags)
		}

		var isExhausted bool
		if nOpenNodesInTargetRegion <= 0 || hasFoundGoal || (targetIsDestination && allNodes.Size() < 150) {
			isExhausted = isTired
		} else {
			isExhausted = true
			isTired = true
		}
		if nOpenNodesInNeighborRegion != 0 && !hasFoundGoal {
			if allNodes.Size() >= 200 {
				isExhausted = true
				isTired = true
			}
		}
		if cur.isTargetRegion && isExhausted {
			goalNode = cur
			break
		}
		cur.visited = true
		if cur.isTargetRegion {
			nOpenNodesInTargetRegion--
		}
		if cur.isNeighborRegion {
			nOpenNodesInNeighborRegion--
			if isExhausted {
				break
			}
		}
		if hasFoundGoal || isExhausted || cur.directionalFlags == 0 {
			continue
		}
		pfGenerateNeighbors(cur.pos, cur.directionalFlags)

		foundGoal := false
		for vi := 0; vi != w.neighbors.Size(); vi++ {
			v := w.neighbors.Ptr(vi)
			if v.isGoal {
				foundGoal = true
			} else {
				if v.flags != 0 {
					v.flags = pfRemoveVisitedFlags(v.pos, v.flags)
					if v.flags == 0 {
						continue
					}
				}
			}
			nRegion := f.getRegionAt(v.pos)
			cost := f.xyLengthFP8(XyFp8Sub(f.toXyFp8(v.pos), f.toXyFp8(cur.pos)))
			if targetRegionWalkable && nRegion != targetRegion && nRegion != sourceRegion {
				cost = cost.MulInt(2)
			}
			totalCost := cur.totalCost.Add(cost)
			var n *shortPathNode
			for i := 1; i != allNodes.Size(); i++ {
				if allNodes.Ptr(i).pos == v.pos {
					n = allNodes.Ptr(i)
					break
				}
			}
			if n == nil {
				allNodes.PushBack(shortPathNode{})
				n = allNodes.Ptr(allNodes.Size() - 1)
				n.prev = cur
				n.pos = v.pos
				n.region = nRegion
				n.depth = cur.depth + 1
				n.directionalFlags = v.flags
				n.totalCost = totalCost
				n.estimatedRemainingCost = f.xyLengthFP8(XyFp8Sub(f.toXyFp8(target), f.toXyFp8(n.pos)))
				n.estimatedFinalCost = n.totalCost.Add(n.estimatedRemainingCost)
				n.visited = n.directionalFlags == 0 && !n.isGoal
				n.isTargetRegion = n.region == targetRegion
				n.isNeighborRegion = n.region.PathfinderFlag != 0
				n.isGoal = v.isGoal
				if !n.visited {
					pushOpen(n)
				}
				if open.Size() == 150 {
					break
				}
				if allNodes.Size() == 250 {
					break
				}
			} else if cur.prev != n {
				if totalCost.Less(n.totalCost) {
					n.prev = cur
					n.depth = cur.depth + 1
					n.totalCost = totalCost
					estimatedFinalCost := n.totalCost.Add(n.estimatedRemainingCost)
					if n.visited {
						n.directionalFlags = v.flags
						n.estimatedFinalCost = estimatedFinalCost
						n.visited = false
						n.isGoal = v.isGoal
						pushOpen(n)
					} else {
						n.estimatedFinalCost = estimatedFinalCost
						binaryHeapUp(open.Values(), slices.Index(open.Values(), n), shortPathNodeLess)
					}
					if open.Size() == 150 {
						break
					}
				} else if !n.visited {
					n.directionalFlags &= v.flags
				}
			}
		}

		if uint(open.Size()) > pf.ShortHighestOpenSize {
			pf.ShortHighestOpenSize = uint(open.Size())
		}

		hasFoundGoal = foundGoal
		if open.Size() == 150 || allNodes.Size() == 250 {
			if !foundGoal && nOpenNodesInTargetRegion == 0 {
				break
			}
		}
	}

	if goalNode == nil {
		nUnvisitedNodes := int32(0)
		nUnvisitedDestinationRegionNodes := int32(0)

		for i := 1; i != allNodes.Size(); i++ {
			n := allNodes.Ptr(i)
			if n.region.PathfinderFlag != 0 {
				n.region.PathfinderFlag++
			}
			if !n.visited {
				if n.directionalFlags != 0 {
					n.directionalFlags = pfRemoveVisitedFlags(n.pos, n.directionalFlags)
				}
				if n.directionalFlags != 0 {
					nUnvisitedNodes++
					if n.region == destinationRegion {
						nUnvisitedDestinationRegionNodes++
					}
				} else {
					n.visited = true
					if n.isTargetRegion {
						nOpenNodesInTargetRegion--
					}
				}
			}
		}

		if targetIsDestination {
			nUnvisitedNodes = nUnvisitedDestinationRegionNodes
			for _, nr := range moveToRegion.WalkableNeighbors {
				if nr == sourceRegion {
					continue
				}
				if nr.PathfinderFlag < 2 {
					nUnvisitedNodes++
					break
				}
				nUnvisitedNodes -= nr.PathfinderFlag / 2
				if nUnvisitedNodes < 0 {
					nUnvisitedNodes = 0
				}
			}
		}
		goalNode = startNode
		startNode.estimatedRemainingCost = f.xyLengthFP8(f.toXyFp8(XySub(target, startNode.pos)))
		if nUnvisitedNodes == 0 && nOpenNodesInTargetRegion == 0 {
			bestCost := Fp8Integer(128 * 128)
			if nOpenNodesInTargetRegion == 0 {
				bestCost = startNode.estimatedRemainingCost
			}
			bestNode := startNode
			for i := 1; i != allNodes.Size(); i++ {
				n := allNodes.Ptr(i)
				if n.estimatedRemainingCost.Less(bestCost) {
					if targetIsDestination || n.region == sourceRegion || (nOpenNodesInTargetRegion != 0 && n.region == destinationRegion) {
						bestCost = n.estimatedRemainingCost
						bestNode = n
					}
				}
			}
			goalNode = bestNode
			pf.DestinationReached = true
			pf.IsStuck = true
		} else {
			bestCost := Fp8Integer(128 * 128)
			bestNode := startNode
			for i := 1; i != allNodes.Size(); i++ {
				n := allNodes.Ptr(i)
				cost := n.estimatedRemainingCost
				if n.region == destinationRegion || n.region == targetRegion {
					cost = cost.Add(n.totalCost.DivInt(2))
				} else {
					if n.region.PathfinderFlag != 0 {
						cost = cost.MulInt(3).DivInt(2)
					} else {
						if n.region == sourceRegion {
							cost = cost.MulInt(2)
						} else {
							cost = cost.MulInt(4)
						}
					}
				}
				if n.visited {
					cost = cost.MulInt(32)
				}
				if cost.Less(bestCost) {
					bestCost = cost
					bestNode = n
				}
			}
			goalNode = bestNode
		}
	}

	pf.ShortAllNodesSize = uint(allNodes.Size())

	pf.ShortPath = pf.ShortPath[:0]
	pf.CurrentShortPathIndex = 0
	for n := goalNode; n != startNode && len(pf.ShortPath) < 128; n = n.prev {
		pf.ShortPath = prependSlice(pf.ShortPath, n.pos)
	}
	if len(pf.ShortPath) > 50 {
		pf.ShortPath[49] = pf.ShortPath[len(pf.ShortPath)-1]
		pf.ShortPath = resizeSlice(pf.ShortPath, 50)
	}
	for _, nr := range moveToRegion.WalkableNeighbors {
		if nr == sourceRegion {
			continue
		}
		nr.PathfinderFlag = 0
	}
}
