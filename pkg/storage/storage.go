package storage

import (
	"context"
	"time"
)

type Object struct {
	Key          string
	LastModified time.Time
}

type Storage interface {
	List(ctx context.Context) ([]Object, error)
}
