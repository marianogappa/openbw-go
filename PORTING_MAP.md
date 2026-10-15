# Engine porting map

This map splits the OpenBW engine into 14 chunks that agents can port in parallel. It covers `state_functions`, `game_load_functions` and `actions.h`.

- Source: `ref/openbw` at commit `4b046d5`. All line numbers refer to that commit.
- `bwgame.h` lines are written as `L1234`. `actions.h` lines are written as `A123`.
- Target: package `bwgame` (`github.com/marianogappa/openbw-go/bwgame`).
- Function lists use `name:line`, where `line` is the line of the signature (a `template<...>` line, if any, is the line before). C14 lists use `actions.h` lines. A `struct`/`enum` entry is a type declared in the range.
- Line ranges are inclusive. They start on the first line of a function (or its `template<...>` line) and end on the blank line before the next chunk. A script checked that the ranges of C00 (`L13197`–`L13206`) and C01–C12 cover `L332`–`L19553` with no gap and no overlap.

## Summary

| Chunk | Subsystem | Go files | C++ ranges | Lines |
|---|---|---|---|---|
| C00 | Foundation (prerequisite) | `state.go`, `engine_types.go`, `lcg.go`, `hooks.go`, `stubs_cNN.go` | `L1`–`L331`, `L13197`–`L13206`, plus the shared structs listed below | ~350 |
| C01 | Core queries and flag helpers | `queries.go` | `L332`–`L1934`, `L12204`–`L12310`, `L17242`–`L17323` | 1,792 |
| C02 | Unit operations (gather, build, own, load, addons) | `unit_ops.go` | `L1935`–`L3723` | 1,789 |
| C03 | Orders: basic, workers, Terran | `orders_basic.go` | `L3724`–`L3812`, `L3817`–`L3863`, `L4028`–`L5732` | 1,841 |
| C04 | Orders: spells, Zerg, Protoss, carriers | `orders_advanced.go` | `L5733`–`L7622` | 1,890 |
| C05 | Order dispatch, vision, frame loop | `order_dispatch.go`, `frame.go` | `L7623`–`L8468`, `L12311`–`L13021`, `L13051`–`L13105`, `L13191`–`L13196`, `L13207`–`L13210` | 1,622 |
| C06 | Flingy movement and movement state machine | `movement.go` | `L8469`–`L9161`, `L10933`–`L12203`, `L13359`–`L13426` | 2,032 |
| C07 | Pathfinder | `pathfinder.go` | `L9162`–`L10932` | 1,771 |
| C08 | Combat, spells, bullets | `combat.go`, `bullets.go` | `L13435`–`L14535`, `L14573`–`L14637` | 1,166 |
| C09 | Iscript, images, sprites, thingies | `iscript.go`, `sprites.go` | `L13211`–`L13358`, `L13427`–`L13434`, `L14536`–`L14572`, `L14638`–`L15369` | 925 |
| C10 | Unit lifecycle, unit finder, order queue | `units.go`, `unit_finder.go` | `L15370`–`L17241` | 1,872 |
| C11 | Tech tree and order requirements | `requirements.go` | `L17324`–`L18784` | 1,461 |
| C12 | Triggers and creep | `triggers.go`, `creep.go` | `L3813`–`L3816`, `L3864`–`L4027`, `L13022`–`L13050`, `L13106`–`L13190`, `L18785`–`L19553` | 1,051 |
| C13 | Map loading (CHK) and region creation | `map_load.go`, `regions_create.go` | `L19764`–`L21964` | 2,201 |
| C14 | Player actions | `actions.go` | `A1`–`A1478` | 1,478 |

Not in any chunk (see "Skip or port elsewhere"): `L19556`–`L19761` (`state_copier`, `copy_state`), `L21966`–`L21980` (`string_table_data`), `L21982`–`L22336` (`global_init`), `L22338`–`L22394` (`game_player`).

## Dependency order

Every chunk calls into almost every other chunk, so the call graph has cycles. Stubs break the cycles (see "Stub plan"). With stubs, all 14 chunks can be ported at the same time after C00.

### Waves

| Wave | Chunks | Why |
|---|---|---|
| 0 (serial, first) | C00 | Types, containers, `lcg_rand`, hooks, and the stub files with every Go signature. Nothing compiles without it. |
| 1 (parallel) | C01, C07, C09, C10, C11, C13 | Most-called code. C11 calls only C01. C07 calls only C01 and a few C06/C10 helpers. C13 makes the first integration test possible. |
| 2 (parallel) | C02, C05, C06, C08, C12 | Mid-level mechanics. |
| 3 (parallel) | C03, C04, C14 | Leaf callers: orders and actions. Nothing calls them except C05 (`execute_main_order`) and the replay driver. |

If agents are limited, use this priority: C00, C01, C09, C10, C13, C05, C06, C07, C12, C02, C08, C11, C03, C04, C14.

### Integration milestones

1. **Empty game**: C00, C01, C05, C06, C09, C10, C12, C13. Load a melee map, create the starting units, run `process_frame` twice (`L21908`–`L21911`), then run N frames without actions. Compare `lcg_rand_state`, `random_counts` and unit positions with the C++ harness.
2. **Economy**: add C02, C03, C11, C14. Replays with worker and build commands only.
3. **Full game**: add C04, C07 (if still stubbed), C08. All replays.

### Calls between chunks

Counts are call sites found by name. They include some false positives from overloads that share a name.

| Chunk | Calls into |
|---|---|
| C01 | C02 (`auto_attack_target`, `kill_unit`, `try_gather_resource`), C06 (`unit_dead`, `unit_halt_distance`, `unit_target_is_enemy`), C08 (`unit_deal_damage`, `weapon_deal_damage`, `deal_defensive_matrix_damage`, `update_acid_spore_image`), C09 (`create_image`, `destroy_image`, `get_image_lo_offset`, `iscript_run_anim`, `sprite_run_anim`), C10 (`find_units_noexpand`, `set_unit_order`, `remove_one_order`, `unit_is_factory`, `update_unit_speed`, `damage_overlay_states`), C14 hook `on_unit_deselect` |
| C02 | C01 (421), C06 (28), C07 (`free_path`, `pathfinder_find_long_path`), C08 (`spell_order_valid`), C09 (23), C10 (80), C12 (`remove_creep_provider`) |
| C03 | C01 (365), C02 (216), C05 (`refresh_unit_vision`, `resume_building_unit`, `try_attack_something`, `update_unit_sprite`), C06 (44), C09 (31), C10 (125), C11 (`unit_can_use_tech`), C12 (`remove_creep_provider`) |
| C04 | C01 (372), C02 (206), C03 (21), C05 (`rally_unit`, `remove_target_references`, `try_attack_something`), C06 (48), C08 (9), C09 (60), C10 (154), C11 (`unit_can_receive_order`), C12 (`add_creep_provider`, `set_creep_building_tiles`, `tile_has_creep`, `zerg_building_start_construction`) |
| C05 | C01 (193), C02 (14), C03 (every `order_*`), C04 (every `order_*`), C06 (`execute_movement`, `get_modified_weapon_cooldown`, `unit_dead`), C08 (`bullet_execute`), C09 (28), C10 (22), C12 (`process_triggers`, `recede_creep`, `spawn_larva`, `spread_creep`, `player_slot_active`) |
| C06 | C01 (214), C02 (`set_next_target_waypoint`, `stop_unit`), C07 (`unit_path_to`, `unit_update_path_movement_state`, `create_single_step_path`, `free_path`, `pathfinder_adjust_destination`, `pathfinder_unit_can_collide_with`), C09 (`move_sprite`, `set_image_heading`, `sprite_run_anim`), C10 (`find_units*`, `unit_finder_reinsert`, repulse field, `check_unit_collision`, `is_intersecting`) |
| C07 | C01 (51), C02 (`set_next_target_waypoint`), C06 (`is_moving_along_path`, `set_flingy_move_target`, `set_unit_move_target`, `unit_halt_distance`, `unit_is_at_move_target`), C10 (`find_units_noexpand`, `is_intersecting`) |
| C08 | C01 (207), C02 (`kill_unit`, `order_done`, `stop_unit`, `set_next_target_waypoint`), C03 (`set_remove_timer`), C05 (`reveal_sight_at`), C06 (21), C09 (22), C10 (27) |
| C09 | C01 (59), C02 (`set_next_target_waypoint`), C06 (`get_modified_unit_speed`, `set_flingy_move_target`, `set_next_speed`, `unit_type_can_fit_at`), C08 (`bullet_hit`, `create_bullet`, `fire_weapon`, `get_bullet_appear_at_target_pos`, `melee_deal_damage`, `spell_order_valid`), C10 (`activate_next_order`, `iscript_run_to_idle`, `set_unit_heading`) |
| C10 | C01 (266), C02 (23), C03 (`set_remove_timer`), C05 (`local_unit_status_error`, `refresh_unit_vision`, `remove_target_references`, `update_unit_detected_flags`, `update_unit_sprite`), C06 (15), C07 (`free_path`), C08 (10), C09 (28), C12 (`remove_creep_provider`, `remove_extractor_creep`, `spread_creep_completely`), C14 hook `on_unit_deselect` |
| C11 | C01 only |
| C12 | C01 (81), C02 (13), C05 (`active_players`, `remove_player`), C06 (`unit_dying`), C08 (`unburrow_unit`), C10 (21), C11 (`unit_can_receive_order`) |
| C13 | C01 (46), C02 (`kill_unit`), C03 (`make_unit_hallucination`, `set_remove_timer`), C05 (`active_players`, `process_frame`, `update_unit_sprite`), C09 (`create_thingy`), C10 (`create_unit`, `complete_unit`, `create_initial_unit`, `finish_building_unit`, `place_completed_unit`, `find_units`), C12 (`spawn_larva`, `spread_creep_completely`) |
| C14 | C01 (152), C02 (25), C05 (`player_left`, `player_dropped`), C06 (`rectangle_can_fit_at`, `unit_dead`, `unit_target_is_enemy`, `unit_type_can_fit_at`), C07 (`pathfinder_adjust_destination`, `pathfinder_adjust_target_pos`, `pathfinder_find_long_path`), C08 (`unburrow_unit`, `unit_deal_damage`), C09 (`sprite_run_anim`), C10 (`find_unit_noexpand`, `issue_order`, `set_queued_order`, `set_unit_order`, `unit_is_factory`, `update_unit_speed`), C11 (`unit_can_*`, `unit_build_order_valid`) |

## Stub plan

The goal: every chunk compiles and runs its tests before its dependencies exist.

1. C00 writes one file per chunk: `bwgame/stubs_c01.go` … `bwgame/stubs_c14.go`.
2. Each stub file holds the final Go signature of every function in that chunk, with the body `panic(notPorted("<c++ name>"))`.
3. A chunk agent owns its stub file. It deletes each stub as it ports the function, and deletes the file at the end. It must not edit another chunk's stub file.
4. Signatures are fixed by C00. If a chunk needs a signature change, it asks the C00 owner. Do not change a signature locally.
5. A test that reaches an unported function panics with the C++ name. This tells the agent which chunk is missing.
6. `notPorted` lives in C00 and returns an error value that names the function, so `go vet` stays clean.

C00 writes the signatures from the function lists in this file. It uses the naming rules and the overload table below.

## Go conventions for all chunks

- **Receiver**: `func (f *StateFunctions) unitCanAttackTarget(u, target *Unit) bool`. `game_load_functions` becomes `GameLoadFunctions` that embeds `*StateFunctions`. `action_functions` becomes `ActionFunctions` that embeds `*StateFunctions`.
- **Names**: snake_case to unexported lowerCamelCase. `order_Stop` → `orderStop`. `secondary_order_Train` → `secondaryOrderTrain`. `movement_UM_AtRest` → `movementUMAtRest`. `bullet_state_init` → `bulletStateInit`.
- **One Go function per C++ function.** `const` and non-const twins (`unit_main_unit`, `unit_attacking_unit`) become one Go function. Lambdas become closures inside the Go function.
- **Default arguments**: Go has none. Every caller passes the C++ default value explicitly.
- **`unit_type_autocast`** (`L84`–`L100`): the Go parameter is `*UnitType`. Callers that pass a unit pass `u.UnitType`.
- **Inheritance**: `unit_t` and `bullet_t` extend `flingy_t`, which extends `thingy_t`. Embed `Flingy` in `Unit` and `Bullet`, and `Thingy` in `Flingy`. Functions that take `flingy_t*` take `*Flingy`. Compare identity with the embedded pointer (`iscript_flingy == &u.Flingy`).
- **Integer widths**: C++ `int` is 32-bit: use `int32` where the value can overflow or is stored in state. `size_t` is 64-bit unsigned on the harness platform: use `uint` where C++ can wrap (for example `(size_t)-1` or unsigned subtraction). Keep `uint8_t`/`int16_t`/`uint32_t` exact. Keep the C++ fixed-point types (`fp8`, `ufp8`, `fp1`, `direction_t`, `xy_fp8`) as Go types with the same raw width and the same multiply, divide and shift rules (C00).
- **`error(...)`**: C++ throws. Go panics with a typed error. The public API recovers it.
- **Ranges returned by `auto`** (`loaded_units`, `active_players`, `trigger_players`, `selected_units`, `control_group`): return `iter.Seq[...]`. These ranges are lazy in C++ and read state while the caller loops. A slice copy can change behaviour.
- **RAII setters**: `make_thingy_setter(iscript_unit, x)` (`L364`–`L380`) restores the old value at the end of the C++ scope. In Go, save and restore by hand at the same scope end. Do not use `defer` when the C++ scope is a block inside the function.
- **Short-circuit order**: keep the `&&`/`||` operand order exactly, because some operands call `lcg_rand` (for example `L14448`, `L14456`). No line in the source has two `lcg_rand` calls in one expression, so there is no unspecified evaluation order.
- **Iteration with removal**: C++ uses two loop styles. `for (auto i = list.begin(); i != end;) { x = &*i++; ... }` takes the next pointer before the body runs. `for (x : ptr(list))` does not. Port each loop with the same style.
- **No pointer ordering**: the source compares pointers only for equality. Order comes from intrusive lists, pool free lists and the unit finder arrays. Port those containers exactly.

### Overload names

| C++ | Go |
|---|---|
| `action_select(owner, units)` A268 / `(owner, u)` A286 | `actionSelect` / `actionSelectUnit` |
| `action_shift_select` A291 / A309 | `actionShiftSelect` / `actionShiftSelectUnit` |
| `action_deselect` A314 / A329 | `actionDeselect` / `actionDeselectUnit` |
| `add_acid_spore(u)` L14069 / `(pos, owner)` L14075 | `addAcidSpore` / `addAcidSporeAt` |
| `bullet_deal_damage(b, t, div)` L13735 / `(b, t)` L13742 | `bulletDealDamageDivisor` / `bulletDealDamage` |
| `bullet_deal_splash_damage<bool>` L13774 | `bulletDealSplashDamage(b, isAirSplash bool)` |
| `cancel_build_queue(u, index)` L2430 / `(u)` L2443 | `cancelBuildQueueIndex` / `cancelBuildQueue` |
| `check_unit_movement_terrain_collision(u, xy)` L8807 / `(u, ems)` L8981 | `checkUnitMovementTerrainCollision` / `checkUnitMovementTerrainCollisionEMS` |
| `create_unit(type*)` L16317 / `(UnitTypes)` L16360 | `createUnit` / `createUnitByID` |
| `destroy_image_from_to(sprite, …)` L987 / `(u, …)` L992 | `destroyImageFromTo` / `destroyImageFromToUnit` |
| `find_image(sprite, …)` L974 / `(u, …)` L981 | `findImage` / `findImageUnit` |
| `direction_xy(dir, fp8)` L1661 / `(dir, int)` L1665 / `(dir)` L1669 | `directionXY` / `directionXYInt` / `directionXYOne` |
| `find_nearest_unit(pos, area, 4 iterators, pred)` L15706 / `(pos, area, pred)` L15795 / `(u, area, pred)` L15807 | `findNearestUnitFromIndices` / `findNearestUnit` / `findNearestUnitToUnit` |
| `find_unit_placement(u, pos, bounds, t)` L16501 / `(u, pos, t)` L16692 | `findUnitPlacementInBounds` / `findUnitPlacement` |
| `free_path(u)` L10763 / `(path)` L10769 | `freeUnitPath` / `freePath` |
| `get_unit(index)` L626 / `(unit_id)` L634 | `getUnit` / `getUnitByID` |
| `get_unit_strength(u, ground)` L849 / game load `(type, weapon)` L20828 | `getUnitStrength` / `(*GameLoadFunctions).getUnitTypeStrength` |
| game load `get_unit_type`, `get_weapon_type`, `get_upgrade_type`, `get_tech_type`, `get_flingy_type` L19781–L19800 | Not ported: the `StateFunctions` getters already return mutable pointers |
| `has_available_resources_for(owner, unit_type)` L2241 / `tech` L2255 / `upgrade` L2279 | `hasAvailableResourcesFor` / `hasAvailableResourcesForTech` / `hasAvailableResourcesForUpgrade` |
| `i_set_flag(i, f)` L400 / `(i, f, v)` L406 | `iSetFlag` / `iSetFlagValue` |
| `u_set_status_flag(u, f)` L381 / `(u, f, v)` L388 | `uSetStatusFlag` / `uSetStatusFlagValue` |
| `is_in_bounds(rect, rect)` L678 / `(xy, rect)` L694 | `isInBounds` / `isPosInBounds` |
| `is_in_inner_bounds(rect, rect)` L686 / `(xy, rect)` L698 | `isInInnerBounds` / `isPosInInnerBounds` |
| `is_in_map_bounds(type, xy)` L653 / `(rect)` L660 / `(xy)` L674 | `isUnitTypeInMapBounds` / `isInMapBounds` / `isPosInMapBounds` |
| `is_reachable(xy, xy)` L1518 / `(u, xy)` L1522 / `(u, u)` L1528 | `isReachable` / `isReachableUnitToPos` / `isReachableUnitToUnit` |
| `lcg_rand(source)` L13197 / `(source, from, to)` L13203 | `lcgRand` / `lcgRandRange` |
| `pathfinder_find_long_path(pf)` L9191 / `(pf, from, to)` L10710 | `pathfinderFindLongPath` / `pathfinderFindLongPathFromTo` |
| `pathfinder_unit_can_collide_with(u, t, cu, cm)` L9406 / `(pf, t)` L9417 | `pathfinderUnitCanCollideWith` / `pathfinderUnitCanCollideWithPF` |
| `play_sound(id, pos, src, race)` L334 / `(id, u, race)` L618 / `(id, race)` L622 | hook `playSoundAt` / `playSoundUnit` / `playSound` |
| `read_action(r)` A1340 / `(owner, r)` A1348 / `(data, size)` A1446 / `(owner, data, size)` A1451 | `readAction` / `readActionOwner` / `readActionBytes` / `readActionOwnerBytes` |
| `remove_creep_provider(type, pos, done)` L3907 / `(u)` L3932 | `removeCreepProvider` / `removeUnitCreepProvider` |
| `remove_target_references(u, t)` L12716 / `(b, t)` L12753 | `removeTargetReferences` / `removeBulletTargetReferences` |
| `set_remove_timer(u, timer)` L3724 / `(u)` L3734 | `setRemoveTimer` / `setDefaultRemoveTimer` |
| `set_unit_move_target(u, xy)` L8580 / `(u, unit)` L8594 | `setUnitMoveTarget` / `setUnitMoveTargetUnit` |
| `set_unit_order(u, ot)` L16935 / `(u, ot, order_target_t)` L16941 / `(u, ot, unit)` L16947 / `(u, ot, xy)` L16958 | `setUnitOrder` / `setUnitOrderTarget` / `setUnitOrderUnit` / `setUnitOrderPos` |
| `unit_can_research(u, t, owner)` L18198 / `(u, t)` L18203 | `unitCanResearchOwner` / `unitCanResearch` |
| `unit_can_upgrade` L18333 / L18338 | `unitCanUpgradeOwner` / `unitCanUpgrade` |
| `unit_can_use_tech` L18640 / L18645 | `unitCanUseTechOwner` / `unitCanUseTech` |
| `unit_finder_insert(u)` L15486 / `(u, bb)` L15517 | `unitFinderInsert` / `unitFinderInsertBB` |
| `unit_finder_reinsert(u)` L15493 / `(u, bb)` L15532 | `unitFinderReinsert` / `unitFinderReinsertBB` |
| `unit_type_bounding_box(t)` L710 / `(t, origin)` L713 | `unitTypeBoundingBox` / `unitTypeBoundingBoxAt` |
| `unit_type_inner_bounding_box(t)` L724 / `(t, origin)` L727 | `unitTypeInnerBoundingBox` / `unitTypeInnerBoundingBoxAt` |
| `weapon_can_target_unit(w, t)` L13757 / `(w, t, src)` L13768 | `weaponCanTargetUnit` / `weaponCanTargetUnitFrom` |
| `xy_direction(xy_fp8)` L1645 / `(xy)` L1653 | `xyDirectionFP8` / `xyDirection` |
| `xy_length(xy)` L1607 / `(xy_fp8)` L1615 | `xyLength` / `xyLengthFP8` |
| `atan<integer_bits>` L1636 | One Go function per instantiation that the code uses, named by type (for example `atanFP8`) |

## C00 Foundation (prerequisite, one agent)

Not part of the 14 chunks, but every chunk needs it.

- `bwgame.h` `L1`–`L331`: lookup tables (`tan_table`, `repulse_adjust_table`, `cardinal_direction_xy`, `direction_table`, `hit_near_target_positions`, `psi_field_mask`, `all_player_slots`), `autocast`, `global_state`, `game_state`, `state_base_copyable`, `psionic_matrix_link_f`, `state_base_non_copyable`, `state`.
- `lcg_rand` `L13197`–`L13206`: uint32 wrap, `(state >> 16) & 0x7fff`, increments `random_counts[source]` and `total_random_counts`.
- Types from `game_types.h`, `data_types.h`, `util.h` (fixed point, `xy`, `rect`, `direction_t`, `get_best_score`, `isqrt`), `intrusive_list.h`, `static_vector.h`, `circular_vector.h`. Enums already exist in `bwgame/enums.go`.
- Shared structs declared inside chunk ranges but used by more than one chunk. C00 declares them in `engine_types.go`:
  - `execute_movement_struct` `L8749`–`L8759` (C06, C08)
  - `pathfinder` `L9162`–`L9190` (C02, C07, C14)
  - image order enum `L15264`–`L15269` (`image_order_top`, …; C01, C02, C04, C08, C09, C10)
  - `state::unit_finder_entry` `L318`
- `StateFunctions` struct with the members `L345`–`L357`: `st`, `global_st`, `game_st`, `update_tiles`, `iscript_flingy`, `iscript_bullet`, `iscript_unit`, `unit_finder_search_index`.
- Hooks (`L334`–`L341`): `play_sound`, `on_unit_deselect`, `on_unit_destroy`, `on_kill_unit`, `on_player_eliminated`, `on_victory_state`, plus `on_action` (`A1336`). Go embedding is not virtual, so use a hooks struct of function fields with no-op defaults. `ActionFunctions` sets `onUnitDeselect` to its own method (`A91`): this one changes selections, so it is gameplay, not UI.
- Pitfalls:
  - `object_container` (`game_types.h` L48): `get(index)` maps a non-zero index to `max_size - index`. Free-list order decides which object (and which unit index) comes next. Unit indices appear in replay commands, so this must be exact.
  - `bw_insert_list` (`game_types.h` L42) inserts after the first element, not at the front. It is used for the pool free lists and the sprite tile lines.
  - `static_vector` throws when full. Go must panic at the same capacity.
  - `intrusive_list::remove` must leave links exactly as C++ does, because some loops read the next pointer of a removed node.

## C01 Core queries and flag helpers → `queries.go`

Ranges: `L332`–`L1934`, `L12204`–`L12310`, `L17242`–`L17323`.

Functions:
`get_order_type:359`, `struct thingy_setter:365`, `make_thingy_setter:377`, `u_set_status_flag:381`, `u_unset_status_flag:384`, `u_set_status_flag:388`, `u_set_movement_flag:393`, `u_unset_movement_flag:396`, `i_set_flag:400`, `i_unset_flag:403`, `i_set_flag:406`, `ut_flag:411`, `u_status_flag:414`, `us_flag:417`, `s_flag:420`, `u_movement_flag:423`, `i_flag:426`, `u_completed:430`, `u_in_bunker:433`, `u_loaded:436`, `u_immovable:439`, `u_disabled:442`, `u_burrowed:445`, `u_can_turn:448`, `u_can_move:451`, `u_grounded_building:454`, `u_hallucination:457`, `u_lifetime_expired:460`, `u_flying:463`, `u_speed_upgrade:466`, `u_cooldown_upgrade:469`, `u_gathering:472`, `u_requires_detector:475`, `u_cloaked:478`, `u_cannot_attack:481`, `u_order_not_interruptible:484`, `u_iscript_nobrk:487`, `u_collision:490`, `u_ground_unit:493`, `u_no_collide:496`, `u_invincible:499`, `u_ready_to_attack:502`, `u_passively_cloaked:505`, `ut_turret:509`, `ut_worker:512`, `ut_hero:515`, `ut_building:518`, `ut_addon:521`, `ut_flyer:524`, `ut_can_turn:527`, `ut_can_move:530`, `ut_invincible:533`, `ut_two_units_in_one_egg:536`, `ut_regens_hp:539`, `ut_flying_building:542`, `ut_requires_psionic_matrix:545`, `ut_can_burrow:548`, `ut_has_energy:551`, `ut_resource_depot:554`, `ut_resource:557`, `ut_initially_cloaked:560`, `ut_detector:563`, `ut_requires_creep:566`, `ut_mechanical:569`, `ut_organic:572`, `ut_robotic:575`, `ut_powerup:578`, `us_hidden:582`, `s_hidden:585`, `get_unit_type:589`, `get_image_type:593`, `get_weapon_type:597`, `get_sprite_type:601`, `get_upgrade_type:605`, `get_tech_type:609`, `get_flingy_type:613`, `play_sound:618`, `play_sound:622`, `get_unit:626`, `get_unit:634`, `get_unit_id:643`, `get_unit_id_32:648`, `is_in_map_bounds:653`, `is_in_map_bounds:660`, `is_in_inner_map_bounds:667`, `is_in_map_bounds:674`, `is_in_bounds:678`, `is_in_inner_bounds:686`, `is_in_bounds:694`, `is_in_inner_bounds:698`, `translate_rect:702`, `square_at:706`, `unit_type_bounding_box:710`, `unit_type_bounding_box:713`, `unit_bounding_box:716`, `unit_sprite_bounding_box:720`, `unit_type_inner_bounding_box:724`, `unit_type_inner_bounding_box:727`, `unit_inner_bounding_box:730`, `unit_sprite_inner_bounding_box:734`, `map_bounds:738`, `restrict_unit_pos_to_bounds:742`, `restrict_pos_to_map_bounds:751`, `is_inner_bb_move_target_in_valid_bounds:759`, `restrict_move_target_to_valid_bounds:763`, `is_walkable:767`, `tiles_flags_and:782`, `tiles_flags_or:793`, `unit_type_spreads_creep:805`, `visible_hp_plus_shields:815`, `max_visible_hp:821`, `max_visible_hp_plus_shields:827`, `loaded_units:833`, `unit_space_occupied:841`, `get_unit_strength:849`, `update_unit_damage_overlay:869`, `set_unit_hp:926`, `set_unit_shield_points:936`, `set_unit_energy:940`, `unit_is_mineral_field:944`, `set_unit_resources:951`, `unit_is_disabled:966`, `find_image:974`, `find_image:981`, `destroy_image_from_to:987`, `destroy_image_from_to:992`, `disable_effect_end:997`, `remove_stasis:1023`, `deal_irradiate_damage:1029`, `remove_irradiate:1054`, `remove_ensnare:1061`, `remove_lockdown:1067`, `remove_plague:1072`, `remove_maelstrom:1077`, `remove_acid_spores:1082`, `update_unit_status_timers:1088`, `unit_cloak_energy_cost:1159`, `set_secondary_order:1176`, `update_unit_energy:1183`, `unit_hp_percent:1203`, `update_unit_values:1209`, `unit_turret:1254`, `unit_main_unit:1260`, `unit_attacking_unit:1264`, `unit_main_unit:1268`, `unit_attacking_unit:1272`, `unit_ground_weapon:1276`, `unit_air_weapon:1281`, `unit_or_subunit_ground_weapon:1285`, `unit_or_subunit_air_weapon:1291`, `unit_target_weapon:1297`, `unit_is_carrier:1301`, `unit_is_reaver:1305`, `unit_is_queen:1309`, `unit_is_hatchery:1313`, `unit_is_marine:1320`, `unit_is_firebat:1326`, `unit_is_ghost:1332`, `unit_is_wraith:1341`, `unit_is_map_revealer:1347`, `unit_is_refinery:1353`, `unit_is_non_flag_beacon:1360`, `unit_is_special_beacon:1364`, `unit_is_scout:1368`, `unit_is_vulture:1372`, `unit_is_nydus:1376`, `unit_is_defiler:1380`, `unit_is_ultralisk:1386`, `unit_is_arbiter:1392`, `unit_is_egg:1398`, `unit_is_goliath:1405`, `unit_is_critter:1411`, `unit_is_sieged_tank:1421`, `unit_is_unsieged_tank:1427`, `unit_is_tank:1433`, `unit_is_trap:1437`, `unit_is_zerg_building:1441`, `unit_is_fighter:1445`, `unit_is_undetected:1451`, `unit_target_is_undetected:1457`, `unit_target_is_visible:1461`, `unit_position_is_visible:1465`, `unit_position_is_explored:1469`, `unit_can_see_target:1473`, `unit_order_can_target_self:1478`, `player_position_is_visible:1492`, `player_position_is_explored:1496`, `get_region_at:1500`, `get_region_at_prefer_walkable:1510`, `is_reachable:1518`, `is_reachable:1522`, `is_reachable:1528`, `unit_can_be_infested:1532`, `unit_can_attack_target:1538`, `unit_autoattack:1555`, `unit_can_collide_with:1567`, `rect_difference:1587`, `nearest_pos_in_rect:1599`, `xy_length:1607`, `xy_length:1615`, `to_xy_fp8:1619`, `to_xy:1622`, `units_distance:1626`, `unit_distance_to:1632`, `atan:1637`, `xy_direction:1645`, `xy_direction:1653`, `direction_xy:1661`, `direction_xy:1665`, `direction_xy:1669`, `direction_index:1673`, `direction_from_index:1679`, `units_direction:1685`, `unit_target_in_attack_angle:1690`, `weapon_max_range:1702`, `unit_target_movement_range:1729`, `unit_target_in_weapon_movement_range:1737`, `unit_target_out_of_max_range:1750`, `unit_first_loaded_unit:1760`, `unit_race:1767`, `unit_provides_space:1774`, `unit_space_left:1781`, `unit_can_load_target:1789`, `unit_interceptor_count:1801`, `unit_scarab_count:1805`, `unit_max_interceptor_count:1810`, `unit_max_scarab_count:1815`, `unit_queued_fighter_units:1821`, `unit_spider_mine_count:1829`, `unit_can_attack:1834`, `unit_target_attack_priority:1841`, `unit_target_in_range:1864`, `get_default_priority_targets:1868`, `find_acquire_target:1893`, `find_acquire_random_target:1918`, `unit_type_is_morphing_building:12204`, `unit_is_morphing_building:12213`, `unit_sight_range:12219`, `unit_sight_range_ignore_blindness:12228`, `unit_target_acquisition_range:12232`, `unit_max_energy:12248`, `visible_to_everyone:12272`, `tile_index:12284`, `tile_visibility:12291`, `tile_explored:12295`, `get_ground_height_at:12299`, `display_last_error_for_player:17242`, `player_has_unit:17246`, `player_has_completed_unit:17249`, `player_tech_available:17252`, `player_has_researched:17255`, `player_is_researching:17258`, `player_max_upgrade_level:17261`, `player_has_upgrade:17264`, `player_is_upgrading:17267`, `player_upgrade_level:17270`, `unit_is:17273`, `unit_addon:17276`, `unit_has_addon:17279`, `unit_is_building_addon:17282`, `unit_is_building_protoss_thing:17288`, `unit_is_constructing:17294`, `unit_is_researching:17300`, `unit_is_upgrading:17303`, `unit_can_move:17306`, `unit_can_hold_position:17313`, `unit_is_immovable_attacker:17318`

Calls: see the dependency table. Almost every other chunk calls C01 (C11 calls nothing else).

Pitfalls:
- `atan` (`L1636`) is a template on the fixed-point integer bits. It runs `lower_bound` on the unsigned raw value. Keep the raw unsigned width per instantiation.
- `xy_length` (`L1607`–`L1617`) uses `unsigned` math. `direction_xy(dir, fp8)` and `direction_xy(dir, int)` multiply differently (fixed-point multiply against plain integer multiply).
- `direction_t` is a signed 8-bit angle that wraps. `direction_index` and `direction_from_index` (`L1673`–`L1684`) convert between signed and 0–255.
- `update_unit_damage_overlay` (`L869`) calls `lcg_rand(8)` and reads lo offsets by frame index.
- `get_default_priority_targets` (`L1868`) fills 6 buckets of at most 16 units in unit finder order. `find_acquire_target` picks the first best score (`get_best_score` keeps the first on ties). `find_acquire_random_target` calls `lcg_rand(19)` only when a bucket has more than one unit.
- `loaded_units` (`L833`) returns a lazy filtered range. Return `iter.Seq`.
- `update_unit_status_timers` (`L1088`) and the `remove_*` effect functions call into C08 and C09 (images).
- `thingy_setter` / `make_thingy_setter` (`L364`–`L380`) are RAII. See the conventions above.
- `display_last_error_for_player` (`L17242`) is empty in C++. Keep it as a no-op.

## C02 Unit operations → `unit_ops.go`

Range: `L1935`–`L3723`.

Functions:
`stop_flingy:1935`, `stop_unit:1951`, `order_done:1962`, `set_next_target_waypoint:1969`, `set_unit_gathering:1974`, `any_neighbor_tile_unoccupied:1980`, `unit_pos_is_bordering_target:2012`, `move_to_target:2030`, `move_to_target_reset:2038`, `try_gather_resource:2043`, `wait_for_resource:2058`, `is_facing_next_target_waypoint:2070`, `long_path_distance:2076`, `unit_long_path_distance:2091`, `destroy_carrying_images:2096`, `unit_gather_resources_from:2102`, `gather_queue_next:2168`, `unit_is_active_resource_depot:2191`, `find_nearest_active_resource_depot:2202`, `drop_carried_items:2212`, `has_available_supply_for:2217`, `has_available_resources_for:2241`, `has_available_resources_for:2255`, `upgrade_mineral_cost:2269`, `upgrade_gas_cost:2272`, `upgrade_time_cost:2275`, `has_available_resources_for:2279`, `build_queue_push:2293`, `build_queue_clear_single:2309`, `train_unit:2316`, `refund_unit_costs:2323`, `partially_refund_unit_costs:2328`, `cancel_morphing_building:2333`, `cancel_building_unit:2383`, `cancel_build_queue:2430`, `cancel_build_queue:2443`, `place_building:2447`, `player_has_supply_and_resources_for:2475`, `get_building_at_center_position:2481`, `is_in_psionic_matrix_range:2490`, `is_in_psionic_matrix:2500`, `can_place_building:2508`, `kill_unit:2625`, `unit_can_enter_nydus:2634`, `unit_can_use_shield_battery:2644`, `unit_is_rescuable:2660`, `try_follow_unit:2666`, `follow_ahead_of_unit:2681`, `auto_attack_target:2694`, `attack_move_movement:2708`, `attack_move_acquire_target:2729`, `attack_unit_reacquire_target:2743`, `attack_unit_move_in_range:2785`, `unit_can_fire_weapon:2830`, `attack_unit_fire_weapon:2853`, `set_unit_owner:2882`, `set_sprite_owner:2921`, `reinitialize_unit_type:2926`, `morph_unit:2972`, `build_refinery:3034`, `unit_can_gather_gas:3050`, `find_gas_exit_position:3058`, `is_at_next_target_waypoint_or_within_attack_angle:3081`, `unit_load_target:3091`, `unit_unload_impl:3120`, `unit_unload:3154`, `trigger_give_unit_to:3182`, `give_unit_to:3228`, `make_unit_neutral:3417`, `building_abandon_addon:3425`, `create_dust_sprite:3435`, `unit_can_attach_addon:3454`, `find_connecting_addon:3471`, `connect_addon:3485`, `find_and_connect_addon:3496`, `cancel_research:3501`, `cancel_upgrade:3517`, `apply_upgrades_to_player_units:3534`, `set_unit_burrowed:3540`, `set_unit_cloaked:3553`, `crappy_move_to_target:3560`, `get_default_gather_order:3587`, `get_default_order:3602`, `default_action:3672`, `spell_cast_target_movement:3690`

Pitfalls:
- `long_path_distance` (`L2076`) iterates a `circular_vector` from the second element.
- `gather_queue_next` (`L2168`) iterates `reverse(ptr(...gather_queue))`.
- `can_place_building` (`L2508`–`L2624`) is long and depends on tile flags and the unit finder. Keep the check order: it decides which error code is stored.
- `set_unit_owner` (`L2882`) loops over `visible_units` and then `hidden_units`.
- `morph_unit` (`L2972`) uses `make_thingy_setter(iscript_unit, u->subunit)` inside an inner block (`L3021`).
- `give_unit_to` (`L3228`–`L3416`) has two recursive lambdas (`L3230`, `L3233`). It moves loaded units and carrier or reaver children. Keep the loop order over `inside_units` and `outside_units`.
- `create_dust_sprite` (`L3435`) creates thingies. They run iscript, which can call `lcg_rand`.
- `partially_refund_unit_costs` (`L2328`) computes `cost * 3 / 4`. Keep multiply before divide.

## C03 Orders: basic, workers, Terran → `orders_basic.go`

Ranges: `L3724`–`L3812`, `L3817`–`L3863`, `L4028`–`L5732`.

Functions:
`set_remove_timer:3724`, `default_remove_timer:3728`, `set_remove_timer:3734`, `make_unit_hallucination:3738`, `unit_can_see_order_target:3767`, `medic_can_heal_target:3772`, `find_medic_target:3783`, `medic_try_heal:3791`, `order_SelfDestructing:3817`, `copy_status_effects:3821`, `merge_status_effects:3844`, `refresh_unit_position:4028`, `units_share_unions:4051`, `interceptors_attack:4066`, `carrier_reaver_attack:4074`, `release_fighter:4151`, `create_hallucination:4186`, `order_Stop:4224`, `order_Guard:4231`, `order_PlayerGuard:4236`, `order_TurretGuard:4251`, `order_Move:4256`, `order_Nothing:4272`, `order_MoveToMinerals:4278`, `order_WaitForMinerals:4370`, `order_MiningMinerals:4380`, `order_ResetHarvestCollision:4414`, `order_ReturnGas:4447`, `order_ReturnMinerals:4451`, `order_Die:4521`, `order_GatherWaitInterrupted:4552`, `order_GatheringInterrupted:4561`, `order_PlaceBuilding:4569`, `order_IncompleteBuilding:4620`, `order_ConstructingBuilding:4642`, `order_ResetCollision:4742`, `order_Follow:4754`, `order_AttackMove:4807`, `order_AttackUnit:4820`, `order_MoveToGas:4827`, `order_WaitForGas:4868`, `order_HarvestGas:4882`, `order_Repair:4925`, `order_EnterTransport:5049`, `order_Unload:5071`, `order_BuildingLiftoff:5093`, `order_LiftingOff:5120`, `order_PlaceAddon:5158`, `order_BuildingLand:5202`, `order_ResearchTech:5296`, `order_Upgrade:5315`, `order_StayInRange:5339`, `order_TurretAttack:5372`, `order_Sieging:5397`, `order_Unsieging:5452`, `order_WatchTarget:5493`, `order_PickupIdle:5507`, `order_PickupTransport:5515`, `order_MoveUnload:5541`, `order_PlaceMine:5558`, `order_SpiderMine:5593`, `order_Critter:5681`, `order_HoldPosition:5710`

Pitfalls:
- `order_Guard` (`L4231`) sets `main_order_timer = lcg_rand(29, 0, 15)`.
- `order_ConstructingBuilding` (`L4642`) calls `lcg_rand(9, ...)` for x and then for y (`L4699`–`L4704`), then `lcg_rand(10)` (`L4722`). Keep this order.
- `order_MoveToMinerals` (`L4278`) and `order_Repair` (`L4925`) use lambdas with early returns.
- `order_Sieging`, `order_Unsieging` (`L5397`–`L5492`) set `iscript_unit` to the turret with `make_thingy_setter` inside inner blocks (`L5416`, `L5463`, `L5471`).
- `order_BuildingLiftoff` and `order_BuildingLand` set `frozen_y_value` from `get_image_map_position` (`L5112`, `L5226`). This is rendering only (see "Skip").
- `order_Unload` (`L5071`) checks `std::next(i) == end` to detect the last loaded unit.
- `release_fighter` (`L4151`) sets `best_hp` only for the first candidate, then compares every later unit against that first value. This is a Blizzard quirk. Keep it.

## C04 Orders: spells, Zerg, Protoss, carriers → `orders_advanced.go`

Range: `L5733`–`L7622`.

Functions:
`order_RightClickAction:5733`, `order_Patrol:5781`, `order_CastScannerSweep:5820`, `order_Scanner:5842`, `order_TowerGuard:5849`, `order_TowerAttack:5863`, `order_CastDefensiveMatrix:5872`, `order_Spell:5886`, `order_MoveToTargetOrder:5955`, `order_PickupBunker:5988`, `order_HealMove:5996`, `order_MedicHeal:6026`, `order_MedicHealToIdle:6058`, `order_MedicIdle:6066`, `order_MedicHoldPosition:6074`, `order_NukeTrain:6104`, `order_NukePaint:6122`, `order_NukeUnit:6127`, `order_CastNuclearStrike:6137`, `order_NukeLaunch:6168`, `order_NukeTrack:6226`, `order_Larva:6253`, `order_InitCreepGrowth:6293`, `order_ZergUnitMorph:6298`, `order_ZergBirth:6346`, `order_DroneStartBuild:6386`, `order_DroneLand:6391`, `order_DroneBuild:6462`, `order_IncompleteMorphing:6508`, `order_ZergBuildingMorph:6569`, `order_Burrowing:6586`, `order_Burrowed:6641`, `order_AttackFixedRange:6653`, `order_Unburrowing:6665`, `order_BuildNydusExit:6703`, `order_EnterNydusCanal:6724`, `order_CastInfestation:6762`, `order_InfestingCommandCenter:6785`, `order_InfestedCommandCenter:6820`, `order_SuicideUnit:6860`, `order_SuicideLocation:6898`, `order_SuicideHoldPosition:6924`, `order_QueenHoldPosition:6933`, `order_PlaceProtossBuilding:6937`, `order_IncompleteWarping:6997`, `order_InitializePsiProvider:7048`, `order_InitializeArbiter:7065`, `order_ArchonWarp:7076`, `order_CompletingArchonSummon:7123`, `order_CastRecall:7148`, `order_Carrier:7224`, `order_CarrierStop:7236`, `order_CarrierAttack:7243`, `order_InterceptorAttack:7262`, `order_InterceptorReturn:7336`, `order_DarkArchonMeld:7370`, `order_CastMindControl:7374`, `order_Reaver:7396`, `order_ReaverAttack:7415`, `order_ScarabAttack:7433`, `order_ReaverStop:7475`, `order_RechargeShieldsUnit:7485`, `order_RechargeShieldsUnitRemoveOverlay:7535`, `order_RechargeShieldsBattery:7542`, `order_Feedback:7555`, `order_CastHallucination:7583`, `order_Neutral:7610`

Pitfalls:
- `order_Critter` (`L5681`) calls `lcg_rand(48, ...)` up to two times, then `lcg_rand(47, ...)`.
- `order_CastRecall` (`L7148`): `play_sound(550 + lcg_rand(17) % 2, ...)` (`L7162`) and `play_sound(552 + lcg_rand(18) % 2, u)` (`L7218`). The sound is a no-op, but the `lcg_rand` call must stay.
- `order_InterceptorAttack` (`L7262`) has a lambda that calls `lcg_rand(11, -127, 128)` for x and then y (`L7287`–`L7288`). `L7597` (`order_CastHallucination`) calls `lcg_rand(32, 0, 32)`.
- `order_Larva` (`L6253`) uses a lambda (`L6273`).
- `order_Burrowed` (`L6641`) uses `queue_order(std::next(order_queue.begin()), ...)` (`L6647`): insert after the first queued order.
- `order_Burrowing` and `order_Unburrowing` iterate `u->sprite->images` (`L6613`, `L6691`).
- `order_CastRecall` loops over `visible_units` and then `active_bullets` (`L7192`, `L7195`).
- `order_NukeLaunch` and `order_NukeTrack` touch sprites and visibility.

## C05 Order dispatch, vision, frame loop → `order_dispatch.go`, `frame.go`

Ranges: `L7623`–`L8468`, `L12311`–`L13021`, `L13051`–`L13105`, `L13191`–`L13196`, `L13207`–`L13210`.

Functions:
`execute_main_order:7623`, `resume_building_unit:8153`, `rally_unit:8181`, `secondary_order_Train:8194`, `secondary_order_BuildAddon:8242`, `secondary_order_Cloak:8265`, `secondary_order_Decloak:8269`, `secondary_order_SpawningLarva:8273`, `secondary_order_SpreadCreep:8292`, `secondary_order_CloakNearbyUnits:8306`, `secondary_order_TrainFighter:8329`, `secondary_order_ShieldBattery:8384`, `execute_secondary_order:8407`, `update_unit:8448`, `reveal_sight_at:12311`, `refresh_unit_vision:12369`, `turn_turret:12387`, `update_unit_movement:12407`, `update_thingy_visibility:12439`, `update_unit_sprite:12465`, `unit_can_fire_from_bunker:12473`, `create_bunker_fire_animation:12480`, `try_attack_something:12502`, `order_hidden_BunkerGuard:12557`, `execute_hidden_unit_main_order:12572`, `execute_hidden_unit_secondary_order:12640`, `update_hidden_unit:12656`, `unit_can_detect:12672`, `unit_calculate_detected_flags:12680`, `remove_target_references:12716`, `remove_target_references:12753`, `update_unit_detected_flags:12758`, `update_psionic_matrix:12777`, `update_dead_unit:12801`, `update_disruption_web:12828`, `update_units:12879`, `update_bullets:12980`, `sprite_is_doodad:12998`, `update_thingy:13004`, `update_thingies:13015`, `active_players:13051`, `player_left:13057`, `player_dropped:13066`, `remove_player:13075`, `process_frame:13088`, `next_frame:13191`, `local_unit_status_error:13207`

Pitfalls:
- `execute_main_order` (`L7623`–`L8152`) is a 530-line switch over every order id. It calls all of C03 and C04. Port it first with stubs.
- `update_units` (`L12879`–`L12978`) is the main tick. Keep the exact pass order: order timers, disruption web, dead units, movement, map revealers vision, sprites and cloak counters, `update_unit` on visible units, hidden units, cloaked units, psionic matrix, lurker hit ring, map revealers. Each pass sets `iscript_flingy` and `iscript_unit`. Some passes take the next pointer before the body (`&*i++`) and some use range-for. Copy each one.
- `update_bullets` (`L12980`) sets `iscript_bullet` and `iscript_flingy`.
- `process_frame` (`L13088`) and `next_frame` (`L13191`): `update_tiles` is true one frame in 100. On those frames all tiles `visible` are reset to `0xff`.
- `reveal_sight_at` (`L12311`) packs height and visibility masks into `uint32_t` (`L12322`) and casts positions to `size_t`. Keep the unsigned types.
- `try_attack_something` (`L12502`) calls `lcg_rand(12)` for cooldown jitter (`L12550`).
- `update_unit` (`L8448`) runs the subunit with `make_thingy_setter` (`L8456`).
- `update_tiles` is a member of each `StateFunctions` instance. OpenBW uses one `game_load_functions` instance for loading and one `replay_functions` (an `action_functions`) instance for actions and frames. Mirror this: one Go instance for loading, one for playback.

## C06 Flingy movement and movement state machine → `movement.go`

Ranges: `L8469`–`L9161`, `L10933`–`L12203`, `L13359`–`L13426`.

Functions:
`unit_is_at_move_target:8469`, `unit_dead:8473`, `unit_dying:8478`, `rectangle_can_fit_at:8483`, `unit_type_can_fit_at:8497`, `can_fit_at:8508`, `unit_target_is_enemy:8542`, `get_largest_blocking_unit:8548`, `is_blocked:8563`, `set_flingy_move_target:8571`, `set_unit_move_target:8580`, `set_unit_move_target:8594`, `set_current_velocity_direction:8608`, `set_current_speed:8613`, `set_next_speed:8619`, `unit_turn_rate:8624`, `set_desired_velocity_direction:8634`, `update_current_velocity_direction_towards_waypoint:8644`, `update_unit_heading:8656`, `update_current_speed_towards_waypoint:8680`, `struct execute_movement_struct:8749`, `set_movement_flags:8760`, `set_movement_values:8776`, `update_unit_movement_values:8797`, `check_unit_movement_terrain_collision:8807`, `check_unit_movement_terrain_collision:8981`, `check_unit_movement_unit_collision:8989`, `check_ground_unit_movement_unit_collision:9051`, `update_repulse_direction:9059`, `apply_repulse_field:9084`, `finish_flingy_movement:9121`, `finish_unit_movement:9139`, `is_moving_along_path:9147`, `set_unit_immovable:10933`, `update_unit_pathing_collision:10941`, `cardinal_direction_from_to:10951`, `collision_get_slide_free_direction:10967`, `reset_movement_state:11024`, `movement_UM_Init:11031`, `movement_UM_AtRest:11062`, `get_unique_sided_positions_within_bounds:11123`, `movement_UM_CheckIllegal:11175`, `movement_UM_Dormant:11382`, `movement_UM_Turret:11394`, `movement_UM_Flyer:11413`, `movement_UM_AnotherPath:11441`, `movement_UM_StartPath:11447`, `movement_UM_FollowPath:11465`, `movement_UM_ScoutPath:11530`, `movement_UM_AtMoveTarget:11537`, `movement_UM_NewMoveTarget:11551`, `movement_UM_Repath:11563`, `movement_UM_UIOrderDelay:11576`, `movement_UM_FixTerrain:11586`, `movement_UM_FailedPath:11628`, `movement_UM_TerrainSlide:11635`, `movement_UM_TurnAndStart:11669`, `movement_UM_FaceTarget:11691`, `movement_UM_RetryPath:11708`, `movement_UM_InitSeq:11723`, `movement_UM_ForceMoveFree:11729`, `movement_UM_FixCollision:11741`, `movement_UM_WaitFree:11852`, `movement_UM_SlidePrep:11881`, `movement_UM_SlideFree:11893`, `movement_UM_RepathMovers:11950`, `movement_UM_GetFree:11963`, `movement_UM_Lump:11985`, `movement_UM_MoveToLegal:11989`, `movement_UM_Hidden:12038`, `movement_UM_Bunker:12043`, `movement_UM_LumpWannabe:12051`, `movement_UM_BldgTurret:12078`, `execute_movement:12085`, `get_modified_unit_speed:13359`, `get_modified_unit_acceleration:13378`, `get_modified_unit_turn_rate:13389`, `get_modified_weapon_cooldown:13400`, `unit_halt_distance:13416`

Pitfalls:
- `unit_turn_rate` (`L8624`) and `update_current_speed_towards_waypoint` (`L8680`) mix signed and unsigned fixed point (`L8625`–`L8627`, `L8728`–`L8729`).
- `get_modified_unit_speed`, `get_modified_unit_acceleration`, `get_modified_unit_turn_rate`, `unit_halt_distance` (`L13359`–`L13426`) use `ufp8` (`as_unsigned`). Keep the unsigned math.
- `check_unit_movement_terrain_collision` (`L8807`–`L8980`) has three lambdas and swaps min/max.
- `movement_UM_CheckIllegal` (`L11175`–`L11381`) has many `lcg_rand(50, ...)` calls inside nested lambdas (`L11196`, `L11266`–`L11279`, `L11301`–`L11302`, `L11336`). Keep the call order and the early returns.
- `movement_UM_StartPath` (`L11453`) and `movement_UM_WaitFree` (`L11864`) call `lcg_rand`.
- `execute_movement` (`L12085`–`L12203`) is a switch over the movement states. The `UM_*` functions return `bool` "continue the loop".
- `can_fit_at` (`L8508`) and `check_unit_movement_unit_collision` (`L8989`) use the unit finder. A search must finish (visited flags reset) before the unit finder changes.
- `increment_repulse_field` (in C10, `L17003`) calls `lcg_rand(37)`. C06 calls it.

## C07 Pathfinder → `pathfinder.go`

Range: `L9162`–`L10932` (the `pathfinder` struct `L9162`–`L9190` moves to C00).

Functions:
`struct pathfinder:9162`, `pathfinder_find_long_path:9191`, `binary_heap_up:9374`, `binary_heap_down:9386`, `pathfinder_unit_can_collide_with:9406`, `pathfinder_unit_can_collide_with:9417`, `pathfinder_find_short_path:9421`, `pathfinder_adjust_target_pos:10562`, `pathfinder_adjust_destination:10596`, `pathfinder_find_next_short_path:10614`, `pathfinder_find_long_path:10710`, `pathfinder_find:10729`, `new_path:10751`, `free_path:10763`, `free_path:10769`, `path_progress:10773`, `create_single_step_path:10861`, `unit_path_to:10890`, `unit_update_path_movement_state:10908`

Pitfalls:
- `binary_heap_up` and `binary_heap_down` (`L9373`–`L9404`) are a custom heap. Do not use `container/heap`: tie order must match. `std::find` on the heap (`L9287`, `L10455`) finds the first match.
- `pathfinder_find_long_path` (`L9191`–`L9372`) caps the open list at 125 and all nodes at 350. It stores a node pointer in `region->pathfinder_node` (a `void*` on a const region). In Go, add a field on `Region` or a side slice by region index, and clear it the same way.
- `pathfinder_find_short_path` (`L9421`–`L10561`) is one 1,140-line function with about 30 lambdas, `static_vector` caps (32 neighbors, 150 open, 250 nodes) and linked-list edits with `std::prev`/`std::next` (`L10224`–`L10260`). Port it as one function with closures.
- `a_circular_vector` for `long_path` and `short_path`.
- `new_path` and `free_path` (`L10751`–`L10772`) use the `free_paths` intrusive list and the `paths` `std::list`. Reuse order matters.
- `st.consider_collision_with_unit_bug` is pointer state that reproduces a Blizzard bug. Keep it.

## C08 Combat, spells, bullets → `combat.go`, `bullets.go`

Ranges: `L13435`–`L14535`, `L14573`–`L14637`.

Functions:
`is_spell:13435`, `weapon_damage_amount:13463`, `bullet_damage_amount:13467`, `unit_sprite_size:13477`, `create_sized_image:13483`, `create_defensive_matrix_image:13488`, `deal_defensive_matrix_damage:13495`, `unit_armor:13509`, `on_hit_should_change_target:13516`, `unburrow_unit:13527`, `unit_can_move_from_to_region:13535`, `on_hit_change_target:13550`, `on_hit_aoe_change_target:13586`, `on_unit_damage:13601`, `unit_deal_damage:13621`, `create_shield_damage_effect:13651`, `weapon_deal_damage:13658`, `hallucinated_weapon_hit:13720`, `bullet_deal_damage:13735`, `bullet_deal_damage:13742`, `melee_deal_damage:13746`, `weapon_can_target_unit:13757`, `weapon_can_target_unit:13768`, `bullet_deal_splash_damage:13775`, `irradiate_unit:13824`, `return_interceptors:13833`, `set_unit_disabled:13840`, `lockdown_unit:13853`, `blind_unit:13861`, `restore_unit:13872`, `emp_shockwave:13888`, `spawn_broodlings:13903`, `ensnare_unit:13925`, `ensnare:13933`, `consume_unit:13946`, `dark_swarm:13956`, `plague_unit:13969`, `plague:13976`, `parasite_unit:13986`, `stasis_unit:13990`, `stasis_field:13999`, `maelstrom_unit:14014`, `maelstrom:14022`, `disruption_web:14043`, `acid_spore_image:14056`, `update_acid_spore_image:14061`, `add_acid_spore:14069`, `add_acid_spore:14075`, `bullet_hit:14086`, `bullet_kill:14200`, `bullet_move:14205`, `set_bullet_move_target:14217`, `bullet_state_init:14222`, `bullet_state_dying:14258`, `bullet_state_move:14266`, `bullet_state_follow:14274`, `bullet_state_hit_near_target:14286`, `bullet_state_bounce:14300`, `bullet_state_damage_over_time:14331`, `bullet_execute:14340`, `unit_is_under_dark_swarm:14372`, `unit_dodge_chance:14380`, `unit_target_miss_chance:14387`, `initialize_bullet:14397`, `create_bullet:14494`, `get_bullet_appear_at_target_pos:14508`, `fire_weapon:14519`, `unit_tech_target_valid:14573`, `spell_order_valid:14611`

Pitfalls:
- `bullet_deal_splash_damage<bool is_air_splash>` (`L13774`) is a template on a bool. A runtime bool is fine. It calls `lcg_rand(58)` when there is more than one target (`L13793`).
- `bullet_hit` (`L14086`) checks `recent_lurker_hits` with `std::find` on pairs of unit indices (`L14114`).
- `initialize_bullet` (`L14397`): `lcg_rand(0) & 1` only when the source unit differs from `st.prev_bullet_source_unit` (`L14415`). Miss checks use `lcg_rand(1)` behind `&&` (`L14448`, `L14456`).
- `bullet_state_bounce` (`L14300`) has a lambda (`L14312`).
- `unit_deal_damage`, `weapon_deal_damage` (`L13621`–`L13719`) use `fp8` and integer division. Keep the operation order.
- `unit_target_miss_chance`, `unit_dodge_chance` (`L14380`–`L14396`) return `fp8`.

## C09 Iscript, images, sprites, thingies → `iscript.go`, `sprites.go`

Ranges: `L13211`–`L13358`, `L13427`–`L13434`, `L14536`–`L14572`, `L14638`–`L15369`.

Functions:
`get_sprite_tile_line_index:13211`, `add_sprite_to_tile_line:13217`, `remove_sprite_from_tile_line:13221`, `move_sprite:13226`, `set_sprite_visibility:13237`, `set_image_offset:13242`, `set_image_modifier:13248`, `show_image:13257`, `hide_image:13263`, `update_image_special_offset:13268`, `update_image_frame_index:13272`, `set_image_frame_index_offset:13280`, `set_image_heading:13287`, `set_image_heading_by_index:13302`, `set_sprite_images_heading_by_index:13316`, `set_sprite_images_heading_by_image_index:13322`, `get_image_center_map_position:13330`, `get_image_map_position:13334`, `get_image_lo_offset:13347`, `iscript_set_script:13427`, `new_thingy:14536`, `initialize_thingy:14546`, `create_thingy:14553`, `create_thingy_at_image:14565`, `iscript_execute:14638`, `iscript_run_anim:15136`, `image_update_cloak:15159`, `image_update_decloak:15174`, `image_update_warpin:15189`, `iscript_execute_sprite:15201`, `sprite_run_anim:15221`, `initialize_image:15228`, `destroy_image:15258`, `enum (anon):15264`, `create_image:15270`, `destroy_sprite:15297`, `create_sprite:15306`, `initialize_flingy:15342`

Pitfalls:
- `iscript_execute` (`L14638`–`L15135`) is the interpreter. It reads `global_st.iscript.program_data`, an `int` program that `global_init` builds (decoder lambda from `L22095`; each opcode is stored as `opc + 0x808091`). The data phase must port that decoder.
- It reads and writes `iscript_unit`, `iscript_flingy`, `iscript_bullet`. It has eight lambdas (`play_frame`, `add_image`, `use_weapon`, `attack_with*`, `playsndrand`).
- `lcg_rand` calls inside the interpreter run only when `noop` is false: `L14710` (`playsndrand`, source 4), `L14765` (`waitrand`, source 3), `L14876` (`playsndbtwn`, source 5), `L14924` (source 6). `randcondjmp` (`L14899`–`L14905`, source 7) has no `noop` check: it always calls `lcg_rand(7)`. Port each `noop` check exactly as it is.
- `iscript_execute_sprite` (`L15201`) takes the next image before running the current one. An image can delete itself, and the main image moves to `images.front()`.
- `create_image` (`L15270`) inserts at front, back, above or below a related image. It panics if the `Init` animation destroys the image.
- `create_sprite` (`L15306`) copies GRP width and height (capped at 255). `update_thingy_visibility` (C05) uses them, so GRP frame sizes are gameplay data.
- `sprites_container`, `images_container`: pool order (`top`, `pop`, `push`) decides reuse.

## C10 Unit lifecycle, unit finder, order queue → `units.go`, `unit_finder.go`

Range: `L15370`–`L17241`.

Functions:
`update_unit_speed_upgrades:15370`, `update_unit_speed:15415`, `increment_unit_counts:15452`, `unit_finder_insert:15486`, `unit_finder_reinsert:15493`, `unit_finder_remove:15499`, `unit_finder_insert:15517`, `unit_finder_reinsert:15532`, `struct unit_finder_search:15580`, `find_units:15681`, `find_units_noexpand:15685`, `find_unit:15690`, `find_unit_noexpand:15698`, `find_nearest_unit:15706`, `find_nearest_unit:15795`, `find_nearest_unit:15807`, `unit_is_factory:15827`, `set_unit_tiles_occupied:15843`, `set_unit_tiles_unoccupied:15849`, `initialize_unit_type:15855`, `kill_interceptors:16006`, `destroy_unit_impl:16022`, `destroy_unit:16220`, `initialize_unit:16233`, `create_unit:16317`, `create_unit:16360`, `replace_sprite_images:16365`, `apply_unit_effects:16378`, `damage_overlay_states:16434`, `set_construction_graphic:16443`, `set_unit_heading:16465`, `finish_building_unit:16477`, `find_unit_placement:16501`, `find_unit_placement:16692`, `move_unit:16696`, `place_completed_unit:16711`, `add_completed_unit:16731`, `remove_queued_order:16771`, `create_order:16779`, `queue_order:16789`, `queue_order_back:16799`, `queue_order_front:16803`, `remove_one_order:16807`, `set_queued_order:16816`, `issue_order:16831`, `iscript_run_to_idle:16859`, `activate_next_order:16893`, `set_unit_order:16935`, `set_unit_order:16941`, `set_unit_order:16947`, `set_unit_order:16958`, `is_intersecting:16966`, `unit_finder_units_intersecting:16974`, `unit_finder_unit_in_bounds:16978`, `check_unit_collision:16982`, `repulse_index:16995`, `increment_repulse_field:17003`, `decrement_repulse_field:17020`, `show_unit:17028`, `hide_unit:17045`, `set_sprite_cloak_modifier:17073`, `cloak_unit:17107`, `decloak_unit:17141`, `complete_unit:17174`, `create_initial_unit:17215`, `create_completed_unit:17234`

Pitfalls:
- Unit finder (`L15486`–`L15826`): two sorted arrays of `{unit, value}`. Insert uses `lower_bound`, so a new entry goes before equal values. `unit_finder_reinsert` shifts entries by hand. Keep it exact: search order decides target choice.
- `unit_finder_search` (`L15580`–`L15680`) is an RAII iterator. It increments `unit_finder_search_index` (max depth 4), marks `unit_finder_visited[search_index]`, and clears the marks in the destructor, also on early `return` from the loop. In Go use `iter.Seq[*Unit]` and clear the marks with `defer` inside the iterator function. C++ increments the index when the search object is built, before the loop. Go increments it when the loop starts. This is the same as long as the result is always ranged over.
- `find_nearest_unit` (`L15706`–`L15826`) walks four iterators out from a position. Use indices into the arrays.
- `create_unit` (`L16317`) always calls `lcg_rand(14)` first (`L16320`). `initialize_unit` (`L16233`) calls `lcg_rand(15)` for `wireframe_randomizer` (`L16280`). `kill_interceptors` calls `lcg_rand(39, 15, 45)` (`L16015`). `finish_building_unit` area `L16491` calls `lcg_rand(36)`. `increment_repulse_field` calls `lcg_rand(37)` (`L17008`).
- `initialize_unit` uses placement `new` to reset `build_queue` (`L16277`–`L16278`).
- `destroy_unit_impl` (`L16022`–`L16219`) calls `on_unit_destroy` (`L16208`). Keep the step order.
- `find_unit_placement` (`L16501`–`L16691`) has two lambdas and a spiral search.
- `queue_order` (`L16789`) takes an intrusive list iterator. In Go pass the node to insert before.
- `create_order` uses `orders_container`. `local_unit_status_error(u, 872)` (`L16792`) is a no-op.
- `cloak_unit` and `decloak_unit` (`L17107`–`L17173`) have lambdas over sprite images.

## C11 Tech tree and order requirements → `requirements.go`

Range: `L17324`–`L18784`.

Functions:
`unit_can_build:17324`, `unit_can_receive_order:17750`, `unit_order_allowed:17767`, `unit_build_order_valid:18173`, `unit_can_research:18198`, `unit_can_research:18203`, `unit_can_upgrade:18333`, `unit_can_upgrade:18338`, `unit_can_use_tech:18640`, `unit_can_use_tech:18645`

Pitfalls:
- Five very long functions with large `switch` statements: `unit_can_build` (`L17324`–`L17749`), `unit_order_allowed` (`L17767`–`L18172`), `unit_can_research` (`L18203`–`L18332`), `unit_can_upgrade` (`L18338`–`L18639`), `unit_can_use_tech` (`L18645`–`L18784`).
- Pure functions of state. Easy to unit test against the C++ harness.

## C12 Triggers and creep → `triggers.go`, `creep.go`

Ranges: `L3813`–`L3816`, `L3864`–`L4027`, `L13022`–`L13050`, `L13106`–`L13190`, `L18785`–`L19553`.

Functions:
`tile_has_creep:3813`, `creep_is_being_sustained:3864`, `set_creep_receding:3890`, `remove_creep_provider:3907`, `remove_creep_provider:3932`, `remove_extractor_creep:3936`, `add_creep_provider:3956`, `zerg_building_start_construction:3986`, `set_creep_building_tiles:3994`, `recede_creep:13022`, `process_triggers:13106`, `trigger_players:18785`, `trigger_command_count:18792`, `trigger_count_comparison:18812`, `struct bring_unit_counters:18819`, `trigger_bring_count:18832`, `trigger_opponent_count:18879`, `test_trigger_condition:18887`, `trigger_create_unit:18905`, `player_won:18950`, `player_defeated:18954`, `player_slot_active:18959`, `trigger_players_pred:18970`, `unit_is_at_elevation_flags:18989`, `trigger_unit_pred:19003`, `trigger_find_unit:19029`, `struct execute_trigger_struct:19037`, `execute_trigger_action:19041`, `execute_trigger:19307`, `tile_can_have_creep:19325`, `get_max_creep_bb:19333`, `count_neighboring_creep_tiles:19354`, `set_tile_creep:19387`, `spread_creep:19444`, `spread_creep_completely:19480`, `get_spawn_larva_position:19497`, `spawn_larva:19533`

Pitfalls:
- `recede_creep` (`L13022`) picks a random list element with `std::advance(it, lcg_rand(27, 0, size - 1))` on `creep_life.lists`.
- `spread_creep` (`L19444`) uses `std::array<static_vector<size_t, 240>, 8>` and `lcg_rand(26)` (`L19473`).
- `set_tile_creep` (`L19387`) and `count_neighboring_creep_tiles` (`L19354`) use lambdas and unsigned tile coordinates.
- `trigger_bring_count` (`L18832`) uses a recursive `std::function` (`L18842`) for loaded units. `execute_trigger_action` (`L19041`–`L19306`) has a recursive `std::function` (`L19149`).
- `process_triggers` (`L13106`) decides victory and defeat and calls `on_victory_state` and `on_player_eliminated`. Melee games use `melee.trg` (loaded by C13).
- `trigger_command_count` (`L18792`) is a template over a count array.
- `spawn_larva` and `get_spawn_larva_position` (`L19497`–`L19552`) are here because they belong to creep and hatcheries.

## C13 Map loading (CHK) and region creation → `map_load.go`, `regions_create.go`

Range: `L19764`–`L21964` (includes the free function `read_grp` `L21916`–`L21964`).

Functions:
`struct setup_info_t:19770`, `get_unit_type:19781`, `get_weapon_type:19785`, `get_upgrade_type:19789`, `get_tech_type:19793`, `get_flingy_type:19797`, `reset:19802`, `get_new_region:20026`, `regions_create:20035`, `get_unit_strength:20828`, `calculate_unit_strengths:20881`, `generate_sight_values:20912`, `load_tile_stuff:21080`, `create_starting_unit:21144`, `create_starting_units:21152`, `get_map_string:21196`, `struct tag_t:21206`, `load_map_file:21219`, `load_map:21224`, `load_map_data:21230`, `read_grp:21917`

Pitfalls:
- `reset` (`L19802`–`L20025`) copies type tables from `global_st` into `game_st` and fixes type pointers with lambdas (`L19809`–`L19837`). It sets `lcg_rand_state = 42`.
- `get_new_region` (`L20026`) reserves 5,000 regions so region pointers stay valid. In Go, allocate the slice with capacity 5,000 and never grow it past that.
- `regions_create` (`L20035`–`L20827`) has about 20 lambdas. `std::sort` of contours (`L20797`) compares only `v[0]` and `v[1]`. A `contour` (`game_types.h` L287) also has `v[2]`, `dir` and `flags`. `std::sort` is not stable, so two contours with equal keys can come out in either order. Add a test-only check that reports such ties on real maps. If ties exist, match the standard library that the C++ harness uses.
- Region neighbor lists use `std::find` before `push_back` (`L20419`–`L20423`). Order of insertion matters.
- `get_unit_strength` (game load, `L20828`) uses `unsigned` and `int64_t` math and a `sqrt(x) * 7.58` integer approximation (from `L20847`). Keep it integer.
- `generate_sight_values` (`L20912`–`L21079`) has `static const` direction arrays (`L21011`–`L21012`).
- `load_map_data` (`L21230`–`L21914`): chunks are read in a fixed tag order per map version. Each tag can appear many times, and each copy is applied in file order. A chunk whose length is larger than the remaining data stops the scan (`L21247`).
- `setup_f` (`L21797`) is a callback in the middle of loading. The replay layer uses it to set players and races.
- `UNIT` invalid units call `lcg_rand(14)` (`L21625`).
- `load_map_data` ends with two `process_frame` calls (`L21908`–`L21911`).
- `read_grp` may already be ported by the data phase (PLAN phase 1). Reuse it if so.

## C14 Player actions → `actions.go`

Range: `actions.h` `A1`–`A1478` (all of it).

Functions:
`unit_can_be_multi_selected:43`, `selected_units:62`, `control_group:68`, `get_first_selected_unit:76`, `get_single_selected_unit:83`, `on_unit_deselect:91`, `struct group_move_t:99`, `calc_group_move:107`, `get_group_move_pos:183`, `morph_archon_impl:231`, `action_select:268`, `action_select:286`, `action_shift_select:291`, `action_shift_select:309`, `action_deselect:314`, `action_deselect:329`, `action_train:333`, `action_build:344`, `action_default_order:368`, `action_order:438`, `action_stop:545`, `action_cancel_building_unit:556`, `action_cancel_build_queue:563`, `action_control_group:581`, `action_unload_all:643`, `action_unload:653`, `action_liftoff:660`, `action_research:672`, `action_cancel_research:687`, `action_upgrade:697`, `action_cancel_upgrade:713`, `action_unsiege:723`, `action_siege:736`, `action_return_cargo:749`, `action_hold_position:763`, `action_cloak:779`, `action_decloak:796`, `action_set_alliances:810`, `action_set_shared_vision:822`, `action_cancel_addon:829`, `action_stim_pack:842`, `action_cancel_nuke:859`, `action_morph:871`, `action_morph_building:888`, `action_burrow:902`, `action_unburrow:914`, `action_cancel_morph:925`, `action_morph_archon:934`, `action_carrier_stop:938`, `action_train_fighter:948`, `action_morph_dark_archon:967`, `action_reaver_stop:971`, `action_player_leave:981`, `action_cheat:987`, `read_action_select:1005`, `read_action_shift_select:1017`, `read_action_deselect:1029`, `read_action_train:1041`, `read_action_default_order:1047`, `read_action_player_leave:1058`, `read_action_build:1064`, `read_action_stop:1073`, `read_action_order:1079`, `read_action_cancel_building_unit:1092`, `read_action_cancel_build_queue:1097`, `read_action_control_group:1103`, `read_action_unload_all:1111`, `read_action_unload:1117`, `read_action_liftoff:1123`, `read_action_research:1130`, `read_action_cancel_research:1136`, `read_action_upgrade:1141`, `read_action_cancel_upgrade:1147`, `read_action_unsiege:1152`, `read_action_siege:1158`, `read_action_chat:1164`, `read_action_return_cargo:1175`, `read_action_hold_position:1181`, `read_action_cloak:1187`, `read_action_decloak:1193`, `read_action_set_alliances:1199`, `read_action_set_shared_vision:1210`, `read_action_cancel_addon:1216`, `read_action_stim_pack:1221`, `read_action_cancel_nuke:1226`, `read_action_ping_minimap:1231`, `read_action_morph:1238`, `read_action_morph_building:1244`, `read_action_burrow:1250`, `read_action_unburrow:1256`, `read_action_cancel_morph:1262`, `read_action_morph_archon:1267`, `read_action_carrier_stop:1272`, `read_action_train_fighter:1277`, `read_action_morph_dark_archon:1282`, `read_action_reaver_stop:1287`, `read_action_cheat:1292`, `read_action_ext_cheat:1298`, `on_action:1336`, `read_action:1340`, `read_action:1348`, `read_action:1446`, `read_action:1451`, `execute_actions:1456`

Pitfalls:
- `action_state` (`A8`–`A22`): selections are `static_vector<unit_t*, 12>` per player, and control groups store `unit_id`.
- `on_unit_deselect` (`A91`) overrides the `StateFunctions` hook. It is gameplay: it removes dead units from selections.
- `action_stim_pack` (`A842`) calls `play_sound(lcg_rand(31, 278, 279), u)` (`A848`). Keep the `lcg_rand` call.
- `calc_group_move` (`A107`) and `get_group_move_pos` (`A183`) decide formation move targets. They use lambdas (`A133`).
- `action_default_order` (`A368`) uses `find_unit_noexpand` with a lambda (`A373`).
- `read_action_*` (`A1004`–`A1335`) are templates on the reader. In Go use one byte reader type.
- `read_action(owner, r)` (`A1348`) switches on the action id. Unknown ids panic. The replay translator must drop them first (PLAN).
- `execute_actions` (`A1456`) reads `int32 frame, uint8 size, bytes` records. PLAN's translator should write this format.
- `copy_state` (`A24`–`A37`) is not needed.

## Skip or port elsewhere

These parts are for rendering, sound, UI or other tools. A replay engine does not need them.

| Part | Lines | What to do |
|---|---|---|
| `play_sound` hooks | `L334`, `L618`–`L625` | No-op hooks. Keep every `lcg_rand` in their arguments: `L7162`, `L7218`, `L14710`, `L14876`, `A848`. |
| `on_unit_destroy`, `on_kill_unit`, `on_player_eliminated`, `on_victory_state`, `on_action` | `L337`–`L341`, `A1336` | Keep as hooks. The API can use them for deaths and results. |
| `local_unit_status_error`, `display_last_error_for_player` | `L13207`, `L17242` | Empty in C++. Keep as no-ops. |
| Sprite tile lines | `L13211`–`L13224`, list moves in `move_sprite` `L13226` | Draw order only. Optional; it is small, so porting it is also fine. `move_sprite` must still set the position. |
| `flag_redraw` writes | `L13245`, `L13254`, `L13260`, `L13276`, `L14738`, `L14747` | Rendering only. Harmless to keep. |
| `get_image_map_position`, `frozen_y_value` | `L13334`–`L13345`, `L5112`, `L5226`, `L6422` | Rendering only. Optional. |
| `get_map_string` and Korean decoding | `L21196`–`L21204` | Trigger text only. Return raw bytes. |
| `load_map_file`, `load_map` (MPQ) | `L21219`–`L21228` | Skip. The replay's CHK bytes go to `load_map_data`. |
| `state_copier`, `copy_state` | `L19556`–`L19761`, `A24`–`A37` | Skip. Port later only if the API needs snapshots. |
| `string_table_data` | `L21966`–`L21980` | Skip. |
| `global_init` | `L21982`–`L22336` | Data phase (PLAN phase 1), not this map. It builds `iscript.program_data`, image GRPs, lo offsets and `melee_trg`, which C09 and C13 need. |
| `game_player` | `L22338`–`L22394` | Replaced by the Go API. |
| `read_action_chat`, `read_action_ping_minimap` | `A1164`, `A1231` | Read and skip the bytes only. |

Do not skip these, even though they look like UI:

- Images, sprites, thingies and their iscript: they call `lcg_rand`, use pool slots and drive attack timing.
- Tile `visible` and `explored`, sprite `visibility_flags`: targeting checks read them (`L1462`).
- GRP width and height: used by `update_thingy_visibility`.
- Selections and control groups in C14: they decide which units get each command.
