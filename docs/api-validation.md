# API validation

Validation of package `openbw` on 210 replays with [tools/apivalidate](../tools/apivalidate), a test tool. Run date: 2026-10-15, at commit `6a614f2`.

Re-run after the screp split (engine terms only, screp links moved to [screpsim](https://github.com/marianogappa/screpsim)): all semantic checks pass with the same counts. screpsim's `tools/linkcheck` on the same 210 replays: every one of the 1,184,569 screp commands has exactly one command result, in order.

Re-run on 2026-10-15 after the `EngineOptions` change (the validator reads the engine through `Game.Engine` on a private data copy): all semantic checks pass with the same counts.

Replay paths in the list are relative to `$OPENBW_REPLAYS`.

```
go run ./tools/apivalidate -list tools/apivalidate/replays.txt               # semantic checks
go run ./tools/apivalidate -mode cost -j 8 -list tools/apivalidate/replays.txt  # cost
go run ./tools/apivalidate -mode scaling -scaling-n 12 -scaling-rounds 3 -list tools/apivalidate/replays.txt  # scaling
```

## Replay set

| Source | Replays | Notes |
|---|---|---|
| cwalggdl, SC:R 1v1 | 170 | Round robin over all 14 maps: 13 each, Litmus 1 (its only replay). Includes 13 Pole Star 1.1. |
| scmapanalyzer, team and older | 40 | 25 legacy (1.16) team games, 15 SC:R (2 more Pole Star). |
| Total | 210 | 2,944,838 frames, 1,641,090 events, 1,184,569 command results. |

## Semantic checks

Each replay runs with `AllKinds`, `AllClasses` and `TickEvery: 2880` (a checkpoint every 2 minutes).

| Check | How | Replays | Result |
|---|---|---|---|
| Lifecycle | Every `UnitCompleted`, `UnitMorphed`, `MorphStarted`, `UnitOwnerChanged`, `UnitDestroyed` refers to a created unit; no event after a non-`Reverted` death; no second `UnitCreated` for an ID. | 210 | Pass (455,574 unit events) |
| Event fields | `Type`, `Owner`, `From` of each event equal the values that the earlier events give. | 210 | Pass |
| No double completion | `UnitCompleted` only for a unit that is not complete (an Archon summon is the one exception, as documented). | 210 | Pass |
| Alive units | At 1,157 checkpoints and at the end: alive units by ID, type and owner from the events equal `Game.Units()` (not dying), and type and owner counts equal the engine's own `PlayerUnits` lists (L0). | 210 | Pass (638,609 units compared) |
| Completed state | The completed state from the events equals the engine's completed flag for every alive unit at each checkpoint. | 210 | Pass |
| Command results | One `CommandResult` per executed action, `Action` 0, 1, 2, ... in order, `Cmd == Action`; frame and player equal the downgrade's source command; count equals a plain engine run. | 210 | Pass (0 source commands dropped) |
| Outcome and selection | `ActionID`, `Owner`, accepted or rejected, and `Selection` equal what a plain engine run executed (`BeforeAction`/`AfterAction`). | 210 | Pass (1,184,569 selections) |
| Command sources | `Source` names the source command of the action, with the same frame and the slot's player ID (`Header`). | 210 | Pass |
| Event links | No event has a `Cmd` of a later action. `Seq` is 1, 2, 3, ...; `Frame` never decreases. | 210 | Pass |
| Engine unchanged | Frame and state hash at the end equal a plain engine run. | 210 | Pass |
| Rules on and off, legacy | `RulesAuto` and `RulesOpenBW` give the same stream. | 25 | Pass |
| Rules on and off, SC:R | Same stream with `RulesOpenBW`. | 185 | 168 same. 15 Pole Star do not load without SC:R rules (disabled `THG2` doors). 2 differ late (frames 28,029 and 28,072); the SC:R object limits alone explain both. |
| Determinism | SHA-256 of the full stream: a parallel run (12 workers) equals a serial run. | 210 | Pass |

### Problems found and fixed

| Problem | Fix |
|---|---|
| `UnitOwnerChanged` after `UnitDestroyed` (2 replays): when a player leaves, the engine also gives that player's dying units to neutral. | The stream leaves out owner changes of dying units (`7fb5da9`). |
| The docs said `UnitCompleted` fires once per type. A cancelled morph completes the old type again (Hydralisk after a Lurker Egg, Drone after a Zerg building, Creep Colony after a Sunken). | Docs: it fires each time a unit becomes complete. |
| `MorphStarted` order was not documented: it comes after the `UnitMorphed` into an egg, cocoon, Zerg building or refinery, but before the `UnitMorphed` of a Lair, Hive, Greater Spire, Sunken or Spore morph. | Docs, with the completed state from `MorphStarted` to the next `UnitCompleted`. |
| A refinery's `Reverted` death is followed by `UnitMorphed` to the geyser and `UnitOwnerChanged` to 11; the docs did not say so. | Docs. |

## Cost

Apple M4 Pro (8 performance and 4 efficiency cores), Go 1.27. Measured again on 2026-10-15 after the screp split and the `EngineOptions` change, with only the usual background work running (about one core); times vary by 2-5%.

Serial, 20 games (15,601 frames on average), each configuration on the same replay one after the other:

| Configuration | ms/game | CPU ms/game | CPU vs plain | Events/game | Allocs/game | MB allocated/game | MB live/game |
|---|---|---|---|---|---|---|---|
| Plain engine (`bwgame`) | 1,766 | 1,767 | 0 | 0 | 322,988 | 18.4 | 5.2 |
| `NoKinds` | 1,756 | 1,757 | -0.6% | 0 | 313,197 | 14.4 | 4.1 |
| `DefaultKinds` | 1,772 | 1,774 | +0.4% | 2,203 | 315,421 | 14.5 | 4.1 |
| `AllKinds` (no `Tick`) | 1,749 | 1,750 | -0.9% | 8,006 | 327,514 | 15.3 | 4.2 |

The stream's CPU cost is below the measurement noise. `AllKinds` adds about 14,000 allocations and 0.9 MB per game (events, rejections and selections). The plain engine row includes the one-time load of its own game data in the first game. Live memory per game is the engine state, about 4 MB. The game data is shared and not counted.

Throughput, the 210 games on 8 workers, longest replay (`EndFrame`) first, best of 2 rounds:

| Configuration | Wall | Games/s | Frames/s | Events/s | CPU ms/game |
|---|---|---|---|---|---|
| Plain engine | 48.9 s | 4.30 | 60k | 0 | 1,850 |
| `NoKinds` | 50.4 s | 4.17 | 58k | 0 | 1,905 |
| `DefaultKinds` | 49.9 s | 4.21 | 59k | 9k | 1,892 |
| `AllKinds` | 50.5 s | 4.16 | 58k | 32k | 1,908 |

The set has 8-player team games of up to 131,000 frames; the longest takes 38 s alone. In list order the long games come last and the run takes 74 s (40k frames/s), because most workers wait for them. The earlier measurement (about 2 games/s on 12 workers) was on a heavily loaded machine and in list order.

Scaling with the worker count is in the [README](../README.md#performance) (`-mode scaling`).

## Engine checks

| Check | Result |
|---|---|
| `tools/difftest -list tools/difftest/replays.txt` against C++ OpenBW | 20/20 match |
| The same with the SC:R harness (`tools/scrharness`) and `-obwhash-args -scr-rules` | 20/20 match |
| `go vet ./...` | No issues |
| `go test ./...` | Pass |
