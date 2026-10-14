package bwgame

func (f *GameLoadFunctions) reset() {
	panic(notPorted("reset"))
}

func (f *GameLoadFunctions) getNewRegion() *Region {
	panic(notPorted("get_new_region"))
}

func (f *GameLoadFunctions) regionsCreate() {
	panic(notPorted("regions_create"))
}

func (f *GameLoadFunctions) getUnitTypeStrength(unitType *UnitTypeT, weaponType *WeaponTypeT) int32 {
	panic(notPorted("get_unit_strength"))
}

func (f *GameLoadFunctions) calculateUnitStrengths() {
	panic(notPorted("calculate_unit_strengths"))
}

func (f *GameLoadFunctions) generateSightValues() {
	panic(notPorted("generate_sight_values"))
}

func (f *GameLoadFunctions) loadTileStuff() {
	panic(notPorted("load_tile_stuff"))
}

func (f *GameLoadFunctions) createStartingUnit(unitType *UnitTypeT, pos Xy, owner int32) *Unit {
	panic(notPorted("create_starting_unit"))
}

func (f *GameLoadFunctions) createStartingUnits(owner int32, position Xy, race Race) {
	panic(notPorted("create_starting_units"))
}

func (f *GameLoadFunctions) getMapString(index uint) string {
	panic(notPorted("get_map_string"))
}

func (f *GameLoadFunctions) loadMap(loadDataFile LoadDataFileFunc, setupF func(), initialProcessing bool) {
	panic(notPorted("load_map"))
}

func (f *GameLoadFunctions) loadMapData(data []byte, setupF func(), initialProcessing bool) {
	panic(notPorted("load_map_data"))
}
