package syncer

import (
	"bytes"
	"context"
	"fmt"
	"log"
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
	repoObjects, err := c.Repository.List(ctx)
	if err != nil {
		return fmt.Errorf("failed to list objects from repository: %w", err)
	}
	inRepo := map[string]RepositoryObject{}
	for _, obj := range repoObjects {
		log.Printf("Object in repo: %s", strings.TrimSuffix(obj.Key, ".tar"))
		inRepo[strings.TrimSuffix(obj.Key, ".tar")] = obj
	}

	localObjects, err := c.LocalStorage.List(ctx, in.Path, in.Depth)
	if err != nil {
		return fmt.Errorf("failed to list objects from local storage: %w", err)
	}

	eg, ctx := errgroup.WithContext(ctx)
	for _, v := range localObjects {
		localObj := v
		key, err := filepath.Rel(in.Path, localObj.Key)
		if err != nil {
			return fmt.Errorf("failed to get relative path of local object: %w", err)
		}

		log.Printf("Object in local: %s", key)
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

				log.Printf("Uploading: %s", localObj.Key)
				if err := c.Archiver.Do(ctx, localObj.Key, b); err != nil {
					return fmt.Errorf("failed to archive %q: %w", localObj.Key, err)
				}
				if err := c.Repository.Upload(ctx, key+".tar", b); err != nil {
					return fmt.Errorf("failed to upload %q to repository: %w", localObj.Key, err)
				}
				return nil
			})
		}
	}

	if err := eg.Wait(); err != nil {
		return fmt.Errorf("errgroup failed: %w", err)
	}

	if len(inRepo) > 0 {
		keys := make([]string, len(inRepo))
		for _, v := range inRepo {
			keys = append(keys, v.Key)
			log.Printf("Deleting: %s", v.Key)
		}
		if err := c.Repository.Delete(ctx, keys); err != nil {
			return fmt.Errorf("failed to delete objects: %w", err)
		}
	}

	return nil
}
