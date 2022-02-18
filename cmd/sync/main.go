package main

import (
	"context"
	"log"
	"runtime"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/hareku/smart-syncer/pkg/syncer"
)

func main() {
	cfg := aws.Config{
		Credentials:      credentials.NewStaticCredentials("minio", "minio123", ""),
		Region:           aws.String("ap-northeast-1"),
		Endpoint:         aws.String("http://127.0.0.1:9000"),
		S3ForcePathStyle: aws.Bool(true),
	}

	s3Client := s3.New(session.Must(session.NewSession(&cfg)))
	// s3Client := s3.New(session.Must(session.NewSession(aws.NewConfig().WithRegion("us-west-2"))))

	concurrency := runtime.NumCPU()
	log.Printf("Running concurrency: %d", concurrency)

	uploader := s3manager.NewUploaderWithClient(s3Client)
	uploader.Concurrency = concurrency

	client := &syncer.Client{
		Concurrency:  concurrency,
		LocalStorage: syncer.NewLocalStorage(),
		Archiver:     syncer.NewArchiver(),
		Repository: syncer.NewRepositoryS3(&syncer.NewRepositoryS3Input{
			Bucket:   "testing",
			Prefix:   "Doujinshi2",
			API:      s3Client,
			Uploader: uploader,
		}),
		Dryrun: false,
	}

	s := time.Now()
	if err := client.Run(context.Background(), &syncer.ClientRunInput{
		Path:  "D:/User/Desktop/同人誌2",
		Depth: 2,
		// Path:  "D:/Code/github/smart-syncer/syncer-test",
		// Depth: 1,
	}); err != nil {
		log.Fatal(err)
	}
	log.Printf("done in %v", time.Since(s))

	stat := runtime.MemStats{}
	runtime.ReadMemStats(&stat)
	log.Printf("mem stats: %+v", stat)
}
