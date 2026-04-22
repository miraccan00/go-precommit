package upstreamhooks

import (
	"testing"
)

func TestRunDestroyedSymlinks(t *testing.T) {
	t.Run("no files returns 0", func(t *testing.T) {
		// Arrange + Act
		got := RunDestroyedSymlinks(nil)

		// Assert
		if got != 0 {
			t.Errorf("got %d, want 0", got)
		}
	})

	t.Run("invalid flag causes early exit with 1", func(t *testing.T) {
		// Arrange
		args := []string{"--unknown-flag-xyzzy"}

		// Act
		got := RunDestroyedSymlinks(args)

		// Assert
		if got != 1 {
			t.Errorf("got %d, want 1 for unrecognised flag", got)
		}
	})

	t.Run("file not in git index returns 0", func(t *testing.T) {
		// Arrange — file is in a temp dir so it will not be in the git index;
		// findDestroyedSymlinks skips any path not present in the index.
		repo, err := openRepo()
		if err != nil {
			t.Skip("not inside a git repository")
		}
		if _, err := repo.Head(); err != nil {
			t.Skip("HEAD is not available")
		}
		dir := t.TempDir()
		path := writeTestFile(t, dir, "untracked.txt", []byte("content"))

		// Act
		got := RunDestroyedSymlinks([]string{path})

		// Assert
		if got != 0 {
			t.Errorf("got %d, want 0 for untracked file", got)
		}
	})
}

func TestTrimBytes(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "no trailing newlines unchanged",
			input: "hello",
			want:  "hello",
		},
		{
			name:  "trailing LF stripped",
			input: "hello\n",
			want:  "hello",
		},
		{
			name:  "trailing CRLF stripped",
			input: "hello\r\n",
			want:  "hello",
		},
		{
			name:  "multiple trailing newlines stripped",
			input: "hello\n\n\r\n",
			want:  "hello",
		},
		{
			name:  "internal newlines preserved",
			input: "hel\nlo\n",
			want:  "hel\nlo",
		},
		{
			name:  "empty string unchanged",
			input: "",
			want:  "",
		},
		{
			name:  "only newlines becomes empty",
			input: "\n\r\n",
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange: input prepared above

			// Act
			got := trimBytes(tt.input)

			// Assert
			if got != tt.want {
				t.Errorf("trimBytes(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
