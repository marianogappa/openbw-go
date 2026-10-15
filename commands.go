package openbw

import (
	"slices"

	"github.com/icza/screp/rep"
	"github.com/icza/screp/rep/repcmd"
	"github.com/marianogappa/openbw-go/bwgame"
	"github.com/marianogappa/openbw-go/downgrade"
)

const (
	actionSelect      = 0x09
	actionShiftSelect = 0x0a
	actionDeselect    = 0x0b
	actionBuild       = 0x0c
	legacyTagSize     = 2
	selectionOwners   = 8
)

func isSelectAction(id byte) bool {
	return id == actionSelect || id == actionShiftSelect || id == actionDeselect
}

type shadowQueue struct {
	queue   []int32
	limbo   []int32
	cleared int32
}

type commandTracker struct {
	executed int32
	current  *CommandResult
	queues   [bwgame.UnitLimit]shadowQueue
	addons   [bwgame.UnitLimit]int32
	eggs     map[UnitID]int32
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

func (t *commandTracker) abort() { t.current = nil }

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
		Action:     t.executed,
		ScrepIndex: g.links.screpIndex(t.executed),
		PlayerID:   byte(g.funcs.ActionState().PlayerID[owner]),
		Owner:      int8(owner),
		ActionID:   byte(actionID),
	}
	switch byte(actionID) {
	case actionSelect, actionShiftSelect:
		r.MissingTags, r.ForeignTags = g.tagTest(owner, payload)
	case actionDeselect:
	default:
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
	if r == nil {
		r = &CommandResult{Action: t.executed, ScrepIndex: g.links.screpIndex(t.executed), Owner: int8(owner), ActionID: byte(actionID)}
	}
	t.current = nil
	t.executed++
	switch {
	case !ok:
		r.Outcome = Rejected
	case byte(actionID) == actionBuild && len(r.Rejections) > 0:
		r.Outcome = NoEffect
	default:
		r.Outcome = Accepted
	}
	if isSelectAction(byte(actionID)) {
		r.Selection = g.selection(owner)
	}
	g.emit(r, r.Action)
}

type screpLinks []int32

func (l screpLinks) screpIndex(action int32) int32 {
	if action < 0 || int(action) >= len(l) {
		return -1
	}
	return l[action]
}

func screpIDMatches(sourceID, screpID byte) bool {
	return sourceID == screpID || sourceID == repcmd.TypeIDBuild && screpID == repcmd.VirtualTypeIDLand
}

func linkCommands(r *rep.Replay, p *downgrade.Prepared) screpLinks {
	actions := 0
	for _, c := range p.Commands {
		if c.Action >= 0 {
			actions++
		}
	}
	links := make(screpLinks, actions)
	for i := range links {
		links[i] = -1
	}
	if r.Commands == nil {
		return links
	}
	cmds := r.Commands.Cmds
	byFrame := map[uint32][]int{}
	for i, c := range cmds {
		f := uint32(c.BaseCmd().Frame)
		byFrame[f] = append(byFrame[f], i)
	}
	cutoff := parseErrorCutoffs(r.Commands, byFrame)
	stopped := map[uint32]bool{}
	for _, c := range p.Commands {
		f := c.Frame
		if stopped[f] {
			continue
		}
		frameCmds := byFrame[f]
		limit, hasLimit := cutoff[f]
		switch {
		case c.DropReason == downgrade.ReasonUnknownLength || c.DropReason == downgrade.ReasonTruncated,
			hasLimit && c.Position >= limit,
			c.Position >= len(frameCmds):
			stopped[f] = true
			continue
		}
		base := cmds[frameCmds[c.Position]].BaseCmd()
		if base.PlayerID != c.PlayerID || base.Type == nil || !screpIDMatches(c.ID, base.Type.ID) {
			stopped[f] = true
			continue
		}
		if c.Action >= 0 && int(c.Action) < actions {
			links[c.Action] = int32(frameCmds[c.Position])
		}
	}
	return links
}

func parseErrorCutoffs(cs *rep.Commands, byFrame map[uint32][]int) map[uint32]int {
	cutoff := map[uint32]int{}
	if len(cs.ParseErrCmds) == 0 {
		return cutoff
	}
	index := make(map[repcmd.Cmd]int, len(cs.Cmds))
	for i, c := range cs.Cmds {
		index[c] = i
	}
	for _, pec := range cs.ParseErrCmds {
		f := uint32(pec.Frame)
		prev := -1
		if pec.PrevCmd != nil {
			if i, ok := index[pec.PrevCmd]; ok {
				prev = i
			}
		}
		n := 0
		for _, i := range byFrame[f] {
			if i <= prev {
				n++
			}
		}
		if old, ok := cutoff[f]; !ok || n < old {
			cutoff[f] = n
		}
	}
	return cutoff
}
