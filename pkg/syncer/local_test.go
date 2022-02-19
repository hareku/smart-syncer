package syncer_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

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

	writeFile := func(filename string) error {
		return os.WriteFile(filepath.Join(dir, filename), []byte(fmt.Sprintf("data for %s", filename)), 0655)
	}
	mkdir := func(name string) error {
		return os.Mkdir(filepath.Join(dir, name), 0655)
	}

	require.NoError(t, writeFile("abc"))

	require.NoError(t, mkdir("def"))
	require.NoError(t, writeFile("def/ghi1"))
	require.NoError(t, writeFile("def/ghi2"))

	require.NoError(t, mkdir("jkl"))
	require.NoError(t, mkdir("jkl/mno"))
	require.NoError(t, writeFile("jkl/mno/pqr"))
	require.NoError(t, writeFile("jkl/mno1"))

	s := syncer.NewLocalStorage()

	tests := []struct {
		depth    int
		wantKeys []string
	}{
		{
			depth:    1,
			wantKeys: []string{"abc", "def", "jkl"},
		},
		{
			depth:    2,
			wantKeys: []string{"abc", "def/ghi1", "def/ghi2", "jkl/mno", "jkl/mno1"},
		},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("depth: %d", tt.depth), func(t *testing.T) {
			got, err := s.List(context.Background(), dir, tt.depth)
			require.NoError(t, err)
			gotKeys := make([]string, len(got))
			for i, v := range got {
				gotKeys[i] = filepath.ToSlash(v.Key)
			}
			assert.Equal(t, tt.wantKeys, gotKeys)
		})
	}
}
