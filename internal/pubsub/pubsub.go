/*
Package pubsub provides simple pub/sub feature.
Messages are guaranteed to be in order and subscribers receive them through channels.
This package does not create goroutines for you; subscribers must create and manage goroutines to receive messages.
*/
package pubsub

import (
	"sync"
)

type Topic[M any] struct {
	mu   sync.Mutex
	subs map[*Subscription[M]]struct{}
}

func New[M any]() *Topic[M] {
	return &Topic[M]{
		subs: map[*Subscription[M]]struct{}{},
	}
}

func (t *Topic[M]) Subscribe(bufsize int) *Subscription[M] {
	t.mu.Lock()
	defer t.mu.Unlock()
	s := newSub[M](max(bufsize, 1))
	t.subs[s] = struct{}{}
	return s
}

func (t *Topic[M]) unsubscribe(s *Subscription[M]) {
	if _, ok := t.subs[s]; !ok {
		return
	}
	delete(t.subs, s)
	close(s.ch)
}

func (t *Topic[M]) Unsubscribe(s *Subscription[M]) {
	t.mu.Lock()
	t.unsubscribe(s)
	t.mu.Unlock()
}

func (t *Topic[M]) Publish(message M) {
	t.mu.Lock()
	defer t.mu.Unlock()
	for s := range t.subs {
		select {
		case s.ch <- message:
		default:
			// Drop a message when channel is full/busy
			continue
		}
	}
}

func (t *Topic[M]) Close() {
	t.mu.Lock()
	defer t.mu.Unlock()
	for s := range t.subs {
		t.unsubscribe(s)
	}
}

type Subscription[M any] struct {
	ch chan M
}

func newSub[M any](bufferSize int) *Subscription[M] {
	ch := make(chan M, bufferSize)
	s := &Subscription[M]{
		ch: ch,
	}
	return s
}

func (s *Subscription[M]) Channel() <-chan M {
	return s.ch
}
