package templatefs

import (
	"html/template"
	"io/fs"
)

type FS struct {
	fs.FS
	cache map[string]*template.Template
}

func New(fsys fs.FS) *FS {
	return &FS{
		FS:    fsys,
		cache: map[string]*template.Template{},
	}
}

func (fs *FS) Template(name string) (*template.Template, error) {
	// Check cache for a parsed Template
	if t, ok := fs.cache[name]; ok {
		return t, nil
	}
	// Load a file and parse as a Template
	tmpl, err := template.ParseFS(fs, name)
	if err != nil {
		return nil, err
	}
	fs.cache[name] = tmpl
	return tmpl, nil
}
