package repo

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/attachment"
)

func TestAttachmentRepo_Create(t *testing.T) {
	doc := useDocs(t, 1)[0]
	item := useItems(t, 1)[0]

	ids := []uuid.UUID{doc.ID, item.ID}
	t.Cleanup(func() {
		for _, id := range ids {
			_ = tRepos.Attachments.Delete(context.Background(), id)
		}
	})

	type args struct {
		ctx    context.Context
		itemID uuid.UUID
		docID  uuid.UUID
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
				docID:  doc.ID,
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
				docID:  doc.ID,
				typ:    "blarg",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tRepos.Attachments.Create(tt.args.ctx, tt.args.itemID, tt.args.docID, tt.args.typ)
			if (err != nil) != tt.wantErr {
				t.Errorf("AttachmentRepo.Create() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			assert.Equal(t, tt.want.Type, got.Type)

			withItems, err := tRepos.Attachments.Get(tt.args.ctx, got.ID)
			require.NoError(t, err)
			assert.Equal(t, tt.args.itemID, withItems.Edges.Item.ID)
			assert.Equal(t, tt.args.docID, withItems.Edges.Document.ID)

			ids = append(ids, got.ID)
		})
	}
}

func useAttachments(t *testing.T, n int) []*ent.Attachment {
	t.Helper()

	doc := useDocs(t, 1)[0]
	item := useItems(t, 1)[0]

	ids := make([]uuid.UUID, 0, n)
	t.Cleanup(func() {
		for _, id := range ids {
			_ = tRepos.Attachments.Delete(context.Background(), id)
		}
	})

	attachments := make([]*ent.Attachment, n)
	for i := 0; i < n; i++ {
		attachment, err := tRepos.Attachments.Create(context.Background(), item.ID, doc.ID, attachment.TypePhoto)
		require.NoError(t, err)
		attachments[i] = attachment

		ids = append(ids, attachment.ID)
	}

	return attachments
}

func TestAttachmentRepo_Update(t *testing.T) {
	entity := useAttachments(t, 1)[0]

	for _, typ := range []attachment.Type{"photo", "manual", "warranty", "attachment"} {
		t.Run(string(typ), func(t *testing.T) {
			_, err := tRepos.Attachments.Update(context.Background(), entity.ID, &ItemAttachmentUpdate{
				Type: string(typ),
			})

			require.NoError(t, err)

			updated, err := tRepos.Attachments.Get(context.Background(), entity.ID)
			require.NoError(t, err)
			assert.Equal(t, typ, updated.Type)
		})
	}
}

func TestAttachmentRepo_Delete(t *testing.T) {
	entity := useAttachments(t, 1)[0]

	err := tRepos.Attachments.Delete(context.Background(), entity.ID)
	require.NoError(t, err)

	_, err = tRepos.Attachments.Get(context.Background(), entity.ID)
	require.Error(t, err)
}

func TestAttachmentRepo_EnsureSinglePrimaryAttachment(t *testing.T) {
	ctx := context.Background()
	attachments := useAttachments(t, 2)

	setAndVerifyPrimary := func(primaryAttachmentID, nonPrimaryAttachmentID uuid.UUID) {
		primaryAttachment, err := tRepos.Attachments.Update(ctx, primaryAttachmentID, &ItemAttachmentUpdate{
			Type:    attachment.TypePhoto.String(),
			Primary: true,
		})
		require.NoError(t, err)

		nonPrimaryAttachment, err := tRepos.Attachments.Get(ctx, nonPrimaryAttachmentID)
		require.NoError(t, err)

		assert.True(t, primaryAttachment.Primary)
		assert.False(t, nonPrimaryAttachment.Primary)
	}

	setAndVerifyPrimary(attachments[0].ID, attachments[1].ID)
	setAndVerifyPrimary(attachments[1].ID, attachments[0].ID)
}
