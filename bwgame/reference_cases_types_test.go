package bwgame

const referenceCasesCpp = `#include "bwgame.h"
#include <cstdio>
#include <cinttypes>
#include <vector>
#include <map>
using namespace bwgame;

static uint64_t rng_state = 0x9E3779B97F4A7C15ull;
static uint64_t next_rand() {
	rng_state ^= rng_state << 13;
	rng_state ^= rng_state >> 7;
	rng_state ^= rng_state << 17;
	return rng_state;
}
static const int32_t edge_values[] = {0, 1, -1, 2, -2, 255, 256, -256, 257, 0x7fff, -0x8000, 0x7fffffff, (int32_t)0x80000000, 12345678, -98765, 100, -100};
static int32_t random_i32() {
	uint64_t r = next_rand();
	switch (r % 4) {
	case 0: return edge_values[(r >> 8) % (sizeof(edge_values) / sizeof(edge_values[0]))];
	case 1: return (int32_t)(r >> 16);
	case 2: return (int32_t)((r >> 16) % 2000001) - 1000000;
	default: return (int32_t)((r >> 16) % 20001) - 10000;
	}
}
static int8_t random_i8() {
	return (int8_t)random_i32();
}
static bool safe_div(int32_t a, int32_t b) {
	return b != 0 && !(a == (int32_t)0x80000000 && b == -1);
}

static void fp8_cases() {
	for (int i = 0; i != 3000; ++i) {
		int32_t a = random_i32();
		int32_t b = random_i32();
		int32_t c = random_i32();
		fp8 fa = fp8::from_raw(a), fb = fp8::from_raw(b), fc = fp8::from_raw(c);
		printf("fp8_mul %d %d = %d\n", a, b, (fa * fb).raw_value);
		if (b != 0) printf("fp8_div %d %d = %d\n", a, b, (fa / fb).raw_value);
		printf("fp8_mulint %d %d = %d\n", a, b, (fa * b).raw_value);
		if (safe_div(a, b)) printf("fp8_divint %d %d = %d\n", a, b, (fa / b).raw_value);
		printf("fp8_integer %d = %d\n", a, fp8::integer(a).raw_value);
		printf("fp8_intpart %d = %d\n", a, fa.integer_part());
		printf("fp8_fracpart %d = %d\n", a, fa.fractional_part());
		printf("fp8_floor %d = %d\n", a, fa.floor().raw_value);
		printf("fp8_ceil %d = %d\n", a, fa.ceil().raw_value);
		printf("fp8_abs %d = %d\n", a, fa.abs().raw_value);
		printf("fp8_neg %d = %d\n", a, (-fa).raw_value);
		printf("fp8_add %d %d = %d\n", a, b, (fa + fb).raw_value);
		printf("fp8_sub %d %d = %d\n", a, b, (fa - fb).raw_value);
		if (c != 0) printf("fp8_muldiv %d %d %d = %d\n", a, b, c, fp8::multiply_divide(fa, fb, fc).raw_value);
		if (safe_div(a, b)) printf("fp8_divmul %d %d %d = %d\n", a, b, c, fp8::divide_multiply(fa, fb, fc).raw_value);
		printf("fp8_less %d %d = %d\n", a, b, fa < fb ? 1 : 0);
		printf("fp8_asunsigned %d = %u\n", a, fa.as_unsigned().raw_value);
		printf("fp8_truncate_dir %d = %d\n", a, (int)direction_t::truncate(fa).raw_value);
		using tmp_t = fixed_point<32, 8, true>;
		int32_t percent = (int32_t)(next_rand() % 101);
		tmp_t tmp = tmp_t::extend(fa);
		tmp = std::max(tmp * percent / 100, tmp_t::from_raw(1));
		printf("fp8_hp_percent %d %d = %d\n", a, percent, fp8::truncate(tmp).raw_value);
	}
}

static void ufp8_cases() {
	for (int i = 0; i != 2000; ++i) {
		uint32_t a = (uint32_t)random_i32();
		uint32_t b = (uint32_t)random_i32();
		uint32_t c = (uint32_t)random_i32();
		ufp8 fa = ufp8::from_raw(a), fb = ufp8::from_raw(b), fc = ufp8::from_raw(c);
		printf("ufp8_mul %u %u = %u\n", a, b, (fa * fb).raw_value);
		if (b != 0) printf("ufp8_div %u %u = %u\n", a, b, (fa / fb).raw_value);
		if (b != 0) printf("ufp8_divint %u %u = %u\n", a, b, (fa / b).raw_value);
		printf("ufp8_mulint %u %u = %u\n", a, b, (fa * b).raw_value);
		if (c != 0) printf("ufp8_muldiv %u %u %u = %u\n", a, b, c, ufp8::multiply_divide(fa, fb, fc).raw_value);
		if (b != 0) printf("ufp8_divmul %u %u %u = %u\n", a, b, c, ufp8::divide_multiply(fa, fb, fc).raw_value);
		printf("ufp8_integer %u = %u\n", a, ufp8::integer(a).raw_value);
		printf("ufp8_intpart %u = %u\n", a, fa.integer_part());
		printf("ufp8_ceil %u = %u\n", a, fa.ceil().raw_value);
		printf("ufp8_assigned %u = %d\n", a, fa.as_signed().raw_value);
	}
	printf("ufp8_third = %u\n", (ufp8::integer(6) + (ufp8::integer(1) - ufp8::integer(1) / 3u)).raw_value);
}

static void fp1_fp16_cases() {
	for (int i = 0; i != 2000; ++i) {
		int32_t a = random_i32();
		int32_t b = random_i32();
		fp1 a1 = fp1::from_raw(a), b1 = fp1::from_raw(b);
		printf("fp1_mul %d %d = %d\n", a, b, (a1 * b1).raw_value);
		if (b != 0) printf("fp1_div %d %d = %d\n", a, b, (a1 / b1).raw_value);
		printf("fp1_integer %d = %d\n", a, fp1::integer(a).raw_value);
		printf("fp1_intpart %d = %d\n", a, a1.integer_part());
		printf("fp1_ceil %d = %d\n", a, a1.ceil().raw_value);
		fp16 a16 = fp16::from_raw(a), b16 = fp16::from_raw(b);
		printf("fp16_mul %d %d = %d\n", a, b, (a16 * b16).raw_value);
		if (b != 0) printf("fp16_div %d %d = %d\n", a, b, (a16 / b16).raw_value);
		printf("fp16_integer %d = %d\n", a, fp16::integer(a).raw_value);
		printf("fp16_intpart %d = %d\n", a, a16.integer_part());
		printf("fp16_fracpart %d = %d\n", a, a16.fractional_part());
		printf("fp16_ceil %d = %d\n", a, a16.ceil().raw_value);
		ufp16 ua = ufp16::from_raw((uint32_t)a), ub = ufp16::from_raw((uint32_t)b);
		printf("ufp16_mul %u %u = %u\n", (uint32_t)a, (uint32_t)b, (ua * ub).raw_value);
		if (b != 0) printf("ufp16_div %u %u = %u\n", (uint32_t)a, (uint32_t)b, (ua / ub).raw_value);
	}
}

static void direction_cases() {
	for (int i = 0; i != 2000; ++i) {
		int8_t a = random_i8();
		int8_t b = random_i8();
		int32_t v = random_i32() % 1000;
		direction_t da = direction_t::from_raw(a), db = direction_t::from_raw(b);
		printf("dir_add %d %d = %d\n", a, b, (int)(da + db).raw_value);
		direction_t acc = da;
		acc += db;
		printf("dir_addassign %d %d = %d\n", a, b, (int)acc.raw_value);
		printf("dir_sub %d %d = %d\n", a, b, (int)(da - db).raw_value);
		printf("dir_neg %d = %d\n", a, (int)(-da).raw_value);
		printf("dir_mulint %d %d = %d\n", a, v, (int)(da * v).raw_value);
		if (v != 0) printf("dir_divint %d %d = %d\n", a, v, (int)(da / v).raw_value);
		printf("dir_intpart %d = %d\n", a, (int)da.integer_part());
		printf("dir_fracpart %d = %d\n", a, (int)da.fractional_part());
		printf("dir_abs %d = %d\n", a, (int)da.abs().raw_value);
		printf("dir_extend %d = %d\n", a, fp8::extend(da).raw_value);
		printf("dir_extend_sub_abs %d %d = %d\n", a, b, fp8::extend(da - db).abs().raw_value);
		printf("dir_less %d %d = %d\n", a, b, da < db ? 1 : 0);
	}
	printf("dir_literal_neg128 = %d\n", (int)(-128_dir).raw_value);
}

static void isqrt_cases() {
	for (int i = 0; i != 2000; ++i) {
		uint64_t r = next_rand();
		uint32_t a = (uint32_t)r;
		if (i % 3 == 0) a %= 100000;
		printf("isqrt32 %u = %u\n", a, isqrt(a));
		uint64_t b = next_rand();
		if (i % 3 == 0) b %= 1000000;
		printf("isqrt64 %" PRIu64 " = %" PRIu64 "\n", b, isqrt(b));
	}
}

static void id_cases() {
	for (int i = 0; i != 500; ++i) {
		size_t index = next_rand() % 0x800;
		unsigned int generation = (unsigned int)(next_rand() % 64);
		unit_id id(index, generation);
		printf("unit_id %zu %u = %u %zu %u\n", index, generation, (unsigned)id.raw_value, id.index(), id.generation());
		unit_id_32 id32(index, generation);
		printf("unit_id_32 %zu %u = %u %zu %u\n", index, generation, id32.raw_value, id32.index(), id32.generation());
		uint16_t raw = (uint16_t)next_rand();
		tile_id t(raw);
		printf("tile_id %u = %d %zu %zu\n", (unsigned)raw, t.has_creep() ? 1 : 0, t.group_index(), t.subtile_index());
	}
}

static void lcg_cases() {
	state st{};
	state_functions f(st);
	st.lcg_rand_state = 42;
	for (int i = 0; i != 2000; ++i) {
		int source = (int)(next_rand() % 64);
		if (i % 2) {
			int from = (int)(next_rand() % 200) - 100;
			int to = from + (int)(next_rand() % 1000);
			int v = f.lcg_rand(source, from, to);
			printf("lcg_range %d %d %d = %d %u\n", source, from, to, v, st.lcg_rand_state);
		} else {
			int v = f.lcg_rand(source);
			printf("lcg %d = %d %u\n", source, v, st.lcg_rand_state);
		}
	}
	printf("lcg_counts =");
	for (int i = 0; i != 64; ++i) printf(" %d", st.random_counts[i]);
	printf(" %d\n", st.total_random_counts);
}

static void object_container_cases() {
	object_container<image_t, 5000, 50> c;
	std::vector<image_t*> taken;
	for (int i = 0; i != 20000; ++i) {
		uint64_t r = next_rand();
		if (r % 5 < 3 || taken.empty()) {
			image_t* t = c.top();
			if (!t) {
				printf("oc top = -1\n");
				continue;
			}
			c.pop();
			taken.push_back(t);
			printf("oc top = %zu\n", t->index);
		} else {
			size_t n = (r >> 8) % taken.size();
			image_t* t = taken[n];
			taken.erase(taken.begin() + n);
			c.push(t);
			printf("oc push %zu =\n", t->index);
		}
	}
	printf("oc_free =");
	int n = 0;
	for (auto& v : c.free_list) {
		if (n++ == 200) break;
		printf(" %zu", v.index);
	}
	printf("\n");
	printf("oc_size = %zu\n", c.size);
}

static void intrusive_list_cases() {
	std::vector<image_t> images(32);
	for (size_t i = 0; i != images.size(); ++i) images[i].index = i;
	std::vector<bool> in_list(32);
	intrusive_list<image_t, default_link_f> list;
	for (int i = 0; i != 3000; ++i) {
		uint64_t r = next_rand();
		size_t n = (r >> 8) % 32;
		int op = (int)(r % 6);
		if (in_list[n]) {
			list.remove(images[n]);
			in_list[n] = false;
			printf("il remove %zu =", n);
		} else if (op == 0) {
			list.push_front(images[n]);
			in_list[n] = true;
			printf("il push_front %zu =", n);
		} else if (op == 1) {
			list.push_back(images[n]);
			in_list[n] = true;
			printf("il push_back %zu =", n);
		} else if (op == 2) {
			bw_insert_list(list, images[n]);
			in_list[n] = true;
			printf("il bw_insert %zu =", n);
		} else if (op == 3 && !list.empty()) {
			list.pop_front();
			for (size_t k = 0; k != 32; ++k) in_list[k] = false;
			for (auto& v : list) in_list[v.index] = true;
			printf("il pop_front %zu =", n);
		} else if (op == 4 && !list.empty()) {
			list.pop_back();
			for (size_t k = 0; k != 32; ++k) in_list[k] = false;
			for (auto& v : list) in_list[v.index] = true;
			printf("il pop_back %zu =", n);
		} else {
			auto pos = list.begin();
			size_t steps = (r >> 20) % 4;
			for (size_t k = 0; k != steps && pos != list.end(); ++k) ++pos;
			list.insert(pos, images[n]);
			in_list[n] = true;
			printf("il insert %zu %zu =", n, steps);
		}
		for (auto& v : list) printf(" %zu", v.index);
		printf(" |");
		for (auto& v : reverse(list)) printf(" %zu", v.index);
		printf("\n");
	}
}

static void table_cases() {
	for (size_t i = 0; i != tan_table.size(); ++i) printf("table_tan %zu = %u\n", i, tan_table[i]);
	for (size_t i = 0; i != repulse_adjust_table.size(); ++i) printf("table_repulse %zu = %d\n", i, repulse_adjust_table[i]);
	for (size_t i = 0; i != cardinal_direction_xy.size(); ++i) printf("table_cardinal %zu = %d %d\n", i, cardinal_direction_xy[i].x, cardinal_direction_xy[i].y);
	for (size_t i = 0; i != 256; ++i) printf("table_direction %zu = %d %d\n", i, direction_table[i].x.raw_value, direction_table[i].y.raw_value);
	for (size_t i = 0; i != 14; ++i) printf("table_hit_near %zu = %d %d\n", i, hit_near_target_positions[i].x, hit_near_target_positions[i].y);
	for (size_t i = 0; i != 5; ++i) {
		for (size_t j = 0; j != 8; ++j) printf("table_psi %zu %zu = %d\n", i, j, psi_field_mask[i][j] ? 1 : 0);
	}
	for (size_t i = 0; i != all_player_slots.size(); ++i) printf("table_player_slot %zu = %d\n", i, all_player_slots[i]);
}

static void static_vector_cases() {
	static_vector<int, 16> v;
	for (int i = 0; i != 3000; ++i) {
		uint64_t r = next_rand();
		int op = (int)(r % 4);
		int value = (int)((r >> 8) % 1000);
		if (op == 0 && v.size() != v.capacity()) {
			v.push_back(value);
			printf("sv push %d =", value);
		} else if (op == 1 && !v.empty()) {
			size_t pos = (r >> 20) % v.size();
			v.erase(v.begin() + pos);
			printf("sv erase %zu =", pos);
		} else if (op == 2 && !v.empty()) {
			v.pop_back();
			printf("sv pop 0 =");
		} else {
			size_t n = (r >> 20) % 17;
			v.resize(n);
			printf("sv resize %zu =", n);
		}
		for (int x : v) printf(" %d", x);
		printf("\n");
	}
}

int main() {
	static_vector_cases();
	table_cases();
	fp8_cases();
	ufp8_cases();
	fp1_fp16_cases();
	direction_cases();
	isqrt_cases();
	id_cases();
	lcg_cases();
	object_container_cases();
	intrusive_list_cases();
	return 0;
}
`
