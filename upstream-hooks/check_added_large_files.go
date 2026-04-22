package upstreamhooks

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// RunCheckAddedLargeFiles is the Go equivalent of check_added_large_files.py.
// Prevents giant files from being committed.
// Returns 0 on success, 1 if oversized files are found.
func RunCheckAddedLargeFiles(args []string) int {
	fs := flag.NewFlagSet("check-added-large-files", flag.ContinueOnError)
	maxKB := fs.Int("maxkb", 500, "maximum allowable KB for added files")
	enforceAll := fs.Bool("enforce-all", false, "enforce all files, not just staged")
	if err := fs.Parse(args); err != nil {
		return 1
	}
	filenames := fs.Args()

	candidates := make(map[string]bool, len(filenames))
	for _, f := range filenames {
		candidates[f] = true
	}

	// Filter to LFS-tracked files and remove them.
	filterLFSFiles(candidates)

	// Unless --enforce-all, intersect with the set of actually-added files.
	if !*enforceAll {
		added, err := addedFiles()
		if err == nil {
			for f := range candidates {
				if !added[f] {
					delete(candidates, f)
				}
			}
		}
	}

	maxBytes := int64(*maxKB) * 1024
	retv := 0
	for f := range candidates {
		info, err := os.Stat(f)
		if err != nil {
			continue
		}
		kb := (info.Size() + 1023) / 1024
		if info.Size() > maxBytes {
			fmt.Printf("%s (%d KB) exceeds %d KB.\n", f, kb, *maxKB)
			retv = 1
		}
	}
	return retv
}

// filterLFSFiles removes LFS-tracked files from the set in-place.
func filterLFSFiles(files map[string]bool) {
	if len(files) == 0 {
		return
	}
	names := make([]string, 0, len(files))
	for f := range files {
		names = append(names, f)
	}
	cmd := exec.Command("git", "check-attr", "filter", "-z", "--stdin")
	cmd.Stdin = strings.NewReader(strings.Join(names, "\x00"))
	out, err := cmd.Output()
	if err != nil {
		return
	}
	parts := zsplit(string(out))
	for i := 0; i+2 < len(parts); i += 3 {
		filename := parts[i]
		filterTag := parts[i+2]
		if filterTag == "lfs" {
			delete(files, filename)
		}
	}
}
