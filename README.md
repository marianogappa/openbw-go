# openbw-go

A pure Go port of the [OpenBW](https://github.com/OpenBW/openbw) StarCraft: Brood War engine, with a library that plays replays and streams what the engine knows as Go values.

Replay parsers such as [screp](https://github.com/icza/screp) tell you what each player clicked. The engine tells you what happened: which units were made and when they finished, who killed what, which commands the engine refused and why, and the live state of every unit and player. The library writes no files and does no analysis; it is a foundation for tools that do.

- Package `openbw` (this module's root): open a replay, range over typed events, read live state between them. See [API.md](API.md).
- Package `bwgame`: the engine port, for low-level use. No stability promise.
- Package `downgrade`: turns SC:R replays into legacy (1.16) replays in memory, so the engine can play both.
- [screpsim](https://github.com/marianogappa/screpsim) joins the engine facts to screp's own commands, players and unit tags.

## Status

The engine is complete and plays SC:R and legacy replays to the end. The API is young: expect changes before v1.

### Accuracy against C++ OpenBW

The Go engine is bit for bit equal to C++ OpenBW at commit `4b046d5`. [tools/difftest](tools/difftest) compares a hash of the full engine state after every frame with a C++ harness:

- 20 of 20 replays equal in default mode ([tools/difftest/replays.txt](tools/difftest/replays.txt)).
- 20 of 20 equal with SC:R rules against a C++ OpenBW with the same rules ([tools/scrharness](tools/scrharness)).
- The desync investigation repeats both checks on about 100 more replays, with the same result ([docs/desync.md](docs/desync.md)).

### Accuracy against SC:R

OpenBW is a reimplementation of the 1.16 engine. SC:R (Remastered) has a few gameplay differences, and a replay only plays correctly when the engine follows them. By default (`RulesAuto`), SC:R replays run with the SC:R rules that [docs/desync.md](docs/desync.md) found: larger object limits (the bullet limit matters in big fights), disabled map doors on Pole Star 1.1, the SC:R order allocation rule and Top vs Bottom team alliances.

A desynced game shows itself in the command stream: players select units that do not exist in the engine. On all 1,462 distinct Terran games of a sample of 2,934 SC:R 1v1 replays ([docs/desync.md](docs/desync.md)):

| | OpenBW rules | SC:R rules |
|---|---|---|
| Plays in sync | 1,133 (77%) | 1,401 (96%) |
| Desyncs | 87 | 61, of which 45 on one map (Attitude 1.0) |
| Does not load (Pole Star 1.1) | 242 | 0 |

Games without a Terran desync less often: about 2% with OpenBW rules ([docs/integration.md](docs/integration.md)). The cause of the remaining desyncs is not known.

Team games: SC:R rules also ally each Top vs Bottom team at game start, as BW does. In a sample of one player's own replays, Top vs Bottom games desync in 8 of 137 (102 before). Melee-lobby team games, where players ally with commands, still desync in 33 of 95, mostly long games ([docs/desync.md](docs/desync.md), "Team games").

### API validation

[tools/apivalidate](tools/apivalidate) checks the event stream against the engine on 210 replays (SC:R and legacy, 1v1 and team): unit lifecycles, event fields, alive units and completed state at checkpoints, one command result per engine action, the same engine state as a plain engine run, and determinism under parallel runs. All checks pass. See [docs/api-validation.md](docs/api-validation.md).

## Performance

Apple M4 Pro (8 performance and 4 efficiency cores), Go 1.27. One game runs at about 10,500 frames per second on one core, so a 15-minute game (21,600 frames) takes about 2 seconds. The event stream costs less than the measurement noise.

Games are independent, so run several in parallel. Scaling, with every worker playing the same 12 replays ([tools/apivalidate](tools/apivalidate) `-mode scaling`):

| Workers | Frames/s | Speedup | Per worker |
|---|---|---|---|
| 1 | 10,566 | 1.00x | 100% |
| 2 | 20,756 | 1.96x | 98% |
| 4 | 40,150 | 3.80x | 95% |
| 6 | 58,813 | 5.57x | 93% |
| 8 | 71,493 | 6.77x | 85% |
| 10 | 77,409 | 7.33x | 73% |
| 12 | 80,052 | 7.58x | 63% |

What limits it is the hardware, not the library:

- **Performance cores.** Up to 8 workers, each one gets a performance core. Workers 9 to 12 run on the efficiency cores and add 12% in total.
- **Memory.** Each game has about 4 MB of live engine state that it reads and writes every frame. At 8 workers the per-worker speed drops by 15%; a compute-only loop drops by 2.5% on the same cores, so the rest is the shared caches and memory.
- **Not Go.** Separate processes scale the same as goroutines (8.9k frames/s each at 8, 6.7k at 12). The garbage collector uses 0.1% of the CPU, and `GOGC`, `GOMAXPROCS` and preemption settings change nothing measurable. The shared `Data` is read-only, so there is no false sharing.

**Recommended worker count: the number of performance cores** (`sysctl -n hw.perflevel0.physicalcpu` on Apple silicon; on a CPU with one core type, `runtime.NumCPU()`). More workers add little throughput and make each game slower. One `Data` serves all workers; each `Game` belongs to one goroutine.

For a batch, start the longest replays first (`Game.EndFrame` after `Open`, or the file size). A long 8-player game can take 40 seconds alone; at the end of the queue it leaves the other cores idle. On the 210-replay validation set with 8 workers, longest first takes 49 s instead of 74 s ([docs/api-validation.md](docs/api-validation.md)).

## Game data

The engine needs the game data files of StarCraft: Brood War. This repository does not include them, and you need your own copy of the game. [tools/bwdata-export](tools/bwdata-export) extracts them from a StarCraft: Remastered install into `testdata/bwdata`, which is git-ignored.

Pass the folder to `openbw.LoadData`, or set `OPENBW_DATA`. The files are the classic (SD) ones, with lowercase paths:

```
arr/{units,weapons,upgrades,techdata,flingy,sprites,images,orders}.dat
arr/images.tbl
scripts/iscript.bin
triggers/melee.trg
unit/**/*.grp, unit/**/*.lo?   (every file that arr/images.tbl names)
tileset/{badlands,platform,install,ashworld,jungle,desert,ice,twilight}.{cv5,vf4}
```

## Quick start

```
go get github.com/marianogappa/openbw-go
```

```go
package main

import (
	"fmt"
	"log"
	"os"

	openbw "github.com/marianogappa/openbw-go"
)

func main() {
	data, err := openbw.LoadData(os.Getenv("OPENBW_DATA"))
	if err != nil {
		log.Fatal(err)
	}
	replay, err := os.ReadFile(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	g, err := openbw.Open(replay, openbw.Options{
		Data:  data,
		Kinds: openbw.KindUnitDestroyed | openbw.KindCommandResult,
	})
	if err != nil {
		log.Fatal(err)
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
	if err := g.Err(); err != nil {
		log.Fatal(err)
	}
}
```

The engine runs only as far as you read. `break` stops early, and a new `range g.Events()` continues. `g.Unit`, `g.Units` and `g.Player` read the live state inside the loop. `Options.TickEvery` adds a `Tick` event every N frames for sampling state, and `Options.MaxFrame` stops at a frame. `g.Header()` has the replay's game info, and `g.SourceCommands()` maps the replay's commands to the engine's actions.

More: [API.md](API.md) (the whole API and its design decisions), [example_test.go](example_test.go) and the [package documentation](https://pkg.go.dev/github.com/marianogappa/openbw-go).

## SC:R rules

`Options.Rules` selects the engine rules. `RulesAuto` (the default) uses the replay's SC:R settings (`LMTS`, `BFIX`, `GCFG`) for SC:R replays, and exact OpenBW behaviour for legacy replays. `RulesOpenBW` forces exact OpenBW behaviour. `Game.Info().Rules` tells which rules a game runs with.

- Object limits from `LMTS`: bullets, sprites, images, orders and lone sprites. Only the pool sizes change; the allocation order stays the same.
- Disabled `THG2` units (the Pole Star 1.1 doors): the engine creates them disabled and without collision. OpenBW stops with `disable thingy unit`.
- The `BFIX`/`GCFG` order allocation rule and the flying SCV fix.

The unit limit stays 1,700. `downgrade` maps the 1.21 unit indexes to 1.16 tags on that assumption, and no game in the samples comes close. Details: [docs/desync.md](docs/desync.md) and [downgrade/README.md](downgrade/README.md).

## Development

Tests read the game data from `$OPENBW_DATA` (default `testdata/bwdata`) and replays from `$OPENBW_REPLAYS` (default `testdata/replays`). Both defaults are git-ignored, and tests that need the files skip when they are missing, so `go test ./...` passes on a fresh clone. Replay paths in tests and in the `replays.txt` lists are relative to `$OPENBW_REPLAYS`; the lists use SC:R 1v1 replays from [cwal.gg](https://cwal.gg) and older team replays.

```
go test ./...
go run ./tools/difftest -out $TMPDIR/difftest -list tools/difftest/replays.txt
tools/scrharness/build.sh $TMPDIR/scrharness
go run ./tools/difftest -out $TMPDIR/scr-difftest -harness $TMPDIR/scrharness/harness -obwhash-args "-scr-rules" -list tools/difftest/replays.txt
```

difftest and the C++ tools need a clone of OpenBW at commit `4b046d5` in `ref/openbw` (git-ignored).

| Tool | What it does |
|---|---|
| [cmd/obwhash](cmd/obwhash) | Plays a replay and prints state hashes; difftest's Go side. |
| [cmd/downgrade](cmd/downgrade) | Writes the legacy replay that `downgrade` makes. |
| [tools/difftest](tools/difftest), [tools/harness](tools/harness), [tools/scrharness](tools/scrharness) | Frame-by-frame comparison with C++ OpenBW. |
| [tools/apivalidate](tools/apivalidate) | Semantic checks, cost and scaling of the event stream. |
| [tools/bench](tools/bench) | Go against C++ speed, and the PGO profile of `obwhash`. |
| [tools/desync](tools/desync) | The SC:R desync investigation ([docs/desync.md](docs/desync.md)). |
| [tools/genenums](tools/genenums) | Generates `bwgame/enums.go` and the name tables from OpenBW's `bwenums.h`. |
| [tools/screp](tools/screp) | A separate module with the checks that use screp ([docs/integration.md](docs/integration.md)), so this module does not require screp. Run with `go -C tools/screp test ./...`. |

Design and investigation notes are in [docs/](docs): [design](docs/design.md), [porting map](docs/porting-map.md), [integration](docs/integration.md), [desync](docs/desync.md) and [API validation](docs/api-validation.md).

## Licence

LGPL-3.0: see [COPYING.LESSER](COPYING.LESSER) and [COPYING](COPYING). The engine is derived from OpenBW, which has no licence file; [NOTICE](NOTICE) explains the choice. No Blizzard data is included.
