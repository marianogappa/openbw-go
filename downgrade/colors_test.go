package downgrade

import (
	"encoding/binary"
	"testing"
)

func footprint(c colorFootprint, alpha uint32) []byte {
	var b []byte
	for _, v := range []uint32{c.r, c.g, c.b, alpha} {
		b = binary.LittleEndian.AppendUint32(b, v)
	}
	return b
}

func TestColorByFootprint(t *testing.T) {
	for _, c := range colorFootprints {
		if id, ok := colorByFootprint(footprint(c, opaqueAlpha)); !ok || id != c.id {
			t.Errorf("color %d: got %d %v", c.id, id, ok)
		}
	}
	if _, ok := colorByFootprint(footprint(colorFootprints[0], 0)); ok {
		t.Error("transparent footprint matched")
	}
	if _, ok := colorByFootprint(make([]byte, cclrFootprintSize)); ok {
		t.Error("zero footprint matched")
	}
}

func TestSourceFormat(t *testing.T) {
	legacy := make([]byte, formatHeaderSize)
	modern := make([]byte, formatHeaderSize)
	modern[formatFirstDataOffset] = zlibMagic
	modern121 := make([]byte, formatHeaderSize)
	modern121[formatIDOffset] = 's'
	for want, data := range map[string][]byte{"legacy": legacy, "1.18-1.20": modern, "1.21+": modern121, "unknown": nil} {
		if got := SourceFormat(data); got != want {
			t.Errorf("SourceFormat: %q, want %q", got, want)
		}
	}
}
