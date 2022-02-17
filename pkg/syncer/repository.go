package syncer

import (
	"context"
	"time"
)

type RepositoryObject struct {
	Key          string
	LastModified time.Time
}

type Repository interface {
	List(ctx context.Context) ([]RepositoryObject, error)
}
