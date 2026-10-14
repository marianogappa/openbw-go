package downgrade

import (
	"encoding/binary"
	"fmt"
)

const (
	chunkHeaderSize   = 8
	maxLegacyStrSize  = 0xffff
	defaultColorCount = 8
)

var legacyMapVersionOf = map[uint16]uint16{
	64:  63,
	206: 205,
}

type chunk struct {
	tag        string
	dataOffset int
	data       []byte
}

func splitChunks(chk []byte) []chunk {
	var chunks []chunk
	for pos := 0; len(chk)-pos >= chunkHeaderSize; {
		tag := string(chk[pos : pos+4])
		size := binary.LittleEndian.Uint32(chk[pos+4:])
		dataOffset := pos + chunkHeaderSize
		if uint64(size) > uint64(len(chk)-dataOffset) {
			break
		}
		end := dataOffset + int(size)
		chunks = append(chunks, chunk{tag: tag, dataOffset: dataOffset, data: chk[dataOffset:end]})
		pos = end
	}
	return chunks
}

func findChunk(chunks []chunk, tag string) (chunk, bool) {
	for _, c := range chunks {
		if c.tag == tag {
			return c, true
		}
	}
	return chunk{}, false
}

func downgradeMap(chk []byte) ([]byte, []string) {
	out := append([]byte(nil), chk...)
	chunks := splitChunks(out)
	changes, version := clampVersions(chunks)
	var prefix []byte
	if added, change := legacyStringsChunk(chunks); added != nil {
		prefix = append(prefix, added...)
		changes = append(changes, change)
	}
	if added, change := defaultColorsChunk(chunks, version); added != nil {
		prefix = append(prefix, added...)
		changes = append(changes, change)
	}
	return append(prefix, out...), changes
}

func clampVersions(chunks []chunk) ([]string, uint16) {
	var changes []string
	var version uint16
	for _, c := range chunks {
		if c.tag != "VER " || len(c.data) < 2 {
			continue
		}
		version = binary.LittleEndian.Uint16(c.data)
		legacy, found := legacyMapVersionOf[version]
		if !found {
			continue
		}
		binary.LittleEndian.PutUint16(c.data, legacy)
		changes = append(changes, fmt.Sprintf("map VER %d -> %d", version, legacy))
		version = legacy
	}
	return changes, version
}

func defaultColorsChunk(chunks []chunk, version uint16) ([]byte, string) {
	if version != 205 {
		return nil, ""
	}
	if _, found := findChunk(chunks, "COLR"); found {
		return nil, ""
	}
	colors := make([]byte, defaultColorCount)
	for i := range colors {
		colors[i] = byte(i)
	}
	return encodeChunk("COLR", colors), "map COLR added with default colors"
}

func legacyStringsChunk(chunks []chunk) ([]byte, string) {
	if _, found := findChunk(chunks, "STR "); found {
		return nil, ""
	}
	strx, found := findChunk(chunks, "STRx")
	if !found {
		return nil, ""
	}
	strs := decodeExtendedStrings(strx.data)
	data, dropped := encodeLegacyStrings(strs)
	change := fmt.Sprintf("map STR added from STRx (%d strings)", len(strs))
	if dropped > 0 {
		change += fmt.Sprintf(", %d emptied to fit 64 KiB", dropped)
	}
	return encodeChunk("STR ", data), change
}

func encodeChunk(tag string, data []byte) []byte {
	out := make([]byte, 0, chunkHeaderSize+len(data))
	out = append(out, tag...)
	out = binary.LittleEndian.AppendUint32(out, uint32(len(data)))
	return append(out, data...)
}

func decodeExtendedStrings(data []byte) [][]byte {
	if len(data) < 4 {
		return nil
	}
	count := int(binary.LittleEndian.Uint32(data))
	if count > (len(data)-4)/4 {
		count = (len(data) - 4) / 4
	}
	strs := make([][]byte, count)
	for i := range strs {
		offset := binary.LittleEndian.Uint32(data[4+i*4:])
		strs[i] = cString(data, offset)
	}
	return strs
}

func cString(data []byte, offset uint32) []byte {
	if uint64(offset) >= uint64(len(data)) {
		return nil
	}
	rest := data[offset:]
	for i, b := range rest {
		if b == 0 {
			return rest[:i]
		}
	}
	return rest
}

func encodeLegacyStrings(strs [][]byte) ([]byte, int) {
	count := min(len(strs), (maxLegacyStrSize-2-1)/2)
	tableEnd := 2 + count*2
	emptyOffset := tableEnd
	data := make([]byte, tableEnd+1)
	binary.LittleEndian.PutUint16(data, uint16(count))
	dropped := len(strs) - count
	for i := range count {
		offset := emptyOffset
		if len(strs[i]) > 0 && len(data)+len(strs[i])+1 <= maxLegacyStrSize {
			offset = len(data)
			data = append(data, strs[i]...)
			data = append(data, 0)
		} else if len(strs[i]) > 0 {
			dropped++
		}
		binary.LittleEndian.PutUint16(data[2+i*2:], uint16(offset))
	}
	return data, dropped
}
