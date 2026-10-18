package bwgame

type ExecuteMovementStruct struct {
	RefreshVision     bool
	StartingMovement  bool
	StoppingMovement  bool
	Speed             Fp8
	Position          Xy
	ExactPosition     XyFp8
	PreMovementFlags  int32
	PostMovementFlags int32
}

type Pathfinder struct {
	U                 *Unit
	TargetUnit        *Unit
	Source            Xy
	SourceRegion      *Region
	DestinationRegion *Region
	UnitBB            Rect
	TargetUnitBB      Rect
	Destination       Xy

	LongPath         []*Region
	FullLongPathSize uint
	ShortPath        []Xy

	CurrentLongPathIndex  uint
	CurrentShortPathIndex uint

	ShortHighestOpenSize uint
	ShortAllNodesSize    uint
	LongHighestOpenSize  uint
	LongAllNodesSize     uint

	DestinationReached bool
	IsStuck            bool

	ConsiderCollisionWithUnit        *Unit
	ConsiderCollisionWithMovingUnits bool
}

const (
	imageOrderTop int32 = iota
	imageOrderBottom
	imageOrderAbove
	imageOrderBelow
)

type UnitCounters struct {
	UnitCounts          [12][UnitTypesNone]int32
	CompletedUnitCounts [12][UnitTypesNone]int32

	FactoryCounts     [12]int32
	BuildingCounts    [12]int32
	NonBuildingCounts [12]int32

	CompletedFactoryCounts     [12]int32
	CompletedBuildingCounts    [12]int32
	CompletedNonBuildingCounts [12]int32
}

type BringUnitCounters = UnitCounters

type ExecuteTriggerStruct struct {
	VictoryState [12]int32
}

type SetupInfo struct {
	CreateMeleeUnitsForPlayer [12]bool
	VictoryCondition          int32
	TournamentMode            int32
	StartingUnits             int32
	ResourceType              int32
	StartingMinerals          int32
	CreateNoUnits             bool
	TopVsBottom               bool
}

func NewSetupInfo() SetupInfo {
	return SetupInfo{StartingMinerals: 50}
}

type ActionState struct {
	PlayerID [12]int32

	ActionsDataPosition uint
	NextActionFrame     int32

	Selection     [8]StaticVector[*Unit, [12]*Unit]
	ControlGroups [8][10]StaticVector[UnitID, [12]UnitID]
}

type GroupMove struct {
	OriginalTargetPos    Xy
	TargetPos            Xy
	MoveOffset           Xy
	HasMoveOffset        bool
	TargetIsInUnitBounds bool
}
