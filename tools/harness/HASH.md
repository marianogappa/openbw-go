# State hash

The harness prints one 64-bit hash of the game state per frame. A Go port must produce the same value from the same state.

## Algorithm

FNV-1a 64:

- Start value: `0xcbf29ce484222325`.
- For each byte: `h ^= byte`, then `h *= 0x100000001b3` (mod 2^64).

The input is a sequence of `int32` values. Each value is fed as 4 bytes, little-endian, lowest byte first. Cast each field to `int32` with two's-complement truncation (for `uint32_t`, keep the bits).

Print the hash as 16 lowercase hex digits with leading zeros.

## When

- Frame 0: after `replay_functions::load_replay` returns.
- Frame n: after the n-th call of `next_frame` (actions of frame n-1 run, then `state_functions::next_frame`). `st.current_frame` is n.

## Field order

Fields are from `bwgame::state st` and `bwgame::unit_t* u`.

1. `st.current_frame`
2. `st.lcg_rand_state`
3. For each player slot `i` in 0..11:
   1. `st.current_minerals[i]`
   2. `st.current_gas[i]`
   3. `st.supply_used[i][r].raw_value` for `r` = 0, 1, 2 (zerg, terran, protoss)
   4. `st.supply_available[i][r].raw_value` for `r` = 0, 1, 2
4. For each unit list, in this order: `st.visible_units`, `st.hidden_units`, `st.map_revealer_units`:
   1. The number of units in the list.
   2. For each unit in list order (begin to end):
      1. `u->index`
      2. `(int)u->unit_type->id`
      3. `u->owner`
      4. `u->exact_position.x.raw_value`
      5. `u->exact_position.y.raw_value`
      6. `u->hp.raw_value`
      7. `u->shield_points.raw_value`
      8. `u->energy.raw_value`
      9. `(int)u->order_type->id`, or `-1` if `u->order_type` is null
      10. `u->status_flags`

Raw values:

- `fp8` (`exact_position`, `hp`, `shield_points`, `energy`): value × 256.
- `fp1` (`supply_used`, `supply_available`): value × 2. One zergling is 1, one marine is 2.

`st.dead_units` is not part of the hash.

## Other output

- `frame <n> <hash>` every `-every` frames (when `n % every == 0`).
- `end <frame> <hash>` after the last frame.
- `-dump-units F`: first a line `rng state=<lcg_rand_state> total=<total_random_counts> source_<i>=<count> ...` (non-zero `random_counts` only). Then a TSV block between `units_at_frame F` and `end_units`, in the same unit order as the hash. Columns: `index`, `type`, `owner`, `x` and `y` (`u->position`), `hp`, `shields`, `energy` (each `raw_value >> 8`), `order` (as in the hash), `completed` (`status_flag_completed`), then `exact_x`, `exact_y`, `hp_raw` (raw values), `status` (`status_flags`), `order_state`, `movement_state`, `heading` (raw), `order_timer` (`main_order_timer`), `speed` (`current_speed` raw) and `images` (the number of images of the unit's sprite, or -1).
- `-no-actions`: do not execute replay actions. The map, the players and the starting units still load from the replay. This is the "empty game" milestone.
- `player <slot> ...` and `units <slot> <type_id>=<count> ...` for each slot with an active controller. Unit counts come from `st.player_units[slot]`. Supply prints as `used/available` in game units.
- `action_fails <count>`: the number of actions where `read_action` returned `false`. Each one also prints `action_fail <frame> <slot> <action_id>` on stderr.
- `error <frame> <message>` on stdout, exit code 2, when OpenBW throws.
