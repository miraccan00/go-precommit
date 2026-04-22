package upstreamhooks

import (
	"flag"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// RunCheckYAML is the Go equivalent of check_yaml.py.
// Checks YAML files for parseable syntax.
// Returns 0 on success, 1 if any file fails to parse.
func RunCheckYAML(args []string) int {
	fs := flag.NewFlagSet("check-yaml", flag.ContinueOnError)
	multi := fs.Bool("allow-multiple-documents", false, "allow multiple YAML documents per file")
	_ = fs.Bool("unsafe", false, "parse-only mode (syntax check without safe-load constraints)")
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
		if err := parseYAML(data, *multi); err != nil {
			fmt.Printf("%s\n", err)
			retval = 1
		}
	}
	return retval
}

func parseYAML(data []byte, multi bool) error {
	dec := yaml.NewDecoder(bytesReader(data))
	for {
		var v interface{}
		err := dec.Decode(&v)
		if err != nil {
			if err.Error() == "EOF" {
				return nil
			}
			return err
		}
		if !multi {
			return nil
		}
	}
}
