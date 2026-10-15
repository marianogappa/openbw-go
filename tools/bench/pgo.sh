#!/bin/sh
set -eu

repo_root=$(cd "$(dirname "$0")/../.." && pwd)
replay_list=${REPLAYS:-$repo_root/tools/bench/replays.txt}
replays_dir=${OPENBW_REPLAYS:-$repo_root/testdata/replays}
work=${OUT:-${TMPDIR:-/tmp}/openbw-pgo}

mkdir -p "$work"
cd "$repo_root"
go build -pgo=off -trimpath -o "$work/obwhash" ./cmd/obwhash
go build -o "$work/downgrade" ./cmd/downgrade

rm -f "$work"/*.pprof
sed -e '/^#/d' -e '/^$/d' "$replay_list" | while IFS= read -r entry; do
	replay="$replays_dir/$entry"
	name=$(basename "$replay" .rep)
	"$work/downgrade" "$replay" "$work/$name.116.rep" 2>/dev/null
	"$work/obwhash" -every 1000000000 -cpuprofile "$work/$name.pprof" "$work/$name.116.rep" >/dev/null 2>&1
done
go tool pprof -proto "$work"/*.pprof > "$repo_root/cmd/obwhash/default.pgo"
