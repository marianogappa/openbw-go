package openbw

// Event is one of the *UnitCreated, *UnitCompleted, ... types of this package.
type Event interface {
	Base() *EventBase
	Kind() Kind
	isEvent()
}

// EventBase has the fields that every event has.
// Frame and Cmd are written by the event's own JSON encoding; Cmd is omitted when -1.
type EventBase struct {
	Seq   int64 `json:"-"`
	Frame int32 `json:"-"`
	Cmd   int32 `json:"-"`
}

// Base gives the common fields.
func (b *EventBase) Base() *EventBase { return b }

func (*EventBase) isEvent() {}

// UnitCreated fires when the engine creates a unit. A trained unit is created when its production starts.
type UnitCreated struct {
	EventBase
	Unit   UnitID
	Type   UnitType
	Owner  int8
	X, Y   int32
	Parent UnitID `json:",omitzero"`
	Source CreateSource
	Class  UnitClass
}

// UnitCompleted fires when a unit completes, once for each type it completes as.
type UnitCompleted struct {
	EventBase
	Unit  UnitID
	Type  UnitType
	Owner int8
	Class UnitClass
}

// UnitMorphed fires each time a unit's type changes.
type UnitMorphed struct {
	EventBase
	Unit     UnitID
	From, To UnitType
	Owner    int8
	Class    UnitClass
}

// MorphStarted fires when a Zerg morph is paid for.
type MorphStarted struct {
	EventBase
	Unit    UnitID
	Morpher UnitID
	From    UnitType
	Into    UnitType
	Owner   int8
	Class   UnitClass
}

// UnitDestroyed fires when the engine decides that a unit dies.
// A refinery death has Reverted set: the ID lives on as a neutral Vespene Geyser.
type UnitDestroyed struct {
	EventBase
	Unit         UnitID
	Type         UnitType
	Owner        int8
	X, Y         int32
	Reason       DeathReason
	Killer       UnitID
	KillerType   UnitType
	KillerOwner  int8
	KillerParent UnitID
	Container    UnitID
	Reverted     bool
	Class        UnitClass
}

// UnitOwnerChanged fires when a unit changes owner.
type UnitOwnerChanged struct {
	EventBase
	Unit     UnitID
	Type     UnitType
	From, To int8
	Reason   OwnerReason
	Class    UnitClass
}

// ProductionCancelled fires when queued or started production is cancelled.
type ProductionCancelled struct {
	EventBase
	Owner    int8
	Producer UnitID
	Unit     UnitID
	What     ProductionKind
	Type     UnitType
	Research Research
	Why      CancelReason
}

// ResearchStarted fires when a tech or upgrade is paid for.
type ResearchStarted struct {
	EventBase
	Owner    int8
	Unit     UnitID
	Research Research
	Level    int32
}

// ResearchCompleted fires when a tech or upgrade level is gained.
type ResearchCompleted struct {
	EventBase
	Owner    int8
	Unit     UnitID
	Research Research
	Level    int32
}

// PlayerLeft fires when a player leaves (Dropped: dropped by the game).
type PlayerLeft struct {
	EventBase
	Owner   int8
	Dropped bool
}

// PlayerDefeated fires when the victory triggers eliminate a player.
type PlayerDefeated struct {
	EventBase
	Owner int8
}

// PlayerWon fires when the victory triggers make a player win.
type PlayerWon struct {
	EventBase
	Owner int8
	State int32
}

// CommandResult is what the engine did with one command.
type CommandResult struct {
	EventBase
	Action      int32
	ScrepIndex  int32
	PlayerID    byte
	Owner       int8
	ActionID    byte
	Outcome     Outcome
	Rejections  []Rejection `json:",omitempty"`
	Selection   []UnitID    `json:",omitempty"`
	MissingTags uint8       `json:",omitzero"`
	ForeignTags uint8       `json:",omitzero"`
}

// Rejection is one reason why the engine refused a command.
type Rejection struct {
	Unit   UnitID `json:",omitzero"`
	Reason RejectReason
}

// Tick is emitted every StreamOptions.TickEvery frames, last in the frame.
type Tick struct{ EventBase }

func (*UnitCreated) Kind() Kind         { return KindUnitCreated }
func (*UnitCompleted) Kind() Kind       { return KindUnitCompleted }
func (*UnitMorphed) Kind() Kind         { return KindUnitMorphed }
func (*MorphStarted) Kind() Kind        { return KindMorphStarted }
func (*UnitDestroyed) Kind() Kind       { return KindUnitDestroyed }
func (*UnitOwnerChanged) Kind() Kind    { return KindUnitOwnerChanged }
func (*ProductionCancelled) Kind() Kind { return KindProductionCancelled }
func (*ResearchStarted) Kind() Kind     { return KindResearchStarted }
func (*ResearchCompleted) Kind() Kind   { return KindResearchCompleted }
func (*PlayerLeft) Kind() Kind          { return KindPlayerLeft }
func (*PlayerDefeated) Kind() Kind      { return KindPlayerDefeated }
func (*PlayerWon) Kind() Kind           { return KindPlayerWon }
func (*CommandResult) Kind() Kind       { return KindCommandResult }
func (*Tick) Kind() Kind                { return KindTick }

func newEvent(kind Kind) Event {
	switch kind {
	case KindUnitCreated:
		return &UnitCreated{}
	case KindUnitCompleted:
		return &UnitCompleted{}
	case KindUnitMorphed:
		return &UnitMorphed{}
	case KindMorphStarted:
		return &MorphStarted{}
	case KindUnitDestroyed:
		return &UnitDestroyed{}
	case KindUnitOwnerChanged:
		return &UnitOwnerChanged{}
	case KindProductionCancelled:
		return &ProductionCancelled{}
	case KindResearchStarted:
		return &ResearchStarted{}
	case KindResearchCompleted:
		return &ResearchCompleted{}
	case KindPlayerLeft:
		return &PlayerLeft{}
	case KindPlayerDefeated:
		return &PlayerDefeated{}
	case KindPlayerWon:
		return &PlayerWon{}
	case KindCommandResult:
		return &CommandResult{}
	case KindTick:
		return &Tick{}
	}
	return nil
}
