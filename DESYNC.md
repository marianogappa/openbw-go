# SC:R vs OpenBW desync

Investigation of the desyncs in `INTEGRATION.md`: C++ OpenBW (`ref/openbw`, unmodified) playing downgraded SC:R replays. The Go port is not involved. Run date: 2026-10-14.

## Result

| Finding | Status | Effect |
|---|---|---|
| OpenBW keeps the 1.16 object limits. SC:R replays declare larger ones in `LMTS`. The bullet limit (100 vs 400) is reached in big fights, so OpenBW drops attacks that SC:R fires. | **Proven** | Terran desyncs drop from 21/149 to 13/149. Not one replay gets worse by the tag test. |
| Disabled `THG2` units (Pole Star 1.1) are doodad doors. SC:R creates them, sets the disabled doodad state, and units walk through them. | **Proven** | Pole Star: 30/30 replays load and stay in sync. Before: 30/30 fail at frame 0. |
| The remaining Terran desyncs are mostly on one map: **Attitude 1.0**. 8 of 18 Terran games there desync, against 5 of 131 Terran games on the other 12 maps. MM-029782AE is one of them. | **Root cause not found** | See "Remaining desyncs". |
| The INTEGRATION.md Siege/Unsiege lead | **Rejected** | The lead was the bullet limit. Tank games have the most bullets. Of the 4 remaining non-Attitude Terran desyncs checked, 3 have no Siege command before the onset. |
| SC:R order allocation rule behind `BFIX` bit 2 | **Real, but no effect in the sample** | The diff makes no difference on 117 Terran replays. See "BFIX". |

## Method

The ground truth is the command stream. In SC:R every tag a player selects exists. A multi-unit selection contains only the player's own units. `tools/desync/tracer` replays a downgraded replay on any OpenBW copy. At each Select or Shift Select it checks each tag against the engine:

- the tag does not exist, or
- the selection has more than one unit and a unit belongs to another player.

Each such command is "inconsistent". Clean games have 0-10 inconsistent commands, from units that die between the click and the command. A game with 20 or more is desynced (the "tag test"). On the 117 Terran baseline replays, the tag test and the `tools/integration` verdict agree on 13 of 14 desynced replays. Each flags one replay the other does not: MM-342E0750 (20 inconsistent commands) and MM-A20D6A70. It also finds the first divergent frame, which is usually minutes before the integration onset.

The test is very sensitive. `tracer -rng F N` moves `lcg_rand` by N steps at frame F. One step before frame 3500 of MM-029782AE causes an inconsistency by frame 3929. Clean games stay consistent to the end, so in them OpenBW's state and RNG match SC:R exactly. A candidate fix that is wrong breaks clean games at once. Two examples: zeroing reused path objects breaks 82 more replays, and never retrying SCV construction spark positions breaks 55 more.

## Root cause 1: object limits (bullets)

Every replay in the sample has `LMTS` = images 10000, sprites 5000, lone sprites 1000, units 3400, bullets 400, orders 4000, fog sprites 1000. It also has `GCFG` build 13515 and `BFIX` `04 00 00 00 00 00 00 00`. OpenBW uses the 1.16 limits: bullets 100, sprites 2500, images 5000, orders 2000, lone sprites (thingies) 500, units 1700. In big fights OpenBW reaches 100 bullets, and `create_bullet` fails silently. SC:R fires those shots.

Evidence: the probe's peak bullet count is 93-100 in 7 of the 14 desynced baseline replays. Raising only the bullet limit to 400 makes 6 of those 14 consistent: MM-13376D90, MM-43C15566, MM-4B4403AA, MM-857B7720, MM-F5ADC66C and MM-F62EF5D2, with 253, 492, 157, 1043, 20 and 93 inconsistent commands before. In the wide sample it does the same for MM-5139D282 and MM-9C050330. With all limits raised the results are identical. Only bullets matter in this sample.

## Root cause 2: disabled THG2 units (Pole Star 1.1)

Pole Star 1.1 has 95 neutral `THG2` units of type 205/206 (Pit Door, Right Pit Door) with flags `0xa280`. Bit 15 (`0x8000`) is "disabled". `tools/desync/mapinfo` lists them. OpenBW stops with `disable thingy unit`.

Tested on 30 Pole Star replays (with the bullet fix):

| Handling of disabled THG2 units | Desynced (tag test) |
|---|---|
| Skip them | 29/30, from frame 3-17. They take unit indexes in SC:R. |
| Create them as normal units | 26/30, from frame 3339-8356 |
| Create them and set `status_flag_disabled` (0x400, BW's doodad-state bit) | 26/30, the same frames |
| Create them and set `status_flag_no_collide` | **0/30** |
| Create them and set both flags | **0/30** |

So SC:R creates the units, which takes indexes, and puts them in the disabled doodad state. In that state a door is open: it has no unit collision. Setting `no_collide` is necessary and sufficient. The `disabled` bit makes no measurable difference here, but it is the bit BW uses for the doodad state, so the diff sets both. The doors' order processing (`OpenDoor`, `CloseDoor`, `EnableDoodad`, `DisableDoodad`) is still unimplemented in OpenBW. Nothing in these games issues those orders.

## Exact change (minimal diff against `ref/openbw`)

```diff
--- ref/openbw/bwgame.h
+++ bwgame.h
@@ -299,14 +299,14 @@
 	object_container<unit_t, 1700, 17> units_container;
 
 	intrusive_list<bullet_t, default_link_f> active_bullets;
-	object_container<bullet_t, 100, 10> bullets_container;
+	object_container<bullet_t, 400, 10> bullets_container;
 
 	a_vector<intrusive_list<sprite_t, default_link_f>> sprites_on_tile_line;
-	object_container<sprite_t, 2500, 25> sprites_container;
-
-	object_container<image_t, 5000, 50> images_container;
+	object_container<sprite_t, 5000, 25> sprites_container;
 
-	object_container<order_t, 2000, 20> orders_container;
+	object_container<image_t, 10000, 50> images_container;
+
+	object_container<order_t, 4000, 20> orders_container;
 
 	intrusive_list<path_t, default_link_f> free_paths;
 	a_list<path_t> paths;
@@ -16848,7 +16848,7 @@
 			}
 		}
 		if (queue) {
-			if (u->order_queue_count < 8 && st.active_orders_size < 1800) {
+			if (u->order_queue_count < 8 && st.active_orders_size < 3600) {
 				queue_order_back(u, order_type, target);
 			}
 		} else {
@@ -21392,8 +21392,11 @@
 					if (unit_type == UnitTypes::Special_Pit_Door) owner = 11;
 					if (unit_type == UnitTypes::Special_Right_Pit_Door) owner = 11;
 					if (use_map_settings || owner == 11) {
-						create_initial_unit(get_unit_type(unit_type), xy(x, y), owner);
-						if (flags & 0x80) error("disable thingy unit");
+						unit_t* u = create_initial_unit(get_unit_type(unit_type), xy(x, y), owner);
+						if (u && (flags & 0x80)) {
+							u_set_status_flag(u, unit_t::status_flag_disabled);
+							u_set_status_flag(u, unit_t::status_flag_no_collide);
+						}
 					}
 				}
 			}
```

Behaviour, in words:

- Use the SC:R object limits from `LMTS`: bullets 400, sprites 5000, images 10000, orders 4000. The "order supply low" threshold is 90% of the order limit (3600). Only the bullet limit changes a result in the sample. For full fidelity the lone-sprite limit (`new_thingy`, 500) should be 1000. The unit limit (1700) can stay: the 1.21 index mapping in `downgrade/` depends on it, and no game comes close.
- When the map loads, create each disabled `THG2` unit, then set the disabled and no-collide status flags on it instead of stopping.

For the Go port: take the limits from `LMTS` when it exists, or use the SC:R values above. Do not use the 1.16 constants.

## Before and after

The baseline is unmodified `ref/openbw`. "Fixed" is the diff above. Tag test: desynced means 20 or more inconsistent commands.

| Sample | Baseline | Fixed |
|---|---|---|
| Terran, 2026-01-01..03-15 (`run.sh -race T -long 0 -n 60`), integration verdict | 9 desynced, 6 suspicious, 1 failed, 44 clean | 5 desynced, 6 suspicious, 0 failed, 49 clean |
| Terran, 2026-03-15..04-24 (57 found), integration verdict | 5 desynced, 4 suspicious, 1 failed, 47 clean | 4 desynced, 4 suspicious, 0 failed, 49 clean |
| Wide sample (`-n 200 -long 30`, 98 replays), integration verdict | 9 desynced, 4 suspicious, 1 failed, 84 clean | 7 desynced, 4 suspicious, 0 failed, 87 clean |
| Both Terran samples (117 that load), tag test | 14 desynced | 8 desynced |
| All 203 distinct replays that load, tag test, Terran games | 21/149 | 13/149 |
| Same, games without a Terran | 1/54 | 1/54 |
| Pole Star 1.1 (`collect -map "Pole Star"`, 30 replays) | 30 fail at frame 0 | 0 desynced |

By the tag test, no replay gets worse. One replay changes from clean to desynced in the integration verdict: MM-85EF47A6 (투혼 1.4, TvZ). In the last minute it has 7 missing tags instead of 4, out of 3,000+. Its bullet count reaches 99 at the baseline limit and 106 with the fix, so the baseline was capped too. The first new inconsistency is a click on a Mutalisk that dies a few frames earlier in OpenBW. This is at the clean-game noise level (0-10).

## MM-029782AE (the short ZvT desync)

The first inconsistent command is at **frame 5148**, before the integration onset (5408). Zerg (p3) selects tags 1513 and 1516 together. In OpenBW, 1513 is a Terran SCV. The allocation order of the two OpenBW units created just before:

```
new 5104 idx=1513 SCV   (Terran, Command Center train at 5103)
new 5113 idx=1512 Larva (Zerg, second Hatchery 1563)
```

SC:R has them in the other order. Two commands confirm it: Zerg morphs 1513 into a Drone at 5408, and Terran selects 1512 with another SCV at 5513. With only these two indexes swapped, every other selection up to frame 6000 matches OpenBW. Later problems follow from this one: in OpenBW the Zerg Drone morph at 5408 fails, so Zerg has 50 more minerals, and the Extractor started at 5960 succeeds in OpenBW but has to be ordered again in SC:R at 6072.

So in SC:R, the Larva spawned before the SCV was created. Either the Larva was at least 10 frames earlier or the SCV at least 10 frames later. Results:

- The SCV is not late for a mechanical reason. Minerals (53 ≥ 50), supply and the Command Center's secondary order state are normal. The train at 5103 creates the unit in the same frame. A cancel at 4844 left no state behind.
- The Larva cannot be 10 frames early because of phase. Hatchery larva timing depends on `order_process_timer`, which the game resets every 150 frames from the unit's position in `visible_units`. All forced reset values (`FORCE_<frame>`, see the tracer variants below) still give 5113 or later.
- If the second Hatchery's morph finishes 10 or 18 frames earlier (`tracer -build-time 3000 1563 -10`), the game stays consistent until frame 7401. The baseline breaks at 5148. This shows the second Hatchery's larva timing is what is off, but not why.
- RNG: one-step shifts before frame ~3500 cause new inconsistencies at 2576-3929. So the RNG matches SC:R until then. No single shift between 3500 and 5100 makes the whole game consistent, so this is not one missing or extra `lcg_rand` call.
- Allocation-model variants all fail at once: inserting into `visible_units` at the front or the back instead of second, and inserting freed units at a different free-list position. OpenBW's list and free-list model is right.

The other Attitude desyncs have the same signature: an early allocation-order swap or an extra or missing unit, between units of different players created a few frames apart. MM-9DAD46D0 (5148), MM-9EE86DE2 (5742), MM-12E10EEC (6595) and MM-84F120D4 (6931) are examples. In MM-12E10EEC, an SCV dies in SC:R after it escapes a Zergling chase in OpenBW. RNG shifts never change that outcome, so it is a deterministic mechanical difference, not an RNG tie.

## Remaining desyncs

After the fix, 14 replays fail the tag test:

| Replay | Map | Matchup | First inconsistency | Inconsistent commands |
|---|---|---|---|---|
| MM-029782AE | Attitude 1.0 | ZvT | 5148 | 106 |
| MM-9DAD46D0 | Attitude 1.0 | PvT | 5148 | 849 |
| MM-9EE86DE2 | Attitude 1.0 | TvZ | 5742 | 262 |
| MM-12E10EEC | Attitude 1.0 | ZvT | 6595 | 224 |
| MM-84F120D4 | Attitude 1.0 | TvP | 6931 | 1862 |
| MM-EA8AF812 | Attitude 1.0 | PvT | 9380 | 500 |
| MM-3F070AD4 | Attitude 1.0 | PvT | 12728 | 40 |
| MM-FFFD042C | Attitude 1.0 | PvT | 22932 | 210 |
| MM-08118F44 | Octagon 1.0 | ZvT | 6329 | 2828 |
| MM-C02C2BC0 | MatchPoint Remastered | TvZ | 9309 | 799 |
| MM-D50574B0 | Jane Doe 1.2 | TvP | 8441 | 81 |
| MM-4B7A5BEA | Jane Doe 1.2 | PvT | 20485 | 30 |
| MM-342E0750 | KnockOut 1.1 | TvZ | 14385 | 20 |
| MM-F4A3BB2C | KnockOut 1.4 | ZvP | 40623 | 512 |

Ranked hypotheses for the Attitude cluster:

1. **A rare map-geometry interaction on Attitude 1.0 that OpenBW resolves differently.** Evidence: 8/18 Terran games on the map desync, against 5/131 elsewhere. The onset is early and has the allocation-tie signature. Not involved: the tileset (4 other jungle maps are clean), Terran construction spark retries (7-8 per game on Attitude, 11-18 elsewhere) and the bottom 32-pixel move bound (removing it breaks 7 more Attitude replays). Attitude has 4 Bengalaas at the map centre, and moving units collide with them. Removing that collision breaks 3 more replays. Critter HP and invincible flags from `UNIT` make no difference: nothing damages the critters. In MM-029782AE a scouting SCV passes a critter at frame 3928, one frame before an event the RNG test shows to be sensitive (3929). In MM-12E10EEC the divergent chase passes the centre. This is the strongest unproven lead.
2. **Path and unit-id reuse with wider SC:R generations.** OpenBW keeps 5 generation bits (mod 32) in `unit_id`, `path_t::last_collision_unit` and control groups. SC:R tags have at least 13 index bits plus a wider generation. A stale id resolves differently after 32 reuses of one slot. This needs long games, so it can explain only the late onsets (MM-FFFD042C, MM-F4A3BB2C), not Attitude's early ones.
3. **The BFIX/GCFG order allocation rule** (next section). It is implemented in a test copy, and it changes no replay in the sample.

## BFIX and GCFG

The ShieldBattery client implements this rule (`game/src/bw_scr/game.rs`, `can_allocate_order`), and `samase_scarf` reverse-engineers it (`analyze_prepare_issue_order`). SC:R's `prepare_issue_order` (OpenBW's `set_queued_order`) checks `BFIX` byte 0 bit `0x4`, `GCFG` `unk10 != 5` and `GCFG` build > 8153. If any is true (here all are, build 13515), SC:R allocates the new order only if one of these holds:

- order supply is not low (allocated orders < 90% of the `LMTS` order limit) and the unit has fewer than 8 queued orders with a highlight icon;
- the unit's *current* order is `Die` or `ResetHarvestCollision`;
- (build ≥ 9713, the "flying SCV" fix) the unit has 8 or more highlighted orders, its current order is `ConstructingBuilding`, and the new order is `ResetCollision`.

Otherwise the order is dropped after the queue cleanup. A test copy with this check in `set_queued_order` (`can_allocate_order`) gives results identical to the baseline on all 117 Terran replays. The units never have 8 highlighted orders when the game queues one itself. It is real SC:R behaviour, so the Go port should implement it. It is not a cause of the measured desyncs.

## SC:R patch notes

Liquipedia's notes for patches 1.18-1.20 contain no unit-behaviour fix that survived. The PTR A Dragoon and Valkyrie fixes were reverted. The OpenBW issue tracker has no Remastered-specific gameplay issue. The SC:R-only gameplay differences found are the `LMTS` limits and the BFIX/GCFG order rule above.

## Implemented

- `downgrade.ReadSCRSettings` and `Report.SCRSettings` read `LMTS`, `BFIX` and `GCFG` from the original SC:R replay. Legacy replays get the 1.16 defaults. `OrderAllocationRule()`, `FlyingSCVFix()` and `OrderSupplyLowThreshold()` give the conditions in "BFIX and GCFG".
- `obwhash` loads SC:R replays directly (downgrade in memory). `-scr-rules` reads the settings.
- Not yet implemented: the engine option that applies the limits, the disabled `THG2` units and the order allocation rule. Until then `-scr-rules` stops with an error.

## Tools (`tools/desync/`)

- `tracer.cpp`, `Makefile` (`make OPENBW=<copy> OUT=<binary>`): plays a downgraded replay on an OpenBW copy. Flags:
  - `-check` prints each inconsistent command and `first_inconsistent <frame> count=<n>`.
  - Without `-check` it prints every command with resolved tags, minerals, gas and supply, plus unit creation, removal and morph events (`-from`/`-to`).
  - `-watch <index>` traces one unit's order, state, position and hatchery larva timer. `-near <radius>` lists other players' units near the first watched unit. `-visible-list` prints changes in `visible_units` around the first watched unit.
  - Counterfactuals: `-rng <frame> <steps>` moves `lcg_rand`. `-drop <frame> <slot>` runs that slot's commands with an empty selection. `-starve <frame> <slot>` runs them with 0 minerals and gas. `-minerals <frame> <slot> <delta>` changes minerals. `-build-time <frame> <index> <delta>` changes a unit's remaining build time.
- `check.sh <openbw> <out> <replays...>` builds the tracer against an OpenBW copy and writes `name first_inconsistent count end_frame` for each replay, in parallel.
- `variant.sh <openbw> <out> [integration flags]` runs `tools/integration` with harness and probe built against an OpenBW copy.
- `collect` downgrades SC:R replays filtered by map, race and length into `.116.rep` files, for example `go run ./tools/desync/collect -replays <dir> -out <dir> -map "Pole Star" -n 30`.
- `mapinfo <legacy.rep>` lists CHK chunks, `UNIT` entries (type, owner, position, resources, property and flag masks, HP%), `THG2` entries with the disabled flag, the tileset, and the number of MTXM/TILE differences.

To reproduce the numbers: copy `ref/openbw` to `$TMPDIR`, apply the diff, then run `check.sh` and `variant.sh` with the same integration flags as the samples above.
