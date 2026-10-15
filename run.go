package openbw

import (
	"encoding/json"

	"github.com/icza/screp/rep"
	"github.com/marianogappa/openbw-go/bwgame"
)

// SchemaVersion is the version of Result's JSON shape.
const SchemaVersion = 1

// DefaultStatsEvery is the default player stats period in frames (1 second).
const DefaultStatsEvery = 24

const likelyDesyncSelects = 20

// Options configures Run.
type Options struct {
	// StreamOptions are passed to Open. Run sets TickEvery. Classes 0 here means all but ClassFighter and ClassSpell.
	StreamOptions
	// Commands includes screp's Commands, like screp -cmds.
	Commands bool
	// MapData includes screp's MapData, like screp -map.
	MapData bool
	// CommandResults selects the command results to keep; the default is ResultsActions.
	CommandResults ResultsMode
	// StatsEvery is the player stats period in frames; 0 means DefaultStatsEvery, -1 disables stats.
	StatsEvery int32
	// UnitSnapshots enables unit snapshots; nil means off.
	UnitSnapshots *UnitSnapshotOptions
}

// UnitSnapshotOptions configures unit snapshots.
type UnitSnapshotOptions struct {
	Every          int32
	Fields         UnitField
	IncludeNeutral bool
}

// Result is what Run records. Its JSON is the output of cmd/openbw.
type Result struct {
	SchemaVersion  int
	Header         *rep.Header
	Commands       *rep.Commands `json:",omitempty"`
	MapData        *rep.MapData  `json:",omitempty"`
	Computed       *rep.Computed
	Engine         EngineInfo
	Trust          Trust
	Events         []Event            `json:",omitempty"`
	CommandResults []CommandResultRow `json:",omitempty"`
	Stats          []PlayerStats      `json:",omitempty"`
	UnitSnapshots  []UnitSnapshot     `json:",omitempty"`
}

// Trust has the signals that tell if the engine output can be trusted.
type Trust struct {
	DroppedCommands  int
	DroppedByReason  map[string]int `json:",omitempty"`
	UnlinkedCommands int
	NotExecuted      int
	Desync           DesyncSignal
	Production       []ProductionTrust
}

// DesyncSignal counts selections of units that do not exist or belong to another player.
// Likely is a signal, not a proof.
type DesyncSignal struct {
	InconsistentSelects int
	FirstFrame          int32
	Likely              bool
}

// ProductionTrust counts one player's production commands and how many the engine refused.
type ProductionTrust struct {
	Owner    int8
	PlayerID byte
	Issued   int
	Rejected int
	NoEffect int
}

// CommandResultRow is a CommandResult plus the units it produced.
type CommandResultRow struct {
	CommandResult
	Produced []UnitID `json:",omitempty"`
}

type commandResultRowJSON struct {
	commandResultFields
	Produced []UnitID `json:",omitempty"`
}

func (r CommandResultRow) MarshalJSON() ([]byte, error) {
	return marshalEvent(0, &r.EventBase, commandResultRowJSON{commandResultFields(r.CommandResult), r.Produced})
}

func (r *CommandResultRow) UnmarshalJSON(data []byte) error {
	var v commandResultRowJSON
	if err := unmarshalEvent(data, &r.EventBase, &v); err != nil {
		return err
	}
	base := r.EventBase
	r.CommandResult = CommandResult(v.commandResultFields)
	r.EventBase = base
	r.Produced = v.Produced
	return nil
}

// UnitSnapshot is the units at one frame, in columns. Only the selected fields are present.
type UnitSnapshot struct {
	Frame     int32
	ID        []UnitID
	Type      []uint16
	Owner     []int8
	X         []int32  `json:",omitempty"`
	Y         []int32  `json:",omitempty"`
	HP        []int32  `json:",omitempty"`
	Shields   []int32  `json:",omitempty"`
	Energy    []int32  `json:",omitempty"`
	Order     []uint16 `json:",omitempty"`
	State     []uint16 `json:",omitempty"`
	BuildTime []int32  `json:",omitempty"`
}

// UnmarshalJSON decodes a Result, with its events.
func (r *Result) UnmarshalJSON(data []byte) error {
	type plain Result
	var v struct {
		plain
		Events []json.RawMessage
	}
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	*r = Result(v.plain)
	r.Events = make([]Event, 0, len(v.Events))
	for _, raw := range v.Events {
		e, err := UnmarshalEvent(raw)
		if err != nil {
			return err
		}
		r.Events = append(r.Events, e)
	}
	return nil
}

func gcd(a, b int32) int32 {
	for b != 0 {
		a, b = b, a%b
	}
	return a
}

// Run plays the whole replay (or up to MaxFrame) with the stream and records a Result.
// On an engine error it returns the partial Result and the error.
func Run(replay []byte, opts Options) (*Result, error) {
	stream := opts.StreamOptions
	outputKinds := stream.Kinds
	if outputKinds == 0 {
		outputKinds = DefaultKinds
	}
	outputKinds &^= KindCommandResult | KindTick
	if stream.Classes == 0 {
		stream.Classes = AllClasses &^ Classes(ClassFighter, ClassSpell)
	}
	statsEvery := opts.StatsEvery
	switch {
	case statsEvery == 0:
		statsEvery = DefaultStatsEvery
	case statsEvery < 0:
		statsEvery = 0
	}
	var snapshotEvery int32
	if opts.UnitSnapshots != nil && opts.UnitSnapshots.Every > 0 {
		snapshotEvery = opts.UnitSnapshots.Every
	}
	stream.TickEvery = gcd(statsEvery, snapshotEvery)
	stream.Kinds = outputKinds | KindCommandResult | KindUnitCreated | KindMorphStarted
	g, err := Open(replay, stream)
	if err != nil {
		return nil, err
	}
	rec := &recorder{
		g:             g,
		opts:          opts,
		outputKinds:   outputKinds,
		statsEvery:    statsEvery,
		snapshotEvery: snapshotEvery,
		classes:       stream.Classes,
		rowOfAction:   map[int32]int{},
		produced:      map[int32][]UnitID{},
		production:    map[int8]*ProductionTrust{},
	}
	for e := range g.Events() {
		rec.add(e)
	}
	return rec.result(), g.Err()
}

type recorder struct {
	g             *Game
	opts          Options
	outputKinds   Kind
	statsEvery    int32
	snapshotEvery int32
	classes       ClassSet

	events      []Event
	rows        []CommandResultRow
	rowOfAction map[int32]int
	produced    map[int32][]UnitID
	stats       []PlayerStats
	snapshots   []UnitSnapshot
	desync      DesyncSignal
	production  map[int8]*ProductionTrust
}

func (r *recorder) add(e Event) {
	switch e := e.(type) {
	case *CommandResult:
		r.addCommandResult(e)
		return
	case *Tick:
		r.addTick(e.Frame)
		return
	case *UnitCreated:
		r.addProduced(e.Cmd, e.Unit)
	case *MorphStarted:
		r.addProduced(e.Cmd, e.Unit)
	}
	if r.outputKinds&e.Kind() != 0 {
		r.events = append(r.events, e)
	}
}

func (r *recorder) addProduced(cmd int32, id UnitID) {
	if cmd >= 0 {
		r.produced[cmd] = append(r.produced[cmd], id)
	}
}

func (r *recorder) addCommandResult(c *CommandResult) {
	inconsistent := (c.ActionID == actionSelect || c.ActionID == actionShiftSelect) && c.MissingTags+c.ForeignTags > 0
	if inconsistent {
		r.desync.InconsistentSelects++
		if r.desync.FirstFrame == 0 {
			r.desync.FirstFrame = c.Frame
		}
	}
	if isProductionAction(c.ActionID) {
		p := r.production[c.Owner]
		if p == nil {
			p = &ProductionTrust{Owner: c.Owner, PlayerID: c.PlayerID}
			r.production[c.Owner] = p
		}
		p.Issued++
		switch c.Outcome {
		case Rejected:
			p.Rejected++
		case NoEffect:
			p.NoEffect++
		}
	}
	keep := false
	switch r.opts.CommandResults {
	case ResultsAll:
		keep = true
	case ResultsActions:
		keep = !isSelectAction(c.ActionID) || inconsistent
	}
	if keep {
		r.rowOfAction[c.Action] = len(r.rows)
		r.rows = append(r.rows, CommandResultRow{CommandResult: *c})
	}
}

func (r *recorder) addTick(frame int32) {
	if r.statsEvery > 0 && frame%r.statsEvery == 0 {
		for _, owner := range r.g.playing {
			r.stats = append(r.stats, r.g.Stats(owner))
		}
	}
	if r.snapshotEvery > 0 && frame%r.snapshotEvery == 0 {
		r.snapshots = append(r.snapshots, r.snapshot(frame))
	}
}

func classOfValue(u Unit) UnitClass {
	if u.Hallucination {
		return ClassHallucination
	}
	switch bwgame.UnitTypes(u.Type) {
	case bwgame.UnitTypesProtossInterceptor, bwgame.UnitTypesProtossScarab:
		return ClassFighter
	case bwgame.UnitTypesTerranVultureSpiderMine:
		return ClassSpiderMine
	case bwgame.UnitTypesSpellDarkSwarm, bwgame.UnitTypesSpellDisruptionWeb, bwgame.UnitTypesSpellScannerSweep:
		return ClassSpell
	}
	return ClassNormal
}

func stateBits(u Unit) uint16 {
	var s uint8
	flags := []struct {
		set bool
		bit uint8
	}{
		{u.Completed, StateCompleted}, {u.Burrowed, StateBurrowed}, {u.Cloaked, StateCloaked},
		{u.Loaded, StateLoaded}, {u.Dying, StateDying}, {u.Hallucination, StateHallucination},
	}
	for _, f := range flags {
		if f.set {
			s |= f.bit
		}
	}
	return uint16(s)
}

func (r *recorder) snapshot(frame int32) UnitSnapshot {
	opts := r.opts.UnitSnapshots
	fields := opts.Fields
	s := UnitSnapshot{Frame: frame}
	for u := range r.g.Units() {
		if u.Owner == neutralOwner && !opts.IncludeNeutral || !r.classes.Has(classOfValue(u)) {
			continue
		}
		s.ID = append(s.ID, u.ID)
		s.Type = append(s.Type, uint16(u.Type))
		s.Owner = append(s.Owner, u.Owner)
		if fields&Position != 0 {
			s.X = append(s.X, u.X)
			s.Y = append(s.Y, u.Y)
		}
		if fields&HP != 0 {
			s.HP = append(s.HP, u.HP)
		}
		if fields&Shields != 0 {
			s.Shields = append(s.Shields, u.Shields)
		}
		if fields&Energy != 0 {
			s.Energy = append(s.Energy, u.Energy)
		}
		if fields&Order != 0 {
			s.Order = append(s.Order, uint16(u.Order))
		}
		if fields&State != 0 {
			s.State = append(s.State, stateBits(u))
		}
		if fields&BuildTime != 0 {
			s.BuildTime = append(s.BuildTime, u.RemainingBuild)
		}
	}
	return s
}

func (r *recorder) result() *Result {
	g := r.g
	replay := g.Replay()
	replay.Compute()
	res := &Result{
		SchemaVersion: SchemaVersion,
		Header:        replay.Header,
		Computed:      replay.Computed,
		Engine:        g.Info(),
		Trust:         r.trust(),
		Events:        r.events,
		Stats:         r.stats,
		UnitSnapshots: r.snapshots,
	}
	if r.opts.Commands {
		res.Commands = replay.Commands
	}
	if r.opts.MapData {
		res.MapData = replay.MapData
	}
	for action, ids := range r.produced {
		if i, ok := r.rowOfAction[action]; ok {
			r.rows[i].Produced = ids
		}
	}
	res.CommandResults = r.rows
	return res
}

func (r *recorder) trust() Trust {
	g := r.g
	report := g.prepared.Report
	t := Trust{
		DroppedCommands:  report.DroppedCount(),
		UnlinkedCommands: g.links.unlinked,
		Desync:           r.desync,
	}
	for _, d := range report.Dropped {
		if t.DroppedByReason == nil {
			t.DroppedByReason = map[string]int{}
		}
		t.DroppedByReason[d.Reason] += d.Count
	}
	if g.cmd != nil {
		t.NotExecuted = len(g.links.screp) - int(g.cmd.executed)
	}
	t.Desync.Likely = t.Desync.InconsistentSelects >= likelyDesyncSelects
	playerIDs := g.funcs.ActionState().PlayerID
	for _, owner := range g.playing {
		p := ProductionTrust{Owner: int8(owner), PlayerID: byte(playerIDs[owner])}
		if counted := r.production[int8(owner)]; counted != nil {
			p = *counted
		}
		t.Production = append(t.Production, p)
	}
	return t
}
