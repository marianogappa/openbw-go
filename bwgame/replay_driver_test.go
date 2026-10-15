package bwgame

import (
	"bytes"
	"encoding/binary"
	"errors"
	"hash/crc32"
	"math/rand"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/marianogappa/openbw-go/downgrade"
)

var (
	compressedLegacyReplayPath = testReplayPath("scmapanalyzer/replays/2486 LimperT qeywro78Z.rep")
	scrReplayDir               = testReplayPath("cwalggdl")
)

func testReplayPath(name string) string {
	dir := os.Getenv("OPENBW_REPLAYS")
	if dir == "" {
		dir = "../testdata/replays"
	}
	return filepath.Join(dir, name)
}

type replaySections struct {
	gameInfo []byte
	actions  []byte
	mapData  []byte
}

func readReplaySections(data []byte) (sections replaySections, err error) {
	defer recoverEngineError(&err)
	dr := NewDataReaderLe(data)
	r := makeReplayFileReader(&dr)
	if id := r.getU32(); id != replayIdentifier {
		throwError("invalid identifier %#x", id)
	}
	sections.gameInfo = make([]byte, replayGameInfoSize)
	r.getBytes(sections.gameInfo)
	sections.actions = make([]byte, r.getU32())
	r.getBytes(sections.actions)
	sections.mapData = make([]byte, r.getU32())
	r.getBytes(sections.mapData)
	return sections, nil
}

func appendTestSection(out, data []byte) []byte {
	segments := (len(data) + replaySegmentMaxSize - 1) / replaySegmentMaxSize
	out = binary.LittleEndian.AppendUint32(out, ^crc32.ChecksumIEEE(data))
	out = binary.LittleEndian.AppendUint32(out, uint32(segments))
	for start := 0; start < len(data); start += replaySegmentMaxSize {
		segment := data[start:min(start+replaySegmentMaxSize, len(data))]
		out = binary.LittleEndian.AppendUint32(out, uint32(len(segment)))
		out = append(out, segment...)
	}
	return out
}

func appendTestSizedSection(out, data []byte) []byte {
	out = appendTestSection(out, binary.LittleEndian.AppendUint32(nil, uint32(len(data))))
	return appendTestSection(out, data)
}

func encodeTestReplay(sections replaySections) []byte {
	out := appendTestSection(nil, binary.LittleEndian.AppendUint32(nil, replayIdentifier))
	out = appendTestSection(out, sections.gameInfo)
	out = appendTestSizedSection(out, sections.actions)
	return appendTestSizedSection(out, sections.mapData)
}

func randomBytes(seed int64, n int) []byte {
	b := make([]byte, n)
	rand.New(rand.NewSource(seed)).Read(b)
	return b
}

func readTestFile(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Skipf("replay not available: %v", err)
	}
	return data
}

func TestCrc32MatchesReplayH(t *testing.T) {
	c := newCrc32T()
	for _, data := range [][]byte{nil, []byte("openbw"), randomBytes(1, 10000)} {
		if got, want := c.checksum(data), ^crc32.ChecksumIEEE(data); got != want {
			t.Fatalf("crc32 %08x, want %08x", got, want)
		}
	}
}

func TestReplayFileReaderReadsMultiSegmentSections(t *testing.T) {
	want := replaySections{gameInfo: randomBytes(2, replayGameInfoSize), actions: randomBytes(3, 3*8192+17), mapData: randomBytes(4, 8192)}
	got, err := readReplaySections(encodeTestReplay(want))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got.gameInfo, want.gameInfo) || !bytes.Equal(got.actions, want.actions) || !bytes.Equal(got.mapData, want.mapData) {
		t.Fatal("sections differ")
	}
}

func TestReplayFileReaderReadsEmptySection(t *testing.T) {
	want := replaySections{gameInfo: randomBytes(5, replayGameInfoSize), mapData: randomBytes(6, 10)}
	got, err := readReplaySections(encodeTestReplay(want))
	if err != nil {
		t.Fatal(err)
	}
	if len(got.actions) != 0 || !bytes.Equal(got.mapData, want.mapData) {
		t.Fatal("sections differ")
	}
}

func TestReplayFileReaderChecksCrc32(t *testing.T) {
	data := encodeTestReplay(replaySections{gameInfo: randomBytes(7, replayGameInfoSize), actions: randomBytes(8, 100), mapData: randomBytes(9, 100)})
	data[len(data)-1] ^= 0xff
	_, err := readReplaySections(data)
	if err == nil || !strings.Contains(err.Error(), "crc32 mismatch") {
		t.Fatalf("error %v, want crc32 mismatch", err)
	}
}

func TestReplayFileReaderRejectsOversizedSegment(t *testing.T) {
	data := binary.LittleEndian.AppendUint32(nil, 0)
	data = binary.LittleEndian.AppendUint32(data, 1)
	data = binary.LittleEndian.AppendUint32(data, 5)
	data = append(data, 1, 2, 3, 4, 5)
	_, err := readReplaySections(data)
	if err == nil || !strings.Contains(err.Error(), "output buffer too small") {
		t.Fatalf("error %v, want output buffer too small", err)
	}
}

func TestDecompressRejectsUnsupportedType(t *testing.T) {
	err := func() (err error) {
		defer recoverEngineError(&err)
		decompress([]byte{1, 6, 0, 0}, make([]byte, 4))
		return nil
	}()
	if err == nil || !strings.Contains(err.Error(), "type 1 not supported") {
		t.Fatalf("error %v", err)
	}
}

func TestDecompressLiteralsAndCopy(t *testing.T) {
	input := []byte{0x00, 0x04, 0x82, 0x00, 0xcd, 0x01}
	output := make([]byte, 6)
	err := func() (err error) {
		defer recoverEngineError(&err)
		decompress(input, output)
		return nil
	}()
	if err != nil {
		t.Fatal(err)
	}
	if want := []byte{0x41, 0x40, 0x41, 0x40, 0x41, 0x40}; !bytes.Equal(output, want) {
		t.Fatalf("output %x, want %x", output, want)
	}
}

func TestDecompressCompressedLegacyReplay(t *testing.T) {
	data := readTestFile(t, compressedLegacyReplayPath)
	got, err := readReplaySections(data)
	if err != nil {
		t.Fatal(err)
	}
	reencoded, _, err := downgrade.Downgrade(data)
	if err != nil {
		t.Fatal(err)
	}
	want, err := downgrade.ReadUncompressedLegacy(reencoded)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got.gameInfo, want.GameInfo) {
		t.Error("game info differs from downgrade.ReadUncompressedLegacy")
	}
	if !bytes.Equal(got.actions, want.Actions) {
		t.Errorf("actions differ from downgrade.ReadUncompressedLegacy: %d vs %d bytes", len(got.actions), len(want.Actions))
	}
	if !bytes.Equal(got.mapData, want.MapData) {
		t.Errorf("map data differs from downgrade.ReadUncompressedLegacy: %d vs %d bytes", len(got.mapData), len(want.MapData))
	}
}

func firstScrReplay(t *testing.T) string {
	t.Helper()
	var found string
	filepath.WalkDir(scrReplayDir, func(path string, d os.DirEntry, err error) error {
		if err != nil || found != "" {
			return filepath.SkipDir
		}
		if !d.IsDir() && strings.HasSuffix(path, ".rep") {
			found = path
		}
		return nil
	})
	if found == "" {
		t.Skip("no SC:R replay found")
	}
	return found
}

func TestReplayFileReaderReadsDowngradedReplay(t *testing.T) {
	legacy, _, err := downgrade.Downgrade(readTestFile(t, firstScrReplay(t)))
	if err != nil {
		t.Fatal(err)
	}
	want, err := downgrade.ReadUncompressedLegacy(legacy)
	if err != nil {
		t.Fatal(err)
	}
	got, err := readReplaySections(legacy)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got.actions, want.Actions) || !bytes.Equal(got.mapData, want.MapData) {
		t.Fatal("sections differ")
	}
}

func loadReplayHeaderOnly(t *testing.T, data []byte) (*State, *ReplayFunctions) {
	t.Helper()
	st := NewState(testGlobalState(t), NewGameState())
	funcs := NewReplayFunctions(st, &ActionState{}, &ReplayState{})
	err := funcs.LoadReplayData(data)
	var notPorted *NotPortedError
	if err != nil && !errors.As(err, &notPorted) {
		t.Fatalf("load replay: %v", err)
	}
	return st, funcs
}

func TestLoadReplaySetsReplayStateFromHeader(t *testing.T) {
	data := readTestFile(t, compressedLegacyReplayPath)
	sections, err := readReplaySections(data)
	if err != nil {
		t.Fatal(err)
	}
	_, funcs := loadReplayHeaderOnly(t, data)
	replaySt := funcs.ReplayState()
	if want := int32(binary.LittleEndian.Uint32(sections.gameInfo[1:])); replaySt.EndFrame != want {
		t.Errorf("end frame %d, want %d", replaySt.EndFrame, want)
	}
	if !bytes.Equal(replaySt.ActionsDataBuffer, sections.actions) {
		t.Error("actions buffer differs")
	}
	for i := range 12 {
		slot := sections.gameInfo[161+36*i:]
		if want := int32(binary.LittleEndian.Uint32(slot[4:])); funcs.ActionState().PlayerID[i] != want {
			t.Errorf("slot %d player id %d, want %d", i, funcs.ActionState().PlayerID[i], want)
		}
	}
	names := funcs.ReplayState().PlayerName[:]
	for _, name := range []string{"Limper", "qeywro78"} {
		if !slices.ContainsFunc(names, func(n string) bool { return strings.Contains(n, name) }) {
			t.Errorf("player %q not in %q", name, names)
		}
	}
	if replaySt.MapName == "" {
		t.Error("empty map name")
	}
}

func TestReplayArrStrStopsAtZeroAndDropsControlBytes(t *testing.T) {
	if got := replayArrStr([]byte{'a', 3, 'b', 20, 21, 0, 'c'}); got != "ab\x15" {
		t.Fatalf("got %q", got)
	}
}

func TestKoreanLocaleToUtf8(t *testing.T) {
	var dst string
	if !koreanLocaleToUtf8("\xc5\xf5\xc8\xa5", &dst) || dst != "투혼" {
		t.Fatalf("got %q", dst)
	}
	if !koreanLocaleToUtf8("Fighting Spirit", &dst) || dst != "Fighting Spirit" {
		t.Fatalf("got %q", dst)
	}
}

func TestIsDoneAtEndFrame(t *testing.T) {
	st := NewState(nil, NewGameState())
	funcs := NewReplayFunctions(st, &ActionState{}, &ReplayState{EndFrame: 3})
	st.CurrentFrame = 3
	if !funcs.IsDone() {
		t.Fatal("not done at end frame")
	}
	err := funcs.NextFrame()
	if err == nil || err.Error() != "replay: attempt to play past end" {
		t.Fatalf("error %v", err)
	}
}
