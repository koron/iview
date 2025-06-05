package main

import (
	"embed"
	"errors"
	"flag"
	"html"
	"html/template"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path"
)

// static assets
//go:embed _
var assetsFS embed.FS

// default HTML template
//go:embed default.html
var defaultView string

type Server struct {
	root http.FileSystem
	base http.Handler
}

func New(dir string) *Server {
	root := http.FS(os.DirFS(dir))
	return &Server{
		root: root,
		base: http.FileServer(root),
	}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// If "raw" query parameter is provided, defer to http.FileServer.
	if r.URL.Query().Has("raw") {
		s.base.ServeHTTP(w, r)
		return
	}

	upath := path.Clean(r.URL.Path)
	f, err := s.root.Open(upath)
	if err != nil {
		log.Printf("failed to open %s: %s", upath, err)
		w.WriteHeader(s.toHTTPError(err))
		return
	}
	defer f.Close()

	if err = s.serveView(w, r, f); err != nil {
		log.Printf("failed to serve view %s: %s", err)
		w.WriteHeader(s.toHTTPError(err))
		return
	}
}

type File struct {
	http.File
}

func (f *File) Name() (string, error) {
	fi, err := f.Stat()
	if err != nil {
		return "", err
	}
	return fi.Name(), nil
}

func (f *File) Content() (string, error) {
	b, err := io.ReadAll(f)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (s *Server) serveView(w http.ResponseWriter, r *http.Request, f http.File) error {
	// Examine file metadata
	fi, err := f.Stat()
	if err != nil {
		return err
	}

	// Prepare the content
	if fi.IsDir() {
		// FIXME: output the directory contents to data.Content.
		s.base.ServeHTTP(w, r)
		return nil
	}

	if r.Method == "HEAD" {
		w.Header().Set("Date", fi.ModTime().UTC().Format(http.TimeFormat))
		w.WriteHeader(http.StatusOK)
		return nil
	}

	// Load "default" template
	tmpl, err := template.New("default").Parse(defaultView)
	if err != nil {
		return err
	}
	// Execute the template and output as the response
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Date", fi.ModTime().UTC().Format(http.TimeFormat))
	w.WriteHeader(http.StatusOK)
	err = tmpl.Execute(w, &File{f})
	if err != nil {
		log.Printf("template failure: %s", err)
		io.WriteString(w, "<h1>Template Failure</h1>")
		io.WriteString(w, "<div>")
		io.WriteString(w, html.EscapeString(err.Error()))
		io.WriteString(w, "</div>")
	}
	return nil
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

func main() {
	var (
		addr string
		dir  string
	)

	flag.StringVar(&addr, "addr", "localhost:8000", `address that hosts the HTTP server`)
	flag.StringVar(&dir, "dir", ".", `root directory for the content to host`)
	flag.Parse()

	// Provide static contents at "/_/"
	http.Handle("/_/", http.FileServerFS(assetsFS))

	// Provide dynamic contents at others
	http.Handle("/", New(dir))

	log.Printf("start to listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
