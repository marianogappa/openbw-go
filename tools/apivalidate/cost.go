package main

import (
	"cmp"
	"fmt"
	"os"
	"runtime"
	"slices"
	"sync/atomic"
	"syscall"
	"time"

	openbw "github.com/marianogappa/openbw-go"
	"github.com/marianogappa/openbw-go/bwgame"
	"github.com/marianogappa/openbw-go/downgrade"
)

type costConfig struct {
	name string
	run  func(data *openbw.Data, replay []byte) (frames int32, events int, keep any, err error)
}

func runPlain(data *openbw.Data, replay []byte) (int32, int, any, error) {
	prepared, err := downgrade.Prepare(replay)
	if err != nil {
		return 0, 0, nil, err
	}
	global, err := plainGlobal()
	if err != nil {
		return 0, 0, nil, err
	}
	st := bwgame.NewState(global, bwgame.NewGameState())
	st.Game.Rules = engineRules(prepared, openbw.RulesAuto)
	f := bwgame.NewReplayFunctions(st, &bwgame.ActionState{}, &bwgame.ReplayState{})
	if err := f.LoadReplayData(prepared.Loadable); err != nil {
		return 0, 0, nil, err
	}
	for !f.IsDone() {
		if err := f.NextFrame(); err != nil {
			return 0, 0, nil, err
		}
	}
	return st.CurrentFrame, 0, f, nil
}

func runStream(kinds openbw.Kind) func(*openbw.Data, []byte) (int32, int, any, error) {
	return func(data *openbw.Data, replay []byte) (int32, int, any, error) {
		g, err := openbw.Open(replay, openbw.Options{Data: data, Kinds: kinds})
		if err != nil {
			return 0, 0, nil, err
		}
		events := 0
		for range g.Events() {
			events++
		}
		return g.Frame(), events, g, g.Err()
	}
}

var costConfigs = []costConfig{
	{"plain engine", runPlain},
	{"NoKinds", runStream(openbw.NoKinds)},
	{"DefaultKinds", runStream(openbw.DefaultKinds)},
	{"AllKinds", runStream(openbw.AllKinds &^ openbw.KindTick)},
}

type costTotals struct {
	elapsed, cpu   time.Duration
	frames         int64
	events         int
	mallocs, bytes uint64
	retained       uint64
	games          int
}

func cpuTime() time.Duration {
	var u syscall.Rusage
	if err := syscall.Getrusage(syscall.RUSAGE_SELF, &u); err != nil {
		return 0
	}
	return time.Duration(u.Utime.Nano() + u.Stime.Nano())
}

func heapAfterGC() runtime.MemStats {
	runtime.GC()
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return m
}

func measureOne(cfg costConfig, data *openbw.Data, replay []byte, t *costTotals) error {
	before := heapAfterGC()
	start, startCPU := time.Now(), cpuTime()
	frames, events, keep, err := cfg.run(data, replay)
	elapsed, cpu := time.Since(start), cpuTime()-startCPU
	if err != nil {
		return err
	}
	var after runtime.MemStats
	runtime.ReadMemStats(&after)
	retained := heapAfterGC()
	runtime.KeepAlive(keep)
	t.elapsed += elapsed
	t.cpu += cpu
	t.frames += int64(frames)
	t.events += events
	t.mallocs += after.Mallocs - before.Mallocs
	t.bytes += after.TotalAlloc - before.TotalAlloc
	t.retained += retained.HeapAlloc - min(retained.HeapAlloc, before.HeapAlloc)
	t.games++
	return nil
}

func measureCost(data *openbw.Data, paths []string, workers, games int) error {
	sample := spread(paths, games)
	totals := make([]costTotals, len(costConfigs))
	for _, path := range sample {
		replay, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		for i, cfg := range costConfigs {
			if err := measureOne(cfg, data, replay, &totals[i]); err != nil {
				return fmt.Errorf("%s %s: %w", cfg.name, path, err)
			}
		}
	}
	plain := totals[0]
	fmt.Printf("serial, %d games, %d frames each on average\n", len(sample), plain.frames/int64(plain.games))
	fmt.Printf("%-14s %8s %12s %9s %12s %12s %14s %13s\n", "config", "ms/game", "CPU ms/game", "CPU +%", "events/game", "allocs/game", "MB alloc/game", "MB live/game")
	for i, cfg := range costConfigs {
		t := totals[i]
		n := float64(t.games)
		fmt.Printf("%-14s %8.0f %12.0f %8.1f%% %12.0f %12.0f %14.1f %13.1f\n", cfg.name,
			float64(t.elapsed.Milliseconds())/n, float64(t.cpu.Milliseconds())/n,
			100*(float64(t.cpu)/float64(plain.cpu)-1),
			float64(t.events)/n, float64(t.mallocs)/n, float64(t.bytes)/n/1e6, float64(t.retained)/n/1e6)
	}
	return measureThroughput(data, paths, workers)
}

type throughputRun struct {
	elapsed, cpu   time.Duration
	frames, events int64
	failed         int32
}

func measureThroughput(data *openbw.Data, paths []string, workers int) error {
	replays := make([][]byte, len(paths))
	for i, path := range paths {
		replay, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		replays[i] = replay
	}
	if err := sortLongestFirst(data, replays); err != nil {
		return err
	}
	fmt.Printf("\nthroughput, %d games, longest first, %d workers, best of 2 rounds in alternating order\n", len(paths), workers)
	best := make([]throughputRun, len(costConfigs))
	for _, i := range []int{0, 1, 2, 3, 3, 2, 1, 0} {
		run := throughput(costConfigs[i], data, replays, workers)
		if best[i].elapsed == 0 || run.elapsed < best[i].elapsed {
			best[i] = run
		}
	}
	for i, cfg := range costConfigs {
		b := best[i]
		s := b.elapsed.Seconds()
		fmt.Printf("%-14s %6.1fs  %5.2f games/s  %5.0fk frames/s  %5.0fk events/s  CPU %5.0f ms/game  failed %d\n", cfg.name, s,
			float64(len(replays))/s, float64(b.frames)/s/1e3, float64(b.events)/s/1e3,
			float64(b.cpu.Milliseconds())/float64(len(replays)), b.failed)
	}
	return nil
}

func throughput(cfg costConfig, data *openbw.Data, replays [][]byte, workers int) throughputRun {
	var frames, events atomic.Int64
	var failed atomic.Int32
	runtime.GC()
	start, startCPU := time.Now(), cpuTime()
	parallel(len(replays), workers, func(i int) struct{} {
		f, e, _, err := cfg.run(data, replays[i])
		if err != nil {
			failed.Add(1)
		}
		frames.Add(int64(f))
		events.Add(int64(e))
		return struct{}{}
	})
	return throughputRun{elapsed: time.Since(start), cpu: cpuTime() - startCPU, frames: frames.Load(), events: events.Load(), failed: failed.Load()}
}

func sortLongestFirst(data *openbw.Data, replays [][]byte) error {
	endFrames := make(map[*byte]int32, len(replays))
	for _, replay := range replays {
		g, err := openbw.Open(replay, openbw.Options{Data: data, Kinds: openbw.NoKinds})
		if err != nil {
			return err
		}
		endFrames[&replay[0]] = g.EndFrame()
	}
	slices.SortStableFunc(replays, func(a, b []byte) int { return cmp.Compare(endFrames[&b[0]], endFrames[&a[0]]) })
	return nil
}
