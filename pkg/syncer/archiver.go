package syncer

import (
	"archive/tar"
	"context"
	"fmt"
	"io"
	"io/fs"
	"log"
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
		log.Printf("%v: %s", d.IsDir(), path)
		if d.IsDir() {
			return nil
		}

		info, err := d.Info()
		if err != nil {
			return fmt.Errorf("failed to get info %q: %w", path, err)
		}
		h := &tar.Header{
			Name: path,
			Size: info.Size(),
		}
		err = tw.WriteHeader(h)
		if err != nil {
			return fmt.Errorf("failed to write tar header %+v: %w", h, err)
		}

		f, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read file %q: %w", path, err)
		}
		_, err = tw.Write(f)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	if err := tw.Close(); err != nil {
		return err
	}
	return nil
}
