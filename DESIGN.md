# Design: a screp that knows what the engine knows

Status: draft for discussion.

## The key idea: the replay is the storage

The engine is deterministic. The replay plus the engine version together hold all game state for every frame, in a few hundred KB. So:

- Do not store state. Store **events** (small) and let people **re-run** the replay when they need something new.
- Any output is a cache that can be thrown away. A new question costs one re-run, not a new database schema.

## Layers

```
 L3  analyses       build order, army value curve, fights, ...   (optional packages)
 L2  recorder       events + sampled stats → Result (JSON / rows)  (the screp-like product)
 L1  observer hooks  OnUnitCreated, OnUnitDestroyed, ...           (zero cost if unused)
 L0  engine         Load, Step, full state access                  (the Go port)
```

Each layer uses only the layer below it. People who need something unusual go down one layer instead of asking us for a new flag.

- **L0 engine**: for power users and viewers. Full access, no stability promise on internals.
- **L1 hooks**: a small, stable set of callbacks at points where the engine already makes a decision. A hook that nobody registers costs one nil check.
- **L2 recorder**: the default product, like screp's output. It is built only from L1 hooks and public L0 reads.
- **L3 analyses**: separate packages. They read an L2 `Result`. They never touch the engine, so they work the same on cached results.

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

## L2 recorder: what costs what

| Output | Default | Size (15 min 1v1) | Notes |
|---|---|---|---|
| Header, players, map | on | tiny | from screp |
| Events | on | ~5–20k rows | |
| Enriched commands | on | ~same as screp | screp commands plus engine outcome: accepted or rejected, which units were selected, what each one produced |
| Player stats every N frames | on, N=24 (1 s) | ~1k rows/player | minerals, gas, supply used/max, workers, army value, income |
| Unit snapshots every M frames | **off** | large | opt-in field set: position, hp, shields, energy, order |
| State checkpoints | **off** | MBs | for viewers that seek; uses OpenBW's state_copier |

Options follow screp's style (opt-in flags):

```go
res, err := openbw.Run(rep, openbw.Options{
    MaxFrame:       8 * 60 * 24,
    StatsEvery:     24,
    UnitSnapshots:  &openbw.UnitSnapshotOptions{Every: 240, Fields: openbw.Position | openbw.HP},
})
```

Cost control:
- **CPU**: the simulation time grows with frames, not with the output. `MaxFrame` lets classifiers stop after the opening. A full game is still a fixed cost per replay, so measure it early (target: under 1 s for a 15 min 1v1).
- **Storage**: events and stats are small and fit in rows (frame, type, player, unit, ...), so they load straight into screpdb's SQL. Large output is opt-in and can be thrown away.

## Enriched commands: the link to screp

Today screp says "player 1 issued Train Zergling at frame 3000". With the engine we can add:
- The command was rejected (not enough minerals, no larva, supply blocked).
- It made units 812 and 813.

That answers the question screp users ask most: "what actually happened, not just what was clicked". It also gives a quality signal for free.

## Trust signals in every Result

A desync makes every engine-derived number wrong without any error. So every `Result` carries:
- engine version and a hash of the game data files,
- frames simulated, and whether the game reached its end,
- the rate of rejected Train/Build/Morph commands per player (high means a likely desync),
- commands dropped by the downgrade.

Consumers can filter out results that look wrong instead of trusting them.

## Open questions

1. One `Run` with options, or a streaming API (`for ev := range sim.Events()`) that lets the caller decide what to keep? The streaming API is more general, and `Run` can be built on it.
2. Should L3 analyses (build order, fights) live in this repo or in screpdb?
3. Output format: JSON like screp, plus a row/columnar export (Parquet or SQLite) for screpdb?
4. Name of the public package: `openbw` (root), or a separate repo when it is stable?
