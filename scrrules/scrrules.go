package scrrules

import (
	"github.com/marianogappa/openbw-go/bwgame"
	"github.com/marianogappa/openbw-go/downgrade"
)

func FromSettings(settings downgrade.SCRSettings) bwgame.Rules {
	return bwgame.Rules{
		Limits: bwgame.ObjectLimits{
			Images:      uint(settings.Limits.Images),
			Sprites:     uint(settings.Limits.Sprites),
			LoneSprites: uint(settings.Limits.LoneSprites),
			Bullets:     uint(settings.Limits.Bullets),
			Orders:      uint(settings.Limits.Orders),
		},
		DisabledThingyUnits: true,
		OrderAllocationRule: settings.OrderAllocationRule(),
		FlyingSCVFix:        settings.FlyingSCVFix(),
	}
}
