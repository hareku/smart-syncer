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

	if err := s.recurse(ctx, path, depth, 1); err != nil {
		return nil, fmt.Errorf("recursive walking failed: %w", err)
	}
	return s.res, nil
}

func (s *localStorage) recurse(ctx context.Context, path string, depth int, currentDepth int) error {
	entries, err := os.ReadDir(path)
	if err != nil {
		return fmt.Errorf("failed to read %q as a directory: %w", path, err)
	}
	for _, e := range entries {
		key := filepath.Join(path, e.Name())
		if e.IsDir() && currentDepth < depth {
			if err := s.recurse(ctx, key, depth, currentDepth+1); err != nil {
				return fmt.Errorf("error in path %q depth %d: %w", key, currentDepth+1, err)
			}
			continue
		}

		info, err := e.Info()
		if err != nil {
			return fmt.Errorf("failed to get info of %q: %w", key, err)
		}

		s.res = append(s.res, LocalObject{
			Key:          key,
			LastModified: info.ModTime(),
		})
	}
	return nil
}
