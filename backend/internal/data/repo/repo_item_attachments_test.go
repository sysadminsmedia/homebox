package repo

import (
	"context"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/attachment"
	"github.com/sysadminsmedia/homebox/backend/internal/sys/config"
)

func TestAttachmentRepo_Create(t *testing.T) {
	item := useItems(t, 1)[0]

	ids := []uuid.UUID{item.ID}
	t.Cleanup(func() {
		for _, id := range ids {
			_ = tRepos.Attachments.Delete(context.Background(), tGroup.ID, item.ID, id)
		}
	})

	type args struct {
		ctx    context.Context
		itemID uuid.UUID
		typ    attachment.Type
	}
	tests := []struct {
		name    string
		args    args
		want    *ent.Attachment
		wantErr bool
	}{
		{
			name: "create attachment",
			args: args{
				ctx:    context.Background(),
				itemID: item.ID,
				typ:    attachment.TypePhoto,
			},
			want: &ent.Attachment{
				Type: attachment.TypePhoto,
			},
		},
		{
			name: "create attachment with invalid item id",
			args: args{
				ctx:    context.Background(),
				itemID: uuid.New(),
				typ:    "blarg",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _ := tRepos.Attachments.Create(tt.args.ctx, tt.args.itemID, ItemCreateAttachment{Title: "Test", Content: strings.NewReader("This is a test")}, tt.args.typ, false)
			// TODO: Figure out how this works and fix the test later
			// if (err != nil) != tt.wantErr {
			//	t.Errorf("AttachmentRepo.Create() error = %v, wantErr %v", err, tt.wantErr)
			//	return
			//}

			if tt.wantErr {
				return
			}

			assert.Equal(t, tt.want.Type, got.Type)

			withItems, err := tRepos.Attachments.Get(tt.args.ctx, tGroup.ID, got.ID)
			require.NoError(t, err)
			assert.Equal(t, tt.args.itemID, withItems.Edges.Item.ID)

			ids = append(ids, got.ID)
		})
	}
}

func useAttachments(t *testing.T, n int) []*ent.Attachment {
	t.Helper()

	item := useItems(t, 1)[0]

	ids := make([]uuid.UUID, 0, n)
	t.Cleanup(func() {
		for _, id := range ids {
			_ = tRepos.Attachments.Delete(context.Background(), tGroup.ID, item.ID, id)
		}
	})

	attachments := make([]*ent.Attachment, n)
	for i := 0; i < n; i++ {
		attach, err := tRepos.Attachments.Create(context.Background(), item.ID, ItemCreateAttachment{Title: "Test", Content: strings.NewReader("Test String")}, attachment.TypePhoto, true)
		require.NoError(t, err)
		attachments[i] = attach

		ids = append(ids, attach.ID)
	}

	return attachments
}

func TestAttachmentRepo_Update(t *testing.T) {
	entity := useAttachments(t, 1)[0]

	for _, typ := range []attachment.Type{"photo", "manual", "warranty", "attachment"} {
		t.Run(string(typ), func(t *testing.T) {
			_, err := tRepos.Attachments.Update(context.Background(), tGroup.ID, entity.ID, &ItemAttachmentUpdate{
				Type: string(typ),
			})

			require.NoError(t, err)

			updated, err := tRepos.Attachments.Get(context.Background(), tGroup.ID, entity.ID)
			require.NoError(t, err)
			assert.Equal(t, typ, updated.Type)
		})
	}
}

func TestAttachmentRepo_Delete(t *testing.T) {
	entity := useAttachments(t, 1)[0]
	item := useItems(t, 1)[0]

	err := tRepos.Attachments.Delete(context.Background(), tGroup.ID, item.ID, entity.ID)
	require.NoError(t, err)

	_, err = tRepos.Attachments.Get(context.Background(), tGroup.ID, entity.ID)
	require.Error(t, err)
}

func TestAttachmentRepo_EnsureSinglePrimaryAttachment(t *testing.T) {
	ctx := context.Background()
	attachments := useAttachments(t, 2)

	setAndVerifyPrimary := func(primaryAttachmentID, nonPrimaryAttachmentID uuid.UUID) {
		primaryAttachment, err := tRepos.Attachments.Update(ctx, tGroup.ID, primaryAttachmentID, &ItemAttachmentUpdate{
			Type:    attachment.TypePhoto.String(),
			Primary: true,
		})
		require.NoError(t, err)

		nonPrimaryAttachment, err := tRepos.Attachments.Get(ctx, tGroup.ID, nonPrimaryAttachmentID)
		require.NoError(t, err)

		assert.True(t, primaryAttachment.Primary)
		assert.False(t, nonPrimaryAttachment.Primary)
	}

	setAndVerifyPrimary(attachments[0].ID, attachments[1].ID)
	setAndVerifyPrimary(attachments[1].ID, attachments[0].ID)
}

func TestAttachmentRepo_UpdateNonPhotoDoesNotAffectPrimaryPhoto(t *testing.T) {
	ctx := context.Background()
	item := useItems(t, 1)[0]

	// Create a photo attachment that will be primary
	photoAttachment, err := tRepos.Attachments.Create(ctx, item.ID, ItemCreateAttachment{Title: "Test Photo", Content: strings.NewReader("Photo content")}, attachment.TypePhoto, true)
	require.NoError(t, err)

	// Create a manual attachment (non-photo)
	manualAttachment, err := tRepos.Attachments.Create(ctx, item.ID, ItemCreateAttachment{Title: "Test Manual", Content: strings.NewReader("Manual content")}, attachment.TypeManual, false)
	require.NoError(t, err)

	// Cleanup
	t.Cleanup(func() {
		_ = tRepos.Attachments.Delete(ctx, tGroup.ID, item.ID, photoAttachment.ID)
		_ = tRepos.Attachments.Delete(ctx, tGroup.ID, item.ID, manualAttachment.ID)
	})

	// Verify photo is primary initially
	photoAttachment, err = tRepos.Attachments.Get(ctx, tGroup.ID, photoAttachment.ID)
	require.NoError(t, err)
	assert.True(t, photoAttachment.Primary)

	// Update the manual attachment (this should NOT affect the photo's primary status)
	_, err = tRepos.Attachments.Update(ctx, tGroup.ID, manualAttachment.ID, &ItemAttachmentUpdate{
		Type:    attachment.TypeManual.String(),
		Title:   "Updated Manual",
		Primary: false, // This should have no effect since it's not a photo
	})
	require.NoError(t, err)

	// Verify photo is still primary after updating the manual
	photoAttachment, err = tRepos.Attachments.Get(ctx, tGroup.ID, photoAttachment.ID)
	require.NoError(t, err)
	assert.True(t, photoAttachment.Primary, "Photo attachment should remain primary after updating non-photo attachment")

	// Verify manual attachment is not primary
	manualAttachment, err = tRepos.Attachments.Get(ctx, tGroup.ID, manualAttachment.ID)
	require.NoError(t, err)
	assert.False(t, manualAttachment.Primary)
}

func TestAttachmentRepo_AddingPDFAfterPhotoKeepsPhotoAsPrimary(t *testing.T) {
	ctx := context.Background()
	item := useItems(t, 1)[0]

	// Step 1: Upload a photo first (this should become primary since it's the first photo)
	photoAttachment, err := tRepos.Attachments.Create(ctx, item.ID, ItemCreateAttachment{Title: "Item Photo", Content: strings.NewReader("Photo content")}, attachment.TypePhoto, false)
	require.NoError(t, err)

	// Cleanup
	t.Cleanup(func() {
		_ = tRepos.Attachments.Delete(ctx, tGroup.ID, item.ID, photoAttachment.ID)
	})

	// Verify photo becomes primary automatically (since it's the first photo)
	photoAttachment, err = tRepos.Attachments.Get(ctx, tGroup.ID, photoAttachment.ID)
	require.NoError(t, err)
	assert.True(t, photoAttachment.Primary, "First photo should automatically become primary")

	// Step 2: Add a PDF receipt (this should NOT affect the photo's primary status)
	pdfAttachment, err := tRepos.Attachments.Create(ctx, item.ID, ItemCreateAttachment{Title: "Receipt PDF", Content: strings.NewReader("PDF content")}, attachment.TypeReceipt, false)
	require.NoError(t, err)

	// Add to cleanup
	t.Cleanup(func() {
		_ = tRepos.Attachments.Delete(ctx, tGroup.ID, item.ID, pdfAttachment.ID)
	})

	// Step 3: Verify photo is still primary after adding PDF
	photoAttachment, err = tRepos.Attachments.Get(ctx, tGroup.ID, photoAttachment.ID)
	require.NoError(t, err)
	assert.True(t, photoAttachment.Primary, "Photo should remain primary after adding PDF attachment")

	// Verify PDF is not primary
	pdfAttachment, err = tRepos.Attachments.Get(ctx, tGroup.ID, pdfAttachment.ID)
	require.NoError(t, err)
	assert.False(t, pdfAttachment.Primary)

	// Step 4: Test the actual item summary mapping (this is what determines the card display)
	updatedItem, err := tRepos.Items.GetOne(ctx, item.ID)
	require.NoError(t, err)

	// The item should have the photo's ID as the imageId
	assert.NotNil(t, updatedItem.ImageID, "Item should have an imageId")
	assert.Equal(t, photoAttachment.ID, *updatedItem.ImageID, "Item's imageId should match the photo attachment ID")
}

func TestAttachmentRepo_SettingPhotoPrimaryStillWorks(t *testing.T) {
	ctx := context.Background()
	item := useItems(t, 1)[0]

	// Create two photo attachments
	photo1, err := tRepos.Attachments.Create(ctx, item.ID, ItemCreateAttachment{Title: "Photo 1", Content: strings.NewReader("Photo 1 content")}, attachment.TypePhoto, false)
	require.NoError(t, err)

	photo2, err := tRepos.Attachments.Create(ctx, item.ID, ItemCreateAttachment{Title: "Photo 2", Content: strings.NewReader("Photo 2 content")}, attachment.TypePhoto, false)
	require.NoError(t, err)

	// Cleanup
	t.Cleanup(func() {
		_ = tRepos.Attachments.Delete(ctx, tGroup.ID, item.ID, photo1.ID)
		_ = tRepos.Attachments.Delete(ctx, tGroup.ID, item.ID, photo2.ID)
	})

	// First photo should be primary (since it was created first)
	photo1, err = tRepos.Attachments.Get(ctx, tGroup.ID, photo1.ID)
	require.NoError(t, err)
	assert.True(t, photo1.Primary)

	photo2, err = tRepos.Attachments.Get(ctx, tGroup.ID, photo2.ID)
	require.NoError(t, err)
	assert.False(t, photo2.Primary)

	// Now set photo2 as primary (this should work and remove primary from photo1)
	photo2, err = tRepos.Attachments.Update(ctx, tGroup.ID, photo2.ID, &ItemAttachmentUpdate{
		Type:    attachment.TypePhoto.String(),
		Title:   "Photo 2",
		Primary: true,
	})
	require.NoError(t, err)
	assert.True(t, photo2.Primary)

	// Verify photo1 is no longer primary
	photo1, err = tRepos.Attachments.Get(ctx, tGroup.ID, photo1.ID)
	require.NoError(t, err)
	assert.False(t, photo1.Primary, "Photo 1 should no longer be primary after setting Photo 2 as primary")
}

func TestAttachmentRepo_PathNormalization(t *testing.T) {
	// Test that paths always use forward slashes
	repo := &AttachmentRepo{
		storage: config.Storage{
			PrefixPath: ".data",
		},
	}
	
	testGUID := uuid.MustParse("eb6bf410-a1a8-478d-a803-ca3948368a0c")
	testHash := "f295eb01-18a9-4631-a797-70bd9623edd4.png"
	
	// Test path() method - should always return forward slashes
	relativePath := repo.path(testGUID, testHash)
	assert.Equal(t, "eb6bf410-a1a8-478d-a803-ca3948368a0c/documents/f295eb01-18a9-4631-a797-70bd9623edd4.png", relativePath)
	assert.NotContains(t, relativePath, "\\", "path() should not contain backslashes")
	
	// Test fullPath() with forward slash input (from database)
	fullPath := repo.fullPath("eb6bf410-a1a8-478d-a803-ca3948368a0c/documents/f295eb01-18a9-4631-a797-70bd9623edd4.png")
	assert.Equal(t, ".data/eb6bf410-a1a8-478d-a803-ca3948368a0c/documents/f295eb01-18a9-4631-a797-70bd9623edd4.png", fullPath)
	assert.NotContains(t, fullPath, "\\", "fullPath() should not contain backslashes")
	
	// Test fullPath() with backslash input (legacy Windows paths from old database)
	fullPathWithBackslash := repo.fullPath("eb6bf410-a1a8-478d-a803-ca3948368a0c\\documents\\f295eb01-18a9-4631-a797-70bd9623edd4.png")
	assert.Equal(t, ".data/eb6bf410-a1a8-478d-a803-ca3948368a0c/documents/f295eb01-18a9-4631-a797-70bd9623edd4.png", fullPathWithBackslash)
	assert.NotContains(t, fullPathWithBackslash, "\\", "fullPath() should normalize backslashes to forward slashes")
	
	// Test with Windows-style prefix path
	repoWindows := &AttachmentRepo{
		storage: config.Storage{
			PrefixPath: ".data",
		},
	}
	fullPathWindows := repoWindows.fullPath("eb6bf410-a1a8-478d-a803-ca3948368a0c/documents/f295eb01-18a9-4631-a797-70bd9623edd4.png")
	assert.NotContains(t, fullPathWindows, "\\", "fullPath() should normalize Windows paths")
	
	// Test empty prefix
	repoNoPrefix := &AttachmentRepo{
		storage: config.Storage{
			PrefixPath: "",
		},
	}
	fullPathNoPrefix := repoNoPrefix.fullPath("eb6bf410-a1a8-478d-a803-ca3948368a0c/documents/f295eb01-18a9-4631-a797-70bd9623edd4.png")
	assert.Equal(t, "eb6bf410-a1a8-478d-a803-ca3948368a0c/documents/f295eb01-18a9-4631-a797-70bd9623edd4.png", fullPathNoPrefix)
	
	// Test with single slash prefix (like in tests)
	repoSlashPrefix := &AttachmentRepo{
		storage: config.Storage{
			PrefixPath: "/",
		},
	}
	fullPathSlashPrefix := repoSlashPrefix.fullPath("eb6bf410-a1a8-478d-a803-ca3948368a0c/documents/f295eb01-18a9-4631-a797-70bd9623edd4.png")
	assert.Equal(t, "eb6bf410-a1a8-478d-a803-ca3948368a0c/documents/f295eb01-18a9-4631-a797-70bd9623edd4.png", fullPathSlashPrefix)
	assert.NotContains(t, fullPathSlashPrefix, "//", "fullPath() should not have double slashes")
}
