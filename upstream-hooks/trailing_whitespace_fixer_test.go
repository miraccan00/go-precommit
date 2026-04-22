package upstreamhooks

import (
	"testing"
)

func TestFixTrailingWhitespace(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		isMarkdown  bool
		chars       string
		wantChanged bool
		wantContent string
	}{
		{
			name:        "no trailing whitespace unchanged",
			input:       "hello world\n",
			wantChanged: false,
			wantContent: "hello world\n",
		},
		{
			name:        "trailing spaces removed",
			input:       "hello   \nworld\n",
			wantChanged: true,
			wantContent: "hello\nworld\n",
		},
		{
			name:        "trailing tabs removed",
			input:       "hello\t\t\nworld\n",
			wantChanged: true,
			wantContent: "hello\nworld\n",
		},
		{
			name:        "markdown preserves double trailing space",
			input:       "hello  \n",
			isMarkdown:  true,
			wantChanged: false,
			wantContent: "hello  \n",
		},
		{
			name:        "markdown strips single trailing space",
			input:       "hello \n",
			isMarkdown:  true,
			wantChanged: true,
			wantContent: "hello\n",
		},
		{
			name:        "custom chars removed",
			input:       "hello!\n",
			chars:       "!",
			wantChanged: true,
			wantContent: "hello\n",
		},
		{
			name:        "empty file unchanged",
			input:       "",
			wantChanged: false,
			wantContent: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			dir := t.TempDir()
			path := writeTestFile(t, dir, "f.txt", []byte(tt.input))

			// Act
			changed := fixTrailingWhitespace(path, tt.isMarkdown, tt.chars)

			// Assert
			if changed != tt.wantChanged {
				t.Errorf("changed=%v, want %v", changed, tt.wantChanged)
			}
			got := string(readTestFile(t, path))
			if got != tt.wantContent {
				t.Errorf("content = %q, want %q", got, tt.wantContent)
			}
		})
	}
}

func TestRunTrailingWhitespaceFixer(t *testing.T) {
	t.Run("no files returns 0", func(t *testing.T) {
		// Arrange + Act
		got := RunTrailingWhitespaceFixer(nil)

		// Assert
		if got != 0 {
			t.Errorf("got %d, want 0", got)
		}
	})

	t.Run("file with trailing spaces returns 1 and fixes file", func(t *testing.T) {
		// Arrange
		dir := t.TempDir()
		path := writeTestFile(t, dir, "f.txt", []byte("line   \n"))

		// Act
		got := RunTrailingWhitespaceFixer([]string{path})

		// Assert
		if got != 1 {
			t.Errorf("got %d, want 1", got)
		}
		if string(readTestFile(t, path)) != "line\n" {
			t.Errorf("file not fixed")
		}
	})
}
