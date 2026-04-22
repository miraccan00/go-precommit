package upstreamhooks

import (
	"flag"
	"fmt"
	"os"
)

// RunDetectPrivateKey is the Go equivalent of detect_private_key.py.
// Detects the presence of private keys in committed files.
// Returns 0 on success, 1 if a private key is found.
func RunDetectPrivateKey(args []string) int {
	fs := flag.NewFlagSet("detect-private-key", flag.ContinueOnError)
	if err := fs.Parse(args); err != nil {
		return 1
	}

	blacklist := [][]byte{
		[]byte("BEGIN RSA PRIVATE KEY"),
		[]byte("BEGIN DSA PRIVATE KEY"),
		[]byte("BEGIN EC PRIVATE KEY"),
		[]byte("BEGIN OPENSSH PRIVATE KEY"),
		[]byte("BEGIN PRIVATE KEY"),
		[]byte("PuTTY-User-Key-File-2"),
		[]byte("BEGIN SSH2 ENCRYPTED PRIVATE KEY"),
		[]byte("BEGIN PGP PRIVATE KEY BLOCK"),
		[]byte("BEGIN ENCRYPTED PRIVATE KEY"),
		[]byte("BEGIN OpenVPN Static key V1"),
	}

	retv := 0
	for _, f := range fs.Args() {
		content, err := os.ReadFile(f)
		if err != nil {
			continue
		}
		for _, pattern := range blacklist {
			if containsBytes(content, pattern) {
				fmt.Printf("Private key found: %s\n", f)
				retv = 1
				break
			}
		}
	}
	return retv
}

func containsBytes(haystack, needle []byte) bool {
	if len(needle) == 0 {
		return true
	}
	for i := 0; i <= len(haystack)-len(needle); i++ {
		if haystack[i] == needle[0] {
			match := true
			for j := 1; j < len(needle); j++ {
				if haystack[i+j] != needle[j] {
					match = false
					break
				}
			}
			if match {
				return true
			}
		}
	}
	return false
}
