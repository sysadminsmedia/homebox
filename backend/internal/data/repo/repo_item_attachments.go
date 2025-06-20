package repo

import (
	"bytes"
	"context"
	"crypto/md5"
	"fmt"
	"github.com/gen2brain/go-fitz"
	"github.com/rs/zerolog/log"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/group"
	"github.com/sysadminsmedia/homebox/backend/internal/sys/config"
	"github.com/sysadminsmedia/homebox/backend/pkgs/utils"
	"github.com/zeebo/blake3"
	"gocloud.dev/pubsub"
	"image"
	"image/png"
	"io"
	"io/fs"
	"net/http"
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
	db         *ent.Client
	storage    config.Storage
	pubSubConn string
	thumbnail  config.Thumbnail
}

type (
	ItemAttachment struct {
		ID        uuid.UUID       `json:"id"`
		CreatedAt time.Time       `json:"createdAt"`
		UpdatedAt time.Time       `json:"updatedAt"`
		Type      string          `json:"type"`
		Primary   bool            `json:"primary"`
		Path      string          `json:"path"`
		Title     string          `json:"title"`
		Thumbnail *ent.Attachment `json:"thumbnail,omitempty"`
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
		return fmt.Sprintf("file://%s?no_tmp_dir=true", dir)
	}
	return r.storage.ConnString
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

	// Upload the file to the storage bucket
	path, err := r.UploadFile(ctx, itemGroup, doc)
	if err != nil {
		err := tx.Rollback()
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

	if r.thumbnail.Enabled {
		topic, err := pubsub.OpenTopic(ctx, utils.GenerateSubPubConn(r.pubSubConn, "thumbnails"))
		if err != nil {
			log.Err(err).Msg("failed to open pubsub topic")
			return nil, err
		}
		defer func(topic *pubsub.Topic, ctx context.Context) {
			err := topic.Shutdown(ctx)
			if err != nil {
				log.Err(err).Msg("failtask ed to shutdown pubsub topic")
			}
		}(topic, ctx)

		err = topic.Send(ctx, &pubsub.Message{
			Body: []byte(fmt.Sprintf("attachment_created:%s", attachmentDb.ID.String())),
			Metadata: map[string]string{
				"attachment_id": attachmentDb.ID.String(),
				"title":         doc.Title,
				"path":          attachmentDb.Path,
			},
		})
		if err != nil {
			log.Err(err).Msg("failed to send message to topic")
			return nil, err
		}
	}

	return attachmentDb, nil
}

func (r *AttachmentRepo) Get(ctx context.Context, id uuid.UUID) (*ent.Attachment, error) {
	return r.db.Attachment.
		Query().
		Where(attachment.ID(id)).
		WithItem().
		WithThumbnail().
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

func (r *AttachmentRepo) CreateThumbnail(ctx context.Context, groupId, attachmentId uuid.UUID, title string, path string) error {
	tx, err := r.db.Tx(ctx)
	if err != nil {
		return nil
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

	att := tx.Attachment.Create().
		SetID(uuid.New()).
		SetOriginalID(attachmentId).
		SetTitle(fmt.Sprintf("%s-thumb", title)).
		SetType("thumbnail")
	orig := tx.Attachment.GetX(ctx, attachmentId)

	bucket, err := blob.OpenBucket(ctx, r.GetConnString())
	if err != nil {
		log.Err(err).Msg("failed to open bucket")
		err := tx.Rollback()
		if err != nil {
			return err
		}
		return err
	}
	defer func(bucket *blob.Bucket) {
		err := bucket.Close()
		if err != nil {
			err := tx.Rollback()
			if err != nil {
				return
			}
			log.Err(err).Msg("failed to close bucket")
		}
	}(bucket)

	origFile, err := bucket.Open(path)
	if err != nil {
		err := tx.Rollback()
		if err != nil {
			return err
		}
		return err
	}
	defer func(file fs.File) {
		err := file.Close()
		if err != nil {
			err := tx.Rollback()
			if err != nil {
				return
			}
			log.Err(err).Msg("failed to close file")
		}
	}(origFile)

	if isImageFile(title) {

		img, _, err := image.Decode(origFile)
		if err != nil {
			err := tx.Rollback()
			if err != nil {
				return err
			}
			return err
		}

		bounds := img.Bounds()
		cropRect := image.Rect(
			bounds.Min.X,
			bounds.Min.Y,
			bounds.Min.X+r.thumbnail.Width,
			bounds.Min.Y+r.thumbnail.Height,
		)
		croppedImg := img.(interface {
			SubImage(r image.Rectangle) image.Image
		}).SubImage(cropRect)

		buf := new(bytes.Buffer)
		err = png.Encode(buf, croppedImg)
		if err != nil {
			err := tx.Rollback()
			if err != nil {
				return err
			}
			return err
		}

		contentBytes := buf.Bytes()
		thumbnailFile, err := r.UploadFile(ctx, tx.Group.GetX(ctx, groupId), ItemCreateAttachment{
			Title:   fmt.Sprintf("%s-thumb", title),
			Content: bytes.NewReader(contentBytes),
		})
		if err != nil {
			log.Err(err).Msg("failed to upload thumbnail file")
			err := tx.Rollback()
			if err != nil {
				return err
			}
			return err
		}

		att.SetPath(thumbnailFile)
	} else if isDocumentFile(title) && r.thumbnail.NonImageEnabled {
		fitz.FzVersion = r.thumbnail.MuPDFVersion
		doc, err := fitz.NewFromReader(origFile)
		if err != nil {
			err := tx.Rollback()
			if err != nil {
				return err
			}
			return err
		}
		defer func(doc *fitz.Document) {
			err := doc.Close()
			if err != nil {
				err := tx.Rollback()
				if err != nil {
					return
				}
				log.Err(err).Msg("failed to close document")
			}
		}(doc)

		img, err := doc.Image(0)
		if err != nil {
			err := tx.Rollback()
			if err != nil {
				return err
			}
			return err
		}

		buf := new(bytes.Buffer)
		if err := png.Encode(buf, img); err != nil {
			err := tx.Rollback()
			if err != nil {
				return err
			}
			return err
		}

		thumbnailFile, err := r.UploadFile(ctx, orig.Edges.Item.QueryGroup().FirstX(ctx), ItemCreateAttachment{
			Title:   fmt.Sprintf("%s-thumb", title),
			Content: bytes.NewReader(buf.Bytes()),
		})
		if err != nil {
			err := tx.Rollback()
			if err != nil {
				return err
			}
			log.Err(err).Msg("failed to upload thumbnail file")
			return err
		}
		att.SetPath(thumbnailFile)
	} else {
		return fmt.Errorf("unsupported file type for thumbnail generation")
	}
	_, err = att.Save(ctx)
	if err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		log.Err(err).Msg("failed to commit transaction")
		return nil
	}
	return nil
}

func (r *AttachmentRepo) CreateMissingThumbnails(ctx context.Context, groupId uuid.UUID) (int, error) {
	attachments, err := r.db.Attachment.Query().
		Where(
			attachment.HasItemWith(item.HasGroupWith(group.ID(groupId))),
			attachment.TypeNEQ("thumbnail"),
		).
		All(ctx)
	if err != nil {
		return 0, err
	}

	count := 0
	for _, attachment := range attachments {
		if !attachment.QueryThumbnail().ExistX(ctx) {
			err = r.CreateThumbnail(ctx, groupId, attachment.ID, attachment.Title, attachment.Path)
			if err != nil {
				log.Err(err).Msg("failed to create thumbnail")
				continue
			}
			count++
		}
	}

	return count, nil
}

func (r *AttachmentRepo) UploadFile(ctx context.Context, itemGroup *ent.Group, doc ItemCreateAttachment) (string, error) {
	// Prepare for the hashing of the file contents
	hashOut := make([]byte, 32)

	// Read all content into a buffer
	buf := new(bytes.Buffer)
	_, err := io.Copy(buf, doc.Content)
	if err != nil {
		log.Err(err).Msg("failed to read file content")
		return "", err
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
		return "", err
	}
	defer func(bucket *blob.Bucket) {
		err := bucket.Close()
		if err != nil {
			log.Err(err).Msg("failed to close bucket")
		}
	}(bucket)
	md5hash := md5.New()
	_, err = md5hash.Write(contentBytes)
	if err != nil {
		log.Err(err).Msg("failed to generate MD5 hash for storage")
		return "", err
	}
	contentType := http.DetectContentType(contentBytes[:min(512, len(contentBytes))])
	options := &blob.WriterOptions{
		ContentType: contentType,
		ContentMD5:  md5hash.Sum(nil),
	}
	path := r.path(itemGroup.ID, fmt.Sprintf("%x", hashOut))
	err = bucket.WriteAll(ctx, path, contentBytes, options)
	if err != nil {
		log.Err(err).Msg("failed to write file to bucket")
		return "", err
	}

	return path, nil
}

func isImageFile(title string) bool {
	// Check file extension for image types
	return strings.HasSuffix(title, ".jpg") || strings.HasSuffix(title, ".jpeg") || strings.HasSuffix(title, ".png")
}

func isDocumentFile(title string) bool {
	// Check file extension for document types
	return strings.HasSuffix(title, ".pdf") || strings.HasSuffix(title, ".epub") || strings.HasSuffix(title, ".mobi") || strings.HasSuffix(title, ".docx") || strings.HasSuffix(title, ".xlsx") || strings.HasSuffix(title, ".pptx") || strings.HasSuffix(title, ".txt") || strings.HasSuffix(title, ".html")
}
