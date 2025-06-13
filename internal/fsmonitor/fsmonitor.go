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
	cancel   context.CancelFunc
	wg       *sync.WaitGroup
	rootDir  string
	watcher  *fsnotify.Watcher
	topic    *pubsub.Topic[Event]
	excludes map[string]struct{}
}

type Type = fsnotify.Op

type Event struct {
	Path string
	Type Type
}

func New(ctx context.Context, dir string, opts ...Option) (*Monitor, error) {
	ctx2, cancel := context.WithCancel(ctx)
	w, err := fsnotify.NewWatcher()
	if err != nil {
		cancel()
		return nil, err
	}
	m := &Monitor{
		cancel:   cancel,
		wg:       &sync.WaitGroup{},
		rootDir:  dir,
		watcher:  w,
		topic:    pubsub.New[Event](),
		excludes: map[string]struct{}{},
	}
	for _, o := range opts {
		o.apply(m)
	}

	// Start monitoring
	m.wg.Add(1)
	go m.run(ctx2)

	return m, nil
}

type entryInfo interface {
	Name() string
	IsDir() bool
}

var _ entryInfo = (fs.DirEntry)(nil)
var _ entryInfo = (fs.FileInfo)(nil)

type entryType int

const (
	etFile entryType = iota
	etWatch
	etExclude
)

func (m *Monitor) targetType(entry entryInfo) entryType {
	if !entry.IsDir() {
		return etFile
	}
	if _, ok := m.excludes[entry.Name()]; ok {
		return etExclude
	}
	return etWatch
}

func (m *Monitor) addWatch(dir string) {
	//log.Printf("addWatch: %s", dir)
	m.watcher.Add(dir)
}

func (m *Monitor) run(ctx context.Context) {
	// Add target directory and its sub directories to the watch list
	// recursively
	m.addWatch(m.rootDir)
	filepath.WalkDir(m.rootDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		switch m.targetType(d) {
		case etWatch:
			m.addWatch(path)
		case etExclude:
			return filepath.SkipDir
		}
		return nil
	})

	// Monitoring main loop
	for {
		select {
		case <-ctx.Done():
			m.watcher.Close()
			m.wg.Done()
			return
		case e := <-m.watcher.Events:
			//log.Printf("fsnotify detected: %+v", e)
			switch e.Op {
			case fsnotify.Create:
				// Add a newly created directory to the watch list.
				fi, err := os.Stat(e.Name)
				if err != nil {
					log.Printf("fail to stat on %s: %s", e.Name, err)
					break
				}
				if m.targetType(fi) == etWatch {
					m.addWatch(e.Name)
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

type Option interface {
	apply(*Monitor)
}

type optionFunc func(*Monitor)

func (f optionFunc) apply(m *Monitor) { f(m) }

func WithExcludeDirs(dirs ...string) Option {
	return optionFunc(func(m *Monitor) {
		for _, dir := range dirs {
			m.excludes[dir] = struct{}{}
		}
	})
}
