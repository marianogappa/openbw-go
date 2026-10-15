# Public API

Package `openbw`, import `github.com/marianogappa/openbw-go`. It is a library: it runs the engine on a replay and gives raw engine facts as Go values through a stream. It writes no files, has no output format and does no analysis. Recorders, statistics, build orders and desync heuristics belong to the tools that use it.

```
stream  openbw.Open(replay, Options) → *Game; for e := range g.Events(); g.Unit, g.Units, g.Player
L1      Options.Hooks: raw bwgame.Hooks, called after the stream's own hooks
L0      g.Engine() → *bwgame.ReplayFunctions (no stability promise)
```

Hard rule: the stream does not change engine behaviour. Hooks only read state. `tools/difftest` stays 20/20 in default mode and with `-scr-rules` against the patched harness.

## Loading

```go
func LoadData(dir string) (*Data, error) // "" means $OPENBW_DATA
func (d *Data) FilesHash() string        // SHA-256 of every data file's path, size and content
func (d *Data) Global() *bwgame.GlobalState

func Open(replay []byte, opts Options) (*Game, error)

type Options struct {
	Data        *Data         // nil: LoadData("") once per process
	Kinds       Kind          // event kinds to emit; 0 means DefaultKinds (all but CommandResult and Tick)
	Classes     ClassSet      // unit classes for unit events; 0 means AllClasses
	TickEvery   int32         // emit Tick every N frames; 0: no ticks
	MaxFrame    int32         // stop after processing this frame; 0: the replay end
	Rules       RulesMode     // RulesAuto (default) or RulesOpenBW
	CustomRules *bwgame.Rules // overrides Rules when not nil
	Hooks       bwgame.Hooks  // raw L1 hooks; they must not change engine state
}
```

- One `Data` is safe for concurrent use by any number of games. The engine never writes to `bwgame.GlobalState`; a `-race` test runs games in parallel on one `Data`.
- `Open` parses the replay with screp, downgrades it in memory (`downgrade.Prepare`), selects the rules and loads the map. The map units are frame 0 events and come first.
- `RulesAuto` gives SC:R rules (`scrrules.FromSettings`) for SC:R replays and exact OpenBW rules for legacy replays. `RulesOpenBW` forces OpenBW rules.

## The stream

```go
func (g *Game) Events() iter.Seq[Event]
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

### Frame semantics

Each frame has two half-steps, in the engine's own order:

| Half-step | `Event.Frame` | Events |
|---|---|---|
| Load | 0 | map and start units |
| Commands of frame F | F | `CommandResult` and the events that commands cause at once (research start, cancels, player left) |
| Processing of frame F+1 | F+1 | what units, bullets and orders do; then triggers (`PlayerDefeated`, `PlayerWon`); then `Tick` |

- `Seq` is strictly increasing and `Frame` never decreases.
- A `CommandResult` comes after the events it caused.
- `MaxFrame = M` stops after the processing half-step that makes `Frame() == M`. Commands of frame `M` are not executed (`obwhash -max-frames`).
- State read inside the loop is the state at the end of the half-step that produced the event. Event fields hold the values at emission time.

## State (L0 reads)

```go
func (g *Game) Unit(id UnitID) (Unit, bool)          // false if removed
func (g *Game) Units() iter.Seq[Unit]                // not yet removed (dying too), owner 0..11, no turrets
func (g *Game) Player(owner int) Player              // owner 0..11
func (g *Game) PlayingSlots() []int                  // occupied, computer or left at load
func (g *Game) UnitTypeInfo(t UnitType) UnitTypeInfo // this game's units.dat values
func (g *Game) Replay() *rep.Replay                  // screp's parse: header, players, map
func (g *Game) Prepared() *downgrade.Prepared        // SC:R settings, dropped commands, command index
func (g *Game) Info() EngineInfo
func (g *Game) Engine() *bwgame.ReplayFunctions
```

`Unit`, `Player` and `UnitTypeInfo` are copies. `Player` has resources, gathered totals, supply used and max, victory state and `Left`. `EngineInfo` has the module version, screp's version, the OpenBW commit, the data hash, the replay's source format and `RulesInfo` (`SCR`, `Custom`, object limits and the rule flags).

`Owner` is the engine slot (0-11, 11 is neutral): the index in screp's `Header.Slots`. Commands also carry screp's `PlayerID`.

## Events

Every event embeds `EventBase{Seq, Frame, Cmd}` and has `Kind()`. Select with a type switch. Every unit event has a `Class`. Turret subunits are never reported.

| Event | Fires | Notes |
|---|---|---|
| `UnitCreated` | right after `createUnit` returns | A trained unit is created when its production starts. `Source` and `Parent` say how and from what. Zerg buildings, refineries, eggs and Archons are morphs, not creations. |
| `UnitCompleted` | end of `completeUnit` | Once for each type a unit completes as (Larva, then Zergling; Hatchery, then Lair). |
| `UnitMorphed` | end of `morphUnit`, and Infest Command Center | Every type change, also siege mode and cancels back to Drone or Hydralisk. |
| `MorphStarted` | a Zerg morph is paid for | `Into` is the target type; `Morpher` is the Drone for an Extractor. |
| `UnitDestroyed` | start of `killUnit` | One per unit ID. `Reason`, `Killer`, `KillerType`, `KillerOwner` (-1 if unknown), `KillerParent` (Carrier or Reaver of a fighter), `Container`. A refinery death has `Reverted`: the ID lives on as a neutral geyser. |
| `UnitOwnerChanged` | `setUnitOwner` | `Reason`: Mind Control, player left, refinery, infest, addon, trigger. |
| `ProductionCancelled` | each cancel site | `What`, `Type` or `Research`, `Why` (command, producer destroyed, owner changed, no supply). |
| `ResearchStarted`, `ResearchCompleted` | cost paid; level gained | `Level` is 1 for tech. |
| `PlayerLeft`, `PlayerDefeated`, `PlayerWon` | leave command or drop; victory triggers | `PlayerWon` can be missing when the replay ends first. |
| `CommandResult` | after each executed command | See below. |
| `Tick` | every `TickEvery` frames, last in the frame | Read state here. |

`UnitClass`: `ClassNormal`, `ClassFighter` (Interceptor, Scarab), `ClassSpiderMine`, `ClassHallucination`, `ClassSpell` (Dark Swarm, Disruption Web, Scanner Sweep). The class is known at creation.

`OnUnitDamaged` and the other raw hooks have no event kind. Use `Options.Hooks` for them.

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
	ScrepIndex  int32       // index in screp's Commands.Cmds; -1 if the link is not exact
	PlayerID    byte        // screp player ID
	Owner       int8        // engine slot
	ActionID    byte        // legacy action ID the engine executed
	Outcome     Outcome     // Accepted, Rejected, NoEffect
	Rejections  []Rejection // {Unit, Reason}: the engine's own refusing checks
	Selection   []UnitID    // the units the command acts on; for selects, the selection after it
	MissingTags uint8       // Select/Shift Select tags that resolve to no unit
	ForeignTags uint8       // tags of another player's units in a Select of 2+ tags
}
```

- `Accepted`/`Rejected` is the `bool` the engine's action function returns. For orders it means at least one unit took the order, not that the order succeeded.
- `NoEffect` is an accepted Build with a `Rejection`: `actionBuild` returns true also when placement, supply or resources fail.
- `RejectReason`: `NoUnit`, `NotOwner`, `Requirements`, `Invalid`, `QueueFull`, `Busy`, `Supply`, `Minerals`, `Gas`, `Placement`. Reasons exist for Train, Train Fighter, Unit and Building Morph, Build, Research and Upgrade.
- `ScrepIndex`: the downgrade numbers every source command (`Prepared.Commands`). Both lists are walked per frame in step; linking stops in a frame at the first ID mismatch, parse error or length drop. Exact or unlinked, never a guess.

`EventBase.Cmd` (only with `KindCommandResult`) links an event to its command in two exact ways:

1. Events between the start and end of command `n` get `Cmd = n`.
2. A shadow of every build queue: `UnitCreated` of trained units, buildings and addons, and `MorphStarted`, take the command that pushed the production. The second unit of an egg takes the egg's command.

Effects of orders after the command returns (kills, spells, Mind Control), Larva spawns, map and trigger units, and completions are not linked (`Cmd = -1`).

## Decisions and alternatives rejected

| Decision | Rejected alternatives and why |
|---|---|
| A library with Go values only: no output files, no JSON format, no CLI. | **A screp-like `Run` with a JSON `Result`**: it fixes one storage shape and one set of statistics for every user. The replay plus the engine is the storage; tools derive what they need. |
| Raw engine facts only. | **Army value, income, desync and trust heuristics in the library**: they are opinions on top of facts and would need a second copy of rules that tools want to own. |
| Streaming base: `Events() iter.Seq[Event]` + `Err()`, like `bufio.Scanner`. | **Channels**: a goroutine per game, leaks on early stop, slower. **Callbacks only** (raw `bwgame.Hooks`): no natural early stop, state is mid-mutation. **`iter.Seq2[Event, error]`**: the error in every element although there is at most one. **`Next()`/`Event()`**: `iter.Pull` gives that. |
| Yield at half-step boundaries (after commands, after processing). | **Yield inside hooks**: state is mid-engine-function, and `break` would need to unwind the engine. **One yield per frame**: command events would see state after the next processing step. |
| Events are pointer structs behind a sealed interface; a type switch selects them. | **One struct with a Kind and all fields**: unclear which fields are valid. |
| Filter at subscription: hooks for unsubscribed kinds stay nil. | **Emit everything and filter after**: allocations for events nobody reads. |
| `UnitCreated` at production start, plus `MorphStarted` for Zerg. | **Creation at completion**: the engine creates the unit at start; it would hide cancels. **A generic ProductionStarted**: duplicates `UnitCreated` for Terran and Protoss. |
| Report fighters, mines, hallucinations and spell units, with a `Class`. | **Leave them out**: kills by Scarabs and Interceptors could not name the killer. |
| `UnitDestroyed` at `killUnit` with a reason from the call site. | **At removal**: the killer is gone. **Killer from `LastAttackingPlayer`**: set by any damage, not the kill. |
| Refinery death is `UnitDestroyed` with `Reverted`; the ID lives on. | **A new geyser ID**: the engine keeps the same unit; later events use the old ID. |
| Unique `UnitID` from the 21-bit generation. | **16-bit `unit_id_t`**: wraps. **A creation counter**: cannot map to replay tags. |
| Rejection reasons from hooks at the engine's own `return false` sites. | **Re-check costs and supply outside the engine**: a second copy of engine logic that can drift. |
| Production links by mirroring `BuildQueue` operations. | **Match by player, type and time**: guesses. **An origin field in `bwgame.Unit`**: changes state that the hash sees. |
| Map units are frame 0 events. | **Leave them out**: minerals and geysers die or become refineries. |
| Owner is the engine slot; command results also carry screp's `PlayerID`. | **Only `PlayerID`**: neutral units and computer slots have none. |
| One `Data` per process, shared. | **One `GlobalState` per game**: memory per game for no gain. |
| `RulesAuto`: SC:R rules for SC:R replays. | **OpenBW rules by default**: Pole Star replays fail to load and bullet-limit games desync (DESYNC.md). |
| Enums have `String` only; consumers choose their own encoding. | **`MarshalJSON`/`MarshalText` on every type**: an output format by the back door. |
