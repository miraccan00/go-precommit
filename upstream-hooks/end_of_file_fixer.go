package upstreamhooks

import (
	"flag"
	"fmt"
	"os"
)

// RunEndOfFileFixer is the Go equivalent of end_of_file_fixer.py.
// Ensures that a file is either empty, or ends with exactly one newline.
// Returns 0 if no files changed, 1 if any file was modified.
func RunEndOfFileFixer(args []string) int {
	fs := flag.NewFlagSet("end-of-file-fixer", flag.ContinueOnError)
	if err := fs.Parse(args); err != nil {
		return 1
	}

	retv := 0
	for _, f := range fs.Args() {
		if fixEndOfFile(f) {
			fmt.Printf("Fixing %s\n", f)
			retv = 1
		}
	}
	return retv
}

// fixEndOfFile ensures the file ends with exactly one newline.
// Returns true if the file was modified.
func fixEndOfFile(filename string) bool {
	data, err := os.ReadFile(filename)
	if err != nil {
		return false
	}
	if len(data) == 0 {
		return false
	}

	// Strip all trailing CR/LF bytes, then add exactly one LF.
	stripped := data
	for len(stripped) > 0 && (stripped[len(stripped)-1] == '\n' || stripped[len(stripped)-1] == '\r') {
		stripped = stripped[:len(stripped)-1]
	}

	var fixed []byte
	if len(stripped) == 0 {
		// File was only newlines — make it empty.
		fixed = []byte{}
	} else {
		fixed = append(stripped, '\n')
	}

	if string(fixed) == string(data) {
		return false
	}
	_ = os.WriteFile(filename, fixed, 0o644)
	return true
}
