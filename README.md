# openbw-go

A pure Go port of the [OpenBW](https://github.com/OpenBW/openbw) StarCraft: Brood War engine.

A library that runs the game engine on a replay and streams what the engine knows as Go values: unit lifecycles with killers, owner changes, production, research, player results, what the engine did with each command, and the live state of units and players. It writes no files and does no analysis; it is a foundation for tools that do.

Status: the engine matches C++ OpenBW bit for bit. See [API.md](API.md) for the API, [DESIGN.md](DESIGN.md) for the layers and [API_VALIDATION.md](API_VALIDATION.md) for the validation on 210 replays.

## Usage

The engine only runs as far as you read:

```go
import openbw "github.com/marianogappa/openbw-go"

g, err := openbw.Open(replay, openbw.Options{
	Kinds:    openbw.KindUnitDestroyed | openbw.KindCommandResult,
	MaxFrame: 10 * 60 * 24,
})
if err != nil {
	return err
}
for e := range g.Events() {
	switch e := e.(type) {
	case *openbw.UnitDestroyed:
		fmt.Println(e.Frame, e.Type, "killed by", e.KillerType)
	case *openbw.CommandResult:
		if e.Outcome == openbw.Rejected {
			fmt.Println(e.Frame, "player", e.Owner, "minerals", g.Player(e.Owner).Minerals, "rejected:", e.Rejections)
		}
	}
}
return g.Err()
```

Inside the loop, `g.Unit`, `g.Units` and `g.Player` read the live state. `break` stops early; a new `range g.Events()` continues. `Options.TickEvery` adds a `Tick` event every N frames for sampling state. `Options.Hooks` adds raw engine hooks, and `g.Engine()` gives the engine itself. See [example_test.go](example_test.go).

The game data comes from `$OPENBW_DATA`, or from `openbw.LoadData(dir)` passed as `Options.Data`. One `Data` is safe to share between games running in parallel.

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
