package bwgame

import "iter"

func (f *StateFunctions) tileHasCreep(pos Xy) bool {
	panic(notPorted("tile_has_creep"))
}

func (f *StateFunctions) creepIsBeingSustained(tilePos XyT[uint]) bool {
	panic(notPorted("creep_is_being_sustained"))
}

func (f *StateFunctions) setCreepReceding(tilePos XyT[uint]) bool {
	panic(notPorted("set_creep_receding"))
}

func (f *StateFunctions) removeCreepProvider(unitType *UnitTypeT, pos Xy, isCompleted bool) bool {
	panic(notPorted("remove_creep_provider"))
}

func (f *StateFunctions) removeUnitCreepProvider(u *Unit) bool {
	panic(notPorted("remove_creep_provider"))
}

func (f *StateFunctions) removeExtractorCreep(unitType *UnitTypeT, pos Xy, isCompleted bool) bool {
	panic(notPorted("remove_extractor_creep"))
}

func (f *StateFunctions) addCreepProvider(u *Unit) {
	panic(notPorted("add_creep_provider"))
}

func (f *StateFunctions) zergBuildingStartConstruction(u *Unit) {
	panic(notPorted("zerg_building_start_construction"))
}

func (f *StateFunctions) setCreepBuildingTiles(unitType *UnitTypeT, pos Xy) {
	panic(notPorted("set_creep_building_tiles"))
}

func (f *StateFunctions) recedeCreep() {
	panic(notPorted("recede_creep"))
}

func (f *StateFunctions) processTriggers() {
	panic(notPorted("process_triggers"))
}

func (f *StateFunctions) triggerPlayers(owner int32, player int32) iter.Seq[int32] {
	panic(notPorted("trigger_players"))
}

func (f *StateFunctions) triggerCommandCount(countObj *UnitCounters, owner int32, player int32, unitID int32, completedUnits bool) int32 {
	panic(notPorted("trigger_command_count"))
}

func (f *StateFunctions) triggerCountComparison(c *TriggerCondition, count int32) bool {
	panic(notPorted("trigger_count_comparison"))
}

func (f *StateFunctions) triggerBringCount(loc *Location) BringUnitCounters {
	panic(notPorted("trigger_bring_count"))
}

func (f *StateFunctions) triggerOpponentCount(owner int32, player int32) int32 {
	panic(notPorted("trigger_opponent_count"))
}

func (f *StateFunctions) testTriggerCondition(c *TriggerCondition, owner int32) bool {
	panic(notPorted("test_trigger_condition"))
}

func (f *StateFunctions) triggerCreateUnit(unitType *UnitTypeT, pos Xy, owner int32) *Unit {
	panic(notPorted("trigger_create_unit"))
}

func (f *StateFunctions) playerWon(owner int32) bool {
	panic(notPorted("player_won"))
}

func (f *StateFunctions) playerDefeated(owner int32) bool {
	panic(notPorted("player_defeated"))
}

func (f *StateFunctions) playerSlotActive(n int32) bool {
	panic(notPorted("player_slot_active"))
}

func (f *StateFunctions) triggerPlayersPred(owner int32, player int32, n int32) bool {
	panic(notPorted("trigger_players_pred"))
}

func (f *StateFunctions) unitIsAtElevationFlags(u *Unit, elevationFlags int32) bool {
	panic(notPorted("unit_is_at_elevation_flags"))
}

func (f *StateFunctions) triggerUnitPred(owner int32, u *Unit, player int32, uid int32) bool {
	panic(notPorted("trigger_unit_pred"))
}

func (f *StateFunctions) triggerFindUnit(owner int32, loc *Location, player int32, unitID int32) *Unit {
	panic(notPorted("trigger_find_unit"))
}

func (f *StateFunctions) executeTriggerAction(ets *ExecuteTriggerStruct, owner int32, rt *RunningTrigger, ra *RunningTriggerAction, a *TriggerAction) bool {
	panic(notPorted("execute_trigger_action"))
}

func (f *StateFunctions) executeTrigger(ets *ExecuteTriggerStruct, owner int32, rt *RunningTrigger, t *Trigger) {
	panic(notPorted("execute_trigger"))
}

func (f *StateFunctions) tileCanHaveCreep(tilePos XyT[uint]) bool {
	panic(notPorted("tile_can_have_creep"))
}

func (f *StateFunctions) getMaxCreepBb(unitType *UnitTypeT, pos Xy, unitIsCompleted bool) RectT[XyT[uint]] {
	panic(notPorted("get_max_creep_bb"))
}

func (f *StateFunctions) countNeighboringCreepTiles(tilePos XyT[uint]) uint {
	panic(notPorted("count_neighboring_creep_tiles"))
}

func (f *StateFunctions) setTileCreep(tilePos XyT[uint], hasCreep bool) {
	panic(notPorted("set_tile_creep"))
}

func (f *StateFunctions) spreadCreep(unitType *UnitTypeT, pos Xy, outAnyTilesOccupied *bool) bool {
	panic(notPorted("spread_creep"))
}

func (f *StateFunctions) spreadCreepCompletely(unitType *UnitTypeT, pos Xy) {
	panic(notPorted("spread_creep_completely"))
}

func (f *StateFunctions) getSpawnLarvaPosition(u *Unit) Xy {
	panic(notPorted("get_spawn_larva_position"))
}

func (f *StateFunctions) spawnLarva(u *Unit) *Unit {
	panic(notPorted("spawn_larva"))
}
