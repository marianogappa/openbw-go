package bwgame

import (
	"errors"
	"os"
	"strings"

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

type bitReader struct {
	r     *DataReader
	data  uint64
	bitsN uint
}

func makeBitReader(r *DataReader) *bitReader {
	return &bitReader{r: r}
}

func (br *bitReader) next() {
	left := br.r.Left()
	if left >= 4 {
		br.data |= uint64(br.r.GetU32()) << br.bitsN
		br.bitsN += 32
		return
	}
	switch left {
	case 0, 1:
		br.data |= uint64(br.r.GetU8()) << br.bitsN
		br.bitsN += 8
	case 2:
		br.data |= uint64(br.r.GetU16()) << br.bitsN
		br.bitsN += 16
	case 3:
		br.data |= uint64(br.r.GetU16()) << br.bitsN
		br.bitsN += 16
		br.data |= uint64(br.r.GetU8()) << br.bitsN
		br.bitsN += 8
	}
}

func (br *bitReader) getBits(bits uint) uint32 {
	if br.bitsN < bits {
		br.next()
	}
	r := uint32(br.data & ((uint64(1) << bits) - 1))
	br.data >>= bits
	br.bitsN -= bits
	return r
}

func (br *bitReader) getU8() uint8 {
	return uint8(br.getBits(8))
}

func (br *bitReader) getBit() bool {
	return br.getBits(1) != 0
}

func pickByBit(br *bitReader, ifSet, ifClear uint) uint {
	if br.getBit() {
		return ifSet
	}
	return ifClear
}

var decompressLength6BitsTable = [64]uint{
	3: 22, 7: 23, 11: 24, 15: 25, 19: 26, 23: 27, 27: 28, 31: 29,
	35: 30, 39: 31, 43: 32, 47: 33, 51: 34, 55: 35, 59: 36, 63: 37,
}

func decompressGetLength6Bits(br *bitReader) uint {
	v := uint(br.getBits(6))
	switch v % 4 {
	case 3:
		return decompressLength6BitsTable[v]
	case 0:
		step := v / 4
		if step%2 == 0 {
			return 262 + step/2 + 8*uint(br.getBits(5))
		}
		return 134 + step/2 + 8*uint(br.getBits(4))
	case 1:
		k := v / 4
		return pickByBit(br, 54+k, 38+k)
	default:
		return 70 + v/4 + 16*uint(br.getBits(2))
	}
}

func decompressGetLength(br *bitReader) uint {
	switch br.getBits(2) {
	case 3:
		return 1
	case 0:
		switch br.getBits(2) {
		case 3:
			return 6
		case 0:
			return decompressGetLength6Bits(br)
		case 1:
			if br.getBit() {
				return 7
			}
			return pickByBit(br, 9, 8)
		default:
			v := uint(br.getBits(3))
			if v%2 == 1 {
				return 10 + v/2
			}
			return pickByBit(br, 18+v/2, 14+v/2)
		}
	case 1:
		return pickByBit(br, 0, 2)
	default:
		if br.getBit() {
			return 3
		}
		return pickByBit(br, 4, 5)
	}
}

var decompressDistance5BitsTable = [32]uint{
	1: 39, 2: 47, 3: 31, 5: 35, 6: 43, 7: 27, 9: 37, 10: 45, 11: 29, 13: 33, 14: 41, 15: 25,
	17: 38, 18: 46, 19: 30, 21: 34, 22: 42, 23: 26, 25: 36, 26: 44, 27: 28, 29: 32, 30: 40, 31: 24,
}

var decompressDistance5BitsPairTable = [8]uint{62, 54, 58, 50, 60, 52, 56, 48}

var decompressDistance4BitsTable = [16]uint{
	1: 14, 2: 18, 3: 10, 4: 20, 5: 12, 6: 16, 7: 8, 8: 21, 9: 13, 10: 17, 11: 9, 12: 19, 13: 11, 14: 15, 15: 7,
}

func decompressGetDistance(br *bitReader) uint {
	switch br.getBits(2) {
	case 3:
		return 0
	case 0:
		v := br.getBits(5)
		if v%4 == 0 {
			base := decompressDistance5BitsPairTable[v/4]
			return pickByBit(br, base, base+1)
		}
		return decompressDistance5BitsTable[v]
	case 1:
		switch br.getBits(2) {
		case 1:
			return 2
		case 3:
			return 1
		case 0:
			return pickByBit(br, 5, 6)
		default:
			return pickByBit(br, 3, 4)
		}
	default:
		v := br.getBits(4)
		if v == 0 {
			return pickByBit(br, 22, 23)
		}
		return decompressDistance4BitsTable[v]
	}
}

func decompress(input []byte, output []byte) {
	sourceR := NewDataReaderLe(input)
	br := makeBitReader(&sourceR)
	compressionType := br.getU8()
	distanceBits := uint(br.getU8())

	if distanceBits != 4 && distanceBits != 5 && distanceBits != 6 {
		throwError("decompress: invalid distance bits %d", distanceBits)
	}
	if compressionType != 0 {
		throwError("decompress: type %d not supported", compressionType)
	}

	outputSize := uint(len(output))
	outPos := uint(0)
	for outPos != outputSize {
		if !br.getBit() {
			output[outPos] = br.getU8()
			outPos++
			continue
		}
		length := 2 + decompressGetLength(br)
		if length == 519 {
			throwError("decompress: eof marker found too early")
		}
		var distance uint
		if length == 2 {
			distance = decompressGetDistance(br)<<2 | uint(br.getBits(2))
		} else {
			distance = decompressGetDistance(br)<<distanceBits | uint(br.getBits(distanceBits))
		}
		srcPos := outPos - 1 - distance
		if srcPos > outputSize {
			throwError("decompress: distance %d out of range at %d", distance, outPos)
		}
		if srcPos+length > outputSize {
			length = outputSize - srcPos
		}
		if outPos+length > outputSize {
			length = outputSize - outPos
		}
		for i := uint(0); i != length; i++ {
			output[outPos+i] = output[srcPos+i]
		}
		outPos += length
	}
}

type ReplayState struct {
	ActionsDataBuffer []byte
	EndFrame          int32
	MapName           string
	PlayerName        [12]string
	GameType          int32
}

type ReplayFunctions struct {
	*ActionFunctions

	replaySt *ReplayState

	OnActionFail func(frame int32, owner int32, actionID int32)
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

type replayGameInfo struct {
	frameCount                uint32
	randomSeed                uint32
	gameType                  uint16
	mapName                   []byte
	victoryCondition          uint8
	resourceType              uint8
	createInitialUnits        uint8
	tournamentMode            uint8
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
	gir.GetU8()
	gir.GetU8()
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
		f.actionSt.PlayerID[i] = gi.slotPlayerID[i]
	}

	f.replaySt.EndFrame = int32(gi.frameCount)
	f.replaySt.GameType = int32(gi.gameType)

	f.replaySt.ActionsDataBuffer = make([]byte, r.getU32())
	r.getBytes(f.replaySt.ActionsDataBuffer)

	mapBuffer := make([]byte, r.getU32())
	r.getBytes(mapBuffer)

	if getMapData != nil {
		*getMapData = append([]byte(nil), mapBuffer...)
	}

	gameLoadFuncs := NewGameLoadFunctions(f.st)
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
	if f.st.CurrentFrame == f.replaySt.EndFrame {
		throwError("replay: attempt to play past end")
	}
	if f.OnActionFail != nil {
		f.executeActionsReportingFails(f.replaySt.ActionsDataBuffer)
	} else {
		f.executeActions(f.replaySt.ActionsDataBuffer)
	}
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
		if !f.readAction(r) {
			f.OnActionFail(f.st.CurrentFrame, owner, actionID)
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
