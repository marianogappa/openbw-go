package downgrade

import "encoding/binary"

// ObjectLimits are the object pool sizes of a game (the SC:R LMTS section).
type ObjectLimits struct {
	Images      uint32
	Sprites     uint32
	LoneSprites uint32
	Units       uint32
	Bullets     uint32
	Orders      uint32
	FogSprites  uint32
}

// Legacy116Limits are the object pool sizes of 1.16.
var Legacy116Limits = ObjectLimits{
	Images:      5000,
	Sprites:     2500,
	LoneSprites: 500,
	Units:       1700,
	Bullets:     100,
	Orders:      2000,
	FogSprites:  500,
}

// GameConfig is the part of the SC:R GCFG section that selects engine rules.
type GameConfig struct {
	Build     uint32
	Unknown10 uint8
}

// SCRSettings are the SC:R sections that change engine rules: LMTS, BFIX and GCFG.
type SCRSettings struct {
	Limits   ObjectLimits
	BugFixes [8]byte
	// GameConfig is nil when the replay has no GCFG section.
	GameConfig *GameConfig
}

const (
	gameConfigBuildOffset     = 4
	gameConfigUnknown10Offset = 0x10
	orderAllocationFixBit     = 0x4
	legacyGameConfigUnknown10 = 5
	orderAllocationMinBuild   = 8153
	flyingSCVFixMinBuild      = 9713
)

// LegacySettings are the settings of a legacy replay: the 1.16 limits.
func LegacySettings() SCRSettings {
	return SCRSettings{Limits: Legacy116Limits}
}

// ReadSCRSettings reads the settings of a replay; a legacy replay gives LegacySettings.
func ReadSCRSettings(rep []byte) (SCRSettings, error) {
	src, err := readSourceSections(rep)
	if err != nil {
		return SCRSettings{}, err
	}
	return src.settings(), nil
}

func (s sourceSections) settings() SCRSettings {
	settings := LegacySettings()
	settings.Limits = parseLimits(s.limits, settings.Limits)
	copy(settings.BugFixes[:], s.bugFixes)
	settings.GameConfig = parseGameConfig(s.gameConfig)
	return settings
}

func parseLimits(data []byte, defaults ObjectLimits) ObjectLimits {
	limits := defaults
	fields := []*uint32{&limits.Images, &limits.Sprites, &limits.LoneSprites, &limits.Units, &limits.Bullets, &limits.Orders, &limits.FogSprites}
	for i, field := range fields {
		if offset := i * 4; offset+4 <= len(data) {
			*field = binary.LittleEndian.Uint32(data[offset:])
		}
	}
	return limits
}

func parseGameConfig(data []byte) *GameConfig {
	if len(data) <= gameConfigUnknown10Offset {
		return nil
	}
	return &GameConfig{
		Build:     binary.LittleEndian.Uint32(data[gameConfigBuildOffset:]),
		Unknown10: data[gameConfigUnknown10Offset],
	}
}

// OrderAllocationRule tells if the game uses the SC:R order allocation rule.
func (s SCRSettings) OrderAllocationRule() bool {
	if s.BugFixes[0]&orderAllocationFixBit != 0 {
		return true
	}
	if s.GameConfig == nil {
		return false
	}
	return s.GameConfig.Unknown10 != legacyGameConfigUnknown10 || s.GameConfig.Build > orderAllocationMinBuild
}

// FlyingSCVFix tells if the game has the SC:R fix for flying SCVs.
func (s SCRSettings) FlyingSCVFix() bool {
	return s.OrderAllocationRule() && s.GameConfig != nil && s.GameConfig.Build >= flyingSCVFixMinBuild
}

// OrderSupplyLowThreshold is 90% of the order limit.
func (s SCRSettings) OrderSupplyLowThreshold() uint32 {
	return s.Limits.Orders * 9 / 10
}
