package repo

import (
	"bytes"
	"context"
	"crypto/md5"
	"fmt"
	"github.com/rs/zerolog/log"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/group"
	"github.com/sysadminsmedia/homebox/backend/internal/sys/config"
	"github.com/sysadminsmedia/homebox/backend/pkgs/utils"
	"github.com/zeebo/blake3"

	"github.com/gen2brain/avif"
	"github.com/gen2brain/heic"
	"github.com/gen2brain/jpegxl"
	"github.com/gen2brain/webp"
	"golang.org/x/image/draw"
	"image"
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

	"gocloud.dev/pubsub"
	_ "gocloud.dev/pubsub/awssnssqs"
	_ "gocloud.dev/pubsub/azuresb"
	_ "gocloud.dev/pubsub/gcppubsub"
	_ "gocloud.dev/pubsub/kafkapubsub"
	_ "gocloud.dev/pubsub/mempubsub"
	_ "gocloud.dev/pubsub/natspubsub"
	_ "gocloud.dev/pubsub/rabbitpubsub"
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
		MimeType  string          `json:"mimeType,omitempty"`
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
		MimeType:  attachment.MimeType,
		Thumbnail: attachment.QueryThumbnail().FirstX(context.Background()),
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

	limitedReader := io.LimitReader(doc.Content, 1024*128)
	file, err := io.ReadAll(limitedReader)
	if err != nil {
		log.Err(err).Msg("failed to read file content")
		err = tx.Rollback()
		if err != nil {
			return nil, err
		}
		return nil, err
	}
	bldr = bldr.SetMimeType(http.DetectContentType(file[:min(512, len(file))]))
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
		pubsubString, err := utils.GenerateSubPubConn(r.pubSubConn, "thumbnails")
		if err != nil {
			log.Err(err).Msg("failed to generate pubsub connection string")
			return nil, err
		}
		topic, err := pubsub.OpenTopic(ctx, pubsubString)
		if err != nil {
			log.Err(err).Msg("failed to open pubsub topic")
			return nil, err
		}

		err = topic.Send(ctx, &pubsub.Message{
			Body: []byte(fmt.Sprintf("attachment_created:%s", attachmentDb.ID.String())),
			Metadata: map[string]string{
				"group_id":      itemGroup.ID.String(),
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

func (r *AttachmentRepo) Get(ctx context.Context, gid uuid.UUID, id uuid.UUID) (*ent.Attachment, error) {
	return r.db.Attachment.
		Query().
		Where(
			attachment.ID(id),
			attachment.HasItemWith(item.HasGroupWith(group.ID(gid))),
		).
		WithItem().
		WithThumbnail().
		Only(ctx)
}

func (r *AttachmentRepo) Update(ctx context.Context, gid uuid.UUID, id uuid.UUID, data *ItemAttachmentUpdate) (*ent.Attachment, error) {
	// Validate that the attachment belongs to the specified group
	_, err := r.db.Attachment.Query().
		Where(
			attachment.ID(id),
			attachment.HasItemWith(item.HasGroupWith(group.ID(gid))),
		).
		Only(ctx)
	if err != nil {
		return nil, err
	}

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

	return r.Get(ctx, gid, updatedAttachment.ID)
}

func (r *AttachmentRepo) Delete(ctx context.Context, gid uuid.UUID, itemId uuid.UUID, id uuid.UUID) error {
	// Validate that the attachment belongs to the specified group
	doc, err := r.db.Attachment.Query().
		Where(
			attachment.ID(id),
			attachment.HasItemWith(item.HasGroupWith(group.ID(gid))),
		).
		Only(ctx)
	if err != nil {
		return err
	}

	all, err := r.db.Attachment.Query().Where(attachment.Path(doc.Path)).All(ctx)
	if err != nil {
		return err
	}
	// If this is the last attachment for this path, delete the file
	if len(all) == 1 {
		thumb, err := doc.QueryThumbnail().First(ctx)
		if err != nil && !ent.IsNotFound(err) {
			log.Err(err).Msg("failed to query thumbnail for attachment")
			return err
		}
		if thumb != nil {
			thumbBucket, err := blob.OpenBucket(ctx, r.GetConnString())
			if err != nil {
				log.Err(err).Msg("failed to open bucket for thumbnail deletion")
				return err
			}
			err = thumbBucket.Delete(ctx, thumb.Path)
			if err != nil {
				return err
			}
			_ = doc.Update().SetNillableThumbnailID(nil).SaveX(ctx)
			_ = thumb.Update().SetNillableThumbnailID(nil).SaveX(ctx)
			err = r.db.Attachment.DeleteOneID(thumb.ID).Exec(ctx)
			if err != nil {
				return err
			}
		}
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

func (r *AttachmentRepo) Rename(ctx context.Context, gid uuid.UUID, id uuid.UUID, title string) (*ent.Attachment, error) {
	// Validate that the attachment belongs to the specified group
	_, err := r.db.Attachment.Query().
		Where(
			attachment.ID(id),
			attachment.HasItemWith(item.HasGroupWith(group.ID(gid))),
		).
		Only(ctx)
	if err != nil {
		return nil, err
	}

	return r.db.Attachment.UpdateOneID(id).SetTitle(title).Save(ctx)
}

//nolint:gocyclo
func (r *AttachmentRepo) CreateThumbnail(ctx context.Context, groupId, attachmentId uuid.UUID, title string, path string) error {
	log.Debug().Msg("starting thumbnail creation")
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

	log.Debug().Msg("set initial database transaction")
	att := tx.Attachment.Create().
		SetID(uuid.New()).
		SetTitle(fmt.Sprintf("%s-thumb", title)).
		SetType("thumbnail")

	log.Debug().Msg("opening original file")
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

	log.Debug().Msg("stat original file for file size")
	stats, err := origFile.Stat()
	if err != nil {
		err := tx.Rollback()
		if err != nil {
			return err
		}
		log.Err(err).Msg("failed to stat original file")
		return err
	}

	if stats.Size() > 100*1024*1024 {
		return fmt.Errorf("original file %s is too large to create a thumbnail", title)
	}

	log.Debug().Msg("reading original file content")
	contentBytes, err := io.ReadAll(origFile)
	if err != nil {
		err := tx.Rollback()
		if err != nil {
			return err
		}
		log.Err(err).Msg("failed to read original file content")
		return err
	}

	log.Debug().Msg("detecting content type of original file")
	contentType := http.DetectContentType(contentBytes[:min(512, len(contentBytes))])

	if contentType == "application/octet-stream" {
		if strings.HasSuffix(title, ".heic") || strings.HasSuffix(title, ".heif") {
			contentType = "image/heic"
		} else if strings.HasSuffix(title, ".avif") {
			contentType = "image/avif"
		}
	}

	switch {
	case isImageFile(contentType):
		log.Debug().Msg("creating thumbnail for image file")
		img, _, err := image.Decode(bytes.NewReader(contentBytes))
		if err != nil {
			log.Err(err).Msg("failed to decode image file")
			err := tx.Rollback()
			if err != nil {
				log.Err(err).Msg("failed to rollback transaction")
				return err
			}
			return err
		}
		dst := image.NewRGBA(image.Rect(0, 0, r.thumbnail.Width, r.thumbnail.Height))
		draw.ApproxBiLinear.Scale(dst, dst.Rect, img, img.Bounds(), draw.Over, nil)
		buf := new(bytes.Buffer)
		err = webp.Encode(buf, dst, webp.Options{Quality: 80, Lossless: false})
		if err != nil {
			err := tx.Rollback()
			if err != nil {
				return err
			}
			return err
		}
		contentBytes := buf.Bytes()
		log.Debug().Msg("uploading thumbnail file")
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
		log.Debug().Msg("setting thumbnail file path in attachment")
		att.SetPath(thumbnailFile)
	case contentType == "image/webp":
		log.Debug().Msg("creating thumbnail for webp file")
		img, err := webp.Decode(bytes.NewReader(contentBytes))
		if err != nil {
			log.Err(err).Msg("failed to decode webp image")
			err := tx.Rollback()
			if err != nil {
				return err
			}
			return err
		}
		dst := image.NewRGBA(image.Rect(0, 0, r.thumbnail.Width, r.thumbnail.Height))
		draw.ApproxBiLinear.Scale(dst, dst.Rect, img, img.Bounds(), draw.Over, nil)
		buf := new(bytes.Buffer)
		err = webp.Encode(buf, dst, webp.Options{Quality: 80, Lossless: false})
		if err != nil {
			err := tx.Rollback()
			if err != nil {
				return err
			}
			return err
		}
		contentBytes := buf.Bytes()
		log.Debug().Msg("uploading thumbnail file")
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
		log.Debug().Msg("setting thumbnail file path in attachment")
		att.SetPath(thumbnailFile)
	case contentType == "image/avif":
		log.Debug().Msg("creating thumbnail for avif file")
		img, err := avif.Decode(bytes.NewReader(contentBytes))
		if err != nil {
			log.Err(err).Msg("failed to decode avif image")
			err := tx.Rollback()
			if err != nil {
				return err
			}
			return err
		}
		dst := image.NewRGBA(image.Rect(0, 0, r.thumbnail.Width, r.thumbnail.Height))
		draw.ApproxBiLinear.Scale(dst, dst.Rect, img, img.Bounds(), draw.Over, nil)
		buf := new(bytes.Buffer)
		err = webp.Encode(buf, dst, webp.Options{Quality: 80, Lossless: false})
		if err != nil {
			err := tx.Rollback()
			if err != nil {
				return err
			}
			return err
		}
		contentBytes := buf.Bytes()
		log.Debug().Msg("uploading thumbnail file")
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
		log.Debug().Msg("setting thumbnail file path in attachment")
		att.SetPath(thumbnailFile)
	case contentType == "image/heic" || contentType == "image/heif":
		log.Debug().Msg("creating thumbnail for heic file")
		img, err := heic.Decode(bytes.NewReader(contentBytes))
		if err != nil {
			log.Err(err).Msg("failed to decode avif image")
			err := tx.Rollback()
			if err != nil {
				return err
			}
			return err
		}
		dst := image.NewRGBA(image.Rect(0, 0, r.thumbnail.Width, r.thumbnail.Height))
		draw.ApproxBiLinear.Scale(dst, dst.Rect, img, img.Bounds(), draw.Over, nil)
		buf := new(bytes.Buffer)
		err = webp.Encode(buf, dst, webp.Options{Quality: 80, Lossless: false})
		if err != nil {
			err := tx.Rollback()
			if err != nil {
				return err
			}
			return err
		}
		contentBytes := buf.Bytes()
		log.Debug().Msg("uploading thumbnail file")
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
		log.Debug().Msg("setting thumbnail file path in attachment")
		att.SetPath(thumbnailFile)
	case contentType == "image/jxl":
		log.Debug().Msg("creating thumbnail for jpegxl file")
		img, err := jpegxl.Decode(bytes.NewReader(contentBytes))
		if err != nil {
			log.Err(err).Msg("failed to decode avif image")
			err := tx.Rollback()
			if err != nil {
				return err
			}
			return err
		}
		dst := image.NewRGBA(image.Rect(0, 0, r.thumbnail.Width, r.thumbnail.Height))
		draw.ApproxBiLinear.Scale(dst, dst.Rect, img, img.Bounds(), draw.Over, nil)
		buf := new(bytes.Buffer)
		err = webp.Encode(buf, dst, webp.Options{Quality: 80, Lossless: false})
		if err != nil {
			err := tx.Rollback()
			if err != nil {
				return err
			}
			return err
		}
		contentBytes := buf.Bytes()
		log.Debug().Msg("uploading thumbnail file")
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
		log.Debug().Msg("setting thumbnail file path in attachment")
		att.SetPath(thumbnailFile)
	default:
		return fmt.Errorf("file type %s is not supported for thumbnail creation or document thumnails disabled", title)
	}

	att.SetMimeType("image/webp")

	log.Debug().Msg("saving thumbnail attachment to database")
	thumbnail, err := att.Save(ctx)
	if err != nil {
		return err
	}

	_, err = tx.Attachment.UpdateOneID(attachmentId).SetThumbnail(thumbnail).Save(ctx)
	if err != nil {
		return err
	}

	log.Debug().Msg("finishing thumbnail creation transaction")
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

	pubsubString, err := utils.GenerateSubPubConn(r.pubSubConn, "thumbnails")
	if err != nil {
		log.Err(err).Msg("failed to generate pubsub connection string")
	}
	topic, err := pubsub.OpenTopic(ctx, pubsubString)
	if err != nil {
		log.Err(err).Msg("failed to open pubsub topic")
	}

	count := 0
	for _, attachment := range attachments {
		if r.thumbnail.Enabled {
			if !attachment.QueryThumbnail().ExistX(ctx) {
				if count > 0 && count%100 == 0 {
					time.Sleep(2 * time.Second)
				}
				err = topic.Send(ctx, &pubsub.Message{
					Body: []byte(fmt.Sprintf("attachment_created:%s", attachment.ID.String())),
					Metadata: map[string]string{
						"group_id":      groupId.String(),
						"attachment_id": attachment.ID.String(),
						"title":         attachment.Title,
						"path":          attachment.Path,
					},
				})
				if err != nil {
					log.Err(err).Msg("failed to send message to topic")
					continue
				} else {
					count++
				}
			}
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

func isImageFile(mimetype string) bool {
	// Check file extension for image types
	return strings.Contains(mimetype, "image/jpeg") || strings.Contains(mimetype, "image/png") || strings.Contains(mimetype, "image/gif")
}
