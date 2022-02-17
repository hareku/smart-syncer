package storage

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
)

type StorageS3 struct {
	api *s3.S3
}

func (s *StorageS3) List(ctx context.Context) ([]Object, error) {
	res := []Object{}

	err := s.api.ListObjectsV2PagesWithContext(ctx, &s3.ListObjectsV2Input{
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
		return nil, fmt.Errorf("s3 api error: %w", err)
	}
	return res, nil
}
