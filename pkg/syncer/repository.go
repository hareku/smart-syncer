package syncer

import (
	"context"
	"io"
	"time"
)

type RepositoryObject struct {
	Key          string
	LastModified time.Time
}

type Repository interface {
	List(ctx context.Context) ([]RepositoryObject, error)
	Upload(ctx context.Context, key string, r io.Reader) error
}
