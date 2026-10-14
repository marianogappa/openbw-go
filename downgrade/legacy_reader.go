package downgrade

import (
	"encoding/binary"
	"errors"
	"fmt"
)

var errCompressedSegment = errors.New("compressed segment")

type LegacyReplay struct {
	GameInfo []byte
	Actions  []byte
	MapData  []byte
}

type replayFileReader struct {
	data []byte
	pos  int
}

func (r *replayFileReader) getUint32() (uint32, error) {
	if len(r.data)-r.pos < 4 {
		return 0, fmt.Errorf("unexpected end of file at offset %d", r.pos)
	}
	v := binary.LittleEndian.Uint32(r.data[r.pos:])
	r.pos += 4
	return v, nil
}

func (r *replayFileReader) getRaw(n int) ([]byte, error) {
	if len(r.data)-r.pos < n {
		return nil, fmt.Errorf("unexpected end of file at offset %d", r.pos)
	}
	b := r.data[r.pos : r.pos+n]
	r.pos += n
	return b, nil
}

func (r *replayFileReader) getBytes(outputSize int) ([]byte, error) {
	crc, err := r.getUint32()
	if err != nil {
		return nil, err
	}
	segments, err := r.getUint32()
	if err != nil {
		return nil, err
	}
	output := make([]byte, 0, outputSize)
	for range segments {
		segment, err := r.getSegment(outputSize - len(output))
		if err != nil {
			return nil, err
		}
		output = append(output, segment...)
	}
	if len(output) != outputSize {
		return nil, fmt.Errorf("read %d bytes, expected %d", len(output), outputSize)
	}
	if got := openBWCRC32(output); got != crc {
		return nil, fmt.Errorf("crc32 mismatch: got %08x, expected %08x", got, crc)
	}
	return output, nil
}

func (r *replayFileReader) getSegment(remaining int) ([]byte, error) {
	inputSize, err := r.getUint32()
	if err != nil {
		return nil, err
	}
	outputSize := min(remaining, segmentSize)
	if uint64(inputSize) > uint64(outputSize) {
		return nil, errors.New("output buffer too small")
	}
	if int(inputSize) != outputSize {
		return nil, errCompressedSegment
	}
	return r.getRaw(outputSize)
}

func (r *replayFileReader) getUint32Section() (uint32, error) {
	b, err := r.getBytes(4)
	if err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint32(b), nil
}

func (r *replayFileReader) getSizedSection() ([]byte, error) {
	size, err := r.getUint32Section()
	if err != nil {
		return nil, err
	}
	if uint64(size) > uint64(len(r.data)) {
		return nil, fmt.Errorf("section size %d larger than the file", size)
	}
	return r.getBytes(int(size))
}

func ReadUncompressedLegacy(data []byte) (LegacyReplay, error) {
	r := &replayFileReader{data: data}
	var rep LegacyReplay
	identifier, err := r.getUint32Section()
	if err != nil {
		return rep, fmt.Errorf("identifier: %w", err)
	}
	if identifier != legacyReplayID {
		return rep, fmt.Errorf("invalid identifier %#x", identifier)
	}
	if rep.GameInfo, err = r.getBytes(gameInfoSize); err != nil {
		return rep, fmt.Errorf("game info: %w", err)
	}
	if rep.Actions, err = r.getSizedSection(); err != nil {
		return rep, fmt.Errorf("actions: %w", err)
	}
	if rep.MapData, err = r.getSizedSection(); err != nil {
		return rep, fmt.Errorf("map data: %w", err)
	}
	return rep, nil
}
