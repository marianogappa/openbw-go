package bwgame

import "iter"

const (
	triggerUnitIDAnyUnit   = 229
	triggerUnitIDMen       = 230
	triggerUnitIDBuildings = 231
	triggerUnitIDFactories = 232
)

const (
	triggerPlayerNonAlliedVictoryPlayers = 26
)

func (f *StateFunctions) processTriggers() {
	timerStep := int32(42)

	for i := 0; i != 12; i++ {
		if !f.st.TriggerWaiting[i] {
			continue
		}
		t := &f.st.TriggerWaitTimers[i]
		if *t == -1 {
			continue
		}
		if *t < timerStep {
			*t = 0
			f.st.TriggerWaiting[i] = false
			f.st.TriggerTimer = 0
		} else {
			*t -= timerStep
		}
	}

	if f.st.TriggerTimer != 0 {
		f.st.TriggerTimer--
		return
	}
	f.st.TriggerTimer = 30

	var ets ExecuteTriggerStruct

	anyTriggersExecuted := false
	for i := range f.activePlayers() {
		runningTriggers := f.st.RunningTriggers[i]
		for j := range runningTriggers {
			rt := &runningTriggers[j]
			if rt.Flags&8 != 0 {
				continue
			}
			t := rt.T
			executeNow := true
			if rt.Flags&1 == 0 {
				for k := range t.Conditions {
					c := &t.Conditions[k]
					if c.Type == 0 {
						break
					}
					if !f.testTriggerCondition(c, i) {
						executeNow = false
						break
					}
				}
			}
			if executeNow {
				rt.CurrentActionIndex = 0
				f.executeTrigger(&ets, i, rt, t)
				anyTriggersExecuted = true
			}
		}
	}

	if !anyTriggersExecuted {
		return
	}
	winners := 0
	for p := range f.activePlayers() {
		if f.playerWon(p) {
			winners++
		}
	}
	if winners != 0 {
		for p := range f.activePlayers() {
			if ets.VictoryState[p] == 0 {
				ets.VictoryState[p] = 2
			}
		}
		initiallyActiveCount := 0
		for i := range f.st.Players {
			if f.st.Players[i].InitiallyActive {
				initiallyActiveCount++
			}
		}
		if winners == initiallyActiveCount {
			for i := range f.st.Players {
				if f.st.Players[i].InitiallyActive {
					ets.VictoryState[i] = 5
				}
			}
		}
	}
	for i := int32(0); i != 8; i++ {
		s := ets.VictoryState[i]
		if s != f.st.Players[i].VictoryState {
			f.st.Players[i].VictoryState = s
			f.onVictoryState(i, s)
			if s == 2 {
				f.onPlayerEliminated(i)
				f.removePlayer(i)
			}
		}
		if s != 0 && s != 4 {
			f.st.RunningTriggers[i] = f.st.RunningTriggers[i][:0]
		}
	}
}

func (f *StateFunctions) triggerPlayers(owner int32, player int32) iter.Seq[int32] {
	return func(yield func(int32) bool) {
		for _, n := range allPlayerSlots {
			if !f.triggerPlayersPred(owner, player, n) {
				continue
			}
			if !yield(n) {
				return
			}
		}
	}
}

func (f *StateFunctions) triggerCommandCount(countObj *UnitCounters, owner int32, player int32, unitID int32, completedUnits bool) int32 {
	r := int32(0)
	for p := range f.triggerPlayers(owner, player) {
		if completedUnits {
			switch unitID {
			case triggerUnitIDAnyUnit:
				r += countObj.CompletedNonBuildingCounts[p] + countObj.CompletedBuildingCounts[p]
			case triggerUnitIDMen:
				r += countObj.CompletedNonBuildingCounts[p]
			case triggerUnitIDBuildings:
				r += countObj.CompletedBuildingCounts[p]
			case triggerUnitIDFactories:
				r += countObj.CompletedFactoryCounts[p]
			default:
				r += countObj.CompletedUnitCounts[p][unitID]
			}
		} else {
			switch unitID {
			case triggerUnitIDAnyUnit:
				r += countObj.NonBuildingCounts[p] + countObj.BuildingCounts[p]
			case triggerUnitIDMen:
				r += countObj.NonBuildingCounts[p]
			case triggerUnitIDBuildings:
				r += countObj.BuildingCounts[p]
			case triggerUnitIDFactories:
				r += countObj.FactoryCounts[p]
			default:
				r += countObj.UnitCounts[owner][unitID]
			}
		}
	}
	return r
}

func (f *StateFunctions) triggerCountComparison(c *TriggerCondition, count int32) bool {
	switch c.NumN {
	case 0:
		return count >= c.CountN
	case 1:
		return count <= c.CountN
	case 2:
		return count == c.CountN
	default:
		return false
	}
}

func (f *StateFunctions) triggerBringCount(loc *Location) BringUnitCounters {
	var r BringUnitCounters
	var addCompleted func(u *Unit)
	addCompleted = func(u *Unit) {
		r.CompletedUnitCounts[u.Owner][u.UnitType.ID]++
		if f.unitProvidesSpace(u) {
			for n := range f.loadedUnits(u) {
				addCompleted(n)
			}
		}
		if f.unitIsCarrier(u.UnitType) {
			r.CompletedUnitCounts[u.Owner][UnitTypesProtossInterceptor] += int32(u.Carrier.InsideCount)
		} else if f.unitIsReaver(u.UnitType) {
			r.CompletedUnitCounts[u.Owner][UnitTypesProtossReaver] += int32(u.Carrier.InsideCount)
		}
		if f.utWorker(u.UnitType) && u.Worker.Powerup != nil {
			if f.unitIs(u.Worker.Powerup.UnitType, UnitTypesPowerupFlag) {
				addCompleted(u.Worker.Powerup)
			} else {
				r.CompletedUnitCounts[u.Owner][u.Worker.Powerup.UnitType.ID]++
			}
		}
		if f.unitIs(u.UnitType, UnitTypesTerranNuclearSilo) && u.Building.Silo.Ready {
			r.CompletedUnitCounts[u.Owner][UnitTypesTerranNuclearMissile]++
		}
		if u.UnitType.GroupFlags&GroupFlagsFactory != 0 {
			r.CompletedFactoryCounts[u.Owner]++
		}
		if u.UnitType.GroupFlags&GroupFlagsMen != 0 {
			r.CompletedNonBuildingCounts[u.Owner]++
		} else if u.UnitType.GroupFlags&GroupFlagsBuilding != 0 {
			r.CompletedBuildingCounts[u.Owner]++
		}
	}
	for u := range f.findUnits(loc.Area) {
		if f.utTurret(u.UnitType) {
			continue
		}
		if !f.unitIsAtElevationFlags(u, loc.ElevationFlags) {
			continue
		}

		r.UnitCounts[u.Owner][u.UnitType.ID]++
		if u.UnitType.GroupFlags&GroupFlagsFactory != 0 {
			r.FactoryCounts[u.Owner]++
		}
		if u.UnitType.GroupFlags&GroupFlagsMen != 0 {
			r.NonBuildingCounts[u.Owner]++
		} else if u.UnitType.GroupFlags&GroupFlagsBuilding != 0 {
			r.BuildingCounts[u.Owner]++
		} else if f.unitIsEgg(u.UnitType) {
			r.NonBuildingCounts[u.Owner]++
		}

		if f.uCompleted(u) {
			addCompleted(u)
		}
	}
	return r
}

func (f *StateFunctions) triggerOpponentCount(owner int32, player int32) int32 {
	r := int32(0)
	for p := range f.triggerPlayers(owner, player) {
		for range f.triggerPlayers(p, triggerPlayerNonAlliedVictoryPlayers) {
			r++
		}
	}
	return r
}

func (f *StateFunctions) testTriggerCondition(c *TriggerCondition, owner int32) bool {
	switch c.Type {
	case 2:
		return f.triggerCountComparison(c, f.triggerCommandCount(&f.st.UnitCounters, owner, c.Group, c.UnitID, c.NumN != 1))
	case 3:
		bringCounts := f.triggerBringCount(&f.st.Locations[c.Location-1])
		return f.triggerCountComparison(c, f.triggerCommandCount(&bringCounts, owner, c.Group, c.UnitID, c.NumN != 1))
	case 12:
		return f.triggerCountComparison(c, f.st.CurrentFrame)
	case 14:
		return f.triggerCountComparison(c, f.triggerOpponentCount(owner, c.Group))
	case 22:
		return true
	default:
		throwError("unknown trigger condition %d\n", c.Type)
		return false
	}
}

func (f *StateFunctions) triggerCreateUnit(unitType *UnitTypeT, pos Xy, owner int32) *Unit {
	if unitType.StareditAvailabilityFlags&2 == 0 {
		return nil
	}
	pos = f.restrictUnitPosToBounds(pos, unitType, f.mapBounds())
	if f.utBuilding(unitType) {
		halfPlacementSize := XyDivScalar(unitType.PlacementSize, 2)
		topLeft := f.restrictPosToMapBounds(XySub(pos, halfPlacementSize))
		pos = XyAdd(XyMulScalar(XyDivScalar(topLeft, 32), 32), halfPlacementSize)
		if !f.canPlaceBuilding(nil, owner, unitType, pos, true, false) {
			return nil
		}
	}
	u := f.createUnit(unitType, pos, owner)
	if u == nil {
		f.displayLastErrorForPlayer(owner)
		return nil
	}
	f.onUnitCreated(u, SourceTrigger, nil)
	spreadsCreep := f.unitTypeSpreadsCreep(u.UnitType, true)
	if spreadsCreep {
		f.setCreepBuildingTiles(u.UnitType, u.Sprite.Position)
		f.addCreepProvider(u)
	} else if f.utRequiresCreep(u.UnitType) {
		f.spreadCreepCompletely(u.UnitType, u.Sprite.Position)
		f.addCreepProvider(u)
	}

	f.finishBuildingUnit(u)
	if !f.placeCompletedUnit(u) {
		f.killUnitWith(u, noKiller(ReasonTrigger))
		return nil
	}
	f.completeUnit(u)

	if spreadsCreep {
		f.addCreepProvider(u)
	}
	if f.utAddon(u.UnitType) {
		f.setUnitOwnerWith(u, 11, false, OwnerTrigger)
	}
	if f.uGroundedBuilding(u) && f.unitRace(u.UnitType) == RaceTerran {
		f.findAndConnectAddon(u)
	}
	f.st.UpdatePsionicMatrix = true
	if f.unitIs(u.UnitType, UnitTypesResourceVespeneGeyser) || f.unitIsRefinery(u.UnitType) {
		f.setUnitResources(u, 5000)
	} else if f.unitIsMineralField(u.UnitType) {
		f.setUnitResources(u, 1500)
	}
	return u
}

func (f *StateFunctions) playerWon(owner int32) bool {
	return f.st.Players[owner].VictoryState >= 3
}

func (f *StateFunctions) playerDefeated(owner int32) bool {
	s := f.st.Players[owner].VictoryState
	return s != 0 && !f.playerWon(owner)
}

func (f *StateFunctions) playerSlotActive(n int32) bool {
	if n >= 8 {
		return false
	}
	switch f.st.Players[n].Controller {
	case PlayerControllerOccupied:
		return !f.playerDefeated(n)
	case PlayerControllerComputerGame:
		return !f.playerDefeated(n)
	case PlayerControllerNeutral:
		return true
	case PlayerControllerRescuePassive:
		return true
	case PlayerControllerUnusedRescueActive:
		return true
	default:
		return false
	}
}

func (f *StateFunctions) triggerPlayersPred(owner int32, player int32, n int32) bool {
	if !f.playerSlotActive(n) {
		return false
	}
	if player < 12 {
		return n == player
	}
	switch player {
	case 12:
		return false
	case 13:
		return n == owner
	case 14:
		return owner != n && f.st.Alliances[owner][n] == 0
	case 15:
		return owner != n && f.st.Alliances[owner][n] == 2
	case 16:
		return owner != n && f.st.Alliances[owner][n] == 1
	case 17:
		return true
	case 18:
		return n < 8 && f.st.Players[n].Force == 1
	case 19:
		return n < 8 && f.st.Players[n].Force == 2
	case 20:
		return n < 8 && f.st.Players[n].Force == 3
	case 21:
		return n < 8 && f.st.Players[n].Force == 4
	case triggerPlayerNonAlliedVictoryPlayers:
		return owner != n && (f.st.Alliances[owner][n] != 2 || f.st.Alliances[n][owner] != 2)
	default:
		return false
	}
}

func (f *StateFunctions) unitIsAtElevationFlags(u *Unit, elevationFlags int32) bool {
	groundHeight := f.getGroundHeightAt(u.Sprite.Position)
	if f.uFlying(u) {
		switch groundHeight {
		case 0:
			return elevationFlags&1 == 0
		case 1:
			return elevationFlags&2 == 0
		case 2:
			return elevationFlags&4 == 0
		}
	} else {
		switch groundHeight {
		case 0:
			return elevationFlags&8 == 0
		case 1:
			return elevationFlags&0x10 == 0
		case 2:
			return elevationFlags&0x20 == 0
		}
	}
	return true
}

func (f *StateFunctions) triggerUnitPred(owner int32, u *Unit, player int32, uid int32) bool {
	if f.unitProvidesSpace(u) {
		for n := range f.loadedUnits(u) {
			if f.triggerUnitPred(owner, n, player, uid) {
				return true
			}
		}
	}
	if f.utWorker(u.UnitType) && u.Worker.Powerup != nil {
		if f.triggerUnitPred(owner, u.Worker.Powerup, player, uid) {
			return true
		}
	}

	if f.unitDying(u) {
		return false
	}
	if f.utPowerup(u.UnitType) {
		throwError("trigger_unit_pred: powerup")
	}
	if !f.triggerPlayersPred(owner, player, u.Owner) {
		return false
	}
	return f.triggerUnitTypeMatches(u, uid)
}

func (f *StateFunctions) triggerUnitTypeMatches(u *Unit, uid int32) bool {
	switch uid {
	case triggerUnitIDAnyUnit:
		return true
	case triggerUnitIDMen:
		return u.UnitType.GroupFlags&GroupFlagsMen != 0
	case triggerUnitIDBuildings:
		return u.UnitType.GroupFlags&GroupFlagsBuilding != 0
	case triggerUnitIDFactories:
		return u.UnitType.GroupFlags&GroupFlagsFactory != 0
	default:
		return f.unitIs(u.UnitType, UnitTypes(uid))
	}
}

func (f *StateFunctions) triggerFindUnit(owner int32, loc *Location, player int32, unitID int32) *Unit {
	for u := range f.findUnits(loc.Area) {
		if !f.unitIsAtElevationFlags(u, loc.ElevationFlags) {
			continue
		}
		if f.triggerUnitPred(owner, u, player, unitID) {
			return u
		}
	}
	return nil
}

func (f *StateFunctions) executeTriggerAction(ets *ExecuteTriggerStruct, owner int32, rt *RunningTrigger, ra *RunningTriggerAction, a *TriggerAction) bool {
	switch a.Type {
	case 1:
		f.triggerActionSetVictoryState(ets, owner, 3)
		return true
	case 2:
		f.triggerActionSetVictoryState(ets, owner, 2)
		return true
	case 3:
		rt.Flags |= 4
		return true
	case 4:
		return f.triggerActionWait(owner, rt, ra, a)
	case 15:
		f.triggerActionAiScript(owner, a)
		return true
	case 22:
		f.triggerActionKillUnit(owner, a)
		return true
	case 24:
		f.triggerActionRemoveUnit(owner, a)
		return true
	case 25:
		f.triggerActionRemoveUnitAtLocation(owner, a)
		return true
	case 26:
		f.triggerActionSetResources(owner, a)
		return true
	case 38:
		f.triggerActionMoveLocation(owner, a)
		return true
	case 44:
		f.triggerActionCreateUnit(owner, a)
		return true
	case 46:
		f.triggerActionOrder(owner, a)
		return true
	case 52:
		f.triggerActionSetUnitResources(owner, a)
		return true
	default:
		throwError("unknown trigger action %d", a.Type)
		return false
	}
}

func (f *StateFunctions) triggerActionSetVictoryState(ets *ExecuteTriggerStruct, owner int32, state int32) {
	controller := f.st.Players[owner].Controller
	if controller == PlayerControllerOccupied || controller == PlayerControllerComputerGame {
		ets.VictoryState[owner] = state
	}
}

func (f *StateFunctions) triggerActionWait(owner int32, rt *RunningTrigger, ra *RunningTriggerAction, a *TriggerAction) bool {
	if f.st.TriggerWaiting[owner] {
		return false
	}
	if ra.Flags&1 != 0 {
		ra.Flags &^= 1
		return true
	}
	if rt.Flags&0x10 != 0 {
		return true
	}
	f.st.TriggerWaiting[owner] = true
	f.st.TriggerWaitTimers[owner] = a.TimeN
	ra.Flags |= 1
	return false
}

func (f *StateFunctions) triggerActionAiScript(owner int32, a *TriggerAction) {
	switch a.Group2N {
	case 0x3069562b:
		f.st.SharedVision[0] |= 1 << owner
	case 0x3169562b:
		f.st.SharedVision[1] |= 1 << owner
	case 0x3269562b:
		f.st.SharedVision[2] |= 1 << owner
	case 0x3369562b:
		f.st.SharedVision[3] |= 1 << owner
	case 0x3469562b:
		f.st.SharedVision[4] |= 1 << owner
	case 0x3569562b:
		f.st.SharedVision[5] |= 1 << owner
	case 0x3669562b:
		f.st.SharedVision[6] |= 1 << owner
	case 0x3769562b:
		f.st.SharedVision[7] |= 1 << owner
	default:
		throwError("unknown ai script")
	}
}

func (f *StateFunctions) triggerActionKillUnit(owner int32, a *TriggerAction) {
	for p := range f.triggerPlayers(owner, a.GroupN) {
		uid := a.ExtraN
		list := &f.st.PlayerUnits[p]
		for i := list.Begin(); i != list.End(); {
			u := i
			i = list.Next(i)
			if f.utTurret(u.UnitType) {
				continue
			}
			if f.unitDying(u) {
				continue
			}
			if f.utPowerup(u.UnitType) {
				throwError("trigger kill unit fixme: powerup")
			}
			if u.Owner != p {
				continue
			}
			if f.triggerUnitTypeMatches(u, uid) {
				f.killUnitWith(u, noKiller(ReasonTrigger))
			}
		}
	}
}

func (f *StateFunctions) triggerActionRemoveUnit(owner int32, a *TriggerAction) {
	for p := range f.triggerPlayers(owner, a.GroupN) {
		uid := a.ExtraN
		list := &f.st.PlayerUnits[p]
		for i := list.Begin(); i != list.End(); {
			u := i
			i = list.Next(i)
			if f.utTurret(u.UnitType) {
				continue
			}
			if f.unitDying(u) {
				continue
			}
			if f.utPowerup(u.UnitType) {
				throwError("trigger kill unit fixme: powerup")
			}
			if u.Owner != p {
				continue
			}
			if !f.triggerUnitTypeMatches(u, uid) {
				continue
			}
			f.hideUnit(u, true)
			f.killUnitWith(u, noKiller(ReasonRemoved))
		}
	}
}

func (f *StateFunctions) triggerActionRemoveUnitAtLocation(owner int32, a *TriggerAction) {
	uid := a.ExtraN
	var proc func(u *Unit)
	proc = func(u *Unit) {
		if !f.unitIsAtElevationFlags(u, f.st.Locations[a.Location-1].ElevationFlags) {
			return
		}

		if f.unitProvidesSpace(u) {
			for n := range f.loadedUnits(u) {
				proc(n)
			}
		}
		if f.utWorker(u.UnitType) && u.Worker.Powerup != nil {
			proc(u.Worker.Powerup)
		}

		if f.utTurret(u.UnitType) {
			return
		}
		if f.unitDying(u) {
			return
		}
		if f.utPowerup(u.UnitType) {
			throwError("trigger remove unit fixme: powerup")
		}
		if !f.triggerPlayersPred(owner, a.GroupN, u.Owner) {
			return
		}
		if !f.triggerUnitTypeMatches(u, uid) {
			return
		}
		f.hideUnit(u, true)
		f.killUnitWith(u, noKiller(ReasonRemoved))
	}
	for u := range f.findUnits(f.st.Locations[a.Location-1].Area) {
		proc(u)
	}
}

func (f *StateFunctions) triggerActionSetResources(owner int32, a *TriggerAction) {
	for p := range f.triggerPlayers(owner, a.GroupN) {
		if p >= 8 {
			continue
		}
		setsMinerals := a.ExtraN == 0 || a.ExtraN == 2
		setsGas := a.ExtraN == 1 || a.ExtraN == 2
		switch a.NumN {
		case 7:
			if setsMinerals {
				f.st.CurrentMinerals[p] = a.Group2N
				f.st.TotalMineralsGathered[p] = a.Group2N
			}
			if setsGas {
				f.st.CurrentGas[p] = a.Group2N
				f.st.TotalGasGathered[p] = a.Group2N
			}
		case 8:
			if setsMinerals {
				f.st.CurrentMinerals[p] += a.Group2N
				f.st.TotalMineralsGathered[p] += a.Group2N
			}
			if setsGas {
				f.st.CurrentGas[p] += a.Group2N
				f.st.TotalGasGathered[p] += a.Group2N
			}
		}
	}
}

func locationCenter(loc *Location) Xy {
	return XyDivScalar(XyAdd(loc.Area.From, loc.Area.To), 2)
}

func (f *StateFunctions) triggerActionMoveLocation(owner int32, a *TriggerAction) {
	loc := &f.st.Locations[a.Location-1]
	pos := locationCenter(loc)
	u := f.triggerFindUnit(owner, loc, a.GroupN, a.ExtraN)
	if u != nil {
		pos = u.Sprite.Position
	}
	targetLoc := &f.st.Locations[a.Group2N-1]
	diff := XySub(pos, locationCenter(loc))
	from := XyAdd(targetLoc.Area.From, diff)
	to := XyAdd(targetLoc.Area.To, diff)
	if from.X < 0 {
		to.X -= from.X
		from.X = 0
	}
	if from.Y < 0 {
		to.Y -= from.Y
		from.Y = 0
	}
	mapWidth := int32(f.gameSt.MapWidth)
	mapHeight := int32(f.gameSt.MapHeight)
	if to.X >= mapWidth {
		from.X += mapWidth - 1 - to.X
		to.X = mapWidth - 1
	}
	if to.Y >= mapHeight {
		from.Y += mapHeight - 1 - to.Y
		to.Y = mapHeight - 1
	}
	targetLoc.Area.From = from
	targetLoc.Area.To = to
}

func (f *StateFunctions) triggerActionCreateUnit(owner int32, a *TriggerAction) {
	for p := range f.triggerPlayers(owner, a.GroupN) {
		ut := f.getUnitType(UnitTypes(a.ExtraN))
		if f.unitIsMineralField(ut) || f.unitIs(ut, UnitTypesResourceVespeneGeyser) {
			p = 11
		}
		pos := locationCenter(&f.st.Locations[a.Location-1])
		for i := int32(0); i != a.NumN; i++ {
			f.triggerCreateUnit(ut, pos, p)
		}
	}
}

func (f *StateFunctions) triggerActionOrder(owner int32, a *TriggerAction) {
	for u := range f.findUnits(f.st.Locations[a.Location-1].Area) {
		if !f.triggerPlayersPred(owner, a.GroupN, u.Owner) {
			continue
		}
		if !f.unitIsAtElevationFlags(u, f.st.Locations[a.Location-1].ElevationFlags) {
			continue
		}
		if f.utBuilding(u.UnitType) {
			if u.OrderType.ID == OrdersBuildingLand || u.OrderType.ID == OrdersBuildingLiftoff {
				continue
			}
		}
		if !f.uCanMove(u) {
			continue
		}
		if f.unitIs(u.UnitType, UnitTypesTerranNuclearMissile) {
			continue
		}
		if f.unitIsFighter(u.UnitType) {
			continue
		}
		if f.unitIsRescuable(u) {
			continue
		}
		var ok bool
		switch a.ExtraN {
		case triggerUnitIDAnyUnit:
			ok = true
		case triggerUnitIDMen:
			ok = u.UnitType.GroupFlags&GroupFlagsMen != 0
		case triggerUnitIDBuildings:
			ok = u.UnitType.GroupFlags&GroupFlagsBuilding != 0
		case triggerUnitIDFactories:
			ok = u.UnitType.GroupFlags&GroupFlagsFactory != 0
		default:
			ok = int32(u.UnitType.ID) == a.ExtraN
		}
		if !ok {
			continue
		}
		o := OrdersNothing
		switch a.NumN {
		case 0:
			o = OrdersMove
		case 1:
			o = OrdersPatrol
		case 2:
			o = OrdersAttackMove
		}
		order := f.getOrderType(o)
		targetPos := locationCenter(&f.st.Locations[a.Group2N-1])
		if f.uBurrowed(u) {
			f.unburrowUnit(u)
			f.setQueuedOrder(u, false, order, OrderTarget{Position: targetPos})
		} else {
			if !f.unitCanReceiveOrder(u, order, u.Owner) {
				continue
			}
			f.setUnitOrderPos(u, order, targetPos)
		}
	}
}

func (f *StateFunctions) triggerActionSetUnitResources(owner int32, a *TriggerAction) {
	for u := range f.findUnits(f.st.Locations[a.Location-1].Area) {
		if !f.triggerPlayersPred(owner, a.GroupN, u.Owner) {
			continue
		}
		if !f.unitIsAtElevationFlags(u, f.st.Locations[a.Location-1].ElevationFlags) {
			continue
		}
		f.setUnitResources(u, a.Group2N)
	}
}

func (f *StateFunctions) executeTrigger(ets *ExecuteTriggerStruct, owner int32, rt *RunningTrigger, t *Trigger) {
	rt.Flags |= 1
	index := rt.CurrentActionIndex
	for ; index != 64; index++ {
		a := &t.Actions[index]
		if a.Flags&2 != 0 {
			continue
		}
		if a.Type == 0 {
			index = 63
		} else if !f.executeTriggerAction(ets, owner, rt, &rt.Actions[index], a) {
			break
		}
	}
	rt.CurrentActionIndex = index
	if index == 64 {
		if rt.Flags&4 != 0 {
			rt.CurrentActionIndex = 0
			rt.Flags &^= 0x51
		} else {
			rt.Flags |= 8
		}
	}
}
