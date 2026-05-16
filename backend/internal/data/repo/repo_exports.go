package repo

import (
	"context"
	"time"
	"unicode/utf8"

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
	ID      uuid.UUID `json:"id"`
	GroupID uuid.UUID `json:"groupId"`
	// Kind is "export" for server-produced backup artifacts, "import" for
	// user-uploaded restore zips. The lifecycle fields below behave the
	// same for both.
	Kind         string    `json:"kind"`
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
		Kind:         string(e.Kind),
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

// CreateImport stages a new pending row representing an upload that the
// worker will restore. The uploadKey points at the blob already written
// to "{gid}/imports/{uuid}.zip", and sizeBytes is the streamed upload
// size so the UI can show "X MB queued" before the worker even starts.
func (r *ExportRepository) CreateImport(ctx context.Context, gid uuid.UUID, uploadKey string, sizeBytes int64) (ExportOut, error) {
	e, err := r.db.Export.Create().
		SetGroupID(gid).
		SetKind(export.KindImport).
		SetArtifactPath(uploadKey).
		SetSizeBytes(sizeBytes).
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

// SetRunning, SetProgress, SetCompleted, and SetFailed all carry gid so the
// underlying UPDATE matches only when the row belongs to that group. A
// mismatched gid yields ent.NotFoundError rather than a silent cross-tenant
// mutation — matching the package contract documented on ExportRepository.
func (r *ExportRepository) SetRunning(ctx context.Context, gid, id uuid.UUID) error {
	return r.db.Export.UpdateOneID(id).
		Where(export.GroupID(gid)).
		SetStatus(export.StatusRunning).
		SetProgress(0).
		Exec(ctx)
}

func (r *ExportRepository) SetProgress(ctx context.Context, gid, id uuid.UUID, pct int) error {
	if pct < 0 {
		pct = 0
	} else if pct > 100 {
		pct = 100
	}
	return r.db.Export.UpdateOneID(id).
		Where(export.GroupID(gid)).
		SetProgress(pct).
		Exec(ctx)
}

func (r *ExportRepository) SetCompleted(ctx context.Context, gid, id uuid.UUID, artifactPath string, sizeBytes int64) error {
	return r.db.Export.UpdateOneID(id).
		Where(export.GroupID(gid)).
		SetStatus(export.StatusCompleted).
		SetProgress(100).
		SetArtifactPath(artifactPath).
		SetSizeBytes(sizeBytes).
		Exec(ctx)
}

func (r *ExportRepository) SetFailed(ctx context.Context, gid, id uuid.UUID, errMsg string) error {
	const maxErrBytes = 1000
	if len(errMsg) > maxErrBytes {
		// Cut at the last rune boundary that keeps the total ≤ maxErrBytes.
		// Plain byte-slicing can split a multibyte rune and the resulting
		// invalid UTF-8 fails Postgres' UTF8 encoding check on insert,
		// masking the real failure with a database error.
		cut := 0
		for i, r := range errMsg {
			end := i + utf8.RuneLen(r)
			if end > maxErrBytes {
				break
			}
			cut = end
		}
		errMsg = errMsg[:cut]
	}
	return r.db.Export.UpdateOneID(id).
		Where(export.GroupID(gid)).
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

// ListOlderThan returns rows older than cutoff so the sweep task can drop
// each one's blob artifact before removing the DB row. The row carries the
// only persisted pointer to the blob, so the caller MUST delete the row only
// after the blob is gone (or confirmed absent) — otherwise a transient bucket
// outage would orphan the blob with no path to find it again. Not scoped by
// group on purpose: this is the cleanup task that sweeps every tenant.
func (r *ExportRepository) ListOlderThan(ctx context.Context, cutoff time.Time) ([]ExportOut, error) {
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
	return out, nil
}
