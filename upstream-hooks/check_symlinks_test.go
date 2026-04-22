package upstreamhooks

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRunCheckSymlinks(t *testing.T) {
	t.Run("no files returns 0", func(t *testing.T) {
		// Arrange + Act
		got := RunCheckSymlinks(nil)

		// Assert
		if got != 0 {
			t.Errorf("got %d, want 0", got)
		}
	})

	t.Run("regular file returns 0", func(t *testing.T) {
		// Arrange
		dir := t.TempDir()
		path := writeTestFile(t, dir, "regular.txt", []byte("content"))

		// Act
		got := RunCheckSymlinks([]string{path})

		// Assert
		if got != 0 {
			t.Errorf("got %d, want 0 for regular file", got)
		}
	})

	t.Run("valid symlink returns 0", func(t *testing.T) {
		// Arrange
		dir := t.TempDir()
		target := writeTestFile(t, dir, "target.txt", []byte("content"))
		link := filepath.Join(dir, "link.txt")
		if err := os.Symlink(target, link); err != nil {
			t.Fatal(err)
		}

		// Act
		got := RunCheckSymlinks([]string{link})

		// Assert
		if got != 0 {
			t.Errorf("got %d, want 0 for valid symlink", got)
		}
	})

	t.Run("broken symlink returns 1", func(t *testing.T) {
		// Arrange
		dir := t.TempDir()
		link := filepath.Join(dir, "broken.txt")
		if err := os.Symlink(filepath.Join(dir, "nonexistent.txt"), link); err != nil {
			t.Fatal(err)
		}

		// Act
		got := RunCheckSymlinks([]string{link})

		// Assert
		if got != 1 {
			t.Errorf("got %d, want 1 for broken symlink", got)
		}
	})

	t.Run("mix of regular file and broken symlink returns 1", func(t *testing.T) {
		// Arrange
		dir := t.TempDir()
		regular := writeTestFile(t, dir, "regular.txt", []byte("ok"))
		broken := filepath.Join(dir, "broken.txt")
		if err := os.Symlink(filepath.Join(dir, "nowhere"), broken); err != nil {
			t.Fatal(err)
		}

		// Act
		got := RunCheckSymlinks([]string{regular, broken})

		// Assert
		if got != 1 {
			t.Errorf("got %d, want 1", got)
		}
	})
}
