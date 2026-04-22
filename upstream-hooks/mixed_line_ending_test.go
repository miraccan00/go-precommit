package upstreamhooks

import (
	"reflect"
	"testing"
)

func TestSplitOnAnyEnding(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
		want  [][]byte
	}{
		{
			name:  "pure LF",
			input: []byte("a\nb\nc"),
			want:  [][]byte{[]byte("a"), []byte("b"), []byte("c")},
		},
		{
			name:  "pure LF with trailing newline",
			input: []byte("a\nb\n"),
			want:  [][]byte{[]byte("a"), []byte("b"), []byte("")},
		},
		{
			name:  "pure CRLF",
			input: []byte("a\r\nb\r\n"),
			want:  [][]byte{[]byte("a"), []byte("b"), []byte("")},
		},
		{
			name:  "pure CR",
			input: []byte("a\rb"),
			want:  [][]byte{[]byte("a"), []byte("b")},
		},
		{
			name:  "mixed CRLF and LF",
			input: []byte("a\r\nb\nc"),
			want:  [][]byte{[]byte("a"), []byte("b"), []byte("c")},
		},
		{
			name:  "empty input",
			input: []byte(""),
			want:  [][]byte{[]byte("")},
		},
		{
			name:  "no line endings",
			input: []byte("hello"),
			want:  [][]byte{[]byte("hello")},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange: input prepared above

			// Act
			got := splitOnAnyEnding(tt.input)

			// Assert
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("splitOnAnyEnding(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestFixMixedLineEnding(t *testing.T) {
	tests := []struct {
		name      string
		input     []byte
		fix       string
		wantChg   bool
		wantMixed bool
	}{
		{
			name:      "pure LF with fix=lf no change",
			input:     []byte("a\nb\n"),
			fix:       "lf",
			wantChg:   false,
			wantMixed: false,
		},
		{
			name:      "pure CRLF with fix=crlf no change",
			input:     []byte("a\r\nb\r\n"),
			fix:       "crlf",
			wantChg:   false,
			wantMixed: false,
		},
		{
			name:      "mixed CRLF and LF detected with fix=no",
			input:     []byte("a\r\nb\n"),
			fix:       "no",
			wantChg:   false,
			wantMixed: true,
		},
		{
			name:      "mixed endings fixed to lf",
			input:     []byte("a\r\nb\n"),
			fix:       "lf",
			wantChg:   true,
			wantMixed: true,
		},
		{
			name:      "mixed endings fixed to crlf",
			input:     []byte("a\r\nb\n"),
			fix:       "crlf",
			wantChg:   true,
			wantMixed: true,
		},
		{
			name:      "auto selects dominant ending",
			input:     []byte("a\nb\nc\r\n"),
			fix:       "auto",
			wantChg:   true,
			wantMixed: true,
		},
		{
			name:      "pure LF with fix=auto no change",
			input:     []byte("a\nb\n"),
			fix:       "auto",
			wantChg:   false,
			wantMixed: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			dir := t.TempDir()
			path := writeTestFile(t, dir, "f.txt", tt.input)

			// Act
			gotChg, gotMixed := fixMixedLineEnding(path, tt.fix)

			// Assert
			if gotChg != tt.wantChg {
				t.Errorf("changed=%v, want %v", gotChg, tt.wantChg)
			}
			if gotMixed != tt.wantMixed {
				t.Errorf("wasMixed=%v, want %v", gotMixed, tt.wantMixed)
			}
		})
	}
}

func TestRunMixedLineEnding(t *testing.T) {
	t.Run("pure LF file with default fix returns 0", func(t *testing.T) {
		// Arrange
		dir := t.TempDir()
		path := writeTestFile(t, dir, "f.txt", []byte("a\nb\n"))

		// Act
		got := RunMixedLineEnding([]string{path})

		// Assert
		if got != 0 {
			t.Errorf("got %d, want 0", got)
		}
	})

	t.Run("mixed line endings with fix=no returns 1", func(t *testing.T) {
		// Arrange
		dir := t.TempDir()
		path := writeTestFile(t, dir, "f.txt", []byte("a\r\nb\n"))

		// Act
		got := RunMixedLineEnding([]string{"--fix", "no", path})

		// Assert
		if got != 1 {
			t.Errorf("got %d, want 1", got)
		}
	})
}
