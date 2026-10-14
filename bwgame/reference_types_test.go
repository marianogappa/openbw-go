package bwgame

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func referenceSourceDirectory(t *testing.T) string {
	dir, err := filepath.Abs(filepath.Join("..", "ref", "openbw"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, "bwgame.h")); err != nil {
		t.Skipf("C++ reference source not found at %s", dir)
	}
	return dir
}

func runReferenceCases(t *testing.T) []string {
	refDir := referenceSourceDirectory(t)
	compiler, err := exec.LookPath("clang++")
	if err != nil {
		t.Skip("clang++ not found")
	}
	workDir := t.TempDir()
	sourcePath := filepath.Join(workDir, "cases.cpp")
	binaryPath := filepath.Join(workDir, "cases")
	if err := os.WriteFile(sourcePath, []byte(referenceCasesCpp), 0o644); err != nil {
		t.Fatal(err)
	}
	build := exec.Command(compiler, "-std=c++14", "-fwrapv", "-O1", "-w", "-I"+refDir, sourcePath, "-o", binaryPath)
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("compile C++ reference cases: %v\n%s", err, output)
	}
	output, err := exec.Command(binaryPath).Output()
	if err != nil {
		t.Fatalf("run C++ reference cases: %v", err)
	}
	var lines []string
	scanner := bufio.NewScanner(bytes.NewReader(output))
	scanner.Buffer(make([]byte, 1<<20), 1<<20)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines
}

type referenceCase struct {
	op       string
	args     []string
	expected string
}

func parseReferenceCase(t *testing.T, line string) referenceCase {
	left, right, ok := strings.Cut(line, "=")
	if !ok {
		t.Fatalf("malformed reference line %q", line)
	}
	fields := strings.Fields(left)
	return referenceCase{op: fields[0], args: fields[1:], expected: strings.Join(strings.Fields(right), " ")}
}

func (c referenceCase) i32(t *testing.T, n int) int32 {
	v, err := strconv.ParseInt(c.args[n], 10, 32)
	if err != nil {
		t.Fatal(err)
	}
	return int32(v)
}

func (c referenceCase) u32(t *testing.T, n int) uint32 {
	v, err := strconv.ParseUint(c.args[n], 10, 32)
	if err != nil {
		t.Fatal(err)
	}
	return uint32(v)
}

func (c referenceCase) u64(t *testing.T, n int) uint64 {
	v, err := strconv.ParseUint(c.args[n], 10, 64)
	if err != nil {
		t.Fatal(err)
	}
	return v
}

func (c referenceCase) uintArg(t *testing.T, n int) uint {
	return uint(c.u64(t, n))
}

func boolToInt(v bool) int {
	if v {
		return 1
	}
	return 0
}

func joinValues(values ...any) string {
	parts := make([]string, len(values))
	for i, v := range values {
		parts[i] = fmt.Sprint(v)
	}
	return strings.Join(parts, " ")
}

type referenceReplayer struct {
	t *testing.T

	lcgState StateBaseCopyable

	images        ObjectContainer[Image, *Image]
	takenImages   map[uint]*Image
	listImages    [32]Image
	list          ImageList
	staticVector  StaticVector[int32, [16]int32]
	numericCases  int
	stateful      int
	unknownOpSeen map[string]bool
}

func newReferenceReplayer(t *testing.T) *referenceReplayer {
	r := &referenceReplayer{
		t:             t,
		images:        NewObjectContainer[Image](5000, 50),
		takenImages:   map[uint]*Image{},
		unknownOpSeen: map[string]bool{},
	}
	r.lcgState.LcgRandState = 42
	for i := range r.listImages {
		r.listImages[i].Index = uint(i)
	}
	return r
}

func (r *referenceReplayer) fp8Result(c referenceCase) (string, bool) {
	t := r.t
	switch c.op {
	case "fp8_mul":
		return fmt.Sprint(Fp8FromRaw(c.i32(t, 0)).Mul(Fp8FromRaw(c.i32(t, 1))).RawValue), true
	case "fp8_div":
		return fmt.Sprint(Fp8FromRaw(c.i32(t, 0)).Div(Fp8FromRaw(c.i32(t, 1))).RawValue), true
	case "fp8_mulint":
		return fmt.Sprint(Fp8FromRaw(c.i32(t, 0)).MulInt(c.i32(t, 1)).RawValue), true
	case "fp8_divint":
		return fmt.Sprint(Fp8FromRaw(c.i32(t, 0)).DivInt(c.i32(t, 1)).RawValue), true
	case "fp8_integer":
		return fmt.Sprint(Fp8Integer(c.i32(t, 0)).RawValue), true
	case "fp8_intpart":
		return fmt.Sprint(Fp8FromRaw(c.i32(t, 0)).IntegerPart()), true
	case "fp8_fracpart":
		return fmt.Sprint(Fp8FromRaw(c.i32(t, 0)).FractionalPart()), true
	case "fp8_floor":
		return fmt.Sprint(Fp8FromRaw(c.i32(t, 0)).Floor().RawValue), true
	case "fp8_ceil":
		return fmt.Sprint(Fp8FromRaw(c.i32(t, 0)).Ceil().RawValue), true
	case "fp8_abs":
		return fmt.Sprint(Fp8FromRaw(c.i32(t, 0)).Abs().RawValue), true
	case "fp8_neg":
		return fmt.Sprint(Fp8FromRaw(c.i32(t, 0)).Neg().RawValue), true
	case "fp8_add":
		return fmt.Sprint(Fp8FromRaw(c.i32(t, 0)).Add(Fp8FromRaw(c.i32(t, 1))).RawValue), true
	case "fp8_sub":
		return fmt.Sprint(Fp8FromRaw(c.i32(t, 0)).Sub(Fp8FromRaw(c.i32(t, 1))).RawValue), true
	case "fp8_muldiv":
		return fmt.Sprint(Fp8MultiplyDivide(Fp8FromRaw(c.i32(t, 0)), Fp8FromRaw(c.i32(t, 1)), Fp8FromRaw(c.i32(t, 2))).RawValue), true
	case "fp8_divmul":
		return fmt.Sprint(Fp8DivideMultiply(Fp8FromRaw(c.i32(t, 0)), Fp8FromRaw(c.i32(t, 1)), Fp8FromRaw(c.i32(t, 2))).RawValue), true
	case "fp8_less":
		return fmt.Sprint(boolToInt(Fp8FromRaw(c.i32(t, 0)).Less(Fp8FromRaw(c.i32(t, 1))))), true
	case "fp8_asunsigned":
		return fmt.Sprint(Fp8FromRaw(c.i32(t, 0)).AsUnsigned().RawValue), true
	case "fp8_truncate_dir":
		return fmt.Sprint(DirectionTTruncate(Fp8FromRaw(c.i32(t, 0))).RawValue), true
	case "fp8_hp_percent":
		return fmt.Sprint(hpPercentLikeInitialUnits(Fp8FromRaw(c.i32(t, 0)), c.i32(t, 1)).RawValue), true
	}
	return "", false
}

func hpPercentLikeInitialUnits(hitpoints Fp8, percent int32) Fp8 {
	tmp := Fp32x8Extend(hitpoints)
	tmp = tmp.MulInt(percent).DivInt(100)
	if tmp.Less(Fp32x8FromRaw(1)) {
		tmp = Fp32x8FromRaw(1)
	}
	return Fp8Truncate(tmp)
}

func (r *referenceReplayer) ufp8Result(c referenceCase) (string, bool) {
	t := r.t
	switch c.op {
	case "ufp8_mul":
		return fmt.Sprint(Ufp8FromRaw(c.u32(t, 0)).Mul(Ufp8FromRaw(c.u32(t, 1))).RawValue), true
	case "ufp8_div":
		return fmt.Sprint(Ufp8FromRaw(c.u32(t, 0)).Div(Ufp8FromRaw(c.u32(t, 1))).RawValue), true
	case "ufp8_divint":
		return fmt.Sprint(Ufp8FromRaw(c.u32(t, 0)).DivInt(c.u32(t, 1)).RawValue), true
	case "ufp8_mulint":
		return fmt.Sprint(Ufp8FromRaw(c.u32(t, 0)).MulInt(c.u32(t, 1)).RawValue), true
	case "ufp8_muldiv":
		return fmt.Sprint(Ufp8MultiplyDivide(Ufp8FromRaw(c.u32(t, 0)), Ufp8FromRaw(c.u32(t, 1)), Ufp8FromRaw(c.u32(t, 2))).RawValue), true
	case "ufp8_divmul":
		return fmt.Sprint(Ufp8DivideMultiply(Ufp8FromRaw(c.u32(t, 0)), Ufp8FromRaw(c.u32(t, 1)), Ufp8FromRaw(c.u32(t, 2))).RawValue), true
	case "ufp8_integer":
		return fmt.Sprint(Ufp8Integer(c.u32(t, 0)).RawValue), true
	case "ufp8_intpart":
		return fmt.Sprint(Ufp8FromRaw(c.u32(t, 0)).IntegerPart()), true
	case "ufp8_ceil":
		return fmt.Sprint(Ufp8FromRaw(c.u32(t, 0)).Ceil().RawValue), true
	case "ufp8_assigned":
		return fmt.Sprint(Ufp8FromRaw(c.u32(t, 0)).AsSigned().RawValue), true
	case "ufp8_third":
		return fmt.Sprint(Ufp8Integer(6).Add(Ufp8Integer(1).Sub(Ufp8Integer(1).DivInt(3))).RawValue), true
	}
	return "", false
}

func (r *referenceReplayer) fp1Fp16Result(c referenceCase) (string, bool) {
	t := r.t
	switch c.op {
	case "fp1_mul":
		return fmt.Sprint(Fp1FromRaw(c.i32(t, 0)).Mul(Fp1FromRaw(c.i32(t, 1))).RawValue), true
	case "fp1_div":
		return fmt.Sprint(Fp1FromRaw(c.i32(t, 0)).Div(Fp1FromRaw(c.i32(t, 1))).RawValue), true
	case "fp1_integer":
		return fmt.Sprint(Fp1Integer(c.i32(t, 0)).RawValue), true
	case "fp1_intpart":
		return fmt.Sprint(Fp1FromRaw(c.i32(t, 0)).IntegerPart()), true
	case "fp1_ceil":
		return fmt.Sprint(Fp1FromRaw(c.i32(t, 0)).Ceil().RawValue), true
	case "fp16_mul":
		return fmt.Sprint(Fp16FromRaw(c.i32(t, 0)).Mul(Fp16FromRaw(c.i32(t, 1))).RawValue), true
	case "fp16_div":
		return fmt.Sprint(Fp16FromRaw(c.i32(t, 0)).Div(Fp16FromRaw(c.i32(t, 1))).RawValue), true
	case "fp16_integer":
		return fmt.Sprint(Fp16Integer(c.i32(t, 0)).RawValue), true
	case "fp16_intpart":
		return fmt.Sprint(Fp16FromRaw(c.i32(t, 0)).IntegerPart()), true
	case "fp16_fracpart":
		return fmt.Sprint(Fp16FromRaw(c.i32(t, 0)).FractionalPart()), true
	case "fp16_ceil":
		return fmt.Sprint(Fp16FromRaw(c.i32(t, 0)).Ceil().RawValue), true
	case "ufp16_mul":
		return fmt.Sprint(Ufp16FromRaw(c.u32(t, 0)).Mul(Ufp16FromRaw(c.u32(t, 1))).RawValue), true
	case "ufp16_div":
		return fmt.Sprint(Ufp16FromRaw(c.u32(t, 0)).Div(Ufp16FromRaw(c.u32(t, 1))).RawValue), true
	}
	return "", false
}

func (r *referenceReplayer) directionResult(c referenceCase) (string, bool) {
	t := r.t
	dir := func(n int) DirectionT { return DirectionTFromRaw(int8(c.i32(t, n))) }
	switch c.op {
	case "dir_add", "dir_addassign":
		return fmt.Sprint(dir(0).Add(dir(1)).RawValue), true
	case "dir_sub":
		return fmt.Sprint(dir(0).Sub(dir(1)).RawValue), true
	case "dir_neg":
		return fmt.Sprint(dir(0).Neg().RawValue), true
	case "dir_mulint":
		return fmt.Sprint(dir(0).MulInt(c.i32(t, 1)).RawValue), true
	case "dir_divint":
		return fmt.Sprint(dir(0).DivInt(c.i32(t, 1)).RawValue), true
	case "dir_intpart":
		return fmt.Sprint(dir(0).IntegerPart()), true
	case "dir_fracpart":
		return fmt.Sprint(dir(0).FractionalPart()), true
	case "dir_abs":
		return fmt.Sprint(dir(0).Abs().RawValue), true
	case "dir_extend":
		return fmt.Sprint(Fp8Extend(dir(0)).RawValue), true
	case "dir_extend_sub_abs":
		return fmt.Sprint(Fp8Extend(dir(0).Sub(dir(1))).Abs().RawValue), true
	case "dir_less":
		return fmt.Sprint(boolToInt(dir(0).Less(dir(1)))), true
	case "dir_literal_neg128":
		return fmt.Sprint(DirectionTFromRaw(-128).Neg().RawValue), true
	}
	return "", false
}

func (r *referenceReplayer) integerResult(c referenceCase) (string, bool) {
	t := r.t
	switch c.op {
	case "isqrt32":
		return fmt.Sprint(isqrt(c.u32(t, 0))), true
	case "isqrt64":
		return fmt.Sprint(isqrt(c.u64(t, 0))), true
	case "unit_id":
		id := NewUnitID(c.uintArg(t, 0), c.u32(t, 1))
		return joinValues(id.RawValue, id.Index(), id.Generation()), true
	case "unit_id_32":
		id := NewUnitID32(c.uintArg(t, 0), c.u32(t, 1))
		return joinValues(id.RawValue, id.Index(), id.Generation()), true
	case "tile_id":
		id := TileID{RawValue: uint16(c.u32(t, 0))}
		return joinValues(boolToInt(id.HasCreep()), id.GroupIndex(), id.SubtileIndex()), true
	}
	return "", false
}

func (r *referenceReplayer) lcgResult(c referenceCase) (string, bool) {
	t := r.t
	switch c.op {
	case "lcg":
		v := r.lcgState.lcgRand(c.i32(t, 0))
		return joinValues(v, r.lcgState.LcgRandState), true
	case "lcg_range":
		v := r.lcgState.lcgRandRange(c.i32(t, 0), c.i32(t, 1), c.i32(t, 2))
		return joinValues(v, r.lcgState.LcgRandState), true
	case "lcg_counts":
		var values []any
		for _, count := range r.lcgState.RandomCounts[:64] {
			values = append(values, count)
		}
		values = append(values, r.lcgState.TotalRandomCounts)
		return joinValues(values...), true
	}
	return "", false
}

func (r *referenceReplayer) objectContainerResult(c referenceCase) (string, bool) {
	switch c.op {
	case "oc":
		if c.args[0] == "top" {
			image := r.images.Top()
			if image == nil {
				return "-1", true
			}
			r.images.Pop()
			r.takenImages[image.Index] = image
			return fmt.Sprint(image.Index), true
		}
		index := c.uintArg(r.t, 1)
		r.images.Push(r.takenImages[index])
		delete(r.takenImages, index)
		return "", true
	case "oc_free":
		var values []any
		for image := range r.images.FreeList.All() {
			if len(values) == 200 {
				break
			}
			values = append(values, image.Index)
		}
		return joinValues(values...), true
	case "oc_size":
		return fmt.Sprint(r.images.Size), true
	}
	return "", false
}

func (r *referenceReplayer) intrusiveListResult(c referenceCase) (string, bool) {
	if c.op != "il" {
		return "", false
	}
	image := &r.listImages[c.uintArg(r.t, 1)]
	switch c.args[0] {
	case "remove":
		r.list.Remove(image)
	case "push_front":
		r.list.PushFront(image)
	case "push_back":
		r.list.PushBack(image)
	case "bw_insert":
		bwInsertList(&r.list, image)
	case "pop_front":
		r.list.PopFront()
	case "pop_back":
		r.list.PopBack()
	case "insert":
		pos := r.list.Begin()
		steps := c.uintArg(r.t, 2)
		for k := uint(0); k != steps && pos != r.list.End(); k++ {
			pos = r.list.Next(pos)
		}
		r.list.Insert(pos, image)
	default:
		r.t.Fatalf("unknown intrusive list op %q", c.args[0])
	}
	var values []any
	for v := range r.list.All() {
		values = append(values, v.Index)
	}
	values = append(values, "|")
	for v := range r.list.Backward() {
		values = append(values, v.Index)
	}
	return joinValues(values...), true
}

func (r *referenceReplayer) staticVectorResult(c referenceCase) (string, bool) {
	if c.op != "sv" {
		return "", false
	}
	argument := c.uintArg(r.t, 1)
	switch c.args[0] {
	case "push":
		r.staticVector.PushBack(int32(argument))
	case "erase":
		r.staticVector.Erase(int(argument))
	case "pop":
		r.staticVector.PopBack()
	case "resize":
		r.staticVector.Resize(int(argument))
	default:
		r.t.Fatalf("unknown static vector op %q", c.args[0])
	}
	var values []any
	for _, v := range r.staticVector.All() {
		values = append(values, v)
	}
	return joinValues(values...), true
}

func (r *referenceReplayer) tableResult(c referenceCase) (string, bool) {
	t := r.t
	if !strings.HasPrefix(c.op, "table_") {
		return "", false
	}
	i := c.uintArg(t, 0)
	switch c.op {
	case "table_tan":
		return fmt.Sprint(tanTable[i]), true
	case "table_repulse":
		return fmt.Sprint(repulseAdjustTable[i]), true
	case "table_cardinal":
		return joinValues(cardinalDirectionXy[i].X, cardinalDirectionXy[i].Y), true
	case "table_direction":
		return joinValues(directionTable[i].X.RawValue, directionTable[i].Y.RawValue), true
	case "table_hit_near":
		return joinValues(hitNearTargetPositions[i].X, hitNearTargetPositions[i].Y), true
	case "table_psi":
		return fmt.Sprint(boolToInt(psiFieldMask[i][c.uintArg(t, 1)])), true
	case "table_player_slot":
		return fmt.Sprint(allPlayerSlots[i]), true
	}
	return "", false
}

func (r *referenceReplayer) result(c referenceCase) string {
	producers := []func(referenceCase) (string, bool){
		r.fp8Result,
		r.ufp8Result,
		r.fp1Fp16Result,
		r.directionResult,
		r.integerResult,
		r.lcgResult,
		r.objectContainerResult,
		r.intrusiveListResult,
		r.tableResult,
		r.staticVectorResult,
	}
	for _, produce := range producers {
		if result, ok := produce(c); ok {
			return result
		}
	}
	r.t.Fatalf("unknown reference op %q", c.op)
	return ""
}

func TestTypesMatchCppReference(t *testing.T) {
	lines := runReferenceCases(t)
	if len(lines) == 0 {
		t.Fatal("C++ reference cases printed nothing")
	}
	r := newReferenceReplayer(t)
	failures := 0
	for lineNumber, line := range lines {
		c := parseReferenceCase(t, line)
		got := r.result(c)
		if got != c.expected {
			failures++
			t.Errorf("line %d: %s: got %q, want %q", lineNumber+1, line, got, c.expected)
			if failures == 20 {
				t.Fatal("too many failures")
			}
		}
	}
}
