package openbw

import (
	"errors"
	"fmt"
	"iter"
	"runtime/debug"

	"github.com/marianogappa/openbw-go/bwgame"
	"github.com/marianogappa/openbw-go/downgrade"
	"github.com/marianogappa/openbw-go/scrrules"
)

// Options configures Open.
type Options struct {
	// Data is the game data; nil loads $OPENBW_DATA once per process.
	Data *Data
	// Kinds are the event kinds to emit; 0 means DefaultKinds and NoKinds means none.
	Kinds Kind
	// Classes filters unit events by class; 0 means AllClasses.
	Classes ClassSet
	// TickEvery emits a Tick every N frames; 0 emits no ticks. It must not be negative.
	TickEvery int32
	// MaxFrame stops after processing this frame; 0 plays to the replay end. It must not be negative.
	MaxFrame int32
	// Rules selects the engine rules.
	Rules RulesMode
	// CustomRules overrides Rules when not nil.
	CustomRules *bwgame.Rules
	// Hooks are raw engine callbacks (L1), called after the stream's own hooks. They must not change engine state.
	// They run inside Open for the map units, before Open returns the Game, and then inside Events and AdvanceTo.
	// The Hooks API is package bwgame's and has no stability promise.
	Hooks bwgame.Hooks
}

// Game is one replay being played. It is not safe for concurrent use.
type Game struct {
	kinds    Kind
	classes  ClassSet
	maxFrame int32
	tick     int32

	data     *Data
	prepared *downgrade.Prepared
	st       *bwgame.State
	funcs    *bwgame.ReplayFunctions
	rules    RulesInfo
	sources  []int32
	header   Header
	playing  []int8

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
// The caller must not change replay while the game is in use.
func Open(replay []byte, opts Options) (*Game, error) {
	if opts.MaxFrame < 0 || opts.TickEvery < 0 {
		return nil, errors.New("openbw: negative MaxFrame or TickEvery")
	}
	data, err := dataOrDefault(opts.Data)
	if err != nil {
		return nil, err
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
		prepared:       prepared,
		containerKills: map[UnitID]killerInfo{},
	}
	switch {
	case g.kinds == 0:
		g.kinds = DefaultKinds
	case g.kinds&NoKinds != 0:
		g.kinds = 0
	}
	if g.classes == 0 {
		g.classes = AllClasses
	}
	rules, source := chooseRules(opts, prepared)
	g.rules = rulesInfo(source, rules)
	g.st = bwgame.NewState(data.global, bwgame.NewGameState())
	g.st.Game.Rules = rules
	g.funcs = bwgame.NewReplayFunctions(g.st, &bwgame.ActionState{}, &bwgame.ReplayState{})
	g.install(opts.Hooks)
	if g.wants(KindCommandResult) {
		g.sources = actionSources(prepared.Commands)
	}
	if err := g.load(prepared.Loadable); err != nil {
		return nil, err
	}
	g.commandsNext = true
	g.playing = playingSlots(g.st)
	g.header = readHeader(g.funcs)
	return g, nil
}

type rulesSource uint8

const (
	rulesFromOpenBW rulesSource = iota
	rulesFromSCR
	rulesFromCustom
)

func chooseRules(opts Options, prepared *downgrade.Prepared) (bwgame.Rules, rulesSource) {
	switch {
	case opts.CustomRules != nil:
		return *opts.CustomRules, rulesFromCustom
	case opts.Rules == RulesOpenBW, prepared.Report.SourceFormat == legacyFormat:
		return bwgame.DefaultRules(), rulesFromOpenBW
	}
	return scrrules.FromSettings(prepared.Settings), rulesFromSCR
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
// The loop body runs only when an event arrives; with few Kinds, TickEvery gives a heartbeat to stop early.
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

// AdvanceTo steps the engine until the processing half-step that makes Frame() >= frame, or until Done.
// It yields nothing: events of the subscribed Kinds stay buffered, and Events gives them next.
// With Kinds: NoKinds, nothing is buffered. It returns Err.
func (g *Game) AdvanceTo(frame int32) error {
	for !g.Done() && !(g.commandsNext && g.st.CurrentFrame >= frame) {
		g.step()
	}
	return g.err
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

// Engine gives the engine (L0) for reads. It has no stability promise.
// The caller must not step the engine or change its state; use Events and AdvanceTo.
func (g *Game) Engine() *bwgame.ReplayFunctions { return g.funcs }

// Prepared is the downgrade output: the loadable replay, the SC:R settings, the dropped commands and the command index.
// It has no stability promise.
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
