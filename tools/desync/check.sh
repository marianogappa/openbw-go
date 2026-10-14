#!/bin/sh
set -eu

if [ "${1:-}" = "-one" ]; then
	result=$("$TRACER" -check -no-units "$2" | awk '/^first_inconsistent/ {print $2, $3} /^end/ {print $2} /^error/ {print "error", $2}' | tr '\n' ' ')
	echo "$(basename "$2" .116.rep) $result"
	exit 0
fi

if [ $# -lt 3 ]; then
	echo "usage: check.sh openbw_dir out_file replay.116.rep..." >&2
	exit 2
fi

openbw=$(cd "$1" && pwd)
out=$2
shift 2

repo_root=$(cd "$(dirname "$0")/../.." && pwd)
TRACER=$(mktemp "${TMPDIR:-/tmp}/tracer.XXXXXX")
c++ -std=c++17 -O2 -w -isystem "$openbw" -o "$TRACER" "$repo_root/tools/desync/tracer.cpp"
export OPENBW_DATA="${OPENBW_DATA:-$repo_root/testdata/bwdata}"
export TRACER

printf '%s\n' "$@" | xargs -P "${WORKERS:-8}" -n 1 "$0" -one | sort > "$out"
rm -f "$TRACER"
