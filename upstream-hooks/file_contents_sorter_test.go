package upstreamhooks

import (
	"testing"
)

func TestSortFileContents(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		ignoreCase  bool
		unique      bool
		wantChanged bool
		wantContent string
	}{
		{
			name:        "already sorted unchanged",
			input:       "apple\nbanana\ncherry\n",
			wantChanged: false,
			wantContent: "apple\nbanana\ncherry\n",
		},
		{
			name:        "unsorted gets sorted",
			input:       "cherry\napple\nbanana\n",
			wantChanged: true,
			wantContent: "apple\nbanana\ncherry\n",
		},
		{
			name:        "case-sensitive sort uppercase before lowercase",
			input:       "banana\nApple\n",
			ignoreCase:  false,
			wantChanged: true,
			wantContent: "Apple\nbanana\n",
		},
		{
			name:        "case-insensitive sort",
			input:       "Cherry\napple\nBanana\n",
			ignoreCase:  true,
			wantChanged: true,
			wantContent: "apple\nBanana\nCherry\n",
		},
		{
			name:        "unique removes duplicates",
			input:       "banana\napple\nbanana\n",
			unique:      true,
			wantChanged: true,
			wantContent: "apple\nbanana\n",
		},
		{
			name:        "empty lines skipped",
			input:       "banana\n\napple\n",
			wantChanged: true,
			wantContent: "apple\nbanana\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			dir := t.TempDir()
			path := writeTestFile(t, dir, "f.txt", []byte(tt.input))

			// Act
			changed := sortFileContents(path, tt.ignoreCase, tt.unique)

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

func TestRunFileContentsSorter(t *testing.T) {
	t.Run("no files returns 0", func(t *testing.T) {
		// Arrange + Act
		got := RunFileContentsSorter(nil)

		// Assert
		if got != 0 {
			t.Errorf("got %d, want 0", got)
		}
	})

	t.Run("unsorted file returns 1", func(t *testing.T) {
		// Arrange
		dir := t.TempDir()
		path := writeTestFile(t, dir, "f.txt", []byte("z\na\nm\n"))

		// Act
		got := RunFileContentsSorter([]string{path})

		// Assert
		if got != 1 {
			t.Errorf("got %d, want 1", got)
		}
		if string(readTestFile(t, path)) != "a\nm\nz\n" {
			t.Errorf("file not sorted correctly")
		}
	})

	t.Run("ignore-case flag respected", func(t *testing.T) {
		// Arrange
		dir := t.TempDir()
		path := writeTestFile(t, dir, "f.txt", []byte("Zebra\napple\nMango\n"))

		// Act
		got := RunFileContentsSorter([]string{"--ignore-case", path})

		// Assert
		if got != 1 {
			t.Errorf("got %d, want 1", got)
		}
	})
}
