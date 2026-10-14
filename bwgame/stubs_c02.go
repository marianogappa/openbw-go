package bwgame

func (f *StateFunctions) stopFlingy(u *Flingy) {
	panic(notPorted("stop_flingy"))
}

func (f *StateFunctions) stopUnit(u *Unit) {
	panic(notPorted("stop_unit"))
}

func (f *StateFunctions) orderDone(u *Unit, target OrderTarget) {
	panic(notPorted("order_done"))
}

func (f *StateFunctions) setNextTargetWaypoint(u *Flingy, pos Xy) {
	panic(notPorted("set_next_target_waypoint"))
}

func (f *StateFunctions) setUnitGathering(u *Unit) {
	panic(notPorted("set_unit_gathering"))
}

func (f *StateFunctions) anyNeighborTileUnoccupied(u *Unit) bool {
	panic(notPorted("any_neighbor_tile_unoccupied"))
}

func (f *StateFunctions) unitPosIsBorderingTarget(u *Unit, pos Xy, target *Unit) bool {
	panic(notPorted("unit_pos_is_bordering_target"))
}

func (f *StateFunctions) moveToTarget(u *Unit, target *Unit) {
	panic(notPorted("move_to_target"))
}

func (f *StateFunctions) moveToTargetReset(u *Unit, target *Unit) {
	panic(notPorted("move_to_target_reset"))
}

func (f *StateFunctions) tryGatherResource(u *Unit, target *Unit) bool {
	panic(notPorted("try_gather_resource"))
}

func (f *StateFunctions) waitForResource(u *Unit, target *Unit) {
	panic(notPorted("wait_for_resource"))
}

func (f *StateFunctions) isFacingNextTargetWaypoint(u *Unit, angle Fp8) bool {
	panic(notPorted("is_facing_next_target_waypoint"))
}

func (f *StateFunctions) longPathDistance(from Xy, to Xy) int32 {
	panic(notPorted("long_path_distance"))
}

func (f *StateFunctions) unitLongPathDistance(u *Unit, from Xy, to Xy) int32 {
	panic(notPorted("unit_long_path_distance"))
}

func (f *StateFunctions) destroyCarryingImages(u *Unit) {
	panic(notPorted("destroy_carrying_images"))
}

func (f *StateFunctions) unitGatherResourcesFrom(u *Unit, resource *Unit) {
	panic(notPorted("unit_gather_resources_from"))
}

func (f *StateFunctions) gatherQueueNext(u *Unit, resource *Unit) {
	panic(notPorted("gather_queue_next"))
}

func (f *StateFunctions) unitIsActiveResourceDepot(u *Unit) bool {
	panic(notPorted("unit_is_active_resource_depot"))
}

func (f *StateFunctions) findNearestActiveResourceDepot(u *Unit) *Unit {
	panic(notPorted("find_nearest_active_resource_depot"))
}

func (f *StateFunctions) dropCarriedItems(u *Unit) {
	panic(notPorted("drop_carried_items"))
}

func (f *StateFunctions) hasAvailableSupplyFor(owner int32, unitType *UnitTypeT, showError bool) bool {
	panic(notPorted("has_available_supply_for"))
}

func (f *StateFunctions) hasAvailableResourcesFor(owner int32, unitType *UnitTypeT, showError bool) bool {
	panic(notPorted("has_available_resources_for"))
}

func (f *StateFunctions) hasAvailableResourcesForTech(owner int32, techType *TechTypeT, showError bool) bool {
	panic(notPorted("has_available_resources_for"))
}

func (f *StateFunctions) upgradeMineralCost(owner int32, upgradeType *UpgradeTypeT) int32 {
	panic(notPorted("upgrade_mineral_cost"))
}

func (f *StateFunctions) upgradeGasCost(owner int32, upgradeType *UpgradeTypeT) int32 {
	panic(notPorted("upgrade_gas_cost"))
}

func (f *StateFunctions) upgradeTimeCost(owner int32, upgradeType *UpgradeTypeT) int32 {
	panic(notPorted("upgrade_time_cost"))
}

func (f *StateFunctions) hasAvailableResourcesForUpgrade(owner int32, upgradeType *UpgradeTypeT, showError bool) bool {
	panic(notPorted("has_available_resources_for"))
}

func (f *StateFunctions) buildQueuePush(u *Unit, unitType *UnitTypeT) bool {
	panic(notPorted("build_queue_push"))
}

func (f *StateFunctions) buildQueueClearSingle(u *Unit) {
	panic(notPorted("build_queue_clear_single"))
}

func (f *StateFunctions) trainUnit(u *Unit, unitType *UnitTypeT) bool {
	panic(notPorted("train_unit"))
}

func (f *StateFunctions) refundUnitCosts(owner int32, unitType *UnitTypeT) {
	panic(notPorted("refund_unit_costs"))
}

func (f *StateFunctions) partiallyRefundUnitCosts(owner int32, unitType *UnitTypeT) {
	panic(notPorted("partially_refund_unit_costs"))
}

func (f *StateFunctions) cancelMorphingBuilding(u *Unit) {
	panic(notPorted("cancel_morphing_building"))
}

func (f *StateFunctions) cancelBuildingUnit(u *Unit) bool {
	panic(notPorted("cancel_building_unit"))
}

func (f *StateFunctions) cancelBuildQueueIndex(u *Unit, index uint) {
	panic(notPorted("cancel_build_queue"))
}

func (f *StateFunctions) cancelBuildQueue(u *Unit) {
	panic(notPorted("cancel_build_queue"))
}

func (f *StateFunctions) placeBuilding(u *Unit, orderType *OrderTypeT, unitType *UnitTypeT, position Xy) {
	panic(notPorted("place_building"))
}

func (f *StateFunctions) playerHasSupplyAndResourcesFor(owner int32, unitType *UnitTypeT, showError bool) bool {
	panic(notPorted("player_has_supply_and_resources_for"))
}

func (f *StateFunctions) getBuildingAtCenterPosition(pos Xy, unitType *UnitTypeT) *Unit {
	panic(notPorted("get_building_at_center_position"))
}

func (f *StateFunctions) isInPsionicMatrixRange(rel Xy) bool {
	panic(notPorted("is_in_psionic_matrix_range"))
}

func (f *StateFunctions) isInPsionicMatrix(owner int32, pos Xy) bool {
	panic(notPorted("is_in_psionic_matrix"))
}

func (f *StateFunctions) canPlaceBuilding(u *Unit, owner int32, unitType *UnitTypeT, pos Xy, checkUndetectedUnits bool, checkInvisibleTiles bool) bool {
	panic(notPorted("can_place_building"))
}

func (f *StateFunctions) killUnit(u *Unit) {
	panic(notPorted("kill_unit"))
}

func (f *StateFunctions) unitCanEnterNydus(u *Unit, target *Unit) bool {
	panic(notPorted("unit_can_enter_nydus"))
}

func (f *StateFunctions) unitCanUseShieldBattery(u *Unit, target *Unit) bool {
	panic(notPorted("unit_can_use_shield_battery"))
}

func (f *StateFunctions) unitIsRescuable(u *Unit) bool {
	panic(notPorted("unit_is_rescuable"))
}

func (f *StateFunctions) tryFollowUnit(u *Unit, target *Unit) bool {
	panic(notPorted("try_follow_unit"))
}

func (f *StateFunctions) followAheadOfUnit(u *Unit, target *Unit) {
	panic(notPorted("follow_ahead_of_unit"))
}

func (f *StateFunctions) autoAttackTarget(u *Unit, target *Unit) {
	panic(notPorted("auto_attack_target"))
}

func (f *StateFunctions) attackMoveMovement(u *Unit, stopBeforeAttacking bool) bool {
	panic(notPorted("attack_move_movement"))
}

func (f *StateFunctions) attackMoveAcquireTarget(u *Unit, stopBeforeAttacking bool) {
	panic(notPorted("attack_move_acquire_target"))
}

func (f *StateFunctions) attackUnitReacquireTarget(u *Unit) {
	panic(notPorted("attack_unit_reacquire_target"))
}

func (f *StateFunctions) attackUnitMoveInRange(u *Unit) bool {
	panic(notPorted("attack_unit_move_in_range"))
}

func (f *StateFunctions) unitCanFireWeapon(u *Unit, weapon *WeaponTypeT) bool {
	panic(notPorted("unit_can_fire_weapon"))
}

func (f *StateFunctions) attackUnitFireWeapon(u *Unit) {
	panic(notPorted("attack_unit_fire_weapon"))
}

func (f *StateFunctions) setUnitOwner(u *Unit, owner int32, incrementScore bool) {
	panic(notPorted("set_unit_owner"))
}

func (f *StateFunctions) setSpriteOwner(u *Unit, owner int32) {
	panic(notPorted("set_sprite_owner"))
}

func (f *StateFunctions) reinitializeUnitType(u *Unit, unitType *UnitTypeT) {
	panic(notPorted("reinitialize_unit_type"))
}

func (f *StateFunctions) morphUnit(u *Unit, unitType *UnitTypeT) {
	panic(notPorted("morph_unit"))
}

func (f *StateFunctions) buildRefinery(u *Unit, unitType *UnitTypeT) *Unit {
	panic(notPorted("build_refinery"))
}

func (f *StateFunctions) unitCanGatherGas(u *Unit, target *Unit) bool {
	panic(notPorted("unit_can_gather_gas"))
}

func (f *StateFunctions) findGasExitPosition(u *Unit, gas *Unit, pos *Xy, heading *DirectionT) bool {
	panic(notPorted("find_gas_exit_position"))
}

func (f *StateFunctions) isAtNextTargetWaypointOrWithinAttackAngle(u *Unit) bool {
	panic(notPorted("is_at_next_target_waypoint_or_within_attack_angle"))
}

func (f *StateFunctions) unitLoadTarget(u *Unit, target *Unit) {
	panic(notPorted("unit_load_target"))
}

func (f *StateFunctions) unitUnloadImpl(u *Unit, containerDestroyed bool) {
	panic(notPorted("unit_unload_impl"))
}

func (f *StateFunctions) unitUnload(u *Unit) bool {
	panic(notPorted("unit_unload"))
}

func (f *StateFunctions) triggerGiveUnitTo(u *Unit, newOwner int32) bool {
	panic(notPorted("trigger_give_unit_to"))
}

func (f *StateFunctions) giveUnitTo(u *Unit, newOwner int32) {
	panic(notPorted("give_unit_to"))
}

func (f *StateFunctions) makeUnitNeutral(u *Unit) {
	panic(notPorted("make_unit_neutral"))
}

func (f *StateFunctions) buildingAbandonAddon(u *Unit) {
	panic(notPorted("building_abandon_addon"))
}

func (f *StateFunctions) createDustSprite(sprite *Sprite, spriteID SpriteTypes, loIndex uint, offsetFrom uint, offsetTo uint, flipped bool) {
	panic(notPorted("create_dust_sprite"))
}

func (f *StateFunctions) unitCanAttachAddon(u *Unit, addon *Unit) bool {
	panic(notPorted("unit_can_attach_addon"))
}

func (f *StateFunctions) findConnectingAddon(u *Unit) *Unit {
	panic(notPorted("find_connecting_addon"))
}

func (f *StateFunctions) connectAddon(u *Unit, addon *Unit) {
	panic(notPorted("connect_addon"))
}

func (f *StateFunctions) findAndConnectAddon(u *Unit) {
	panic(notPorted("find_and_connect_addon"))
}

func (f *StateFunctions) cancelResearch(u *Unit, unitDestroyed bool) {
	panic(notPorted("cancel_research"))
}

func (f *StateFunctions) cancelUpgrade(u *Unit, unitDestroyed bool) {
	panic(notPorted("cancel_upgrade"))
}

func (f *StateFunctions) applyUpgradesToPlayerUnits(owner int32) {
	panic(notPorted("apply_upgrades_to_player_units"))
}

func (f *StateFunctions) setUnitBurrowed(u *Unit) {
	panic(notPorted("set_unit_burrowed"))
}

func (f *StateFunctions) setUnitCloaked(u *Unit) {
	panic(notPorted("set_unit_cloaked"))
}

func (f *StateFunctions) crappyMoveToTarget(u *Unit, target *Unit) {
	panic(notPorted("crappy_move_to_target"))
}

func (f *StateFunctions) getDefaultGatherOrder(u *Unit, target *Unit) *OrderTypeT {
	panic(notPorted("get_default_gather_order"))
}

func (f *StateFunctions) getDefaultOrder(action uint, u *Unit, pos Xy, target *Unit, targetUnitType *UnitTypeT) *OrderTypeT {
	panic(notPorted("get_default_order"))
}

func (f *StateFunctions) defaultAction(u *Unit) uint {
	panic(notPorted("default_action"))
}

func (f *StateFunctions) spellCastTargetMovement(u *Unit, tech *TechTypeT, rangeValue int32) bool {
	panic(notPorted("spell_cast_target_movement"))
}
