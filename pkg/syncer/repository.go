package syncer

import (
	"context"
	"io"
)

type RepositoryObject struct {
	Key              string
	LastModifiedUnix int64
}

type Repository interface {
	List(ctx context.Context) ([]RepositoryObject, error)
	Upload(ctx context.Context, key string, r io.Reader) error
	Delete(ctx context.Context, keys []string) error
}
