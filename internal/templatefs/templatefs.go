package templatefs

import (
	"html/template"
	"io"
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

type Option interface {
	apply(*template.Template) (*template.Template, error)
}

type OptionFunc func(*template.Template) (*template.Template, error)

func (fn OptionFunc) apply(tmpl *template.Template) (*template.Template, error) {
	return fn(tmpl)
}

var _ Option = (OptionFunc)(nil)

type options []Option

func (opts options) apply(tmpl *template.Template) (*template.Template, error) {
	var err error
	for _, opt := range opts {
		tmpl, err = opt.apply(tmpl)
		if err != nil {
			return nil, err
		}
	}
	return tmpl, nil
}

var _ Option = (options)(nil)

func (fs *FS) Template(name string, opts ...Option) (*template.Template, error) {
	// Check cache for a parsed Template
	if t, ok := fs.cache[name]; ok {
		return t, nil
	}

	var loader = OptionFunc(func(tmpl *template.Template) (*template.Template, error) {
		// Load a file and parse as a Template
		f, err := fs.Open(name)
		if err != nil {
			return nil, err
		}
		b, err := io.ReadAll(f)
		f.Close()
		if err != nil {
			return nil, err
		}
		return tmpl.Parse(string(b))
	})

	// Apply options
	tmpl, err := options(append(opts, loader)).apply(template.New(name))
	if err != nil {
		return nil, err
	}

	fs.cache[name] = tmpl
	return tmpl, nil
}
