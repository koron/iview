package main

import (
	"fmt"
	"log"
	"os"

	"github.com/koron/iview/internal/gitfunc"
)

func gitstatus(dir string) error {
	status, err := gitfunc.DirStatus(dir)
	if err != nil {
		return err
	}
	fmt.Printf("status of %s\n", dir)
	for n, s := range status {
		fmt.Printf("\t%s (S:'%c', W:'%c', %s)\n", n, s.Staging, s.Worktree, s.Extra)
	}
	return nil
}

func main() {
	dir := "."
	if len(os.Args) >= 2 {
		dir = os.Args[1]
	}
	if err := gitstatus(dir); err != nil {
		log.Fatal(err)
	}
}
