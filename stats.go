package openbw

import "github.com/marianogappa/openbw-go/bwgame"

// PlayerStats is one player's numbers at one frame.
type PlayerStats struct {
	Frame            int32
	Owner            int8
	Minerals, Gas    int32
	GatheredMinerals int32
	GatheredGas      int32
	SupplyUsed       float64
	SupplyMax        float64
	Workers          int32
	Army             int32
	ArmyMinerals     int32
	ArmyGas          int32
	Buildings        int32
	UnitScore        int32
	BuildingScore    int32
}

type armyCost struct {
	unitType      bwgame.UnitTypes
	minerals, gas int32
	halve         bool
}

var notArmy = map[bwgame.UnitTypes]bool{
	bwgame.UnitTypesZergLarva:               true,
	bwgame.UnitTypesZergEgg:                 true,
	bwgame.UnitTypesZergCocoon:              true,
	bwgame.UnitTypesZergLurkerEgg:           true,
	bwgame.UnitTypesProtossInterceptor:      true,
	bwgame.UnitTypesProtossScarab:           true,
	bwgame.UnitTypesTerranVultureSpiderMine: true,
	bwgame.UnitTypesTerranNuclearMissile:    true,
}

var morphSources = map[bwgame.UnitTypes]bwgame.UnitTypes{
	bwgame.UnitTypesZergLurker:   bwgame.UnitTypesZergHydralisk,
	bwgame.UnitTypesZergGuardian: bwgame.UnitTypesZergMutalisk,
	bwgame.UnitTypesZergDevourer: bwgame.UnitTypesZergMutalisk,
}

var mergeSources = map[bwgame.UnitTypes]bwgame.UnitTypes{
	bwgame.UnitTypesProtossArchon:     bwgame.UnitTypesProtossHighTemplar,
	bwgame.UnitTypesProtossDarkArchon: bwgame.UnitTypesProtossDarkTemplar,
}

var workerTypes = []bwgame.UnitTypes{bwgame.UnitTypesTerranSCV, bwgame.UnitTypesZergDrone, bwgame.UnitTypesProtossProbe}

func armyCosts(st *bwgame.State) []armyCost {
	types := st.Game.UnitTypes.Vec
	var costs []armyCost
	for i := range types {
		ut := &types[i]
		if ut.GroupFlags&bwgame.GroupFlagsMen == 0 || ut.Flags&bwgame.UnitTypeTFlagWorker != 0 || notArmy[ut.ID] {
			continue
		}
		c := armyCost{unitType: ut.ID, minerals: ut.MineralCost, gas: ut.GasCost}
		if source, ok := morphSources[ut.ID]; ok {
			c.minerals += types[source].MineralCost
			c.gas += types[source].GasCost
		}
		if source, ok := mergeSources[ut.ID]; ok {
			c.minerals = 2 * types[source].MineralCost
			c.gas = 2 * types[source].GasCost
		}
		c.halve = ut.ID == bwgame.UnitTypesZergZergling || ut.ID == bwgame.UnitTypesZergScourge
		costs = append(costs, c)
	}
	return costs
}

// Stats is the PlayerStats row that Run records, computed now. Owner is 0..11.
func (g *Game) Stats(owner int) PlayerStats {
	p := g.Player(owner)
	s := PlayerStats{
		Frame:            g.st.CurrentFrame,
		Owner:            int8(owner),
		Minerals:         p.Minerals,
		Gas:              p.Gas,
		GatheredMinerals: p.GatheredMinerals,
		GatheredGas:      p.GatheredGas,
		SupplyUsed:       p.SupplyUsed,
		SupplyMax:        p.SupplyMax,
	}
	if owner < 0 || owner >= playerSlots {
		return s
	}
	counts := &g.st.CompletedUnitCounts[owner]
	for _, t := range workerTypes {
		s.Workers += counts[t]
	}
	for _, c := range g.army {
		n := counts[c.unitType]
		if n == 0 {
			continue
		}
		s.Army += n
		minerals, gas := n*c.minerals, n*c.gas
		if c.halve {
			minerals, gas = minerals/2, gas/2
		}
		s.ArmyMinerals += minerals
		s.ArmyGas += gas
	}
	s.Buildings = g.st.CompletedBuildingCounts[owner]
	s.UnitScore = g.st.UnitScore[owner]
	s.BuildingScore = g.st.BuildingScore[owner]
	return s
}
