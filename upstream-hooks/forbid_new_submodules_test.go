package upstreamhooks

import (
	"testing"
)

func TestRunForbidNewSubmodules(t *testing.T) {
	t.Run("repository with no new submodules staged returns 0", func(t *testing.T) {
		// Arrange — the go-precommit repo itself has no submodules; we rely on the
		// real git index which is available because tests run inside the repo.
		if _, err := openRepo(); err != nil {
			t.Skip("not inside a git repository")
		}
		args := []string{}

		// Act
		got := RunForbidNewSubmodules(args)

		// Assert
		if got != 0 {
			t.Errorf("got %d, want 0 when no new submodules are staged", got)
		}
	})

	t.Run("invalid flag causes early exit with 1", func(t *testing.T) {
		// Arrange
		args := []string{"--unknown-flag-xyzzy"}

		// Act
		got := RunForbidNewSubmodules(args)

		// Assert
		if got != 1 {
			t.Errorf("got %d, want 1 for unrecognised flag", got)
		}
	})
}
