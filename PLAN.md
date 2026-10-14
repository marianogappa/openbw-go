# Port plan

## Scope

About 29,000 lines of C++ in `ref/openbw`:

| File | Lines | Port? |
|---|---|---|
| `bwgame.h` | 22,400 | Yes: the engine |
| `actions.h` | 1,480 | Yes: command execution |
| `data_loading.h` | 1,800 | Yes: `.dat` parsers |
| `bwenums.h` | 2,250 | Yes: generate Go enums |
| `game_types.h`, `data_types.h`, `util.h` | 1,700 | Yes, as needed |
| `replay.h` | 280 | No: screp parses replays |
| `sync*.h`, `replay_saver.h`, `ui/`, `korean.h` | — | No |

## Rules

- Port function by function. Keep the C++ names, in snake_case converted to Go style, so each Go function maps to one C++ function.
- Keep the exact integer types, overflow, fixed-point math and the order of `lcg_rand` calls. One different random call desyncs the whole game.
- No floats where the C++ uses integers.

## Replay input: no 1.16 file

OpenBW needs a 1.16 replay only because `replay.h` reads that format. In Go we skip the file:

```
SC:R .rep ──screp──▶ header + map CHK + commands ──translate──▶ legacy 1.16 action bytes ──▶ engine
```

The translation table:

- `0x60`–`0x65` (Right Click, Targeted Order, Unload, Select, Select Add, Select Remove 1.21): the same as the legacy commands, but with 4-byte unit tags. Keep the low 16 bits.
- Drop commands that the 1.16 engine does not know. Log each drop.
- Map player slots and races from the SC:R header.

## Validation: differential testing against C++ OpenBW

1. Build a small C++ harness on `ref/openbw` that reads a map and a legacy action stream and prints a hash of the game state every frame.
2. The Go side writes the same action stream (step "translate" above) and prints the same hash.
3. The first frame where the hashes differ points to the bug.

This also solves the downgrade: the C++ harness never reads a replay file.

Test replays: `~/Code/go/src/github.com/marianogappa/cwalggdl` (about 4,000 cwal.gg replays).

## Phases

1. **Data**: `.dat` files, `iscript.bin`, `images.tbl` and GRP frame sizes, tilesets. Test against the C++ parser output.
2. **Map**: CHK parsing, tile and minitile walkability, region and pathing creation.
3. **Engine core**: game state, units, sprites, images, iscript interpreter, orders, movement, pathfinding, combat. This is most of the work.
4. **Actions**: `actions.h`.
5. **Replay input**: screp to legacy action stream.
6. **API**: `Load(rep)`, `Step(n)`, `Frame()`, unit table and player stats snapshots.
7. **Validation**: the differential harness over a large replay sample.

## Open risks

- SC:R gameplay changes after 1.16 (for example 1.18+ balance or bug fixes) could desync replays even with a perfect port.
- SC:R allows more than 1,700 units in some modes. OpenBW does not.
- OpenBW licence: ask tscmoo before publishing.
