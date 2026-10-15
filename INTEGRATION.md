# Integration: SC:R replay downgrade on unmodified OpenBW

End-to-end test of `downgrade` + `tools/harness` on C++ OpenBW at `4b046d5`. Run date: 2026-10-14.

## Result

- 30 replays from `cwalggdl`: **26 clean, 1 suspicious, 2 desynced, 1 failed** (Pole Star map).
- All 29 replays that load play to `end_frame` without an engine error. The downgrade dropped 0 commands.
- No failure was caused by `downgrade/` or `tools/harness/`. Nothing in those directories was changed.
- Desyncs are real gameplay divergence in OpenBW, not translation errors. They almost always involve a Terran player.

A wider sample with the same tool (98 replays) gives 84 clean, 4 suspicious, 9 desynced, 1 failed. Two extra Terran-only samples (60 replays each) give 9 and 6 desynced. Estimated desync rate: **about 13% of games with a Terran, about 2% of games without one** (1 of 53).

## How to run

```
tools/screp/integration/run.sh                       # the 30-replay set below
tools/screp/integration/run.sh -n 200 -long 30       # the 98-replay wide sample
tools/screp/integration/run.sh -race T -from 2026-01-01 -to 2026-03-15 -long 0 -n 60
```

The tool is in the `tools/screp` module, because it uses screp. `run.sh` builds `tools/harness/harness`, `tools/screp/integration/probe/probe` and the tool, then runs it on `$REPLAYS` (default `$OPENBW_REPLAYS/cwalggdl`). It prints the Markdown tables below. `OUT` (default `$TMPDIR/openbw-integration`) gets the downgraded replays, the raw harness and probe output and `results.json`.

For each replay the tool:

1. Downgrades it in-process with `downgrade.Downgrade`.
2. Runs `tools/harness/harness` to the end: load, end frame, `action_fails`, final per-player state.
3. Runs `tools/screp/integration/probe/probe` (a second program on unmodified OpenBW) for desync evidence:
   - units each player actually made, per type (completed units seen by index, generation and type);
   - why each Train, Build, Unit Morph and Building Morph was rejected: `no_unit`/`cannot_build` (the selection cannot do it) or `resources`/`supply`/`queue`/`already_morphing` (normal spam);
   - every selected or targeted unit tag that does not exist in the engine (`missing_tag` on stderr);
   - peak use of the 1.16 object limits (units 1,700, bullets 100, sprites 2,500, images 5,000, orders 2,000).
4. Compares this with screp's command stream for the same player (effective Train/Build/Morph counts per unit type, leave frame, last command).

## Selection

All 2,934 `cwalggdl` replays are 1v1 (Top vs Bottom), 2025-10-01 to 2026-04-23. The 1,782 that last at least 5 minutes are candidates. The tool picks the 4 longest games under 60 minutes (distinct maps), then goes round-robin over the 14 maps at different date quantiles. Pole Star 1.1 is capped at 1 replay, because OpenBW cannot load it.

## Classification

The strongest signal is the unit tags. In a synced game, every unit a player selects or right-clicks exists. A desynced engine has different units, so the tags stop resolving.

| Verdict | Rule |
|---|---|
| desynced | In some minute, at least 5 selected/targeted tags (and at least 5% of that minute's tags) do not exist in the engine; or 10+ Train/Build/Morph commands whose selection cannot do them; or 2+ unit types ordered 5+ times but never made; or 20+ commands after the engine removed all of the player's units |
| suspicious | 5+ missing tags overall (at least 0.25%); or 3+ "no unit" production rejects; or 1 unit type ordered 5+ times but never made; or an OpenBW object limit reached |
| clean | none of the above |

Noise level in clean games: 0-5 missing tags in 1,000-5,000, from units that die in the frames between click and execution.

Rejected Train commands for money, supply or a full queue are not evidence. OpenBW `build_queue_push` rejects them exactly like SC:R does, and some players spam the train hotkey (for example Horang2[._.]: 207 of 266 Train commands rejected for money, in a game with 0 missing tags). The "made units / orders" ratio and unspent minerals are shown, but alone they do not make a replay suspicious.

## Results: 30 replays

`First anomaly cluster` is the first frame that starts 5+ missing tags or "no unit" rejects within one minute.

| # | Replay | Date | Map | Matchup | Frames | Loads | To end | Dropped | action_fail | Peak units / bullets / sprites / images / orders | First anomaly cluster (frame) | Verdict |
|---|---|---|---|---|---|---|---|---|---|---|---|---|
| 1 | 30-MORE-mentalgap/MM-6CE4E808 | 2025-10-14 | Dominator SE 2. | ZvT | 17571 | yes | yes | 0 | 61 | 319 / 39 / 470 / 840 / 120 | - | suspicious |
| 2 | 30-MORE-mentalgap/MM-5071BE2E | 2025-10-15 | Metropolis 1.1 | ZvP | 24337 | yes | yes | 0 | 182 | 477 / 45 / 591 / 1161 / 187 | - | clean |
| 3 | 30-MORE-mentalgap/MM-FC802C7E | 2025-10-15 | Radeon 1.2 | ZvZ | 7369 | yes | yes | 0 | 18 | 220 / 0 / 246 / 483 / 66 | - | clean |
| 4 | 30-MORE-sdfsf22sfa/MM-3D3E81A4 | 2025-10-23 | KnockOut 1.1 | TvP | 31549 | yes | yes | 0 | 303 | 510 / 75 / 781 / 1596 / 203 | - | clean |
| 5 | 30-MORE-sdfsf22sfa/MM-4985AD6C | 2025-10-24 | 투혼 1.4 | TvP | 27032 | yes | yes | 0 | 162 | 512 / 76 / 717 / 1469 / 197 | - | clean |
| 6 | 30-MORE-3050_KzerG/MM-E8DD6C38 | 2025-10-25 | Litmus 1.1 | TvZ | 16035 | yes | yes | 0 | 54 | 289 / 46 / 433 / 761 / 109 | - | clean |
| 7 | 30-NEW-Horang2[._.]/MM-AB99F700 | 2026-01-02 | Polypoid 1.75 | PvT | 35115 | yes | yes | 0 | 220 | 650 / 76 / 934 / 1849 / 225 | - | clean |
| 8 | 30-NEW-Horang2[._.]/MM-0BAED626 | 2026-01-05 | Dominator SE 2. | ZvP | 11461 | yes | yes | 0 | 216 | 236 / 23 / 331 / 609 / 68 | - | clean |
| 9 | 30-NEW-Horang2[._.]/MM-7FCE8310 | 2026-01-16 | Radeon 1.2 | PvZ | 17257 | yes | yes | 0 | 262 | 347 / 43 / 453 / 880 / 143 | - | clean |
| 10 | 30-NEW-lillljilililili/MM-F7581750 | 2026-02-04 | Metropolis 1.1 | ZvP | 35692 | yes | yes | 0 | 182 | 519 / 51 / 604 / 1195 / 190 | - | clean |
| 11 | 30-NEW-lillljilililili/MM-000E17C0 | 2026-02-04 | KnockOut 1.1 | ZvZ | 10647 | yes | yes | 0 | 29 | 207 / 8 / 412 / 743 / 56 | - | clean |
| 12 | 30-NEW-lillljilililili/MM-724993B8 | 2026-02-24 | Jane Doe 1.2 | ZvT | 25097 | yes | yes | 0 | 141 | 530 / 73 / 692 / 1262 / 228 | - | clean |
| 13 | 30-lIlIIIIIIllllll/MM-D4DCAE08 | 2026-02-25 | MatchPoint Remastered | TvP | 8409 | yes | yes | 0 | 51 | 172 / 26 / 213 / 427 / 75 | - | clean |
| 14 | 30-NEW-lillljilililili/MM-491926A2 | 2026-02-25 | Octagon 1.0 | ZvZ | 12161 | yes | yes | 0 | 14 | 268 / 21 / 414 / 694 / 90 | - | clean |
| 15 | 30-NEW-lillljilililili/MM-CB157510 | 2026-02-25 | Neo Sylphid 3.2 | ZvT | 21140 | yes | yes | 0 | 142 | 391 / 64 / 624 / 1231 / 164 | - | clean |
| 16 | 30-NEW-lillljilililili/MM-E0A8B85E | 2026-02-26 | KnockOut 1.4 | TvZ | 16448 | yes | yes | 0 | 74 | 343 / 46 / 663 / 1194 / 127 | - | clean |
| 17 | 30-MORE-SST_JumJaJungJi/MM-AAA58B20 | 2026-02-26 | Attitude 1.0 | TvT | 20411 | yes | yes | 0 | 189 | 469 / 50 / 585 / 1183 / 177 | - | clean |
| 18 | 30-Esccccccc/MM-4862526E | 2026-03-22 | Jane Doe 1.2 | ZvP | 22865 | yes | yes | 0 | 267 | 460 / 60 / 617 / 1144 / 209 | - | clean |
| 19 | 30-Esccccccc/MM-AD3C6A4C | 2026-03-22 | MatchPoint Remastered | ZvT | 15819 | yes | yes | 0 | 74 | 261 / 52 / 401 / 709 / 115 | - | clean |
| 20 | 30-MORE-Skins_/MM-75695598 | 2026-04-06 | Pole Star 1.1 | ZvZ | 9374 | no | no (0) | 0 | 0 | - | - | failed |
| 21 | 30-MORE-BBBuuuUU[kS]/MM-0E7AF1D0 | 2026-04-10 | Neo Sylphid 3.2 | ZvZ | 7684 | yes | yes | 0 | 5 | 207 / 0 / 373 / 743 / 59 | - | clean |
| 22 | 30-MORE-Skins_/MM-96471080 | 2026-04-10 | Octagon 1.0 | ZvZ | 8835 | yes | yes | 0 | 46 | 206 / 4 / 233 / 445 / 53 | - | clean |
| 23 | 30-NEW-Horang2[._.]/MM-A1B13D46 | 2026-04-11 | 투혼 1.4 | TvP | 26128 | yes | yes | 0 | 243 | 396 / 52 / 551 / 1124 / 145 | - | clean |
| 24 | 30-MORE-Skins_/MM-71C039A8 | 2026-04-11 | Polypoid 1.75 | ZvT | 11373 | yes | yes | 0 | 118 | 268 / 31 / 522 / 925 / 95 | - | clean |
| 25 | 30-NEW-Horang2[._.]/MM-39C48C10 | 2026-04-11 | KnockOut 1.4 | PvT | 7250 | yes | yes | 0 | 106 | 214 / 26 / 425 / 812 / 64 | - | clean |
| 26 | 30-MORE-231314231313131/MM-40AD2926 | 2026-04-15 | Attitude 1.0 | PvP | 13224 | yes | yes | 0 | 262 | 209 / 31 / 314 / 585 / 52 | - | clean |
| 27 | 30-MORE-4023b/MM-F4C14D6E | 2026-04-17 | Neo Sylphid 3.2 | ZvT | 83841 | yes | yes | 0 | 102 | 811 / 74 / 1001 / 2275 / 127 | - | clean |
| 28 | 30-MORE-BBBuuuUU[kS]/MM-657AB750 | 2026-04-18 | Polypoid 1.75 | PvZ | 79906 | yes | yes | 0 | 244 | 576 / 44 / 827 / 1584 / 247 | - | clean |
| 29 | 30-MORE-4023b/MM-D50574B0 | 2026-04-21 | Jane Doe 1.2 | TvP | 83569 | yes | yes | 0 | 427 | 628 / 52 / 680 / 1543 / 123 | 45140 (min 31 of 58) | desynced |
| 30 | 30-MORE-4023b/MM-FFFD042C | 2026-04-23 | Attitude 1.0 | PvT | 82196 | yes | yes | 0 | 1174 | 658 / 84 / 803 / 1664 / 243 | 29113 (min 20 of 57) | desynced |

Final per-player state (from the harness; "made" and rejects from the probe; orders from screp, effective commands only):

| # | Player | Race | Minerals | Gas | Supply | Alive | Made units / orders | Made buildings / orders | Rejected: no unit / economy / cmds | Missing tags | No units at | Left at |
|---|---|---|---|---|---|---|---|---|---|---|---|---|
| 1 | mentalgap | Zerg | 594 | 478 | 60/58 | 67 | 92 / 82 | 24 / 30 | 1 / 10 / 112 | 0 / 937 | - | - |
| 1 | kdqj | Terran | 784 | 663 | 0/0 | 0 | 119 / 169 | 43 / 57 | 0 / 41 / 226 | 1 / 1435 | 17512 | 17511 |
| 2 | 340iM3 | Protoss | 1075 | 599 | 0/0 | 0 | 153 / 317 | 49 / 56 | 0 / 151 / 373 | 1 / 1766 | 24276 | 24273 |
| 2 | mentalgap | Zerg | 768 | 1059 | 135/160 | 191 | 404 / 228 | 24 / 25 | 0 / 12 / 253 | 0 / 2123 | - | - |
| 3 | mentalgap | Zerg | 69 | 518 | 0/0 | 0 | 38 / 35 | 5 / 16 | 0 / 7 / 51 | 0 / 543 | 7308 | 7305 |
| 3 | ArmaniSC2 | Zerg | 70 | 450 | 26/34 | 45 | 38 / 30 | 4 / 5 | 0 / 2 / 35 | 0 / 1122 | - | - |
| 4 | ff291 | Terran | 675 | 2152 | 0/0 | 0 | 323 / 495 | 59 / 75 | 1 / 196 / 570 | 3 / 2882 | 31484 | 31482 |
| 4 | sdfsf22sfa | Protoss | 561 | 2209 | 190/261 | 192 | 271 / 357 | 69 / 97 | 0 / 67 / 454 | 0 / 2537 | - | - |
| 5 | hyeonsu0623 | Terran | 2528 | 1111 | 200/200 | 253 | 244 / 325 | 62 / 78 | 0 / 87 / 403 | 4 / 2274 | - | - |
| 5 | sdfsf22sfa | Protoss | 574 | 979 | 0/0 | 0 | 224 / 298 | 77 / 110 | 0 / 62 / 408 | 0 / 2052 | 26972 | 26969 |
| 6 | 3050_KzerG | Zerg | 624 | 483 | 0/0 | 0 | 98 / 69 | 10 / 11 | 0 / 1 / 80 | 1 / 806 | 15976 | 15975 |
| 6 | FlOwEr)CoNtRoL | Terran | 602 | 350 | 83/84 | 108 | 78 / 120 | 29 / 36 | 0 / 38 / 156 | 0 / 1808 | - | - |
| 7 | v_[0_0]_v | Terran | 310 | 520 | 0/0 | 0 | 323 / 320 | 93 / 132 | 0 / 32 / 452 | 3 / 2803 | 35056 | 35055 |
| 7 | Horang2[._.] | Protoss | 3759 | 4789 | 196/402 | 275 | 320 / 472 | 135 / 171 | 0 / 135 / 643 | 4 / 3140 | - | - |
| 8 | UtataneLeina | Zerg | 935 | 234 | 47.5/50 | 71 | 66 / 46 | 11 / 20 | 2 / 3 / 66 | 1 / 644 | - | - |
| 8 | Horang2[._.] | Protoss | 151 | 138 | 0/0 | 0 | 37 / 246 | 18 / 20 | 0 / 207 / 266 | 0 / 1446 | 11388 | 11387 |
| 9 | Iksansi- | Zerg | 75 | 934 | 82.5/133 | 108 | 151 / 132 | 19 / 28 | 1 / 6 / 160 | 1 / 1563 | - | - |
| 9 | Horang2[._.] | Protoss | 1780 | 709 | 0/0 | 0 | 81 / 329 | 41 / 53 | 1 / 238 / 382 | 0 / 1426 | 17040 | 17037 |
| 10 | JG_Quickly | Protoss | 950 | 221 | 0/0 | 0 | 245 / 394 | 54 / 70 | 0 / 142 / 464 | 2 / 2422 | 34480 | 34478 |
| 10 | lillljilililili | Zerg | 1139 | 2244 | 187.5/225 | 313 | 815 / 436 | 29 / 38 | 0 / 14 / 474 | 5 / 5270 | - | - |
| 11 | lillljilililili | Zerg | 35 | 27 | 36/42 | 45 | 59 / 53 | 6 / 6 | 0 / 10 / 59 | 0 / 1251 | - | - |
| 11 | ssssowhat | Zerg | 87 | 55 | 0/0 | 0 | 40 / 33 | 4 / 4 | 0 / 5 / 37 | 0 / 1253 | 9104 | 9101 |
| 12 | lillljilililili | Zerg | 2557 | 1281 | 193/193 | 247 | 351 / 307 | 30 / 37 | 1 / 14 / 344 | 3 / 3122 | - | - |
| 12 | cklvgmnklthb | Terran | 1745 | 250 | 0/0 | 0 | 185 / 271 | 66 / 105 | 0 / 88 / 376 | 0 / 1718 | 23624 | 23623 |
| 13 | JSU_susik | Terran | 214 | 166 | 0/0 | 0 | 32 / 41 | 7 / 11 | 0 / 6 / 52 | 0 / 632 | 8324 | 8323 |
| 13 | lIlIIIIIIllllll | Protoss | 261 | 224 | 41/41 | 48 | 29 / 77 | 10 / 13 | 0 / 42 / 90 | 0 / 393 | - | - |
| 14 | lIllIll1Illllll | Zerg | 464 | 407 | 30.5/67 | 55 | 75 / 55 | 9 / 11 | 0 / 2 / 66 | 0 / 920 | - | - |
| 14 | lillljilililili | Zerg | 59 | 8 | 51/66 | 58 | 72 / 63 | 6 / 9 | 1 / 8 / 72 | 0 / 1828 | - | - |
| 15 | lillljilililili | Zerg | 1287 | 557 | 170.5/173 | 191 | 198 / 156 | 50 / 60 | 2 / 12 / 216 | 0 / 2286 | - | - |
| 15 | 564376 | Terran | 575 | 1078 | 0/0 | 0 | 124 / 257 | 40 / 57 | 0 / 114 / 314 | 1 / 1260 | 18020 | 18016 |
| 16 | lillljilililili | Zerg | 800 | 269 | 0/0 | 0 | 107 / 96 | 22 / 26 | 2 / 5 / 122 | 0 / 1189 | 16280 | 16276 |
| 16 | f189asdfasd165f | Terran | 329 | 248 | 97/100 | 137 | 137 / 212 | 38 / 44 | 0 / 57 / 256 | 1 / 958 | - | - |
| 17 | SST_JumJaJungJi | Terran | 1000 | 255 | 136/200 | 189 | 158 / 266 | 48 / 80 | 0 / 109 / 346 | 2 / 1603 | - | - |
| 17 | PrimeBoxer | Terran | 1647 | 527 | 0/0 | 0 | 133 / 200 | 35 / 40 | 0 / 63 / 240 | 1 / 1244 | 20348 | 20344 |
| 18 | Esccccccc | Zerg | 1264 | 620 | 0/0 | 0 | 327 / 238 | 20 / 31 | 2 / 6 / 269 | 5 / 2684 | 22808 | 22805 |
| 18 | MiniMaxii | Protoss | 787 | 484 | 162/212 | 176 | 180 / 438 | 65 / 78 | 0 / 246 / 516 | 1 / 2506 | - | - |
| 19 | lll1Ill1Illllll | Terran | 377 | 282 | 0/0 | 0 | 66 / 134 | 29 / 37 | 0 / 60 / 171 | 1 / 828 | 12032 | 12030 |
| 19 | Esccccccc | Zerg | 211 | 254 | 107.5/116 | 126 | 142 / 112 | 17 / 22 | 0 / 4 / 134 | 0 / 1389 | - | - |
| 20 | Skins_ | Zerg | 0 | 0 |  | 0 | 0 / 32 | 0 / 10 | 0 / 0 / 42 | 0 / 0 | - | - |
| 20 | Stay[WHITE] | Zerg | 0 | 0 |  | 0 | 0 / 32 | 0 / 12 | 0 / 0 / 44 | 0 / 0 | - | 9310 |
| 21 | neryj94 | Zerg | 216 | 362 | 19/26 | 34 | 53 / 30 | 4 / 6 | 0 / 1 / 36 | 1 / 597 | - | - |
| 21 | BBBuuuUU[kS] | Zerg | 126 | 436 | 0/0 | 0 | 56 / 31 | 3 / 3 | 0 / 0 / 34 | 0 / 291 | 7628 | 7624 |
| 22 | Skins_ | Zerg | 34 | 50 | 17.5/26 | 31 | 57 / 41 | 5 / 7 | 1 / 11 / 48 | 0 / 946 | - | - |
| 22 | eOnzErG | Zerg | 95 | 230 | 0/0 | 0 | 34 / 32 | 7 / 10 | 2 / 8 / 42 | 1 / 954 | 8788 | 8786 |
| 23 | fnq;ltkvkdl'dj+ | Terran | 861 | 468 | 198/214 | 251 | 186 / 247 | 63 / 67 | 0 / 67 / 314 | 1 / 2358 | - | - |
| 23 | Horang2[._.] | Protoss | 2223 | 1686 | 0/0 | 0 | 104 / 226 | 33 / 43 | 0 / 95 / 269 | 5 / 2043 | 21464 | 21462 |
| 24 | Samsung_Paup_Go | Terran | 765 | 586 | 0/0 | 0 | 64 / 167 | 22 / 29 | 0 / 98 / 196 | 1 / 692 | 11316 | 11315 |
| 24 | Skins_ | Zerg | 295 | 438 | 56/67 | 66 | 60 / 55 | 13 / 17 | 0 / 6 / 72 | 1 / 933 | - | - |
| 25 | Horang2[._.] | Protoss | 121 | 60 | 36/42 | 43 | 25 / 92 | 11 / 14 | 0 / 64 / 106 | 0 / 901 | - | - |
| 25 | IIIIlIlllllllI | Terran | 454 | 476 | 0/0 | 0 | 27 / 67 | 7 / 8 | 0 / 39 / 75 | 1 / 314 | 7200 | 7199 |
| 26 | kydkyd | Protoss | 300 | 460 | 26/66 | 40 | 45 / 79 | 20 / 22 | 0 / 31 / 101 | 0 / 944 | - | - |
| 26 | 231314231313131 | Protoss | 201 | 296 | 0/0 | 0 | 36 / 252 | 9 / 9 | 0 / 212 / 261 | 0 / 988 | 13164 | 13160 |
| 27 | 4023b | Terran | 58129 | 19997 | 200/418 | 760 | 270 / 302 | 356 / 542 | 0 / 70 / 844 | 5 / 5720 | - | - |
| 27 | RK_ANNA | Zerg | 438 | 1801 | 0/0 | 0 | 117 / 65 | 35 / 39 | 0 / 4 / 104 | 0 / 866 | 18860 | 18858 |
| 28 | devilmoney | Protoss | 831 | 1783 | 144/286 | 202 | 606 / 767 | 104 / 119 | 0 / 184 / 886 | 3 / 5468 | - | - |
| 28 | BBBuuuUU[kS] | Zerg | 726 | 3639 | 0/0 | 0 | 1675 / 638 | 45 / 53 | 0 / 17 / 691 | 9 / 5618 | 79816 | 79815 |
| 29 | SKT1JSPARK | Protoss | 371 | 420 | 0/0 | 0 | 24 / 66 | 9 / 14 | 0 / 39 / 80 | 1 / 2649 | 8532 | 8529 |
| 29 | 4023b | Terran | 46524 | 26951 | 178/256 | 563 | 187 / 192 | 259 / 467 | 32 / 16 / 659 | 183 / 4885 | - | - |
| 30 | 4023b | Terran | 42826 | 16257 | 192/358 | 530 | 471 / 472 | 194 / 350 | 90 / 26 / 822 | 410 / 5460 | - | - |
| 30 | Tvbxxx | Protoss | 6025 | 3698 | 0/0 | 0 | 410 / 554 | 78 / 87 | 4 / 139 / 641 | 7 / 2318 | 50256 | 50254 |


Evidence per replay:

12 matches in 1 files:

final.md:102:- 1 (suspicious): kdqj: never made [Science Vessel (14 orders)]
final.md:103:- 7 (clean): Horang2[._.]: 3759 unspent minerals at the end
final.md:104:- 8 (clean): Horang2[._.]: made 37 units for 246 train orders (207 rejected f...
final.md:105:- 9 (clean): Horang2[._.]: made 81 units for 329 train orders (238 rejected f...
final.md:106:- 20 (failed): engine stopped at frame 0 of 9374: 0 disable thingy unit
final.md:107:- 25 (clean): Horang2[._.]: made 25 units for 92 train orders (64 rejected fo...
final.md:108:- 26 (clean): 231314231313131: made 36 units for 252 train orders (212 reject...
final.md:109:- 27 (clean): 4023b: 58129 unspent minerals at the end
final.md:110:- 29 (desynced): selected or targeted units stop existing in the engine from ...
final.md:111:- 30 (desynced): selected or targeted units stop existing in the engine from ...

## Wide sample: desynced replays (98 replays)

| Replay | Date | Matchup | Minutes | First anomaly cluster (frame, minute) |
|---|---|---|---|---|
| 30-MORE-SST_JumJaJungJi/MM-EA8AF812 | 2026-03-19 | PvT | 17 | 9174, 6 |
| 11-MORE-F1SSasad11s32dd/MM-08118F44 | 2026-04-10 | ZvT | 45 | 6329, 4 |
| 30-MORE-4023b/MM-9C050330 | 2026-04-18 | TvP | 31 | 23262, 16 |
| 30-MORE-LYX2008/MM-5139D282 | 2026-04-19 | TvP | 32 | 37539, 26 |
| 30-MORE-Skins_/MM-029782AE | 2026-04-19 | ZvT | 8 | 5408, 3 |
| 30-MORE-4023b/MM-D50574B0 | 2026-04-21 | TvP | 58 | 45140, 31 |
| 30-MORE-4023b/MM-C02C2BC0 | 2026-04-21 | TvZ | 24 | 13832, 9 |
| 30-MORE-231314231313131/MM-F4A3BB2C | 2026-04-22 | ZvP | 47 | 44019, 30 |
| 30-MORE-4023b/MM-FFFD042C | 2026-04-23 | PvT | 57 | 29113, 20 |

| Group | Desynced |
|---|---|
| With a Terran (wide sample) | 8 / 44 |
| Without a Terran (wide sample) | 1 / 53 |
| Games over 25 minutes | 6 / 14 |
| Games up to 25 minutes | 3 / 83 |
| Terran games 2026-01-01 to 2026-03-14 | 9 / 59 |
| Terran games 2026-03-19 to 2026-04-23 | 6 / 56 |

The date does not matter: Terran games desync at a similar rate in both periods. Long games desync more often because they have more time to diverge.

## Failure patterns

1. **Map does not load (Pole Star 1.1).** `error 0 disable thingy unit`. The map has a `THG2` unit with the disabled flag `0x80`, which OpenBW does not support. This is 497 of 2,934 cwal.gg replays. Not a downgrade bug: clearing the flag would change the game.
2. **Gradual desync, Terran games.** Missing tags start in a few, then grow. Both players are affected after the onset. The desynced player shows the signs described in the task:
   - large unspent minerals (4023b: 42,826 and 46,524; F1SSasad11s32dd: 15,288; osolx;: 6,497),
   - unit types ordered many times but never made (F1SSasad11s32dd: 129 Siege Tank orders, 0 made; IIlllIIlIIlIll: 112 Hydralisk orders, 0 made),
   - many Train/Build commands whose selected unit cannot do them (541 and 454 in MM-08118F44).
3. **Early divergence is visible in unit allocation.** In MM-029782AE (ZvT), at frame 5399 Zerg selects 1.21 index 1513 and morphs a Drone. In OpenBW, index 1512 is a Terran SCV in training, so the select and the morph fail. At frame 5960 Zerg orders an Extractor; OpenBW builds it at about frame 6040, but in SC:R the player selects the same Drone again at 6072 and orders the Extractor again. So the game state already differed before frame 5400 (minute 3.8).
4. **No object limit is reached.** Peaks in the 30 replays: 811 units, 84 bullets, 1,001 sprites, 2,275 images, 247 orders. SC:R's `LMTS` limits (10,000 images, 5,000 sprites, 3,400 units, 400 bullets, 4,000 orders) are higher than OpenBW's, but no game came near OpenBW's limits.

## Most likely causes

| Cause | Assessment |
|---|---|
| Downgrade bug (tag mapping, dropped or misread commands) | **Unlikely.** 0 commands dropped. In 84 of 98 replays, every selected and targeted tag resolves for the whole game, including units reused 5-9 times (generation mod 32 matches). A tag or command layout error would show from the first minute in every game. |
| Harness bug | **No.** The harness and the probe agree on reject counts (for example 484 failed Train actions in both). |
| SC:R gameplay or data changes after 1.16 | **Most likely.** The divergence is in game state, not in commands, and is tied to Terran (8 of 9 desyncs). Every SC:R replay has `BFIX` = `04 00 00 00 00 00 00 00`, an undocumented bug-fix flag that 1.16 does not have. Lead: in the minute before onset, Terran players issued Unsiege in 42% of desynced games versus 19% of a clean-game baseline, and Siege in 50% versus 30%. Siege Tank mode switching (or another Terran mechanic) is the first thing to check against SC:R. |
| Unsupported commands | **No.** Every command ID in the sample is read by OpenBW. |
| OpenBW object limits lower than SC:R | **Not observed** (see pattern 4). It can matter for very large games. |

## Next steps

- Find the first divergent frame for a short desync (MM-029782AE, before frame 5400) with a frame-by-frame comparison against SC:R state, for example from a BWAPI/SC:R memory dump.
- Check SC:R `BFIX` bit 2 and Siege Tank behaviour changes against `bwgame.h`.
- Pole Star needs OpenBW support for disabled `THG2` units.
