package upstreamhooks

import (
	"testing"
)

func TestRunCheckVCSPermalinks(t *testing.T) {
	t.Run("no files returns 0", func(t *testing.T) {
		// Arrange + Act
		got := RunCheckVCSPermalinks(nil)

		// Assert
		if got != 0 {
			t.Errorf("got %d, want 0", got)
		}
	})

	t.Run("permanent link with full commit hash returns 0", func(t *testing.T) {
		// Arrange — 40-char hex hash is a permalink
		dir := t.TempDir()
		path := writeTestFile(t, dir, "f.md",
			[]byte("see https://github.com/user/repo/blob/abc123def456abc123def456abc123def456abc1/file.py#L10\n"))

		// Act
		got := RunCheckVCSPermalinks([]string{path})

		// Assert
		if got != 0 {
			t.Errorf("got %d, want 0 for permanent link", got)
		}
	})

	t.Run("non-permanent branch link returns 1", func(t *testing.T) {
		// Arrange — branch name "main" is not a hex hash
		dir := t.TempDir()
		path := writeTestFile(t, dir, "f.md",
			[]byte("see https://github.com/user/repo/blob/main/file.py#L10\n"))

		// Act
		got := RunCheckVCSPermalinks([]string{path})

		// Assert
		if got != 1 {
			t.Errorf("got %d, want 1 for non-permanent link", got)
		}
	})

	t.Run("non-permanent master branch link returns 1", func(t *testing.T) {
		// Arrange
		dir := t.TempDir()
		path := writeTestFile(t, dir, "f.md",
			[]byte("ref https://github.com/user/repo/blob/master/src/main.go#L42\n"))

		// Act
		got := RunCheckVCSPermalinks([]string{path})

		// Assert
		if got != 1 {
			t.Errorf("got %d, want 1", got)
		}
	})

	t.Run("file without any github links returns 0", func(t *testing.T) {
		// Arrange
		dir := t.TempDir()
		path := writeTestFile(t, dir, "f.md", []byte("No links here.\n"))

		// Act
		got := RunCheckVCSPermalinks([]string{path})

		// Assert
		if got != 0 {
			t.Errorf("got %d, want 0", got)
		}
	})

	t.Run("custom domain non-permanent link returns 1", func(t *testing.T) {
		// Arrange
		dir := t.TempDir()
		path := writeTestFile(t, dir, "f.md",
			[]byte("see https://github.example.com/user/repo/blob/develop/file.py#L5\n"))

		// Act
		got := RunCheckVCSPermalinks([]string{"--additional-github-domain", "github.example.com", path})

		// Assert
		if got != 1 {
			t.Errorf("got %d, want 1 for non-permanent link on custom domain", got)
		}
	})

	t.Run("custom domain permanent link returns 0", func(t *testing.T) {
		// Arrange
		dir := t.TempDir()
		path := writeTestFile(t, dir, "f.md",
			[]byte("see https://github.example.com/user/repo/blob/abc123def456abc1234567890abcdef123456789a/file.py#L5\n"))

		// Act
		got := RunCheckVCSPermalinks([]string{"--additional-github-domain", "github.example.com", path})

		// Assert
		if got != 0 {
			t.Errorf("got %d, want 0 for permanent link on custom domain", got)
		}
	})
}
