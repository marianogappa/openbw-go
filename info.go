package openbw

import (
	"runtime/debug"
	"sync"

	"github.com/icza/screp/repparser"
	"github.com/marianogappa/openbw-go/bwgame"
)

// OpenBWCommit is the OpenBW commit that the engine port matches.
const OpenBWCommit = "4b046d5"

const modulePath = "github.com/marianogappa/openbw-go"

// EngineInfo identifies the engine, the data and the rules of a game.
type EngineInfo struct {
	Version      string
	ScrepVersion string
	OpenBWCommit string
	DataHash     string
	// SourceFormat is the replay format that the downgrade read, for example "legacy".
	SourceFormat string
	Rules        RulesInfo
}

// RulesInfo is the engine rules a game runs with.
type RulesInfo struct {
	// SCR is true when the rules come from the replay's SC:R settings.
	SCR bool
	// Custom is true when Options.CustomRules gave the rules.
	Custom              bool
	Limits              ObjectLimits
	DisabledThingyUnits bool
	OrderAllocationRule bool
	FlyingSCVFix        bool
}

// ObjectLimits are the engine's object pool sizes.
type ObjectLimits struct {
	Images, Sprites, LoneSprites, Bullets, Orders uint
}

func rulesInfo(source rulesSource, rules bwgame.Rules) RulesInfo {
	limits := rules.ObjectLimits()
	return RulesInfo{
		SCR:                 source == rulesFromSCR,
		Custom:              source == rulesFromCustom,
		Limits:              ObjectLimits{Images: limits.Images, Sprites: limits.Sprites, LoneSprites: limits.LoneSprites, Bullets: limits.Bullets, Orders: limits.Orders},
		DisabledThingyUnits: rules.DisabledThingyUnits,
		OrderAllocationRule: rules.OrderAllocationRule,
		FlyingSCVFix:        rules.FlyingSCVFix,
	}
}

var moduleVersion = sync.OnceValue(func() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "devel"
	}
	if info.Main.Path == modulePath && info.Main.Version != "" && info.Main.Version != "(devel)" {
		return info.Main.Version
	}
	for _, dep := range info.Deps {
		if dep.Path == modulePath && dep.Version != "" {
			return dep.Version
		}
	}
	return "devel"
})

// Version is this module's version from the build info, or "devel".
func Version() string { return moduleVersion() }

// Info identifies the engine, the data and the rules of this game.
func (g *Game) Info() EngineInfo {
	return EngineInfo{
		Version:      Version(),
		ScrepVersion: repparser.Version,
		OpenBWCommit: OpenBWCommit,
		DataHash:     g.data.FilesHash(),
		SourceFormat: g.prepared.Report.SourceFormat,
		Rules:        g.rules,
	}
}
