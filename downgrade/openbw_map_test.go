package downgrade

import (
	"encoding/binary"
	"fmt"
)

const (
	frameCountOffset         = 1
	victoryConditionOffset   = 137
	createInitialUnitsOffset = 141
	tournamentModeOffset     = 147
)

var openBWBaseChunks = []string{"VER ", "DIM ", "ERA ", "OWNR", "SIDE", "STR ", "SPRP", "FORC", "VCOD"}

var openBWVersionChunks = map[uint16]map[bool][]string{
	59: {
		true:  {"STR ", "MTXM", "THG2", "MASK", "UNIS", "UPGS", "TECS", "PUNI", "UPGR", "PTEC", "UNIT", "UPRP", "MRGN", "TRIG"},
		false: {"STR ", "MTXM", "THG2", "UNIT"},
	},
	63: {
		true:  {"STR ", "MTXM", "THG2", "MASK", "UNIS", "UPGS", "TECS", "PUNI", "UPGR", "PTEC", "UNIT", "UPRP", "MRGN", "TRIG"},
		false: {"STR ", "MTXM", "THG2", "UNIT"},
	},
	205: {
		true:  {"STR ", "MTXM", "THG2", "MASK", "UNIx", "UPGx", "TECx", "PUNI", "PUPx", "PTEx", "UNIT", "UPRP", "MRGN", "TRIG", "COLR"},
		false: {"STR ", "MTXM", "THG2", "UNIT", "COLR"},
	},
}

func usesMapSettings(gameInfo []byte) bool {
	return gameInfo[victoryConditionOffset] == 0 && gameInfo[tournamentModeOffset] == 0 && gameInfo[createInitialUnitsOffset] == 0
}

const openBWColorChunkSize = 8

func checkOpenBWChunkHeaders(chk []byte) error {
	pos := 0
	for pos < len(chk) {
		if len(chk)-pos < chunkHeaderSize {
			return fmt.Errorf("partial chunk header at offset %d", pos)
		}
		size := binary.LittleEndian.Uint32(chk[pos+4:])
		if uint64(size) > uint64(len(chk)-pos-chunkHeaderSize) {
			return nil
		}
		pos += chunkHeaderSize + int(size)
	}
	return nil
}

func checkOpenBWCanLoadMap(chk, gameInfo []byte) error {
	if err := checkOpenBWChunkHeaders(chk); err != nil {
		return err
	}
	chunks := splitChunks(chk)
	present := map[string]bool{}
	var version uint16
	for _, c := range chunks {
		present[c.tag] = true
		if c.tag != "VER " {
			continue
		}
		if len(c.data) < 2 {
			return fmt.Errorf("VER chunk has %d bytes", len(c.data))
		}
		version = binary.LittleEndian.Uint16(c.data)
	}
	required, supported := openBWVersionChunks[version]
	if !supported {
		return fmt.Errorf("unsupported map version %d", version)
	}
	if version == 205 {
		if err := checkOpenBWColorChunks(chunks); err != nil {
			return err
		}
	}
	for _, tag := range append(openBWBaseChunks, required[usesMapSettings(gameInfo)]...) {
		if !present[tag] {
			return fmt.Errorf("map is missing required chunk %q", tag)
		}
	}
	return nil
}

func checkOpenBWColorChunks(chunks []chunk) error {
	for _, c := range chunks {
		if c.tag == "COLR" && len(c.data) < openBWColorChunkSize {
			return fmt.Errorf("COLR chunk has %d bytes", len(c.data))
		}
	}
	return nil
}
