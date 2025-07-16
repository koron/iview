package layout

import (
	"io"
	"net/http"
	"strings"
)

type Data struct {
	http.File

	RawPath          string
	LayoutExtensions *LayoutExtensions
}

var _ Document = (*Data)(nil)

type LayoutExtensions struct {
	Head any
}

func (f *Data) Name() (string, error) {
	fi, err := f.Stat()
	if err != nil {
		return "", err
	}
	return fi.Name(), nil
}

func (f *Data) Path() (string, error) {
	fi, err := f.Stat()
	if err != nil {
		return "", err
	}
	if fi.IsDir() {
		return f.RawPath + "/", nil
	}
	return f.RawPath, nil
}

func (f *Data) ReadAllString() (string, error) {
	b, err := io.ReadAll(f)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (f *Data) Breadcrumbs() ([]Link, error) {
	dirs := strings.Split(f.RawPath, "/")
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

func (f *Data) Content() (any, error) {
	b, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}
	return string(b), nil
}
