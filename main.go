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

	contents, err := fs.Sub(assetsFS, "assets")
	if err != nil {
		log.Fatal(err)
	}
	http.Handle("/_/", http.StripPrefix("/_/", http.FileServer(http.FS(contents))))

	//os.DirFS()
	//root, err := os.OpenRoot(dirpath)
	//if err != kkkkkkkk
	//http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
	//	fmt.Printf("path: %s\n", r.URL.Path)
	//})

	http.Handle("/", http.FileServer(http.FS(os.DirFS(dirpath))))

	log.Fatal(http.ListenAndServe(address, nil))
}
