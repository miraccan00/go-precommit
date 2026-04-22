package upstreamhooks

import (
	"os"
	"testing"
)

func TestRunCheckShebangScriptsAreExecutable(t *testing.T) {
	t.Run("no files returns 0", func(t *testing.T) {
		// Arrange + Act
		got := RunCheckShebangScriptsAreExecutable(nil)

		// Assert
		if got != 0 {
			t.Errorf("got %d, want 0", got)
		}
	})

	t.Run("shebang script that is executable returns 0", func(t *testing.T) {
		// Arrange
		dir := t.TempDir()
		path := writeTestFile(t, dir, "script.sh", []byte("#!/bin/bash\necho hi\n"))
		if err := os.Chmod(path, 0o755); err != nil {
			t.Fatal(err)
		}

		// Act
		got := RunCheckShebangScriptsAreExecutable([]string{path})

		// Assert
		if got != 0 {
			t.Errorf("got %d, want 0", got)
		}
	})

	t.Run("shebang script that is not executable returns 1", func(t *testing.T) {
		// Arrange
		dir := t.TempDir()
		path := writeTestFile(t, dir, "script.sh", []byte("#!/bin/bash\necho hi\n"))
		// mode 0o644 — not executable (default from writeTestFile)

		// Act
		got := RunCheckShebangScriptsAreExecutable([]string{path})

		// Assert
		if got != 1 {
			t.Errorf("got %d, want 1", got)
		}
	})

	t.Run("file without shebang but not executable returns 0", func(t *testing.T) {
		// Arrange — no shebang, so the hook ignores it
		dir := t.TempDir()
		path := writeTestFile(t, dir, "plain.txt", []byte("no shebang here\n"))

		// Act
		got := RunCheckShebangScriptsAreExecutable([]string{path})

		// Assert
		if got != 0 {
			t.Errorf("got %d, want 0 for file without shebang", got)
		}
	})

	t.Run("file without shebang but executable returns 0", func(t *testing.T) {
		// Arrange — no shebang, hook does not care about executability
		dir := t.TempDir()
		path := writeTestFile(t, dir, "exec.bin", []byte{0x7F, 0x45, 0x4C, 0x46})
		if err := os.Chmod(path, 0o755); err != nil {
			t.Fatal(err)
		}

		// Act
		got := RunCheckShebangScriptsAreExecutable([]string{path})

		// Assert
		if got != 0 {
			t.Errorf("got %d, want 0 for binary without shebang marker", got)
		}
	})
}
