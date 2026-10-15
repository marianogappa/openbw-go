# bwdata-export

Extracts the game data files that the engine needs from your own StarCraft: Remastered install. This repository does not include them.

1. Install [CascLib](https://github.com/ladislav-zezula/CascLib) and build the extractor:

   ```
   c++ -std=c++17 -I<casclib>/src -o tools/bwdata-export/casc-tools/casc_extract_one \
       tools/bwdata-export/casc-tools/casc_extract_one.cpp -L<casclib-build> -lcasc
   ```

2. Run the export. It writes the classic (SD) files to `testdata/bwdata` (git-ignored):

   ```
   python3 tools/bwdata-export/export_bwdata.py --casc "/Applications/StarCraft"
   ```

   `--casc` is the SC:R install folder. `--out` changes the output folder, and `--zip` also writes a zip.

3. Use the folder: it is the default of the tests and tools, or set `OPENBW_DATA` to it.

`casc_list.cpp` lists the files in a CASC storage, for when a path changes.
