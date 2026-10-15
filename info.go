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

// EngineInfo says which engine, data and rules produced the output, and how far the game ran.
type EngineInfo struct {
	Version         string
	ScrepVersion    string
	OpenBWCommit    string
	DataHash        string
	SourceFormat    string
	Rules           RulesInfo
	FramesSimulated int32
	EndFrame        int32
	ReachedEnd      bool
	Error           string `json:",omitempty"`
}

// RulesInfo is the engine rules a game ran with. Mode is "Auto", "OpenBW" or "Custom".
type RulesInfo struct {
	Mode                string
	Limits              ObjectLimits
	DisabledThingyUnits bool
	OrderAllocationRule bool
	FlyingSCVFix        bool
}

// ObjectLimits are the engine's object pool sizes.
type ObjectLimits struct {
	Images, Sprites, LoneSprites, Bullets, Orders uint
}

func rulesInfo(opts StreamOptions, rules bwgame.Rules) RulesInfo {
	mode := opts.Rules.String()
	if opts.CustomRules != nil {
		mode = "Custom"
	}
	limits := rules.ObjectLimits()
	return RulesInfo{
		Mode:                mode,
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

// Info says which engine, data and rules this game uses, and how far it ran.
func (g *Game) Info() EngineInfo {
	info := EngineInfo{
		Version:         Version(),
		ScrepVersion:    repparser.Version,
		OpenBWCommit:    OpenBWCommit,
		DataHash:        g.data.FilesHash(),
		SourceFormat:    g.prepared.Report.SourceFormat,
		Rules:           g.rules,
		FramesSimulated: g.Frame(),
		EndFrame:        g.EndFrame(),
		ReachedEnd:      g.Frame() == g.EndFrame(),
	}
	if g.err != nil {
		info.Error = g.err.Error()
	}
	return info
}
