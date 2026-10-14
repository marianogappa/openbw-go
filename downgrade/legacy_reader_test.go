package downgrade

import (
	"bytes"
	"testing"
)

func TestLegacyRoundTripMultiSegment(t *testing.T) {
	gameInfo := bytes.Repeat([]byte{1}, gameInfoSize)
	actions := bytes.Repeat([]byte{2}, 3*segmentSize+17)
	mapData := bytes.Repeat([]byte{3}, segmentSize)
	rep, err := ReadUncompressedLegacy(encodeLegacyReplay(gameInfo, actions, mapData))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(rep.GameInfo, gameInfo) || !bytes.Equal(rep.Actions, actions) || !bytes.Equal(rep.MapData, mapData) {
		t.Error("sections changed")
	}
}

func TestLegacyRoundTripEmptyActions(t *testing.T) {
	encoded := encodeLegacyReplay(make([]byte, gameInfoSize), nil, []byte("VER "))
	rep, err := ReadUncompressedLegacy(encoded)
	if err != nil {
		t.Fatal(err)
	}
	if len(rep.Actions) != 0 {
		t.Errorf("actions % x", rep.Actions)
	}
}

func TestLegacyReaderChecksCRC(t *testing.T) {
	encoded := encodeLegacyReplay(make([]byte, gameInfoSize), []byte{1, 2, 3}, []byte("VER "))
	encoded[len(encoded)-1] ^= 0xff
	if _, err := ReadUncompressedLegacy(encoded); err == nil {
		t.Error("want a crc32 error")
	}
}

func TestDowngradeKeepsLegacyInput(t *testing.T) {
	encoded := encodeLegacyReplay(make([]byte, gameInfoSize), []byte{1, 0, 0, 0, 3, 0, 0x1a, 0}, []byte("VER "))
	out, report, err := Downgrade(encoded)
	if err != nil {
		t.Fatal(err)
	}
	if report.SourceFormat != "legacy" || !bytes.Equal(out, encoded) {
		t.Errorf("format %q, changed=%v", report.SourceFormat, !bytes.Equal(out, encoded))
	}
}

func TestOpenBWCRC32MatchesReplayH(t *testing.T) {
	if got := openBWCRC32(nil); got != 0xffffffff {
		t.Errorf("empty crc %08x", got)
	}
	if got := openBWCRC32([]byte("123456789")); got != ^uint32(0xcbf43926) {
		t.Errorf("crc %08x", got)
	}
}
