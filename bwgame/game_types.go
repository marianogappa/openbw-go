package bwgame

import (
	"iter"
	"unsafe"
)

type unitIDRaw interface {
	~uint16 | ~uint32
}

type UnitIDT[T unitIDRaw] struct {
	RawValue T
}

type UnitID = UnitIDT[uint16]
type UnitID32 = UnitIDT[uint32]

func NewUnitID(index uint, generation uint32) UnitID {
	return UnitID{RawValue: uint16(index | uint(generation<<11))}
}

func NewUnitID32(index uint, generation uint32) UnitID32 {
	return UnitID32{RawValue: uint32(index | uint(generation<<11))}
}

func (id UnitIDT[T]) Index() uint {
	return uint(id.RawValue & 0x7ff)
}

func (id UnitIDT[T]) Generation() uint32 {
	return uint32(id.RawValue >> 11)
}

type Link[T any] struct {
	Prev *T
	Next *T
}

type LinkAccessor[T any] interface {
	Link(*T) *Link[T]
}

type IntrusiveList[T any, A LinkAccessor[T]] struct {
	first             *T
	last              *T
	linkOffsetPlusOne uintptr
}

func (l *IntrusiveList[T, A]) link(v *T) *Link[T] {
	if l.linkOffsetPlusOne == 0 {
		l.linkOffsetPlusOne = linkOffset[T, A](v) + 1
	}
	return (*Link[T])(unsafe.Add(unsafe.Pointer(v), l.linkOffsetPlusOne-1))
}

func linkOffset[T any, A LinkAccessor[T]](v *T) uintptr {
	var accessor A
	return uintptr(unsafe.Pointer(accessor.Link(v))) - uintptr(unsafe.Pointer(v))
}

func (l *IntrusiveList[T, A]) Empty() bool {
	return l.first == nil
}

func (l *IntrusiveList[T, A]) Front() *T {
	return l.first
}

func (l *IntrusiveList[T, A]) Back() *T {
	return l.last
}

func (l *IntrusiveList[T, A]) Begin() *T {
	return l.first
}

func (l *IntrusiveList[T, A]) End() *T {
	return nil
}

func (l *IntrusiveList[T, A]) Next(v *T) *T {
	if v == nil {
		return l.first
	}
	return l.link(v).Next
}

func (l *IntrusiveList[T, A]) Prev(v *T) *T {
	if v == nil {
		return l.last
	}
	return l.link(v).Prev
}

func (l *IntrusiveList[T, A]) Clear() {
	l.first = nil
	l.last = nil
}

func (l *IntrusiveList[T, A]) setNextOf(v *T, next *T) {
	if v == nil {
		l.first = next
		return
	}
	l.link(v).Next = next
}

func (l *IntrusiveList[T, A]) setPrevOf(v *T, prev *T) {
	if v == nil {
		l.last = prev
		return
	}
	l.link(v).Prev = prev
}

func (l *IntrusiveList[T, A]) Insert(pos *T, v *T) *T {
	before := pos
	after := l.Prev(before)
	l.setPrevOf(before, v)
	l.setNextOf(after, v)
	vLink := l.link(v)
	vLink.Prev = after
	vLink.Next = before
	return v
}

func (l *IntrusiveList[T, A]) Erase(pos *T) *T {
	posLink := l.link(pos)
	prev := posLink.Prev
	next := posLink.Next
	l.setNextOf(prev, next)
	l.setPrevOf(next, prev)
	return next
}

func (l *IntrusiveList[T, A]) Remove(v *T) {
	l.Erase(v)
}

func (l *IntrusiveList[T, A]) PushBack(v *T) {
	l.Insert(nil, v)
}

func (l *IntrusiveList[T, A]) PopBack() {
	l.Erase(l.last)
}

func (l *IntrusiveList[T, A]) PushFront(v *T) {
	l.Insert(l.first, v)
}

func (l *IntrusiveList[T, A]) PopFront() {
	l.Erase(l.first)
}

func (l *IntrusiveList[T, A]) Swap(n *IntrusiveList[T, A]) {
	l.first, n.first = n.first, l.first
	l.last, n.last = n.last, l.last
}

func (l *IntrusiveList[T, A]) All() iter.Seq[*T] {
	return func(yield func(*T) bool) {
		for v := l.first; v != nil; v = l.link(v).Next {
			if !yield(v) {
				return
			}
		}
	}
}

func (l *IntrusiveList[T, A]) Backward() iter.Seq[*T] {
	return func(yield func(*T) bool) {
		first := l.first
		if first == nil {
			return
		}
		for v := l.last; v != nil; v = l.link(v).Prev {
			if !yield(v) || v == first {
				return
			}
		}
	}
}

func bwInsertList[T any, A LinkAccessor[T]](l *IntrusiveList[T, A], v *T) {
	if l.Empty() {
		l.PushFront(v)
		return
	}
	l.Insert(l.Next(l.Begin()), v)
}

type defaultLinked[T any] interface {
	*T
	DefaultLink() *Link[T]
}

type DefaultLinkF[T any, PT defaultLinked[T]] struct{}

func (DefaultLinkF[T, PT]) Link(v *T) *Link[T] {
	return PT(v).DefaultLink()
}

type containerObject[T any] interface {
	defaultLinked[T]
	ContainerIndex() *uint
}

type ObjectContainer[T any, PT containerObject[T]] struct {
	List     [][]T
	FreeList IntrusiveList[T, DefaultLinkF[T, PT]]
	Size     uint

	MaxSize               uint
	AllocationGranularity uint
}

func NewObjectContainer[T any, PT containerObject[T]](maxSize, allocationGranularity uint) ObjectContainer[T, PT] {
	return ObjectContainer[T, PT]{MaxSize: maxSize, AllocationGranularity: allocationGranularity}
}

func (c *ObjectContainer[T, PT]) storageIndex(index uint) uint {
	if index != 0 {
		return c.MaxSize - index
	}
	return index
}

func (c *ObjectContainer[T, PT]) slot(storageIndex uint) *T {
	return &c.List[storageIndex/c.AllocationGranularity][storageIndex%c.AllocationGranularity]
}

func (c *ObjectContainer[T, PT]) Get(index uint, addNewToFree bool) *T {
	index = c.storageIndex(index)
	for c.Size <= index {
		c.Grow(addNewToFree)
	}
	return c.slot(index)
}

func (c *ObjectContainer[T, PT]) TryGet(index uint) *T {
	index = c.storageIndex(index)
	if c.Size <= index {
		return nil
	}
	return c.slot(index)
}

func (c *ObjectContainer[T, PT]) At(index uint) *T {
	index = c.storageIndex(index)
	if c.Size <= index {
		throwError("object_container::get const: invalid index %d", index)
	}
	return c.slot(index)
}

func (c *ObjectContainer[T, PT]) Grow(addNewToFree bool) {
	if c.Size == c.MaxSize {
		throwError("object_container: attempt to grow beyond max_size")
	}
	chunk := make([]T, c.AllocationGranularity)
	c.List = append(c.List, chunk)
	n := min(c.AllocationGranularity, c.MaxSize-c.Size)
	for i := uint(0); i != n; i++ {
		obj := &chunk[i]
		*PT(obj).ContainerIndex() = c.newObjectIndex()
		if addNewToFree {
			c.FreeList.PushBack(obj)
		}
		c.Size++
	}
}

func (c *ObjectContainer[T, PT]) newObjectIndex() uint {
	if c.Size == 0 {
		return 0
	}
	return c.MaxSize - c.Size
}

func (c *ObjectContainer[T, PT]) Top() *T {
	if c.FreeList.Empty() {
		if c.Size == c.MaxSize {
			return nil
		}
		c.Grow(true)
	} else if c.Size != c.MaxSize && c.FreeList.Next(c.FreeList.Begin()) == c.FreeList.End() {
		c.Grow(true)
	}
	r := c.FreeList.Front()
	rIndex := *PT(r).ContainerIndex()
	if r != c.Get(rIndex, true) {
		throwError("index mismatch for %d\n", rIndex)
	}
	return r
}

func (c *ObjectContainer[T, PT]) Pop() {
	c.FreeList.PopFront()
}

func (c *ObjectContainer[T, PT]) Push(obj *T) {
	bwInsertList(&c.FreeList, obj)
}

type Race int32

const (
	RaceZerg Race = iota
	RaceTerran
	RaceProtoss
	RaceNone
)

const (
	PlayerControllerInactive = iota
	PlayerControllerComputerGame
	PlayerControllerOccupied
	PlayerControllerRescuePassive
	PlayerControllerUnusedRescueActive
	PlayerControllerComputer
	PlayerControllerOpen
	PlayerControllerNeutral
	PlayerControllerClosed
	PlayerControllerUnusedObserver
	PlayerControllerUserLeft
	PlayerControllerComputerDefeated
)

type Player struct {
	Controller int32
	Race       Race
	Force      int32
	Color      int32

	InitiallyActive bool
	VictoryState    int32
}

type SightValuesMaskdatNode struct {
	Prev              uint
	Prev2             uint
	RelativeTileIndex int32
	X                 int32
	Y                 int32
}

type SightValues struct {
	MaxWidth       int32
	MaxHeight      int32
	MinWidth       int32
	MinHeight      int32
	MinMaskSize    int32
	ExtMaskedCount int32
	Maskdat        []SightValuesMaskdatNode
}

type TriggerCondition struct {
	Location int32
	Group    int32
	CountN   int32
	UnitID   int32
	NumN     int32
	Type     int32
	ExtraN   int32
	Flags    int32
	Unk      int32
}

type TriggerAction struct {
	Location    int32
	StringIndex int32
	SoundIndex  int32
	TimeN       int32
	GroupN      int32
	Group2N     int32
	ExtraN      int32
	Type        int32
	NumN        int32
	Flags       int32
	Unk         int32
}

type Trigger struct {
	Conditions     [16]TriggerCondition
	Actions        [64]TriggerAction
	ExecutionFlags int32
	Enabled        [28]bool
}

type RunningTriggerAction struct {
	Flags int32
}

type RunningTrigger struct {
	Actions            [64]RunningTriggerAction
	T                  *Trigger
	Flags              int32
	CurrentActionIndex uint
}

type Location struct {
	Area           Rect
	ElevationFlags int32
}

type Cv5Entry struct {
	Flags         uint16
	MegaTileIndex [16]uint16
}

const (
	Vf4EntryFlagWalkable uint16 = 1
	Vf4EntryFlagMiddle   uint16 = 2
	Vf4EntryFlagHigh     uint16 = 4
	Vf4EntryFlagVeryHigh uint16 = 8
)

type Vf4Entry struct {
	Flags [16]uint16
}

type TileID struct {
	RawValue uint16
}

func NewTileID(groupIndex uint, subtileIndex uint) TileID {
	return TileID{RawValue: uint16(groupIndex<<4 | subtileIndex)}
}

func (t TileID) HasCreep() bool {
	return (t.RawValue>>4)&0x8000 != 0
}

func (t TileID) GroupIndex() uint {
	return uint((t.RawValue >> 4) & 0x7ff)
}

func (t TileID) SubtileIndex() uint {
	return uint(t.RawValue & 0xf)
}

func (t TileID) Valid() bool {
	return t.RawValue != 0
}

const (
	TileFlagWalkable          = 1
	TileFlagUnk0              = 2
	TileFlagUnwalkable        = 4
	TileFlagUnk1              = 8
	TileFlagProvidesCover     = 0x10
	TileFlagUnk3              = 0x20
	TileFlagHasCreep          = 0x40
	TileFlagUnbuildable       = 0x80
	TileFlagVeryHigh          = 0x100
	TileFlagMiddle            = 0x200
	TileFlagHigh              = 0x400
	TileFlagOccupied          = 0x800
	TileFlagCreepReceding     = 0x1000
	TileFlagPartiallyWalkable = 0x2000
	TileFlagTemporaryCreep    = 0x4000
	TileFlagUnk4              = 0x8000
)

type Tile struct {
	Visible  uint8
	Explored uint8
	Flags    uint16
}

type Region struct {
	Flags                uint16
	Index                uint
	TileCenter           XyT[uint]
	TileArea             RectT[XyT[uint]]
	Center               XyFp8
	Area                 Rect
	TileCount            uint
	GroupIndex           uint
	WalkableNeighbors    []*Region
	NonWalkableNeighbors []*Region

	PathfinderFlag int32
	PathfinderNode any
}

func NewRegion() Region {
	return Region{Flags: 0x1FFD, Index: ^uint(0)}
}

func (r *Region) Walkable() bool {
	return r.Flags != 0x1ffd
}

type SplitRegion struct {
	Mask uint16
	A    *Region
	B    *Region
}

type Contour struct {
	V     [3]int32
	Dir   uint
	Flags uint8
}

type Regions struct {
	TileRegionIndex []uint

	TileBoundingBox RectT[XyT[uint]]

	Regions []Region

	SplitRegions []SplitRegion

	Contours [4][]Contour

	contourIndexes [4]contourIndex
}

func NewRegions() Regions {
	return Regions{TileRegionIndex: make([]uint, 256*256)}
}

type CreepLifeEntry struct {
	HashLink               Link[CreepLifeEntry]
	ListLink               Link[CreepLifeEntry]
	TilePos                XyT[uint]
	NNeighboringCreepTiles uint
}

type CreepEntryHashLinkF struct{}

func (CreepEntryHashLinkF) Link(e *CreepLifeEntry) *Link[CreepLifeEntry] {
	return &e.HashLink
}

type CreepEntryListLinkF struct{}

func (CreepEntryListLinkF) Link(e *CreepLifeEntry) *Link[CreepLifeEntry] {
	return &e.ListLink
}

type CreepEntryHashList = IntrusiveList[CreepLifeEntry, CreepEntryHashLinkF]
type CreepEntryList = IntrusiveList[CreepLifeEntry, CreepEntryListLinkF]

type CreepLifeEntryHashTable struct {
	Buckets [0x200]CreepEntryHashList
}

func (t *CreepLifeEntryHashTable) bucketIndex(tilePos XyT[uint]) uint {
	return (tilePos.Y*7 + tilePos.X) % uint(len(t.Buckets))
}

func (t *CreepLifeEntryHashTable) Find(tilePos XyT[uint]) *CreepLifeEntry {
	for v := range t.Buckets[t.bucketIndex(tilePos)].All() {
		if v.TilePos == tilePos {
			return v
		}
	}
	return nil
}

func (t *CreepLifeEntryHashTable) Remove(v *CreepLifeEntry) {
	t.Buckets[t.bucketIndex(v.TilePos)].Remove(v)
}

func (t *CreepLifeEntryHashTable) Insert(v *CreepLifeEntry) {
	t.Buckets[t.bucketIndex(v.TilePos)].PushFront(v)
}

type CreepLife struct {
	RecedeTimer        int32
	CheckDeadUnitTimer int32

	Lists        [9]CreepEntryList
	ListsSize    [9]uint
	FreeList     CreepEntryList
	FreeListSize uint
	Table        CreepLifeEntryHashTable

	EntryContainer []CreepLifeEntry
}

func NewCreepLife() CreepLife {
	c := CreepLife{EntryContainer: make([]CreepLifeEntry, 1024)}
	for i := range c.EntryContainer {
		c.FreeList.PushBack(&c.EntryContainer[i])
		c.FreeListSize++
	}
	return c
}

type Target struct {
	Pos  Xy
	Unit *Unit
}

type IscriptState struct {
	CurrentScript  *IscriptTScript
	ProgramCounter uint
	ReturnAddress  uint
	Animation      int32
	Wait           int32
}

const (
	ImageFlagRedraw               = 1
	ImageFlagHorizontallyFlipped  = 2
	ImageFlagYFrozen              = 4
	ImageFlagHasDirectionalFrames = 8
	ImageFlagHasIscriptAnimations = 0x10
	ImageFlagClickable            = 0x20
	ImageFlagHidden               = 0x40
	ImageFlagUsesSpecialOffset    = 0x80
)

type Image struct {
	Link Link[Image]

	Index            uint
	ImageType        *ImageTypeT
	Modifier         int32
	FrameIndexOffset uint
	Flags            int32
	Offset           Xy
	IscriptState     IscriptState
	FrameIndexBase   uint
	FrameIndex       uint
	Grp              *GrpT
	ModifierData1    int32
	ModifierData2    int32
	Sprite           *Sprite
	FrozenYValue     int32
}

func (i *Image) DefaultLink() *Link[Image] { return &i.Link }
func (i *Image) ContainerIndex() *uint     { return &i.Index }

type ImageList = IntrusiveList[Image, DefaultLinkF[Image, *Image]]

const (
	SpriteFlagSelected     = 0x8
	SpriteFlagTurret       = 0x10
	SpriteFlagHidden       = 0x20
	SpriteFlagBurrowed     = 0x40
	SpriteFlagIscriptNobrk = 0x80
)

type Sprite struct {
	Link Link[Sprite]

	Index           uint
	SpriteType      *SpriteTypeT
	Owner           int32
	SelectionIndex  int32
	VisibilityFlags int32
	ElevationLevel  int32
	Flags           int32
	SelectionTimer  int32
	Width           uint
	Height          uint
	Position        Xy
	MainImage       *Image
	Images          ImageList
}

func (s *Sprite) DefaultLink() *Link[Sprite] { return &s.Link }
func (s *Sprite) ContainerIndex() *uint      { return &s.Index }

type SpriteList = IntrusiveList[Sprite, DefaultLinkF[Sprite, *Sprite]]

type Thingy struct {
	Link Link[Thingy]

	Hp     Fp8
	Sprite *Sprite
}

func (t *Thingy) DefaultLink() *Link[Thingy] { return &t.Link }

type ThingyList = IntrusiveList[Thingy, DefaultLinkF[Thingy, *Thingy]]

type Flingy struct {
	Thingy

	Index                    uint
	MoveTarget               Target
	NextMovementWaypoint     Xy
	NextTargetWaypoint       Xy
	MovementFlags            int32
	Heading                  DirectionT
	FlingyTurnRate           Fp8
	NextVelocityDirection    DirectionT
	FlingyType               *FlingyTypeT
	FlingyMovementType       int32
	Position                 Xy
	ExactPosition            XyFp8
	FlingyTopSpeed           Fp8
	CurrentSpeed             Fp8
	NextSpeed                Fp8
	Velocity                 XyFp8
	FlingyAcceleration       Fp8
	CurrentVelocityDirection DirectionT
	DesiredVelocityDirection DirectionT
	OrderSignal              int32
}

const (
	BulletStateInit = iota
	BulletStateMove
	BulletStateFollow
	BulletStateBounce
	BulletStateDamageOverTime
	BulletStateDying
	BulletStateHitNearTarget
)

type Bullet struct {
	Flingy
	Link Link[Bullet]

	Index                      uint
	BulletState                int32
	BulletTarget               *Unit
	BulletTargetPos            Xy
	WeaponType                 *WeaponTypeT
	RemainingTime              int32
	HitFlags                   int32
	RemainingBounces           int32
	Owner                      int32
	BulletOwnerUnit            *Unit
	PrevBounceUnit             *Unit
	HitNearTargetPositionIndex uint
}

func (b *Bullet) DefaultLink() *Link[Bullet] { return &b.Link }
func (b *Bullet) ContainerIndex() *uint      { return &b.Index }

type BulletList = IntrusiveList[Bullet, DefaultLinkF[Bullet, *Bullet]]

type OrderTarget struct {
	Position Xy
	Unit     *Unit
	UnitType *UnitTypeT
}

type Order struct {
	Link Link[Order]

	Index     uint
	OrderType *OrderTypeT
	Target    OrderTarget
}

func (o *Order) DefaultLink() *Link[Order] { return &o.Link }
func (o *Order) ContainerIndex() *uint     { return &o.Index }

type OrderList = IntrusiveList[Order, DefaultLinkF[Order, *Order]]

type Path struct {
	Link Link[Path]

	Delay         int32
	CreationFrame int32
	StateFlags    int32

	LongPath         []*Region
	FullLongPathSize uint
	ShortPath        []Xy

	CurrentLongPathIndex  uint
	CurrentShortPathIndex uint

	Source      Xy
	Destination Xy
	Next        Xy

	LastCollisionUnit  UnitID
	LastCollisionSpeed Fp8
	SlideFreeDirection DirectionT
}

func (p *Path) DefaultLink() *Link[Path] { return &p.Link }

type PathList = IntrusiveList[Path, DefaultLinkF[Path, *Path]]

const (
	UnitStatusFlagCompleted                   = 1
	UnitStatusFlagGroundedBuilding            = 2
	UnitStatusFlagFlying                      = 4
	UnitStatusFlag8                           = 8
	UnitStatusFlagBurrowed                    = 0x10
	UnitStatusFlagInBunker                    = 0x20
	UnitStatusFlagLoaded                      = 0x40
	UnitStatusFlagRequiresDetector            = 0x100
	UnitStatusFlagCloaked                     = 0x200
	UnitStatusFlagDisabled                    = 0x400
	UnitStatusFlagPassivelyCloaked            = 0x800
	UnitStatusFlagOrderNotInterruptible       = 0x1000
	UnitStatusFlagIscriptNobrk                = 0x2000
	UnitStatusFlag4000                        = 0x4000
	UnitStatusFlagCannotAttack                = 0x8000
	UnitStatusFlagCanTurn                     = 0x10000
	UnitStatusFlagCanMove                     = 0x20000
	UnitStatusFlagCollision                   = 0x40000
	UnitStatusFlagImmovable                   = 0x80000
	UnitStatusFlagGroundUnit                  = 0x100000
	UnitStatusFlagNoCollide                   = 0x200000
	UnitStatusFlag400000                      = 0x400000
	UnitStatusFlagGathering                   = 0x800000
	UnitStatusFlagTurretWalking               = 0x1000000
	UnitStatusFlagInvincible                  = 0x4000000
	UnitStatusFlagReadyToAttack               = 0x8000000
	UnitStatusFlagSpeedUpgrade                = 0x10000000
	UnitStatusFlagCooldownUpgrade             = 0x20000000
	UnitStatusFlagHallucination               = 0x40000000
	UnitStatusFlagLifetimeExpired       int32 = -0x80000000
)

type UnitVulture struct {
	SpiderMineCount uint
}

type UnitFighter struct {
	Parent      *Unit
	FighterLink Link[Unit]
	IsOutside   bool
}

type UnitCarrier struct {
	InsideUnits  FighterList
	OutsideUnits FighterList
	InsideCount  uint
	OutsideCount uint
}

type UnitBeacon struct {
	Unknown00      int32
	Unknown04      int32
	FlagSpawnFrame int32
}

type UnitGhost struct {
	NukeDot *Thingy
}

type UnitWorker struct {
	Powerup                *Unit
	TargetResourcePosition Xy
	TargetResourceUnit     *Unit
	RepairTimer            int32
	IsGathering            bool
	ResourcesCarried       int32
	GatherTarget           *Unit
	GatherLink             Link[Unit]
}

type UnitBuildingResource struct {
	ResourceCount   int32
	ResourceIscript int32
	IsBeingGathered bool
	GatherQueue     WorkerGatherList
}

type UnitBuildingNydus struct {
	Exit *Unit
}

type UnitBuildingPylon struct {
	PsiFieldSprite    *Sprite
	PsionicMatrixLink Link[Unit]
}

type UnitBuildingSilo struct {
	Nuke  *Unit
	Ready bool
}

type UnitBuildingPowerup struct {
	Origin Xy
}

type UnitBuildingHatchery struct {
	LarvaSpawnSideValues [4]int32
}

type UnitBuilding struct {
	Addon               *Unit
	AddonBuildType      *UnitTypeT
	UpgradeResearchTime int32
	ResearchingType     *TechTypeT
	UpgradingType       *UpgradeTypeT
	LarvaTimer          int32
	IsLanding           bool
	CreepTimer          int32
	UpgradingLevel      int32
	Rally               Target

	Resource UnitBuildingResource
	Nydus    UnitBuildingNydus
	Pylon    UnitBuildingPylon
	Silo     UnitBuildingSilo
	Powerup  UnitBuildingPowerup
	Hatchery UnitBuildingHatchery
}

type Unit struct {
	Flingy
	Link Link[Unit]

	Owner                int32
	OrderType            *OrderTypeT
	OrderState           int32
	OrderUnitType        *UnitTypeT
	MainOrderTimer       int32
	GroundWeaponCooldown int32
	AirWeaponCooldown    int32
	SpellCooldown        int32
	OrderTarget          Target

	ShieldPoints Fp8
	UnitType     *UnitTypeT

	PlayerUnitsLink Link[Unit]

	Subunit                *Unit
	OrderQueue             OrderList
	AutoTargetUnit         *Unit
	ConnectedUnit          *Unit
	OrderQueueCount        int32
	OrderProcessTimer      int32
	Unknown0x086           int32
	AttackNotifyTimer      int32
	PreviousUnitType       *UnitTypeT
	LastEventTimer         int32
	LastEventColor         int32
	RankIncrease           int32
	KillCount              int32
	LastAttackingPlayer    int32
	SecondaryOrderTimer    int32
	UserActionFlags        int32
	CloakCounter           int32
	MovementState          int32
	BuildQueue             StaticVector[*UnitTypeT, [5]*UnitTypeT]
	BuildQueueLimbo        StaticVector[*UnitTypeT, [5]*UnitTypeT]
	Energy                 Fp8
	UnitIDGeneration       uint32
	SecondaryOrderType     *OrderTypeT
	DamageOverlayState     int32
	HpConstructionRate     Fp8
	ShieldConstructionRate Fp8
	RemainingBuildTime     int32
	PreviousHp             int32
	LoadedUnits            [8]UnitID

	Vulture UnitVulture
	Fighter UnitFighter
	Carrier UnitCarrier
	Beacon  UnitBeacon
	Ghost   UnitGhost

	Worker UnitWorker

	Building UnitBuilding

	StatusFlags         int32
	CarryingFlags       int32
	WireframeRandomizer int32
	SecondaryOrderState int32
	MoveTargetTimer     int32
	DetectedFlags       uint32
	CurrentBuildUnit    *Unit
	CloakedUnitLink     Link[Unit]

	Path                     *Path
	PathingCollisionCounter  int32
	PathingFlags             int32
	Unused0x106              int32
	IsBeingHealed            bool
	TerrainNoCollisionBounds Rect

	RemoveTimer          int32
	DefensiveMatrixHp    Fp8
	DefensiveMatrixTimer int32
	StimTimer            int32
	EnsnareTimer         int32
	LockdownTimer        int32
	IrradiateTimer       int32
	StasisTimer          int32
	PlagueTimer          int32
	StormTimer           int32
	IrradiatedBy         *Unit
	IrradiateOwner       int32
	ParasiteFlags        int32
	CycleCounter         int32
	BlindedBy            int32
	MaelstromTimer       int32
	AcidSporeCount       int32
	AcidSporeTime        [9]int32

	NextHitNearTargetPositionIndex int32
	AirStrength                    int32
	GroundStrength                 int32
	RepulseFlags                   int32
	RepulseDirection               DirectionT
	RepulseIndex                   uint

	UnitFinderBoundingBox Rect
	UnitFinderVisited     [4]bool
	UnitFinderIndexFrom   uint
	UnitFinderIndexTo     uint
}

func (u *Unit) DefaultLink() *Link[Unit] { return &u.Link }
func (u *Unit) ContainerIndex() *uint    { return &u.Index }

func (u *Unit) Reaver() *UnitCarrier {
	return &u.Carrier
}

type UnitList = IntrusiveList[Unit, DefaultLinkF[Unit, *Unit]]

type PlayerUnitsLinkF struct{}

func (PlayerUnitsLinkF) Link(u *Unit) *Link[Unit] { return &u.PlayerUnitsLink }

type PlayerUnitsList = IntrusiveList[Unit, PlayerUnitsLinkF]

type CloakedUnitLinkF struct{}

func (CloakedUnitLinkF) Link(u *Unit) *Link[Unit] { return &u.CloakedUnitLink }

type CloakedUnitList = IntrusiveList[Unit, CloakedUnitLinkF]

type FighterLinkF struct{}

func (FighterLinkF) Link(u *Unit) *Link[Unit] { return &u.Fighter.FighterLink }

type FighterList = IntrusiveList[Unit, FighterLinkF]

type WorkerGatherLinkF struct{}

func (WorkerGatherLinkF) Link(u *Unit) *Link[Unit] { return &u.Worker.GatherLink }

type WorkerGatherList = IntrusiveList[Unit, WorkerGatherLinkF]
