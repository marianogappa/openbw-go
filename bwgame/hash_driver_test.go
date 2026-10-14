package bwgame

import (
	"encoding/binary"
	"fmt"
	"hash/fnv"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

const hashCheckExpected = "5d737723392c1182"

func TestStateHasherIsFnv1aOverLittleEndianInt32(t *testing.T) {
	values := []int32{0, 1, -1, 0x7fffffff, -0x80000000, 0x12345678}
	h := NewStateHasher()
	reference := fnv.New64a()
	for _, v := range values {
		h.AddInt32(v)
		reference.Write(binary.LittleEndian.AppendUint32(nil, uint32(v)))
	}
	if h.Value != reference.Sum64() {
		t.Fatalf("hash %016x, want %016x", h.Value, reference.Sum64())
	}
}

func TestEmptyHasherIsOffsetBasis(t *testing.T) {
	if got := NewStateHasher().Value; got != 0xcbf29ce484222325 {
		t.Fatalf("start value %016x", got)
	}
}

func hashCheckState() *State {
	st := NewState(nil, NewGameState())
	st.CurrentFrame = 1234
	st.LcgRandState = 0xdeadbeef
	for i := range int32(12) {
		st.CurrentMinerals[i] = i*100 + 7
		st.CurrentGas[i] = -i * 3
		for race := range int32(3) {
			st.SupplyUsed[i][race] = Fp1FromRaw(i*10 + race)
			st.SupplyAvailable[i][race] = Fp1FromRaw(-(i + race))
		}
	}
	marine := &UnitTypeT{ID: UnitTypesTerranMarine}
	drone := &UnitTypeT{ID: UnitTypesZergDrone}
	move := &OrderTypeT{ID: OrdersMove}
	units := []*Unit{
		hashCheckUnit(1, marine, 0, 0x1234, 0x5678, 40*256, move, 1),
		hashCheckUnit(1699, drone, 11, -5, 70000, 1, move, -0x7fffffff),
		hashCheckUnit(42, drone, 3, 256, 512, 0, nil, 0x4000),
	}
	st.VisibleUnits.PushBack(units[0])
	st.VisibleUnits.PushBack(units[1])
	st.HiddenUnits.PushBack(units[2])
	return st
}

func hashCheckUnit(index uint, unitType *UnitTypeT, owner, x, y, hp int32, order *OrderTypeT, statusFlags int32) *Unit {
	u := &Unit{}
	u.Index = index
	u.UnitType = unitType
	u.Owner = owner
	u.ExactPosition.X = Fp8FromRaw(x)
	u.ExactPosition.Y = Fp8FromRaw(y)
	u.Hp = Fp8FromRaw(hp)
	u.ShieldPoints = Fp8FromRaw(hp / 2)
	u.Energy = Fp8FromRaw(-hp)
	u.OrderType = order
	u.StatusFlags = statusFlags
	return u
}

func TestStateHashMatchesRecordedCppValue(t *testing.T) {
	if got := fmt.Sprintf("%016x", StateHash(hashCheckState())); got != hashCheckExpected {
		t.Fatalf("state hash %s, want %s", got, hashCheckExpected)
	}
}

func TestStateHashMatchesCppProgram(t *testing.T) {
	compiler, err := exec.LookPath("clang++")
	if err != nil {
		t.Skip("clang++ not found")
	}
	source := filepath.Join("..", "tools", "difftest", "hashcheck", "main.cpp")
	openbw := filepath.Join("..", "ref", "openbw")
	if _, err := os.Stat(filepath.Join(openbw, "bwgame.h")); err != nil {
		t.Skip("ref/openbw not found")
	}
	binary := filepath.Join(t.TempDir(), "hashcheck")
	build := exec.Command(compiler, "-std=c++17", "-O1", "-isystem", openbw, "-o", binary, source)
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build hashcheck: %v\n%s", err, out)
	}
	out, err := exec.Command(binary).Output()
	if err != nil {
		t.Fatalf("run hashcheck: %v", err)
	}
	want := strings.TrimSpace(string(out))
	if got := fmt.Sprintf("%016x", StateHash(hashCheckState())); got != want {
		t.Fatalf("state hash %s, C++ %s", got, want)
	}
}
