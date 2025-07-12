// Package gitfunc provides wrappers for the functions of go-git that iview uses.
package gitfunc

import (
	"path/filepath"

	"github.com/go-git/go-git/v5"
)

// Worktree returns the *git.Worktree of the specified directory if it is under git control.
func Worktree(dir string) (*git.Worktree, error) {
	r, err := git.PlainOpenWithOptions(dir, &git.PlainOpenOptions{DetectDotGit: true})
	if err != nil {
		return nil, err
	}
	return r.Worktree()
}

// DirStatus returns git.Status limited to the specified directory.
// If the specified directory is not under git control, it returns git.ErrRepositoryNotExists.
func DirStatus(dir string) (git.Status, error) {
	wt, err := Worktree(dir)
	if err != nil {
		return nil, err
	}
	status, err := wt.Status()
	if err != nil {
		return nil, err
	}

	root := wt.Filesystem.Root()
	absdir, err := filepath.Abs(dir)
	if err != nil {
		return nil, err
	}

	dirStatus := git.Status{}
	for n, s := range status {
		fullpath := filepath.Join(root, filepath.FromSlash(n))
		if filepath.Dir(fullpath) != absdir {
			continue
		}
		relpath, err := filepath.Rel(absdir, fullpath)
		if err != nil {
			return nil, err
		}
		dirStatus[relpath] = s
	}
	return dirStatus, nil
}
