package syncer

import (
	"context"
	"fmt"
	"io"
	"log"
	"path/filepath"
	"strings"

	"golang.org/x/sync/errgroup"
)

type Client struct {
	LocalStorage LocalStorage
	Repository   Repository
	Archiver     Archiver
	Dryrun       bool
	Concurrency  int
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
		inRepo[strings.TrimSuffix(obj.Key, ".tar")] = obj
	}

	localObjects, err := c.LocalStorage.List(ctx, in.Path, in.Depth)
	if err != nil {
		return fmt.Errorf("failed to list objects from local storage: %w", err)
	}

	queue := []LocalObject{}
	for _, v := range localObjects {
		localObj := v

		repoObj, ok := inRepo[localObj.Key]
		if ok {
			delete(inRepo, localObj.Key)
		}
		if ok && localObj.LastModifiedUnix <= repoObj.LastModifiedUnix {
			continue
		}

		queue = append(queue, localObj)
	}

	if len(queue) > 0 {
		eg, ctx := errgroup.WithContext(ctx)
		ch := make(chan LocalObject, c.Concurrency)

		eg.Go(func() error {
			defer close(ch)
			for i, v := range queue {
				select {
				case <-ctx.Done():
					return fmt.Errorf("uploading cancelled: %w", ctx.Err())
				case ch <- v:
					log.Printf("Uploading(%d/%d): %s", i+1, len(queue), v.Key)
				}
			}
			return nil
		})

		eg.Go(func() error {
			if err := c.upload(ctx, in.Path, ch); err != nil {
				return fmt.Errorf("uploading failed: %w", err)
			}
			return nil
		})

		if err := eg.Wait(); err != nil {
			return err
		}
	}

	if len(inRepo) > 0 {
		keys := make([]string, 0, len(inRepo))
		for _, v := range inRepo {
			keys = append(keys, v.Key)
		}
		for i, k := range keys {
			log.Printf("Deleting(%d/%d): %s", i+1, len(keys), k)
		}
		if !c.Dryrun {
			if err := c.Repository.Delete(ctx, keys); err != nil {
				return fmt.Errorf("failed to delete objects: %w", err)
			}
		}
	}

	return nil
}

func (c *Client) upload(ctx context.Context, root string, ch <-chan LocalObject) error {
	for localObj := range ch {
		if c.Dryrun {
			continue
		}

		pr, pw := io.Pipe()
		eg, ctx := errgroup.WithContext(ctx)
		eg.Go(func() error {
			defer pw.Close()
			if err := c.Archiver.Do(ctx, filepath.Join(root, localObj.Key), pw); err != nil {
				return fmt.Errorf("failed to archive %q: %w", localObj.Key, err)
			}
			return nil
		})
		eg.Go(func() error {
			if err := c.Repository.Upload(ctx, localObj.Key+".tar", pr); err != nil {
				return fmt.Errorf("failed to upload %q to repository: %w", localObj.Key, err)
			}
			return nil
		})

		if err := eg.Wait(); err != nil {
			return err
		}
	}
	return nil
}
