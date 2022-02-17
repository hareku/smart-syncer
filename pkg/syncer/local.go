package syncer

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type LocalObject struct {
	Key          string
	LastModified time.Time
}

type LocalStorage interface {
	List(ctx context.Context, path string, depth int) ([]LocalObject, error)
}

func NewLocalStorage() LocalStorage {
	return &localStorage{}
}

type localStorage struct {
	res []LocalObject
}

func (s *localStorage) List(ctx context.Context, path string, depth int) ([]LocalObject, error) {
	s.res = []LocalObject{}

	if err := s.recursive(ctx, path, depth, 1); err != nil {
		return nil, fmt.Errorf("recursive walk failed: %w", err)
	}
	return s.res, nil
}

func (s *localStorage) recursive(ctx context.Context, path string, depth int, currentDepth int) error {
	entries, err := os.ReadDir(path)
	if err != nil {
		return fmt.Errorf("failed to read %q as a directory: %w", path, err)
	}
	for _, e := range entries {
		if e.IsDir() && currentDepth < depth {
			nextPath := filepath.Join(path, e.Name())
			if err := s.recursive(ctx, nextPath, depth, currentDepth+1); err != nil {
				return fmt.Errorf("error in path %q depth %d: %w", nextPath, currentDepth+1, err)
			}
			continue
		}

		s.res = append(s.res, LocalObject{
			Key: e.Name(),
		})
	}
	return nil
}
