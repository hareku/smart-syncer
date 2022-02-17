package syncer

import (
	"bytes"
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"sync"

	"golang.org/x/sync/errgroup"
)

var bufPool = sync.Pool{
	New: func() interface{} {
		return new(bytes.Buffer)
	},
}

type Client struct {
	LocalStorage LocalStorage
	Repository   Repository
	Archiver     Archiver
}

type ClientRunInput struct {
	Path  string
	Depth int
}

func (c *Client) Run(ctx context.Context, in *ClientRunInput) error {
	objects, err := c.Repository.List(ctx)
	if err != nil {
		return err
	}
	inRepo := map[string]RepositoryObject{}
	for _, obj := range objects {
		fmt.Printf("in repo: %s\n", strings.TrimSuffix(obj.Key, ".tar"))
		inRepo[strings.TrimSuffix(obj.Key, ".tar")] = obj
	}

	localObjects, err := c.LocalStorage.List(ctx, in.Path, in.Depth)
	if err != nil {
		return err
	}

	eg, ctx := errgroup.WithContext(ctx)
	for _, v := range localObjects {
		localObj := v
		key, err := filepath.Rel(in.Path, localObj.Key)
		if err != nil {
			return fmt.Errorf("failed to get relative path: %w", err)
		}

		fmt.Printf("local obj key: %+v\n", key)
		repoObj, ok := inRepo[key]
		if ok {
			delete(inRepo, key)
		}

		if !ok || repoObj.LastModified.Before(localObj.LastModified) {
			eg.Go(func() error {
				b := bufPool.Get().(*bytes.Buffer)
				defer func() {
					b.Reset()
					bufPool.Put(b)
				}()
				fmt.Printf("uploading: %+v\n", localObj.Key+".tar")

				if err := c.Archiver.Do(ctx, localObj.Key, b); err != nil {
					return fmt.Errorf("failed to archive %q: %w", localObj.Key, err)
				}
				if err := c.Repository.Upload(ctx, key+".tar", b); err != nil {
					return fmt.Errorf("failed to upload object %q: %w", localObj.Key, err)
				}
				return nil
			})
		}
	}

	if err := eg.Wait(); err != nil {
		return err
	}
	return nil
}
