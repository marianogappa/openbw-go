package bwgame

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/marianogappa/openbw-go/downgrade"
)

const defaultC13ReplayDir = "/Users/mariano.gappa/Code/go/src/github.com/marianogappa/cwalggdl"

type fnvHasher struct {
	value uint64
}

func newFnvHasher() *fnvHasher {
	return &fnvHasher{value: 0xcbf29ce484222325}
}

func (h *fnvHasher) add(v int32) {
	bits := uint32(v)
	for i := 0; i != 4; i++ {
		h.value ^= uint64((bits >> (8 * i)) & 0xff)
		h.value *= 0x100000001b3
	}
}

func (h *fnvHasher) hex() string {
	return fmt.Sprintf("%016x", h.value)
}

func regionIndexForHash(r *Region) int32 {
	if r == nil {
		return -1
	}
	return int32(r.Index)
}

func hashRegions(g *GameState) string {
	h := newFnvHasher()
	rs := &g.Regions
	h.add(int32(len(rs.Regions)))
	for i := range rs.Regions {
		r := &rs.Regions[i]
		for _, v := range []int32{
			int32(r.Flags), int32(r.Index),
			int32(r.TileCenter.X), int32(r.TileCenter.Y),
			int32(r.TileArea.From.X), int32(r.TileArea.From.Y), int32(r.TileArea.To.X), int32(r.TileArea.To.Y),
			r.Center.X.RawValue, r.Center.Y.RawValue,
			r.Area.From.X, r.Area.From.Y, r.Area.To.X, r.Area.To.Y,
			int32(r.TileCount), int32(r.GroupIndex),
		} {
			h.add(v)
		}
		h.add(int32(len(r.WalkableNeighbors)))
		for _, n := range r.WalkableNeighbors {
			h.add(regionIndexForHash(n))
		}
		h.add(int32(len(r.NonWalkableNeighbors)))
		for _, n := range r.NonWalkableNeighbors {
			h.add(regionIndexForHash(n))
		}
	}
	for y := uint(0); y != g.MapTileHeight; y++ {
		for x := uint(0); x != g.MapTileWidth; x++ {
			h.add(int32(rs.TileRegionIndex[y*256+x]))
		}
	}
	h.add(int32(len(rs.SplitRegions)))
	for _, s := range rs.SplitRegions {
		h.add(int32(s.Mask))
		h.add(regionIndexForHash(s.A))
		h.add(regionIndexForHash(s.B))
	}
	for _, contours := range rs.Contours {
		h.add(int32(len(contours)))
		for _, c := range contours {
			h.add(c.V[0])
			h.add(c.V[1])
			h.add(c.V[2])
			h.add(int32(c.Dir))
			h.add(int32(c.Flags))
		}
	}
	return h.hex()
}

func hashSightValues(g *GameState) string {
	h := newFnvHasher()
	for _, v := range g.SightValues {
		for _, x := range []int32{v.MaxWidth, v.MaxHeight, v.MinWidth, v.MinHeight, v.MinMaskSize, v.ExtMaskedCount, int32(len(v.Maskdat))} {
			h.add(x)
		}
		for _, m := range v.Maskdat {
			h.add(int32(m.Prev))
			h.add(int32(m.Prev2))
			h.add(m.RelativeTileIndex)
			h.add(m.X)
			h.add(m.Y)
		}
	}
	return h.hex()
}

func hashTiles(g *GameState) string {
	h := newFnvHasher()
	h.add(int32(len(g.MegaTileFlags)))
	for _, v := range g.MegaTileFlags {
		h.add(int32(v))
	}
	h.add(int32(len(g.GfxTiles)))
	for _, v := range g.GfxTiles {
		h.add(int32(v.RawValue))
	}
	return h.hex()
}

func countContourTies(g *GameState) int {
	ties := 0
	for _, contours := range g.Regions.Contours {
		for i := 1; i < len(contours); i++ {
			if contours[i].V[0] == contours[i-1].V[0] && contours[i].V[1] == contours[i-1].V[1] {
				ties++
			}
		}
	}
	return ties
}

func c13ReplayDir() string {
	if dir := os.Getenv("DOWNGRADE_REPLAY_DIR"); dir != "" {
		return dir
	}
	return defaultC13ReplayDir
}

func loadReplayChk(t *testing.T, relativePath string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(c13ReplayDir(), relativePath))
	if err != nil {
		t.Skipf("replay not found: %v", err)
	}
	legacy, _, err := downgrade.Downgrade(data)
	if err != nil {
		t.Fatal(err)
	}
	rep, err := downgrade.ReadUncompressedLegacy(legacy)
	if err != nil {
		t.Fatal(err)
	}
	return rep.MapData
}

func testMaskTiles(f *GameLoadFunctions, offsetX, offsetY, width, height uint, and, or int32) {
	g := f.gameSt
	for y := offsetY; y != offsetY+height; y++ {
		for x := offsetX; x != offsetX+width; x++ {
			t := &f.st.Tiles[x+y*g.MapTileWidth]
			t.Flags = uint16((int32(t.Flags) & and) | or)
		}
	}
}

func loadMapTerrainForTest(t *testing.T, chk []byte) *GameLoadFunctions {
	t.Helper()
	f := NewGameLoadFunctions(NewState(testGlobalState(t), NewGameState()))
	l := &mapDataLoader{f: f, data: chk}
	l.registerTagFuncs()
	l.readChunks(chunkRequirements(true, "VER ", "DIM ", "ERA "))

	g := f.gameSt
	f.generateSightValues()
	f.loadTileStuff()
	f.st.Tiles = make([]Tile, g.MapTileWidth*g.MapTileHeight)
	f.st.TilesMegaTileIndex = make([]uint16, len(f.st.Tiles))

	l.tagFuncs[newTag("MTXM")] = func(r DataReader) {
		l.loadMtxmTiles(r)
		cleared := ^int32(TileFlagWalkable | TileFlagHasCreep | TileFlagPartiallyWalkable)
		testMaskTiles(f, 0, g.MapTileHeight-2, 5, 1, cleared, TileFlagUnbuildable)
		testMaskTiles(f, g.MapTileWidth-5, g.MapTileHeight-2, 5, 1, cleared, TileFlagUnbuildable)
		testMaskTiles(f, 0, g.MapTileHeight-1, g.MapTileWidth, 1, cleared, TileFlagUnbuildable)
		f.regionsCreate()
	}
	l.readChunks(chunkRequirements(true, "MTXM"))
	return f
}

func TestRegionsCreateMatchesCpp(t *testing.T) {
	cases := []struct {
		replay  string
		regions string
		sight   string
		tiles   string
	}{
		{"replays/30-llIlIlIIIIIlII1/MM-0CF890C4-160B-11F1-A3DD-FA167B5461B6.rep", "471793aef21d022b", "ba5f4434be6af6f0", "888863ef3a4f4788"},
		{"replays-cwal-dl/30-MORE-BBBuuuUU[kS]/MM-47F3007A-33E2-11F1-A3DD-FA167B5461B6.rep", "f9c7993abcfe5c8f", "ba5f4434be6af6f0", "01a081db2b87dac3"},
		{"replays-cwal-dl/30-NEW-lillljilililili/MM-9A601962-140F-11F1-A3DD-FA167B5461B6.rep", "2e189ea7d047082e", "b854d72a98ca5962", "33fcfc1810ed9050"},
		{"replays-cwal-dl/30-MORE-LYX2008/MM-624B86DE-37F9-11F1-A3DD-FA167B5461B6.rep", "db562eaf1105ff79", "ba5f4434be6af6f0", "bdf6c95dcd57de7a"},
		{"replays-cwal-dl/30-NEW-Horang2[._.]/MM-826F51D0-E82A-11F0-A3DD-FA167B5461B6.rep", "96e312f93cfcafe1", "ba5f4434be6af6f0", "bd6d05b73c303f83"},
		{"replays-cwal-dl/30-MORE-Skins_/MM-8EC5E794-39A3-11F1-A3DD-FA167B5461B6.rep", "8bcfce8a3fb3b29c", "ba5f4434be6af6f0", "9d9caa5c0a996ad6"},
		{"replays-cwal-dl/30-NEW-58BoJi4485/MM-1762CCF0-0BD9-11F1-A3DD-FA167B5461B6.rep", "773188f24d8a872f", "ba5f4434be6af6f0", "171d35ac55e56175"},
	}
	for _, tc := range cases {
		t.Run(filepath.Base(tc.replay), func(t *testing.T) {
			f := loadMapTerrainForTest(t, loadReplayChk(t, tc.replay))
			if got := hashTiles(f.gameSt); got != tc.tiles {
				t.Errorf("tiles hash = %s, want %s", got, tc.tiles)
			}
			if got := hashSightValues(f.gameSt); got != tc.sight {
				t.Errorf("sight values hash = %s, want %s", got, tc.sight)
			}
			if got := hashRegions(f.gameSt); got != tc.regions {
				t.Errorf("regions hash = %s, want %s", got, tc.regions)
			}
			if ties := countContourTies(f.gameSt); ties != 0 {
				t.Errorf("%d contours share a sort key; std::sort order is unspecified for them", ties)
			}
		})
	}
}

func TestGetUnitTypeStrengthMatchesCpp(t *testing.T) {
	global := testGlobalState(t)
	units, err := LoadUnitsDat(global.UnitsDat)
	if err != nil {
		t.Fatal(err)
	}
	weapons, err := LoadWeaponsDat(global.WeaponsDat)
	if err != nil {
		t.Fatal(err)
	}
	f := &GameLoadFunctions{}
	h := newFnvHasher()
	n := 0
	for i := range units.Vec {
		for j := range weapons.Vec {
			if weapons.Vec[j].Cooldown == 0 {
				continue
			}
			h.add(f.getUnitTypeStrength(&units.Vec[i], &weapons.Vec[j]))
			n++
		}
	}
	if n != 29640 || h.hex() != "970d7004e89e7aca" {
		t.Errorf("strength count %d hash %s, want 29640 970d7004e89e7aca", n, h.hex())
	}
}

func TestSqrtXTimes758(t *testing.T) {
	cases := map[uint32]uint32{0: 0, 1: 7, 2: 10, 100: 75, 12345: 842, 1 << 20: 7761, 0xffffffff: 496763}
	for x, want := range cases {
		if got := sqrtXTimes758(x); got != want {
			t.Errorf("sqrtXTimes758(%d) = %d, want %d", x, got, want)
		}
	}
}
