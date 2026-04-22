package upstreamhooks

import (
	"bytes"
	"io"
	"strings"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/storage/filesystem"
)

func bytesReader(b []byte) io.Reader { return bytes.NewReader(b) }

// openRepo opens the nearest git repository, walking up from the current directory.
func openRepo() (*git.Repository, error) {
	return git.PlainOpenWithOptions(".", &git.PlainOpenOptions{DetectDotGit: true})
}

// repoGitDir returns the path to the .git directory (e.g. "/proj/.git").
// Returns "" when the storer is not filesystem-backed (e.g. in-memory repos in tests).
func repoGitDir(repo *git.Repository) string {
	if fs, ok := repo.Storer.(*filesystem.Storage); ok {
		return fs.Filesystem().Root()
	}
	return ""
}

// addedFiles returns the set of files staged for addition (not present in HEAD).
func addedFiles() (map[string]bool, error) {
	repo, err := openRepo()
	if err != nil {
		return nil, err
	}
	idx, err := repo.Storer.Index()
	if err != nil {
		return nil, err
	}

	// Build a set of every file in the HEAD tree so we can detect what is new.
	headFiles := make(map[string]struct{})
	if ref, err := repo.Head(); err == nil {
		if commit, err := repo.CommitObject(ref.Hash()); err == nil {
			if tree, err := commit.Tree(); err == nil {
				_ = tree.Files().ForEach(func(f *object.File) error {
					headFiles[f.Name] = struct{}{}
					return nil
				})
			}
		}
	}
	// When HEAD doesn't exist (initial commit) headFiles stays empty, so all
	// index entries are treated as added — which is correct.

	set := make(map[string]bool)
	for _, entry := range idx.Entries {
		if _, inHEAD := headFiles[entry.Name]; !inHEAD {
			set[entry.Name] = true
		}
	}
	return set, nil
}

// trackedFiles returns all files currently in the index (equivalent to git ls-files).
func trackedFiles() ([]string, error) {
	repo, err := openRepo()
	if err != nil {
		return nil, err
	}
	idx, err := repo.Storer.Index()
	if err != nil {
		return nil, err
	}
	files := make([]string, 0, len(idx.Entries))
	for _, entry := range idx.Entries {
		files = append(files, entry.Name)
	}
	return files, nil
}

func splitLines(s string) []string {
	s = strings.TrimRight(s, "\n\r")
	if s == "" {
		return nil
	}
	return strings.Split(s, "\n")
}

func zsplit(s string) []string {
	s = strings.Trim(s, "\x00")
	if s == "" {
		return nil
	}
	return strings.Split(s, "\x00")
}
