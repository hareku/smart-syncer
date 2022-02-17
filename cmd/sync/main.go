package main

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/hareku/smart-syncer/pkg/syncer"
)

func main() {
	s3Client := s3.New(session.Must(session.NewSession(aws.NewConfig().WithRegion("us-west-2"))))

	client := &syncer.Client{
		LocalStorage: syncer.NewLocalStorage(),
		Archiver:     syncer.NewArchiver(),
		Repository: syncer.NewRepositoryS3(&syncer.NewRepositoryS3Input{
			Bucket:   "syncer-bucket",
			Prefix:   "prefix",
			API:      s3Client,
			Uploader: s3manager.NewUploaderWithClient(s3Client),
		}),
	}

	if err := client.Run(context.Background(), &syncer.ClientRunInput{
		// Path: "D:/User/Desktop/同人誌",
		Path:  "D:/Code/github/smart-syncer/syncer-test",
		Depth: 1,
	}); err != nil {
		log.Fatal(err)
	}
}
