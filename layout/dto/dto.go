// Package dto provides Data Transfer Object for layout system.
package dto

import (
	"html/template"
	"io/fs"
)

type Document interface {
	// Name returns file name.
	Name() (string, error)
	// Path returns a physical source file path.
	Path() (string, error)
	// Filepath returns a physical source file path.
	Filepath() (string, error)

	// Breadcrumbs returns links of breadcrumbs navigator.
	Breadcrumbs() ([]Link, error)

	Read([]byte) (int, error)
	Readdir(count int) ([]fs.FileInfo, error)
	ReadAllString() (string, error)

	IsHighlighted() bool
	HighlightName() string
	HightlightCSS() (template.CSS, error)
	HightlightedHTML() (template.HTML, error)

	ExtHead() (template.HTML, error)
}

type DocumentFilter interface {
	Apply(doc Document) Document
}

type DocumentFilterFunc func(doc Document) Document

func (f DocumentFilterFunc) Apply(doc Document) Document {
	return f(doc)
}

type Link struct {
	Name string
	Path string
}
