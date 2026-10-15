package openbw

import (
	"fmt"
	"math/bits"
	"strconv"
	"strings"
)

// UnknownEnum is the value that UnmarshalText gives an enum name it does not know.
const UnknownEnum = 255

func enumName[T ~uint8](v T, names []string) string {
	if int(v) < len(names) {
		return names[v]
	}
	return strconv.Itoa(int(v))
}

func parseEnum[T ~uint8](text []byte, names []string, v *T) error {
	s := string(text)
	for i, name := range names {
		if name == s {
			*v = T(i)
			return nil
		}
	}
	if n, err := strconv.ParseUint(s, 10, 8); err == nil {
		*v = T(n)
		return nil
	}
	*v = UnknownEnum
	return nil
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
)

// DefaultKinds is every kind except CommandResult and Tick.
const DefaultKinds = AllKinds &^ (KindCommandResult | KindTick)

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

// MarshalText writes the kind names.
func (k Kind) MarshalText() ([]byte, error) { return []byte(k.String()), nil }

// UnmarshalText reads kind names. Unknown names are ignored.
func (k *Kind) UnmarshalText(text []byte) error {
	kinds, _ := ParseKinds(string(text))
	*k = kinds
	return nil
}

// ParseKinds reads a comma-separated list of kind names, for example "UnitCreated,UnitDestroyed".
func ParseKinds(list string) (Kind, error) {
	var kinds Kind
	for _, name := range strings.Split(list, ",") {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		bit := -1
		for i, known := range kindNames {
			if strings.EqualFold(known, name) {
				bit = i
			}
		}
		if bit < 0 {
			return kinds, fmt.Errorf("openbw: unknown event kind %q", name)
		}
		kinds |= 1 << bit
	}
	return kinds, nil
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

func (c UnitClass) String() string                   { return enumName(c, classNames) }
func (c UnitClass) MarshalText() ([]byte, error)     { return []byte(c.String()), nil }
func (c *UnitClass) UnmarshalText(text []byte) error { return parseEnum(text, classNames, c) }

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

func (s CreateSource) String() string                   { return enumName(s, sourceNames) }
func (s CreateSource) MarshalText() ([]byte, error)     { return []byte(s.String()), nil }
func (s *CreateSource) UnmarshalText(text []byte) error { return parseEnum(text, sourceNames, s) }

// DeathReason tells why a unit died.
type DeathReason uint8

// The death reasons.
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
)

var deathNames = []string{
	"Other", "Damage", "Container", "Spell", "Cancelled", "Expired", "Depleted", "Removed", "Detonated", "Linked", "Trigger",
}

func (r DeathReason) String() string                   { return enumName(r, deathNames) }
func (r DeathReason) MarshalText() ([]byte, error)     { return []byte(r.String()), nil }
func (r *DeathReason) UnmarshalText(text []byte) error { return parseEnum(text, deathNames, r) }

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

func (r OwnerReason) String() string                   { return enumName(r, ownerNames) }
func (r OwnerReason) MarshalText() ([]byte, error)     { return []byte(r.String()), nil }
func (r *OwnerReason) UnmarshalText(text []byte) error { return parseEnum(text, ownerNames, r) }

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

func (k ProductionKind) String() string                   { return enumName(k, productionNames) }
func (k ProductionKind) MarshalText() ([]byte, error)     { return []byte(k.String()), nil }
func (k *ProductionKind) UnmarshalText(text []byte) error { return parseEnum(text, productionNames, k) }

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

func (r CancelReason) String() string                   { return enumName(r, cancelNames) }
func (r CancelReason) MarshalText() ([]byte, error)     { return []byte(r.String()), nil }
func (r *CancelReason) UnmarshalText(text []byte) error { return parseEnum(text, cancelNames, r) }

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

func (r RejectReason) String() string                   { return enumName(r, rejectNames) }
func (r RejectReason) MarshalText() ([]byte, error)     { return []byte(r.String()), nil }
func (r *RejectReason) UnmarshalText(text []byte) error { return parseEnum(text, rejectNames, r) }

// Outcome is what the engine did with a command.
type Outcome uint8

// The outcomes.
const (
	Accepted Outcome = iota
	Rejected
	NoEffect
)

var outcomeNames = []string{"Accepted", "Rejected", "NoEffect"}

func (o Outcome) String() string                   { return enumName(o, outcomeNames) }
func (o Outcome) MarshalText() ([]byte, error)     { return []byte(o.String()), nil }
func (o *Outcome) UnmarshalText(text []byte) error { return parseEnum(text, outcomeNames, o) }

// RulesMode selects the engine rules.
type RulesMode uint8

// The rules modes.
const (
	RulesAuto RulesMode = iota
	RulesOpenBW
)

var rulesNames = []string{"Auto", "OpenBW"}

func (m RulesMode) String() string                   { return enumName(m, rulesNames) }
func (m RulesMode) MarshalText() ([]byte, error)     { return []byte(m.String()), nil }
func (m *RulesMode) UnmarshalText(text []byte) error { return parseEnum(text, rulesNames, m) }

// ResultsMode selects the command results that Run keeps.
type ResultsMode uint8

// The results modes.
const (
	ResultsActions ResultsMode = iota
	ResultsAll
	ResultsNone
)

var resultsNames = []string{"actions", "all", "none"}

func (m ResultsMode) String() string { return enumName(m, resultsNames) }

// ParseResultsMode reads "actions", "all" or "none".
func ParseResultsMode(s string) (ResultsMode, error) {
	for i, name := range resultsNames {
		if name == s {
			return ResultsMode(i), nil
		}
	}
	return 0, fmt.Errorf("openbw: unknown results mode %q", s)
}

// UnitField is a bit set of unit snapshot fields.
type UnitField uint16

// The unit snapshot fields. ID, Type and Owner are always present.
const (
	Position UnitField = 1 << iota
	HP
	Shields
	Energy
	Order
	State
	BuildTime
)

var unitFieldNames = []string{"pos", "hp", "shields", "energy", "order", "state", "buildtime"}

// ParseUnitFields reads a comma-separated list such as "pos,hp".
func ParseUnitFields(list string) (UnitField, error) {
	var fields UnitField
	for _, name := range strings.Split(list, ",") {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		bit := -1
		for i, known := range unitFieldNames {
			if known == name {
				bit = i
			}
		}
		if bit < 0 {
			return fields, fmt.Errorf("openbw: unknown unit field %q", name)
		}
		fields |= 1 << bit
	}
	return fields, nil
}

// The State bits of a unit snapshot.
const (
	StateCompleted uint8 = 1 << iota
	StateBurrowed
	StateCloaked
	StateLoaded
	StateDying
	StateHallucination
)
