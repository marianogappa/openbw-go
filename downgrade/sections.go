package downgrade

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"errors"
	"fmt"
	"io"

	"github.com/marianogappa/openbw-go/internal/pkware"
)

const (
	replayIDSize          = 4
	gameInfoSize          = 633
	playerNamesSize       = 0x300
	sectionIDCCLR         = 0x524c4343
	sectionIDLMTS         = 0x53544d4c
	sectionIDBFIX         = 0x58494642
	sectionIDGCFG         = 0x47464347
	sectionIDSKIN         = 0x4e494b53
	legacyReplayID        = 0x53526572
	modernReplayID        = 0x53526573
	formatHeaderSize      = 30
	formatIDOffset        = 12
	formatFirstDataOffset = 28
	zlibMagic             = 0x78
	zlibHeaderCheck       = 31
	maxSectionSize        = 1 << 30
)

type replayFormat uint8

const (
	formatUnknown replayFormat = iota
	formatLegacy
	formatModern
	formatModern121
)

func (f replayFormat) String() string {
	switch f {
	case formatLegacy:
		return "legacy"
	case formatModern:
		return "1.18-1.20"
	case formatModern121:
		return "1.21+"
	}
	return "unknown"
}

// SourceFormat names the format of a replay: "legacy", "1.18-1.20", "1.21+" or "unknown".
func SourceFormat(replay []byte) string { return detectFormat(replay).String() }

func detectFormat(data []byte) replayFormat {
	switch {
	case len(data) < formatHeaderSize:
		return formatUnknown
	case data[formatIDOffset] == 's':
		return formatModern121
	case data[formatFirstDataOffset] != zlibMagic:
		return formatLegacy
	}
	return formatModern
}

var trailingSectionSizes = map[uint32]int{
	sectionIDSKIN: 0x15e0,
	sectionIDLMTS: 0x1c,
	sectionIDBFIX: 0x08,
	sectionIDCCLR: 0xc0,
	sectionIDGCFG: 0x19,
}

type sourceSections struct {
	format       replayFormat
	gameInfo     []byte
	actions      []byte
	mapData      []byte
	playerColors []byte
	limits       []byte
	bugFixes     []byte
	gameConfig   []byte
}

type sectionReader struct {
	data   []byte
	pos    int
	format replayFormat
}

func (r *sectionReader) uint32() (uint32, error) {
	b, err := r.raw(4)
	if err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint32(b), nil
}

func (r *sectionReader) raw(n int) ([]byte, error) {
	if n < 0 || len(r.data)-r.pos < n {
		return nil, io.ErrUnexpectedEOF
	}
	b := r.data[r.pos : r.pos+n]
	r.pos += n
	return b, nil
}

func (r *sectionReader) section(size int) ([]byte, error) {
	if size == 0 {
		return []byte{}, nil
	}
	if _, err := r.uint32(); err != nil {
		return nil, err
	}
	chunks, err := r.uint32()
	if err != nil {
		return nil, err
	}
	if r.format == formatLegacy {
		return r.legacyChunks(int(chunks), size)
	}
	return r.modernChunks(int(chunks), size)
}

func (r *sectionReader) legacyChunks(chunks, size int) ([]byte, error) {
	out := make([]byte, size)
	pos := 0
	for range chunks {
		length, err := r.uint32()
		if err != nil {
			return nil, err
		}
		chunkSize := min(size-pos, segmentSize)
		if uint64(length) > uint64(chunkSize) {
			return nil, errors.New("mismatched section")
		}
		chunk, err := r.raw(int(length))
		if err != nil {
			return nil, err
		}
		if int(length) == chunkSize {
			copy(out[pos:], chunk)
		} else if err := pkware.Decompress(chunk, out[pos:pos+chunkSize]); err != nil {
			return nil, err
		}
		pos += chunkSize
	}
	return out[:pos], nil
}

func isZlibChunk(chunk []byte) bool {
	return len(chunk) > 4 && chunk[0] == zlibMagic && (uint16(chunk[0])<<8|uint16(chunk[1]))%zlibHeaderCheck == 0
}

func (r *sectionReader) modernChunks(chunks, size int) ([]byte, error) {
	out := bytes.NewBuffer(make([]byte, 0, size))
	for range chunks {
		length, err := r.uint32()
		if err != nil {
			return nil, err
		}
		chunk, err := r.raw(int(length))
		if err != nil {
			return nil, err
		}
		if !isZlibChunk(chunk) {
			out.Write(chunk)
			continue
		}
		zr, err := zlib.NewReader(bytes.NewReader(chunk))
		if err != nil {
			return nil, err
		}
		_, err = io.Copy(out, zr)
		zr.Close()
		if err != nil {
			return nil, err
		}
	}
	return out.Bytes(), nil
}

func (r *sectionReader) fixedSection(size int) ([]byte, error) {
	data, err := r.section(size)
	if err != nil {
		return nil, err
	}
	if len(data) != size {
		return nil, fmt.Errorf("got %d bytes, expected %d", len(data), size)
	}
	return data, nil
}

func (r *sectionReader) sizedSection() ([]byte, error) {
	sizeData, err := r.fixedSection(4)
	if err != nil {
		return nil, fmt.Errorf("size: %w", err)
	}
	size := binary.LittleEndian.Uint32(sizeData)
	if size > maxSectionSize {
		return nil, fmt.Errorf("section size %d too large", size)
	}
	return r.fixedSection(int(size))
}

func (r *sectionReader) trailingSection() (id uint32, data []byte, err error) {
	if id, err = r.uint32(); err != nil {
		return 0, nil, err
	}
	rawSize, err := r.uint32()
	if err != nil {
		return 0, nil, err
	}
	size, known := trailingSectionSizes[id]
	if !known {
		data, err = r.raw(int(rawSize))
		return id, data, err
	}
	data, err = r.section(size)
	return id, data, err
}

func readSourceSections(data []byte) (s sourceSections, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("decode replay: %v", r)
		}
	}()
	if legacy, err := ReadUncompressedLegacy(data); err == nil {
		return sourceSections{format: formatLegacy, gameInfo: legacy.GameInfo, actions: legacy.Actions, mapData: legacy.MapData}, nil
	}
	s.format = detectFormat(data)
	if s.format == formatUnknown {
		return s, errors.New("unknown replay format")
	}
	r := &sectionReader{data: data, format: s.format}
	if err = readReplayID(r); err != nil {
		return s, err
	}
	if s.format == formatModern121 {
		if _, err = r.uint32(); err != nil {
			return s, fmt.Errorf("game info: %w", err)
		}
	}
	if s.gameInfo, err = r.fixedSection(gameInfoSize); err != nil {
		return s, fmt.Errorf("game info: %w", err)
	}
	if s.actions, err = r.sizedSection(); err != nil {
		return s, fmt.Errorf("actions: %w", err)
	}
	if s.mapData, err = r.sizedSection(); err != nil {
		return s, fmt.Errorf("map data: %w", err)
	}
	s.readTrailingSections(r)
	return s, nil
}

func readReplayID(r *sectionReader) error {
	id, err := r.fixedSection(replayIDSize)
	if err != nil {
		return fmt.Errorf("replay id: %w", err)
	}
	switch binary.LittleEndian.Uint32(id) {
	case legacyReplayID, modernReplayID:
		return nil
	}
	return fmt.Errorf("invalid replay id %q", id)
}

func (s *sourceSections) readTrailingSections(r *sectionReader) {
	if s.format == formatLegacy {
		return
	}
	if _, err := r.section(playerNamesSize); err != nil {
		return
	}
	for r.pos < len(r.data) {
		id, data, err := r.trailingSection()
		if err != nil {
			return
		}
		s.keepTrailingSection(id, data)
	}
}

func (s *sourceSections) keepTrailingSection(sectionID uint32, data []byte) {
	switch sectionID {
	case sectionIDCCLR:
		s.playerColors = data
	case sectionIDLMTS:
		s.limits = data
	case sectionIDBFIX:
		s.bugFixes = data
	case sectionIDGCFG:
		s.gameConfig = data
	}
}
