package upstreamhooks

import (
	"os"
	"testing"
)

func TestHasShebang(t *testing.T) {
	tests := []struct {
		name    string
		content []byte
		want    bool
	}{
		{
			name:    "file starting with shebang",
			content: []byte("#!/usr/bin/env bash\necho hello\n"),
			want:    true,
		},
		{
			name:    "file starting with python shebang",
			content: []byte("#!/usr/bin/python3\nprint('hi')\n"),
			want:    true,
		},
		{
			name:    "file without shebang",
			content: []byte("echo hello\n"),
			want:    false,
		},
		{
			name:    "empty file",
			content: []byte{},
			want:    false,
		},
		{
			name:    "only one byte",
			content: []byte("#"),
			want:    false,
		},
		{
			name:    "binary file not starting with shebang",
			content: []byte{0x7F, 0x45, 0x4C, 0x46}, // ELF magic
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			dir := t.TempDir()
			path := writeTestFile(t, dir, "script", tt.content)

			// Act
			got := hasShebang(path)

			// Assert
			if got != tt.want {
				t.Errorf("hasShebang(%q) = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

func TestRunCheckExecutablesHaveShebangs(t *testing.T) {
	t.Run("no files returns 0", func(t *testing.T) {
		// Arrange + Act
		got := RunCheckExecutablesHaveShebangs(nil)

		// Assert
		if got != 0 {
			t.Errorf("got %d, want 0", got)
		}
	})

	t.Run("executable with shebang returns 0", func(t *testing.T) {
		// Arrange
		dir := t.TempDir()
		path := writeTestFile(t, dir, "script.sh", []byte("#!/bin/bash\necho hi\n"))
		if err := os.Chmod(path, 0o755); err != nil {
			t.Fatal(err)
		}

		// Act
		got := RunCheckExecutablesHaveShebangs([]string{path})

		// Assert
		if got != 0 {
			t.Errorf("got %d, want 0", got)
		}
	})

	t.Run("executable without shebang returns 1", func(t *testing.T) {
		// Arrange
		dir := t.TempDir()
		path := writeTestFile(t, dir, "script.sh", []byte("echo hi\n"))
		if err := os.Chmod(path, 0o755); err != nil {
			t.Fatal(err)
		}

		// Act
		got := RunCheckExecutablesHaveShebangs([]string{path})

		// Assert
		if got != 1 {
			t.Errorf("got %d, want 1", got)
		}
	})

	t.Run("non-executable file without shebang returns 0", func(t *testing.T) {
		// Arrange
		dir := t.TempDir()
		path := writeTestFile(t, dir, "readme.txt", []byte("just a readme\n"))
		// mode 0o644 — not executable (default from writeTestFile)

		// Act
		got := RunCheckExecutablesHaveShebangs([]string{path})

		// Assert
		if got != 0 {
			t.Errorf("got %d, want 0 for non-executable", got)
		}
	})
}
