package main

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"github.com/koron/iview/internal/templatefs"
	"github.com/koron/iview/plugin"
)

type Server struct {
	rootDir string
	rootFS  http.FileSystem
	base    http.Handler

	templateFS *templatefs.FS
}

func New(rootDir string, templateFS fs.FS) *Server {
	root := http.FS(os.DirFS(rootDir))
	return &Server{
		rootDir: rootDir,
		rootFS:  root,
		base:    http.FileServer(root),

		templateFS: templatefs.New(templateFS),
	}
}

func (s *Server) fileToMediaType(f http.File) (string, error) {
	fi, err := f.Stat()
	if err != nil {
		return "", err
	}
	if fi.IsDir() {
		return plugin.MediaTypeDirectory, nil
	}
	ext := path.Ext(fi.Name())
	if mediaTypes, ok := plugin.ExtToMediaType[ext]; ok {
		switch len(mediaTypes) {
		case 0:
			return "", fmt.Errorf("no media types found for extension: %s", ext)
		case 1:
			return mediaTypes[0], nil
		default:
			return plugin.InferMediaType(f, ext, mediaTypes)
		}
	}
	return plugin.MediaTypeDefault, nil
}

type templateRenderer struct {
	*template.Template
}

func (r *templateRenderer) Render(w io.Writer, upath string, f http.File) error {
	return r.Execute(w, &RawFile{File: f, path: upath})
}

var _ HTMLRenderer = (*templateRenderer)(nil)

func (s *Server) determineRenderer(f http.File) (HTMLRenderer, error) {
	mediaType, err := s.fileToMediaType(f)
	if err != nil {
		return nil, err
	}
	// Custom renderer
	if r, ok := plugin.MediaTypeToRenderer[mediaType]; ok {
		return r, nil
	}
	// Default layout template renderer.
	tmpl, err := s.layoutTemplate(mediaType)
	if err != nil {
		return nil, err
	}
	return &templateRenderer{tmpl}, nil
}

func (s *Server) openFile(upath string) (http.File, fs.FileInfo, error) {
	f, err := s.rootFS.Open(upath)
	if err != nil {
		return nil, nil, err
	}
	fi, err := f.Stat()
	if err != nil {
		f.Close()
		return nil, nil, err
	}
	return f, fi, nil
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	upath := path.Clean(r.URL.Path)

	// Open a file and get its information. Resource existence proof.
	f, fi, err := s.openFile(upath)
	if err != nil {
		// Should be 404 not found
		s.serveError(w, r, err)
		return
	}
	defer f.Close()

	if r.Method == "HEAD" {
		w.Header().Set("Date", fi.ModTime().UTC().Format(http.TimeFormat))
		w.WriteHeader(http.StatusOK)
		return
	}

	// Path of directory should be end with "/", normalization.
	if fi.IsDir() && !strings.HasSuffix(r.URL.Path, "/") {
		s.serveRedirect(w, r.URL.Path+"/")
		return
	}

	// If "raw" query parameter is provided, defer to http.FileServer.
	if r.URL.Query().Has("raw") {
		s.base.ServeHTTP(w, r)
		return
	}

	// If "edit" parameter is provided, open with editor.
	if r.URL.Query().Has("edit") {
		fpath := filepath.FromSlash(strings.TrimLeft(upath, "/"))
		cmd := exec.Command("gvim", fpath)
		cmd.Dir = s.rootDir
		err := cmd.Start()
		if err != nil {
			s.serveError(w, r, err)
			return
		}
		w.WriteHeader(http.StatusOK)
		return
	}

	// Determine renderer for the file.
	renderer, err := s.determineRenderer(f)
	if err != nil {
		s.serveError(w, r, err)
		return
	}

	// Render as HTML
	bb := &bytes.Buffer{}
	err = renderer.Render(bb, upath, f)
	if err != nil {
		s.serveError(w, r, err)
		return
	}

	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Date", fi.ModTime().UTC().Format(http.TimeFormat))
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	io.Copy(w, bb)

	return
}

func (s *Server) toHTTPError(err error) int {
	if errors.Is(err, fs.ErrNotExist) {
		return http.StatusNotFound
	}
	if errors.Is(err, fs.ErrPermission) {
		return http.StatusForbidden
	}
	return http.StatusInternalServerError
}

func (s *Server) serveError(w http.ResponseWriter, r *http.Request, err error) {
	statusCode := s.toHTTPError(err)
	log.Printf("request failed: %d %s %s: %s", statusCode, r.Method, r.URL, err)

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(statusCode)
	io.WriteString(w, err.Error())
}

func (s *Server) serveRedirect(w http.ResponseWriter, newURL string) {
	w.Header().Set("Location", newURL)
	w.WriteHeader(http.StatusMovedPermanently)
}

// TODO: Move to plugin package
var funcMap = template.FuncMap{
	"markdown": markdownFunc,
}

type HTMLRenderer interface {
	Render(w io.Writer, upath string, f http.File) error
}

var templatefsOptions = []templatefs.Option{
	templatefs.OptionFunc(func(tmpl *template.Template) (*template.Template, error) {
		return tmpl.Funcs(funcMap), nil
	}),
}

func (s *Server) layoutTemplate(mediaType string) (*template.Template, error) {
	layout, err := s.templateFS.Template("layout.html", templatefsOptions...)
	if err != nil {
		return nil, err
	}
	layout, err = layout.Clone()
	if err != nil {
		return nil, err
	}
	main, err := s.templateFS.Template(path.Join(mediaType, "main.html"), templatefsOptions...)
	if err != nil {
		return nil, err
	}
	_, err = layout.AddParseTree("main", main.Tree)
	if err != nil {
		return nil, err
	}
	return layout, nil
}

type Link struct {
	Name string
	Path string
}

type RawFile struct {
	http.File
	path string
}

func (f *RawFile) Name() (any, error) {
	fi, err := f.Stat()
	if err != nil {
		return nil, err
	}
	return fi.Name(), nil
}

func (f *RawFile) Content() (any, error) {
	b, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}
	return string(b), nil
}

func (f *RawFile) Breadcrumbs() ([]Link, error) {
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
