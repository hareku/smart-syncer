package syncer

import (
	"context"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

type RepositoryS3 struct {
	bucket   *string
	prefix   *string
	api      *s3.S3
	uploader *s3manager.Uploader
}

type NewRepositoryS3Input struct {
	Bucket   *string
	Prefix   *string
	API      *s3.S3
	Uploader *s3manager.Uploader
}

func NewRepositoryS3(in *NewRepositoryS3Input) Repository {
	return &RepositoryS3{
		bucket:   in.Bucket,
		prefix:   in.Prefix,
		api:      in.API,
		uploader: in.Uploader,
	}
}

func (s *RepositoryS3) List(ctx context.Context) ([]RepositoryObject, error) {
	res := []RepositoryObject{}

	err := s.api.ListObjectsV2PagesWithContext(ctx, &s3.ListObjectsV2Input{
		Bucket: s.bucket,
		Prefix: s.prefix,
	}, func(lovo *s3.ListObjectsV2Output, b bool) bool {
		for _, o := range lovo.Contents {
			res = append(res, RepositoryObject{
				Key:          *o.Key,
				LastModified: *o.LastModified,
			})
		}
		return true
	})
	if err != nil {
		return nil, fmt.Errorf("s3 api error: %w", err)
	}
	return res, nil
}

func (s *RepositoryS3) Upload(ctx context.Context, key string, r io.Reader) error {
	_, err := s.uploader.UploadWithContext(ctx, &s3manager.UploadInput{
		Bucket: s.bucket,
		Key:    aws.String(key),
		Body:   r,
	})
	if err != nil {
		return fmt.Errorf("s3 uploader failed: %w", err)
	}
	return nil
}

func (s *RepositoryS3) Delete(ctx context.Context, keys []string) error {
	ids := make([]*s3.ObjectIdentifier, len(keys))
	for i, k := range keys {
		ids[i] = &s3.ObjectIdentifier{
			Key: aws.String(k),
		}
	}

	_, err := s.api.DeleteObjects(&s3.DeleteObjectsInput{
		Delete: &s3.Delete{
			Objects: ids,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to delete objects: %w", err)
	}
	return nil
}
