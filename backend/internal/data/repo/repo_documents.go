package repo

import (
	"context"
	"crypto/md5"
	"errors"
	"github.com/rs/zerolog/log"
	"io"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/document"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/group"
	"github.com/sysadminsmedia/homebox/backend/pkgs/pathlib"

	"gocloud.dev/blob"
	_ "gocloud.dev/blob/azureblob"
	_ "gocloud.dev/blob/fileblob"
	_ "gocloud.dev/blob/gcsblob"
	_ "gocloud.dev/blob/s3blob"
)

var ErrInvalidDocExtension = errors.New("invalid document extension")

type DocumentRepository struct {
	db         *ent.Client
	storePath  string
	connString string
}

type (
	DocumentCreate struct {
		Title   string    `json:"title"`
		Content io.Reader `json:"content"`
	}

	DocumentOut struct {
		ID    uuid.UUID `json:"id"`
		Title string    `json:"title"`
		Path  string    `json:"path"`
	}
)

func mapDocumentOut(doc *ent.Document) DocumentOut {
	return DocumentOut{
		ID:    doc.ID,
		Title: doc.Title,
		Path:  doc.Path,
	}
}

var (
	mapDocumentOutErr     = mapTErrFunc(mapDocumentOut)
	mapDocumentOutEachErr = mapTEachErrFunc(mapDocumentOut)
)

func (r *DocumentRepository) path(gid uuid.UUID, ext string) string {
	return pathlib.Safe(filepath.Join(r.storePath, gid.String(), "documents", uuid.NewString()+ext))
}

func (r *DocumentRepository) GetAll(ctx context.Context, gid uuid.UUID) ([]DocumentOut, error) {
	return mapDocumentOutEachErr(r.db.Document.
		Query().
		Where(document.HasGroupWith(group.ID(gid))).
		All(ctx),
	)
}

func (r *DocumentRepository) Get(ctx context.Context, id uuid.UUID) (DocumentOut, error) {
	return mapDocumentOutErr(r.db.Document.Get(ctx, id))
}

func (r *DocumentRepository) Create(ctx context.Context, gid uuid.UUID, doc DocumentCreate) (DocumentOut, error) {
	ext := filepath.Ext(doc.Title)
	if ext == "" {
		return DocumentOut{}, ErrInvalidDocExtension
	}

	path := r.path(gid, ext)

	bucket, err := blob.OpenBucket(context.Background(), r.connString)
	if err != nil {
		log.Err(err).Msg("failed to open bucket")
		return DocumentOut{}, err
	}

	fileData, err := io.ReadAll(doc.Content)
	if err != nil {
		log.Err(err).Msg("failed to read all from content")
		return DocumentOut{}, err
	}

	hash := md5.New()
	hash.Write(fileData)
	options := &blob.WriterOptions{
		ContentType:                 "application/octet-stream",
		DisableContentTypeDetection: false,
		ContentMD5:                  hash.Sum(nil),
	}

	err = bucket.WriteAll(ctx, path, fileData, options)
	if err != nil {
		log.Err(err).Msg("failed to write all to bucket")
		return DocumentOut{}, err
	}
	defer func(bucket *blob.Bucket) {
		err := bucket.Close()
		if err != nil {
			log.Err(err).Msg("failed to close bucket")
		}
	}(bucket)

	return mapDocumentOutErr(r.db.Document.Create().
		SetGroupID(gid).
		SetTitle(doc.Title).
		SetPath(path).
		Save(ctx),
	)
}

func (r *DocumentRepository) Rename(ctx context.Context, id uuid.UUID, title string) (DocumentOut, error) {
	return mapDocumentOutErr(r.db.Document.UpdateOneID(id).
		SetTitle(title).
		Save(ctx))
}

func (r *DocumentRepository) Delete(ctx context.Context, id uuid.UUID) error {

	bucket, err := blob.OpenBucket(context.Background(), r.connString)
	if err != nil {
		log.Err(err).Msg("failed to open bucket")
		return err
	}

	doc, err := r.db.Document.Get(ctx, id)
	if err != nil {
		return err
	}

	err = bucket.Delete(ctx, doc.Path)
	if err != nil {
		return err
	}

	defer func(bucket *blob.Bucket) {
		err := bucket.Close()
		if err != nil {
			log.Err(err).Msg("failed to close bucket")
		}
	}(bucket)

	return r.db.Document.DeleteOneID(id).Exec(ctx)
}
