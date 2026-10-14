package downgrade

import (
	"encoding/binary"
	"fmt"
	"slices"
)

type parsedAction struct {
	frame    uint32
	playerID byte
	id       byte
	fields   string
}

type actionReader struct {
	data []byte
	pos  int
}

func (r *actionReader) skip(n int) error {
	if len(r.data)-r.pos < n {
		return fmt.Errorf("read past end: need %d bytes at offset %d of %d", n, r.pos, len(r.data))
	}
	r.pos += n
	return nil
}

func (r *actionReader) getUint8() (byte, error) {
	if err := r.skip(1); err != nil {
		return 0, err
	}
	return r.data[r.pos-1], nil
}

var openBWActionPayloadSizes = map[byte]int{
	12: 7, 13: 2, 14: 4, 18: 4, 19: 2, 20: 9, 21: 10, 24: 0, 25: 0, 26: 1,
	27: 0, 28: 0, 30: 1, 31: 2, 32: 2, 33: 1, 34: 1, 35: 2, 37: 1, 38: 1,
	39: 0, 40: 1, 41: 2, 42: 0, 43: 1, 44: 1, 45: 1, 46: 0, 47: 4, 48: 1,
	49: 0, 50: 1, 51: 0, 52: 0, 53: 2, 54: 0, 87: 1, 88: 4, 90: 0, 92: 81,
}

func readOpenBWSelection(r *actionReader) error {
	n, err := r.getUint8()
	if err != nil {
		return err
	}
	if n > 12 {
		return fmt.Errorf("invalid selection of %d units", n)
	}
	return r.skip(int(n) * 2)
}

func readOpenBWAction(r *actionReader, id byte) error {
	switch id {
	case 9, 10, 11:
		return readOpenBWSelection(r)
	}
	size, known := openBWActionPayloadSizes[id]
	if !known {
		return fmt.Errorf("unknown action %d", id)
	}
	if err := r.skip(size); err != nil {
		return err
	}
	if id == 19 {
		return checkOpenBWControlGroup(r.data[r.pos-size : r.pos])
	}
	return nil
}

func checkOpenBWControlGroup(payload []byte) error {
	subaction, group := payload[0], payload[1]
	if group < 10 && subaction > 2 {
		return fmt.Errorf("action_control_group: unknown subaction %d", subaction)
	}
	return nil
}

func u16(b []byte) uint16 {
	return binary.LittleEndian.Uint16(b)
}

func unitIndex(b []byte) uint16 {
	return u16(b) & 0x7ff
}

func openBWActionFields(id byte, payload []byte) string {
	switch id {
	case 9, 10, 11:
		indexes := []uint16{}
		for i := range int(payload[0]) {
			indexes = append(indexes, unitIndex(payload[1+i*2:]))
		}
		return fmt.Sprint(indexes)
	case 20:
		return fmt.Sprint(u16(payload), u16(payload[2:]), unitIndex(payload[4:]), u16(payload[6:]), payload[8] != 0)
	case 21:
		return fmt.Sprint(u16(payload), u16(payload[2:]), unitIndex(payload[4:]), u16(payload[6:]), payload[8], payload[9] != 0)
	case 41:
		return fmt.Sprint(unitIndex(payload))
	}
	return ""
}

func parseOpenBWActions(actions []byte, slotIDs [slotCount]uint32) ([]parsedAction, error) {
	var parsed []parsedAction
	var lastFrame uint32
	r := &actionReader{data: actions}
	for r.pos != len(actions) {
		if err := r.skip(5); err != nil {
			return nil, err
		}
		frame := binary.LittleEndian.Uint32(actions[r.pos-5:])
		if frame < lastFrame {
			return nil, fmt.Errorf("frame %d after frame %d", frame, lastFrame)
		}
		lastFrame = frame
		size := int(actions[r.pos-1])
		if err := r.skip(size); err != nil {
			return nil, err
		}
		block, err := parseOpenBWBlock(frame, actions[r.pos-size:r.pos], slotIDs)
		if err != nil {
			return nil, fmt.Errorf("frame %d: %w", frame, err)
		}
		parsed = append(parsed, block...)
	}
	return parsed, nil
}

func parseOpenBWBlock(frame uint32, block []byte, slotIDs [slotCount]uint32) ([]parsedAction, error) {
	var parsed []parsedAction
	r := &actionReader{data: block}
	for r.pos != len(block) {
		playerID, err := r.getUint8()
		if err != nil {
			return nil, err
		}
		if !slices.Contains(slotIDs[:], uint32(playerID)) {
			return nil, fmt.Errorf("player id %d not found", playerID)
		}
		id, err := r.getUint8()
		if err != nil {
			return nil, err
		}
		payloadStart := r.pos
		if err := readOpenBWAction(r, id); err != nil {
			return nil, err
		}
		fields := openBWActionFields(id, block[payloadStart:r.pos])
		parsed = append(parsed, parsedAction{frame: frame, playerID: playerID, id: id, fields: fields})
	}
	return parsed, nil
}
