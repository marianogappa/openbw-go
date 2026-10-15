package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

const (
	bwenumsPath   = "../../ref/openbw/bwenums.h"
	generatedPath = "../../bwgame/enums.go"
	namesPath     = "../../names.go"
)

func readBwenums(t *testing.T) []byte {
	t.Helper()
	source, err := os.ReadFile(bwenumsPath)
	if err != nil {
		t.Skipf("C++ reference not found: %v", err)
	}
	return source
}

func TestGeneratedFileIsUpToDate(t *testing.T) {
	blocks, err := parseBlocks(readBwenums(t))
	if err != nil {
		t.Fatal(err)
	}
	want, err := render(blocks)
	if err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(generatedPath)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want) {
		t.Fatal("bwgame/enums.go is stale: run go run ./tools/genenums")
	}
	wantNames, err := renderNames(blocks)
	if err != nil {
		t.Fatal(err)
	}
	gotNames, err := os.ReadFile(namesPath)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(gotNames, wantNames) {
		t.Fatal("names.go is stale: run go run ./tools/genenums")
	}
}

func cppQualifiedName(block enumBlock, e enumerator) string {
	return "bwgame::" + block.cppName + "::" + e.name
}

func cppEnumPrinter(blocks []enumBlock) string {
	var b strings.Builder
	b.WriteString("#include \"bwenums.h\"\n#include <cstdio>\nint main() {\n")
	for _, block := range blocks {
		for _, e := range block.enumerators {
			fmt.Fprintf(&b, "\tprintf(\"%%d\\n\", (int)%s);\n", cppQualifiedName(block, e))
		}
	}
	b.WriteString("\treturn 0;\n}\n")
	return b.String()
}

func runCppEnumPrinter(t *testing.T, source string) []string {
	t.Helper()
	compiler, err := exec.LookPath("clang++")
	if err != nil {
		t.Skip("clang++ not found")
	}
	refDir, err := filepath.Abs(filepath.Dir(bwenumsPath))
	if err != nil {
		t.Fatal(err)
	}
	workDir := t.TempDir()
	sourcePath := filepath.Join(workDir, "enums.cpp")
	binaryPath := filepath.Join(workDir, "enums")
	if err := os.WriteFile(sourcePath, []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}
	if output, err := exec.Command(compiler, "-std=c++14", "-w", "-I"+refDir, sourcePath, "-o", binaryPath).CombinedOutput(); err != nil {
		t.Fatalf("compile: %v\n%s", err, output)
	}
	output, err := exec.Command(binaryPath).Output()
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	return strings.Fields(string(output))
}

func TestParsedValuesMatchCpp(t *testing.T) {
	blocks, err := parseBlocks(readBwenums(t))
	if err != nil {
		t.Fatal(err)
	}
	values := runCppEnumPrinter(t, cppEnumPrinter(blocks))
	i := 0
	for _, block := range blocks {
		for _, e := range block.enumerators {
			if i >= len(values) {
				t.Fatalf("C++ printed %d values, want more", len(values))
			}
			if want := values[i]; fmt.Sprint(e.value) != want {
				t.Errorf("%s = %d, C++ says %s", cppQualifiedName(block, e), e.value, want)
			}
			i++
		}
	}
	if i != len(values) || i < 2000 {
		t.Fatalf("compared %d enumerators, C++ printed %d", i, len(values))
	}
}
