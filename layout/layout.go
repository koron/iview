package layout

import (
	"html/template"
	"io/fs"
)

type Document interface {
	Name() (string, error)
	Path() (string, error)
	Breadcrumbs() ([]Link, error)

	Read([]byte) (int, error)
	Readdir(count int) ([]fs.FileInfo, error)
	ReadAllString() (string, error)

	ExtHead() (template.HTML, error)
}

type Link struct {
	Name string
	Path string
}

type LayoutExtensions struct {
	Head any
}
