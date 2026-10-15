#!/bin/sh
set -eu

repo_root=$(cd "$(dirname "$0")/../../.." && pwd)
replays=${REPLAYS:-${OPENBW_REPLAYS:-$repo_root/testdata/replays}/cwalggdl}
out=${OUT:-${TMPDIR:-/tmp}/openbw-integration}

make -C "$repo_root/tools/harness" harness >&2
make -C "$repo_root/tools/screp/integration/probe" probe >&2
go -C "$repo_root/tools/screp" build -o "$out/integration" ./integration
mkdir -p "$out"
cd "$repo_root"
"$out/integration" -replays "$replays" -out "$out" "$@"
