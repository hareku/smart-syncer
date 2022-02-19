package syncer_test

import (
	"archive/tar"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/hareku/smart-syncer/pkg/syncer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/mod/sumdb/dirhash"
)

func TestArchiver_Do(t *testing.T) {
	a := syncer.NewArchiver()

	targetDir, err := os.MkdirTemp("", "archiver-do-test-target-")
	require.NoError(t, err)
	t.Cleanup(func() {
		assert.NoError(t, os.RemoveAll(targetDir))
	})

	require.NoError(t, os.WriteFile(filepath.Join(targetDir, "abc"), []byte("data for abc"), 0777))
	require.NoError(t, os.Mkdir(filepath.Join(targetDir, "def"), 0777))
	require.NoError(t, os.WriteFile(filepath.Join(targetDir, "def/ghi"), []byte("data for ghi"), 0777))

	tempDir, err := os.MkdirTemp("", "archiver-do-test-temp")
	require.NoError(t, err)
	t.Cleanup(func() {
		assert.NoError(t, os.RemoveAll(tempDir))
	})

	w, err := os.Create(filepath.Join(tempDir, "out.tar"))
	require.NoError(t, err)
	require.NoError(t, a.Do(context.Background(), targetDir, w))
	require.NoError(t, w.Close())

	out, err := os.Open(filepath.Join(tempDir, "out.tar"))
	require.NoError(t, err)
	t.Cleanup(func() {
		assert.NoError(t, out.Close())
	})

	expected, err := dirhash.HashDir(targetDir, "", dirhash.Hash1)
	require.NoError(t, err)
	got, err := hashTar(out)
	require.NoError(t, err)
	assert.Equal(t, expected, got)
}

func hashTar(tarReader io.Reader) (string, error) {
	tr := tar.NewReader(tarReader)

	bodies := make(map[string]*[]byte)
	var files []string
	for {
		h, err := tr.Next()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return "", err
		}
		files = append(files, h.Name)
		b, err := io.ReadAll(tr)
		if err != nil {
			return "", err
		}
		bodies[h.Name] = &b
	}

	open := func(name string) (io.ReadCloser, error) {
		b := bodies[name]
		if b == nil {
			return nil, fmt.Errorf("file %q not found in tar", name) // should never happen
		}
		return io.NopCloser(bytes.NewReader(*b)), nil
	}
	return dirhash.Hash1(files, open)
}
