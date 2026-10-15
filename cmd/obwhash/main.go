package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"runtime/debug"
	"sort"
	"strconv"
	"strings"

	"github.com/marianogappa/openbw-go/bwgame"
	"github.com/marianogappa/openbw-go/downgrade"
)

const defaultDataDirectory = "/Users/mariano.gappa/Code/go/src/github.com/marianogappa/openbw-go/testdata/bwdata"

type options struct {
	dataDirectory  string
	replayPath     string
	every          int32
	dumpUnitsFrame int32
	maxFrames      int32
	noActions      bool
	scrRules       bool
}

func exitWithUsage() {
	fmt.Fprintf(os.Stderr, "usage: obwhash [-data dir] [-every N] [-dump-units F] [-max-frames M] [-no-actions] [-scr-rules] replay.rep\n")
	os.Exit(1)
}

func parseIntArgument(flag, value string) int32 {
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		fmt.Fprintf(os.Stderr, "invalid value for %s: %s\n", flag, value)
		exitWithUsage()
	}
	return int32(parsed)
}

func initialDataDirectory() string {
	if fromEnv := os.Getenv("OPENBW_DATA"); fromEnv != "" {
		return fromEnv
	}
	return defaultDataDirectory
}

func parseOptions(args []string) options {
	opts := options{dataDirectory: initialDataDirectory(), every: 1, dumpUnitsFrame: -1, maxFrames: -1}
	for i := 0; i < len(args); i++ {
		arg := args[i]
		hasValue := i+1 < len(args)
		switch {
		case arg == "-data" && hasValue:
			i++
			opts.dataDirectory = args[i]
		case arg == "-every" && hasValue:
			i++
			opts.every = parseIntArgument("-every", args[i])
		case arg == "-dump-units" && hasValue:
			i++
			opts.dumpUnitsFrame = parseIntArgument("-dump-units", args[i])
		case arg == "-max-frames" && hasValue:
			i++
			opts.maxFrames = parseIntArgument("-max-frames", args[i])
		case arg == "-no-actions":
			opts.noActions = true
		case arg == "-scr-rules":
			opts.scrRules = true
		case arg != "" && arg[0] != '-' && opts.replayPath == "":
			opts.replayPath = arg
		default:
			exitWithUsage()
		}
	}
	if opts.replayPath == "" || opts.every < 1 {
		exitWithUsage()
	}
	return opts
}

type runner struct {
	opts                  options
	out                   *bufio.Writer
	funcs                 *bwgame.ReplayFunctions
	st                    *bwgame.State
	currentFrameForErrors int32
	actionFailCount       int
}

func (r *runner) reportActionFail(frame, owner, actionID int32) {
	r.actionFailCount++
	fmt.Fprintf(os.Stderr, "action_fail %d %d %d\n", frame, owner, actionID)
}

func (r *runner) printFrameHash() {
	fmt.Fprintf(r.out, "frame %d %016x\n", r.st.CurrentFrame, bwgame.StateHash(r.st))
}

func (r *runner) printUnitRow(u *bwgame.Unit) {
	completed := 0
	if u.StatusFlags&bwgame.UnitStatusFlagCompleted != 0 {
		completed = 1
	}
	fmt.Fprintf(r.out, "%d\t%d\t%d\t%d\t%d\t%d\t%d\t%d\t%d\t%d\t%d\t%d\t%d\t%d\t%d\t%d\t%d\t%d\t%d\t%d\n",
		int32(u.Index), int32(u.UnitType.ID), u.Owner, u.Position.X, u.Position.Y,
		u.Hp.RawValue>>8, u.ShieldPoints.RawValue>>8, u.Energy.RawValue>>8,
		bwgame.UnitOrderID(u), completed,
		u.ExactPosition.X.RawValue, u.ExactPosition.Y.RawValue, u.Hp.RawValue,
		u.StatusFlags, u.OrderState, u.MovementState, u.Heading.RawValue,
		u.MainOrderTimer, u.CurrentSpeed.RawValue, imageCount(u))
}

func imageCount(u *bwgame.Unit) int {
	if u.Sprite == nil {
		return -1
	}
	count := 0
	for range u.Sprite.Images.All() {
		count++
	}
	return count
}

func (r *runner) printRandomState() {
	fmt.Fprintf(r.out, "rng state=%d total=%d", r.st.LcgRandState, r.st.TotalRandomCounts)
	for source, count := range r.st.RandomCounts {
		if count != 0 {
			fmt.Fprintf(r.out, " source_%d=%d", source, count)
		}
	}
	fmt.Fprintf(r.out, "\n")
}

func (r *runner) dumpUnits() {
	r.printRandomState()
	fmt.Fprintf(r.out, "units_at_frame %d\n", r.st.CurrentFrame)
	fmt.Fprintf(r.out, "index\ttype\towner\tx\ty\thp\tshields\tenergy\torder\tcompleted\texact_x\texact_y\thp_raw\tstatus\torder_state\tmovement_state\theading\torder_timer\tspeed\timages\n")
	for _, units := range bwgame.HashedUnitLists(r.st) {
		for u := range units.All() {
			r.printUnitRow(u)
		}
	}
	fmt.Fprintf(r.out, "end_units\n")
}

func isPlayingController(controller int32) bool {
	switch controller {
	case bwgame.PlayerControllerOccupied, bwgame.PlayerControllerComputerGame, bwgame.PlayerControllerComputer,
		bwgame.PlayerControllerUserLeft, bwgame.PlayerControllerComputerDefeated:
		return true
	}
	return false
}

func formatSupply(supply bwgame.Fp1) string {
	raw := supply.RawValue
	s := strconv.Itoa(int(raw / 2))
	if raw%2 != 0 {
		s += ".5"
	}
	return s
}

var raceNames = [3]string{"zerg", "terran", "protoss"}

func (r *runner) printPlayerSummary(i int) {
	st := r.st
	fmt.Fprintf(r.out, "player %d name=%s race=%d minerals=%d gas=%d", i, r.funcs.ReplayState().PlayerName[i],
		int32(st.Players[i].Race), st.CurrentMinerals[i], st.CurrentGas[i])
	for race := range 3 {
		fmt.Fprintf(r.out, " supply_%s=%s/%s", raceNames[race], formatSupply(st.SupplyUsed[i][race]),
			formatSupply(st.SupplyAvailable[i][race]))
	}
	fmt.Fprintf(r.out, "\n")
}

func (r *runner) printPlayerUnits(i int) {
	countByType := map[int]int{}
	for u := range r.st.PlayerUnits[i].All() {
		countByType[int(u.UnitType.ID)]++
	}
	types := make([]int, 0, len(countByType))
	for t := range countByType {
		types = append(types, t)
	}
	sort.Ints(types)
	var line strings.Builder
	fmt.Fprintf(&line, "units %d", i)
	for _, t := range types {
		fmt.Fprintf(&line, " %d=%d", t, countByType[t])
	}
	fmt.Fprintf(r.out, "%s\n", line.String())
}

func (r *runner) printSummary() {
	for i := range 12 {
		if !isPlayingController(r.st.Players[i].Controller) {
			continue
		}
		r.printPlayerSummary(i)
		r.printPlayerUnits(i)
	}
	fmt.Fprintf(r.out, "action_fails %d\n", r.actionFailCount)
}

func (r *runner) afterFrame() {
	r.currentFrameForErrors = r.st.CurrentFrame
	if r.st.CurrentFrame%r.opts.every == 0 {
		r.printFrameHash()
	}
	if r.st.CurrentFrame == r.opts.dumpUnitsFrame {
		r.dumpUnits()
	}
}

func (r *runner) shouldContinue() bool {
	if r.funcs.IsDone() {
		return false
	}
	return r.opts.maxFrames < 0 || r.st.CurrentFrame < r.opts.maxFrames
}

func (r *runner) run() error {
	global, err := bwgame.GlobalInitFromDirectory(r.opts.dataDirectory)
	if err != nil {
		return err
	}
	r.st = bwgame.NewState(global, bwgame.NewGameState())
	r.funcs = bwgame.NewReplayFunctions(r.st, &bwgame.ActionState{}, &bwgame.ReplayState{})
	r.funcs.OnActionFail = r.reportActionFail
	r.funcs.SkipActions = r.opts.noActions
	replay, err := r.loadableReplay()
	if err != nil {
		return err
	}
	if err := r.funcs.LoadReplayData(replay); err != nil {
		return err
	}
	r.afterFrame()
	for r.shouldContinue() {
		if err := r.funcs.NextFrame(); err != nil {
			return err
		}
		r.afterFrame()
	}
	fmt.Fprintf(r.out, "end %d %016x\n", r.st.CurrentFrame, bwgame.StateHash(r.st))
	r.printSummary()
	return nil
}

func (r *runner) loadableReplay() ([]byte, error) {
	data, err := os.ReadFile(r.opts.replayPath)
	if err != nil {
		return nil, err
	}
	loadable, settings, err := downgrade.PrepareForOpenBW(data)
	if r.opts.scrRules {
		if err != nil {
			return nil, fmt.Errorf("read SC:R settings: %w", err)
		}
		return loadable, applySCRRules(settings)
	}
	if err != nil {
		return data, nil
	}
	return loadable, nil
}

func applySCRRules(settings downgrade.SCRSettings) error {
	fmt.Fprintf(os.Stderr, "scr_settings limits=%+v bfix=%x game_config=%+v order_allocation_rule=%t\n",
		settings.Limits, settings.BugFixes, gameConfigText(settings.GameConfig), settings.OrderAllocationRule())
	return errors.New("-scr-rules: the engine has no SC:R rules option yet")
}

func gameConfigText(config *downgrade.GameConfig) string {
	if config == nil {
		return "none"
	}
	return fmt.Sprintf("build=%d unknown10=%d", config.Build, config.Unknown10)
}

func (r *runner) exitWithError(err any) {
	fmt.Fprintf(r.out, "error %d %v\n", r.currentFrameForErrors, err)
	r.out.Flush()
	os.Exit(2)
}

func (r *runner) exitOnPanic() {
	if recovered := recover(); recovered != nil {
		fmt.Fprintf(os.Stderr, "panic: %v\n%s", recovered, debug.Stack())
		r.exitWithError(recovered)
	}
}

func main() {
	opts := parseOptions(os.Args[1:])
	r := &runner{opts: opts, out: bufio.NewWriter(os.Stdout)}
	defer r.exitOnPanic()
	if err := r.run(); err != nil {
		r.exitWithError(err)
	}
	r.out.Flush()
}
