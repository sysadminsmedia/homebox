package repo

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/document"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/group"
	"github.com/sysadminsmedia/homebox/backend/pkgs/pathlib"
)

var ErrInvalidDocExtension = errors.New("invalid document extension")

type DocumentRepository struct {
	db  *ent.Client
	dir string
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
	return pathlib.Safe(filepath.Join(r.dir, gid.String(), "documents", uuid.NewString()+ext))
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

	parent := filepath.Dir(path)
	err := os.MkdirAll(parent, 0o755)
	if err != nil {
		return DocumentOut{}, err
	}

	f, err := os.Create(path)
	if err != nil {
		return DocumentOut{}, err
	}

	_, err = io.Copy(f, doc.Content)
	if err != nil {
		return DocumentOut{}, err
	}

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
	doc, err := r.db.Document.Get(ctx, id)
	if err != nil {
		return err
	}

	err = os.Remove(doc.Path)
	if err != nil {
		return err
	}

	return r.db.Document.DeleteOneID(id).Exec(ctx)
}
