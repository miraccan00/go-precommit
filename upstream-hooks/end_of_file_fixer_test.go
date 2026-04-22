package upstreamhooks

import (
	"testing"
)

func TestFixEndOfFile(t *testing.T) {
	tests := []struct {
		name        string
		input       []byte
		wantChanged bool
		wantContent []byte
	}{
		{
			name:        "already ends with single newline",
			input:       []byte("hello\n"),
			wantChanged: false,
			wantContent: []byte("hello\n"),
		},
		{
			name:        "missing trailing newline",
			input:       []byte("hello"),
			wantChanged: true,
			wantContent: []byte("hello\n"),
		},
		{
			name:        "multiple trailing newlines",
			input:       []byte("hello\n\n\n"),
			wantChanged: true,
			wantContent: []byte("hello\n"),
		},
		{
			name:        "only newlines becomes empty",
			input:       []byte("\n\n"),
			wantChanged: true,
			wantContent: []byte{},
		},
		{
			name:        "empty file unchanged",
			input:       []byte{},
			wantChanged: false,
			wantContent: []byte{},
		},
		{
			name:        "CRLF trailing stripped to single LF",
			input:       []byte("hello\r\n\r\n"),
			wantChanged: true,
			wantContent: []byte("hello\n"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			dir := t.TempDir()
			path := writeTestFile(t, dir, "file.txt", tt.input)

			// Act
			changed := fixEndOfFile(path)

			// Assert
			if changed != tt.wantChanged {
				t.Errorf("fixEndOfFile changed=%v, want %v", changed, tt.wantChanged)
			}
			got := readTestFile(t, path)
			if string(got) != string(tt.wantContent) {
				t.Errorf("content = %q, want %q", got, tt.wantContent)
			}
		})
	}
}

func TestRunEndOfFileFixer(t *testing.T) {
	t.Run("no files returns 0", func(t *testing.T) {
		// Arrange: no args

		// Act
		got := RunEndOfFileFixer(nil)

		// Assert
		if got != 0 {
			t.Errorf("RunEndOfFileFixer(nil) = %d, want 0", got)
		}
	})

	t.Run("file needing fix returns 1", func(t *testing.T) {
		// Arrange
		dir := t.TempDir()
		path := writeTestFile(t, dir, "noeol.txt", []byte("hello"))

		// Act
		got := RunEndOfFileFixer([]string{path})

		// Assert
		if got != 1 {
			t.Errorf("RunEndOfFileFixer = %d, want 1", got)
		}
		if string(readTestFile(t, path)) != "hello\n" {
			t.Errorf("file not fixed")
		}
	})

	t.Run("already correct file returns 0", func(t *testing.T) {
		// Arrange
		dir := t.TempDir()
		path := writeTestFile(t, dir, "ok.txt", []byte("hello\n"))

		// Act
		got := RunEndOfFileFixer([]string{path})

		// Assert
		if got != 0 {
			t.Errorf("RunEndOfFileFixer = %d, want 0", got)
		}
	})

	t.Run("mixed files returns 1", func(t *testing.T) {
		// Arrange
		dir := t.TempDir()
		ok := writeTestFile(t, dir, "ok.txt", []byte("ok\n"))
		bad := writeTestFile(t, dir, "bad.txt", []byte("bad"))

		// Act
		got := RunEndOfFileFixer([]string{ok, bad})

		// Assert
		if got != 1 {
			t.Errorf("RunEndOfFileFixer = %d, want 1", got)
		}
	})
}
