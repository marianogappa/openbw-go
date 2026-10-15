// Package openbw runs the OpenBW StarCraft: Brood War engine on a replay and reports what the engine knows:
// unit lifecycles, kills, production, research, command outcomes and player stats.
//
// The base API is a stream: Open returns a Game, and Game.Events steps the engine as the caller pulls events.
// Run is built on the stream and records a Result, which marshals to JSON in the style of screp.
// Game.Engine gives the engine itself (package bwgame) for what the events leave out.
//
// The engine needs the Blizzard game data files; see LoadData.
package openbw
