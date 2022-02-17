package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

var c *s3.S3

func main() {
	c = s3.New(session.Must(session.NewSession(aws.NewConfig().WithRegion("us-west-2"))))

	if err := run(context.Background()); err != nil {
		log.Fatal(err)
	}
}

func run(ctx context.Context) error {
	objects, err := listLocal(ctx)
	if err != nil {
		return err
	}

	for _, v := range objects {
		fmt.Println(v.Key)
	}
	return nil
}

type LocalObject struct {
	Key          string
	LastModified time.Time
}

func listLocal(ctx context.Context) ([]LocalObject, error) {
	res := []LocalObject{}

	files, err := os.ReadDir("D:/User/Desktop/同人誌")
	if err != nil {
		return nil, err
	}
	for _, f := range files {
		res = append(res, LocalObject{
			Key:          f.Name(),
			LastModified: time.Now(),
		})
	}

	return res, nil
}
