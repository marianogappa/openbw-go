package downgrade

import "encoding/binary"

type ObjectLimits struct {
	Images      uint32
	Sprites     uint32
	LoneSprites uint32
	Units       uint32
	Bullets     uint32
	Orders      uint32
	FogSprites  uint32
}

var Legacy116Limits = ObjectLimits{
	Images:      5000,
	Sprites:     2500,
	LoneSprites: 500,
	Units:       1700,
	Bullets:     100,
	Orders:      2000,
	FogSprites:  500,
}

type GameConfig struct {
	Build     uint32
	Unknown10 uint8
}

type SCRSettings struct {
	Limits     ObjectLimits
	BugFixes   [8]byte
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

func LegacySettings() SCRSettings {
	return SCRSettings{Limits: Legacy116Limits}
}

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

func (s SCRSettings) OrderAllocationRule() bool {
	if s.BugFixes[0]&orderAllocationFixBit != 0 {
		return true
	}
	if s.GameConfig == nil {
		return false
	}
	return s.GameConfig.Unknown10 != legacyGameConfigUnknown10 || s.GameConfig.Build > orderAllocationMinBuild
}

func (s SCRSettings) FlyingSCVFix() bool {
	return s.OrderAllocationRule() && s.GameConfig != nil && s.GameConfig.Build >= flyingSCVFixMinBuild
}

func (s SCRSettings) OrderSupplyLowThreshold() uint32 {
	return s.Limits.Orders * 9 / 10
}
