package layout

import (
	"io/fs"
)

type Document interface {
	Name() (string, error)
	Path() (string, error)
	Breadcrumbs() ([]Link, error)

	Read([]byte) (int, error)
	Readdir(count int) ([]fs.FileInfo, error)
	ReadAllString() (string, error)
}

type Link struct {
	Name string
	Path string
}
