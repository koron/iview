package main

import (
	"embed"
	"errors"
	"flag"
	"html/template"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path"
)

//go:embed _
var assetsFS embed.FS

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

//go:embed default.html
var defaultView string

type Data struct {
	Name    string
	Content string
}

func (s *Server) serveView(w http.ResponseWriter, r *http.Request, f http.File) error {
	// Examine file metadata
	fi, err := f.Stat()
	if err != nil {
		return err
	}
	data := Data{
		Name: fi.Name(),
	}
	// Prepare the content
	if fi.IsDir() {
		// FIXME: output the directory contents to data.Content.
		s.base.ServeHTTP(w, r)
		return nil
	} else {
		if r.Method == "HEAD" {
			w.Header().Set("Date", fi.ModTime().UTC().Format(http.TimeFormat))
			w.WriteHeader(http.StatusOK)
			return nil
		}
		b, err := io.ReadAll(f)
		if err != nil {
			return err
		}
		data.Content = string(b)
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
	err = tmpl.Execute(w, &data)
	if err != nil {
		return err
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

	flag.StringVar(&addr, "addr", ":8000", `address that hosts the HTTP server`)
	flag.StringVar(&dir, "dir", ".", `root directory for the content to host`)
	flag.Parse()

	// Provide static contents at "/_/"
	http.Handle("/_/", http.FileServerFS(assetsFS))

	// Provide dynamic contents at others
	http.Handle("/", New(dir))

	log.Printf("start to listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
