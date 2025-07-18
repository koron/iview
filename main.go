package main

import (
	"embed"
	"errors"
	"flag"
	"io/fs"
	"log"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/koron/iview/internal/browser"
	"github.com/koron/iview/internal/fschanges"
)

//go:embed _resource
var embedFS embed.FS

var (
	flagAddr   string
	flagDir    string
	flagRsrc   string
	flagEditor string
	flagWeb    bool
)

func editorCommand() (string, error) {
	// TODO: parse special placeholders: "%file", "%line" for flagEditor and
	// IVIEW_EDITOR.
	if flagEditor != "" {
		return flagEditor, nil
	}
	if s := os.Getenv("IVIEW_EDITOR"); s != "" {
		return s, nil
	}
	if s := os.Getenv("EDITOR"); s != "" {
		return s, nil
	}
	switch runtime.GOOS {
	case "darwin":
		return "open", nil
	case "freebsd":
		return "xdg-open", nil
	case "linux":
		return "xdg-open", nil
	case "windows":
		return "notepad", nil
	}
	return "", errors.New("no default editors. please set environment variables EDITOR, IVIEW_EDITOR, or -editor flag on start up")
}

func main() {
	flag.StringVar(&flagAddr, "addr", "localhost:8000", `address that hosts the HTTP server`)
	flag.StringVar(&flagDir, "dir", ".", `root directory for the content to host`)
	flag.StringVar(&flagRsrc, "rsrc", "", `resource directory for debug`)
	flag.StringVar(&flagEditor, "editor", "", `editor to open the file`)
	flag.BoolVar(&flagWeb, "web", false, `start the browser`)
	flag.Parse()

	var err error
	var rsrcFS fs.FS

	if flagRsrc != "" {
		rsrcFS = os.DirFS(flagRsrc)
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

	// Handle favicon.ico differently using redirects.
	http.Handle("/favicon.ico", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/_/static/favicon.ico", http.StatusMovedPermanently)
	}))

	es := fschanges.New(flagDir, fschanges.WithExcludeDirs(".git"))
	http.Handle("/_/stream/", http.StripPrefix("/_/stream/", es))

	// Provide dynamic contents at others
	tmplFS, err := fs.Sub(rsrcFS, "template")
	if err != nil {
		log.Fatal(err)
	}
	http.Handle("/", New(flagDir, tmplFS))

	if flagWeb {
		// Start the web browser
		go func() {
			time.Sleep(200 * time.Millisecond)
			browser.Open("http://" + flagAddr)
		}()
	}

	log.Printf("start to listening on %s", flagAddr)
	log.Fatal(http.ListenAndServe(flagAddr, nil))
}
