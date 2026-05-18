/*
Package fsmonitor monitors change events in the specified directory.
The detected Event are sent as messages to subscribers who have subscribed to the Topic.
*/
package fsmonitor

import (
	"context"
	"io/fs"
	"log/slog"
	"path/filepath"
	"sync"

	"github.com/fswatcher/fswatcher"
	"github.com/koron/iview/internal/pubsub"
)

type Monitor struct {
	cancel   context.CancelFunc
	wg       *sync.WaitGroup
	rootDir  string
	watcher  *fswatcher.Watcher
	topic    *pubsub.Topic[Event]
	excludes map[string]struct{}
}

type Type = fswatcher.Op

type Event struct {
	Path string
	Type Type
}

func regulateRootDir(dir string) (string, error) {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}
	expanded, err := filepath.EvalSymlinks(abs)
	if err != nil {
		return "", err
	}
	return expanded, nil
}

func New(ctx context.Context, dir string, opts ...Option) (*Monitor, error) {
	ctx2, cancel := context.WithCancel(ctx)
	w, err := fswatcher.NewWatcher()
	if err != nil {
		cancel()
		return nil, err
	}
	rootDir, err := regulateRootDir(dir)
	if err != nil {
		cancel()
		return nil, err
	}
	m := &Monitor{
		cancel:   cancel,
		wg:       &sync.WaitGroup{},
		rootDir:  rootDir,
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

func (m *Monitor) isExcluded(entry entryInfo) bool {
	_, ok := m.excludes[entry.Name()]
	return ok
}

func (m *Monitor) run(ctx context.Context) {
	// Add target directory and its sub directories to the watch list
	// recursively
	m.watcher.AddRecursive(m.rootDir, fswatcher.All)
	filepath.WalkDir(m.rootDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if m.isExcluded(d) {
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
			slog.Debug("fswatcher detected", "event", e)
			// Compose a path of the event target on the HTTP server
			name, err := filepath.Rel(m.rootDir, e.Name)
			if err != nil {
				slog.Warn("fail to calc relative path", "error", err)
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
