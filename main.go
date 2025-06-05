package main

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
	"os"
)

//go:embed assets
var assetsFS embed.FS

func main() {
	var (
		address = ":8000"
		dirpath = "."
	)

	// Provide static contents under "/_/"
	contents, err := fs.Sub(assetsFS, "assets")
	if err != nil {
		log.Fatal(err)
	}
	http.Handle("/_/", http.StripPrefix("/_/", http.FileServer(http.FS(contents))))

	// Provide dynamic contents
	http.Handle("/", http.FileServer(http.FS(os.DirFS(dirpath))))

	log.Fatal(http.ListenAndServe(address, nil))
}
