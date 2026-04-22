package upstreamhooks

import (
	"flag"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
)

// RunCheckCaseConflict is the Go equivalent of check_case_conflict.py.
// Checks for files that would conflict in case-insensitive filesystems.
// Returns 0 on success, 1 if conflicts are found.
func RunCheckCaseConflict(args []string) int {
	fs := flag.NewFlagSet("check-case-conflict", flag.ContinueOnError)
	if err := fs.Parse(args); err != nil {
		return 1
	}
	filenames := fs.Args()

	// Collect all repo files plus parent directories.
	tracked, err := trackedFiles()
	if err != nil {
		return 1
	}
	repoFiles := make(map[string]bool)
	for _, f := range tracked {
		repoFiles[f] = true
		for _, d := range parentDirs(f) {
			repoFiles[d] = true
		}
	}

	added, _ := addedFiles()
	relevant := make(map[string]bool)
	for _, f := range filenames {
		relevant[f] = true
	}
	for f := range added {
		relevant[f] = true
	}
	for f := range relevant {
		for _, d := range parentDirs(f) {
			relevant[d] = true
		}
	}

	// Remove relevant from repo so we compare new files against existing ones.
	for f := range relevant {
		delete(repoFiles, f)
	}

	// Build lowercase sets for conflict detection.
	lowerRepo := lowerSet(repoFiles)
	lowerRelevant := lowerSet(relevant)

	conflicts := make(map[string]bool)

	// New file conflicts with existing file.
	for lower := range lowerRepo {
		if lowerRelevant[lower] {
			conflicts[lower] = true
		}
	}

	// New file conflicts with another new file.
	seen := make(map[string]bool, len(relevant))
	for f := range relevant {
		lower := strings.ToLower(f)
		if seen[lower] {
			conflicts[lower] = true
		} else {
			seen[lower] = true
		}
	}

	if len(conflicts) == 0 {
		return 0
	}

	// Collect and sort conflicting filenames for deterministic output.
	var conflicting []string
	for f := range repoFiles {
		if conflicts[strings.ToLower(f)] {
			conflicting = append(conflicting, f)
		}
	}
	for f := range relevant {
		if conflicts[strings.ToLower(f)] {
			conflicting = append(conflicting, f)
		}
	}
	sort.Strings(conflicting)
	for _, f := range conflicting {
		fmt.Printf("Case-insensitivity conflict found: %s\n", f)
	}
	return 1
}

func lowerSet(m map[string]bool) map[string]bool {
	out := make(map[string]bool, len(m))
	for k := range m {
		out[strings.ToLower(k)] = true
	}
	return out
}

func parentDirs(file string) []string {
	parts := strings.Split(filepath.ToSlash(file), "/")
	parts = parts[:len(parts)-1]
	var result []string
	for len(parts) > 0 {
		result = append(result, strings.Join(parts, "/"))
		parts = parts[:len(parts)-1]
	}
	return result
}
