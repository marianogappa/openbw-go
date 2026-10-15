package openbw

import (
	"math/bits"
	"strconv"
	"strings"
)

func enumName[T ~uint8](v T, names []string) string {
	if int(v) < len(names) {
		return names[v]
	}
	return strconv.Itoa(int(v))
}

// Kind is a bit set of event kinds.
type Kind uint32

// The event kinds. Each one is one bit.
const (
	KindUnitCreated Kind = 1 << iota
	KindUnitCompleted
	KindUnitMorphed
	KindMorphStarted
	KindUnitDestroyed
	KindUnitOwnerChanged
	KindProductionCancelled
	KindResearchStarted
	KindResearchCompleted
	KindPlayerLeft
	KindPlayerDefeated
	KindPlayerWon
	KindCommandResult
	KindTick

	AllKinds Kind = KindTick<<1 - 1

	// NoKinds in Options.Kinds subscribes to no kind, for games read through state, AdvanceTo or Options.Hooks.
	NoKinds Kind = 1 << 31
)

// DefaultKinds are the unit, production, research and player kinds; not CommandResult or Tick.
// A kind added in a later version is not in DefaultKinds.
const DefaultKinds = KindUnitCreated | KindUnitCompleted | KindUnitMorphed | KindMorphStarted | KindUnitDestroyed |
	KindUnitOwnerChanged | KindProductionCancelled | KindResearchStarted | KindResearchCompleted |
	KindPlayerLeft | KindPlayerDefeated | KindPlayerWon

var kindNames = []string{
	"UnitCreated", "UnitCompleted", "UnitMorphed", "MorphStarted", "UnitDestroyed", "UnitOwnerChanged",
	"ProductionCancelled", "ResearchStarted", "ResearchCompleted", "PlayerLeft", "PlayerDefeated", "PlayerWon",
	"CommandResult", "Tick",
}

// String gives the kind names, joined with commas.
func (k Kind) String() string {
	var names []string
	for rest := k; rest != 0; rest &= rest - 1 {
		bit := bits.TrailingZeros32(uint32(rest))
		if bit < len(kindNames) {
			names = append(names, kindNames[bit])
		} else {
			names = append(names, "Kind"+strconv.Itoa(bit))
		}
	}
	return strings.Join(names, ",")
}

// UnitClass groups units that consumers often treat apart.
type UnitClass uint8

// The unit classes.
const (
	ClassNormal UnitClass = iota
	ClassFighter
	ClassSpiderMine
	ClassHallucination
	ClassSpell
)

var classNames = []string{"Normal", "Fighter", "SpiderMine", "Hallucination", "Spell"}

func (c UnitClass) String() string { return enumName(c, classNames) }

// ClassSet is a bit set: bit c is UnitClass c.
type ClassSet uint8

// AllClasses has every unit class.
const AllClasses ClassSet = 1<<(ClassSpell+1) - 1

// Classes makes a set of the given classes.
func Classes(classes ...UnitClass) ClassSet {
	var set ClassSet
	for _, c := range classes {
		set |= 1 << c
	}
	return set
}

// Has tells if c is in the set.
func (s ClassSet) Has(c UnitClass) bool { return s&(1<<c) != 0 }

// CreateSource tells how a unit was created.
type CreateSource uint8

// The create sources.
const (
	SourceMap CreateSource = iota
	SourceStart
	SourceTrained
	SourceBuilt
	SourceLarva
	SourceSecondEggUnit
	SourceBroodling
	SourceHallucination
	SourceSpell
	SourceScanner
	SourceSpiderMine
	SourceDroneReturned
	SourceTrigger
)

var sourceNames = []string{
	"Map", "Start", "Trained", "Built", "Larva", "SecondEggUnit", "Broodling", "Hallucination", "Spell",
	"Scanner", "SpiderMine", "DroneReturned", "Trigger",
}

func (s CreateSource) String() string { return enumName(s, sourceNames) }

// DeathReason tells why a unit died.
type DeathReason uint8

// The death reasons. ReasonOffCreep is a Larva off creep, ReasonCrushed a unit under a landing building
// (Killer is the building) and ReasonBlocked a Siege Tank that sieges on a building.
const (
	ReasonOther DeathReason = iota
	ReasonDamage
	ReasonContainer
	ReasonSpell
	ReasonCancelled
	ReasonExpired
	ReasonDepleted
	ReasonRemoved
	ReasonDetonated
	ReasonLinked
	ReasonTrigger
	ReasonOffCreep
	ReasonCrushed
	ReasonBlocked
)

var deathNames = []string{
	"Other", "Damage", "Container", "Spell", "Cancelled", "Expired", "Depleted", "Removed", "Detonated", "Linked", "Trigger",
	"OffCreep", "Crushed", "Blocked",
}

func (r DeathReason) String() string { return enumName(r, deathNames) }

// OwnerReason tells why a unit changed owner.
type OwnerReason uint8

// The owner change reasons.
const (
	OwnerOther OwnerReason = iota
	OwnerMindControl
	OwnerPlayerLeft
	OwnerRefinery
	OwnerInfest
	OwnerAddon
	OwnerTrigger
)

var ownerNames = []string{"Other", "MindControl", "PlayerLeft", "Refinery", "Infest", "Addon", "Trigger"}

func (r OwnerReason) String() string { return enumName(r, ownerNames) }

// ProductionKind is what a cancelled production was.
type ProductionKind uint8

// The production kinds.
const (
	ProductionUnit ProductionKind = iota
	ProductionBuilding
	ProductionMorph
	ProductionResearch
	ProductionUpgrade
	ProductionAddon
)

var productionNames = []string{"Unit", "Building", "Morph", "Research", "Upgrade", "Addon"}

func (k ProductionKind) String() string { return enumName(k, productionNames) }

// CancelReason tells why production was cancelled.
type CancelReason uint8

// The cancel reasons. CancelOther is a cancel the engine makes on its own.
const (
	CancelOther CancelReason = iota
	CancelCommand
	CancelProducerDestroyed
	CancelOwnerChanged
	CancelNoSupply
)

var cancelNames = []string{"Other", "Command", "ProducerDestroyed", "OwnerChanged", "NoSupply"}

func (r CancelReason) String() string { return enumName(r, cancelNames) }

// RejectReason tells why the engine refused a command.
type RejectReason uint8

// The reject reasons.
const (
	RejectNoUnit RejectReason = iota
	RejectNotOwner
	RejectRequirements
	RejectInvalid
	RejectQueueFull
	RejectBusy
	RejectSupply
	RejectMinerals
	RejectGas
	RejectPlacement
)

var rejectNames = []string{"NoUnit", "NotOwner", "Requirements", "Invalid", "QueueFull", "Busy", "Supply", "Minerals", "Gas", "Placement"}

func (r RejectReason) String() string { return enumName(r, rejectNames) }

// Outcome is what the engine did with a command.
type Outcome uint8

// The outcomes.
const (
	Accepted Outcome = iota
	Rejected
	NoEffect
)

var outcomeNames = []string{"Accepted", "Rejected", "NoEffect"}

func (o Outcome) String() string { return enumName(o, outcomeNames) }

// RulesMode selects the engine rules.
type RulesMode uint8

// The rules modes.
const (
	RulesAuto RulesMode = iota
	RulesOpenBW
)

var rulesNames = []string{"Auto", "OpenBW"}

func (m RulesMode) String() string { return enumName(m, rulesNames) }
