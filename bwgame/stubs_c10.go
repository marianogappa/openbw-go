package bwgame

import "iter"

func (f *StateFunctions) updateUnitSpeedUpgrades(u *Unit) {
	panic(notPorted("update_unit_speed_upgrades"))
}

func (f *StateFunctions) updateUnitSpeed(u *Unit) {
	panic(notPorted("update_unit_speed"))
}

func (f *StateFunctions) incrementUnitCounts(u *Unit, count int32) {
	panic(notPorted("increment_unit_counts"))
}

func (f *StateFunctions) unitFinderInsert(u *Unit) {
	panic(notPorted("unit_finder_insert"))
}

func (f *StateFunctions) unitFinderReinsert(u *Unit) {
	panic(notPorted("unit_finder_reinsert"))
}

func (f *StateFunctions) unitFinderRemove(u *Unit) {
	panic(notPorted("unit_finder_remove"))
}

func (f *StateFunctions) unitFinderInsertBB(u *Unit, bb Rect) {
	panic(notPorted("unit_finder_insert"))
}

func (f *StateFunctions) unitFinderReinsertBB(u *Unit, bb Rect) {
	panic(notPorted("unit_finder_reinsert"))
}

func (f *StateFunctions) findUnits(area Rect) iter.Seq[*Unit] {
	panic(notPorted("find_units"))
}

func (f *StateFunctions) findUnitsNoexpand(area Rect) iter.Seq[*Unit] {
	panic(notPorted("find_units_noexpand"))
}

func (f *StateFunctions) findUnit(area Rect, predicate func(*Unit) bool) *Unit {
	panic(notPorted("find_unit"))
}

func (f *StateFunctions) findUnitNoexpand(area Rect, predicate func(*Unit) bool) *Unit {
	panic(notPorted("find_unit_noexpand"))
}

func (f *StateFunctions) findNearestUnitFromIndices(pos Xy, searchArea Rect, leftI uint, upI uint, rightI uint, downI uint, predicate func(*Unit) bool) *Unit {
	panic(notPorted("find_nearest_unit"))
}

func (f *StateFunctions) findNearestUnit(pos Xy, searchArea Rect, predicate func(*Unit) bool) *Unit {
	panic(notPorted("find_nearest_unit"))
}

func (f *StateFunctions) findNearestUnitToUnit(u *Unit, searchArea Rect, predicate func(*Unit) bool) *Unit {
	panic(notPorted("find_nearest_unit"))
}

func (f *StateFunctions) unitIsFactory(u *Unit) bool {
	panic(notPorted("unit_is_factory"))
}

func (f *StateFunctions) setUnitTilesOccupied(ut *UnitTypeT, position Xy) {
	panic(notPorted("set_unit_tiles_occupied"))
}

func (f *StateFunctions) setUnitTilesUnoccupied(ut *UnitTypeT, position Xy) {
	panic(notPorted("set_unit_tiles_unoccupied"))
}

func (f *StateFunctions) initializeUnitType(u *Unit, unitType *UnitTypeT, pos Xy, owner int32) bool {
	panic(notPorted("initialize_unit_type"))
}

func (f *StateFunctions) killInterceptors(u *Unit) {
	panic(notPorted("kill_interceptors"))
}

func (f *StateFunctions) destroyUnitImpl(u *Unit) {
	panic(notPorted("destroy_unit_impl"))
}

func (f *StateFunctions) destroyUnit(u *Unit) {
	panic(notPorted("destroy_unit"))
}

func (f *StateFunctions) initializeUnit(u *Unit, unitType *UnitTypeT, pos Xy, owner int32) bool {
	panic(notPorted("initialize_unit"))
}

func (f *StateFunctions) createUnit(unitType *UnitTypeT, pos Xy, owner int32) *Unit {
	panic(notPorted("create_unit"))
}

func (f *StateFunctions) createUnitByID(unitTypeID UnitTypes, pos Xy, owner int32) *Unit {
	panic(notPorted("create_unit"))
}

func (f *StateFunctions) replaceSpriteImages(sprite *Sprite, newImageType *ImageTypeT, heading DirectionT) {
	panic(notPorted("replace_sprite_images"))
}

func (f *StateFunctions) applyUnitEffects(u *Unit) {
	panic(notPorted("apply_unit_effects"))
}

func (f *StateFunctions) damageOverlayStates(u *Unit) int32 {
	panic(notPorted("damage_overlay_states"))
}

func (f *StateFunctions) setConstructionGraphic(u *Unit, animated bool) {
	panic(notPorted("set_construction_graphic"))
}

func (f *StateFunctions) setUnitHeading(u *Unit, heading DirectionT) {
	panic(notPorted("set_unit_heading"))
}

func (f *StateFunctions) finishBuildingUnit(u *Unit) {
	panic(notPorted("finish_building_unit"))
}

func (f *StateFunctions) findUnitPlacementInBounds(u *Unit, pos Xy, bounds Rect, terrainDisplacesUnit bool) (bool, Xy) {
	panic(notPorted("find_unit_placement"))
}

func (f *StateFunctions) findUnitPlacement(u *Unit, pos Xy, terrainDisplacesUnit bool) (bool, Xy) {
	panic(notPorted("find_unit_placement"))
}

func (f *StateFunctions) moveUnit(u *Unit, pos Xy) {
	panic(notPorted("move_unit"))
}

func (f *StateFunctions) placeCompletedUnit(u *Unit) bool {
	panic(notPorted("place_completed_unit"))
}

func (f *StateFunctions) addCompletedUnit(u *Unit, count int32, incrementScore bool) {
	panic(notPorted("add_completed_unit"))
}

func (f *StateFunctions) removeQueuedOrder(u *Unit, o *Order) {
	panic(notPorted("remove_queued_order"))
}

func (f *StateFunctions) createOrder(orderType *OrderTypeT, target OrderTarget) *Order {
	panic(notPorted("create_order"))
}

func (f *StateFunctions) queueOrder(pos *Order, u *Unit, orderType *OrderTypeT, target OrderTarget) {
	panic(notPorted("queue_order"))
}

func (f *StateFunctions) queueOrderBack(u *Unit, orderType *OrderTypeT, target OrderTarget) {
	panic(notPorted("queue_order_back"))
}

func (f *StateFunctions) queueOrderFront(u *Unit, orderType *OrderTypeT, target OrderTarget) {
	panic(notPorted("queue_order_front"))
}

func (f *StateFunctions) removeOneOrder(u *Unit, orderType *OrderTypeT) {
	panic(notPorted("remove_one_order"))
}

func (f *StateFunctions) setQueuedOrder(u *Unit, interrupt bool, orderType *OrderTypeT, target OrderTarget) {
	panic(notPorted("set_queued_order"))
}

func (f *StateFunctions) issueOrder(u *Unit, queue bool, orderType *OrderTypeT, target OrderTarget) {
	panic(notPorted("issue_order"))
}

func (f *StateFunctions) iscriptRunToIdle(u *Unit) {
	panic(notPorted("iscript_run_to_idle"))
}

func (f *StateFunctions) activateNextOrder(u *Unit) {
	panic(notPorted("activate_next_order"))
}

func (f *StateFunctions) setUnitOrder(u *Unit, orderType *OrderTypeT) {
	panic(notPorted("set_unit_order"))
}

func (f *StateFunctions) setUnitOrderTarget(u *Unit, orderType *OrderTypeT, target OrderTarget) {
	panic(notPorted("set_unit_order"))
}

func (f *StateFunctions) setUnitOrderUnit(u *Unit, orderType *OrderTypeT, targetUnit *Unit) {
	panic(notPorted("set_unit_order"))
}

func (f *StateFunctions) setUnitOrderPos(u *Unit, orderType *OrderTypeT, position Xy) {
	panic(notPorted("set_unit_order"))
}

func (f *StateFunctions) isIntersecting(a Rect, b Rect) bool {
	panic(notPorted("is_intersecting"))
}

func (f *StateFunctions) unitFinderUnitsIntersecting(a *Unit, b *Unit) bool {
	panic(notPorted("unit_finder_units_intersecting"))
}

func (f *StateFunctions) unitFinderUnitInBounds(u *Unit, bounds Rect) bool {
	panic(notPorted("unit_finder_unit_in_bounds"))
}

func (f *StateFunctions) checkUnitCollision(u *Unit) {
	panic(notPorted("check_unit_collision"))
}

func (f *StateFunctions) repulseIndex(pos Xy) uint {
	panic(notPorted("repulse_index"))
}

func (f *StateFunctions) incrementRepulseField(u *Unit) {
	panic(notPorted("increment_repulse_field"))
}

func (f *StateFunctions) decrementRepulseField(u *Unit) {
	panic(notPorted("decrement_repulse_field"))
}

func (f *StateFunctions) showUnit(u *Unit) {
	panic(notPorted("show_unit"))
}

func (f *StateFunctions) hideUnit(u *Unit, deselect bool) {
	panic(notPorted("hide_unit"))
}

func (f *StateFunctions) setSpriteCloakModifier(sprite *Sprite, requiresDetector bool, cloaked bool, burrowed bool, data1 int32, data2 int32) {
	panic(notPorted("set_sprite_cloak_modifier"))
}

func (f *StateFunctions) cloakUnit(u *Unit) {
	panic(notPorted("cloak_unit"))
}

func (f *StateFunctions) decloakUnit(u *Unit) {
	panic(notPorted("decloak_unit"))
}

func (f *StateFunctions) completeUnit(u *Unit) {
	panic(notPorted("complete_unit"))
}

func (f *StateFunctions) createInitialUnit(unitType *UnitTypeT, pos Xy, owner int32) *Unit {
	panic(notPorted("create_initial_unit"))
}

func (f *StateFunctions) createCompletedUnit(unitType *UnitTypeT, pos Xy, owner int32) *Unit {
	panic(notPorted("create_completed_unit"))
}
