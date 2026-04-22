package upstreamhooks

import (
	"strings"
	"testing"
)

func TestSortSimpleYAML(t *testing.T) {
	t.Run("already sorted returns false", func(t *testing.T) {
		// Arrange
		dir := t.TempDir()
		input := "apple: 1\nbanana: 2\ncherry: 3\n"
		path := writeTestFile(t, dir, "f.yaml", []byte(input))

		// Act
		changed := sortSimpleYAML(path)

		// Assert
		if changed {
			t.Errorf("expected no change for already-sorted YAML")
		}
	})

	t.Run("unsorted keys get sorted and returns true", func(t *testing.T) {
		// Arrange
		dir := t.TempDir()
		input := "cherry: 3\napple: 1\nbanana: 2\n"
		path := writeTestFile(t, dir, "f.yaml", []byte(input))

		// Act
		changed := sortSimpleYAML(path)

		// Assert
		if !changed {
			t.Errorf("expected change for unsorted YAML")
		}
		got := string(readTestFile(t, path))
		applePos := strings.Index(got, "apple")
		bananaPos := strings.Index(got, "banana")
		cherryPos := strings.Index(got, "cherry")
		if applePos > bananaPos || bananaPos > cherryPos {
			t.Errorf("keys not in sorted order: got %q", got)
		}
	})

	t.Run("case-insensitive sort order", func(t *testing.T) {
		// Arrange
		dir := t.TempDir()
		input := "Zebra: 1\napple: 2\nMango: 3\n"
		path := writeTestFile(t, dir, "f.yaml", []byte(input))

		// Act
		changed := sortSimpleYAML(path)

		// Assert
		if !changed {
			t.Errorf("expected change")
		}
		got := string(readTestFile(t, path))
		applePos := strings.Index(got, "apple")
		mangoPos := strings.Index(got, "Mango")
		zebraPos := strings.Index(got, "Zebra")
		if applePos > mangoPos || mangoPos > zebraPos {
			t.Errorf("keys not in case-insensitive sorted order: got %q", got)
		}
	})

	t.Run("non-mapping YAML unchanged returns false", func(t *testing.T) {
		// Arrange
		dir := t.TempDir()
		input := "- item1\n- item2\n"
		path := writeTestFile(t, dir, "f.yaml", []byte(input))

		// Act
		changed := sortSimpleYAML(path)

		// Assert
		if changed {
			t.Errorf("sequence YAML should not be modified")
		}
	})

	t.Run("invalid YAML returns false", func(t *testing.T) {
		// Arrange
		dir := t.TempDir()
		path := writeTestFile(t, dir, "bad.yaml", []byte(":\t:invalid"))

		// Act
		changed := sortSimpleYAML(path)

		// Assert
		if changed {
			t.Errorf("invalid YAML should return false")
		}
	})
}

func TestRunSortSimpleYAML(t *testing.T) {
	t.Run("no files returns 0", func(t *testing.T) {
		// Arrange + Act
		got := RunSortSimpleYAML(nil)

		// Assert
		if got != 0 {
			t.Errorf("got %d, want 0", got)
		}
	})

	t.Run("unsorted YAML returns 1", func(t *testing.T) {
		// Arrange
		dir := t.TempDir()
		path := writeTestFile(t, dir, "f.yaml", []byte("z: 1\na: 2\n"))

		// Act
		got := RunSortSimpleYAML([]string{path})

		// Assert
		if got != 1 {
			t.Errorf("got %d, want 1", got)
		}
	})
}
