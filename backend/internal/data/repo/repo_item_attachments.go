package repo

import (
	"context"
	"github.com/sysadminsmedia/homebox/backend/internal/core/services"
	"io"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/attachment"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/item"
)

// AttachmentRepo is a repository for Attachments table that links Items to Documents
// While also specifying the type of the attachment.
type AttachmentRepo struct {
	db *ent.Client
}

type (
	ItemAttachment struct {
		ID        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"createdAt"`
		UpdatedAt time.Time `json:"updatedAt"`
		Type      string    `json:"type"`
		Primary   bool      `json:"primary"`
		Path      string    `json:"path"`
		Title     string    `json:"title"`
	}

	ItemAttachmentUpdate struct {
		ID      uuid.UUID `json:"-"`
		Type    string    `json:"type"`
		Title   string    `json:"title"`
		Primary bool      `json:"primary"`
	}

	ItemCreateAttachment struct {
		Title   string    `json:"title"`
		Content io.Reader `json:"content"`
	}
)

func ToItemAttachment(attachment *ent.Attachment) ItemAttachment {
	return ItemAttachment{
		ID:        attachment.ID,
		CreatedAt: attachment.CreatedAt,
		UpdatedAt: attachment.UpdatedAt,
		Type:      attachment.Type.String(),
		Primary:   attachment.Primary,
		Path:      attachment.Path,
		Title:     attachment.Title,
	}
}

func (r *AttachmentRepo) Create(ctx context.Context, itemID uuid.UUID, doc ItemCreateAttachment, typ attachment.Type) (*ent.Attachment, error) {
	bldr := r.db.Attachment.Create().
		SetType(typ).
		SetItemID(itemID)

	// Autoset primary to true if this is the first attachment
	// that is of type photo
	if typ == attachment.TypePhoto {
		cnt, err := r.db.Attachment.Query().
			Where(
				attachment.HasItemWith(item.ID(itemID)),
				attachment.TypeEQ(typ),
			).
			Count(ctx)
		if err != nil {
			return nil, err
		}

		if cnt == 0 {
			bldr = bldr.SetPrimary(true)
		}
	}

	return bldr.Save(ctx)
}

func (r *AttachmentRepo) Get(ctx context.Context, id uuid.UUID) (*ent.Attachment, error) {
	return r.db.Attachment.
		Query().
		Where(attachment.ID(id)).
		WithItem().
		Only(ctx)
}

func (r *AttachmentRepo) Update(ctx context.Context, id uuid.UUID, data *ItemAttachmentUpdate) (*ent.Attachment, error) {
	// TODO: execute within Tx
	typ := attachment.Type(data.Type)

	bldr := r.db.Attachment.UpdateOneID(id).
		SetType(typ)

	// Primary only applies to photos
	if typ == attachment.TypePhoto {
		bldr = bldr.SetPrimary(data.Primary)
	} else {
		bldr = bldr.SetPrimary(false)
	}

	updatedAttachment, err := bldr.Save(ctx)
	if err != nil {
		return nil, err
	}

	attachmentItem, err := updatedAttachment.QueryItem().Only(ctx)
	if err != nil {
		return nil, err
	}

	// Ensure all other attachments are not primary
	err = r.db.Attachment.Update().
		Where(
			attachment.HasItemWith(item.ID(attachmentItem.ID)),
			attachment.IDNEQ(updatedAttachment.ID),
		).
		SetPrimary(false).
		Exec(ctx)
	if err != nil {
		return nil, err
	}

	return r.Get(ctx, updatedAttachment.ID)
}

func (r *AttachmentRepo) Delete(ctx context.Context, id uuid.UUID) error {
	doc, error := r.db.Attachment.Get(ctx, id)
	if error != nil {
		return error
	}

	err := os.Remove(doc.Path)
	if err != nil {
		return err
	}
	return r.db.Attachment.DeleteOneID(id).Exec(ctx)
}

func (r *AttachmentRepo) Rename(ctx services.Context, id uuid.UUID, title string) (interface{}, error) {
	return r.db.Attachment.UpdateOneID(id).SetTitle(title).Save(ctx)
}
