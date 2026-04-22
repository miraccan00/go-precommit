package upstreamhooks

import (
	"testing"
)

func TestRunFixByteOrderMarker(t *testing.T) {
	bom := []byte{0xEF, 0xBB, 0xBF}

	t.Run("file with BOM is fixed and returns 1", func(t *testing.T) {
		// Arrange
		dir := t.TempDir()
		content := append(bom, []byte("hello")...)
		path := writeTestFile(t, dir, "f.txt", content)

		// Act
		got := RunFixByteOrderMarker([]string{path})

		// Assert
		if got != 1 {
			t.Errorf("got %d, want 1", got)
		}
		if string(readTestFile(t, path)) != "hello" {
			t.Errorf("BOM not removed")
		}
	})

	t.Run("file without BOM unchanged returns 0", func(t *testing.T) {
		// Arrange
		dir := t.TempDir()
		path := writeTestFile(t, dir, "f.txt", []byte("hello"))

		// Act
		got := RunFixByteOrderMarker([]string{path})

		// Assert
		if got != 0 {
			t.Errorf("got %d, want 0", got)
		}
		if string(readTestFile(t, path)) != "hello" {
			t.Errorf("file should not have changed")
		}
	})

	t.Run("no files returns 0", func(t *testing.T) {
		// Arrange + Act
		got := RunFixByteOrderMarker(nil)

		// Assert
		if got != 0 {
			t.Errorf("got %d, want 0", got)
		}
	})

	t.Run("only BOM bytes becomes empty file", func(t *testing.T) {
		// Arrange
		dir := t.TempDir()
		path := writeTestFile(t, dir, "bom-only.txt", bom)

		// Act
		got := RunFixByteOrderMarker([]string{path})

		// Assert
		if got != 1 {
			t.Errorf("got %d, want 1", got)
		}
		if len(readTestFile(t, path)) != 0 {
			t.Errorf("expected empty file after stripping BOM")
		}
	})
}
