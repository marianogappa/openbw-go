package openbw_test

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func captureStdout(t *testing.T, run func()) string {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	stdout := os.Stdout
	os.Stdout = w
	done := make(chan string)
	go func() {
		out, _ := io.ReadAll(r)
		done <- string(out)
	}()
	defer func() { os.Stdout = stdout }()
	run()
	w.Close()
	return <-done
}

func TestExamples(t *testing.T) {
	if os.Getenv("OPENBW_DATA") == "" {
		dir, _ := filepath.Abs("testdata/bwdata")
		if _, err := os.Stat(filepath.Join(dir, "arr", "units.dat")); err != nil {
			t.Skip("no game data")
		}
		t.Setenv("OPENBW_DATA", dir)
	}
	for _, path := range exampleReplays {
		if _, err := os.Stat(path); err != nil {
			t.Skip("no example replay")
		}
	}
	examples := []struct {
		name string
		run  func()
		want []string
	}{
		{"Example", Example, []string{"3380 Terran SCV of player 1 killed by Zerg Zergling of player 0\n", "end frame 21140\n"}},
		{"filterKinds", Example_filterKinds, []string{"1039 1 Built Terran Barracks\n", "1513 0 Morph Zerg Spawning Pool\n"}},
		{"readState", Example_readState, []string{"frame 4320 player 1: 172 minerals, 0 gas, supply 19/26\n"}},
		{"earlyStop", Example_earlyStop, []string{"frame 370: player 1 command 0x1f rejected: [{3615 Minerals}]\nselected Terran Command Center at 288,2960\nstopped at frame 370 of 21140\n"}},
		{"parallel", Example_parallel, []string{"game 0: 224 units destroyed\ngame 1: 338 units destroyed\n"}},
	}
	for _, ex := range examples {
		out := captureStdout(t, ex.run)
		for _, line := range ex.want {
			if !strings.Contains(out, line) {
				t.Errorf("Example %s output has no line %q", ex.name, line)
			}
		}
	}
}
