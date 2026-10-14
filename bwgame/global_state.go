package bwgame

import (
	"os"
	"path/filepath"
	"strings"
)

type GlobalState struct {
	FlingyTypes FlingyTypesT
	SpriteTypes SpriteTypesT
	ImageTypes  ImageTypesT
	OrderTypes  OrderTypesT
	Iscript     IscriptT

	Grps           []GrpT
	ImageGrp       []*GrpT
	LoOffsets      [][][]Xy
	ImageLoOffsets [][6]*[][]Xy

	UnitsDat    []byte
	WeaponsDat  []byte
	UpgradesDat []byte
	TechdataDat []byte

	MeleeTrg []byte

	TilesetVf4 [8][]byte
	TilesetCv5 [8][]byte
}

type LoadDataFileFunc func(path string) ([]byte, error)

func DataPathToLowerSlash(path string) string {
	return strings.ToLower(strings.ReplaceAll(path, "\\", "/"))
}

func DataFilesDirectory(dir string) LoadDataFileFunc {
	return func(path string) ([]byte, error) {
		return os.ReadFile(filepath.Join(dir, filepath.FromSlash(DataPathToLowerSlash(path))))
	}
}

func GlobalInitFromDirectory(dir string) (*GlobalState, error) {
	return GlobalInit(DataFilesDirectory(dir))
}

func GlobalInit(loadFile func(path string) ([]byte, error)) (st *GlobalState, err error) {
	defer recoverDataLoadingError(&err)
	st = &GlobalState{}
	globalInit(st, loadFile)
	return st, nil
}

func ReadGrp(r DataReader) (grp GrpT, err error) {
	defer recoverDataLoadingError(&err)
	return readGrp(r), nil
}

func readGrp(r DataReader) GrpT {
	baseR := r
	var grp GrpT
	frameCount := uint(r.GetU16())
	grp.Width = uint(r.GetU16())
	grp.Height = uint(r.GetU16())
	grp.Frames = make([]GrpTFrameT, frameCount)
	for i := range grp.Frames {
		f := &grp.Frames[i]
		f.Offset.X = uint(r.GetU8())
		f.Offset.Y = uint(r.GetU8())
		f.Size.X = uint(r.GetU8())
		f.Size.Y = uint(r.GetU8())
		fileOffset := uint(r.GetU32())
		lineOffsetR := baseR
		lineOffsetR.Skip(fileOffset)
		f.LineDataOffset = make([]uint, 0, f.Size.Y)
		for y := uint(0); y != f.Size.Y; y++ {
			lineR := baseR
			lineR.Skip(fileOffset + uint(lineOffsetR.GetU16()))
			f.LineDataOffset = append(f.LineDataOffset, uint(len(f.DataContainer)))
			readGrpLine(&lineR, f)
		}
	}
	return grp
}

func readGrpLine(lineR *DataReader, f *GrpTFrameT) {
	for x := uint(0); x != f.Size.X; {
		v := lineR.GetU8()
		remaining := f.Size.X - x
		switch {
		case v&0x80 != 0:
			v = clampGrpRun(v&0x7f, remaining)
			f.DataContainer = append(f.DataContainer, 0x80|v)
		case v&0x40 != 0:
			v = clampGrpRun(v&0x3f, remaining)
			f.DataContainer = append(f.DataContainer, 0x40|v, lineR.GetU8())
		default:
			v = clampGrpRun(v, remaining)
			f.DataContainer = append(f.DataContainer, v)
			for i := uint8(0); i != v; i++ {
				f.DataContainer = append(f.DataContainer, lineR.GetU8())
			}
		}
		x += uint(v)
	}
}

func clampGrpRun(v uint8, remaining uint) uint8 {
	if uint(v) > remaining {
		return uint8(remaining)
	}
	return v
}

type globalInitContext struct {
	st           *GlobalState
	loadDataFile LoadDataFileFunc
}

func globalInit(st *GlobalState, loadDataFile LoadDataFileFunc) {
	c := &globalInitContext{st: st, loadDataFile: loadDataFile}

	st.UnitsDat = c.load("arr/units.dat")
	st.WeaponsDat = c.load("arr/weapons.dat")
	st.UpgradesDat = c.load("arr/upgrades.dat")
	st.TechdataDat = c.load("arr/techdata.dat")

	st.MeleeTrg = c.load("triggers/Melee.trg")

	st.FlingyTypes = mustLoad(LoadFlingyDat(c.load("arr/flingy.dat")))
	st.SpriteTypes = mustLoad(LoadSpritesDat(c.load("arr/sprites.dat")))
	st.ImageTypes = mustLoad(LoadImagesDat(c.load("arr/images.dat")))
	st.OrderTypes = mustLoad(LoadOrdersDat(c.load("arr/orders.dat")))

	for i := range st.FlingyTypes.Vec {
		c.fixupSpriteType(&st.FlingyTypes.Vec[i].Sprite)
	}
	for i := range st.SpriteTypes.Vec {
		c.fixupImageType(&st.SpriteTypes.Vec[i].Image)
	}

	c.loadIscriptBin()
	c.loadImages()

	tilesetNames := [8]string{"badlands", "platform", "install", "AshWorld", "Jungle", "Desert", "Ice", "Twilight"}
	for i, name := range tilesetNames {
		st.TilesetVf4[i] = c.load("Tileset/" + name + ".vf4")
		st.TilesetCv5[i] = c.load("Tileset/" + name + ".cv5")
	}
}

func mustLoad[T any](value T, err error) T {
	if err != nil {
		panic(err)
	}
	return value
}

func (c *globalInitContext) load(path string) []byte {
	data, err := c.loadDataFile(path)
	if err != nil {
		throwDataLoadingError("%s: %v", path, err)
	}
	return data
}

func (c *globalInitContext) getSpriteType(id SpriteTypes) *SpriteTypeT {
	if uint(id) >= 517 {
		throwDataLoadingError("invalid sprite id %d", uint(id))
	}
	return &c.st.SpriteTypes.Vec[id]
}

func (c *globalInitContext) getImageType(id ImageTypes) *ImageTypeT {
	if uint(id) >= 999 {
		throwDataLoadingError("invalid image id %d", uint(id))
	}
	return &c.st.ImageTypes.Vec[id]
}

func (c *globalInitContext) fixupSpriteType(t *TypeID[SpriteTypeT, SpriteTypes]) {
	if t.ID == SpriteTypesNone {
		t.Ptr = nil
		return
	}
	t.Ptr = c.getSpriteType(t.ID)
}

func (c *globalInitContext) fixupImageType(t *TypeID[ImageTypeT, ImageTypes]) {
	if t.ID == ImageTypesNone {
		t.Ptr = nil
		return
	}
	t.Ptr = c.getImageType(t.ID)
}

var iscriptInstructionData = [69]string{
	IscriptOpcodesOpcPlayfram:          "2",
	IscriptOpcodesOpcPlayframtile:      "2",
	IscriptOpcodesOpcSethorpos:         "s1",
	IscriptOpcodesOpcSetvertpos:        "s1",
	IscriptOpcodesOpcSetpos:            "s1s1",
	IscriptOpcodesOpcWait:              "1",
	IscriptOpcodesOpcWaitrand:          "11",
	IscriptOpcodesOpcGoto:              "j",
	IscriptOpcodesOpcImgol:             "211",
	IscriptOpcodesOpcImgul:             "211",
	IscriptOpcodesOpcImgolorig:         "2",
	IscriptOpcodesOpcSwitchul:          "2",
	IscriptOpcodesOpc0c:                "",
	IscriptOpcodesOpcImgoluselo:        "211",
	IscriptOpcodesOpcImguluselo:        "211",
	IscriptOpcodesOpcSprol:             "2s1s1",
	IscriptOpcodesOpcHighsprol:         "211",
	IscriptOpcodesOpcLowsprul:          "211",
	IscriptOpcodesOpcUflunstable:       "2",
	IscriptOpcodesOpcSpruluselo:        "211",
	IscriptOpcodesOpcSprul:             "211",
	IscriptOpcodesOpcSproluselo:        "21",
	IscriptOpcodesOpcEnd:               "e",
	IscriptOpcodesOpcSetflipstate:      "1",
	IscriptOpcodesOpcPlaysnd:           "2",
	IscriptOpcodesOpcPlaysndrand:       "v",
	IscriptOpcodesOpcPlaysndbtwn:       "22",
	IscriptOpcodesOpcDomissiledmg:      "",
	IscriptOpcodesOpcAttackmelee:       "v",
	IscriptOpcodesOpcFollowmaingraphic: "",
	IscriptOpcodesOpcRandcondjmp:       "1b",
	IscriptOpcodesOpcTurnccwise:        "1",
	IscriptOpcodesOpcTurncwise:         "1",
	IscriptOpcodesOpcTurn1cwise:        "",
	IscriptOpcodesOpcTurnrand:          "1",
	IscriptOpcodesOpcSetspawnframe:     "1",
	IscriptOpcodesOpcSigorder:          "1",
	IscriptOpcodesOpcAttackwith:        "1",
	IscriptOpcodesOpcAttack:            "",
	IscriptOpcodesOpcCastspell:         "",
	IscriptOpcodesOpcUseweapon:         "1",
	IscriptOpcodesOpcMove:              "1",
	IscriptOpcodesOpcGotorepeatattk:    "",
	IscriptOpcodesOpcEngframe:          "1",
	IscriptOpcodesOpcEngset:            "1",
	IscriptOpcodesOpc2d:                "",
	IscriptOpcodesOpcNobrkcodestart:    "",
	IscriptOpcodesOpcNobrkcodeend:      "",
	IscriptOpcodesOpcIgnorerest:        "",
	IscriptOpcodesOpcAttkshiftproj:     "1",
	IscriptOpcodesOpcTmprmgraphicstart: "",
	IscriptOpcodesOpcTmprmgraphicend:   "",
	IscriptOpcodesOpcSetfldirect:       "1",
	IscriptOpcodesOpcCall:              "b",
	IscriptOpcodesOpcReturn:            "",
	IscriptOpcodesOpcSetflspeed:        "2",
	IscriptOpcodesOpcCreategasoverlays: "1",
	IscriptOpcodesOpcPwrupcondjmp:      "b",
	IscriptOpcodesOpcTrgtrangecondjmp:  "2b",
	IscriptOpcodesOpcTrgtarccondjmp:    "22b",
	IscriptOpcodesOpcCurdirectcondjmp:  "22b",
	IscriptOpcodesOpcImgulnextid:       "11",
	IscriptOpcodesOpc3e:                "",
	IscriptOpcodesOpcLiftoffcondjmp:    "b",
	IscriptOpcodesOpcWarpoverlay:       "2",
	IscriptOpcodesOpcOrderdone:         "1",
	IscriptOpcodesOpcGrdsprol:          "211",
	IscriptOpcodesOpc43:                "",
	IscriptOpcodesOpcDogrddamage:       "",
}

const iscriptOpcodeBias = 0x808091

type iscriptBranch struct {
	address      uint
	programIndex uint
}

type iscriptDecoder struct {
	baseR       DataReader
	programData []int32
	decodeMap   map[uint]uint
	branches    []iscriptBranch
}

func (c *globalInitContext) loadIscriptBin() {
	scripts := map[int32]*IscriptTScript{}
	programData := []int32{0}

	data := c.load("scripts/iscript.bin")
	baseR := NewDataReaderLe(data)
	r := baseR
	idListOffset := uint(r.GetU32())
	r.Seek(idListOffset)
	for r.Left() != 0 {
		id := int32(r.GetI16())
		if id == -1 {
			break
		}
		scriptAddress := uint(r.GetU16())
		scriptR := baseR
		scriptR.Skip(scriptAddress)
		scriptR.GetN(4)

		decoder := &iscriptDecoder{baseR: baseR, programData: programData, decodeMap: map[uint]uint{}}

		script := scripts[id]
		if script == nil {
			script = &IscriptTScript{ID: id}
			scripts[id] = script
		}

		highestAnimation := uint(scriptR.GetU32())
		animations := (highestAnimation + 1 + 1) &^ 1
		for i := uint(0); i != animations; i++ {
			animAddress := uint(scriptR.GetU16())
			if animAddress == 0 {
				script.AnimationPc = append(script.AnimationPc, 0)
				continue
			}
			animR := baseR
			animR.Skip(animAddress)
			script.AnimationPc = append(script.AnimationPc, decoder.decodeAt(animAddress))
		}
		programData = decoder.programData
	}

	c.st.Iscript = IscriptT{Scripts: scripts, ProgramData: programData}
}

func (d *iscriptDecoder) decodeAt(initialAddress uint) uint {
	d.branches = nil
	initialPc := d.decode(initialAddress)
	for len(d.branches) != 0 {
		v := d.branches[0]
		d.branches = d.branches[1:]
		pc := d.decode(v.address)
		if !fitsInInt32(pc) {
			throwDataLoadingError("iscript load: 0x%x does not fit in an int", pc)
		}
		d.programData[v.programIndex] = int32(pc)
	}
	return initialPc
}

func (d *iscriptDecoder) decode(initialAddress uint) uint {
	if initialAddress == 0 {
		throwDataLoadingError("iscript load: attempt to decode instruction at null address")
	}
	if pc, ok := d.decodeMap[initialAddress]; ok {
		return pc
	}
	initialPc := uint(len(d.programData))
	d.decodeMap[initialAddress] = initialPc
	r := d.baseR
	r.Skip(initialAddress)
	done := false
	for !done {
		pc := uint(len(d.programData))
		curAddress := r.Tell()
		if !fitsInInt32(pc) || !fitsInInt32(curAddress) {
			throwDataLoadingError("iscript too big")
		}
		if curAddress != initialAddress {
			if existingPc, ok := d.decodeMap[curAddress]; ok {
				d.push(IscriptOpcodesOpcGoto+iscriptOpcodeBias, int32(existingPc))
				break
			}
			d.decodeMap[curAddress] = pc
		}
		opcode := int(r.GetU8())
		if opcode >= len(iscriptInstructionData) {
			throwDataLoadingError("iscript load: at 0x%04x: invalid instruction %d", curAddress, opcode)
		}
		d.push(int32(opcode + iscriptOpcodeBias))
		done = d.decodeOperands(&r, iscriptInstructionData[opcode])
	}
	return initialPc
}

func (d *iscriptDecoder) decodeOperands(r *DataReader, format string) bool {
	done := false
	for c := 0; c < len(format); c++ {
		switch format[c] {
		case 's':
			c++
			if format[c] == '1' {
				d.push(int32(r.GetI8()))
			} else if format[c] == '2' {
				d.push(int32(r.GetI16()))
			}
		case '1':
			d.push(int32(r.GetU8()))
		case '2':
			d.push(int32(r.GetU16()))
		case 'v':
			n := int32(r.GetU8())
			d.push(n)
			for ; n != 0; n-- {
				d.push(int32(r.GetU16()))
			}
		case 'j':
			done = d.decodeJump(r)
		case 'b':
			branchAddress := uint(r.GetU16())
			d.branches = append(d.branches, iscriptBranch{address: branchAddress, programIndex: uint(len(d.programData))})
			d.push(0)
		case 'e':
			done = true
		}
	}
	return done
}

func (d *iscriptDecoder) decodeJump(r *DataReader) bool {
	jumpAddress := uint(r.GetU16())
	jumpPc, ok := d.decodeMap[jumpAddress]
	if !ok {
		d.programData = d.programData[:len(d.programData)-1]
		*r = d.baseR
		r.Skip(jumpAddress)
		return false
	}
	d.push(int32(jumpPc))
	return true
}

func (d *iscriptDecoder) push(values ...int32) {
	d.programData = append(d.programData, values...)
}

func fitsInInt32(v uint) bool {
	return uint(int32(v)) == v
}

type imagesLoader struct {
	c         *globalInitContext
	baseR     DataReader
	grps      []GrpT
	loOffsets [][][]Xy
	loaded    map[int32]uint
}

func (c *globalInitContext) loadImages() {
	data := c.load("arr/images.tbl")
	l := &imagesLoader{c: c, baseR: NewDataReaderLe(data), loaded: map[int32]uint{}}

	r := l.baseR
	r.GetU16()

	l.grps = append(l.grps, GrpT{})
	l.loOffsets = append(l.loOffsets, nil)

	var imageGrpIndex []uint
	var loIndices [6][]uint
	for i := 0; i != 999; i++ {
		imageType := c.getImageType(ImageTypes(i))
		imageGrpIndex = append(imageGrpIndex, l.load(imageType.GrpFilenameIndex, l.loadGrp))
		loIndices[0] = append(loIndices[0], l.load(imageType.AttackFilenameIndex, l.loadOffsets))
		loIndices[1] = append(loIndices[1], l.load(imageType.DamageFilenameIndex, l.loadOffsets))
		loIndices[2] = append(loIndices[2], l.load(imageType.SpecialFilenameIndex, l.loadOffsets))
		loIndices[3] = append(loIndices[3], l.load(imageType.LandingDustFilenameIndex, l.loadOffsets))
		loIndices[4] = append(loIndices[4], l.load(imageType.LiftOffFilenameIndex, l.loadOffsets))
		loIndices[5] = append(loIndices[5], l.load(imageType.ShieldFilenameIndex, l.loadOffsets))
	}

	st := c.st
	st.Grps = l.grps
	st.ImageGrp = make([]*GrpT, len(imageGrpIndex))
	for i, index := range imageGrpIndex {
		st.ImageGrp[i] = &st.Grps[index]
	}
	st.LoOffsets = l.loOffsets
	st.ImageLoOffsets = make([][6]*[][]Xy, 999)
	for i := 0; i != 6; i++ {
		for i2 := 0; i2 != 999; i2++ {
			st.ImageLoOffsets[i2][i] = &st.LoOffsets[loIndices[i][i2]]
		}
	}
}

func (l *imagesLoader) load(index int32, f func(DataReader) uint) uint {
	if index == 0 {
		return 0
	}
	if loadedIndex, ok := l.loaded[index]; ok {
		return loadedIndex
	}
	l.loaded[index] = 0
	r := l.baseR
	r.Skip(uint(int64(2 + (index-1)*2)))
	fnOffset := uint(r.GetU16())
	r = l.baseR
	r.Skip(fnOffset)
	fn := readCString(&r)

	data := l.c.load("unit/" + fn)
	loadedIndex := f(NewDataReaderLe(data))
	l.loaded[index] = loadedIndex
	return loadedIndex
}

func readCString(r *DataReader) string {
	var b []byte
	for c := r.GetU8(); c != 0; c = r.GetU8() {
		b = append(b, c)
	}
	return string(b)
}

func (l *imagesLoader) loadGrp(r DataReader) uint {
	index := uint(len(l.grps))
	l.grps = append(l.grps, readGrp(r))
	return index
}

func (l *imagesLoader) loadOffsets(r DataReader) uint {
	baseR := r
	frameCount := uint(r.GetU32())
	offsetCount := uint(r.GetU32())
	var offs [][]Xy
	for f := uint(0); f < frameCount; f++ {
		fileOffset := uint(r.GetU32())
		r2 := baseR
		r2.Skip(fileOffset)
		vec := make([]Xy, offsetCount)
		for i := range vec {
			x := int32(r2.GetI8())
			y := int32(r2.GetI8())
			vec[i] = Xy{X: x, Y: y}
		}
		offs = append(offs, vec)
	}
	l.loOffsets = append(l.loOffsets, offs)
	return uint(len(l.loOffsets) - 1)
}
