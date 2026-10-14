package downgrade

import (
	"encoding/binary"
	"hash/crc32"
)

const segmentSize = 8192

func openBWCRC32(data []byte) uint32 {
	return ^crc32.ChecksumIEEE(data)
}

func appendSection(out, data []byte) []byte {
	segments := (len(data) + segmentSize - 1) / segmentSize
	out = binary.LittleEndian.AppendUint32(out, openBWCRC32(data))
	out = binary.LittleEndian.AppendUint32(out, uint32(segments))
	for start := 0; start < len(data); start += segmentSize {
		segment := data[start:min(start+segmentSize, len(data))]
		out = binary.LittleEndian.AppendUint32(out, uint32(len(segment)))
		out = append(out, segment...)
	}
	return out
}

func appendUint32Section(out []byte, v uint32) []byte {
	return appendSection(out, binary.LittleEndian.AppendUint32(nil, v))
}

func encodeLegacyReplay(gameInfo, actions, mapData []byte) []byte {
	size := len(gameInfo) + len(actions) + len(mapData) + 256
	out := make([]byte, 0, size)
	out = appendUint32Section(out, legacyReplayID)
	out = appendSection(out, gameInfo)
	out = appendUint32Section(out, uint32(len(actions)))
	out = appendSection(out, actions)
	out = appendUint32Section(out, uint32(len(mapData)))
	return appendSection(out, mapData)
}
