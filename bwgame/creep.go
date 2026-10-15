package bwgame

const creepSpreadMaxDistanceSquared int32 = 320 * 320 * 25

func creepDistanceSquared(dx, dy int32) int32 {
	return dx*dx*25 + dy*dy*64
}

func (f *StateFunctions) tileHasCreep(pos Xy) bool {
	return f.st.Tiles[f.tileIndex(pos)].Flags&TileFlagHasCreep != 0
}

func (f *StateFunctions) creepIsBeingSustained(tilePos XyT[uint]) bool {
	pos := Xy{X: int32(tilePos.X)*32 + 16, Y: int32(tilePos.Y)*32 + 16}
	var area Rect
	area.From = XySub(pos, Xy{X: 640, Y: 400})
	area.To = XyAdd(pos, Xy{X: 640, Y: 400})
	mapWidth := int32(f.gameSt.MapWidth)
	mapHeight := int32(f.gameSt.MapHeight)
	if area.From.X < 0 {
		area.From.X = 0
	} else if area.To.X >= mapWidth {
		area.To.X = mapWidth - 1
	}
	if area.From.Y < 0 {
		area.From.Y = 0
	} else if area.To.Y >= mapHeight {
		area.To.Y = mapHeight - 1
	}
	for u := range f.findUnitsNoexpand(area) {
		if !f.uGroundedBuilding(u) {
			continue
		}
		if !f.utBuilding(u.UnitType) {
			continue
		}
		if f.unitRace(u.UnitType) != RaceZerg {
			continue
		}
		bb := f.getMaxCreepBb(u.UnitType, u.Sprite.Position, f.uCompleted(u))
		if tilePos.X < bb.From.X {
			continue
		}
		if tilePos.Y < bb.From.Y {
			continue
		}
		if tilePos.X > bb.To.X {
			continue
		}
		if tilePos.Y > bb.To.Y {
			continue
		}
		diff := XySub(pos, u.Sprite.Position)
		if creepDistanceSquared(diff.X, diff.Y) > creepSpreadMaxDistanceSquared {
			continue
		}
		return true
	}
	return false
}

func (f *StateFunctions) setCreepReceding(tilePos XyT[uint]) bool {
	life := &f.st.CreepLife
	if life.FreeList.Empty() {
		return false
	}
	v := life.FreeList.Front()
	life.FreeList.PopFront()
	life.FreeListSize--
	nNeighbors := f.countNeighboringCreepTiles(tilePos)
	v.NNeighboringCreepTiles = nNeighbors
	v.TilePos = tilePos
	life.Lists[nNeighbors].PushFront(v)
	life.ListsSize[nNeighbors]++
	life.Table.Insert(v)

	index := tilePos.Y*f.gameSt.MapTileWidth + tilePos.X
	f.st.Tiles[index].Flags |= TileFlagCreepReceding
	return true
}

func (f *StateFunctions) removeCreepProvider(unitType *UnitTypeT, pos Xy, isCompleted bool) bool {
	var flagsMask uint16 = TileFlagCreepReceding
	if !f.unitIs(unitType, UnitTypesZergExtractor) {
		flagsMask |= TileFlagOccupied
	}
	spreadsCreep := f.unitTypeSpreadsCreep(unitType, isCompleted)
	area := f.getMaxCreepBb(unitType, pos, isCompleted)
	dy := int32(area.From.Y)*32 - pos.Y + 16
	for tileY := area.From.Y; tileY != area.To.Y+1; tileY, dy = tileY+1, dy+32 {
		dx := int32(area.From.X)*32 - pos.X + 16
		for tileX := area.From.X; tileX != area.To.X+1; tileX, dx = tileX+1, dx+32 {
			index := tileY*f.gameSt.MapTileWidth + tileX
			if f.st.Tiles[index].Flags&TileFlagHasCreep == 0 {
				continue
			}
			if f.st.Tiles[index].Flags&flagsMask != 0 {
				continue
			}
			if spreadsCreep && creepDistanceSquared(dx, dy) > creepSpreadMaxDistanceSquared {
				continue
			}
			tilePos := XyT[uint]{X: tileX, Y: tileY}
			if !f.creepIsBeingSustained(tilePos) {
				if !f.setCreepReceding(tilePos) {
					return false
				}
			}
		}
	}
	f.st.CreepLife.CheckDeadUnitTimer = 0
	return true
}

func (f *StateFunctions) removeUnitCreepProvider(u *Unit) bool {
	return f.removeCreepProvider(u.UnitType, u.Position, f.uCompleted(u))
}

func (f *StateFunctions) removeExtractorCreep(unitType *UnitTypeT, pos Xy, isCompleted bool) bool {
	spreadsCreep := f.unitTypeSpreadsCreep(unitType, isCompleted)
	area := f.getMaxCreepBb(unitType, pos, isCompleted)
	dy := int32(area.From.Y)*32 - pos.Y + 16
	for tileY := area.From.Y; tileY != area.To.Y+1; tileY, dy = tileY+1, dy+32 {
		dx := int32(area.From.X)*32 - pos.X + 16
		for tileX := area.From.X; tileX != area.To.X+1; tileX, dx = tileX+1, dx+32 {
			index := tileY*f.gameSt.MapTileWidth + tileX
			if f.st.Tiles[index].Flags&TileFlagHasCreep == 0 {
				continue
			}
			if spreadsCreep && creepDistanceSquared(dx, dy) > creepSpreadMaxDistanceSquared {
				continue
			}
			f.setTileCreep(XyT[uint]{X: tileX, Y: tileY}, false)
		}
	}
	f.st.CreepLife.CheckDeadUnitTimer = 0
	return true
}

func (f *StateFunctions) addCreepProvider(u *Unit) {
	life := &f.st.CreepLife
	pos := u.Position
	spreadsCreep := f.unitTypeSpreadsCreep(u.UnitType, f.uCompleted(u))
	area := f.getMaxCreepBb(u.UnitType, pos, f.uCompleted(u))
	dy := int32(area.From.Y)*32 - pos.Y + 16
	for tileY := area.From.Y; tileY != area.To.Y+1; tileY, dy = tileY+1, dy+32 {
		dx := int32(area.From.X)*32 - pos.X + 16
		for tileX := area.From.X; tileX != area.To.X+1; tileX, dx = tileX+1, dx+32 {
			index := tileY*f.gameSt.MapTileWidth + tileX
			if f.st.Tiles[index].Flags&TileFlagCreepReceding == 0 {
				continue
			}
			if spreadsCreep && creepDistanceSquared(dx, dy) > creepSpreadMaxDistanceSquared {
				continue
			}
			v := life.Table.Find(XyT[uint]{X: tileX, Y: tileY})
			if v == nil {
				continue
			}
			life.Table.Remove(v)
			life.Lists[v.NNeighboringCreepTiles].Remove(v)
			life.ListsSize[v.NNeighboringCreepTiles]--
			v.NNeighboringCreepTiles = 9
			life.FreeList.PushFront(v)
			life.FreeListSize++

			f.st.Tiles[index].Flags &^= TileFlagCreepReceding
		}
	}
}

func (f *StateFunctions) zergBuildingStartConstruction(u *Unit) {
	f.addCreepProvider(u)
	f.setConstructionGraphic(u, true)
	u.Hp = u.UnitType.Hitpoints.DivInt(10)
	if !u.BuildQueue.Empty() {
		f.buildQueueClearSingle(u)
	}
	f.setUnitOrder(u, f.getOrderType(OrdersIncompleteMorphing))
}

func creepUnitTileArea(unitType *UnitTypeT, pos Xy) RectT[XyT[uint]] {
	var area RectT[XyT[uint]]
	area.From.X = uint(uint32(pos.X)/32 - uint32(unitType.PlacementSize.X)/32/2)
	area.From.Y = uint(uint32(pos.Y)/32 - uint32(unitType.PlacementSize.Y)/32/2)
	area.To.X = area.From.X + uint(uint32(unitType.PlacementSize.X)/32)
	area.To.Y = area.From.Y + uint(uint32(unitType.PlacementSize.Y)/32)
	return area
}

func (f *StateFunctions) setCreepBuildingTiles(unitType *UnitTypeT, pos Xy) {
	if !f.unitTypeSpreadsCreep(unitType, true) && f.utRequiresCreep(unitType) {
		return
	}

	f.setUnitTilesUnoccupied(unitType, pos)

	creepArea := creepUnitTileArea(unitType, pos)
	roundedCorners := false
	if f.utFlag(unitType, UnitTypeTFlag80000000) {
		creepArea.From.X -= 2
		creepArea.From.Y -= 1
		creepArea.To.X += 2
		creepArea.To.Y += 1
		roundedCorners = true
	}
	for y := creepArea.From.Y; y != creepArea.To.Y; y++ {
		if y >= f.gameSt.MapTileHeight {
			continue
		}
		for x := creepArea.From.X; x != creepArea.To.X; x++ {
			if x >= f.gameSt.MapTileWidth {
				continue
			}
			if roundedCorners {
				onCornerColumn := x == creepArea.From.X || x == creepArea.To.X-1
				onCornerRow := y == creepArea.From.Y || y == creepArea.To.Y-1
				if onCornerColumn && onCornerRow {
					continue
				}
			}
			if !f.tileCanHaveCreep(XyT[uint]{X: x, Y: y}) {
				continue
			}
			if f.st.Tiles[y*f.gameSt.MapTileWidth+x].Flags&TileFlagOccupied != 0 {
				continue
			}
			f.setTileCreep(XyT[uint]{X: x, Y: y}, true)
		}
	}

	f.setUnitTilesOccupied(unitType, pos)
}

var creepRecedeTimerByFreeListSize = [9]int32{1, 3, 5, 6, 7, 8, 9, 9, 9}

func (f *StateFunctions) recedeCreep() {
	life := &f.st.CreepLife
	if life.RecedeTimer != 0 {
		life.RecedeTimer--
		return
	}
	lutIndex := life.FreeListSize >> 7
	if lutIndex >= uint(len(creepRecedeTimerByFreeListSize)) {
		throwError("array::at: __n (which is %d) >= _Nm (which is 9)", lutIndex)
	}
	life.RecedeTimer = creepRecedeTimerByFreeListSize[lutIndex]

	for i := range life.Lists {
		list := &life.Lists[i]
		if list.Empty() {
			continue
		}
		v := list.Begin()
		for steps := f.lcgRandRange(27, 0, int32(life.ListsSize[i])-1); steps > 0; steps-- {
			v = list.Next(v)
		}
		list.Remove(v)
		life.ListsSize[i]--
		life.Table.Remove(v)
		v.NNeighboringCreepTiles = 9
		life.FreeList.PushFront(v)
		life.FreeListSize++

		index := v.TilePos.Y*f.gameSt.MapTileWidth + v.TilePos.X
		f.st.Tiles[index].Flags &^= TileFlagCreepReceding
		f.setTileCreep(v.TilePos, false)
		break
	}
}

func (f *StateFunctions) tileCanHaveCreep(tilePos XyT[uint]) bool {
	index := tilePos.Y*f.gameSt.MapTileWidth + tilePos.X
	if f.st.Tiles[index].Flags&(TileFlagUnbuildable|TileFlagPartiallyWalkable) != 0 {
		return false
	}
	if tilePos.Y == f.gameSt.MapTileHeight-1 {
		return true
	}
	if f.st.Tiles[index+f.gameSt.MapTileWidth].Flags&TileFlagUnbuildable != 0 {
		return false
	}
	return true
}

func (f *StateFunctions) getMaxCreepBb(unitType *UnitTypeT, pos Xy, unitIsCompleted bool) RectT[XyT[uint]] {
	var r Rect
	if f.unitTypeSpreadsCreep(unitType, unitIsCompleted) {
		r.From = XySub(pos, Xy{X: 320, Y: 200})
		r.To = XyAdd(pos, Xy{X: 320, Y: 200})
	} else {
		halfPlacementTiles := XyDivScalar(XyDivScalar(unitType.PlacementSize, 32), 2)
		r.From = XyMulScalar(XySub(XyDivScalar(pos, 32), halfPlacementTiles), 32)
		placementTiles := XyDivScalar(unitType.PlacementSize, 32)
		r.To = XyMulScalar(XySub(XyAdd(XyDivScalar(r.From, 32), placementTiles), Xy{X: 1, Y: 1}), 32)
	}
	var rt RectT[XyT[uint]]
	if r.From.X <= 0 {
		rt.From.X = 0
	} else {
		rt.From.X = uint(uint32(r.From.X) / 32)
	}
	if r.From.Y <= 0 {
		rt.From.Y = 0
	} else {
		rt.From.Y = uint(uint32(r.From.Y) / 32)
	}
	if r.To.X >= int32(f.gameSt.MapWidth) {
		rt.To.X = f.gameSt.MapTileWidth - 1
	} else {
		rt.To.X = uint(uint32(r.To.X) / 32)
	}
	if r.To.Y >= int32(f.gameSt.MapHeight) {
		rt.To.Y = f.gameSt.MapTileHeight - 1
	} else {
		rt.To.Y = uint(uint32(r.To.Y) / 32)
	}
	return rt
}

func (f *StateFunctions) countNeighboringCreepTiles(tilePos XyT[uint]) uint {
	r := uint(0)
	width := f.gameSt.MapTileWidth
	index := tilePos.Y*width + tilePos.X
	index -= width
	index -= 1
	test := func() {
		if f.st.Tiles[index].Flags&TileFlagHasCreep != 0 {
			r++
		}
	}
	if tilePos.Y != 0 {
		if tilePos.X != 0 {
			test()
		}
		index++
		test()
		index++
		if tilePos.X != width-1 {
			test()
		}
		index -= 2
	}
	index += width
	if tilePos.X != 0 {
		test()
	}
	index += 2
	if tilePos.X != width-1 {
		test()
	}
	index -= 2
	index += width
	if tilePos.Y != f.gameSt.MapTileHeight-1 {
		if tilePos.X != 0 {
			test()
		}
		index++
		test()
		index++
		if tilePos.X != width-1 {
			test()
		}
	}
	return r
}

func (f *StateFunctions) setTileCreep(tilePos XyT[uint], hasCreep bool) {
	life := &f.st.CreepLife
	index := tilePos.Y*f.gameSt.MapTileWidth + tilePos.X
	if hasCreep {
		f.st.Tiles[index].Flags |= TileFlagHasCreep
	} else {
		f.st.Tiles[index].Flags &^= TileFlagHasCreep
	}

	width := f.gameSt.MapTileWidth
	height := f.gameSt.MapTileHeight
	index -= width
	tilePos.X--
	index -= 1
	tilePos.Y--
	test := func() {
		if f.st.Tiles[index].Flags&TileFlagHasCreep == 0 {
			return
		}
		if f.st.Tiles[index].Flags&TileFlagCreepReceding == 0 {
			return
		}
		v := life.Table.Find(tilePos)
		if v == nil {
			throwError("set_tile_creep: receding creep not found")
		}
		nNeighbors := f.countNeighboringCreepTiles(tilePos)
		if v.NNeighboringCreepTiles == nNeighbors {
			return
		}

		life.Lists[v.NNeighboringCreepTiles].Remove(v)
		life.ListsSize[v.NNeighboringCreepTiles]--
		v.NNeighboringCreepTiles = nNeighbors
		life.Lists[nNeighbors].PushFront(v)
		life.ListsSize[nNeighbors]++
	}
	if tilePos.Y < height {
		if tilePos.X < width {
			test()
		}
		index++
		tilePos.X++
		test()
		index++
		tilePos.X++
		if tilePos.X < width {
			test()
		}
		index -= 2
		tilePos.X -= 2
	}
	index += width
	tilePos.Y++
	if tilePos.X < width {
		test()
	}
	index += 2
	tilePos.X += 2
	if tilePos.X < width {
		test()
	}
	index -= 2
	tilePos.X -= 2
	index += width
	tilePos.Y++
	if tilePos.Y < height {
		if tilePos.X < width {
			test()
		}
		index++
		tilePos.X++
		test()
		index++
		tilePos.X++
		if tilePos.X < width {
			test()
		}
	}
}

func (f *StateFunctions) spreadCreep(unitType *UnitTypeT, pos Xy, outAnyTilesOccupied *bool) bool {
	var targetTiles [8]StaticVector[uint, [240]uint]
	spreadsCreep := f.unitTypeSpreadsCreep(unitType, true)
	area := f.getMaxCreepBb(unitType, pos, true)
	dy := int32(area.From.Y)*32 - pos.Y + 16
	anyTilesOccupied := false
	for tileY := area.From.Y; tileY != area.To.Y+1; tileY, dy = tileY+1, dy+32 {
		dx := int32(area.From.X)*32 - pos.X + 16
		for tileX := area.From.X; tileX != area.To.X+1; tileX, dx = tileX+1, dx+32 {
			index := tileY*f.gameSt.MapTileWidth + tileX
			flags := f.st.Tiles[index].Flags
			if flags&TileFlagHasCreep != 0 {
				continue
			}
			tilePos := XyT[uint]{X: tileX, Y: tileY}
			if !f.tileCanHaveCreep(tilePos) {
				continue
			}
			if flags&TileFlagOccupied != 0 {
				anyTilesOccupied = true
				continue
			}
			if spreadsCreep && creepDistanceSquared(dx, dy) > creepSpreadMaxDistanceSquared {
				continue
			}
			n := f.countNeighboringCreepTiles(tilePos)
			if n == 0 {
				continue
			}
			targetTiles[n-1].PushBack(index)
		}
	}
	if outAnyTilesOccupied != nil {
		*outAnyTilesOccupied = anyTilesOccupied
	}
	for i := len(targetTiles) - 1; i >= 0; i-- {
		v := &targetTiles[i]
		if v.Empty() {
			continue
		}
		index := v.At(int(uint(f.lcgRand(26)>>4) % uint(v.Size())))
		f.setTileCreep(XyT[uint]{X: index % f.gameSt.MapTileWidth, Y: index / f.gameSt.MapTileWidth}, true)
		return true
	}
	return false
}

func (f *StateFunctions) checkCreepTileIndexInRange(index uint) {
	if index >= uint(len(f.st.Tiles)) {
		throwError("vector::_M_range_check: __n (which is %d) >= this->size() (which is %d)", index, len(f.st.Tiles))
	}
}

func (f *StateFunctions) spreadCreepCompletely(unitType *UnitTypeT, pos Xy) {
	unitArea := creepUnitTileArea(unitType, pos)
	f.checkCreepTileIndexInRange(unitArea.From.Y*f.gameSt.MapTileWidth + unitArea.From.X)
	if unitArea.From != unitArea.To {
		f.checkCreepTileIndexInRange((unitArea.To.Y-1)*f.gameSt.MapTileWidth + unitArea.To.X - 1)
	}
	for y := unitArea.From.Y; y != unitArea.To.Y; y++ {
		for x := unitArea.From.X; x != unitArea.To.X; x++ {
			if !f.tileCanHaveCreep(XyT[uint]{X: x, Y: y}) {
				continue
			}
			f.setTileCreep(XyT[uint]{X: x, Y: y}, true)
		}
	}
	for f.spreadCreep(unitType, pos, nil) {
	}
}

func (f *StateFunctions) getSpawnLarvaPosition(u *Unit) Xy {
	if !f.unitIsHatchery(u.UnitType) {
		throwError("get_spawn_larva_position: unit is not a hatchery")
	}
	bestScore := int32(101)
	noPosition := Xy{X: -1, Y: -1}
	bestPos := noPosition
	test := func(index int, pos Xy, neighborOffset Xy) bool {
		val := u.Building.Hatchery.LarvaSpawnSideValues[index]
		if val >= bestScore {
			return false
		}
		if f.restrictMoveTargetToValidBounds(f.getUnitType(UnitTypesZergLarva), pos) != pos {
			return false
		}
		if f.st.Tiles[f.tileIndex(pos)].Flags&TileFlagHasCreep == 0 {
			return false
		}
		op := XyAdd(pos, neighborOffset)
		if f.restrictMoveTargetToValidBounds(f.getUnitType(UnitTypesZergLarva), op) != op {
			return false
		}
		flags := f.st.Tiles[f.tileIndex(op)].Flags
		if flags&TileFlagOccupied != 0 {
			return false
		}
		if flags&(TileFlagWalkable|TileFlagHasCreep) == 0 {
			return false
		}
		bestScore = val
		bestPos = pos
		return true
	}
	bb := f.unitSpriteInnerBoundingBox(u)
	spritePos := u.Sprite.Position
	test(3, Xy{X: spritePos.X, Y: bb.To.Y + 10}, Xy{X: 0, Y: 22})
	test(0, Xy{X: bb.From.X - 10, Y: spritePos.Y}, Xy{X: -22, Y: 0})
	test(2, Xy{X: bb.To.X + 10, Y: spritePos.Y}, Xy{X: 22, Y: 0})
	test(1, Xy{X: spritePos.X, Y: bb.From.Y - 10}, Xy{X: 0, Y: -22})
	if bestPos != noPosition {
		return bestPos
	}
	w := (u.UnitType.Dimensions.From.X + u.UnitType.Dimensions.To.X + 1) / 2
	h := (u.UnitType.Dimensions.From.Y + u.UnitType.Dimensions.To.Y + 1) / 2
	bestScore = 0x10000
	if !test(3, Xy{X: spritePos.X - w, Y: bb.To.Y + 10}, Xy{X: 0, Y: 22}) && !test(3, Xy{X: spritePos.X + w, Y: bb.To.Y + 10}, Xy{X: 0, Y: 22}) {
		bestScore = 101
	}
	if !test(0, Xy{X: bb.From.X - 10, Y: spritePos.Y - h}, Xy{X: -22, Y: 0}) {
		test(0, Xy{X: bb.From.X - 10, Y: spritePos.Y + h}, Xy{X: -22, Y: 0})
	}
	if !test(2, Xy{X: bb.To.X + 10, Y: spritePos.Y - h}, Xy{X: 22, Y: 0}) {
		test(2, Xy{X: bb.To.X + 10, Y: spritePos.Y + h}, Xy{X: 22, Y: 0})
	}
	if !test(1, Xy{X: spritePos.X - w, Y: bb.From.Y - 10}, Xy{X: 0, Y: -22}) {
		test(1, Xy{X: spritePos.X + w, Y: bb.From.Y - 10}, Xy{X: 0, Y: -22})
	}
	return bestPos
}

func (f *StateFunctions) spawnLarva(u *Unit) *Unit {
	pos := f.getSpawnLarvaPosition(u)
	if pos == (Xy{X: -1, Y: -1}) {
		return nil
	}
	larva := f.createUnit(f.getUnitType(UnitTypesZergLarva), pos, u.Owner)
	if larva != nil {
		f.onUnitCreated(larva, SourceLarva, u)
		f.finishBuildingUnit(larva)
		f.completeUnit(larva)
		larva.ConnectedUnit = u
		larvaPos := larva.Sprite.Position
		hatcheryPos := u.Sprite.Position
		dimensions := u.UnitType.Dimensions
		if larvaPos.X < hatcheryPos.X-dimensions.From.X {
			larva.OrderState = 0
		} else if larvaPos.Y < hatcheryPos.Y-dimensions.From.Y {
			larva.OrderState = 1
		} else if larvaPos.X > hatcheryPos.X+dimensions.To.X {
			larva.OrderState = 2
		} else {
			larva.OrderState = 3
		}
	}
	return larva
}
