package upstreamhooks

import (
	"strings"
	"testing"
)

func TestPrettyJSON(t *testing.T) {
	t.Run("basic indentation applied", func(t *testing.T) {
		// Arrange
		input := []byte(`{"b":1,"a":2}`)

		// Act
		out, err := prettyJSON(input, "  ", true, nil)

		// Assert
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		got := string(out)
		if !strings.Contains(got, "\n  ") {
			t.Errorf("expected indented output, got %q", got)
		}
		if !strings.HasSuffix(got, "\n") {
			t.Errorf("output should end with newline")
		}
	})

	t.Run("keys sorted alphabetically", func(t *testing.T) {
		// Arrange
		input := []byte(`{"z": 1, "a": 2, "m": 3}`)

		// Act
		out, err := prettyJSON(input, "  ", true, nil)

		// Assert
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		got := string(out)
		aPos := strings.Index(got, `"a"`)
		mPos := strings.Index(got, `"m"`)
		zPos := strings.Index(got, `"z"`)
		if aPos > mPos || mPos > zPos {
			t.Errorf("keys not sorted: %q", got)
		}
	})

	t.Run("invalid JSON returns error", func(t *testing.T) {
		// Arrange
		input := []byte(`not json`)

		// Act
		_, err := prettyJSON(input, "  ", true, nil)

		// Assert
		if err == nil {
			t.Errorf("expected error for invalid JSON")
		}
	})

	t.Run("numeric indent string expanded to spaces", func(t *testing.T) {
		// Arrange
		dir := t.TempDir()
		path := writeTestFile(t, dir, "f.json", []byte(`{"a":1}`))

		// Act — RunPrettyFormatJSON resolves numeric indent
		got := RunPrettyFormatJSON([]string{"--indent", "2", "--autofix", path})

		// Assert
		if got != 1 {
			// File was not already pretty-formatted with 2-space indent
			t.Errorf("got %d, want 1 (file should be reformatted)", got)
		}
	})
}

func TestRunPrettyFormatJSON(t *testing.T) {
	t.Run("no files returns 0", func(t *testing.T) {
		// Arrange + Act
		got := RunPrettyFormatJSON(nil)

		// Assert
		if got != 0 {
			t.Errorf("got %d, want 0", got)
		}
	})

	t.Run("already pretty formatted returns 0", func(t *testing.T) {
		// Arrange: 4-space indent (default) with sorted keys
		pretty := "{\n    \"a\": 1,\n    \"b\": 2\n}\n"
		dir := t.TempDir()
		path := writeTestFile(t, dir, "f.json", []byte(pretty))

		// Act
		got := RunPrettyFormatJSON([]string{path})

		// Assert
		if got != 0 {
			t.Errorf("got %d, want 0 for already-formatted file", got)
		}
	})

	t.Run("not pretty formatted without autofix returns 1 but does not modify", func(t *testing.T) {
		// Arrange
		dir := t.TempDir()
		original := `{"b":1,"a":2}`
		path := writeTestFile(t, dir, "f.json", []byte(original))

		// Act
		got := RunPrettyFormatJSON([]string{path})

		// Assert
		if got != 1 {
			t.Errorf("got %d, want 1", got)
		}
		if string(readTestFile(t, path)) != original {
			t.Errorf("file should not be modified without --autofix")
		}
	})

	t.Run("autofix rewrites file and returns 1", func(t *testing.T) {
		// Arrange
		dir := t.TempDir()
		path := writeTestFile(t, dir, "f.json", []byte(`{"b":1,"a":2}`))

		// Act
		got := RunPrettyFormatJSON([]string{"--autofix", path})

		// Assert
		if got != 1 {
			t.Errorf("got %d, want 1", got)
		}
		content := string(readTestFile(t, path))
		if !strings.Contains(content, "\n") {
			t.Errorf("expected formatted multi-line JSON, got %q", content)
		}
	})
}
