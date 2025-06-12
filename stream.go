package main

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/koron/iview/internal/fsmonitor"
)

var monitorDir = "."

var monitor = sync.OnceValues(func() (*fsmonitor.Monitor, error) {
	//log.Printf("fsmonitor start on %s", monitorDir)
	return fsmonitor.New(context.Background(), monitorDir, fsmonitor.WithExcludeDirs(".git"))
})

type fsChangeEvent struct {
	Path string   `json:"path"`
	Type []string `json:"type"`
}

func toFSChangeEvent(src fsmonitor.Event) fsChangeEvent {
	var typ []string
	if src.Type.Has(fsnotify.Create) {
		typ = append(typ, "create")
	}
	if src.Type.Has(fsnotify.Write) {
		typ = append(typ, "write")
	}
	if src.Type.Has(fsnotify.Remove) {
		typ = append(typ, "remove")
	}
	if src.Type.Has(fsnotify.Rename) {
		typ = append(typ, "rename")
	}
	if src.Type.Has(fsnotify.Chmod) {
		typ = append(typ, "chmod")
	}
	return fsChangeEvent{
		Path: src.Path,
		Type: typ,
	}
}

func flushWriter(w http.ResponseWriter) {
	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}
}

func serveStream(w http.ResponseWriter, r *http.Request) {
	m, err := monitor()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, err.Error())
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	flushWriter(w)

	s := m.Topic().Subscribe(10)
	defer m.Topic().Unsubscribe(s)
	for {
		select {
		case <-r.Context().Done():
			return
		case ev := <-s.Channel():
			//log.Printf("fsmonitor receive: %+v", ev)
			b, _ := json.Marshal(toFSChangeEvent(ev))
			io.WriteString(w, "data: ")
			w.Write(b)
			io.WriteString(w, "\n\n")
		}
		flushWriter(w)
	}
}
