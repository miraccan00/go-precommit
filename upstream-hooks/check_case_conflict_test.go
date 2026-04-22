package upstreamhooks

import (
	"reflect"
	"testing"
)

func TestParentDirs(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{
			name:  "top-level file has no parents",
			input: "file.go",
			want:  nil,
		},
		{
			name:  "one directory deep",
			input: "src/file.go",
			want:  []string{"src"},
		},
		{
			name:  "two directories deep",
			input: "a/b/file.go",
			want:  []string{"a/b", "a"},
		},
		{
			name:  "three directories deep",
			input: "a/b/c/file.go",
			want:  []string{"a/b/c", "a/b", "a"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange: input prepared above

			// Act
			got := parentDirs(tt.input)

			// Assert
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parentDirs(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestLowerSet(t *testing.T) {
	tests := []struct {
		name  string
		input map[string]bool
		want  map[string]bool
	}{
		{
			name:  "empty map returns empty map",
			input: map[string]bool{},
			want:  map[string]bool{},
		},
		{
			name:  "already lowercase unchanged",
			input: map[string]bool{"foo": true, "bar": true},
			want:  map[string]bool{"foo": true, "bar": true},
		},
		{
			name:  "uppercase keys are lowercased",
			input: map[string]bool{"FOO": true, "Bar": true},
			want:  map[string]bool{"foo": true, "bar": true},
		},
		{
			name:  "mixed-case duplicates collapse to one key",
			input: map[string]bool{"README": true, "readme": true},
			want:  map[string]bool{"readme": true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange: input prepared above

			// Act
			got := lowerSet(tt.input)

			// Assert
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("lowerSet(%v) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestRunCheckCaseConflict(t *testing.T) {
	t.Run("no files returns 0", func(t *testing.T) {
		// Arrange — RunCheckCaseConflict calls trackedFiles() which requires a git repo.
		if _, err := openRepo(); err != nil {
			t.Skip("not inside a git repository")
		}

		// Act
		got := RunCheckCaseConflict(nil)

		// Assert
		if got != 0 {
			t.Errorf("got %d, want 0", got)
		}
	})

	t.Run("two files differing only in case within the same call conflict", func(t *testing.T) {
		// Arrange — pass two filenames that differ only in case; both are treated
		// as "relevant" new files and the hook detects the collision between them.
		if _, err := openRepo(); err != nil {
			t.Skip("not inside a git repository")
		}
		dir := t.TempDir()
		a := writeTestFile(t, dir, "Readme.md", []byte("a"))
		b := writeTestFile(t, dir, "readme.md", []byte("b"))

		// Act
		got := RunCheckCaseConflict([]string{a, b})

		// Assert
		if got != 1 {
			t.Errorf("got %d, want 1 for case-conflicting filenames", got)
		}
	})

	t.Run("files with distinct names in the same directory do not conflict", func(t *testing.T) {
		// Arrange
		if _, err := openRepo(); err != nil {
			t.Skip("not inside a git repository")
		}
		dir := t.TempDir()
		a := writeTestFile(t, dir, "alpha.txt", []byte("a"))
		b := writeTestFile(t, dir, "beta.txt", []byte("b"))

		// Act
		got := RunCheckCaseConflict([]string{a, b})

		// Assert
		if got != 0 {
			t.Errorf("got %d, want 0 for distinct filenames", got)
		}
	})
}
