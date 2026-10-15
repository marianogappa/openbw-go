# API validation

Validation of package `openbw` on 210 replays with [tools/apivalidate](tools/apivalidate), a test tool. Run date: 2026-10-15, at commit `6a614f2`.

```
go run ./tools/apivalidate -list tools/apivalidate/replays.txt               # semantic checks
go run ./tools/apivalidate -mode cost -list tools/apivalidate/replays.txt    # cost
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
| screp links | Every `ScrepIndex` names a screp command with the same frame and player. | 210 | Pass (1,184,569 linked) |
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

Apple M4 Pro (8 performance and 4 efficiency cores), shared with other work; times vary by about 5%.

Serial, 20 games (15,601 frames on average), each configuration on the same replay one after the other:

| Configuration | ms/game | CPU ms/game | CPU vs plain | Events/game | Allocs/game | MB allocated/game | MB live/game |
|---|---|---|---|---|---|---|---|
| Plain engine (`bwgame`) | 2,137 | 2,042 | 0 | 0 | 313,077 | 13.4 | 3.6 |
| `NoKinds` | 2,050 | 1,997 | -2.2% | 0 | 313,083 | 13.4 | 4.1 |
| `DefaultKinds` | 2,101 | 2,033 | -0.5% | 2,203 | 315,307 | 13.5 | 4.1 |
| `AllKinds` (no `Tick`) | 2,107 | 2,040 | -0.1% | 8,006 | 346,532 | 16.1 | 4.7 |

The stream's CPU cost is below the measurement noise. `AllKinds` adds about 33,000 allocations and 2.7 MB per game (events, rejections, selections and the screp parse for command links). Live memory per game is the engine state, plus 0.5 MB that the `Game` keeps (mostly the prepared replay and its command index), plus 0.6 MB with `KindCommandResult` (the screp parse and the command tracking). The game data is shared and not counted.

Throughput, 210 games on 12 workers, best of 2 rounds:

| Configuration | Wall | Games/s | Frames/s | Events/s | CPU ms/game |
|---|---|---|---|---|---|
| Plain engine | 101.1 s | 2.08 | 29k | 0 | 2,607 |
| `NoKinds` | 102.0 s | 2.06 | 29k | 0 | 2,474 |
| `DefaultKinds` | 108.0 s | 1.94 | 27k | 4k | 2,625 |
| `AllKinds` | 98.9 s | 2.12 | 30k | 17k | 2,547 |

## Engine checks

| Check | Result |
|---|---|
| `tools/difftest -list tools/difftest/replays.txt` against C++ OpenBW | 20/20 match |
| The same with the SC:R harness (`tools/scrharness`) and `-obwhash-args -scr-rules` | 20/20 match |
| `go vet ./...` | No issues |
| `go test ./...` | Pass (194 tests in 15 packages) |
