package bwgame

import (
	"errors"
	"fmt"
	"hash/fnv"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestGetSpriteTileLineIndex(t *testing.T) {
	f := &StateFunctions{gameSt: &GameState{MapTileHeight: 64}}
	cases := []struct {
		y    int32
		want uint
	}{
		{-1, 0}, {-2147483648, 0}, {0, 0}, {31, 0}, {32, 1}, {2047, 63}, {2048, 63}, {2147483647, 63},
	}
	for _, c := range cases {
		if got := f.getSpriteTileLineIndex(c.y); got != c.want {
			t.Errorf("getSpriteTileLineIndex(%d) = %d, want %d", c.y, got, c.want)
		}
	}
}

func TestGetImageMapPosition(t *testing.T) {
	grp := &GrpT{Width: 64, Height: 48, Frames: []GrpTFrameT{{Offset: XyT[uint]{X: 10, Y: 5}, Size: XyT[uint]{X: 30, Y: 20}}}}
	sprite := &Sprite{Position: Xy{X: 100, Y: 200}}
	image := &Image{Sprite: sprite, Grp: grp, Offset: Xy{X: 3, Y: -4}}
	f := &StateFunctions{}

	if got, want := f.getImageMapPosition(image), (Xy{X: 81, Y: 177}); got != want {
		t.Errorf("unflipped = %v, want %v", got, want)
	}
	image.Flags = ImageFlagHorizontallyFlipped
	if got, want := f.getImageMapPosition(image), (Xy{X: 95, Y: 177}); got != want {
		t.Errorf("flipped = %v, want %v", got, want)
	}
	image.Flags = ImageFlagYFrozen
	image.FrozenYValue = 7
	if got, want := f.getImageMapPosition(image), (Xy{X: 81, Y: 7}); got != want {
		t.Errorf("y frozen = %v, want %v", got, want)
	}
}

func TestSpriteIscriptMatchesCpp(t *testing.T) {
	global := testGlobalState(t)
	want := readC09Hashes(t)
	only := os.Getenv("OPENBW_C09_SPRITE")
	ported, skipped := 0, 0
	for id := range global.SpriteTypes.Vec {
		if only != "" && only != strconv.Itoa(id) {
			continue
		}
		if want[id] == "crash" {
			continue
		}
		trace, notPortedName := traceSpriteIscript(global, id)
		if notPortedName != "" {
			skipped++
			continue
		}
		ported++
		if only != "" {
			fmt.Print(trace)
		}
		if got := fnvHex(trace); got != want[id] {
			t.Errorf("sprite type %d: trace hash %s, want %s", id, got, want[id])
		}
	}
	t.Logf("compared %d sprite types, skipped %d that reach unported code", ported, skipped)
}

func readC09Hashes(t *testing.T) map[int]string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", "c09_sprite_iscript_hashes.txt"))
	if err != nil {
		t.Fatal(err)
	}
	r := map[int]string{}
	for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
		fields := strings.Fields(line)
		id, err := strconv.Atoi(fields[0])
		if err != nil {
			t.Fatal(err)
		}
		r[id] = fields[1]
	}
	return r
}

func fnvHex(s string) string {
	h := fnv.New64a()
	h.Write([]byte(s))
	return fmt.Sprintf("%016x", h.Sum64())
}

func newC09TestState(global *GlobalState) (*State, *StateFunctions) {
	game := NewGameState()
	game.MapTileWidth = 64
	game.MapTileHeight = 64
	game.MapWidth = 64 * 32
	game.MapHeight = 64 * 32
	st := NewState(global, game)
	st.LcgRandState = 42
	st.SpritesOnTileLine = make([]SpriteList, 64)
	st.Tiles = make([]Tile, 64*64)
	return st, NewStateFunctions(st)
}

func traceSpriteIscript(global *GlobalState, id int) (trace string, notPortedName string) {
	st, f := newC09TestState(global)
	var b strings.Builder
	defer func() {
		r := recover()
		if r == nil {
			return
		}
		var np *NotPortedError
		if err, ok := r.(error); ok && errors.As(err, &np) {
			notPortedName = np.Name
			return
		}
		b.WriteString("error\n")
		trace = b.String()
	}()

	sprite := f.createSprite(&global.SpriteTypes.Vec[id], Xy{X: 300, Y: 400}, 1)
	alive := sprite != nil
	if !alive {
		b.WriteString("null\n")
	}
	writeC09Frame(&b, st, sprite, alive)
	for frame := 0; frame != 40 && alive; frame++ {
		if frame == 20 {
			f.spriteRunAnim(sprite, IscriptAnimsDeath)
		}
		alive = f.iscriptExecuteSprite(sprite)
		executeActiveThingies(st, f)
		writeC09Frame(&b, st, sprite, alive)
	}
	return b.String(), ""
}

func executeActiveThingies(st *State, f *StateFunctions) {
	for i := st.ActiveThingies.Begin(); i != st.ActiveThingies.End(); {
		th := i
		i = st.ActiveThingies.Next(i)
		if !f.iscriptExecuteSprite(th.Sprite) {
			st.ActiveThingiesSize--
			st.ActiveThingies.Remove(th)
			bwInsertList(&st.FreeThingies, th)
		}
	}
}

func writeC09Frame(b *strings.Builder, st *State, sprite *Sprite, alive bool) {
	fmt.Fprintf(b, "lcg %d thingies %d\n", st.LcgRandState, st.ActiveThingiesSize)
	if alive {
		writeC09Sprite(b, sprite)
	}
	for th := range st.ActiveThingies.All() {
		writeC09Sprite(b, th.Sprite)
	}
}

func writeC09Sprite(b *strings.Builder, sprite *Sprite) {
	mainIndex := int64(-1)
	if sprite.MainImage != nil {
		mainIndex = int64(sprite.MainImage.Index)
	}
	fmt.Fprintf(b, "sprite %d w%d h%d v%d f%d main%d\n", sprite.Index, sprite.Width, sprite.Height, sprite.VisibilityFlags, sprite.Flags, mainIndex)
	for i := range sprite.Images.All() {
		s := &i.IscriptState
		fmt.Fprintf(b, " img%d t%d fi%d fb%d fo%d fl%d ox%d oy%d m%d d%d e%d pc%d w%d a%d\n",
			i.Index, int(i.ImageType.ID), i.FrameIndex, i.FrameIndexBase, i.FrameIndexOffset, i.Flags,
			i.Offset.X, i.Offset.Y, i.Modifier, i.ModifierData1, i.ModifierData2, s.ProgramCounter, s.Wait, s.Animation)
	}
}
