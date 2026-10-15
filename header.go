package openbw

import (
	"github.com/marianogappa/openbw-go/bwgame"
	"github.com/marianogappa/openbw-go/downgrade"
)

// Header is the replay's game info as the engine read it at load.
type Header struct {
	MapName string
	// EndFrame is the replay's frame count.
	EndFrame int32
	GameType int32
	// Slots are the 12 player slots; a slot's index is its Owner.
	Slots [playerSlots]Slot
}

// Slot is one player slot of the game info, before the map load changes it.
type Slot struct {
	Owner int8
	// PlayerID is the player ID that the slot's commands carry; -1 for none.
	PlayerID int32
	Name     string
	// Controller is the game info controller: 0 inactive, 1 computer game, 2 occupied, ..., 6 open, 7 neutral.
	Controller int32
	// Race is the game info race: 0 Zerg, 1 Terran, 2 Protoss, 6 random.
	Race  uint8
	Force int32
}

func readHeader(funcs *bwgame.ReplayFunctions) Header {
	rs := funcs.ReplayState()
	h := Header{MapName: rs.MapName, EndFrame: rs.EndFrame, GameType: rs.GameType}
	for i := range h.Slots {
		h.Slots[i] = Slot{
			Owner:      int8(i),
			PlayerID:   funcs.ActionState().PlayerID[i],
			Name:       rs.PlayerName[i],
			Controller: rs.SlotController[i],
			Race:       uint8(rs.SlotRace[i]),
			Force:      rs.SlotForce[i],
		}
	}
	return h
}

// Header is the replay's game info as the engine read it.
func (g *Game) Header() Header { return g.header }

// SourceCommand is one command of the original replay's command stream, as the downgrade read it.
type SourceCommand struct {
	Frame    int32
	PlayerID byte
	// ID is the command ID in the original replay; 1.21 IDs are not translated.
	ID byte
	// Action is the engine action index of the command, or -1 if the downgrade dropped it.
	Action int32
	// Dropped is why the downgrade dropped the command (a downgrade.Reason constant), or "".
	Dropped string
	// EndsBlock tells that the downgrade could not read past this command: the rest of its frame block is missing.
	EndsBlock bool
}

// SourceCommands are the commands of the original replay's command stream, in stream order.
// CommandResult.Source indexes this slice. Each call returns a new slice.
func (g *Game) SourceCommands() []SourceCommand {
	commands := make([]SourceCommand, len(g.prepared.Commands))
	for i, c := range g.prepared.Commands {
		commands[i] = SourceCommand{
			Frame:     int32(c.Frame),
			PlayerID:  c.PlayerID,
			ID:        c.ID,
			Action:    c.Action,
			Dropped:   c.DropReason,
			EndsBlock: c.DropReason == downgrade.ReasonUnknownLength || c.DropReason == downgrade.ReasonTruncated,
		}
	}
	return commands
}
