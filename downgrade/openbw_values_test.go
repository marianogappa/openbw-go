package downgrade

import (
	"encoding/binary"
	"testing"
)

func TestOpenBWCRC32MatchesCppValues(t *testing.T) {
	cases := []struct {
		input string
		want  uint32
	}{
		{"", 0xffffffff},
		{"a", 0x174841bc},
		{"123456789", 0x340bc6d9},
		{"reRS", 0x2b7e7ea7},
		{"The quick brown fox jumps over the lazy dog", 0xbeb05cc6},
	}
	for _, c := range cases {
		if got := openBWCRC32([]byte(c.input)); got != c.want {
			t.Errorf("crc32(%q) = %08x, want %08x", c.input, got, c.want)
		}
	}
	big := make([]byte, 20000)
	for i := range big {
		big[i] = byte(i*7 + 3)
	}
	if got := openBWCRC32(big); got != 0x21425e9c {
		t.Errorf("crc32(big) = %08x, want 21425e9c", got)
	}
}

func TestLegacyUnitIndexMatchesOpenBWAllocationOrder(t *testing.T) {
	cases := []struct{ scrTagIndex, openBWTagIndex uint32 }{
		{1, 1}, {3400, 1700}, {3399, 1699}, {3398, 1698}, {3385, 1685}, {3384, 1684},
		{3383, 1683}, {3301, 1601}, {3248, 1548}, {2401, 701}, {1703, 3}, {1702, 2},
	}
	for _, c := range cases {
		if got, ok := legacyUnitIndex(c.scrTagIndex); !ok || got != c.openBWTagIndex {
			t.Errorf("legacyUnitIndex(%d) = %d %v, want %d", c.scrTagIndex, got, ok, c.openBWTagIndex)
		}
	}
	for _, unrepresentable := range []uint32{0, 2, 1700, 3401, 8191} {
		if got, ok := legacyUnitIndex(unrepresentable); ok {
			t.Errorf("legacyUnitIndex(%d) = %d, want no OpenBW index", unrepresentable, got)
		}
	}
}

func TestLegacyUnitTag(t *testing.T) {
	cases := []struct {
		tag121     uint32
		index      uint16
		generation uint16
	}{
		{0x00002cb4, 1552, 1},
		{0x00008bca, 1318, 4},
		{0x00010c13, 1391, 8},
		{0x00042001, 1, 1},
		{0x00400d48, 1700, 0},
		{0x0000ffff, 0, 0},
		{0x00000000, 0, 0},
	}
	for _, c := range cases {
		tag := legacyUnitTag(c.tag121)
		if index, generation := tag&0x7ff, tag>>11; index != c.index || generation != c.generation {
			t.Errorf("legacyUnitTag(%08x): index %d generation %d, want %d %d", c.tag121, index, generation, c.index, c.generation)
		}
	}
}

func TestNarrowSelectionTags(t *testing.T) {
	payload := []byte{3}
	for _, tag121 := range []uint32{0x00002cb4, 0x00010c13, 0} {
		payload = binary.LittleEndian.AppendUint32(payload, tag121)
	}
	want := []byte{3, 0x10, 0x0e, 0x6f, 0x45, 0, 0}
	if got := narrowSelectionTags(payload); string(got) != string(want) {
		t.Errorf("got % x, want % x", got, want)
	}
}

func TestDropUnknownControlGroupSubaction(t *testing.T) {
	unknownSubaction := []byte{0, 0x13, 3, 2}
	unknownSubactionPastLastGroup := []byte{0, 0x13, 3, 10}
	recall := []byte{1, 0x13, 1, 2}
	out, dropped := translateForTest(t, frameBlock(8, unknownSubaction, unknownSubactionPastLastGroup, recall))
	if want := frameBlock(8, unknownSubactionPastLastGroup, recall); string(out) != string(want) {
		t.Errorf("got % x, want % x", out, want)
	}
	assertDropped(t, dropped, []DroppedCommand{{ID: 0x13, Frame: 8, PlayerID: 0, Count: 1, Reason: ReasonUnknownGroupAction}})
}
