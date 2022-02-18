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

		// log.Printf("Object in local: %s", key)
		repoObj, ok := inRepo[key]
		if ok {
			delete(inRepo, key)
		}
		if ok && localObj.LastModified.Before(repoObj.LastModified) {
			continue
		}

		log.Printf("Uploading: %s", localObj.Key)
		if c.Dryrun {
			continue
		}

		pr, pw := io.Pipe()
		eg.Go(func() error {
			defer pw.Close()
			if err := c.Archiver.Do(ctx, localObj.Key, pw); err != nil {
				return fmt.Errorf("failed to archive %q: %w", localObj.Key, err)
			}
			return nil
		})
		eg.Go(func() error {
			if err := c.Repository.Upload(ctx, key+".tar", pr); err != nil {
				return fmt.Errorf("failed to upload %q to repository: %w", localObj.Key, err)
			}
			return nil
		})
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
		if !c.Dryrun {
			if err := c.Repository.Delete(ctx, keys); err != nil {
				return fmt.Errorf("failed to delete objects: %w", err)
			}
		}
	}

	return nil
}
