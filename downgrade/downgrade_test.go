package downgrade

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/icza/screp/rep/repcmd"
	"github.com/icza/screp/repparser"
)

const (
	defaultReplayDir     = "/Users/mariano.gappa/Code/go/src/github.com/marianogappa/cwalggdl"
	defaultReplaysPerDir = 4
	minimumSampleCount   = 50
)

func replayDir() string {
	if dir := os.Getenv("DOWNGRADE_REPLAY_DIR"); dir != "" {
		return dir
	}
	return defaultReplayDir
}

func replaysPerDirectory() int {
	if n, err := strconv.Atoi(os.Getenv("DOWNGRADE_REPLAYS_PER_DIR")); err == nil && n > 0 {
		return n
	}
	return defaultReplaysPerDir
}

func findReplays(t *testing.T) map[string][]string {
	byDir := map[string][]string{}
	err := filepath.WalkDir(replayDir(), func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.EqualFold(filepath.Ext(path), ".rep") {
			byDir[filepath.Dir(path)] = append(byDir[filepath.Dir(path)], path)
		}
		return nil
	})
	if err != nil {
		t.Skipf("replay dir not readable: %v", err)
	}
	return byDir
}

func evenlySpaced(files []string, n int) []string {
	sort.Strings(files)
	if len(files) <= n {
		return files
	}
	picked := make([]string, n)
	for i := range picked {
		picked[i] = files[i*len(files)/n]
	}
	return picked
}

func sampleReplays(t *testing.T) []string {
	var sample []string
	for _, files := range findReplays(t) {
		sample = append(sample, evenlySpaced(files, replaysPerDirectory())...)
	}
	sort.Strings(sample)
	if len(sample) == 0 {
		t.Skip("no replays found")
	}
	return sample
}

type dropStats struct {
	replays          int
	replaysWithDrops int
	commands         int
	dropped          int
	commandsPastEnd  int
	byID             map[byte]int
	byReason         map[string]int
	formats          map[string]int
	changes          map[string]int
}

func newDropStats() *dropStats {
	return &dropStats{byID: map[byte]int{}, byReason: map[string]int{}, formats: map[string]int{}, changes: map[string]int{}}
}

func (s *dropStats) add(result sampleResult) {
	report := result.report
	s.replays++
	s.commands += result.commands
	s.commandsPastEnd += result.commandsPastEnd
	s.dropped += report.DroppedCount()
	s.formats[report.SourceFormat]++
	if report.DroppedCount() > 0 {
		s.replaysWithDrops++
	}
	for _, d := range report.Dropped {
		s.byID[d.ID] += d.Count
		s.byReason[d.Reason] += d.Count
	}
	for _, c := range report.Changes {
		s.changes[changeKind(c)]++
	}
}

func changeKind(change string) string {
	if i := strings.Index(change, ":"); i >= 0 {
		return change[:i]
	}
	if i := strings.Index(change, "("); i >= 0 {
		return strings.TrimSpace(change[:i])
	}
	return change
}

func (s *dropStats) log(t *testing.T) {
	t.Logf("replays: %d %v, with drops: %d", s.replays, s.formats, s.replaysWithDrops)
	t.Logf("commands kept: %d, dropped: %d, at or after the end frame: %d", s.commands, s.dropped, s.commandsPastEnd)
	for id, n := range s.byID {
		t.Logf("dropped id %#02x: %d", id, n)
	}
	for reason, n := range s.byReason {
		t.Logf("dropped reason %q: %d", reason, n)
	}
	for kind, n := range s.changes {
		t.Logf("change %q: %d replays", kind, n)
	}
}

func TestDowngradeSample(t *testing.T) {
	sample := sampleReplays(t)
	if len(sample) < minimumSampleCount {
		t.Logf("only %d replays in sample", len(sample))
	}
	stats := newDropStats()
	for _, path := range sample {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		result, err := checkDowngrade(data)
		if err != nil {
			t.Errorf("%s: %v", path, err)
			continue
		}
		stats.add(result)
	}
	stats.log(t)
}

type sampleResult struct {
	report          Report
	commands        int
	commandsPastEnd int
}

func checkDowngrade(data []byte) (sampleResult, error) {
	out, report, err := Downgrade(data)
	result := sampleResult{report: report}
	if err != nil {
		return result, err
	}
	legacy, err := ReadUncompressedLegacy(out)
	if err != nil {
		return result, fmt.Errorf("read output: %w", err)
	}
	slotIDs := slotPlayerIDs(legacy.GameInfo)
	parsed, err := parseOpenBWActions(legacy.Actions, slotIDs)
	if err != nil {
		return result, fmt.Errorf("parse output actions: %w", err)
	}
	if err := checkOpenBWCanLoadMap(legacy.MapData, legacy.GameInfo); err != nil {
		return result, err
	}
	if err := checkGameInfoKept(data, legacy.GameInfo); err != nil {
		return result, err
	}
	if err := checkActionsMatchScrep(data, parsed, slotIDs); err != nil {
		return result, err
	}
	if err := checkRoundTrip(out); err != nil {
		return result, err
	}
	result.commands = len(parsed)
	result.commandsPastEnd = countPastEndFrame(parsed, legacy.GameInfo)
	return result, nil
}

func countPastEndFrame(actions []parsedAction, gameInfo []byte) int {
	endFrame := binary.LittleEndian.Uint32(gameInfo[frameCountOffset:])
	count := 0
	for _, a := range actions {
		if a.frame >= endFrame {
			count++
		}
	}
	return count
}

func checkGameInfoKept(original, gameInfo []byte) error {
	src, err := readSourceSections(original)
	if err != nil {
		return err
	}
	before := slices.Clone(src.gameInfo)
	after := slices.Clone(gameInfo)
	for slot := range colorSlotCount {
		setPlayerColor(before, slot, 0)
		setPlayerColor(after, slot, 0)
	}
	if !bytes.Equal(before, after) {
		return fmt.Errorf("game info changed outside player colors")
	}
	for slot := range colorSlotCount {
		if playerColor(gameInfo, slot) >= openBWPlayerColorSize {
			return fmt.Errorf("slot %d color %d out of range", slot, playerColor(gameInfo, slot))
		}
	}
	return nil
}

func checkRoundTrip(legacy []byte) error {
	again, report, err := Downgrade(legacy)
	if err != nil {
		return fmt.Errorf("round trip: %w", err)
	}
	if report.SourceFormat != "legacy" {
		return fmt.Errorf("round trip format %q", report.SourceFormat)
	}
	if !bytes.Equal(again, legacy) {
		return fmt.Errorf("round trip changed the replay")
	}
	return nil
}

func screpCommands(data []byte) ([]repcmd.Cmd, error) {
	r, err := repparser.ParseConfig(data, repparser.Config{Commands: true, Logger: log.New(io.Discard, "", 0)})
	if err != nil {
		return nil, err
	}
	return r.Commands.Cmds, nil
}

func asParsedAction(cmd repcmd.Cmd) parsedAction {
	base := cmd.BaseCmd()
	return parsedAction{frame: uint32(base.Frame), playerID: base.PlayerID, id: screpLegacyID(cmd), fields: screpFields(cmd)}
}

func screpUnitIndex(cmd repcmd.Cmd, tag repcmd.UnitTag) uint16 {
	if _, is121 := legacyIDOf121[cmd.BaseCmd().Type.ID]; !is121 {
		return uint16(tag) & 0x7ff
	}
	index := uint16(tag) & 0x1fff
	switch {
	case index == 1:
		return 1
	case index > 1700 && index <= 3400:
		return index - 1700
	}
	return 0
}

func screpFields(cmd repcmd.Cmd) string {
	switch c := cmd.(type) {
	case *repcmd.SelectCmd:
		indexes := make([]uint16, len(c.UnitTags))
		for i, tag := range c.UnitTags {
			indexes[i] = screpUnitIndex(cmd, tag)
		}
		return fmt.Sprint(indexes)
	case *repcmd.RightClickCmd:
		return fmt.Sprint(c.Pos.X, c.Pos.Y, screpUnitIndex(cmd, c.UnitTag), c.Unit.ID, c.Queued)
	case *repcmd.TargetedOrderCmd:
		return fmt.Sprint(c.Pos.X, c.Pos.Y, screpUnitIndex(cmd, c.UnitTag), c.Unit.ID, c.Order.ID, c.Queued)
	case *repcmd.UnloadCmd:
		return fmt.Sprint(screpUnitIndex(cmd, c.UnitTag))
	}
	return ""
}

func screpLegacyID(cmd repcmd.Cmd) byte {
	id := cmd.BaseCmd().Type.ID
	if id == repcmd.VirtualTypeIDLand {
		return repcmd.TypeIDBuild
	}
	if legacyID, is121 := legacyIDOf121[id]; is121 {
		return legacyID
	}
	return id
}

func keptByOpenBW(action parsedAction, cmd repcmd.Cmd, slotIDs [slotCount]uint32) bool {
	if !openBWReadableIDs[action.id] || !slices.Contains(slotIDs[:], uint32(action.playerID)) {
		return false
	}
	switch c := cmd.(type) {
	case *repcmd.SelectCmd:
		return len(c.UnitTags) <= maxSelectionCount
	case *repcmd.HotkeyCmd:
		return c.Group >= controlGroupCount || c.HotkeyType.ID <= maxControlGroupSubaction
	}
	return true
}

func checkActionsMatchScrep(original []byte, got []parsedAction, slotIDs [slotCount]uint32) error {
	cmds, err := screpCommands(original)
	if err != nil {
		return err
	}
	var want []parsedAction
	for _, cmd := range cmds {
		if action := asParsedAction(cmd); keptByOpenBW(action, cmd, slotIDs) {
			want = append(want, action)
		}
	}
	if len(want) != len(got) {
		return fmt.Errorf("screp keeps %d commands, output has %d", len(want), len(got))
	}
	for i := range want {
		if want[i] != got[i] {
			return fmt.Errorf("command %d: screp %+v, output %+v", i, want[i], got[i])
		}
	}
	return nil
}
