package syncer

import (
	"archive/tar"
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

type Archiver interface {
	Do(ctx context.Context, root string, w io.Writer) error
}

func NewArchiver() Archiver {
	return &archiver{}
}

type archiver struct{}

func (a *archiver) Do(ctx context.Context, root string, w io.Writer) error {
	tw := tar.NewWriter(w)

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		rel, err := filepath.Rel(root, path)
		if err != nil {
			return fmt.Errorf("failed to get relative path: %w", err)
		}

		// if rel is ".", root is file (not dir).
		if rel == "." {
			rel = d.Name()
		}

		f, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read file %q: %w", path, err)
		}

		h := &tar.Header{
			Name: rel,
			Size: int64(len(f)),
		}
		if err := tw.WriteHeader(h); err != nil {
			return fmt.Errorf("failed to write tar header %+v: %w", h, err)
		}

		if _, err := tw.Write(f); err != nil {
			return fmt.Errorf("failed to write tar content: %w", err)
		}
		return nil
	})
	if err != nil {
		return err
	}
	if err := tw.Close(); err != nil {
		return fmt.Errorf("failed to close tar writer: %w", err)
	}
	return nil
}
