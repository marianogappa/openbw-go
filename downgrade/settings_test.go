package downgrade

import (
	"os"
	"testing"
)

const settingsSampleReplay = defaultReplayDir + "/replays/30-llIlIlIIIIIlII1/MM-0CF890C4-160B-11F1-A3DD-FA167B5461B6.rep"

var scrSampleLimits = ObjectLimits{Images: 10000, Sprites: 5000, LoneSprites: 1000, Units: 3400, Bullets: 400, Orders: 4000, FogSprites: 1000}

func readSampleReplay(t *testing.T) []byte {
	data, err := os.ReadFile(settingsSampleReplay)
	if err != nil {
		t.Skipf("sample replay not readable: %v", err)
	}
	return data
}

func TestReadSCRSettingsFromSCRReplay(t *testing.T) {
	settings, err := ReadSCRSettings(readSampleReplay(t))
	if err != nil {
		t.Fatal(err)
	}
	if settings.Limits != scrSampleLimits {
		t.Errorf("limits = %+v, want %+v", settings.Limits, scrSampleLimits)
	}
	if settings.BugFixes != [8]byte{4} {
		t.Errorf("bug fixes = %x, want 0400000000000000", settings.BugFixes)
	}
	if settings.GameConfig == nil || *settings.GameConfig != (GameConfig{Build: 13515, Unknown10: 5}) {
		t.Errorf("game config = %+v, want build 13515 and unknown10 5", settings.GameConfig)
	}
	if !settings.OrderAllocationRule() || !settings.FlyingSCVFix() {
		t.Error("order allocation rule and flying SCV fix must be on")
	}
	if got := settings.OrderSupplyLowThreshold(); got != 3600 {
		t.Errorf("order supply low threshold = %d, want 3600", got)
	}
}

func TestDowngradeReportsSCRSettings(t *testing.T) {
	_, report, err := Downgrade(readSampleReplay(t))
	if err != nil {
		t.Fatal(err)
	}
	if report.SCRSettings.Limits != scrSampleLimits {
		t.Errorf("report limits = %+v, want %+v", report.SCRSettings.Limits, scrSampleLimits)
	}
}

func TestReadSCRSettingsFromLegacyReplay(t *testing.T) {
	legacy, _, err := Downgrade(readSampleReplay(t))
	if err != nil {
		t.Fatal(err)
	}
	settings, err := ReadSCRSettings(legacy)
	if err != nil {
		t.Fatal(err)
	}
	if settings != LegacySettings() {
		t.Errorf("settings = %+v, want the 1.16 defaults", settings)
	}
	if settings.OrderAllocationRule() || settings.FlyingSCVFix() {
		t.Error("legacy replays must not use the order allocation rule")
	}
	if got := settings.OrderSupplyLowThreshold(); got != 1800 {
		t.Errorf("order supply low threshold = %d, want 1800", got)
	}
}

func TestParseLimitsKeepsDefaultsForMissingFields(t *testing.T) {
	limits := parseLimits([]byte{0x10, 0x27, 0, 0}, Legacy116Limits)
	want := Legacy116Limits
	want.Images = 10000
	if limits != want {
		t.Errorf("limits = %+v, want %+v", limits, want)
	}
}

func TestOrderAllocationRuleFromGameConfigAlone(t *testing.T) {
	cases := []struct {
		config GameConfig
		want   bool
	}{
		{GameConfig{Build: 8153, Unknown10: 5}, false},
		{GameConfig{Build: 8154, Unknown10: 5}, true},
		{GameConfig{Build: 8000, Unknown10: 4}, true},
	}
	for _, c := range cases {
		settings := LegacySettings()
		settings.GameConfig = &c.config
		if got := settings.OrderAllocationRule(); got != c.want {
			t.Errorf("config %+v: rule = %v, want %v", c.config, got, c.want)
		}
	}
}
