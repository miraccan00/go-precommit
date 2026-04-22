package upstreamhooks

import (
	"flag"
	"fmt"
	"io"
	"strings"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/filemode"
	"github.com/go-git/go-git/v5/plumbing/object"
)

// RunDestroyedSymlinks is the Go equivalent of destroyed_symlinks.py.
// Detects symlinks that have been changed to regular files containing the
// symlink path — a common accident when core.symlinks=false.
// Returns 0 on success, 1 if destroyed symlinks are detected.
func RunDestroyedSymlinks(args []string) int {
	fs := flag.NewFlagSet("destroyed-symlinks", flag.ContinueOnError)
	if err := fs.Parse(args); err != nil {
		return 1
	}
	files := fs.Args()
	if len(files) == 0 {
		return 0
	}

	destroyed := findDestroyedSymlinks(files)
	if len(destroyed) == 0 {
		return 0
	}

	fmt.Println("Destroyed symlinks:")
	for _, d := range destroyed {
		fmt.Printf("- %s\n", d)
	}
	fmt.Println("You should unstage affected files:")
	fmt.Printf("\tgit reset HEAD -- %s\n", strings.Join(destroyed, " "))
	fmt.Println("And retry commit. As a long term solution you may try to explicitly")
	fmt.Println("tell git that your environment does not support symlinks:")
	fmt.Println("\tgit config core.symlinks false")
	return 1
}

func findDestroyedSymlinks(files []string) []string {
	repo, err := openRepo()
	if err != nil {
		return nil
	}

	idx, err := repo.Storer.Index()
	if err != nil {
		return nil
	}

	// Build a fast name→entry map from the index.
	indexByName := make(map[string]struct {
		mode filemode.FileMode
		hash plumbing.Hash
	}, len(idx.Entries))
	for _, e := range idx.Entries {
		indexByName[e.Name] = struct {
			mode filemode.FileMode
			hash plumbing.Hash
		}{e.Mode, e.Hash}
	}

	// HEAD tree — nil on very first commit.
	var headTree *object.Tree
	if ref, err := repo.Head(); err == nil {
		if commit, err := repo.CommitObject(ref.Hash()); err == nil {
			headTree, _ = commit.Tree()
		}
	}

	var destroyed []string
	for _, path := range files {
		idxEntry, inIndex := indexByName[path]
		if !inIndex {
			continue
		}
		if headTree == nil {
			continue
		}
		headFile, err := headTree.File(path)
		if err != nil {
			// File not in HEAD — no symlink to destroy.
			continue
		}

		modeHEAD := headFile.Mode
		modeIndex := idxEntry.mode
		hashHEAD := headFile.Hash
		hashIndex := idxEntry.hash

		if modeHEAD != filemode.Symlink || modeIndex == filemode.Symlink || modeIndex == filemode.Empty {
			continue
		}

		if hashHEAD == hashIndex {
			destroyed = append(destroyed, path)
			continue
		}

		// Hashes differ: compare sizes and content to handle hook-modified content.
		sizeIndex := blobSize(repo, hashIndex)
		sizeHEAD := blobSize(repo, hashHEAD)
		if sizeIndex <= sizeHEAD+2 {
			headContent := blobContent(repo, hashHEAD)
			indexContent := blobContent(repo, hashIndex)
			if trimBytes(headContent) == trimBytes(indexContent) {
				destroyed = append(destroyed, path)
			}
		}
	}
	return destroyed
}

func blobSize(repo *git.Repository, hash plumbing.Hash) int64 {
	blob, err := repo.BlobObject(hash)
	if err != nil {
		return -1
	}
	return blob.Size
}

func blobContent(repo *git.Repository, hash plumbing.Hash) string {
	blob, err := repo.BlobObject(hash)
	if err != nil {
		return ""
	}
	r, err := blob.Reader()
	if err != nil {
		return ""
	}
	data, err := io.ReadAll(r)
	_ = r.Close()
	if err != nil {
		return ""
	}
	return string(data)
}

func trimBytes(s string) string {
	return strings.TrimRight(s, "\r\n")
}
