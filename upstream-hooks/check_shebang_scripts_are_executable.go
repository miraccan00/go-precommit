package upstreamhooks

import (
	"flag"
	"fmt"
	"os"
)

// RunCheckShebangScriptsAreExecutable is the Go equivalent of check_shebang_scripts_are_executable.py.
// Ensures that (non-binary) files with a shebang are executable.
// Returns 0 on success, 1 if non-executable shebang scripts are found.
func RunCheckShebangScriptsAreExecutable(args []string) int {
	fs := flag.NewFlagSet("check-shebang-scripts-are-executable", flag.ContinueOnError)
	if err := fs.Parse(args); err != nil {
		return 1
	}

	retv := 0
	for _, f := range fs.Args() {
		if !hasShebang(f) {
			continue // no shebang — not our concern
		}
		info, err := os.Stat(f)
		if err != nil {
			continue
		}
		if info.Mode()&0o111 == 0 {
			fmt.Printf("%s: has shebang but is not executable\n", f)
			retv = 1
		}
	}
	return retv
}
