package upstreamhooks

import (
	"testing"
)

func TestRunCheckMergeConflict(t *testing.T) {
	// All tests use --assume-in-merge to bypass the git-dir check.

	t.Run("no files returns 0", func(t *testing.T) {
		// Arrange + Act
		got := RunCheckMergeConflict([]string{"--assume-in-merge"})

		// Assert
		if got != 0 {
			t.Errorf("got %d, want 0", got)
		}
	})

	t.Run("file with conflict start marker returns 1", func(t *testing.T) {
		// Arrange
		dir := t.TempDir()
		path := writeTestFile(t, dir, "f.go", []byte("func foo() {\n<<<<<<< HEAD\n\treturn 1\n=======\n\treturn 2\n>>>>>>> branch\n}\n"))

		// Act
		got := RunCheckMergeConflict([]string{"--assume-in-merge", path})

		// Assert
		if got != 1 {
			t.Errorf("got %d, want 1", got)
		}
	})

	t.Run("file without conflict markers returns 0", func(t *testing.T) {
		// Arrange
		dir := t.TempDir()
		path := writeTestFile(t, dir, "f.go", []byte("func foo() {\n\treturn 1\n}\n"))

		// Act
		got := RunCheckMergeConflict([]string{"--assume-in-merge", path})

		// Assert
		if got != 0 {
			t.Errorf("got %d, want 0", got)
		}
	})

	t.Run("file with only equals separator returns 1", func(t *testing.T) {
		// Arrange
		dir := t.TempDir()
		path := writeTestFile(t, dir, "f.txt", []byte("line1\n======= \nline2\n"))

		// Act
		got := RunCheckMergeConflict([]string{"--assume-in-merge", path})

		// Assert
		if got != 1 {
			t.Errorf("got %d, want 1", got)
		}
	})

	t.Run("partial conflict text not matching prefix does not trigger", func(t *testing.T) {
		// Arrange — "<<<< not a conflict" doesn't start with "<<<<<<< "
		dir := t.TempDir()
		path := writeTestFile(t, dir, "f.txt", []byte("<<<< not a conflict marker\n"))

		// Act
		got := RunCheckMergeConflict([]string{"--assume-in-merge", path})

		// Assert
		if got != 0 {
			t.Errorf("got %d, want 0 for non-matching pattern", got)
		}
	})

	t.Run("without assume-in-merge not in git merge returns 0", func(t *testing.T) {
		// Arrange
		dir := t.TempDir()
		path := writeTestFile(t, dir, "f.txt", []byte("<<<<<<< HEAD\n"))

		// Act — no --assume-in-merge; isInMerge() should return false outside a real merge
		got := RunCheckMergeConflict([]string{path})

		// Assert: outside a real merge the hook skips all files
		if got != 0 {
			t.Errorf("got %d, want 0 (not in a real merge)", got)
		}
	})
}
