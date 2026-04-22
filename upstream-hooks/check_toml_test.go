package upstreamhooks

import (
	"testing"
)

func TestCheckTOML(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "valid key-value pair",
			input:   "key = \"value\"\n",
			wantErr: false,
		},
		{
			name:    "valid section header",
			input:   "[section]\nkey = 1\n",
			wantErr: false,
		},
		{
			name:    "comment line skipped",
			input:   "# this is a comment\nkey = 1\n",
			wantErr: false,
		},
		{
			name:    "empty file valid",
			input:   "",
			wantErr: false,
		},
		{
			name:    "blank lines skipped",
			input:   "\n\nkey = 1\n\n",
			wantErr: false,
		},
		{
			name:    "unexpected bare value is error",
			input:   "just_a_value\n",
			wantErr: true,
		},
		{
			name:    "valid numeric value",
			input:   "count = 42\n",
			wantErr: false,
		},
		{
			name:    "valid boolean value",
			input:   "enabled = true\n",
			wantErr: false,
		},
		{
			name:    "dotted key valid",
			input:   "a.b = 1\n",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			data := []byte(tt.input)

			// Act
			err := checkTOML("test.toml", data)

			// Assert
			if (err != nil) != tt.wantErr {
				t.Errorf("checkTOML error=%v, wantErr=%v", err, tt.wantErr)
			}
		})
	}
}

func TestRunCheckTOML(t *testing.T) {
	t.Run("no files returns 0", func(t *testing.T) {
		// Arrange + Act
		got := RunCheckTOML(nil)

		// Assert
		if got != 0 {
			t.Errorf("got %d, want 0", got)
		}
	})

	t.Run("valid TOML file returns 0", func(t *testing.T) {
		// Arrange
		dir := t.TempDir()
		path := writeTestFile(t, dir, "f.toml", []byte("[package]\nname = \"myapp\"\nversion = \"1.0\"\n"))

		// Act
		got := RunCheckTOML([]string{path})

		// Assert
		if got != 0 {
			t.Errorf("got %d, want 0", got)
		}
	})

	t.Run("invalid TOML file returns 1", func(t *testing.T) {
		// Arrange
		dir := t.TempDir()
		path := writeTestFile(t, dir, "bad.toml", []byte("this is not toml\n"))

		// Act
		got := RunCheckTOML([]string{path})

		// Assert
		if got != 1 {
			t.Errorf("got %d, want 1", got)
		}
	})
}
