package main

import (
	"encoding/binary"
	"fmt"
	"os"
	"sort"

	"github.com/marianogappa/openbw-go/downgrade"
)

const (
	unitEntrySize    = 36
	thg2EntrySize    = 10
	flagDrawAsSprite = 0x1000
	flagDisabled     = 0x8000
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: mapinfo legacy.rep")
		os.Exit(2)
	}
	if err := run(os.Args[1]); err != nil {
		fmt.Fprintln(os.Stderr, "mapinfo:", err)
		os.Exit(1)
	}
}

func run(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	rep, err := downgrade.ReadUncompressedLegacy(data)
	if err != nil {
		return err
	}
	printChunkList(rep.MapData)
	printTileDifferences(rep.MapData)
	for _, chunk := range chunksNamed(rep.MapData, "ERA ") {
		if len(chunk) >= 2 {
			fmt.Printf("tileset %d\n", binary.LittleEndian.Uint16(chunk)&7)
		}
	}
	for _, chunk := range chunksNamed(rep.MapData, "UNIT") {
		printUnitCounts(chunk)
	}
	for _, chunk := range chunksNamed(rep.MapData, "THG2") {
		printEntries(chunk)
	}
	return nil
}

func printChunkList(chk []byte) {
	for pos := 0; pos+8 <= len(chk); {
		size := int(int32(binary.LittleEndian.Uint32(chk[pos+4:])))
		fmt.Printf("chunk %q size=%d\n", string(chk[pos:pos+4]), size)
		end := pos + 8 + size
		if size < 0 || end > len(chk) {
			break
		}
		pos = end
	}
}

func printTileDifferences(chk []byte) {
	mtxm := chunksNamed(chk, "MTXM")
	tile := chunksNamed(chk, "TILE")
	if len(mtxm) == 0 || len(tile) == 0 {
		return
	}
	a, b := mtxm[len(mtxm)-1], tile[len(tile)-1]
	differences := 0
	for i := 0; i+1 < len(a) && i+1 < len(b); i += 2 {
		if binary.LittleEndian.Uint16(a[i:]) != binary.LittleEndian.Uint16(b[i:]) {
			differences++
		}
	}
	fmt.Printf("mtxm_tile_differences %d\n", differences)
}

func printUnitCounts(chunk []byte) {
	counts := map[[2]int]int{}
	for pos := 0; pos+unitEntrySize <= len(chunk); pos += unitEntrySize {
		e := chunk[pos : pos+unitEntrySize]
		id := int(binary.LittleEndian.Uint16(e[8:]))
		owner := int(e[16])
		resources := int(binary.LittleEndian.Uint32(e[20:]))
		x := int(binary.LittleEndian.Uint16(e[4:]))
		y := int(binary.LittleEndian.Uint16(e[6:]))
		validProperties := binary.LittleEndian.Uint16(e[12:])
		validFlags := binary.LittleEndian.Uint16(e[14:])
		hpPercent, shieldPercent, energyPercent := e[17], e[18], e[19]
		stateFlags := binary.LittleEndian.Uint16(e[26:])
		fmt.Printf("map_unit type=%d owner=%d x=%d y=%d resources=%d properties=%#x valid_flags=%#x flags=%#x hp=%d shield=%d energy=%d\n",
			id, owner, x, y, resources, validProperties, validFlags, stateFlags, hpPercent, shieldPercent, energyPercent)
		counts[[2]int{id, owner}]++
	}
	keys := make([][2]int, 0, len(counts))
	for k := range counts {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		return keys[i][0] < keys[j][0] || keys[i][0] == keys[j][0] && keys[i][1] < keys[j][1]
	})
	for _, k := range keys {
		fmt.Printf("unit type=%d owner=%d count=%d\n", k[0], k[1], counts[k])
	}
}

func chunksNamed(chk []byte, name string) [][]byte {
	var chunks [][]byte
	for pos := 0; pos+8 <= len(chk); {
		id := string(chk[pos : pos+4])
		size := int(int32(binary.LittleEndian.Uint32(chk[pos+4:])))
		start := pos + 8
		end := start + size
		if size < 0 || end > len(chk) {
			break
		}
		if id == name {
			chunks = append(chunks, chk[start:end])
		}
		pos = end
	}
	return chunks
}

func printEntries(chunk []byte) {
	for pos := 0; pos+thg2EntrySize <= len(chunk); pos += thg2EntrySize {
		e := chunk[pos : pos+thg2EntrySize]
		id := binary.LittleEndian.Uint16(e[0:])
		x := binary.LittleEndian.Uint16(e[2:])
		y := binary.LittleEndian.Uint16(e[4:])
		owner := e[6]
		flags := binary.LittleEndian.Uint16(e[8:])
		kind := "unit"
		if flags&flagDrawAsSprite != 0 {
			kind = "sprite"
		}
		disabled := ""
		if flags&flagDisabled != 0 {
			disabled = " disabled"
		}
		fmt.Printf("%s id=%d x=%d y=%d owner=%d flags=%#04x%s\n", kind, id, x, y, owner, flags, disabled)
	}
}
