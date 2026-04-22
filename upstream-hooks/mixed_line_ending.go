package upstreamhooks

import (
	"bytes"
	"flag"
	"fmt"
	"os"
)

// RunMixedLineEnding is the Go equivalent of mixed_line_ending.py.
// Replaces or reports mixed line endings in files.
// Returns 0 on success, 1 if any file has (or had) mixed endings.
func RunMixedLineEnding(args []string) int {
	fs := flag.NewFlagSet("mixed-line-ending", flag.ContinueOnError)
	fix := fs.String("fix", "auto", "replace line ending with: auto | no | cr | crlf | lf")
	if err := fs.Parse(args); err != nil {
		return 1
	}

	retv := 0
	for _, f := range fs.Args() {
		changed, wasMixed := fixMixedLineEnding(f, *fix)
		if changed || wasMixed {
			if *fix == "no" {
				fmt.Printf("%s: mixed line endings\n", f)
			} else if changed {
				fmt.Printf("%s: fixed mixed line endings\n", f)
			}
			retv = 1
		}
	}
	return retv
}

var (
	crlf = []byte("\r\n")
	lf   = []byte("\n")
	cr   = []byte("\r")
)

// fixMixedLineEnding fixes or reports mixed line endings.
// Returns (changed, wasMixed).
func fixMixedLineEnding(filename, fix string) (bool, bool) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return false, false
	}

	// Count each ending type.
	crlfCount := bytes.Count(data, crlf)
	// For bare LF: remove CRLF occurrences first to avoid double-counting.
	stripped := bytes.ReplaceAll(data, crlf, nil)
	lfCount := bytes.Count(stripped, lf)
	crCount := bytes.Count(bytes.ReplaceAll(stripped, lf, nil), cr)

	mixed := (crlfCount > 0 && lfCount > 0) ||
		(crlfCount > 0 && crCount > 0) ||
		(lfCount > 0 && crCount > 0)

	if fix == "no" {
		return false, mixed
	}

	if fix == "auto" && !mixed {
		return false, false
	}

	var target []byte
	switch fix {
	case "cr":
		target = cr
	case "crlf":
		target = crlf
	case "lf":
		target = lf
	default: // "auto" with mixed endings — choose the most common
		target = lf
		max := lfCount
		if crlfCount > max {
			max = crlfCount
			target = crlf
		}
		if crCount >= max {
			target = cr
		}
	}

	// Normalise: split on any line ending, rejoin with target.
	lines := splitOnAnyEnding(data)
	newData := bytes.Join(lines, target)
	if bytes.Equal(newData, data) {
		return false, mixed
	}
	_ = os.WriteFile(filename, newData, 0o644)
	return true, mixed
}

// splitOnAnyEnding splits data on \r\n, \n, or \r, returning the bare lines
// (without line endings).
func splitOnAnyEnding(data []byte) [][]byte {
	var lines [][]byte
	start := 0
	for i := 0; i < len(data); i++ {
		switch data[i] {
		case '\r':
			lines = append(lines, data[start:i])
			if i+1 < len(data) && data[i+1] == '\n' {
				i++
			}
			start = i + 1
		case '\n':
			lines = append(lines, data[start:i])
			start = i + 1
		}
	}
	lines = append(lines, data[start:])
	return lines
}
