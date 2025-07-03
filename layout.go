package main

import (
	"html/template"
	"io"
	"net/http"
	"strings"
)

type Link struct {
	Name string
	Path string
}

type TemplateData struct {
	http.File
	path      string
	LayoutExt *LayoutExtensions
}

type LayoutExtensions struct {
	Head any
}

type templateRenderer struct {
	*template.Template
	ext *LayoutExtensions
}

func (r *templateRenderer) Render(w io.Writer, upath string, f http.File) error {
	return r.Execute(w, &TemplateData{
		File:      f,
		path:      upath,
		LayoutExt: r.ext,
	})
}

func (f *TemplateData) Name() (any, error) {
	fi, err := f.Stat()
	if err != nil {
		return nil, err
	}
	return fi.Name(), nil
}

func (f *TemplateData) Content() (any, error) {
	b, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}
	return string(b), nil
}

func (f *TemplateData) ContentBytes(nbytes int) (any, error) {
	if nbytes < 0 {
		return io.ReadAll(f)
	}
	b := make([]byte, nbytes)
	n, err := f.Read(b)
	if err != nil {
		return nil, err
	}
	return b[:n], nil
}

func (f *TemplateData) Breadcrumbs() ([]Link, error) {
	dirs := strings.Split(f.path, "/")
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
