package bwgame

import (
	"iter"
	"slices"
)

func unitFinderIndexOf(vec []UnitFinderEntry, value int32, u *Unit) int {
	i := unitFinderLowerBound(vec, value)
	for vec[i].U != u {
		i++
	}
	return i
}

func unitFinderIndexOfWithHint(vec []UnitFinderEntry, value int32, u *Unit, hint int) int {
	for _, offset := range [5]int{0, -1, 1, -2, 2} {
		i := hint + offset
		if uint(i) < uint(len(vec)) && vec[i].U == u && vec[i].Value == value {
			return i
		}
	}
	return unitFinderIndexOf(vec, value, u)
}

func (f *StateFunctions) errorIfUnitFinderSearchActive() {
	if f.unitFinderSearchIndex != 0 {
		throwError("attempt to modify unit finder while search is active")
	}
}

func (f *StateFunctions) unitFinderInsert(u *Unit) {
	if f.utTurret(u.UnitType) {
		return
	}
	bb := f.unitSpriteInnerBoundingBox(u)
	f.unitFinderInsertBB(u, bb)
}

func (f *StateFunctions) unitFinderReinsert(u *Unit) {
	if u.UnitFinderBoundingBox.To.X == -1 {
		return
	}
	bb := f.unitSpriteInnerBoundingBox(u)
	f.unitFinderReinsertBB(u, bb)
}

func (f *StateFunctions) unitFinderRemove(u *Unit) {
	if u.UnitFinderBoundingBox.To.X == -1 {
		return
	}
	f.errorIfUnitFinderSearchActive()
	remove := func(vec []UnitFinderEntry, value int32) []UnitFinderEntry {
		i := unitFinderIndexOf(vec, value, u)
		return slices.Delete(vec, i, i+1)
	}
	st := f.st
	st.UnitFinderX = remove(st.UnitFinderX, u.UnitFinderBoundingBox.From.X)
	st.UnitFinderX = remove(st.UnitFinderX, u.UnitFinderBoundingBox.To.X)
	st.UnitFinderY = remove(st.UnitFinderY, u.UnitFinderBoundingBox.From.Y)
	st.UnitFinderY = remove(st.UnitFinderY, u.UnitFinderBoundingBox.To.Y)
	u.UnitFinderBoundingBox = Rect{From: Xy{X: -1, Y: -1}, To: Xy{X: -1, Y: -1}}
}

func (f *StateFunctions) unitFinderInsertBB(u *Unit, bb Rect) {
	f.errorIfUnitFinderSearchActive()
	insert := func(vec []UnitFinderEntry, fromValue int32, toValue int32) []UnitFinderEntry {
		fromI := unitFinderLowerBound(vec, fromValue)
		vec = slices.Insert(vec, fromI, UnitFinderEntry{U: u, Value: fromValue})
		toI := unitFinderLowerBound(vec, toValue)
		return slices.Insert(vec, toI, UnitFinderEntry{U: u, Value: toValue})
	}
	st := f.st
	st.UnitFinderX = insert(st.UnitFinderX, bb.From.X, bb.To.X)
	st.UnitFinderY = insert(st.UnitFinderY, bb.From.Y, bb.To.Y)
	u.UnitFinderBoundingBox = bb
}

func unitFinderReinsertValue(vec []UnitFinderEntry, u *Unit, oldValue int32, newValue int32, indexHint *int) {
	if oldValue == newValue {
		return
	}
	i := unitFinderIndexOfWithHint(vec, oldValue, u, *indexHint)
	*indexHint = unitFinderMoveEntry(vec, i, u, oldValue, newValue)
}

func unitFinderMoveEntry(vec []UnitFinderEntry, i int, u *Unit, oldValue int32, newValue int32) int {
	if newValue > oldValue {
		ni := i + 1
		for ni != len(vec) && vec[ni].Value < newValue {
			vec[i] = vec[ni]
			i++
			ni++
		}
		vec[i] = UnitFinderEntry{U: u, Value: newValue}
		return i
	}
	for i != 0 {
		ni := i
		i--
		if vec[i].Value <= newValue {
			i++
			break
		}
		vec[ni] = vec[i]
	}
	vec[i] = UnitFinderEntry{U: u, Value: newValue}
	return i
}

func (f *StateFunctions) unitFinderReinsertBB(u *Unit, bb Rect) {
	f.errorIfUnitFinderSearchActive()
	st := f.st
	old := u.UnitFinderBoundingBox
	hints := &u.unitFinderIndexHints
	if bb.From.X <= old.From.X {
		unitFinderReinsertValue(st.UnitFinderX, u, old.From.X, bb.From.X, &hints[0])
		unitFinderReinsertValue(st.UnitFinderX, u, old.To.X, bb.To.X, &hints[1])
	} else {
		unitFinderReinsertValue(st.UnitFinderX, u, old.To.X, bb.To.X, &hints[1])
		unitFinderReinsertValue(st.UnitFinderX, u, old.From.X, bb.From.X, &hints[0])
	}
	if bb.From.Y <= old.From.Y {
		unitFinderReinsertValue(st.UnitFinderY, u, old.From.Y, bb.From.Y, &hints[2])
		unitFinderReinsertValue(st.UnitFinderY, u, old.To.Y, bb.To.Y, &hints[3])
	} else {
		unitFinderReinsertValue(st.UnitFinderY, u, old.To.Y, bb.To.Y, &hints[3])
		unitFinderReinsertValue(st.UnitFinderY, u, old.From.Y, bb.From.Y, &hints[2])
	}
	u.UnitFinderBoundingBox = bb
}

type unitFinderSearch struct {
	vec         []UnitFinderEntry
	iBegin      int
	iEnd        int
	area        Rect
	searchIndex uint
}

func (s *unitFinderSearch) inBounds(i int) bool {
	u := s.vec[i].U
	if u.UnitFinderBoundingBox.From.X >= s.area.To.X {
		return false
	}
	if u.UnitFinderBoundingBox.From.Y >= s.area.To.Y {
		return false
	}
	if u.UnitFinderBoundingBox.To.Y < s.area.From.Y {
		return false
	}
	return true
}

func (s *unitFinderSearch) visited(i int) bool {
	return s.vec[i].U.UnitFinderVisited[s.searchIndex]
}

func (s *unitFinderSearch) markVisited(i int) {
	s.vec[i].U.UnitFinderVisited[s.searchIndex] = true
}

func (s *unitFinderSearch) next(i int) int {
	for {
		i++
		if i == s.iEnd {
			return i
		}
		if s.inBounds(i) && !s.visited(i) {
			break
		}
	}
	s.markVisited(i)
	return i
}

func (s *unitFinderSearch) begin() int {
	i := s.iBegin
	if i != s.iEnd {
		if s.inBounds(i) {
			s.markVisited(i)
		} else {
			i = s.next(i)
		}
	}
	return i
}

func (s *unitFinderSearch) clearVisited() {
	for i := s.iBegin; i != s.iEnd; i++ {
		s.vec[i].U.UnitFinderVisited[s.searchIndex] = false
	}
}

func (f *StateFunctions) newUnitFinderSearch(area Rect, expand bool) unitFinderSearch {
	if f.unitFinderSearchIndex == 4 {
		throwError("unit_finder_search maximum recursive depth reached")
	}
	s := unitFinderSearch{vec: f.st.UnitFinderX, area: area, searchIndex: f.unitFinderSearchIndex}
	f.unitFinderSearchIndex++

	beginX := area.From.X
	endX := area.To.X
	if expand {
		if endX-beginX+1 < f.gameSt.MaxUnitWidth {
			endX = beginX + f.gameSt.MaxUnitWidth - 1
			s.area.To.X++
		}
		if area.To.Y-area.From.Y+1 < f.gameSt.MaxUnitHeight {
			s.area.To.Y++
		}
	}
	s.iBegin = unitFinderLowerBound(s.vec, beginX)
	s.iEnd = unitFinderLowerBound(s.vec, endX)
	return s
}

func (f *StateFunctions) unitFinderSearchUnits(area Rect, expand bool) iter.Seq[*Unit] {
	return func(yield func(*Unit) bool) {
		s := f.newUnitFinderSearch(area, expand)
		defer func() {
			f.unitFinderSearchIndex--
			s.clearVisited()
		}()
		for i := s.begin(); i != s.iEnd; i = s.next(i) {
			if !yield(s.vec[i].U) {
				return
			}
		}
	}
}

func (f *StateFunctions) findUnits(area Rect) iter.Seq[*Unit] {
	return f.unitFinderSearchUnits(area, true)
}

func (f *StateFunctions) findUnitsNoexpand(area Rect) iter.Seq[*Unit] {
	return f.unitFinderSearchUnits(area, false)
}

func (f *StateFunctions) findUnit(area Rect, predicate func(*Unit) bool) *Unit {
	for u := range f.findUnits(area) {
		if predicate(u) {
			return u
		}
	}
	return nil
}

func (f *StateFunctions) findUnitNoexpand(area Rect, predicate func(*Unit) bool) *Unit {
	for u := range f.findUnitsNoexpand(area) {
		if predicate(u) {
			return u
		}
	}
	return nil
}

type nearestUnitSearch struct {
	f            *StateFunctions
	pos          Xy
	searchArea   Rect
	predicate    func(*Unit) bool
	bestDistance int32
	bestUnit     *Unit
}

func (n *nearestUnitSearch) consider(u *Unit) {
	d := n.f.xyLength(XySub(n.pos, u.Sprite.Position))
	if d < n.bestDistance && n.predicate(u) {
		n.bestDistance = d
		n.bestUnit = u
	}
}

func (n *nearestUnitSearch) yInArea(u *Unit) bool {
	return u.Sprite.Position.Y >= n.searchArea.From.Y && u.Sprite.Position.Y < n.searchArea.To.Y
}

func (n *nearestUnitSearch) xInArea(u *Unit) bool {
	return u.Sprite.Position.X >= n.searchArea.From.X && u.Sprite.Position.X < n.searchArea.To.X
}

func (n *nearestUnitSearch) shrinkArea() {
	pos := n.pos
	d := n.bestDistance
	if n.searchArea.From.X < pos.X-d {
		n.searchArea.From.X = pos.X - d
	}
	if n.searchArea.From.Y < pos.Y-d {
		n.searchArea.From.Y = pos.Y - d
	}
	if n.searchArea.To.X > pos.X+d {
		n.searchArea.To.X = pos.X + d
	}
	if n.searchArea.To.Y > pos.Y+d {
		n.searchArea.To.Y = pos.Y + d
	}
}

func (f *StateFunctions) findNearestUnitFromIndices(pos Xy, searchArea Rect, leftI uint, upI uint, rightI uint, downI uint, predicate func(*Unit) bool) *Unit {
	xVec := f.st.UnitFinderX
	yVec := f.st.UnitFinderY
	xEnd := uint(len(xVec))
	yEnd := uint(len(yVec))

	n := &nearestUnitSearch{f: f, pos: pos, searchArea: searchArea, predicate: predicate}
	n.bestDistance = f.xyLength(Xy{
		X: max(pos.X-searchArea.From.X, searchArea.To.X-pos.X),
		Y: max(pos.Y-searchArea.From.Y, searchArea.To.Y-pos.Y),
	})

	for {
		done := true
		prevBestDistance := n.bestDistance
		if leftI != 0 {
			leftI--
			done = false
			u := xVec[leftI].U
			if u.Sprite.Position.X >= n.searchArea.From.X {
				if n.yInArea(u) {
					n.consider(u)
				}
			} else {
				leftI = 0
			}
		}
		if rightI != xEnd {
			done = false
			u := xVec[rightI].U
			if u.Sprite.Position.X < n.searchArea.To.X {
				if n.yInArea(u) {
					n.consider(u)
				}
				rightI++
			} else {
				rightI = xEnd
			}
		}
		if upI != 0 {
			done = false
			upI--
			u := yVec[upI].U
			if u.Sprite.Position.Y >= n.searchArea.From.Y {
				if n.xInArea(u) {
					n.consider(u)
				}
			} else {
				upI = 0
			}
		}
		if downI != yEnd {
			done = false
			u := yVec[downI].U
			if u.Sprite.Position.Y < n.searchArea.To.Y {
				if n.xInArea(u) {
					n.consider(u)
				}
				downI++
			} else {
				downI = yEnd
			}
		}
		if n.bestDistance != prevBestDistance {
			n.shrinkArea()
		}
		if done {
			break
		}
	}
	return n.bestUnit
}

func (f *StateFunctions) findNearestUnit(pos Xy, searchArea Rect, predicate func(*Unit) bool) *Unit {
	xI := uint(unitFinderLowerBound(f.st.UnitFinderX, pos.X))
	yI := uint(unitFinderLowerBound(f.st.UnitFinderY, pos.Y))
	return f.findNearestUnitFromIndices(pos, searchArea, xI, yI, xI, yI, predicate)
}

func (f *StateFunctions) findNearestUnitToUnit(u *Unit, searchArea Rect, predicate func(*Unit) bool) *Unit {
	if f.usHidden(&u.Thingy) {
		return f.findNearestUnit(u.Sprite.Position, searchArea, predicate)
	}
	bb := u.UnitFinderBoundingBox
	leftI := uint(unitFinderIndexOf(f.st.UnitFinderX, bb.To.X, u))
	upI := uint(unitFinderIndexOf(f.st.UnitFinderY, bb.To.Y, u))
	rightI := uint(unitFinderIndexOf(f.st.UnitFinderX, bb.From.X, u)) + 1
	downI := uint(unitFinderIndexOf(f.st.UnitFinderY, bb.From.Y, u)) + 1
	return f.findNearestUnitFromIndices(u.Sprite.Position, searchArea, leftI, upI, rightI, downI, predicate)
}

func (f *StateFunctions) isIntersecting(a Rect, b Rect) bool {
	if b.From.X > a.To.X {
		return false
	}
	if b.From.Y > a.To.Y {
		return false
	}
	if a.From.X > b.To.X {
		return false
	}
	if a.From.Y > b.To.Y {
		return false
	}
	return true
}

func (f *StateFunctions) unitFinderUnitsIntersecting(a *Unit, b *Unit) bool {
	return f.isIntersecting(a.UnitFinderBoundingBox, b.UnitFinderBoundingBox)
}

func (f *StateFunctions) unitFinderUnitInBounds(u *Unit, bounds Rect) bool {
	return f.isIntersecting(u.UnitFinderBoundingBox, bounds)
}

func (f *StateFunctions) checkUnitCollision(u *Unit) {
	for nu := range f.findUnits(f.unitSpriteInnerBoundingBox(u)) {
		if f.uGroundedBuilding(nu) {
			f.uSetStatusFlag(u, UnitStatusFlagCollision)
		} else if !f.uFlying(nu) && (!f.uGathering(nu) || f.uGroundedBuilding(u)) {
			if f.unitFinderUnitsIntersecting(u, nu) {
				f.uSetStatusFlag(nu, UnitStatusFlagCollision)
			}
		}
	}
}
