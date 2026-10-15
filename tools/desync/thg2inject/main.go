package main

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
)

const (
	sectionsBeforeTrailing = 6
	mapLengthSection       = 4
	mapSection             = 5
	encodedLengthOffset    = 16
	startLocationUnitType  = 214
	pitDoorUnitType        = 205
	neutralOwner           = 11
	disabledThingyFlags    = 0x8000
	chkUnitEntrySize       = 36
)

type section struct{ start, end int }

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintln(os.Stderr, "usage: thg2inject in.rep out.rep")
		os.Exit(2)
	}
	doors, err := run(os.Args[1], os.Args[2])
	if err != nil {
		fmt.Fprintln(os.Stderr, "thg2inject:", err)
		os.Exit(1)
	}
	fmt.Printf("%s disabled_doors=%d\n", os.Args[2], doors)
}

func run(in, out string) (int, error) {
	data, err := os.ReadFile(in)
	if err != nil {
		return 0, err
	}
	if len(data) < 30 || data[12] != 's' {
		return 0, errors.New("not a 1.21 SC:R replay")
	}
	sections, err := readSections(data)
	if err != nil {
		return 0, err
	}
	chk, err := sectionData(data, sections[mapSection])
	if err != nil {
		return 0, err
	}
	doors := doorPositions(chunks(chk))
	newChk := appendChunk(chk, "THG2", disabledDoors(doors))
	result := append([]byte{}, data[:sections[mapLengthSection].start]...)
	result = append(result, encodeSection(binary.LittleEndian.AppendUint32(nil, uint32(len(newChk))))...)
	result = append(result, encodeSection(newChk)...)
	result = append(result, data[sections[mapSection].end:]...)
	encodedLength := int(binary.LittleEndian.Uint32(data[encodedLengthOffset:]))
	binary.LittleEndian.PutUint32(result[encodedLengthOffset:], uint32(encodedLength+len(result)-len(data)))
	return len(doors), os.WriteFile(out, result, 0o644)
}

func readSections(data []byte) ([]section, error) {
	var sections []section
	pos := 0
	for len(sections) != sectionsBeforeTrailing {
		if len(sections) == 1 {
			pos += 4
		}
		if pos+8 > len(data) {
			return nil, errors.New("truncated replay")
		}
		start := pos
		count := int(binary.LittleEndian.Uint32(data[pos+4:]))
		pos += 8
		for range count {
			if pos+4 > len(data) {
				return nil, errors.New("truncated replay")
			}
			pos += 4 + int(binary.LittleEndian.Uint32(data[pos:]))
		}
		sections = append(sections, section{start, pos})
	}
	return sections, nil
}

func sectionData(data []byte, s section) ([]byte, error) {
	pos := s.start + 4
	count := int(binary.LittleEndian.Uint32(data[pos:]))
	pos += 4
	var out bytes.Buffer
	for range count {
		n := int(binary.LittleEndian.Uint32(data[pos:]))
		chunk := data[pos+4 : pos+4+n]
		pos += 4 + n
		if !isZlib(chunk) {
			out.Write(chunk)
			continue
		}
		zr, err := zlib.NewReader(bytes.NewReader(chunk))
		if err != nil {
			return nil, err
		}
		if _, err := io.Copy(&out, zr); err != nil {
			return nil, err
		}
	}
	return out.Bytes(), nil
}

func isZlib(chunk []byte) bool {
	return len(chunk) > 4 && chunk[0] == 0x78 && (uint16(chunk[0])<<8|uint16(chunk[1]))%31 == 0
}

func encodeSection(payload []byte) []byte {
	var compressed bytes.Buffer
	w := zlib.NewWriter(&compressed)
	w.Write(payload)
	w.Close()
	out := binary.LittleEndian.AppendUint32(nil, 0)
	out = binary.LittleEndian.AppendUint32(out, 1)
	out = binary.LittleEndian.AppendUint32(out, uint32(compressed.Len()))
	return append(out, compressed.Bytes()...)
}

func chunks(chk []byte) map[string][][]byte {
	m := map[string][][]byte{}
	for pos := 0; pos+8 <= len(chk); {
		tag := string(chk[pos : pos+4])
		n := int(int32(binary.LittleEndian.Uint32(chk[pos+4:])))
		pos += 8
		if n < 0 || pos+n > len(chk) {
			break
		}
		m[tag] = append(m[tag], chk[pos:pos+n])
		pos += n
	}
	return m
}

func doorPositions(chunks map[string][][]byte) [][2]int {
	dim := chunks["DIM "][0]
	cx := int(binary.LittleEndian.Uint16(dim)) * 16
	cy := int(binary.LittleEndian.Uint16(dim[2:])) * 16
	positions := [][2]int{{cx, cy}}
	for _, units := range chunks["UNIT"] {
		for i := 0; i+chkUnitEntrySize <= len(units); i += chkUnitEntrySize {
			if binary.LittleEndian.Uint16(units[i+8:]) != startLocationUnitType {
				continue
			}
			x := int(binary.LittleEndian.Uint16(units[i+4:]))
			y := int(binary.LittleEndian.Uint16(units[i+6:]))
			positions = append(positions, [2]int{(x + cx) / 2, (y + cy) / 2}, [2]int{(3*x + cx) / 4, (3*y + cy) / 4})
		}
	}
	return positions
}

func disabledDoors(positions [][2]int) []byte {
	var thg2 []byte
	for i, p := range positions {
		thg2 = binary.LittleEndian.AppendUint16(thg2, uint16(pitDoorUnitType+i%2))
		thg2 = binary.LittleEndian.AppendUint16(thg2, uint16(p[0]))
		thg2 = binary.LittleEndian.AppendUint16(thg2, uint16(p[1]))
		thg2 = append(thg2, neutralOwner, 0)
		thg2 = binary.LittleEndian.AppendUint16(thg2, disabledThingyFlags)
	}
	return thg2
}

func appendChunk(chk []byte, tag string, payload []byte) []byte {
	out := append(append([]byte{}, chk...), tag...)
	out = binary.LittleEndian.AppendUint32(out, uint32(len(payload)))
	return append(out, payload...)
}
