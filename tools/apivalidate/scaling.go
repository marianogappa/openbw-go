package main

import (
	"fmt"
	"os"
	"runtime"
	"runtime/metrics"
	"strconv"
	"strings"
	"sync"
	"time"

	openbw "github.com/marianogappa/openbw-go"
)

type scalingRun struct {
	elapsed, cpu time.Duration
	frames       int64
	gcCPU        float64
	gcCycles     uint64
	failed       int
}

var scalingMetrics = []metrics.Sample{
	{Name: "/cpu/classes/gc/total:cpu-seconds"},
	{Name: "/gc/cycles/total:gc-cycles"},
}

func readScalingMetrics() (gcCPU float64, gcCycles uint64) {
	metrics.Read(scalingMetrics)
	return scalingMetrics[0].Value.Float64(), scalingMetrics[1].Value.Uint64()
}

func parseWorkerCounts(s string) ([]int, error) {
	var counts []int
	for field := range strings.SplitSeq(s, ",") {
		n, err := strconv.Atoi(strings.TrimSpace(field))
		if err != nil || n < 1 {
			return nil, fmt.Errorf("bad worker count %q", field)
		}
		counts = append(counts, n)
	}
	return counts, nil
}

func measureScaling(data *openbw.Data, paths []string, workerCounts []int, games, rounds int, kinds openbw.Kind) error {
	sample := spread(paths, games)
	replays := make([][]byte, len(sample))
	for i, path := range sample {
		replay, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		replays[i] = replay
	}
	fmt.Printf("scaling: every worker plays the same %d games in a rotated order; kinds %v; GOMAXPROCS %d; GOGC %q; best of %d rounds\n",
		len(replays), kinds, runtime.GOMAXPROCS(0), os.Getenv("GOGC"), rounds)
	fmt.Printf("%8s %8s %12s %10s %12s %10s %9s %9s\n", "workers", "wall s", "frames/s", "games/s", "per worker", "speedup", "GC CPU%", "GC/s")
	var base float64
	for _, workers := range workerCounts {
		var best scalingRun
		for range rounds {
			run := scalingRound(data, replays, workers, kinds)
			if best.elapsed == 0 || run.elapsed < best.elapsed {
				best = run
			}
		}
		s := best.elapsed.Seconds()
		rate := float64(best.frames) / s
		if base == 0 {
			base = rate / float64(workers)
		}
		fmt.Printf("%8d %8.1f %12.0f %10.2f %12.0f %9.2fx %8.1f%% %9.1f  failed %d\n", workers, s, rate,
			float64(workers*len(replays))/s, rate/float64(workers), rate/base,
			100*best.gcCPU/best.cpu.Seconds(), float64(best.gcCycles)/s, best.failed)
	}
	return nil
}

func scalingRound(data *openbw.Data, replays [][]byte, workers int, kinds openbw.Kind) scalingRun {
	runtime.GC()
	frames := make([]int64, workers)
	failed := make([]int, workers)
	gcCPU0, gcCycles0 := readScalingMetrics()
	start, startCPU := time.Now(), cpuTime()
	var wg sync.WaitGroup
	for w := range workers {
		wg.Go(func() {
			for i := range replays {
				replay := replays[(w+i)%len(replays)]
				g, err := openbw.Open(replay, openbw.Options{Data: data, Kinds: kinds})
				if err == nil {
					for range g.Events() {
					}
					err = g.Err()
					frames[w] += int64(g.Frame())
				}
				if err != nil {
					failed[w]++
				}
			}
		})
	}
	wg.Wait()
	run := scalingRun{elapsed: time.Since(start), cpu: cpuTime() - startCPU}
	gcCPU1, gcCycles1 := readScalingMetrics()
	run.gcCPU, run.gcCycles = gcCPU1-gcCPU0, gcCycles1-gcCycles0
	for w := range workers {
		run.frames += frames[w]
		run.failed += failed[w]
	}
	return run
}
