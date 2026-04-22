package upstreamhooks

import (
	"os"
	"path/filepath"
	"testing"
)

// writeTestFile creates a file in dir with the given content and returns its path.
func writeTestFile(t *testing.T, dir, name string, content []byte) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatalf("writeTestFile: %v", err)
	}
	return path
}

// readTestFile reads a file and returns its content.
func readTestFile(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("readTestFile: %v", err)
	}
	return data
}
