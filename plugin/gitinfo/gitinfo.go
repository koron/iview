package gitinfo

import (
	"path/filepath"
	"strings"
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
	p, err := gi.Path()
	if err != nil {
		return nil, err
	}
	p = filepath.FromSlash(strings.Trim(p, "/"))
	if p == "" {
		p = "."
	}
	return gitfunc.DirStatus(p)
}

type GitStatus = git.FileStatus

func (gi *gitInfo) GitStatus(name string) (*GitStatus, error) {
	stat, err := gi.gitDirStatus()
	if err != nil {
		return nil, err
	}
	fstat, _ := stat[name]
	return fstat, nil
}
