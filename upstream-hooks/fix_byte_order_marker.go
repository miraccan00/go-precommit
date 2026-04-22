package upstreamhooks

import (
	"flag"
	"fmt"
	"os"
)

// RunFixByteOrderMarker is the Go equivalent of fix_byte_order_marker.py.
// Removes the UTF-8 byte order marker (BOM: 0xEF 0xBB 0xBF) from text files.
// Returns 0 if no files changed, 1 if any file was modified.
func RunFixByteOrderMarker(args []string) int {
	fs := flag.NewFlagSet("fix-byte-order-marker", flag.ContinueOnError)
	if err := fs.Parse(args); err != nil {
		return 1
	}

	bom := []byte{0xEF, 0xBB, 0xBF}
	retv := 0
	for _, f := range fs.Args() {
		data, err := os.ReadFile(f)
		if err != nil {
			continue
		}
		if len(data) >= 3 && data[0] == bom[0] && data[1] == bom[1] && data[2] == bom[2] {
			if err := os.WriteFile(f, data[3:], 0o644); err == nil {
				fmt.Printf("%s: removed byte-order marker\n", f)
				retv = 1
			}
		}
	}
	return retv
}
