package bwgame

func (f *StateFunctions) isSpell(weaponType *WeaponTypeT) bool {
	panic(notPorted("is_spell"))
}

func (f *StateFunctions) weaponDamageAmount(w *WeaponTypeT, owner int32) Fp8 {
	panic(notPorted("weapon_damage_amount"))
}

func (f *StateFunctions) bulletDamageAmount(b *Bullet, target *Unit) Fp8 {
	panic(notPorted("bullet_damage_amount"))
}

func (f *StateFunctions) unitSpriteSize(ut *UnitTypeT) int32 {
	panic(notPorted("unit_sprite_size"))
}

func (f *StateFunctions) createSizedImage(u *Unit, smallImage ImageTypes, onSubunit bool, imageOrder int32) *Image {
	panic(notPorted("create_sized_image"))
}

func (f *StateFunctions) createDefensiveMatrixImage(u *Unit) {
	panic(notPorted("create_defensive_matrix_image"))
}

func (f *StateFunctions) dealDefensiveMatrixDamage(u *Unit, damage Fp8) {
	panic(notPorted("deal_defensive_matrix_damage"))
}

func (f *StateFunctions) unitArmor(u *Unit) Fp8 {
	panic(notPorted("unit_armor"))
}

func (f *StateFunctions) onHitShouldChangeTarget(u *Unit, oldTarget *Unit, newTarget *Unit) bool {
	panic(notPorted("on_hit_should_change_target"))
}

func (f *StateFunctions) unburrowUnit(u *Unit) {
	panic(notPorted("unburrow_unit"))
}

func (f *StateFunctions) unitCanMoveFromToRegion(owner int32, from *Region, to *Region) bool {
	panic(notPorted("unit_can_move_from_to_region"))
}

func (f *StateFunctions) onHitChangeTarget(u *Unit, target *Unit) {
	panic(notPorted("on_hit_change_target"))
}

func (f *StateFunctions) onHitAoeChangeTarget(u *Unit, target *Unit) {
	panic(notPorted("on_hit_aoe_change_target"))
}

func (f *StateFunctions) onUnitDamage(u *Unit, sourceUnit *Unit, revealSource bool) {
	panic(notPorted("on_unit_damage"))
}

func (f *StateFunctions) unitDealDamage(u *Unit, damage Fp8, sourceUnit *Unit, sourceOwner int32, revealSource bool) {
	panic(notPorted("unit_deal_damage"))
}

func (f *StateFunctions) createShieldDamageEffect(target *Unit, heading DirectionT) {
	panic(notPorted("create_shield_damage_effect"))
}

func (f *StateFunctions) weaponDealDamage(weapon *WeaponTypeT, damage Fp8, damageDivisor int32, target *Unit, heading DirectionT, sourceUnit *Unit, sourceOwner int32) {
	panic(notPorted("weapon_deal_damage"))
}

func (f *StateFunctions) hallucinatedWeaponHit(weapon *WeaponTypeT, target *Unit, heading DirectionT, sourceUnit *Unit) {
	panic(notPorted("hallucinated_weapon_hit"))
}

func (f *StateFunctions) bulletDealDamageDivisor(b *Bullet, target *Unit, damageDivisor int32) {
	panic(notPorted("bullet_deal_damage"))
}

func (f *StateFunctions) bulletDealDamage(b *Bullet, target *Unit) {
	panic(notPorted("bullet_deal_damage"))
}

func (f *StateFunctions) meleeDealDamage(u *Unit) {
	panic(notPorted("melee_deal_damage"))
}

func (f *StateFunctions) weaponCanTargetUnit(weapon *WeaponTypeT, target *Unit) bool {
	panic(notPorted("weapon_can_target_unit"))
}

func (f *StateFunctions) weaponCanTargetUnitFrom(weapon *WeaponTypeT, target *Unit, sourceUnit *Unit) bool {
	panic(notPorted("weapon_can_target_unit"))
}

func (f *StateFunctions) bulletDealSplashDamage(b *Bullet, isAirSplash bool) {
	panic(notPorted("bullet_deal_splash_damage"))
}

func (f *StateFunctions) irradiateUnit(u *Unit, sourceUnit *Unit, sourceOwner int32) {
	panic(notPorted("irradiate_unit"))
}

func (f *StateFunctions) returnInterceptors(u *Unit) {
	panic(notPorted("return_interceptors"))
}

func (f *StateFunctions) setUnitDisabled(u *Unit) {
	panic(notPorted("set_unit_disabled"))
}

func (f *StateFunctions) lockdownUnit(u *Unit) {
	panic(notPorted("lockdown_unit"))
}

func (f *StateFunctions) blindUnit(u *Unit, sourceOwner int32) {
	panic(notPorted("blind_unit"))
}

func (f *StateFunctions) restoreUnit(u *Unit) {
	panic(notPorted("restore_unit"))
}

func (f *StateFunctions) empShockwave(position Xy, sourceUnit *Unit) {
	panic(notPorted("emp_shockwave"))
}

func (f *StateFunctions) spawnBroodlings(target *Unit, sourceUnit *Unit) {
	panic(notPorted("spawn_broodlings"))
}

func (f *StateFunctions) ensnareUnit(u *Unit) {
	panic(notPorted("ensnare_unit"))
}

func (f *StateFunctions) ensnare(pos Xy, sourceUnit *Unit) {
	panic(notPorted("ensnare"))
}

func (f *StateFunctions) consumeUnit(target *Unit, sourceUnit *Unit) {
	panic(notPorted("consume_unit"))
}

func (f *StateFunctions) darkSwarm(pos Xy, owner int32) {
	panic(notPorted("dark_swarm"))
}

func (f *StateFunctions) plagueUnit(u *Unit) {
	panic(notPorted("plague_unit"))
}

func (f *StateFunctions) plague(pos Xy, sourceUnit *Unit) {
	panic(notPorted("plague"))
}

func (f *StateFunctions) parasiteUnit(u *Unit, owner int32) {
	panic(notPorted("parasite_unit"))
}

func (f *StateFunctions) stasisUnit(u *Unit) {
	panic(notPorted("stasis_unit"))
}

func (f *StateFunctions) stasisField(pos Xy, sourceUnit *Unit) {
	panic(notPorted("stasis_field"))
}

func (f *StateFunctions) maelstromUnit(u *Unit) {
	panic(notPorted("maelstrom_unit"))
}

func (f *StateFunctions) maelstrom(pos Xy, sourceUnit *Unit) {
	panic(notPorted("maelstrom"))
}

func (f *StateFunctions) disruptionWeb(pos Xy, owner int32) {
	panic(notPorted("disruption_web"))
}

func (f *StateFunctions) acidSporeImage(u *Unit) ImageTypes {
	panic(notPorted("acid_spore_image"))
}

func (f *StateFunctions) updateAcidSporeImage(u *Unit) {
	panic(notPorted("update_acid_spore_image"))
}

func (f *StateFunctions) addAcidSpore(u *Unit) {
	panic(notPorted("add_acid_spore"))
}

func (f *StateFunctions) addAcidSporeAt(pos Xy, owner int32) {
	panic(notPorted("add_acid_spore"))
}

func (f *StateFunctions) bulletHit(b *Bullet) {
	panic(notPorted("bullet_hit"))
}

func (f *StateFunctions) bulletKill(b *Bullet) {
	panic(notPorted("bullet_kill"))
}

func (f *StateFunctions) bulletMove(b *Bullet, ems *ExecuteMovementStruct) {
	panic(notPorted("bullet_move"))
}

func (f *StateFunctions) setBulletMoveTarget(b *Bullet, target Xy) {
	panic(notPorted("set_bullet_move_target"))
}

func (f *StateFunctions) bulletStateInit(b *Bullet, ems *ExecuteMovementStruct) bool {
	panic(notPorted("bullet_state_init"))
}

func (f *StateFunctions) bulletStateDying(b *Bullet, ems *ExecuteMovementStruct) bool {
	panic(notPorted("bullet_state_dying"))
}

func (f *StateFunctions) bulletStateMove(b *Bullet, ems *ExecuteMovementStruct) bool {
	panic(notPorted("bullet_state_move"))
}

func (f *StateFunctions) bulletStateFollow(b *Bullet, ems *ExecuteMovementStruct) bool {
	panic(notPorted("bullet_state_follow"))
}

func (f *StateFunctions) bulletStateHitNearTarget(b *Bullet, ems *ExecuteMovementStruct) bool {
	panic(notPorted("bullet_state_hit_near_target"))
}

func (f *StateFunctions) bulletStateBounce(b *Bullet, ems *ExecuteMovementStruct) bool {
	panic(notPorted("bullet_state_bounce"))
}

func (f *StateFunctions) bulletStateDamageOverTime(b *Bullet, ems *ExecuteMovementStruct) bool {
	panic(notPorted("bullet_state_damage_over_time"))
}

func (f *StateFunctions) bulletExecute(b *Bullet) {
	panic(notPorted("bullet_execute"))
}

func (f *StateFunctions) unitIsUnderDarkSwarm(u *Unit) bool {
	panic(notPorted("unit_is_under_dark_swarm"))
}

func (f *StateFunctions) unitDodgeChance(u *Unit) Fp8 {
	panic(notPorted("unit_dodge_chance"))
}

func (f *StateFunctions) unitTargetMissChance(u *Unit, target *Unit) Fp8 {
	panic(notPorted("unit_target_miss_chance"))
}

func (f *StateFunctions) initializeBullet(b *Bullet, weaponType *WeaponTypeT, sourceUnit *Unit, pos Xy, owner int32, heading DirectionT) bool {
	panic(notPorted("initialize_bullet"))
}

func (f *StateFunctions) createBullet(weaponType *WeaponTypeT, sourceUnit *Unit, pos Xy, owner int32, heading DirectionT) *Bullet {
	panic(notPorted("create_bullet"))
}

func (f *StateFunctions) getBulletAppearAtTargetPos(u *Unit, target *Unit) Xy {
	panic(notPorted("get_bullet_appear_at_target_pos"))
}

func (f *StateFunctions) fireWeapon(u *Unit, weaponType *WeaponTypeT, forwardOffset int32) {
	panic(notPorted("fire_weapon"))
}

func (f *StateFunctions) unitTechTargetValid(u *Unit, tech *TechTypeT, target *Unit) bool {
	panic(notPorted("unit_tech_target_valid"))
}

func (f *StateFunctions) spellOrderValid(u *Unit) bool {
	panic(notPorted("spell_order_valid"))
}
