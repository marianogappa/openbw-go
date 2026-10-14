package bwgame

import "testing"

func TestDirectionLiteralC04(t *testing.T) {
	cases := []struct {
		value uint8
		want  int8
	}{{128, -128}, {144, -112}, {16, 16}, {1, 1}, {0, 0}}
	for _, c := range cases {
		if got := directionLiteral(c.value).RawValue; got != c.want {
			t.Errorf("directionLiteral(%d) = %d, want %d", c.value, got, c.want)
		}
	}
}

func TestBurrowDirectionC04(t *testing.T) {
	want := map[int32]int8{11: 88, 12: 96, 13: 104}
	for dir, raw := range want {
		if got := directionLiteral(8).MulInt(dir).RawValue; got != raw {
			t.Errorf("8_dir * %d = %d, want %d", dir, got, raw)
		}
	}
}

func TestUnitIsPlacedAtC04(t *testing.T) {
	cases := []struct {
		nx, ny, psx, psy, px, py int32
		want                     bool
	}{
		{100, 100, 64, 64, 100, 100, true},
		{100, 100, 64, 64, 132, 100, true},
		{100, 100, 64, 64, 131, 100, true},
		{100, 100, 64, 64, 68, 100, false},
		{100, 100, 64, 64, 67, 100, false},
		{100, 100, 64, 64, 100, 132, true},
		{100, 100, 64, 64, 100, 131, true},
		{0, 0, 32, 32, -16, 0, false},
		{0, 0, 32, 32, 17, 0, false},
		{5, 7, 48, 32, -2147483647, 7, false},
	}
	for _, c := range cases {
		unitType := &UnitTypeT{PlacementSize: Xy{X: c.psx, Y: c.psy}}
		n := &Unit{UnitType: unitType}
		n.Sprite = &Sprite{Position: Xy{X: c.nx, Y: c.ny}}
		if got := unitIsPlacedAt(n, unitType, Xy{X: c.px, Y: c.py}); got != c.want {
			t.Errorf("unitIsPlacedAt(%+v) = %v, want %v", c, got, c.want)
		}
		if unitIsPlacedAt(n, &UnitTypeT{}, Xy{X: c.px, Y: c.py}) {
			t.Errorf("unitIsPlacedAt matched a different unit type for %+v", c)
		}
	}
}
