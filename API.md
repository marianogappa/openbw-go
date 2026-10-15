# Public API spec

Status: implemented in package `openbw` (module root) and `cmd/openbw`. It refines [DESIGN.md](DESIGN.md) into Go signatures. Where the implementation differs from the first draft, the text below is updated and [Implementation notes](#implementation-notes) lists the change.

Package `openbw`, import `github.com/marianogappa/openbw-go`. The output is JSON in screp's style. L3 analyses are out of scope, except one example program.

```
L2  openbw.Run(rep, Options) → *Result → JSON (cmd/openbw)
L1  openbw.Open(rep, StreamOptions) → *Game; for e := range g.Events()
    bwgame.Hooks (raw engine callbacks under the stream)
L0  g.Engine() → *bwgame.ReplayFunctions (no stability promise)
```

Hard rule: no hook changes engine behaviour. Every engine change below only adds `if f.Hooks.X != nil { ... }` calls, hook-only fields, local variables that hooks read (for example the type before `morphUnit`), and two exported methods that split `ReplayFunctions.nextFrame` into its halves (2). `tools/difftest` must stay 20/20 in default mode and with `-scr-rules` against the patched harness.

## 1. Loading

### Game data

```go
// Data is the game data: .dat files, iscript, GRP frame sizes, tilesets.
// One Data is safe for concurrent use by any number of games.
type Data struct{ /* global *bwgame.GlobalState; filesHash string */ }

// LoadData reads the game data from dir. An empty dir means $OPENBW_DATA.
func LoadData(dir string) (*Data, error)

// FilesHash is the SHA-256 of every data file, in load order, as hex.
func (d *Data) FilesHash() string

// Global gives the engine data (L0).
func (d *Data) Global() *bwgame.GlobalState

// ErrNoData is returned when no directory is given and OPENBW_DATA is not set.
var ErrNoData = errors.New("openbw: no game data: set Options.Data or OPENBW_DATA")
```

- `Options.Data == nil` uses a process-wide `Data` from `LoadData("")`, loaded once (`sync.Once`). A load error is returned by every `Open`/`Run` call.
- `FilesHash` is computed while `bwgame.GlobalInit` reads the files: the loader function is wrapped. It goes into every `Result`.

**Is `bwgame.GlobalState` safe to share? Yes, with no change.**

- The engine never writes to it. No assignment to `globalSt.*` or `Global.*` exists outside `globalInit`. Pointers into it (`ImageTypes.Vec`, `ImageLoOffsets`, iscript) are only read.
- Per-game data is already per game: `map_load.go` parses `UnitsDat` into `GameState.UnitTypes` for each game. `TilesetVf4/Cv5` and `MeleeTrg` are read through new `DataReader`s.
- Package-level variables in `bwgame` are constant lookup tables. `giveUnitToGrants` is a map that is only read.
- Test: 4 full games in parallel goroutines on one `GlobalState`, built with `-race`: no race reported (7,369-17,257 frames each).

What must change: nothing in `bwgame`. `tools/desync/gotags` loads one `GlobalState` per worker; it can share one. Add a `-race` test that runs 4 replays in parallel on one `Data`, so a future write to the global state fails CI.

### Replay input

```go
// Open prepares a replay (SC:R or legacy bytes) and returns a game at frame 0.
func Open(replay []byte, opts StreamOptions) (*Game, error)
```

`Open` does, in order:

1. `repparser.ParseConfig(replay, Config{Commands: true, MapData: true})` (screp). Needed for the header and for command indexes.
2. `downgrade.Prepare(replay)` (new, see 5.1): the loadable legacy bytes, the SC:R settings, the drop report and the command index.
3. Rules: `RulesAuto` (default) sets `GameState.Rules = scrrules.FromSettings(settings)`. For a legacy replay the settings are the 1.16 defaults, so the rules are `bwgame.DefaultRules()`: exact OpenBW. `RulesOpenBW` forces `DefaultRules()` on an SC:R replay. `CustomRules` overrides both.
4. `LoadReplayData`. Map load creates the initial units. Their events are frame 0 and come first from `Events()`.

```go
type RulesMode uint8

const (
	RulesAuto   RulesMode = iota // SC:R rules for SC:R replays, OpenBW rules for legacy replays
	RulesOpenBW                  // exact C++ OpenBW for every replay
)
```

## 2. The stream

```go
type StreamOptions struct {
	Data        *Data         // nil: LoadData("") once per process
	Kinds       Kind          // event kinds to emit; 0 means DefaultKinds
	Classes     ClassSet      // unit classes for unit events; 0 means AllClasses
	TickEvery   int32         // emit Tick every N frames; 0: no ticks
	MaxFrame    int32         // stop after processing this frame; 0: the replay end
	Rules       RulesMode
	CustomRules *bwgame.Rules // overrides Rules when not nil
	Hooks       bwgame.Hooks  // raw L1 hooks, called after the stream's own hooks
}

// Game is one replay being played. It is not safe for concurrent use.
type Game struct{ /* unexported */ }

// Events steps the engine as the caller pulls events.
// After a break, a new call continues with the next event.
func (g *Game) Events() iter.Seq[Event]

// Err is the error that stopped the game, or nil. Check it after the loop.
func (g *Game) Err() error

func (g *Game) Frame() int32    // engine CurrentFrame
func (g *Game) EndFrame() int32 // the replay's last frame
func (g *Game) Done() bool      // EndFrame or MaxFrame reached, or Err != nil

func (g *Game) Replay() *rep.Replay          // the screp parse
func (g *Game) Info() EngineInfo             // see 6
func (g *Game) Unit(id UnitID) (Unit, bool)  // false if the unit was removed
func (g *Game) Units() iter.Seq[Unit]        // units not yet removed (dying ones too), owner 0..11, engine list order
func (g *Game) Player(owner int) Player      // owner 0..11
func (g *Game) UnitTypeInfo(t UnitType) UnitTypeInfo // this game's units.dat values (a map can change them)
func (g *Game) Engine() *bwgame.ReplayFunctions // L0
```

The value types read from live state:

```go
// UnitType is a units.dat ID. It marshals to JSON as screp's {"Name", "ID"} object.
type UnitType uint16

type Unit struct {
	ID              UnitID
	Type            UnitType
	Owner           int8
	X, Y            int32
	HP, Shields     int32 // integer part
	Energy          int32
	Order           uint8
	RemainingBuild  int32
	Container       UnitID // the transport, Bunker or Carrier/Reaver; 0 if none
	Completed, Burrowed, Cloaked, Loaded, Dying, Hallucination bool
}

type Player struct {
	Owner                       int8
	Race                        uint8
	Minerals, Gas               int32
	GatheredMinerals, GatheredGas int32
	SupplyUsed, SupplyMax       float64
	VictoryState                int32
	Left                        bool
}

type UnitTypeInfo struct {
	Minerals, Gas, BuildTime int32
	Supply                   float64 // supply used; 0.5 for Zergling and Scourge
	Building, Worker, Flyer  bool
}

// Stats is the PlayerStats row that Run records, computed now.
func (g *Game) Stats(owner int) PlayerStats
```

`Stats` is public so that stream users get the same army and worker numbers as `Run`, with no second copy of the formula.

Example:

```go
g, err := openbw.Open(replay, openbw.StreamOptions{Kinds: openbw.KindUnitDestroyed, MaxFrame: 10 * 60 * 24})
if err != nil {
	return err
}
for e := range g.Events() {
	d := e.(*openbw.UnitDestroyed)
	if d.Killer != 0 {
		fmt.Println(d.Frame, d.Type, "killed by", d.KillerType)
	}
}
return g.Err()
```

### Lazy stepping

- `Events()` yields buffered events. When the buffer is empty, it runs one half-step (below), then yields what that half-step produced.
- The engine never runs ahead of the caller by more than one half-step.
- Breaking the loop leaves the game paused. Calling `Events()` again continues with the next buffered event. The prototype checked this.
- The stream ends when `Done()` is true. `MaxFrame = M` stops after the processing half-step that makes `Frame() == M`. Commands of frame `M` are not executed. This is `obwhash -max-frames` behaviour.
- There is no `Close`: the stream holds no goroutine, so an abandoned `Game` is garbage.
- Each event is a new value that the caller owns. The stream never reuses it, so callers can keep it.
- Calling `Events()` inside its own loop body panics.
- Frame-by-frame stepping (a viewer): subscribe `KindTick` with `TickEvery: 1` and read state on each `Tick`.

The two halves need new exported methods in `bwgame`, with the same code and order as `nextFrame`:

```go
func (f *ReplayFunctions) ExecuteFrameActions() // the actions of CurrentFrame
func (f *ReplayFunctions) ProcessFrame()        // StateFunctions.nextFrame: CurrentFrame++, processFrame, processTriggers
```

They do not recover. `Game` wraps each call in its own `recover` because `recoverEngineError` re-panics runtime errors.

### Filtering

- `Kinds` is a bit set. Hooks for kinds that are not subscribed are not installed. An unsubscribed kind costs one nil check at each engine call site.
- Some kinds need internal hooks: `CommandResult` needs `BeforeAction`, `AfterAction`, the reject hook and the build queue hooks (5.3). They are installed only when `KindCommandResult` is subscribed. `EventBase.Cmd` is set only then; otherwise it is -1 on every event.
- Call sites that set hook context (kill reason, owner reason) do it through small helpers such as `killUnitWith(u, ctx)`, which store the context only when the hook is not nil.
- `StreamOptions.Hooks` gives raw engine callbacks (for example a new `OnUnitDamaged`) to people who need what the events leave out. They run in the middle of engine functions: read state, never change it.
- `Classes` filters unit events (`UnitCreated`, `UnitCompleted`, `UnitMorphed`, `MorphStarted`, `UnitDestroyed`, `UnitOwnerChanged`) by `UnitClass`. A filtered event is not allocated.

### Frame semantics

The engine runs each frame as `ReplayFunctions.nextFrame`: it executes the commands with frame `F` while `CurrentFrame == F`, then `StateFunctions.nextFrame` does `CurrentFrame++`, `processFrame` and `processTriggers`. `Game` calls the two halves separately (same package code, same order), so there are two yield points per frame:

| Half-step | `Frame()` during it | `Event.Frame` | Events |
|---|---|---|---|
| Load | 0 | 0 | Map and start units: `UnitCreated`, `UnitCompleted` |
| Commands of frame F | F | F | `CommandResult` and the events that commands cause synchronously (research start, cancels, player left) |
| Processing of frame F+1 | F+1 | F+1 | Everything units, bullets and orders do; then trigger results (`PlayerDefeated`, `PlayerWon`); then `Tick` if `(F+1) % TickEvery == 0` |

- `Event.Frame` is `CurrentFrame` when the engine emits it. Command events have the command's own frame, as in screp.
- Order: events are in engine call order. `Seq` is strictly increasing and `Frame` never decreases.
- A command's `CommandResult` comes after the events it caused (they have `Cmd` equal to its `Action`).
- **State at this point** is the engine state at the end of the half-step that produced the event. All events of one half-step see the same state. A unit created and killed in one half-step is not in `Units()` when you get its `UnitCreated`. Event fields hold the values at emission time.
- `Unit`, `Units` and `Player` read live state. Call them only between events (inside the loop body or after it). The result is a copy.

### Errors

- `Open` returns parse, downgrade and map load errors.
- During the stream, the half-step function recovers every panic from the engine: `bwgame` engine errors and runtime panics (nil dereference, index out of range). The loop stops and `Err()` returns `*EngineError`.
- A panic in the caller's loop body is not recovered: `yield` is called outside the recover scope. The prototype checked this.

```go
type EngineError struct {
	Frame int32  // CurrentFrame when the engine failed
	Value any    // the recovered value; an error for bwgame errors
	Stack []byte // the engine goroutine stack
}

func (e *EngineError) Error() string
func (e *EngineError) Unwrap() error // Value if it is an error
```

### Rejected stream shapes

See "Decisions and alternatives rejected": channels, callbacks, `iter.Seq2[Event, error]`, a `Next()` method.

## 3. Events

```go
// Event is one of the *UnitCreated, *UnitCompleted, ... types below.
type Event interface {
	Base() *EventBase
	Kind() Kind
	isEvent() // sealed: only this package defines events
}

// UnmarshalEvent decodes one event of a Result's JSON by its "Kind".
func UnmarshalEvent(data []byte) (Event, error)

type EventBase struct {
	Seq   int64 `json:"-"` // 1, 2, 3, ... per game
	Frame int32
	Cmd   int32 // the Action of the command this event is linked to exactly (5.3), or -1
}

type Kind uint32

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
```

`Owner` fields are engine slots (0-11, 11 is neutral): the index in screp's `Header.Slots`. The engine reads slot `i` of the replay header into player `i`. This is usually screp's `Player.SlotID`, but the index is the definition. Commands carry screp's `Player.ID`; the engine maps it to the owner with `ownerOfPlayerID`.

### Unit classes

Every unit event has a `Class`. Nothing is left out of the engine view. Callers and `Run` filter by class.

```go
type UnitClass uint8

const (
	ClassNormal        UnitClass = iota
	ClassFighter                 // Interceptor, Scarab
	ClassSpiderMine
	ClassHallucination           // status_flag_hallucination, set at creation, never cleared
	ClassSpell                   // Dark Swarm, Disruption Web, Scanner Sweep
)

// ClassSet is a bit set: bit c is UnitClass c.
type ClassSet uint8

const AllClasses ClassSet = 1<<(ClassSpell+1) - 1
```

The class comes from the call site and the type, not from unit flags, so it is known when `UnitCreated` fires: `SourceHallucination` gives `ClassHallucination`, Interceptor and Scarab give `ClassFighter`, and so on.

- Turret subunits (Tank and Goliath turrets) are never reported. They live and die with their parent.
- Neutral units (minerals, geysers, critters, map doodad units) are `ClassNormal` with `Owner` 11.

### Engine hooks

New fields in `bwgame.Hooks`. `ReplayFunctions` copies its `Hooks` into the `GameLoadFunctions` that `loadReplay` creates, so map load is reported. Today map load uses a separate `StateFunctions` with empty hooks.

```go
OnUnitCreated       func(u *Unit, source CreateSource, parent *Unit)
OnUnitCompleted     func(u *Unit)
OnUnitMorphed       func(u *Unit, from *UnitTypeT)
OnMorphStarted      func(u, morpher *Unit, from, into *UnitTypeT)
OnKillUnit          func(u *Unit, kill KillContext) // replaces OnKillUnit(u)
OnOwnerChanged      func(u *Unit, from int32, reason OwnerReason)
OnProductionCancel  func(c ProductionCancel)
OnResearch          func(u *Unit, r Research, level int32, completed bool)
OnPlayerLeft        func(owner int32, dropped bool)
OnPlayerEliminated  func(owner int32)          // exists
OnVictoryState      func(owner, state int32)   // exists
OnActionRejected    func(u *Unit, reason RejectReason)
OnBuildQueue        func(u *Unit, op BuildQueueOp, index int)
OnAddonBuildType    func(u *Unit, set bool)
OnUnitDamaged       func(u, source *Unit, sourceOwner int32, damage Fp8) // raw only; no event kind
```

Every unit event (`UnitCreated`, `UnitCompleted`, `UnitMorphed`, `MorphStarted`, `UnitDestroyed`, `UnitOwnerChanged`) has a `Class UnitClass` field. The struct listings below show it only on `UnitCreated`.

Implemented in `bwgame/hooks.go`. `OnMorphStarted` also passes `from`, because the unit's type has already changed to the egg or building when it fires. `CancelReason` has a zero value `CancelOther` for cancels the engine makes on its own (a lifted building with an addon in construction, a failed training unit placement). Hook context (kill reason, owner reason, cancel reason) is stored only when its hook is set.

`ReplayFunctions` gets `AfterAction func(owner, actionID int32, ok bool)` next to the existing `BeforeAction`. `nextFrame` must take the reporting path (`executeActionsReportingFails`) when `AfterAction` is set, as it does for `BeforeAction`.

### UnitCreated

```go
type UnitCreated struct {
	EventBase
	Unit   UnitID
	Type   UnitType
	Owner  int8
	X, Y   int32
	Parent UnitID // 0 if none; see the table
	Source CreateSource
	Class  UnitClass
}
```

Fires at each `createUnit` call site, right after `createUnit` returns a unit, before the caller calls anything else. The call site passes the source and the parent; the class follows from them. It does not fire inside `createUnit`, because the source and the parent are not known there. A failed `createUnit` (nil) fires nothing.

It must fire before the caller continues because many callers call `completeUnit` at once (`spawnLarva`, `createInitialUnit`, `spawnBroodlings`, `orderPlaceMine`, `orderCastScannerSweep`, `triggerCreateUnit`, the returned Drone). A hook "after the caller finishes the unit" would emit `UnitCompleted` before `UnitCreated`.

A trained unit is created when its production **starts**, not when it finishes: `secondaryOrderTrain` creates it hidden and incomplete. `UnitCompleted` comes later.

| `Source` | Call site | `Parent` |
|---|---|---|
| `SourceMap` | `mapDataLoader.tagUnit`, `tagThg2` (`createInitialUnit`) | 0 |
| `SourceStart` | `createStartingUnit` (melee start units) | 0 |
| `SourceTrained` | `secondaryOrderTrain`, `secondaryOrderTrainFighter` | the producer (building, Carrier, Reaver) |
| `SourceBuilt` | `startPlacedBuilding` (SCV), `warpInProtossBuilding` (Probe), `placeAddonAtTarget`, `orderBuildNydusExit` | the SCV, the Probe, the building, the Nydus Canal |
| `SourceHallucination` (recursive) | `addHallucinatedFighter` (a hallucinated Reaver's Scarabs) | the hallucinated Reaver |
| `SourceLarva` | `spawnLarva` | the Hatchery, Lair or Hive |
| `SourceSecondEggUnit` | `spawnSecondEggUnit` (second Zergling or Scourge) | the first unit of the egg |
| `SourceBroodling` | `spawnBroodlings` | the Queen |
| `SourceHallucination` | `createHallucination` | the unit that is copied |
| `SourceSpell` | `createSpellUnit` (Dark Swarm, Disruption Web, owner 11) | 0 (the engine does not pass the caster) |
| `SourceScanner` | `orderCastScannerSweep` | the Comsat Station |
| `SourceSpiderMine` | `orderPlaceMine` | the Vulture |
| `SourceDroneReturned` | `cancelZergBuildingConstruction` for an Extractor (a new Drone) | the Extractor |
| `SourceTrigger` | `triggerCreateUnit` | 0 |

`UnitCreated` is always the first event of a unit. Where a call site changes the new unit with an evented function (`triggerCreateUnit` gives an addon to 11, or kills the unit when placement fails), that change comes after `UnitCreated`.

Every unit with a `UnitCreated` gets a `UnitDestroyed`, also when the engine removes it without `killUnit`. One site does that: `placeHallucination` calls `destroyUnit` when it cannot place the hallucination. A hook call at that `destroyUnit` emits `UnitDestroyed` with `ReasonRemoved`. The other `destroyUnit` calls are for turret subunits, which are never reported, and inside `createUnit` before it returns.

Not created, so no `UnitCreated`:

- **Zerg buildings.** The Drone becomes the building: `orderDroneBuild` calls `morphUnit`. You get `MorphStarted` and `UnitMorphed` on the Drone's ID.
- **Refinery, Extractor, Assimilator.** `buildRefinery` takes the neutral Vespene Geyser: `UnitOwnerChanged` (11 to the player), then `UnitMorphed` (Geyser to Refinery) on the geyser's ID.
- **Eggs, cocoons, Lurker eggs, Archons, siege mode.** These are morphs of an existing unit.

### UnitCompleted

```go
type UnitCompleted struct {
	EventBase
	Unit  UnitID
	Type  UnitType // the type at completion
	Owner int8
}
```

Fires at the end of `completeUnit`. One unit can complete more than once, once for each type it completes as: Larva, then Zergling; Hatchery, then Lair, then Hive; a Drone again after its building is cancelled; a Hatchery again after a cancelled Lair morph. No event for:

- Archons and Dark Archons: the merging Templar was complete, and `morphUnit` keeps the completed flag.
- A geyser after its refinery dies: `revertRefineryToGeyser` sets the flag directly.

### UnitMorphed and MorphStarted

```go
type UnitMorphed struct {
	EventBase
	Unit     UnitID
	From, To UnitType
	Owner    int8
}

// MorphStarted fires when a Zerg morph is paid for.
type MorphStarted struct {
	EventBase
	Unit    UnitID   // the unit that becomes Into
	Morpher UnitID   // the Drone for buildings; equals Unit for the other morphs
	From    UnitType // Unit's type before the morph started
	Into    UnitType
	Owner   int8
}
```

`UnitMorphed` fires at the end of `morphUnit`, every time the type changes:

- Larva → Egg → Zergling, Hydralisk → Lurker Egg → Lurker, Mutalisk → Cocoon → Guardian.
- Drone → Hatchery (incomplete) at build start. Hatchery → Lair at the **end** of the morph: the type stays Hatchery while it morphs.
- High Templar → Archon at merge (`mergeArchon`). The other Templar gets `UnitDestroyed` with `ReasonRemoved`.
- Siege Tank Tank Mode ↔ Siege Mode on each siege and unsiege.
- Vespene Geyser → Refinery (built), Refinery → Vespene Geyser (destroyed, see `Reverted`).
- Cancels: Cocoon → Mutalisk, Lurker Egg → Hydralisk (`revertEggMorph`), building → Drone (`cancelZergBuildingConstruction`).
- Infest Command Center: Command Center → Infested Command Center, with `UnitOwnerChanged`. `orderInfestedCommandCenter` does not call `morphUnit`: it assigns `u.UnitType` directly. It needs its own `OnUnitMorphed` call after that assignment, before `setUnitOwner`.

`OnUnitMorphed` is called at the end of `morphUnit` for every call, with the type read at its start. It is not filtered by "type changed", because no caller passes the current type.

`MorphStarted` exists because two morph starts have no type change or hide the target type:

| Call site | `Unit` | `Morpher` | `From` → `Into` |
|---|---|---|---|
| `orderZergUnitMorph` state 0, after the morph to the egg | the egg | same | Larva → Zergling (the egg's `BuildQueue` front) |
| `orderZergBuildingMorph`, after the cost is paid | the building | same | Hatchery → Lair |
| `orderDroneBuild`, after `morphUnit` | the Drone/building | same | Drone → Spawning Pool |
| `orderDroneBuild` for an Extractor, after `buildRefinery` | the Extractor (former geyser) | the Drone | Vespene Geyser → Extractor |

So every Zerg production start is one `MorphStarted`, and every non-Zerg production start is one `UnitCreated` (or `UnitMorphed` Geyser → Refinery/Assimilator).

A paid-for morph has a `MorphStarted`; a refused one has none. `orderZergBuildingMorph` without supply or resources sends the building back to idle and leaves the type in its build queue. That is neither a start nor a `ProductionCancelled`; the command shows it only as `Accepted` with no `Produced` unit.

### UnitDestroyed

```go
type UnitDestroyed struct {
	EventBase
	Unit         UnitID
	Type         UnitType
	Owner        int8
	X, Y         int32
	Reason       DeathReason
	Killer       UnitID   // 0 if none
	KillerType   UnitType // valid if Killer != 0
	KillerOwner  int8     // -1 if unknown; can be set when Killer is 0 (Irradiate after the caster died)
	KillerParent UnitID   // the Carrier or Reaver when Killer is an Interceptor or Scarab
	Container    UnitID   // the transport or bunker when Reason is ReasonContainer
	Reverted     bool     // a refinery: the ID stays alive as a neutral Vespene Geyser
}
```

Fires in `killUnit` (C++ `kill_unit`), when the engine decides the unit dies. The engine removes the unit later, in `orderDie`. Between the two, `Unit()` returns it with `Dying == true`.

`killUnit` has no guard: the engine can call it again on a dying unit (for example `destroyLoadedUnits` or `destroyRegularUnit` on a unit that `unitDealDamage` already killed). So `killUnit` reads `u.OrderType.ID == OrdersDie` at its start, before `setUnitOrderPos`, and fires only when that was false. The existing `onKillUnit` call at the end stays as it is. Result: one `UnitDestroyed` per unit ID, except a geyser, which gets one with `Reverted` each time a refinery on it dies.

The reason comes from a hook-only `killContext` field on `StateFunctions`. Call sites set it before `killUnit` and clear it after. A site that sets nothing gives `ReasonOther`.

| `Reason` | Call sites | Killer |
|---|---|---|
| `ReasonDamage` | `unitDealDamage`: weapons, Psionic Storm, Irradiate, nukes. Plague never kills: `updateUnitStatusTimers` checks `Hp > damage`. | `sourceUnit`, `sourceOwner` as passed |
| `ReasonContainer` | `unitDealDamage` loop over loaded units, `destroyLoadedUnits` (units that a Bunker cannot unload) | the container's killer; `Container` set |
| `ReasonSpell` | `consumeUnit`, `spawnBroodlings` (the target), hallucinations killed by spells (`blindUnit`, `restoreUnit`, `empShockwave`, `maelstrom`, `orderCastMindControl`) | the caster |
| `ReasonSpell` | a hallucination that gets a status effect (`executeSecondaryOrder`, `Hallucination2`) | 0 |
| `ReasonCancelled` | `cancelBuildingUnit` (building, egg, training unit), `cancelZergBuildingConstruction` (Extractor) | 0 |
| `ReasonExpired` | remove timer at 0 (`updateUnitValues`): hallucinations, Broodlings, spell units, orphan Interceptors; `orderScanner` | 0 |
| `ReasonDepleted` | `unitGatherResourcesFrom`: a mineral field mined out | 0 |
| `ReasonRemoved` | `UserActionFlags & 4` set before `killUnit`: the second Archon Templar, the Drone that becomes an Extractor, trigger remove, map units under a start location (`createStartingUnits`); and `placeHallucination` failure | 0 |
| `ReasonDetonated` | the unit kills itself: `initializeBullet` (Scarab, Infested Terran), `orderSpiderMine`, `orderSelfDestructing`, `explodeOnSignal`, `orderNukeLaunch` | 0 |
| `ReasonLinked` | killed because another unit died: units in training in a destroyed building, Interceptors and Scarabs inside a dead Carrier or Reaver, Interceptors with no parent (`orderInterceptorAttack`, `orderInterceptorReturn`), Nydus exit, nuke in a silo | 0 |
| `ReasonTrigger` | trigger kill actions | 0 |
| `ReasonOther` | any other `killUnit` call, for example `orderLarva` (Larva off creep), `startSieging`, `finishLanding` | 0 |

Edge cases:

- **Units in a transport that dies from damage** (`unitDealDamage`, not a building) fire before the transport, in the same half-step. Their `X, Y` is the transport position.
- **Units in a Bunker, or in a transport killed any other way** (trigger, `ReasonOther`), fire later, from `destroyLoadedUnits` when the container is removed in `orderDie`, so after the container's event. A Bunker first unloads them; only those it cannot unload die. The recorder keeps the container's kill context by unit ID to fill `Killer` for these.
- **Carrier or Reaver death**: fighters inside die with it (`ReasonLinked`); fighters outside get a remove timer and die later (`ReasonExpired`). Fighters still in its build queue fire no `ProductionCancelled`: `destroyGroundedBuilding` does not run for a unit.
- **Refinery, Extractor, Assimilator**: `UnitDestroyed` with `Reverted: true`, then later (in `orderDie`) `UnitMorphed` to Vespene Geyser and `UnitOwnerChanged` to 11. The ID stays valid. Unit-lifecycle consumers must not treat `Reverted` IDs as gone.
- **Archon merge** and **Drone to Extractor**: the removed unit fires `ReasonRemoved`. It is not a loss in combat.
- **Hallucinations** are classed `ClassHallucination`. They are in the stream and in `Run` output by default; consumers that count losses filter by `Class`.

### UnitOwnerChanged

```go
type UnitOwnerChanged struct {
	EventBase
	Unit     UnitID
	Type     UnitType
	From, To int8
	Reason   OwnerReason
}
```

Fires in `setUnitOwner` when the owner changes, not for turret subunits.

| `Reason` | Call site |
|---|---|
| `OwnerMindControl` | `triggerGiveUnitTo` from `orderCastMindControl`, with the units it carries: loaded units, Interceptors and Scarabs, the worker inside a refinery, addon, addon in construction, Nydus exit |
| `OwnerPlayerLeft` | `removePlayer` → `makeUnitNeutral`: every unit of a leaving or defeated player goes to 11 |
| `OwnerRefinery` | `buildRefinery` (11 to player), `revertRefineryToGeyser` (player to 11) |
| `OwnerInfest` | `orderInfestedCommandCenter` |
| `OwnerAddon` | `connectAddon`: an addon and its nuke join the building's owner; `buildingAbandonAddon` → `makeUnitNeutral`: the addon of a dead, lifted or infested building goes to 11 |
| `OwnerTrigger` | `triggerCreateUnit` gives a created addon to 11 |
| `OwnerOther` | any other `setUnitOwner` call |

### ProductionCancelled

```go
type ProductionCancelled struct {
	EventBase
	Owner    int8
	Producer UnitID // the unit whose queue held it; for a cancel of the incomplete unit itself, its ConnectedUnit or 0
	Unit     UnitID // the incomplete unit, 0 if it was not created yet
	What     ProductionKind // Unit, Building, Morph, Research, Upgrade, Addon
	Type     UnitType       // for Unit, Building, Morph, Addon
	Research Research       // for Research, Upgrade
	Why      CancelReason   // Command, ProducerDestroyed, OwnerChanged, NoSupply
}
```

| Call site | `What` | `Why` |
|---|---|---|
| `cancelBuildingUnit` (returns true) | Unit, Building, Morph (egg, cocoon, Lurker egg) | Command; ProducerDestroyed from `destroyGroundedBuilding`; OwnerChanged from `giveUnitTo` through `cancelBuildQueueIndex` |
| `cancelBuildQueueIndex` for a slot with no unit yet | Unit | Command, ProducerDestroyed, OwnerChanged (`giveUnitTo`) |
| `cancelBuildingMorph` (Lair, Hive, Greater Spire, Sunken, Spore) | Morph | Command or ProducerDestroyed |
| `cancelResearch`, `cancelUpgrade` | Research, Upgrade | `unitDestroyed` argument: ProducerDestroyed; from `giveUnitTo`: OwnerChanged; else Command |
| `orderZergUnitMorph` state 0 without supply (`cancelBuildQueue`) | Morph | NoSupply |
| cancel addon (`actionCancelAddon`) | Addon | Command |

### ResearchStarted and ResearchCompleted

```go
type Research struct {
	Upgrade bool  // false: tech
	ID      uint8 // TechType or UpgradeType
}

type ResearchStarted struct {
	EventBase
	Owner    int8
	Unit     UnitID // the building
	Research Research
	Level    int32 // 1 for tech
}

type ResearchCompleted struct {
	EventBase
	Owner    int8
	Unit     UnitID
	Research Research
	Level    int32
}
```

- Started: `actionResearch` and `actionUpgrade` after they charge the cost. They run inside a command, so `Cmd` is set.
- Completed: `orderResearchTech` when `TechResearched` goes from false to true; `orderUpgrade` when it writes `UpgradeLevels`. If the player already has it (`alreadyUpgraded`, `playerHasResearched`), nothing fires: the building only stops.
- Cancels are `ProductionCancelled`.

### PlayerLeft, PlayerDefeated, PlayerWon

```go
type PlayerLeft struct {
	EventBase
	Owner   int8
	Dropped bool // playerDropped (victory state 1) instead of playerLeft (2)
}

type PlayerDefeated struct {
	EventBase
	Owner int8
}

type PlayerWon struct {
	EventBase
	Owner int8
	State int32 // the victory state, 3 or more
}
```

- `PlayerLeft`: `playerLeft` and `playerDropped` (`frame.go`). The leave command calls `playerLeft`, so `Cmd` is set. The player's units then fire `UnitOwnerChanged` with `OwnerPlayerLeft`.
- `PlayerDefeated`: `onPlayerEliminated` in the trigger victory check.
- `PlayerWon`: `onVictoryState` with a state of 3 or more (`player_won` in the engine).
- A replay can end before the melee victory trigger runs (for example right after the loser's leave command), so `PlayerWon` can be missing. Use screp's `Computed.WinnerTeam` for the winner. `PlayerLeft` could reuse `OnVictoryState` (states 1 and 2 inside a command), but a dedicated hook keeps it apart from trigger defeat, which also sets state 2.

### Tick

```go
type Tick struct{ EventBase }
```

Emitted last in a processing half-step when `Frame % TickEvery == 0`. `Run` samples player stats and unit snapshots on ticks.

## 4. Unit IDs

```go
// UnitID is unique within a game: bits 0-10 are the unit index + 1 (1..1700),
// bits 11-31 are the slot's generation modulo 2^21. 0 means no unit.
type UnitID uint32

func (id UnitID) Index() uint16      // 1..1700
func (id UnitID) Generation() uint32
func (id UnitID) LegacyTag() uint16           // the 1.16 tag and OpenBW unit_id_t: index | (generation%32)<<11
func (id UnitID) MatchesSCRTag(tag uint32) bool // the downgrade maps tag to LegacyTag()
```

`MatchesSCRTag` applies the downgrade's own `legacyUnitTag` mapping, so it is exact by construction. A function that builds a 1.21 tag from a `UnitID` is not offered: index 1 can come from 1.21 index 1 or 1701, and SC:R's generation above 5 bits is unknown to the engine.

- It is `bwgame.UnitID32` (`getUnitID32`): `u.Index+1 | (u.UnitIDGeneration % 1<<21) << 11`. `UnitIDGeneration` is a `uint32` that `initializeUnit` increments on each reuse of the slot, so the ID does not wrap in a real game. OpenBW's 16-bit `unit_id_t` wraps after 32 reuses of one slot; it is not unique.
- screp's `UnitTag` is the raw tag in the command. For legacy replays it equals `LegacyTag()`. For 1.21 replays use `MatchesSCRTag`. To get the unit of a command, read `CommandResult.Selection`: the engine resolved it.
- `UnitID` is a plain number in JSON.

## 5. Enriched commands

### 5.1 Command index from the downgrade

```go
package downgrade

// Prepare is PrepareForOpenBW plus the report and the command index.
func Prepare(replay []byte) (*Prepared, error)

type Prepared struct {
	Loadable []byte
	Settings SCRSettings
	Report   Report
	Commands []SourceCommand // every command in the source stream, in order
}

type SourceCommand struct {
	Frame      uint32
	PlayerID   byte
	ID         byte   // the source command ID (0x63 stays 0x63)
	Position   int    // position among the commands of its frame, over all blocks of that frame
	Action     int32  // index in Loadable's action stream; -1 if dropped
	DropReason string // a Reason* constant if dropped
}
```

- `translateBlock` appends one `SourceCommand` for each command it splits, with `Action` = the count of commands written so far, or -1 for a drop. For legacy input, `Prepare` walks the actions the same way and copies them unchanged.
- Commands after a `ReasonUnknownLength` or `ReasonTruncated` drop are not split, so they are not in `Commands`. `PrepareForOpenBW` becomes a wrapper of `Prepare`.

Mapping to screp's `Commands.Cmds`: group both by frame and walk the two lists of a frame in step. Position `p` links to the `p`-th screp command of that frame while the IDs match. screp's Land command is the Build command (0x0c) with its virtual ID `0xfe` (`VirtualTypeIDLand`), so 0x0c matches 0xfe. Linking stops for the rest of the frame at the first ID mismatch, or at the first `ParseErrCmd` of screp or length drop of the downgrade in that frame. The rest of the frame has `ScrepIndex = -1`. The block size is one byte, so a busy frame can span several blocks; a parse error ends only its own block in both parsers, so positions after it in the same frame are not reliable. The result is exact or unlinked, never a guess.

Engine side: `Action` `n` is the `n`-th `readAction` call (counted where `OnAction` is called today). `executeActionGroup` reads the actions in the order the downgrade wrote them. An unknown player ID throws, so no action is skipped without an error.

### 5.2 CommandResult

```go
type CommandResult struct {
	EventBase               // Cmd == Action
	Action      int32       // index in the loadable action stream
	ScrepIndex  int32       // index in screp's Commands.Cmds, -1 if not linked
	PlayerID    byte        // screp player ID
	Owner       int8        // engine slot
	ActionID    byte        // legacy action ID executed by the engine
	Outcome     Outcome
	Rejections  []Rejection // why the engine refused; empty if accepted
	Selection   []UnitID    // see below
	MissingTags uint8       // Select/Shift Select: tags that resolve to no unit
	ForeignTags uint8       // Select/Shift Select with 2+ tags: tags of another player's units
}

type Outcome uint8

const (
	Accepted Outcome = iota // the action function returned true and had an effect
	Rejected                // the action function returned false
	NoEffect                // returned true but did nothing (Build without placement or resources)
)

type Rejection struct {
	Unit   UnitID // the selected unit refused; 0 for the command as a whole
	Reason RejectReason
}

type RejectReason uint8 // NoUnit, NotOwner, Requirements, Invalid, QueueFull, Busy, Supply, Minerals, Gas, Placement
```

`Outcome`:

- `Accepted`/`Rejected` is the `bool` that `readAction` returns, through `AfterAction`. For order commands (Right Click, Stop, Attack...) it means at least one selected unit took the order. It says nothing about the result of the order.
- `actionBuild` returns true also when `canPlaceBuilding` fails or `placeBuilding` returns early (supply, resources, or a Drone already in `DroneBuild` or landing). The game client checks most of these before it sends the command, so in a replay this signals a desync. Each of these early exits calls `OnActionRejected`. `NoEffect` is an accepted Build with at least one `Rejection`. A rule based on "no build queue push" would be wrong: the Nydus exit (`BuildNydusExit`), landing and `CTFCOP2` set an order and push nothing.

`Rejections`: `OnActionRejected` is called at the refusing `return` of these functions, with the reason of the check that failed. The checks are the engine's own, in the engine's order. Nothing is re-implemented.

| Command | Functions |
|---|---|
| Train, Train Fighter | `actionTrain`, `actionTrainFighter`, `buildQueuePush` |
| Unit Morph, Building Morph | `actionMorph`, `actionMorphBuilding` (one `Rejection` for each refused selected unit) |
| Build | `actionBuild`, `canPlaceBuilding` result in `actionBuild`, every early `return` in `placeBuilding` before `setUnitOrderPos` |
| Research, Upgrade | `actionResearch`, `actionUpgrade` |

`Minerals` or `Gas`: `hasAvailableResourcesFor` compares both; the hook reads the same two values and reports the first one that is short. Other commands have no reasons, only `Outcome`.

`Selection`:

- For Select (0x09), Shift Select (0x0a) and Deselect (0x0b): the selection after the command.
- For any other command: `ActionState.Selection[owner]` when `BeforeAction` runs, which is the set of units the command acts on.
- `MissingTags` and `ForeignTags` are the tag test of `tools/desync/gotags`, computed in `BeforeAction` with the same rules: tag index 0 is ignored; a missing unit counts; a foreign unit counts only in a selection of 2 or more.

### 5.3 What links exactly to a command

`EventBase.Cmd` is set by two mechanisms. Both are exact by construction. Nothing else is linked.

1. **Synchronous**: every event emitted between `BeforeAction` and `AfterAction` of action `n` gets `Cmd = n`. Examples: `ResearchStarted`, `ProductionCancelled` from a cancel command, `PlayerLeft`, `UnitOwnerChanged` of the leaving player's units.
2. **Shadow build queue**: production goes through `Unit.BuildQueue` and `BuildQueueLimbo` (and `Building.AddonBuildType` for addons). The game mirrors each queue with a slice of `Action` values, keyed by unit index. `OnBuildQueue(u, op, index)` is called at each mutation site: `buildQueuePush` (and its limbo merge), `buildQueueClearSingle`, `BuildQueue.Erase` in `secondaryOrderTrain` (two sites), `secondaryOrderTrainFighter` and `cancelBuildQueueIndex`, and the reset in `initializeUnit`. This is every write to these fields in `bwgame` today. `OnAddonBuildType(u, set)` is called where `AddonBuildType` is set (`placeBuilding`) or cleared (`secondaryOrderBuildAddon`, three sites, and `initializeUnitType`). A push inside command `n` stores `n`; any other push stores -1. The mirror applies the same operations as the engine, so the front of the mirror is the front of the queue.
   - `UnitCreated` (Trained, Built by SCV or Probe), `MorphStarted` and `UnitCreated` of an addon take `Cmd` from the producer's mirror front at that moment.
   - The second unit of an egg takes the `Cmd` of the egg's `MorphStarted`: the recorder keeps it by unit ID. The mirror front is gone by then (`orderZergUnitMorph` state 2 calls `buildQueueClearSingle` before `orderZergBirth`).
   - `placeBuilding` pushes the building type to the worker's queue, and the SCV or Probe creates the building from it, so Terran and Protoss buildings link to their Build command. Drone buildings link through `orderDroneBuild`.
   - The Nydus exit is not linked: the Build command only sets the Canal's order.

`Run` inverts the links: `CommandResult.Produced` lists the `UnitID`s of events with that `Cmd`.

**Not linked** (`Cmd = -1`), because the engine has no exact path from the command:

- Effects of orders after the command returns: kills, Mind Control, spells (Broodlings, hallucinations, mines, Scanner Sweep, Dark Swarm), Archon merge, siege, lift-off. A queued (shift) order can run many seconds later, after other commands. `Parent` and `Killer` give the unit; the command is not guessed.
- Units that a Hatchery spawns (Larva), map units, trigger units.
- Production completion (`UnitCompleted`): it links to the unit, and the unit's `UnitCreated` or `MorphStarted` links to the command.

## 6. L2 recorder: Run and Result

```go
// Run plays the whole replay (or up to MaxFrame) with the stream and records a Result.
// On an engine error it returns the partial Result and the error.
func Run(replay []byte, opts Options) (*Result, error)

type Options struct {
	StreamOptions                // Data, Kinds, MaxFrame, Rules, CustomRules, Hooks; Run sets TickEvery
	                             // Classes 0 here means all but ClassFighter and ClassSpell
	Commands       bool          // include screp's Commands, like screp -cmds
	MapData        bool          // include screp's MapData, like screp -map
	CommandResults ResultsMode   // default ResultsActions
	StatsEvery     int32         // default 24; -1 disables stats
	UnitSnapshots  *UnitSnapshotOptions // nil: off
}

const DefaultKinds = AllKinds &^ (KindCommandResult | KindTick)

type ResultsMode uint8

const (
	ResultsActions ResultsMode = iota // all but Select/Shift Select/Deselect, plus selects with MissingTags+ForeignTags > 0
	ResultsAll
	ResultsNone
)

type UnitSnapshotOptions struct {
	Every          int32     // frames, for example 240
	Fields         UnitField // ID, Type and Owner are always present
	IncludeNeutral bool      // owner 11; off by default (about 130 resources per map)
}

type UnitField uint16

const (
	Position UnitField = 1 << iota // X, Y
	HP
	Shields
	Energy
	Order       // order ID
	State       // Completed, Burrowed, Cloaked, Loaded, Dying bits
	BuildTime   // remaining build time
)
```

`Run` uses only `Open`, `Events`, `Stats`, `Units` and `Info`. It always subscribes `KindCommandResult` (for trust signals), `KindUnitCreated` and `KindMorphStarted` (for `Produced`), and ticks (at the GCD of `StatsEvery` and `UnitSnapshots.Every`). It drops the kinds that the caller did not ask for, and always `CommandResult` and `Tick`, from `Events` in the output. Unit snapshots use the same `Classes` filter as the events.

### Result

```go
type Result struct {
	SchemaVersion  int // 1
	Header         *rep.Header   // screp
	Commands       *rep.Commands `json:",omitempty"` // screp, opt-in
	MapData        *rep.MapData  `json:",omitempty"` // screp, opt-in
	Computed       *rep.Computed // screp
	Engine         EngineInfo
	Trust          Trust
	Events         []Event
	CommandResults []CommandResultRow `json:",omitempty"`
	Stats          []PlayerStats      `json:",omitempty"`
	UnitSnapshots  []UnitSnapshot     `json:",omitempty"`
}

type EngineInfo struct {
	Version         string // module version from debug.ReadBuildInfo, or "devel"
	ScrepVersion    string // the screp module version: Header, Computed and Commands are screp's JSON
	OpenBWCommit    string // "4b046d5", the reference the port matches
	DataHash        string // Data.FilesHash()
	SourceFormat    string // "legacy", "1.18-1.20", "1.21+"
	Rules           RulesInfo // mode, object limits, DisabledThingyUnits, OrderAllocationRule, FlyingSCVFix
	FramesSimulated int32
	EndFrame        int32  // from the replay
	ReachedEnd      bool   // FramesSimulated == EndFrame
	Error           string `json:",omitempty"`
}

type Trust struct {
	DroppedCommands  int            // downgrade drops
	DroppedByReason  map[string]int `json:",omitempty"`
	UnlinkedCommands int            // screp commands with no SourceCommand or Action
	NotExecuted      int            // actions in Loadable that the engine never read (stop frame, replay end, or out-of-order frames)
	Desync           DesyncSignal
	Production       []ProductionTrust // one per playing slot
}

type DesyncSignal struct {
	InconsistentSelects int   // Select/Shift Select results with MissingTags+ForeignTags > 0
	FirstFrame          int32 // frame of the first one, 0 if none
	Likely              bool  // InconsistentSelects >= 20
}

type ProductionTrust struct {
	Owner    int8
	PlayerID byte
	Issued   int // Train, Train Fighter, Unit Morph, Building Morph, Build commands
	Rejected int
	NoEffect int
}

type CommandResultRow struct {
	CommandResult
	Produced []UnitID `json:",omitempty"`
}
```

- The desync threshold (20) is the tag test's: clean games have 0-10 inconsistent commands (DESYNC.md). `Likely` is a signal, not a proof.
- `(Rejected + NoEffect) / Issued` is not a desync signal by itself. On clean SC:R games it is often 20-60%: players spam Train when they have no minerals, and the engine refuses each one (`Minerals`). Compare it between players and games, not against a fixed threshold.
- The output is deterministic: no times, no map iteration order in arrays. Two runs give the same bytes.

### Player stats

```go
type PlayerStats struct {
	Frame            int32
	Owner            int8
	Minerals, Gas    int32   // CurrentMinerals, CurrentGas
	GatheredMinerals int32   // TotalMineralsGathered (includes the starting minerals)
	GatheredGas      int32   // TotalGasGathered
	SupplyUsed       float64 // the player's race: SupplyUsed[slot][race] / 2 (x.5 possible)
	SupplyMax        float64 // min(SupplyAvailable / 2, 200)
	Workers          int32   // completed SCV + Drone + Probe
	Army             int32   // completed army units, see below
	ArmyMinerals     int32   // army value, see below
	ArmyGas          int32
	Buildings        int32   // CompletedBuildingCounts
	UnitScore        int32   // engine UnitScore
	BuildingScore    int32   // engine BuildingScore
}
```

- One row per playing slot (controller occupied, computer or left) on each tick, from `g.Player(slot)`. A player who left keeps rows; the values go to zero when the units become neutral.
- **Army** = sum of `CompletedUnitCounts[slot][t]` over the types `t` with `GroupFlagsMen`, except workers, Larva, Egg, Cocoon, Lurker Egg, Interceptor, Scarab, Spider Mine and Nuclear Missile. `CompletedUnitCounts` already leaves out hallucinations (`addCompletedUnit`).
- **Army value** = the same sum, each count times the units.dat cost, divided by 2 for Zergling and Scourge (one egg makes two). Morphed units add their source cost: Lurker + Hydralisk, Guardian and Devourer + Mutalisk, Archon = 2 × High Templar, Dark Archon = 2 × Dark Templar.
- Income is not a field. It is the difference of `GatheredMinerals` between two rows.

### Unit snapshots

Columnar, one object per snapshot frame. Only the selected fields are present.

```go
type UnitSnapshot struct {
	Frame     int32
	ID        []UnitID
	Type      []uint16
	Owner     []int8
	X, Y      []int32 `json:",omitempty"`
	HP        []int32 `json:",omitempty"` // integer part
	Shields   []int32 `json:",omitempty"`
	Energy    []int32 `json:",omitempty"`
	Order     []uint16 `json:",omitempty"` // []uint8 would be base64 in encoding/json
	State     []uint16 `json:",omitempty"` // StateCompleted, StateBurrowed, StateCloaked, StateLoaded, StateDying, StateHallucination
	BuildTime []int32 `json:",omitempty"`
}
```

### JSON shape

- Field names are Go names (PascalCase), as screp writes them. They are stable within a `SchemaVersion`. New fields and new enum values do not change the version; readers must ignore unknown fields and accept unknown enum strings. A rename, removal or change of meaning increments it. screp's parts follow screp's version, which `Engine.ScrepVersion` records.
- The JSON round-trips: `Result` has an `UnmarshalJSON` that decodes `Events` with `UnmarshalEvent`, so L3 code can read cached results with the same types.
- Until v1, `bwgame` (L0) has no compatibility promise, although it is importable. The `openbw` package follows semver.
- Enum values are strings: `"Kind": "UnitDestroyed"`, `"Reason": "Damage"`, `"Outcome": "Rejected"`.
- `Type` in events is screp's object `{"Name": "Zerg Zergling", "ID": 37}` (`repcmd.UnitByID`). `Research` is `{"Tech": {...}}` or `{"Upgrade": {...}}`, with screp's `TechByID`/`UpgradeByID`. Snapshots use plain IDs for size.
- Each event is one object with `"Kind"` and its own fields. `Cmd` and zero `UnitID`s are omitted when -1 or 0.

## 7. CLI: cmd/openbw

```
openbw [FLAGS] repfile.rep
```

| Flag | Default | Meaning |
|---|---|---|
| `-data dir` | `$OPENBW_DATA` | game data directory |
| `-header` | true | screp header |
| `-map` | false | screp map data |
| `-cmds` | false | screp commands |
| `-computed` | true | screp computed data |
| `-events` | true | engine events |
| `-kinds list` | all event kinds | comma list, for example `UnitCreated,UnitDestroyed` |
| `-fighters`, `-spells` | false | include Interceptor/Scarab and spell units in events |
| `-results mode` | `actions` | `actions`, `all` or `none` |
| `-stats` | true | player stats |
| `-statsevery N` | 24 | frames |
| `-units` | false | unit snapshots |
| `-unitsevery N` | 240 | frames |
| `-unitfields list` | `pos,hp` | `pos,hp,shields,energy,order,state,buildtime` |
| `-neutral` | false | neutral units in snapshots |
| `-maxframe N` | 0 | stop after frame N |
| `-rules mode` | `auto` | `auto` or `openbw` |
| `-overview` | false | human-readable summary, no JSON |
| `-stdin`, `-outfile`, `-indent`, `-version` | as screp | |

Exit codes: screp's 1-3, plus 5 for an engine error. With exit code 5 the JSON is still written, with `Engine.Error` set and `ReachedEnd: false`.

`openbw rep.rep` (indented by default as in screp; values are illustrative, arrays shortened):

```json
{
  "SchemaVersion": 1,
  "Header": { "Engine": {"Name": "Brood War", "ID": 1, "ShortName": "BW"}, "Version": "1.21+", "Frames": 21140, "Players": [ ... ] },
  "Computed": { "WinnerTeam": 1, "PlayerDescs": [ ... ] },
  "Engine": {
    "Version": "v0.1.0", "OpenBWCommit": "4b046d5", "DataHash": "9f2c…", "SourceFormat": "1.21+",
    "Rules": {"Mode": "Auto", "Limits": {"Images": 10000, "Sprites": 5000, "LoneSprites": 1000, "Bullets": 400, "Orders": 4000},
              "DisabledThingyUnits": true, "OrderAllocationRule": true, "FlyingSCVFix": true},
    "FramesSimulated": 21140, "EndFrame": 21140, "ReachedEnd": true
  },
  "Trust": {
    "DroppedCommands": 0, "UnlinkedCommands": 0, "NotExecuted": 0,
    "Desync": {"InconsistentSelects": 3, "FirstFrame": 15022, "Likely": false},
    "Production": [{"Owner": 0, "PlayerID": 0, "Issued": 412, "Rejected": 6, "NoEffect": 0},
                   {"Owner": 1, "PlayerID": 1, "Issued": 268, "Rejected": 2, "NoEffect": 0}]
  },
  "Events": [
    {"Kind": "UnitCreated", "Frame": 0, "Unit": 1, "Type": {"Name": "Resource Mineral Field", "ID": 176}, "Owner": 11, "X": 2272, "Y": 176, "Source": "Map", "Class": "Normal"},
    {"Kind": "MorphStarted", "Frame": 13, "Cmd": 2, "Unit": 1693, "Morpher": 1693, "From": {"Name": "Zerg Larva", "ID": 35}, "Into": {"Name": "Zerg Drone", "ID": 41}, "Owner": 0},
    {"Kind": "UnitDestroyed", "Frame": 9120, "Unit": 6301, "Type": {"Name": "Zerg Zergling", "ID": 37}, "Owner": 0, "X": 1510, "Y": 1822,
     "Reason": "Damage", "Killer": 1489, "KillerType": {"Name": "Terran Marine", "ID": 0}, "KillerOwner": 1}
  ],
  "CommandResults": [
    {"Frame": 13, "Cmd": 2, "Action": 2, "ScrepIndex": 2, "PlayerID": 0, "Owner": 0, "ActionID": 35, "Outcome": "Accepted", "Selection": [1693], "Produced": [1693]},
    {"Frame": 4410, "Cmd": 812, "Action": 812, "ScrepIndex": 812, "PlayerID": 1, "Owner": 1, "ActionID": 31, "Outcome": "Rejected",
     "Rejections": [{"Reason": "Supply"}], "Selection": [1702]}
  ],
  "Stats": [
    {"Frame": 24, "Owner": 0, "Minerals": 58, "Gas": 0, "GatheredMinerals": 58, "GatheredGas": 0, "SupplyUsed": 4.5, "SupplyMax": 9,
     "Workers": 4, "Army": 0, "ArmyMinerals": 0, "ArmyGas": 0, "Buildings": 1, "UnitScore": 0, "BuildingScore": 0}
  ]
}
```

`openbw -overview rep.rep`:

```
Engine  : BW 1.21+   openbw v0.1.0 (OpenBW 4b046d5), rules auto (SC:R)
Map     : Neo Sylphid 3.2
Length  : 14:48 (21140 frames), simulated to the end
Players : lillljilililili (Z)  vs  564376 (T)
Trust   : no dropped commands, 3 inconsistent selections (clean), rejected production 1.5% / 0.7%
Final   : Z  supply 142/200  workers 58  army value 4150/2900
          T  supply  96/134  workers 41  army value 2900/1050
```

The L3 examples are Go examples in `example_test.go`: `Example_buildOrder` (the stream with `MaxFrame`: `MorphStarted`, `UnitCreated` with `Source` Built or Trained, and `ResearchStarted` in the first 8 minutes), `Example_killsPerPlayer` and `ExampleRun`. `TestExamples` runs them against real data and checks their output; it skips when the data or the replay is missing.

## 8. Cost

Counts measured on the 20 replays in `tools/difftest/replays.txt` with a polling prototype (a scan of all units after each frame). Polling cannot see units that live less than one frame and has no `MorphStarted`, research or cancel events, so the real counts are a little higher.

| Replay (1v1) | Length | Commands | Created (neutral) | Completed | Morphed | Destroyed | Owner changes |
|---|---|---|---|---|---|---|---|
| MM-CB157510 | 14.8 min | 10,677 | 532 (130) | 745 | 407 | 217 | 121 |
| MM-AAA58B20 | 14.3 min | 8,182 | 645 (136) | 625 | 95 | 214 | 121 |
| MM-4862526E | 16.0 min | 11,130 | 716 (130) | 1,013 | 609 | 338 | 106 |
| MM-657AB750 | 55.9 min | 24,734 | 2,555 (153) | 3,829 | 2,550 | 2,219 | 123 |

A 15-minute 1v1, default `Run`:

| Part | Rows | Bytes per row (compact JSON) | Size |
|---|---|---|---|
| Header, computed (screp) | 1 | | 2.5 KB |
| Events | 2,000-3,000 | ~150 | 0.3-0.45 MB |
| Command results (`actions`, about 60% of commands) | ~6,500 | ~80 | ~0.5 MB |
| Player stats every 24 frames | 2 × ~880 | ~230 | ~0.4 MB |
| Unit snapshots (opt-in, every 240 frames, pos+hp, no neutral) | ~90 × ~100 units | ~20 per unit | ~0.2 MB |
| **Total default** | | | **~1.3 MB** (gzip about 10×) |

For comparison, `screp -cmds -indent=false` on MM-CB157510 is 1.34 MB; `screp` without commands is 2.4 KB.

Overhead targets, measured with `tools/bench` against the commit before the hooks, same replays, best of 5:

| Mode | Target |
|---|---|
| Hooks compiled in, nothing subscribed (`obwhash`) | < 2% |
| `Open` with `AllKinds`, empty loop body | < 5% |
| `Run` with defaults, JSON encoding excluded | < 10% |

Engine time today is about 1-1.5 s for a 15-minute game (1.2 s for MM-CB157510 in the prototype, including the polling).

- Allocation: one per emitted event (a few thousand per game) plus one `Selection` slice per `CommandResult`. `Units()` yields values and allocates nothing.
- Concurrency: one `Data` for any number of games; one goroutine per `Game`. Games share only read-only data, so throughput should scale with cores.
- Early stop: `break` or `MaxFrame` stops the engine at once; the cost is the frames simulated so far.

## 9. Verification plan

- `tools/difftest` on `tools/difftest/replays.txt`: 20/20 with `-out $TMPDIR/api-difftest`, and 20/20 with `-harness` from `tools/scrharness/build.sh` and `-obwhash-args -scr-rules`.
- `obwhash` output does not change: it installs no new hooks.
- Event invariants on all 20 replays: each `UnitID` has at most one `UnitCreated` and at most one `UnitDestroyed` without `Reverted`; every `UnitDestroyed` has a `UnitCreated` (map units included); no unit event before the unit's `UnitCreated`; `UnitCompleted` never before `UnitCreated`; `Seq` increases, `Frame` never decreases.
- Link invariants: every `ScrepIndex` points to a screp command with the same frame, player ID and (mapped) type ID.
- `gotags` and `Trust.Desync` give the same counts on the replays in DESYNC.md.
- The shadow queue equals `Unit.BuildQueue` in length and types after every half-step (debug build check).
- `-race` test with 4 parallel games on one `Data`.

## Decisions and alternatives rejected

| Decision | Rejected alternatives and why |
|---|---|
| Streaming base: `Events() iter.Seq[Event]` + `Err()`, like `bufio.Scanner`. `Run` is built on it. | **Channels**: a goroutine per game, leaks on early stop, slower. **Callbacks only** (the raw `bwgame.Hooks`): no natural early stop, state is mid-mutation when they run. **`iter.Seq2[Event, error]`**: the error is in every element although there is at most one, at the end. **`Next()`/`Event()` methods**: `iter.Pull` gives that from the `iter.Seq`. |
| Yield at half-step boundaries (after commands, after processing). | **Yield inside hooks**: the caller would read state in the middle of an engine function, and `break` would need to unwind the engine. **One yield per frame**: command events would see state after the next processing step. |
| Events are pointer structs behind a small interface; a type switch selects them. | **One struct with a Kind and all fields**: unclear which fields are valid, larger JSON. |
| Filter at subscription: hooks for unsubscribed kinds stay nil. | **Emit everything and filter after**: allocations for events nobody reads. |
| `UnitCreated` at production start (where the engine creates the unit), plus `MorphStarted` for Zerg. | **Creation at completion**: the engine creates the unit at start; a different meaning would hide cancels. **A generic ProductionStarted for every race**: duplicates `UnitCreated` for Terran and Protoss. |
| Report fighters, mines, hallucinations and spell units, with a `Class`. `Run` leaves out fighters and spell units by default. | **Leave them out of the engine view**: kills by Scarabs and Interceptors could not name the killer. `KillerParent` gives the Reaver or Carrier. |
| `UnitDestroyed` at `killUnit` with a reason set at the call sites. | **At removal (`onUnitDestroy`)**: the killer is gone by then. **Killer from `LastAttackingPlayer`**: it is a player, set by any damage, not the kill. The engine does not count kills (`KillCount` only changes at Archon merge). |
| Refinery death is `UnitDestroyed` with `Reverted: true`; the ID lives on. | **A new geyser ID**: the engine keeps the same `unit_t`; later commands and events use the old ID. |
| Unique `UnitID` from the 21-bit generation. | **OpenBW's 16-bit `unit_id_t`**: wraps after 32 reuses. **A creation counter**: cannot map to replay tags. |
| Rejection reasons from hooks at the engine's own `return false` sites. | **Re-check costs and supply in the recorder**: a second copy of engine logic that can drift. **`LastError`**: set only by unit creation, not by actions. |
| Production links by mirroring `BuildQueue` operations. | **Match by player, type and time**: guesses, wrong with several producers. **An origin field in `bwgame.Unit`**: changes engine state that the hash and state copies see. |
| `Run` default: events, action results, stats every 24 frames. Snapshots, fighters, spell units, select results and screp commands are opt-in. | **Snapshots on by default**: the largest output, rarely needed. **All command results**: Select commands are about 40% and add little; the inconsistent ones stay. |
| Map units are frame 0 events. | **Leave them out**: minerals and geysers die or become refineries; their events need a `UnitCreated`. |
| Owner is the engine slot (screp `SlotID`); command rows also carry screp's `PlayerID`. | **Only `PlayerID`**: neutral units and computer slots have no unique player ID. |
| PascalCase JSON with screp's type objects; columnar snapshots. | **snake_case**: unlike screp, so mixed output. **Row snapshots**: field names repeat for every unit. |
| One `Data` per process, shared. | **One `GlobalState` per game**: memory per game for no gain; the race test shows sharing is safe. |
| `RulesAuto`: SC:R rules for SC:R replays. | **OpenBW rules by default**: Pole Star replays fail to load and bullet-limit games desync (DESYNC.md). |
| No state checkpoints, no L3 analyses (one example). | Not needed for the first users. A checkpoint is a large opt-in output and can come later with `state_copier`. |

## Review changes

An adversarial review checked the spec against `bwgame/` and screp. These are the changes and why.

**Event semantics (checked in the engine code)**

| Change | Why |
|---|---|
| `UnitCreated` fires right after `createUnit` returns, not "after the caller finishes the unit" | `spawnLarva`, `createInitialUnit`, Broodlings, mines, Scanner, triggers and the returned Drone call `completeUnit` at once, so `UnitCompleted` came first. The class comes from the source and the type. |
| `killUnit` reads `OrdersDie` at its start | `killUnit` has no guard and can run twice on one unit; the old text read the order after `setUnitOrderPos` had set it. |
| `placeHallucination` failure emits `UnitDestroyed` (`ReasonRemoved`) | It calls `destroyUnit` without `killUnit`, so the unit had a `UnitCreated` and no end. |
| Infest Command Center needs its own `OnUnitMorphed` call | `orderInfestedCommandCenter` sets `u.UnitType` directly, not through `morphUnit`. |
| Transport deaths split by cause | Only `unitDealDamage` kills passengers first. Bunkers and non-damage deaths kill them later in `destroyLoadedUnits`. |
| One `UnitDestroyed` per ID, except geysers | A geyser can carry several refineries over a game; each death is `Reverted` on the same ID. |
| Reason and owner tables completed | `createStartingUnits` sets flag 4 (Removed, not Other); Mind Control on a hallucination; parentless Interceptors; `buildingAbandonAddon`; Mind Control also moves fighters and the refinery worker. |
| Second egg unit links through the egg's `MorphStarted` | The egg's queue front is already cleared when `orderZergBirth` spawns it. |
| `AddonBuildType` gets its own hook; the Nydus exit is not linked | The mirror listed no addon sites; the Nydus exit Build only sets an order. |
| Hallucinations stay in `Run` output | Two places in the spec disagreed; the CLI has no flag for them. |
| `PlayerWon` can be missing | Replays can end before the victory trigger runs. |

**Enriched commands**

| Change | Why |
|---|---|
| Land matches by screp's virtual ID `0xfe` | screp turns Build with `BuildingLand` into `TypeLand`; the ID check unlinked every Land. |
| `Position` counts over all blocks of a frame; linking stops at the first mismatch or parse error in a frame | Block size is one byte, so a frame can span several blocks; "frames are unique per block" was false. |
| `NoEffect` = accepted Build with a `Rejection` | The old "no queue push" rule called every Nydus exit, landing and `CTFCOP2` Build `NoEffect`. |
| `SCRTag()` replaced by `MatchesSCRTag` | The downgrade maps 1.21 index 1 and 1701 to the same index and drops generation bits, so a 1.21 tag cannot be rebuilt. The check is exact. |
| `NotExecuted` counts unread actions | The old "at or after the stop frame" misses actions stalled by out-of-order frames. |
| `Cmd` is set only with `KindCommandResult` | The spec did not say which hooks fill `Cmd` when only unit events are subscribed. |

**API and Go idiom**

| Change | Why |
|---|---|
| `ExecuteFrameActions` and `ProcessFrame` in `bwgame` | `Game` cannot call the unexported halves; `recoverEngineError` re-panics runtime errors, so `Game` recovers itself. |
| `AfterAction` selects the reporting action path | `nextFrame` only takes it for `OnActionFail` or `BeforeAction`. |
| `Unit`, `Player`, `UnitType`, `ClassSet`, `UnitTypeInfo`, `Stats` defined | They were used but not defined. Army value needs this game's costs, and stream users need the same stats as `Run`. |
| `Options` embeds `StreamOptions`; `Kinds` 0 means `DefaultKinds` in both | Duplicate fields and a zero value that emitted nothing. |
| `StreamOptions.Hooks` and a raw `OnUnitDamaged` | DESIGN.md promises raw L1 hooks for what events leave out; the spec gave no way to add them next to the stream's own hooks. |
| Sealed `Event`, `UnmarshalEvent`, `Result.UnmarshalJSON` | L3 reads cached results (DESIGN.md), which needs decoding of the event list. |
| `EngineError` implements `error`; `Game` is not concurrent-safe; events are never reused; no `Close` | Missing contracts. |
| `ScrepVersion`; unknown fields and enum values are tolerated; `bwgame` has no promise before v1 | Output includes screp's JSON, which changes with screp, not with `SchemaVersion`. |
| `Owner` = index in screp's `Header.Slots` | The engine reads slot `i` into player `i`; `SlotID` is a header field that usually, not by definition, equals it. |

**Foundation check: what third parties build with no library change**

| Consumer | How | Gap |
|---|---|---|
| Build order | Stream `MorphStarted`, `UnitCreated` with `SourceBuilt`/`SourceTrained`, `ResearchStarted`, `ProductionCancelled`, `MaxFrame` | None |
| Army value curve | `Run` stats, or `Tick` + `g.Stats` in the stream | None (was: no unit costs, no shared formula) |
| Fight detection | `UnitDestroyed` with position, killer and class, clustered in time and space; HP from `Units()` on ticks | Fights with no deaths need `OnUnitDamaged` through `StreamOptions.Hooks` |
| Map control | `Units()` positions on ticks, or unit snapshots | Vision and creep only through `Engine()` (L0) |
| Replay viewer | `KindTick`, `TickEvery: 1`, `Units()`, `Engine()` for sprites and bullets | No seek: replay from frame 0 (about 1 s per game) until checkpoints exist |
| Classifier at 8 minutes | `MaxFrame: 8*60*24` or `break` | None |
| screpdb ingestion | `Run` JSON; rows keyed by replay and event index (`Seq` is not in the JSON); `ScrepIndex` joins screp commands | None |

## Implementation notes

Changes from the first draft of this spec, found while implementing it:

| Change | Why |
|---|---|
| `Class` on every unit event | The text said "every unit event has a `Class`", but only `UnitCreated` had the field; loss counters filter `UnitDestroyed` by class. |
| `MorphStarted` of a Drone building takes the command from the front that `reinitializeUnitType` cleared | `morphUnit` of a Drone into a grounded building calls `buildQueueClearSingle`, so the queue is empty when `MorphStarted` fires. The shadow queue keeps the last cleared front of each unit. Only `MorphStarted` reads it; the Nydus exit stays unlinked. |
| `Run` also subscribes `UnitCreated` and `MorphStarted` | `Produced` is built from them. They are dropped from the output when not asked for. |
| `UnitSnapshot.Order` and `State` are `[]uint16`; `State` has a `StateHallucination` bit | `encoding/json` writes `[]uint8` as base64. |
| `TickEvery > 0` alone enables ticks | Ticks are opt-in through `TickEvery`; `KindTick` in `Kinds` is not needed. |
| `EventBase.Frame` and `Cmd` are written by each event's `MarshalJSON`; `KillerType` is omitted without a killer, `KillerOwner` when -1; `ProductionCancelled` has either `Type` or `Research` | Smaller JSON, and no ambiguous zero values (unit type 0 is the Marine, tech 0 is Stim Packs). Decoding restores -1. |
| `Result.Events` is `omitempty` | `cmd/openbw -events=false` omits the field. |
| `RulesInfo.Mode` is a string: `Auto`, `OpenBW` or `Custom` | `CustomRules` has no `RulesMode` value. |
| `Data.FilesHash` hashes each file's path, size and content | Two data sets with the same bytes in other files must not collide. |
| `Units()` leaves out turret subunits | Events never report them. |
| `Player.Left` is also true after a defeat | `removePlayer` sets the controller to "user left" in both cases. |
| Production rejection rate is not a desync signal | See Trust above. |
| Extra exported helpers: `Game.PlayingSlots`, `Game.Prepared`, `Version`, `ParseKinds`, `ParseResultsMode`, `ParseUnitFields`, `Classes`, `ClassSet.Has`, `OpenBWCommit`, `SchemaVersion`, `DefaultStatsEvery`, `UnknownEnum`, `downgrade.LegacyUnitTag` | The CLI, `Run` and `UnitID.MatchesSCRTag` need them, and stream users need the same playing slots as `Run`. |
| The L3 example is Go examples, not `examples/buildorder` | Examples show in godoc and are compile-checked; `TestExamples` runs them. |
| `cmd/openbw` also has screp's `-maptiles` and `-mapres`; exit code 2 also covers data and map load errors | Same output as screp for `-map`. |

Verification (all pass): `go test ./...` with stream invariants on 6 replays (4 SC:R 1v1, an 8-player team game, a legacy replay): unit lifecycle order, `Seq` and `Frame` order, every `ScrepIndex` checked against screp, shadow queue length equal to the engine's on every tick, the final state hash equal to a run without hooks, `Trust.Desync` equal to the tag test, stream resume after `break`, deterministic and round-tripping JSON, and 4 parallel games on one `Data` under `-race`. `tools/difftest` stays 20/20 in default mode and with `-scr-rules` against the patched harness. Overhead of `Open` with `AllKinds` and of `Run` with defaults is below the noise of the benchmark (about 0.9 s for MM-CB157510 in all three modes).
