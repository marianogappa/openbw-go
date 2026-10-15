package bwgame

type Hooks struct {
	PlaySound          func(id int32, position Xy, sourceUnit *Unit, addRaceIndex bool)
	OnUnitDeselect     func(u *Unit)
	OnUnitDestroy      func(u *Unit)
	OnPlayerEliminated func(owner int32)
	OnVictoryState     func(owner int32, state int32)
	OnAction           func(owner int32, action int32)

	OnUnitCreated      func(u *Unit, source CreateSource, parent *Unit)
	OnUnitCompleted    func(u *Unit)
	OnUnitMorphed      func(u *Unit, from *UnitTypeT)
	OnMorphStarted     func(u, morpher *Unit, from, into *UnitTypeT)
	OnKillUnit         func(u *Unit, kill KillContext)
	OnOwnerChanged     func(u *Unit, from int32, reason OwnerReason)
	OnProductionCancel func(c ProductionCancel)
	OnResearch         func(u *Unit, r Research, level int32, completed bool)
	OnPlayerLeft       func(owner int32, dropped bool)
	OnActionRejected   func(u *Unit, reason RejectReason)
	OnBuildQueue       func(u *Unit, op BuildQueueOp, index int)
	OnAddonBuildType   func(u *Unit, set bool)
	OnUnitDamaged      func(u, source *Unit, sourceOwner int32, damage Fp8)
}

type CreateSource uint8

const (
	SourceMap CreateSource = iota
	SourceStart
	SourceTrained
	SourceBuilt
	SourceLarva
	SourceSecondEggUnit
	SourceBroodling
	SourceHallucination
	SourceSpell
	SourceScanner
	SourceSpiderMine
	SourceDroneReturned
	SourceTrigger
)

type DeathReason uint8

const (
	ReasonOther DeathReason = iota
	ReasonDamage
	ReasonContainer
	ReasonSpell
	ReasonCancelled
	ReasonExpired
	ReasonDepleted
	ReasonRemoved
	ReasonDetonated
	ReasonLinked
	ReasonTrigger
	ReasonOffCreep
	ReasonCrushed
	ReasonBlocked
)

type KillContext struct {
	Reason      DeathReason
	Killer      *Unit
	KillerOwner int32
	Container   *Unit
}

type OwnerReason uint8

const (
	OwnerOther OwnerReason = iota
	OwnerMindControl
	OwnerPlayerLeft
	OwnerRefinery
	OwnerInfest
	OwnerAddon
	OwnerTrigger
)

type ProductionKind uint8

const (
	ProductionUnit ProductionKind = iota
	ProductionBuilding
	ProductionMorph
	ProductionResearch
	ProductionUpgrade
	ProductionAddon
)

type CancelReason uint8

const (
	CancelOther CancelReason = iota
	CancelCommand
	CancelProducerDestroyed
	CancelOwnerChanged
	CancelNoSupply
)

type Research struct {
	Upgrade bool
	ID      uint8
}

type ProductionCancel struct {
	Owner    int32
	Producer *Unit
	Unit     *Unit
	What     ProductionKind
	Type     *UnitTypeT
	Research Research
	Why      CancelReason
}

type RejectReason uint8

const (
	RejectNoUnit RejectReason = iota
	RejectNotOwner
	RejectRequirements
	RejectInvalid
	RejectQueueFull
	RejectBusy
	RejectSupply
	RejectMinerals
	RejectGas
	RejectPlacement
)

type BuildQueueOp uint8

const (
	BuildQueuePush BuildQueueOp = iota
	BuildQueueClearSingle
	BuildQueueErase
	BuildQueueReset
)

func (f *StateFunctions) playSoundAt(id int32, position Xy, sourceUnit *Unit, addRaceIndex bool) {
	if f.Hooks.PlaySound != nil {
		f.Hooks.PlaySound(id, position, sourceUnit, addRaceIndex)
	}
}

func (f *StateFunctions) onUnitDeselect(u *Unit) {
	if f.deselectUnit != nil {
		f.deselectUnit(u)
	}
	if f.Hooks.OnUnitDeselect != nil {
		f.Hooks.OnUnitDeselect(u)
	}
}

func (f *StateFunctions) onUnitDestroy(u *Unit) {
	if f.Hooks.OnUnitDestroy != nil {
		f.Hooks.OnUnitDestroy(u)
	}
}

func (f *StateFunctions) onPlayerEliminated(owner int32) {
	if f.Hooks.OnPlayerEliminated != nil {
		f.Hooks.OnPlayerEliminated(owner)
	}
}

func (f *StateFunctions) onVictoryState(owner int32, state int32) {
	if f.Hooks.OnVictoryState != nil {
		f.Hooks.OnVictoryState(owner, state)
	}
}

func (f *ActionFunctions) onAction(owner int32, action int32) {
	if f.Hooks.OnAction != nil {
		f.Hooks.OnAction(owner, action)
	}
}

func (f *StateFunctions) onUnitCreated(u *Unit, source CreateSource, parent *Unit) {
	if f.Hooks.OnUnitCreated != nil {
		f.Hooks.OnUnitCreated(u, source, parent)
	}
}

func (f *StateFunctions) onUnitCompleted(u *Unit) {
	if f.Hooks.OnUnitCompleted != nil {
		f.Hooks.OnUnitCompleted(u)
	}
}

func (f *StateFunctions) onUnitMorphed(u *Unit, from *UnitTypeT) {
	if f.Hooks.OnUnitMorphed != nil {
		f.Hooks.OnUnitMorphed(u, from)
	}
}

func (f *StateFunctions) onMorphStarted(u, morpher *Unit, from, into *UnitTypeT) {
	if f.Hooks.OnMorphStarted != nil {
		f.Hooks.OnMorphStarted(u, morpher, from, into)
	}
}

func killedBy(reason DeathReason, killer *Unit) KillContext {
	if killer == nil {
		return KillContext{Reason: reason, KillerOwner: -1}
	}
	return KillContext{Reason: reason, Killer: killer, KillerOwner: killer.Owner}
}

func noKiller(reason DeathReason) KillContext {
	return KillContext{Reason: reason, KillerOwner: -1}
}

func killedInContainer(container *Unit) KillContext {
	return KillContext{Reason: ReasonContainer, KillerOwner: -1, Container: container}
}

func (f *StateFunctions) killUnitWith(u *Unit, kill KillContext) {
	if f.Hooks.OnKillUnit == nil {
		f.killUnit(u)
		return
	}
	prev := f.killContext
	f.killContext = kill
	f.killUnit(u)
	f.killContext = prev
}

func (f *StateFunctions) onUnitRemovedWithoutKill(u *Unit) {
	if f.Hooks.OnKillUnit != nil {
		f.Hooks.OnKillUnit(u, noKiller(ReasonRemoved))
	}
}

func (f *StateFunctions) setUnitOwnerWith(u *Unit, owner int32, incrementScore bool, reason OwnerReason) {
	prev := f.swapOwnerReason(reason)
	f.setUnitOwner(u, owner, incrementScore)
	f.ownerReason = prev
}

func (f *StateFunctions) swapOwnerReason(reason OwnerReason) OwnerReason {
	prev := f.ownerReason
	if f.Hooks.OnOwnerChanged != nil {
		f.ownerReason = reason
	}
	return prev
}

func (f *StateFunctions) onOwnerChanged(u *Unit, from int32) {
	if f.Hooks.OnOwnerChanged != nil && !f.utTurret(u.UnitType) {
		f.Hooks.OnOwnerChanged(u, from, f.ownerReason)
	}
}

func (f *StateFunctions) swapCancelReason(why CancelReason) CancelReason {
	prev := f.cancelReason
	if f.Hooks.OnProductionCancel != nil {
		f.cancelReason = why
	}
	return prev
}

func (f *StateFunctions) onBuildingUnitCancelled(u *Unit) {
	if f.Hooks.OnProductionCancel == nil {
		return
	}
	producer := f.cancelProducer
	if producer == nil {
		producer = u.ConnectedUnit
	}
	what, unitType := f.cancelledUnitKind(u)
	f.Hooks.OnProductionCancel(ProductionCancel{Owner: u.Owner, Producer: producer, Unit: u, What: what, Type: unitType, Why: f.cancelReason})
}

func (f *StateFunctions) cancelledUnitKind(u *Unit) (ProductionKind, *UnitTypeT) {
	if f.uGroundedBuilding(u) {
		switch {
		case f.unitRace(u.UnitType) == RaceZerg && f.unitIsMorphingBuilding(u) && !u.BuildQueue.Empty():
			return ProductionMorph, u.BuildQueue.Front()
		case f.utAddon(u.UnitType):
			return ProductionAddon, u.UnitType
		}
		return ProductionBuilding, u.UnitType
	}
	if f.unitIsEgg(u.UnitType) || f.unitIs(u.UnitType, UnitTypesZergCocoon) || f.unitIs(u.UnitType, UnitTypesZergLurkerEgg) {
		if !u.BuildQueue.Empty() {
			return ProductionMorph, u.BuildQueue.Front()
		}
		return ProductionMorph, u.UnitType
	}
	return ProductionUnit, u.UnitType
}

func (f *StateFunctions) cancelBuildingUnitOf(producer *Unit, u *Unit) bool {
	if f.Hooks.OnProductionCancel == nil {
		return f.cancelBuildingUnit(u)
	}
	prev := f.cancelProducer
	f.cancelProducer = producer
	cancelled := f.cancelBuildingUnit(u)
	f.cancelProducer = prev
	return cancelled
}

func (f *StateFunctions) onQueuedProductionCancelled(u *Unit, unitType *UnitTypeT) {
	if f.Hooks.OnProductionCancel == nil {
		return
	}
	what := ProductionUnit
	switch {
	case f.unitTypeIsMorphingBuilding(unitType):
		what = ProductionMorph
	case f.utBuilding(unitType):
		what = ProductionBuilding
	case f.unitRace(u.UnitType) == RaceZerg && !f.unitIs(u.UnitType, UnitTypesZergInfestedCommandCenter):
		what = ProductionMorph
	}
	f.Hooks.OnProductionCancel(ProductionCancel{Owner: u.Owner, Producer: u, What: what, Type: unitType, Why: f.cancelReason})
}

func (f *StateFunctions) onResearchCancelled(u *Unit, r Research, unitDestroyed bool) {
	if f.Hooks.OnProductionCancel == nil {
		return
	}
	what := ProductionResearch
	if r.Upgrade {
		what = ProductionUpgrade
	}
	why := f.cancelReason
	if unitDestroyed {
		why = CancelProducerDestroyed
	}
	f.Hooks.OnProductionCancel(ProductionCancel{Owner: u.Owner, Producer: u, What: what, Research: r, Why: why})
}

func (f *StateFunctions) onResearch(u *Unit, r Research, level int32, completed bool) {
	if f.Hooks.OnResearch != nil {
		f.Hooks.OnResearch(u, r, level, completed)
	}
}

func (f *StateFunctions) onPlayerLeft(owner int32, dropped bool) {
	if f.Hooks.OnPlayerLeft != nil {
		f.Hooks.OnPlayerLeft(owner, dropped)
	}
}

func (f *StateFunctions) onActionRejected(u *Unit, reason RejectReason) {
	if f.Hooks.OnActionRejected != nil {
		f.Hooks.OnActionRejected(u, reason)
	}
}

func (f *StateFunctions) onActionRejectedForCost(u *Unit, owner int32, minerals int32) {
	if f.Hooks.OnActionRejected == nil {
		return
	}
	if f.st.CurrentMinerals[owner] < minerals {
		f.Hooks.OnActionRejected(u, RejectMinerals)
	} else {
		f.Hooks.OnActionRejected(u, RejectGas)
	}
}

func (f *StateFunctions) onBuildQueue(u *Unit, op BuildQueueOp, index int) {
	if f.Hooks.OnBuildQueue != nil {
		f.Hooks.OnBuildQueue(u, op, index)
	}
}

func (f *StateFunctions) onAddonBuildType(u *Unit, set bool) {
	if f.Hooks.OnAddonBuildType != nil {
		f.Hooks.OnAddonBuildType(u, set)
	}
}

func (f *StateFunctions) onUnitDamaged(u, source *Unit, sourceOwner int32, damage Fp8) {
	if f.Hooks.OnUnitDamaged != nil {
		f.Hooks.OnUnitDamaged(u, source, sourceOwner, damage)
	}
}
