# downgrade

Converts an SC:R replay into a legacy 1.16 replay that unmodified OpenBW can load.

```
go run ./cmd/downgrade in.rep out.rep
```

```go
legacy, report, err := downgrade.Downgrade(scr)
```

A legacy replay as input comes back re-encoded, with the same content.

## Output format

The output is what `ref/openbw/replay.h` reads:

1. Section: `u32` identifier `0x53526572` (`reRS`).
2. Section: the 633-byte game info.
3. Section: `u32` actions length.
4. Section: the actions.
5. Section: `u32` map length.
6. Section: the map CHK.

Each section is `u32` crc32, `u32` segment count, then for each segment a `u32` input size and the bytes. The crc32 is the one in `replay.h` (no final XOR). Each segment holds 8,192 bytes, stored uncompressed (input size equals output size), so no PKWARE compression is necessary.

The SC:R sections after the map (player names, `SKIN`, `LMTS`, `BFIX`, `CCLR`, `GCFG` and others) are dropped.

Load the output with `replay_player::load_replay_file` or `replay_functions::load_replay_data`. `replay_player::load_replay_data` does not read the section framing.

## What changes

### Game info

The 633 raw bytes are copied. Only one field changes:

| Offset | Field | Change |
|---|---|---|
| `0x251 + 4*slot` (slots 0-7) | Player color | If the `CCLR` section has a known color footprint for the slot, write that color ID. If the result is 16 or more, write the slot index. OpenBW's UI has 16 player colors and uses `.at()`. |

In the cwal.gg sample, `CCLR` is all zeros except when a player picked a custom color. The frame count is not clamped: in all 2,934 replays, only 3 commands are at or after the end frame, and OpenBW never runs them.

### Commands

The stream is frame blocks: `u32` frame, `u8` block size, then commands. Each command is a `u8` player ID, a `u8` command ID and a payload. Lengths come from `actions.h` and screp.

- `0x60`-`0x65` (1.21 Right Click, Targeted Order, Unload, Select, Select Add, Select Remove) become `0x14`, `0x15`, `0x29`, `0x09`, `0x0A`, `0x0B`. Each 4-byte unit tag becomes a 2-byte OpenBW `unit_id`:
  - A 1.21 tag is a 13-bit index (unit array index + 1) and a generation in the bits above it. An OpenBW `unit_id` is an 11-bit index and a 5-bit generation.
  - SC:R has a 3,400-unit array. OpenBW has a 1,700-unit array. Both use the `object_container` order: the first unit gets index 0, the next units get the highest free index. Thus 1.21 index 1 stays 1, and 1.21 index 1,701-3,400 becomes 1-1,700.
  - The generation is kept modulo 32.
  - A tag with index 0 or an index that OpenBW cannot have becomes 0 (no unit).
  - In the cwal.gg sample, every 1.21 tag has index 0, 1 or 1,701-3,400. The old conversion (keep the low 16 bits) made OpenBW miss almost every selection.
- These commands are dropped. Each drop goes into `Report.Dropped` (ID, frame, player, count, reason):
  - The command ID is not in `actions.h` `read_action` (for example `0x05` keep alive, `0x37` sync, `0x55` latency).
  - The player ID is not a slot player ID in the game info. OpenBW stops with an error on this.
  - A selection has more than 12 units. OpenBW stops with an error on this.
  - A control group command (`0x13`) has a group below 10 and a subaction other than 0, 1 or 2. OpenBW stops with an error on this.
  - The command ID has no known length (for example `0xD2`). The rest of that frame block is dropped, because the next command boundary is unknown.
- Block sizes are recalculated. Empty blocks are removed.

### Map CHK

OpenBW reads map `VER` 59, 63 and 205 only, and needs a `STR ` chunk. SC:R maps use `VER` 206 (or 64) and `STRx`.

| Chunk | Change |
|---|---|
| `VER ` | 206 becomes 205, 64 becomes 63. This is the "clamping VER" in the cwal.gg converter. |
| `STR ` | If missing and `STRx` exists, a `STR ` chunk is built from `STRx` and put first in the CHK. Strings that do not fit in 64 KiB become empty. |
| `COLR` | If the version is 205 and `COLR` is missing, a `COLR` with colors 0-7 is put first. |

New chunks go first, because OpenBW stops reading at the first chunk whose size is past the end of the data.

## Results

`go test ./downgrade` checks 4 replays from each directory of `cwalggdl` (108 replays). Set `DOWNGRADE_REPLAYS_PER_DIR=1000` to check all of them. For each replay, the test:

- reads the output with a reader that mirrors `replay.h`, including the crc32,
- parses every command with the lengths in `actions.h`, and checks the player IDs and selection sizes,
- compares each command (frame, player, ID) with screp's parse of the original, and for selections, right clicks, targeted orders and unloads also the positions, unit types, orders and unit indexes,
- checks that OpenBW finds every chunk it needs in the map,
- checks that downgrading the output again gives the same bytes.

All 2,934 cwal.gg replays (all 1.21+): 15,063,239 commands, 0 dropped. 2,373 maps changed `VER` 206 to 205 and got `STR ` from `STRx`. 11 player colors came from `CCLR`.

524 replays from `scmapanalyzer/replays` (323 legacy, 201 1.21+): 0 commands dropped.

### Unmodified OpenBW

A C++ program built on `ref/openbw` loaded each output with `load_replay_file` and played it to the end frame. 2,437 of the 2,934 cwal.gg replays play to the end. The 497 that fail all use the map Pole Star 1.1. OpenBW stops with `disable thingy unit`: the map has a `THG2` unit with the disabled flag `0x80`, and OpenBW does not support it. The downgrade does not change `THG2`, because clearing the flag would change the game.

`tools/harness` counts the actions where `read_action` returns `false`. With the old tag conversion, cwal.gg games had thousands of failed actions (for example 3,655 in 9,234 frames), because almost no selection found its unit. That game now has 91. With the new conversion, 42 cwal.gg replays have 0-481 failed actions each, a rate similar to legacy 1.16 replays that OpenBW reads directly.
