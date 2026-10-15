package bwgame

const UnitLimit = 1700

type ObjectLimits struct {
	Images      uint
	Sprites     uint
	LoneSprites uint
	Bullets     uint
	Orders      uint
}

var Legacy116ObjectLimits = ObjectLimits{
	Images:      5000,
	Sprites:     2500,
	LoneSprites: 500,
	Bullets:     100,
	Orders:      2000,
}

type Rules struct {
	Limits              ObjectLimits
	DisabledThingyUnits bool
	OrderAllocationRule bool
	FlyingSCVFix        bool
}

func DefaultRules() Rules {
	return Rules{Limits: Legacy116ObjectLimits}
}

func (r Rules) ObjectLimits() ObjectLimits {
	limits := r.Limits
	orDefault(&limits.Images, Legacy116ObjectLimits.Images)
	orDefault(&limits.Sprites, Legacy116ObjectLimits.Sprites)
	orDefault(&limits.LoneSprites, Legacy116ObjectLimits.LoneSprites)
	orDefault(&limits.Bullets, Legacy116ObjectLimits.Bullets)
	orDefault(&limits.Orders, Legacy116ObjectLimits.Orders)
	return limits
}

func (r Rules) OrderSupplyLowThreshold() uint {
	return r.ObjectLimits().Orders * 9 / 10
}

func orDefault(value *uint, fallback uint) {
	if *value == 0 {
		*value = fallback
	}
}

func (f *StateFunctions) canAllocateOrder(u *Unit, orderType *OrderTypeT) bool {
	rules := &f.gameSt.Rules
	if !rules.OrderAllocationRule {
		return true
	}
	if f.st.ActiveOrdersSize < rules.OrderSupplyLowThreshold() && u.OrderQueueCount < 8 {
		return true
	}
	if u.OrderType.ID == OrdersDie || u.OrderType.ID == OrdersResetHarvestCollision {
		return true
	}
	return rules.FlyingSCVFix && u.OrderQueueCount >= 8 && u.OrderType.ID == OrdersConstructingBuilding && orderType.ID == OrdersResetCollision
}
