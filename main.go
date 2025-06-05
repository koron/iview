package main

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
	"os"
)

//go:embed _
var assetsFS embed.FS

type Server struct {
	root fs.FS
	base http.Handler
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.base.ServeHTTP(w, r)
}

func main() {
	var (
		address = ":8000"
		dirpath = "."
	)

	// Provide static contents at "/_/"
	http.Handle("/_/", http.FileServerFS(assetsFS))

	// Provide dynamic contents at others
	root := os.DirFS(dirpath)
	http.Handle("/", &Server{
		root:   root,
		base: http.FileServerFS(root),
	})

	log.Printf("start to listening on %s", address)
	log.Fatal(http.ListenAndServe(address, nil))
}
