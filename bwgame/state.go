package bwgame

type Force struct {
	Name  string
	Flags uint8
}

type GameState struct {
	MapTileWidth  uint
	MapTileHeight uint
	MapWalkWidth  uint
	MapWalkHeight uint
	MapWidth      uint
	MapHeight     uint

	MapStrings []string

	ScenarioName        string
	ScenarioDescription string

	UnitAirStrength    [UnitTypesNone]int32
	UnitGroundStrength [UnitTypesNone]int32

	Forces [4]Force

	SightValues [12]SightValues

	TilesetIndex uint

	GfxTiles      []TileID
	Cv5           []Cv5Entry
	Vf4           []Vf4Entry
	MegaTileFlags []uint16

	UnitTypes    UnitTypesT
	WeaponTypes  WeaponTypesT
	UpgradeTypes UpgradeTypesT
	TechTypes    TechTypesT

	UnitTypeAllowed  [12][UnitTypesNone]bool
	MaxUpgradeLevels [12][UpgradeTypesNone]int32
	TechAvailable    [12][TechTypesNone]bool

	StartLocations [12]Xy

	MaxUnitWidth  int32
	MaxUnitHeight int32

	RepulseFieldWidth  uint
	RepulseFieldHeight uint

	Regions Regions

	Triggers []Trigger

	Rules Rules
}

func NewGameState() *GameState {
	return &GameState{Regions: NewRegions(), Rules: DefaultRules()}
}

type StateBaseCopyable struct {
	Global *GlobalState
	Game   *GameState

	UpdateTilesCountdown int32

	OrderTimerCounter          int32
	SecondaryOrderTimerCounter int32
	CurrentFrame               int32

	Players [12]Player

	Alliances [12][12]int32

	UpgradeLevels    [12][UpgradeTypesNone]int32
	UpgradeUpgrading [12][UpgradeTypesNone]bool
	TechResearched   [12][TechTypesNone]bool
	TechResearching  [12][TechTypesNone]bool

	UnitCounters

	TotalBuildingsEverCompleted    [12]int32
	TotalNonBuildingsEverCompleted [12]int32

	UnitScore     [12]int32
	BuildingScore [12]int32

	SupplyUsed      [12][3]Fp1
	SupplyAvailable [12][3]Fp1

	SharedVision [12]uint32

	Tiles              []Tile
	TilesMegaTileIndex []uint16

	RandomCounts      [0x100]int32
	TotalRandomCounts int32
	LcgRandState      uint32

	LastError int32

	TriggerTimer      int32
	RunningTriggers   [8][]RunningTrigger
	TriggerWaitTimers [12]int32
	TriggerWaiting    [12]bool

	ActiveOrdersSize   uint
	ActiveBulletsSize  uint
	ActiveThingiesSize uint

	RepulseField []uint8

	PrevBulletHeadingOffsetClockwise bool

	CurrentMinerals       [12]int32
	CurrentGas            [12]int32
	TotalMineralsGathered [12]int32
	TotalGasGathered      [12]int32

	RecentLurkerHits            [32]StaticVector[Pair[uint, uint], [16]Pair[uint, uint]]
	RecentLurkerHitCurrentIndex uint

	CreepLife             CreepLife
	UpdatePsionicMatrix   bool
	DisruptionWebbedUnits int32
	CheatsEnabled         bool
	CheatOperationCwal    bool

	Locations []Location
}

type PsionicMatrixLinkF struct{}

func (PsionicMatrixLinkF) Link(u *Unit) *Link[Unit] { return &u.Building.Pylon.PsionicMatrixLink }

type PsionicMatrixList = IntrusiveList[Unit, PsionicMatrixLinkF]

type UnitFinderEntry struct {
	U     *Unit
	Value int32
}

type StateBaseNonCopyable struct {
	VisibleUnits     UnitList
	HiddenUnits      UnitList
	MapRevealerUnits UnitList
	DeadUnits        UnitList

	PlayerUnits        [12]PlayerUnitsList
	CloakedUnits       CloakedUnitList
	PsionicMatrixUnits PsionicMatrixList

	UnitsContainer ObjectContainer[Unit, *Unit]

	ActiveBullets    BulletList
	BulletsContainer ObjectContainer[Bullet, *Bullet]

	SpritesOnTileLine []SpriteList
	SpritesContainer  ObjectContainer[Sprite, *Sprite]

	ImagesContainer ObjectContainer[Image, *Image]

	OrdersContainer ObjectContainer[Order, *Order]

	FreePaths PathList
	Paths     []*Path

	ActiveThingies ThingyList
	FreeThingies   ThingyList
	Thingies       []*Thingy

	UnitFinderX []UnitFinderEntry
	UnitFinderY []UnitFinderEntry

	unitFinderSearchGenerations [4]uint32

	ConsiderCollisionWithUnitBug *Unit
	PrevBulletSourceUnit         *Unit
}

type State struct {
	StateBaseCopyable
	StateBaseNonCopyable
}

func NewState(global *GlobalState, game *GameState) *State {
	st := &State{}
	st.Global = global
	st.Game = game
	st.CreepLife = NewCreepLife()
	st.newObjectContainers()
	return st
}

func (st *State) newObjectContainers() {
	limits := DefaultRules().ObjectLimits()
	if st.Game != nil {
		limits = st.Game.Rules.ObjectLimits()
	}
	st.UnitsContainer = NewObjectContainer[Unit](UnitLimit, 17)
	st.BulletsContainer = NewObjectContainer[Bullet](limits.Bullets, 10)
	st.SpritesContainer = NewObjectContainer[Sprite](limits.Sprites, 25)
	st.ImagesContainer = NewObjectContainer[Image](limits.Images, 50)
	st.OrdersContainer = NewObjectContainer[Order](limits.Orders, 20)
}
