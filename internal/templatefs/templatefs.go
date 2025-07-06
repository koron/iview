/*
Package templatefs provides Template set.
This template set is linked to fs.FS and returns a template parsed from the specified file in fs.FS.
*/
package templatefs

import (
	"html/template"
	"io/fs"
	"time"
)

type FS struct {
	fs.FS
	cache map[string]*cacheEntry
}

type cacheEntry struct {
	*template.Template
	ModTime time.Time
}

func New(fsys fs.FS) *FS {
	return &FS{
		FS:    fsys,
		cache: map[string]*cacheEntry{},
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

func (fs *FS) modTime(name string) (time.Time, error) {
	f, err := fs.Open(name)
	if err != nil {
		return time.Time{}, err
	}
	defer f.Close()
	fi, err := f.Stat()
	if err != nil {
		return time.Time{}, err
	}
	return fi.ModTime(), nil
}

func (fs *FS) latestModTime(names ...string) (time.Time, error) {
	var latest time.Time
	for _, name := range names {
		ti, err := fs.modTime(name)
		if err != nil {
			return time.Time{}, err
		}
		if ti.After(latest) {
			latest = ti
		}
	}
	return latest, nil
}

func (fs *FS) Template2(layoutName, contentName string, opts ...Option) (*template.Template, error) {
	latest, err := fs.latestModTime(layoutName, contentName)
	if err != nil {
		return nil, err
	}

	// Check cache for a parsed Template
	cacheName := layoutName + ":" + contentName
	if entry, ok := fs.cache[cacheName]; ok && !latest.After(entry.ModTime) {
		return entry.Template, nil
	}

	// Create a new template and apply options
	layout, err := options(opts).apply(template.New(layoutName))
	if err != nil {
		return nil, err
	}

	// Parse template files.
	_, err = layout.ParseFS(fs, layoutName, contentName)
	if err != nil {
		return nil, err
	}

	fs.cache[cacheName] = &cacheEntry{
		Template: layout,
		ModTime:  latest,
	}
	return layout, nil
}
