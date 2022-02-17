package main

import (
	"context"
	"fmt"
	"log"

	"github.com/hareku/smart-syncer/pkg/syncer"
)

func main() {
	// c = s3.New(session.Must(session.NewSession(aws.NewConfig().WithRegion("us-west-2"))))

	if err := run(context.Background()); err != nil {
		log.Fatal(err)
	}
}

func run(ctx context.Context) error {
	storage := syncer.NewLocalStorage()

	objects, err := storage.List(ctx, "D:/syncer-test", 1)
	if err != nil {
		return err
	}

	for _, v := range objects {
		fmt.Println(v.Key)
	}
	return nil
}
