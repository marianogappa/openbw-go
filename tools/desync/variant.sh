#!/bin/sh
set -eu

if [ $# -lt 2 ]; then
	echo "usage: variant.sh openbw_dir out_dir [integration flags...]" >&2
	exit 2
fi

openbw=$(cd "$1" && pwd)
out=$2
shift 2

repo_root=$(cd "$(dirname "$0")/../.." && pwd)
replays=${REPLAYS:-${OPENBW_REPLAYS:-$repo_root/testdata/replays}/cwalggdl}
mkdir -p "$out/bin"

c++ -std=c++17 -O2 -w -isystem "$openbw" -o "$out/bin/harness" "$repo_root/tools/harness/main.cpp"
c++ -std=c++17 -O2 -w -isystem "$openbw" -o "$out/bin/probe" "$repo_root/tools/screp/integration/probe/probe.cpp"
go -C "$repo_root/tools/screp" build -o "$out/bin/integration" ./integration
cd "$repo_root"
"$out/bin/integration" -replays "$replays" -out "$out" -harness "$out/bin/harness" -probe "$out/bin/probe" "$@"
