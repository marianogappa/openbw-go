package openbw

import "github.com/marianogappa/openbw-go/bwgame"

func chain1[A any](first, second func(A)) func(A) {
	if first == nil || second == nil {
		if first == nil {
			return second
		}
		return first
	}
	return func(a A) { first(a); second(a) }
}

func chain2[A, B any](first, second func(A, B)) func(A, B) {
	if first == nil || second == nil {
		if first == nil {
			return second
		}
		return first
	}
	return func(a A, b B) { first(a, b); second(a, b) }
}

func chain3[A, B, C any](first, second func(A, B, C)) func(A, B, C) {
	if first == nil || second == nil {
		if first == nil {
			return second
		}
		return first
	}
	return func(a A, b B, c C) { first(a, b, c); second(a, b, c) }
}

func chain4[A, B, C, D any](first, second func(A, B, C, D)) func(A, B, C, D) {
	if first == nil || second == nil {
		if first == nil {
			return second
		}
		return first
	}
	return func(a A, b B, c C, d D) { first(a, b, c, d); second(a, b, c, d) }
}

func chainHooks(own, user bwgame.Hooks) bwgame.Hooks {
	return bwgame.Hooks{
		PlaySound:          chain4(own.PlaySound, user.PlaySound),
		OnUnitDeselect:     chain1(own.OnUnitDeselect, user.OnUnitDeselect),
		OnUnitDestroy:      chain1(own.OnUnitDestroy, user.OnUnitDestroy),
		OnPlayerEliminated: chain1(own.OnPlayerEliminated, user.OnPlayerEliminated),
		OnVictoryState:     chain2(own.OnVictoryState, user.OnVictoryState),
		OnAction:           chain2(own.OnAction, user.OnAction),
		OnUnitCreated:      chain3(own.OnUnitCreated, user.OnUnitCreated),
		OnUnitCompleted:    chain1(own.OnUnitCompleted, user.OnUnitCompleted),
		OnUnitMorphed:      chain2(own.OnUnitMorphed, user.OnUnitMorphed),
		OnMorphStarted:     chain4(own.OnMorphStarted, user.OnMorphStarted),
		OnKillUnit:         chain2(own.OnKillUnit, user.OnKillUnit),
		OnOwnerChanged:     chain3(own.OnOwnerChanged, user.OnOwnerChanged),
		OnProductionCancel: chain1(own.OnProductionCancel, user.OnProductionCancel),
		OnResearch:         chain4(own.OnResearch, user.OnResearch),
		OnPlayerLeft:       chain2(own.OnPlayerLeft, user.OnPlayerLeft),
		OnActionRejected:   chain2(own.OnActionRejected, user.OnActionRejected),
		OnBuildQueue:       chain3(own.OnBuildQueue, user.OnBuildQueue),
		OnAddonBuildType:   chain2(own.OnAddonBuildType, user.OnAddonBuildType),
		OnUnitDamaged:      chain4(own.OnUnitDamaged, user.OnUnitDamaged),
	}
}
