package bwgame

func (f *StateFunctions) setRemoveTimer(u *Unit, timer int32) {
	panic(notPorted("set_remove_timer"))
}

func (f *StateFunctions) defaultRemoveTimer(u *Unit) int32 {
	panic(notPorted("default_remove_timer"))
}

func (f *StateFunctions) setDefaultRemoveTimer(u *Unit) {
	panic(notPorted("set_remove_timer"))
}

func (f *StateFunctions) makeUnitHallucination(u *Unit) {
	panic(notPorted("make_unit_hallucination"))
}

func (f *StateFunctions) unitCanSeeOrderTarget(u *Unit) bool {
	panic(notPorted("unit_can_see_order_target"))
}

func (f *StateFunctions) medicCanHealTarget(u *Unit, target *Unit) bool {
	panic(notPorted("medic_can_heal_target"))
}

func (f *StateFunctions) findMedicTarget(u *Unit) *Unit {
	panic(notPorted("find_medic_target"))
}

func (f *StateFunctions) medicTryHeal(u *Unit) int32 {
	panic(notPorted("medic_try_heal"))
}

func (f *StateFunctions) orderSelfDestructing(u *Unit) {
	panic(notPorted("order_SelfDestructing"))
}

func (f *StateFunctions) copyStatusEffects(to *Unit, from *Unit) {
	panic(notPorted("copy_status_effects"))
}

func (f *StateFunctions) mergeStatusEffects(to *Unit, from *Unit) {
	panic(notPorted("merge_status_effects"))
}

func (f *StateFunctions) refreshUnitPosition(u *Unit) {
	panic(notPorted("refresh_unit_position"))
}

func (f *StateFunctions) unitsShareUnions(a *UnitTypeT, b *UnitTypeT) bool {
	panic(notPorted("units_share_unions"))
}

func (f *StateFunctions) interceptorsAttack(u *Unit) {
	panic(notPorted("interceptors_attack"))
}

func (f *StateFunctions) carrierReaverAttack(u *Unit, acquireRange int32, maxGroundDistance int32) bool {
	panic(notPorted("carrier_reaver_attack"))
}

func (f *StateFunctions) releaseFighter(u *Unit) *Unit {
	panic(notPorted("release_fighter"))
}

func (f *StateFunctions) createHallucination(sourceUnit *Unit, owner int32) *Unit {
	panic(notPorted("create_hallucination"))
}

func (f *StateFunctions) orderStop(u *Unit) {
	panic(notPorted("order_Stop"))
}

func (f *StateFunctions) orderGuard(u *Unit) {
	panic(notPorted("order_Guard"))
}

func (f *StateFunctions) orderPlayerGuard(u *Unit) {
	panic(notPorted("order_PlayerGuard"))
}

func (f *StateFunctions) orderTurretGuard(u *Unit) {
	panic(notPorted("order_TurretGuard"))
}

func (f *StateFunctions) orderMove(u *Unit) {
	panic(notPorted("order_Move"))
}

func (f *StateFunctions) orderNothing(u *Unit) {
	panic(notPorted("order_Nothing"))
}

func (f *StateFunctions) orderMoveToMinerals(u *Unit) {
	panic(notPorted("order_MoveToMinerals"))
}

func (f *StateFunctions) orderWaitForMinerals(u *Unit) {
	panic(notPorted("order_WaitForMinerals"))
}

func (f *StateFunctions) orderMiningMinerals(u *Unit) {
	panic(notPorted("order_MiningMinerals"))
}

func (f *StateFunctions) orderResetHarvestCollision(u *Unit) {
	panic(notPorted("order_ResetHarvestCollision"))
}

func (f *StateFunctions) orderReturnGas(u *Unit) {
	panic(notPorted("order_ReturnGas"))
}

func (f *StateFunctions) orderReturnMinerals(u *Unit) {
	panic(notPorted("order_ReturnMinerals"))
}

func (f *StateFunctions) orderDie(u *Unit) {
	panic(notPorted("order_Die"))
}

func (f *StateFunctions) orderGatherWaitInterrupted(u *Unit) {
	panic(notPorted("order_GatherWaitInterrupted"))
}

func (f *StateFunctions) orderGatheringInterrupted(u *Unit) {
	panic(notPorted("order_GatheringInterrupted"))
}

func (f *StateFunctions) orderPlaceBuilding(u *Unit) {
	panic(notPorted("order_PlaceBuilding"))
}

func (f *StateFunctions) orderIncompleteBuilding(u *Unit) {
	panic(notPorted("order_IncompleteBuilding"))
}

func (f *StateFunctions) orderConstructingBuilding(u *Unit) {
	panic(notPorted("order_ConstructingBuilding"))
}

func (f *StateFunctions) orderResetCollision(u *Unit) {
	panic(notPorted("order_ResetCollision"))
}

func (f *StateFunctions) orderFollow(u *Unit) {
	panic(notPorted("order_Follow"))
}

func (f *StateFunctions) orderAttackMove(u *Unit) {
	panic(notPorted("order_AttackMove"))
}

func (f *StateFunctions) orderAttackUnit(u *Unit) {
	panic(notPorted("order_AttackUnit"))
}

func (f *StateFunctions) orderMoveToGas(u *Unit) {
	panic(notPorted("order_MoveToGas"))
}

func (f *StateFunctions) orderWaitForGas(u *Unit) {
	panic(notPorted("order_WaitForGas"))
}

func (f *StateFunctions) orderHarvestGas(u *Unit) {
	panic(notPorted("order_HarvestGas"))
}

func (f *StateFunctions) orderRepair(u *Unit) {
	panic(notPorted("order_Repair"))
}

func (f *StateFunctions) orderEnterTransport(u *Unit) {
	panic(notPorted("order_EnterTransport"))
}

func (f *StateFunctions) orderUnload(u *Unit) {
	panic(notPorted("order_Unload"))
}

func (f *StateFunctions) orderBuildingLiftoff(u *Unit) {
	panic(notPorted("order_BuildingLiftoff"))
}

func (f *StateFunctions) orderLiftingOff(u *Unit) {
	panic(notPorted("order_LiftingOff"))
}

func (f *StateFunctions) orderPlaceAddon(u *Unit) {
	panic(notPorted("order_PlaceAddon"))
}

func (f *StateFunctions) orderBuildingLand(u *Unit) {
	panic(notPorted("order_BuildingLand"))
}

func (f *StateFunctions) orderResearchTech(u *Unit) {
	panic(notPorted("order_ResearchTech"))
}

func (f *StateFunctions) orderUpgrade(u *Unit) {
	panic(notPorted("order_Upgrade"))
}

func (f *StateFunctions) orderStayInRange(u *Unit) {
	panic(notPorted("order_StayInRange"))
}

func (f *StateFunctions) orderTurretAttack(u *Unit) {
	panic(notPorted("order_TurretAttack"))
}

func (f *StateFunctions) orderSieging(u *Unit) {
	panic(notPorted("order_Sieging"))
}

func (f *StateFunctions) orderUnsieging(u *Unit) {
	panic(notPorted("order_Unsieging"))
}

func (f *StateFunctions) orderWatchTarget(u *Unit) {
	panic(notPorted("order_WatchTarget"))
}

func (f *StateFunctions) orderPickupIdle(u *Unit) {
	panic(notPorted("order_PickupIdle"))
}

func (f *StateFunctions) orderPickupTransport(u *Unit) {
	panic(notPorted("order_PickupTransport"))
}

func (f *StateFunctions) orderMoveUnload(u *Unit) {
	panic(notPorted("order_MoveUnload"))
}

func (f *StateFunctions) orderPlaceMine(u *Unit) {
	panic(notPorted("order_PlaceMine"))
}

func (f *StateFunctions) orderSpiderMine(u *Unit) {
	panic(notPorted("order_SpiderMine"))
}

func (f *StateFunctions) orderCritter(u *Unit) {
	panic(notPorted("order_Critter"))
}

func (f *StateFunctions) orderHoldPosition(u *Unit) {
	panic(notPorted("order_HoldPosition"))
}
