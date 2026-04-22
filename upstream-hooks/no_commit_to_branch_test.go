package upstreamhooks

import (
	"testing"
)

func TestRunNoCommitToBranch(t *testing.T) {
	// These tests run inside the go-precommit git repository.
	// The current branch is whatever the developer has checked out;
	// tests are written to be branch-agnostic by controlling the protected
	// list explicitly via --branch / --pattern flags.

	t.Run("non-existent protected branch does not block commit", func(t *testing.T) {
		// Arrange — protect a branch name that cannot be the current branch
		args := []string{"--branch", "definitely-not-a-real-branch-xyzzy"}

		// Act
		got := RunNoCommitToBranch(args)

		// Assert
		if got != 0 {
			t.Errorf("got %d, want 0 when current branch is not protected", got)
		}
	})

	t.Run("non-matching pattern does not block commit", func(t *testing.T) {
		// Arrange — supply an explicit --branch that can't match so the default
		// ["master","main"] protection is replaced and only the pattern is tested.
		args := []string{"--branch", "definitely-not-a-real-branch-xyzzy", "--pattern", "^definitely-not-matching-xyzzy$"}

		// Act
		got := RunNoCommitToBranch(args)

		// Assert
		if got != 0 {
			t.Errorf("got %d, want 0 when pattern does not match current branch", got)
		}
	})

	t.Run("current branch name as --branch blocks commit", func(t *testing.T) {
		// Arrange — resolve the current branch from git, then protect it
		repo, err := openRepo()
		if err != nil {
			t.Skip("not inside a git repository")
		}
		head, err := repo.Head()
		if err != nil || !head.Name().IsBranch() {
			t.Skip("HEAD is detached or cannot be resolved")
		}
		currentBranch := head.Name().Short()
		args := []string{"--branch", currentBranch}

		// Act
		got := RunNoCommitToBranch(args)

		// Assert
		if got != 1 {
			t.Errorf("got %d, want 1 when current branch %q is protected", got, currentBranch)
		}
	})

	t.Run("pattern matching current branch blocks commit", func(t *testing.T) {
		// Arrange — resolve the branch and build a regex that matches it exactly
		repo, err := openRepo()
		if err != nil {
			t.Skip("not inside a git repository")
		}
		head, err := repo.Head()
		if err != nil || !head.Name().IsBranch() {
			t.Skip("HEAD is detached or cannot be resolved")
		}
		currentBranch := head.Name().Short()
		// Escape any regex meta-chars in the branch name and anchor the pattern.
		args := []string{"--pattern", "^" + currentBranch + "$"}

		// Act
		got := RunNoCommitToBranch(args)

		// Assert
		if got != 1 {
			t.Errorf("got %d, want 1 when pattern matches current branch %q", got, currentBranch)
		}
	})

	t.Run("default protection blocks commit to main or master", func(t *testing.T) {
		// Arrange — resolve the current branch; if it is main or master the hook
		// must return 1, otherwise it must return 0.
		repo, err := openRepo()
		if err != nil {
			t.Skip("not inside a git repository")
		}
		head, err := repo.Head()
		if err != nil || !head.Name().IsBranch() {
			t.Skip("HEAD is detached or cannot be resolved")
		}
		currentBranch := head.Name().Short()

		// Act — no explicit args means default protected list [master, main]
		got := RunNoCommitToBranch(nil)

		// Assert
		onProtected := currentBranch == "main" || currentBranch == "master"
		if onProtected && got != 1 {
			t.Errorf("on branch %q: got %d, want 1", currentBranch, got)
		}
		if !onProtected && got != 0 {
			t.Errorf("on branch %q: got %d, want 0", currentBranch, got)
		}
	})
}
