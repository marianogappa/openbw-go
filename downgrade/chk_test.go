package downgrade

import (
	"bytes"
	"encoding/binary"
	"testing"
)

func versionChunk(version uint16) []byte {
	return encodeChunk("VER ", binary.LittleEndian.AppendUint16(nil, version))
}

func extendedStrings(strs ...string) []byte {
	data := binary.LittleEndian.AppendUint32(nil, uint32(len(strs)))
	tableEnd := 4 + 4*len(strs)
	var text []byte
	for _, s := range strs {
		data = binary.LittleEndian.AppendUint32(data, uint32(tableEnd+len(text)))
		text = append(append(text, s...), 0)
	}
	return append(data, text...)
}

func legacyStrings(t *testing.T, chk []byte) []string {
	t.Helper()
	strChunk, found := findChunk(splitChunks(chk), "STR ")
	if !found {
		t.Fatal("no STR chunk")
	}
	count := int(binary.LittleEndian.Uint16(strChunk.data))
	strs := make([]string, count)
	for i := range strs {
		strs[i] = string(cString(strChunk.data, uint32(binary.LittleEndian.Uint16(strChunk.data[2+i*2:]))))
	}
	return strs
}

func mapVersion(chk []byte) uint16 {
	ver, _ := findChunk(splitChunks(chk), "VER ")
	return binary.LittleEndian.Uint16(ver.data)
}

func TestDowngradeMapConvertsStrx(t *testing.T) {
	chk := append(versionChunk(206), encodeChunk("STRx", extendedStrings("Fighting Spirit", "", "Force 1"))...)
	chk = append(chk, encodeChunk("COLR", []byte{0, 1, 2, 3, 4, 5, 6, 7})...)
	out, changes := downgradeMap(chk)
	if got := mapVersion(out); got != 205 {
		t.Errorf("VER %d, want 205", got)
	}
	want := []string{"Fighting Spirit", "", "Force 1"}
	got := legacyStrings(t, out)
	if len(got) != len(want) {
		t.Fatalf("strings %q, want %q", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("string %d = %q, want %q", i, got[i], want[i])
		}
	}
	if len(changes) != 2 {
		t.Errorf("changes %q", changes)
	}
}

func TestDowngradeMapAddsMissingColors(t *testing.T) {
	out, _ := downgradeMap(append(versionChunk(206), encodeChunk("STR ", []byte{0, 0})...))
	colors, found := findChunk(splitChunks(out), "COLR")
	if !found || !bytes.Equal(colors.data, []byte{0, 1, 2, 3, 4, 5, 6, 7}) {
		t.Errorf("COLR %v found=%v", colors.data, found)
	}
}

func TestDowngradeMapKeepsLegacyMaps(t *testing.T) {
	chk := append(versionChunk(59), encodeChunk("STR ", []byte{0, 0})...)
	out, changes := downgradeMap(chk)
	if !bytes.Equal(out, chk) || len(changes) != 0 {
		t.Errorf("map changed: %q", changes)
	}
}

func TestEncodeLegacyStringsEmptiesOverflow(t *testing.T) {
	long := string(bytes.Repeat([]byte{'a'}, 40000))
	data, dropped := encodeLegacyStrings([][]byte{[]byte(long), []byte(long), []byte("short")})
	if dropped != 1 || len(data) > maxLegacyStrSize {
		t.Errorf("dropped %d, size %d", dropped, len(data))
	}
	third := binary.LittleEndian.Uint16(data[2+2*2:])
	if string(cString(data, uint32(third))) != "short" {
		t.Error("short string lost")
	}
}
