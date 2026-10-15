package bench

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/marianogappa/openbw-go/bwgame"
	"github.com/marianogappa/openbw-go/downgrade"
)

const defaultDataDirectory = "../../testdata/bwdata"

func dataDirectory() string {
	if fromEnv := os.Getenv("OPENBW_DATA"); fromEnv != "" {
		return fromEnv
	}
	return defaultDataDirectory
}

func replaysDirectory() string {
	if fromEnv := os.Getenv("OPENBW_REPLAYS"); fromEnv != "" {
		return fromEnv
	}
	return "../../testdata/replays"
}

func replayListPath() string {
	if fromEnv := os.Getenv("REPLAYS"); fromEnv != "" {
		return fromEnv
	}
	return "replays.txt"
}

func readReplayList(b *testing.B) []string {
	data, err := os.ReadFile(replayListPath())
	if err != nil {
		b.Fatal(err)
	}
	var replays []string
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "#") {
			replays = append(replays, filepath.Join(replaysDirectory(), line))
		}
	}
	return replays
}

func legacyReplay(b *testing.B, path string) []byte {
	data, err := os.ReadFile(path)
	if err != nil {
		b.Skipf("replay not available: %v", err)
	}
	legacy, _, err := downgrade.Downgrade(data)
	if err != nil {
		b.Fatal(err)
	}
	return legacy
}

func runReplay(b *testing.B, global *bwgame.GlobalState, replay []byte) int32 {
	st := bwgame.NewState(global, bwgame.NewGameState())
	funcs := bwgame.NewReplayFunctions(st, &bwgame.ActionState{}, &bwgame.ReplayState{})
	if err := funcs.LoadReplayData(replay); err != nil {
		b.Fatal(err)
	}
	for !funcs.IsDone() {
		if err := funcs.NextFrame(); err != nil {
			b.Fatal(err)
		}
	}
	return st.CurrentFrame
}

func BenchmarkReplay(b *testing.B) {
	global, err := bwgame.GlobalInitFromDirectory(dataDirectory())
	if err != nil {
		b.Skipf("game data not available: %v", err)
	}
	for _, path := range readReplayList(b) {
		replay := legacyReplay(b, path)
		b.Run(strings.TrimSuffix(filepath.Base(path), ".rep"), func(b *testing.B) {
			var frames int64
			for b.Loop() {
				frames += int64(runReplay(b, global, replay))
			}
			b.ReportMetric(float64(frames)/b.Elapsed().Seconds(), "frames/s")
		})
	}
}
