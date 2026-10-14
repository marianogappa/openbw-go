package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

type unitRow struct {
	index  int
	fields map[string]string
}

type runOutput struct {
	frameHashes map[int]string
	lastFrame   int
	endLine     string
	errorLine   string
	units       []unitRow
	unitColumns []string
	players     map[string]string
	exitErr     error
	stderrTail  string
}

func newRunOutput() *runOutput {
	return &runOutput{frameHashes: map[int]string{}, lastFrame: -1, players: map[string]string{}}
}

type engine struct {
	name string
	path string
	data string
}

func (e engine) run(replay string, args ...string) *runOutput {
	if e.data != "" {
		args = append([]string{"-data", e.data}, args...)
	}
	cmd := exec.Command(e.path, append(args, replay)...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	out := parseRunOutput(stdout.Bytes())
	out.exitErr = err
	out.stderrTail = lastLines(stderr.String(), 5)
	return out
}

func lastLines(s string, n int) string {
	lines := strings.Split(strings.TrimRight(s, "\n"), "\n")
	return strings.Join(lines[max(0, len(lines)-n):], "\n")
}

func parseRunOutput(stdout []byte) *runOutput {
	out := newRunOutput()
	scanner := bufio.NewScanner(bytes.NewReader(stdout))
	scanner.Buffer(make([]byte, 1<<20), 1<<20)
	inUnits := false
	for scanner.Scan() {
		line := scanner.Text()
		switch {
		case inUnits:
			inUnits = out.parseUnitLine(line)
		case strings.HasPrefix(line, "units_at_frame "):
			inUnits = true
		default:
			out.parseLine(line)
		}
	}
	return out
}

func (out *runOutput) parseUnitLine(line string) bool {
	if line == "end_units" {
		return false
	}
	values := strings.Split(line, "\t")
	if out.unitColumns == nil {
		out.unitColumns = values
		return true
	}
	row := unitRow{fields: map[string]string{}}
	for i, column := range out.unitColumns {
		if i < len(values) {
			row.fields[column] = values[i]
		}
	}
	row.index, _ = strconv.Atoi(row.fields["index"])
	out.units = append(out.units, row)
	return true
}

func (out *runOutput) parseLine(line string) {
	fields := strings.Fields(line)
	if len(fields) == 0 {
		return
	}
	switch fields[0] {
	case "frame":
		if len(fields) == 3 {
			frame, err := strconv.Atoi(fields[1])
			if err == nil {
				out.frameHashes[frame] = fields[2]
				out.lastFrame = max(out.lastFrame, frame)
			}
		}
	case "end":
		out.endLine = line
	case "error":
		out.errorLine = line
	case "player", "units":
		if len(fields) >= 2 {
			out.parsePlayerLine(fields)
		}
	case "action_fails":
		out.players["action_fails"] = strings.Join(fields[1:], " ")
	}
}

func (out *runOutput) parsePlayerLine(fields []string) {
	prefix := fields[0] + " " + fields[1]
	if fields[0] == "units" {
		out.players[prefix] = strings.Join(fields[2:], " ")
		return
	}
	for _, kv := range fields[2:] {
		key, value, ok := strings.Cut(kv, "=")
		if ok {
			out.players[prefix+" "+key] = value
		}
	}
}

func (out *runOutput) describeFailure() string {
	var parts []string
	if out.errorLine != "" {
		parts = append(parts, out.errorLine)
	}
	if out.exitErr != nil {
		parts = append(parts, "exit: "+out.exitErr.Error())
	}
	if out.stderrTail != "" && out.errorLine == "" {
		parts = append(parts, "stderr: "+out.stderrTail)
	}
	return strings.Join(parts, "; ")
}

func frameArgs(maxFrames int) []string {
	if maxFrames < 0 {
		return nil
	}
	return []string{"-max-frames", fmt.Sprint(maxFrames)}
}
