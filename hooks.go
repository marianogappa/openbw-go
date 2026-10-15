package openbw

import (
	"github.com/marianogappa/openbw-go/bwgame"
)

type killerInfo struct {
	killer      UnitID
	killerType  UnitType
	killerOwner int8
	parent      UnitID
}

func (g *Game) install(user bwgame.Hooks) {
	var h bwgame.Hooks
	if g.wants(KindUnitCreated) {
		h.OnUnitCreated = g.onUnitCreated
	}
	if g.wants(KindUnitCompleted) {
		h.OnUnitCompleted = g.onUnitCompleted
	}
	if g.wants(KindUnitMorphed) {
		h.OnUnitMorphed = g.onUnitMorphed
	}
	if g.wants(KindUnitDestroyed) {
		h.OnKillUnit = g.onKillUnit
	}
	if g.wants(KindUnitOwnerChanged) {
		h.OnOwnerChanged = g.onOwnerChanged
	}
	if g.wants(KindProductionCancelled) {
		h.OnProductionCancel = g.onProductionCancel
	}
	if g.wants(KindResearchStarted | KindResearchCompleted) {
		h.OnResearch = g.onResearch
	}
	if g.wants(KindPlayerLeft) {
		h.OnPlayerLeft = g.onPlayerLeft
	}
	if g.wants(KindPlayerDefeated) {
		h.OnPlayerEliminated = g.onPlayerEliminated
	}
	if g.wants(KindPlayerWon) {
		h.OnVictoryState = g.onVictoryState
	}
	if g.wants(KindCommandResult) {
		g.cmd = newCommandTracker()
		g.funcs.BeforeAction = g.beforeAction
		g.funcs.AfterAction = g.afterAction
		h.OnActionRejected = g.cmd.onRejected
		h.OnBuildQueue = g.cmd.onBuildQueue
		h.OnAddonBuildType = g.cmd.onAddonBuildType
	}
	if g.wants(KindMorphStarted) || g.cmd != nil && g.wants(KindUnitCreated) {
		h.OnMorphStarted = g.onMorphStarted
	}
	g.funcs.Hooks = chainHooks(h, user)
}

func isDying(u *bwgame.Unit) bool {
	return u.OrderType != nil && u.OrderType.ID == bwgame.OrdersDie
}

func isTurret(u *bwgame.Unit) bool {
	return u.UnitType.Flags&bwgame.UnitTypeTFlagTurret != 0
}

func classOf(u *bwgame.Unit) UnitClass {
	if u.StatusFlags&bwgame.UnitStatusFlagHallucination != 0 {
		return ClassHallucination
	}
	switch u.UnitType.ID {
	case bwgame.UnitTypesProtossInterceptor, bwgame.UnitTypesProtossScarab:
		return ClassFighter
	case bwgame.UnitTypesTerranVultureSpiderMine:
		return ClassSpiderMine
	case bwgame.UnitTypesSpellDarkSwarm, bwgame.UnitTypesSpellDisruptionWeb, bwgame.UnitTypesSpellScannerSweep:
		return ClassSpell
	}
	return ClassNormal
}

func (g *Game) reported(u *bwgame.Unit) (UnitClass, bool) {
	if isTurret(u) {
		return 0, false
	}
	class := classOf(u)
	return class, g.classes.Has(class)
}

func isAddonType(t *bwgame.UnitTypeT) bool {
	return t.Flags&bwgame.UnitTypeTFlagAddon != 0
}

func (g *Game) onUnitCreated(u *bwgame.Unit, source bwgame.CreateSource, parent *bwgame.Unit) {
	if isTurret(u) {
		return
	}
	class := classOf(u)
	if source == bwgame.SourceHallucination {
		class = ClassHallucination
	}
	if !g.classes.Has(class) {
		return
	}
	g.emit(&UnitCreated{
		Unit:   unitIDOf(u),
		Type:   unitTypeOf(u.UnitType),
		Owner:  int8(u.Owner),
		X:      u.Position.X,
		Y:      u.Position.Y,
		Parent: unitIDOf(parent),
		Source: CreateSource(source),
		Class:  class,
	}, g.createdCmd(u, source, parent))
}

func (g *Game) createdCmd(u *bwgame.Unit, source bwgame.CreateSource, parent *bwgame.Unit) int32 {
	if g.cmd == nil {
		return -1
	}
	linked := int32(-1)
	switch {
	case parent == nil:
	case source == bwgame.SourceTrained:
		linked = g.cmd.front(parent)
	case source == bwgame.SourceBuilt && isAddonType(u.UnitType):
		linked = g.cmd.addons[parent.Index]
	case source == bwgame.SourceBuilt:
		linked = g.cmd.front(parent)
	case source == bwgame.SourceSecondEggUnit:
		id := unitIDOf(parent)
		if cmd, ok := g.cmd.eggs[id]; ok {
			delete(g.cmd.eggs, id)
			linked = cmd
		}
	}
	if linked >= 0 {
		return linked
	}
	return g.cmd.sync()
}

func (g *Game) syncCmd() int32 {
	if g.cmd == nil {
		return -1
	}
	return g.cmd.sync()
}

func (g *Game) onUnitCompleted(u *bwgame.Unit) {
	class, ok := g.reported(u)
	if !ok {
		return
	}
	g.emit(&UnitCompleted{Unit: unitIDOf(u), Type: unitTypeOf(u.UnitType), Owner: int8(u.Owner), Class: class}, g.syncCmd())
}

func (g *Game) onUnitMorphed(u *bwgame.Unit, from *bwgame.UnitTypeT) {
	class, ok := g.reported(u)
	if !ok {
		return
	}
	g.emit(&UnitMorphed{Unit: unitIDOf(u), From: unitTypeOf(from), To: unitTypeOf(u.UnitType), Owner: int8(u.Owner), Class: class}, g.syncCmd())
}

func (g *Game) onMorphStarted(u, morpher *bwgame.Unit, from, into *bwgame.UnitTypeT) {
	cmd := g.syncCmd()
	if g.cmd != nil {
		if linked := g.cmd.frontOrCleared(morpher); linked >= 0 {
			cmd = linked
		}
		if into.Flags&bwgame.UnitTypeTFlagTwoUnitsInOneEgg != 0 {
			g.cmd.eggs[unitIDOf(u)] = cmd
		}
	}
	if !g.wants(KindMorphStarted) {
		return
	}
	class, ok := g.reported(u)
	if !ok {
		return
	}
	g.emit(&MorphStarted{
		Unit:    unitIDOf(u),
		Morpher: unitIDOf(morpher),
		From:    unitTypeOf(from),
		Into:    unitTypeOf(into),
		Owner:   int8(u.Owner),
		Class:   class,
	}, cmd)
}

func isRefinery(t *bwgame.UnitTypeT) bool {
	switch t.ID {
	case bwgame.UnitTypesTerranRefinery, bwgame.UnitTypesZergExtractor, bwgame.UnitTypesProtossAssimilator:
		return true
	}
	return false
}

func hasLoadedUnits(u *bwgame.Unit) bool {
	for _, id := range u.LoadedUnits {
		if id.RawValue != 0 {
			return true
		}
	}
	return false
}

func killerOf(kill bwgame.KillContext) killerInfo {
	info := killerInfo{killerType: UnitTypeNone, killerOwner: int8(kill.KillerOwner)}
	if k := kill.Killer; k != nil {
		info.killer = unitIDOf(k)
		info.killerType = unitTypeOf(k.UnitType)
		if k.UnitType.ID == bwgame.UnitTypesProtossInterceptor {
			info.parent = unitIDOf(k.Fighter.Parent)
		}
	}
	return info
}

func (g *Game) onKillUnit(u *bwgame.Unit, kill bwgame.KillContext) {
	if isTurret(u) {
		return
	}
	id := unitIDOf(u)
	killer := killerOf(kill)
	if kill.Reason == bwgame.ReasonContainer && kill.Killer == nil && kill.Container != nil {
		if stored, ok := g.containerKills[unitIDOf(kill.Container)]; ok {
			killer = stored
		}
	}
	if hasLoadedUnits(u) {
		g.containerKills[id] = killer
	}
	class := classOf(u)
	if !g.classes.Has(class) {
		return
	}
	g.emit(&UnitDestroyed{
		Unit:         id,
		Type:         unitTypeOf(u.UnitType),
		Owner:        int8(u.Owner),
		X:            u.Position.X,
		Y:            u.Position.Y,
		Reason:       DeathReason(kill.Reason),
		Killer:       killer.killer,
		KillerType:   killer.killerType,
		KillerOwner:  killer.killerOwner,
		KillerParent: killer.parent,
		Container:    unitIDOf(kill.Container),
		Reverted:     isRefinery(u.UnitType),
		Class:        class,
	}, g.syncCmd())
}

func (g *Game) onOwnerChanged(u *bwgame.Unit, from int32, reason bwgame.OwnerReason) {
	class, ok := g.reported(u)
	if !ok || isDying(u) {
		return
	}
	g.emit(&UnitOwnerChanged{
		Unit:   unitIDOf(u),
		Type:   unitTypeOf(u.UnitType),
		From:   int8(from),
		To:     int8(u.Owner),
		Reason: OwnerReason(reason),
		Class:  class,
	}, g.syncCmd())
}

func (g *Game) onProductionCancel(c bwgame.ProductionCancel) {
	e := &ProductionCancelled{
		Owner:    int8(c.Owner),
		Producer: unitIDOf(c.Producer),
		Unit:     unitIDOf(c.Unit),
		What:     ProductionKind(c.What),
		Type:     unitTypeOf(c.Type),
		Research: NoResearch,
		Why:      CancelReason(c.Why),
	}
	if e.What == ProductionResearch || e.What == ProductionUpgrade {
		e.Research = researchOf(c.Research)
	}
	g.emit(e, g.syncCmd())
}

func (g *Game) onResearch(u *bwgame.Unit, r bwgame.Research, level int32, completed bool) {
	switch {
	case completed && g.wants(KindResearchCompleted):
		g.emit(&ResearchCompleted{Owner: int8(u.Owner), Unit: unitIDOf(u), Research: researchOf(r), Level: level}, g.syncCmd())
	case !completed && g.wants(KindResearchStarted):
		g.emit(&ResearchStarted{Owner: int8(u.Owner), Unit: unitIDOf(u), Research: researchOf(r), Level: level}, g.syncCmd())
	}
}

func (g *Game) onPlayerLeft(owner int32, dropped bool) {
	g.emit(&PlayerLeft{Owner: int8(owner), Dropped: dropped}, g.syncCmd())
}

func (g *Game) onPlayerEliminated(owner int32) {
	g.emit(&PlayerDefeated{Owner: int8(owner)}, g.syncCmd())
}

func (g *Game) onVictoryState(owner, state int32) {
	if s := VictoryState(state); s.Won() {
		g.emit(&PlayerWon{Owner: int8(owner), State: s}, g.syncCmd())
	}
}
