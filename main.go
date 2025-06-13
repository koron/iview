package main

import (
	"embed"
	"flag"
	"io/fs"
	"log"
	"net/http"
	"os"

	"github.com/koron/iview/internal/fschanges"
)

//go:embed _resource
var embedFS embed.FS

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

	// Provide static contents at "/_/"
	staticFS, err := fs.Sub(rsrcFS, "static")
	if err != nil {
		log.Fatal(err)
	}
	http.Handle("/_/static/", http.StripPrefix("/_/static/", http.FileServerFS(staticFS)))

	es := fschanges.New(dir, fschanges.WithExcludeDirs(".git"))
	http.Handle("/_/stream/", http.StripPrefix("/_/stream/", es))

	// Provide dynamic contents at others
	tmplFS, err := fs.Sub(rsrcFS, "template")
	if err != nil {
		log.Fatal(err)
	}
	http.Handle("/", New(dir, tmplFS))

	log.Printf("start to listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
