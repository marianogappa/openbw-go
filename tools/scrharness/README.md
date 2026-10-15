# scrharness

Builds `tools/harness` against a copy of `ref/openbw` with the SC:R rules, to check `obwhash -scr-rules` frame by frame.

```
tools/scrharness/build.sh $TMPDIR/scrharness
go run ./tools/difftest -out $TMPDIR/scr-difftest -harness $TMPDIR/scrharness/harness -obwhash-args "-scr-rules" -list tools/difftest/replays.txt
```

`build.sh [out_dir]` copies `ref/openbw` to `out_dir/openbw`, applies `openbw-scr.patch` and writes `out_dir/harness`. The default `out_dir` is `$TMPDIR/scrharness`.

`openbw-scr.patch` is the diff in [DESYNC.md](../../DESYNC.md) ("Exact change") plus the rest of the rules:

- Object pools with the cwal.gg `LMTS` limits: bullets 400, sprites 5000, images 10000, orders 4000. The unit pool stays 1700.
- Lone sprites (`new_thingy`): 1000 instead of 500.
- `issue_order` queues an order only below 3600 active orders (90% of 4000) instead of 1800.
- `can_allocate_order` in `set_queued_order`: the `BFIX`/`GCFG` order allocation rule with the flying SCV fix.
- Disabled `THG2` units get the disabled and no-collide status flags instead of `disable thingy unit`.
- The `state_copier` arrays get the new pool sizes.

The limits are compile-time constants, so the patched harness matches `-scr-rules` only for replays with these settings. All 2,934 cwal.gg replays have them. `obwhash -scr-rules` prints the settings it uses to stderr.
