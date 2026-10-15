package openbw

import (
	"encoding/json"
	"fmt"

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

// UnitType is a units.dat ID. It marshals to JSON as screp's {"Name", "ID"} object.
type UnitType uint16

func unitTypeOf(t *bwgame.UnitTypeT) UnitType {
	if t == nil {
		return UnitType(bwgame.UnitTypesNone)
	}
	return UnitType(t.ID)
}

// Name is screp's name of the type.
func (t UnitType) Name() string { return repcmd.UnitByID(uint16(t)).Name }

func (t UnitType) String() string { return t.Name() }

func (t UnitType) MarshalJSON() ([]byte, error) {
	return json.Marshal(repcmd.UnitByID(uint16(t)))
}

func (t *UnitType) UnmarshalJSON(data []byte) error {
	var v struct{ ID uint16 }
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	*t = UnitType(v.ID)
	return nil
}

// Research is a tech or an upgrade.
type Research struct {
	Upgrade bool
	ID      uint8
}

func researchOf(r bwgame.Research) Research { return Research{Upgrade: r.Upgrade, ID: r.ID} }

func (r Research) String() string {
	if r.Upgrade {
		return repcmd.UpgradeByID(r.ID).Name
	}
	return repcmd.TechByID(r.ID).Name
}

func (r Research) MarshalJSON() ([]byte, error) {
	if r.Upgrade {
		return json.Marshal(struct{ Upgrade *repcmd.Upgrade }{repcmd.UpgradeByID(r.ID)})
	}
	return json.Marshal(struct{ Tech *repcmd.Tech }{repcmd.TechByID(r.ID)})
}

func (r *Research) UnmarshalJSON(data []byte) error {
	var v struct {
		Tech    *struct{ ID uint8 }
		Upgrade *struct{ ID uint8 }
	}
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	switch {
	case v.Upgrade != nil:
		*r = Research{Upgrade: true, ID: v.Upgrade.ID}
	case v.Tech != nil:
		*r = Research{ID: v.Tech.ID}
	default:
		return fmt.Errorf("openbw: research without Tech or Upgrade: %s", data)
	}
	return nil
}
