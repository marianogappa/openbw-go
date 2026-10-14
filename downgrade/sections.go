package downgrade

import (
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/icza/screp/repparser/repdecoder"
)

const (
	replayIDSize    = 4
	gameInfoSize    = 633
	playerNamesSize = 0x300
	sectionIDCCLR   = 0x524c4343
	legacyReplayID  = 0x53526572
	modernReplayID  = 0x53526573
)

type sourceSections struct {
	format       repdecoder.RepFormat
	gameInfo     []byte
	actions      []byte
	mapData      []byte
	playerColors []byte
}

func readSourceSections(data []byte) (s sourceSections, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("decode replay: %v", r)
		}
	}()
	if legacy, err := ReadUncompressedLegacy(data); err == nil {
		return sourceSections{format: repdecoder.RepFormatLegacy, gameInfo: legacy.GameInfo, actions: legacy.Actions, mapData: legacy.MapData}, nil
	}
	dec := repdecoder.New(data)
	defer dec.Close()
	s.format = dec.RepFormat()
	if s.format == repdecoder.RepFormatUnknown {
		return s, errors.New("unknown replay format")
	}
	if err = readReplayID(dec); err != nil {
		return s, err
	}
	if s.gameInfo, err = readFixedSection(dec, gameInfoSize); err != nil {
		return s, fmt.Errorf("game info: %w", err)
	}
	if s.actions, err = readSizedSection(dec); err != nil {
		return s, fmt.Errorf("actions: %w", err)
	}
	if s.mapData, err = readSizedSection(dec); err != nil {
		return s, fmt.Errorf("map data: %w", err)
	}
	s.playerColors = readOptionalPlayerColors(dec)
	return s, nil
}

func readReplayID(dec repdecoder.Decoder) error {
	id, err := readFixedSection(dec, replayIDSize)
	if err != nil {
		return fmt.Errorf("replay id: %w", err)
	}
	switch binary.LittleEndian.Uint32(id) {
	case legacyReplayID, modernReplayID:
		return nil
	}
	return fmt.Errorf("invalid replay id %q", id)
}

func readFixedSection(dec repdecoder.Decoder, size int32) ([]byte, error) {
	if err := dec.NewSection(); err != nil {
		return nil, err
	}
	return readSectionData(dec, size)
}

func readSizedSection(dec repdecoder.Decoder) ([]byte, error) {
	if err := dec.NewSection(); err != nil {
		return nil, err
	}
	sizeData, err := readSectionData(dec, 4)
	if err != nil {
		return nil, fmt.Errorf("size: %w", err)
	}
	size := binary.LittleEndian.Uint32(sizeData)
	if size > 1<<30 {
		return nil, fmt.Errorf("section size %d too large", size)
	}
	return readSectionData(dec, int32(size))
}

func readSectionData(dec repdecoder.Decoder, size int32) ([]byte, error) {
	data, _, err := dec.Section(size)
	if err != nil {
		return nil, err
	}
	if int32(len(data)) != size {
		return nil, fmt.Errorf("got %d bytes, expected %d", len(data), size)
	}
	return data, nil
}

func skipPlayerNames(dec repdecoder.Decoder) error {
	if err := dec.NewSection(); err != nil {
		return err
	}
	_, _, err := dec.Section(playerNamesSize)
	return err
}

func readOptionalPlayerColors(dec repdecoder.Decoder) []byte {
	if dec.RepFormat() == repdecoder.RepFormatLegacy {
		return nil
	}
	if err := skipPlayerNames(dec); err != nil {
		return nil
	}
	for {
		if err := dec.NewSection(); err != nil {
			return nil
		}
		data, sectionID, err := dec.Section(0)
		if err != nil {
			return nil
		}
		if uint32(sectionID) == sectionIDCCLR {
			return data
		}
	}
}
