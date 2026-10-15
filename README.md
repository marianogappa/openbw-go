# openbw-go

A pure Go port of the [OpenBW](https://github.com/OpenBW/openbw) StarCraft: Brood War engine.

The goal is a library like [screp](https://github.com/icza/screp) that runs the game engine on a replay's commands, so you get engine-derived data: units, supply, deaths, resources and build timings per frame.

Status: planning. See [PLAN.md](PLAN.md).

## Game data

The engine needs Blizzard data files. They are not in this repository. Put them in `testdata/bwdata/` (git-ignored) or set `OPENBW_DATA`:

```
arr/{units,weapons,upgrades,techdata,flingy,sprites,images,orders}.dat
arr/images.tbl
scripts/iscript.bin
triggers/melee.trg
unit/**/*.grp, unit/**/*.lo?   (every file that arr/images.tbl names)
tileset/{badlands,platform,install,ashworld,jungle,desert,ice,twilight}.{cv5,vf4}
```

All paths are lowercase. The loader converts each requested path with `strings.ToLower(strings.ReplaceAll(p, "\\", "/"))`.

These are the classic (SD) files from an SC:R install. To extract them again, use [tools/bwdata-export](tools/bwdata-export): it needs [CascLib](https://github.com/ladislav-zezula/CascLib) and SC:R installed.

## Reference source

`ref/openbw` is a git-ignored clone of OpenBW at commit `4b046d5` (2026-08-13), the latest upstream. OpenBW has no licence, so its source is not committed here.

## SC:R replays

`cmd/obwhash` loads SC:R replays directly: it downgrades them in memory with `downgrade.PrepareForOpenBW`. Legacy replays load unchanged.

`downgrade.ReadSCRSettings` reads the SC:R object limits (`LMTS`), bug fixes (`BFIX`) and build (`GCFG`). Legacy replays get the 1.16 defaults. See [downgrade/README.md](downgrade/README.md).

## SC:R rules

`bwgame.GameState.Rules` selects the engine rules. Set it before the replay loads. The zero value and `bwgame.DefaultRules()` give exact OpenBW behaviour. `scrrules.FromSettings(settings)` gives the SC:R rules for a replay's `downgrade.SCRSettings`:

- Object limits from `LMTS`: bullets, sprites, images, orders and lone sprites (thingies). The "order supply low" threshold is 90% of the order limit. Only the pool sizes change. The free-list order stays the same.
- Disabled `THG2` units (Pole Star 1.1 doors): the engine creates them and sets the disabled and no-collide status flags. Default rules stop with `disable thingy unit`, as OpenBW does.
- The `BFIX`/`GCFG` order allocation rule and the flying SCV fix in `set_queued_order`. See [DESYNC.md](DESYNC.md).

The unit limit stays 1700, also with SC:R rules. An OpenBW `unit_id` has an 11-bit index, and `downgrade` maps the 1.21 unit indexes 1,701-3,400 to 1-1,700 on the assumption of a 1,700-unit array. A 3,400-unit array gives the same allocation order while fewer than 1,700 units exist, and no game in the sample comes close.

`obwhash -scr-rules` runs the game with the replay's settings and prints them to stderr. Without the flag, the engine stays bit-for-bit identical to C++ OpenBW.

[tools/scrharness](tools/scrharness) builds a C++ harness from a patched OpenBW with the same rules. Compare the two engines with:

```
tools/scrharness/build.sh $TMPDIR/scrharness
go run ./tools/difftest -out $TMPDIR/scr-difftest -harness $TMPDIR/scrharness/harness -obwhash-args "-scr-rules" -list tools/difftest/replays.txt
```

With `-scr-rules` in `-obwhash-args`, difftest gives obwhash the original SC:R replay, because the downgraded copy has no `LMTS`, `BFIX` or `GCFG`.
