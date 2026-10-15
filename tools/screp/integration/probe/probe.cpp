#include "bwgame.h"
#include "actions.h"
#include "replay.h"
#include "data_loading.h"

#include <algorithm>
#include <array>
#include <cctype>
#include <cstdint>
#include <cstdio>
#include <cstdlib>
#include <map>
#include <set>
#include <string>
#include <tuple>

namespace {

const int scan_interval = 4;
const int timeline_interval = 1429;
const int player_slots = 8;

struct options {
	std::string data_directory;
	std::string replay_path;
};

[[noreturn]] void exit_with_usage() {
	std::fprintf(stderr, "usage: probe [-data dir] replay.rep\n");
	std::exit(1);
}

options parse_options(int argc, char** argv) {
	options opts;
	const char* from_env = std::getenv("OPENBW_DATA");
	if (from_env && *from_env) opts.data_directory = from_env;
	for (int i = 1; i < argc; ++i) {
		std::string arg = argv[i];
		if (arg == "-data" && i + 1 < argc) opts.data_directory = argv[++i];
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

using unit_key = std::tuple<size_t, uint32_t, int>;

struct production_tracker {
	std::set<unit_key> seen;
	std::array<std::map<int, int>, player_slots> made;

	void scan(const bwgame::state& st, bool count_new) {
		for (auto& u : st.visible_units) observe(&u, count_new);
		for (auto& u : st.hidden_units) observe(&u, count_new);
	}

	void observe(const bwgame::unit_t* u, bool count_new) {
		if (u->owner < 0 || u->owner >= player_slots) return;
		if (!(u->status_flags & bwgame::unit_t::status_flag_completed)) return;
		unit_key key{u->index, u->unit_id_generation, (int)u->unit_type->id};
		if (!seen.insert(key).second) return;
		if (count_new) ++made[u->owner][(int)u->unit_type->id];
	}
};

bool is_playing_controller(int controller) {
	using p = bwgame::player_t;
	return controller == p::controller_occupied || controller == p::controller_computer_game ||
		controller == p::controller_computer || controller == p::controller_user_left ||
		controller == p::controller_computer_defeated;
}

int player_unit_count(const bwgame::state& st, int slot) {
	int count = 0;
	for (auto& u : st.player_units[slot]) {
		(void)u;
		++count;
	}
	return count;
}

int supply_used_of_own_race(const bwgame::state& st, int slot) {
	int race = (int)st.players[slot].race;
	if (race < 0 || race > 2) return 0;
	return (int)st.supply_used[slot][race].raw_value / 2;
}

void print_timeline(const bwgame::state& st) {
	for (int slot = 0; slot != player_slots; ++slot) {
		if (!is_playing_controller(st.players[slot].controller)) continue;
		std::printf("timeline %d %d units=%d minerals=%d gas=%d supply=%d gathered_minerals=%d gathered_gas=%d\n",
			st.current_frame, slot, player_unit_count(st, slot), st.current_minerals[slot], st.current_gas[slot],
			supply_used_of_own_race(st, slot), st.total_minerals_gathered[slot], st.total_gas_gathered[slot]);
	}
}

struct elimination_tracker {
	std::array<int, player_slots> frame_without_units{};

	void update(const bwgame::state& st) {
		for (int slot = 0; slot != player_slots; ++slot) {
			if (frame_without_units[slot] || !is_playing_controller(st.players[slot].controller)) continue;
			if (player_unit_count(st, slot) == 0) frame_without_units[slot] = st.current_frame;
		}
	}
};

bool is_by_product(int type_id) {
	using namespace bwgame;
	static const std::set<int> by_products = {
		(int)UnitTypes::Terran_Vulture_Spider_Mine, (int)UnitTypes::Terran_Nuclear_Missile,
		(int)UnitTypes::Zerg_Larva, (int)UnitTypes::Zerg_Egg, (int)UnitTypes::Zerg_Cocoon,
		(int)UnitTypes::Zerg_Lurker_Egg, (int)UnitTypes::Protoss_Interceptor, (int)UnitTypes::Protoss_Scarab,
	};
	return by_products.count(type_id) != 0;
}

struct made_totals {
	int men = 0;
	int buildings = 0;
};

made_totals count_made(const bwgame::state& st, const std::map<int, int>& made) {
	made_totals totals;
	for (auto& v : made) {
		if (is_by_product(v.first)) continue;
		int flags = st.game->unit_types.vec[(size_t)v.first].group_flags;
		if (flags & bwgame::GroupFlags::Men) totals.men += v.second;
		else if (flags & bwgame::GroupFlags::Building) totals.buildings += v.second;
	}
	return totals;
}

void print_made(const bwgame::state& st, const production_tracker& production) {
	for (int slot = 0; slot != player_slots; ++slot) {
		if (!is_playing_controller(st.players[slot].controller)) continue;
		std::printf("made %d", slot);
		for (auto& v : production.made[slot]) std::printf(" %d=%d", v.first, v.second);
		std::printf("\n");
		made_totals totals = count_made(st, production.made[slot]);
		std::printf("made_totals %d men=%d buildings=%d\n", slot, totals.men, totals.buildings);
	}
}

void print_eliminations(const elimination_tracker& eliminations) {
	for (int slot = 0; slot != player_slots; ++slot) {
		if (eliminations.frame_without_units[slot]) std::printf("no_units %d %d\n", slot, eliminations.frame_without_units[slot]);
	}
}

const int action_id_build = 0x0c;
const int action_id_train = 0x1f;
const int action_id_unit_morph = 0x23;
const int action_id_building_morph = 0x35;

uint16_t read_u16(const uint8_t* p) {
	return (uint16_t)(p[0] | (p[1] << 8));
}

struct reject_key {
	int slot;
	int action_id;
	std::string reason;
	bool operator<(const reject_key& o) const {
		return std::tie(slot, action_id, reason) < std::tie(o.slot, o.action_id, o.reason);
	}
};

const int action_id_select = 0x09;
const int action_id_select_add = 0x0a;
const int action_id_select_remove = 0x0b;
const int action_id_right_click = 0x14;
const int action_id_targeted_order = 0x15;
const int target_tag_offset = 4;

struct tag_counts {
	int total = 0;
	int missing = 0;
};

struct tag_key {
	int slot;
	int minute;
	bool operator<(const tag_key& o) const {
		return std::tie(slot, minute) < std::tie(o.slot, o.minute);
	}
};

struct probe_functions : bwgame::replay_functions {
	using bwgame::replay_functions::replay_functions;

	std::map<reject_key, int> rejects;
	std::map<tag_key, tag_counts> selection_tags;
	std::map<tag_key, tag_counts> target_tags;

	bool tag_resolves(uint16_t raw_tag) const {
		return get_unit(bwgame::unit_id(raw_tag)) != nullptr;
	}

	void count_tag(std::map<tag_key, tag_counts>& counts, int owner, uint16_t raw_tag) {
		tag_counts& c = counts[{owner, st.current_frame / timeline_interval}];
		++c.total;
		if (tag_resolves(raw_tag)) return;
		++c.missing;
		report_missing_tag(owner, raw_tag);
	}

	void report_missing_tag(int owner, uint16_t raw_tag) const {
		bwgame::unit_id id(raw_tag);
		const bwgame::unit_t* slot_unit = id.index() ? get_unit(id.index() - 1) : nullptr;
		std::fprintf(stderr, "missing_tag frame=%d owner=%d index=%d generation=%d", st.current_frame, owner, (int)id.index(), (int)id.generation());
		if (slot_unit) {
			std::fprintf(stderr, " slot_type=%d slot_owner=%d slot_generation=%d", (int)slot_unit->unit_type->id, slot_unit->owner, (int)(slot_unit->unit_id_generation % 32));
		}
		std::fprintf(stderr, "\n");
	}

	bool is_selection_action(int action_id) const {
		return action_id == action_id_select || action_id == action_id_select_add || action_id == action_id_select_remove;
	}

	void check_tags(int owner, int action_id, const uint8_t* payload) {
		if (is_selection_action(action_id)) {
			int n = payload[0];
			for (int i = 0; i != n; ++i) count_tag(selection_tags, owner, read_u16(payload + 1 + 2 * i));
		} else if (action_id == action_id_right_click || action_id == action_id_targeted_order) {
			uint16_t raw_tag = read_u16(payload + target_tag_offset);
			if (raw_tag) count_tag(target_tags, owner, raw_tag);
		}
	}

	int owner_of_player_id(int player_id) const {
		auto i = std::find(action_st.player_id.begin(), action_st.player_id.end(), player_id);
		if (i == action_st.player_id.end()) return -1;
		return (int)(i - action_st.player_id.begin());
	}

	const bwgame::unit_type_t* unit_type_or_null(uint16_t id) const {
		if (id >= (uint16_t)bwgame::UnitTypes::None) return nullptr;
		return get_unit_type((bwgame::UnitTypes)id);
	}

	std::string economy_reason(const bwgame::unit_t* u, const bwgame::unit_type_t* unit_type) const {
		if (u->build_queue.size() >= 5) return "queue";
		if (u->build_queue.empty() && !has_available_supply_for(u->owner, unit_type)) return "supply";
		if (!has_available_resources_for(u->owner, unit_type)) return "resources";
		return "other";
	}

	std::string train_reject_reason(int owner, const bwgame::unit_type_t* unit_type) const {
		const bwgame::unit_t* u = get_single_selected_unit(owner);
		if (!u || u->owner != owner) return "no_unit";
		if (!unit_type || !unit_can_build(u, unit_type)) return "cannot_build";
		return economy_reason(u, unit_type);
	}

	std::string build_reject_reason(int owner, const bwgame::unit_type_t* unit_type) const {
		const bwgame::unit_t* u = get_single_selected_unit(owner);
		if (!u || u->owner != owner) return "no_unit";
		return "cannot_build";
	}

	bool can_start_unit_morph(const bwgame::unit_t* u, int owner, const bwgame::unit_type_t* unit_type) const {
		using bwgame::UnitTypes;
		if (u->owner != owner) return false;
		if (unit_is(u, UnitTypes::Zerg_Hydralisk) && unit_is(unit_type, UnitTypes::Zerg_Lurker)) {
			return unit_can_use_tech(u, get_tech_type(bwgame::TechTypes::Lurker_Aspect));
		}
		if (unit_is(u, UnitTypes::Zerg_Larva) || unit_is(u, UnitTypes::Zerg_Mutalisk)) return unit_can_build(u, unit_type);
		return false;
	}

	bool can_start_building_morph(const bwgame::unit_t* u, int owner, const bwgame::unit_type_t* unit_type) const {
		return u->owner == owner && unit_is_zerg_building(u) && unit_can_build(u, unit_type);
	}

	bool is_morphing(const bwgame::unit_t* u) const {
		using bwgame::UnitTypes;
		if (unit_is(u, UnitTypes::Zerg_Egg) || unit_is(u, UnitTypes::Zerg_Cocoon) || unit_is(u, UnitTypes::Zerg_Lurker_Egg)) return true;
		auto order = u->order_type->id;
		return order == bwgame::Orders::ZergUnitMorph || order == bwgame::Orders::ZergBuildingMorph;
	}

	template<typename can_start_T>
	std::string morph_reject_reason(int owner, const bwgame::unit_type_t* unit_type, can_start_T can_start) const {
		if (!unit_type) return "cannot_build";
		bool has_selection = false;
		bool has_morphing = false;
		for (const bwgame::unit_t* u : selected_units(owner)) {
			has_selection = true;
			if (can_start(u)) return economy_reason(u, unit_type);
			if (is_morphing(u)) has_morphing = true;
		}
		if (has_morphing) return "already_morphing";
		return has_selection ? "cannot_build" : "no_unit";
	}

	std::string reject_reason(int owner, int action_id, const uint8_t* payload) const {
		switch (action_id) {
		case action_id_train:
			return train_reject_reason(owner, unit_type_or_null(read_u16(payload)));
		case action_id_build:
			return build_reject_reason(owner, unit_type_or_null(read_u16(payload + 5)));
		case action_id_unit_morph: {
			auto* unit_type = unit_type_or_null(read_u16(payload));
			return morph_reject_reason(owner, unit_type, [&](const bwgame::unit_t* u) { return can_start_unit_morph(u, owner, unit_type); });
		}
		case action_id_building_morph: {
			auto* unit_type = unit_type_or_null(read_u16(payload));
			return morph_reject_reason(owner, unit_type, [&](const bwgame::unit_t* u) { return can_start_building_morph(u, owner, unit_type); });
		}
		}
		return "";
	}

	void record_reject(int owner, int action_id, const std::string& reason) {
		++rejects[{owner, action_id, reason}];
		std::fprintf(stderr, "reject frame=%d owner=%d action=%d reason=%s\n", st.current_frame, owner, action_id, reason.c_str());
	}

	bool is_tracked_action(int action_id) const {
		return action_id == action_id_train || action_id == action_id_build || action_id == action_id_unit_morph || action_id == action_id_building_morph;
	}

	void execute_action_group(bwgame::data_loading::data_reader_le& r, const uint8_t* end) {
		while (r.ptr != end) {
			int owner = owner_of_player_id(r.ptr[0]);
			int action_id = r.ptr + 1 != end ? r.ptr[1] : -1;
			std::string reason;
			if (owner >= 0 && is_tracked_action(action_id)) reason = reject_reason(owner, action_id, r.ptr + 2);
			if (owner >= 0) check_tags(owner, action_id, r.ptr + 2);
			if (!read_action(r) && !reason.empty()) record_reject(owner, action_id, reason);
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

	void next_frame() {
		if (st.current_frame == replay_st.end_frame) bwgame::error("replay: attempt to play past end");
		auto& buffer = replay_st.actions_data_buffer;
		execute_actions(buffer.data(), buffer.data() + buffer.size());
		bwgame::state_functions::next_frame();
	}
};

void print_rejects(const probe_functions& funcs) {
	for (auto& v : funcs.rejects) std::printf("reject %d %d %s %d\n", v.first.slot, v.first.action_id, v.first.reason.c_str(), v.second);
}

void print_tag_counts(const char* name, const std::map<tag_key, tag_counts>& counts) {
	for (auto& v : counts) {
		std::printf("%s %d %d total=%d missing=%d\n", name, v.first.slot, v.first.minute, v.second.total, v.second.missing);
	}
}

struct container_usage {
	const char* name;
	size_t max_size = 0;
	size_t peak = 0;
	int first_full_frame = 0;

	template<typename T, size_t container_max_size, size_t granularity>
	void update(const bwgame::object_container<T, container_max_size, granularity>& container, int frame) {
		max_size = container_max_size;
		size_t free = 0;
		for (auto& v : container.free_list) {
			(void)v;
			++free;
		}
		size_t in_use = container.size - free;
		peak = std::max(peak, in_use);
		if (in_use == container_max_size && !first_full_frame) first_full_frame = frame;
	}

	void print() const {
		std::printf("limit %s peak=%zu max=%zu first_full=%d\n", name, peak, max_size, first_full_frame);
	}
};

struct limit_tracker {
	container_usage units{"units"};
	container_usage bullets{"bullets"};
	container_usage sprites{"sprites"};
	container_usage images{"images"};
	container_usage orders{"orders"};

	void update(const bwgame::state& st) {
		units.update(st.units_container, st.current_frame);
		bullets.update(st.bullets_container, st.current_frame);
		sprites.update(st.sprites_container, st.current_frame);
		images.update(st.images_container, st.current_frame);
		orders.update(st.orders_container, st.current_frame);
	}

	void print() const {
		units.print();
		bullets.print();
		sprites.print();
		images.print();
		orders.print();
	}
};

int current_frame_for_errors = 0;

int run(const options& opts) {
	bwgame::game_player player(directory_data_loader{opts.data_directory});
	bwgame::action_state action_st;
	bwgame::replay_state replay_st;
	probe_functions funcs(player.st(), action_st, replay_st);
	funcs.load_replay_file(opts.replay_path.c_str());
	bwgame::state& st = player.st();
	production_tracker production;
	elimination_tracker eliminations;
	limit_tracker limits;
	production.scan(st, false);
	while (!funcs.is_done()) {
		funcs.next_frame();
		current_frame_for_errors = st.current_frame;
		limits.update(st);
		if (st.current_frame % scan_interval == 0) {
			production.scan(st, true);
			eliminations.update(st);
		}
		if (st.current_frame % timeline_interval == 0) print_timeline(st);
	}
	production.scan(st, true);
	print_timeline(st);
	print_made(st, production);
	print_eliminations(eliminations);
	print_rejects(funcs);
	limits.print();
	print_tag_counts("selection_tags", funcs.selection_tags);
	print_tag_counts("target_tags", funcs.target_tags);
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
		std::fflush(stdout);
		return 2;
	}
}
