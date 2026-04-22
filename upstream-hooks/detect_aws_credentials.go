package upstreamhooks

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// RunDetectAWSCredentials is the Go equivalent of detect_aws_credentials.py.
// Detects AWS credentials that match secrets from your local credential files.
// Returns 0 on success, 1 if credentials are found in committed files, 2 if no keys configured.
func RunDetectAWSCredentials(args []string) int {
	fs := flag.NewFlagSet("detect-aws-credentials", flag.ContinueOnError)
	allowMissing := fs.Bool("allow-missing-credentials", false,
		"allow hook to pass when no credentials are detected")
	var credFiles multiStringFlag
	credFiles = []string{"~/.aws/config", "~/.aws/credentials", "/etc/boto.cfg", "~/.boto"}
	fs.Var(&credFiles, "credentials-file",
		"additional credential file path (may be repeated)")
	if err := fs.Parse(args); err != nil {
		return 1
	}

	// Collect credential file paths from environment variables.
	for _, env := range []string{"AWS_CONFIG_FILE", "AWS_CREDENTIAL_FILE", "AWS_SHARED_CREDENTIALS_FILE", "BOTO_CONFIG"} {
		if v := os.Getenv(env); v != "" {
			credFiles = append(credFiles, v)
		}
	}

	// Read all secrets from credential files.
	var secrets [][]byte
	for _, cf := range credFiles {
		secrets = append(secrets, readAWSSecretsFromFile(cf)...)
	}

	// Read secrets from environment variables.
	for _, env := range []string{"AWS_SECRET_ACCESS_KEY", "AWS_SECURITY_TOKEN", "AWS_SESSION_TOKEN"} {
		if v := os.Getenv(env); v != "" {
			secrets = append(secrets, []byte(v))
		}
	}

	if len(secrets) == 0 {
		if *allowMissing {
			return 0
		}
		fmt.Println("No AWS keys were found in the configured credential files and environment variables.")
		fmt.Println("Please ensure you have the correct setting for --credentials-file")
		return 2
	}

	retv := 0
	for _, f := range fs.Args() {
		content, err := os.ReadFile(f)
		if err != nil {
			continue
		}
		for _, secret := range secrets {
			if containsBytes(content, secret) {
				hidden := string(secret)
				if len(hidden) > 4 {
					hidden = hidden[:4] + strings.Repeat("*", len(hidden)-4)
				}
				fmt.Printf("AWS secret found in %s: %s\n", f, hidden)
				retv = 1
				break
			}
		}
	}
	return retv
}

// readAWSSecretsFromFile parses an INI-style AWS credential file and returns secret values.
func readAWSSecretsFromFile(credFile string) [][]byte {
	path := credFile
	if strings.HasPrefix(path, "~/") {
		home, _ := os.UserHomeDir()
		path = filepath.Join(home, path[2:])
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}

	secretKeys := map[string]bool{
		"aws_secret_access_key": true,
		"aws_security_token":    true,
		"aws_session_token":     true,
	}
	var secrets [][]byte
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "#") || strings.HasPrefix(line, "[") {
			continue
		}
		idx := strings.IndexByte(line, '=')
		if idx < 0 {
			continue
		}
		key := strings.ToLower(strings.TrimSpace(line[:idx]))
		val := strings.TrimSpace(line[idx+1:])
		if secretKeys[key] && val != "" {
			secrets = append(secrets, []byte(val))
		}
	}
	return secrets
}

// multiStringFlag is a flag.Value that accumulates repeated string flags.
type multiStringFlag []string

func (m *multiStringFlag) String() string { return strings.Join(*m, ", ") }
func (m *multiStringFlag) Set(v string) error {
	*m = append(*m, v)
	return nil
}
