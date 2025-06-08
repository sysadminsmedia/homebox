package repo

import (
	"bytes"
	"context"
	"crypto/md5"
	"fmt"
	"github.com/rs/zerolog/log"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/group"
	"github.com/sysadminsmedia/homebox/backend/internal/sys/config"
	"github.com/zeebo/blake3"
	"io"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/attachment"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/item"

	"gocloud.dev/blob"
	_ "gocloud.dev/blob/azureblob"
	_ "gocloud.dev/blob/fileblob"
	_ "gocloud.dev/blob/gcsblob"
	_ "gocloud.dev/blob/memblob"
	_ "gocloud.dev/blob/s3blob"
)

// AttachmentRepo is a repository for Attachments table that links Items to their
// associated files while also specifying the type of the attachment.
type AttachmentRepo struct {
	db      *ent.Client
	storage config.Storage
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

func (r *AttachmentRepo) path(gid uuid.UUID, hash string) string {
	return filepath.Join(r.storage.PrefixPath, gid.String(), "documents", hash)
}

func (r *AttachmentRepo) GetConnString() string {
	if strings.HasPrefix(r.storage.ConnString, "file:///./") {
		dir, err := filepath.Abs(strings.TrimPrefix(r.storage.ConnString, "file:///./"))
		if err != nil {
			log.Err(err).Msg("failed to get absolute path for attachment directory")
			return r.storage.ConnString
		}
		return fmt.Sprintf("file://%x?no_tmp_dir=true", dir+r.storage.PrefixPath)
	}
	return r.storage.ConnString + r.storage.PrefixPath
}

func (r *AttachmentRepo) Create(ctx context.Context, itemID uuid.UUID, doc ItemCreateAttachment, typ attachment.Type, primary bool) (*ent.Attachment, error) {
	tx, err := r.db.Tx(ctx)
	if err != nil {
		return nil, err
	}

	// If there is an error during file creation rollback the database
	defer func() {
		if v := recover(); v != nil {
			err := tx.Rollback()
			if err != nil {
				return
			}
		}
	}()

	bldrId := uuid.New()

	bldr := tx.Attachment.Create().
		SetID(bldrId).
		SetCreatedAt(time.Now()).
		SetUpdatedAt(time.Now()).
		SetType(typ).
		SetItemID(itemID).
		SetTitle(doc.Title)

	if typ == attachment.TypePhoto && primary {
		bldr = bldr.SetPrimary(true)
		err := r.db.Attachment.Update().
			Where(
				attachment.HasItemWith(item.ID(itemID)),
				attachment.IDNEQ(bldrId),
			).
			SetPrimary(false).
			Exec(ctx)
		if err != nil {
			log.Err(err).Msg("failed to remove primary from other attachments")
			err := tx.Rollback()
			if err != nil {
				return nil, err
			}
			return nil, err
		}
	} else if typ == attachment.TypePhoto {
		// Autoset primary to true if this is the first attachment
		// that is of type photo
		cnt, err := tx.Attachment.Query().
			Where(
				attachment.HasItemWith(item.ID(itemID)),
				attachment.TypeEQ(typ),
			).
			Count(ctx)
		if err != nil {
			log.Err(err).Msg("failed to count attachments")
			err := tx.Rollback()
			if err != nil {
				return nil, err
			}
			return nil, err
		}

		if cnt == 0 {
			bldr = bldr.SetPrimary(true)
		}
	}

	// Get the group ID for the item the attachment is being created for
	itemGroup, err := tx.Item.Query().QueryGroup().Where(group.HasItemsWith(item.ID(itemID))).First(ctx)
	if err != nil {
		log.Err(err).Msg("failed to get item group")
		err := tx.Rollback()
		if err != nil {
			return nil, err
		}
		return nil, err
	}

	// Prepare for the hashing of the file contents
	hashOut := make([]byte, 32)

	// Read all content into a buffer
	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, doc.Content)
	if err != nil {
		log.Err(err).Msg("failed to read file content")
		if rbErr := tx.Rollback(); rbErr != nil {
			return nil, rbErr
		}
		return nil, err
	}
	// Now the buffer contains all the data, use it for hashing
	contentBytes := buf.Bytes()

	// We use blake3 to generate a hash of the file contents, the group ID is used as context to ensure unique hashes
	// for the same file across different groups to reduce the chance of collisions
	// additionally, the hash can be used to validate the file contents if needed
	blake3.DeriveKey(itemGroup.ID.String(), contentBytes, hashOut)

	// Write the file to the blob storage bucket which might be a local file system or cloud storage
	bucket, err := blob.OpenBucket(ctx, r.GetConnString())
	if err != nil {
		log.Err(err).Msg("failed to open bucket")
		err := tx.Rollback()
		if err != nil {
			return nil, err
		}
		return nil, err
	}
	defer func(bucket *blob.Bucket) {
		err := bucket.Close()
		if err != nil {
			log.Err(err).Msg("failed to close bucket")
			err := tx.Rollback()
			if err != nil {
				log.Err(err).Msg("failed to rollback transaction after closing bucket")
			}
		}
	}(bucket)
	md5hash := md5.New()
	_, err = md5hash.Write(contentBytes)
	if err != nil {
		log.Err(err).Msg("failed to generate MD5 hash for storage")
		err = tx.Rollback()
		if err != nil {
			return nil, err
		}
		return nil, err
	}
	options := &blob.WriterOptions{
		ContentType:                 "application/octet-stream",
		DisableContentTypeDetection: false,
		ContentMD5:                  md5hash.Sum(nil),
	}
	path := r.path(itemGroup.ID, fmt.Sprintf("%x", hashOut))
	err = bucket.WriteAll(ctx, path, contentBytes, options)
	if err != nil {
		log.Err(err).Msg("failed to write file to bucket")
		err = tx.Rollback()
		if err != nil {
			return nil, err
		}
		return nil, err
	}

	bldr = bldr.SetPath(path)

	attachmentDb, err := bldr.Save(ctx)
	if err != nil {
		log.Err(err).Msg("failed to save attachment to database")
		err = tx.Rollback()
		if err != nil {
			return nil, err
		}
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		log.Err(err).Msg("failed to commit transaction")
		return nil, err
	}
	return attachmentDb, nil
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

	all, err := r.db.Attachment.Query().Where(attachment.Path(doc.Path)).All(ctx)
	if err != nil {
		return err
	}
	// If this is the last attachment for this path, delete the file
	if len(all) == 1 {
		bucket, err := blob.OpenBucket(ctx, r.GetConnString())
		if err != nil {
			log.Err(err).Msg("failed to open bucket")
			return err
		}
		defer func(bucket *blob.Bucket) {
			err := bucket.Close()
			if err != nil {
				log.Err(err).Msg("failed to close bucket")
			}
		}(bucket)
		err = bucket.Delete(ctx, doc.Path)
		if err != nil {
			return err
		}
	}

	return r.db.Attachment.DeleteOneID(id).Exec(ctx)
}

func (r *AttachmentRepo) Rename(ctx context.Context, id uuid.UUID, title string) (*ent.Attachment, error) {
	return r.db.Attachment.UpdateOneID(id).SetTitle(title).Save(ctx)
}
