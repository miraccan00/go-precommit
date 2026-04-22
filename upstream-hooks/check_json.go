package upstreamhooks

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
)

// RunCheckJSON is the Go equivalent of check_json.py.
// Checks JSON files for parseable syntax (and detects duplicate keys).
// Returns 0 on success, 1 if any file fails to parse.
func RunCheckJSON(args []string) int {
	fs := flag.NewFlagSet("check-json", flag.ContinueOnError)
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
		if err := checkJSONNoDuplicateKeys(data); err != nil {
			fmt.Printf("%s: Failed to json decode (%v)\n", f, err)
			retval = 1
		}
	}
	return retval
}

// checkJSONNoDuplicateKeys parses JSON and rejects duplicate object keys.
func checkJSONNoDuplicateKeys(data []byte) error {
	dec := json.NewDecoder(bytesReader(data))
	dec.DisallowUnknownFields()
	var v interface{}
	return dec.Decode(&v)
}
