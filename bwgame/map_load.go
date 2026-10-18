package bwgame

func fixupTypeID[T any, I ~int](t *TypeID[T, I], none I, get func(I) *T) {
	if t.ID == none {
		t.Ptr = nil
		return
	}
	t.Ptr = get(t.ID)
}

func fixupRequiredTypeID[T any, I ~int](t *TypeID[T, I], get func(I) *T) {
	t.Ptr = get(t.ID)
}

func clampMapPosCoord(v *int32) {
	if *v < 0 {
		*v = 0
	}
	if *v >= 256*32 {
		*v = 256*32 - 1
	}
}

func (f *GameLoadFunctions) reset() {
	st := f.st
	g := f.gameSt

	g.UnitTypes = mustLoad(LoadUnitsDat(f.globalSt.UnitsDat))
	g.WeaponTypes = mustLoad(LoadWeaponsDat(f.globalSt.WeaponsDat))
	g.UpgradeTypes = mustLoad(LoadUpgradesDat(f.globalSt.UpgradesDat))
	g.TechTypes = mustLoad(LoadTechdataDat(f.globalSt.TechdataDat))

	f.fixupUnitTypes()
	f.fixupWeaponTypes()

	for i := range g.UnitTypeAllowed {
		for j := range g.UnitTypeAllowed[i] {
			g.UnitTypeAllowed[i][j] = true
		}
	}
	for i := range g.TechAvailable {
		for j := range g.TechAvailable[i] {
			g.TechAvailable[i][j] = true
		}
	}
	for i := range g.MaxUpgradeLevels {
		for upgrade := UpgradeTypes(0); upgrade != 61; upgrade++ {
			g.MaxUpgradeLevels[i][upgrade] = f.getUpgradeType(upgrade).MaxLevel
		}
	}

	st.Alliances = [12][12]int32{}
	for i := 0; i != 12; i++ {
		st.Alliances[i] = [12]int32{}
		st.Alliances[i][i] = 1
	}
	for i := 0; i != 12; i++ {
		st.Alliances[i][11] = 1
		st.Alliances[11][i] = 1
	}

	st.UpgradeLevels = [12][UpgradeTypesNone]int32{}
	st.UpgradeUpgrading = [12][UpgradeTypesNone]bool{}
	st.TechResearched = [12][TechTypesNone]bool{}
	st.TechResearching = [12][TechTypesNone]bool{}

	st.UnitCounters = UnitCounters{}

	st.TotalBuildingsEverCompleted = [12]int32{}
	st.TotalNonBuildingsEverCompleted = [12]int32{}

	st.UnitScore = [12]int32{}
	st.BuildingScore = [12]int32{}

	st.SupplyUsed = [12][3]Fp1{}
	st.SupplyAvailable = [12][3]Fp1{}

	f.setAcquisitionRanges()

	f.calculateUnitStrengths()

	f.generateSightValues()

	f.loadTileStuff()

	st.Tiles = make([]Tile, g.MapTileWidth*g.MapTileHeight)
	for i := range st.Tiles {
		st.Tiles[i].Visible = 0xff
		st.Tiles[i].Explored = 0xff
	}
	st.TilesMegaTileIndex = make([]uint16, len(st.Tiles))

	st.UpdateTilesCountdown = 1

	st.OrderTimerCounter = 10
	st.SecondaryOrderTimerCounter = 150
	st.CurrentFrame = 0

	f.resetObjectContainers()

	cursor := f.createThingy(f.getSpriteType(SpriteTypesSPRITEIDCursorMarker), Xy{}, 0)
	if cursor != nil {
		cursor.Sprite.Flags |= SpriteFlagHidden
	}

	st.LastError = 0
	st.RecentLurkerHitCurrentIndex = 0

	st.TriggerTimer = 1
	st.RunningTriggers = [8][]RunningTrigger{}
	st.TriggerWaitTimers = [12]int32{}
	st.TriggerWaiting = [12]bool{}

	f.setMaxUnitDimensions()

	st.RandomCounts = [0x100]int32{}
	st.TotalRandomCounts = 0
	st.LcgRandState = 42

	g.RepulseFieldWidth = (g.MapWidth + 47) / 48
	g.RepulseFieldHeight = (g.MapHeight + 47) / 48
	st.RepulseField = resizeSlice(st.RepulseField, int(g.RepulseFieldWidth*g.RepulseFieldHeight))

	g.Triggers = nil

	st.PrevBulletSourceUnit = nil
	st.ConsiderCollisionWithUnitBug = nil

	st.CurrentMinerals = [12]int32{}
	st.CurrentGas = [12]int32{}
	st.TotalMineralsGathered = [12]int32{}
	st.TotalGasGathered = [12]int32{}

	st.CreepLife = NewCreepLife()
	st.UpdatePsionicMatrix = false
	st.DisruptionWebbedUnits = 0
	st.CheatsEnabled = false
	st.CheatOperationCwal = false

	st.Locations = nil
}

func (f *GameLoadFunctions) fixupUnitTypes() {
	for i := range f.gameSt.UnitTypes.Vec {
		v := &f.gameSt.UnitTypes.Vec[i]
		fixupRequiredTypeID(&v.Flingy, f.getFlingyType)
		fixupTypeID(&v.TurretUnitType, UnitTypesNone, f.getUnitType)
		fixupTypeID(&v.Subunit2, UnitTypesNone, f.getUnitType)
		fixupTypeID(&v.InfestationUnit, UnitTypesNone, f.getUnitType)
		fixupTypeID(&v.ConstructionAnimation, ImageTypesNone, f.getImageType)
		fixupTypeID(&v.GroundWeapon, WeaponTypesNone, f.getWeaponType)
		fixupTypeID(&v.AirWeapon, WeaponTypesNone, f.getWeaponType)
		fixupTypeID(&v.ArmorUpgrade, UpgradeTypesNone, f.getUpgradeType)
		fixupRequiredTypeID(&v.ComputerAiIdle, f.getOrderType)
		fixupRequiredTypeID(&v.HumanAiIdle, f.getOrderType)
		fixupRequiredTypeID(&v.ReturnToIdle, f.getOrderType)
		fixupRequiredTypeID(&v.AttackUnit, f.getOrderType)
		fixupRequiredTypeID(&v.AttackMove, f.getOrderType)

		clampMapPosCoord(&v.PlacementSize.X)
		clampMapPosCoord(&v.PlacementSize.Y)
		clampMapPosCoord(&v.Dimensions.From.X)
		clampMapPosCoord(&v.Dimensions.From.Y)
		clampMapPosCoord(&v.Dimensions.To.X)
		clampMapPosCoord(&v.Dimensions.To.Y)
	}
}

func (f *GameLoadFunctions) fixupWeaponTypes() {
	for i := range f.gameSt.WeaponTypes.Vec {
		v := &f.gameSt.WeaponTypes.Vec[i]
		fixupRequiredTypeID(&v.Flingy, f.getFlingyType)
		fixupTypeID(&v.DamageUpgrade, UpgradeTypesNone, f.getUpgradeType)
	}
}

func (f *GameLoadFunctions) setAcquisitionRanges() {
	for i := UnitTypes(0); i != 228; i++ {
		unitType := f.getUnitType(i)
		attackingType := unitType
		if unitType.TurretUnitType.Ptr != nil {
			attackingType = unitType.TurretUnitType.Ptr
		}
		groundWeapon := attackingType.GroundWeapon.Ptr
		airWeapon := attackingType.AirWeapon.Ptr
		acqRange := attackingType.TargetAcquisitionRange
		if groundWeapon != nil {
			acqRange = max(acqRange, groundWeapon.MaxRange/32)
		}
		if airWeapon != nil {
			acqRange = max(acqRange, airWeapon.MaxRange/32)
		}
		unitType.TargetAcquisitionRange = acqRange
	}
}

func (f *GameLoadFunctions) resetObjectContainers() {
	st := f.st

	st.VisibleUnits.Clear()
	st.HiddenUnits.Clear()
	st.MapRevealerUnits.Clear()
	st.DeadUnits.Clear()
	for i := range st.PlayerUnits {
		st.PlayerUnits[i].Clear()
	}

	st.newObjectContainers()

	st.ActiveBulletsSize = 0
	st.ActiveBullets.Clear()

	st.SpritesOnTileLine = make([]SpriteList, f.gameSt.MapTileHeight)

	st.ActiveOrdersSize = 0

	st.ActiveThingiesSize = 0
	st.ActiveThingies.Clear()
	st.FreeThingies.Clear()
	st.Thingies = nil
}

func (f *GameLoadFunctions) setMaxUnitDimensions() {
	maxUnitWidth := int32(0)
	maxUnitHeight := int32(0)
	for i := range f.gameSt.UnitTypes.Vec {
		v := &f.gameSt.UnitTypes.Vec[i]
		width := v.Dimensions.From.X + 1 + v.Dimensions.To.X
		height := v.Dimensions.From.Y + 1 + v.Dimensions.To.Y
		if width > maxUnitWidth {
			maxUnitWidth = width
		}
		if height > maxUnitHeight {
			maxUnitHeight = height
		}
	}
	f.gameSt.MaxUnitWidth = maxUnitWidth
	f.gameSt.MaxUnitHeight = maxUnitHeight
}

func (f *GameLoadFunctions) getUnitTypeStrength(unitType *UnitTypeT, weaponType *WeaponTypeT) int32 {
	switch unitType.ID {
	case UnitTypesTerranVultureSpiderMine, UnitTypesProtossInterceptor, UnitTypesProtossScarab:
		return 0
	}
	hp := unitType.Hitpoints.IntegerPart()
	if unitType.HasShield {
		hp += unitType.ShieldPoints
	}
	if hp == 0 {
		return 0
	}
	bulletCount := weaponType.BulletCount
	cd := weaponType.Cooldown
	dmg := weaponType.DamageAmount
	rangeValue := weaponType.MaxRange
	a := (uint32(rangeValue) / uint32(cd)) * uint32(bulletCount) * uint32(dmg)
	b := uint32((int64(hp) * (int64((bulletCount*dmg)<<11) / int64(cd))) >> 8)
	score := int32(sqrtXTimes758(a + b))
	switch unitType.ID {
	case UnitTypesTerranSCV, UnitTypesZergDrone, UnitTypesProtossProbe:
		return score / 4
	case UnitTypesTerranFirebat, UnitTypesZergMutalisk, UnitTypesProtossZealot:
		return score * 2
	case UnitTypesZergScourge, UnitTypesZergInfestedTerran:
		return score / 16
	case UnitTypesProtossReaver:
		return score / 10
	default:
		return score
	}
}

func sqrtXTimes758(x uint32) uint32 {
	if x == 0 {
		return 0
	}
	value := isqrt(x) * 758 / 100
	n := uint32(8)
	for n > 0 {
		nv := value + n/2 + 1
		r := uint32((uint64(nv) * uint64(nv) * 10000) / (758 * 758))
		if r < x {
			value = nv
			n -= n/2 + 1
		} else {
			n /= 2
		}
	}
	return value
}

func (f *GameLoadFunctions) calculateUnitStrengths() {
	for idx := UnitTypes(0); idx < 228; idx++ {
		unitType := f.getUnitType(idx)
		attackingType := unitType
		airStrength := int32(0)
		groundStrength := int32(0)
		if attackingType.ID != UnitTypesZergLarva && !f.unitIsEgg(attackingType) {
			if f.unitIsCarrier(attackingType) {
				attackingType = f.getUnitType(UnitTypesProtossInterceptor)
			} else if f.unitIsReaver(attackingType) {
				attackingType = f.getUnitType(UnitTypesProtossScarab)
			} else if attackingType.TurretUnitType.Ptr != nil {
				attackingType = attackingType.TurretUnitType.Ptr
			}

			airWeapon := attackingType.AirWeapon.Ptr
			if airWeapon == nil {
				airStrength = 1
			} else {
				airStrength = f.getUnitTypeStrength(unitType, airWeapon)
			}

			groundWeapon := attackingType.GroundWeapon.Ptr
			if groundWeapon == nil {
				groundStrength = 1
			} else {
				groundStrength = f.getUnitTypeStrength(unitType, groundWeapon)
			}
		}
		if airStrength == 1 && groundStrength > airStrength {
			airStrength = 0
		}
		if groundStrength == 1 && airStrength > groundStrength {
			groundStrength = 0
		}

		f.gameSt.UnitAirStrength[idx] = airStrength
		f.gameSt.UnitGroundStrength[idx] = groundStrength
	}
}

const noMaskdatNode = ^uint(0)

type sightBaseMask struct {
	maskdatNode uint
	masked      bool
}

type sightValuesGenerator struct {
	v              *SightValues
	baseMask       []sightBaseMask
	centerIndex    uint
	mapTileWidth   int32
	nextEntryIndex uint
	addedCount     int32
}

var sightDirectionX = [4]int32{1, 0, -1, 0}
var sightDirectionY = [4]int32{0, 1, 0, -1}

func (f *GameLoadFunctions) generateSightValues() {
	g := f.gameSt
	for i := range g.SightValues {
		v := &g.SightValues[i]
		v.MaxWidth = 3 + int32(i)*2
		v.MaxHeight = 3 + int32(i)*2
		v.MinWidth = 3
		v.MinHeight = 3
		v.MinMaskSize = 0
		v.ExtMaskedCount = 0
	}

	for i := range g.SightValues {
		s := &sightValuesGenerator{v: &g.SightValues[i], mapTileWidth: int32(g.MapTileWidth)}
		s.generate()
	}
}

func (s *sightValuesGenerator) mask(index uint) {
	if index >= uint(len(s.baseMask)) {
		throwError("attempt to mask invalid base mask index %d (size %d)", index, len(s.baseMask))
	}
	s.baseMask[index].masked = true
}

func (s *sightValuesGenerator) generate() {
	v := s.v
	s.baseMask = make([]sightBaseMask, v.MaxWidth*v.MaxHeight)
	for i := range s.baseMask {
		s.baseMask[i].maskdatNode = noMaskdatNode
	}
	v.MinMaskSize = v.MinWidth * v.MinHeight
	offx := v.MaxWidth/2 - v.MinWidth/2
	offy := v.MaxHeight/2 - v.MinHeight/2
	for y := int32(0); y < v.MinHeight; y++ {
		for x := int32(0); x < v.MinWidth; x++ {
			s.mask(uint((offy+y)*v.MaxWidth + offx + x))
		}
	}
	s.generateBaseMask()
	maskedCount := int32(0)
	for _, m := range s.baseMask {
		if m.masked {
			maskedCount++
		}
	}

	v.ExtMaskedCount = maskedCount - v.MinMaskSize
	v.Maskdat = make([]SightValuesMaskdatNode, maskedCount)

	s.centerIndex = uint(v.MaxHeight/2*v.MaxWidth + v.MaxWidth/2)
	s.baseMask[s.centerIndex].maskdatNode = 0

	s.nextEntryIndex = 1

	curX := int32(-1)
	curY := int32(-1)
	s.addedCount = 1
	for i := int32(2); s.addedCount < maskedCount; i += 2 {
		for dir := 0; dir < 4; dir++ {
			maxI := [4]int32{v.MaxHeight, v.MaxWidth, v.MaxHeight, v.MaxWidth}
			var thisX, thisY int32
			if i > maxI[dir] {
				thisX = curX + i*sightDirectionX[dir]
				thisY = curY + i*sightDirectionY[dir]
				thisX, thisY = s.doN(dir, thisX, thisY, 1)
			} else {
				thisX = curX + sightDirectionX[dir]
				thisY = curY + sightDirectionY[dir]
				thisX, thisY = s.doN(dir, thisX, thisY, min(maxI[(dir+1)%4]-1, i))
			}
			curX = thisX - sightDirectionX[dir]
			curY = thisY - sightDirectionY[dir]
		}
		if i < v.MaxWidth-1 {
			curX--
		}
		if i < v.MaxHeight-1 {
			curY--
		}
	}
}

func (s *sightValuesGenerator) generateBaseMask() {
	v := s.v
	offset := v.MaxHeight/2 - v.MaxWidth/2
	halfWidth := v.MaxWidth / 2
	maxX2 := halfWidth
	maxX1 := halfWidth * 2
	curX1 := int32(0)
	curX2 := halfWidth
	i := int32(0)
	maxI := halfWidth
	cursize1 := int32(0)
	cursize2 := halfWidth * halfWidth
	minCursize2 := halfWidth * (halfWidth - 1)
	minCursize2Chg := halfWidth * 2
	for {
		if curX1 <= maxX1 {
			for k := int32(0); k <= maxX1-curX1; k++ {
				s.mask(uint((offset+curX2)*v.MaxWidth + curX1 + k))
				s.mask(uint((offset+maxX2)*v.MaxWidth + curX1 + k))
			}
		}
		if curX2 <= maxX2 {
			for k := int32(0); k <= maxX2-curX2; k++ {
				s.mask(uint((offset+curX1)*v.MaxWidth + curX2 + k))
				s.mask(uint((offset+maxX1)*v.MaxWidth + curX2 + k))
			}
		}
		cursize2 += 1 - cursize1 - 2
		cursize1 += 2
		curX2--
		maxX2++
		if cursize2 <= minCursize2 {
			maxI--
			curX1++
			maxX1--
			minCursize2 -= minCursize2Chg - 2
			minCursize2Chg -= 2
		}

		i++
		if i > maxI {
			break
		}
	}
}

func (s *sightValuesGenerator) at(relativeIndex int32) *sightBaseMask {
	index := s.centerIndex + uint(relativeIndex)
	if index >= uint(len(s.baseMask)) {
		throwError("attempt to access invalid base mask center-relative index %d (size %d)", index, len(s.baseMask))
	}
	return &s.baseMask[index]
}

func (s *sightValuesGenerator) doN(dir int, thisX, thisY int32, n int32) (int32, int32) {
	v := s.v
	for k := int32(0); k < n; k++ {
		if s.at(thisY*v.MaxWidth+thisX).masked && (thisX != 0 || thisY != 0) {
			s.addMaskdatEntry(dir, thisX, thisY)
		}
		thisX += sightDirectionX[dir]
		thisY += sightDirectionY[dir]
	}
	return thisX, thisY
}

func (s *sightValuesGenerator) addMaskdatEntry(dir int, thisX, thisY int32) {
	v := s.v
	entryIndex := s.nextEntryIndex
	s.nextEntryIndex++
	thisEntry := &v.Maskdat[entryIndex]

	prevX := thisX
	prevY := thisY
	if prevX > 0 {
		prevX--
	} else if prevX < 0 {
		prevX++
	}
	if prevY > 0 {
		prevY--
	} else if prevY < 0 {
		prevY++
	}
	if absInt32(prevX) == absInt32(prevY) || (thisX == 0 && sightDirectionX[dir] != 0) || (thisY == 0 && sightDirectionY[dir] != 0) {
		thisEntry.Prev = s.at(prevY*v.MaxWidth + prevX).maskdatNode
		thisEntry.Prev2 = noMaskdatNode
	} else {
		thisEntry.Prev = s.at(prevY*v.MaxWidth + prevX).maskdatNode
		prev2X := prevX
		prev2Y := prevY
		if absInt32(prev2X) <= absInt32(prev2Y) {
			if thisX >= 0 {
				prev2X++
			} else {
				prev2X--
			}
		} else {
			if thisY >= 0 {
				prev2Y++
			} else {
				prev2Y--
			}
		}
		thisEntry.Prev2 = s.at(prev2Y*v.MaxWidth + prev2X).maskdatNode
	}
	thisEntry.RelativeTileIndex = thisY*s.mapTileWidth + thisX
	thisEntry.X = thisX
	thisEntry.Y = thisY
	s.at(thisY*v.MaxWidth + thisX).maskdatNode = entryIndex
	s.addedCount++
}

func (f *GameLoadFunctions) loadTileStuff() {
	g := f.gameSt

	r := NewDataReaderLe(f.globalSt.TilesetVf4[g.TilesetIndex])
	for r.Left() != 0 {
		var e Vf4Entry
		for i := range e.Flags {
			e.Flags[i] = r.GetU16()
		}
		g.Vf4 = append(g.Vf4, e)
	}
	r = NewDataReaderLe(f.globalSt.TilesetCv5[g.TilesetIndex])
	for r.Left() != 0 {
		var e Cv5Entry
		r.GetU16()
		e.Flags = r.GetU16()
		for i := 0; i != 8; i++ {
			r.GetU16()
		}
		for i := range e.MegaTileIndex {
			e.MegaTileIndex[i] = r.GetU16()
		}
		g.Cv5 = append(g.Cv5, e)
	}

	f.setMegaTileFlags()
}

func (f *GameLoadFunctions) setMegaTileFlags() {
	g := f.gameSt
	g.MegaTileFlags = resizeSlice(g.MegaTileFlags, len(g.Vf4))
	for i := range g.MegaTileFlags {
		g.MegaTileFlags[i] = megaTileFlags(&g.Vf4[i])
	}
}

func megaTileFlags(mt *Vf4Entry) uint16 {
	flags := 0
	walkableCount := 0
	middleCount := 0
	highCount := 0
	veryHighCount := 0
	for _, v := range mt.Flags {
		if v&Vf4EntryFlagWalkable != 0 {
			walkableCount++
		}
		if v&Vf4EntryFlagMiddle != 0 {
			middleCount++
		}
		if v&Vf4EntryFlagHigh != 0 {
			highCount++
		}
		if v&Vf4EntryFlagVeryHigh != 0 {
			veryHighCount++
		}
	}
	if walkableCount > 12 {
		flags |= TileFlagWalkable
	} else {
		flags |= TileFlagUnwalkable
	}
	if walkableCount != 0 && walkableCount != 0x10 {
		flags |= TileFlagPartiallyWalkable
	}
	if highCount < 12 && middleCount+highCount >= 12 {
		flags |= TileFlagMiddle
	}
	if highCount >= 12 {
		flags |= TileFlagHigh
	}
	if veryHighCount != 0 {
		flags |= TileFlagVeryHigh
	}
	return uint16(flags)
}

func (f *GameLoadFunctions) createStartingUnit(unitType *UnitTypeT, pos Xy, owner int32) *Unit {
	u := f.createUnit(unitType, pos, owner)
	if u == nil {
		return nil
	}
	f.onUnitCreated(u, SourceStart, nil)
	f.finishBuildingUnit(u)
	if f.placeCompletedUnit(u) {
		f.completeUnit(u)
	}
	return u
}

func alignXyDownToTileUnsigned(v Xy) Xy {
	return Xy{X: int32(uint32(v.X) / 32 * 32), Y: int32(uint32(v.Y) / 32 * 32)}
}

func (f *GameLoadFunctions) raceUnitType(race Race, zerg, terran, protoss UnitTypes) *UnitTypeT {
	switch race {
	case RaceZerg:
		return f.getUnitType(zerg)
	case RaceTerran:
		return f.getUnitType(terran)
	default:
		return f.getUnitType(protoss)
	}
}

func (f *GameLoadFunctions) createStartingUnits(owner int32, position Xy, race Race) {
	resourceDepotType := f.raceUnitType(race, UnitTypesZergHatchery, UnitTypesTerranCommandCenter, UnitTypesProtossNexus)

	halfPlacementSize := XyDivScalar(resourceDepotType.PlacementSize, 2)
	resourceDepotPos := XyAdd(alignXyDownToTileUnsigned(XySub(position, halfPlacementSize)), halfPlacementSize)
	var placementRect Rect
	placementRect.From = XySub(resourceDepotPos, halfPlacementSize)
	placementRect.To = XySub(XyAdd(placementRect.From, resourceDepotType.PlacementSize), Xy{X: 1, Y: 1})
	for u := range f.findUnits(placementRect) {
		u.UserActionFlags |= 4
		f.killUnitWith(u, noKiller(ReasonRemoved))
	}
	resourceDepot := f.createStartingUnit(resourceDepotType, resourceDepotPos, owner)
	if resourceDepot != nil {
		if f.unitTypeSpreadsCreep(resourceDepotType, true) || f.utRequiresCreep(resourceDepot.UnitType) {
			f.spreadCreepCompletely(resourceDepot.UnitType, resourceDepot.Sprite.Position)
			f.spawnLarva(resourceDepot)
			f.spawnLarva(resourceDepot)
		}
	}
	if race == RaceZerg {
		f.createStartingOverlord(owner, position, resourceDepotType)
	}

	workerType := f.raceUnitType(race, UnitTypesZergDrone, UnitTypesTerranSCV, UnitTypesProtossProbe)

	for i := 0; i != 4; i++ {
		f.createStartingUnit(workerType, position, owner)
	}
}

func (f *GameLoadFunctions) createStartingOverlord(owner int32, position Xy, resourceDepotType *UnitTypeT) {
	overlordPos := position
	add := XyAdd(XyAdd(resourceDepotType.Dimensions.From, resourceDepotType.Dimensions.To), Xy{X: 1, Y: 1})
	if overlordPos.X >= int32(f.gameSt.MapWidth/2) {
		overlordPos.X -= add.X
	} else {
		overlordPos.X += add.X
	}
	if overlordPos.Y >= int32(f.gameSt.MapHeight/2) {
		overlordPos.Y -= add.Y
	} else {
		overlordPos.Y += add.Y
	}
	f.createStartingUnit(f.getUnitType(UnitTypesZergOverlord), overlordPos, owner)
}

func (f *GameLoadFunctions) getMapString(index uint) string {
	if index == 0 {
		return "<null string>"
	}
	index--
	if index >= uint(len(f.gameSt.MapStrings)) {
		return "<invalid string index>"
	}
	return f.gameSt.MapStrings[index]
}

func (f *GameLoadFunctions) loadMap(loadDataFile LoadDataFileFunc, setupF func(), initialProcessing bool) {
	data, err := loadDataFile("staredit/scenario.chk")
	if err != nil {
		throwError("%v", err)
	}
	f.loadMapData(data, setupF, initialProcessing)
}

type tagT [4]byte

func newTag(s string) tagT {
	return tagT{s[0], s[1], s[2], s[3]}
}

func (t tagT) String() string {
	return string(t[:])
}

type mapChunkRequirement struct {
	tag      tagT
	required bool
}

type mapDataLoader struct {
	f              *GameLoadFunctions
	data           []byte
	version        int32
	useMapSettings bool
	tagFuncs       map[tagT]func(r DataReader)
}

func (f *GameLoadFunctions) loadMapData(data []byte, setupF func(), initialProcessing bool) {
	st := f.st
	l := &mapDataLoader{f: f, data: data}
	l.registerTagFuncs()

	st.Players = [12]Player{}

	l.readChunks(chunkRequirements(true, "VER ", "DIM ", "ERA ", "OWNR", "SIDE", "STR ", "SPRP", "FORC", "VCOD"))

	f.reset()

	l.initPlayers()

	if setupF != nil {
		setupF()
	} else {
		l.occupyOpenAndComputerSlots()
	}

	l.useMapSettings = f.setupInfo.VictoryCondition == 0 && f.setupInfo.TournamentMode == 0 && f.setupInfo.StartingUnits == 0

	l.readVersionChunks()

	if !l.useMapSettings && f.setupInfo.VictoryCondition == 1 {
		l.tagFuncs[newTag("TRIG")](NewDataReaderLe(f.globalSt.MeleeTrg))
	}

	for i := range st.Players {
		st.Players[i].InitiallyActive = false
	}
	for p := range f.activePlayers() {
		st.Players[p].InitiallyActive = true
	}

	l.createAllStartingUnits()
	if f.setupInfo.ResourceType == 1 {
		for i := 0; i != 12; i++ {
			st.CurrentMinerals[i] = f.setupInfo.StartingMinerals
			st.TotalMineralsGathered[i] = f.setupInfo.StartingMinerals
		}
	}
	l.shareForceVision()
	f.gameSt.AlliedVictoryWithinForces = f.gameSt.Rules.TopVsBottomAlliances && f.setupInfo.TopVsBottom
	if f.gameSt.AlliedVictoryWithinForces {
		l.allyForces()
	}

	if initialProcessing {
		f.processFrame()
		f.processFrame()
	}
}

func chunkRequirements(required bool, tags ...string) []mapChunkRequirement {
	r := make([]mapChunkRequirement, len(tags))
	for i, tag := range tags {
		r[i] = mapChunkRequirement{tag: newTag(tag), required: required}
	}
	return r
}

func (l *mapDataLoader) readChunks(tags []mapChunkRequirement) {
	r := NewDataReaderLe(l.data)
	chunks := map[tagT][]DataReader{}
	for r.Left() != 0 {
		var tag tagT
		copy(tag[:], r.GetN(4))
		length := r.GetU32()
		if uint(length) > r.Left() {
			break
		}
		chunks[tag] = append(chunks[tag], NewDataReaderLe(r.GetN(uint(length))))
	}
	for _, v := range tags {
		readers, ok := chunks[v.tag]
		if !ok {
			if v.required {
				throwError("map is missing required chunk '%s'", v.tag)
			}
			continue
		}
		tagFunc := l.tagFuncs[v.tag]
		if tagFunc == nil {
			throwError("tag '%s' is missing a function", v.tag)
		}
		for _, chunk := range readers {
			tagFunc(chunk)
		}
	}
}

func (l *mapDataLoader) readVersionChunks() {
	switch l.version {
	case 59, 63:
		if l.useMapSettings {
			l.readChunks(append(
				chunkRequirements(true, "STR ", "MTXM", "THG2", "MASK", "UNIS", "UPGS", "TECS", "PUNI", "UPGR", "PTEC"),
				append(
					chunkRequirements(false, "UNIx", "UPGx", "TECx", "PUPx", "PTEx"),
					chunkRequirements(true, "UNIT", "UPRP", "MRGN", "TRIG")...,
				)...,
			))
		} else {
			l.readChunks(chunkRequirements(true, "STR ", "MTXM", "THG2", "UNIT"))
		}
	case 205:
		if l.useMapSettings {
			l.readChunks(chunkRequirements(true, "STR ", "MTXM", "THG2", "MASK", "UNIx", "UPGx", "TECx", "PUNI", "PUPx", "PTEx", "UNIT", "UPRP", "MRGN", "TRIG", "COLR"))
		} else {
			l.readChunks(chunkRequirements(true, "STR ", "MTXM", "THG2", "UNIT", "COLR"))
		}
	default:
		throwError("unsupported map version %d", l.version)
	}
}

func (l *mapDataLoader) initPlayers() {
	st := l.f.st
	for i := 0; i != 12; i++ {
		st.SharedVision[i] = 1 << i
		st.Players[i].Color = int32(i)
		if st.Players[i].Controller == PlayerControllerRescuePassive || st.Players[i].Controller == PlayerControllerNeutral {
			for i2 := 0; i2 < 12; i2++ {
				st.Alliances[i][i2] = 1
				st.Alliances[i2][i] = 1
			}
		}
		if !l.useMapSettings {
			if st.Players[i].Controller == PlayerControllerOpen || st.Players[i].Controller == PlayerControllerComputer {
				st.Players[i].Race = Race(5)
			}
		}
	}
}

func (l *mapDataLoader) occupyOpenAndComputerSlots() {
	st := l.f.st
	for i := 0; i != 12; i++ {
		if st.Players[i].Controller == PlayerControllerOpen {
			st.Players[i].Controller = PlayerControllerOccupied
		}
		if st.Players[i].Controller == PlayerControllerComputer {
			st.Players[i].Controller = PlayerControllerComputerGame
		}
	}
}

func (l *mapDataLoader) createAllStartingUnits() {
	f := l.f
	st := f.st
	for i := 8; i != 0; {
		i--
		controller := st.Players[i].Controller
		race := st.Players[i].Race
		if controller != PlayerControllerOccupied && controller != PlayerControllerComputerGame {
			continue
		}
		if (!l.useMapSettings || f.setupInfo.CreateMeleeUnitsForPlayer[i]) && !f.setupInfo.CreateNoUnits {
			f.createStartingUnits(int32(i), f.gameSt.StartLocations[i], race)
		}
	}
}

func (l *mapDataLoader) shareForceVision() {
	st := l.f.st
	for i := 0; i != 12; i++ {
		if st.Players[i].Controller != PlayerControllerOccupied {
			continue
		}
		force := st.Players[i].Force
		if force < 1 || force > 4 {
			continue
		}
		if l.f.gameSt.Forces[force-1].Flags&8 == 0 {
			continue
		}
		for i2 := 0; i2 != 12; i2++ {
			if st.Players[i2].Controller != PlayerControllerOccupied {
				continue
			}
			if st.Players[i2].Force == force {
				st.SharedVision[i] |= 1 << i2
			}
		}
	}
}

func (l *mapDataLoader) allyForces() {
	st := l.f.st
	for i := 0; i != 8; i++ {
		for i2 := 0; i2 != 8; i2++ {
			if i == i2 || st.Players[i].Force != st.Players[i2].Force {
				continue
			}
			if st.Players[i].Controller != PlayerControllerOccupied || st.Players[i2].Controller != PlayerControllerOccupied {
				continue
			}
			st.Alliances[i][i2] = 2
			st.SharedVision[i] |= 1 << i2
		}
	}
}

func (l *mapDataLoader) registerTagFuncs() {
	l.tagFuncs = map[tagT]func(r DataReader){
		newTag("VER "): l.tagVer,
		newTag("DIM "): l.tagDim,
		newTag("ERA "): l.tagEra,
		newTag("OWNR"): l.tagOwnr,
		newTag("SIDE"): l.tagSide,
		newTag("STR "): l.tagStr,
		newTag("SPRP"): l.tagSprp,
		newTag("FORC"): l.tagForc,
		newTag("VCOD"): func(DataReader) {},
		newTag("MTXM"): l.tagMtxm,
		newTag("THG2"): l.tagThg2,
		newTag("MASK"): l.tagMask,
		newTag("UNIS"): l.mapSettingsOnly(func(r DataReader) { l.units(r, false) }),
		newTag("UPGS"): l.mapSettingsOnly(func(r DataReader) { l.upgrades(r, false) }),
		newTag("TECS"): l.mapSettingsOnly(func(r DataReader) { l.techdata(r, false) }),
		newTag("PUNI"): l.mapSettingsOnly(l.tagPuni),
		newTag("UPGR"): l.mapSettingsOnly(func(r DataReader) { l.upgradeRestrictions(r, false) }),
		newTag("PTEC"): l.mapSettingsOnly(func(r DataReader) { l.techRestrictions(r, false) }),
		newTag("UNIx"): l.mapSettingsOnly(func(r DataReader) { l.units(r, true) }),
		newTag("UPGx"): l.mapSettingsOnly(func(r DataReader) { l.upgrades(r, true) }),
		newTag("TECx"): l.mapSettingsOnly(func(r DataReader) { l.techdata(r, true) }),
		newTag("PUPx"): l.mapSettingsOnly(func(r DataReader) { l.upgradeRestrictions(r, true) }),
		newTag("PTEx"): l.mapSettingsOnly(func(r DataReader) { l.techRestrictions(r, true) }),
		newTag("UNIT"): l.tagUnit,
		newTag("UPRP"): l.tagUprp,
		newTag("MRGN"): l.tagMrgn,
		newTag("TRIG"): l.tagTrig,
		newTag("COLR"): l.tagColr,
	}
}

func (l *mapDataLoader) mapSettingsOnly(tagFunc func(r DataReader)) func(r DataReader) {
	return func(r DataReader) {
		if !l.useMapSettings {
			throwError("wrong game mode")
		}
		tagFunc(r)
	}
}

func (l *mapDataLoader) tagVer(r DataReader) {
	l.version = int32(r.GetU16())
}

func (l *mapDataLoader) tagDim(r DataReader) {
	g := l.f.gameSt
	g.MapTileWidth = uint(r.GetU16())
	g.MapTileHeight = uint(r.GetU16())
	g.MapWidth = g.MapTileWidth * 32
	g.MapHeight = g.MapTileHeight * 32
	g.MapWalkWidth = g.MapTileWidth * 4
	g.MapWalkHeight = g.MapTileHeight * 4
}

func (l *mapDataLoader) tagEra(r DataReader) {
	l.f.gameSt.TilesetIndex = uint(r.GetU16()) % 8
}

func (l *mapDataLoader) tagOwnr(r DataReader) {
	for i := 0; i != 12; i++ {
		l.f.st.Players[i].Controller = int32(r.GetI8())
	}
}

func (l *mapDataLoader) tagSide(r DataReader) {
	for i := 0; i != 12; i++ {
		l.f.st.Players[i].Race = Race(r.GetI8())
	}
}

func (l *mapDataLoader) tagStr(r DataReader) {
	g := l.f.gameSt
	if r.Left() < 2 {
		return
	}
	start := r
	num := uint(r.GetU16())
	g.MapStrings = make([]string, num)
	for i := uint(0); i != num; i++ {
		offset := uint(r.GetU16())
		t := start
		if offset < t.Left() {
			t.Skip(offset)
			begin := t.Tell()
			for t.GetU8() != 0 {
			}
			g.MapStrings[i] = string(t.data[begin : t.Tell()-1])
		} else {
			g.MapStrings[i] = "<invalid string offset>"
		}
	}
}

func (l *mapDataLoader) tagSprp(r DataReader) {
	g := l.f.gameSt
	g.ScenarioName = l.f.getMapString(uint(r.GetU16()))
	g.ScenarioDescription = l.f.getMapString(uint(r.GetU16()))
}

func (l *mapDataLoader) tagForc(r DataReader) {
	st := l.f.st
	g := l.f.gameSt
	for i := 0; i != 12; i++ {
		st.Players[i].Force = 0
	}
	for i := 0; i != 4; i++ {
		g.Forces[i].Name = ""
		g.Forces[i].Flags = 0
	}
	if r.Left() == 0 {
		return
	}
	for i := 0; i != 8; i++ {
		st.Players[i].Force = int32(r.GetU8())
	}
	for i := 0; i != 4; i++ {
		g.Forces[i].Name = l.f.getMapString(uint(r.GetU16()))
	}
	for i := 0; i != 4; i++ {
		g.Forces[i].Flags = r.GetU8()
	}
}

const tileFlagsSetFromMegaTile = TileFlagWalkable | TileFlagUnwalkable | TileFlagVeryHigh | TileFlagMiddle | TileFlagHigh | TileFlagPartiallyWalkable

func (l *mapDataLoader) tagMtxm(r DataReader) {
	f := l.f
	g := f.gameSt
	l.loadMtxmTiles(r)

	const clearedFlags = ^int32(TileFlagWalkable | TileFlagHasCreep | TileFlagPartiallyWalkable)
	f.tilesFlagsAnd(0, g.MapTileHeight-2, 5, 1, clearedFlags)
	f.tilesFlagsOr(0, g.MapTileHeight-2, 5, 1, TileFlagUnbuildable)
	f.tilesFlagsAnd(g.MapTileWidth-5, g.MapTileHeight-2, 5, 1, clearedFlags)
	f.tilesFlagsOr(g.MapTileWidth-5, g.MapTileHeight-2, 5, 1, TileFlagUnbuildable)

	f.tilesFlagsAnd(0, g.MapTileHeight-1, g.MapTileWidth, 1, clearedFlags)
	f.tilesFlagsOr(0, g.MapTileHeight-1, g.MapTileWidth, 1, TileFlagUnbuildable)

	f.regionsCreate()
}

func (l *mapDataLoader) loadMtxmTiles(r DataReader) {
	g := l.f.gameSt
	st := l.f.st
	g.GfxTiles = resizeSlice(g.GfxTiles, int(g.MapTileWidth*g.MapTileHeight))
	for i := 0; r.Left() != 0; i++ {
		if r.Left() == 1 {
			g.GfxTiles[i].RawValue &= 0xff00
			g.GfxTiles[i].RawValue |= uint16(r.GetU8())
			break
		}
		if i >= len(g.GfxTiles) {
			break
		}
		g.GfxTiles[i] = TileID{RawValue: r.GetU16()}
	}
	for i := range g.GfxTiles {
		tileID := g.GfxTiles[i]
		if tileID.GroupIndex() >= uint(len(g.Cv5)) {
			tileID = TileID{}
		}
		cv5 := &g.Cv5[tileID.GroupIndex()]
		megatileIndex := uint(cv5.MegaTileIndex[tileID.SubtileIndex()])
		cv5Flags := int32(cv5.Flags) &^ tileFlagsSetFromMegaTile
		st.TilesMegaTileIndex[i] = uint16(megatileIndex)
		st.Tiles[i].Flags = uint16(int32(g.MegaTileFlags[megatileIndex]) | cv5Flags)
		if tileID.HasCreep() {
			st.TilesMegaTileIndex[i] |= 0x8000
			st.Tiles[i].Flags |= TileFlagHasCreep
		}
	}
}

func isMapDoorOwnedByNeutral(unitType UnitTypes) bool {
	switch unitType {
	case UnitTypesSpecialUpperLevelDoor, UnitTypesSpecialRightUpperLevelDoor, UnitTypesSpecialPitDoor, UnitTypesSpecialRightPitDoor:
		return true
	}
	return false
}

func (l *mapDataLoader) tagThg2(r DataReader) {
	f := l.f
	for r.Left() != 0 {
		id := int32(r.GetU16())
		x := int32(r.GetU16())
		y := int32(r.GetU16())
		owner := int32(r.GetU8())
		r.GetU8()
		r.GetU8()
		flags := int32(r.GetU8())
		if flags&0x10 != 0 {
			if id == 0xffff {
				continue
			}
			f.createThingy(f.getSpriteType(SpriteTypes(id)), Xy{X: x, Y: y}, owner)
			continue
		}
		unitType := UnitTypes(id)
		if isMapDoorOwnedByNeutral(unitType) {
			owner = 11
		}
		if l.useMapSettings || owner == 11 {
			u := f.createInitialUnit(f.getUnitType(unitType), Xy{X: x, Y: y}, owner)
			if flags&0x80 != 0 {
				f.disableThingyUnit(u)
			}
		}
	}
}

func (f *GameLoadFunctions) disableThingyUnit(u *Unit) {
	if !f.gameSt.Rules.DisabledThingyUnits {
		throwError("disable thingy unit")
	}
	if u != nil {
		f.uSetStatusFlag(u, UnitStatusFlagDisabled)
		f.uSetStatusFlag(u, UnitStatusFlagNoCollide)
	}
}

func (l *mapDataLoader) tagMask(r DataReader) {
	g := l.f.gameSt
	st := l.f.st
	mask := r.GetN(min(g.MapTileWidth*g.MapTileHeight, r.Left()))
	for i, v := range mask {
		st.Tiles[i].Visible = v
		st.Tiles[i].Explored = v
	}
}

func weaponCount(broodwar bool) uint {
	if broodwar {
		return 130
	}
	return 100
}

func upgradeCount(broodwar bool) uint {
	if broodwar {
		return 61
	}
	return 46
}

func techCount(broodwar bool) uint {
	if broodwar {
		return 44
	}
	return 24
}

func (l *mapDataLoader) units(r DataReader, broodwar bool) {
	f := l.f
	usesDefaultSettings := DataReaderGetVec[uint8](&r, 228)
	hp := DataReaderGetVec[uint32](&r, 228)
	shieldPoints := DataReaderGetVec[uint16](&r, 228)
	armor := DataReaderGetVec[uint8](&r, 228)
	buildTime := DataReaderGetVec[uint16](&r, 228)
	mineralCost := DataReaderGetVec[uint16](&r, 228)
	gasCost := DataReaderGetVec[uint16](&r, 228)
	stringIndex := DataReaderGetVec[uint16](&r, 228)
	weaponDamage := DataReaderGetVec[uint16](&r, weaponCount(broodwar))
	weaponBonusDamage := DataReaderGetVec[uint16](&r, weaponCount(broodwar))
	for i := 0; i != 228; i++ {
		if usesDefaultSettings[i] != 0 {
			continue
		}
		unitType := f.getUnitType(UnitTypes(i))
		unitType.Hitpoints = Fp8FromRaw(int32(hp[i]))
		unitType.ShieldPoints = int32(shieldPoints[i])
		unitType.Armor = int32(armor[i])
		unitType.BuildTime = int32(buildTime[i])
		unitType.MineralCost = int32(mineralCost[i])
		unitType.GasCost = int32(gasCost[i])
		unitType.UnitMapStringIndex = uint(stringIndex[i])
		attackingType := unitType
		if unitType.TurretUnitType.Ptr != nil {
			attackingType = unitType.TurretUnitType.Ptr
		}
		setWeaponDamage(attackingType.GroundWeapon.Ptr, weaponDamage, weaponBonusDamage)
		setWeaponDamage(attackingType.AirWeapon.Ptr, weaponDamage, weaponBonusDamage)
	}
}

func setWeaponDamage(weapon *WeaponTypeT, weaponDamage, weaponBonusDamage []uint16) {
	if weapon == nil || uint(weapon.ID) >= uint(len(weaponDamage)) {
		return
	}
	weapon.DamageAmount = int32(weaponDamage[weapon.ID])
	weapon.DamageBonus = int32(weaponBonusDamage[weapon.ID])
}

func (l *mapDataLoader) upgrades(r DataReader, broodwar bool) {
	count := upgradeCount(broodwar)
	usesDefaultSettings := DataReaderGetVec[uint8](&r, count)
	if broodwar {
		r.GetU8()
	}
	mineralCost := DataReaderGetVec[uint16](&r, count)
	mineralCostFactor := DataReaderGetVec[uint16](&r, count)
	gasCost := DataReaderGetVec[uint16](&r, count)
	gasCostFactor := DataReaderGetVec[uint16](&r, count)
	timeCost := DataReaderGetVec[uint16](&r, count)
	timeCostFactor := DataReaderGetVec[uint16](&r, count)
	for i := uint(0); i != count; i++ {
		if usesDefaultSettings[i] != 0 {
			continue
		}
		upg := l.f.getUpgradeType(UpgradeTypes(i))
		upg.MineralCostBase = int32(mineralCost[i])
		upg.MineralCostFactor = int32(mineralCostFactor[i])
		upg.GasCostBase = int32(gasCost[i])
		upg.GasCostFactor = int32(gasCostFactor[i])
		upg.TimeCostBase = int32(timeCost[i])
		upg.TimeCostFactor = int32(timeCostFactor[i])
	}
}

func (l *mapDataLoader) techdata(r DataReader, broodwar bool) {
	count := techCount(broodwar)
	usesDefaultSettings := DataReaderGetVec[uint8](&r, count)
	mineralCost := DataReaderGetVec[uint16](&r, count)
	gasCost := DataReaderGetVec[uint16](&r, count)
	buildTime := DataReaderGetVec[uint16](&r, count)
	energyCost := DataReaderGetVec[uint16](&r, count)
	for i := uint(0); i != count; i++ {
		if usesDefaultSettings[i] != 0 {
			continue
		}
		tech := l.f.getTechType(TechTypes(i))
		tech.MineralCost = int32(mineralCost[i])
		tech.GasCost = int32(gasCost[i])
		tech.ResearchTime = int32(buildTime[i])
		tech.EnergyCost = int32(energyCost[i])
	}
}

func playerOrGlobalSetting(playerUsesGlobalDefault uint8, globalValue, playerValue uint8) uint8 {
	if playerUsesGlobalDefault != 0 {
		return globalValue
	}
	return playerValue
}

func (l *mapDataLoader) upgradeRestrictions(r DataReader, broodwar bool) {
	g := l.f.gameSt
	st := l.f.st
	count := upgradeCount(broodwar)
	playerMaxLevel := DataReaderGetVec[uint8](&r, 12*count)
	playerCurLevel := DataReaderGetVec[uint8](&r, 12*count)
	globalMaxLevel := DataReaderGetVec[uint8](&r, count)
	globalCurLevel := DataReaderGetVec[uint8](&r, count)
	playerUsesGlobalDefault := DataReaderGetVec[uint8](&r, 12*count)
	for player := uint(0); player != 12; player++ {
		for upgrade := uint(0); upgrade != count; upgrade++ {
			index := player*count + upgrade
			usesGlobal := playerUsesGlobalDefault[index]
			g.MaxUpgradeLevels[player][upgrade] = int32(playerOrGlobalSetting(usesGlobal, globalMaxLevel[upgrade], playerMaxLevel[index]))
			st.UpgradeLevels[player][upgrade] = int32(playerOrGlobalSetting(usesGlobal, globalCurLevel[upgrade], playerCurLevel[index]))
		}
	}
}

func (l *mapDataLoader) techRestrictions(r DataReader, broodwar bool) {
	g := l.f.gameSt
	st := l.f.st
	count := techCount(broodwar)
	playerAvailable := DataReaderGetVec[uint8](&r, 12*count)
	playerResearched := DataReaderGetVec[uint8](&r, 12*count)
	globalAvailable := DataReaderGetVec[uint8](&r, count)
	globalResearched := DataReaderGetVec[uint8](&r, count)
	playerUsesGlobalDefault := DataReaderGetVec[uint8](&r, 12*count)
	for player := uint(0); player != 12; player++ {
		for tech := uint(0); tech != count; tech++ {
			index := player*count + tech
			usesGlobal := playerUsesGlobalDefault[index]
			g.TechAvailable[player][tech] = playerOrGlobalSetting(usesGlobal, globalAvailable[tech], playerAvailable[index]) != 0
			st.TechResearched[player][tech] = playerOrGlobalSetting(usesGlobal, globalResearched[tech], playerResearched[index]) != 0
		}
	}
}

func (l *mapDataLoader) tagPuni(r DataReader) {
	g := l.f.gameSt
	playerAvailable := DataReaderGetVec[uint8](&r, 12*228)
	globalAvailable := DataReaderGetVec[uint8](&r, 228)
	playerUsesGlobalDefault := DataReaderGetVec[uint8](&r, 12*228)
	for player := 0; player != 12; player++ {
		for unit := 0; unit != 228; unit++ {
			index := player*228 + unit
			g.UnitTypeAllowed[player][unit] = playerOrGlobalSetting(playerUsesGlobalDefault[index], globalAvailable[unit], playerAvailable[index]) != 0
		}
	}
}

type chkUnitEntry struct {
	x               int32
	y               int32
	unitTypeID      UnitTypes
	validFlags      int32
	validProperties int32
	owner           int32
	hpPercent       int32
	shieldPercent   int32
	energyPercent   int32
	resources       int32
	flags           int32
	isInvalid       bool
}

func readChkUnitEntry(r *DataReader) chkUnitEntry {
	var e chkUnitEntry
	r.GetU32()
	e.x = int32(r.GetU16())
	e.y = int32(r.GetU16())
	e.unitTypeID = UnitTypes(r.GetU16())
	r.GetU16()
	e.validFlags = int32(r.GetU16())
	e.validProperties = int32(r.GetU16())
	e.owner = int32(r.GetU8())
	e.hpPercent = int32(r.GetU8())
	e.shieldPercent = int32(r.GetU8())
	e.energyPercent = int32(r.GetU8())
	e.resources = int32(r.GetU32())
	r.GetU16()
	e.flags = int32(r.GetU16())
	r.GetU32()
	r.GetU32()
	e.isInvalid = e.unitTypeID == 0xffff
	if e.isInvalid {
		e.unitTypeID = 0
		e.owner = 8
	}
	return e
}

func (l *mapDataLoader) shouldCreateUnitsForPlayer(owner int32) bool {
	f := l.f
	if f.setupInfo.CreateNoUnits {
		return false
	}
	if owner >= 8 {
		return true
	}
	switch f.st.Players[owner].Controller {
	case PlayerControllerComputerGame, PlayerControllerOccupied, PlayerControllerRescuePassive, PlayerControllerUnusedRescueActive, PlayerControllerNeutral:
		return true
	}
	return false
}

func (l *mapDataLoader) isNeutralUnit(owner int32, unitType *UnitTypeT) bool {
	f := l.f
	if owner == 11 {
		return true
	}
	if f.unitIsMineralField(unitType) {
		return true
	}
	if unitType.ID == UnitTypesResourceVespeneGeyser {
		return true
	}
	if f.unitIsCritter(unitType) {
		return true
	}
	return false
}

func (l *mapDataLoader) tagUnit(r DataReader) {
	f := l.f
	st := f.st
	for r.Left() != 0 {
		e := readChkUnitEntry(&r)

		if uint(e.unitTypeID) >= 228 {
			throwError("UNIT: invalid unit type %d", int32(e.unitTypeID))
		}
		if uint(e.owner) >= 12 {
			throwError("UNIT: invalid owner %d", e.owner)
		}

		unitType := f.getUnitType(e.unitTypeID)

		if unitType.ID == UnitTypesSpecialStartLocation {
			f.gameSt.StartLocations[e.owner] = Xy{X: e.x, Y: e.y}
			continue
		}
		if !l.shouldCreateUnitsForPlayer(e.owner) {
			continue
		}
		if !l.useMapSettings && !l.isNeutralUnit(e.owner, unitType) {
			continue
		}
		if l.useMapSettings {
			if e.owner < 8 && f.setupInfo.CreateMeleeUnitsForPlayer[e.owner] && ^unitType.GroupFlags&GroupFlagsNeutral != 0 {
				continue
			}
		}

		if e.isInvalid {
			f.lcgRand(14)
			if f.isPosInMapBounds(Xy{X: e.x, Y: e.y}) {
				throwError("attempt to create invalid initial unit in valid map bounds")
			}
			continue
		}

		u := f.createInitialUnit(unitType, Xy{X: e.x, Y: e.y}, e.owner)
		if u == nil {
			continue
		}
		l.applyChkUnitProperties(u, e)
	}

	for u := range st.VisibleUnits.All() {
		f.updateUnitSprite(u)
	}
	st.UpdatePsionicMatrix = true
}

func (l *mapDataLoader) applyChkUnitProperties(u *Unit, e chkUnitEntry) {
	f := l.f
	if e.validProperties&0x2 != 0 {
		tmp := int64(u.UnitType.Hitpoints.RawValue)
		tmp = max(tmp*int64(e.hpPercent)/100, 1)
		f.setUnitHp(u, Fp8FromRaw(int32(tmp)))
	}
	if e.validProperties&0x4 != 0 {
		f.setUnitShieldPoints(u, Fp8Integer(u.UnitType.ShieldPoints*e.shieldPercent/100))
	}
	if e.validProperties&0x8 != 0 {
		f.setUnitEnergy(u, f.unitMaxEnergy(u).MulInt(e.energyPercent).DivInt(100))
	}
	if e.validProperties&0x10 != 0 {
		f.setUnitResources(u, e.resources)
	}

	if e.validFlags&8 != 0 && e.flags&8 != 0 {
		f.makeUnitHallucination(u)
	}

	if e.validFlags&0x10 != 0 && e.flags&0x10 != 0 {
		f.uSetStatusFlag(u, UnitStatusFlagInvincible)
	}

	if f.unitIs(u.UnitType, UnitTypesZergBroodling) {
		f.setDefaultRemoveTimer(u)
	}
}

func (l *mapDataLoader) tagUprp(r DataReader) {
	for i := 0; i < 64; i++ {
		r.GetU16()
		r.GetU16()
		r.GetU8()
		r.GetU8()
		r.GetU8()
		r.GetU8()
		r.GetU32()
		r.GetU16()
		r.GetU16()
		r.GetU32()
	}
}

func (l *mapDataLoader) tagMrgn(r DataReader) {
	st := l.f.st
	for r.Left() != 0 {
		left := r.GetI32()
		top := r.GetI32()
		right := r.GetI32()
		bottom := r.GetI32()
		l.f.getMapString(uint(r.GetU16()))
		elevationFlags := int32(r.GetU16())
		st.Locations = append(st.Locations, Location{
			Area:           Rect{From: Xy{X: left, Y: top}, To: Xy{X: right, Y: bottom}},
			ElevationFlags: elevationFlags,
		})
	}
}

func readTriggerCondition(r *DataReader, c *TriggerCondition) {
	c.Location = int32(r.GetU32())
	c.Group = int32(r.GetU32())
	c.CountN = int32(r.GetU32())
	c.UnitID = int32(r.GetU16())
	c.NumN = int32(r.GetU8())
	c.Type = int32(r.GetU8())
	c.ExtraN = int32(r.GetU8())
	c.Flags = int32(r.GetU8())
	c.Unk = int32(r.GetU16())
}

func readTriggerAction(r *DataReader, a *TriggerAction) {
	a.Location = int32(r.GetU32())
	a.StringIndex = int32(r.GetU32())
	a.SoundIndex = int32(r.GetU32())
	a.TimeN = int32(r.GetU32())
	a.GroupN = int32(r.GetU32())
	a.Group2N = int32(r.GetU32())
	a.ExtraN = int32(r.GetU16())
	a.Type = int32(r.GetU8())
	a.NumN = int32(r.GetU8())
	a.Flags = int32(r.GetU8())
	a.Unk = int32(r.GetU16())
	r.GetU8()
}

func (l *mapDataLoader) tagTrig(r DataReader) {
	g := l.f.gameSt
	st := l.f.st
	for r.Left() != 0 {
		g.Triggers = append(g.Triggers, Trigger{})
		t := &g.Triggers[len(g.Triggers)-1]
		for i := range t.Conditions {
			readTriggerCondition(&r, &t.Conditions[i])
		}
		for i := range t.Actions {
			readTriggerAction(&r, &t.Actions[i])
		}
		t.ExecutionFlags = int32(r.GetU32())

		enabledForAny := false
		for i := range t.Enabled {
			t.Enabled[i] = r.GetU8() != 0
			if !enabledForAny && t.Enabled[i] {
				enabledForAny = true
			}
		}
		if !enabledForAny {
			g.Triggers = g.Triggers[:len(g.Triggers)-1]
		}
	}
	for ti := range g.Triggers {
		t := &g.Triggers[ti]
		for i := 0; i != 8; i++ {
			if st.Players[i].Controller != PlayerControllerOccupied {
				continue
			}
			if !t.Enabled[i] && !t.Enabled[17+st.Players[i].Force] && !t.Enabled[17] {
				continue
			}
			st.RunningTriggers[i] = append(st.RunningTriggers[i], RunningTrigger{T: t, Flags: t.ExecutionFlags})
		}
	}
}

func (l *mapDataLoader) tagColr(r DataReader) {
	for i := 0; i != 8; i++ {
		l.f.st.Players[i].Color = int32(r.GetU8())
	}
}
