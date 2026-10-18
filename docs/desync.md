# SC:R vs OpenBW desync

Investigation of the desyncs in [integration.md](integration.md): C++ OpenBW (`ref/openbw`, unmodified) playing downgraded SC:R replays. The Go port is not involved. Run date: 2026-10-14.

## Result

| Finding | Status | Effect |
|---|---|---|
| OpenBW keeps the 1.16 object limits. SC:R replays declare larger ones in `LMTS`. The bullet limit (100 vs 400) is reached in big fights, so OpenBW drops attacks that SC:R fires. | **Proven** | Terran desyncs drop from 21/149 to 13/149. Not one replay gets worse by the tag test. |
| Disabled `THG2` units (Pole Star 1.1) are doodad doors. SC:R creates them, sets the disabled doodad state, and units walk through them. | **Proven** | Pole Star: 30/30 replays load and stay in sync. Before: 30/30 fail at frame 0. |
| The remaining Terran desyncs are mostly on one map: **Attitude 1.0**. 8 of 18 Terran games there desync, against 5 of 131 Terran games on the other 12 maps. MM-029782AE is one of them. | **Root cause not found** | See "Remaining desyncs". |
| The Siege/Unsiege lead of integration.md | **Rejected** | The lead was the bullet limit. Tank games have the most bullets. Of the 4 remaining non-Attitude Terran desyncs checked, 3 have no Siege command before the onset. |
| Top vs Bottom teams start as enemies in OpenBW. BW allies each team at game start, and the replay does not record it. | **Proven** | Team games (see "Team games"): Top vs Bottom desyncs drop from 102/137 to 8/137. |
| Early 1.21 replays without `LMTS` use 1.16-layout tags, which the downgrade turned into tag 0 | **Proven** | Commands in those games now take effect. 7/7 stay in sync. |
| Halo Rockets fail at a fixed 80 active bullets. SC:R scales the cap with the `LMTS` bullet limit (80% of it). | **Proven** | Melee-lobby team games: 58/159 → 48/159 desynced in a 400-replay sample. Not one replay gets worse. See "Root cause 5". |
| Melee-lobby team games (alliances by command), the rest | **Root cause not found** | Mostly long games. See "Remaining: Melee-lobby team games". |
| SC:R order allocation rule behind `BFIX` bit 2 | **Real, but no effect in the sample** | The diff makes no difference on 117 Terran replays. See "BFIX". |

## Method

The ground truth is the command stream. In SC:R every tag a player selects exists. A multi-unit selection contains only the player's own units. `tools/desync/tracer` replays a downgraded replay on any OpenBW copy. At each Select or Shift Select it checks each tag against the engine:

- the tag does not exist, or
- the selection has more than one unit and a unit belongs to another player.

Each such command is "inconsistent". Clean games have 0-10 inconsistent commands, from units that die between the click and the command. A game with 20 or more is desynced (the "tag test"). On the 117 Terran baseline replays, the tag test and the `tools/screp/integration` verdict agree on 13 of 14 desynced replays. Each flags one replay the other does not: MM-342E0750 (20 inconsistent commands) and MM-A20D6A70. It also finds the first divergent frame, which is usually minutes before the integration onset.

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
- `obwhash` loads SC:R replays directly (downgrade in memory).
- `bwgame.Rules` (in `GameState.Rules`) is the engine option. The zero value and `DefaultRules()` give exact OpenBW behaviour. `scrrules.FromSettings` turns `SCRSettings` into SC:R rules:
  - Pool sizes from `LMTS`: bullets, sprites, images, orders and lone sprites (`new_thingy`). `issue_order` and the order allocation rule use 90% of the order limit. Only the capacity changes. Index mapping (`max_size - index`) and free-list order stay the same.
  - The unit pool stays 1700. A `unit_id` has an 11-bit index, and `downgrade` maps 1.21 indexes 1,701-3,400 to 1-1,700 for a 1,700-unit array. With fewer than 1,700 live units the allocation order is the same.
  - Disabled `THG2` units are created with `status_flag_disabled` and `status_flag_no_collide`.
  - `set_queued_order` drops the order when `can_allocate_order` is false ("BFIX and GCFG"). The flying SCV fix is a separate flag (`GCFG` build ≥ 9713).
  - Halo Rockets fail to fire at 80% of the bullet limit (`Rules.HaloRocketsBulletCap`), not at a fixed 80. See "Root cause 5". With the 1.16 limit it is 80, so OpenBW rules are unchanged.
- `obwhash -scr-rules` runs the game with these rules.
- `tools/scrharness` builds the C++ harness against OpenBW with the same rules (`openbw-scr.patch`). `difftest -obwhash-args "-scr-rules"` compares the two engines.

### Verification (2026-10-15)

Frame-by-frame state hashes, `difftest`, Go `-scr-rules` against the patched C++ harness:

| Set | Replays | Match |
|---|---|---|
| `tools/difftest/replays.txt` | 20 | 20 |
| Pole Star 1.1 (fail with `disable thingy unit` in default mode) | 20 | 20 |
| Other maps with injected disabled `THG2` doors (`thg2inject`), 13 maps | 24 | 24 |
| Fixed by the bullet limit: the 8 listed in "Root cause 1" plus MM-85EF47A6 | 9 | 9 |
| Terran replays that the tag test shows fixed by `-scr-rules` (includes the 8 above) | 28 | 28 |

No replay on another map fails with `disable thingy unit`: only Pole Star 1.1 does, in the 2,934 cwal.gg replays and the 524 `scmapanalyzer` replays. For that reason the "other maps" set uses replays from the other 13 maps with 5-9 disabled Pit Doors added to `THG2` (map centre and the way to each start location). They all fail with `disable thingy unit` in default mode.

Controls on the 9 bullet-limit replays: Go `-scr-rules` against the unmodified harness differs in 9/9, and Go default against the patched harness differs in 9/9. Default mode against the unmodified harness: 70/70 match (`replays.txt` and 50 other replays from 11 maps).

Tag test (`tools/desync/gotags`) on all 1,462 distinct Terran replays in cwal.gg:

| | Default rules | `-scr-rules` |
|---|---|---|
| Fail to load (Pole Star 1.1, `disable thingy unit`) | 242 | 0 |
| Desynced (20+ inconsistent commands) | 87 | 61 |
| Clean | 1,133 | 1,401 |

Without Pole Star: 87 → 59 desynced of 1,220. 28 replays change from desynced to clean, and none from clean to desynced. 2 of the 242 Pole Star replays are desynced. 45 of the 61 desynced replays are on Attitude 1.0 (see "Remaining desyncs"). The Go tag counts are the same as the C++ tracer's for the replays in "Root cause 1" and for MM-029782AE.

## Team games (2026-10-15)

The sections above cover 1v1 ladder games. A random sample of 400 SC:R replays from one player's own folder (`Maps/Replays`, 2017-2026, mostly Big Game Hunters team games) desynced in most team games: 136 of 234 that load, with SC:R rules. 1v1 games were already fine (0 of 117). Two engine and downgrade bugs caused most of it. Both are fixed.

| Game type (3+ players) | Before | After |
|---|---|---|
| Top vs Bottom | 102/137 desynced | **8/137** |
| Melee lobby (teams made with in-game alliances), FFA | 33/96 | 33/96 (partly fixed later by root cause 5, see below) |
| 1v1 (control) | 0/117 | 0/117 |

Tag test, `gotags -scr-rules`, desynced means 20 or more inconsistent commands. No replay changes from clean to desynced.

### Root cause 3: Top vs Bottom teams start as enemies

In a Top vs Bottom game (game info game type 15), BW allies each team when the game starts: alliance 2 (allied victory) between teammates, and shared vision. The replay does not record it: there are no alliance (`0x0E`) or vision (`0x0D`) commands for it. OpenBW reads the game type but ignores it, so every player starts as everyone's enemy. Teammates' units attack each other on first contact. The desync shows as soon as two teammates' armies or workers meet, which is why the median onset was minutes into the game.

Evidence:

- Start locations, unit tags and mineral right-clicks match the replay from frame 0. The engine was right until first contact.
- Allying same-force players at load fixes 5 of the 9 desynced games in a 12-game BGH subset. Adding shared vision fixes the other 2 that alliance alone did not (234410: 1,609 → 14 inconsistent commands).
- The one alliance command in 234410 (frame 108) sets the same value the rule gives (teammates 2), so it changes nothing. That fits a starting alliance that already exists.
- The game info's "teams enabled" byte is 0 in Top vs Bottom, so the trigger is the game type, not that byte.

A second rule is needed for 2 games where players ally across teams with alliance commands (182824, 170727). In SC:R, such a cross-team alliance does not give allied victory: with it, OpenBW's melee victory trigger gives one player the victory at frame 3133 and removes the others, while the SC:R game goes on to frame 28049. Capping a cross-team alliance at 1 (allied, no allied victory) fixes both and changes nothing else in the 238 Top vs Bottom replays. Ignoring alliance commands in Top vs Bottom instead (alliances locked) breaks 3 clean games, so SC:R does apply them.

### Root cause 4: 1.21 tags of games without `LMTS`

Early 1.21 replays (2017-2019 in this sample) have no `LMTS` section. Their unit array is 1,700, and their 4-byte tags use the 1.16 layout: an 11-bit index and the generation above it. `downgrade` read every 1.21 tag as a 13-bit index into a 3,400-unit array, so these tags got index 1,701+ and became 0 (no unit). Every Select and targeted command did nothing, and the players trained nothing. The tag test did not see it, because it ignores tag 0. Example: 165606 (3v3, 2018-09) selects tag `0x0DA8`, which is index 1448 (a Drone of that player), generation 1.

7 of the 389 replays have this format. All 7 now load with working commands and stay in sync.

### Exact change

- `downgrade`: the tag layout comes from the replay's unit limit (`LMTS`, or 1,700 without it). With 1,700 units a tag is `index | generation << 11`, with no index remapping. `LegacyUnitTagFor(settings, tag)` exposes it. `LegacyUnitTag` keeps the 3,400-unit mapping.
- `bwgame.Rules.TopVsBottomAlliances` (on in `scrrules.FromSettings`, so in `RulesAuto` for SC:R replays): when the replay's game type is Top vs Bottom, after the starting units and force vision, each pair of occupied slots in the same force gets alliance 2 and shared vision. In such a game an alliance command sets at most 1 toward a player of another force.
- `RulesInfo.TopVsBottomAlliances` reports it. Default (`RulesOpenBW`) behaviour is unchanged, so `difftest` against unpatched C++ OpenBW is unaffected. `tools/scrharness/openbw-scr.patch` does not have this rule, so `difftest -scr-rules` on a Top vs Bottom replay will differ. `tools/difftest/replays.txt` has none.
- Legacy (1.16) Top vs Bottom replays still run with OpenBW rules under `RulesAuto`, so they still have this bug. Use a custom `Rules` with `TopVsBottomAlliances` for them.

### Root cause 5: the Halo Rockets cap (2026-10-16)

BW drops a Valkyrie's Halo Rockets when 80 or more bullets are active: 80% of the 1.16 bullet limit of 100. OpenBW has the fixed 80 (`create_bullet`). SC:R keeps the ratio when `LMTS` raises the limit: the cap is 160 with 200 bullets and 320 with 400. Liquipedia's Valkyrie page describes the cap as a known BW problem in team battles that the larger SC:R limits removed. Big team fights pass 80 active bullets often, so OpenBW dropped rockets that SC:R fired.

Evidence:

- 203541 (200 bullets): 1,269 → 17 inconsistent commands. 213744 (400 bullets): 309 → 12. At the onset of 213744, Valkyries fight next to Siege Tanks and Zerglings.
- A cap at 80% of the limit and no cap at all give the same result on 399 of 400 replays of the sample. The exception is 233905, a game created with the original CCMU limits (100 bullets): clean with the cap of 80 (9 inconsistent commands), desynced without it (86). So SC:R still has the cap. It is not removed, and it moves with the limit.

The fix: `Rules.HaloRocketsBulletCap()` is `ObjectLimits().Bullets * 4 / 5`. With OpenBW rules (100 bullets) it is 80, so `difftest` against unpatched C++ OpenBW is unchanged. `tools/scrharness/openbw-scr.patch` uses 320.

Tag test, `gotags -scr-rules`, before and after:

| Set | Before | After | Changed |
|---|---|---|---|
| 400 random Big Game Hunters replays of the player's folder: Melee lobby, 7-8 players | 58/159 desynced | **48/159** | 10 fixed |
| Same sample: Melee lobby, 2-6 players | 0/4 | 0/4 | none |
| Same sample: Top vs Bottom | 12/207 | **10/207** | 2 fixed |
| 300 random 1v1 ladder replays (pro accounts, 400 bullets) | 5/300 | **4/300** | MM-B625CDC6: 500 → 3 |
| 200 random other replays of the player's folder (160 load) | 4/160 | **2/160** | 2 fixed |

No replay changes from clean to desynced.

### Remaining: Melee-lobby team games

48 of 159 Melee-lobby team games (7-8 players) still desync. These players form teams with alliance commands, which OpenBW already handles. Top vs Bottom games show the same effect of length (before root cause 5, above 30,000 frames: 9 of 25 desync, against 3 of 186 below), so it is one general mechanism whose rate grows with the amount of fighting, not something specific to alliance commands. Melee lobbies have 8 players and 195 of the 211 Top vs Bottom games have 6 or fewer, which explains the higher rate.

| Melee lobby, 7-8 players, by length | Before root cause 5 | After |
|---|---|---|
| under 30,000 frames | 9/94 | 4/94 |
| 30,000-50,000 | 28/43 | 23/43 |
| over 50,000 | 21/22 | 21/22 |

How to read the tag test in these games: isolated inconsistent commands are normal. A unit dies a frame early in OpenBW, the player clicks it, and the game heals, because the RNG and the unit allocation stay the same (180911: a Marine duel at equal HP at frame 6902; the slot is reused identically at 6963). The real onset is where the misses become sustained (`tools/desync/onset`: 6 inconsistent commands within 3,000 frames). From that point the unit allocation order differs, and players select tags that OpenBW gives to other players' units. The sustained onset is usually 20,000-30,000 frames into these games, minutes after the first isolated miss.

Ruled out:

- **5-bit unit-id generation wrap.** A stale id resolves to a new unit only after 32 reuses of one slot. The highest generation reached is 7 in 180911 (61,163 frames) and 10 in 213744. No slot wraps.
- **Players leaving or being defeated.** 20 of 58 desynced games desync before any player is defeated or leaves (for example 180911 at 23,772, first defeat at 27,251). Keeping a leaver's units with their owner instead of making them neutral makes 213744 desync at 5,412 instead of 27,905, so OpenBW's `make_unit_neutral` is right. Killing a defeated player's units instead changes nothing in 234855.
- **Object and creep pools.** Unit, bullet, sprite and image pools stay below the limits. The receding-creep pool (1,024 entries) has at least 636 free entries in 234855.

The best lead is 234855 (sustained from frame 6,000). Player 1 (Zerg) loses its last hatchery at 4,821 and is defeated at 5,861. Its units become neutral. Its larvae keep the Larva order, wander on the receding creep and die off creep at 6,487 and 6,517. In SC:R those unit slots are already reused before frame 6,273. So the larvae of a destroyed hatchery die earlier in SC:R: either the creep under them recedes faster, or SC:R ends the Larva order differently when the hatchery is gone or the larva is neutral. 180911 has a different trigger: its Valkyries never reach the cap, and its first sustained misses (frames 23,768 and 26,214) are a Hydralisk and an SCV that die to units of two different players.

### Load failures

All 37 replays of the sample that fail to load fail on OpenBW features that are not implemented. None is a Go port bug.

| Error | Replays | Cause | Triage |
|---|---|---|---|
| `ComputerAI` (16), `AIPatrol` (9) | 25 | A computer player's units run AI orders. OpenBW has no computer AI. | Not fixable without porting BW's AI script engine. Skip games with a computer slot (the header says so). |
| `unknown trigger condition 4` | 8 | UMS map trigger condition 4 (Accumulate) is not implemented. | UMS only. Skip UMS or implement the trigger condition. |
| `fixme unknown flag` | 2 | `complete_unit` stops on units.dat flag `0x800` (single entity: powerups). OpenBW never sets `building.powerup.origin`. | UMS micro maps only. Not fixed: the exact BW behaviour is not known. |
| `HiddenGun`, `UNIT: invalid owner 12` | 2 | An unimplemented order and a map unit owned by player 12. | UMS-style maps only. |
| `disable thingy unit` | 2 | Pole Star 1.1 with OpenBW rules only. | Already fixed by SC:R rules (root cause 2). |

The 16 1v1 load failures are all games against a computer or UMS micro maps.

## Tools (`tools/desync/`)

- `tracer.cpp`, `Makefile` (`make OPENBW=<copy> OUT=<binary>`): plays a downgraded replay on an OpenBW copy. Flags:
  - `-check` prints each inconsistent command and `first_inconsistent <frame> count=<n>`.
  - Without `-check` it prints every command with resolved tags, minerals, gas and supply, plus unit creation, removal and morph events (`-from`/`-to`).
  - `-watch <index>` traces one unit's order, state, position and hatchery larva timer. `-near <radius>` lists other players' units near the first watched unit. `-visible-list` prints changes in `visible_units` around the first watched unit.
  - Counterfactuals: `-rng <frame> <steps>` moves `lcg_rand`. `-drop <frame> <slot>` runs that slot's commands with an empty selection. `-starve <frame> <slot>` runs them with 0 minerals and gas. `-minerals <frame> <slot> <delta>` changes minerals. `-build-time <frame> <index> <delta>` changes a unit's remaining build time.
- `check.sh <openbw> <out> <replays...>` builds the tracer against an OpenBW copy and writes `name first_inconsistent count end_frame` for each replay, in parallel.
- `variant.sh <openbw> <out> [integration flags]` runs `tools/screp/integration` with harness and probe built against an OpenBW copy.
- `collect` downgrades SC:R replays filtered by map, race and length into `.116.rep` files, for example `go run ./tools/desync/collect -replays <dir> -out <dir> -map "Pole Star" -n 30`. It reads the map, races and length from the engine's `Header`, so it needs the game data (`-data` or `$OPENBW_DATA`).
- `gotags [-scr-rules] <replays...>` runs the tag test on the Go engine, in parallel, and prints `name first_inconsistent count=<n> end_frame` (or `name error ...`). It reads SC:R replays directly.
- `thg2inject in.rep out.rep` adds disabled Pit Doors to the `THG2` chunk of a 1.21 SC:R replay's map: one at the map centre and two on the way to each start location.
- `teamprobe [-scr-rules] [-from f] [-to f] [-stats n] <replay>` prints the players (controller, race, force, start location), the units at load, and each command in the frame window with resolved tags and right-click targets, plus unit creation, destruction, morph, victory and leave events. `-stats n` prints the object pool sizes every n frames. It reads SC:R replays directly.
- `onset <replays...>` runs the tag test with SC:R rules on Melee and Top vs Bottom replays, in parallel, and prints one line per replay as it finishes: game type, occupied player slots, end frame, inconsistent commands, the sustained onset (6 inconsistent commands within 3,000 frames) and every victory-state change (`frame:player:state`).
- `watch -idx <index> [-every n] [-x x -y y -r half-size] -from f -to f <replay>` prints the damage and death of one unit (and of every unit in a square), and every n frames its order, position and tile flags, and the receding-creep pool.
- `mapinfo <legacy.rep>` lists CHK chunks, `UNIT` entries (type, owner, position, resources, property and flag masks, HP%), `THG2` entries with the disabled flag, the tileset, and the number of MTXM/TILE differences.

To reproduce the numbers: copy `ref/openbw` to `$TMPDIR`, apply the diff, then run `check.sh` and `variant.sh` with the same integration flags as the samples above.
