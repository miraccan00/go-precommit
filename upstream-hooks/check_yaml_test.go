package upstreamhooks

import (
	"testing"
)

func TestParseYAML(t *testing.T) {
	tests := []struct {
		name    string
		input   []byte
		multi   bool
		wantErr bool
	}{
		{
			name:    "valid single document",
			input:   []byte("key: value\n"),
			wantErr: false,
		},
		{
			name:    "valid mapping",
			input:   []byte("a: 1\nb: 2\n"),
			wantErr: false,
		},
		{
			name:    "invalid YAML",
			input:   []byte("key: :\tbad"),
			wantErr: true,
		},
		{
			name:    "multi-doc with flag stops at first without multi",
			input:   []byte("---\na: 1\n---\nb: 2\n"),
			multi:   false,
			wantErr: false,
		},
		{
			name:    "multi-doc with multi flag parses all",
			input:   []byte("---\na: 1\n---\nb: 2\n"),
			multi:   true,
			wantErr: false,
		},
		{
			name:    "empty document",
			input:   []byte(""),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange: input prepared above

			// Act
			err := parseYAML(tt.input, tt.multi)

			// Assert
			if (err != nil) != tt.wantErr {
				t.Errorf("parseYAML error=%v, wantErr=%v", err, tt.wantErr)
			}
		})
	}
}

func TestRunCheckYAML(t *testing.T) {
	t.Run("no files returns 0", func(t *testing.T) {
		// Arrange + Act
		got := RunCheckYAML(nil)

		// Assert
		if got != 0 {
			t.Errorf("got %d, want 0", got)
		}
	})

	t.Run("valid YAML file returns 0", func(t *testing.T) {
		// Arrange
		dir := t.TempDir()
		path := writeTestFile(t, dir, "f.yaml", []byte("name: test\nvalue: 42\n"))

		// Act
		got := RunCheckYAML([]string{path})

		// Assert
		if got != 0 {
			t.Errorf("got %d, want 0", got)
		}
	})

	t.Run("invalid YAML file returns 1", func(t *testing.T) {
		// Arrange
		dir := t.TempDir()
		path := writeTestFile(t, dir, "bad.yaml", []byte("key: :\tinvalid"))

		// Act
		got := RunCheckYAML([]string{path})

		// Assert
		if got != 1 {
			t.Errorf("got %d, want 1", got)
		}
	})
}
