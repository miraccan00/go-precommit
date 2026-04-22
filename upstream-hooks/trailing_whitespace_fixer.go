package upstreamhooks

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

// RunTrailingWhitespaceFixer is the Go equivalent of trailing_whitespace_fixer.py.
// Trims trailing whitespace from each line.
// Returns 0 if no files changed, 1 if any file was modified.
func RunTrailingWhitespaceFixer(args []string) int {
	fs := flag.NewFlagSet("trailing-whitespace", flag.ContinueOnError)
	markdownLinebreak := fs.Bool("markdown-linebreak-ext", false,
		"preserve trailing two-space markdown line breaks (or use '*' for all files)")
	noMarkdown := fs.Bool("no-markdown-linebreak-ext", false, "deprecated no-op")
	chars := fs.String("chars", "", "set of characters to strip (default: all whitespace)")
	if err := fs.Parse(args); err != nil {
		return 1
	}
	_ = *noMarkdown

	retv := 0
	for _, f := range fs.Args() {
		isMD := *markdownLinebreak
		if fixTrailingWhitespace(f, isMD, *chars) {
			fmt.Printf("Fixing %s\n", f)
			retv = 1
		}
	}
	return retv
}

func fixTrailingWhitespace(filename string, isMarkdown bool, chars string) bool {
	data, err := os.ReadFile(filename)
	if err != nil {
		return false
	}

	lines := strings.Split(string(data), "\n")
	changed := false
	for i, line := range lines {
		var eol string
		if strings.HasSuffix(line, "\r") {
			eol = "\r"
			line = line[:len(line)-1]
		}

		var trimmed string
		if chars == "" {
			trimmed = strings.TrimRight(line, " \t")
		} else {
			trimmed = strings.TrimRight(line, chars)
		}

		// Preserve trailing two-space markdown line break.
		if isMarkdown && strings.TrimSpace(line) != "" && strings.HasSuffix(line, "  ") {
			core := strings.TrimRight(line[:len(line)-2], " \t"+chars)
			trimmed = core + "  "
		}

		lines[i] = trimmed + eol
		if lines[i] != line+eol {
			changed = true
		}
	}

	if !changed {
		return false
	}
	_ = os.WriteFile(filename, []byte(strings.Join(lines, "\n")), 0o644)
	return true
}
