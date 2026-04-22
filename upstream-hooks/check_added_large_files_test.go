package upstreamhooks

import (
	"bytes"
	"testing"
)

func TestRunCheckAddedLargeFiles(t *testing.T) {
	t.Run("no files returns 0", func(t *testing.T) {
		// Arrange + Act
		got := RunCheckAddedLargeFiles([]string{"--enforce-all"})

		// Assert
		if got != 0 {
			t.Errorf("got %d, want 0", got)
		}
	})

	t.Run("small file under default limit returns 0", func(t *testing.T) {
		// Arrange
		dir := t.TempDir()
		path := writeTestFile(t, dir, "small.txt", bytes.Repeat([]byte("a"), 1024)) // 1 KB

		// Act
		got := RunCheckAddedLargeFiles([]string{"--enforce-all", path})

		// Assert
		if got != 0 {
			t.Errorf("got %d, want 0 for file under limit", got)
		}
	})

	t.Run("file exceeding default 500 KB limit returns 1", func(t *testing.T) {
		// Arrange
		dir := t.TempDir()
		path := writeTestFile(t, dir, "big.bin", bytes.Repeat([]byte("x"), 600*1024)) // 600 KB

		// Act
		got := RunCheckAddedLargeFiles([]string{"--enforce-all", path})

		// Assert
		if got != 1 {
			t.Errorf("got %d, want 1 for file over limit", got)
		}
	})

	t.Run("file exceeding custom --maxkb threshold returns 1", func(t *testing.T) {
		// Arrange
		dir := t.TempDir()
		path := writeTestFile(t, dir, "medium.bin", bytes.Repeat([]byte("y"), 3*1024)) // 3 KB

		// Act — set limit to 1 KB
		got := RunCheckAddedLargeFiles([]string{"--enforce-all", "--maxkb", "1", path})

		// Assert
		if got != 1 {
			t.Errorf("got %d, want 1 for file over custom limit", got)
		}
	})

	t.Run("file under custom --maxkb threshold returns 0", func(t *testing.T) {
		// Arrange
		dir := t.TempDir()
		path := writeTestFile(t, dir, "tiny.txt", bytes.Repeat([]byte("z"), 512)) // 0.5 KB

		// Act — set limit to 1 KB
		got := RunCheckAddedLargeFiles([]string{"--enforce-all", "--maxkb", "1", path})

		// Assert
		if got != 0 {
			t.Errorf("got %d, want 0 for file under custom limit", got)
		}
	})

	t.Run("multiple files only one oversized returns 1", func(t *testing.T) {
		// Arrange
		dir := t.TempDir()
		small := writeTestFile(t, dir, "small.txt", bytes.Repeat([]byte("a"), 1024))
		large := writeTestFile(t, dir, "large.bin", bytes.Repeat([]byte("b"), 600*1024))

		// Act
		got := RunCheckAddedLargeFiles([]string{"--enforce-all", small, large})

		// Assert
		if got != 1 {
			t.Errorf("got %d, want 1 when at least one file is oversized", got)
		}
	})
}
