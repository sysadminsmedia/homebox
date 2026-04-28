package repo

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/export"
)

// ExportRepository persists Export job rows. Every method is group-scoped:
// callers pass the requesting tenant's gid and the repo refuses to act on
// rows owned by a different group.
type ExportRepository struct {
	db *ent.Client
}

type ExportOut struct {
	ID           uuid.UUID `json:"id"`
	GroupID      uuid.UUID `json:"groupId"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
	Status       string    `json:"status"`
	Progress     int       `json:"progress"`
	ArtifactPath string    `json:"artifactPath,omitempty"`
	SizeBytes    int64     `json:"sizeBytes"`
	Error        string    `json:"error,omitempty"`
}

func mapExport(e *ent.Export) ExportOut {
	return ExportOut{
		ID:           e.ID,
		GroupID:      e.GroupID,
		CreatedAt:    e.CreatedAt,
		UpdatedAt:    e.UpdatedAt,
		Status:       string(e.Status),
		Progress:     e.Progress,
		ArtifactPath: e.ArtifactPath,
		SizeBytes:    e.SizeBytes,
		Error:        e.Error,
	}
}

func (r *ExportRepository) Create(ctx context.Context, gid uuid.UUID) (ExportOut, error) {
	e, err := r.db.Export.Create().
		SetGroupID(gid).
		Save(ctx)
	if err != nil {
		return ExportOut{}, err
	}
	return mapExport(e), nil
}

func (r *ExportRepository) ListByGroup(ctx context.Context, gid uuid.UUID) ([]ExportOut, error) {
	rows, err := r.db.Export.Query().
		Where(export.GroupID(gid)).
		Order(ent.Desc(export.FieldCreatedAt)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]ExportOut, len(rows))
	for i, e := range rows {
		out[i] = mapExport(e)
	}
	return out, nil
}

// Get returns an export iff it exists AND is owned by gid.
func (r *ExportRepository) Get(ctx context.Context, gid uuid.UUID, id uuid.UUID) (ExportOut, error) {
	e, err := r.db.Export.Query().
		Where(export.ID(id), export.GroupID(gid)).
		Only(ctx)
	if err != nil {
		return ExportOut{}, err
	}
	return mapExport(e), nil
}

func (r *ExportRepository) SetRunning(ctx context.Context, id uuid.UUID) error {
	return r.db.Export.UpdateOneID(id).
		SetStatus(export.StatusRunning).
		SetProgress(0).
		Exec(ctx)
}

func (r *ExportRepository) SetProgress(ctx context.Context, id uuid.UUID, pct int) error {
	if pct < 0 {
		pct = 0
	} else if pct > 100 {
		pct = 100
	}
	return r.db.Export.UpdateOneID(id).SetProgress(pct).Exec(ctx)
}

func (r *ExportRepository) SetCompleted(ctx context.Context, id uuid.UUID, artifactPath string, sizeBytes int64) error {
	return r.db.Export.UpdateOneID(id).
		SetStatus(export.StatusCompleted).
		SetProgress(100).
		SetArtifactPath(artifactPath).
		SetSizeBytes(sizeBytes).
		Exec(ctx)
}

func (r *ExportRepository) SetFailed(ctx context.Context, id uuid.UUID, errMsg string) error {
	if len(errMsg) > 1000 {
		errMsg = errMsg[:1000]
	}
	return r.db.Export.UpdateOneID(id).
		SetStatus(export.StatusFailed).
		SetError(errMsg).
		Exec(ctx)
}

// Delete removes an export row scoped to gid. Callers must remove the blob
// artifact separately if one exists.
func (r *ExportRepository) Delete(ctx context.Context, gid uuid.UUID, id uuid.UUID) (int, error) {
	return r.db.Export.Delete().
		Where(export.ID(id), export.GroupID(gid)).
		Exec(ctx)
}

// PurgeOlderThan returns rows that would be purged (so callers can drop the
// blob artifacts) and then deletes them. Not scoped by group on purpose: this
// is the cleanup task that sweeps every tenant.
func (r *ExportRepository) PurgeOlderThan(ctx context.Context, cutoff time.Time) ([]ExportOut, error) {
	rows, err := r.db.Export.Query().
		Where(export.CreatedAtLT(cutoff)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]ExportOut, len(rows))
	for i, e := range rows {
		out[i] = mapExport(e)
	}
	if _, err := r.db.Export.Delete().Where(export.CreatedAtLT(cutoff)).Exec(ctx); err != nil {
		return nil, err
	}
	return out, nil
}
