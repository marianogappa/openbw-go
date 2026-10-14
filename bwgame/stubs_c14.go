package bwgame

import "iter"

func (f *ActionFunctions) unitCanBeMultiSelected(u *Unit) bool {
	panic(notPorted("unit_can_be_multi_selected"))
}

func (f *ActionFunctions) selectedUnits(owner int32) iter.Seq[*Unit] {
	panic(notPorted("selected_units"))
}

func (f *ActionFunctions) controlGroup(owner int32, groupN uint) iter.Seq[*Unit] {
	panic(notPorted("control_group"))
}

func (f *ActionFunctions) getFirstSelectedUnit(owner int32) *Unit {
	panic(notPorted("get_first_selected_unit"))
}

func (f *ActionFunctions) getSingleSelectedUnit(owner int32) *Unit {
	panic(notPorted("get_single_selected_unit"))
}

func (f *ActionFunctions) onUnitDeselect(u *Unit) {
	panic(notPorted("on_unit_deselect"))
}

func (f *ActionFunctions) calcGroupMove(g *GroupMove, owner int32, targetPos Xy, canBeObstructed bool) {
	panic(notPorted("calc_group_move"))
}

func (f *ActionFunctions) getGroupMovePos(u *Unit, pos Xy, g *GroupMove) Xy {
	panic(notPorted("get_group_move_pos"))
}

func (f *ActionFunctions) morphArchonImpl(owner int32, isDarkArchon bool) bool {
	panic(notPorted("morph_archon_impl"))
}

func (f *ActionFunctions) actionSelect(owner int32, units []*Unit) bool {
	panic(notPorted("action_select"))
}

func (f *ActionFunctions) actionSelectUnit(owner int32, u *Unit) bool {
	panic(notPorted("action_select"))
}

func (f *ActionFunctions) actionShiftSelect(owner int32, units []*Unit) bool {
	panic(notPorted("action_shift_select"))
}

func (f *ActionFunctions) actionShiftSelectUnit(owner int32, u *Unit) bool {
	panic(notPorted("action_shift_select"))
}

func (f *ActionFunctions) actionDeselect(owner int32, units []*Unit) bool {
	panic(notPorted("action_deselect"))
}

func (f *ActionFunctions) actionDeselectUnit(owner int32, u *Unit) bool {
	panic(notPorted("action_deselect"))
}

func (f *ActionFunctions) actionTrain(owner int32, unitType *UnitTypeT) bool {
	panic(notPorted("action_train"))
}

func (f *ActionFunctions) actionBuild(owner int32, orderType *OrderTypeT, unitType *UnitTypeT, tilePos XyT[uint]) bool {
	panic(notPorted("action_build"))
}

func (f *ActionFunctions) actionDefaultOrder(owner int32, pos Xy, target *Unit, targetUnitType *UnitTypeT, queue bool) bool {
	panic(notPorted("action_default_order"))
}

func (f *ActionFunctions) actionOrder(owner int32, inputOrder *OrderTypeT, pos Xy, target *Unit, targetUnitType *UnitTypeT, queue bool) bool {
	panic(notPorted("action_order"))
}

func (f *ActionFunctions) actionStop(owner int32, queue bool) bool {
	panic(notPorted("action_stop"))
}

func (f *ActionFunctions) actionCancelBuildingUnit(owner int32) bool {
	panic(notPorted("action_cancel_building_unit"))
}

func (f *ActionFunctions) actionCancelBuildQueue(owner int32, slot uint) bool {
	panic(notPorted("action_cancel_build_queue"))
}

func (f *ActionFunctions) actionControlGroup(owner int32, groupN uint, subaction int32) bool {
	panic(notPorted("action_control_group"))
}

func (f *ActionFunctions) actionUnloadAll(owner int32, queue bool) bool {
	panic(notPorted("action_unload_all"))
}

func (f *ActionFunctions) actionUnload(owner int32, target *Unit) bool {
	panic(notPorted("action_unload"))
}

func (f *ActionFunctions) actionLiftoff(owner int32, pos Xy) bool {
	panic(notPorted("action_liftoff"))
}

func (f *ActionFunctions) actionResearch(owner int32, tech *TechTypeT) bool {
	panic(notPorted("action_research"))
}

func (f *ActionFunctions) actionCancelResearch(owner int32) bool {
	panic(notPorted("action_cancel_research"))
}

func (f *ActionFunctions) actionUpgrade(owner int32, upgrade *UpgradeTypeT) bool {
	panic(notPorted("action_upgrade"))
}

func (f *ActionFunctions) actionCancelUpgrade(owner int32) bool {
	panic(notPorted("action_cancel_upgrade"))
}

func (f *ActionFunctions) actionUnsiege(owner int32, queue bool) bool {
	panic(notPorted("action_unsiege"))
}

func (f *ActionFunctions) actionSiege(owner int32, queue bool) bool {
	panic(notPorted("action_siege"))
}

func (f *ActionFunctions) actionReturnCargo(owner int32, queue bool) bool {
	panic(notPorted("action_return_cargo"))
}

func (f *ActionFunctions) actionHoldPosition(owner int32, queue bool) bool {
	panic(notPorted("action_hold_position"))
}

func (f *ActionFunctions) actionCloak(owner int32) bool {
	panic(notPorted("action_cloak"))
}

func (f *ActionFunctions) actionDecloak(owner int32) bool {
	panic(notPorted("action_decloak"))
}

func (f *ActionFunctions) actionSetAlliances(owner int32, alliances [12]int32) bool {
	panic(notPorted("action_set_alliances"))
}

func (f *ActionFunctions) actionSetSharedVision(owner int32, flags uint32) bool {
	panic(notPorted("action_set_shared_vision"))
}

func (f *ActionFunctions) actionCancelAddon(owner int32) bool {
	panic(notPorted("action_cancel_addon"))
}

func (f *ActionFunctions) actionStimPack(owner int32) bool {
	panic(notPorted("action_stim_pack"))
}

func (f *ActionFunctions) actionCancelNuke(owner int32) bool {
	panic(notPorted("action_cancel_nuke"))
}

func (f *ActionFunctions) actionMorph(owner int32, unitType *UnitTypeT) bool {
	panic(notPorted("action_morph"))
}

func (f *ActionFunctions) actionMorphBuilding(owner int32, unitType *UnitTypeT) bool {
	panic(notPorted("action_morph_building"))
}

func (f *ActionFunctions) actionBurrow(owner int32, queue bool) bool {
	panic(notPorted("action_burrow"))
}

func (f *ActionFunctions) actionUnburrow(owner int32) bool {
	panic(notPorted("action_unburrow"))
}

func (f *ActionFunctions) actionCancelMorph(owner int32) bool {
	panic(notPorted("action_cancel_morph"))
}

func (f *ActionFunctions) actionMorphArchon(owner int32) bool {
	panic(notPorted("action_morph_archon"))
}

func (f *ActionFunctions) actionCarrierStop(owner int32) bool {
	panic(notPorted("action_carrier_stop"))
}

func (f *ActionFunctions) actionTrainFighter(owner int32) bool {
	panic(notPorted("action_train_fighter"))
}

func (f *ActionFunctions) actionMorphDarkArchon(owner int32) bool {
	panic(notPorted("action_morph_dark_archon"))
}

func (f *ActionFunctions) actionReaverStop(owner int32) bool {
	panic(notPorted("action_reaver_stop"))
}

func (f *ActionFunctions) actionPlayerLeave(owner int32, reason int32) bool {
	panic(notPorted("action_player_leave"))
}

func (f *ActionFunctions) actionCheat(owner int32, flags int32) bool {
	panic(notPorted("action_cheat"))
}

func (f *ActionFunctions) readActionSelect(owner int32, r *DataReader) bool {
	panic(notPorted("read_action_select"))
}

func (f *ActionFunctions) readActionShiftSelect(owner int32, r *DataReader) bool {
	panic(notPorted("read_action_shift_select"))
}

func (f *ActionFunctions) readActionDeselect(owner int32, r *DataReader) bool {
	panic(notPorted("read_action_deselect"))
}

func (f *ActionFunctions) readActionTrain(owner int32, r *DataReader) bool {
	panic(notPorted("read_action_train"))
}

func (f *ActionFunctions) readActionDefaultOrder(owner int32, r *DataReader) bool {
	panic(notPorted("read_action_default_order"))
}

func (f *ActionFunctions) readActionPlayerLeave(owner int32, r *DataReader) bool {
	panic(notPorted("read_action_player_leave"))
}

func (f *ActionFunctions) readActionBuild(owner int32, r *DataReader) bool {
	panic(notPorted("read_action_build"))
}

func (f *ActionFunctions) readActionStop(owner int32, r *DataReader) bool {
	panic(notPorted("read_action_stop"))
}

func (f *ActionFunctions) readActionOrder(owner int32, r *DataReader) bool {
	panic(notPorted("read_action_order"))
}

func (f *ActionFunctions) readActionCancelBuildingUnit(owner int32, r *DataReader) bool {
	panic(notPorted("read_action_cancel_building_unit"))
}

func (f *ActionFunctions) readActionCancelBuildQueue(owner int32, r *DataReader) bool {
	panic(notPorted("read_action_cancel_build_queue"))
}

func (f *ActionFunctions) readActionControlGroup(owner int32, r *DataReader) bool {
	panic(notPorted("read_action_control_group"))
}

func (f *ActionFunctions) readActionUnloadAll(owner int32, r *DataReader) bool {
	panic(notPorted("read_action_unload_all"))
}

func (f *ActionFunctions) readActionUnload(owner int32, r *DataReader) bool {
	panic(notPorted("read_action_unload"))
}

func (f *ActionFunctions) readActionLiftoff(owner int32, r *DataReader) bool {
	panic(notPorted("read_action_liftoff"))
}

func (f *ActionFunctions) readActionResearch(owner int32, r *DataReader) bool {
	panic(notPorted("read_action_research"))
}

func (f *ActionFunctions) readActionCancelResearch(owner int32, r *DataReader) bool {
	panic(notPorted("read_action_cancel_research"))
}

func (f *ActionFunctions) readActionUpgrade(owner int32, r *DataReader) bool {
	panic(notPorted("read_action_upgrade"))
}

func (f *ActionFunctions) readActionCancelUpgrade(owner int32, r *DataReader) bool {
	panic(notPorted("read_action_cancel_upgrade"))
}

func (f *ActionFunctions) readActionUnsiege(owner int32, r *DataReader) bool {
	panic(notPorted("read_action_unsiege"))
}

func (f *ActionFunctions) readActionSiege(owner int32, r *DataReader) bool {
	panic(notPorted("read_action_siege"))
}

func (f *ActionFunctions) readActionChat(owner int32, r *DataReader) bool {
	panic(notPorted("read_action_chat"))
}

func (f *ActionFunctions) readActionReturnCargo(owner int32, r *DataReader) bool {
	panic(notPorted("read_action_return_cargo"))
}

func (f *ActionFunctions) readActionHoldPosition(owner int32, r *DataReader) bool {
	panic(notPorted("read_action_hold_position"))
}

func (f *ActionFunctions) readActionCloak(owner int32, r *DataReader) bool {
	panic(notPorted("read_action_cloak"))
}

func (f *ActionFunctions) readActionDecloak(owner int32, r *DataReader) bool {
	panic(notPorted("read_action_decloak"))
}

func (f *ActionFunctions) readActionSetAlliances(owner int32, r *DataReader) bool {
	panic(notPorted("read_action_set_alliances"))
}

func (f *ActionFunctions) readActionSetSharedVision(owner int32, r *DataReader) bool {
	panic(notPorted("read_action_set_shared_vision"))
}

func (f *ActionFunctions) readActionCancelAddon(owner int32, r *DataReader) bool {
	panic(notPorted("read_action_cancel_addon"))
}

func (f *ActionFunctions) readActionStimPack(owner int32, r *DataReader) bool {
	panic(notPorted("read_action_stim_pack"))
}

func (f *ActionFunctions) readActionCancelNuke(owner int32, r *DataReader) bool {
	panic(notPorted("read_action_cancel_nuke"))
}

func (f *ActionFunctions) readActionPingMinimap(owner int32, r *DataReader) bool {
	panic(notPorted("read_action_ping_minimap"))
}

func (f *ActionFunctions) readActionMorph(owner int32, r *DataReader) bool {
	panic(notPorted("read_action_morph"))
}

func (f *ActionFunctions) readActionMorphBuilding(owner int32, r *DataReader) bool {
	panic(notPorted("read_action_morph_building"))
}

func (f *ActionFunctions) readActionBurrow(owner int32, r *DataReader) bool {
	panic(notPorted("read_action_burrow"))
}

func (f *ActionFunctions) readActionUnburrow(owner int32, r *DataReader) bool {
	panic(notPorted("read_action_unburrow"))
}

func (f *ActionFunctions) readActionCancelMorph(owner int32, r *DataReader) bool {
	panic(notPorted("read_action_cancel_morph"))
}

func (f *ActionFunctions) readActionMorphArchon(owner int32, r *DataReader) bool {
	panic(notPorted("read_action_morph_archon"))
}

func (f *ActionFunctions) readActionCarrierStop(owner int32, r *DataReader) bool {
	panic(notPorted("read_action_carrier_stop"))
}

func (f *ActionFunctions) readActionTrainFighter(owner int32, r *DataReader) bool {
	panic(notPorted("read_action_train_fighter"))
}

func (f *ActionFunctions) readActionMorphDarkArchon(owner int32, r *DataReader) bool {
	panic(notPorted("read_action_morph_dark_archon"))
}

func (f *ActionFunctions) readActionReaverStop(owner int32, r *DataReader) bool {
	panic(notPorted("read_action_reaver_stop"))
}

func (f *ActionFunctions) readActionCheat(owner int32, r *DataReader) bool {
	panic(notPorted("read_action_cheat"))
}

func (f *ActionFunctions) readActionExtCheat(owner int32, r *DataReader) bool {
	panic(notPorted("read_action_ext_cheat"))
}

func (f *ActionFunctions) readAction(r *DataReader) bool {
	panic(notPorted("read_action"))
}

func (f *ActionFunctions) readActionOwner(owner int32, r *DataReader) bool {
	panic(notPorted("read_action"))
}

func (f *ActionFunctions) readActionBytes(data []byte) bool {
	panic(notPorted("read_action"))
}

func (f *ActionFunctions) readActionOwnerBytes(owner int32, data []byte) bool {
	panic(notPorted("read_action"))
}

func (f *ActionFunctions) executeActions(actionsData []byte) {
	panic(notPorted("execute_actions"))
}
