package openbw

import (
	"slices"

	"github.com/marianogappa/openbw-go/bwgame"
	"github.com/marianogappa/openbw-go/downgrade"
)

const (
	actionSelect      = 0x09
	actionShiftSelect = 0x0a
	actionDeselect    = 0x0b
	actionBuild       = 0x0c
	actionHotkey      = 0x13
	hotkeyRecall      = 1
	legacyTagSize     = 2
	selectionOwners   = 8
)

func changesSelection(id byte, payload []byte) bool {
	switch id {
	case actionSelect, actionShiftSelect, actionDeselect:
		return true
	case actionHotkey:
		return len(payload) > 0 && payload[0] == hotkeyRecall
	}
	return false
}

type shadowQueue struct {
	queue   []int32
	limbo   []int32
	cleared int32
}

type commandTracker struct {
	executed       int32
	current        *CommandResult
	selectionAfter bool
	queues         [bwgame.UnitLimit]shadowQueue
	addons         [bwgame.UnitLimit]int32
	eggs           map[UnitID]int32
}

func newCommandTracker() *commandTracker {
	t := &commandTracker{eggs: map[UnitID]int32{}}
	for i := range t.addons {
		t.addons[i] = -1
		t.queues[i].cleared = -1
	}
	return t
}

func (t *commandTracker) sync() int32 {
	if t.current == nil {
		return -1
	}
	return t.current.Action
}

func (t *commandTracker) abort() { t.current, t.selectionAfter = nil, false }

func (t *commandTracker) front(u *bwgame.Unit) int32 {
	if u == nil {
		return -1
	}
	q := t.queues[u.Index].queue
	if len(q) == 0 {
		return -1
	}
	return q[0]
}

func (t *commandTracker) frontOrCleared(u *bwgame.Unit) int32 {
	if front := t.front(u); front >= 0 || u == nil {
		return front
	}
	return t.queues[u.Index].cleared
}

func (t *commandTracker) onBuildQueue(u *bwgame.Unit, op bwgame.BuildQueueOp, index int) {
	m := &t.queues[u.Index]
	switch op {
	case bwgame.BuildQueuePush:
		m.queue = append(m.queue[:min(index, len(m.queue))], t.sync())
		m.queue = append(m.queue, m.limbo...)
		m.limbo = m.limbo[:0]
		m.cleared = -1
	case bwgame.BuildQueueClearSingle:
		m.cleared = -1
		if len(m.queue) > 0 {
			m.cleared = m.queue[0]
			m.limbo = append(m.limbo[:0], m.queue[1:]...)
		}
		m.queue = m.queue[:0]
	case bwgame.BuildQueueErase:
		if index < len(m.queue) {
			m.queue = slices.Delete(m.queue, index, index+1)
		}
	case bwgame.BuildQueueReset:
		m.queue, m.limbo, m.cleared = m.queue[:0], m.limbo[:0], -1
	}
}

func (t *commandTracker) onAddonBuildType(u *bwgame.Unit, set bool) {
	if set {
		t.addons[u.Index] = t.sync()
	} else {
		t.addons[u.Index] = -1
	}
}

func (t *commandTracker) onRejected(u *bwgame.Unit, reason bwgame.RejectReason) {
	if t.current != nil {
		t.current.Rejections = append(t.current.Rejections, Rejection{Unit: unitIDOf(u), Reason: RejectReason(reason)})
	}
}

func (g *Game) selection(owner int32) []UnitID {
	if owner < 0 || owner >= selectionOwners {
		return nil
	}
	units := g.funcs.ActionState().Selection[owner].Values()
	if len(units) == 0 {
		return nil
	}
	ids := make([]UnitID, len(units))
	for i, u := range units {
		ids[i] = unitIDOf(u)
	}
	return ids
}

func (g *Game) beforeAction(owner, actionID int32, payload []byte) {
	t := g.cmd
	r := &CommandResult{
		Action:   t.executed,
		Source:   g.source(t.executed),
		Owner:    int8(owner),
		ActionID: byte(actionID),
	}
	if byte(actionID) == actionSelect || byte(actionID) == actionShiftSelect {
		r.MissingTags, r.ForeignTags = g.tagTest(owner, payload)
	}
	t.selectionAfter = changesSelection(byte(actionID), payload)
	if !t.selectionAfter {
		r.Selection = g.selection(owner)
	}
	t.current = r
}

func (g *Game) tagTest(owner int32, payload []byte) (missing, foreign uint8) {
	if len(payload) == 0 {
		return 0, 0
	}
	n := int(payload[0])
	if len(payload) < 1+legacyTagSize*n {
		return 0, 0
	}
	for i := range n {
		raw := uint16(payload[1+legacyTagSize*i]) | uint16(payload[2+legacyTagSize*i])<<8
		id := bwgame.UnitID{RawValue: raw}
		if id.Index() == 0 {
			continue
		}
		u := g.funcs.UnitByID(id)
		switch {
		case u == nil:
			missing++
		case n > 1 && u.Owner != owner:
			foreign++
		}
	}
	return missing, foreign
}

func (g *Game) afterAction(owner, actionID int32, ok bool) {
	t := g.cmd
	r := t.current
	selectionAfter := t.selectionAfter
	if r == nil {
		r = &CommandResult{Action: t.executed, Source: g.source(t.executed), Owner: int8(owner), ActionID: byte(actionID)}
		selectionAfter = false
	}
	t.current, t.selectionAfter = nil, false
	t.executed++
	switch {
	case !ok:
		r.Outcome = Rejected
	case byte(actionID) == actionBuild && len(r.Rejections) > 0:
		r.Outcome = NoEffect
	default:
		r.Outcome = Accepted
	}
	if selectionAfter {
		r.Selection = g.selection(owner)
	}
	g.emit(r, r.Action)
}

func actionSources(commands []downgrade.SourceCommand) []int32 {
	var sources []int32
	for i, c := range commands {
		if c.Action < 0 {
			continue
		}
		for int(c.Action) >= len(sources) {
			sources = append(sources, -1)
		}
		sources[c.Action] = int32(i)
	}
	return sources
}

func (g *Game) source(action int32) int32 {
	if action < 0 || int(action) >= len(g.sources) {
		return -1
	}
	return g.sources[action]
}
