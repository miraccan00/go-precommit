package upstreamhooks

import (
	"flag"
	"fmt"
	"regexp"
	"strings"

	"github.com/go-git/go-git/v5/plumbing"
)

// RunNoCommitToBranch is the Go equivalent of no_commit_to_branch.py.
// Prevents direct commits to protected branches (default: main, master).
// Returns 0 if the current branch is allowed, 1 if it is protected.
func RunNoCommitToBranch(args []string) int {
	fs := flag.NewFlagSet("no-commit-to-branch", flag.ContinueOnError)
	var branches, patterns multiStringFlag
	fs.Var(&branches, "branch", "branch to disallow commits to (may be repeated)")
	fs.Var(&patterns, "pattern", "regex pattern for branch to disallow (may be repeated)")
	if err := fs.Parse(args); err != nil {
		return 1
	}

	protected := []string{"master", "main"}
	if len(branches) > 0 {
		protected = branches
	}

	repo, err := openRepo()
	if err != nil {
		return 0 // can't open repo — allow commit
	}
	headRef, err := repo.Storer.Reference(plumbing.HEAD)
	if err != nil || headRef.Type() != plumbing.SymbolicReference {
		// Detached HEAD — not on any branch.
		return 0
	}
	// target is e.g. "refs/heads/main" — strip the "refs/heads/" prefix
	// the same way the original did: split on "/" and join from index 2.
	chunks := strings.Split(string(headRef.Target()), "/")
	branch := strings.Join(chunks[2:], "/")

	for _, b := range protected {
		if branch == b {
			fmt.Printf("do not commit to %q\n", branch)
			return 1
		}
	}
	for _, p := range patterns {
		if matched, _ := regexp.MatchString(p, branch); matched {
			fmt.Printf("do not commit to %q (matches pattern %q)\n", branch, p)
			return 1
		}
	}
	return 0
}
