package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"runtime"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/hareku/smart-syncer/pkg/syncer"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name: "smart-syncer",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "src",
				Required: true,
			},
			&cli.UintFlag{
				Name:     "depth",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "region",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "bucket",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "prefix",
				Required: true,
			},
			&cli.BoolFlag{
				Name: "dryrun",
			},
			&cli.BoolFlag{
				Name:  "minio",
				Usage: "use minio instead of s3",
			},
		},
		Action: func(c *cli.Context) error {
			var cfg *aws.Config
			if c.Bool("minio") {
				cfg = &aws.Config{
					Credentials:      credentials.NewStaticCredentials("minio", "minio123", ""),
					Region:           aws.String(c.String("region")),
					Endpoint:         aws.String("http://127.0.0.1:9000"),
					S3ForcePathStyle: aws.Bool(true),
				}
			} else {
				cfg = aws.NewConfig().WithRegion(c.String("region"))
			}
			s3Client := s3.New(session.Must(session.NewSession(cfg)))

			concurrency := runtime.NumCPU()
			if concurrency > 5 {
				concurrency = 5
			}
			log.Printf("Running concurrency: %d", concurrency)

			uploader := s3manager.NewUploaderWithClient(s3Client)
			uploader.Concurrency = concurrency

			client := &syncer.Client{
				Concurrency:  concurrency,
				LocalStorage: syncer.NewLocalStorage(),
				Archiver:     syncer.NewArchiver(),
				Repository: syncer.NewRepositoryS3(&syncer.NewRepositoryS3Input{
					Bucket:   c.String("bucket"),
					Prefix:   c.String("prefix"),
					API:      s3Client,
					Uploader: uploader,
				}),
				Dryrun: c.Bool("dryrun"),
			}

			if c.Int("depth") < 1 {
				return fmt.Errorf("option -depth must be greater than 0")
			}

			begin := time.Now()
			if err := client.Run(context.Background(), &syncer.ClientRunInput{
				Path:  c.String("src"),
				Depth: c.Int("depth"),
			}); err != nil {
				return err
			}
			log.Printf("Done in %v", time.Since(begin))

			stat := runtime.MemStats{}
			runtime.ReadMemStats(&stat)
			log.Printf("Allocated memory: %d bytes", stat.Alloc)
			log.Printf("Allocated heap: %d bytes", stat.HeapAlloc)
			log.Printf("Allocated total: %d bytes", stat.TotalAlloc)
			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
	os.Exit(0)
}
