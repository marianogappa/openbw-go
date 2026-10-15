# Design: a foundation that knows what the engine knows

Status: implemented. [API.md](../API.md) has the Go API.

## The key idea: the replay is the storage

The engine is deterministic. The replay plus the engine version together hold all game state for every frame, in a few hundred KB. So:

- Do not store state. Store **events** (small) and let people **re-run** the replay when they need something new.
- Any output is a cache that can be thrown away. A new question costs one re-run, not a new database schema.

## Layers

```
 L3  analyses       build order, army value curve, fights, ...   (other tools)
 L2  recorders      events + sampled state → stored rows         (other tools)
 --- this module ---------------------------------------------------------
 API stream         Open, Events, Unit/Units/Player, CommandResult
 L1  observer hooks  OnUnitCreated, OnUnitDestroyed, ...           (zero cost if unused)
 L0  engine         Load, Step, full state access                  (the Go port)
```

The product is L0 + L1 + the streaming API. Each layer uses only the layer below it. People who need something unusual go down one layer instead of asking for a new flag.

- **L0 engine**: for power users and viewers. Full access through `EngineOptions` and `Game.Engine`, on a private copy of the game data. No stability promise on internals.
- **L1 hooks**: callbacks at points where the engine already makes a decision (`EngineOptions.Hooks`). A hook that nobody registers costs one nil check.
- **Streaming API**: `openbw.Open` and `Game.Events`. It turns L1 hooks into typed events, links them to commands, and gives copies of live state between events. The output is Go values, never a file.
- **L2 recorders and L3 analyses** belong to other tools. They choose what to keep, how to store it and what to derive. This module has no recorder, no output format and no derived statistics.

## L1 events (the stable foundation)

These are sparse: a normal 1v1 has thousands of events, not millions.

| Event | Fields |
|---|---|
| UnitCreated | frame, unit, type, owner, x, y, parent unit (for example the Hatchery for a Larva) |
| UnitCompleted | frame, unit |
| UnitMorphed | frame, unit, from type, to type (eggs, Lurkers, Archons, Zerg buildings) |
| UnitDestroyed | frame, unit, x, y, killer unit, killer owner |
| UnitOwnerChanged | frame, unit, from, to (Mind Control, give_unit_to) |
| ProductionCancelled | frame, unit, what (unit, building, research, upgrade) |
| ResearchStarted / ResearchCompleted | frame, player, tech or upgrade, level |
| PlayerDefeated / PlayerLeft | frame, player |
| CommandResult | frame, player, command index, accepted or rejected, units it affected |

Unit IDs: `unit_id_t` (index + generation) gives an ID that is unique within a game. We expose it as one `uint32`.

The left-out ones are on purpose: order changes, every attack and every mining trip would add millions of events. Anyone who needs them can register an L1 hook.

## Cost

- **CPU**: the simulation time grows with frames, not with the output. `MaxFrame` or a `break` stops early. Unsubscribed kinds install no hooks.
- **Memory**: one shared `Data` per process; a game holds the engine state and a small event buffer for one half-step.

## Enriched commands: the link to screp

Today screp says "player 1 issued Train Zergling at frame 3000". With the engine we can add:
- The command was rejected (not enough minerals, no larva, supply blocked).
- It made units 812 and 813 (`UnitCreated` events whose `Cmd` is the command).

That answers the question screp users ask most: "what actually happened, not just what was clicked".

The join lives in a separate module:

```
screp  <-  screpsim  ->  openbw-go
```

- openbw-go speaks engine terms only: player slots, action indexes, source command indexes and `UnitID`s. `Game.SourceCommands` and `CommandResult.Source` are the exact map from the replay's command stream to the engine's actions. `Game.Header` has the game info facts. The library packages never import screp, so an engine release never waits for a screp release.
- [screpsim](https://github.com/marianogappa/screpsim) parses with screp and returns screp's own types (`*rep.Replay`, `repcmd.Cmd`). It adds the engine facts keyed by screp's identifiers: the command index in `Commands.Cmds`, the player slot and the unit tag. A link is exact or absent.

## Out of scope

- Recorders, output formats (JSON, Parquet, SQLite) and CLIs that write reports.
- Derived numbers: army value, income, build orders, fights, desync or trust heuristics.
- State checkpoints for seeking. A viewer replays from frame 0 (about 1 s per game).
