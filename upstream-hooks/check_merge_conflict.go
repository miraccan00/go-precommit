package upstreamhooks

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5/plumbing"
)

// RunCheckMergeConflict is the Go equivalent of check_merge_conflict.py.
// Checks for files that contain merge conflict strings.
// Returns 0 on success, 1 if conflict markers are found.
func RunCheckMergeConflict(args []string) int {
	fs := flag.NewFlagSet("check-merge-conflict", flag.ContinueOnError)
	assumeInMerge := fs.Bool("assume-in-merge", false, "assume we are in a merge even if git dir does not indicate so")
	if err := fs.Parse(args); err != nil {
		return 1
	}

	if !*assumeInMerge && !isInMerge() {
		return 0
	}

	conflictPatterns := [][]byte{
		[]byte("<<<<<<< "),
		[]byte("======= "),
		[]byte("=======\r\n"),
		[]byte("=======\n"),
		[]byte(">>>>>>> "),
	}

	retcode := 0
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
			for _, pat := range conflictPatterns {
				if bytes.HasPrefix(line, pat) {
					fmt.Printf("%s:%d: Merge conflict string %q found\n",
						f, lineNum, strings.TrimSpace(string(pat)))
					retcode = 1
				}
			}
		}
	}
	return retcode
}

func isInMerge() bool {
	repo, err := openRepo()
	if err != nil {
		return false
	}
	gitDir := repoGitDir(repo)
	if gitDir == "" {
		return false
	}
	// MERGE_MSG must be present before we check the rest.
	if _, err := os.Stat(filepath.Join(gitDir, "MERGE_MSG")); err != nil {
		return false
	}
	// MERGE_HEAD is written by go-git / git during an active merge.
	if _, err := repo.Storer.Reference(plumbing.ReferenceName("MERGE_HEAD")); err == nil {
		return true
	}
	// rebase-apply / rebase-merge directories indicate an active rebase.
	for _, dir := range []string{"rebase-apply", "rebase-merge"} {
		if _, err := os.Stat(filepath.Join(gitDir, dir)); err == nil {
			return true
		}
	}
	return false
}
