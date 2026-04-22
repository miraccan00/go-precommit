package upstreamhooks

import (
	"flag"
	"fmt"
	"os"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/filemode"
	"github.com/go-git/go-git/v5/plumbing/object"
)

// RunForbidNewSubmodules is the Go equivalent of forbid_new_submodules.py.
// Prevents addition of new git submodules.
// Returns 0 on success, 1 if a new submodule is detected.
func RunForbidNewSubmodules(args []string) int {
	fs := flag.NewFlagSet("forbid-new-submodules", flag.ContinueOnError)
	if err := fs.Parse(args); err != nil {
		return 1
	}

	fromRef := os.Getenv("PRE_COMMIT_FROM_REF")
	toRef := os.Getenv("PRE_COMMIT_TO_REF")

	var retv int
	if fromRef != "" && toRef != "" {
		retv = forbidSubmodulesInRange(fromRef, toRef)
	} else {
		retv = forbidSubmodulesStaged()
	}

	if retv != 0 {
		fmt.Println()
		fmt.Println("This commit introduces new submodules.")
		fmt.Println("Did you unintentionally `git add .`?")
		fmt.Println("To fix: git rm {thesubmodule}  # no trailing slash")
		fmt.Println("Also check .gitmodules")
	}
	return retv
}

// forbidSubmodulesStaged checks the staging area (index vs HEAD) for newly added submodules.
func forbidSubmodulesStaged() int {
	repo, err := openRepo()
	if err != nil {
		return 0
	}

	// HEAD tree — nil on the very first commit (no HEAD yet).
	var headTree *object.Tree
	if ref, err := repo.Head(); err == nil {
		if commit, err := repo.CommitObject(ref.Hash()); err == nil {
			headTree, _ = commit.Tree()
		}
	}

	idx, err := repo.Storer.Index()
	if err != nil {
		return 0
	}

	retv := 0
	for _, entry := range idx.Entries {
		if entry.Mode != filemode.Submodule {
			continue
		}
		// Newly added = not present in HEAD.
		if headTree != nil {
			if _, err := headTree.File(entry.Name); err == nil {
				continue // already existed in HEAD
			}
		}
		fmt.Printf("%s: new submodule introduced\n", entry.Name)
		retv = 1
	}
	return retv
}

// forbidSubmodulesInRange checks commits between fromRefStr..toRefStr (pre-push).
func forbidSubmodulesInRange(fromRefStr, toRefStr string) int {
	repo, err := openRepo()
	if err != nil {
		return 0
	}

	fromHash, err := repo.ResolveRevision(plumbing.Revision(fromRefStr))
	if err != nil {
		return 0
	}
	toHash, err := repo.ResolveRevision(plumbing.Revision(toRefStr))
	if err != nil {
		return 0
	}

	fromCommit, err := repo.CommitObject(*fromHash)
	if err != nil {
		return 0
	}
	toCommit, err := repo.CommitObject(*toHash)
	if err != nil {
		return 0
	}

	fromTree, err := fromCommit.Tree()
	if err != nil {
		return 0
	}
	toTree, err := toCommit.Tree()
	if err != nil {
		return 0
	}

	changes, err := fromTree.Diff(toTree)
	if err != nil {
		return 0
	}

	retv := 0
	for _, change := range changes {
		// An added entry has an empty From side.
		if change.From.Name == "" && change.To.TreeEntry.Mode == filemode.Submodule {
			fmt.Printf("%s: new submodule introduced\n", change.To.Name)
			retv = 1
		}
	}
	return retv
}
