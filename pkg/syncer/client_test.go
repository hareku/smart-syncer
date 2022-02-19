package syncer_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/hareku/smart-syncer/pkg/syncer"
	"github.com/hareku/smart-syncer/pkg/syncer/syncermock"
	"github.com/stretchr/testify/require"
)

func TestClient_Run(t *testing.T) {
	ctrl := gomock.NewController(t)

	repo := syncermock.NewMockRepository(ctrl)
	repo.EXPECT().List(gomock.Any()).Times(1).Return([]syncer.RepositoryObject{
		{
			Key:              "obj1.tar",
			LastModifiedUnix: 1,
		},
		{
			Key:              "obj2.tar",
			LastModifiedUnix: 20,
		},
		{
			Key:              "obj4.tar",
			LastModifiedUnix: 40,
		},
	}, nil)
	repo.EXPECT().Upload(gomock.Any(), "obj1.tar", gomock.Any()).Times(1).Return(nil)
	repo.EXPECT().Upload(gomock.Any(), "new/obj3.tar", gomock.Any()).Times(1).Return(nil)
	repo.EXPECT().Delete(gomock.Any(), []string{"obj4.tar"}).Times(1).Return(nil)

	local := syncermock.NewMockLocalStorage(ctrl)
	local.EXPECT().List(gomock.Any(), "target", 1).Times(1).Return([]syncer.LocalObject{
		{
			Key:              "obj1",
			LastModifiedUnix: 10,
		},
		{
			Key:              "obj2",
			LastModifiedUnix: 20,
		},
		{
			Key:              "new/obj3",
			LastModifiedUnix: 30,
		},
	}, nil)

	arc := syncermock.NewMockArchiver(ctrl)
	arc.EXPECT().Do(gomock.Any(), filepath.Join("target/obj1"), gomock.Any()).Times(1).Return(nil)
	arc.EXPECT().Do(gomock.Any(), filepath.Join("target/new/obj3"), gomock.Any()).Times(1).Return(nil)

	c := &syncer.Client{
		LocalStorage: local,
		Repository:   repo,
		Archiver:     arc,
		Dryrun:       false,
		Concurrency:  1,
	}
	err := c.Run(context.Background(), &syncer.ClientRunInput{
		Path:  "target",
		Depth: 1,
	})
	require.NoError(t, err)
}
