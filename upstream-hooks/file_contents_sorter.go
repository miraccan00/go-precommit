package upstreamhooks

import (
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"
)

// RunFileContentsSorter is the Go equivalent of file_contents_sorter.py.
// Sorts the lines in specified files alphabetically.
// Returns 0 if no files changed, 1 if any file was sorted.
func RunFileContentsSorter(args []string) int {
	fs := flag.NewFlagSet("file-contents-sorter", flag.ContinueOnError)
	ignoreCase := fs.Bool("ignore-case", false, "fold lower case to upper case for comparison")
	unique := fs.Bool("unique", false, "ensure each line is unique")
	if err := fs.Parse(args); err != nil {
		return 1
	}

	retv := 0
	for _, f := range fs.Args() {
		if sortFileContents(f, *ignoreCase, *unique) {
			fmt.Printf("Sorting %s\n", f)
			retv = 1
		}
	}
	return retv
}

// sortFileContents sorts lines in the file. Returns true if the file was modified.
func sortFileContents(filename string, ignoreCase, unique bool) bool {
	data, err := os.ReadFile(filename)
	if err != nil {
		return false
	}

	// Collect non-empty lines, stripping trailing CR/LF.
	var lines []string
	for _, l := range strings.Split(strings.TrimRight(string(data), "\r\n"), "\n") {
		l = strings.TrimRight(l, "\r")
		if strings.TrimSpace(l) != "" {
			lines = append(lines, l)
		}
	}

	if unique {
		seen := make(map[string]bool, len(lines))
		deduped := lines[:0]
		for _, l := range lines {
			if !seen[l] {
				seen[l] = true
				deduped = append(deduped, l)
			}
		}
		lines = deduped
	}

	sorted := make([]string, len(lines))
	copy(sorted, lines)
	sort.Slice(sorted, func(i, j int) bool {
		a, b := sorted[i], sorted[j]
		if ignoreCase {
			a, b = strings.ToLower(a), strings.ToLower(b)
		}
		return a < b
	})

	newContent := strings.Join(sorted, "\n")
	if newContent != "" {
		newContent += "\n"
	}
	if newContent == string(data) {
		return false
	}
	_ = os.WriteFile(filename, []byte(newContent), 0o644)
	return true
}
