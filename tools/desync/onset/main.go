package main

import (
	"cmp"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/marianogappa/openbw-go/bwgame"
	"github.com/marianogappa/openbw-go/downgrade"
	"github.com/marianogappa/openbw-go/scrrules"
)

const (
	gameTypeMelee       = 2
	gameTypeTopVsBottom = 15
)

type victoryEvent struct {
	frame, owner, state int32
}

type result struct {
	gameType           int32
	players            int
	inconsistentFrames []int32
	victoryEvents      []victoryEvent
	endFrame           int32
	err                error
}

func run(global *bwgame.GlobalState, path string) (res result) {
	defer func() {
		if r := recover(); r != nil {
			res.err = fmt.Errorf("%v", r)
		}
	}()
	data, err := os.ReadFile(path)
	if err != nil {
		return result{err: err}
	}
	loadable, settings, err := downgrade.PrepareForOpenBW(data)
	if err != nil {
		return result{err: err}
	}
	st := bwgame.NewState(global, bwgame.NewGameState())
	st.Game.Rules = scrrules.FromSettings(settings)
	f := bwgame.NewReplayFunctions(st, &bwgame.ActionState{}, &bwgame.ReplayState{})
	f.BeforeAction = func(owner, actionID int32, payload []byte) {
		if actionID != 0x09 && actionID != 0x0a || len(payload) == 0 {
			return
		}
		n := int(payload[0])
		if len(payload) < 1+2*n {
			return
		}
		for i := range n {
			id := bwgame.UnitID{RawValue: uint16(payload[1+2*i]) | uint16(payload[2+2*i])<<8}
			if id.Index() == 0 {
				continue
			}
			if u := f.UnitByID(id); u == nil || n > 1 && u.Owner != owner {
				res.inconsistentFrames = append(res.inconsistentFrames, st.CurrentFrame)
				return
			}
		}
	}
	f.Hooks.OnVictoryState = func(owner, state int32) {
		res.victoryEvents = append(res.victoryEvents, victoryEvent{st.CurrentFrame, owner, state})
	}
	if err := f.LoadReplayData(loadable); err != nil {
		res.err = err
		return res
	}
	res.gameType = f.ReplayState().GameType
	for i := range 8 {
		if st.Players[i].Controller == bwgame.PlayerControllerOccupied {
			res.players++
		}
	}
	if res.gameType != gameTypeMelee && res.gameType != gameTypeTopVsBottom {
		return res
	}
	for !f.IsDone() {
		if err := f.NextFrame(); err != nil {
			res.err = err
			break
		}
	}
	res.endFrame = st.CurrentFrame
	return res
}

// sustainedOnset skips the isolated inconsistencies that clean games also have (a unit dies between the
// click and the command) and returns the first of 6 inconsistent commands within 3,000 frames.
func sustainedOnset(frames []int32) int32 {
	for i := range frames {
		if i+5 < len(frames) && frames[i+5]-frames[i] < 3000 {
			return frames[i]
		}
	}
	return 0
}

func format(path string, r result) string {
	name := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	if r.err != nil {
		return fmt.Sprintf("%s\terror\t%v", name, r.err)
	}
	var events []string
	for _, e := range r.victoryEvents {
		events = append(events, fmt.Sprintf("%d:p%d:s%d", e.frame, e.owner, e.state))
	}
	return fmt.Sprintf("%s\tgt=%d\tpl=%d\tend=%d\tcount=%d\tonset=%d\tev=%s",
		name, r.gameType, r.players, r.endFrame, len(r.inconsistentFrames), sustainedOnset(r.inconsistentFrames), strings.Join(events, ","))
}

func main() {
	data := flag.String("data", cmp.Or(os.Getenv("OPENBW_DATA"), "testdata/bwdata"), "game data directory")
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "usage: onset [-data dir] replay.rep...")
		flag.PrintDefaults()
	}
	flag.Parse()
	if flag.NArg() == 0 {
		flag.Usage()
		os.Exit(2)
	}
	paths := flag.Args()
	jobs := make(chan string)
	var mu sync.Mutex
	var wg sync.WaitGroup
	for range runtime.NumCPU() {
		global, err := bwgame.GlobalInitFromDirectory(*data)
		if err != nil {
			fmt.Fprintln(os.Stderr, "onset:", err)
			os.Exit(1)
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			for path := range jobs {
				line := format(path, run(global, path))
				mu.Lock()
				fmt.Println(line)
				mu.Unlock()
			}
		}()
	}
	for _, path := range paths {
		jobs <- path
	}
	close(jobs)
	wg.Wait()
}
