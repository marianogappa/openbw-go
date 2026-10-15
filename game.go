package openbw

import (
	"fmt"
	"iter"
	"runtime/debug"

	"github.com/icza/screp/rep"
	"github.com/icza/screp/repparser"
	"github.com/marianogappa/openbw-go/bwgame"
	"github.com/marianogappa/openbw-go/downgrade"
	"github.com/marianogappa/openbw-go/scrrules"
)

// StreamOptions configures Open.
type StreamOptions struct {
	// Data is the game data; nil loads $OPENBW_DATA once per process.
	Data *Data
	// Kinds are the event kinds to emit; 0 means DefaultKinds.
	Kinds Kind
	// Classes filters unit events by class; 0 means AllClasses.
	Classes ClassSet
	// TickEvery emits a Tick every N frames; 0 emits no ticks.
	TickEvery int32
	// MaxFrame stops after processing this frame; 0 plays to the replay end.
	MaxFrame int32
	// Rules selects the engine rules.
	Rules RulesMode
	// CustomRules overrides Rules when not nil.
	CustomRules *bwgame.Rules
	// Hooks are raw engine callbacks, called after the stream's own hooks. They must not change state.
	Hooks bwgame.Hooks
}

// Game is one replay being played. It is not safe for concurrent use.
type Game struct {
	kinds    Kind
	classes  ClassSet
	maxFrame int32
	tick     int32

	data     *Data
	replay   *rep.Replay
	prepared *downgrade.Prepared
	st       *bwgame.State
	funcs    *bwgame.ReplayFunctions
	rules    RulesInfo
	links    commandLinks
	playing  []int
	army     []armyCost

	buf          []Event
	head         int
	seq          int64
	err          error
	commandsNext bool
	iterating    bool

	cmd            *commandTracker
	containerKills map[UnitID]killerInfo
}

const legacyFormat = "legacy"

// Open prepares a replay (SC:R or legacy bytes) and returns a game at frame 0.
// The events of the map load (frame 0) come first from Events.
func Open(replay []byte, opts StreamOptions) (*Game, error) {
	data, err := dataOrDefault(opts.Data)
	if err != nil {
		return nil, err
	}
	parsed, err := repparser.ParseConfig(replay, repparser.Config{Commands: true, MapData: true})
	if err != nil {
		return nil, fmt.Errorf("openbw: parse replay: %w", err)
	}
	prepared, err := downgrade.Prepare(replay)
	if err != nil {
		return nil, fmt.Errorf("openbw: prepare replay: %w", err)
	}
	g := &Game{
		kinds:          opts.Kinds,
		classes:        opts.Classes,
		maxFrame:       opts.MaxFrame,
		tick:           opts.TickEvery,
		data:           data,
		replay:         parsed,
		prepared:       prepared,
		containerKills: map[UnitID]killerInfo{},
	}
	if g.kinds == 0 {
		g.kinds = DefaultKinds
	}
	if g.classes == 0 {
		g.classes = AllClasses
	}
	rules := chooseRules(opts, prepared)
	g.rules = rulesInfo(opts, rules)
	g.st = bwgame.NewState(data.global, bwgame.NewGameState())
	g.st.Game.Rules = rules
	g.funcs = bwgame.NewReplayFunctions(g.st, &bwgame.ActionState{}, &bwgame.ReplayState{})
	g.install(opts.Hooks)
	g.links = linkCommands(parsed, prepared)
	if err := g.load(prepared.Loadable); err != nil {
		return nil, err
	}
	g.commandsNext = true
	g.playing = playingSlots(g.st)
	g.army = armyCosts(g.st)
	return g, nil
}

func chooseRules(opts StreamOptions, prepared *downgrade.Prepared) bwgame.Rules {
	switch {
	case opts.CustomRules != nil:
		return *opts.CustomRules
	case opts.Rules == RulesOpenBW, prepared.Report.SourceFormat == legacyFormat:
		return bwgame.DefaultRules()
	}
	return scrrules.FromSettings(prepared.Settings)
}

func (g *Game) load(loadable []byte) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = &EngineError{Frame: g.st.CurrentFrame, Value: r, Stack: debug.Stack()}
		}
	}()
	if err := g.funcs.LoadReplayData(loadable); err != nil {
		return fmt.Errorf("openbw: load replay: %w", err)
	}
	return nil
}

// Events steps the engine as the caller pulls events.
// After a break, a new call continues with the next event. Calling it inside its own loop panics.
func (g *Game) Events() iter.Seq[Event] {
	return func(yield func(Event) bool) {
		if g.iterating {
			panic("openbw: Game.Events called inside its own loop")
		}
		g.iterating = true
		defer func() { g.iterating = false }()
		for {
			for g.head < len(g.buf) {
				e := g.buf[g.head]
				g.buf[g.head] = nil
				g.head++
				if !yield(e) {
					return
				}
			}
			g.buf, g.head = g.buf[:0], 0
			if g.Done() {
				return
			}
			g.step()
		}
	}
}

func (g *Game) step() {
	defer func() {
		if r := recover(); r != nil {
			g.err = &EngineError{Frame: g.st.CurrentFrame, Value: r, Stack: debug.Stack()}
			if g.cmd != nil {
				g.cmd.abort()
			}
		}
	}()
	if g.commandsNext {
		g.commandsNext = false
		g.funcs.ExecuteFrameActions()
		return
	}
	g.commandsNext = true
	g.funcs.ProcessFrame()
	if g.tick > 0 && g.st.CurrentFrame%g.tick == 0 {
		g.emit(&Tick{}, -1)
	}
}

// Err is the error that stopped the game, or nil. Check it after the loop.
func (g *Game) Err() error { return g.err }

// Frame is the engine's current frame.
func (g *Game) Frame() int32 { return g.st.CurrentFrame }

// EndFrame is the replay's last frame.
func (g *Game) EndFrame() int32 { return g.funcs.ReplayState().EndFrame }

// Done tells if the end frame or MaxFrame is reached, or the engine failed.
func (g *Game) Done() bool {
	if g.err != nil {
		return true
	}
	if !g.commandsNext {
		return false
	}
	return g.funcs.IsDone() || g.maxFrame > 0 && g.st.CurrentFrame >= g.maxFrame
}

// Replay is the screp parse of the replay.
func (g *Game) Replay() *rep.Replay { return g.replay }

// Engine gives the engine (L0). It has no stability promise.
func (g *Game) Engine() *bwgame.ReplayFunctions { return g.funcs }

// Prepared is the downgrade output: the loadable replay, the SC:R settings, the drop report and the command index.
func (g *Game) Prepared() *downgrade.Prepared { return g.prepared }

func (g *Game) emit(e Event, cmd int32) {
	b := e.Base()
	g.seq++
	b.Seq = g.seq
	b.Frame = g.st.CurrentFrame
	b.Cmd = cmd
	g.buf = append(g.buf, e)
}

func (g *Game) wants(kind Kind) bool { return g.kinds&kind != 0 }
