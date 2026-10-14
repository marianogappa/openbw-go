package bwgame

func (f *StateFunctions) unitCanBuild(u *Unit, unitType *UnitTypeT) bool {
	panic(notPorted("unit_can_build"))
}

func (f *StateFunctions) unitCanReceiveOrder(u *Unit, order *OrderTypeT, owner int32) bool {
	panic(notPorted("unit_can_receive_order"))
}

func (f *StateFunctions) unitOrderAllowed(u *Unit, order *OrderTypeT) bool {
	panic(notPorted("unit_order_allowed"))
}

func (f *StateFunctions) unitBuildOrderValid(u *Unit, order *OrderTypeT, unitType *UnitTypeT, owner int32) bool {
	panic(notPorted("unit_build_order_valid"))
}

func (f *StateFunctions) unitCanResearchOwner(u *Unit, tech *TechTypeT, owner int32) bool {
	panic(notPorted("unit_can_research"))
}

func (f *StateFunctions) unitCanResearch(u *Unit, tech *TechTypeT) bool {
	panic(notPorted("unit_can_research"))
}

func (f *StateFunctions) unitCanUpgradeOwner(u *Unit, upgrade *UpgradeTypeT, owner int32) bool {
	panic(notPorted("unit_can_upgrade"))
}

func (f *StateFunctions) unitCanUpgrade(u *Unit, upgrade *UpgradeTypeT) bool {
	panic(notPorted("unit_can_upgrade"))
}

func (f *StateFunctions) unitCanUseTechOwner(u *Unit, tech *TechTypeT, owner int32) bool {
	panic(notPorted("unit_can_use_tech"))
}

func (f *StateFunctions) unitCanUseTech(u *Unit, tech *TechTypeT) bool {
	panic(notPorted("unit_can_use_tech"))
}
