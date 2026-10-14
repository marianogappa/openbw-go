#!/bin/sh
set -eu

repo_root=$(cd "$(dirname "$0")/../.." && pwd)
replays=${REPLAYS:-$HOME/Code/go/src/github.com/marianogappa/cwalggdl}
out=${OUT:-${TMPDIR:-/tmp}/openbw-integration}

make -C "$repo_root/tools/harness" harness >&2
make -C "$repo_root/tools/integration/probe" probe >&2
cd "$repo_root"
go run ./tools/integration -replays "$replays" -out "$out" "$@"
