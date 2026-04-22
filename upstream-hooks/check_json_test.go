package upstreamhooks

import (
	"testing"
)

func TestCheckJSONNoDuplicateKeys(t *testing.T) {
	tests := []struct {
		name    string
		input   []byte
		wantErr bool
	}{
		{
			name:    "valid object",
			input:   []byte(`{"a": 1, "b": 2}`),
			wantErr: false,
		},
		{
			name:    "valid array",
			input:   []byte(`[1, 2, 3]`),
			wantErr: false,
		},
		{
			name:    "valid nested",
			input:   []byte(`{"a": {"b": 1}}`),
			wantErr: false,
		},
		{
			name:    "invalid JSON missing bracket",
			input:   []byte(`{"a": 1`),
			wantErr: true,
		},
		{
			name:    "invalid JSON unquoted key",
			input:   []byte(`{a: 1}`),
			wantErr: true,
		},
		{
			name:    "empty object",
			input:   []byte(`{}`),
			wantErr: false,
		},
		{
			name:    "empty array",
			input:   []byte(`[]`),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange: input prepared above

			// Act
			err := checkJSONNoDuplicateKeys(tt.input)

			// Assert
			if (err != nil) != tt.wantErr {
				t.Errorf("checkJSONNoDuplicateKeys error=%v, wantErr=%v", err, tt.wantErr)
			}
		})
	}
}

func TestRunCheckJSON(t *testing.T) {
	t.Run("no files returns 0", func(t *testing.T) {
		// Arrange + Act
		got := RunCheckJSON(nil)

		// Assert
		if got != 0 {
			t.Errorf("got %d, want 0", got)
		}
	})

	t.Run("valid JSON file returns 0", func(t *testing.T) {
		// Arrange
		dir := t.TempDir()
		path := writeTestFile(t, dir, "f.json", []byte(`{"key": "value"}`))

		// Act
		got := RunCheckJSON([]string{path})

		// Assert
		if got != 0 {
			t.Errorf("got %d, want 0", got)
		}
	})

	t.Run("invalid JSON file returns 1", func(t *testing.T) {
		// Arrange
		dir := t.TempDir()
		path := writeTestFile(t, dir, "bad.json", []byte(`{bad json}`))

		// Act
		got := RunCheckJSON([]string{path})

		// Assert
		if got != 1 {
			t.Errorf("got %d, want 1", got)
		}
	})

	t.Run("mixed valid and invalid returns 1", func(t *testing.T) {
		// Arrange
		dir := t.TempDir()
		valid := writeTestFile(t, dir, "ok.json", []byte(`{"a":1}`))
		invalid := writeTestFile(t, dir, "bad.json", []byte(`not json`))

		// Act
		got := RunCheckJSON([]string{valid, invalid})

		// Assert
		if got != 1 {
			t.Errorf("got %d, want 1", got)
		}
	})
}
