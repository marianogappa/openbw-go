package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash"
	"hash/fnv"
	"os"
	"slices"
	"strings"

	openbw "github.com/marianogappa/openbw-go"
	"github.com/marianogappa/openbw-go/bwgame"
	"github.com/marianogappa/openbw-go/downgrade"
	"github.com/marianogappa/openbw-go/scrrules"
)

const (
	checkpointEvery = 24 * 60 * 2
	maxFailures     = 20
)

type life struct {
	typ       openbw.UnitType
	before    openbw.UnitType
	morphedAt int32
	owner     int8
	completed bool
	summoning bool
}

type result struct {
	path           string
	format         string
	err            error
	frames         int32
	events         int
	unitEvents     int
	checkpoints    int
	unitsCompared  int
	commands       int
	dropped        int
	selections     int
	digest         string
	serialDigest   string
	openbwDigest   string
	openbwDiverge  int32
	openbwLoadErr  error
	mapName        string
	rulesCause     []string
	rulesSCR       bool
	failures       []string
	failureCount   int
	failureByCheck map[string]int
}

func (r *result) fail(check, format string, args ...any) {
	if r.failureByCheck == nil {
		r.failureByCheck = map[string]int{}
	}
	r.failureByCheck[check]++
	r.failureCount++
	if len(r.failures) < maxFailures {
		r.failures = append(r.failures, check+": "+fmt.Sprintf(format, args...))
	}
}

type digester struct {
	sum    hash.Hash
	frames []int32
	hashes []uint64
}

func newDigester(keepTrail bool) *digester {
	d := &digester{sum: sha256.New()}
	if keepTrail {
		d.frames, d.hashes = []int32{}, []uint64{}
	}
	return d
}

func (d *digester) add(e openbw.Event) {
	line := fmt.Sprintf("%T %+v\n", e, e)
	d.sum.Write([]byte(line))
	if d.hashes != nil {
		h := fnv.New64a()
		h.Write([]byte(line))
		d.frames = append(d.frames, e.Base().Frame)
		d.hashes = append(d.hashes, h.Sum64())
	}
}

func (d *digester) hex() string { return hex.EncodeToString(d.sum.Sum(nil)) }

func firstDivergence(a, b *digester) int32 {
	n := min(len(a.hashes), len(b.hashes))
	for i := range n {
		if a.hashes[i] != b.hashes[i] {
			return min(a.frames[i], b.frames[i])
		}
	}
	if len(a.hashes) != len(b.hashes) {
		if n < len(a.hashes) {
			return a.frames[n]
		}
		return b.frames[n]
	}
	return -1
}

func engineRules(prepared *downgrade.Prepared, mode openbw.RulesMode) bwgame.Rules {
	if mode == openbw.RulesOpenBW || prepared.Report.SourceFormat == "legacy" {
		return bwgame.DefaultRules()
	}
	return scrrules.FromSettings(prepared.Settings)
}

type validator struct {
	r       *result
	g       *openbw.Game
	alive   map[openbw.UnitID]*life
	dead    map[openbw.UnitID]bool
	lastSeq int64
	lastFr  int32
	results int
	plain   *plainRun
	source  map[int32]*openbw.SourceCommand
	header  openbw.Header
	sources []openbw.SourceCommand
}

func (v *validator) liveUnit(check string, id openbw.UnitID, e openbw.Event) *life {
	if v.dead[id] {
		v.r.fail(check, "frame %d: %T for unit %d after its UnitDestroyed", e.Base().Frame, e, id)
		return nil
	}
	l := v.alive[id]
	if l == nil {
		v.r.fail(check, "frame %d: %T for unit %d before UnitCreated", e.Base().Frame, e, id)
	}
	return l
}

func (v *validator) sameUnit(l *life, typ openbw.UnitType, owner int8, e openbw.Event) {
	if l != nil && (l.typ != typ || l.owner != owner) {
		v.r.fail("event fields", "frame %d: %T says %v of %d, events say %v of %d", e.Base().Frame, e, typ, owner, l.typ, l.owner)
	}
}

func (v *validator) unitEvent(e openbw.Event) {
	v.r.unitEvents++
	switch e := e.(type) {
	case *openbw.UnitCreated:
		if v.alive[e.Unit] != nil || v.dead[e.Unit] {
			v.r.fail("lifecycle", "frame %d: UnitCreated twice for %d", e.Frame, e.Unit)
		}
		v.alive[e.Unit] = &life{typ: e.Type, owner: e.Owner}
	case *openbw.UnitCompleted:
		l := v.liveUnit("lifecycle", e.Unit, e)
		v.sameUnit(l, e.Type, e.Owner, e)
		if l != nil {
			if l.completed && !l.summoning {
				v.r.fail("double completion", "frame %d: %v %d completed while complete", e.Frame, e.Type, e.Unit)
			}
			l.completed, l.summoning = true, false
		}
	case *openbw.UnitMorphed:
		l := v.liveUnit("lifecycle", e.Unit, e)
		v.sameUnit(l, e.From, e.Owner, e)
		if l != nil {
			l.before, l.typ, l.morphedAt = l.typ, e.To, e.Frame
			l.summoning = isArchon(e.To)
		}
	case *openbw.MorphStarted:
		l := v.liveUnit("lifecycle", e.Unit, e)
		if l != nil && l.morphedAt == e.Frame && l.before == e.From {
			v.sameUnit(l, l.typ, e.Owner, e)
		} else {
			v.sameUnit(l, e.From, e.Owner, e)
		}
		if l != nil {
			l.completed = false
		}
	case *openbw.UnitOwnerChanged:
		l := v.liveUnit("lifecycle", e.Unit, e)
		v.sameUnit(l, e.Type, e.From, e)
		if l != nil {
			l.owner = e.To
		}
	case *openbw.UnitDestroyed:
		l := v.liveUnit("lifecycle", e.Unit, e)
		v.sameUnit(l, e.Type, e.Owner, e)
		switch {
		case l == nil:
		case e.Reverted:
			l.completed = true
		default:
			delete(v.alive, e.Unit)
			v.dead[e.Unit] = true
		}
	}
}

func isArchon(t openbw.UnitType) bool {
	return t == openbw.UnitType(bwgame.UnitTypesProtossArchon) || t == openbw.UnitType(bwgame.UnitTypesProtossDarkArchon)
}

type typeOwner struct {
	typ   openbw.UnitType
	owner int8
}

func (v *validator) checkpoint() {
	v.r.checkpoints++
	frame := v.g.Frame()
	engine := map[openbw.UnitID]openbw.Unit{}
	byTypeOwner := map[typeOwner]int{}
	for u := range v.g.Units() {
		if u.Dying {
			continue
		}
		engine[u.ID] = u
		byTypeOwner[typeOwner{u.Type, u.Owner}]++
	}
	raw := map[typeOwner]int{}
	st := v.g.Engine().State()
	for owner := range st.PlayerUnits {
		for u := range st.PlayerUnits[owner].All() {
			if u.UnitType.Flags&bwgame.UnitTypeTFlagTurret != 0 || u.OrderType == nil || u.OrderType.ID == bwgame.OrdersDie {
				continue
			}
			raw[typeOwner{openbw.UnitType(u.UnitType.ID), int8(u.Owner)}]++
		}
	}
	events := map[typeOwner]int{}
	for id, l := range v.alive {
		events[typeOwner{l.typ, l.owner}]++
		u, ok := engine[id]
		if !ok {
			v.r.fail("alive units", "frame %d: events have %v %d of %d alive, engine does not", frame, l.typ, id, l.owner)
			continue
		}
		if u.Completed != l.completed {
			v.r.fail("completed flag", "frame %d: %v %d engine completed %v, events %v", frame, u.Type, id, u.Completed, l.completed)
		}
	}
	for id, u := range engine {
		if v.alive[id] == nil {
			v.r.fail("alive units", "frame %d: engine has %v %d of %d, events do not", frame, u.Type, id, u.Owner)
		}
	}
	for _, other := range []map[typeOwner]int{byTypeOwner, raw} {
		for k, n := range other {
			if events[k] != n {
				v.r.fail("type and owner counts", "frame %d: %v of %d: events %d, engine %d", frame, k.typ, k.owner, events[k], n)
			}
		}
		for k, n := range events {
			if other[k] != n {
				v.r.fail("type and owner counts", "frame %d: %v of %d: events %d, engine %d", frame, k.typ, k.owner, n, other[k])
			}
		}
	}
	v.r.unitsCompared += len(engine)
}

func (v *validator) commandResult(e *openbw.CommandResult) {
	i := v.results
	v.results++
	if e.Action != int32(i) || e.Cmd != e.Action {
		v.r.fail("command order", "result %d has Action %d Cmd %d", i, e.Action, e.Cmd)
		return
	}
	playerID := v.header.Slots[e.Owner].PlayerID
	if src := v.source[e.Action]; src == nil {
		v.r.fail("command source", "action %d has no source command", e.Action)
	} else if int32(src.Frame) != e.Frame || int32(src.PlayerID) != playerID {
		v.r.fail("command source", "action %d: result frame %d player %d, source frame %d player %d", e.Action, e.Frame, playerID, src.Frame, src.PlayerID)
	} else if e.Source < 0 || int(e.Source) >= len(v.sources) || v.sources[e.Source].Action != e.Action {
		v.r.fail("command source", "action %d: Source %d is not its source command", e.Action, e.Source)
	}
	if i >= len(v.plain.actions) {
		v.r.fail("command engine", "action %d not executed by the plain engine", i)
		return
	}
	p := v.plain.actions[i]
	if p.frame != e.Frame || p.owner != int32(e.Owner) || p.actionID != e.ActionID || p.ok != (e.Outcome != openbw.Rejected) {
		v.r.fail("command engine", "action %d: result frame %d owner %d id %#x %v; engine frame %d owner %d id %#x ok %v",
			i, e.Frame, e.Owner, e.ActionID, e.Outcome, p.frame, p.owner, p.actionID, p.ok)
	}
	got := make([]uint32, len(e.Selection))
	for j, id := range e.Selection {
		got[j] = uint32(id)
	}
	if !slices.Equal(got, p.selection) {
		v.r.fail("selection", "action %d %#x at frame %d: result %v, engine %v", i, e.ActionID, e.Frame, got, p.selection)
	}
	v.r.selections++
}

func (v *validator) event(e openbw.Event) {
	v.r.events++
	b := e.Base()
	if b.Seq != v.lastSeq+1 || b.Frame < v.lastFr {
		v.r.fail("order", "%T seq %d frame %d after seq %d frame %d", e, b.Seq, b.Frame, v.lastSeq, v.lastFr)
	}
	v.lastSeq, v.lastFr = b.Seq, b.Frame
	if b.Cmd >= int32(v.results)+1 {
		v.r.fail("command link", "frame %d: %T linked to future action %d", b.Frame, e, b.Cmd)
	}
	switch e := e.(type) {
	case *openbw.CommandResult:
		v.commandResult(e)
	case *openbw.Tick:
		v.checkpoint()
	default:
		v.unitEvent(e)
	}
}

func validateReplay(data *openbw.Data, path string) *result {
	r := &result{path: path, openbwDiverge: -1}
	replay, err := os.ReadFile(path)
	if err != nil {
		r.err = err
		return r
	}
	prepared, err := downgrade.Prepare(replay)
	if err != nil {
		r.err = fmt.Errorf("prepare: %w", err)
		return r
	}
	r.format = prepared.Report.SourceFormat
	global, err := plainGlobal()
	if err != nil {
		r.err = err
		return r
	}
	plain, err := runPlainEngine(global, prepared, engineRules(prepared, openbw.RulesAuto))
	if err != nil {
		r.err = fmt.Errorf("plain engine: %w", err)
		return r
	}
	g, err := openbw.Open(replay, openbw.Options{Data: data, Kinds: openbw.AllKinds, Classes: openbw.AllClasses, TickEvery: checkpointEvery, Engine: &openbw.EngineOptions{}})
	if err != nil {
		r.err = err
		return r
	}
	r.rulesSCR = g.Info().Rules.SCR
	r.mapName = cleanMapName(g.Header().MapName)
	v := &validator{r: r, g: g, header: g.Header(), sources: g.SourceCommands(), alive: map[openbw.UnitID]*life{}, dead: map[openbw.UnitID]bool{}, plain: plain, source: map[int32]*openbw.SourceCommand{}}
	for i := range v.sources {
		c := &v.sources[i]
		if c.Action >= 0 {
			v.source[c.Action] = c
		} else {
			r.dropped++
		}
	}
	d := newDigester(true)
	for e := range g.Events() {
		d.add(e)
		v.event(e)
	}
	if err := g.Err(); err != nil {
		r.err = err
		return r
	}
	v.checkpoint()
	r.frames = g.Frame()
	r.commands = v.results
	r.digest = d.hex()
	if v.results != len(plain.actions) {
		r.fail("command count", "%d results, plain engine executed %d actions", v.results, len(plain.actions))
	}
	for action := range v.source {
		if action >= int32(v.results) && int32(v.source[action].Frame) <= g.Frame() {
			r.fail("command count", "source action %d at frame %d has no result", action, v.source[action].Frame)
		}
	}
	if g.Frame() != plain.frame || bwgame.StateHash(g.Engine().State()) != plain.hash {
		r.fail("engine state", "stream ends at frame %d hash %x, plain engine frame %d hash %x", g.Frame(), bwgame.StateHash(g.Engine().State()), plain.frame, plain.hash)
	}
	if want := r.format != "legacy"; r.rulesSCR != want {
		r.fail("rules", "SC:R rules %v for format %s", r.rulesSCR, r.format)
	}
	off := newDigester(true)
	r.openbwDigest, err = streamDigest(data, replay, openbw.Options{Rules: openbw.RulesOpenBW}, off)
	switch {
	case err != nil && r.rulesSCR && strings.Contains(err.Error(), "disable thingy unit"):
		r.openbwLoadErr = err
	case err != nil:
		r.fail("rules on and off", "OpenBW rules: %v", err)
	case r.openbwDigest != r.digest:
		r.openbwDiverge = firstDivergence(d, off)
		if !r.rulesSCR {
			r.fail("rules on and off", "legacy replay streams differ at frame %d", r.openbwDiverge)
		} else {
			r.rulesCause = explainRules(data, replay, prepared, r.digest)
		}
	}
	return r
}

func explainRules(data *openbw.Data, replay []byte, prepared *downgrade.Prepared, scrDigest string) []string {
	scr := engineRules(prepared, openbw.RulesAuto)
	singles := []struct {
		name string
		set  func(*bwgame.Rules)
	}{
		{"object limits", func(r *bwgame.Rules) { r.Limits = scr.Limits }},
		{"disabled thingy units", func(r *bwgame.Rules) { r.DisabledThingyUnits = scr.DisabledThingyUnits }},
		{"order allocation", func(r *bwgame.Rules) { r.OrderAllocationRule = scr.OrderAllocationRule }},
		{"flying SCV fix", func(r *bwgame.Rules) { r.FlyingSCVFix = scr.FlyingSCVFix }},
	}
	var causes []string
	for _, s := range singles {
		rules := bwgame.DefaultRules()
		s.set(&rules)
		digest, err := streamDigest(data, replay, openbw.Options{Engine: &openbw.EngineOptions{Rules: &rules}}, newDigester(false))
		if err == nil && digest == scrDigest {
			causes = append(causes, s.name)
		}
	}
	return causes
}

func streamDigest(data *openbw.Data, replay []byte, opts openbw.Options, d *digester) (string, error) {
	opts.Data, opts.Kinds, opts.Classes, opts.TickEvery = data, openbw.AllKinds, openbw.AllClasses, checkpointEvery
	g, err := openbw.Open(replay, opts)
	if err != nil {
		return "", err
	}
	for e := range g.Events() {
		d.add(e)
	}
	if err := g.Err(); err != nil {
		return "", err
	}
	return d.hex(), nil
}
