#include "bwgame.h"
#include <cstdio>
#include <string>
#include <cctype>
#include <unistd.h>
#include <sys/wait.h>

using namespace bwgame;

static std::string normalize(std::string s) {
	for (auto& c : s) { if (c == '\\') c = '/'; c = (char)std::tolower((unsigned char)c); }
	return s;
}

struct loader {
	std::string dir;
	void operator()(a_vector<uint8_t>& dst, a_string filename) const {
		std::string p = dir + "/" + normalize(std::string(filename.c_str()));
		FILE* f = std::fopen(p.c_str(), "rb");
		if (!f) error("not found %s", p.c_str());
		std::fseek(f, 0, SEEK_END); long n = std::ftell(f); std::fseek(f, 0, SEEK_SET);
		dst.resize((size_t)n);
		if (n > 0) std::fread(dst.data(), 1, (size_t)n, f);
		std::fclose(f);
	}
};

struct tracer {
	std::string s;
	void add(const char* fmt, long long a) { char b[64]; std::snprintf(b, sizeof(b), fmt, a); s += b; }
};

static void trace_sprite(tracer& t, state& st, sprite_t* sprite) {
	t.add("sprite %lld", (long long)sprite->index);
	t.add(" w%lld", (long long)sprite->width);
	t.add(" h%lld", (long long)sprite->height);
	t.add(" v%lld", (long long)sprite->visibility_flags);
	t.add(" f%lld", (long long)sprite->flags);
	t.add(" main%lld\n", sprite->main_image ? (long long)sprite->main_image->index : -1);
	for (auto& i : sprite->images) {
		t.add(" img%lld", (long long)i.index);
		t.add(" t%lld", (long long)(int)i.image_type->id);
		t.add(" fi%lld", (long long)i.frame_index);
		t.add(" fb%lld", (long long)i.frame_index_base);
		t.add(" fo%lld", (long long)i.frame_index_offset);
		t.add(" fl%lld", (long long)i.flags);
		t.add(" ox%lld", (long long)i.offset.x);
		t.add(" oy%lld", (long long)i.offset.y);
		t.add(" m%lld", (long long)i.modifier);
		t.add(" d%lld", (long long)i.modifier_data1);
		t.add(" e%lld", (long long)i.modifier_data2);
		t.add(" pc%lld", (long long)i.iscript_state.program_counter);
		t.add(" w%lld", (long long)i.iscript_state.wait);
		t.add(" a%lld\n", (long long)i.iscript_state.animation);
	}
}

static void trace_frame(tracer& t, state& st, sprite_t* sprite, bool alive) {
	t.add("lcg %lld", (long long)st.lcg_rand_state);
	t.add(" thingies %lld\n", (long long)st.active_thingies_size);
	if (alive) trace_sprite(t, st, sprite);
	for (auto& th : st.active_thingies) trace_sprite(t, st, th.sprite);
}

static unsigned long long fnv(const std::string& s) {
	unsigned long long h = 0xcbf29ce484222325ull;
	for (unsigned char c : s) { h ^= c; h *= 0x100000001b3ull; }
	return h;
}

int main(int argc, char** argv) {
	global_state global_st;
	global_init(global_st, loader{argv[1]});
	int only = argc > 2 ? std::atoi(argv[2]) : -1;
	for (size_t id = 0; id != global_st.sprite_types.vec.size(); ++id) {
		if (only >= 0 && (int)id != only) continue;
		std::fflush(stdout);
		pid_t pid = only >= 0 ? 0 : fork();
		if (pid != 0) {
			int status = 0;
			waitpid(pid, &status, 0);
			if (!WIFEXITED(status)) std::printf("%d crash\n", (int)id);
			continue;
		}
		auto game_st = std::make_unique<game_state>();
		game_st->map_tile_width = 64;
		game_st->map_tile_height = 64;
		game_st->map_width = 64 * 32;
		game_st->map_height = 64 * 32;
		game_st->tileset_index = 0;
		auto st = std::make_unique<state>();
		st->global = &global_st;
		st->game = game_st.get();
		st->lcg_rand_state = 42;
		st->sprites_on_tile_line.resize(64);
		st->tiles.resize(64 * 64);
		state_functions f(*st);
		tracer t;
		bool alive = true;
		sprite_t* sprite = nullptr;
		try {
			sprite = f.create_sprite(&global_st.sprite_types.vec[id], {300, 400}, 1);
			if (!sprite) { t.s += "null\n"; alive = false; }
			trace_frame(t, *st, sprite, alive);
			for (int frame = 0; frame != 40 && alive; ++frame) {
				if (frame == 20) f.sprite_run_anim(sprite, iscript_anims::Death);
				alive = f.iscript_execute_sprite(sprite);
				for (auto i = st->active_thingies.begin(); i != st->active_thingies.end();) {
					thingy_t* th = &*i++;
					if (!f.iscript_execute_sprite(th->sprite)) {
						--st->active_thingies_size;
						st->active_thingies.remove(*th);
						bw_insert_list(st->free_thingies, *th);
					}
				}
				trace_frame(t, *st, sprite, alive);
			}
		} catch (...) {
			t.s += "error\n";
		}
		if (only >= 0) std::fputs(t.s.c_str(), stdout);
		else {
			std::printf("%d %016llx\n", (int)id, fnv(t.s));
			std::fflush(stdout);
			_exit(0);
		}
	}
}
