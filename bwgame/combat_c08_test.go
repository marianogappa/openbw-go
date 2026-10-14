package bwgame

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

const combatCasesCpp = `#include "bwgame.h"
#include <cstdio>
using namespace bwgame;

static uint64_t rng_state = 0x2545F4914F6CDD1Dull;
static uint64_t next_rand() {
	rng_state ^= rng_state << 13;
	rng_state ^= rng_state >> 7;
	rng_state ^= rng_state << 17;
	return rng_state;
}

int main() {
	global_state gs;
	game_state game;
	state st;
	st.global = &gs;
	st.game = &game;
	state_functions f(st);

	for (int i = 0; i != (int)WeaponTypes::None + 1; ++i) {
		weapon_type_t w;
		w.id = (WeaponTypes)i;
		printf("is_spell %d = %d\n", i, f.is_spell(&w) ? 1 : 0);
	}

	for (int g = 0; g != 200; ++g) {
		size_t n = 2 + next_rand() % 10;
		game.regions.regions.clear();
		game.regions.regions.resize(n);
		printf("graph %zu\n", n);
		for (size_t i = 0; i != n; ++i) {
			auto& r = game.regions.regions[i];
			r.group_index = next_rand() % 3;
			size_t nw = next_rand() % 4;
			size_t nnw = next_rand() % 4;
			printf("region %zu %zu", i, r.group_index);
			printf(" %zu", nw);
			for (size_t j = 0; j != nw; ++j) {
				size_t k = next_rand() % n;
				r.walkable_neighbors.push_back(&game.regions.regions[k]);
				printf(" %zu", k);
			}
			printf(" %zu", nnw);
			for (size_t j = 0; j != nnw; ++j) {
				size_t k = next_rand() % n;
				r.non_walkable_neighbors.push_back(&game.regions.regions[k]);
				printf(" %zu", k);
			}
			printf("\n");
		}
		for (int q = 0; q != 20; ++q) {
			int owner = (int)(next_rand() % 12);
			size_t from = next_rand() % n;
			size_t to = next_rand() % n;
			bool r = f.unit_can_move_from_to_region(owner, &game.regions.regions[from], &game.regions.regions[to]);
			printf("query %d %zu %zu = %d\n", owner, from, to, r ? 1 : 0);
		}
	}
	return 0;
}
`

func runCombatCases(t *testing.T) []string {
	refDir := referenceSourceDirectory(t)
	compiler, err := exec.LookPath("clang++")
	if err != nil {
		t.Skip("clang++ not found")
	}
	workDir := t.TempDir()
	sourcePath := filepath.Join(workDir, "combat.cpp")
	binaryPath := filepath.Join(workDir, "combat")
	if err := os.WriteFile(sourcePath, []byte(combatCasesCpp), 0o644); err != nil {
		t.Fatal(err)
	}
	build := exec.Command(compiler, "-std=c++14", "-fwrapv", "-O1", "-w", "-I"+refDir, sourcePath, "-o", binaryPath)
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("compile C++ combat cases: %v\n%s", err, output)
	}
	output, err := exec.Command(binaryPath).Output()
	if err != nil {
		t.Fatalf("run C++ combat cases: %v", err)
	}
	var lines []string
	scanner := bufio.NewScanner(bytes.NewReader(output))
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines
}

type combatCaseReplayer struct {
	t    *testing.T
	f    *StateFunctions
	game *GameState
}

func newCombatCaseReplayer(t *testing.T) *combatCaseReplayer {
	game := NewGameState()
	st := NewState(&GlobalState{}, game)
	return &combatCaseReplayer{t: t, f: NewStateFunctions(st), game: game}
}

func (r *combatCaseReplayer) atoi(s string) int {
	v, err := strconv.Atoi(s)
	if err != nil {
		r.t.Fatalf("bad integer %q: %v", s, err)
	}
	return v
}

func (r *combatCaseReplayer) readGraph(fields []string) {
	r.game.Regions.Regions = make([]Region, r.atoi(fields[1]))
}

func (r *combatCaseReplayer) readRegion(fields []string) {
	values := make([]int, len(fields)-1)
	for i, s := range fields[1:] {
		values[i] = r.atoi(s)
	}
	regions := r.game.Regions.Regions
	region := &regions[values[0]]
	region.GroupIndex = uint(values[1])
	rest := values[2:]
	region.WalkableNeighbors, rest = r.neighbors(rest)
	region.NonWalkableNeighbors, _ = r.neighbors(rest)
}

func (r *combatCaseReplayer) neighbors(values []int) ([]*Region, []int) {
	count := values[0]
	var result []*Region
	for _, index := range values[1 : 1+count] {
		result = append(result, &r.game.Regions.Regions[index])
	}
	return result, values[1+count:]
}

func (r *combatCaseReplayer) isSpell(fields []string) string {
	w := WeaponTypeT{ID: WeaponTypes(r.atoi(fields[1]))}
	return fmt.Sprint(boolToInt(r.f.isSpell(&w)))
}

func (r *combatCaseReplayer) query(fields []string) string {
	regions := r.game.Regions.Regions
	owner := int32(r.atoi(fields[1]))
	from := &regions[r.atoi(fields[2])]
	to := &regions[r.atoi(fields[3])]
	return fmt.Sprint(boolToInt(r.f.unitCanMoveFromToRegion(owner, from, to)))
}

func TestCombatMatchesCppReference(t *testing.T) {
	r := newCombatCaseReplayer(t)
	checked := 0
	for _, line := range runCombatCases(t) {
		input, expected, hasResult := strings.Cut(line, " = ")
		fields := strings.Fields(input)
		switch fields[0] {
		case "graph":
			r.readGraph(fields)
		case "region":
			r.readRegion(fields)
		case "is_spell":
			checked += r.check(line, r.isSpell(fields), expected, hasResult)
		case "query":
			checked += r.check(line, r.query(fields), expected, hasResult)
		default:
			t.Fatalf("unknown case %q", line)
		}
	}
	if checked == 0 {
		t.Fatal("no cases checked")
	}
}

func (r *combatCaseReplayer) check(line, got, expected string, hasResult bool) int {
	if !hasResult {
		r.t.Fatalf("case without result: %q", line)
	}
	if got != expected {
		r.t.Errorf("%s: got %s", line, got)
	}
	return 1
}

func TestBounceDamageDivisor(t *testing.T) {
	bounce := &WeaponTypeT{BulletType: WeaponTypeTBulletTypeBounce}
	fly := &WeaponTypeT{BulletType: WeaponTypeTBulletTypeFly}
	cases := []struct {
		weapon           *WeaponTypeT
		remainingBounces int32
		want             int32
	}{
		{bounce, 3, 1},
		{bounce, 2, 1},
		{bounce, 1, 3},
		{bounce, 0, 9},
		{fly, 0, 1},
	}
	for _, c := range cases {
		b := &Bullet{WeaponType: c.weapon, RemainingBounces: c.remainingBounces}
		if got := bounceDamageDivisor(b); got != c.want {
			t.Errorf("bullet type %d, remaining bounces %d: got %d, want %d", c.weapon.BulletType, c.remainingBounces, got, c.want)
		}
	}
}
