package bwgame

import (
	"slices"
	"testing"
)

func newRevealSightTestFunctions() *StateFunctions {
	game := NewGameState()
	game.MapTileWidth = 8
	game.MapTileHeight = 6
	sv := &game.SightValues[3]
	sv.MinMaskSize = 2
	sv.ExtMaskedCount = 5
	xs := []int32{0, 1, -1, 2, 0, 7, -3}
	ys := []int32{0, 0, 1, -1, 2, 1, 0}
	for i := range xs {
		sv.Maskdat = append(sv.Maskdat, SightValuesMaskdatNode{
			Prev:              0,
			Prev2:             ^uint(0),
			X:                 xs[i],
			Y:                 ys[i],
			RelativeTileIndex: xs[i] + ys[i]*8,
		})
	}
	st := NewState(&GlobalState{}, game)
	st.Tiles = make([]Tile, 48)
	for i := range st.Tiles {
		st.Tiles[i] = Tile{Visible: 0xff, Explored: uint8(0xf0 | i)}
	}
	return NewStateFunctions(st)
}

func TestRevealSightAtInAirMatchesCpp(t *testing.T) {
	f := newRevealSightTestFunctions()
	f.revealSightAt(Xy{X: 32*1 + 5, Y: 32*2 + 7}, 3, 0x5, true)
	f.revealSightAt(Xy{X: 32*7 + 1, Y: 32*0 + 1}, 3, 0x30, true)
	want := []uint8{255, 240, 255, 241, 255, 242, 255, 243, 255, 244, 255, 245, 255, 246, 207, 199, 255, 248, 255, 249, 255, 250, 250, 250, 255, 252, 255, 253, 207, 206, 255, 255, 255, 240, 250, 240, 250, 242, 255, 243, 255, 244, 255, 245, 255, 246, 207, 199, 250, 248, 255, 249, 255, 250, 255, 251, 255, 252, 255, 253, 255, 254, 255, 255, 255, 240, 250, 240, 255, 242, 255, 243, 255, 244, 255, 245, 255, 246, 255, 247, 255, 248, 255, 249, 255, 250, 255, 251, 255, 252, 255, 253, 255, 254, 255, 255}
	var got []uint8
	for _, tile := range f.st.Tiles {
		got = append(got, tile.Visible, tile.Explored)
	}
	if !slices.Equal(got, want) {
		t.Fatalf("tiles = %v, want %v", got, want)
	}
}

func TestSpriteIsDoodadMatchesCpp(t *testing.T) {
	f := &StateFunctions{}
	want := []SpriteTypes{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32, 33, 34, 35, 36, 37, 38, 39, 40, 41, 42, 43, 44, 45, 46, 47, 48, 49, 50, 51, 52, 53, 54, 55, 56, 57, 58, 59, 60, 61, 62, 63, 64, 65, 66, 67, 68, 69, 70, 71, 72, 73, 74, 75, 76, 77, 78, 79, 80, 81, 82, 83, 84, 85, 86, 87, 88, 89, 90, 91, 92, 93, 94, 95, 96, 97, 98, 99, 100, 101, 102, 103, 104, 105, 106, 107, 108, 109, 110, 111, 112, 113, 114, 115, 116, 117, 118, 119, 120, 121, 122, 123, 124, 125, 126, 127, 128, 129, 386, 387, 388, 389, 390, 391, 392, 393, 394, 395, 396, 397, 398, 399, 400, 401, 402, 403, 404, 405, 406, 407, 408, 409, 410, 411, 412, 413, 414, 415, 416, 417, 418, 419, 420, 421, 422, 423, 424, 425, 426, 427, 428, 429, 430, 431, 432, 433, 434, 435, 436, 437, 438, 439, 440, 441, 442, 443, 444, 445, 446, 447, 448, 449, 450, 451, 452, 453, 454, 455, 456, 457, 458, 459, 460, 461, 462, 463, 464, 465, 466, 467, 468, 469, 470, 471, 472, 473, 474, 475, 476, 477, 478, 479, 480}
	var got []SpriteTypes
	for id := range SpriteTypesNone {
		if f.spriteIsDoodad(&SpriteTypeT{ID: id}) {
			got = append(got, id)
		}
	}
	if !slices.Equal(got, want) {
		t.Fatalf("doodads = %v, want %v", got, want)
	}
}
