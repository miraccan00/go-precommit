package upstreamhooks

import (
	"testing"
)

func TestReadAWSSecretsFromFile(t *testing.T) {
	t.Run("reads aws_secret_access_key from credential file", func(t *testing.T) {
		// Arrange
		dir := t.TempDir()
		path := writeTestFile(t, dir, "credentials", []byte(
			"[default]\naws_access_key_id = AKIAIOSFODNN7EXAMPLE\naws_secret_access_key = wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY\n",
		))

		// Act
		secrets := readAWSSecretsFromFile(path)

		// Assert
		if len(secrets) != 1 {
			t.Fatalf("got %d secrets, want 1", len(secrets))
		}
		if string(secrets[0]) != "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY" {
			t.Errorf("secret = %q, want key value", secrets[0])
		}
	})

	t.Run("reads multiple secret types", func(t *testing.T) {
		// Arrange
		dir := t.TempDir()
		path := writeTestFile(t, dir, "credentials", []byte(
			"[default]\naws_secret_access_key = secret1\naws_session_token = token1\naws_security_token = sectoken1\n",
		))

		// Act
		secrets := readAWSSecretsFromFile(path)

		// Assert
		if len(secrets) != 3 {
			t.Fatalf("got %d secrets, want 3", len(secrets))
		}
	})

	t.Run("comment lines and section headers skipped", func(t *testing.T) {
		// Arrange
		dir := t.TempDir()
		path := writeTestFile(t, dir, "credentials", []byte(
			"# this is a comment\n[profile1]\n# aws_secret_access_key = notasecret\naws_access_key_id = AKID\n",
		))

		// Act
		secrets := readAWSSecretsFromFile(path)

		// Assert
		if len(secrets) != 0 {
			t.Errorf("got %d secrets, want 0 (commented key should be ignored)", len(secrets))
		}
	})

	t.Run("non-existent file returns nil", func(t *testing.T) {
		// Arrange
		path := "/tmp/does-not-exist-aws-creds-test"

		// Act
		secrets := readAWSSecretsFromFile(path)

		// Assert
		if secrets != nil {
			t.Errorf("expected nil for missing file, got %v", secrets)
		}
	})

	t.Run("empty value skipped", func(t *testing.T) {
		// Arrange
		dir := t.TempDir()
		path := writeTestFile(t, dir, "credentials", []byte("aws_secret_access_key = \n"))

		// Act
		secrets := readAWSSecretsFromFile(path)

		// Assert
		if len(secrets) != 0 {
			t.Errorf("empty value should not be collected, got %v", secrets)
		}
	})
}

func TestRunDetectAWSCredentials(t *testing.T) {
	t.Run("allow-missing-credentials with no cred files returns 0", func(t *testing.T) {
		// Arrange — point to a non-existent credentials file so secrets list is empty
		dir := t.TempDir()
		nonexistent := dir + "/no-such-file"

		// Act
		got := RunDetectAWSCredentials([]string{
			"--allow-missing-credentials",
			"--credentials-file", nonexistent,
		})

		// Assert
		if got != 0 {
			t.Errorf("got %d, want 0 with --allow-missing-credentials", got)
		}
	})

	t.Run("no cred files without allow-missing returns 2", func(t *testing.T) {
		// Arrange — point to non-existent cred files so no secrets are loaded
		dir := t.TempDir()
		nonexistent := dir + "/no-such-file"

		// Act
		got := RunDetectAWSCredentials([]string{
			"--credentials-file", nonexistent,
		})

		// Assert
		if got != 2 {
			t.Errorf("got %d, want 2 when no credentials configured", got)
		}
	})

	t.Run("secret found in committed file returns 1", func(t *testing.T) {
		// Arrange — create a credentials file and a source file containing the secret
		dir := t.TempDir()
		credsPath := writeTestFile(t, dir, "credentials",
			[]byte("[default]\naws_secret_access_key = SUPERSECRETKEY123\n"))
		sourcePath := writeTestFile(t, dir, "config.py",
			[]byte(`AWS_SECRET = "SUPERSECRETKEY123"\n`))

		// Act
		got := RunDetectAWSCredentials([]string{
			"--credentials-file", credsPath,
			sourcePath,
		})

		// Assert
		if got != 1 {
			t.Errorf("got %d, want 1 when secret found in file", got)
		}
	})

	t.Run("secret not in any committed file returns 0", func(t *testing.T) {
		// Arrange
		dir := t.TempDir()
		credsPath := writeTestFile(t, dir, "credentials",
			[]byte("[default]\naws_secret_access_key = SUPERSECRETKEY123\n"))
		cleanPath := writeTestFile(t, dir, "clean.py",
			[]byte("# no secrets here\n"))

		// Act
		got := RunDetectAWSCredentials([]string{
			"--credentials-file", credsPath,
			cleanPath,
		})

		// Assert
		if got != 0 {
			t.Errorf("got %d, want 0 when secret not present", got)
		}
	})
}
