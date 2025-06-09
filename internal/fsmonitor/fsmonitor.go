/*
Package fsmonitor monitors change events in the specified directory.
The detected Event are sent as messages to subscribers who have subscribed to the Topic.
*/
package fsmonitor

import (
	"context"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/koron/iview/internal/pubsub"
)

type Monitor struct {
	cancel  context.CancelFunc
	wg      *sync.WaitGroup
	rootDir string
	watcher *fsnotify.Watcher
	topic   *pubsub.Topic[Event]
}

type Type = fsnotify.Op

type Event struct {
	Path string
	Type Type
}

func New(ctx context.Context, dir string) (*Monitor, error) {
	ctx2, cancel := context.WithCancel(ctx)
	w, err := fsnotify.NewWatcher()
	if err != nil {
		cancel()
		return nil, err
	}
	m := &Monitor{
		cancel:  cancel,
		wg:      &sync.WaitGroup{},
		rootDir: dir,
		watcher: w,
		topic:   pubsub.New[Event](),
	}

	// Add target directory and its sub directories to the watch list
	// recursively
	w.Add(dir)
	filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err == nil && d.IsDir() {
			// TODO: Customize ignore directories
			if d.Name() == ".git" {
				return fs.SkipDir
			}
			w.Add(path)
		}
		return err
	})

	m.wg.Add(1)
	go m.run(ctx2)

	return m, nil
}

func (m *Monitor) run(ctx context.Context) {
	for {
		select {
		case e := <-m.watcher.Events:
			//log.Printf("%+v", e)
			switch e.Op {
			case fsnotify.Create:
				// Add a newly created directory to the watch list.
				fi, err := os.Stat(e.Name)
				if err != nil {
					log.Printf("fail to stat on %s: %s", e.Name, err)
					break
				}
				if fi.IsDir() {
					m.watcher.Add(e.Name)
				}
			}
			// Compose a path of the event target on the HTTP server
			name, err := filepath.Rel(m.rootDir, e.Name)
			if err != nil {
				log.Printf("fail to calc relative path: %s", err)
				break
			}
			m.topic.Publish(Event{
				Path: "/" + filepath.ToSlash(name),
				Type: Type(e.Op),
			})
		case <-ctx.Done():
			m.watcher.Close()
			m.wg.Done()
			return
		}
	}
}

func (m *Monitor) Topic() *pubsub.Topic[Event] {
	return m.topic
}

func (m *Monitor) Close() {
	m.cancel()
	m.wg.Wait()
	m.topic.Close()
}
