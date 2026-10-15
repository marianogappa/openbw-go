package downgradecheck

import (
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
	"github.com/marianogappa/openbw-go/downgrade"
)

const (
	slotCount             = 12
	slotsOffset           = 0xa1
	slotSize              = 36
	slotPlayerIDOffset    = 4
	maxSelectionCount     = 12
	controlGroupCount     = 10
	maxGroupSubaction     = 2
	defaultReplaysPerDir  = 4
	legacyUnitIndexMask   = 0x7ff
	scrUnitIndexMask      = 0x1fff
	openBWUnitCapacity    = 1700
	scrUnitCapacity       = 3400
	firstAllocatedUnitTag = 1
)

var legacyIDOf121 = map[byte]byte{0x60: 0x14, 0x61: 0x15, 0x62: 0x29, 0x63: 0x09, 0x64: 0x0a, 0x65: 0x0b}

var openBWActionPayloadSizes = map[byte]int{
	12: 7, 13: 2, 14: 4, 18: 4, 19: 2, 20: 9, 21: 10, 24: 0, 25: 0, 26: 1,
	27: 0, 28: 0, 30: 1, 31: 2, 32: 2, 33: 1, 34: 1, 35: 2, 37: 1, 38: 1,
	39: 0, 40: 1, 41: 2, 42: 0, 43: 1, 44: 1, 45: 1, 46: 0, 47: 4, 48: 1,
	49: 0, 50: 1, 51: 0, 52: 0, 53: 2, 54: 0, 87: 1, 88: 4, 90: 0, 92: 81,
}

type action struct {
	frame    uint32
	playerID byte
	id       byte
	fields   string
}

func replaysDir() string {
	if dir := os.Getenv("OPENBW_REPLAYS"); dir != "" {
		return filepath.Join(dir, "cwalggdl")
	}
	return "../../../testdata/replays/cwalggdl"
}

func replaysPerDirectory() int {
	if n, err := strconv.Atoi(os.Getenv("DOWNGRADE_REPLAYS_PER_DIR")); err == nil && n > 0 {
		return n
	}
	return defaultReplaysPerDir
}

func sampleReplays(t *testing.T) []string {
	byDir := map[string][]string{}
	err := filepath.WalkDir(replaysDir(), func(path string, d os.DirEntry, err error) error {
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
	var sample []string
	for _, files := range byDir {
		sort.Strings(files)
		n := min(len(files), replaysPerDirectory())
		for i := range n {
			sample = append(sample, files[i*len(files)/n])
		}
	}
	sort.Strings(sample)
	if len(sample) == 0 {
		t.Skip("no replays found")
	}
	return sample
}

func TestDowngradeKeepsScrepCommands(t *testing.T) {
	commands := 0
	for _, path := range sampleReplays(t) {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		n, err := check(data)
		if err != nil {
			t.Errorf("%s: %v", path, err)
		}
		commands += n
	}
	t.Logf("%d commands compared", commands)
}

func check(original []byte) (int, error) {
	out, _, err := downgrade.Downgrade(original)
	if err != nil {
		return 0, err
	}
	legacy, err := downgrade.ReadUncompressedLegacy(out)
	if err != nil {
		return 0, fmt.Errorf("read output: %w", err)
	}
	slotIDs := slotPlayerIDs(legacy.GameInfo)
	got, err := parseActions(legacy.Actions, slotIDs)
	if err != nil {
		return 0, fmt.Errorf("parse output actions: %w", err)
	}
	r, err := repparser.ParseConfig(original, repparser.Config{Commands: true, Logger: log.New(io.Discard, "", 0)})
	if err != nil {
		return 0, err
	}
	var want []action
	for _, cmd := range r.Commands.Cmds {
		if a := screpAction(cmd); keptByOpenBW(a, cmd, slotIDs) {
			want = append(want, a)
		}
	}
	if len(want) != len(got) {
		return 0, fmt.Errorf("screp keeps %d commands, output has %d", len(want), len(got))
	}
	for i := range want {
		if want[i] != got[i] {
			return 0, fmt.Errorf("command %d: screp %+v, output %+v", i, want[i], got[i])
		}
	}
	return len(got), nil
}

func slotPlayerIDs(gameInfo []byte) [slotCount]uint32 {
	var ids [slotCount]uint32
	for i := range ids {
		ids[i] = binary.LittleEndian.Uint32(gameInfo[slotsOffset+i*slotSize+slotPlayerIDOffset:])
	}
	return ids
}

func screpAction(cmd repcmd.Cmd) action {
	base := cmd.BaseCmd()
	return action{frame: uint32(base.Frame), playerID: base.PlayerID, id: screpLegacyID(cmd), fields: screpFields(cmd)}
}

func screpUnitIndex(cmd repcmd.Cmd, tag repcmd.UnitTag) uint16 {
	if _, is121 := legacyIDOf121[cmd.BaseCmd().Type.ID]; !is121 {
		return uint16(tag) & legacyUnitIndexMask
	}
	index := uint16(tag) & scrUnitIndexMask
	switch {
	case index == firstAllocatedUnitTag:
		return index
	case index > openBWUnitCapacity && index <= scrUnitCapacity:
		return index - openBWUnitCapacity
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

func keptByOpenBW(a action, cmd repcmd.Cmd, slotIDs [slotCount]uint32) bool {
	_, readable := openBWActionPayloadSizes[a.id]
	readable = readable || a.id == 0x09 || a.id == 0x0a || a.id == 0x0b
	if !readable || !slices.Contains(slotIDs[:], uint32(a.playerID)) {
		return false
	}
	switch c := cmd.(type) {
	case *repcmd.SelectCmd:
		return len(c.UnitTags) <= maxSelectionCount
	case *repcmd.HotkeyCmd:
		return c.Group >= controlGroupCount || c.HotkeyType.ID <= maxGroupSubaction
	}
	return true
}

type reader struct {
	data []byte
	pos  int
}

func (r *reader) skip(n int) error {
	if len(r.data)-r.pos < n {
		return fmt.Errorf("read past end: need %d bytes at offset %d of %d", n, r.pos, len(r.data))
	}
	r.pos += n
	return nil
}

func (r *reader) byte() (byte, error) {
	if err := r.skip(1); err != nil {
		return 0, err
	}
	return r.data[r.pos-1], nil
}

func readPayload(r *reader, id byte) error {
	switch id {
	case 0x09, 0x0a, 0x0b:
		n, err := r.byte()
		if err != nil {
			return err
		}
		if n > maxSelectionCount {
			return fmt.Errorf("invalid selection of %d units", n)
		}
		return r.skip(int(n) * 2)
	}
	size, known := openBWActionPayloadSizes[id]
	if !known {
		return fmt.Errorf("unknown action %d", id)
	}
	if err := r.skip(size); err != nil {
		return err
	}
	if payload := r.data[r.pos-size : r.pos]; id == 0x13 && payload[1] < controlGroupCount && payload[0] > maxGroupSubaction {
		return fmt.Errorf("unknown control group subaction %d", payload[0])
	}
	return nil
}

func u16(b []byte) uint16 { return binary.LittleEndian.Uint16(b) }

func unitIndex(b []byte) uint16 { return u16(b) & legacyUnitIndexMask }

func actionFields(id byte, payload []byte) string {
	switch id {
	case 0x09, 0x0a, 0x0b:
		indexes := []uint16{}
		for i := range int(payload[0]) {
			indexes = append(indexes, unitIndex(payload[1+i*2:]))
		}
		return fmt.Sprint(indexes)
	case 0x14:
		return fmt.Sprint(u16(payload), u16(payload[2:]), unitIndex(payload[4:]), u16(payload[6:]), payload[8] != 0)
	case 0x15:
		return fmt.Sprint(u16(payload), u16(payload[2:]), unitIndex(payload[4:]), u16(payload[6:]), payload[8], payload[9] != 0)
	case 0x29:
		return fmt.Sprint(unitIndex(payload))
	}
	return ""
}

func parseActions(actions []byte, slotIDs [slotCount]uint32) ([]action, error) {
	var parsed []action
	for r := (&reader{data: actions}); r.pos != len(actions); {
		if err := r.skip(5); err != nil {
			return nil, err
		}
		frame := binary.LittleEndian.Uint32(actions[r.pos-5:])
		size := int(actions[r.pos-1])
		if err := r.skip(size); err != nil {
			return nil, err
		}
		block := &reader{data: actions[r.pos-size : r.pos]}
		for block.pos != len(block.data) {
			playerID, err := block.byte()
			if err != nil {
				return nil, err
			}
			if !slices.Contains(slotIDs[:], uint32(playerID)) {
				return nil, fmt.Errorf("frame %d: player id %d not found", frame, playerID)
			}
			id, err := block.byte()
			if err != nil {
				return nil, err
			}
			start := block.pos
			if err := readPayload(block, id); err != nil {
				return nil, fmt.Errorf("frame %d: %w", frame, err)
			}
			parsed = append(parsed, action{frame: frame, playerID: playerID, id: id, fields: actionFields(id, block.data[start:block.pos])})
		}
	}
	return parsed, nil
}
