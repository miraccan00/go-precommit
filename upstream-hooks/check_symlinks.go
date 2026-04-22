package upstreamhooks

import (
	"flag"
	"fmt"
	"os"
)

// RunCheckSymlinks is the Go equivalent of check_symlinks.py.
// Checks for symlinks that do not point to anything (broken symlinks).
// Returns 0 on success, 1 if broken symlinks are found.
func RunCheckSymlinks(args []string) int {
	fs := flag.NewFlagSet("check-symlinks", flag.ContinueOnError)
	if err := fs.Parse(args); err != nil {
		return 1
	}

	retv := 0
	for _, f := range fs.Args() {
		info, err := os.Lstat(f)
		if err != nil {
			continue
		}
		if info.Mode()&os.ModeSymlink != 0 {
			if _, err := os.Stat(f); err != nil {
				fmt.Printf("%s: Broken symlink\n", f)
				retv = 1
			}
		}
	}
	return retv
}
