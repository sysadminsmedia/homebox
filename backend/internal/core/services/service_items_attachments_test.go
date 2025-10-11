package services

import (
	"context"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/sysadminsmedia/homebox/backend/internal/data/repo"
	"github.com/sysadminsmedia/homebox/backend/internal/sys/config"
)

func TestItemService_AddAttachment(t *testing.T) {
	temp := os.TempDir()

	svc := &ItemService{
		repo:     tRepos,
		filepath: temp,
	}

	loc, err := tRepos.Locations.Create(context.Background(), tGroup.ID, repo.LocationCreate{
		Description: "test",
		Name:        "test",
	})
	require.NoError(t, err)
	assert.NotNil(t, loc)

	itmC := repo.ItemCreate{
		Name:        fk.Str(10),
		Description: fk.Str(10),
		LocationID:  loc.ID,
	}

	itm, err := svc.repo.Items.Create(context.Background(), tGroup.ID, itmC)
	require.NoError(t, err)
	assert.NotNil(t, itm)
	t.Cleanup(func() {
		err := svc.repo.Items.Delete(context.Background(), itm.ID)
		require.NoError(t, err)
	})

	contents := fk.Str(1000)
	reader := strings.NewReader(contents)

	// Setup
	afterAttachment, err := svc.AttachmentAdd(tCtx, itm.ID, "testfile.txt", "attachment", false, reader)
	require.NoError(t, err)
	assert.NotNil(t, afterAttachment)

	// Check that the file exists
	storedPath := afterAttachment.Attachments[0].Path

	// path should now be relative: {group}/{documents}
	assert.Equal(t, path.Join(tGroup.ID.String(), "documents"), path.Dir(storedPath))

	// Check that the file contents are correct
	bts, err := os.ReadFile(path.Join(os.TempDir(), storedPath))
	require.NoError(t, err)
	assert.Equal(t, contents, string(bts))
}

func TestItemService_AddAttachment_InvalidStorage(t *testing.T) {
	// Create a service with an invalid storage path to simulate the issue
	svc := &ItemService{
		repo:     tRepos,
		filepath: "/nonexistent/path/that/should/not/exist",
	}

	// Create a temporary repo with invalid storage config
	invalidRepos := repo.New(tClient, tbus, config.Storage{
		PrefixPath: "/",
		ConnString: "file:///nonexistent/directory/that/does/not/exist",
	}, "mem://{{ .Topic }}", config.Thumbnail{
		Enabled: false,
		Width:   0,
		Height:  0,
	})

	svc.repo = invalidRepos

	loc, err := invalidRepos.Locations.Create(context.Background(), tGroup.ID, repo.LocationCreate{
		Description: "test",
		Name:        "test-invalid",
	})
	require.NoError(t, err)
	assert.NotNil(t, loc)

	itmC := repo.ItemCreate{
		Name:        fk.Str(10),
		Description: fk.Str(10),
		LocationID:  loc.ID,
	}

	itm, err := invalidRepos.Items.Create(context.Background(), tGroup.ID, itmC)
	require.NoError(t, err)
	assert.NotNil(t, itm)
	t.Cleanup(func() {
		err := invalidRepos.Items.Delete(context.Background(), itm.ID)
		require.NoError(t, err)
	})

	contents := fk.Str(1000)
	reader := strings.NewReader(contents)

	// Attempt to add attachment with invalid storage - should return an error
	_, err = svc.AttachmentAdd(tCtx, itm.ID, "testfile.txt", "attachment", false, reader)
	
	// This should return an error now (after the fix)
	assert.Error(t, err, "AttachmentAdd should return an error when storage is invalid")
}
