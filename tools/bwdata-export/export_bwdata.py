#!/usr/bin/env python3
"""Extract the classic (SD) StarCraft data files OpenBW needs from a StarCraft: Remastered CASC install."""
import argparse
import shutil
import struct
import subprocess
import sys
import zipfile
from pathlib import Path

HERE = Path(__file__).resolve().parent
DEFAULT_TOOL = HERE.parent / "tools" / "casc_extract_one" / "casc_extract_one"

DAT_FILES = ["units", "weapons", "upgrades", "techdata", "flingy", "sprites", "images", "orders"]
TILESETS = ["badlands", "platform", "install", "AshWorld", "Jungle", "Desert", "Ice", "Twilight"]


def fixed_files():
    files = [f"arr/{name}.dat" for name in DAT_FILES]
    files += ["arr/images.tbl", "scripts/iscript.bin", "triggers/Melee.trg"]
    files += [f"Tileset/{t}.{ext}" for t in TILESETS for ext in ("cv5", "vf4")]
    return files


def read_tbl(data):
    count = struct.unpack_from("<H", data, 0)[0]
    offsets = struct.unpack_from(f"<{count}H", data, 2)
    return [data[o:data.index(b"\0", o)].decode("latin-1") for o in offsets]


def extract(tool, casc_root, inner_path, out_path):
    out_path.parent.mkdir(parents=True, exist_ok=True)
    casc_path = inner_path.replace("/", "\\")
    result = subprocess.run([str(tool), str(casc_root), casc_path, str(out_path)],
                            capture_output=True, text=True)
    return result.returncode == 0 and out_path.stat().st_size > 0


def main():
    ap = argparse.ArgumentParser(description=__doc__)
    ap.add_argument("--casc", default="/Applications/StarCraft", help="SC:R install folder")
    ap.add_argument("--tool", default=str(DEFAULT_TOOL), help="path to casc_extract_one binary")
    ap.add_argument("--out", default=str(HERE / "bwdata"), help="output folder")
    ap.add_argument("--zip", default=str(HERE / "bwdata.zip"), help="output zip")
    args = ap.parse_args()

    out = Path(args.out)
    if out.exists():
        shutil.rmtree(out)
    failed = []

    for rel in fixed_files():
        if not extract(args.tool, args.casc, rel, out / rel.lower()):
            failed.append(rel)

    images_tbl = out / "arr/images.tbl"
    if not images_tbl.exists():
        sys.exit("arr/images.tbl missing; cannot continue")
    # images.tbl entries are relative to unit\ and include both .grp and .lo? files.
    # BW paths are case-insensitive and images.tbl mixes cases (Zerg\ vs zerg\), so store everything lowercase.
    entries = sorted({e.replace("\\", "/").lower() for e in read_tbl(images_tbl.read_bytes()) if e})
    for entry in entries:
        rel = f"unit/{entry}"
        if not extract(args.tool, args.casc, rel, out / rel):
            failed.append(rel)

    files = sorted(p for p in out.rglob("*") if p.is_file())
    with zipfile.ZipFile(args.zip, "w", zipfile.ZIP_DEFLATED) as z:
        for p in files:
            z.write(p, p.relative_to(out).as_posix())

    print(f"extracted {len(files)} files ({len(entries)} from images.tbl) -> {args.zip}")
    if failed:
        print(f"{len(failed)} failed:", *failed, sep="\n  ")
        sys.exit(1)


if __name__ == "__main__":
    main()
