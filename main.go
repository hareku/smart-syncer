package main

import (
	"context"
	"fmt"
	"log"
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
	objects, err := listS3(ctx)
	if err != nil {
		return err
	}

	for _, v := range objects {
		fmt.Println(v.Key)
	}
	return nil
}

type Object struct {
	Key          string
	LastModified time.Time
}

func listS3(ctx context.Context) ([]Object, error) {
	res := []Object{}

	err := c.ListObjectsV2PagesWithContext(ctx, &s3.ListObjectsV2Input{
		Bucket: aws.String("h4reku-backup"),
	}, func(lovo *s3.ListObjectsV2Output, b bool) bool {
		for _, o := range lovo.Contents {
			res = append(res, Object{
				Key:          *o.Key,
				LastModified: *o.LastModified,
			})
		}
		return true
	})
	if err != nil {
		return nil, err
	}
	return res, nil
}
