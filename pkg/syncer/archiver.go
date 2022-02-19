package syncer

import (
	"archive/tar"
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
)

//go:generate mockgen -destination ./${GOPACKAGE}mock/${GOFILE} -package ${GOPACKAGE}mock -source ./${GOFILE}

type Archiver interface {
	Do(ctx context.Context, root string, w io.Writer) error
}

func NewArchiver() Archiver {
	return &archiver{}
}

// pool for io.CopyBuffer
var bufPool = sync.Pool{
	New: func() interface{} {
		s := make([]byte, 128*1024)
		return &s
	},
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

		f, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("failed to read file %q: %w", path, err)
		}
		defer f.Close()

		info, err := f.Stat()
		if err != nil {
			return fmt.Errorf("failed to stat file %q: %w", path, err)
		}
		h, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return fmt.Errorf("failed to create tar header for %q: %w", path, err)
		}
		h.Name = filepath.ToSlash(rel)
		if err := tw.WriteHeader(h); err != nil {
			return fmt.Errorf("failed to write tar header %+v: %w", h, err)
		}

		buf := bufPool.Get().(*[]byte)
		defer bufPool.Put(buf)

		if _, err := io.CopyBuffer(tw, f, *buf); err != nil {
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
