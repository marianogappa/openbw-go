/*
Command openbw runs the OpenBW engine on a StarCraft: Brood War replay and prints
screp's replay info plus engine-derived data as JSON.
*/
package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"runtime"

	"github.com/icza/screp/repparser"
	openbw "github.com/marianogappa/openbw-go"
)

const appName = "openbw"

const (
	ExitCodeMissingArguments         = 1
	ExitCodeFailedToParseReplay      = 2
	ExitCodeFailedToCreateOutputFile = 3
	ExitCodeEngineError              = 5
)

var (
	version = flag.Bool("version", false, "print version info and exit")

	dataDir    = flag.String("data", "", "game data directory (default $OPENBW_DATA)")
	overview   = flag.Bool("overview", false, "print a human-readable summary (no JSON)")
	header     = flag.Bool("header", true, "print the replay header")
	mapData    = flag.Bool("map", false, "print map data")
	mapTiles   = flag.Bool("maptiles", false, "print map data tiles; valid with 'map'")
	mapResLoc  = flag.Bool("mapres", false, "print map resource locations; valid with 'map'")
	cmds       = flag.Bool("cmds", false, "print player commands")
	computed   = flag.Bool("computed", true, "print computed / derived data")
	events     = flag.Bool("events", true, "print engine events")
	kinds      = flag.String("kinds", "", "comma-separated event kinds, for example UnitCreated,UnitDestroyed (default all)")
	fighters   = flag.Bool("fighters", false, "include Interceptors and Scarabs in events")
	spells     = flag.Bool("spells", false, "include spell units (Dark Swarm, Disruption Web, Scanner Sweep) in events")
	results    = flag.String("results", "actions", "command results: 'actions', 'all' or 'none'")
	stats      = flag.Bool("stats", true, "print player stats")
	statsEvery = flag.Int("statsevery", openbw.DefaultStatsEvery, "player stats period in frames")
	units      = flag.Bool("units", false, "print unit snapshots")
	unitsEvery = flag.Int("unitsevery", 240, "unit snapshot period in frames")
	unitFields = flag.String("unitfields", "pos,hp", "unit snapshot fields: pos,hp,shields,energy,order,state,buildtime")
	neutral    = flag.Bool("neutral", false, "include neutral units in unit snapshots")
	maxFrame   = flag.Int("maxframe", 0, "stop after this frame (0: the replay end)")
	rulesMode  = flag.String("rules", "auto", "engine rules: 'auto' or 'openbw'")
	stdin      = flag.Bool("stdin", false, "read replay content from standard input instead of a file")
	outFile    = flag.String("outfile", "", "optional output file name")
	indent     = flag.Bool("indent", true, "use indentation when formatting output")
)

func main() {
	flag.Parse()
	if *version {
		printVersion()
		return
	}
	args := flag.Args()
	if !*stdin && len(args) < 1 {
		printUsage()
		os.Exit(ExitCodeMissingArguments)
	}
	opts, err := options()
	if err != nil {
		fmt.Println(err)
		os.Exit(ExitCodeMissingArguments)
	}
	replay, err := readReplay(args)
	if err != nil {
		fmt.Printf("Failed to read replay: %v\n", err)
		os.Exit(ExitCodeFailedToParseReplay)
	}
	res, runErr := openbw.Run(replay, opts)
	if res == nil {
		fmt.Printf("Failed to run replay: %v\n", runErr)
		os.Exit(ExitCodeFailedToParseReplay)
	}
	destination := os.Stdout
	if *outFile != "" {
		f, err := os.Create(*outFile)
		if err != nil {
			fmt.Printf("Failed to create output file: %v\n", err)
			os.Exit(ExitCodeFailedToCreateOutputFile)
		}
		defer f.Close()
		destination = f
	}
	if *overview {
		printOverview(destination, res)
	} else if err := writeJSON(destination, res); err != nil {
		fmt.Printf("Failed to encode output: %v\n", err)
	}
	var engineErr *openbw.EngineError
	if errors.As(runErr, &engineErr) {
		fmt.Fprintln(os.Stderr, runErr)
		destination.Close()
		os.Exit(ExitCodeEngineError)
	}
}

func options() (openbw.Options, error) {
	var opts openbw.Options
	if *dataDir != "" {
		data, err := openbw.LoadData(*dataDir)
		if err != nil {
			return opts, err
		}
		opts.Data = data
	}
	if *kinds != "" {
		k, err := openbw.ParseKinds(*kinds)
		if err != nil {
			return opts, err
		}
		opts.Kinds = k
	}
	opts.Classes = openbw.AllClasses &^ openbw.Classes(openbw.ClassFighter, openbw.ClassSpell)
	if *fighters {
		opts.Classes |= openbw.Classes(openbw.ClassFighter)
	}
	if *spells {
		opts.Classes |= openbw.Classes(openbw.ClassSpell)
	}
	mode, err := openbw.ParseResultsMode(*results)
	if err != nil {
		return opts, err
	}
	opts.CommandResults = mode
	opts.StatsEvery = int32(*statsEvery)
	if !*stats {
		opts.StatsEvery = -1
	}
	if *units {
		fields, err := openbw.ParseUnitFields(*unitFields)
		if err != nil {
			return opts, err
		}
		opts.UnitSnapshots = &openbw.UnitSnapshotOptions{Every: int32(*unitsEvery), Fields: fields, IncludeNeutral: *neutral}
	}
	opts.MaxFrame = int32(*maxFrame)
	switch *rulesMode {
	case "auto":
		opts.Rules = openbw.RulesAuto
	case "openbw":
		opts.Rules = openbw.RulesOpenBW
	default:
		return opts, fmt.Errorf("Invalid rules: %v (valid values are 'auto' and 'openbw')", *rulesMode)
	}
	opts.Commands = *cmds
	opts.MapData = *mapData
	return opts, nil
}

func readReplay(args []string) ([]byte, error) {
	if *stdin {
		return io.ReadAll(os.Stdin)
	}
	return os.ReadFile(args[0])
}

func writeJSON(w io.Writer, res *openbw.Result) error {
	if !*header {
		res.Header = nil
	}
	if !*computed {
		res.Computed = nil
	}
	if !*events {
		res.Events = nil
	}
	if res.MapData != nil {
		if !*mapTiles {
			res.MapData.Tiles = nil
		}
		if !*mapResLoc {
			res.MapData.MineralFields = nil
			res.MapData.Geysers = nil
		}
	}
	enc := json.NewEncoder(w)
	if *indent {
		enc.SetIndent("", "  ")
	}
	return enc.Encode(res)
}

func printVersion() {
	fmt.Println(appName, "version:", openbw.Version())
	fmt.Println("OpenBW commit:", openbw.OpenBWCommit)
	fmt.Println("screp version:", repparser.Version)
	fmt.Println("Platform:", runtime.GOOS, runtime.GOARCH)
	fmt.Println("Built with:", runtime.Version())
}

func printUsage() {
	fmt.Println("Usage:")
	fmt.Printf("\t%s [FLAGS] repfile.rep\n", os.Args[0])
	fmt.Println("\tRun with '-h' to see a list of available flags.")
}
