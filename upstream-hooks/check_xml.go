package upstreamhooks

import (
	"encoding/xml"
	"flag"
	"fmt"
	"io"
	"os"
)

// RunCheckXML is the Go equivalent of check_xml.py.
// Checks XML files for parseable syntax.
// Returns 0 on success, 1 if any file fails to parse.
func RunCheckXML(args []string) int {
	fs := flag.NewFlagSet("check-xml", flag.ContinueOnError)
	if err := fs.Parse(args); err != nil {
		return 1
	}

	retval := 0
	for _, f := range fs.Args() {
		data, err := os.ReadFile(f)
		if err != nil {
			fmt.Printf("%s: %v\n", f, err)
			retval = 1
			continue
		}
		dec := xml.NewDecoder(bytesReader(data))
		for {
			_, err := dec.Token()
			if err == io.EOF {
				break
			}
			if err != nil {
				fmt.Printf("%s: Failed to xml parse (%v)\n", f, err)
				retval = 1
				break
			}
		}
	}
	return retval
}
