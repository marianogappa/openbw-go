package bwgame

func (f *StateFunctions) unitIsAtMoveTarget(u *Flingy) bool {
	panic(notPorted("unit_is_at_move_target"))
}

func (f *StateFunctions) unitDead(u *Unit) bool {
	panic(notPorted("unit_dead"))
}

func (f *StateFunctions) unitDying(u *Unit) bool {
	panic(notPorted("unit_dying"))
}

func (f *StateFunctions) rectangleCanFitAt(pos Xy, width int32, height int32) bool {
	panic(notPorted("rectangle_can_fit_at"))
}

func (f *StateFunctions) unitTypeCanFitAt(unitType *UnitTypeT, pos Xy) bool {
	panic(notPorted("unit_type_can_fit_at"))
}

func (f *StateFunctions) canFitAt(pos Xy, inner [4]int32) bool {
	panic(notPorted("can_fit_at"))
}

func (f *StateFunctions) unitTargetIsEnemy(u *Unit, target *Unit) bool {
	panic(notPorted("unit_target_is_enemy"))
}

func (f *StateFunctions) getLargestBlockingUnit(u *Unit, bounds Rect) *Unit {
	panic(notPorted("get_largest_blocking_unit"))
}

func (f *StateFunctions) isBlocked(u *Unit, pos Xy) (bool, *Unit) {
	panic(notPorted("is_blocked"))
}

func (f *StateFunctions) setFlingyMoveTarget(u *Flingy, targetPos Xy, targetUnit *Unit) {
	panic(notPorted("set_flingy_move_target"))
}

func (f *StateFunctions) setUnitMoveTarget(u *Unit, moveTarget Xy) {
	panic(notPorted("set_unit_move_target"))
}

func (f *StateFunctions) setUnitMoveTargetUnit(u *Unit, target *Unit) {
	panic(notPorted("set_unit_move_target"))
}

func (f *StateFunctions) setCurrentVelocityDirection(u *Flingy, currentVelocityDirection DirectionT) {
	panic(notPorted("set_current_velocity_direction"))
}

func (f *StateFunctions) setCurrentSpeed(u *Flingy, currentSpeed Fp8) {
	panic(notPorted("set_current_speed"))
}

func (f *StateFunctions) setNextSpeed(u *Flingy, nextSpeed Fp8) {
	panic(notPorted("set_next_speed"))
}

func (f *StateFunctions) unitTurnRate(u *Flingy, desiredTurn DirectionT) DirectionT {
	panic(notPorted("unit_turn_rate"))
}

func (f *StateFunctions) setDesiredVelocityDirection(u *Flingy, desiredVelocityDirection DirectionT) {
	panic(notPorted("set_desired_velocity_direction"))
}

func (f *StateFunctions) updateCurrentVelocityDirectionTowardsWaypoint(u *Flingy) {
	panic(notPorted("update_current_velocity_direction_towards_waypoint"))
}

func (f *StateFunctions) updateUnitHeading(u *Flingy, velocityDirection DirectionT) {
	panic(notPorted("update_unit_heading"))
}

func (f *StateFunctions) updateCurrentSpeedTowardsWaypoint(u *Flingy) {
	panic(notPorted("update_current_speed_towards_waypoint"))
}

func (f *StateFunctions) setMovementFlags(u *Flingy, ems *ExecuteMovementStruct) {
	panic(notPorted("set_movement_flags"))
}

func (f *StateFunctions) setMovementValues(u *Flingy, ems *ExecuteMovementStruct) {
	panic(notPorted("set_movement_values"))
}

func (f *StateFunctions) updateUnitMovementValues(u *Flingy, ems *ExecuteMovementStruct) {
	panic(notPorted("update_unit_movement_values"))
}

func (f *StateFunctions) checkUnitMovementTerrainCollision(u *Unit, movement Xy) bool {
	panic(notPorted("check_unit_movement_terrain_collision"))
}

func (f *StateFunctions) checkUnitMovementTerrainCollisionEMS(u *Unit, ems *ExecuteMovementStruct) bool {
	panic(notPorted("check_unit_movement_terrain_collision"))
}

func (f *StateFunctions) checkUnitMovementUnitCollision(u *Unit, ems *ExecuteMovementStruct) *Unit {
	panic(notPorted("check_unit_movement_unit_collision"))
}

func (f *StateFunctions) checkGroundUnitMovementUnitCollision(u *Unit, ems *ExecuteMovementStruct) *Unit {
	panic(notPorted("check_ground_unit_movement_unit_collision"))
}

func (f *StateFunctions) updateRepulseDirection(u *Unit) {
	panic(notPorted("update_repulse_direction"))
}

func (f *StateFunctions) applyRepulseField(u *Unit, ems *ExecuteMovementStruct) bool {
	panic(notPorted("apply_repulse_field"))
}

func (f *StateFunctions) finishFlingyMovement(u *Flingy, ems *ExecuteMovementStruct) bool {
	panic(notPorted("finish_flingy_movement"))
}

func (f *StateFunctions) finishUnitMovement(u *Unit, ems *ExecuteMovementStruct) bool {
	panic(notPorted("finish_unit_movement"))
}

func (f *StateFunctions) isMovingAlongPath(u *Unit) bool {
	panic(notPorted("is_moving_along_path"))
}

func (f *StateFunctions) setUnitImmovable(u *Unit) {
	panic(notPorted("set_unit_immovable"))
}

func (f *StateFunctions) updateUnitPathingCollision(u *Unit) {
	panic(notPorted("update_unit_pathing_collision"))
}

func (f *StateFunctions) cardinalDirectionFromTo(fromU *Unit, toU *Unit) int32 {
	panic(notPorted("cardinal_direction_from_to"))
}

func (f *StateFunctions) collisionGetSlideFreeDirection(u *Unit, collisionUnit *Unit, slideFreeDirection *DirectionT) bool {
	panic(notPorted("collision_get_slide_free_direction"))
}

func (f *StateFunctions) resetMovementState(u *Unit) {
	panic(notPorted("reset_movement_state"))
}

func (f *StateFunctions) movementUMInit(u *Unit, ems *ExecuteMovementStruct) bool {
	panic(notPorted("movement_UM_Init"))
}

func (f *StateFunctions) movementUMAtRest(u *Unit, ems *ExecuteMovementStruct) bool {
	panic(notPorted("movement_UM_AtRest"))
}

func (f *StateFunctions) getUniqueSidedPositionsWithinBounds(from *Xy, to *Xy, bounds Rect) bool {
	panic(notPorted("get_unique_sided_positions_within_bounds"))
}

func (f *StateFunctions) movementUMCheckIllegal(u *Unit, ems *ExecuteMovementStruct) bool {
	panic(notPorted("movement_UM_CheckIllegal"))
}

func (f *StateFunctions) movementUMDormant(u *Unit, ems *ExecuteMovementStruct) bool {
	panic(notPorted("movement_UM_Dormant"))
}

func (f *StateFunctions) movementUMTurret(u *Unit, ems *ExecuteMovementStruct) bool {
	panic(notPorted("movement_UM_Turret"))
}

func (f *StateFunctions) movementUMFlyer(u *Unit, ems *ExecuteMovementStruct) bool {
	panic(notPorted("movement_UM_Flyer"))
}

func (f *StateFunctions) movementUMAnotherPath(u *Unit, ems *ExecuteMovementStruct) bool {
	panic(notPorted("movement_UM_AnotherPath"))
}

func (f *StateFunctions) movementUMStartPath(u *Unit, ems *ExecuteMovementStruct) bool {
	panic(notPorted("movement_UM_StartPath"))
}

func (f *StateFunctions) movementUMFollowPath(u *Unit, ems *ExecuteMovementStruct) bool {
	panic(notPorted("movement_UM_FollowPath"))
}

func (f *StateFunctions) movementUMScoutPath(u *Unit, ems *ExecuteMovementStruct) bool {
	panic(notPorted("movement_UM_ScoutPath"))
}

func (f *StateFunctions) movementUMAtMoveTarget(u *Unit, ems *ExecuteMovementStruct) bool {
	panic(notPorted("movement_UM_AtMoveTarget"))
}

func (f *StateFunctions) movementUMNewMoveTarget(u *Unit, ems *ExecuteMovementStruct) bool {
	panic(notPorted("movement_UM_NewMoveTarget"))
}

func (f *StateFunctions) movementUMRepath(u *Unit, ems *ExecuteMovementStruct) bool {
	panic(notPorted("movement_UM_Repath"))
}

func (f *StateFunctions) movementUMUIOrderDelay(u *Unit, ems *ExecuteMovementStruct) bool {
	panic(notPorted("movement_UM_UIOrderDelay"))
}

func (f *StateFunctions) movementUMFixTerrain(u *Unit, ems *ExecuteMovementStruct) bool {
	panic(notPorted("movement_UM_FixTerrain"))
}

func (f *StateFunctions) movementUMFailedPath(u *Unit, ems *ExecuteMovementStruct) bool {
	panic(notPorted("movement_UM_FailedPath"))
}

func (f *StateFunctions) movementUMTerrainSlide(u *Unit, ems *ExecuteMovementStruct) bool {
	panic(notPorted("movement_UM_TerrainSlide"))
}

func (f *StateFunctions) movementUMTurnAndStart(u *Unit, ems *ExecuteMovementStruct) bool {
	panic(notPorted("movement_UM_TurnAndStart"))
}

func (f *StateFunctions) movementUMFaceTarget(u *Unit, ems *ExecuteMovementStruct) bool {
	panic(notPorted("movement_UM_FaceTarget"))
}

func (f *StateFunctions) movementUMRetryPath(u *Unit, ems *ExecuteMovementStruct) bool {
	panic(notPorted("movement_UM_RetryPath"))
}

func (f *StateFunctions) movementUMInitSeq(u *Unit, ems *ExecuteMovementStruct) bool {
	panic(notPorted("movement_UM_InitSeq"))
}

func (f *StateFunctions) movementUMForceMoveFree(u *Unit, ems *ExecuteMovementStruct) bool {
	panic(notPorted("movement_UM_ForceMoveFree"))
}

func (f *StateFunctions) movementUMFixCollision(u *Unit, ems *ExecuteMovementStruct) bool {
	panic(notPorted("movement_UM_FixCollision"))
}

func (f *StateFunctions) movementUMWaitFree(u *Unit, ems *ExecuteMovementStruct) bool {
	panic(notPorted("movement_UM_WaitFree"))
}

func (f *StateFunctions) movementUMSlidePrep(u *Unit, ems *ExecuteMovementStruct) bool {
	panic(notPorted("movement_UM_SlidePrep"))
}

func (f *StateFunctions) movementUMSlideFree(u *Unit, ems *ExecuteMovementStruct) bool {
	panic(notPorted("movement_UM_SlideFree"))
}

func (f *StateFunctions) movementUMRepathMovers(u *Unit, ems *ExecuteMovementStruct) bool {
	panic(notPorted("movement_UM_RepathMovers"))
}

func (f *StateFunctions) movementUMGetFree(u *Unit, ems *ExecuteMovementStruct) bool {
	panic(notPorted("movement_UM_GetFree"))
}

func (f *StateFunctions) movementUMLump(u *Unit, ems *ExecuteMovementStruct) bool {
	panic(notPorted("movement_UM_Lump"))
}

func (f *StateFunctions) movementUMMoveToLegal(u *Unit, ems *ExecuteMovementStruct) bool {
	panic(notPorted("movement_UM_MoveToLegal"))
}

func (f *StateFunctions) movementUMHidden(u *Unit, ems *ExecuteMovementStruct) bool {
	panic(notPorted("movement_UM_Hidden"))
}

func (f *StateFunctions) movementUMBunker(u *Unit, ems *ExecuteMovementStruct) bool {
	panic(notPorted("movement_UM_Bunker"))
}

func (f *StateFunctions) movementUMLumpWannabe(u *Unit, ems *ExecuteMovementStruct) bool {
	panic(notPorted("movement_UM_LumpWannabe"))
}

func (f *StateFunctions) movementUMBldgTurret(u *Unit, ems *ExecuteMovementStruct) bool {
	panic(notPorted("movement_UM_BldgTurret"))
}

func (f *StateFunctions) executeMovement(u *Unit) bool {
	panic(notPorted("execute_movement"))
}

func (f *StateFunctions) getModifiedUnitSpeed(u *Unit, baseSpeed Fp8) Fp8 {
	panic(notPorted("get_modified_unit_speed"))
}

func (f *StateFunctions) getModifiedUnitAcceleration(u *Unit, baseAcceleration Fp8) Fp8 {
	panic(notPorted("get_modified_unit_acceleration"))
}

func (f *StateFunctions) getModifiedUnitTurnRate(u *Unit, baseTurnRate Fp8) Fp8 {
	panic(notPorted("get_modified_unit_turn_rate"))
}

func (f *StateFunctions) getModifiedWeaponCooldown(u *Unit, weapon *WeaponTypeT) int32 {
	panic(notPorted("get_modified_weapon_cooldown"))
}

func (f *StateFunctions) unitHaltDistance(u *Flingy) Fp8 {
	panic(notPorted("unit_halt_distance"))
}
