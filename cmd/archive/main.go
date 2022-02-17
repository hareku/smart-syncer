package main

import (
	"context"
	"log"
	"os"

	"github.com/hareku/smart-syncer/pkg/syncer"
)

func main() {
	err := run()
	if err != nil {
		log.Fatal(err)
	}
}

func run() error {
	f, err := os.Create("./build/test.tar")
	if err != nil {
		return err
	}
	defer f.Close()

	a := syncer.NewArchiver()
	return a.Do(context.Background(), "D:/syncer-test", f)
}
