package bwgame

import (
	"errors"
	"testing"
)

func unitListIndices(l *UnitList) []uint {
	var indices []uint
	for u := range l.All() {
		indices = append(indices, u.Index)
	}
	return indices
}

func equalIndices(a, b []uint) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestIntrusiveListRemoveCurrentDuringRangeContinues(t *testing.T) {
	var units [5]Unit
	var list UnitList
	for i := range units {
		units[i].Index = uint(i)
		list.PushBack(&units[i])
	}
	var visited []uint
	for u := range list.All() {
		visited = append(visited, u.Index)
		if u.Index%2 == 0 {
			list.Remove(u)
		}
	}
	if !equalIndices(visited, []uint{0, 1, 2, 3, 4}) {
		t.Fatalf("visited %v", visited)
	}
	if got := unitListIndices(&list); !equalIndices(got, []uint{1, 3}) {
		t.Fatalf("list %v", got)
	}
}

func TestIntrusiveListsWithDifferentLinksAreIndependent(t *testing.T) {
	var units [3]Unit
	var visible UnitList
	var playerUnits PlayerUnitsList
	for i := range units {
		units[i].Index = uint(i)
		visible.PushBack(&units[i])
		playerUnits.PushFront(&units[i])
	}
	if got := unitListIndices(&visible); !equalIndices(got, []uint{0, 1, 2}) {
		t.Fatalf("visible %v", got)
	}
	var player []uint
	for u := range playerUnits.All() {
		player = append(player, u.Index)
	}
	if !equalIndices(player, []uint{2, 1, 0}) {
		t.Fatalf("player units %v", player)
	}
}

func TestBwInsertListInsertsAfterFirst(t *testing.T) {
	var units [3]Unit
	var list UnitList
	for i := range units {
		units[i].Index = uint(i)
		bwInsertList(&list, &units[i])
	}
	if got := unitListIndices(&list); !equalIndices(got, []uint{0, 2, 1}) {
		t.Fatalf("list %v", got)
	}
}

func TestObjectContainerHandsOutUnitIndicesInBroodWarOrder(t *testing.T) {
	st := NewState(nil, NewGameState())
	var indices []uint
	for i := 0; i != 20; i++ {
		u := st.UnitsContainer.Top()
		st.UnitsContainer.Pop()
		indices = append(indices, u.Index)
	}
	want := []uint{0, 1699, 1698, 1697, 1696, 1695, 1694, 1693, 1692, 1691, 1690, 1689, 1688, 1687, 1686, 1685, 1684, 1683, 1682, 1681}
	if !equalIndices(indices, want) {
		t.Fatalf("indices %v, want %v", indices, want)
	}
	if u := st.UnitsContainer.At(1699); u.Index != 1699 {
		t.Fatalf("At(1699) has index %d", u.Index)
	}
}

func TestNewStateInitialisesCreepLifeAndRegions(t *testing.T) {
	game := NewGameState()
	st := NewState(nil, game)
	if st.Game != game {
		t.Fatal("game state not set")
	}
	if st.CreepLife.FreeListSize != 1024 {
		t.Fatalf("free list size %d", st.CreepLife.FreeListSize)
	}
	n := 0
	for e := range st.CreepLife.FreeList.All() {
		if e != &st.CreepLife.EntryContainer[n] {
			t.Fatalf("free list entry %d out of order", n)
		}
		n++
	}
	if n != 1024 {
		t.Fatalf("free list has %d entries", n)
	}
	if len(game.Regions.TileRegionIndex) != 256*256 {
		t.Fatalf("tile region index size %d", len(game.Regions.TileRegionIndex))
	}
	region := NewRegion()
	if region.Flags != 0x1ffd || region.Index != ^uint(0) || region.Walkable() {
		t.Fatalf("unexpected default region %+v", region)
	}
}

func TestCreepHashTableFindsInsertedEntries(t *testing.T) {
	c := NewCreepLife()
	a := &c.EntryContainer[0]
	b := &c.EntryContainer[1]
	a.TilePos = XyT[uint]{X: 3, Y: 4}
	b.TilePos = XyT[uint]{X: 3 + 0x200, Y: 4}
	c.Table.Insert(a)
	c.Table.Insert(b)
	if c.Table.Find(a.TilePos) != a || c.Table.Find(b.TilePos) != b {
		t.Fatal("hash table lookup failed")
	}
	c.Table.Remove(a)
	if c.Table.Find(a.TilePos) != nil {
		t.Fatal("removed entry still found")
	}
}

func TestStaticVectorPanicsWhenFull(t *testing.T) {
	var queue StaticVector[*UnitTypeT, [5]*UnitTypeT]
	for i := 0; i != 5; i++ {
		queue.PushBack(nil)
	}
	defer func() {
		recovered := recover()
		err, ok := recovered.(error)
		var exception *Exception
		if !ok || !errors.As(err, &exception) {
			t.Fatalf("recovered %v, want *Exception", recovered)
		}
	}()
	queue.PushBack(nil)
}

func TestStaticVectorEraseAndInsertKeepOrder(t *testing.T) {
	var v StaticVector[int32, [16]int32]
	for i := int32(0); i != 5; i++ {
		v.PushBack(i)
	}
	v.Erase(1)
	v.Insert(0, 9)
	want := []int32{9, 0, 2, 3, 4}
	for i, value := range v.Values() {
		if value != want[i] {
			t.Fatalf("values %v, want %v", v.Values(), want)
		}
	}
}

func TestXyOrderingComparesYFirst(t *testing.T) {
	if !XyLess(Xy{X: 5, Y: 1}, Xy{X: 1, Y: 2}) {
		t.Fatal("y must decide first")
	}
	if !XyLess(Xy{X: 1, Y: 2}, Xy{X: 5, Y: 2}) {
		t.Fatal("x must decide on equal y")
	}
	if XyLess(Xy{X: 1, Y: 2}, Xy{X: 1, Y: 2}) || !XyLessEqual(Xy{X: 1, Y: 2}, Xy{X: 1, Y: 2}) {
		t.Fatal("equal values")
	}
}

func TestGetBestScoreKeepsFirstOnTies(t *testing.T) {
	values := []int32{4, 2, 7, 2}
	if got := getBestScore(values, func(v int32) int32 { return v }); got != 1 {
		t.Fatalf("best index %d", got)
	}
	if got := getBestScore([]int32{}, func(v int32) int32 { return v }); got != 0 {
		t.Fatalf("empty best index %d", got)
	}
}

func TestDirectionTableMatchesCardinalDirections(t *testing.T) {
	cases := map[int]XyFp8{
		0:   {X: Fp8FromRaw(0), Y: Fp8FromRaw(-256)},
		64:  {X: Fp8FromRaw(256), Y: Fp8FromRaw(0)},
		128: {X: Fp8FromRaw(0), Y: Fp8FromRaw(256)},
		192: {X: Fp8FromRaw(-256), Y: Fp8FromRaw(0)},
		255: {X: Fp8FromRaw(-6), Y: Fp8FromRaw(-256)},
	}
	for index, want := range cases {
		if directionTable[index] != want {
			t.Fatalf("directionTable[%d] = %+v, want %+v", index, directionTable[index], want)
		}
	}
}

func TestStaticVectorAssignmentCopiesValues(t *testing.T) {
	marine := &UnitTypeT{ID: UnitTypesTerranMarine}
	ghost := &UnitTypeT{ID: UnitTypesTerranGhost}
	var u Unit
	u.BuildQueue.PushBack(marine)
	u.BuildQueue.PushBack(ghost)
	u.BuildQueue.Erase(0)
	u.BuildQueueLimbo = u.BuildQueue
	u.BuildQueue.Clear()
	u.BuildQueue.PushBack(marine)
	if u.BuildQueueLimbo.Size() != 1 || u.BuildQueueLimbo.At(0) != ghost {
		t.Fatalf("limbo = %v, want [ghost]", u.BuildQueueLimbo.Values())
	}
	if u.BuildQueue.Size() != 1 || u.BuildQueue.At(0) != marine {
		t.Fatalf("queue = %v, want [marine]", u.BuildQueue.Values())
	}
}
