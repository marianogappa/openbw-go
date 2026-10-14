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
