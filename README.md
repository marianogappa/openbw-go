# openbw-go

A pure Go port of the [OpenBW](https://github.com/OpenBW/openbw) StarCraft: Brood War engine.

The goal is a library like [screp](https://github.com/icza/screp) that runs the game engine on a replay's commands, so you get engine-derived data: units, supply, deaths, resources and build timings per frame.

Status: planning. See [PLAN.md](PLAN.md).

## Game data

The engine needs Blizzard data files. They are not in this repository. Put them in `testdata/bwdata/` (git-ignored) or set `OPENBW_DATA`:

```
arr/units.dat  arr/weapons.dat  arr/upgrades.dat  arr/techdata.dat
arr/flingy.dat arr/sprites.dat  arr/images.dat    arr/orders.dat
arr/images.tbl
scripts/iscript.bin
triggers/Melee.trg
unit/**/*.grp, unit/**/*.lo?   (every file that arr/images.tbl names)
Tileset/{badlands,platform,install,AshWorld,Jungle,Desert,Ice,Twilight}.{cv5,vf4}
```

## Reference source

`ref/openbw` is a git-ignored clone of OpenBW at commit `4b046d5` (2026-08-13), the latest upstream. OpenBW has no licence, so its source is not committed here.
