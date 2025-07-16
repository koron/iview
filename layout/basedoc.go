package layout

import (
	"html/template"
	"io"
	"net/http"
	"strings"
)

type BaseDoc struct {
	http.File
	rawPath string
	extHead template.HTML
}

var _ Document = (*BaseDoc)(nil)

func (doc *BaseDoc) Name() (string, error) {
	fi, err := doc.Stat()
	if err != nil {
		return "", err
	}
	return fi.Name(), nil
}

func (doc *BaseDoc) Path() (string, error) {
	fi, err := doc.Stat()
	if err != nil {
		return "", err
	}
	if fi.IsDir() {
		return doc.rawPath + "/", nil
	}
	return doc.rawPath, nil
}

func (doc *BaseDoc) Breadcrumbs() ([]Link, error) {
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

func (doc *BaseDoc) ReadAllString() (string, error) {
	b, err := io.ReadAll(doc)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (doc *BaseDoc) ExtHead() (template.HTML, error) {
	return doc.extHead, nil
}
