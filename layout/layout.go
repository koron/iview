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
	http.File
	rawPath string
	extHead template.HTML
}

var _ Document = (*DocBase)(nil)

func NewDoc() Document {
	// TODO:
	return &DocBase{}
}

func (doc *DocBase) Name() (string, error) {
	fi, err := doc.Stat()
	if err != nil {
		return "", err
	}
	return fi.Name(), nil
}

func (doc *DocBase) Path() (string, error) {
	fi, err := doc.Stat()
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
	doc := &DocBase{
		File:    f,
		rawPath: rawPath,
		extHead: r.ExtHead,
	}
	// TODO:
	return r.Execute(w, doc)
}
