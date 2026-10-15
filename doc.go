// Package openbw runs the OpenBW StarCraft: Brood War engine on a replay and streams what the engine knows:
// unit lifecycles with killers, owner changes, production, research, player results and what the engine
// did with each command. Between events, the caller can read units, players and resources.
//
// Open returns a Game, and Game.Events steps the engine as the caller pulls events. Events are buffered
// for one engine half-step, so a read inside the loop shows the state at the end of the half-step that
// emitted the event, the same for every event of that half-step. The package reports raw engine facts
// only; aggregation and analysis belong to the caller.
//
// One Data serves any number of games in parallel; each Game belongs to one goroutine. For what the
// events leave out, EngineOptions give low-level access to the engine (package bwgame) on a private copy
// of the data.
//
// The engine needs the Blizzard game data files; see LoadData.
package openbw
