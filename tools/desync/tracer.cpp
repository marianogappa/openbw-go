#include "bwgame.h"
#include "actions.h"
#include "replay.h"
#include "data_loading.h"

#include <algorithm>
#include <cctype>
#include <cstdint>
#include <cstdio>
#include <cstdlib>
#include <map>
#include <string>
#include <vector>
#include <array>

namespace {

struct options {
	std::string data_directory;
	std::string replay_path;
	int from_frame = 0;
	int to_frame = 1 << 30;
	bool units = true;
	bool visible_list = false;
	bool check = false;
	std::vector<size_t> watched;
	std::vector<std::pair<int, int>> starved_actions;
	std::vector<std::pair<int, int>> rng_shifts;
	std::vector<std::pair<int, int>> dropped_actions;
	std::vector<std::array<int, 3>> mineral_changes;
	std::vector<std::array<int, 3>> build_time_changes;
	int near_radius = 0;
};

[[noreturn]] void exit_with_usage() {
	std::fprintf(stderr, "usage: tracer [-data dir] [-from F] [-to F] [-no-units] [-visible-list] [-check] [-watch index]... [-starve frame slot]... [-rng frame steps]... [-drop frame slot]... [-minerals frame slot delta]... [-build-time frame index delta]... [-near radius] replay.rep\n");
	std::exit(1);
}

options parse_options(int argc, char** argv) {
	options opts;
	const char* from_env = std::getenv("OPENBW_DATA");
	if (from_env && *from_env) opts.data_directory = from_env;
	for (int i = 1; i < argc; ++i) {
		std::string arg = argv[i];
		if (arg == "-data" && i + 1 < argc) opts.data_directory = argv[++i];
		else if (arg == "-from" && i + 1 < argc) opts.from_frame = std::atoi(argv[++i]);
		else if (arg == "-to" && i + 1 < argc) opts.to_frame = std::atoi(argv[++i]);
		else if (arg == "-no-units") opts.units = false;
		else if (arg == "-visible-list") opts.visible_list = true;
		else if (arg == "-check") opts.check = true;
		else if (arg == "-near" && i + 1 < argc) opts.near_radius = std::atoi(argv[++i]);
		else if (arg == "-build-time" && i + 3 < argc) {
			int frame = std::atoi(argv[++i]);
			int index = std::atoi(argv[++i]);
			opts.build_time_changes.push_back({frame, index, std::atoi(argv[++i])});
		} else if (arg == "-minerals" && i + 3 < argc) {
			int frame = std::atoi(argv[++i]);
			int slot = std::atoi(argv[++i]);
			opts.mineral_changes.push_back({frame, slot, std::atoi(argv[++i])});
		} else if (arg == "-drop" && i + 2 < argc) {
			int frame = std::atoi(argv[++i]);
			opts.dropped_actions.push_back({frame, std::atoi(argv[++i])});
		} else if (arg == "-rng" && i + 2 < argc) {
			int frame = std::atoi(argv[++i]);
			opts.rng_shifts.push_back({frame, std::atoi(argv[++i])});
		}
		else if (arg == "-starve" && i + 2 < argc) {
			int frame = std::atoi(argv[++i]);
			opts.starved_actions.push_back({frame, std::atoi(argv[++i])});
		}
		else if (arg == "-watch" && i + 1 < argc) opts.watched.push_back((size_t)std::atoi(argv[++i]));
		else if (!arg.empty() && arg[0] != '-' && opts.replay_path.empty()) opts.replay_path = arg;
		else exit_with_usage();
	}
	if (opts.replay_path.empty() || opts.data_directory.empty()) exit_with_usage();
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

uint16_t read_u16(const uint8_t* p) {
	return (uint16_t)(p[0] | (p[1] << 8));
}

int16_t read_i16(const uint8_t* p) {
	return (int16_t)read_u16(p);
}

struct unit_snapshot {
	uint32_t generation;
	int type;
	int owner;
	int x;
	int y;
	int order;
};

using unit_table = std::map<size_t, unit_snapshot>;

const int id_select = 0x09;
const int id_shift_select = 0x0a;
const int id_deselect = 0x0b;
const int id_right_click = 0x14;
const int id_targeted_order = 0x15;
const int id_build = 0x0c;
const int id_train = 0x1f;
const int id_unit_morph = 0x23;
const int id_building_morph = 0x35;

struct tracer_functions : bwgame::replay_functions {
	using bwgame::replay_functions::replay_functions;

	options opts;
	unit_table previous_units;
	std::map<size_t, std::string> watched_states;
	std::vector<size_t> previous_visible;

	std::vector<size_t> visible_order() const {
		std::vector<size_t> r;
		for (auto& u : st.visible_units) r.push_back(u.index + 1);
		return r;
	}

	int position_in(const std::vector<size_t>& list, size_t index) const {
		auto i = std::find(list.begin(), list.end(), index);
		return i == list.end() ? -1 : (int)(i - list.begin());
	}

	void print_visible_list_changes() {
		std::vector<size_t> now = visible_order();
		if (in_window() && !opts.watched.empty()) {
			size_t anchor = opts.watched.front();
			int before = position_in(previous_visible, anchor);
			int after = position_in(now, anchor);
			for (size_t index : previous_visible) {
				if (position_in(now, index) < 0) print_visible_change("vis-", index, previous_visible, before);
			}
			for (size_t index : now) {
				if (position_in(previous_visible, index) < 0) print_visible_change("vis+", index, now, after);
			}
			if (before != after) std::printf("vispos %d idx=%zu pos=%d->%d\n", st.current_frame, anchor, before, after);
		}
		previous_visible = std::move(now);
	}

	void print_visible_change(const char* what, size_t index, const std::vector<size_t>& list, int anchor_position) const {
		const bwgame::unit_t* u = st.units_container.try_get(index - 1);
		int position = position_in(list, index);
		std::printf("%s %d idx=%zu type=%d owner=%d order=%d pos=%d %s\n", what, st.current_frame, index, u && u->unit_type ? (int)u->unit_type->id : -1,
			u ? u->owner : -1, u ? order_id(u) : -1, position, position < anchor_position ? "before" : "after");
	}

	std::string watched_state(const bwgame::unit_t* u) const {
		char buf[256];
		std::snprintf(buf, sizeof(buf), "gen=%u type=%d hp=%d cd=%d order=%d state=%d sorder=%d sstate=%d target=%d,%d target_unit=%zu move_target=%d,%d next_target=%d,%d carry=%d",
			u->unit_id_generation % 32, (int)u->unit_type->id, (int)(u->hp.raw_value >> 8), (int)u->ground_weapon_cooldown, order_id(u), u->order_state, u->secondary_order_type ? (int)u->secondary_order_type->id : -1, u->secondary_order_state,
			u->order_target.pos.x, u->order_target.pos.y, u->order_target.unit ? u->order_target.unit->index + 1 : 0,
			u->move_target.pos.x, u->move_target.pos.y, u->next_target_waypoint.x, u->next_target_waypoint.y,
			u->carrying_flags);
		std::string r = buf;
		if (unit_is_hatchery(u)) {
			std::snprintf(buf, sizeof(buf), " larva_timer=%d", u->building.larva_timer);
			r += buf;
		}
		return r;
	}

	void print_near_watched() {
		if (!opts.near_radius || opts.watched.empty() || !in_window() || st.current_frame % 8) return;
		const bwgame::unit_t* anchor = st.units_container.try_get(opts.watched.front() - 1);
		if (!anchor || !anchor->sprite) return;
		for (auto* list : {&st.visible_units, &st.hidden_units}) {
			for (auto& u : *list) {
				if (&u == anchor || u.owner == anchor->owner || u.owner == 11) continue;
				int dx = u.sprite->position.x - anchor->sprite->position.x;
				int dy = u.sprite->position.y - anchor->sprite->position.y;
				if (dx * dx + dy * dy > opts.near_radius * opts.near_radius) continue;
				std::printf("near %d idx=%zu type=%d owner=%d pos=%d,%d order=%d\n", st.current_frame, u.index + 1, (int)u.unit_type->id, u.owner, u.sprite->position.x, u.sprite->position.y, order_id(&u));
			}
		}
	}

	void print_watched() {
		print_near_watched();
		for (size_t index : opts.watched) {
			const bwgame::unit_t* u = st.units_container.try_get(index - 1);
			if (!u || !u->unit_type) continue;
			std::string state = watched_state(u);
			std::string& previous = watched_states[index];
			if (state != previous && in_window()) {
				std::printf("watch %d idx=%zu pos=%d,%d %s\n", st.current_frame, index, u->position.x, u->position.y, state.c_str());
			}
			previous = state;
		}
	}

	bool in_window() const {
		return st.current_frame >= opts.from_frame && st.current_frame <= opts.to_frame;
	}

	int order_id(const bwgame::unit_t* u) const {
		return u->order_type ? (int)u->order_type->id : -1;
	}

	void add_units(unit_table& table, const bwgame::intrusive_list<bwgame::unit_t, bwgame::default_link_f>& list) const {
		for (auto& u : list) {
			table[u.index] = {u.unit_id_generation % 32, (int)u.unit_type->id, u.owner, u.position.x, u.position.y, order_id(&u)};
		}
	}

	unit_table current_units() const {
		unit_table table;
		add_units(table, st.visible_units);
		add_units(table, st.hidden_units);
		add_units(table, st.map_revealer_units);
		return table;
	}

	void print_unit_changes() {
		unit_table now = current_units();
		if (in_window()) {
			for (auto& v : previous_units) {
				auto i = now.find(v.first);
				if (i == now.end() || i->second.generation != v.second.generation) {
					std::printf("del %d idx=%zu gen=%u type=%d owner=%d at=%d,%d\n", st.current_frame, v.first + 1, v.second.generation, v.second.type, v.second.owner, v.second.x, v.second.y);
				}
			}
			for (auto& v : now) {
				auto i = previous_units.find(v.first);
				if (i == previous_units.end() || i->second.generation != v.second.generation) {
					std::printf("new %d idx=%zu gen=%u type=%d owner=%d at=%d,%d\n", st.current_frame, v.first + 1, v.second.generation, v.second.type, v.second.owner, v.second.x, v.second.y);
				} else if (i->second.type != v.second.type) {
					std::printf("morph %d idx=%zu gen=%u type=%d->%d owner=%d\n", st.current_frame, v.first + 1, v.second.generation, i->second.type, v.second.type, v.second.owner);
				}
			}
		}
		previous_units = std::move(now);
	}

	std::string describe_tag(uint16_t raw_tag) const {
		bwgame::unit_id id(raw_tag);
		char buf[160];
		if (!id.index()) return "none";
		const bwgame::unit_t* slot = st.units_container.try_get(id.index() - 1);
		const bwgame::unit_t* u = get_unit(id);
		if (u) {
			std::snprintf(buf, sizeof(buf), "%d.%d=t%d/p%d/o%d", (int)id.index(), (int)id.generation(), (int)u->unit_type->id, u->owner, order_id(u));
		} else if (slot && slot->unit_type) {
			std::snprintf(buf, sizeof(buf), "%d.%d=MISSING(slot gen %u t%d/p%d)", (int)id.index(), (int)id.generation(), slot->unit_id_generation % 32, (int)slot->unit_type->id, slot->owner);
		} else {
			std::snprintf(buf, sizeof(buf), "%d.%d=MISSING", (int)id.index(), (int)id.generation());
		}
		return buf;
	}

	bool is_dropped(int owner) const {
		for (auto& v : opts.dropped_actions) {
			if (v.first == st.current_frame && v.second == owner) return true;
		}
		return false;
	}

	bool is_starved(int owner) const {
		for (auto& v : opts.starved_actions) {
			if (v.first == st.current_frame && v.second == owner) return true;
		}
		return false;
	}

	int first_inconsistent_frame = 0;
	int inconsistent_count = 0;

	bool tag_is_inconsistent(uint16_t raw_tag, int owner, bool multi) const {
		bwgame::unit_id id(raw_tag);
		if (!id.index()) return false;
		const bwgame::unit_t* u = get_unit(id);
		if (!u) return true;
		return multi && u->owner != owner;
	}

	void check_selection(int owner, int action_id, const uint8_t* p) {
		if (action_id != id_select && action_id != id_shift_select) return;
		int n = p[0];
		int bad = 0;
		for (int i = 0; i != n; ++i) {
			if (tag_is_inconsistent(read_u16(p + 1 + 2 * i), owner, n > 1)) ++bad;
		}
		if (!bad) return;
		++inconsistent_count;
		if (!first_inconsistent_frame) first_inconsistent_frame = st.current_frame;
		if (opts.check) std::printf("inconsistent %d p%d bad=%d/%d%s\n", st.current_frame, owner, bad, n, describe_payload(action_id, p, p + 1 + 2 * n, owner).c_str());
	}

	int owner_of_player_id(int player_id) const {
		auto i = std::find(action_st.player_id.begin(), action_st.player_id.end(), player_id);
		if (i == action_st.player_id.end()) return -1;
		return (int)(i - action_st.player_id.begin());
	}

	std::string describe_selection(int owner) const {
		std::string s;
		for (const bwgame::unit_t* u : selected_units(owner)) {
			char buf[64];
			std::snprintf(buf, sizeof(buf), " %zu=t%d/q%zu", u->index + 1, (int)u->unit_type->id, u->build_queue.size());
			s += buf;
		}
		return s;
	}

	std::string hex_payload(const uint8_t* p, const uint8_t* end) const {
		std::string s = " raw=";
		for (const uint8_t* i = p; i < end && i < p + 12; ++i) {
			char buf[4];
			std::snprintf(buf, sizeof(buf), "%02x", *i);
			s += buf;
		}
		return s;
	}

	std::string describe_economy(int owner) const {
		if (owner < 0) return "";
		char buf[96];
		std::snprintf(buf, sizeof(buf), " [m=%d g=%d s=%d/%d]", st.current_minerals[owner], st.current_gas[owner], (int)st.supply_used[owner][(int)st.players[owner].race].raw_value, (int)st.supply_available[owner][(int)st.players[owner].race].raw_value);
		return buf;
	}

	std::string describe_payload(int action_id, const uint8_t* p, const uint8_t* end, int owner) const {
		char buf[256];
		switch (action_id) {
		case id_select:
		case id_shift_select:
		case id_deselect: {
			std::string s;
			for (int i = 0; i != p[0]; ++i) s += " " + describe_tag(read_u16(p + 1 + 2 * i));
			return s;
		}
		case id_right_click:
			std::snprintf(buf, sizeof(buf), " xy=%d,%d target=%s cmdtype=%d queue=%d", read_i16(p), read_i16(p + 2), describe_tag(read_u16(p + 4)).c_str(), read_u16(p + 6), p[8]);
			return buf + describe_selection(owner);
		case id_targeted_order:
			std::snprintf(buf, sizeof(buf), " xy=%d,%d target=%s cmdtype=%d order=%d queue=%d", read_i16(p), read_i16(p + 2), describe_tag(read_u16(p + 4)).c_str(), read_u16(p + 6), p[8], p[9]);
			return buf + describe_selection(owner);
		case id_build:
			std::snprintf(buf, sizeof(buf), " order=%d tile=%d,%d type=%d", p[0], read_u16(p + 1), read_u16(p + 3), read_u16(p + 5));
			return buf + describe_selection(owner);
		case id_train:
		case id_unit_morph:
		case id_building_morph:
			std::snprintf(buf, sizeof(buf), " type=%d", read_u16(p));
			return buf + describe_selection(owner);
		}
		return hex_payload(p, end) + describe_selection(owner);
	}

	void execute_action_group(bwgame::data_loading::data_reader_le& r, const uint8_t* end) {
		while (r.ptr != end) {
			int owner = owner_of_player_id(r.ptr[0]);
			int action_id = r.ptr + 1 != end ? r.ptr[1] : -1;
			std::string description;
			if (in_window()) description = describe_economy(owner) + describe_payload(action_id, r.ptr + 2, end, owner);
			if (owner >= 0 && action_id >= 0) check_selection(owner, action_id, r.ptr + 2);
			bool starve = is_starved(owner);
			int minerals = owner >= 0 ? st.current_minerals[owner] : 0;
			int gas = owner >= 0 ? st.current_gas[owner] : 0;
			if (starve) st.current_minerals[owner] = st.current_gas[owner] = 0;
			bool drop = is_dropped(owner);
			auto saved_selection = action_st.selection[owner >= 0 ? owner : 0];
			if (drop) action_st.selection[owner].clear();
			bool ok = read_action(r);
			if (drop) action_st.selection[owner] = saved_selection;
			if (starve) {
				st.current_minerals[owner] += minerals;
				st.current_gas[owner] += gas;
			}
			if (in_window() && !opts.check) std::printf("act %d p%d 0x%02x %s%s\n", st.current_frame, owner, action_id, ok ? "ok" : "FAIL", description.c_str());
		}
	}

	void execute_actions(uint8_t* actions_data_begin, uint8_t* actions_data_end) {
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

	void shift_rng() {
		const uint32_t multiplier = 22695477u;
		const uint32_t inverse_multiplier = 690295837u;
		for (auto& v : opts.rng_shifts) {
			if (v.first != st.current_frame) continue;
			for (int i = 0; i < v.second; ++i) st.lcg_rand_state = st.lcg_rand_state * multiplier + 1;
			for (int i = 0; i > v.second; --i) st.lcg_rand_state = (st.lcg_rand_state - 1) * inverse_multiplier;
		}
	}

	void change_minerals() {
		for (auto& v : opts.mineral_changes) {
			if (v[0] == st.current_frame) st.current_minerals[v[1]] += v[2];
		}
	}

	void change_build_times() {
		for (auto& v : opts.build_time_changes) {
			if (v[0] != st.current_frame) continue;
			bwgame::unit_t* u = st.units_container.try_get(v[1] - 1);
			if (u) u->remaining_build_time += v[2];
		}
	}

	void next_frame() {
		shift_rng();
		change_minerals();
		change_build_times();
		if (st.current_frame == replay_st.end_frame) bwgame::error("replay: attempt to play past end");
		auto& buffer = replay_st.actions_data_buffer;
		execute_actions(buffer.data(), buffer.data() + buffer.size());
		bwgame::state_functions::next_frame();
		if (opts.units) print_unit_changes();
		print_watched();
		if (opts.visible_list) print_visible_list_changes();
	}
};

int current_frame_for_errors = 0;

int run(const options& opts) {
	bwgame::game_player player(directory_data_loader{opts.data_directory});
	bwgame::action_state action_st;
	bwgame::replay_state replay_st;
	tracer_functions funcs(player.st(), action_st, replay_st);
	funcs.opts = opts;
	funcs.load_replay_file(opts.replay_path.c_str());
	bwgame::state& st = player.st();
	funcs.previous_units = funcs.current_units();
	while (!funcs.is_done() && st.current_frame <= opts.to_frame) {
		funcs.next_frame();
		current_frame_for_errors = st.current_frame;
	}
	for (size_t i = 0; i != st.random_counts.size(); ++i) {
		if (st.random_counts[i]) std::printf("random_count %zu %d\n", i, st.random_counts[i]);
	}
	std::printf("first_inconsistent %d count=%d\n", funcs.first_inconsistent_frame, funcs.inconsistent_count);
	std::printf("end %d\n", st.current_frame);
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
		return 2;
	}
}
