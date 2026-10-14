package main

import (
	"github.com/icza/screp/rep/repcmd"
	"github.com/icza/screp/rep/repcore"
	"github.com/icza/screp/repparser"
)

const minimumOrdersForMissingType = 5

var productionCommandIDs = map[byte]bool{
	repcmd.TypeIDBuild:         true,
	repcmd.TypeIDTrain:         true,
	repcmd.TypeIDUnitMorph:     true,
	repcmd.TypeIDBuildingMorph: true,
}

type commandStats struct {
	Slot             int
	PlayerID         byte
	Name             string
	Race             string
	TotalCmds        int
	ProductionCmds   int
	OrderedUnits     int
	OrderedBuildings int
	OrdersByType     map[int]int
	LastCmdFrame     int
	LeaveFrame       int
	cmdFrames        []int
}

func (s commandStats) cmdsAfter(frame int) int {
	count := 0
	for _, f := range s.cmdFrames {
		if f > frame {
			count++
		}
	}
	return count
}

func readCommandStats(path string) (map[int]*commandStats, error) {
	r, err := repparser.ParseFileConfig(path, repparser.Config{Commands: true})
	if err != nil {
		return nil, err
	}
	bySlot := map[int]*commandStats{}
	byID := map[byte]*commandStats{}
	for _, p := range r.Header.Players {
		s := &commandStats{Slot: int(p.SlotID), PlayerID: p.ID, Name: p.Name, Race: p.Race.Name, OrdersByType: map[int]int{}}
		bySlot[s.Slot] = s
		byID[p.ID] = s
	}
	for _, cmd := range r.Commands.Cmds {
		if s := byID[cmd.BaseCmd().PlayerID]; s != nil {
			addCommand(s, cmd)
		}
	}
	return bySlot, nil
}

func addCommand(s *commandStats, cmd repcmd.Cmd) {
	base := cmd.BaseCmd()
	frame := int(base.Frame)
	if base.Type.ID == repcmd.TypeIDLeaveGame {
		s.LeaveFrame = frame
		return
	}
	s.TotalCmds++
	s.cmdFrames = append(s.cmdFrames, frame)
	if frame > s.LastCmdFrame {
		s.LastCmdFrame = frame
	}
	if !productionCommandIDs[base.Type.ID] {
		return
	}
	s.ProductionCmds++
	if base.IneffKind != repcore.IneffKindEffective {
		return
	}
	addOrder(s, cmd)
}

func addOrder(s *commandStats, cmd repcmd.Cmd) {
	switch c := cmd.(type) {
	case *repcmd.TrainCmd:
		s.OrderedUnits++
		s.OrdersByType[int(c.Unit.ID)]++
	case *repcmd.BuildCmd:
		s.OrderedBuildings++
		s.OrdersByType[int(c.Unit.ID)]++
	case *repcmd.BuildingMorphCmd:
		s.OrderedBuildings++
		s.OrdersByType[int(c.Unit.ID)]++
	}
}

func unitName(id int) string {
	return repcmd.UnitByID(uint16(id)).Name
}
