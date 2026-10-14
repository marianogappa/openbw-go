#include "bwgame.h"
#include "actions.h"
#include "replay.h"
#include "data_loading.h"

#include <algorithm>
#include <cctype>
#include <cstdint>
#include <cstdio>
#include <cstdlib>
#include <cstring>
#include <map>
#include <string>
#include <vector>

namespace {

const char* default_data_directory = "/Users/mariano.gappa/Code/go/src/github.com/marianogappa/openbw-go/testdata/bwdata";

struct options {
	std::string data_directory;
	std::string replay_path;
	int every = 1;
	int dump_units_frame = -1;
	int max_frames = -1;
	bool no_actions = false;
};

[[noreturn]] void exit_with_usage() {
	std::fprintf(stderr, "usage: harness [-data dir] [-every N] [-dump-units F] [-max-frames M] [-no-actions] replay.rep\n");
	std::exit(1);
}

int parse_int_argument(const char* flag, const char* value) {
	char* end = nullptr;
	long parsed = std::strtol(value, &end, 10);
	if (end == value || *end != 0) {
		std::fprintf(stderr, "invalid value for %s: %s\n", flag, value);
		exit_with_usage();
	}
	return (int)parsed;
}

std::string initial_data_directory() {
	const char* from_env = std::getenv("OPENBW_DATA");
	if (from_env && *from_env) return from_env;
	return default_data_directory;
}

options parse_options(int argc, char** argv) {
	options opts;
	opts.data_directory = initial_data_directory();
	for (int i = 1; i < argc; ++i) {
		std::string arg = argv[i];
		bool has_value = i + 1 < argc;
		if (arg == "-data" && has_value) opts.data_directory = argv[++i];
		else if (arg == "-every" && has_value) opts.every = parse_int_argument("-every", argv[++i]);
		else if (arg == "-dump-units" && has_value) opts.dump_units_frame = parse_int_argument("-dump-units", argv[++i]);
		else if (arg == "-max-frames" && has_value) opts.max_frames = parse_int_argument("-max-frames", argv[++i]);
		else if (arg == "-no-actions") opts.no_actions = true;
		else if (!arg.empty() && arg[0] != '-' && opts.replay_path.empty()) opts.replay_path = arg;
		else exit_with_usage();
	}
	if (opts.replay_path.empty() || opts.every < 1) exit_with_usage();
	return opts;
}

std::string normalize_data_path(std::string path) {
	std::replace(path.begin(), path.end(), '\\', '/');
	std::transform(path.begin(), path.end(), path.begin(), [](unsigned char c) { return (char)std::tolower(c); });
	return path;
}

struct directory_data_loader {
	std::string directory;

	void operator()(bwgame::a_vector<uint8_t>& dst, bwgame::a_string filename) const {
		std::string full_path = directory + "/" + normalize_data_path(std::string(filename.c_str()));
		FILE* f = std::fopen(full_path.c_str(), "rb");
		if (!f) bwgame::error("data loader: %s: file not found", full_path.c_str());
		std::fseek(f, 0, SEEK_END);
		long size = std::ftell(f);
		std::fseek(f, 0, SEEK_SET);
		dst.resize((size_t)size);
		size_t read = size > 0 ? std::fread(dst.data(), 1, (size_t)size, f) : 0;
		std::fclose(f);
		if (read != (size_t)size) bwgame::error("data loader: %s: read error", full_path.c_str());
	}
};

struct state_hasher {
	uint64_t value = 0xcbf29ce484222325ull;

	void add_int32(int32_t v) {
		uint32_t bits = (uint32_t)v;
		for (int i = 0; i != 4; ++i) {
			value ^= (bits >> (8 * i)) & 0xff;
			value *= 0x100000001b3ull;
		}
	}
};

int order_id(const bwgame::unit_t* u) {
	return u->order_type ? (int)u->order_type->id : -1;
}

void hash_unit(state_hasher& h, const bwgame::unit_t* u) {
	h.add_int32((int32_t)u->index);
	h.add_int32((int32_t)u->unit_type->id);
	h.add_int32((int32_t)u->owner);
	h.add_int32((int32_t)u->exact_position.x.raw_value);
	h.add_int32((int32_t)u->exact_position.y.raw_value);
	h.add_int32((int32_t)u->hp.raw_value);
	h.add_int32((int32_t)u->shield_points.raw_value);
	h.add_int32((int32_t)u->energy.raw_value);
	h.add_int32((int32_t)order_id(u));
	h.add_int32((int32_t)u->status_flags);
}

template<typename list_T>
void hash_unit_list(state_hasher& h, const list_T& units) {
	int32_t count = 0;
	for (auto& u : units) {
		(void)u;
		++count;
	}
	h.add_int32(count);
	for (auto& u : units) hash_unit(h, &u);
}

void hash_players(state_hasher& h, const bwgame::state& st) {
	for (size_t i = 0; i != 12; ++i) {
		h.add_int32((int32_t)st.current_minerals[i]);
		h.add_int32((int32_t)st.current_gas[i]);
		for (size_t race = 0; race != 3; ++race) h.add_int32((int32_t)st.supply_used[i][race].raw_value);
		for (size_t race = 0; race != 3; ++race) h.add_int32((int32_t)st.supply_available[i][race].raw_value);
	}
}

uint64_t state_hash(const bwgame::state& st) {
	state_hasher h;
	h.add_int32((int32_t)st.current_frame);
	h.add_int32((int32_t)st.lcg_rand_state);
	hash_players(h, st);
	hash_unit_list(h, st.visible_units);
	hash_unit_list(h, st.hidden_units);
	hash_unit_list(h, st.map_revealer_units);
	return h.value;
}

struct harness_functions : bwgame::replay_functions {
	using bwgame::replay_functions::replay_functions;

	int action_fail_count = 0;
	bool no_actions = false;

	int owner_of_player_id(int player_id) const {
		auto i = std::find(action_st.player_id.begin(), action_st.player_id.end(), player_id);
		if (i == action_st.player_id.end()) return -1;
		return (int)(i - action_st.player_id.begin());
	}

	void report_action_fail(int owner, int action_id) {
		++action_fail_count;
		std::fprintf(stderr, "action_fail %d %d %d\n", st.current_frame, owner, action_id);
	}

	void execute_action_group(bwgame::data_loading::data_reader_le& r, const uint8_t* end) {
		while (r.ptr != end) {
			int owner = owner_of_player_id(r.ptr[0]);
			int action_id = r.ptr + 1 != end ? r.ptr[1] : -1;
			if (!read_action(r)) report_action_fail(owner, action_id);
		}
	}

	void execute_actions_reporting_fails(uint8_t* actions_data_begin, uint8_t* actions_data_end) {
		if (st.current_frame != action_st.next_action_frame) return;
		while (action_st.actions_data_position != size_t(actions_data_end - actions_data_begin)) {
			bwgame::data_loading::data_reader_le r(actions_data_begin + action_st.actions_data_position, actions_data_end);
			int frame = r.get<int32_t>();
			if (frame != st.current_frame) {
				action_st.next_action_frame = frame;
				return;
			}
			size_t actions_size = r.get<uint8_t>();
			const uint8_t* ptr = r.get_n(actions_size);
			const uint8_t* end = ptr + actions_size;
			bwgame::data_loading::data_reader_le group_reader(ptr, end);
			execute_action_group(group_reader, end);
			action_st.actions_data_position = end - actions_data_begin;
		}
	}

	void next_frame() {
		if (st.current_frame == replay_st.end_frame) bwgame::error("replay: attempt to play past end");
		auto& buffer = replay_st.actions_data_buffer;
		if (!no_actions) execute_actions_reporting_fails(buffer.data(), buffer.data() + buffer.size());
		bwgame::state_functions::next_frame();
	}
};

int current_frame_for_errors = 0;

void print_frame_hash(const bwgame::state& st) {
	std::printf("frame %d %016llx\n", st.current_frame, (unsigned long long)state_hash(st));
}

int image_count(const bwgame::unit_t* u) {
	if (!u->sprite) return -1;
	int count = 0;
	for (auto& image : u->sprite->images) {
		(void)image;
		++count;
	}
	return count;
}

void print_unit_row(const bwgame::unit_t* u) {
	bool completed = (u->status_flags & bwgame::unit_t::status_flag_completed) != 0;
	std::printf("%d\t%d\t%d\t%d\t%d\t%d\t%d\t%d\t%d\t%d\t%d\t%d\t%d\t%d\t%d\t%d\t%d\t%d\t%d\t%d\n",
		(int)u->index, (int)u->unit_type->id, u->owner, u->position.x, u->position.y,
		(int)(u->hp.raw_value >> 8), (int)(u->shield_points.raw_value >> 8), (int)(u->energy.raw_value >> 8),
		order_id(u), completed ? 1 : 0,
		(int)u->exact_position.x.raw_value, (int)u->exact_position.y.raw_value, (int)u->hp.raw_value,
		(int)u->status_flags, u->order_state, u->movement_state, (int)u->heading.raw_value,
		u->main_order_timer, (int)u->current_speed.raw_value, image_count(u));
}

void dump_units(const bwgame::state& st) {
	std::printf("rng state=%u total=%d", (unsigned)st.lcg_rand_state, st.total_random_counts);
	for (size_t i = 0; i != st.random_counts.size(); ++i) {
		if (st.random_counts[i]) std::printf(" source_%d=%d", (int)i, st.random_counts[i]);
	}
	std::printf("\n");
	std::printf("units_at_frame %d\n", st.current_frame);
	std::printf("index\ttype\towner\tx\ty\thp\tshields\tenergy\torder\tcompleted\texact_x\texact_y\thp_raw\tstatus\torder_state\tmovement_state\theading\torder_timer\tspeed\timages\n");
	for (auto& u : st.visible_units) print_unit_row(&u);
	for (auto& u : st.hidden_units) print_unit_row(&u);
	for (auto& u : st.map_revealer_units) print_unit_row(&u);
	std::printf("end_units\n");
}

bool is_playing_controller(int controller) {
	using p = bwgame::player_t;
	return controller == p::controller_occupied || controller == p::controller_computer_game ||
		controller == p::controller_computer || controller == p::controller_user_left ||
		controller == p::controller_computer_defeated;
}

std::string format_supply(const bwgame::fp1& supply) {
	int raw = (int)supply.raw_value;
	std::string r = std::to_string(raw / 2);
	if (raw % 2) r += ".5";
	return r;
}

const char* race_names[3] = {"zerg", "terran", "protoss"};

void print_player_summary(const bwgame::state& st, const bwgame::replay_state& replay_st, size_t i) {
	std::printf("player %d name=%s race=%d minerals=%d gas=%d", (int)i, replay_st.player_name[i].c_str(),
		(int)st.players[i].race, st.current_minerals[i], st.current_gas[i]);
	for (size_t race = 0; race != 3; ++race) {
		std::printf(" supply_%s=%s/%s", race_names[race], format_supply(st.supply_used[i][race]).c_str(),
			format_supply(st.supply_available[i][race]).c_str());
	}
	std::printf("\n");
}

void print_player_units(const bwgame::state& st, size_t i) {
	std::map<int, int> count_by_type;
	for (auto& u : st.player_units[i]) ++count_by_type[(int)u.unit_type->id];
	std::printf("units %d", (int)i);
	for (auto& v : count_by_type) std::printf(" %d=%d", v.first, v.second);
	std::printf("\n");
}

void print_summary(const bwgame::state& st, const bwgame::replay_state& replay_st, int action_fail_count) {
	for (size_t i = 0; i != 12; ++i) {
		if (!is_playing_controller(st.players[i].controller)) continue;
		print_player_summary(st, replay_st, i);
		print_player_units(st, i);
	}
	std::printf("action_fails %d\n", action_fail_count);
}

void after_frame(const bwgame::state& st, const options& opts) {
	current_frame_for_errors = st.current_frame;
	if (st.current_frame % opts.every == 0) print_frame_hash(st);
	if (st.current_frame == opts.dump_units_frame) dump_units(st);
}

bool should_continue(harness_functions& funcs, const options& opts) {
	if (funcs.is_done()) return false;
	return opts.max_frames < 0 || funcs.st.current_frame < opts.max_frames;
}

int run(const options& opts) {
	bwgame::game_player player(directory_data_loader{opts.data_directory});
	bwgame::action_state action_st;
	bwgame::replay_state replay_st;
	harness_functions funcs(player.st(), action_st, replay_st);
	funcs.no_actions = opts.no_actions;
	funcs.load_replay_file(opts.replay_path.c_str());
	bwgame::state& st = player.st();
	after_frame(st, opts);
	while (should_continue(funcs, opts)) {
		funcs.next_frame();
		after_frame(st, opts);
	}
	std::printf("end %d %016llx\n", st.current_frame, (unsigned long long)state_hash(st));
	print_summary(st, replay_st, funcs.action_fail_count);
	return 0;
}

}

int main(int argc, char** argv) {
	options opts = parse_options(argc, argv);
	try {
		return run(opts);
	} catch (const std::exception& e) {
		std::fflush(stdout);
		std::printf("error %d %s\n", current_frame_for_errors, e.what());
		std::fflush(stdout);
		return 2;
	}
}
