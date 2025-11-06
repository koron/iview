// Package gitinfo provides git information.
package gitinfo

import (
	"errors"
	"path/filepath"
	"sync"

	"github.com/go-git/go-git/v5"
	"github.com/koron/iview/internal/gitfunc"
	layoutdto "github.com/koron/iview/layout/dto"
	"github.com/koron/iview/plugin"
)

func init() {
	plugin.AddLayoutDocumentFilter(plugin.MediaTypeDirectory, layoutdto.DocumentFilterFunc(gitInfoWrap))
}

type gitInfo struct {
	layoutdto.Document

	gitDirStatus func() (git.Status, error)
}

func gitInfoWrap(base layoutdto.Document) layoutdto.Document {
	gi := &gitInfo{
		Document: base,
	}
	gi.gitDirStatus = sync.OnceValues[git.Status, error](gi.getGitDirStatus)
	return gi
}

func (gi *gitInfo) getGitDirStatus() (git.Status, error) {
	p, err := gi.Filepath()
	if err != nil {
		return nil, err
	}
	s, err := gitfunc.DirStatus(filepath.Clean(p))
	// Ignore git.ErrRepositoryNotExists
	if err != nil && errors.Is(err, git.ErrRepositoryNotExists) {
		return nil, nil
	}
	return s, err
}

type GitStatus = git.FileStatus

func (gi *gitInfo) GitStatus(name string) (*GitStatus, error) {
	stat, err := gi.gitDirStatus()
	if err != nil {
		return nil, err
	}
	return stat[name], nil
}
