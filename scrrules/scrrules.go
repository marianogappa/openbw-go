package scrrules

import (
	"github.com/marianogappa/openbw-go/bwgame"
	"github.com/marianogappa/openbw-go/downgrade"
)

// FromSettings gives the rules for settings: the object limits, disabled THG2 units, the order
// allocation rule and the flying SCV fix.
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
