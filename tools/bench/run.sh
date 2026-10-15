#!/bin/sh
set -eu

repo_root=$(cd "$(dirname "$0")/../.." && pwd)
replay_list=${REPLAYS:-$repo_root/tools/bench/replays.txt}
replays_dir=${OPENBW_REPLAYS:-$repo_root/testdata/replays}
runs=${RUNS:-3}
work=${OUT:-${TMPDIR:-/tmp}/openbw-bench}
every=1000000000

mkdir -p "$work"
make -C "$repo_root/tools/harness" harness >/dev/null
cd "$repo_root"
go build -o "$work/obwhash" ./cmd/obwhash
go build -o "$work/downgrade" ./cmd/downgrade

seconds_of_one_run() {
	perl -MTime::HiRes=time -e 'open(my $result, ">&", STDOUT); open(STDOUT, ">", "/dev/null"); open(STDERR, ">", "/dev/null"); $start = time; system(@ARGV); printf $result "%.3f\n", time - $start' "$@"
}

best_seconds() {
	best=""
	i=0
	while [ "$i" -lt "$runs" ]; do
		seconds=$(seconds_of_one_run "$@")
		best=$(awk -v best="$best" -v s="$seconds" 'BEGIN { print (best == "" || s < best) ? s : best }')
		i=$((i + 1))
	done
	echo "$best"
}

frames_of() {
	"$work/obwhash" -every "$every" "$1" 2>/dev/null | awk '$1 == "end" { print $2 }'
}

printf "%-40s %8s %8s %9s %8s %9s %7s\n" replay frames cpp_s cpp_fps go_s go_fps go/cpp
sed -e '/^#/d' -e '/^$/d' "$replay_list" | while IFS= read -r entry; do
	replay="$replays_dir/$entry"
	name=$(basename "$replay" .rep)
	legacy="$work/$name.116.rep"
	"$work/downgrade" "$replay" "$legacy" 2>/dev/null
	frames=$(frames_of "$legacy")
	cpp=$(best_seconds "$repo_root/tools/harness/harness" -every "$every" "$legacy")
	go=$(best_seconds "$work/obwhash" -every "$every" "$legacy")
	awk -v n="$name" -v f="$frames" -v c="$cpp" -v g="$go" 'BEGIN {
		printf "%-40s %8d %8.2f %9.0f %8.2f %9.0f %7.2f\n", n, f, c, f / c, g, f / g, g / c
	}'
done
