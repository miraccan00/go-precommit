package upstreamhooks

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"os"
	"regexp"
)

// RunCheckVCSPermalinks is the Go equivalent of check_vcs_permalinks.py.
// Ensures that links to GitHub are permalinks (commit hash, not branch name).
// Returns 0 on success, 1 if non-permanent links are found.
func RunCheckVCSPermalinks(args []string) int {
	fs := flag.NewFlagSet("check-vcs-permalinks", flag.ContinueOnError)
	var extraDomains domainList
	fs.Var(&extraDomains, "additional-github-domain", "additional GitHub domain to check (may be repeated)")
	if err := fs.Parse(args); err != nil {
		return 1
	}

	domains := append([]string{"github.com"}, extraDomains...)
	// Capture the ref segment (branch or commit hash) after /blob/.
	// Go's regexp does not support negative lookaheads, so we capture and
	// check in code whether the ref is a commit hash.
	patterns := make([]*regexp.Regexp, 0, len(domains))
	for _, d := range domains {
		re := regexp.MustCompile(
			`https://` + regexp.QuoteMeta(d) + `/[^/ ]+/[^/ ]+/blob/([^/. ]+)/[^# ]+#L\d+`,
		)
		patterns = append(patterns, re)
	}

	retv := 0
	for _, f := range fs.Args() {
		data, err := os.ReadFile(f)
		if err != nil {
			continue
		}
		scanner := bufio.NewScanner(bytes.NewReader(data))
		lineNum := 0
		for scanner.Scan() {
			lineNum++
			line := scanner.Bytes()
			for _, pat := range patterns {
				m := pat.FindSubmatch(line)
				if m == nil {
					continue
				}
				ref := m[1]
				if isCommitHash(ref) {
					continue // permalink — allow
				}
				fmt.Printf("%s:%d:", f, lineNum)
				fmt.Printf("%s\n", line)
				retv = 1
			}
		}
	}
	if retv != 0 {
		fmt.Println()
		fmt.Println("Non-permanent github link detected.")
		fmt.Println("On any page on github press [y] to load a permalink.")
	}
	return retv
}

// isCommitHash reports whether ref looks like a git commit hash (4–64 hex chars).
func isCommitHash(ref []byte) bool {
	if len(ref) < 4 || len(ref) > 64 {
		return false
	}
	for _, c := range ref {
		if (c < 'a' || c > 'f') && (c < 'A' || c > 'F') && (c < '0' || c > '9') {
			return false
		}
	}
	return true
}

// domainList implements flag.Value for repeated --additional-github-domain flags.
type domainList []string

func (d *domainList) String() string  { return "" }
func (d *domainList) Set(v string) error {
	*d = append(*d, v)
	return nil
}
