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
	if _, err := os.Stat(exampleReplay); err != nil {
		t.Skip("no example replay")
	}
	examples := []struct {
		name string
		run  func()
		want []string
	}{
		{"buildOrder", Example_buildOrder, []string{" 0:43 1 Barracks\n", " 1:03 0 Spawning Pool\n", " 4:40 0 Metabolic Boost (Zergling Speed)\n"}},
		{"killsPerPlayer", Example_killsPerPlayer, []string{"player 0 killed 58 units\nplayer 1 killed 28 units\n"}},
		{"Run", ExampleRun, []string{"reached end: true\nlikely desync: false\nplayer 0 at 10:00: 45 workers, army value 2375/900\nplayer 1 at 10:00: 34 workers, army value 1350/125\n"}},
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
