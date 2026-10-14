#include "bwgame.h"

#include <cstdint>
#include <cstdio>

namespace {

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

uint64_t state_hash(const bwgame::state& st) {
	state_hasher h;
	h.add_int32((int32_t)st.current_frame);
	h.add_int32((int32_t)st.lcg_rand_state);
	for (size_t i = 0; i != 12; ++i) {
		h.add_int32((int32_t)st.current_minerals[i]);
		h.add_int32((int32_t)st.current_gas[i]);
		for (size_t race = 0; race != 3; ++race) h.add_int32((int32_t)st.supply_used[i][race].raw_value);
		for (size_t race = 0; race != 3; ++race) h.add_int32((int32_t)st.supply_available[i][race].raw_value);
	}
	hash_unit_list(h, st.visible_units);
	hash_unit_list(h, st.hidden_units);
	hash_unit_list(h, st.map_revealer_units);
	return h.value;
}

void fill_players(bwgame::state& st) {
	for (int i = 0; i != 12; ++i) {
		st.current_minerals[i] = i * 100 + 7;
		st.current_gas[i] = -i * 3;
		for (int race = 0; race != 3; ++race) {
			st.supply_used[i][race] = bwgame::fp1::from_raw(i * 10 + race);
			st.supply_available[i][race] = bwgame::fp1::from_raw(-(i + race));
		}
	}
}

void fill_unit(bwgame::unit_t& u, size_t index, const bwgame::unit_type_t* type, int owner, int x, int y, int hp, const bwgame::order_type_t* order, int status_flags) {
	u.index = index;
	u.unit_type = type;
	u.owner = owner;
	u.exact_position.x = bwgame::fp8::from_raw(x);
	u.exact_position.y = bwgame::fp8::from_raw(y);
	u.hp = bwgame::fp8::from_raw(hp);
	u.shield_points = bwgame::fp8::from_raw(hp / 2);
	u.energy = bwgame::fp8::from_raw(-hp);
	u.order_type = order;
	u.status_flags = status_flags;
}

}

int main() {
	static bwgame::state st;
	st.current_frame = 1234;
	st.lcg_rand_state = 0xdeadbeef;
	fill_players(st);

	static bwgame::unit_type_t marine;
	marine.id = bwgame::UnitTypes::Terran_Marine;
	static bwgame::unit_type_t drone;
	drone.id = bwgame::UnitTypes::Zerg_Drone;
	static bwgame::order_type_t move;
	move.id = bwgame::Orders::Move;

	static bwgame::unit_t units[3];
	fill_unit(units[0], 1, &marine, 0, 0x1234, 0x5678, 40 * 256, &move, 1);
	fill_unit(units[1], 1699, &drone, 11, -5, 70000, 1, &move, (int)0x80000001u);
	fill_unit(units[2], 42, &drone, 3, 256, 512, 0, nullptr, 0x4000);
	st.visible_units.push_back(units[0]);
	st.visible_units.push_back(units[1]);
	st.hidden_units.push_back(units[2]);

	std::printf("%016llx\n", (unsigned long long)state_hash(st));
	return 0;
}
