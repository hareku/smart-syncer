package syncer

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/aws/aws-sdk-go/service/s3/s3manager/s3manageriface"
)

type RepositoryS3 struct {
	bucket       string
	prefix       string // prefix with "/" suffix of S3 bucket
	storageClass string
	api          s3iface.S3API
	uploader     s3manageriface.UploaderAPI
}

type NewRepositoryS3Input struct {
	Bucket       string
	Prefix       string
	StorageClass string
	API          s3iface.S3API
	Uploader     s3manageriface.UploaderAPI
}

func NewRepositoryS3(in *NewRepositoryS3Input) Repository {
	return &RepositoryS3{
		bucket:       in.Bucket,
		prefix:       strings.Trim(in.Prefix, "/") + "/",
		storageClass: in.StorageClass,
		api:          in.API,
		uploader:     in.Uploader,
	}
}

func (s *RepositoryS3) List(ctx context.Context) ([]RepositoryObject, error) {
	res := []RepositoryObject{}

	err := s.api.ListObjectsV2PagesWithContext(ctx, &s3.ListObjectsV2Input{
		Bucket: &s.bucket,
		Prefix: &s.prefix,
	}, func(lovo *s3.ListObjectsV2Output, b bool) bool {
		for _, o := range lovo.Contents {
			res = append(res, RepositoryObject{
				Key:              strings.TrimPrefix(*o.Key, s.prefix),
				LastModifiedUnix: (*o.LastModified).Unix(),
			})
		}
		return true
	})
	if err != nil {
		return nil, fmt.Errorf("s3 listing objects failed: %w", err)
	}
	return res, nil
}

func (s *RepositoryS3) Upload(ctx context.Context, key string, r io.Reader) error {
	_, err := s.uploader.UploadWithContext(ctx, &s3manager.UploadInput{
		Bucket:       &s.bucket,
		Key:          aws.String(strings.TrimPrefix(s.prefix+key, "/")),
		Body:         r,
		StorageClass: aws.String(s.storageClass),
	})
	if err != nil {
		return fmt.Errorf("s3 uploading failed: %w", err)
	}
	return nil
}

func (s *RepositoryS3) Delete(ctx context.Context, keys []string) error {
	ids := make([]*s3.ObjectIdentifier, len(keys))
	for i, k := range keys {
		ids[i] = &s3.ObjectIdentifier{
			Key: aws.String(s.prefix + k),
		}
	}

	_, err := s.api.DeleteObjects(&s3.DeleteObjectsInput{
		Bucket: &s.bucket,
		Delete: &s3.Delete{
			Objects: ids,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to delete objects: %w", err)
	}
	return nil
}
