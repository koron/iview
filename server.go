package main

import (
	"errors"
	"html"
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

const MediaTypeDirectory = "application/vnd.iview.directory"

var extToMIMETypes = map[string]string{
	".md": "text/markdown",
}

func (s *Server) fileToMediaType(fi fs.FileInfo, f fs.File) (string, error) {
	if fi.IsDir() {
		return MediaTypeDirectory, nil
	}
	ext := path.Ext(fi.Name())
	// TODO: custom media type
	if mediaType, ok := extToMIMETypes[ext]; ok {
		return mediaType, nil
	}
	return "text/plain", nil
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// If "raw" query parameter is provided, defer to http.FileServer.
	if r.URL.Query().Has("raw") {
		s.base.ServeHTTP(w, r)
		return
	}

	// If "edit" parameter is provided, open with editor.
	if r.URL.Query().Has("edit") {
		fpath := filepath.FromSlash(strings.TrimLeft(r.URL.Path, "/"))
		cmd := exec.Command("gvim", fpath)
		cmd.Dir = s.rootDir
		err := cmd.Start()
		w.WriteHeader(s.toHTTPError(err))
		return
	}

	// Open a file of by path.
	upath := path.Clean(r.URL.Path)
	f, err := s.rootFS.Open(upath)
	if err != nil {
		log.Printf("failed to open %s: %s", upath, err)
		w.WriteHeader(s.toHTTPError(err))
		return
	}
	defer f.Close()

	// Examine file metadata
	fi, err := f.Stat()
	if err != nil {
		log.Printf("failed to stat %s: %s", upath, err)
		w.WriteHeader(s.toHTTPError(err))
		return
	}

	// Path of directory should be end with "/".
	if fi.IsDir() && !strings.HasSuffix(r.URL.Path, "/") {
		newPath := r.URL.Path + "/"
		w.Header().Set("Location", newPath)
		w.WriteHeader(http.StatusMovedPermanently)
		return
	}

	w.Header().Set("Cache-Control", "no-store")

	mediaType, err := s.fileToMediaType(fi, f)
	if err != nil {
		w.WriteHeader(s.toHTTPError(err))
		return
	}

	// Prepare the content

	w.Header().Set("Date", fi.ModTime().UTC().Format(http.TimeFormat))

	if r.Method == "HEAD" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Load template set for layout
	tmpl, err := s.layoutTemplate(s.templateFS, mediaType)
	if err != nil {
		log.Printf("failed to determine template for %s: %s", fi.Name(), err)
		w.WriteHeader(s.toHTTPError(err))
		return
	}
	// Execute the template and output as the response
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	err = tmpl.Execute(w, &RawFile{File: f, path: r.URL.Path})
	if err != nil {
		log.Printf("template failure: %s", err)
		io.WriteString(w, "<h1>Template Failure</h1>")
		io.WriteString(w, "<div>")
		io.WriteString(w, html.EscapeString(err.Error()))
		io.WriteString(w, "</div>")
	}
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

var funcMap = template.FuncMap{
	"markdown": markdownFunc,
}

var templatefsOptions = []templatefs.Option{
	templatefs.OptionFunc(func(tmpl *template.Template) (*template.Template, error) {
		return tmpl.Funcs(funcMap), nil
	}),
}

func (s *Server) layoutTemplate(tfs *templatefs.FS, name string) (*template.Template, error) {
	layout, err := tfs.Template("layout.html", templatefsOptions...)
	if err != nil {
		return nil, err
	}
	layout, err = layout.Clone()
	if err != nil {
		return nil, err
	}
	main, err := tfs.Template(path.Join(name, "main.html"), templatefsOptions...)
	if err != nil {
		return nil, err
	}
	_, err = layout.AddParseTree("main", main.Tree)
	if err != nil {
		return nil, err
	}
	return layout, nil
}

func toMIMEType(name string) string {
	ext := path.Ext(name)
	if typ, ok := extToMIMETypes[ext]; ok {
		return typ
	}
	return "text/plain"
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
