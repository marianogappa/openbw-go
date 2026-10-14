package main

import (
	"fmt"
	"io"
	"sort"
	"strings"
)

func firstDifferentFrame(a, b *runOutput) (int, bool) {
	last := max(a.lastFrame, b.lastFrame)
	for frame := 0; frame <= last; frame++ {
		hashA, okA := a.frameHashes[frame]
		hashB, okB := b.frameHashes[frame]
		if okA != okB || hashA != hashB {
			return frame, true
		}
	}
	return 0, false
}

func unitsByIndex(rows []unitRow) map[int]unitRow {
	byIndex := make(map[int]unitRow, len(rows))
	for _, row := range rows {
		byIndex[row.index] = row
	}
	return byIndex
}

func sortedUnitIndices(a, b map[int]unitRow) []int {
	var indices []int
	for index := range a {
		indices = append(indices, index)
	}
	for index := range b {
		if _, ok := a[index]; !ok {
			indices = append(indices, index)
		}
	}
	sort.Ints(indices)
	return indices
}

func formatUnitRow(columns []string, row unitRow) string {
	var parts []string
	for _, column := range columns {
		parts = append(parts, column+"="+row.fields[column])
	}
	return strings.Join(parts, " ")
}

func unitDiffLines(columns []string, nameA, nameB string, a, b *runOutput) []string {
	byIndexA := unitsByIndex(a.units)
	byIndexB := unitsByIndex(b.units)
	var lines []string
	for _, index := range sortedUnitIndices(byIndexA, byIndexB) {
		rowA, okA := byIndexA[index]
		rowB, okB := byIndexB[index]
		switch {
		case !okB:
			lines = append(lines, fmt.Sprintf("unit %d only in %s: %s", index, nameA, formatUnitRow(columns, rowA)))
		case !okA:
			lines = append(lines, fmt.Sprintf("unit %d only in %s: %s", index, nameB, formatUnitRow(columns, rowB)))
		default:
			lines = append(lines, fieldDiffLines(fmt.Sprintf("unit %d (type %s)", index, rowA.fields["type"]), columns, rowA.fields, rowB.fields, nameA, nameB)...)
		}
	}
	if len(lines) == 0 && unitOrderDiffers(a.units, b.units) {
		lines = append(lines, "same units, different list order")
	}
	return lines
}

func unitOrderDiffers(a, b []unitRow) bool {
	if len(a) != len(b) {
		return true
	}
	for i := range a {
		if a[i].index != b[i].index {
			return true
		}
	}
	return false
}

func fieldDiffLines(label string, keys []string, a, b map[string]string, nameA, nameB string) []string {
	var lines []string
	for _, key := range keys {
		if a[key] != b[key] {
			lines = append(lines, fmt.Sprintf("%s %s: %s=%s %s=%s", label, key, nameA, a[key], nameB, b[key]))
		}
	}
	return lines
}

func playerKeys(a, b map[string]string) []string {
	seen := map[string]bool{}
	var keys []string
	for _, m := range []map[string]string{a, b} {
		for key := range m {
			if !seen[key] {
				seen[key] = true
				keys = append(keys, key)
			}
		}
	}
	sort.Strings(keys)
	return keys
}

func playerDiffLines(nameA, nameB string, a, b *runOutput) []string {
	return fieldDiffLines("", playerKeys(a.players, b.players), a.players, b.players, nameA, nameB)
}

func unitColumns(a, b *runOutput) []string {
	if a.unitColumns != nil {
		return a.unitColumns
	}
	return b.unitColumns
}

func printSnapshotDiff(w io.Writer, frame int, nameA, nameB string, a, b *runOutput, maxLines int) {
	fmt.Fprintf(w, "  at frame %d:\n", frame)
	for _, out := range []struct {
		name string
		run  *runOutput
	}{{nameA, a}, {nameB, b}} {
		if failure := out.run.describeFailure(); failure != "" {
			fmt.Fprintf(w, "    %s failed: %s\n", out.name, failure)
		}
	}
	lines := append(unitDiffLines(unitColumns(a, b), nameA, nameB, a, b), playerDiffLines(nameA, nameB, a, b)...)
	if len(lines) == 0 {
		fmt.Fprintf(w, "    no unit or player differences\n")
	}
	if maxLines > 0 && len(lines) > maxLines {
		hidden := len(lines) - maxLines
		lines = append(lines[:maxLines], fmt.Sprintf("... %d more differences", hidden))
	}
	for _, line := range lines {
		fmt.Fprintf(w, "    %s\n", strings.TrimSpace(line))
	}
}
