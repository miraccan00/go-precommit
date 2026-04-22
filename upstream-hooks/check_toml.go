package upstreamhooks

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"os"
	"regexp"
	"strings"
)

// RunCheckTOML is the Go equivalent of check_toml.py.
// Checks TOML files for parseable syntax.
// Returns 0 on success, 1 if any file has a syntax error.
func RunCheckTOML(args []string) int {
	fs := flag.NewFlagSet("check-toml", flag.ContinueOnError)
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
		if err := checkTOML(f, data); err != nil {
			fmt.Printf("%s\n", err)
			retval = 1
		}
	}
	return retval
}

var (
	tomlKeyRe     = regexp.MustCompile(`^\s*[a-zA-Z0-9_.\-"']+\s*=`)
	tomlSectionRe = regexp.MustCompile(`^\s*\[`)
	tomlCommentRe = regexp.MustCompile(`^\s*#`)
)

func checkTOML(filename string, data []byte) error {
	scanner := bufio.NewScanner(bytes.NewReader(data))
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if tomlCommentRe.MatchString(line) || tomlSectionRe.MatchString(line) || tomlKeyRe.MatchString(line) {
			continue
		}
		return fmt.Errorf("%s:%d: unexpected line: %s", filename, lineNum, line)
	}
	return nil
}
