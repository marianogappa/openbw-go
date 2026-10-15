# openbw-go

A pure Go port of the [OpenBW](https://github.com/OpenBW/openbw) StarCraft: Brood War engine.

The goal is a library like [screp](https://github.com/icza/screp) that runs the game engine on a replay's commands, so you get engine-derived data: units, supply, deaths, resources and build timings per frame.

Status: the engine matches C++ OpenBW bit for bit, and the public API is implemented. See [API.md](API.md) for the full spec and [DESIGN.md](DESIGN.md) for the layers.

## Usage

### Go

Stream events and stop when you have what you need. The engine only runs as far as you read:

```go
import openbw "github.com/marianogappa/openbw-go"

g, err := openbw.Open(replay, openbw.StreamOptions{
	Kinds:    openbw.KindUnitDestroyed,
	MaxFrame: 10 * 60 * 24,
})
if err != nil {
	return err
}
for e := range g.Events() {
	d := e.(*openbw.UnitDestroyed)
	if d.Killer != 0 {
		fmt.Println(d.Frame, d.Type, "killed by", d.KillerType)
	}
}
return g.Err()
```

Or record the whole game in one call, like screp:

```go
res, err := openbw.Run(replay, openbw.Options{})
```

`res` has screp's header and computed data, the engine events, command results (accepted or rejected, and the units each command produced), player stats every second and trust signals. Unit snapshots and screp's commands and map data are opt-in. `Game.Engine()` gives the engine itself for anything else. See the examples in [example_test.go](example_test.go).

The game data comes from `$OPENBW_DATA`, or from `openbw.LoadData(dir)` passed as `Options.Data`. One `Data` is safe to share between games running in parallel.

### CLI

```
go install github.com/marianogappa/openbw-go/cmd/openbw@latest
openbw -data testdata/bwdata game.rep > game.json
openbw -overview game.rep
openbw -kinds UnitCreated,UnitDestroyed -results none -stats=false game.rep
openbw -units -unitsevery 240 -unitfields pos,hp,order game.rep
```

The flags follow screp (`-header`, `-map`, `-cmds`, `-computed`, `-indent`, `-outfile`, `-stdin`). Run `openbw -h` for the rest. [testdata/examples/MM-39C48C10.json](testdata/examples/MM-39C48C10.json) is the output of `openbw -statsevery 240 -results none` on a 5-minute game.

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
