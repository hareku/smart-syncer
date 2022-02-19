package syncer_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/hareku/smart-syncer/pkg/syncer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLocalStorage_List(t *testing.T) {
	dir, err := os.MkdirTemp("", "localstorage-list-test-")
	require.NoError(t, err)
	t.Cleanup(func() {
		assert.NoError(t, os.RemoveAll(dir))
	})

	writeFile := func(filename string, modTime time.Time) error {
		if err := os.WriteFile(filepath.Join(dir, filename), []byte(fmt.Sprintf("data for %s", filename)), 0755); err != nil {
			return err
		}
		return os.Chtimes(filepath.Join(dir, filename), modTime, modTime)
	}
	mkdir := func(name string) error {
		return os.Mkdir(filepath.Join(dir, name), 0755)
	}

	require.NoError(t, writeFile("abc", time.Unix(100, 0)))
	require.NoError(t, mkdir("def"))
	require.NoError(t, writeFile("def/ghi", time.Unix(100, 0)))
	require.NoError(t, mkdir("jkl"))
	require.NoError(t, mkdir("jkl/mno"))
	require.NoError(t, writeFile("jkl/mno/pqr", time.Unix(100, 0)))

	s := syncer.NewLocalStorage()
	res, err := s.List(context.Background(), dir, 1)
	require.NoError(t, err)
	assert.Equal(t, []syncer.LocalObject{}, res)
}
