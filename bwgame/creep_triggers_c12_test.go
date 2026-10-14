package bwgame

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

const c12TileWidth = 20
const c12TileHeight = 14

const c12ReferenceCpp = `#include "bwgame.h"
#include <cstdio>
using namespace bwgame;

static uint64_t rng_state = 0x9E3779B97F4A7C15ull;
static uint64_t next_rand() {
	rng_state ^= rng_state << 13;
	rng_state ^= rng_state >> 7;
	rng_state ^= rng_state << 17;
	return rng_state;
}

static uint64_t tile_flags_hash(state& st) {
	uint64_t h = 14695981039346656037ull;
	for (auto& t : st.tiles) {
		h ^= t.flags;
		h *= 1099511628211ull;
	}
	return h;
}

static void print_creep_life(state& st) {
	printf("life %u %d", st.lcg_rand_state, st.creep_life.recede_timer);
	for (size_t i = 0; i != 9; ++i) printf(" %zu", st.creep_life.lists_size[i]);
	printf(" free %zu hash %llu\n", st.creep_life.free_list_size, (unsigned long long)tile_flags_hash(st));
}

static void creep_cases(state& st, state_functions& f, game_state& game_st) {
	size_t w = game_st.map_tile_width;
	size_t h = game_st.map_tile_height;
	for (auto& t : st.tiles) {
		uint64_t r = next_rand();
		t.flags = 0;
		if (r % 2) t.flags |= tile_t::flag_has_creep;
		if ((r >> 8) % 8 == 0) t.flags |= tile_t::flag_unbuildable;
		if ((r >> 16) % 16 == 0) t.flags |= tile_t::flag_partially_walkable;
	}
	for (size_t y = 0; y != h; ++y) {
		for (size_t x = 0; x != w; ++x) {
			printf("tile %zu %zu %zu %d\n", x, y, f.count_neighboring_creep_tiles({x, y}), f.tile_can_have_creep({x, y}) ? 1 : 0);
		}
	}
	for (size_t y = 0; y != h; ++y) {
		for (size_t x = 0; x != w; ++x) {
			if (~st.tiles[y * w + x].flags & tile_t::flag_has_creep) continue;
			if ((x + y) % 3 != 0) continue;
			printf("receding %zu %zu %d\n", x, y, f.set_creep_receding({x, y}) ? 1 : 0);
		}
	}
	print_creep_life(st);
	st.lcg_rand_state = 42;
	for (int step = 0; step != 400; ++step) {
		uint64_t r = next_rand();
		if (r % 3 == 0) {
			st.creep_life.recede_timer = (int)((r >> 4) % 2);
			f.recede_creep();
		} else {
			size_t x = (r >> 8) % w;
			size_t y = (r >> 20) % h;
			f.set_tile_creep({x, y}, (r >> 32) % 2 == 0);
		}
		print_creep_life(st);
	}
	for (size_t i = 0; i != 9; ++i) {
		printf("list %zu", i);
		for (auto& v : st.creep_life.lists[i]) printf(" %zu,%zu", v.tile_pos.x, v.tile_pos.y);
		printf("\n");
	}
}

static void setup_players(state& st) {
	int controllers[12] = {
		player_t::controller_occupied, player_t::controller_computer_game, player_t::controller_neutral,
		player_t::controller_rescue_passive, player_t::controller_unused_rescue_active, player_t::controller_open,
		player_t::controller_occupied, player_t::controller_closed, player_t::controller_occupied,
		player_t::controller_neutral, player_t::controller_occupied, player_t::controller_neutral};
	for (int i = 0; i != 12; ++i) {
		st.players[i].controller = controllers[i];
		st.players[i].force = i % 5;
		st.players[i].victory_state = (i == 6) ? 2 : (i == 1 ? 3 : 0);
		st.players[i].initially_active = true;
		for (int j = 0; j != 12; ++j) st.alliances[i][j] = (int)((i * 7 + j * 3) % 3);
	}
	for (int i = 0; i != 12; ++i) {
		for (int u = 0; u != 228; ++u) {
			st.unit_counts[i][(UnitTypes)u] = (i * 31 + u * 7) % 11;
			st.completed_unit_counts[i][(UnitTypes)u] = (i * 13 + u * 5) % 7;
		}
		st.factory_counts[i] = i + 1;
		st.building_counts[i] = i * 2 + 1;
		st.non_building_counts[i] = i * 3 + 1;
		st.completed_factory_counts[i] = i;
		st.completed_building_counts[i] = i * 2;
		st.completed_non_building_counts[i] = i * 3;
	}
}

static void trigger_cases(state& st, state_functions& f) {
	setup_players(st);
	for (int n = 0; n != 12; ++n) {
		printf("player %d %d %d %d\n", n, f.player_won(n) ? 1 : 0, f.player_defeated(n) ? 1 : 0, f.player_slot_active(n) ? 1 : 0);
	}
	for (int owner = 0; owner != 12; ++owner) {
		for (int player = 0; player != 28; ++player) {
			printf("pred %d %d", owner, player);
			for (int n = 0; n != 12; ++n) printf(" %d", f.trigger_players_pred(owner, player, n) ? 1 : 0);
			printf(" opp %d", f.trigger_opponent_count(owner, player));
			int unit_ids[] = {0, 7, 106, 227, 229, 230, 231, 232};
			for (int uid : unit_ids) {
				printf(" %d/%d", f.trigger_command_count(st, owner, player, uid, true), f.trigger_command_count(st, owner, player, uid, false));
			}
			printf("\n");
		}
	}
	for (int num_n = 0; num_n != 4; ++num_n) {
		for (int count = 3; count != 8; ++count) {
			trigger::condition c{};
			c.num_n = num_n;
			c.count_n = 5;
			printf("cmp %d %d %d\n", num_n, count, f.trigger_count_comparison(c, count) ? 1 : 0);
		}
	}
	trigger t{};
	t.actions[0].type = 4;
	t.actions[0].time_n = 100;
	t.actions[1].type = 3;
	t.actions[2].type = 15;
	t.actions[2].flags = 2;
	t.actions[3].type = 1;
	t.actions[4].type = 2;
	t.actions[5].type = 0;
	for (int owner : {0, 2, 6}) {
		running_trigger rt;
		rt.t = &t;
		state_functions::execute_trigger_struct ets;
		for (int round = 0; round != 6; ++round) {
			if (round == 2) st.trigger_waiting[owner] = false;
			if (round == 4) rt.flags |= 0x10;
			f.execute_trigger(ets, owner, rt, t);
			printf("exec %d %d %d %zu %d %d %d %d\n", owner, round, rt.flags, rt.current_action_index, rt.actions[0].flags, st.trigger_waiting[owner] ? 1 : 0, st.trigger_wait_timers[owner], ets.victory_state[owner]);
		}
	}
}

int main() {
	global_state global_st;
	game_state game_st;
	game_st.map_tile_width = 20;
	game_st.map_tile_height = 14;
	game_st.map_width = 20 * 32;
	game_st.map_height = 14 * 32;
	state st;
	st.global = &global_st;
	st.game = &game_st;
	st.tiles.resize(game_st.map_tile_width * game_st.map_tile_height);
	state_functions f(st);
	creep_cases(st, f, game_st);
	trigger_cases(st, f);
	return 0;
}
`

type c12Rand struct {
	state uint64
}

func (r *c12Rand) next() uint64 {
	r.state ^= r.state << 13
	r.state ^= r.state >> 7
	r.state ^= r.state << 17
	return r.state
}

type c12Output struct {
	lines []string
}

func (o *c12Output) printf(format string, args ...any) {
	o.lines = append(o.lines, strings.TrimSuffix(fmt.Sprintf(format, args...), "\n"))
}

func c12Bool(v bool) int {
	if v {
		return 1
	}
	return 0
}

func c12TileFlagsHash(st *State) uint64 {
	h := uint64(14695981039346656037)
	for _, t := range st.Tiles {
		h ^= uint64(t.Flags)
		h *= 1099511628211
	}
	return h
}

func c12PrintCreepLife(o *c12Output, st *State) {
	life := &st.CreepLife
	line := fmt.Sprintf("life %d %d", st.LcgRandState, life.RecedeTimer)
	for i := range life.ListsSize {
		line += fmt.Sprintf(" %d", life.ListsSize[i])
	}
	line += fmt.Sprintf(" free %d hash %d", life.FreeListSize, c12TileFlagsHash(st))
	o.printf("%s", line)
}

func c12CreepCases(o *c12Output, rng *c12Rand, st *State, f *StateFunctions) {
	w := f.gameSt.MapTileWidth
	h := f.gameSt.MapTileHeight
	for i := range st.Tiles {
		r := rng.next()
		st.Tiles[i].Flags = 0
		if r%2 != 0 {
			st.Tiles[i].Flags |= TileFlagHasCreep
		}
		if (r>>8)%8 == 0 {
			st.Tiles[i].Flags |= TileFlagUnbuildable
		}
		if (r>>16)%16 == 0 {
			st.Tiles[i].Flags |= TileFlagPartiallyWalkable
		}
	}
	for y := uint(0); y != h; y++ {
		for x := uint(0); x != w; x++ {
			pos := XyT[uint]{X: x, Y: y}
			o.printf("tile %d %d %d %d", x, y, f.countNeighboringCreepTiles(pos), c12Bool(f.tileCanHaveCreep(pos)))
		}
	}
	for y := uint(0); y != h; y++ {
		for x := uint(0); x != w; x++ {
			if st.Tiles[y*w+x].Flags&TileFlagHasCreep == 0 {
				continue
			}
			if (x+y)%3 != 0 {
				continue
			}
			o.printf("receding %d %d %d", x, y, c12Bool(f.setCreepReceding(XyT[uint]{X: x, Y: y})))
		}
	}
	c12PrintCreepLife(o, st)
	st.LcgRandState = 42
	for step := 0; step != 400; step++ {
		r := rng.next()
		if r%3 == 0 {
			st.CreepLife.RecedeTimer = int32((r >> 4) % 2)
			f.recedeCreep()
		} else {
			x := uint((r >> 8) % uint64(w))
			y := uint((r >> 20) % uint64(h))
			f.setTileCreep(XyT[uint]{X: x, Y: y}, (r>>32)%2 == 0)
		}
		c12PrintCreepLife(o, st)
	}
	for i := range st.CreepLife.Lists {
		line := fmt.Sprintf("list %d", i)
		for v := range st.CreepLife.Lists[i].All() {
			line += fmt.Sprintf(" %d,%d", v.TilePos.X, v.TilePos.Y)
		}
		o.printf("%s", line)
	}
}

func c12SetupPlayers(st *State) {
	controllers := [12]int32{
		PlayerControllerOccupied, PlayerControllerComputerGame, PlayerControllerNeutral,
		PlayerControllerRescuePassive, PlayerControllerUnusedRescueActive, PlayerControllerOpen,
		PlayerControllerOccupied, PlayerControllerClosed, PlayerControllerOccupied,
		PlayerControllerNeutral, PlayerControllerOccupied, PlayerControllerNeutral,
	}
	for i := int32(0); i != 12; i++ {
		p := &st.Players[i]
		p.Controller = controllers[i]
		p.Force = i % 5
		switch i {
		case 6:
			p.VictoryState = 2
		case 1:
			p.VictoryState = 3
		default:
			p.VictoryState = 0
		}
		p.InitiallyActive = true
		for j := int32(0); j != 12; j++ {
			st.Alliances[i][j] = (i*7 + j*3) % 3
		}
	}
	for i := int32(0); i != 12; i++ {
		for u := int32(0); u != 228; u++ {
			st.UnitCounts[i][u] = (i*31 + u*7) % 11
			st.CompletedUnitCounts[i][u] = (i*13 + u*5) % 7
		}
		st.FactoryCounts[i] = i + 1
		st.BuildingCounts[i] = i*2 + 1
		st.NonBuildingCounts[i] = i*3 + 1
		st.CompletedFactoryCounts[i] = i
		st.CompletedBuildingCounts[i] = i * 2
		st.CompletedNonBuildingCounts[i] = i * 3
	}
}

func c12TriggerCases(o *c12Output, st *State, f *StateFunctions) {
	c12SetupPlayers(st)
	for n := int32(0); n != 12; n++ {
		o.printf("player %d %d %d %d", n, c12Bool(f.playerWon(n)), c12Bool(f.playerDefeated(n)), c12Bool(f.playerSlotActive(n)))
	}
	for owner := int32(0); owner != 12; owner++ {
		for player := int32(0); player != 28; player++ {
			line := fmt.Sprintf("pred %d %d", owner, player)
			for n := int32(0); n != 12; n++ {
				line += fmt.Sprintf(" %d", c12Bool(f.triggerPlayersPred(owner, player, n)))
			}
			line += fmt.Sprintf(" opp %d", f.triggerOpponentCount(owner, player))
			for _, uid := range []int32{0, 7, 106, 227, 229, 230, 231, 232} {
				line += fmt.Sprintf(" %d/%d", f.triggerCommandCount(&st.UnitCounters, owner, player, uid, true), f.triggerCommandCount(&st.UnitCounters, owner, player, uid, false))
			}
			o.printf("%s", line)
		}
	}
	for numN := int32(0); numN != 4; numN++ {
		for count := int32(3); count != 8; count++ {
			c := TriggerCondition{NumN: numN, CountN: 5}
			o.printf("cmp %d %d %d", numN, count, c12Bool(f.triggerCountComparison(&c, count)))
		}
	}
	var t Trigger
	t.Actions[0].Type = 4
	t.Actions[0].TimeN = 100
	t.Actions[1].Type = 3
	t.Actions[2].Type = 15
	t.Actions[2].Flags = 2
	t.Actions[3].Type = 1
	t.Actions[4].Type = 2
	t.Actions[5].Type = 0
	for _, owner := range []int32{0, 2, 6} {
		rt := RunningTrigger{T: &t}
		var ets ExecuteTriggerStruct
		for round := 0; round != 6; round++ {
			if round == 2 {
				st.TriggerWaiting[owner] = false
			}
			if round == 4 {
				rt.Flags |= 0x10
			}
			f.executeTrigger(&ets, owner, &rt, &t)
			o.printf("exec %d %d %d %d %d %d %d %d", owner, round, rt.Flags, rt.CurrentActionIndex, rt.Actions[0].Flags, c12Bool(st.TriggerWaiting[owner]), st.TriggerWaitTimers[owner], ets.VictoryState[owner])
		}
	}
}

func c12GoLines() []string {
	game := NewGameState()
	game.MapTileWidth = c12TileWidth
	game.MapTileHeight = c12TileHeight
	game.MapWidth = c12TileWidth * 32
	game.MapHeight = c12TileHeight * 32
	st := NewState(&GlobalState{}, game)
	st.Tiles = make([]Tile, c12TileWidth*c12TileHeight)
	f := NewStateFunctions(st)
	var o c12Output
	rng := &c12Rand{state: 0x9E3779B97F4A7C15}
	c12CreepCases(&o, rng, st, f)
	c12TriggerCases(&o, st, f)
	return o.lines
}

func c12CppLines(t *testing.T) []string {
	refDir := referenceSourceDirectory(t)
	compiler, err := exec.LookPath("clang++")
	if err != nil {
		t.Skip("clang++ not found")
	}
	workDir := t.TempDir()
	sourcePath := filepath.Join(workDir, "c12.cpp")
	binaryPath := filepath.Join(workDir, "c12")
	if err := os.WriteFile(sourcePath, []byte(c12ReferenceCpp), 0o644); err != nil {
		t.Fatal(err)
	}
	build := exec.Command(compiler, "-std=c++14", "-fwrapv", "-O1", "-w", "-I"+refDir, sourcePath, "-o", binaryPath)
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("compile C++ reference: %v\n%s", err, output)
	}
	output, err := exec.Command(binaryPath).Output()
	if err != nil {
		t.Fatalf("run C++ reference: %v", err)
	}
	var lines []string
	scanner := bufio.NewScanner(bytes.NewReader(output))
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines
}

func TestCreepAndTriggersMatchCppReference(t *testing.T) {
	expected := c12CppLines(t)
	actual := c12GoLines()
	if len(expected) != len(actual) {
		t.Errorf("line count: C++ %d, Go %d", len(expected), len(actual))
	}
	for i := 0; i < min(len(expected), len(actual)); i++ {
		if expected[i] != actual[i] {
			t.Fatalf("line %d differs\nC++: %s\nGo:  %s", i, expected[i], actual[i])
		}
	}
}
