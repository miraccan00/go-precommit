package repo

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

// Manager clones remote repos and caches them under a platform cache dir.
// Each (url, rev) pair lives in its own subdirectory so multiple revisions
// of the same repo can coexist.
type Manager struct {
	cacheDir string
}

// NewManager creates a Manager, ensuring the cache directory exists.
func NewManager() (*Manager, error) {
	base, err := os.UserCacheDir() // ~/Library/Caches on macOS, ~/.cache on Linux
	if err != nil {
		return nil, fmt.Errorf("cache dir: %w", err)
	}
	cacheDir := filepath.Join(base, "go-precommit", "repos")
	if err := os.MkdirAll(cacheDir, 0o755); err != nil {
		return nil, fmt.Errorf("creating cache dir: %w", err)
	}
	return &Manager{cacheDir: cacheDir}, nil
}

// LocalPath returns the filesystem path to a cloned repo at the given rev.
// On first call it clones; subsequent calls return the cached path.
func (m *Manager) LocalPath(repoURL, rev string) (string, error) {
	dir := filepath.Join(m.cacheDir, dirName(repoURL, rev))

	if _, err := os.Stat(dir); err == nil {
		return dir, nil // cache hit
	}

	fmt.Fprintf(os.Stderr, "[go-precommit] cloning %s @ %s ...\n", repoURL, rev)

	// 1. Try shallow clone via tag reference (fast path for vX.Y.Z).
	_, err := gogit.PlainClone(dir, false, &gogit.CloneOptions{
		URL:           repoURL,
		ReferenceName: plumbing.NewTagReferenceName(rev),
		SingleBranch:  true,
		Depth:         1,
	})
	if err == nil {
		return dir, nil
	}
	_ = os.RemoveAll(dir)

	// 2. Try shallow clone via branch reference.
	_, err = gogit.PlainClone(dir, false, &gogit.CloneOptions{
		URL:           repoURL,
		ReferenceName: plumbing.NewBranchReferenceName(rev),
		SingleBranch:  true,
		Depth:         1,
	})
	if err == nil {
		return dir, nil
	}
	_ = os.RemoveAll(dir)

	// 3. Full clone then checkout by commit hash.
	r, err := gogit.PlainClone(dir, false, &gogit.CloneOptions{
		URL: repoURL,
	})
	if err != nil {
		_ = os.RemoveAll(dir)
		return "", fmt.Errorf("cloning %s: %w", repoURL, err)
	}

	w, err := r.Worktree()
	if err != nil {
		return dir, nil // best effort — checkout failed but repo is there
	}
	_ = w.Checkout(&gogit.CheckoutOptions{
		Hash: plumbing.NewHash(rev),
	})

	return dir, nil
}

// dirName returns a stable, filesystem-safe directory name for (url, rev).
func dirName(repoURL, rev string) string {
	h := sha256.Sum256([]byte(repoURL))
	return fmt.Sprintf("%x-%s", h[:8], sanitize(rev))
}

var unsafeChars = regexp.MustCompile(`[^a-zA-Z0-9._-]`)

func sanitize(s string) string {
	return unsafeChars.ReplaceAllString(s, "_")
}
