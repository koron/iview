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
	"unicode/utf8"

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

// detectMediaType detects media type of the file.
func (s *Server) detectMediaType(f http.File) (string, error) {
	defer f.Seek(0, io.SeekStart)
	fi, err := f.Stat()
	if err != nil {
		return "", err
	}
	if fi.IsDir() {
		return plugin.MediaTypeDirectory, nil
	}
	ext := path.Ext(fi.Name())
	if mediaTypes, ok := plugin.GetMediaType(ext); ok {
		switch len(mediaTypes) {
		case 0:
			return "", fmt.Errorf("no media types found for extension: %s", ext)
		case 1:
			return mediaTypes[0], nil
		default:
			return plugin.InferMediaType(f, ext, mediaTypes)
		}
	}

	// TODO: Reads up to 4096 bytes (4KiB)  and verifies that it is UTF-8 text.
	b := make([]byte, 4096)
	n, err := f.Read(b)
	if err != nil && !errors.Is(err, io.EOF) {
		return "", err
	}
	for i := 0; i < utf8.UTFMax; i++ {
		if utf8.Valid(b[:n-i]) {
			return plugin.MediaTypePlainText, nil
		}
	}

	return plugin.MediaTypeDefault, nil
}

func (s *Server) determineRenderer(f http.File) (plugin.HTMLRenderer, error) {
	mediaType, err := s.detectMediaType(f)
	if err != nil {
		return nil, err
	}
	// Custom renderer
	if r, ok := plugin.MediaTypeToRenderer[mediaType]; ok {
		return r, nil
	}
	// Default layout template renderer.
	return s.layoutTemplate(mediaType)
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

func (s *Server) layoutTemplate(mediaType string) (*templateRenderer, error) {
	opts := []templatefs.Option{
		templatefs.OptionFunc(func(tmpl *template.Template) (*template.Template, error) {
			return tmpl.Funcs(plugin.GetTemplateFuncMap()), nil
		}),
	}

	// Load main contents template.
	main, err := s.templateFS.Template(path.Join(mediaType, "main.html"), opts...)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, fmt.Errorf("media type: %s is not supported: %w", mediaType, err)
		}
		return nil, err
	}

	// Load layout template.
	layout, err := s.templateFS.Template("layout.html", opts...)
	if err != nil {
		return nil, err
	}

	// Merge layout and main templates.
	layout, err = layout.Clone()
	if err != nil {
		return nil, err
	}
	_, err = layout.AddParseTree("main", main.Tree)
	if err != nil {
		return nil, err
	}

	// Load layout extensions by media type
	ext, err := s.loadLayoutExtensions(mediaType)
	if err != nil {
		return nil, err
	}

	return &templateRenderer{
		Template: layout,
		ext:      ext,
	}, nil
}

func loadAsHTML(fsys fs.FS, name string) (template.HTML, error) {
	f, err := fsys.Open(name)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return "", nil
		}
		return "", err
	}
	defer f.Close()
	b, err := io.ReadAll(f)
	if err != nil {
		return "", err
	}
	return template.HTML(b), nil
}

func (s *Server) loadLayoutExtensions(mediaType string) (*LayoutExtensions, error) {
	head, err := loadAsHTML(s.templateFS, path.Join(mediaType, "layout_ext_head.html"))
	if err != nil {
		return nil, err
	}
	if head == "" {
		return nil, nil
	}
	return &LayoutExtensions{
		Head: head,
	}, nil
}
