package bwgame

import (
	"errors"
	"os"
	"strings"

	"github.com/marianogappa/openbw-go/internal/pkware"
	"golang.org/x/text/encoding/korean"
)

const (
	replayIdentifier     = 0x53526572
	replayGameInfoSize   = 633
	replaySegmentMaxSize = 8192
)

type crc32T struct {
	table [256]uint32
}

func newCrc32T() crc32T {
	var c crc32T
	for i := uint32(0); i != 256; i++ {
		v := i
		for b := 0; b != 8; b++ {
			if v&1 != 0 {
				v = (v >> 1) ^ 0xedb88320
			} else {
				v >>= 1
			}
		}
		c.table[i] = v
	}
	return c
}

func (c *crc32T) checksum(data []byte) uint32 {
	r := uint32(0xffffffff)
	for _, b := range data {
		r = (r >> 8) ^ c.table[(r^uint32(b))&0xff]
	}
	return r
}

type replayFileReader struct {
	crc32 crc32T
	r     *DataReader
}

func makeReplayFileReader(r *DataReader) *replayFileReader {
	return &replayFileReader{crc32: newCrc32T(), r: r}
}

func (rr *replayFileReader) getBytes(output []byte) {
	crc32Sum := rr.r.GetU32()
	segments := uint(rr.r.GetU32())

	outputSize := uint(len(output))
	outputPos := uint(0)
	for i := uint(0); i != segments; i++ {
		segmentInputSize := uint(rr.r.GetU32())

		segmentOutputSize := min(outputSize-outputPos, replaySegmentMaxSize)

		if segmentInputSize > segmentOutputSize {
			throwError("replay_file_reader: output buffer too small")
		}
		segmentOutput := output[outputPos : outputPos+segmentOutputSize]
		if segmentInputSize == segmentOutputSize {
			rr.r.GetBytes(segmentOutput)
		} else {
			decompress(rr.r.GetN(segmentInputSize), segmentOutput)
		}
		outputPos += segmentOutputSize
	}

	if outputPos != outputSize {
		throwError("replay_file_reader: read %d bytes, expected %d", outputPos, outputSize)
	}

	calculatedCrc32Sum := rr.crc32.checksum(output)
	if calculatedCrc32Sum != crc32Sum {
		throwError("replay_file_reader: crc32 mismatch: got %08x, expected %08x", calculatedCrc32Sum, crc32Sum)
	}
}

func (rr *replayFileReader) getU32() uint32 {
	var buf [4]byte
	rr.getBytes(buf[:])
	return valueAtLe[uint32](buf[:])
}

func decompress(input []byte, output []byte) {
	if err := pkware.Decompress(input, output); err != nil {
		throwError("decompress: %v", err)
	}
}

type ReplayState struct {
	ActionsDataBuffer []byte
	EndFrame          int32
	MapName           string
	PlayerName        [12]string
	GameType          int32
	SlotController    [12]int32
	SlotRace          [12]int32
	SlotForce         [12]int32
	AlliesEnabled     bool
	TeamsEnabled      bool
}

type ReplayFunctions struct {
	*ActionFunctions

	replaySt *ReplayState

	OnActionFail func(frame int32, owner int32, actionID int32)
	BeforeAction func(owner int32, actionID int32, payload []byte)
	AfterAction  func(owner int32, actionID int32, ok bool)
	SkipActions  bool
}

func NewReplayFunctions(st *State, actionSt *ActionState, replaySt *ReplayState) *ReplayFunctions {
	return &ReplayFunctions{ActionFunctions: NewActionFunctions(st, actionSt), replaySt: replaySt}
}

func (f *ReplayFunctions) State() *State {
	return f.st
}

func (f *ReplayFunctions) ActionState() *ActionState {
	return f.actionSt
}

func (f *ReplayFunctions) ReplayState() *ReplayState {
	return f.replaySt
}

func (f *ReplayFunctions) LoadReplayFile(filename string) (err error) {
	defer recoverEngineError(&err)
	f.loadReplayFile(filename, true, nil)
	return nil
}

func (f *ReplayFunctions) LoadReplayData(data []byte) (err error) {
	defer recoverEngineError(&err)
	f.loadReplayData(data, true, nil)
	return nil
}

func (f *ReplayFunctions) NextFrame() (err error) {
	defer recoverEngineError(&err)
	f.nextFrame()
	return nil
}

func (f *ReplayFunctions) IsDone() bool {
	return f.isDone()
}

func recoverEngineError(err *error) {
	recovered := recover()
	if recovered == nil {
		return
	}
	if asError, ok := recovered.(error); ok && isEngineError(asError) {
		*err = asError
		return
	}
	panic(recovered)
}

func isEngineError(err error) bool {
	var exception *Exception
	var loadingError *DataLoadingError
	var notPortedError *NotPortedError
	return errors.As(err, &exception) || errors.As(err, &loadingError) || errors.As(err, &notPortedError)
}

func (f *ReplayFunctions) loadReplayFile(filename string, initialProcessing bool, getMapData *[]byte) {
	data, err := os.ReadFile(filename)
	if err != nil {
		throwError("file_reader: failed to open %s for reading", filename)
	}
	fileR := NewDataReaderLe(data)
	f.loadReplay(makeReplayFileReader(&fileR), initialProcessing, getMapData)
}

func (f *ReplayFunctions) loadReplayData(data []byte, initialProcessing bool, getMapData *[]byte) {
	r := NewDataReaderLe(data)
	f.loadReplay(makeReplayFileReader(&r), initialProcessing, getMapData)
}

const gameTypeTopVsBottom = 15

type replayGameInfo struct {
	frameCount                uint32
	randomSeed                uint32
	gameType                  uint16
	mapName                   []byte
	victoryCondition          uint8
	resourceType              uint8
	createInitialUnits        uint8
	tournamentMode            uint8
	alliesEnabled             uint8
	teamsEnabled              uint8
	startingMinerals          uint32
	slotPlayerID              [12]int32
	slotController            [12]int32
	slotRace                  [12]int32
	slotForce                 [12]int32
	slotName                  [12][]byte
	playerColor               [8]uint32
	createMeleeUnitsForPlayer [8]uint8
}

func readReplayGameInfo(gir *DataReader) replayGameInfo {
	var gi replayGameInfo
	gir.GetU8()
	gi.frameCount = gir.GetU32()
	gir.GetU16()
	gir.GetU8()
	gi.randomSeed = gir.GetU32()
	gir.Skip(8)
	gir.GetU32()
	gir.Skip(24)
	gir.GetU32()
	gir.GetU16()
	gir.GetU16()
	gir.GetU8()
	gir.GetU8()
	gir.GetU8()
	gir.GetU8()
	gi.gameType = gir.GetU16()
	gir.GetU16()
	gir.GetU32()
	gir.GetU16()
	gir.GetU8()
	gir.GetU8()
	gir.Skip(25)
	gi.mapName = gir.GetN(32)
	gir.GetU16()
	gir.GetU16()
	gir.GetU16()
	gir.GetU16()
	gi.victoryCondition = gir.GetU8()
	gi.resourceType = gir.GetU8()
	gir.GetU8()
	gir.GetU8()
	gi.createInitialUnits = gir.GetU8()
	gir.GetU8()
	gir.GetU8()
	gi.alliesEnabled = gir.GetU8()
	gi.teamsEnabled = gir.GetU8()
	gir.GetU8()
	gi.tournamentMode = gir.GetU8()
	gir.GetU32()
	gi.startingMinerals = gir.GetU32()
	gir.GetU32()
	gir.GetU8()

	for i := range 12 {
		gir.GetU32()
		gi.slotPlayerID[i] = int32(gir.GetU32())
		gi.slotController[i] = int32(gir.GetU8())
		gi.slotRace[i] = int32(gir.GetU8())
		gi.slotForce[i] = int32(gir.GetU8())
		gi.slotName[i] = gir.GetN(25)
	}

	for i := range gi.playerColor {
		gi.playerColor[i] = gir.GetU32()
	}
	for i := range gi.createMeleeUnitsForPlayer {
		gi.createMeleeUnitsForPlayer[i] = gir.GetU8()
	}
	return gi
}

func replayArrStr(str []byte) string {
	var r strings.Builder
	for _, v := range str {
		if v == 0 {
			break
		}
		if v >= 21 {
			r.WriteByte(v)
		}
	}
	return r.String()
}

func (f *ReplayFunctions) loadReplay(r *replayFileReader, initialProcessing bool, getMapData *[]byte) {
	identifier := r.getU32()
	if identifier != replayIdentifier {
		throwError("load_replay: invalid identifier %#x", identifier)
	}

	gameInfoBuffer := make([]byte, replayGameInfoSize)
	r.getBytes(gameInfoBuffer)

	gir := NewDataReaderLe(gameInfoBuffer)
	gi := readReplayGameInfo(&gir)

	f.replaySt.MapName = replayArrStr(gi.mapName)
	var kn string
	if koreanLocaleToUtf8(f.replaySt.MapName, &kn) {
		f.replaySt.MapName = kn
	}

	for i := range 12 {
		f.replaySt.PlayerName[i] = replayArrStr(gi.slotName[i])
		f.replaySt.SlotController[i] = gi.slotController[i]
		f.replaySt.SlotRace[i] = gi.slotRace[i]
		f.replaySt.SlotForce[i] = gi.slotForce[i]
		f.actionSt.PlayerID[i] = gi.slotPlayerID[i]
	}

	f.replaySt.EndFrame = int32(gi.frameCount)
	f.replaySt.GameType = int32(gi.gameType)
	f.replaySt.AlliesEnabled = gi.alliesEnabled != 0
	f.replaySt.TeamsEnabled = gi.teamsEnabled != 0

	f.replaySt.ActionsDataBuffer = make([]byte, r.getU32())
	r.getBytes(f.replaySt.ActionsDataBuffer)

	mapBuffer := make([]byte, r.getU32())
	r.getBytes(mapBuffer)

	if getMapData != nil {
		*getMapData = append([]byte(nil), mapBuffer...)
	}

	gameLoadFuncs := NewGameLoadFunctions(f.st)
	gameLoadFuncs.Hooks = f.Hooks
	gameLoadFuncs.loadMapData(mapBuffer, func() {
		f.applyReplaySetup(gameLoadFuncs, &gi)
	}, initialProcessing)

	for i := range 8 {
		f.st.Players[i].Color = int32(gi.playerColor[i])
	}
}

func (f *ReplayFunctions) applyReplaySetup(gameLoadFuncs *GameLoadFunctions, gi *replayGameInfo) {
	setupInfo := &gameLoadFuncs.setupInfo
	setupInfo.VictoryCondition = int32(gi.victoryCondition)
	setupInfo.StartingUnits = int32(gi.createInitialUnits)
	setupInfo.TournamentMode = int32(gi.tournamentMode)
	setupInfo.ResourceType = int32(gi.resourceType)
	setupInfo.StartingMinerals = int32(gi.startingMinerals)
	setupInfo.TopVsBottom = gi.gameType == gameTypeTopVsBottom
	for i := range 12 {
		f.st.Players[i].Controller = gi.slotController[i]
		f.st.Players[i].Race = Race(gi.slotRace[i])
		f.st.Players[i].Force = gi.slotForce[i]
		if gi.victoryCondition == 0 && gi.tournamentMode == 0 {
			if i >= 8 {
				setupInfo.CreateMeleeUnitsForPlayer[i] = false
			} else {
				setupInfo.CreateMeleeUnitsForPlayer[i] = gi.createMeleeUnitsForPlayer[i] != 0
			}
		}
	}
	f.st.LcgRandState = gi.randomSeed
}

func koreanLocaleToUtf8(src string, dst *string) bool {
	decoded, err := korean.EUCKR.NewDecoder().String(src)
	if err != nil || strings.ContainsRune(decoded, '�') {
		return false
	}
	*dst = decoded
	return true
}

func (f *ReplayFunctions) nextFrame() {
	f.ExecuteFrameActions()
	f.ProcessFrame()
}

func (f *ReplayFunctions) ExecuteFrameActions() {
	if f.st.CurrentFrame == f.replaySt.EndFrame {
		throwError("replay: attempt to play past end")
	}
	switch {
	case f.SkipActions:
	case f.OnActionFail != nil || f.BeforeAction != nil || f.AfterAction != nil:
		f.executeActionsReportingFails(f.replaySt.ActionsDataBuffer)
	default:
		f.executeActions(f.replaySt.ActionsDataBuffer)
	}
}

func (f *ReplayFunctions) ProcessFrame() {
	f.StateFunctions.nextFrame()
}

func (f *ReplayFunctions) isDone() bool {
	return f.st.CurrentFrame == f.replaySt.EndFrame
}

func (f *ReplayFunctions) ownerOfPlayerID(playerID int32) int32 {
	for i, id := range f.actionSt.PlayerID {
		if id == playerID {
			return int32(i)
		}
	}
	return -1
}

func (f *ReplayFunctions) executeActionGroup(r *DataReader) {
	for r.Left() != 0 {
		owner := f.ownerOfPlayerID(int32(r.data[r.ptr]))
		actionID := int32(-1)
		if r.Left() > 1 {
			actionID = int32(r.data[r.ptr+1])
		}
		if f.BeforeAction != nil && owner >= 0 && actionID >= 0 {
			f.BeforeAction(owner, actionID, r.data[r.ptr+2:])
		}
		ok := f.readAction(r)
		if !ok && f.OnActionFail != nil {
			f.OnActionFail(f.st.CurrentFrame, owner, actionID)
		}
		if f.AfterAction != nil {
			f.AfterAction(owner, actionID, ok)
		}
	}
}

func (f *ReplayFunctions) executeActionsReportingFails(actionsData []byte) {
	if f.st.CurrentFrame != f.actionSt.NextActionFrame {
		return
	}
	for f.actionSt.ActionsDataPosition != uint(len(actionsData)) {
		r := NewDataReaderLe(actionsData[f.actionSt.ActionsDataPosition:])
		frame := r.GetI32()
		if frame != f.st.CurrentFrame {
			f.actionSt.NextActionFrame = frame
			return
		}
		actionsSize := uint(r.GetU8())
		groupReader := NewDataReaderLe(r.GetN(actionsSize))
		f.executeActionGroup(&groupReader)
		f.actionSt.ActionsDataPosition += r.Tell()
	}
}
