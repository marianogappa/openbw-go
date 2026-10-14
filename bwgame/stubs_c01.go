package bwgame

import "iter"

func (f *StateFunctions) getOrderType(id Orders) *OrderTypeT {
	panic(notPorted("get_order_type"))
}

func (f *StateFunctions) uSetStatusFlag(u *Unit, flag int32) {
	panic(notPorted("u_set_status_flag"))
}

func (f *StateFunctions) uUnsetStatusFlag(u *Unit, flag int32) {
	panic(notPorted("u_unset_status_flag"))
}

func (f *StateFunctions) uSetStatusFlagValue(u *Unit, flag int32, value bool) {
	panic(notPorted("u_set_status_flag"))
}

func (f *StateFunctions) uSetMovementFlag(u *Flingy, flag int32) {
	panic(notPorted("u_set_movement_flag"))
}

func (f *StateFunctions) uUnsetMovementFlag(u *Flingy, flag int32) {
	panic(notPorted("u_unset_movement_flag"))
}

func (f *StateFunctions) iSetFlag(i *Image, flag int32) {
	panic(notPorted("i_set_flag"))
}

func (f *StateFunctions) iUnsetFlag(i *Image, flag int32) {
	panic(notPorted("i_unset_flag"))
}

func (f *StateFunctions) iSetFlagValue(i *Image, flag int32, value bool) {
	panic(notPorted("i_set_flag"))
}

func (f *StateFunctions) utFlag(ut *UnitTypeT, flag UnitTypeFlagsT) bool {
	panic(notPorted("ut_flag"))
}

func (f *StateFunctions) uStatusFlag(u *Unit, flag int32) bool {
	panic(notPorted("u_status_flag"))
}

func (f *StateFunctions) usFlag(u *Thingy, flag int32) bool {
	panic(notPorted("us_flag"))
}

func (f *StateFunctions) sFlag(s *Sprite, flag int32) bool {
	panic(notPorted("s_flag"))
}

func (f *StateFunctions) uMovementFlag(u *Flingy, flag int32) bool {
	panic(notPorted("u_movement_flag"))
}

func (f *StateFunctions) iFlag(i *Image, flag int32) bool {
	panic(notPorted("i_flag"))
}

func (f *StateFunctions) uCompleted(u *Unit) bool {
	panic(notPorted("u_completed"))
}

func (f *StateFunctions) uInBunker(u *Unit) bool {
	panic(notPorted("u_in_bunker"))
}

func (f *StateFunctions) uLoaded(u *Unit) bool {
	panic(notPorted("u_loaded"))
}

func (f *StateFunctions) uImmovable(u *Unit) bool {
	panic(notPorted("u_immovable"))
}

func (f *StateFunctions) uDisabled(u *Unit) bool {
	panic(notPorted("u_disabled"))
}

func (f *StateFunctions) uBurrowed(u *Unit) bool {
	panic(notPorted("u_burrowed"))
}

func (f *StateFunctions) uCanTurn(u *Unit) bool {
	panic(notPorted("u_can_turn"))
}

func (f *StateFunctions) uCanMove(u *Unit) bool {
	panic(notPorted("u_can_move"))
}

func (f *StateFunctions) uGroundedBuilding(u *Unit) bool {
	panic(notPorted("u_grounded_building"))
}

func (f *StateFunctions) uHallucination(u *Unit) bool {
	panic(notPorted("u_hallucination"))
}

func (f *StateFunctions) uLifetimeExpired(u *Unit) bool {
	panic(notPorted("u_lifetime_expired"))
}

func (f *StateFunctions) uFlying(u *Unit) bool {
	panic(notPorted("u_flying"))
}

func (f *StateFunctions) uSpeedUpgrade(u *Unit) bool {
	panic(notPorted("u_speed_upgrade"))
}

func (f *StateFunctions) uCooldownUpgrade(u *Unit) bool {
	panic(notPorted("u_cooldown_upgrade"))
}

func (f *StateFunctions) uGathering(u *Unit) bool {
	panic(notPorted("u_gathering"))
}

func (f *StateFunctions) uRequiresDetector(u *Unit) bool {
	panic(notPorted("u_requires_detector"))
}

func (f *StateFunctions) uCloaked(u *Unit) bool {
	panic(notPorted("u_cloaked"))
}

func (f *StateFunctions) uCannotAttack(u *Unit) bool {
	panic(notPorted("u_cannot_attack"))
}

func (f *StateFunctions) uOrderNotInterruptible(u *Unit) bool {
	panic(notPorted("u_order_not_interruptible"))
}

func (f *StateFunctions) uIscriptNobrk(u *Unit) bool {
	panic(notPorted("u_iscript_nobrk"))
}

func (f *StateFunctions) uCollision(u *Unit) bool {
	panic(notPorted("u_collision"))
}

func (f *StateFunctions) uGroundUnit(u *Unit) bool {
	panic(notPorted("u_ground_unit"))
}

func (f *StateFunctions) uNoCollide(u *Unit) bool {
	panic(notPorted("u_no_collide"))
}

func (f *StateFunctions) uInvincible(u *Unit) bool {
	panic(notPorted("u_invincible"))
}

func (f *StateFunctions) uReadyToAttack(u *Unit) bool {
	panic(notPorted("u_ready_to_attack"))
}

func (f *StateFunctions) uPassivelyCloaked(u *Unit) bool {
	panic(notPorted("u_passively_cloaked"))
}

func (f *StateFunctions) utTurret(ut *UnitTypeT) bool {
	panic(notPorted("ut_turret"))
}

func (f *StateFunctions) utWorker(ut *UnitTypeT) bool {
	panic(notPorted("ut_worker"))
}

func (f *StateFunctions) utHero(ut *UnitTypeT) bool {
	panic(notPorted("ut_hero"))
}

func (f *StateFunctions) utBuilding(ut *UnitTypeT) bool {
	panic(notPorted("ut_building"))
}

func (f *StateFunctions) utAddon(ut *UnitTypeT) bool {
	panic(notPorted("ut_addon"))
}

func (f *StateFunctions) utFlyer(ut *UnitTypeT) bool {
	panic(notPorted("ut_flyer"))
}

func (f *StateFunctions) utCanTurn(ut *UnitTypeT) bool {
	panic(notPorted("ut_can_turn"))
}

func (f *StateFunctions) utCanMove(ut *UnitTypeT) bool {
	panic(notPorted("ut_can_move"))
}

func (f *StateFunctions) utInvincible(ut *UnitTypeT) bool {
	panic(notPorted("ut_invincible"))
}

func (f *StateFunctions) utTwoUnitsInOneEgg(ut *UnitTypeT) bool {
	panic(notPorted("ut_two_units_in_one_egg"))
}

func (f *StateFunctions) utRegensHp(ut *UnitTypeT) bool {
	panic(notPorted("ut_regens_hp"))
}

func (f *StateFunctions) utFlyingBuilding(ut *UnitTypeT) bool {
	panic(notPorted("ut_flying_building"))
}

func (f *StateFunctions) utRequiresPsionicMatrix(ut *UnitTypeT) bool {
	panic(notPorted("ut_requires_psionic_matrix"))
}

func (f *StateFunctions) utCanBurrow(ut *UnitTypeT) bool {
	panic(notPorted("ut_can_burrow"))
}

func (f *StateFunctions) utHasEnergy(ut *UnitTypeT) bool {
	panic(notPorted("ut_has_energy"))
}

func (f *StateFunctions) utResourceDepot(ut *UnitTypeT) bool {
	panic(notPorted("ut_resource_depot"))
}

func (f *StateFunctions) utResource(ut *UnitTypeT) bool {
	panic(notPorted("ut_resource"))
}

func (f *StateFunctions) utInitiallyCloaked(ut *UnitTypeT) bool {
	panic(notPorted("ut_initially_cloaked"))
}

func (f *StateFunctions) utDetector(ut *UnitTypeT) bool {
	panic(notPorted("ut_detector"))
}

func (f *StateFunctions) utRequiresCreep(ut *UnitTypeT) bool {
	panic(notPorted("ut_requires_creep"))
}

func (f *StateFunctions) utMechanical(ut *UnitTypeT) bool {
	panic(notPorted("ut_mechanical"))
}

func (f *StateFunctions) utOrganic(ut *UnitTypeT) bool {
	panic(notPorted("ut_organic"))
}

func (f *StateFunctions) utRobotic(ut *UnitTypeT) bool {
	panic(notPorted("ut_robotic"))
}

func (f *StateFunctions) utPowerup(ut *UnitTypeT) bool {
	panic(notPorted("ut_powerup"))
}

func (f *StateFunctions) usHidden(u *Thingy) bool {
	panic(notPorted("us_hidden"))
}

func (f *StateFunctions) sHidden(u *Sprite) bool {
	panic(notPorted("s_hidden"))
}

func (f *StateFunctions) getUnitType(id UnitTypes) *UnitTypeT {
	panic(notPorted("get_unit_type"))
}

func (f *StateFunctions) getImageType(id ImageTypes) *ImageTypeT {
	panic(notPorted("get_image_type"))
}

func (f *StateFunctions) getWeaponType(id WeaponTypes) *WeaponTypeT {
	panic(notPorted("get_weapon_type"))
}

func (f *StateFunctions) getSpriteType(id SpriteTypes) *SpriteTypeT {
	panic(notPorted("get_sprite_type"))
}

func (f *StateFunctions) getUpgradeType(id UpgradeTypes) *UpgradeTypeT {
	panic(notPorted("get_upgrade_type"))
}

func (f *StateFunctions) getTechType(id TechTypes) *TechTypeT {
	panic(notPorted("get_tech_type"))
}

func (f *StateFunctions) getFlingyType(id FlingyTypes) *FlingyTypeT {
	panic(notPorted("get_flingy_type"))
}

func (f *StateFunctions) playSoundUnit(id int32, sourceUnit *Unit, addRaceIndex bool) {
	panic(notPorted("play_sound"))
}

func (f *StateFunctions) playSound(id int32, addRaceIndex bool) {
	panic(notPorted("play_sound"))
}

func (f *StateFunctions) getUnit(index uint) *Unit {
	panic(notPorted("get_unit"))
}

func (f *StateFunctions) getUnitByID(id UnitID) *Unit {
	panic(notPorted("get_unit"))
}

func (f *StateFunctions) getUnitID(u *Unit) UnitID {
	panic(notPorted("get_unit_id"))
}

func (f *StateFunctions) getUnitID32(u *Unit) UnitID32 {
	panic(notPorted("get_unit_id_32"))
}

func (f *StateFunctions) isUnitTypeInMapBounds(unitType *UnitTypeT, pos Xy) bool {
	panic(notPorted("is_in_map_bounds"))
}

func (f *StateFunctions) isInMapBounds(area Rect) bool {
	panic(notPorted("is_in_map_bounds"))
}

func (f *StateFunctions) isInInnerMapBounds(area Rect) bool {
	panic(notPorted("is_in_inner_map_bounds"))
}

func (f *StateFunctions) isPosInMapBounds(pos Xy) bool {
	panic(notPorted("is_in_map_bounds"))
}

func (f *StateFunctions) isInBounds(area Rect, bounds Rect) bool {
	panic(notPorted("is_in_bounds"))
}

func (f *StateFunctions) isInInnerBounds(area Rect, bounds Rect) bool {
	panic(notPorted("is_in_inner_bounds"))
}

func (f *StateFunctions) isPosInBounds(pos Xy, bounds Rect) bool {
	panic(notPorted("is_in_bounds"))
}

func (f *StateFunctions) isPosInInnerBounds(pos Xy, bounds Rect) bool {
	panic(notPorted("is_in_inner_bounds"))
}

func (f *StateFunctions) translateRect(src Rect, translation Xy) Rect {
	panic(notPorted("translate_rect"))
}

func (f *StateFunctions) squareAt(pos Xy, halfWidth int32) Rect {
	panic(notPorted("square_at"))
}

func (f *StateFunctions) unitTypeBoundingBox(unitType *UnitTypeT) Rect {
	panic(notPorted("unit_type_bounding_box"))
}

func (f *StateFunctions) unitTypeBoundingBoxAt(unitType *UnitTypeT, origin Xy) Rect {
	panic(notPorted("unit_type_bounding_box"))
}

func (f *StateFunctions) unitBoundingBox(u *Unit, origin Xy) Rect {
	panic(notPorted("unit_bounding_box"))
}

func (f *StateFunctions) unitSpriteBoundingBox(u *Unit) Rect {
	panic(notPorted("unit_sprite_bounding_box"))
}

func (f *StateFunctions) unitTypeInnerBoundingBox(unitType *UnitTypeT) Rect {
	panic(notPorted("unit_type_inner_bounding_box"))
}

func (f *StateFunctions) unitTypeInnerBoundingBoxAt(unitType *UnitTypeT, origin Xy) Rect {
	panic(notPorted("unit_type_inner_bounding_box"))
}

func (f *StateFunctions) unitInnerBoundingBox(u *Unit, origin Xy) Rect {
	panic(notPorted("unit_inner_bounding_box"))
}

func (f *StateFunctions) unitSpriteInnerBoundingBox(u *Unit) Rect {
	panic(notPorted("unit_sprite_inner_bounding_box"))
}

func (f *StateFunctions) mapBounds() Rect {
	panic(notPorted("map_bounds"))
}

func (f *StateFunctions) restrictUnitPosToBounds(pos Xy, ut *UnitTypeT, bounds Rect) Xy {
	panic(notPorted("restrict_unit_pos_to_bounds"))
}

func (f *StateFunctions) restrictPosToMapBounds(pos Xy) Xy {
	panic(notPorted("restrict_pos_to_map_bounds"))
}

func (f *StateFunctions) isInnerBbMoveTargetInValidBounds(bb Rect) bool {
	panic(notPorted("is_inner_bb_move_target_in_valid_bounds"))
}

func (f *StateFunctions) restrictMoveTargetToValidBounds(ut *UnitTypeT, moveTarget Xy) Xy {
	panic(notPorted("restrict_move_target_to_valid_bounds"))
}

func (f *StateFunctions) isWalkable(pos Xy) bool {
	panic(notPorted("is_walkable"))
}

func (f *StateFunctions) tilesFlagsAnd(offsetX uint, offsetY uint, width uint, height uint, flags int32) {
	panic(notPorted("tiles_flags_and"))
}

func (f *StateFunctions) tilesFlagsOr(offsetX uint, offsetY uint, width uint, height uint, flags int32) {
	panic(notPorted("tiles_flags_or"))
}

func (f *StateFunctions) unitTypeSpreadsCreep(ut *UnitTypeT, unitIsCompleted bool) bool {
	panic(notPorted("unit_type_spreads_creep"))
}

func (f *StateFunctions) visibleHpPlusShields(u *Unit) int32 {
	panic(notPorted("visible_hp_plus_shields"))
}

func (f *StateFunctions) maxVisibleHp(u *Unit) int32 {
	panic(notPorted("max_visible_hp"))
}

func (f *StateFunctions) maxVisibleHpPlusShields(u *Unit) int32 {
	panic(notPorted("max_visible_hp_plus_shields"))
}

func (f *StateFunctions) loadedUnits(u *Unit) iter.Seq[*Unit] {
	panic(notPorted("loaded_units"))
}

func (f *StateFunctions) unitSpaceOccupied(u *Unit) uint {
	panic(notPorted("unit_space_occupied"))
}

func (f *StateFunctions) getUnitStrength(u *Unit, ground bool) int32 {
	panic(notPorted("get_unit_strength"))
}

func (f *StateFunctions) updateUnitDamageOverlay(u *Unit) {
	panic(notPorted("update_unit_damage_overlay"))
}

func (f *StateFunctions) setUnitHp(u *Unit, hitpoints Fp8) {
	panic(notPorted("set_unit_hp"))
}

func (f *StateFunctions) setUnitShieldPoints(u *Unit, shieldPoints Fp8) {
	panic(notPorted("set_unit_shield_points"))
}

func (f *StateFunctions) setUnitEnergy(u *Unit, energy Fp8) {
	panic(notPorted("set_unit_energy"))
}

func (f *StateFunctions) unitIsMineralField(ut *UnitTypeT) bool {
	panic(notPorted("unit_is_mineral_field"))
}

func (f *StateFunctions) setUnitResources(u *Unit, resources int32) {
	panic(notPorted("set_unit_resources"))
}

func (f *StateFunctions) unitIsDisabled(u *Unit) bool {
	panic(notPorted("unit_is_disabled"))
}

func (f *StateFunctions) findImage(sprite *Sprite, firstID ImageTypes, lastID ImageTypes) *Image {
	panic(notPorted("find_image"))
}

func (f *StateFunctions) findImageUnit(u *Unit, firstID ImageTypes, lastID ImageTypes) *Image {
	panic(notPorted("find_image"))
}

func (f *StateFunctions) destroyImageFromTo(sprite *Sprite, firstID ImageTypes, lastID ImageTypes) {
	panic(notPorted("destroy_image_from_to"))
}

func (f *StateFunctions) destroyImageFromToUnit(u *Unit, firstID ImageTypes, lastID ImageTypes) {
	panic(notPorted("destroy_image_from_to"))
}

func (f *StateFunctions) disableEffectEnd(u *Unit, first ImageTypes, last ImageTypes) {
	panic(notPorted("disable_effect_end"))
}

func (f *StateFunctions) removeStasis(u *Unit) {
	panic(notPorted("remove_stasis"))
}

func (f *StateFunctions) dealIrradiateDamage(sourceUnit *Unit) {
	panic(notPorted("deal_irradiate_damage"))
}

func (f *StateFunctions) removeIrradiate(u *Unit) {
	panic(notPorted("remove_irradiate"))
}

func (f *StateFunctions) removeEnsnare(u *Unit) {
	panic(notPorted("remove_ensnare"))
}

func (f *StateFunctions) removeLockdown(u *Unit) {
	panic(notPorted("remove_lockdown"))
}

func (f *StateFunctions) removePlague(u *Unit) {
	panic(notPorted("remove_plague"))
}

func (f *StateFunctions) removeMaelstrom(u *Unit) {
	panic(notPorted("remove_maelstrom"))
}

func (f *StateFunctions) removeAcidSpores(u *Unit) {
	panic(notPorted("remove_acid_spores"))
}

func (f *StateFunctions) updateUnitStatusTimers(u *Unit) {
	panic(notPorted("update_unit_status_timers"))
}

func (f *StateFunctions) unitCloakEnergyCost(u *Unit) Fp8 {
	panic(notPorted("unit_cloak_energy_cost"))
}

func (f *StateFunctions) setSecondaryOrder(u *Unit, orderType *OrderTypeT) {
	panic(notPorted("set_secondary_order"))
}

func (f *StateFunctions) updateUnitEnergy(u *Unit) {
	panic(notPorted("update_unit_energy"))
}

func (f *StateFunctions) unitHpPercent(u *Unit) int32 {
	panic(notPorted("unit_hp_percent"))
}

func (f *StateFunctions) updateUnitValues(u *Unit) {
	panic(notPorted("update_unit_values"))
}

func (f *StateFunctions) unitTurret(u *Unit) *Unit {
	panic(notPorted("unit_turret"))
}

func (f *StateFunctions) unitMainUnit(u *Unit) *Unit {
	panic(notPorted("unit_main_unit"))
}

func (f *StateFunctions) unitAttackingUnit(u *Unit) *Unit {
	panic(notPorted("unit_attacking_unit"))
}

func (f *StateFunctions) unitGroundWeapon(u *Unit) *WeaponTypeT {
	panic(notPorted("unit_ground_weapon"))
}

func (f *StateFunctions) unitAirWeapon(u *Unit) *WeaponTypeT {
	panic(notPorted("unit_air_weapon"))
}

func (f *StateFunctions) unitOrSubunitGroundWeapon(u *Unit) *WeaponTypeT {
	panic(notPorted("unit_or_subunit_ground_weapon"))
}

func (f *StateFunctions) unitOrSubunitAirWeapon(u *Unit) *WeaponTypeT {
	panic(notPorted("unit_or_subunit_air_weapon"))
}

func (f *StateFunctions) unitTargetWeapon(u *Unit, target *Unit) *WeaponTypeT {
	panic(notPorted("unit_target_weapon"))
}

func (f *StateFunctions) unitIsCarrier(ut *UnitTypeT) bool {
	panic(notPorted("unit_is_carrier"))
}

func (f *StateFunctions) unitIsReaver(ut *UnitTypeT) bool {
	panic(notPorted("unit_is_reaver"))
}

func (f *StateFunctions) unitIsQueen(ut *UnitTypeT) bool {
	panic(notPorted("unit_is_queen"))
}

func (f *StateFunctions) unitIsHatchery(ut *UnitTypeT) bool {
	panic(notPorted("unit_is_hatchery"))
}

func (f *StateFunctions) unitIsMarine(ut *UnitTypeT) bool {
	panic(notPorted("unit_is_marine"))
}

func (f *StateFunctions) unitIsFirebat(ut *UnitTypeT) bool {
	panic(notPorted("unit_is_firebat"))
}

func (f *StateFunctions) unitIsGhost(ut *UnitTypeT) bool {
	panic(notPorted("unit_is_ghost"))
}

func (f *StateFunctions) unitIsWraith(ut *UnitTypeT) bool {
	panic(notPorted("unit_is_wraith"))
}

func (f *StateFunctions) unitIsMapRevealer(ut *UnitTypeT) bool {
	panic(notPorted("unit_is_map_revealer"))
}

func (f *StateFunctions) unitIsRefinery(ut *UnitTypeT) bool {
	panic(notPorted("unit_is_refinery"))
}

func (f *StateFunctions) unitIsNonFlagBeacon(ut *UnitTypeT) bool {
	panic(notPorted("unit_is_non_flag_beacon"))
}

func (f *StateFunctions) unitIsSpecialBeacon(ut *UnitTypeT) bool {
	panic(notPorted("unit_is_special_beacon"))
}

func (f *StateFunctions) unitIsScout(ut *UnitTypeT) bool {
	panic(notPorted("unit_is_scout"))
}

func (f *StateFunctions) unitIsVulture(ut *UnitTypeT) bool {
	panic(notPorted("unit_is_vulture"))
}

func (f *StateFunctions) unitIsNydus(ut *UnitTypeT) bool {
	panic(notPorted("unit_is_nydus"))
}

func (f *StateFunctions) unitIsDefiler(ut *UnitTypeT) bool {
	panic(notPorted("unit_is_defiler"))
}

func (f *StateFunctions) unitIsUltralisk(ut *UnitTypeT) bool {
	panic(notPorted("unit_is_ultralisk"))
}

func (f *StateFunctions) unitIsArbiter(ut *UnitTypeT) bool {
	panic(notPorted("unit_is_arbiter"))
}

func (f *StateFunctions) unitIsEgg(ut *UnitTypeT) bool {
	panic(notPorted("unit_is_egg"))
}

func (f *StateFunctions) unitIsGoliath(ut *UnitTypeT) bool {
	panic(notPorted("unit_is_goliath"))
}

func (f *StateFunctions) unitIsCritter(ut *UnitTypeT) bool {
	panic(notPorted("unit_is_critter"))
}

func (f *StateFunctions) unitIsSiegedTank(ut *UnitTypeT) bool {
	panic(notPorted("unit_is_sieged_tank"))
}

func (f *StateFunctions) unitIsUnsiegedTank(ut *UnitTypeT) bool {
	panic(notPorted("unit_is_unsieged_tank"))
}

func (f *StateFunctions) unitIsTank(ut *UnitTypeT) bool {
	panic(notPorted("unit_is_tank"))
}

func (f *StateFunctions) unitIsTrap(ut *UnitTypeT) bool {
	panic(notPorted("unit_is_trap"))
}

func (f *StateFunctions) unitIsZergBuilding(ut *UnitTypeT) bool {
	panic(notPorted("unit_is_zerg_building"))
}

func (f *StateFunctions) unitIsFighter(ut *UnitTypeT) bool {
	panic(notPorted("unit_is_fighter"))
}

func (f *StateFunctions) unitIsUndetected(u *Unit, owner int32) bool {
	panic(notPorted("unit_is_undetected"))
}

func (f *StateFunctions) unitTargetIsUndetected(u *Unit, target *Unit) bool {
	panic(notPorted("unit_target_is_undetected"))
}

func (f *StateFunctions) unitTargetIsVisible(u *Unit, target *Unit) bool {
	panic(notPorted("unit_target_is_visible"))
}

func (f *StateFunctions) unitPositionIsVisible(u *Unit, position Xy) bool {
	panic(notPorted("unit_position_is_visible"))
}

func (f *StateFunctions) unitPositionIsExplored(u *Unit, position Xy) bool {
	panic(notPorted("unit_position_is_explored"))
}

func (f *StateFunctions) unitCanSeeTarget(u *Unit, target *Unit) bool {
	panic(notPorted("unit_can_see_target"))
}

func (f *StateFunctions) unitOrderCanTargetSelf(u *Unit, order *OrderTypeT) bool {
	panic(notPorted("unit_order_can_target_self"))
}

func (f *StateFunctions) playerPositionIsVisible(owner int32, position Xy) bool {
	panic(notPorted("player_position_is_visible"))
}

func (f *StateFunctions) playerPositionIsExplored(owner int32, position Xy) bool {
	panic(notPorted("player_position_is_explored"))
}

func (f *StateFunctions) getRegionAt(pos Xy) *Region {
	panic(notPorted("get_region_at"))
}

func (f *StateFunctions) getRegionAtPreferWalkable(pos Xy) *Region {
	panic(notPorted("get_region_at_prefer_walkable"))
}

func (f *StateFunctions) isReachable(from Xy, to Xy) bool {
	panic(notPorted("is_reachable"))
}

func (f *StateFunctions) isReachableUnitToPos(u *Unit, to Xy) bool {
	panic(notPorted("is_reachable"))
}

func (f *StateFunctions) isReachableUnitToUnit(fromU *Unit, toU *Unit) bool {
	panic(notPorted("is_reachable"))
}

func (f *StateFunctions) unitCanBeInfested(u *Unit) bool {
	panic(notPorted("unit_can_be_infested"))
}

func (f *StateFunctions) unitCanAttackTarget(u *Unit, target *Unit) bool {
	panic(notPorted("unit_can_attack_target"))
}

func (f *StateFunctions) unitAutoattack(u *Unit) bool {
	panic(notPorted("unit_autoattack"))
}

func (f *StateFunctions) unitCanCollideWith(u *Unit, target *Unit) bool {
	panic(notPorted("unit_can_collide_with"))
}

func (f *StateFunctions) rectDifference(a Rect, b Rect) Xy {
	panic(notPorted("rect_difference"))
}

func (f *StateFunctions) nearestPosInRect(pos Xy, area Rect) Xy {
	panic(notPorted("nearest_pos_in_rect"))
}

func (f *StateFunctions) xyLength(vec Xy) int32 {
	panic(notPorted("xy_length"))
}

func (f *StateFunctions) xyLengthFP8(vec XyFp8) Fp8 {
	panic(notPorted("xy_length"))
}

func (f *StateFunctions) toXyFp8(position Xy) XyFp8 {
	panic(notPorted("to_xy_fp8"))
}

func (f *StateFunctions) toXy(position XyFp8) Xy {
	panic(notPorted("to_xy"))
}

func (f *StateFunctions) unitsDistance(a *Unit, b *Unit) int32 {
	panic(notPorted("units_distance"))
}

func (f *StateFunctions) unitDistanceTo(u *Unit, pos Xy) int32 {
	panic(notPorted("unit_distance_to"))
}

func (f *StateFunctions) atanFP8(x Fp8) DirectionT {
	panic(notPorted("atan"))
}

func (f *StateFunctions) xyDirectionFP8(pos XyFp8) DirectionT {
	panic(notPorted("xy_direction"))
}

func (f *StateFunctions) xyDirection(pos Xy) DirectionT {
	panic(notPorted("xy_direction"))
}

func (f *StateFunctions) directionXY(dir DirectionT, length Fp8) XyFp8 {
	panic(notPorted("direction_xy"))
}

func (f *StateFunctions) directionXYInt(dir DirectionT, length int32) XyFp8 {
	panic(notPorted("direction_xy"))
}

func (f *StateFunctions) directionXYOne(dir DirectionT) XyFp8 {
	panic(notPorted("direction_xy"))
}

func (f *StateFunctions) directionIndex(dir DirectionT) uint {
	panic(notPorted("direction_index"))
}

func (f *StateFunctions) directionFromIndex(index uint) DirectionT {
	panic(notPorted("direction_from_index"))
}

func (f *StateFunctions) unitsDirection(from *Unit, to *Unit) DirectionT {
	panic(notPorted("units_direction"))
}

func (f *StateFunctions) unitTargetInAttackAngle(u *Unit, targetPos Xy, weapon *WeaponTypeT) bool {
	panic(notPorted("unit_target_in_attack_angle"))
}

func (f *StateFunctions) weaponMaxRange(u *Unit, w *WeaponTypeT) int32 {
	panic(notPorted("weapon_max_range"))
}

func (f *StateFunctions) unitTargetMovementRange(u *Flingy, target *Flingy) int32 {
	panic(notPorted("unit_target_movement_range"))
}

func (f *StateFunctions) unitTargetInWeaponMovementRange(u *Unit, target *Unit) bool {
	panic(notPorted("unit_target_in_weapon_movement_range"))
}

func (f *StateFunctions) unitTargetOutOfMaxRange(u *Unit, target *Unit) bool {
	panic(notPorted("unit_target_out_of_max_range"))
}

func (f *StateFunctions) unitFirstLoadedUnit(u *Unit) *Unit {
	panic(notPorted("unit_first_loaded_unit"))
}

func (f *StateFunctions) unitRace(ut *UnitTypeT) Race {
	panic(notPorted("unit_race"))
}

func (f *StateFunctions) unitProvidesSpace(u *Unit) bool {
	panic(notPorted("unit_provides_space"))
}

func (f *StateFunctions) unitSpaceLeft(u *Unit) uint {
	panic(notPorted("unit_space_left"))
}

func (f *StateFunctions) unitCanLoadTarget(u *Unit, target *Unit) bool {
	panic(notPorted("unit_can_load_target"))
}

func (f *StateFunctions) unitInterceptorCount(u *Unit) uint {
	panic(notPorted("unit_interceptor_count"))
}

func (f *StateFunctions) unitScarabCount(u *Unit) uint {
	panic(notPorted("unit_scarab_count"))
}

func (f *StateFunctions) unitMaxInterceptorCount(u *Unit) uint {
	panic(notPorted("unit_max_interceptor_count"))
}

func (f *StateFunctions) unitMaxScarabCount(u *Unit) uint {
	panic(notPorted("unit_max_scarab_count"))
}

func (f *StateFunctions) unitQueuedFighterUnits(u *Unit) uint {
	panic(notPorted("unit_queued_fighter_units"))
}

func (f *StateFunctions) unitSpiderMineCount(u *Unit) uint {
	panic(notPorted("unit_spider_mine_count"))
}

func (f *StateFunctions) unitCanAttack(u *Unit) bool {
	panic(notPorted("unit_can_attack"))
}

func (f *StateFunctions) unitTargetAttackPriority(u *Unit, target *Unit) int32 {
	panic(notPorted("unit_target_attack_priority"))
}

func (f *StateFunctions) unitTargetInRange(u *Unit, target *Unit, rangeValue int32) bool {
	panic(notPorted("unit_target_in_range"))
}

func (f *StateFunctions) getDefaultPriorityTargets(u *Unit, minDistance int32, maxDistance int32) [6]StaticVector[*Unit, [16]*Unit] {
	panic(notPorted("get_default_priority_targets"))
}

func (f *StateFunctions) findAcquireTarget(u *Unit) *Unit {
	panic(notPorted("find_acquire_target"))
}

func (f *StateFunctions) findAcquireRandomTarget(u *Unit) *Unit {
	panic(notPorted("find_acquire_random_target"))
}

func (f *StateFunctions) unitTypeIsMorphingBuilding(ut *UnitTypeT) bool {
	panic(notPorted("unit_type_is_morphing_building"))
}

func (f *StateFunctions) unitIsMorphingBuilding(u *Unit) bool {
	panic(notPorted("unit_is_morphing_building"))
}

func (f *StateFunctions) unitSightRange(u *Unit, ignoreBlindness bool) int32 {
	panic(notPorted("unit_sight_range"))
}

func (f *StateFunctions) unitSightRangeIgnoreBlindness(u *Unit) int32 {
	panic(notPorted("unit_sight_range_ignore_blindness"))
}

func (f *StateFunctions) unitTargetAcquisitionRange(u *Unit) int32 {
	panic(notPorted("unit_target_acquisition_range"))
}

func (f *StateFunctions) unitMaxEnergy(u *Unit) Fp8 {
	panic(notPorted("unit_max_energy"))
}

func (f *StateFunctions) visibleToEveryone(u *Unit) bool {
	panic(notPorted("visible_to_everyone"))
}

func (f *StateFunctions) tileIndex(pos Xy) uint {
	panic(notPorted("tile_index"))
}

func (f *StateFunctions) tileVisibility(pos Xy) uint8 {
	panic(notPorted("tile_visibility"))
}

func (f *StateFunctions) tileExplored(pos Xy) uint8 {
	panic(notPorted("tile_explored"))
}

func (f *StateFunctions) getGroundHeightAt(pos Xy) int32 {
	panic(notPorted("get_ground_height_at"))
}

func (f *StateFunctions) displayLastErrorForPlayer(player int32) {
	panic(notPorted("display_last_error_for_player"))
}

func (f *StateFunctions) playerHasUnit(owner int32, unitTypeID UnitTypes) bool {
	panic(notPorted("player_has_unit"))
}

func (f *StateFunctions) playerHasCompletedUnit(owner int32, unitTypeID UnitTypes) bool {
	panic(notPorted("player_has_completed_unit"))
}

func (f *StateFunctions) playerTechAvailable(owner int32, techID TechTypes) bool {
	panic(notPorted("player_tech_available"))
}

func (f *StateFunctions) playerHasResearched(owner int32, techID TechTypes) bool {
	panic(notPorted("player_has_researched"))
}

func (f *StateFunctions) playerIsResearching(owner int32, techID TechTypes) bool {
	panic(notPorted("player_is_researching"))
}

func (f *StateFunctions) playerMaxUpgradeLevel(owner int32, upgradeID UpgradeTypes) int32 {
	panic(notPorted("player_max_upgrade_level"))
}

func (f *StateFunctions) playerHasUpgrade(owner int32, upgradeID UpgradeTypes) bool {
	panic(notPorted("player_has_upgrade"))
}

func (f *StateFunctions) playerIsUpgrading(owner int32, upgradeID UpgradeTypes) bool {
	panic(notPorted("player_is_upgrading"))
}

func (f *StateFunctions) playerUpgradeLevel(owner int32, upgradeID UpgradeTypes) int32 {
	panic(notPorted("player_upgrade_level"))
}

func (f *StateFunctions) unitIs(ut *UnitTypeT, unitTypeID UnitTypes) bool {
	panic(notPorted("unit_is"))
}

func (f *StateFunctions) unitAddon(u *Unit) *Unit {
	panic(notPorted("unit_addon"))
}

func (f *StateFunctions) unitHasAddon(u *Unit, unitTypeID UnitTypes) bool {
	panic(notPorted("unit_has_addon"))
}

func (f *StateFunctions) unitIsBuildingAddon(u *Unit) bool {
	panic(notPorted("unit_is_building_addon"))
}

func (f *StateFunctions) unitIsBuildingProtossThing(u *Unit) bool {
	panic(notPorted("unit_is_building_protoss_thing"))
}

func (f *StateFunctions) unitIsConstructing(u *Unit) bool {
	panic(notPorted("unit_is_constructing"))
}

func (f *StateFunctions) unitIsResearching(u *Unit) bool {
	panic(notPorted("unit_is_researching"))
}

func (f *StateFunctions) unitIsUpgrading(u *Unit) bool {
	panic(notPorted("unit_is_upgrading"))
}

func (f *StateFunctions) unitCanMove(u *Unit) bool {
	panic(notPorted("unit_can_move"))
}

func (f *StateFunctions) unitCanHoldPosition(u *Unit) bool {
	panic(notPorted("unit_can_hold_position"))
}

func (f *StateFunctions) unitIsImmovableAttacker(u *Unit) bool {
	panic(notPorted("unit_is_immovable_attacker"))
}
