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
	actionSelect      = 0x09
	actionShiftSelect = 0x0a
)

type result struct {
	firstInconsistentFrame int32
	inconsistentCount      int
	endFrame               int32
	err                    error
}

type checker struct {
	funcs  *bwgame.ReplayFunctions
	result result
}

func (c *checker) beforeAction(owner, actionID int32, payload []byte) {
	if actionID != actionSelect && actionID != actionShiftSelect || len(payload) == 0 {
		return
	}
	n := int(payload[0])
	if len(payload) < 1+2*n {
		return
	}
	bad := 0
	for i := range n {
		raw := uint16(payload[1+2*i]) | uint16(payload[2+2*i])<<8
		if c.tagIsInconsistent(bwgame.UnitID{RawValue: raw}, owner, n > 1) {
			bad++
		}
	}
	if bad == 0 {
		return
	}
	c.result.inconsistentCount++
	if c.result.firstInconsistentFrame == 0 {
		c.result.firstInconsistentFrame = c.funcs.State().CurrentFrame
	}
}

func (c *checker) tagIsInconsistent(id bwgame.UnitID, owner int32, multi bool) bool {
	if id.Index() == 0 {
		return false
	}
	u := c.funcs.UnitByID(id)
	if u == nil {
		return true
	}
	return multi && u.Owner != owner
}

func check(global *bwgame.GlobalState, path string, useSCRRules bool) (res result) {
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
	if useSCRRules {
		st.Game.Rules = scrrules.FromSettings(settings)
	}
	c := &checker{funcs: bwgame.NewReplayFunctions(st, &bwgame.ActionState{}, &bwgame.ReplayState{})}
	c.funcs.BeforeAction = c.beforeAction
	if err := c.funcs.LoadReplayData(loadable); err != nil {
		c.result.err = err
		return c.result
	}
	for !c.funcs.IsDone() {
		if err := c.funcs.NextFrame(); err != nil {
			c.result.err = err
			break
		}
	}
	c.result.endFrame = st.CurrentFrame
	return c.result
}

func main() {
	data := flag.String("data", cmp.Or(os.Getenv("OPENBW_DATA"), "testdata/bwdata"), "game data directory")
	useSCRRules := flag.Bool("scr-rules", false, "run the engine with the replay's SC:R rules")
	workers := flag.Int("workers", runtime.NumCPU(), "replays checked in parallel")
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "usage: gotags [-data dir] [-scr-rules] [-workers n] replay.rep...")
		flag.PrintDefaults()
	}
	flag.Parse()
	if flag.NArg() == 0 {
		flag.Usage()
		os.Exit(2)
	}
	paths := flag.Args()
	lines := make([]string, len(paths))
	jobs := make(chan int)
	var wg sync.WaitGroup
	for range *workers {
		wg.Add(1)
		global, err := bwgame.GlobalInitFromDirectory(*data)
		if err != nil {
			fmt.Fprintln(os.Stderr, "gotags:", err)
			os.Exit(1)
		}
		go func() {
			defer wg.Done()
			for i := range jobs {
				lines[i] = formatResult(paths[i], check(global, paths[i], *useSCRRules))
			}
		}()
	}
	for i := range paths {
		jobs <- i
	}
	close(jobs)
	wg.Wait()
	for _, line := range lines {
		fmt.Println(line)
	}
}

func formatResult(path string, r result) string {
	name := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	if r.err != nil {
		return fmt.Sprintf("%s error %d %v", name, r.endFrame, r.err)
	}
	return fmt.Sprintf("%s %d count=%d %d", name, r.firstInconsistentFrame, r.inconsistentCount, r.endFrame)
}
