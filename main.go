package main

import (
	"embed"
	"log"
	"net/http"
	"os"
)

//go:embed _
var assetsFS embed.FS

func main() {
	var (
		address = ":8000"
		dirpath = "."
	)

	// Provide static contents at "/_/"
	http.Handle("/_/", http.FileServer(http.FS(assetsFS)))

	// Provide dynamic contents at others
	http.Handle("/", http.FileServer(http.FS(os.DirFS(dirpath))))

	log.Printf("start to listening on %s", address)
	log.Fatal(http.ListenAndServe(address, nil))
}
