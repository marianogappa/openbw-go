package bwgame

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"testing"
)

const cppReferenceDumpSha256 = "f6213fb740175a589a023e358b01c493e0ebf021e29a8a45c899267d63ff905e"

func testDataDir(t *testing.T) string {
	t.Helper()
	dir := os.Getenv("OPENBW_DATA")
	if dir == "" {
		dir = filepath.Join("..", "testdata", "bwdata")
	}
	if _, err := os.Stat(filepath.Join(dir, "arr", "units.dat")); err != nil {
		t.Skipf("game data not found in %s", dir)
	}
	return dir
}

var (
	loadGlobalStateOnce sync.Once
	loadedGlobalState   *GlobalState
	loadGlobalStateErr  error
)

func testGlobalState(t *testing.T) *GlobalState {
	t.Helper()
	dir := testDataDir(t)
	loadGlobalStateOnce.Do(func() {
		loadedGlobalState, loadGlobalStateErr = GlobalInitFromDirectory(dir)
	})
	if loadGlobalStateErr != nil {
		t.Fatal(loadGlobalStateErr)
	}
	return loadedGlobalState
}

func TestGlobalInitMatchesCppDump(t *testing.T) {
	st := testGlobalState(t)
	hash := sha256.New()
	if err := writeGlobalStateDump(hash, st); err != nil {
		t.Fatal(err)
	}
	got := hex.EncodeToString(hash.Sum(nil))
	if got != cppReferenceDumpSha256 {
		if path := os.Getenv("OPENBW_DUMP_OUT"); path != "" {
			writeDumpFile(t, path, st)
		}
		t.Fatalf("dump sha256 = %s, want %s (set OPENBW_DUMP_OUT to write the Go dump and diff it)", got, cppReferenceDumpSha256)
	}
}

func writeDumpFile(t *testing.T, path string, st *GlobalState) {
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err := writeGlobalStateDump(f, st); err != nil {
		t.Fatal(err)
	}
}

func TestIscriptLoaded(t *testing.T) {
	st := testGlobalState(t)
	if got := len(st.Iscript.ProgramData); got != 22363 {
		t.Errorf("program data size = %d, want 22363", got)
	}
	if got := len(st.Iscript.Scripts); got != 398 {
		t.Errorf("script count = %d, want 398", got)
	}
	script := st.Iscript.Scripts[0]
	want := []uint{5458, 5497, 0, 5506, 0, 0, 5506, 0, 0, 5467, 0, 5520, 5467, 0}
	if fmt.Sprint(script.AnimationPc) != fmt.Sprint(want) {
		t.Errorf("script 0 animation pcs = %v, want %v", script.AnimationPc, want)
	}
	if st.Iscript.ProgramData[0] != 0 {
		t.Errorf("program data starts with %d, want 0", st.Iscript.ProgramData[0])
	}
}

func TestImagesLoaded(t *testing.T) {
	st := testGlobalState(t)
	if got := len(st.Grps); got != 834 {
		t.Errorf("grp count = %d, want 834", got)
	}
	if got := len(st.LoOffsets); got != 97 {
		t.Errorf("lo offsets count = %d, want 97", got)
	}
	scourge := st.ImageGrp[ImageTypesIMAGEIDScourge]
	if scourge != &st.Grps[1] || scourge.Width != 48 || scourge.Height != 48 || len(scourge.Frames) != 85 {
		t.Errorf("scourge grp = %dx%d with %d frames, want grps[1] 48x48 with 85 frames", scourge.Width, scourge.Height, len(scourge.Frames))
	}
	if st.ImageGrp[939] != &st.Grps[789] {
		t.Errorf("image 939 does not point to grps[789]")
	}
	if got := len(*st.ImageLoOffsets[ImageTypesIMAGEIDScourge][5]); got != 17 {
		t.Errorf("scourge has %d shield lo frames, want 17", got)
	}
}

func TestTilesetsLoaded(t *testing.T) {
	st := testGlobalState(t)
	wantVf4 := [8]int{197504, 182752, 45792, 133152, 225216, 292160, 229568, 273664}
	wantCv5 := [8]int{102908, 106392, 65780, 73736, 106392, 106392, 106028, 106444}
	for i := range wantVf4 {
		if len(st.TilesetVf4[i]) != wantVf4[i] || len(st.TilesetCv5[i]) != wantCv5[i] {
			t.Errorf("tileset %d: vf4 %d cv5 %d, want %d %d", i, len(st.TilesetVf4[i]), len(st.TilesetCv5[i]), wantVf4[i], wantCv5[i])
		}
	}
	if len(st.MeleeTrg) != 4800 {
		t.Errorf("melee.trg size = %d, want 4800", len(st.MeleeTrg))
	}
}

func TestGlobalInitPointerFixups(t *testing.T) {
	st := testGlobalState(t)
	for _, v := range st.FlingyTypes.Vec {
		if v.Sprite.Ptr == nil || v.Sprite.Ptr.ID != v.Sprite.ID {
			t.Fatalf("flingy %d sprite pointer not fixed up", v.ID)
		}
	}
	for _, v := range st.SpriteTypes.Vec {
		if v.Image.Ptr == nil || v.Image.Ptr.ID != v.Image.ID {
			t.Fatalf("sprite %d image pointer not fixed up", v.ID)
		}
	}
}

func TestGlobalInitReportsMissingFile(t *testing.T) {
	_, err := GlobalInit(func(path string) ([]byte, error) {
		return nil, os.ErrNotExist
	})
	if err == nil || !strings.Contains(err.Error(), "arr/units.dat") {
		t.Fatalf("err = %v, want an error that names arr/units.dat", err)
	}
}

func fnv64(data []byte) uint64 {
	h := uint64(14695981039346656037)
	for _, b := range data {
		h ^= uint64(b)
		h *= 1099511628211
	}
	return h
}

func lineOffsetsHash(offsets []uint) uint64 {
	h := uint64(14695981039346656037)
	for _, o := range offsets {
		var b [4]byte
		binary.LittleEndian.PutUint32(b[:], uint32(o))
		h ^= fnv64(b[:])
		h *= 1099511628211
	}
	return h
}

func b2i(b bool) int {
	if b {
		return 1
	}
	return 0
}

func ptrID[T any, I ~int](t TypeID[T, I], id func(*T) I) int {
	if t.Ptr == nil {
		return -1
	}
	return int(id(t.Ptr))
}

type dumpWriter struct {
	w   io.Writer
	err error
}

func (d *dumpWriter) p(format string, args ...any) {
	if d.err != nil {
		return
	}
	_, d.err = fmt.Fprintf(d.w, format, args...)
}

func writeGlobalStateDump(w io.Writer, st *GlobalState) error {
	d := &dumpWriter{w: w}
	dumpGlobalTypes(d, st)
	if err := dumpGameTypes(d, st); err != nil {
		return err
	}
	dumpIscript(d, st)
	dumpGrps(d, st)
	dumpLoOffsets(d, st)
	dumpRawFiles(d, st)
	return d.err
}

func dumpGlobalTypes(d *dumpWriter, st *GlobalState) {
	for _, v := range st.FlingyTypes.Vec {
		d.p("flingy %d %d %d %d %d %d %d %d\n", v.ID, v.Sprite.ID, v.TopSpeed.RawValue, v.Acceleration.RawValue, v.HaltDistance.RawValue, v.TurnRate.RawValue, v.Unused, v.MovementType)
	}
	for _, v := range st.FlingyTypes.Vec {
		d.p("flingy_sprite_ptr %d %d\n", v.ID, ptrID(v.Sprite, func(s *SpriteTypeT) SpriteTypes { return s.ID }))
	}
	for _, v := range st.SpriteTypes.Vec {
		d.p("sprite %d %d %d %d %d %d %d %d\n", v.ID, v.Image.ID, v.HealthBarSize, v.Unk0, b2i(v.Visible), v.SelectionCircle, v.SelectionCircleVpos, ptrID(v.Image, func(i *ImageTypeT) ImageTypes { return i.ID }))
	}
	for _, v := range st.ImageTypes.Vec {
		d.p("image %d %d %d %d %d %d %d %d %d %d %d %d %d %d %d\n", v.ID, v.GrpFilenameIndex, b2i(v.HasDirectionalFrames), b2i(v.IsClickable), b2i(v.HasIscriptAnimations), b2i(v.AlwaysVisible), v.Modifier, v.ColorShift, v.IscriptID, v.ShieldFilenameIndex, v.AttackFilenameIndex, v.DamageFilenameIndex, v.SpecialFilenameIndex, v.LandingDustFilenameIndex, v.LiftOffFilenameIndex)
	}
	for _, v := range st.OrderTypes.Vec {
		d.p("order %d %d %d %d %d %d %d %d %d %d %d %d %d %d %d %d %d %d %d %d\n", v.ID, v.Label, b2i(v.TargetsEnemies), v.Background, v.Unused3, b2i(v.ValidForTurret), v.Unused5, b2i(v.CanBeInterrupted), b2i(v.Unk7), b2i(v.CanBeQueued), v.Unk9, b2i(v.CanBeObstructed), v.Unk11, v.Unused12, v.Weapon, v.TechType, v.Animation, v.Highlight, v.DepIndex, v.TargetOrder)
	}
}

func dumpGameTypes(d *dumpWriter, st *GlobalState) error {
	units, err := LoadUnitsDat(st.UnitsDat)
	if err != nil {
		return err
	}
	for _, v := range units.Vec {
		d.p("unit %d %d %d %d %d %d %d %d %d %d %d %d %d", v.ID, v.Flingy.ID, v.TurretUnitType.ID, v.Subunit2.ID, v.InfestationUnit.ID, v.ConstructionAnimation.ID, v.UnitDirection, b2i(v.HasShield), v.ShieldPoints, v.Hitpoints.RawValue, v.ElevationLevel, v.Unknown1, v.Sublabel)
		d.p(" %d %d %d %d %d %d %d %d %d %d %d", v.ComputerAiIdle.ID, v.HumanAiIdle.ID, v.ReturnToIdle.ID, v.AttackUnit.ID, v.AttackMove.ID, v.GroundWeapon.ID, v.MaxGroundHits, v.AirWeapon.ID, v.MaxAirHits, v.AiInternal, v.Flags)
		d.p(" %d %d %d %d %d %d %d %d %d %d %d %d %d", v.TargetAcquisitionRange, v.SightRange, v.ArmorUpgrade.ID, v.UnitSize, v.Armor, v.RightClickAction, v.ReadySound, v.FirstWhatSound, v.LastWhatSound, v.FirstPissedSound, v.LastPissedSound, v.FirstYesSound, v.LastYesSound)
		d.p(" %d %d %d %d %d %d %d %d", v.PlacementSize.X, v.PlacementSize.Y, v.AddonPosition.X, v.AddonPosition.Y, v.Dimensions.From.X, v.Dimensions.From.Y, v.Dimensions.To.X, v.Dimensions.To.Y)
		d.p(" %d %d %d %d %d %d %d %d %d %d %d %d %d %d %d\n", v.Portrait, v.MineralCost, v.GasCost, v.BuildTime, v.Unknown2, v.GroupFlags, v.SupplyProvided.RawValue, v.SupplyRequired.RawValue, v.SpaceRequired, v.SpaceProvided, v.BuildScore, v.DestroyScore, v.UnitMapStringIndex, b2i(v.IsBroodwar), v.StareditAvailabilityFlags)
	}
	weapons, err := LoadWeaponsDat(st.WeaponsDat)
	if err != nil {
		return err
	}
	for _, v := range weapons.Vec {
		d.p("weapon %d %d %d %d %d %d %d %d %d %d %d %d %d %d %d %d %d %d %d %d %d %d %d %d %d\n", v.ID, v.Label, v.Flingy.ID, v.Unused, v.TargetFlags, v.MinRange, v.MaxRange, v.DamageUpgrade.ID, v.DamageType, v.BulletType, v.Lifetime, v.HitType, v.InnerSplashRadius, v.MediumSplashRadius, v.OuterSplashRadius, v.DamageAmount, v.DamageBonus, v.Cooldown, v.BulletCount, v.AttackAngle.RawValue, v.BulletHeadingOffset.RawValue, v.ForwardOffset, v.UpwardOffset, v.TargetErrorMessage, v.Icon)
	}
	upgrades, err := LoadUpgradesDat(st.UpgradesDat)
	if err != nil {
		return err
	}
	for _, v := range upgrades.Vec {
		d.p("upgrade %d %d %d %d %d %d %d %d %d %d %d %d %d\n", v.ID, v.MineralCostBase, v.MineralCostFactor, v.GasCostBase, v.GasCostFactor, v.TimeCostBase, v.TimeCostFactor, v.Unknown, v.Icon, v.Label, v.Race, v.MaxLevel, b2i(v.IsBroodwar))
	}
	techs, err := LoadTechdataDat(st.TechdataDat)
	if err != nil {
		return err
	}
	for _, v := range techs.Vec {
		d.p("tech %d %d %d %d %d %d %d %d %d %d\n", v.ID, v.MineralCost, v.GasCost, v.ResearchTime, v.EnergyCost, v.Unknown, v.Icon, v.Label, v.Race, v.Flags)
	}
	return nil
}

func dumpIscript(d *dumpWriter, st *GlobalState) {
	d.p("program_data %d\n", len(st.Iscript.ProgramData))
	for i, v := range st.Iscript.ProgramData {
		d.p("pd %d %d\n", i, v)
	}
	ids := make([]int, 0, len(st.Iscript.Scripts))
	for id := range st.Iscript.Scripts {
		ids = append(ids, int(id))
	}
	sort.Ints(ids)
	d.p("scripts %d\n", len(ids))
	for _, id := range ids {
		s := st.Iscript.Scripts[int32(id)]
		d.p("script %d %d", id, s.ID)
		for _, pc := range s.AnimationPc {
			d.p(" %d", pc)
		}
		d.p("\n")
	}
}

func dumpGrps(d *dumpWriter, st *GlobalState) {
	d.p("grps %d\n", len(st.Grps))
	for i, g := range st.Grps {
		d.p("grp %d %d %d %d\n", i, g.Width, g.Height, len(g.Frames))
		for f, fr := range g.Frames {
			d.p("frame %d %d %d %d %d %d %d %d %d\n", f, fr.Offset.X, fr.Offset.Y, fr.Size.X, fr.Size.Y, len(fr.LineDataOffset), len(fr.DataContainer), lineOffsetsHash(fr.LineDataOffset), fnv64(fr.DataContainer))
		}
	}
	for i, g := range st.ImageGrp {
		d.p("image_grp %d %d\n", i, grpIndex(st, g))
	}
}

func grpIndex(st *GlobalState, g *GrpT) int {
	for i := range st.Grps {
		if &st.Grps[i] == g {
			return i
		}
	}
	return -1
}

func loOffsetsIndex(st *GlobalState, offsets *[][]Xy) int {
	for i := range st.LoOffsets {
		if &st.LoOffsets[i] == offsets {
			return i
		}
	}
	return -1
}

func dumpLoOffsets(d *dumpWriter, st *GlobalState) {
	d.p("lo_offsets %d\n", len(st.LoOffsets))
	for i, lo := range st.LoOffsets {
		d.p("lo %d %d\n", i, len(lo))
		for f, frame := range lo {
			d.p("lof %d %d", i, f)
			for _, o := range frame {
				d.p(" %d,%d", o.X, o.Y)
			}
			d.p("\n")
		}
	}
	for i, offsets := range st.ImageLoOffsets {
		d.p("image_lo %d", i)
		for _, p := range offsets {
			d.p(" %d", loOffsetsIndex(st, p))
		}
		d.p("\n")
	}
}

func dumpRawFiles(d *dumpWriter, st *GlobalState) {
	d.p("dat_sizes %d %d %d %d %d\n", len(st.UnitsDat), len(st.WeaponsDat), len(st.UpgradesDat), len(st.TechdataDat), len(st.MeleeTrg))
	d.p("melee_trg %d\n", fnv64(st.MeleeTrg))
	for i := 0; i != 8; i++ {
		d.p("tileset %d %d %d %d %d\n", i, len(st.TilesetVf4[i]), len(st.TilesetCv5[i]), fnv64(st.TilesetVf4[i]), fnv64(st.TilesetCv5[i]))
	}
}
