package downgrade

import (
	"encoding/binary"
	"fmt"
)

const (
	maxSelectionCount        = 12
	frameHeaderSize          = 5
	controlGroupCount        = 10
	maxControlGroupSubaction = 2
)

const (
	idSelect           byte = 0x09
	idSelectAdd        byte = 0x0a
	idSelectRemove     byte = 0x0b
	idRightClick       byte = 0x14
	idTargetedOrder    byte = 0x15
	idControlGroup     byte = 0x13
	idSaveGame         byte = 0x06
	idLoadGame         byte = 0x07
	idUnload           byte = 0x29
	idRightClick121    byte = 0x60
	idTargetedOrder121 byte = 0x61
	idUnload121        byte = 0x62
	idSelect121        byte = 0x63
	idSelectAdd121     byte = 0x64
	idSelectRemove121  byte = 0x65
)

const (
	ReasonUnsupported        = "not read by OpenBW"
	ReasonUnknownLength      = "unknown command, rest of frame block skipped"
	ReasonTruncated          = "truncated command"
	ReasonUnknownPlayer      = "player id not in game info slots"
	ReasonTooManySelected    = "selects more than 12 units"
	ReasonUnknownGroupAction = "unknown control group subaction"
)

var fixedPayloadSizes = map[byte]int{
	0x05: 0, 0x08: 0, 0x0c: 7, 0x0d: 2, 0x0e: 4, 0x0f: 1, 0x10: 0, 0x11: 0,
	0x12: 4, 0x13: 2, 0x14: 9, 0x15: 10, 0x18: 0, 0x19: 0, 0x1a: 1, 0x1b: 0,
	0x1c: 0, 0x1d: 0, 0x1e: 1, 0x1f: 2, 0x20: 2, 0x21: 1, 0x22: 1, 0x23: 2,
	0x25: 1, 0x26: 1, 0x27: 0, 0x28: 1, 0x29: 2, 0x2a: 0, 0x2b: 1, 0x2c: 1,
	0x2d: 1, 0x2e: 0, 0x2f: 4, 0x30: 1, 0x31: 0, 0x32: 1, 0x33: 0, 0x34: 0,
	0x35: 2, 0x36: 0, 0x37: 6, 0x38: 0, 0x39: 0, 0x3a: 1, 0x3b: 1, 0x3c: 0,
	0x3d: 1, 0x3e: 5, 0x3f: 7, 0x40: 17, 0x41: 2, 0x42: 1, 0x43: 1, 0x44: 2,
	0x45: 2, 0x48: 12, 0x54: 0, 0x55: 1, 0x56: 9, 0x57: 1, 0x58: 4, 0x5a: 0,
	0x5b: 0, 0x5c: 81, 0x60: 11, 0x61: 12, 0x62: 4,
}

var openBWReadableIDs = map[byte]bool{
	0x09: true, 0x0a: true, 0x0b: true, 0x0c: true, 0x0d: true, 0x0e: true,
	0x12: true, 0x13: true, 0x14: true, 0x15: true, 0x18: true, 0x19: true,
	0x1a: true, 0x1b: true, 0x1c: true, 0x1e: true, 0x1f: true, 0x20: true,
	0x21: true, 0x22: true, 0x23: true, 0x25: true, 0x26: true, 0x27: true,
	0x28: true, 0x29: true, 0x2a: true, 0x2b: true, 0x2c: true, 0x2d: true,
	0x2e: true, 0x2f: true, 0x30: true, 0x31: true, 0x32: true, 0x33: true,
	0x34: true, 0x35: true, 0x36: true, 0x57: true, 0x58: true, 0x5a: true,
	0x5c: true,
}

var legacyIDOf121 = map[byte]byte{
	idRightClick121:    idRightClick,
	idTargetedOrder121: idTargetedOrder,
	idUnload121:        idUnload,
	idSelect121:        idSelect,
	idSelectAdd121:     idSelectAdd,
	idSelectRemove121:  idSelectRemove,
}

type DroppedCommand struct {
	ID       byte
	Frame    uint32
	PlayerID byte
	Count    int
	Reason   string
}

func isSelection(id byte) bool {
	return id == idSelect || id == idSelectAdd || id == idSelectRemove
}

func isSelection121(id byte) bool {
	return id == idSelect121 || id == idSelectAdd121 || id == idSelectRemove121
}

func commandPayloadSize(id byte, payload []byte) (int, bool) {
	switch {
	case isSelection(id):
		return selectionPayloadSize(payload, 2)
	case isSelection121(id):
		return selectionPayloadSize(payload, 4)
	case id == idSaveGame || id == idLoadGame:
		return saveGamePayloadSize(payload)
	}
	size, known := fixedPayloadSizes[id]
	return size, known
}

func selectionPayloadSize(payload []byte, tagSize int) (int, bool) {
	if len(payload) < 1 {
		return 1, true
	}
	return 1 + int(payload[0])*tagSize, true
}

func saveGamePayloadSize(payload []byte) (int, bool) {
	if len(payload) < 4 {
		return 4, true
	}
	return 4 + int(binary.LittleEndian.Uint32(payload)), true
}

type command struct {
	frame    uint32
	playerID byte
	id       byte
	payload  []byte
}

type SourceCommand struct {
	Frame      uint32
	PlayerID   byte
	ID         byte
	Position   int
	Action     int32
	DropReason string
}

type commandTranslator struct {
	validPlayerIDs map[byte]bool
	dropped        []DroppedCommand
	commands       []SourceCommand
	written        int32
	framePositions map[uint32]int
}

func newCommandIndexer() *commandTranslator {
	return &commandTranslator{framePositions: map[uint32]int{}}
}

func newCommandTranslator(slotPlayerIDs [slotCount]uint32) *commandTranslator {
	t := newCommandIndexer()
	t.validPlayerIDs = map[byte]bool{}
	for _, id := range slotPlayerIDs {
		if id <= 0xff {
			t.validPlayerIDs[byte(id)] = true
		}
	}
	return t
}

func (t *commandTranslator) translateActions(actions []byte) ([]byte, error) {
	out := make([]byte, 0, len(actions))
	for pos := 0; pos < len(actions); {
		if len(actions)-pos < frameHeaderSize {
			return nil, fmt.Errorf("truncated frame header at offset %d", pos)
		}
		frame := binary.LittleEndian.Uint32(actions[pos:])
		blockSize := int(actions[pos+4])
		blockStart := pos + frameHeaderSize
		blockEnd := blockStart + blockSize
		if blockEnd > len(actions) {
			return nil, fmt.Errorf("frame %d block ends past the actions section", frame)
		}
		block := t.translateBlock(frame, actions[blockStart:blockEnd])
		out = appendFrameBlock(out, frame, block)
		pos = blockEnd
	}
	return out, nil
}

func appendFrameBlock(out []byte, frame uint32, block []byte) []byte {
	if len(block) == 0 {
		return out
	}
	out = binary.LittleEndian.AppendUint32(out, frame)
	out = append(out, byte(len(block)))
	return append(out, block...)
}

func (t *commandTranslator) translateBlock(frame uint32, block []byte) []byte {
	var out []byte
	for pos := 0; pos < len(block); {
		cmd, next, ok := t.splitCommand(frame, block, pos)
		if !ok {
			break
		}
		pos = next
		out = t.appendTranslated(out, cmd)
	}
	return out
}

func (t *commandTranslator) splitCommand(frame uint32, block []byte, pos int) (command, int, bool) {
	if len(block)-pos < 2 {
		t.drop(frame, block[pos], 0, ReasonTruncated)
		return command{}, 0, false
	}
	cmd := command{frame: frame, playerID: block[pos], id: block[pos+1]}
	payloadStart := pos + 2
	size, known := commandPayloadSize(cmd.id, block[payloadStart:])
	if !known {
		t.drop(frame, cmd.playerID, cmd.id, ReasonUnknownLength)
		return command{}, 0, false
	}
	if payloadStart+size > len(block) {
		t.drop(frame, cmd.playerID, cmd.id, ReasonTruncated)
		return command{}, 0, false
	}
	cmd.payload = block[payloadStart : payloadStart+size]
	return cmd, payloadStart + size, true
}

func (t *commandTranslator) appendTranslated(out []byte, original command) []byte {
	cmd := toLegacyCommand(original)
	if reason := t.dropReason(cmd); reason != "" {
		t.drop(original.frame, original.playerID, original.id, reason)
		return out
	}
	t.recordWritten(original.frame, original.playerID, original.id)
	out = append(out, cmd.playerID, cmd.id)
	return append(out, cmd.payload...)
}

func (t *commandTranslator) recordWritten(frame uint32, playerID, id byte) {
	t.record(frame, playerID, id, t.written, "")
	t.written++
}

func (t *commandTranslator) record(frame uint32, playerID, id byte, action int32, dropReason string) {
	position := t.framePositions[frame]
	t.framePositions[frame] = position + 1
	t.commands = append(t.commands, SourceCommand{Frame: frame, PlayerID: playerID, ID: id, Position: position, Action: action, DropReason: dropReason})
}

func indexLegacyActions(actions []byte) []SourceCommand {
	t := newCommandIndexer()
	for pos := 0; len(actions)-pos >= frameHeaderSize; {
		frame := binary.LittleEndian.Uint32(actions[pos:])
		blockStart := pos + frameHeaderSize
		blockEnd := blockStart + int(actions[pos+4])
		if blockEnd > len(actions) {
			break
		}
		if !t.indexLegacyBlock(frame, actions[blockStart:blockEnd]) {
			break
		}
		pos = blockEnd
	}
	return t.commands
}

func (t *commandTranslator) indexLegacyBlock(frame uint32, block []byte) bool {
	for pos := 0; pos < len(block); {
		if len(block)-pos < 2 {
			return false
		}
		playerID, id := block[pos], block[pos+1]
		size, known := commandPayloadSize(id, block[pos+2:])
		if !known || pos+2+size > len(block) {
			return false
		}
		t.recordWritten(frame, playerID, id)
		pos += 2 + size
	}
	return true
}

func toLegacyCommand(cmd command) command {
	legacyID, is121 := legacyIDOf121[cmd.id]
	if !is121 {
		return cmd
	}
	return command{frame: cmd.frame, playerID: cmd.playerID, id: legacyID, payload: narrowUnitTags(cmd.id, cmd.payload)}
}

func (t *commandTranslator) dropReason(cmd command) string {
	switch {
	case !openBWReadableIDs[cmd.id]:
		return ReasonUnsupported
	case !t.validPlayerIDs[cmd.playerID]:
		return ReasonUnknownPlayer
	case isSelection(cmd.id) && int(cmd.payload[0]) > maxSelectionCount:
		return ReasonTooManySelected
	case cmd.id == idControlGroup && !isKnownControlGroupAction(cmd.payload):
		return ReasonUnknownGroupAction
	}
	return ""
}

func isKnownControlGroupAction(payload []byte) bool {
	subaction, group := payload[0], payload[1]
	return group >= controlGroupCount || subaction <= maxControlGroupSubaction
}

const (
	unitTag121Size        = 4
	unitTagSize           = 2
	unitIndex121Bits      = 13
	unitIndex121Mask      = 1<<unitIndex121Bits - 1
	scrUnitCapacity       = 3400
	openBWUnitCapacity    = 1700
	unitGenerationShift   = 11
	unitGenerationCount   = 32
	firstAllocatedUnitTag = 1
)

func narrowUnitTags(id121 byte, payload []byte) []byte {
	switch id121 {
	case idRightClick121, idTargetedOrder121:
		return concat(payload[:4], appendLegacyUnitTag(nil, payload[4:8]), payload[8:])
	case idUnload121:
		return appendLegacyUnitTag(nil, payload[:4])
	}
	return narrowSelectionTags(payload)
}

func narrowSelectionTags(payload []byte) []byte {
	count := int(payload[0])
	out := make([]byte, 0, 1+count*unitTagSize)
	out = append(out, payload[0])
	for i := range count {
		out = appendLegacyUnitTag(out, payload[1+i*unitTag121Size:])
	}
	return out
}

func appendLegacyUnitTag(out, tag121 []byte) []byte {
	return binary.LittleEndian.AppendUint16(out, legacyUnitTag(binary.LittleEndian.Uint32(tag121)))
}

func legacyUnitTag(tag121 uint32) uint16 {
	index, representable := legacyUnitIndex(tag121 & unitIndex121Mask)
	if !representable {
		return 0
	}
	generation := (tag121 >> unitIndex121Bits) % unitGenerationCount
	return uint16(index | generation<<unitGenerationShift)
}

func legacyUnitIndex(index121 uint32) (uint32, bool) {
	switch {
	case index121 == 0:
		return 0, false
	case index121 == firstAllocatedUnitTag:
		return index121, true
	case index121 > openBWUnitCapacity && index121 <= scrUnitCapacity:
		return index121 - openBWUnitCapacity, true
	}
	return 0, false
}

func concat(parts ...[]byte) []byte {
	var out []byte
	for _, part := range parts {
		out = append(out, part...)
	}
	return out
}

func (t *commandTranslator) drop(frame uint32, playerID, id byte, reason string) {
	t.record(frame, playerID, id, -1, reason)
	if n := len(t.dropped); n > 0 {
		last := &t.dropped[n-1]
		if last.Frame == frame && last.ID == id && last.PlayerID == playerID && last.Reason == reason {
			last.Count++
			return
		}
	}
	t.dropped = append(t.dropped, DroppedCommand{ID: id, Frame: frame, PlayerID: playerID, Count: 1, Reason: reason})
}
