package openbw

import (
	"github.com/icza/screp/rep/repcmd"
	"github.com/marianogappa/openbw-go/bwgame"
	"github.com/marianogappa/openbw-go/downgrade"
)

// UnitID is unique within a game: bits 0-10 are the unit index + 1 (1..1700),
// bits 11-31 are the slot's generation modulo 2^21. 0 means no unit.
type UnitID uint32

const (
	unitIndexBits       = 11
	unitGenerationLimit = 1 << 21
	legacyGenerations   = 32
)

func unitIDOf(u *bwgame.Unit) UnitID {
	if u == nil {
		return 0
	}
	return UnitID(uint32(u.Index+1) | (u.UnitIDGeneration%unitGenerationLimit)<<unitIndexBits)
}

// Index is the unit index + 1 (1..1700).
func (id UnitID) Index() uint16 { return uint16(id & (1<<unitIndexBits - 1)) }

// Generation is how many times the index was reused, modulo 2^21.
func (id UnitID) Generation() uint32 { return uint32(id) >> unitIndexBits }

// LegacyTag is the 1.16 unit tag and OpenBW unit_id_t: index | (generation%32)<<11.
func (id UnitID) LegacyTag() uint16 {
	return id.Index() | uint16(id.Generation()%legacyGenerations)<<unitIndexBits
}

// MatchesSCRTag tells if a 1.21 unit tag refers to this unit, with the downgrade's own mapping.
func (id UnitID) MatchesSCRTag(tag uint32) bool {
	return id != 0 && downgrade.LegacyUnitTag(tag) == id.LegacyTag()
}

// UnitType is a units.dat ID.
type UnitType uint16

// UnitTypeNone is the type of no unit, for example KillerType when there is no killer.
const UnitTypeNone = UnitType(bwgame.UnitTypesNone)

func unitTypeOf(t *bwgame.UnitTypeT) UnitType {
	if t == nil {
		return UnitTypeNone
	}
	return UnitType(t.ID)
}

// Name is screp's name of the type, or "None" for UnitTypeNone.
func (t UnitType) Name() string {
	if t == UnitTypeNone {
		return "None"
	}
	return repcmd.UnitByID(uint16(t)).Name
}

// String is Name.
func (t UnitType) String() string { return t.Name() }

// Research is a tech (techdata.dat ID) or an upgrade (upgrades.dat ID).
type Research struct {
	Upgrade bool
	ID      uint8
}

const noResearchID = 0xff

// NoResearch is the Research of no tech or upgrade, for example in a ProductionCancelled of a unit.
var NoResearch = Research{ID: noResearchID}

// Valid tells if r is a tech or an upgrade, not NoResearch.
func (r Research) Valid() bool { return r.ID != noResearchID }

func researchOf(r bwgame.Research) Research { return Research{Upgrade: r.Upgrade, ID: r.ID} }

// String is screp's name of the tech or upgrade, or "None" for NoResearch.
func (r Research) String() string {
	if !r.Valid() {
		return "None"
	}
	if r.Upgrade {
		return repcmd.UpgradeByID(r.ID).Name
	}
	return repcmd.TechByID(r.ID).Name
}
