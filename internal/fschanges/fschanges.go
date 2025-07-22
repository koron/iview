// Package fschanges provides HTTP server that streams change events on a filesystem.
package fschanges

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/koron/iview/internal/fsmonitor"
)

type EventServer struct {
	dir string

	mon     *fsmonitor.Monitor
	monMu   sync.Mutex
	monOpts []fsmonitor.Option
}

func New(dir string, opts ...Option) *EventServer {
	es := &EventServer{
		dir: dir,
	}
	for _, o := range opts {
		o.apply(es)
	}
	return es
}

func (es *EventServer) monitor() (*fsmonitor.Monitor, error) {
	es.monMu.Lock()
	defer es.monMu.Unlock()
	if es.mon != nil {
		return es.mon, nil
	}
	m, err := fsmonitor.New(context.Background(), es.dir, es.monOpts...)
	if err != nil {
		return nil, err
	}
	es.mon = m
	return es.mon, nil
}

func (es *EventServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	m, err := es.monitor()
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
			slog.Debug("fsmonitor receive", "event", ev)
			b, _ := json.Marshal(toChangeEvent(ev))
			io.WriteString(w, "data: ")
			w.Write(b)
			io.WriteString(w, "\n\n")
		}
		flushWriter(w)
	}
}

func flushWriter(w http.ResponseWriter) {
	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}
}

type changeEvent struct {
	Path string   `json:"path"`
	Type []string `json:"type"`
}

func toChangeEvent(src fsmonitor.Event) changeEvent {
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
	return changeEvent{
		Path: src.Path,
		Type: typ,
	}
}

type Option interface {
	apply(*EventServer)
}

type optionFunc func(*EventServer)

func (f optionFunc) apply(es *EventServer) { f(es) }

var _ Option = (optionFunc)(nil)

func WithExcludeDirs(dirs ...string) Option {
	return optionFunc(func(es *EventServer) {
		es.monOpts = append(es.monOpts, fsmonitor.WithExcludeDirs(dirs...))
	})
}
