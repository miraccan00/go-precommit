package upstreamhooks

import (
	"testing"
)

func TestContainsBytes(t *testing.T) {
	tests := []struct {
		name     string
		haystack []byte
		needle   []byte
		want     bool
	}{
		{
			name:     "needle found at start",
			haystack: []byte("BEGIN RSA PRIVATE KEY end"),
			needle:   []byte("BEGIN RSA PRIVATE KEY"),
			want:     true,
		},
		{
			name:     "needle found in middle",
			haystack: []byte("some text BEGIN RSA PRIVATE KEY more text"),
			needle:   []byte("BEGIN RSA PRIVATE KEY"),
			want:     true,
		},
		{
			name:     "needle not found",
			haystack: []byte("no private keys here"),
			needle:   []byte("BEGIN RSA PRIVATE KEY"),
			want:     false,
		},
		{
			name:     "empty needle always matches",
			haystack: []byte("anything"),
			needle:   []byte{},
			want:     true,
		},
		{
			name:     "needle longer than haystack",
			haystack: []byte("short"),
			needle:   []byte("much longer needle"),
			want:     false,
		},
		{
			name:     "empty haystack with non-empty needle",
			haystack: []byte{},
			needle:   []byte("key"),
			want:     false,
		},
		{
			name:     "exact match",
			haystack: []byte("key"),
			needle:   []byte("key"),
			want:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange: inputs prepared above

			// Act
			got := containsBytes(tt.haystack, tt.needle)

			// Assert
			if got != tt.want {
				t.Errorf("containsBytes(%q, %q) = %v, want %v",
					tt.haystack, tt.needle, got, tt.want)
			}
		})
	}
}

func TestRunDetectPrivateKey(t *testing.T) {
	t.Run("no files returns 0", func(t *testing.T) {
		// Arrange + Act
		got := RunDetectPrivateKey(nil)

		// Assert
		if got != 0 {
			t.Errorf("got %d, want 0", got)
		}
	})

	t.Run("file with RSA private key header returns 1", func(t *testing.T) {
		// Arrange
		dir := t.TempDir()
		path := writeTestFile(t, dir, "key.pem", []byte("-----BEGIN RSA PRIVATE KEY-----\nMIIE...\n-----END RSA PRIVATE KEY-----\n"))

		// Act
		got := RunDetectPrivateKey([]string{path})

		// Assert
		if got != 1 {
			t.Errorf("got %d, want 1", got)
		}
	})

	t.Run("file with EC private key header returns 1", func(t *testing.T) {
		// Arrange
		dir := t.TempDir()
		path := writeTestFile(t, dir, "key.pem", []byte("-----BEGIN EC PRIVATE KEY-----\n"))

		// Act
		got := RunDetectPrivateKey([]string{path})

		// Assert
		if got != 1 {
			t.Errorf("got %d, want 1", got)
		}
	})

	t.Run("file with OPENSSH private key header returns 1", func(t *testing.T) {
		// Arrange
		dir := t.TempDir()
		path := writeTestFile(t, dir, "key", []byte("-----BEGIN OPENSSH PRIVATE KEY-----\n"))

		// Act
		got := RunDetectPrivateKey([]string{path})

		// Assert
		if got != 1 {
			t.Errorf("got %d, want 1", got)
		}
	})

	t.Run("file with PGP private key block returns 1", func(t *testing.T) {
		// Arrange
		dir := t.TempDir()
		path := writeTestFile(t, dir, "key.asc", []byte("-----BEGIN PGP PRIVATE KEY BLOCK-----\n"))

		// Act
		got := RunDetectPrivateKey([]string{path})

		// Assert
		if got != 1 {
			t.Errorf("got %d, want 1", got)
		}
	})

	t.Run("file without private key returns 0", func(t *testing.T) {
		// Arrange
		dir := t.TempDir()
		path := writeTestFile(t, dir, "safe.txt", []byte("This is a normal file\nNo keys here\n"))

		// Act
		got := RunDetectPrivateKey([]string{path})

		// Assert
		if got != 0 {
			t.Errorf("got %d, want 0", got)
		}
	})

	t.Run("public key file does not trigger", func(t *testing.T) {
		// Arrange
		dir := t.TempDir()
		path := writeTestFile(t, dir, "pub.pem", []byte("-----BEGIN PUBLIC KEY-----\nMIIB...\n-----END PUBLIC KEY-----\n"))

		// Act
		got := RunDetectPrivateKey([]string{path})

		// Assert
		if got != 0 {
			t.Errorf("got %d, want 0 for public key", got)
		}
	})
}
