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
	"strings"

	"github.com/koron/iview/internal/templatefs"
)

//go:embed _resource
var embedFS embed.FS

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
		log.Printf("failed to serve view %s: %s", upath, err)
		w.WriteHeader(s.toHTTPError(err))
		return
	}
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
	dirs = dirs[:len(dirs)-1]
	links := append(make([]Link, 0, len(dirs)), Link{Name: "(Root)", Path: "/"})
	for _, d := range dirs[1:] {
		links = append(links, Link{
			Name: d,
			Path: links[len(links)-1].Path + d + "/",
		})
	}
	return links, nil
}

var templatefsOptions = []templatefs.Option{
	templatefs.OptionFunc(func(tmpl *template.Template) (*template.Template, error) {
		return tmpl.Funcs(funcMap), nil
	}),
}

var funcMap = template.FuncMap{
	"markdown": markdownFunc,
}

var extToMIMETypes = map[string]string{
	".md": "text/markdown",
}

func toMIMEType(name string) string {
	ext := path.Ext(name)
	if typ, ok := extToMIMETypes[ext]; ok {
		return typ
	}
	return "text/plain"
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

	w.Header().Set("Date", fi.ModTime().UTC().Format(http.TimeFormat))

	if r.Method == "HEAD" {
		w.WriteHeader(http.StatusOK)
		return nil
	}

	// Load template set for layout
	tmpl, err := layoutTemplate(templateFS, toMIMEType(fi.Name()))
	if err != nil {
		return err
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

func layoutTemplate(tfs *templatefs.FS, name string) (*template.Template, error) {
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

var templateFS *templatefs.FS

func main() {
	var (
		addr string
		dir  string
		rsrc string
	)

	flag.StringVar(&addr, "addr", "localhost:8000", `address that hosts the HTTP server`)
	flag.StringVar(&dir, "dir", ".", `root directory for the content to host`)
	flag.StringVar(&rsrc, "rsrc", "", `resource directory for debug`)
	flag.Parse()

	var err error
	var rsrcFS fs.FS

	if rsrc != "" {
		rsrcFS = os.DirFS(rsrc)
	} else {
		rsrcFS, err = fs.Sub(embedFS, "_resource")
		if err != nil {
			log.Fatal(err)
		}
	}

	// Setup template file-system
	tmplFS, err := fs.Sub(rsrcFS, "template")
	if err != nil {
		log.Fatal(err)
	}
	templateFS = templatefs.New(tmplFS)

	// Provide static contents at "/_/"
	staticFS, err := fs.Sub(rsrcFS, "static")
	if err != nil {
		log.Fatal(err)
	}
	http.Handle("/_/static/", http.StripPrefix("/_/static/", http.FileServerFS(staticFS)))

	monitorDir = dir
	http.Handle("/_/stream/", http.StripPrefix("/_/stream/", http.HandlerFunc(serveStream)))

	// Provide dynamic contents at others
	http.Handle("/", New(dir))

	log.Printf("start to listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
