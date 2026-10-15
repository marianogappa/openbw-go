package main

import (
	"encoding/binary"
	"fmt"
	"os"

	"github.com/icza/screp/repparser/repdecoder"
)

const (
	sectionIDLimits   = 0x53544d4c
	sectionIDBugFixes = 0x58494642
	sectionIDGameCfg  = 0x47464347
	gameCfgBuildAt    = 4
	replayIDSize      = 4
	gameInfoSize      = 633
	playerNamesSize   = 0x300
)

type extraSections struct {
	Limits   []uint32
	BugFixes string
	Build    uint32
}

func readExtraSections(path string) (extra extraSections, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("decode replay: %v", r)
		}
	}()
	data, err := os.ReadFile(path)
	if err != nil {
		return extra, err
	}
	dec := repdecoder.New(data)
	defer dec.Close()
	if err := skipCoreSections(dec); err != nil {
		return extra, err
	}
	for dec.NewSection() == nil {
		body, id, err := dec.Section(0)
		if err != nil {
			break
		}
		switch uint32(id) {
		case sectionIDLimits:
			extra.Limits = uint32s(body)
		case sectionIDBugFixes:
			extra.BugFixes = fmt.Sprintf("%x", body)
		case sectionIDGameCfg:
			extra.Build = gameBuild(body)
		}
	}
	return extra, nil
}

func skipCoreSections(dec repdecoder.Decoder) error {
	for _, size := range []int32{replayIDSize, gameInfoSize} {
		if err := skipSection(dec, size); err != nil {
			return err
		}
	}
	for range 2 {
		if err := skipSizedSection(dec); err != nil {
			return err
		}
	}
	return skipSection(dec, playerNamesSize)
}

func skipSection(dec repdecoder.Decoder, size int32) error {
	if err := dec.NewSection(); err != nil {
		return err
	}
	_, _, err := dec.Section(size)
	return err
}

func skipSizedSection(dec repdecoder.Decoder) error {
	if err := dec.NewSection(); err != nil {
		return err
	}
	sizeData, _, err := dec.Section(4)
	if err != nil || len(sizeData) != 4 {
		return fmt.Errorf("sized section: %v", err)
	}
	_, _, err = dec.Section(int32(binary.LittleEndian.Uint32(sizeData)))
	return err
}

func uint32s(data []byte) []uint32 {
	var out []uint32
	for i := 0; i+4 <= len(data); i += 4 {
		out = append(out, binary.LittleEndian.Uint32(data[i:]))
	}
	return out
}

func gameBuild(gameCfg []byte) uint32 {
	if len(gameCfg) < gameCfgBuildAt+4 {
		return 0
	}
	return binary.LittleEndian.Uint32(gameCfg[gameCfgBuildAt:])
}
