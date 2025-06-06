package pubsub_test

import (
	"sync"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/koron/iview/internal/pubsub"
)

type History []string

func (h *History) Add(m string) {
	*h = append(*h, m)
}

func runSub(t *testing.T, wg *sync.WaitGroup, s *pubsub.Subscription[string], id int) (*pubsub.Subscription[string], *History) {
	h := &History{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for m := range s.Channel() {
			h.Add(m)
		}
	}()
	return s, h
}

func TestPubsub(t *testing.T) {
	topic := pubsub.New[string]()

	wg := &sync.WaitGroup{}

	bufsize := 8
	_, h1 := runSub(t, wg, topic.Subscribe(bufsize), 1)
	s2, h2 := runSub(t, wg, topic.Subscribe(bufsize), 2)
	_, h3 := runSub(t, wg, topic.Subscribe(bufsize), 3)

	topic.Publish("m1")
	topic.Unsubscribe(s2)
	topic.Publish("m2")
	_, h4 := runSub(t, wg, topic.Subscribe(bufsize), 4)
	topic.Publish("m3")

	time.Sleep(10 * time.Millisecond)

	topic.Close()
	wg.Wait()

	if d := cmp.Diff(History([]string{"m1", "m2", "m3"}), *h1); d != "" {
		t.Errorf("%s unmatch: -want +got\n%s", "h1", d)
	}
	if d := cmp.Diff(History([]string{"m1"}), *h2); d != "" {
		t.Errorf("%s unmatch: -want +got\n%s", "h2", d)
	}
	if d := cmp.Diff(History([]string{"m1", "m2", "m3"}), *h3); d != "" {
		t.Errorf("%s unmatch: -want +got\n%s", "h3", d)
	}
	if d := cmp.Diff(History([]string{"m3"}), *h4); d != "" {
		t.Errorf("%s unmatch: -want +got\n%s", "h4", d)
	}
}
