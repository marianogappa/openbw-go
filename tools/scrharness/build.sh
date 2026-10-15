#!/bin/sh
set -eu

repo_root=$(cd "$(dirname "$0")/../.." && pwd)
out=${1:-${TMPDIR:-/tmp}/scrharness}
openbw="$out/openbw"

rm -rf "$openbw"
mkdir -p "$out"
cp -R "$repo_root/ref/openbw" "$openbw"
rm -rf "$openbw/.git"
patch -s -d "$openbw" -p1 < "$repo_root/tools/scrharness/openbw-scr.patch"
${CXX:-clang++} -std=c++17 -O2 -w -isystem "$openbw" -o "$out/harness" "$repo_root/tools/harness/main.cpp"
echo "$out/harness"
