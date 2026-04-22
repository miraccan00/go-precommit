package upstreamhooks

import (
	"testing"
)

func TestRunCheckXML(t *testing.T) {
	t.Run("no files returns 0", func(t *testing.T) {
		// Arrange + Act
		got := RunCheckXML(nil)

		// Assert
		if got != 0 {
			t.Errorf("got %d, want 0", got)
		}
	})

	t.Run("valid XML returns 0", func(t *testing.T) {
		// Arrange
		dir := t.TempDir()
		path := writeTestFile(t, dir, "f.xml", []byte(`<root><child key="val">text</child></root>`))

		// Act
		got := RunCheckXML([]string{path})

		// Assert
		if got != 0 {
			t.Errorf("got %d, want 0", got)
		}
	})

	t.Run("unclosed tag returns 1", func(t *testing.T) {
		// Arrange
		dir := t.TempDir()
		path := writeTestFile(t, dir, "bad.xml", []byte(`<root><unclosed></root>`))

		// Act
		got := RunCheckXML([]string{path})

		// Assert
		if got != 1 {
			t.Errorf("got %d, want 1", got)
		}
	})

	t.Run("mismatched tags returns 1", func(t *testing.T) {
		// Arrange
		dir := t.TempDir()
		path := writeTestFile(t, dir, "bad.xml", []byte(`<root></other>`))

		// Act
		got := RunCheckXML([]string{path})

		// Assert
		if got != 1 {
			t.Errorf("got %d, want 1", got)
		}
	})

	t.Run("empty XML document returns 0", func(t *testing.T) {
		// Arrange
		dir := t.TempDir()
		path := writeTestFile(t, dir, "empty.xml", []byte(``))

		// Act
		got := RunCheckXML([]string{path})

		// Assert
		if got != 0 {
			t.Errorf("got %d, want 0", got)
		}
	})

	t.Run("XML with declaration returns 0", func(t *testing.T) {
		// Arrange
		dir := t.TempDir()
		path := writeTestFile(t, dir, "f.xml", []byte(`<?xml version="1.0"?><root/>`))

		// Act
		got := RunCheckXML([]string{path})

		// Assert
		if got != 0 {
			t.Errorf("got %d, want 0", got)
		}
	})
}
