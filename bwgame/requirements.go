package bwgame

func (f *StateFunctions) unitCanBuild(u *Unit, unitType *UnitTypeT) bool {
	if !f.uCompleted(u) {
		return false
	}
	if f.unitIsDisabled(u) {
		return false
	}
	if f.uHallucination(u) {
		return false
	}
	owner := u.Owner
	if !f.gameSt.UnitTypeAllowed[owner][unitType.ID] {
		return false
	}
	if u.BuildQueue.Size() >= 5 {
		if u.BuildQueue.Back().ID <= UnitTypesSpellDisruptionWeb {
			return false
		}
	}

	switch unitType.ID {
	case UnitTypesTerranMarine:
		if !f.unitIs(u.UnitType, UnitTypesTerranBarracks) {
			return false
		}
		if f.unitIsBuildingAddon(u) || f.unitIsBuildingProtossThing(u) {
			return false
		}
		if !f.uGroundedBuilding(u) {
			return false
		}
		return true
	case UnitTypesTerranGhost:
		if !f.unitIs(u.UnitType, UnitTypesTerranBarracks) {
			return false
		}
		if f.unitIsBuildingAddon(u) || f.unitIsBuildingProtossThing(u) {
			return false
		}
		if !f.uGroundedBuilding(u) {
			return false
		}
		if !f.playerHasCompletedUnit(owner, UnitTypesTerranCovertOps) {
			return false
		}
		if !f.playerHasCompletedUnit(owner, UnitTypesTerranAcademy) {
			return false
		}
		return true
	case UnitTypesTerranVulture:
		if !f.unitIs(u.UnitType, UnitTypesTerranFactory) {
			return false
		}
		if !f.uGroundedBuilding(u) {
			return false
		}
		if f.unitIsBuildingAddon(u) || f.unitIsBuildingProtossThing(u) {
			return false
		}
		return true
	case UnitTypesTerranGoliath:
		if !f.unitIs(u.UnitType, UnitTypesTerranFactory) {
			return false
		}
		if !f.uGroundedBuilding(u) {
			return false
		}
		if f.unitIsBuildingAddon(u) || f.unitIsBuildingProtossThing(u) {
			return false
		}
		if !f.playerHasCompletedUnit(owner, UnitTypesTerranArmory) {
			return false
		}
		return true
	case UnitTypesTerranSiegeTankTankMode:
		if !f.unitIs(u.UnitType, UnitTypesTerranFactory) {
			return false
		}
		if !f.uGroundedBuilding(u) {
			return false
		}
		if f.unitIsBuildingAddon(u) || f.unitIsBuildingProtossThing(u) {
			return false
		}
		if !f.unitHasAddon(u, UnitTypesTerranMachineShop) {
			return false
		}
		return true
	case UnitTypesTerranSCV:
		if !f.unitIs(u.UnitType, UnitTypesTerranCommandCenter) {
			return false
		}
		if f.unitIsBuildingAddon(u) || f.unitIsBuildingProtossThing(u) {
			return false
		}
		if !f.uGroundedBuilding(u) {
			return false
		}
		return true
	case UnitTypesTerranWraith:
		if !f.unitIs(u.UnitType, UnitTypesTerranStarport) {
			return false
		}
		if f.unitIsBuildingAddon(u) || f.unitIsBuildingProtossThing(u) {
			return false
		}
		if !f.uGroundedBuilding(u) {
			return false
		}
		return true
	case UnitTypesTerranScienceVessel:
		if !f.unitIs(u.UnitType, UnitTypesTerranStarport) {
			return false
		}
		if f.unitIsBuildingAddon(u) || f.unitIsBuildingProtossThing(u) {
			return false
		}
		if !f.uGroundedBuilding(u) {
			return false
		}
		if !f.unitHasAddon(u, UnitTypesTerranControlTower) {
			return false
		}
		if !f.playerHasCompletedUnit(owner, UnitTypesTerranScienceFacility) {
			return false
		}
		return true
	case UnitTypesTerranDropship:
		if !f.unitIs(u.UnitType, UnitTypesTerranStarport) {
			return false
		}
		if f.unitIsBuildingAddon(u) || f.unitIsBuildingProtossThing(u) {
			return false
		}
		if !f.uGroundedBuilding(u) {
			return false
		}
		if !f.unitHasAddon(u, UnitTypesTerranControlTower) {
			return false
		}
		return true
	case UnitTypesTerranBattlecruiser:
		if !f.unitIs(u.UnitType, UnitTypesTerranStarport) {
			return false
		}
		if f.unitIsBuildingAddon(u) || f.unitIsBuildingProtossThing(u) {
			return false
		}
		if !f.uGroundedBuilding(u) {
			return false
		}
		if !f.unitHasAddon(u, UnitTypesTerranControlTower) {
			return false
		}
		if !f.playerHasCompletedUnit(owner, UnitTypesTerranPhysicsLab) {
			return false
		}
		return true
	case UnitTypesTerranFirebat:
		if !f.unitIs(u.UnitType, UnitTypesTerranBarracks) {
			return false
		}
		if f.unitIsBuildingAddon(u) || f.unitIsBuildingProtossThing(u) {
			return false
		}
		if !f.uGroundedBuilding(u) {
			return false
		}
		if !f.playerHasCompletedUnit(owner, UnitTypesTerranAcademy) {
			return false
		}
		return true
	case UnitTypesTerranNuclearMissile:
		if !f.unitIs(u.UnitType, UnitTypesTerranNuclearSilo) {
			return false
		}
		if f.utBuilding(u.UnitType) && u.Building.Silo.Nuke != nil {
			return false
		}
		if f.unitIsConstructing(u) || f.unitIsResearching(u) || f.unitIsUpgrading(u) {
			return false
		}
		return true
	case UnitTypesTerranMedic:
		if !f.unitIs(u.UnitType, UnitTypesTerranBarracks) {
			return false
		}
		if f.unitIsBuildingAddon(u) || f.unitIsBuildingProtossThing(u) {
			return false
		}
		if !f.uGroundedBuilding(u) {
			return false
		}
		if !f.playerHasCompletedUnit(owner, UnitTypesTerranAcademy) {
			return false
		}
		return true
	case UnitTypesTerranValkyrie:
		if !f.unitIs(u.UnitType, UnitTypesTerranStarport) {
			return false
		}
		if f.unitIsBuildingAddon(u) || f.unitIsBuildingProtossThing(u) {
			return false
		}
		if !f.uGroundedBuilding(u) {
			return false
		}
		if !f.unitHasAddon(u, UnitTypesTerranControlTower) {
			return false
		}
		if !f.playerHasCompletedUnit(owner, UnitTypesTerranArmory) {
			return false
		}
		return true
	case UnitTypesZergZergling:
		if !f.unitIs(u.UnitType, UnitTypesZergLarva) {
			return false
		}
		if !f.playerHasCompletedUnit(owner, UnitTypesZergSpawningPool) {
			return false
		}
		return true
	case UnitTypesZergHydralisk:
		if !f.unitIs(u.UnitType, UnitTypesZergLarva) {
			return false
		}
		if !f.playerHasCompletedUnit(owner, UnitTypesZergHydraliskDen) {
			return false
		}
		return true
	case UnitTypesZergUltralisk:
		if !f.unitIs(u.UnitType, UnitTypesZergLarva) {
			return false
		}
		if !f.playerHasCompletedUnit(owner, UnitTypesZergUltraliskCavern) {
			return false
		}
		return true
	case UnitTypesZergDrone:
		if !f.unitIs(u.UnitType, UnitTypesZergLarva) {
			return false
		}
		return true
	case UnitTypesZergOverlord:
		if !f.unitIs(u.UnitType, UnitTypesZergLarva) {
			return false
		}
		return true
	case UnitTypesZergMutalisk:
		if !f.unitIs(u.UnitType, UnitTypesZergLarva) {
			return false
		}
		if !f.playerHasCompletedUnit(owner, UnitTypesZergSpire) && !f.playerHasUnit(owner, UnitTypesZergGreaterSpire) {
			return false
		}
		return true
	case UnitTypesZergGuardian:
		if !f.unitIs(u.UnitType, UnitTypesZergMutalisk) {
			return false
		}
		if !f.playerHasCompletedUnit(owner, UnitTypesZergGreaterSpire) {
			return false
		}
		return true
	case UnitTypesZergQueen:
		if !f.unitIs(u.UnitType, UnitTypesZergLarva) {
			return false
		}
		if !f.playerHasCompletedUnit(owner, UnitTypesZergQueensNest) {
			return false
		}
		return true
	case UnitTypesZergDefiler:
		if !f.unitIs(u.UnitType, UnitTypesZergLarva) {
			return false
		}
		if !f.playerHasCompletedUnit(owner, UnitTypesZergDefilerMound) {
			return false
		}
		return true
	case UnitTypesZergScourge:
		if !f.unitIs(u.UnitType, UnitTypesZergLarva) {
			return false
		}
		if !f.playerHasCompletedUnit(owner, UnitTypesZergSpire) && !f.playerHasUnit(owner, UnitTypesZergGreaterSpire) {
			return false
		}
		return true
	case UnitTypesZergInfestedTerran:
		if !f.unitIs(u.UnitType, UnitTypesZergInfestedCommandCenter) {
			return false
		}
		if !f.uGroundedBuilding(u) {
			return false
		}
		return true
	case UnitTypesZergDevourer:
		if !f.unitIs(u.UnitType, UnitTypesZergMutalisk) {
			return false
		}
		if !f.playerHasCompletedUnit(owner, UnitTypesZergGreaterSpire) {
			return false
		}
		return true
	case UnitTypesZergLurker:
		if !f.unitIs(u.UnitType, UnitTypesZergHydralisk) {
			return false
		}
		if !f.playerHasResearched(owner, TechTypesLurkerAspect) {
			return false
		}
		return true
	case UnitTypesProtossProbe:
		if !f.unitIs(u.UnitType, UnitTypesProtossNexus) {
			return false
		}
		if f.unitIsBuildingAddon(u) || f.unitIsBuildingProtossThing(u) {
			return false
		}
		return true
	case UnitTypesProtossZealot:
		if !f.unitIs(u.UnitType, UnitTypesProtossGateway) {
			return false
		}
		return true
	case UnitTypesProtossDragoon:
		if !f.unitIs(u.UnitType, UnitTypesProtossGateway) {
			return false
		}
		if !f.playerHasCompletedUnit(owner, UnitTypesProtossCyberneticsCore) {
			return false
		}
		return true
	case UnitTypesProtossHighTemplar:
		if !f.unitIs(u.UnitType, UnitTypesProtossGateway) {
			return false
		}
		if !f.playerHasCompletedUnit(owner, UnitTypesProtossTemplarArchives) {
			return false
		}
		return true
	case UnitTypesProtossShuttle:
		if !f.unitIs(u.UnitType, UnitTypesProtossRoboticsFacility) {
			return false
		}
		return true
	case UnitTypesProtossScout:
		if !f.unitIs(u.UnitType, UnitTypesProtossStargate) {
			return false
		}
		return true
	case UnitTypesProtossArbiter:
		if !f.unitIs(u.UnitType, UnitTypesProtossStargate) {
			return false
		}
		if !f.playerHasCompletedUnit(owner, UnitTypesProtossArbiterTribunal) {
			return false
		}
		return true
	case UnitTypesProtossCarrier:
		if !f.unitIs(u.UnitType, UnitTypesProtossStargate) {
			return false
		}
		if !f.playerHasCompletedUnit(owner, UnitTypesProtossFleetBeacon) {
			return false
		}
		return true
	case UnitTypesProtossInterceptor:
		if !f.unitIsCarrier(u.UnitType) {
			return false
		}
		if f.unitInterceptorCount(u)+f.unitQueuedFighterUnits(u) >= f.unitMaxInterceptorCount(u) {
			return false
		}
		return true
	case UnitTypesProtossReaver:
		if !f.unitIs(u.UnitType, UnitTypesProtossRoboticsFacility) {
			return false
		}
		if !f.playerHasCompletedUnit(owner, UnitTypesProtossRoboticsSupportBay) {
			return false
		}
		return true
	case UnitTypesProtossScarab:
		if !f.unitIsReaver(u.UnitType) {
			return false
		}
		if f.unitScarabCount(u)+f.unitQueuedFighterUnits(u) >= f.unitMaxScarabCount(u) {
			return false
		}
		return true
	case UnitTypesProtossObserver:
		if !f.unitIs(u.UnitType, UnitTypesProtossRoboticsFacility) {
			return false
		}
		if !f.playerHasCompletedUnit(owner, UnitTypesProtossObservatory) {
			return false
		}
		return true
	case UnitTypesProtossCorsair:
		if !f.unitIs(u.UnitType, UnitTypesProtossStargate) {
			return false
		}
		return true
	case UnitTypesProtossDarkTemplar:
		if !f.unitIs(u.UnitType, UnitTypesProtossGateway) {
			return false
		}
		if !f.playerHasCompletedUnit(owner, UnitTypesProtossTemplarArchives) {
			return false
		}
		return true
	case UnitTypesTerranCommandCenter:
		if !f.unitIs(u.UnitType, UnitTypesTerranSCV) {
			return false
		}
		if f.unitIs(u.UnitType, UnitTypesTerranSCV) && u.OrderType.ID == OrdersConstructingBuilding {
			return false
		}
		return true
	case UnitTypesTerranComsatStation:
		if !f.unitIs(u.UnitType, UnitTypesTerranCommandCenter) {
			return false
		}
		if f.unitAddon(u) != nil {
			return false
		}
		if f.unitIsConstructing(u) || f.unitIsResearching(u) || f.unitIsUpgrading(u) {
			return false
		}
		if !f.playerHasCompletedUnit(owner, UnitTypesTerranAcademy) {
			return false
		}
		return true
	case UnitTypesTerranNuclearSilo:
		if !f.unitIs(u.UnitType, UnitTypesTerranCommandCenter) {
			return false
		}
		if f.unitAddon(u) != nil {
			return false
		}
		if f.unitIsConstructing(u) || f.unitIsResearching(u) || f.unitIsUpgrading(u) {
			return false
		}
		if !f.playerHasCompletedUnit(owner, UnitTypesTerranCovertOps) {
			return false
		}
		return true
	case UnitTypesTerranSupplyDepot:
		if !f.unitIs(u.UnitType, UnitTypesTerranSCV) {
			return false
		}
		if f.unitIs(u.UnitType, UnitTypesTerranSCV) && u.OrderType.ID == OrdersConstructingBuilding {
			return false
		}
		return true
	case UnitTypesTerranRefinery:
		if !f.unitIs(u.UnitType, UnitTypesTerranSCV) {
			return false
		}
		if f.unitIs(u.UnitType, UnitTypesTerranSCV) && u.OrderType.ID == OrdersConstructingBuilding {
			return false
		}
		return true
	case UnitTypesTerranBarracks:
		if !f.unitIs(u.UnitType, UnitTypesTerranSCV) {
			return false
		}
		if f.unitIs(u.UnitType, UnitTypesTerranSCV) && u.OrderType.ID == OrdersConstructingBuilding {
			return false
		}
		if !f.playerHasCompletedUnit(owner, UnitTypesTerranCommandCenter) {
			return false
		}
		return true
	case UnitTypesTerranAcademy:
		if !f.unitIs(u.UnitType, UnitTypesTerranSCV) {
			return false
		}
		if f.unitIs(u.UnitType, UnitTypesTerranSCV) && u.OrderType.ID == OrdersConstructingBuilding {
			return false
		}
		if !f.playerHasCompletedUnit(owner, UnitTypesTerranBarracks) {
			return false
		}
		return true
	case UnitTypesTerranFactory:
		if !f.unitIs(u.UnitType, UnitTypesTerranSCV) {
			return false
		}
		if f.unitIs(u.UnitType, UnitTypesTerranSCV) && u.OrderType.ID == OrdersConstructingBuilding {
			return false
		}
		if !f.playerHasCompletedUnit(owner, UnitTypesTerranBarracks) {
			return false
		}
		return true
	case UnitTypesTerranStarport:
		if !f.unitIs(u.UnitType, UnitTypesTerranSCV) {
			return false
		}
		if f.unitIs(u.UnitType, UnitTypesTerranSCV) && u.OrderType.ID == OrdersConstructingBuilding {
			return false
		}
		if !f.playerHasCompletedUnit(owner, UnitTypesTerranFactory) {
			return false
		}
		return true
	case UnitTypesTerranControlTower:
		if !f.unitIs(u.UnitType, UnitTypesTerranStarport) {
			return false
		}
		if f.unitAddon(u) != nil {
			return false
		}
		if f.unitIsConstructing(u) || f.unitIsResearching(u) || f.unitIsUpgrading(u) {
			return false
		}
		return true
	case UnitTypesTerranScienceFacility:
		if !f.unitIs(u.UnitType, UnitTypesTerranSCV) {
			return false
		}
		if f.unitIs(u.UnitType, UnitTypesTerranSCV) && u.OrderType.ID == OrdersConstructingBuilding {
			return false
		}
		if !f.playerHasCompletedUnit(owner, UnitTypesTerranStarport) {
			return false
		}
		return true
	case UnitTypesTerranCovertOps:
		if !f.unitIs(u.UnitType, UnitTypesTerranScienceFacility) {
			return false
		}
		if f.unitAddon(u) != nil {
			return false
		}
		if f.unitIsConstructing(u) || f.unitIsResearching(u) || f.unitIsUpgrading(u) {
			return false
		}
		return true
	case UnitTypesTerranPhysicsLab:
		if !f.unitIs(u.UnitType, UnitTypesTerranScienceFacility) {
			return false
		}
		if f.unitAddon(u) != nil {
			return false
		}
		if f.unitIsConstructing(u) || f.unitIsResearching(u) || f.unitIsUpgrading(u) {
			return false
		}
		return true
	case UnitTypesTerranMachineShop:
		if !f.unitIs(u.UnitType, UnitTypesTerranFactory) {
			return false
		}
		if f.unitAddon(u) != nil {
			return false
		}
		if f.unitIsConstructing(u) || f.unitIsResearching(u) || f.unitIsUpgrading(u) {
			return false
		}
		return true
	case UnitTypesTerranEngineeringBay:
		if !f.unitIs(u.UnitType, UnitTypesTerranSCV) {
			return false
		}
		if f.unitIs(u.UnitType, UnitTypesTerranSCV) && u.OrderType.ID == OrdersConstructingBuilding {
			return false
		}
		if !f.playerHasCompletedUnit(owner, UnitTypesTerranCommandCenter) {
			return false
		}
		return true
	case UnitTypesTerranArmory:
		if !f.unitIs(u.UnitType, UnitTypesTerranSCV) {
			return false
		}
		if f.unitIs(u.UnitType, UnitTypesTerranSCV) && u.OrderType.ID == OrdersConstructingBuilding {
			return false
		}
		if !f.playerHasCompletedUnit(owner, UnitTypesTerranFactory) {
			return false
		}
		return true
	case UnitTypesTerranMissileTurret:
		if !f.unitIs(u.UnitType, UnitTypesTerranSCV) {
			return false
		}
		if f.unitIs(u.UnitType, UnitTypesTerranSCV) && u.OrderType.ID == OrdersConstructingBuilding {
			return false
		}
		if !f.playerHasCompletedUnit(owner, UnitTypesTerranEngineeringBay) {
			return false
		}
		return true
	case UnitTypesTerranBunker:
		if !f.unitIs(u.UnitType, UnitTypesTerranSCV) {
			return false
		}
		if f.unitIs(u.UnitType, UnitTypesTerranSCV) && u.OrderType.ID == OrdersConstructingBuilding {
			return false
		}
		if !f.playerHasCompletedUnit(owner, UnitTypesTerranBarracks) {
			return false
		}
		return true
	case UnitTypesZergHatchery:
		if !f.unitIs(u.UnitType, UnitTypesZergDrone) {
			return false
		}
		if f.uBurrowed(u) {
			return false
		}
		return true
	case UnitTypesZergLair:
		if !f.unitIs(u.UnitType, UnitTypesZergHatchery) {
			return false
		}
		if !f.playerHasCompletedUnit(owner, UnitTypesZergSpawningPool) {
			return false
		}
		if f.unitIsConstructing(u) || f.unitIsResearching(u) || f.unitIsUpgrading(u) {
			return false
		}
		return true
	case UnitTypesZergHive:
		if !f.unitIs(u.UnitType, UnitTypesZergLair) {
			return false
		}
		if !f.playerHasCompletedUnit(owner, UnitTypesZergQueensNest) {
			return false
		}
		if f.unitIsConstructing(u) || f.unitIsResearching(u) || f.unitIsUpgrading(u) {
			return false
		}
		return true
	case UnitTypesZergSpawningPool:
		if !f.unitIs(u.UnitType, UnitTypesZergDrone) {
			return false
		}
		if f.uBurrowed(u) {
			return false
		}
		if !f.playerHasCompletedUnit(owner, UnitTypesZergHatchery) && !f.playerHasUnit(owner, UnitTypesZergLair) && !f.playerHasUnit(owner, UnitTypesZergHive) {
			return false
		}
		return true
	case UnitTypesZergHydraliskDen:
		if !f.unitIs(u.UnitType, UnitTypesZergDrone) {
			return false
		}
		if f.uBurrowed(u) {
			return false
		}
		if !f.playerHasCompletedUnit(owner, UnitTypesZergSpawningPool) {
			return false
		}
		return true
	case UnitTypesZergDefilerMound:
		if !f.unitIs(u.UnitType, UnitTypesZergDrone) {
			return false
		}
		if f.uBurrowed(u) {
			return false
		}
		if !f.playerHasCompletedUnit(owner, UnitTypesZergHive) {
			return false
		}
		return true
	case UnitTypesZergGreaterSpire:
		if !f.unitIs(u.UnitType, UnitTypesZergSpire) {
			return false
		}
		if f.unitIsConstructing(u) || f.unitIsResearching(u) || f.unitIsUpgrading(u) {
			return false
		}
		if !f.playerHasCompletedUnit(owner, UnitTypesZergHive) {
			return false
		}
		return true
	case UnitTypesZergQueensNest:
		if !f.unitIs(u.UnitType, UnitTypesZergDrone) {
			return false
		}
		if f.uBurrowed(u) {
			return false
		}
		if !f.playerHasCompletedUnit(owner, UnitTypesZergLair) && !f.playerHasUnit(owner, UnitTypesZergHive) {
			return false
		}
		return true
	case UnitTypesZergEvolutionChamber:
		if !f.unitIs(u.UnitType, UnitTypesZergDrone) {
			return false
		}
		if f.uBurrowed(u) {
			return false
		}
		if !f.playerHasCompletedUnit(owner, UnitTypesZergHatchery) && !f.playerHasUnit(owner, UnitTypesZergLair) && !f.playerHasUnit(owner, UnitTypesZergHive) {
			return false
		}
		return true
	case UnitTypesZergUltraliskCavern:
		if !f.unitIs(u.UnitType, UnitTypesZergDrone) {
			return false
		}
		if f.uBurrowed(u) {
			return false
		}
		if !f.playerHasCompletedUnit(owner, UnitTypesZergHive) {
			return false
		}
		return true
	case UnitTypesZergSpire:
		if !f.unitIs(u.UnitType, UnitTypesZergDrone) {
			return false
		}
		if f.uBurrowed(u) {
			return false
		}
		if !f.playerHasCompletedUnit(owner, UnitTypesZergLair) && !f.playerHasUnit(owner, UnitTypesZergHive) {
			return false
		}
		return true
	case UnitTypesZergNydusCanal:
		if !f.unitIs(u.UnitType, UnitTypesZergDrone) {
			return false
		}
		if f.uBurrowed(u) {
			return false
		}
		if !f.playerHasCompletedUnit(owner, UnitTypesZergHive) {
			return false
		}
		return true
	case UnitTypesZergCreepColony:
		if !f.unitIs(u.UnitType, UnitTypesZergDrone) {
			return false
		}
		if f.uBurrowed(u) {
			return false
		}
		return true
	case UnitTypesZergSporeColony:
		if !f.unitIs(u.UnitType, UnitTypesZergCreepColony) {
			return false
		}
		if !f.playerHasCompletedUnit(owner, UnitTypesZergEvolutionChamber) {
			return false
		}
		return true
	case UnitTypesZergSunkenColony:
		if !f.unitIs(u.UnitType, UnitTypesZergCreepColony) {
			return false
		}
		if !f.playerHasCompletedUnit(owner, UnitTypesZergSpawningPool) {
			return false
		}
		return true
	case UnitTypesZergExtractor:
		if !f.unitIs(u.UnitType, UnitTypesZergDrone) {
			return false
		}
		return true
	case UnitTypesProtossNexus:
		if !f.unitIs(u.UnitType, UnitTypesProtossProbe) {
			return false
		}
		return true
	case UnitTypesProtossRoboticsFacility:
		if !f.unitIs(u.UnitType, UnitTypesProtossProbe) {
			return false
		}
		if !f.playerHasCompletedUnit(owner, UnitTypesProtossCyberneticsCore) {
			return false
		}
		return true
	case UnitTypesProtossPylon:
		if !f.unitIs(u.UnitType, UnitTypesProtossProbe) {
			return false
		}
		return true
	case UnitTypesProtossAssimilator:
		if !f.unitIs(u.UnitType, UnitTypesProtossProbe) {
			return false
		}
		return true
	case UnitTypesProtossObservatory:
		if !f.unitIs(u.UnitType, UnitTypesProtossProbe) {
			return false
		}
		if !f.playerHasCompletedUnit(owner, UnitTypesProtossRoboticsFacility) {
			return false
		}
		return true
	case UnitTypesProtossGateway:
		if !f.unitIs(u.UnitType, UnitTypesProtossProbe) {
			return false
		}
		if !f.playerHasCompletedUnit(owner, UnitTypesProtossNexus) {
			return false
		}
		return true
	case UnitTypesProtossPhotonCannon:
		if !f.unitIs(u.UnitType, UnitTypesProtossProbe) {
			return false
		}
		if !f.playerHasCompletedUnit(owner, UnitTypesProtossForge) {
			return false
		}
		return true
	case UnitTypesProtossCitadelOfAdun:
		if !f.unitIs(u.UnitType, UnitTypesProtossProbe) {
			return false
		}
		if !f.playerHasCompletedUnit(owner, UnitTypesProtossCyberneticsCore) {
			return false
		}
		return true
	case UnitTypesProtossCyberneticsCore:
		if !f.unitIs(u.UnitType, UnitTypesProtossProbe) {
			return false
		}
		if !f.playerHasCompletedUnit(owner, UnitTypesProtossGateway) {
			return false
		}
		return true
	case UnitTypesProtossTemplarArchives:
		if !f.unitIs(u.UnitType, UnitTypesProtossProbe) {
			return false
		}
		if !f.playerHasCompletedUnit(owner, UnitTypesProtossCitadelOfAdun) {
			return false
		}
		return true
	case UnitTypesProtossForge:
		if !f.unitIs(u.UnitType, UnitTypesProtossProbe) {
			return false
		}
		if !f.playerHasCompletedUnit(owner, UnitTypesProtossNexus) {
			return false
		}
		return true
	case UnitTypesProtossStargate:
		if !f.unitIs(u.UnitType, UnitTypesProtossProbe) {
			return false
		}
		if !f.playerHasCompletedUnit(owner, UnitTypesProtossCyberneticsCore) {
			return false
		}
		return true
	case UnitTypesProtossFleetBeacon:
		if !f.unitIs(u.UnitType, UnitTypesProtossProbe) {
			return false
		}
		if !f.playerHasCompletedUnit(owner, UnitTypesProtossStargate) {
			return false
		}
		return true
	case UnitTypesProtossArbiterTribunal:
		if !f.unitIs(u.UnitType, UnitTypesProtossProbe) {
			return false
		}
		if !f.playerHasCompletedUnit(owner, UnitTypesProtossStargate) {
			return false
		}
		if !f.playerHasCompletedUnit(owner, UnitTypesProtossTemplarArchives) {
			return false
		}
		return true
	case UnitTypesProtossRoboticsSupportBay:
		if !f.unitIs(u.UnitType, UnitTypesProtossProbe) {
			return false
		}
		if !f.playerHasCompletedUnit(owner, UnitTypesProtossRoboticsFacility) {
			return false
		}
		return true
	case UnitTypesProtossShieldBattery:
		if !f.unitIs(u.UnitType, UnitTypesProtossProbe) {
			return false
		}
		if !f.playerHasCompletedUnit(owner, UnitTypesProtossGateway) {
			return false
		}
		return true
	default:
		return false
	}
}

func (f *StateFunctions) unitCanReceiveOrder(u *Unit, order *OrderTypeT, owner int32) bool {
	if u.Owner != owner {
		return false
	}
	if order.ID != OrdersRallyPointUnit && order.ID != OrdersRallyPointTile {
		if !f.uCompleted(u) {
			return false
		}
	}
	if f.unitIsDisabled(u) {
		return false
	}
	if !f.unitIs(u.UnitType, UnitTypesZergLurker) && f.uBurrowed(u) {
		if order.ID != OrdersUnburrowing {
			return false
		}
	}
	if f.unitIs(u.UnitType, UnitTypesTerranSCV) && u.OrderType.ID == OrdersConstructingBuilding && order.ID != OrdersStop {
		return false
	}
	if f.unitIsGhost(u.UnitType) && u.OrderType.ID == OrdersNukeTrack {
		return false
	}
	if f.unitIs(u.UnitType, UnitTypesProtossArchon) || f.unitIs(u.UnitType, UnitTypesProtossDarkArchon) {
		if u.OrderType.ID == OrdersCompletingArchonSummon {
			return false
		}
	}
	return f.unitOrderAllowed(u, order)
}

func (f *StateFunctions) unitOrderAllowed(u *Unit, order *OrderTypeT) bool {
	switch order.ID {
	case OrdersDie:
		return true
	case OrdersStop:
		return true
	case OrdersGuard:
		if f.uGroundedBuilding(u) {
			return false
		}
		return true
	case OrdersPlayerGuard:
		if f.uGroundedBuilding(u) {
			return false
		}
		return true
	case OrdersTurretGuard:
		if !f.utTurret(u.UnitType) {
			return false
		}
		return true
	case OrdersBunkerGuard:
		if !f.unitIs(u.UnitType, UnitTypesTerranBunker) {
			return false
		}
		if f.uHallucination(u) {
			return false
		}
		return true
	case OrdersMove:
		if !f.unitCanMove(u) {
			return false
		}
		return true
	case OrdersInterceptorAttack:
		if !f.unitIs(u.UnitType, UnitTypesProtossInterceptor) {
			return false
		}
		return true
	case OrdersScarabAttack:
		if !f.unitIs(u.UnitType, UnitTypesProtossScarab) {
			return false
		}
		if f.uHallucination(u) {
			return false
		}
		return true
	case OrdersSpiderMine:
		if !f.unitIs(u.UnitType, UnitTypesTerranVultureSpiderMine) {
			return false
		}
		if f.uHallucination(u) {
			return false
		}
		return true
	case OrdersCarrier:
		if !f.unitIsCarrier(u.UnitType) {
			return false
		}
		return true
	case OrdersReaver:
		if !f.unitIsReaver(u.UnitType) {
			return false
		}
		return true
	case OrdersReaverCarrierMove:
		if !f.unitIsCarrier(u.UnitType) && !f.unitIsReaver(u.UnitType) {
			return false
		}
		return true
	case OrdersCarrierStop:
		if !f.unitIsCarrier(u.UnitType) {
			return false
		}
		return true
	case OrdersReaverStop:
		if !f.unitIsReaver(u.UnitType) {
			return false
		}
		return true
	case OrdersCarrierHoldPosition:
		if !f.unitIsCarrier(u.UnitType) {
			return false
		}
		return true
	case OrdersReaverHoldPosition:
		if !f.unitIsReaver(u.UnitType) {
			return false
		}
		return true
	case OrdersCarrierAttack:
		if !f.unitIsCarrier(u.UnitType) {
			return false
		}
		return true
	case OrdersReaverAttack:
		if !f.unitIsReaver(u.UnitType) {
			return false
		}
		return true
	case OrdersCarrierMoveToAttack:
		if !f.unitIsCarrier(u.UnitType) {
			return false
		}
		return true
	case OrdersReaverMoveToAttack:
		if !f.unitIsReaver(u.UnitType) {
			return false
		}
		return true
	case OrdersCarrierIgnore2:
		if !f.unitIsCarrier(u.UnitType) && !f.unitIsReaver(u.UnitType) {
			return false
		}
		return true
	case OrdersCarrierFight:
		if !f.unitIsCarrier(u.UnitType) {
			return false
		}
		return true
	case OrdersReaverFight:
		if !f.unitIsReaver(u.UnitType) {
			return false
		}
		return true
	case OrdersAttackDefault:
		if f.uGroundedBuilding(u) {
			return false
		}
		return true
	case OrdersMoveToAttack:
		if f.uGroundedBuilding(u) {
			return false
		}
		return true
	case OrdersAttackUnit:
		if !f.unitCanMove(u) {
			return false
		}
		return true
	case OrdersAttackFixedRange:
		if f.uGroundedBuilding(u) {
			return false
		}
		return true
	case OrdersAttackTile:
		if f.uGroundedBuilding(u) {
			return false
		}
		return true
	case OrdersHover:
		if !f.unitCanMove(u) {
			return false
		}
		return true
	case OrdersAttackMove:
		if !f.unitCanMove(u) {
			return false
		}
		return true
	case OrdersAtkMoveEP:
		if !f.unitCanMove(u) {
			return false
		}
		return true
	case OrdersStayInRange:
		if !f.unitCanMove(u) {
			return false
		}
		return true
	case OrdersTowerGuard:
		if !f.unitIs(u.UnitType, UnitTypesTerranMissileTurret) && !f.unitIs(u.UnitType, UnitTypesZergSporeColony) && !f.unitIs(u.UnitType, UnitTypesZergSunkenColony) && !f.unitIs(u.UnitType, UnitTypesProtossPhotonCannon) {
			return false
		}
		if f.uHallucination(u) {
			return false
		}
		return true
	case OrdersTowerAttack:
		if !f.unitIs(u.UnitType, UnitTypesTerranMissileTurret) && !f.unitIs(u.UnitType, UnitTypesZergSporeColony) && !f.unitIs(u.UnitType, UnitTypesZergSunkenColony) && !f.unitIs(u.UnitType, UnitTypesProtossPhotonCannon) && !f.unitIsTrap(u.UnitType) {
			return false
		}
		if f.uHallucination(u) {
			return false
		}
		return true
	case OrdersTurretAttack:
		if !f.utTurret(u.UnitType) {
			return false
		}
		return true
	case OrdersNothing:
		return true
	case OrdersDroneStartBuild:
		if !f.unitIs(u.UnitType, UnitTypesZergDrone) {
			return false
		}
		if f.uHallucination(u) {
			return false
		}
		return true
	case OrdersDroneBuild:
		if !f.unitIs(u.UnitType, UnitTypesZergDrone) {
			return false
		}
		if f.uHallucination(u) {
			return false
		}
		return true
	case OrdersDroneLand:
		if !f.unitIs(u.UnitType, UnitTypesZergDrone) {
			return false
		}
		if f.uHallucination(u) {
			return false
		}
		return true
	case OrdersPlaceBuilding:
		if !f.unitIs(u.UnitType, UnitTypesTerranSCV) {
			return false
		}
		if f.uHallucination(u) {
			return false
		}
		return true
	case OrdersPlaceProtossBuilding:
		if !f.unitIs(u.UnitType, UnitTypesProtossProbe) {
			return false
		}
		if f.uHallucination(u) {
			return false
		}
		return true
	case OrdersCreateProtossBuilding:
		if !f.unitIs(u.UnitType, UnitTypesProtossProbe) {
			return false
		}
		if f.uHallucination(u) {
			return false
		}
		return true
	case OrdersPlaceAddon:
		if f.uHallucination(u) {
			return false
		}
		return true
	case OrdersRallyPointUnit:
		if !f.uGroundedBuilding(u) {
			return false
		}
		if f.uHallucination(u) {
			return false
		}
		return true
	case OrdersRallyPointTile:
		if !f.uGroundedBuilding(u) {
			return false
		}
		if f.uHallucination(u) {
			return false
		}
		return true
	case OrdersInfestedCommandCenter:
		if !f.unitIs(u.UnitType, UnitTypesTerranCommandCenter) {
			return false
		}
		if f.uHallucination(u) {
			return false
		}
		return true
	case OrdersBuildNydusExit:
		if !f.unitIsNydus(u.UnitType) {
			return false
		}
		if u.Building.Nydus.Exit != nil {
			return false
		}
		if f.uHallucination(u) {
			return false
		}
		return true
	case OrdersQueenHoldPosition:
		if !f.unitIsQueen(u.UnitType) {
			return false
		}
		if f.uHallucination(u) {
			return false
		}
		return true
	case OrdersCastInfestation:
		if !f.unitIsQueen(u.UnitType) {
			return false
		}
		if f.uHallucination(u) {
			return false
		}
		return true
	case OrdersMoveToInfest:
		if !f.unitIsQueen(u.UnitType) {
			return false
		}
		if f.uHallucination(u) {
			return false
		}
		return true
	case OrdersInfestingCommandCenter:
		if !f.unitIsQueen(u.UnitType) {
			return false
		}
		if f.uHallucination(u) {
			return false
		}
		return true
	case OrdersConstructingBuilding:
		if !f.unitIs(u.UnitType, UnitTypesTerranSCV) {
			return false
		}
		if f.uHallucination(u) {
			return false
		}
		return true
	case OrdersRepair:
		if !f.unitIs(u.UnitType, UnitTypesTerranSCV) {
			return false
		}
		if f.uHallucination(u) {
			return false
		}
		return true
	case OrdersMoveToRepair:
		if !f.unitIs(u.UnitType, UnitTypesTerranSCV) {
			return false
		}
		if f.uHallucination(u) {
			return false
		}
		return true
	case OrdersZergBirth:
		if !f.unitIs(u.UnitType, UnitTypesZergEgg) && !f.unitIs(u.UnitType, UnitTypesZergCocoon) {
			return false
		}
		if f.uHallucination(u) {
			return false
		}
		return true
	case OrdersFollow:
		if !f.unitCanMove(u) {
			return false
		}
		return true
	case OrdersInterceptorReturn:
		if !f.utWorker(u.UnitType) {
			return false
		}
		if f.uHallucination(u) {
			return false
		}
		return true
	case OrdersBuildingLand:
		if !f.utFlyingBuilding(u.UnitType) {
			return false
		}
		if f.uGroundedBuilding(u) {
			return false
		}
		if f.uHallucination(u) {
			return false
		}
		return true
	case OrdersBuildingLiftoff:
		if !f.utFlyingBuilding(u.UnitType) {
			return false
		}
		if !f.uGroundedBuilding(u) {
			return false
		}
		if f.uHallucination(u) {
			return false
		}
		return true
	case OrdersLiftingOff:
		if !f.utFlyingBuilding(u.UnitType) {
			return false
		}
		if f.uHallucination(u) {
			return false
		}
		return true
	case OrdersCTFCOP2:
		if !f.unitIs(u.UnitType, UnitTypesSpecialZergFlagBeacon) && !f.unitIs(u.UnitType, UnitTypesSpecialTerranFlagBeacon) && !f.unitIs(u.UnitType, UnitTypesSpecialProtossFlagBeacon) {
			return false
		}
		if f.uHallucination(u) {
			return false
		}
		return true
	case OrdersLarva:
		if !f.unitIs(u.UnitType, UnitTypesZergLarva) {
			return false
		}
		return true
	case OrdersSpawningLarva:
		if f.unitIsHatchery(u.UnitType) {
			return false
		}
		if f.uHallucination(u) {
			return false
		}
		return true
	case OrdersHarvest1:
		if !f.utWorker(u.UnitType) {
			return false
		}
		if f.uHallucination(u) {
			return false
		}
		return true
	case OrdersHarvest2:
		if !f.utWorker(u.UnitType) {
			return false
		}
		if f.uHallucination(u) {
			return false
		}
		return true
	case OrdersMoveToGas:
		if !f.utWorker(u.UnitType) {
			return false
		}
		if f.uHallucination(u) {
			return false
		}
		return true
	case OrdersWaitForGas:
		if !f.utWorker(u.UnitType) {
			return false
		}
		if f.uHallucination(u) {
			return false
		}
		return true
	case OrdersReturnGas:
		if !f.utWorker(u.UnitType) {
			return false
		}
		if f.uHallucination(u) {
			return false
		}
		return true
	case OrdersMoveToMinerals:
		if !f.utWorker(u.UnitType) {
			return false
		}
		if f.uHallucination(u) {
			return false
		}
		return true
	case OrdersWaitForMinerals:
		if !f.utWorker(u.UnitType) {
			return false
		}
		if f.uHallucination(u) {
			return false
		}
		return true
	case OrdersReturnMinerals:
		if !f.utWorker(u.UnitType) {
			return false
		}
		if f.uHallucination(u) {
			return false
		}
		return true
	case OrdersEnterTransport:
		return true
	case OrdersUnload:
		if !f.unitProvidesSpace(u) {
			return false
		}
		if f.uHallucination(u) {
			return false
		}
		return true
	case OrdersMoveUnload:
		if !f.unitProvidesSpace(u) {
			return false
		}
		if f.uHallucination(u) {
			return false
		}
		return true
	case OrdersPickupTransport:
		if !f.unitProvidesSpace(u) {
			return false
		}
		if f.uHallucination(u) {
			return false
		}
		return true
	case OrdersPickupBunker:
		if !f.unitIs(u.UnitType, UnitTypesTerranBunker) {
			return false
		}
		if f.uHallucination(u) {
			return false
		}
		return true
	case OrdersPickupIdle:
		if !f.unitProvidesSpace(u) && !f.unitIs(u.UnitType, UnitTypesZergOverlord) {
			return false
		}
		if f.uHallucination(u) {
			return false
		}
		return true
	case OrdersPowerupIdle:
		if !f.utPowerup(u.UnitType) {
			return false
		}
		if f.uHallucination(u) {
			return false
		}
		return true
	case OrdersWatchTarget:
		return true
	case OrdersInitCreepGrowth:
		if !f.unitIs(u.UnitType, UnitTypesZergCreepColony) && !f.unitIs(u.UnitType, UnitTypesZergHatchery) && !f.unitIs(u.UnitType, UnitTypesZergLair) && !f.unitIs(u.UnitType, UnitTypesZergHive) {
			return false
		}
		if f.uHallucination(u) {
			return false
		}
		return true
	case OrdersStoppingCreepGrowth:
		if !f.unitIs(u.UnitType, UnitTypesZergSporeColony) && !f.unitIs(u.UnitType, UnitTypesZergSunkenColony) {
			return false
		}
		if f.uHallucination(u) {
			return false
		}
		return true
	case OrdersSpreadCreep:
		if f.uHallucination(u) {
			return false
		}
		return true
	case OrdersGuardianAspect:
		if !f.unitIs(u.UnitType, UnitTypesZergMutalisk) {
			return false
		}
		if f.uHallucination(u) {
			return false
		}
		return true
	case OrdersHoldPosition:
		if !f.unitCanHoldPosition(u) {
			return false
		}
		return true
	case OrdersPatrol:
		if !f.unitCanMove(u) {
			return false
		}
		return true
	case OrdersRechargeShieldsBattery:
		if !f.unitIs(u.UnitType, UnitTypesProtossShieldBattery) {
			return false
		}
		if f.uHallucination(u) {
			return false
		}
		return true
	case OrdersScanner:
		if !f.unitIs(u.UnitType, UnitTypesSpellScannerSweep) {
			return false
		}
		if f.uHallucination(u) {
			return false
		}
		return true
	case OrdersNukeWait:
		if !f.unitIs(u.UnitType, UnitTypesTerranNuclearMissile) {
			return false
		}
		if f.uHallucination(u) {
			return false
		}
		return true
	case OrdersNukeLaunch:
		if !f.unitIs(u.UnitType, UnitTypesTerranNuclearMissile) {
			return false
		}
		if f.uHallucination(u) {
			return false
		}
		return true
	case OrdersNukePaint:
		if !f.unitIs(u.UnitType, UnitTypesTerranGhost) {
			return false
		}
		if f.uHallucination(u) {
			return false
		}
		return true
	case OrdersNukeUnit:
		if !f.unitIs(u.UnitType, UnitTypesTerranNuclearMissile) {
			return false
		}
		if f.uHallucination(u) {
			return false
		}
		return true
	case OrdersCastNuclearStrike:
		if !f.unitIs(u.UnitType, UnitTypesTerranNuclearMissile) {
			return false
		}
		if f.uHallucination(u) {
			return false
		}
		return true
	case OrdersNukeTrack:
		if !f.unitIs(u.UnitType, UnitTypesTerranNuclearMissile) {
			return false
		}
		if f.uHallucination(u) {
			return false
		}
		return true
	case OrdersInitializeArbiter:
		if !f.unitIsArbiter(u.UnitType) {
			return false
		}
		return true
	case OrdersCloakNearbyUnits:
		if !f.unitIsArbiter(u.UnitType) {
			return false
		}
		if f.uHallucination(u) {
			return false
		}
		return true
	case OrdersRightClickAction:
		return true
	case OrdersSuicideUnit:
		if !f.unitIs(u.UnitType, UnitTypesZergInfestedTerran) {
			return false
		}
		return true
	case OrdersSuicideLocation:
		if !f.unitIs(u.UnitType, UnitTypesZergInfestedTerran) {
			return false
		}
		return true
	case OrdersSuicideHoldPosition:
		if !f.unitIs(u.UnitType, UnitTypesZergInfestedTerran) && !f.unitIs(u.UnitType, UnitTypesZergScourge) {
			return false
		}
		return true
	case OrdersSelfDestructing:
		if !f.unitIs(u.UnitType, UnitTypesProtossScarab) {
			return false
		}
		if f.uHallucination(u) {
			return false
		}
		return true
	case OrdersCritter:
		if !f.unitIsCritter(u.UnitType) {
			return false
		}
		return true
	case OrdersHiddenGun:
		if !f.unitIsTrap(u.UnitType) {
			return false
		}
		if f.uHallucination(u) {
			return false
		}
		return true
	case OrdersOpenDoor:
		if !f.unitIs(u.UnitType, UnitTypesSpecialUpperLevelDoor) && !f.unitIs(u.UnitType, UnitTypesSpecialRightUpperLevelDoor) && !f.unitIs(u.UnitType, UnitTypesSpecialPitDoor) && !f.unitIs(u.UnitType, UnitTypesSpecialRightPitDoor) {
			return false
		}
		if f.uHallucination(u) {
			return false
		}
		return true
	case OrdersCloseDoor:
		if !f.unitIs(u.UnitType, UnitTypesSpecialUpperLevelDoor) && !f.unitIs(u.UnitType, UnitTypesSpecialRightUpperLevelDoor) && !f.unitIs(u.UnitType, UnitTypesSpecialPitDoor) && !f.unitIs(u.UnitType, UnitTypesSpecialRightPitDoor) {
			return false
		}
		if f.uHallucination(u) {
			return false
		}
		return true
	case OrdersHideTrap:
		if !f.unitIsTrap(u.UnitType) {
			return false
		}
		if f.uHallucination(u) {
			return false
		}
		return true
	case OrdersRevealTrap:
		if !f.unitIsTrap(u.UnitType) {
			return false
		}
		if f.uHallucination(u) {
			return false
		}
		return true
	case OrdersMedicIdle:
		if !f.unitIs(u.UnitType, UnitTypesTerranMedic) {
			return false
		}
		if f.uHallucination(u) {
			return false
		}
		return true
	case OrdersMedicHeal:
		if !f.unitIs(u.UnitType, UnitTypesTerranMedic) {
			return false
		}
		if f.uHallucination(u) {
			return false
		}
		return true
	case OrdersHealMove:
		if !f.unitIs(u.UnitType, UnitTypesTerranMedic) {
			return false
		}
		if f.uHallucination(u) {
			return false
		}
		return true
	case OrdersMedicHoldPosition:
		if !f.unitIs(u.UnitType, UnitTypesTerranMedic) {
			return false
		}
		if f.uHallucination(u) {
			return false
		}
		return true
	case OrdersCastMindControl:
		if !f.unitIs(u.UnitType, UnitTypesProtossDarkArchon) {
			return false
		}
		if f.uHallucination(u) {
			return false
		}
		return true
	default:
		return false
	}
}

func (f *StateFunctions) unitBuildOrderValid(u *Unit, order *OrderTypeT, unitType *UnitTypeT, owner int32) bool {
	switch order.ID {
	case OrdersDroneStartBuild, OrdersPlaceBuilding, OrdersPlaceProtossBuilding, OrdersCreateProtossBuilding, OrdersPlaceAddon:
		if !f.unitCanBuild(u, unitType) {
			return false
		}
		if !f.unitCanReceiveOrder(u, order, owner) {
			return false
		}
		return true
	case OrdersCTFCOP2:
		if f.st.CurrentFrame > 600 {
			return false
		}
		if !f.unitIs(u.UnitType, unitType.ID) {
			return false
		}
		if !f.unitCanReceiveOrder(u, order, owner) {
			return false
		}
		return true
	case OrdersBuildNydusExit, OrdersBuildingLand:
		if !f.unitIs(u.UnitType, unitType.ID) {
			return false
		}
		if !f.unitCanReceiveOrder(u, order, owner) {
			return false
		}
		return true
	default:
		return false
	}
}

func (f *StateFunctions) unitCanResearchOwner(u *Unit, tech *TechTypeT, owner int32) bool {
	if u.Owner != owner {
		return false
	}
	return f.unitCanResearch(u, tech)
}

func (f *StateFunctions) unitCanResearch(u *Unit, tech *TechTypeT) bool {
	if !f.uCompleted(u) {
		return false
	}
	if f.uHallucination(u) {
		return false
	}
	if f.unitIsDisabled(u) {
		return false
	}
	owner := u.Owner
	if !f.playerTechAvailable(owner, tech.ID) {
		return false
	}
	if f.playerHasResearched(owner, tech.ID) {
		return false
	}
	if f.playerIsResearching(owner, tech.ID) {
		return false
	}
	switch tech.ID {
	case TechTypesStimPacks:
		if !f.unitIs(u.UnitType, UnitTypesTerranAcademy) {
			return false
		}
		if f.unitIsConstructing(u) || f.unitIsResearching(u) || f.unitIsUpgrading(u) {
			return false
		}
		if !f.uGroundedBuilding(u) {
			return false
		}
		return true
	case TechTypesLockdown:
		if !f.unitIs(u.UnitType, UnitTypesTerranCovertOps) {
			return false
		}
		if f.unitIsConstructing(u) || f.unitIsResearching(u) || f.unitIsUpgrading(u) {
			return false
		}
		if !f.uGroundedBuilding(u) {
			return false
		}
		return true
	case TechTypesSpiderMines:
		if !f.unitIs(u.UnitType, UnitTypesTerranMachineShop) {
			return false
		}
		if f.unitIsConstructing(u) || f.unitIsResearching(u) || f.unitIsUpgrading(u) {
			return false
		}
		if !f.uGroundedBuilding(u) {
			return false
		}
		return true
	case TechTypesTankSiegeMode:
		if !f.unitIs(u.UnitType, UnitTypesTerranMachineShop) {
			return false
		}
		if f.unitIsConstructing(u) || f.unitIsResearching(u) || f.unitIsUpgrading(u) {
			return false
		}
		if !f.uGroundedBuilding(u) {
			return false
		}
		return true
	case TechTypesEMPShockwave:
		if !f.unitIs(u.UnitType, UnitTypesTerranScienceFacility) {
			return false
		}
		if f.unitIsConstructing(u) || f.unitIsResearching(u) || f.unitIsUpgrading(u) {
			return false
		}
		if !f.uGroundedBuilding(u) {
			return false
		}
		return true
	case TechTypesIrradiate:
		if !f.unitIs(u.UnitType, UnitTypesTerranScienceFacility) {
			return false
		}
		if f.unitIsConstructing(u) || f.unitIsResearching(u) || f.unitIsUpgrading(u) {
			return false
		}
		if !f.uGroundedBuilding(u) {
			return false
		}
		return true
	case TechTypesYamatoGun:
		if !f.unitIs(u.UnitType, UnitTypesTerranPhysicsLab) {
			return false
		}
		if f.unitIsConstructing(u) || f.unitIsResearching(u) || f.unitIsUpgrading(u) {
			return false
		}
		if !f.uGroundedBuilding(u) {
			return false
		}
		return true
	case TechTypesCloakingField:
		if !f.unitIs(u.UnitType, UnitTypesTerranControlTower) {
			return false
		}
		if f.unitIsConstructing(u) || f.unitIsResearching(u) || f.unitIsUpgrading(u) {
			return false
		}
		if !f.uGroundedBuilding(u) {
			return false
		}
		return true
	case TechTypesPersonnelCloaking:
		if !f.unitIs(u.UnitType, UnitTypesTerranCovertOps) {
			return false
		}
		if f.unitIsConstructing(u) || f.unitIsResearching(u) || f.unitIsUpgrading(u) {
			return false
		}
		if !f.uGroundedBuilding(u) {
			return false
		}
		return true
	case TechTypesRestoration:
		if !f.unitIs(u.UnitType, UnitTypesTerranAcademy) {
			return false
		}
		if f.unitIsConstructing(u) || f.unitIsResearching(u) || f.unitIsUpgrading(u) {
			return false
		}
		if !f.uGroundedBuilding(u) {
			return false
		}
		return true
	case TechTypesOpticalFlare:
		if !f.unitIs(u.UnitType, UnitTypesTerranAcademy) {
			return false
		}
		if f.unitIsConstructing(u) || f.unitIsResearching(u) || f.unitIsUpgrading(u) {
			return false
		}
		if !f.uGroundedBuilding(u) {
			return false
		}
		return true
	case TechTypesBurrowing:
		if !f.unitIs(u.UnitType, UnitTypesZergHatchery) && !f.unitIs(u.UnitType, UnitTypesZergLair) && !f.unitIs(u.UnitType, UnitTypesZergHive) {
			return false
		}
		if f.unitIsConstructing(u) || f.unitIsResearching(u) || f.unitIsUpgrading(u) {
			return false
		}
		return true
	case TechTypesSpawnBroodlings:
		if !f.unitIs(u.UnitType, UnitTypesZergQueensNest) {
			return false
		}
		if f.unitIsConstructing(u) || f.unitIsResearching(u) || f.unitIsUpgrading(u) {
			return false
		}
		return true
	case TechTypesParasite:
		if !f.unitIs(u.UnitType, UnitTypesZergQueensNest) {
			return false
		}
		if f.unitIsConstructing(u) || f.unitIsResearching(u) || f.unitIsUpgrading(u) {
			return false
		}
		return true
	case TechTypesPlague:
		if !f.unitIs(u.UnitType, UnitTypesZergDefilerMound) {
			return false
		}
		if f.unitIsConstructing(u) || f.unitIsResearching(u) || f.unitIsUpgrading(u) {
			return false
		}
		return true
	case TechTypesConsume:
		if !f.unitIs(u.UnitType, UnitTypesZergDefilerMound) {
			return false
		}
		if f.unitIsConstructing(u) || f.unitIsResearching(u) || f.unitIsUpgrading(u) {
			return false
		}
		return true
	case TechTypesEnsnare:
		if !f.unitIs(u.UnitType, UnitTypesZergQueensNest) {
			return false
		}
		if f.unitIsConstructing(u) || f.unitIsResearching(u) || f.unitIsUpgrading(u) {
			return false
		}
		return true
	case TechTypesLurkerAspect:
		if !f.unitIs(u.UnitType, UnitTypesZergHydraliskDen) {
			return false
		}
		if !f.playerHasCompletedUnit(owner, UnitTypesZergLair) && !f.playerHasCompletedUnit(owner, UnitTypesZergHive) {
			return false
		}
		if f.unitIsConstructing(u) || f.unitIsResearching(u) || f.unitIsUpgrading(u) {
			return false
		}
		return true
	case TechTypesPsionicStorm:
		if !f.unitIs(u.UnitType, UnitTypesProtossTemplarArchives) {
			return false
		}
		if f.unitIsConstructing(u) || f.unitIsResearching(u) || f.unitIsUpgrading(u) {
			return false
		}
		return true
	case TechTypesHallucination:
		if !f.unitIs(u.UnitType, UnitTypesProtossTemplarArchives) {
			return false
		}
		if f.unitIsConstructing(u) || f.unitIsResearching(u) || f.unitIsUpgrading(u) {
			return false
		}
		return true
	case TechTypesRecall:
		if !f.unitIs(u.UnitType, UnitTypesProtossArbiterTribunal) {
			return false
		}
		if f.unitIsConstructing(u) || f.unitIsResearching(u) || f.unitIsUpgrading(u) {
			return false
		}
		return true
	case TechTypesStasisField:
		if !f.unitIs(u.UnitType, UnitTypesProtossArbiterTribunal) {
			return false
		}
		if f.unitIsConstructing(u) || f.unitIsResearching(u) || f.unitIsUpgrading(u) {
			return false
		}
		return true
	case TechTypesDisruptionWeb:
		if !f.unitIs(u.UnitType, UnitTypesProtossFleetBeacon) {
			return false
		}
		if f.unitIsConstructing(u) || f.unitIsResearching(u) || f.unitIsUpgrading(u) {
			return false
		}
		return true
	case TechTypesMindControl:
		if !f.unitIs(u.UnitType, UnitTypesProtossTemplarArchives) {
			return false
		}
		if f.unitIsConstructing(u) || f.unitIsResearching(u) || f.unitIsUpgrading(u) {
			return false
		}
		return true
	case TechTypesFeedback:
		if !f.unitIs(u.UnitType, UnitTypesProtossTemplarArchives) {
			return false
		}
		if f.unitIsConstructing(u) || f.unitIsResearching(u) || f.unitIsUpgrading(u) {
			return false
		}
		return true
	case TechTypesMaelstrom:
		if !f.unitIs(u.UnitType, UnitTypesProtossTemplarArchives) {
			return false
		}
		if f.unitIsConstructing(u) || f.unitIsResearching(u) || f.unitIsUpgrading(u) {
			return false
		}
		return true
	default:
		return false
	}
}

func (f *StateFunctions) unitCanUpgradeOwner(u *Unit, upgrade *UpgradeTypeT, owner int32) bool {
	if u.Owner != owner {
		return false
	}
	return f.unitCanUpgrade(u, upgrade)
}

func (f *StateFunctions) unitCanUpgrade(u *Unit, upgrade *UpgradeTypeT) bool {
	if !f.uCompleted(u) {
		return false
	}
	if f.uHallucination(u) {
		return false
	}
	if f.unitIsDisabled(u) {
		return false
	}
	owner := u.Owner
	if f.playerUpgradeLevel(owner, upgrade.ID) >= f.playerMaxUpgradeLevel(u.Owner, upgrade.ID) {
		return false
	}
	if f.playerIsUpgrading(owner, upgrade.ID) {
		return false
	}
	switch upgrade.ID {
	case UpgradeTypesTerranInfantryArmor:
		if !f.unitIs(u.UnitType, UnitTypesTerranEngineeringBay) {
			return false
		}
		if !f.uGroundedBuilding(u) {
			return false
		}
		if f.unitIsConstructing(u) || f.unitIsResearching(u) || f.unitIsUpgrading(u) {
			return false
		}
		if f.playerUpgradeLevel(owner, UpgradeTypesTerranInfantryArmor) != 0 {
			if !f.playerHasCompletedUnit(owner, UnitTypesTerranScienceFacility) {
				return false
			}
		}
		return true
	case UpgradeTypesTerranVehiclePlating:
		if !f.unitIs(u.UnitType, UnitTypesTerranArmory) {
			return false
		}
		if !f.uGroundedBuilding(u) {
			return false
		}
		if f.unitIsConstructing(u) || f.unitIsResearching(u) || f.unitIsUpgrading(u) {
			return false
		}
		if f.playerUpgradeLevel(owner, UpgradeTypesTerranVehiclePlating) != 0 {
			if !f.playerHasCompletedUnit(owner, UnitTypesTerranScienceFacility) {
				return false
			}
		}
		return true
	case UpgradeTypesTerranShipPlating:
		if !f.unitIs(u.UnitType, UnitTypesTerranArmory) {
			return false
		}
		if !f.uGroundedBuilding(u) {
			return false
		}
		if f.unitIsConstructing(u) || f.unitIsResearching(u) || f.unitIsUpgrading(u) {
			return false
		}
		if f.playerUpgradeLevel(owner, UpgradeTypesTerranShipPlating) != 0 {
			if !f.playerHasCompletedUnit(owner, UnitTypesTerranScienceFacility) {
				return false
			}
		}
		return true
	case UpgradeTypesZergCarapace:
		if !f.unitIs(u.UnitType, UnitTypesZergEvolutionChamber) {
			return false
		}
		if f.unitIsConstructing(u) || f.unitIsResearching(u) || f.unitIsUpgrading(u) {
			return false
		}
		switch f.playerUpgradeLevel(owner, UpgradeTypesZergCarapace) {
		case 1:
			if !f.playerHasCompletedUnit(owner, UnitTypesZergLair) && !f.playerHasUnit(owner, UnitTypesZergHive) {
				return false
			}
		case 2:
			if !f.playerHasCompletedUnit(owner, UnitTypesZergHive) {
				return false
			}
		}
		return true
	case UpgradeTypesZergFlyerCarapace:
		if !f.unitIs(u.UnitType, UnitTypesZergSpire) && !f.unitIs(u.UnitType, UnitTypesZergGreaterSpire) {
			return false
		}
		if f.unitIsConstructing(u) || f.unitIsResearching(u) || f.unitIsUpgrading(u) {
			return false
		}
		switch f.playerUpgradeLevel(owner, UpgradeTypesZergFlyerCarapace) {
		case 1:
			if !f.playerHasCompletedUnit(owner, UnitTypesZergLair) && !f.playerHasUnit(owner, UnitTypesZergHive) {
				return false
			}
		case 2:
			if !f.playerHasCompletedUnit(owner, UnitTypesZergHive) {
				return false
			}
		}
		return true
	case UpgradeTypesProtossGroundArmor:
		if !f.unitIs(u.UnitType, UnitTypesProtossForge) {
			return false
		}
		if f.unitIsConstructing(u) || f.unitIsResearching(u) || f.unitIsUpgrading(u) {
			return false
		}
		if f.playerUpgradeLevel(owner, UpgradeTypesProtossGroundArmor) != 0 {
			if !f.playerHasCompletedUnit(owner, UnitTypesProtossTemplarArchives) {
				return false
			}
		}
		return true
	case UpgradeTypesProtossAirArmor:
		if !f.unitIs(u.UnitType, UnitTypesProtossCyberneticsCore) {
			return false
		}
		if f.unitIsConstructing(u) || f.unitIsResearching(u) || f.unitIsUpgrading(u) {
			return false
		}
		if f.playerUpgradeLevel(owner, UpgradeTypesProtossAirArmor) != 0 {
			if !f.playerHasCompletedUnit(owner, UnitTypesProtossFleetBeacon) {
				return false
			}
		}
		return true
	case UpgradeTypesTerranInfantryWeapons:
		if !f.unitIs(u.UnitType, UnitTypesTerranEngineeringBay) {
			return false
		}
		if !f.uGroundedBuilding(u) {
			return false
		}
		if f.unitIsConstructing(u) || f.unitIsResearching(u) || f.unitIsUpgrading(u) {
			return false
		}
		if f.playerUpgradeLevel(owner, UpgradeTypesTerranInfantryWeapons) != 0 {
			if !f.playerHasCompletedUnit(owner, UnitTypesTerranScienceFacility) {
				return false
			}
		}
		return true
	case UpgradeTypesTerranVehicleWeapons:
		if !f.unitIs(u.UnitType, UnitTypesTerranArmory) {
			return false
		}
		if !f.uGroundedBuilding(u) {
			return false
		}
		if f.unitIsConstructing(u) || f.unitIsResearching(u) || f.unitIsUpgrading(u) {
			return false
		}
		if f.playerUpgradeLevel(owner, UpgradeTypesTerranVehicleWeapons) != 0 {
			if !f.playerHasCompletedUnit(owner, UnitTypesTerranScienceFacility) {
				return false
			}
		}
		return true
	case UpgradeTypesTerranShipWeapons:
		if !f.unitIs(u.UnitType, UnitTypesTerranArmory) {
			return false
		}
		if !f.uGroundedBuilding(u) {
			return false
		}
		if f.unitIsConstructing(u) || f.unitIsResearching(u) || f.unitIsUpgrading(u) {
			return false
		}
		if f.playerUpgradeLevel(owner, UpgradeTypesTerranShipWeapons) != 0 {
			if !f.playerHasCompletedUnit(owner, UnitTypesTerranScienceFacility) {
				return false
			}
		}
		return true
	case UpgradeTypesZergMeleeAttacks:
		if !f.unitIs(u.UnitType, UnitTypesZergEvolutionChamber) {
			return false
		}
		if f.unitIsConstructing(u) || f.unitIsResearching(u) || f.unitIsUpgrading(u) {
			return false
		}
		switch f.playerUpgradeLevel(owner, UpgradeTypesZergMeleeAttacks) {
		case 1:
			if !f.playerHasCompletedUnit(owner, UnitTypesZergLair) && !f.playerHasUnit(owner, UnitTypesZergHive) {
				return false
			}
		case 2:
			if !f.playerHasCompletedUnit(owner, UnitTypesZergHive) {
				return false
			}
		}
		return true
	case UpgradeTypesZergMissileAttacks:
		if !f.unitIs(u.UnitType, UnitTypesZergEvolutionChamber) {
			return false
		}
		if f.unitIsConstructing(u) || f.unitIsResearching(u) || f.unitIsUpgrading(u) {
			return false
		}
		switch f.playerUpgradeLevel(owner, UpgradeTypesZergMissileAttacks) {
		case 1:
			if !f.playerHasCompletedUnit(owner, UnitTypesZergLair) && !f.playerHasUnit(owner, UnitTypesZergHive) {
				return false
			}
		case 2:
			if !f.playerHasCompletedUnit(owner, UnitTypesZergHive) {
				return false
			}
		}
		return true
	case UpgradeTypesZergFlyerAttacks:
		if !f.unitIs(u.UnitType, UnitTypesZergSpire) && !f.unitIs(u.UnitType, UnitTypesZergGreaterSpire) {
			return false
		}
		if f.unitIsConstructing(u) || f.unitIsResearching(u) || f.unitIsUpgrading(u) {
			return false
		}
		switch f.playerUpgradeLevel(owner, UpgradeTypesZergFlyerAttacks) {
		case 1:
			if !f.playerHasCompletedUnit(owner, UnitTypesZergLair) && !f.playerHasUnit(owner, UnitTypesZergHive) {
				return false
			}
		case 2:
			if !f.playerHasCompletedUnit(owner, UnitTypesZergHive) {
				return false
			}
		}
		return true
	case UpgradeTypesProtossGroundWeapons:
		if !f.unitIs(u.UnitType, UnitTypesProtossForge) {
			return false
		}
		if f.unitIsConstructing(u) || f.unitIsResearching(u) || f.unitIsUpgrading(u) {
			return false
		}
		if f.playerUpgradeLevel(owner, UpgradeTypesProtossGroundWeapons) != 0 {
			if !f.playerHasCompletedUnit(owner, UnitTypesProtossTemplarArchives) {
				return false
			}
		}
		return true
	case UpgradeTypesProtossAirWeapons:
		if !f.unitIs(u.UnitType, UnitTypesProtossCyberneticsCore) {
			return false
		}
		if f.unitIsConstructing(u) || f.unitIsResearching(u) || f.unitIsUpgrading(u) {
			return false
		}
		if f.playerUpgradeLevel(owner, UpgradeTypesProtossAirWeapons) != 0 {
			if !f.playerHasCompletedUnit(owner, UnitTypesProtossFleetBeacon) {
				return false
			}
		}
		return true
	case UpgradeTypesProtossPlasmaShields:
		if !f.unitIs(u.UnitType, UnitTypesProtossForge) {
			return false
		}
		if f.unitIsConstructing(u) || f.unitIsResearching(u) || f.unitIsUpgrading(u) {
			return false
		}
		if f.playerUpgradeLevel(owner, UpgradeTypesProtossPlasmaShields) != 0 {
			if !f.playerHasCompletedUnit(owner, UnitTypesProtossCyberneticsCore) {
				return false
			}
		}
		return true
	case UpgradeTypesU238Shells:
		if !f.unitIs(u.UnitType, UnitTypesTerranAcademy) {
			return false
		}
		if !f.uGroundedBuilding(u) {
			return false
		}
		if f.unitIsConstructing(u) || f.unitIsResearching(u) || f.unitIsUpgrading(u) {
			return false
		}
		return true
	case UpgradeTypesIonThrusters:
		if !f.unitIs(u.UnitType, UnitTypesTerranMachineShop) {
			return false
		}
		if !f.uGroundedBuilding(u) {
			return false
		}
		if f.unitIsConstructing(u) || f.unitIsResearching(u) || f.unitIsUpgrading(u) {
			return false
		}
		return true
	case UpgradeTypesTitanReactor:
		if !f.unitIs(u.UnitType, UnitTypesTerranScienceFacility) {
			return false
		}
		if !f.uGroundedBuilding(u) {
			return false
		}
		if f.unitIsConstructing(u) || f.unitIsResearching(u) || f.unitIsUpgrading(u) {
			return false
		}
		return true
	case UpgradeTypesOcularImplants:
		if !f.unitIs(u.UnitType, UnitTypesTerranCovertOps) {
			return false
		}
		if !f.uGroundedBuilding(u) {
			return false
		}
		if f.unitIsConstructing(u) || f.unitIsResearching(u) || f.unitIsUpgrading(u) {
			return false
		}
		return true
	case UpgradeTypesMoebiusReactor:
		if !f.unitIs(u.UnitType, UnitTypesTerranCovertOps) {
			return false
		}
		if !f.uGroundedBuilding(u) {
			return false
		}
		if f.unitIsConstructing(u) || f.unitIsResearching(u) || f.unitIsUpgrading(u) {
			return false
		}
		return true
	case UpgradeTypesApolloReactor:
		if !f.unitIs(u.UnitType, UnitTypesTerranControlTower) {
			return false
		}
		if !f.uGroundedBuilding(u) {
			return false
		}
		if f.unitIsConstructing(u) || f.unitIsResearching(u) || f.unitIsUpgrading(u) {
			return false
		}
		return true
	case UpgradeTypesColossusReactor:
		if !f.unitIs(u.UnitType, UnitTypesTerranPhysicsLab) {
			return false
		}
		if !f.uGroundedBuilding(u) {
			return false
		}
		if f.unitIsConstructing(u) || f.unitIsResearching(u) || f.unitIsUpgrading(u) {
			return false
		}
		return true
	case UpgradeTypesVentralSacs:
		if !f.unitIs(u.UnitType, UnitTypesZergLair) && !f.unitIs(u.UnitType, UnitTypesZergHive) {
			return false
		}
		if f.unitIsConstructing(u) || f.unitIsResearching(u) || f.unitIsUpgrading(u) {
			return false
		}
		return true
	case UpgradeTypesAntennae:
		if !f.unitIs(u.UnitType, UnitTypesZergLair) && !f.unitIs(u.UnitType, UnitTypesZergHive) {
			return false
		}
		if f.unitIsConstructing(u) || f.unitIsResearching(u) || f.unitIsUpgrading(u) {
			return false
		}
		return true
	case UpgradeTypesPneumatizedCarapace:
		if !f.unitIs(u.UnitType, UnitTypesZergLair) && !f.unitIs(u.UnitType, UnitTypesZergHive) {
			return false
		}
		if f.unitIsConstructing(u) || f.unitIsResearching(u) || f.unitIsUpgrading(u) {
			return false
		}
		return true
	case UpgradeTypesMetabolicBoost:
		if !f.unitIs(u.UnitType, UnitTypesZergSpawningPool) {
			return false
		}
		if f.unitIsConstructing(u) || f.unitIsResearching(u) || f.unitIsUpgrading(u) {
			return false
		}
		return true
	case UpgradeTypesAdrenalGlands:
		if !f.unitIs(u.UnitType, UnitTypesZergSpawningPool) {
			return false
		}
		if f.unitIsConstructing(u) || f.unitIsResearching(u) || f.unitIsUpgrading(u) {
			return false
		}
		if !f.playerHasCompletedUnit(owner, UnitTypesZergHive) {
			return false
		}
		return true
	case UpgradeTypesMuscularAugments:
		if !f.unitIs(u.UnitType, UnitTypesZergHydraliskDen) {
			return false
		}
		if f.unitIsConstructing(u) || f.unitIsResearching(u) || f.unitIsUpgrading(u) {
			return false
		}
		return true
	case UpgradeTypesGroovedSpines:
		if !f.unitIs(u.UnitType, UnitTypesZergHydraliskDen) {
			return false
		}
		if f.unitIsConstructing(u) || f.unitIsResearching(u) || f.unitIsUpgrading(u) {
			return false
		}
		return true
	case UpgradeTypesGameteMeiosis:
		if !f.unitIs(u.UnitType, UnitTypesZergQueensNest) {
			return false
		}
		if f.unitIsConstructing(u) || f.unitIsResearching(u) || f.unitIsUpgrading(u) {
			return false
		}
		return true
	case UpgradeTypesMetasynapticNode:
		if !f.unitIs(u.UnitType, UnitTypesZergDefilerMound) {
			return false
		}
		if f.unitIsConstructing(u) || f.unitIsResearching(u) || f.unitIsUpgrading(u) {
			return false
		}
		return true
	case UpgradeTypesSingularityCharge:
		if !f.unitIs(u.UnitType, UnitTypesProtossCyberneticsCore) {
			return false
		}
		if f.unitIsConstructing(u) || f.unitIsResearching(u) || f.unitIsUpgrading(u) {
			return false
		}
		return true
	case UpgradeTypesLegEnhancements:
		if !f.unitIs(u.UnitType, UnitTypesProtossCitadelOfAdun) {
			return false
		}
		if f.unitIsConstructing(u) || f.unitIsResearching(u) || f.unitIsUpgrading(u) {
			return false
		}
		return true
	case UpgradeTypesScarabDamage:
		if !f.unitIs(u.UnitType, UnitTypesProtossRoboticsSupportBay) {
			return false
		}
		if f.unitIsConstructing(u) || f.unitIsResearching(u) || f.unitIsUpgrading(u) {
			return false
		}
		return true
	case UpgradeTypesReaverCapacity:
		if !f.unitIs(u.UnitType, UnitTypesProtossRoboticsSupportBay) {
			return false
		}
		if f.unitIsConstructing(u) || f.unitIsResearching(u) || f.unitIsUpgrading(u) {
			return false
		}
		return true
	case UpgradeTypesGraviticDrive:
		if !f.unitIs(u.UnitType, UnitTypesProtossRoboticsSupportBay) {
			return false
		}
		if f.unitIsConstructing(u) || f.unitIsResearching(u) || f.unitIsUpgrading(u) {
			return false
		}
		return true
	case UpgradeTypesSensorArray:
		if !f.unitIs(u.UnitType, UnitTypesProtossObservatory) {
			return false
		}
		if f.unitIsConstructing(u) || f.unitIsResearching(u) || f.unitIsUpgrading(u) {
			return false
		}
		return true
	case UpgradeTypesGraviticBoosters:
		if !f.unitIs(u.UnitType, UnitTypesProtossObservatory) {
			return false
		}
		if f.unitIsConstructing(u) || f.unitIsResearching(u) || f.unitIsUpgrading(u) {
			return false
		}
		return true
	case UpgradeTypesKhaydarinAmulet:
		if !f.unitIs(u.UnitType, UnitTypesProtossTemplarArchives) {
			return false
		}
		if f.unitIsConstructing(u) || f.unitIsResearching(u) || f.unitIsUpgrading(u) {
			return false
		}
		return true
	case UpgradeTypesApialSensors:
		if !f.unitIs(u.UnitType, UnitTypesProtossFleetBeacon) {
			return false
		}
		if f.unitIsConstructing(u) || f.unitIsResearching(u) || f.unitIsUpgrading(u) {
			return false
		}
		return true
	case UpgradeTypesGraviticThrusters:
		if !f.unitIs(u.UnitType, UnitTypesProtossFleetBeacon) {
			return false
		}
		if f.unitIsConstructing(u) || f.unitIsResearching(u) || f.unitIsUpgrading(u) {
			return false
		}
		return true
	case UpgradeTypesCarrierCapacity:
		if !f.unitIs(u.UnitType, UnitTypesProtossFleetBeacon) {
			return false
		}
		if f.unitIsConstructing(u) || f.unitIsResearching(u) || f.unitIsUpgrading(u) {
			return false
		}
		return true
	case UpgradeTypesKhaydarinCore:
		if !f.unitIs(u.UnitType, UnitTypesProtossArbiterTribunal) {
			return false
		}
		if f.unitIsConstructing(u) || f.unitIsResearching(u) || f.unitIsUpgrading(u) {
			return false
		}
		return true
	case UpgradeTypesArgusJewel:
		if !f.unitIs(u.UnitType, UnitTypesProtossFleetBeacon) {
			return false
		}
		if f.unitIsConstructing(u) || f.unitIsResearching(u) || f.unitIsUpgrading(u) {
			return false
		}
		return true
	case UpgradeTypesCaduceusReactor:
		if !f.unitIs(u.UnitType, UnitTypesTerranAcademy) {
			return false
		}
		if !f.uGroundedBuilding(u) {
			return false
		}
		if f.unitIsConstructing(u) || f.unitIsResearching(u) || f.unitIsUpgrading(u) {
			return false
		}
		return true
	case UpgradeTypesArgusTalisman:
		if !f.unitIs(u.UnitType, UnitTypesProtossTemplarArchives) {
			return false
		}
		if f.unitIsConstructing(u) || f.unitIsResearching(u) || f.unitIsUpgrading(u) {
			return false
		}
		return true
	case UpgradeTypesChitinousPlating:
		if !f.unitIs(u.UnitType, UnitTypesZergUltraliskCavern) {
			return false
		}
		if f.unitIsConstructing(u) || f.unitIsResearching(u) || f.unitIsUpgrading(u) {
			return false
		}
		return true
	case UpgradeTypesAnabolicSynthesis:
		if !f.unitIs(u.UnitType, UnitTypesZergUltraliskCavern) {
			return false
		}
		if f.unitIsConstructing(u) || f.unitIsResearching(u) || f.unitIsUpgrading(u) {
			return false
		}
		return true
	case UpgradeTypesCharonBoosters:
		if !f.unitIs(u.UnitType, UnitTypesTerranMachineShop) {
			return false
		}
		if f.unitIsConstructing(u) || f.unitIsResearching(u) || f.unitIsUpgrading(u) {
			return false
		}
		if !f.playerHasCompletedUnit(owner, UnitTypesTerranArmory) {
			return false
		}
		return true
	default:
		return false
	}
}

func (f *StateFunctions) unitCanUseTechOwner(u *Unit, tech *TechTypeT, owner int32) bool {
	if u.Owner != owner {
		return false
	}
	return f.unitCanUseTech(u, tech)
}

func (f *StateFunctions) unitCanUseTech(u *Unit, tech *TechTypeT) bool {
	if !f.uCompleted(u) {
		return false
	}
	if f.uHallucination(u) {
		return false
	}
	if f.unitIsDisabled(u) {
		return false
	}
	owner := u.Owner
	if !f.playerTechAvailable(owner, tech.ID) {
		return false
	}
	switch tech.ID {
	case TechTypesStimPacks:
		if !f.utHero(u.UnitType) && !f.playerHasResearched(owner, TechTypesStimPacks) {
			return false
		}
		if !f.unitIsMarine(u.UnitType) && !f.unitIsFirebat(u.UnitType) {
			return false
		}
		return true
	case TechTypesLockdown:
		if !f.utHero(u.UnitType) && !f.playerHasResearched(owner, TechTypesLockdown) {
			return false
		}
		if !f.unitIsGhost(u.UnitType) {
			return false
		}
		return true
	case TechTypesSpiderMines:
		if !f.utHero(u.UnitType) && !f.playerHasResearched(owner, TechTypesSpiderMines) {
			return false
		}
		if f.unitSpiderMineCount(u) == 0 {
			return false
		}
		if !f.unitIsVulture(u.UnitType) {
			return false
		}
		return true
	case TechTypesScannerSweep:
		if !f.unitIs(u.UnitType, UnitTypesTerranComsatStation) {
			return false
		}
		return true
	case TechTypesTankSiegeMode:
		if !f.utHero(u.UnitType) && !f.playerHasResearched(owner, TechTypesTankSiegeMode) {
			return false
		}
		if !f.unitIsTank(u.UnitType) {
			return false
		}
		return true
	case TechTypesEMPShockwave:
		if !f.utHero(u.UnitType) && !f.playerHasResearched(owner, TechTypesEMPShockwave) {
			return false
		}
		if !f.unitIs(u.UnitType, UnitTypesTerranScienceVessel) && !f.unitIs(u.UnitType, UnitTypesHeroMagellan) {
			return false
		}
		return true
	case TechTypesDefensiveMatrix:
		if !f.unitIs(u.UnitType, UnitTypesTerranScienceVessel) && !f.unitIs(u.UnitType, UnitTypesHeroMagellan) {
			return false
		}
		return true
	case TechTypesIrradiate:
		if !f.utHero(u.UnitType) && !f.playerHasResearched(owner, TechTypesIrradiate) {
			return false
		}
		if !f.unitIs(u.UnitType, UnitTypesTerranScienceVessel) && !f.unitIs(u.UnitType, UnitTypesHeroMagellan) {
			return false
		}
		return true
	case TechTypesYamatoGun:
		if !f.utHero(u.UnitType) && !f.playerHasResearched(owner, TechTypesYamatoGun) {
			return false
		}
		if !f.unitIs(u.UnitType, UnitTypesTerranBattlecruiser) && !f.unitIs(u.UnitType, UnitTypesHeroHyperion) && !f.unitIs(u.UnitType, UnitTypesHeroNoradII) && !f.unitIs(u.UnitType, UnitTypesHeroGerardDuGalle) {
			return false
		}
		return true
	case TechTypesCloakingField:
		if !f.utHero(u.UnitType) && !f.playerHasResearched(owner, TechTypesCloakingField) {
			return false
		}
		if !f.unitIsWraith(u.UnitType) {
			return false
		}
		return true
	case TechTypesPersonnelCloaking:
		if !f.utHero(u.UnitType) && !f.playerHasResearched(owner, TechTypesPersonnelCloaking) {
			return false
		}
		if !f.unitIsGhost(u.UnitType) {
			return false
		}
		return true
	case TechTypesRestoration:
		if !f.playerHasResearched(owner, TechTypesRestoration) {
			return false
		}
		if !f.unitIs(u.UnitType, UnitTypesTerranMedic) {
			return false
		}
		return true
	case TechTypesHealing:
		if !f.unitIs(u.UnitType, UnitTypesTerranMedic) {
			return false
		}
		return true
	case TechTypesBurrowing:
		if !f.utHero(u.UnitType) && !f.unitIs(u.UnitType, UnitTypesZergLurker) && !f.playerHasResearched(owner, TechTypesBurrowing) {
			return false
		}
		if !f.unitIs(u.UnitType, UnitTypesZergLurker) && !f.unitIs(u.UnitType, UnitTypesZergDrone) && !f.unitIs(u.UnitType, UnitTypesZergInfestedTerran) && !f.unitIs(u.UnitType, UnitTypesZergDefiler) && !f.unitIs(u.UnitType, UnitTypesZergZergling) && !f.unitIs(u.UnitType, UnitTypesZergHydralisk) && !f.unitIs(u.UnitType, UnitTypesHeroUncleanOne) && !f.unitIs(u.UnitType, UnitTypesHeroDevouringOne) && !f.unitIs(u.UnitType, UnitTypesHeroHunterKiller) {
			return false
		}
		return true
	case TechTypesInfestation:
		if !f.unitIsQueen(u.UnitType) {
			return false
		}
		return true
	case TechTypesSpawnBroodlings:
		if !f.utHero(u.UnitType) && !f.playerHasResearched(owner, TechTypesSpawnBroodlings) {
			return false
		}
		if !f.unitIsQueen(u.UnitType) {
			return false
		}
		return true
	case TechTypesParasite:
		if !f.unitIsQueen(u.UnitType) {
			return false
		}
		return true
	case TechTypesDarkSwarm:
		if !f.unitIsDefiler(u.UnitType) {
			return false
		}
		if f.uBurrowed(u) {
			return false
		}
		return true
	case TechTypesPlague:
		if !f.utHero(u.UnitType) && !f.playerHasResearched(owner, TechTypesPlague) {
			return false
		}
		if !f.unitIsDefiler(u.UnitType) {
			return false
		}
		if f.uBurrowed(u) {
			return false
		}
		return true
	case TechTypesConsume:
		if !f.utHero(u.UnitType) && !f.playerHasResearched(owner, TechTypesConsume) {
			return false
		}
		if !f.unitIsDefiler(u.UnitType) && !f.unitIs(u.UnitType, UnitTypesHeroInfestedDuran) && !f.unitIs(u.UnitType, UnitTypesHeroInfestedKerrigan) {
			return false
		}
		if f.uBurrowed(u) {
			return false
		}
		return true
	case TechTypesEnsnare:
		if !f.utHero(u.UnitType) && !f.playerHasResearched(owner, TechTypesEnsnare) {
			return false
		}
		if !f.unitIsQueen(u.UnitType) && !f.unitIs(u.UnitType, UnitTypesHeroInfestedKerrigan) {
			return false
		}
		if f.uBurrowed(u) {
			return false
		}
		return true
	case TechTypesPsionicStorm:
		if !f.utHero(u.UnitType) && !f.playerHasResearched(owner, TechTypesPsionicStorm) {
			return false
		}
		if !f.unitIs(u.UnitType, UnitTypesProtossHighTemplar) && !f.unitIs(u.UnitType, UnitTypesHeroTassadar) && !f.unitIs(u.UnitType, UnitTypesHeroInfestedKerrigan) {
			return false
		}
		return true
	case TechTypesHallucination:
		if !f.utHero(u.UnitType) && !f.playerHasResearched(owner, TechTypesHallucination) {
			return false
		}
		if !f.unitIs(u.UnitType, UnitTypesProtossHighTemplar) && !f.unitIs(u.UnitType, UnitTypesHeroTassadar) {
			return false
		}
		return true
	case TechTypesArchonWarp:
		if !f.unitIs(u.UnitType, UnitTypesProtossHighTemplar) {
			return false
		}
		return true
	case TechTypesRecall:
		if !f.utHero(u.UnitType) && !f.playerHasResearched(owner, TechTypesRecall) {
			return false
		}
		if !f.unitIsArbiter(u.UnitType) {
			return false
		}
		return true
	case TechTypesStasisField:
		if !f.utHero(u.UnitType) && !f.playerHasResearched(owner, TechTypesStasisField) {
			return false
		}
		if !f.unitIsArbiter(u.UnitType) {
			return false
		}
		return true
	case TechTypesDisruptionWeb:
		if !f.utHero(u.UnitType) && !f.playerHasResearched(owner, TechTypesDisruptionWeb) {
			return false
		}
		if !f.unitIs(u.UnitType, UnitTypesProtossCorsair) && !f.unitIs(u.UnitType, UnitTypesHeroRaszagal) {
			return false
		}
		return true
	case TechTypesDarkArchonMeld:
		if !f.unitIs(u.UnitType, UnitTypesProtossDarkTemplar) {
			return false
		}
		return true
	case TechTypesMindControl:
		if !f.playerHasResearched(owner, TechTypesMindControl) {
			return false
		}
		if !f.unitIs(u.UnitType, UnitTypesProtossDarkArchon) {
			return false
		}
		return true
	case TechTypesFeedback:
		if !f.unitIs(u.UnitType, UnitTypesProtossDarkArchon) {
			return false
		}
		return true
	case TechTypesOpticalFlare:
		if !f.unitIs(u.UnitType, UnitTypesTerranMedic) {
			return false
		}
		if !f.playerHasResearched(owner, TechTypesOpticalFlare) {
			return false
		}
		return true
	case TechTypesMaelstrom:
		if !f.playerHasResearched(owner, TechTypesMaelstrom) {
			return false
		}
		if !f.unitIs(u.UnitType, UnitTypesProtossDarkArchon) {
			return false
		}
		return true
	case TechTypesLurkerAspect:
		if !f.playerHasResearched(owner, TechTypesLurkerAspect) {
			return false
		}
		if !f.unitIs(u.UnitType, UnitTypesZergHydralisk) {
			return false
		}
		return true
	default:
		return false
	}
}
