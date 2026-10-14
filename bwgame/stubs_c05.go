package bwgame

import "iter"

func (f *StateFunctions) executeMainOrder(u *Unit) {
	panic(notPorted("execute_main_order"))
}

func (f *StateFunctions) resumeBuildingUnit(u *Unit, placeWhenCompleted bool) bool {
	panic(notPorted("resume_building_unit"))
}

func (f *StateFunctions) rallyUnit(u *Unit, factoryUnit *Unit) {
	panic(notPorted("rally_unit"))
}

func (f *StateFunctions) secondaryOrderTrain(u *Unit) {
	panic(notPorted("secondary_order_Train"))
}

func (f *StateFunctions) secondaryOrderBuildAddon(u *Unit) {
	panic(notPorted("secondary_order_BuildAddon"))
}

func (f *StateFunctions) secondaryOrderCloak(u *Unit) {
	panic(notPorted("secondary_order_Cloak"))
}

func (f *StateFunctions) secondaryOrderDecloak(u *Unit) {
	panic(notPorted("secondary_order_Decloak"))
}

func (f *StateFunctions) secondaryOrderSpawningLarva(u *Unit) {
	panic(notPorted("secondary_order_SpawningLarva"))
}

func (f *StateFunctions) secondaryOrderSpreadCreep(u *Unit) {
	panic(notPorted("secondary_order_SpreadCreep"))
}

func (f *StateFunctions) secondaryOrderCloakNearbyUnits(u *Unit) {
	panic(notPorted("secondary_order_CloakNearbyUnits"))
}

func (f *StateFunctions) secondaryOrderTrainFighter(u *Unit) {
	panic(notPorted("secondary_order_TrainFighter"))
}

func (f *StateFunctions) secondaryOrderShieldBattery(u *Unit) {
	panic(notPorted("secondary_order_ShieldBattery"))
}

func (f *StateFunctions) executeSecondaryOrder(u *Unit) {
	panic(notPorted("execute_secondary_order"))
}

func (f *StateFunctions) updateUnit(u *Unit) {
	panic(notPorted("update_unit"))
}

func (f *StateFunctions) revealSightAt(pos Xy, rangeValue int32, revealTo int32, inAir bool) {
	panic(notPorted("reveal_sight_at"))
}

func (f *StateFunctions) refreshUnitVision(u *Unit) {
	panic(notPorted("refresh_unit_vision"))
}

func (f *StateFunctions) turnTurret(u *Unit, turn DirectionT) {
	panic(notPorted("turn_turret"))
}

func (f *StateFunctions) updateUnitMovement(u *Unit) {
	panic(notPorted("update_unit_movement"))
}

func (f *StateFunctions) updateThingyVisibility(t *Thingy, size Xy) bool {
	panic(notPorted("update_thingy_visibility"))
}

func (f *StateFunctions) updateUnitSprite(u *Unit) {
	panic(notPorted("update_unit_sprite"))
}

func (f *StateFunctions) unitCanFireFromBunker(u *Unit) bool {
	panic(notPorted("unit_can_fire_from_bunker"))
}

func (f *StateFunctions) createBunkerFireAnimation(u *Unit) {
	panic(notPorted("create_bunker_fire_animation"))
}

func (f *StateFunctions) tryAttackSomething(u *Unit) bool {
	panic(notPorted("try_attack_something"))
}

func (f *StateFunctions) orderHiddenBunkerGuard(u *Unit) {
	panic(notPorted("order_hidden_BunkerGuard"))
}

func (f *StateFunctions) executeHiddenUnitMainOrder(u *Unit) {
	panic(notPorted("execute_hidden_unit_main_order"))
}

func (f *StateFunctions) executeHiddenUnitSecondaryOrder(u *Unit) {
	panic(notPorted("execute_hidden_unit_secondary_order"))
}

func (f *StateFunctions) updateHiddenUnit(u *Unit) {
	panic(notPorted("update_hidden_unit"))
}

func (f *StateFunctions) unitCanDetect(u *Unit) bool {
	panic(notPorted("unit_can_detect"))
}

func (f *StateFunctions) unitCalculateDetectedFlags(u *Unit) uint32 {
	panic(notPorted("unit_calculate_detected_flags"))
}

func (f *StateFunctions) removeTargetReferences(u *Unit, target *Unit) {
	panic(notPorted("remove_target_references"))
}

func (f *StateFunctions) removeBulletTargetReferences(b *Bullet, target *Unit) {
	panic(notPorted("remove_target_references"))
}

func (f *StateFunctions) updateUnitDetectedFlags(u *Unit) {
	panic(notPorted("update_unit_detected_flags"))
}

func (f *StateFunctions) updatePsionicMatrix() {
	panic(notPorted("update_psionic_matrix"))
}

func (f *StateFunctions) updateDeadUnit(u *Unit) {
	panic(notPorted("update_dead_unit"))
}

func (f *StateFunctions) updateDisruptionWeb() {
	panic(notPorted("update_disruption_web"))
}

func (f *StateFunctions) updateUnits() {
	panic(notPorted("update_units"))
}

func (f *StateFunctions) updateBullets() {
	panic(notPorted("update_bullets"))
}

func (f *StateFunctions) spriteIsDoodad(spriteType *SpriteTypeT) bool {
	panic(notPorted("sprite_is_doodad"))
}

func (f *StateFunctions) updateThingy(t *Thingy) {
	panic(notPorted("update_thingy"))
}

func (f *StateFunctions) updateThingies() {
	panic(notPorted("update_thingies"))
}

func (f *StateFunctions) activePlayers() iter.Seq[int32] {
	panic(notPorted("active_players"))
}

func (f *StateFunctions) playerLeft(owner int32) {
	panic(notPorted("player_left"))
}

func (f *StateFunctions) playerDropped(owner int32) {
	panic(notPorted("player_dropped"))
}

func (f *StateFunctions) removePlayer(owner int32) {
	panic(notPorted("remove_player"))
}

func (f *StateFunctions) processFrame() {
	panic(notPorted("process_frame"))
}

func (f *StateFunctions) nextFrame() {
	panic(notPorted("next_frame"))
}

func (f *StateFunctions) localUnitStatusError(u *Unit, err int32) {
	panic(notPorted("local_unit_status_error"))
}
