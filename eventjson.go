package openbw

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
)

// UnmarshalEvent decodes one event of a Result's JSON by its "Kind".
func UnmarshalEvent(data []byte) (Event, error) {
	var head struct{ Kind Kind }
	if err := json.Unmarshal(data, &head); err != nil {
		return nil, err
	}
	e := newEvent(head.Kind)
	if e == nil {
		return nil, fmt.Errorf("openbw: unknown event kind in %.60s", data)
	}
	if err := json.Unmarshal(data, e); err != nil {
		return nil, err
	}
	return e, nil
}

func marshalEvent(kind Kind, b *EventBase, fields any) ([]byte, error) {
	body, err := json.Marshal(fields)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	buf.Grow(len(body) + 48)
	buf.WriteByte('{')
	if kind != 0 {
		buf.WriteString(`"Kind":"`)
		buf.WriteString(kind.String())
		buf.WriteString(`",`)
	}
	buf.WriteString(`"Frame":`)
	buf.WriteString(strconv.Itoa(int(b.Frame)))
	if b.Cmd != -1 {
		buf.WriteString(`,"Cmd":`)
		buf.WriteString(strconv.Itoa(int(b.Cmd)))
	}
	if len(body) > 2 {
		buf.WriteByte(',')
		buf.Write(body[1:])
	} else {
		buf.WriteByte('}')
	}
	return buf.Bytes(), nil
}

func unmarshalEvent(data []byte, b *EventBase, fields any) error {
	var head struct {
		Frame int32
		Cmd   *int32
	}
	if err := json.Unmarshal(data, &head); err != nil {
		return err
	}
	b.Frame = head.Frame
	b.Cmd = -1
	if head.Cmd != nil {
		b.Cmd = *head.Cmd
	}
	return json.Unmarshal(data, fields)
}

type (
	unitCreatedFields       UnitCreated
	unitCompletedFields     UnitCompleted
	unitMorphedFields       UnitMorphed
	morphStartedFields      MorphStarted
	unitOwnerChangedFields  UnitOwnerChanged
	researchStartedFields   ResearchStarted
	researchCompletedFields ResearchCompleted
	playerLeftFields        PlayerLeft
	playerDefeatedFields    PlayerDefeated
	playerWonFields         PlayerWon
	commandResultFields     CommandResult
	tickFields              Tick
)

func (e UnitCreated) MarshalJSON() ([]byte, error) {
	return marshalEvent(KindUnitCreated, &e.EventBase, unitCreatedFields(e))
}
func (e *UnitCreated) UnmarshalJSON(data []byte) error {
	return unmarshalEvent(data, &e.EventBase, (*unitCreatedFields)(e))
}

func (e UnitCompleted) MarshalJSON() ([]byte, error) {
	return marshalEvent(KindUnitCompleted, &e.EventBase, unitCompletedFields(e))
}
func (e *UnitCompleted) UnmarshalJSON(data []byte) error {
	return unmarshalEvent(data, &e.EventBase, (*unitCompletedFields)(e))
}

func (e UnitMorphed) MarshalJSON() ([]byte, error) {
	return marshalEvent(KindUnitMorphed, &e.EventBase, unitMorphedFields(e))
}
func (e *UnitMorphed) UnmarshalJSON(data []byte) error {
	return unmarshalEvent(data, &e.EventBase, (*unitMorphedFields)(e))
}

func (e MorphStarted) MarshalJSON() ([]byte, error) {
	return marshalEvent(KindMorphStarted, &e.EventBase, morphStartedFields(e))
}
func (e *MorphStarted) UnmarshalJSON(data []byte) error {
	return unmarshalEvent(data, &e.EventBase, (*morphStartedFields)(e))
}

func (e UnitOwnerChanged) MarshalJSON() ([]byte, error) {
	return marshalEvent(KindUnitOwnerChanged, &e.EventBase, unitOwnerChangedFields(e))
}
func (e *UnitOwnerChanged) UnmarshalJSON(data []byte) error {
	return unmarshalEvent(data, &e.EventBase, (*unitOwnerChangedFields)(e))
}

func (e ResearchStarted) MarshalJSON() ([]byte, error) {
	return marshalEvent(KindResearchStarted, &e.EventBase, researchStartedFields(e))
}
func (e *ResearchStarted) UnmarshalJSON(data []byte) error {
	return unmarshalEvent(data, &e.EventBase, (*researchStartedFields)(e))
}

func (e ResearchCompleted) MarshalJSON() ([]byte, error) {
	return marshalEvent(KindResearchCompleted, &e.EventBase, researchCompletedFields(e))
}
func (e *ResearchCompleted) UnmarshalJSON(data []byte) error {
	return unmarshalEvent(data, &e.EventBase, (*researchCompletedFields)(e))
}

func (e PlayerLeft) MarshalJSON() ([]byte, error) {
	return marshalEvent(KindPlayerLeft, &e.EventBase, playerLeftFields(e))
}
func (e *PlayerLeft) UnmarshalJSON(data []byte) error {
	return unmarshalEvent(data, &e.EventBase, (*playerLeftFields)(e))
}

func (e PlayerDefeated) MarshalJSON() ([]byte, error) {
	return marshalEvent(KindPlayerDefeated, &e.EventBase, playerDefeatedFields(e))
}
func (e *PlayerDefeated) UnmarshalJSON(data []byte) error {
	return unmarshalEvent(data, &e.EventBase, (*playerDefeatedFields)(e))
}

func (e PlayerWon) MarshalJSON() ([]byte, error) {
	return marshalEvent(KindPlayerWon, &e.EventBase, playerWonFields(e))
}
func (e *PlayerWon) UnmarshalJSON(data []byte) error {
	return unmarshalEvent(data, &e.EventBase, (*playerWonFields)(e))
}

func (e CommandResult) MarshalJSON() ([]byte, error) {
	return marshalEvent(KindCommandResult, &e.EventBase, commandResultFields(e))
}
func (e *CommandResult) UnmarshalJSON(data []byte) error {
	return unmarshalEvent(data, &e.EventBase, (*commandResultFields)(e))
}

func (e Tick) MarshalJSON() ([]byte, error) {
	return marshalEvent(KindTick, &e.EventBase, tickFields(e))
}
func (e *Tick) UnmarshalJSON(data []byte) error {
	return unmarshalEvent(data, &e.EventBase, (*tickFields)(e))
}

type unitDestroyedJSON struct {
	Unit         UnitID
	Type         UnitType
	Owner        int8
	X, Y         int32
	Reason       DeathReason
	Killer       UnitID    `json:",omitzero"`
	KillerType   *UnitType `json:",omitempty"`
	KillerOwner  *int8     `json:",omitempty"`
	KillerParent UnitID    `json:",omitzero"`
	Container    UnitID    `json:",omitzero"`
	Reverted     bool      `json:",omitzero"`
	Class        UnitClass
}

func (e UnitDestroyed) MarshalJSON() ([]byte, error) {
	v := unitDestroyedJSON{
		Unit: e.Unit, Type: e.Type, Owner: e.Owner, X: e.X, Y: e.Y, Reason: e.Reason, Killer: e.Killer,
		KillerParent: e.KillerParent, Container: e.Container, Reverted: e.Reverted, Class: e.Class,
	}
	if e.Killer != 0 {
		v.KillerType = &e.KillerType
	}
	if e.KillerOwner >= 0 {
		v.KillerOwner = &e.KillerOwner
	}
	return marshalEvent(KindUnitDestroyed, &e.EventBase, v)
}

func (e *UnitDestroyed) UnmarshalJSON(data []byte) error {
	var v unitDestroyedJSON
	if err := unmarshalEvent(data, &e.EventBase, &v); err != nil {
		return err
	}
	e.Unit, e.Type, e.Owner, e.X, e.Y, e.Reason, e.Killer = v.Unit, v.Type, v.Owner, v.X, v.Y, v.Reason, v.Killer
	e.KillerParent, e.Container, e.Reverted, e.Class = v.KillerParent, v.Container, v.Reverted, v.Class
	e.KillerType, e.KillerOwner = 0, -1
	if v.KillerType != nil {
		e.KillerType = *v.KillerType
	}
	if v.KillerOwner != nil {
		e.KillerOwner = *v.KillerOwner
	}
	return nil
}

type productionCancelledJSON struct {
	Owner    int8
	Producer UnitID `json:",omitzero"`
	Unit     UnitID `json:",omitzero"`
	What     ProductionKind
	Type     *UnitType `json:",omitempty"`
	Research *Research `json:",omitempty"`
	Why      CancelReason
}

func (e ProductionCancelled) isResearch() bool {
	return e.What == ProductionResearch || e.What == ProductionUpgrade
}

func (e ProductionCancelled) MarshalJSON() ([]byte, error) {
	v := productionCancelledJSON{Owner: e.Owner, Producer: e.Producer, Unit: e.Unit, What: e.What, Why: e.Why}
	if e.isResearch() {
		v.Research = &e.Research
	} else {
		v.Type = &e.Type
	}
	return marshalEvent(KindProductionCancelled, &e.EventBase, v)
}

func (e *ProductionCancelled) UnmarshalJSON(data []byte) error {
	var v productionCancelledJSON
	if err := unmarshalEvent(data, &e.EventBase, &v); err != nil {
		return err
	}
	e.Owner, e.Producer, e.Unit, e.What, e.Why = v.Owner, v.Producer, v.Unit, v.What, v.Why
	if v.Type != nil {
		e.Type = *v.Type
	}
	if v.Research != nil {
		e.Research = *v.Research
	}
	return nil
}
