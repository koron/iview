package layout

import (
	"html/template"
	"io"
	"io/fs"
	"net/http"
	"strings"
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

//////////////////////////////////////////////////////////////////////////////
// DocBase

type DocBase struct {
	file    DocFile
	rawPath string
	extHead template.HTML
}

var _ Document = (*DocBase)(nil)

type DocFile interface {
	Read([]byte) (int, error)
	Readdir(int) ([]fs.FileInfo, error)
	Stat() (fs.FileInfo, error)
}

type DocOption interface {
	apply(*DocBase)
}

type DocOptionFunc func(*DocBase)

func (f DocOptionFunc) apply(doc *DocBase) { f(doc) }

func DocWithPath(path string) DocOption {
	return DocOptionFunc(func(doc *DocBase) {
		doc.rawPath = path
	})
}

func DocWithExtHead(extHead template.HTML) DocOption {
	return DocOptionFunc(func(doc *DocBase) {
		doc.extHead = extHead
	})
}

func NewDoc(file DocFile, options ...DocOption) Document {
	doc := &DocBase{
		file: file,
	}
	for _, opt := range options {
		opt.apply(doc)
	}
	return doc
}

func (doc *DocBase) Name() (string, error) {
	fi, err := doc.file.Stat()
	if err != nil {
		return "", err
	}
	return fi.Name(), nil
}

func (doc *DocBase) Path() (string, error) {
	fi, err := doc.file.Stat()
	if err != nil {
		return "", err
	}
	if fi.IsDir() {
		return doc.rawPath + "/", nil
	}
	return doc.rawPath, nil
}

func (doc *DocBase) Breadcrumbs() ([]Link, error) {
	dirs := strings.Split(doc.rawPath, "/")
	if len(dirs) < 2 {
		return nil, nil
	}
	if dirs[len(dirs)-1] == "" {
		dirs = dirs[:len(dirs)-1]
	}
	links := append(make([]Link, 0, len(dirs)), Link{Name: "(Root)", Path: "/"})
	for _, d := range dirs[1:] {
		links = append(links, Link{
			Name: d,
			Path: links[len(links)-1].Path + d + "/",
		})
	}
	links[len(links)-1].Path = ""
	return links, nil
}

func (doc *DocBase) Read(b []byte) (int, error) {
	return doc.file.Read(b)
}

func (doc *DocBase) Readdir(count int) ([]fs.FileInfo, error) {
	return doc.file.Readdir(count)
}

func (doc *DocBase) ReadAllString() (string, error) {
	b, err := io.ReadAll(doc)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (doc *DocBase) ExtHead() (template.HTML, error) {
	return doc.extHead, nil
}

//////////////////////////////////////////////////////////////////////////////
// Renderer

type Renderer struct {
	*template.Template

	ExtHead template.HTML
}

func (r *Renderer) Render(w io.Writer, rawPath string, f http.File) error {
	doc := NewDoc(f, DocWithPath(rawPath), DocWithExtHead(r.ExtHead))
	// TODO: apply media type filters.
	return r.Execute(w, doc)
}
