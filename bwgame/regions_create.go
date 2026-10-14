package bwgame

import (
	"math"
	"slices"
)

const (
	walkFlagsRowSize          = 256 * 4
	regionUnmappedThreshold   = 5000
	regionMaxCount            = 5000
	regionSplitIndexBase      = 0x2000
	regionNoNeighbor          = 0x1fff
	regionUnwalkableFlags     = 0x1ffd
	regionEmptyFlags          = 0x1fff
	regionUnreachableGroup    = 0x4000
	unwalkableFlagUnwalkable  = 0x80
	unwalkableFlagAnyDirMask  = 0x7f
	regionsMaxNonEmptyRegions = 2500
)

func (f *GameLoadFunctions) getNewRegion() *Region {
	regions := &f.gameSt.Regions
	if cap(regions.Regions) < regionMaxCount {
		reserved := make([]Region, len(regions.Regions), regionMaxCount)
		copy(reserved, regions.Regions)
		regions.Regions = reserved
	}
	if len(regions.Regions) >= regionMaxCount {
		throwError("too many regions")
	}
	regions.Regions = append(regions.Regions, NewRegion())
	r := &regions.Regions[len(regions.Regions)-1]
	r.Index = uint(len(regions.Regions) - 1)
	return r
}

type regionsCreator struct {
	f               *GameLoadFunctions
	gameSt          *GameState
	st              *State
	regions         *Regions
	unwalkableFlags []uint8
}

func (f *GameLoadFunctions) regionsCreate() {
	c := &regionsCreator{
		f:               f,
		gameSt:          f.gameSt,
		st:              f.st,
		regions:         &f.gameSt.Regions,
		unwalkableFlags: make([]uint8, 256*4*256*4),
	}

	c.regions.TileBoundingBox = RectT[XyT[uint]]{To: XyT[uint]{X: c.gameSt.MapTileWidth, Y: c.gameSt.MapTileHeight}}

	c.setUnwalkableFlags()
	c.initTileRegionIndices()
	c.createUnreachableBottomRegion()
	c.createRegions()
	c.createContours()
}

func (c *regionsCreator) walkIndex(walkX, walkY uint) uint {
	return walkY*walkFlagsRowSize + walkX
}

func (c *regionsCreator) isUnwalkable(walkX, walkY uint) bool {
	return c.unwalkableFlags[c.walkIndex(walkX, walkY)]&unwalkableFlagUnwalkable != 0
}

func (c *regionsCreator) isWalkable(walkX, walkY uint) bool {
	return ^c.unwalkableFlags[c.walkIndex(walkX, walkY)]&unwalkableFlagUnwalkable != 0
}

func (c *regionsCreator) setUnwalkable(walkX, walkY uint) {
	c.unwalkableFlags[c.walkIndex(walkX, walkY)] |= unwalkableFlagUnwalkable
}

func (c *regionsCreator) isDirWalkable(walkX, walkY, dir uint) bool {
	return ^c.unwalkableFlags[c.walkIndex(walkX, walkY)]&(1<<dir) != 0
}

func (c *regionsCreator) isDirUnwalkable(walkX, walkY, dir uint) bool {
	return c.unwalkableFlags[c.walkIndex(walkX, walkY)]&(1<<dir) != 0
}

func (c *regionsCreator) setDirUnwalkable(walkX, walkY, dir uint) {
	c.unwalkableFlags[c.walkIndex(walkX, walkY)] |= 1 << dir
}

func (c *regionsCreator) flipDirWalkable(walkX, walkY, dir uint) {
	c.unwalkableFlags[c.walkIndex(walkX, walkY)] ^= 1 << dir
}

func (c *regionsCreator) isEveryDirWalkable(walkX, walkY uint) bool {
	return c.unwalkableFlags[c.walkIndex(walkX, walkY)]&unwalkableFlagAnyDirMask == 0
}

func (c *regionsCreator) tileRegionIndex(x, y uint) *uint {
	return &c.regions.TileRegionIndex[y*256+x]
}

func (c *regionsCreator) setUnwalkableFlags() {
	g := c.gameSt
	for y := uint(0); y != g.MapTileHeight; y++ {
		for x := uint(0); x != g.MapTileWidth; x++ {
			megaTileIndex := c.st.TilesMegaTileIndex[y*g.MapTileWidth+x]
			mt := &g.Vf4[megaTileIndex&0x7fff]
			for sy := uint(0); sy < 4; sy++ {
				for sx := uint(0); sx < 4; sx++ {
					if ^mt.Flags[sy*4+sx]&Vf4EntryFlagWalkable != 0 {
						c.setUnwalkable(x*4+sx, y*4+sy)
					}
				}
			}
		}
	}
	c.markBottomUIUnwalkable()

	if g.MapWalkWidth == 0 || g.MapWalkHeight == 0 {
		throwError("map width/height is zero")
	}

	for y := uint(0); y != g.MapWalkHeight; y++ {
		for x := uint(0); x != g.MapWalkWidth; x++ {
			if c.isUnwalkable(x, y) {
				continue
			}
			if y == 0 || c.isUnwalkable(x, y-1) {
				c.setDirUnwalkable(x, y, 0)
			}
			if x == g.MapWalkWidth-1 || c.isUnwalkable(x+1, y) {
				c.setDirUnwalkable(x, y, 1)
			}
			if y == g.MapWalkHeight-1 || c.isUnwalkable(x, y+1) {
				c.setDirUnwalkable(x, y, 2)
			}
			if x == 0 || c.isUnwalkable(x-1, y) {
				c.setDirUnwalkable(x, y, 3)
			}
		}
	}
}

func (c *regionsCreator) markBottomUIUnwalkable() {
	g := c.gameSt
	if g.MapWalkHeight < 8 {
		return
	}
	for y := g.MapWalkHeight - 8; y != g.MapWalkHeight; y++ {
		for x := uint(0); x != 20; x++ {
			c.setUnwalkable(x, y)
		}
		if g.MapWalkWidth >= 20 {
			for x := g.MapWalkWidth - 20; x != g.MapWalkWidth; x++ {
				c.setUnwalkable(x, y)
			}
		}
		if y >= g.MapWalkHeight-4 {
			for x := uint(0); x != g.MapWalkWidth; x++ {
				c.setUnwalkable(x, y)
			}
		}
	}
}

func (c *regionsCreator) initTileRegionIndices() {
	g := c.gameSt
	for y := uint(0); y != g.MapTileHeight; y++ {
		for x := uint(0); x != g.MapTileWidth; x++ {
			index := c.tileRegionIndex(x, y)
			t := &c.st.Tiles[y*g.MapTileWidth+x]
			switch {
			case ^t.Flags&TileFlagWalkable != 0:
				*index = 0x1ffd
			case t.Flags&TileFlagMiddle != 0:
				*index = 0x1ff9
			case t.Flags&TileFlagHigh != 0:
				*index = 0x1ffa
			default:
				*index = 0x1ffb
			}
		}
	}
}

func (c *regionsCreator) createRegion(area RectT[XyT[uint]]) *Region {
	r := c.f.getNewRegion()
	flags := uint16(*c.tileRegionIndex(area.From.X, area.From.Y))
	if flags < regionUnmappedThreshold {
		throwError("attempt to create region inside another region")
	}
	r.Flags = flags
	r.TileArea = area
	r.TileCenter.X = (area.From.X + area.To.X) / 2
	r.TileCenter.Y = (area.From.Y + area.To.Y) / 2
	tileCount := uint(0)
	index := r.Index
	for y := area.From.Y; y != area.To.Y; y++ {
		for x := area.From.X; x != area.To.X; x++ {
			if *c.tileRegionIndex(x, y) < regionUnmappedThreshold {
				throwError("attempt to create overlapping region")
			}
			*c.tileRegionIndex(x, y) = uint(uint16(index))
			tileCount++
		}
	}
	r.TileCount = tileCount
	return r
}

func (c *regionsCreator) createUnreachableBottomRegion() {
	g := c.gameSt
	r := c.createRegion(RectT[XyT[uint]]{
		From: XyT[uint]{X: 0, Y: g.MapTileHeight - 1},
		To:   XyT[uint]{X: g.MapTileWidth, Y: g.MapTileHeight},
	})
	r.Area = Rect{
		From: Xy{X: 0, Y: int32(g.MapHeight) - 32},
		To:   Xy{X: int32(g.MapWidth), Y: int32(g.MapHeight)},
	}
	r.Center.X = Fp8Integer(r.Area.From.X).Add(Fp8Integer(r.Area.To.X)).DivInt(2)
	r.Center.Y = Fp8Integer(r.Area.From.Y).Add(Fp8Integer(r.Area.To.Y)).DivInt(2)
	r.Flags = regionUnwalkableFlags
	r.GroupIndex = regionUnreachableGroup
}

type regionSearch struct {
	c               *regionsCreator
	bb              RectT[XyT[uint]]
	regionTileIndex uint
	regionX         uint
	regionY         uint
}

func (s *regionSearch) advance(x, y uint) (uint, uint) {
	x = s.bb.From.X
	if y+1 >= s.bb.To.Y {
		y = s.bb.From.Y
	} else {
		y = y + 1
	}
	return x, y
}

func (s *regionSearch) findEmptyRegion(x, y uint) bool {
	if x >= s.bb.To.X {
		x, y = s.advance(x, y)
	}
	startX := x
	startY := y
	for {
		index := *s.c.tileRegionIndex(x, y)
		if index >= regionUnmappedThreshold {
			s.regionTileIndex = index
			s.regionX = x
			s.regionY = y
			return true
		}
		x++
		if x >= s.bb.To.X {
			x, y = s.advance(x, y)
		}
		if x == startX && y == startY {
			return false
		}
	}
}

func (c *regionsCreator) findArea(beginX, beginY, index uint) RectT[XyT[uint]] {
	maxEndX := min(beginX+8, c.gameSt.MapTileWidth)
	maxEndY := min(beginY+7, c.gameSt.MapTileHeight)

	endX := beginX + 1
	endY := beginY + 1
	xIsGood := true
	yIsGood := true
	itsAllGood := true
	for (xIsGood || yIsGood) && (endX != maxEndX && endY != maxEndY) {
		if xIsGood {
			for y := beginY; y != endY; y++ {
				if *c.tileRegionIndex(endX, y) != index {
					xIsGood = false
					break
				}
			}
		}
		if yIsGood {
			for x := beginX; x != endX; x++ {
				if *c.tileRegionIndex(x, endY) != index {
					yIsGood = false
					break
				}
			}
		}

		if *c.tileRegionIndex(endX, endY) != index {
			itsAllGood = false
		}
		if itsAllGood {
			if yIsGood {
				endY++
			}
			if xIsGood {
				endX++
			}
		} else {
			if yIsGood {
				endY++
			} else if xIsGood {
				endX++
			}
		}
	}

	width := endX - beginX
	height := endY - beginY
	if width > height*3 {
		width = height * 3
	} else if height > width*3 {
		height = width * 3
	}
	endX = beginX + width
	endY = beginY + height

	return RectT[XyT[uint]]{From: XyT[uint]{X: beginX, Y: beginY}, To: XyT[uint]{X: endX, Y: endY}}
}

func tileAreaSize(area RectT[XyT[uint]]) uint {
	return (area.To.X - area.From.X) * (area.To.Y - area.From.Y)
}

func (c *regionsCreator) expand(r *Region) {
	g := c.gameSt
	if r.TileArea.From.X > 0 {
		r.TileArea.From.X--
	}
	if r.TileArea.From.Y > 0 {
		r.TileArea.From.Y--
	}
	if r.TileArea.To.X < g.MapTileWidth {
		r.TileArea.To.X++
	}
	if r.TileArea.To.Y < g.MapTileHeight {
		r.TileArea.To.Y++
	}
	beginX := r.TileArea.From.X
	beginY := r.TileArea.From.Y
	endX := r.TileArea.To.X
	endY := r.TileArea.To.Y

	flags := uint(r.Flags)
	index := r.Index

	isNeighbor := func(x, y uint) bool {
		if x != 0 && *c.tileRegionIndex(x-1, y) == index {
			return true
		}
		if x != g.MapTileWidth-1 && *c.tileRegionIndex(x+1, y) == index {
			return true
		}
		if y != 0 && *c.tileRegionIndex(x, y-1) == index {
			return true
		}
		if y != g.MapTileHeight-1 && *c.tileRegionIndex(x, y+1) == index {
			return true
		}
		return false
	}

	for i := 0; i < 2; i++ {
		for y := beginY; y != endY; y++ {
			for x := beginX; x != endX; x++ {
				if *c.tileRegionIndex(x, y) == flags && isNeighbor(x, y) {
					*c.tileRegionIndex(x, y) = index
				}
			}
		}
	}
}

func (c *regionsCreator) createRegions() {
	s := &regionSearch{c: c, bb: c.regions.TileBoundingBox}
	c.fillRegions(s)
	c.refreshRegions()
	c.mergeSmallRegions()
	c.compactRegions()
	c.refreshRegions()
	c.createSplitRegions()
}

func (c *regionsCreator) fillRegions(s *regionSearch) {
	bb := s.bb
	nextX := bb.From.X
	nextY := bb.From.Y

	hasExpandedAll := false
	initialRegionsSize := uint(len(c.regions.Regions))

	prevSize := uint(7 * 8)

	for {
		startX := nextX
		startY := nextY
		if startX >= bb.To.X {
			startX = bb.From.X
			startY++
			if startY >= bb.To.Y {
				startY = bb.From.Y
			}
		}
		if !s.findEmptyRegion(startX, startY) {
			if len(c.regions.Regions) >= regionMaxCount {
				throwError("too many regions (nooks and crannies)")
			}
			break
		}

		area := c.findArea(s.regionX, s.regionY, s.regionTileIndex)

		size := tileAreaSize(area)
		if size < prevSize {
			bestArea := area
			bestSize := size

			for n := 0; n != 25; n++ {
				if !s.findEmptyRegion(area.To.X, s.regionY) {
					break
				}
				area = c.findArea(s.regionX, s.regionY, s.regionTileIndex)
				size := tileAreaSize(area)
				if size > bestSize {
					bestSize = size
					bestArea = area
					if size >= prevSize {
						break
					}
				}
			}

			area = bestArea
			size = bestSize
		}

		prevSize = size

		nextX = area.To.X
		nextY = area.From.Y

		if len(c.regions.Regions) >= regionMaxCount {
			throwError("too many regions (nooks and crannies)")
		}

		r := c.createRegion(area)

		c.expand(r)

		if size <= 6 && !hasExpandedAll {
			hasExpandedAll = true
			for i := initialRegionsSize; i != uint(len(c.regions.Regions)); i++ {
				c.expand(&c.regions.Regions[i])
			}
		}
	}
}

func (c *regionsCreator) getNeighbors(tileX, tileY uint) [8]uint {
	var r [8]uint
	n := 0
	test := func(cond bool, x, y uint) {
		if !cond {
			r[n] = regionNoNeighbor
		} else {
			r[n] = *c.tileRegionIndex(x, y)
		}
		n++
	}
	w := c.gameSt.MapTileWidth
	h := c.gameSt.MapTileHeight
	test(tileY > 0, tileX, tileY-1)
	test(tileX > 0, tileX-1, tileY)
	test(tileX+1 < w, tileX+1, tileY)
	test(tileY+1 < h, tileX, tileY+1)
	test(tileY > 0 && tileX > 0, tileX-1, tileY-1)
	test(tileY > 0 && tileX+1 < w, tileX+1, tileY-1)
	test(tileY+1 < h && tileX > 0, tileX-1, tileY+1)
	test(tileY+1 < h && tileX+1 < w, tileX+1, tileY+1)
	return r
}

func (c *regionsCreator) refreshRegions() {
	g := c.gameSt
	regions := c.regions.Regions
	for i := range regions {
		r := &regions[i]
		r.Area = Rect{
			From: Xy{X: math.MaxInt32, Y: math.MaxInt32},
			To:   Xy{X: math.MinInt32, Y: math.MinInt32},
		}
		r.TileCount = 0
	}
	for y := uint(0); y != g.MapTileHeight; y++ {
		for x := uint(0); x != g.MapTileWidth; x++ {
			index := *c.tileRegionIndex(x, y)
			if index >= regionUnmappedThreshold {
				continue
			}
			r := &regions[index]
			r.TileCount++
			tileX := int32(x)
			tileY := int32(y)
			if r.Area.From.X > tileX*32 {
				r.Area.From.X = 32 * tileX
			}
			if r.Area.From.Y > tileY*32 {
				r.Area.From.Y = 32 * tileY
			}
			if r.Area.To.X < (tileX+1)*32 {
				r.Area.To.X = 32 * (tileX + 1)
			}
			if r.Area.To.Y < (tileY+1)*32 {
				r.Area.To.Y = 32 * (tileY + 1)
			}
		}
	}

	for i := range regions {
		if regions[i].TileCount == 0 {
			regions[i].Flags = regionEmptyFlags
		}
	}

	for i := range regions {
		r := &regions[i]
		if r.TileCount == 0 {
			continue
		}
		c.refreshRegionNeighbors(r)
		c.moveFirstRegionToBackOfNonWalkableNeighbors(r)
	}

	for i := range regions {
		r := &regions[i]
		r.Center = XyFp8{
			X: Fp8Integer(int32(r.TileCenter.X*32 + 16)),
			Y: Fp8Integer(int32(r.TileCenter.Y*32 + 16)),
		}
	}

	c.assignGroupIndices()
}

func (c *regionsCreator) refreshRegionNeighbors(r *Region) {
	r.WalkableNeighbors = r.WalkableNeighbors[:0]
	r.NonWalkableNeighbors = r.NonWalkableNeighbors[:0]

	for y := r.Area.From.Y / 32; y != r.Area.To.Y/32; y++ {
		for x := r.Area.From.X / 32; x != r.Area.To.X/32; x++ {
			if *c.tileRegionIndex(uint(x), uint(y)) != r.Index {
				continue
			}
			neighbors := c.getNeighbors(uint(x), uint(y))
			for i := uint(0); i != 8; i++ {
				nindex := neighbors[i]
				if nindex == regionNoNeighbor || nindex == r.Index {
					continue
				}
				nr := &c.regions.Regions[nindex]
				add := i < 4 || !r.Walkable() || !nr.Walkable() || c.isDiagonalNeighborWalkable(i, uint(x), uint(y))
				if !add {
					continue
				}
				if nr.Walkable() {
					if !slices.Contains(r.WalkableNeighbors, nr) {
						r.WalkableNeighbors = append(r.WalkableNeighbors, nr)
					}
				} else {
					if !slices.Contains(r.NonWalkableNeighbors, nr) {
						r.NonWalkableNeighbors = append(r.NonWalkableNeighbors, nr)
					}
				}
			}
		}
	}
}

func (c *regionsCreator) is2x2Walkable(walkX, walkY uint) bool {
	if !c.isWalkable(walkX, walkY) {
		return false
	}
	if !c.isWalkable(walkX+1, walkY) {
		return false
	}
	if !c.isWalkable(walkX, walkY+1) {
		return false
	}
	if !c.isWalkable(walkX+1, walkY+1) {
		return false
	}
	return true
}

func (c *regionsCreator) isDiagonalNeighborWalkable(i, x, y uint) bool {
	walkX := x * 4
	walkY := y * 4
	switch i {
	case 4:
		if c.is2x2Walkable(walkX-2, walkY-2) && c.is2x2Walkable(walkX, walkY) {
			if c.is2x2Walkable(walkX-2, walkY) {
				return true
			} else if c.is2x2Walkable(walkX, walkY-2) {
				return true
			}
		}
	case 5:
		if c.is2x2Walkable(walkX+4, walkY-2) && c.is2x2Walkable(walkX+2, walkY) {
			if c.is2x2Walkable(walkX+2, walkY-2) {
				return true
			} else if c.is2x2Walkable(walkX+4, walkY) {
				return true
			}
		}
	case 6:
		if c.is2x2Walkable(walkX, walkY+2) && c.is2x2Walkable(walkX-2, walkY+4) {
			if c.is2x2Walkable(walkX-2, walkY+2) {
				return true
			} else if c.is2x2Walkable(walkX, walkY+4) {
				return true
			}
		}
	case 7:
		if c.is2x2Walkable(walkX+2, walkY+2) && c.is2x2Walkable(walkX+4, walkY+4) {
			if c.is2x2Walkable(walkX+4, walkY+2) {
				return true
			} else if c.is2x2Walkable(walkX+2, walkY+4) {
				return true
			}
		}
	}
	return false
}

func (c *regionsCreator) moveFirstRegionToBackOfNonWalkableNeighbors(r *Region) {
	neighbors := r.NonWalkableNeighbors
	if len(neighbors) == 0 {
		return
	}
	front := &c.regions.Regions[0]
	last := len(neighbors) - 1
	for i := range neighbors {
		if neighbors[i] == front && i != last {
			neighbors[i], neighbors[last] = neighbors[last], neighbors[i]
		}
	}
}

func (c *regionsCreator) assignGroupIndices() {
	regions := c.regions.Regions
	for i := range regions {
		if regions[i].GroupIndex < regionUnreachableGroup {
			regions[i].GroupIndex = 0
		}
	}
	var stack []*Region
	nextGroupIndex := uint(1)
	for i := range regions {
		r := &regions[i]
		if r.GroupIndex != 0 || r.TileCount == 0 {
			continue
		}
		groupIndex := nextGroupIndex
		nextGroupIndex++
		r.GroupIndex = groupIndex
		stack = append(stack, r)
		for len(stack) != 0 {
			cr := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			neighbors := cr.NonWalkableNeighbors
			if cr.Walkable() {
				neighbors = cr.WalkableNeighbors
			}
			for _, nr := range neighbors {
				if nr.GroupIndex == 0 {
					nr.GroupIndex = groupIndex
					stack = append(stack, nr)
				}
			}
		}
	}
}

func (c *regionsCreator) mergeSmallRegions() {
	for n := uint(6); ; n += 2 {
		regions := c.regions.Regions
		for i := len(regions) - 1; i >= 0; i-- {
			r := &regions[i]
			if r.TileCount == 0 || r.TileCount >= n || r.GroupIndex >= regionUnreachableGroup {
				continue
			}
			smallestNeighbor := c.findSmallestMergeableNeighbor(r)
			if smallestNeighbor != nil {
				c.mergeRegionInto(r, smallestNeighbor)
			}
		}

		nonEmptyRegions := 0
		for i := range regions {
			if regions[i].TileCount != 0 {
				nonEmptyRegions++
			}
		}
		if nonEmptyRegions < regionsMaxNonEmptyRegions {
			break
		}
	}
}

func (c *regionsCreator) findSmallestMergeableNeighbor(r *Region) *Region {
	var smallestNeighbor *Region
	eval := func(nr *Region) {
		if nr.TileCount == 0 || nr.GroupIndex >= regionUnreachableGroup || nr.Flags != r.Flags {
			return
		}
		if smallestNeighbor == nil || nr.TileCount < smallestNeighbor.TileCount {
			smallestNeighbor = nr
		}
	}
	for _, nr := range r.WalkableNeighbors {
		eval(nr)
	}
	for _, nr := range r.NonWalkableNeighbors {
		eval(nr)
	}
	return smallestNeighbor
}

func (c *regionsCreator) mergeRegionInto(r, mergeInto *Region) {
	for y := uint(r.Area.From.Y) / 32; y != uint(r.Area.To.Y)/32; y++ {
		for x := uint(r.Area.From.X) / 32; x != uint(r.Area.To.X)/32; x++ {
			index := c.tileRegionIndex(x, y)
			if *index == r.Index {
				*index = mergeInto.Index
			}
		}
	}
	mergeInto.TileCount += r.TileCount
	r.TileCount = 0
	r.Flags = regionEmptyFlags
	if r.Area.From.X < mergeInto.Area.From.X {
		mergeInto.Area.From.X = r.Area.From.X
	}
	if r.Area.From.Y < mergeInto.Area.From.Y {
		mergeInto.Area.From.Y = r.Area.From.Y
	}
	if r.Area.To.X > mergeInto.Area.To.X {
		mergeInto.Area.To.X = r.Area.To.X
	}
	if r.Area.To.Y > mergeInto.Area.To.Y {
		mergeInto.Area.To.Y = r.Area.To.Y
	}
}

func (c *regionsCreator) compactRegions() {
	g := c.gameSt
	regions := c.regions.Regions
	reindex := make([]uint, regionMaxCount)
	newRegionCount := uint(0)
	for i := range regions {
		r := &regions[i]
		r.WalkableNeighbors = nil
		r.NonWalkableNeighbors = nil
		if r.TileCount == 0 {
			continue
		}
		newIndex := newRegionCount
		newRegionCount++
		reindex[i] = newIndex
		if uint(i) != newIndex {
			r.Index = newIndex
			regions[newIndex] = *r
		}
	}
	for y := uint(0); y != g.MapTileHeight; y++ {
		for x := uint(0); x != g.MapTileWidth; x++ {
			index := c.tileRegionIndex(x, y)
			*index = reindex[*index]
		}
	}
	c.regions.Regions = regions[:newRegionCount]
}

func (c *regionsCreator) count4x1Walkable(walkX, walkY uint) uint {
	r := uint(0)
	for i := uint(0); i != 4; i++ {
		if c.isWalkable(walkX+i, walkY) {
			r++
		}
	}
	return r
}

func (c *regionsCreator) count1x4Walkable(walkX, walkY uint) uint {
	r := uint(0)
	for i := uint(0); i != 4; i++ {
		if c.isWalkable(walkX, walkY+i) {
			r++
		}
	}
	return r
}

func (c *regionsCreator) tileUnwalkableMask(megatileIndex uint) uint16 {
	mask := uint16(0)
	vf4 := &c.gameSt.Vf4[megatileIndex]
	for sy := uint(0); sy != 4; sy++ {
		for sx := uint(0); sx != 4; sx++ {
			flags := int32(vf4.Flags[sy*4+sx])
			mask >>= 1
			if ^flags&int32(Vf4EntryFlagWalkable) != 0 {
				mask |= 0x8000
			}
		}
	}
	return mask
}

func (c *regionsCreator) createSplitRegions() {
	g := c.gameSt
	c.regions.SplitRegions = c.regions.SplitRegions[:0]

	for y := uint(0); y != g.MapTileHeight; y++ {
		for x := uint(0); x != g.MapTileWidth; x++ {
			index := y*g.MapTileWidth + x
			tile := c.st.Tiles[index]
			if ^tile.Flags&TileFlagPartiallyWalkable != 0 {
				continue
			}
			c.splitTile(x, y, index)
		}
	}
}

func (c *regionsCreator) splitTile(x, y, index uint) {
	neighbors := c.getNeighbors(x, y)
	r := &c.regions.Regions[*c.tileRegionIndex(x, y)]
	mask := c.tileUnwalkableMask(uint(c.st.TilesMegaTileIndex[index]))

	walkX := x * 4
	walkY := y * 4
	var r2 *Region
	if !r.Walkable() {
		r2 = c.findWalkableSplitNeighbor(r, neighbors, walkX, walkY)
	} else {
		r2 = c.findUnwalkableSplitNeighbor(r, neighbors, walkX, walkY)
	}
	if r2 == nil || r2 == r {
		if r.Walkable() {
			mask = 0
		} else {
			mask = 0xffff
		}
	} else if !r.Walkable() && !r2.Walkable() {
		mask = 0xffff
	}
	*c.tileRegionIndex(x, y) = regionSplitIndexBase + uint(len(c.regions.SplitRegions))
	if r2 == nil {
		r2 = r
	}
	if r.Walkable() {
		c.regions.SplitRegions = append(c.regions.SplitRegions, SplitRegion{Mask: mask, A: r, B: r2})
	} else {
		c.regions.SplitRegions = append(c.regions.SplitRegions, SplitRegion{Mask: mask, A: r2, B: r})
	}
}

func (c *regionsCreator) findWalkableSplitNeighbor(r *Region, neighbors [8]uint, walkX, walkY uint) *Region {
	var nWalkable [4]uint
	nWalkable[0] = c.count4x1Walkable(walkX, walkY)
	nWalkable[1] = c.count1x4Walkable(walkX, walkY)
	nWalkable[2] = c.count1x4Walkable(walkX+3, walkY)
	nWalkable[3] = c.count4x1Walkable(walkX, walkY+3)
	highestN := uint(0)
	var highestNIndex uint
	for i := uint(4); i != 0; {
		i--
		n := nWalkable[i]
		if n <= highestN {
			continue
		}
		nindex := neighbors[i]
		if nindex == r.Index {
			continue
		}
		if nindex < regionSplitIndexBase {
			if nindex >= regionUnmappedThreshold {
				continue
			}
			nr := &c.regions.Regions[nindex]
			if nr.Walkable() {
				highestN = n
				highestNIndex = nindex
			}
		} else {
			set := false
			if i == 0 && c.count4x1Walkable(walkX, walkY-1) != 0 {
				set = true
			}
			if i == 1 && c.count1x4Walkable(walkX-1, walkY) != 0 {
				set = true
			}
			if i == 2 && c.count1x4Walkable(walkX+4, walkY) != 0 {
				set = true
			}
			if i == 3 && c.count4x1Walkable(walkX, walkY+4) != 0 {
				set = true
			}
			if set {
				highestN = n
				highestNIndex = nindex
			}
		}
	}
	if highestN == 0 {
		return nil
	}
	if highestNIndex < regionSplitIndexBase {
		return &c.regions.Regions[highestNIndex]
	}
	return c.regions.SplitRegions[highestNIndex-regionSplitIndexBase].A
}

func (c *regionsCreator) findUnwalkableSplitNeighbor(r *Region, neighbors [8]uint, walkX, walkY uint) *Region {
	var nUnwalkable [8]uint
	nUnwalkable[0] = 4 - c.count4x1Walkable(walkX, walkY)
	nUnwalkable[1] = 4 - c.count1x4Walkable(walkX, walkY)
	nUnwalkable[2] = 4 - c.count1x4Walkable(walkX+3, walkY)
	nUnwalkable[3] = 4 - c.count4x1Walkable(walkX, walkY+3)
	nUnwalkable[4] = c.unwalkableCount(walkX, walkY)
	nUnwalkable[5] = c.unwalkableCount(walkX+3, walkY)
	nUnwalkable[6] = c.unwalkableCount(walkX, walkY+3)
	nUnwalkable[7] = c.unwalkableCount(walkX+3, walkY+3)
	highestN := uint(0)
	var highestNIndex uint
	for i := uint(8); i != 0; {
		i--
		n := nUnwalkable[i]
		if n <= highestN {
			continue
		}
		nindex := neighbors[i]
		if nindex == r.Index {
			continue
		}
		if nindex < regionSplitIndexBase {
			if nindex >= regionUnmappedThreshold {
				continue
			}
			nr := &c.regions.Regions[nindex]
			if !nr.Walkable() {
				highestN = n
				highestNIndex = nindex
			}
		} else if c.splitNeighborHasUnwalkableEdge(i, walkX, walkY) {
			highestN = n
			highestNIndex = nindex
		}
	}
	if highestN == 0 {
		return nil
	}
	if highestNIndex < regionSplitIndexBase {
		return &c.regions.Regions[highestNIndex]
	}
	return c.regions.SplitRegions[highestNIndex-regionSplitIndexBase].B
}

func (c *regionsCreator) unwalkableCount(walkX, walkY uint) uint {
	if c.isWalkable(walkX, walkY) {
		return 0
	}
	return 1
}

func (c *regionsCreator) splitNeighborHasUnwalkableEdge(i, walkX, walkY uint) bool {
	set := false
	if i == 0 && c.count4x1Walkable(walkX, walkY-1) != 4 {
		set = true
	}
	if i == 1 && c.count1x4Walkable(walkX-1, walkY) != 4 {
		set = true
	}
	if i == 2 && c.count1x4Walkable(walkX+4, walkY) != 4 {
		set = true
	}
	if i == 3 && c.count4x1Walkable(walkX, walkY+4) != 4 {
		set = true
	}
	if i == 4 && !c.isWalkable(walkX-1, walkY-1) {
		set = true
	}
	if i == 5 && !c.isWalkable(walkX+4, walkY-1) {
		set = true
	}
	if i == 6 && !c.isWalkable(walkX-1, walkY+4) {
		set = true
	}
	if i == 7 && !c.isWalkable(walkX+4, walkY+4) {
		set = true
	}
	return set
}

type contourWalker struct {
	c     *regionsCreator
	nextX uint
	nextY uint
}

func (w *contourWalker) next() bool {
	c := w.c
	g := c.gameSt
	x := w.nextX
	y := w.nextY
	if x >= g.MapWalkWidth {
		throwError("create_contours::next: unreachable")
	}
	x = x / 4 * 4
	startX := x
	startY := y
	for c.isEveryDirWalkable(x, y) && c.isEveryDirWalkable(x+1, y) && c.isEveryDirWalkable(x+2, y) && c.isEveryDirWalkable(x+3, y) {
		x += 4
		if x == g.MapWalkWidth {
			x = 0
			y++
			if y == g.MapWalkHeight {
				y = 0
			}
		}
		if x == startX && y == startY {
			return false
		}
	}
	for !c.isWalkable(x, y) || c.isEveryDirWalkable(x, y) {
		x++
		if x == g.MapWalkWidth {
			throwError("create_contours: out of bounds")
		}
	}
	w.nextX = x
	w.nextY = y
	return true
}

var contourStartLut = [16]int32{-1, -1, 8, -1, 8, 8, -1, 8, 0, -1, 8, 0, 7, 8, -1, 7}
var contourNextLut = [16]int32{8, -1, 8, 8, -1, 8, -1, -1, 7, -1, 8, 7, 0, 8, -1, 0}
var contourNextRelative = [4][2]int{{1, -1}, {1, 1}, {-1, 1}, {-1, -1}}

func (c *regionsCreator) createContours() {
	c.regions.Contours = [4][]Contour{}

	w := &contourWalker{c: c}
	for w.next() {
		c.traceContour(w.nextX, w.nextY)
	}

	for i := range c.regions.Contours {
		slices.SortFunc(c.regions.Contours[i], compareContours)
	}
}

func compareContours(a, b Contour) int {
	if a.V[0] != b.V[0] {
		if a.V[0] < b.V[0] {
			return -1
		}
		return 1
	}
	if a.V[1] < b.V[1] {
		return -1
	}
	if a.V[1] > b.V[1] {
		return 1
	}
	return 0
}

func contourCoordinate(lutValue int32, walkCoordinate uint) int32 {
	return lutValue + int32(walkCoordinate*8)
}

func (c *regionsCreator) traceContour(x, y uint) {
	if c.isDirWalkable(x, y, 0) && c.isDirWalkable(x, y, 1) && c.isDirUnwalkable(x, y, 2) {
		y++
		x--
	}
	firstUnwalkableDir := uint(0)
	lut1val := uint(1)
	if c.isDirUnwalkable(x, y, 0) && c.isDirUnwalkable(x, y, 3) {
		lut1val = 0
	}
	for i := uint(0); i < 4; i++ {
		if c.isDirUnwalkable(x, y, i) {
			firstUnwalkableDir = i
			break
		}
	}

	clutIndex := firstUnwalkableDir + 4*lut1val
	startCx := contourCoordinate(contourStartLut[clutIndex*2], x)
	startCy := contourCoordinate(contourStartLut[clutIndex*2+1], y)

	c.flipDirWalkable(x, y, firstUnwalkableDir)

	cx := startCx
	cy := startCy

	curDir := firstUnwalkableDir
	for {
		nextDir := (curDir + 1) & 3
		relX := contourRelX(nextDir)
		relY := contourRelY(nextDir)
		c.flipDirWalkable(x, y, curDir)
		nextWalkable := uint(0)
		if c.isDirWalkable(x, y, nextDir) {
			nextWalkable = 1
			for c.isDirUnwalkable(x+uint(relX), y+uint(relY), curDir) {
				x += uint(relX)
				y += uint(relY)
				c.flipDirWalkable(x, y, curDir)
				if c.isDirUnwalkable(x, y, nextDir) {
					nextWalkable = 0
					break
				}
			}
		}
		nlutIndex := curDir + 4*nextWalkable

		nx := contourCoordinate(contourNextLut[nlutIndex*2], x)
		ny := contourCoordinate(contourNextLut[nlutIndex*2+1], y)

		c.addContour(curDir, cx, cy, nx, ny, lut1val, nextWalkable)

		if nextWalkable == 0 {
			curDir = nextDir
		} else {
			x += uint(contourNextRelative[curDir][0])
			y += uint(contourNextRelative[curDir][1])
			if curDir == 0 {
				curDir = 3
			} else {
				curDir--
			}
		}
		cx = nx
		cy = ny
		lut1val = nextWalkable

		if cx == startCx && cy == startCy {
			break
		}
	}

	c.flipDirWalkable(x, y, curDir)
}

func contourRelX(dir uint) int {
	switch dir {
	case 1:
		return 1
	case 3:
		return -1
	}
	return 0
}

func contourRelY(dir uint) int {
	switch dir {
	case 0:
		return -1
	case 2:
		return 1
	}
	return 0
}

func contourFlags(curDir, lut1val, nextWalkable uint) uint8 {
	flags0 := uint8(curDir ^ (2 * lut1val) ^ (^(2 * curDir) & 3))
	flags1 := uint8(curDir ^ (2 * (nextWalkable ^ (curDir & 1))) ^ 1)
	if curDir == 0 || curDir == 1 {
		return uint8(uint(flags0&3) | 4*(uint(flags1&3)|4*(lut1val|2*nextWalkable)))
	}
	return uint8(uint(flags1&3) | 4*(uint(flags0&3)|4*(nextWalkable|2*lut1val)))
}

func (c *regionsCreator) addContour(curDir uint, cx, cy, nx, ny int32, lut1val, nextWalkable uint) {
	flags := contourFlags(curDir, lut1val, nextWalkable)
	var v [3]int32
	switch curDir {
	case 0:
		v = [3]int32{cy, cx, nx}
	case 1:
		v = [3]int32{cx, cy, ny}
	case 2:
		v = [3]int32{cy, nx, cx}
	default:
		v = [3]int32{cx, ny, cy}
	}
	c.regions.Contours[curDir] = append(c.regions.Contours[curDir], Contour{V: v, Dir: curDir, Flags: flags})
}
