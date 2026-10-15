package openbw

// Event is one of the *UnitCreated, *UnitCompleted, ... types of this package.
type Event interface {
	Base() *EventBase
	Kind() Kind
	isEvent()
}

// EventBase has the fields that every event has.
type EventBase struct {
	// Seq is 1, 2, 3, ... in emission order within a game.
	Seq int64
	// Frame is the engine frame when the event fired.
	Frame int32
	// Cmd is the Action of the command that caused the event, or -1. It is set only when KindCommandResult is subscribed.
	Cmd int32
}

// Base gives the common fields.
func (b *EventBase) Base() *EventBase { return b }

func (*EventBase) isEvent() {}

// UnitCreated fires when the engine creates a unit. A trained unit is created when its production starts.
type UnitCreated struct {
	EventBase
	Unit  UnitID
	Type  UnitType
	Owner int8
	X, Y  int32
	// Parent is the producer, the Larva's Hatchery, the copied unit, ...; 0 if none.
	Parent UnitID
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
	Unit   UnitID
	Type   UnitType
	Owner  int8
	X, Y   int32
	Reason DeathReason
	// Killer is the unit that killed it, 0 if none. KillerType is valid only when Killer is not 0.
	Killer     UnitID
	KillerType UnitType
	// KillerOwner is -1 if unknown. It can be set when Killer is 0, for example Irradiate after the caster died.
	KillerOwner int8
	// KillerParent is the Carrier or Reaver when Killer is an Interceptor or a Scarab.
	KillerParent UnitID
	// Container is the transport or Bunker when Reason is ReasonContainer.
	Container UnitID
	Reverted  bool
	Class     UnitClass
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
	// Type is valid when What is not ProductionResearch or ProductionUpgrade.
	Type UnitType
	// Research is valid when What is ProductionResearch or ProductionUpgrade.
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

// CommandResult is what the engine did with one command. It comes after the events that the command caused.
type CommandResult struct {
	EventBase
	// Action is the index in the engine's action stream. Cmd equals Action.
	Action int32
	// ScrepIndex is the index in screp's Commands.Cmds, or -1 if the link is not exact.
	ScrepIndex int32
	// PlayerID is screp's player ID; Owner is the engine slot.
	PlayerID byte
	Owner    int8
	// ActionID is the legacy action ID that the engine executed.
	ActionID byte
	Outcome  Outcome
	// Rejections are the engine's reasons for refusing the command or some of its units.
	Rejections []Rejection
	// Selection is the selection after a Select, Shift Select or Deselect, else the units the command acts on.
	Selection []UnitID
	// MissingTags counts Select and Shift Select tags that resolve to no unit.
	MissingTags uint8
	// ForeignTags counts tags of another player's units in a Select or Shift Select of 2 or more tags.
	ForeignTags uint8
}

// Rejection is one reason why the engine refused a command.
type Rejection struct {
	// Unit is the selected unit that refused, or 0 for the command as a whole.
	Unit   UnitID
	Reason RejectReason
}

// Tick is emitted every Options.TickEvery frames, last in the frame.
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
