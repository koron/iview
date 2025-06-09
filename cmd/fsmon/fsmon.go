package main

// A tool to check the differences in fsnotify behavior on each platform

import (
	"context"
	"flag"
	"log"

	"github.com/koron/iview/internal/fsmonitor"
)

func main() {
	var dir string
	flag.StringVar(&dir, "dir", ".", `root directory for the content to host`)
	flag.Parse()
	monitor, err := fsmonitor.New(context.Background(), dir)
	if err != nil {
		log.Fatal(err)
	}
	s := monitor.Topic().Subscribe(10)
	defer monitor.Topic().Unsubscribe(s)
	for ev := range s.Channel() {
		log.Printf("%+v", ev)
	}
}
