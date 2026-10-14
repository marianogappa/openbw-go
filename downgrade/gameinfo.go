package downgrade

import (
	"encoding/binary"
	"fmt"

	"github.com/icza/screp/rep/repcore"
)

const (
	slotCount             = 12
	colorSlotCount        = 8
	slotsOffset           = 0xa1
	slotSize              = 36
	slotPlayerIDOffset    = 4
	playerColorsOffset    = 0x251
	cclrFootprintSize     = 16
	openBWPlayerColorSize = 16
)

func slotPlayerIDs(gameInfo []byte) [slotCount]uint32 {
	var ids [slotCount]uint32
	for i := range ids {
		ids[i] = binary.LittleEndian.Uint32(gameInfo[slotsOffset+i*slotSize+slotPlayerIDOffset:])
	}
	return ids
}

func playerColorOffset(slot int) int {
	return playerColorsOffset + slot*4
}

func playerColor(gameInfo []byte, slot int) uint32 {
	return binary.LittleEndian.Uint32(gameInfo[playerColorOffset(slot):])
}

func setPlayerColor(gameInfo []byte, slot int, color uint32) {
	binary.LittleEndian.PutUint32(gameInfo[playerColorOffset(slot):], color)
}

func colorFromFootprint(cclr []byte, slot int) (uint32, bool) {
	start := slot * cclrFootprintSize
	if start+cclrFootprintSize > len(cclr) {
		return 0, false
	}
	color := repcore.ColorByFootprint(cclr[start : start+cclrFootprintSize])
	if color == nil {
		return 0, false
	}
	return color.ID, true
}

func downgradePlayerColors(gameInfo, cclr []byte) []string {
	var changes []string
	for slot := range colorSlotCount {
		want := playerColor(gameInfo, slot)
		if fromCCLR, ok := colorFromFootprint(cclr, slot); ok {
			want = fromCCLR
		}
		if want >= openBWPlayerColorSize {
			want = uint32(slot)
		}
		if old := playerColor(gameInfo, slot); old != want {
			setPlayerColor(gameInfo, slot, want)
			changes = append(changes, fmt.Sprintf("player color slot %d: %d -> %d", slot, old, want))
		}
	}
	return changes
}
