# Public API

Package `openbw`, import `github.com/marianogappa/openbw-go`. It is a library: it runs the engine on a replay and gives raw engine facts as Go values through a stream. It writes no files, has no output format and does no analysis. Recorders, statistics, build orders and desync heuristics belong to the tools that use it.

The API speaks engine terms only: player slots, engine action indexes, source command indexes and `UnitID`s. The library packages (`openbw`, `bwgame`, `downgrade`) do not import a replay parser. [screpsim](https://github.com/marianogappa/screpsim) joins these facts to [screp](https://github.com/icza/screp)'s commands, players and unit tags.

```
stream     openbw.Open(replay, Options) → *Game; for e := range g.Events(); g.Unit, g.Units, g.Player
low level  Options.Engine: &EngineOptions{Rules, Hooks}; g.Engine() → *bwgame.ReplayFunctions (no stability promise)
```

Hard rule: the stream does not change engine behaviour. Hooks only read state. `tools/difftest` stays 20/20 in default mode and with `-scr-rules` against the patched harness.

## Loading

```go
func LoadData(dir string) (*Data, error) // "" means $OPENBW_DATA
func (d *Data) FilesHash() string        // SHA-256 of every data file's path, size and content

func Open(replay []byte, opts Options) (*Game, error)

type Options struct {
	Data      *Data          // nil: LoadData("") once per process
	Kinds     Kind           // event kinds to emit; 0 means DefaultKinds, NoKinds means none
	Classes   ClassSet       // unit classes for unit events; 0 means AllClasses
	TickEvery int32          // emit Tick every N frames; 0: no ticks; negative: Open fails
	MaxFrame  int32          // stop after processing this frame; 0: the replay end; negative: Open fails
	Rules     RulesMode      // RulesAuto (default) or RulesOpenBW
	Engine    *EngineOptions // low-level access; nil for none
}

type EngineOptions struct {
	Rules *bwgame.Rules // overrides Options.Rules when not nil
	Hooks bwgame.Hooks  // raw engine hooks; they must not change engine state
}
```

- `DefaultKinds` is an explicit list of the unit, production, research and player kinds. A kind added later is opt-in.
- One `Data` is safe for concurrent use by any number of games. The engine never writes to its data (`bwgame.GlobalState`), and nothing in package `openbw` gives the caller a pointer to it. A `-race` test runs games in parallel on one `Data`.
- `EngineOptions` are the only door to package `bwgame`, which has no stability promise. A game opened with them runs on a private copy of the data, loaded again from the `Data`'s directory (about 60 ms, the cost of `LoadData`), so nothing done through `Game.Engine` or a hook can reach other games. `Game.Engine` is nil without them.
- `Hooks` run inside `Open` for the map units, before `Open` returns the `Game`, and then inside `Events` and `AdvanceTo`. A panic in a hook stops the game like an engine panic (`*EngineError`).
- `Open` downgrades the replay in memory (`downgrade.Prepare`), selects the rules and loads the map. The map units are frame 0 events and come first.
- `RulesAuto` gives SC:R rules (`scrrules.FromSettings`) for SC:R replays and exact OpenBW rules for legacy replays. `RulesOpenBW` forces OpenBW rules.

## The stream

```go
func (g *Game) Events() iter.Seq[Event]
func (g *Game) AdvanceTo(frame int32) error // step without yielding; events stay buffered
func (g *Game) Err() error // check after the loop; *EngineError if the engine failed
func (g *Game) Done() bool // end frame or MaxFrame reached, or Err != nil
func (g *Game) Frame() int32
func (g *Game) EndFrame() int32
```

- Lazy: `Events` runs one half-step of the engine when its buffer is empty, then yields what that half-step produced. The engine never runs ahead of the caller by more than one half-step.
- `break` leaves the game paused. A new call to `Events` continues with the next event. Calling `Events` inside its own loop panics.
- Each event is a new value that the caller owns. The stream never reuses it.
- `Game` is not safe for concurrent use. It holds no goroutine, so there is no `Close`.
- `Kinds` filters at the source: hooks for unsubscribed kinds are not installed, and filtered events are not allocated. `Classes` filters unit events the same way.
- A panic in the loop body is not recovered. An engine panic stops the stream and `Err` returns `*EngineError{Frame, Value, Stack}`.
- The loop body runs only when an event arrives. With sparse `Kinds` (for example only `KindPlayerWon`), use `TickEvery` as a heartbeat to stop early.
- `AdvanceTo(frame)` steps until the processing half-step that makes `Frame() >= frame`, or `Done`. It yields nothing; events of the subscribed kinds stay buffered and `Events` gives them next. For state-only or hooks-only use, `Kinds: NoKinds` buffers nothing.

### Frame semantics

Each frame has two half-steps, in the engine's own order:

| Half-step | `Event.Frame` | Events |
|---|---|---|
| Load | 0 | map and start units |
| Commands of frame F | F | `CommandResult` and the events that commands cause at once (research start, cancels, player left) |
| Processing of frame F+1 | F+1 | what units, bullets and orders do; then triggers (`PlayerDefeated`, `PlayerWon`); then `Tick` (F+1) |

So for one `Frame` value F the order is: processing events of F, `Tick` F, then the commands of F with their `CommandResult`s. `Tick` F is not the last event with frame F; state read at `Tick` F is the state that the commands of F see.

- `Seq` is strictly increasing and `Frame` never decreases.
- A `CommandResult` comes after the events it caused.
- `MaxFrame = M` stops after the processing half-step that makes `Frame() == M`. Commands of frame `M` are not executed (`obwhash -max-frames`).
- State read inside the loop is the state at the end of the half-step that produced the event, the same for every event of that half-step. Resources read at a `UnitCreated` can already include later spending of the same half-step, and `Unit(id)` can still be true at a `UnitDestroyed`. Event fields hold the values at emission time; `EngineOptions.Hooks` see the state at the moment of the event.

## State

```go
func (g *Game) Unit(id UnitID) (Unit, bool)          // false if removed
func (g *Game) Units() iter.Seq[Unit]                // not yet removed (dying too), owner 0..11, no turrets
func (g *Game) Player(owner int8) Player             // owner 0..11; others give only Owner
func (g *Game) PlayingSlots() []int8                 // occupied, computer or left at load
func (g *Game) UnitTypeInfo(t UnitType) UnitTypeInfo // this game's units.dat values
func (g *Game) Header() Header                       // the replay's game info as the engine read it
func (g *Game) SourceCommands() []SourceCommand      // the original replay's command stream, as the downgrade read it
func (g *Game) Info() EngineInfo
func (g *Game) Engine() *bwgame.ReplayFunctions      // nil without Options.Engine; reads only: do not step the engine
```

`Unit`, `Player` and `UnitTypeInfo` are copies. `Unit` has the current `Order` (`OrderNone` for none). `Player` has `Race`, resources, gathered totals, supply used and max, `VictoryState` and `Left`.

`Race`, `Controller`, `Order`, `VictoryState` and `GameType` are named types with `String`. `Order` names come from OpenBW's `bwenums.h` ("PlayerGuard"). `VictoryState.Won` is true for `VictoryWon` and `VictoryAllWon`, the engine's own test. `EngineInfo` has the module version, the OpenBW commit, the data hash, the replay's source format and `RulesInfo` (`SCR`, `Custom`, object limits and the rule flags).

`Owner` is the engine slot (0-11, 11 is neutral), the index of the slot in the game info.

## Replay header

```go
type Header struct {
	MapName  string
	EndFrame int32 // the replay's frame count
	GameType GameType
	Slots    [12]Slot
}

type Slot struct {
	Owner      int8
	PlayerID   int32 // the player ID that the slot's commands carry; -1 for none
	Name       string
	Controller Controller // ControllerOccupied, ControllerComputer, ControllerOpen, ...
	Race       Race       // RaceZerg, RaceTerran, RaceProtoss, RaceRandom, ...
	Force      int32
}
```

`Header` is what the engine read from the game info at load, before the map load changes the slots. `MapName` has the engine's Korean decoding; `Name` is the raw name without control characters.

## Source commands

```go
type SourceCommand struct {
	Frame     int32
	PlayerID  byte
	ID        byte   // command ID in the original replay; 1.21 IDs are not translated
	Action    int32  // engine action index; -1 if the downgrade dropped the command
	Dropped   string // a downgrade.Reason constant, or ""
	EndsBlock bool   // the downgrade could not read past this command; the rest of its frame block is missing
}
```

`SourceCommands` is the command stream of the original replay (SC:R or legacy), in stream order, with every command that the downgrade read, kept or dropped. It is the map from the replay's commands to the engine's actions. A joiner walks it frame by frame next to its own parse of the replay: the same frame, position in the frame, player ID and command ID name the same command.

## Events

Every event embeds `EventBase{Seq, Frame, Cmd}` and has `Kind()`. Select with a type switch. Every unit event has a `Class`. Turret subunits are never reported.

| Event | Fires | Notes |
|---|---|---|
| `UnitCreated` | right after `createUnit` returns | A trained unit is created when its production starts. `Source` and `Parent` say how and from what. Zerg buildings, refineries, eggs and Archons are morphs, not creations. |
| `UnitCompleted` | end of `completeUnit`; end of an Archon summon | Each time a unit becomes complete: build end, end or cancel of a morph that `MorphStarted` began (a Hydralisk completes again when its Lurker Egg is cancelled), end of an Archon summon (High Templar, then Archon). |
| `UnitMorphed` | end of `morphUnit`, and Infest Command Center | Every type change, also siege mode and cancels back to Drone or Hydralisk. |
| `MorphStarted` | a Zerg morph or a refinery is paid for | `Into` is the target type. For a Refinery, Extractor or Assimilator, `Unit` is the geyser and `Morpher` the worker. It comes right after the `UnitMorphed` into the egg, cocoon, Zerg building or refinery, and before the `UnitMorphed` of a Zerg building morph (Lair, Hive, Greater Spire, Sunken, Spore), whose type changes at completion. The unit is incomplete from `MorphStarted` to its next `UnitCompleted`; a `UnitMorphed` without one (siege mode, cancels) keeps the completed state. |
| `UnitDestroyed` | start of `killUnit` | One per unit life: a refinery ID can die again after a rebuild (`Reverted` marks each death; the ID lives on as a completed neutral geyser, and a `UnitMorphed` to the geyser and a `UnitOwnerChanged` to 11 follow). `Reason`, `Killer`, `KillerType`, `KillerOwner`, `KillerParent`, `Container`: see below. |
| `UnitOwnerChanged` | `setUnitOwner` | `Reason`: Mind Control, player left, refinery, infest, addon, trigger. Not for a dying unit (after its `UnitDestroyed`), which the engine also gives to neutral when its player leaves. |
| `ProductionCancelled` | each cancel site | `What`, `Type` or `Research`, `Why` (command, producer destroyed, owner changed, no supply). |
| `ResearchStarted`, `ResearchCompleted` | cost paid; level gained | `Level` is 1 for tech. |
| `PlayerLeft`, `PlayerDefeated`, `PlayerWon` | leave command or drop; victory triggers | `PlayerWon` can be missing when the replay ends first. |
| `CommandResult` | after each executed command | See below. |
| `Tick` | every `TickEvery` frames, after the processing half-step | Before the commands of the same frame. Read state here. |

`UnitClass`: `ClassNormal`, `ClassFighter` (Interceptor, Scarab), `ClassSpiderMine`, `ClassHallucination`, `ClassSpell` (Dark Swarm, Disruption Web, Scanner Sweep). The class is known at creation.

`OnUnitDamaged` and the other raw hooks have no event kind. Use `EngineOptions.Hooks` for them.

Killer facts on `UnitDestroyed`:

- `Killer` is the unit that dealt the killing damage as the engine sees it, 0 if none; `KillerType` is then `UnitTypeNone`.
- A Scarab's bullet belongs to its Reaver, so Scarab kills name the Reaver. `KillerParent` is the Carrier when `Killer` is an Interceptor, else 0.
- A Spider Mine or a Scourge dies when it hits, and the engine clears its bullet's owner unit. Its kills have `Killer` 0 and only `KillerOwner`.
- `KillerOwner` is the player who fired the attack (the bullet owner), -1 if unknown. It can differ from the `Killer`'s current owner, for example after the player left.
- With `ReasonContainer` (a unit inside a Bunker or transport), the killer fields are those of the `Container`'s death.
- `DeathReason`: `Damage`, `Container`, `Spell`, `Cancelled`, `Expired`, `Depleted`, `Removed`, `Detonated`, `Linked`, `Trigger`, `OffCreep` (Larva), `Crushed` (under a landing building, which is the `Killer`), `Blocked` (a Siege Tank that sieges on a building). `Other` is not emitted by the current engine.

Fields that do not apply have explicit sentinels: `UnitTypeNone` for a unit type, `NoResearch` (`Research.Valid()` false) for a research. `ProductionCancelled` has `Type` for units and `Research` for tech and upgrades.

## Unit IDs

```go
type UnitID uint32 // bits 0-10: index + 1 (1..1700); bits 11-31: generation mod 2^21; 0: no unit

func (id UnitID) Index() uint16
func (id UnitID) Generation() uint32
func (id UnitID) LegacyTag() uint16             // 1.16 tag and OpenBW unit_id_t
func (id UnitID) MatchesSCRTag(tag uint32) bool // the downgrade's own 1.21 tag mapping
```

The ID is unique within a game. OpenBW's 16-bit `unit_id_t` wraps after 32 reuses of a slot.

## Command results

```go
type CommandResult struct {
	EventBase
	Action      int32       // index in the engine's action stream; Cmd == Action
	Source      int32       // index in SourceCommands(); -1 if the downgrade has no source for the action
	Owner       int8        // player slot that issued the command
	ActionID    byte        // legacy action ID the engine executed
	Outcome     Outcome     // Accepted, Rejected, NoEffect
	Rejections  []Rejection // {Unit, Reason}: the engine's own refusing checks
	Selection   []UnitID    // the units the command acts on; for selects and hotkeys, the selection after it
	MissingTags uint8       // Select/Shift Select tags that resolve to no unit
	ForeignTags uint8       // tags of another player's units in a Select of 2+ tags
}
```

- `Accepted`/`Rejected` is the `bool` the engine's action function returns. For orders it means at least one unit took the order, not that the order succeeded.
- `NoEffect` is an accepted Build with a `Rejection`: `actionBuild` returns true also when placement, supply or resources fail.
- `RejectReason`: `NoUnit`, `NotOwner`, `Requirements`, `Invalid`, `QueueFull`, `Busy`, `Supply`, `Minerals`, `Gas`, `Placement`. Reasons exist for Train, Train Fighter, Unit and Building Morph, Build, Research and Upgrade.
- `Source`: the downgrade numbers every command it reads and records the action that each kept command becomes. `SourceCommands()[Source].Action == Action`. The player ID is `Header().Slots[Owner].PlayerID`.

`EventBase.Cmd` (only with `KindCommandResult`) links an event to its command in two exact ways:

1. Events between the start and end of command `n` get `Cmd = n`.
2. A shadow of every build queue: `UnitCreated` of trained units, buildings and addons, and `MorphStarted` (also of refineries), take the command that pushed the production. The second unit of an egg takes the egg's command.

Effects of orders after the command returns (kills, spells, Mind Control), Larva spawns, map and trigger units, and completions are not linked (`Cmd = -1`).

## Decisions and alternatives rejected

| Decision | Rejected alternatives and why |
|---|---|
| A library with Go values only: no output files, no JSON format, no CLI. | **A screp-like `Run` with a JSON `Result`**: it fixes one storage shape and one set of statistics for every user. The replay plus the engine is the storage; tools derive what they need. |
| Raw engine facts only. | **Army value, income, desync and trust heuristics in the library**: they are opinions on top of facts and would need a second copy of rules that tools want to own. |
| Streaming base: `Events() iter.Seq[Event]` + `Err()`, like `bufio.Scanner`. | **Channels**: a goroutine per game, leaks on early stop, slower. **Callbacks only** (raw `bwgame.Hooks`): no natural early stop, state is mid-mutation. **`iter.Seq2[Event, error]`**: the error in every element although there is at most one. **`Next()`/`Event()`**: `iter.Pull` gives that. |
| Yield at half-step boundaries (after commands, after processing). | **Yield inside hooks**: state is mid-engine-function, and `break` would need to unwind the engine. **One yield per frame**: command events would see state after the next processing step. |
| Events are pointer structs behind a sealed interface; a type switch selects them. | **One struct with a Kind and all fields**: unclear which fields are valid. |
| Filter at subscription: hooks for unsubscribed kinds stay nil. `NoKinds` and `AdvanceTo` for state-only and hooks-only use. | **Emit everything and filter after**: allocations for events nobody reads. **The `KindTick` with `TickEvery: 0` trick**: hides the intent. |
| Explicit sentinels (`UnitTypeNone`, `NoResearch`) for fields that do not apply. | **Zero values**: 0 is Marine and Stim Packs, so a naive ingestion writes wrong facts. |
| Engine terms only; screpsim joins them to screp. The library packages do not import screp. | **screp types in the API (`Replay()`, `ScrepIndex`, the screp `PlayerID`)**: every engine release would depend on a screp release, and screp users would get two copies of the parse. |
| `downgrade` reads the replay sections itself (zlib, PKWARE, the 1.21 command layouts it needs). | **screp's decoder**: a dependency for about 200 lines, and the downgrade must see the same command boundaries as the engine. |
| Unit, tech and upgrade names from OpenBW's `bwenums.h` ("Terran Marine"), generated by `tools/genenums`. | **screp's names**: a screp import for strings; screpsim users have screp's names for screp's IDs. |
| `UnitCreated` at production start, plus `MorphStarted` for Zerg. | **Creation at completion**: the engine creates the unit at start; it would hide cancels. **A generic ProductionStarted**: duplicates `UnitCreated` for Terran and Protoss. |
| Report fighters, mines, hallucinations and spell units, with a `Class`. | **Leave them out**: Interceptor kills could not name the killer, and mine, Scarab and hallucination lives would be missing. |
| `UnitDestroyed` at `killUnit` with a reason from the call site. | **At removal**: the killer is gone. **Killer from `LastAttackingPlayer`**: set by any damage, not the kill. |
| Refinery death is `UnitDestroyed` with `Reverted`; the ID lives on. | **A new geyser ID**: the engine keeps the same unit; later events use the old ID. |
| Unique `UnitID` from the 21-bit generation. | **16-bit `unit_id_t`**: wraps. **A creation counter**: cannot map to replay tags. |
| Rejection reasons from hooks at the engine's own `return false` sites. | **Re-check costs and supply outside the engine**: a second copy of engine logic that can drift. |
| Production links by mirroring `BuildQueue` operations. | **Match by player, type and time**: guesses. **An origin field in `bwgame.Unit`**: changes state that the hash sees. |
| Map units are frame 0 events. | **Leave them out**: minerals and geysers die or become refineries. |
| Owner is the engine slot; `Header().Slots[Owner].PlayerID` gives the replay's player ID. | **Only `PlayerID`**: neutral units and computer slots have none. |
| One `Data` per process, shared. | **One `GlobalState` per game**: memory per game for no gain. |
| Low-level access only through `EngineOptions`, on a private copy of the data. | **`Data.Global()`, `Game.Prepared()` and `Game.Engine()` on every game**: one write through them corrupts every game that shares the `Data`. **A deep copy of `GlobalState`**: pointers between its tables make a copy error-prone; loading again is simple and costs one `LoadData`. |
| Named types with `String` for races, controllers, orders, victory states and game types. | **Raw integers**: callers need the engine's tables to read them. |
| `RulesAuto`: SC:R rules for SC:R replays. | **OpenBW rules by default**: Pole Star replays fail to load and bullet-limit games desync ([docs/desync.md](docs/desync.md)). |
| Enums have `String` only; consumers choose their own encoding. | **`MarshalJSON`/`MarshalText` on every type**: an output format by the back door. |
