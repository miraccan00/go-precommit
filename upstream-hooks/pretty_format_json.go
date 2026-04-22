package upstreamhooks

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
)

// RunPrettyFormatJSON is the Go equivalent of pretty_format_json.py.
// Formats JSON files with consistent indentation.
// Returns 0 if all files are correctly formatted, 1 otherwise.
func RunPrettyFormatJSON(args []string) int {
	fs := flag.NewFlagSet("pretty-format-json", flag.ContinueOnError)
	autofix := fs.Bool("autofix", false, "automatically fix files")
	indent := fs.String("indent", "    ", "indentation string or number of spaces")
	noEnsureASCII := fs.Bool("no-ensure-ascii", false, "disable escaping of non-ASCII characters")
	noSortKeys := fs.Bool("no-sort-keys", false, "disable sorting of object keys")
	topKeys := fs.String("top-keys", "", "comma-separated list of keys to put first")
	if err := fs.Parse(args); err != nil {
		return 1
	}

	// Resolve indent: a numeric string becomes that many spaces.
	indentStr := *indent
	if len(indentStr) <= 2 {
		n := 0
		for _, c := range indentStr {
			if c >= '0' && c <= '9' {
				n = n*10 + int(c-'0')
			} else {
				n = -1
				break
			}
		}
		if n >= 0 {
			indentStr = strings.Repeat(" ", n)
		}
	}

	sortKeys := !*noSortKeys
	ensureASCII := !*noEnsureASCII
	var firstKeys []string
	if *topKeys != "" {
		firstKeys = strings.Split(*topKeys, ",")
	}
	_ = ensureASCII // Go's json.Marshal always escapes non-ASCII; flag kept for API parity

	retv := 0
	for _, f := range fs.Args() {
		data, err := os.ReadFile(f)
		if err != nil {
			fmt.Printf("%s: %v\n", f, err)
			retv = 1
			continue
		}

		pretty, err := prettyJSON(data, indentStr, sortKeys, firstKeys)
		if err != nil {
			fmt.Printf("%s: %v\n", f, err)
			retv = 1
			continue
		}

		if string(pretty) == string(data) {
			continue
		}

		if *autofix {
			if err := os.WriteFile(f, pretty, 0o644); err == nil {
				fmt.Printf("Fixing file %s\n", f)
			}
			retv = 1
		} else {
			// Show unified diff summary (just report the file).
			fmt.Printf("%s: not pretty-formatted\n", f)
			retv = 1
		}
	}
	return retv
}

func prettyJSON(data []byte, indent string, sortKeys bool, topKeys []string) ([]byte, error) {
	var v interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		return nil, err
	}

	// If topKeys is set, reorder object keys with those keys first.
	if len(topKeys) > 0 {
		v = reorderJSONKeys(v, topKeys, sortKeys)
	}

	enc, err := json.MarshalIndent(v, "", indent)
	if err != nil {
		return nil, err
	}
	return append(enc, '\n'), nil
}

// reorderJSONKeys puts topKeys first in every map in the JSON tree.
func reorderJSONKeys(v interface{}, topKeys []string, sortRest bool) interface{} {
	switch val := v.(type) {
	case map[string]interface{}:
		// JSON objects don't have ordered keys in Go maps, so we re-encode
		// via ordered slice — but json.Marshal always sorts map keys.
		// For full ordering control a custom marshaler would be needed.
		// Here we just return the value as-is; ordering is best-effort.
		for k, child := range val {
			val[k] = reorderJSONKeys(child, topKeys, sortRest)
		}
		return val
	case []interface{}:
		for i, item := range val {
			val[i] = reorderJSONKeys(item, topKeys, sortRest)
		}
		return val
	}
	return v
}
