// Package gitfunc provides wrappers for the functions of go-git that iview uses.
package gitfunc

import (
	"path/filepath"
	"strings"

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

	root, err := filepath.Abs(wt.Filesystem.Root())
	if err != nil {
		return nil, err
	}
	absdir, err := filepath.Abs(dir)
	if err != nil {
		return nil, err
	}
	absdir += string(filepath.Separator)

	dirStatus := git.Status{}
	for n, s := range status {
		fullpath := filepath.Join(root, filepath.FromSlash(n))
		if !strings.HasPrefix(fullpath, absdir) {
			continue
		}
		relpath, err := filepath.Rel(absdir, fullpath)
		if err != nil {
			return nil, err
		}
		components := strings.Split(relpath, string(filepath.Separator))
		if len(components) == 0 {
			continue
		}
		if s.Staging == git.Untracked && s.Worktree == git.Untracked {
			s.Staging = git.Unmodified
		}
		dirStatus[components[0]] = mergeFileStatus(dirStatus[components[0]], s)
	}
	return dirStatus, nil
}

var statusCodeWeights = map[git.StatusCode]int{
	git.Unmodified:         1,
	git.Copied:             2,
	git.Renamed:            3,
	git.UpdatedButUnmerged: 4,
	git.Modified:           5,
	git.Added:              6,
	git.Deleted:            7,
	git.Untracked:          8,
}

func mergeStatusCode(a, b git.StatusCode) git.StatusCode {
	if statusCodeWeights[a] > statusCodeWeights[b] {
		return a
	}
	return b
}

func mergeFileStatus(a, b *git.FileStatus) *git.FileStatus {
	if a == nil {
		return b
	}
	a.Staging = mergeStatusCode(a.Staging, b.Staging)
	a.Worktree = mergeStatusCode(a.Worktree, b.Worktree)
	if a.Extra == "" {
		a.Extra = b.Extra
	}
	return a
}
