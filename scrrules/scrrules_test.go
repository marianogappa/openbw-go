package scrrules

import (
	"testing"

	"github.com/marianogappa/openbw-go/bwgame"
	"github.com/marianogappa/openbw-go/downgrade"
)

func TestLegacySettingsKeepOpenBWLimits(t *testing.T) {
	rules := FromSettings(downgrade.LegacySettings())
	if rules.ObjectLimits() != bwgame.DefaultRules().ObjectLimits() {
		t.Fatalf("limits %+v", rules.ObjectLimits())
	}
	if rules.OrderAllocationRule || rules.FlyingSCVFix || !rules.DisabledThingyUnits {
		t.Fatalf("rules %+v", rules)
	}
	if rules.OrderSupplyLowThreshold() != 1800 {
		t.Fatalf("threshold %d", rules.OrderSupplyLowThreshold())
	}
}

func TestSCRSettings(t *testing.T) {
	settings := downgrade.SCRSettings{
		Limits:     downgrade.ObjectLimits{Images: 10000, Sprites: 5000, LoneSprites: 1000, Units: 3400, Bullets: 400, Orders: 4000, FogSprites: 1000},
		BugFixes:   [8]byte{4},
		GameConfig: &downgrade.GameConfig{Build: 13515, Unknown10: 5},
	}
	rules := FromSettings(settings)
	want := bwgame.ObjectLimits{Images: 10000, Sprites: 5000, LoneSprites: 1000, Bullets: 400, Orders: 4000}
	if rules.ObjectLimits() != want {
		t.Fatalf("limits %+v", rules.ObjectLimits())
	}
	if !rules.OrderAllocationRule || !rules.FlyingSCVFix || rules.OrderSupplyLowThreshold() != 3600 {
		t.Fatalf("rules %+v", rules)
	}
}
