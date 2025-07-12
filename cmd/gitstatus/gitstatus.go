package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/go-git/go-git/v5"
)

func gitstatus(dir string) error {
	r, err := git.PlainOpenWithOptions(dir, &git.PlainOpenOptions{DetectDotGit: true})
	if err != nil {
		return err
	}
	wt, err := r.Worktree()
	if err != nil {
		return err
	}

	root := wt.Filesystem.Root()
	fmt.Printf("root=%s\n", root)
	dirabs, err := filepath.Abs(dir)
	if err != nil {
		return err
	}
	fmt.Printf("abs(dir)=%s\n", dirabs)

	status, err := wt.Status()
	if err != nil {
		return err
	}
	for n, s := range status {
		fullpath := filepath.Join(root, filepath.FromSlash(n))
		relpath, err := filepath.Rel(dirabs, fullpath)
		if err != nil {
			return err
		}
		fmt.Printf("\t%s (%c, %c, %s)\n", filepath.ToSlash(relpath), s.Staging, s.Worktree, s.Extra)
	}
	return nil
}

func gitstatus2(dir string) error {
	status, err := DirStatus(dir)
	if err != nil {
		return err
	}
	for n, s := range status {
		fmt.Printf("\t%s (S:'%c', W:'%c', %s)\n", n, s.Staging, s.Worktree, s.Extra)
	}
	return nil
}

func DirStatus(dir string) (git.Status, error) {
	r, err := git.PlainOpenWithOptions(dir, &git.PlainOpenOptions{DetectDotGit: true})
	if err != nil {
		return nil, err
	}
	wt, err := r.Worktree()
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

func main() {
	dir := "."
	if len(os.Args) >= 2 {
		dir = os.Args[1]
	}
	if err := gitstatus2(dir); err != nil {
		log.Fatal(err)
	}
}
