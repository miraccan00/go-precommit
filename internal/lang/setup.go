package lang

import (
	"crypto/sha256"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
)

// Setup installs the required environment for the given language inside envDir.
// It is idempotent: if envDir already contains a marker file the function
// returns immediately.
//
// repoPath  — local path to the cloned remote repo (used for pip/npm install)
// language  — "python", "golang", "node", "system", "script", ...
// additionalDeps — extra packages to install (e.g. ["types-pyyaml"])
func Setup(language, repoPath, envDir string, additionalDeps []string) error {
	marker := filepath.Join(envDir, ".installed")
	if _, err := os.Stat(marker); err == nil {
		return nil // already done
	}

	if err := os.MkdirAll(envDir, 0o755); err != nil {
		return err
	}

	var setupErr error
	switch strings.ToLower(language) {
	case "python", "python3":
		setupErr = setupPython(repoPath, envDir, additionalDeps)
	case "golang", "go":
		setupErr = setupGolang(repoPath, envDir)
	case "node", "nodejs":
		setupErr = setupNode(repoPath, envDir, additionalDeps)
	case "system", "script", "fail", "":
		// No environment needed.
	default:
		return fmt.Errorf("unsupported language: %q", language)
	}

	if setupErr != nil {
		_ = os.RemoveAll(envDir)
		return setupErr
	}

	// Write the marker so we skip setup next time.
	return os.WriteFile(marker, nil, 0o644)
}

// BinDir returns the directory that should be prepended to PATH when running
// hooks for the given language.  Returns "" for system/script (use PATH as-is).
func BinDir(language, envDir string) string {
	switch strings.ToLower(language) {
	case "python", "python3":
		if runtime.GOOS == "windows" {
			return filepath.Join(envDir, "Scripts")
		}
		return filepath.Join(envDir, "bin")
	case "golang", "go":
		return envDir // GOBIN was set to envDir
	case "node", "nodejs":
		// node_modules live inside envDir (symlinked from the repo).
		return filepath.Join(envDir, "node_modules", ".bin")
	default:
		return ""
	}
}

// EnvCacheKey returns a stable, unique key for (repoURL, rev, additionalDeps).
// Used to derive the env directory path so each unique combination gets its own
// isolated environment.
func EnvCacheKey(repoURL, rev string, additionalDeps []string) string {
	deps := make([]string, len(additionalDeps))
	copy(deps, additionalDeps)
	sort.Strings(deps)
	raw := repoURL + "@" + rev + ":" + strings.Join(deps, ",")
	h := sha256.Sum256([]byte(raw))
	return fmt.Sprintf("%x", h[:12])
}

// ─── language-specific setup ──────────────────────────────────────────────────

func setupPython(repoPath, envDir string, additionalDeps []string) error {
	// Create virtualenv.
	if out, err := exec.Command("python3", "-m", "venv", envDir).CombinedOutput(); err != nil {
		return fmt.Errorf("python3 -m venv: %w\n%s", err, out)
	}

	pip := filepath.Join(envDir, "bin", "pip")
	if runtime.GOOS == "windows" {
		pip = filepath.Join(envDir, "Scripts", "pip.exe")
	}

	// Install the package at repoPath.
	cmd := exec.Command(pip, "install", ".")
	cmd.Dir = repoPath
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("pip install .: %w\n%s", err, out)
	}

	// Install additional dependencies.
	if len(additionalDeps) > 0 {
		args := append([]string{"install"}, additionalDeps...)
		if out, err := exec.Command(pip, args...).CombinedOutput(); err != nil {
			return fmt.Errorf("pip install additional deps: %w\n%s", err, out)
		}
	}

	return nil
}

func setupGolang(repoPath, envDir string) error {
	cmd := exec.Command("go", "install", "./...")
	cmd.Dir = repoPath
	cmd.Env = append(os.Environ(), "GOBIN="+envDir)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("go install: %w\n%s", err, out)
	}
	return nil
}

func setupNode(repoPath, envDir string, additionalDeps []string) error {
	// Symlink node_modules inside envDir so BinDir() works.
	modulesInRepo := filepath.Join(repoPath, "node_modules")

	cmd := exec.Command("npm", "install")
	cmd.Dir = repoPath
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("npm install: %w\n%s", err, out)
	}

	if len(additionalDeps) > 0 {
		args := append([]string{"install"}, additionalDeps...)
		cmd = exec.Command("npm", args...)
		cmd.Dir = repoPath
		if out, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("npm install deps: %w\n%s", err, out)
		}
	}

	// Create a symlink envDir/node_modules → repoPath/node_modules.
	linkTarget := filepath.Join(envDir, "node_modules")
	if _, err := os.Lstat(linkTarget); err != nil {
		if err := os.Symlink(modulesInRepo, linkTarget); err != nil {
			return fmt.Errorf("symlink node_modules: %w", err)
		}
	}

	return nil
}
