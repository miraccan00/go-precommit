package upstreamhooks

import (
	"flag"
	"fmt"
	"os"
)

// RunCheckExecutablesHaveShebangs is the Go equivalent of check_executables_have_shebangs.py.
// Ensures that (non-binary) executables have a shebang line.
// Returns 0 on success, 1 if executables without shebangs are found.
func RunCheckExecutablesHaveShebangs(args []string) int {
	fs := flag.NewFlagSet("check-executables-have-shebangs", flag.ContinueOnError)
	if err := fs.Parse(args); err != nil {
		return 1
	}

	retv := 0
	for _, f := range fs.Args() {
		info, err := os.Stat(f)
		if err != nil {
			continue
		}
		if info.Mode()&0o111 == 0 {
			continue // not executable — skip
		}
		if !hasShebang(f) {
			fmt.Printf("%s: executable without shebang\n", f)
			retv = 1
		}
	}
	return retv
}

// hasShebang reports whether the file starts with "#!".
func hasShebang(filename string) bool {
	fh, err := os.Open(filename)
	if err != nil {
		return false
	}
	buf := make([]byte, 2)
	n, _ := fh.Read(buf)
	_ = fh.Close()
	return n == 2 && buf[0] == '#' && buf[1] == '!'
}
