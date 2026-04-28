package services

import (
	"archive/zip"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel"
	"gocloud.dev/blob"
	"gocloud.dev/pubsub"

	"github.com/sysadminsmedia/homebox/backend/internal/core/services/reporting/eventbus"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/entity"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/entitytype"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/group"
	"github.com/sysadminsmedia/homebox/backend/internal/data/repo"
	"github.com/sysadminsmedia/homebox/backend/internal/sys/config"
	"github.com/sysadminsmedia/homebox/backend/pkgs/utils"
)

// ExportSchemaVersion is the on-disk version of the export zip layout.
// Bump this when manifest/file shapes change in incompatible ways and import
// can no longer round-trip an older export.
const ExportSchemaVersion = 1

// Pubsub topic names used by the export and import workers.
const (
	TopicCollectionExport = "collection_export"
	TopicCollectionImport = "collection_import"
)

// ManifestFile is the name of the manifest entry inside the zip artifact.
const manifestFile = "manifest.json"

// attachmentsDir is the prefix inside the zip for attachment blobs.
const attachmentsDir = "attachments/"

// tableSpec describes how to extract one table's rows scoped to a group, and
// how to handle foreign keys on import.
//
// New fields/columns flow through automatically: export uses SELECT * and
// import builds INSERT from the JSON keys. Adding a new TABLE still requires
// editing this list and (probably) the dependency graph; same for adding a
// new FK column to an existing table that points at another exported table.
type tableSpec struct {
	// name is the SQL table name.
	name string
	// scope is a SQL WHERE fragment with one ? placeholder for the group ID.
	// Use "" to fetch every row in the table.
	scope string
	// pkCol is the primary-key column name. "" for junction tables that have
	// no single-column PK (e.g. tag_entities).
	pkCol string
	// groupCols are columns whose values are remapped to the destination
	// group_id on import (the various "group_xxx" FK columns).
	groupCols []string
	// userCols are columns whose values are remapped to the importing user
	// (notifiers being the only example).
	userCols []string
	// fkCols are immediate foreign keys: { column → target table }. The
	// import looks each value up in the id map populated by earlier table
	// inserts and substitutes the new id.
	fkCols map[string]string
	// deferCols are foreign keys whose target row may not exist yet at the
	// time this row is inserted (self-references and forward-circular refs).
	// They are nulled on insert and patched in a second pass.
	deferCols map[string]string
}

// exportTables defines the export/import schema. Order matters: imports run
// in this order, so each table's non-deferred FK targets must already be
// present.
//
// Why every PK is remapped on import: a real "fresh server" import would
// keep original IDs, but if the user re-imports a backup into the same
// server (or a server that already received this export once), reusing PKs
// causes UNIQUE-constraint violations. Remapping always = simpler invariant.
//
// Self-referential FKs (entities.entity_children, tags.tag_children,
// attachments.attachment_thumbnail) and forward-circular FKs
// (entity_types.entity_type_default_template,
// entity_templates.entity_template_location) live in deferCols so the first
// INSERT pass can succeed; the second pass patches them with remapped IDs.
//
// Known gap: entity_templates.default_tag_ids is a JSON list of tag UUIDs.
// We do not currently rewrite UUIDs nested inside JSON columns, so that
// reference is lost on import. Templates and tags both still come across
// individually; only the template→tag default association is dropped.
var exportTables = []tableSpec{
	{
		name:      "entity_types",
		scope:     "group_entity_types = ?",
		pkCol:     "id",
		groupCols: []string{"group_entity_types"},
		deferCols: map[string]string{"entity_type_default_template": "entity_templates"},
	},
	{
		name:      "entity_templates",
		scope:     "group_entity_templates = ?",
		pkCol:     "id",
		groupCols: []string{"group_entity_templates"},
		deferCols: map[string]string{"entity_template_location": "entities"},
	},
	{
		name:   "template_fields",
		scope:  "entity_template_fields IN (SELECT id FROM entity_templates WHERE group_entity_templates = ?)",
		pkCol:  "id",
		fkCols: map[string]string{"entity_template_fields": "entity_templates"},
	},
	{
		name:      "tags",
		scope:     "group_tags = ?",
		pkCol:     "id",
		groupCols: []string{"group_tags"},
		deferCols: map[string]string{"tag_children": "tags"},
	},
	{
		name:      "entities",
		scope:     "group_entities = ?",
		pkCol:     "id",
		groupCols: []string{"group_entities"},
		fkCols:    map[string]string{"entity_type_entities": "entity_types"},
		deferCols: map[string]string{"entity_children": "entities"},
	},
	{
		name:   "entity_fields",
		scope:  "entity_fields IN (SELECT id FROM entities WHERE group_entities = ?)",
		pkCol:  "id",
		fkCols: map[string]string{"entity_fields": "entities"},
	},
	{
		name:   "maintenance_entries",
		scope:  "entity_id IN (SELECT id FROM entities WHERE group_entities = ?)",
		pkCol:  "id",
		fkCols: map[string]string{"entity_id": "entities"},
	},
	{
		// Two-part scope: the regular attachments owned by an entity in this
		// group, PLUS the thumbnail rows those attachments point at (which
		// have entity_attachments=NULL and are linked only via
		// attachment_thumbnail on the parent). Each ? is the same gid;
		// dumpTable/wipeGroup expand based on placeholder count.
		name: "attachments",
		scope: "entity_attachments IN (SELECT id FROM entities WHERE group_entities = ?)" +
			" OR id IN (SELECT attachment_thumbnail FROM attachments" +
			" WHERE attachment_thumbnail IS NOT NULL" +
			" AND entity_attachments IN (SELECT id FROM entities WHERE group_entities = ?))",
		pkCol:     "id",
		fkCols:    map[string]string{"entity_attachments": "entities"},
		deferCols: map[string]string{"attachment_thumbnail": "attachments"},
	},
	{
		name:   "tag_entities",
		scope:  "tag_id IN (SELECT id FROM tags WHERE group_tags = ?)",
		fkCols: map[string]string{"tag_id": "tags", "entity_id": "entities"},
	},
	{
		name:      "notifiers",
		scope:     "group_id = ?",
		pkCol:     "id",
		groupCols: []string{"group_id"},
		userCols:  []string{"user_id"},
	},
}

// Manifest is the contents of manifest.json inside the export zip.
type Manifest struct {
	SchemaVersion  int            `json:"schemaVersion"`
	ExportedAt     time.Time      `json:"exportedAt"`
	GroupID        uuid.UUID      `json:"groupId"`
	HomeboxVersion string         `json:"homeboxVersion,omitempty"`
	Counts         map[string]int `json:"counts"`
}

// ExportService orchestrates the export and import jobs. It is wired into
// AllServices and invoked by the pubsub workers in app/api/recurring.go.
//
// Every public method takes the requesting tenant's group id and refuses to
// operate on data that does not belong to that group.
type ExportService struct {
	db         *ent.Client
	repos      *repo.AllRepos
	bus        *eventbus.EventBus
	storage    config.Storage
	pubSubConn string
	dialect    string // "sqlite3" or "postgres"
}

// Enqueue creates a pending Export row for gid and publishes a job to the
// export topic. The actual zip-building happens in the worker.
func (s *ExportService) Enqueue(ctx context.Context, gid uuid.UUID) (repo.ExportOut, error) {
	ctx, span := otel.Tracer("services").Start(ctx, "ExportService.Enqueue")
	defer span.End()

	out, err := s.repos.Exports.Create(ctx, gid)
	if err != nil {
		return out, err
	}

	if err := s.publishExportJob(ctx, gid, out.ID); err != nil {
		_ = s.repos.Exports.SetFailed(ctx, out.ID, "failed to enqueue: "+err.Error())
		return out, err
	}

	s.publishMutation(gid)
	return out, nil
}

// EnqueueImport publishes a job to import the zip already staged at
// uploadKey. uploadKey must live under "{gid}/imports/" — the worker will
// re-validate this before reading.
func (s *ExportService) EnqueueImport(ctx context.Context, gid uuid.UUID, userID uuid.UUID, uploadKey string) error {
	ctx, span := otel.Tracer("services").Start(ctx, "ExportService.EnqueueImport")
	defer span.End()
	return s.publishImportJob(ctx, gid, userID, uploadKey)
}

// IsGroupReadyForImport returns true if gid contains no items (i.e. no
// entities whose type is not a location). Default locations and tags
// auto-seeded at registration are tolerated — the import wipes them before
// restoring. Hard "no entities at all" was too strict because a freshly
// registered group already has the seeded locations/tags.
func (s *ExportService) IsGroupReadyForImport(ctx context.Context, gid uuid.UUID) (bool, error) {
	n, err := s.db.Entity.Query().Where(
		entity.HasGroupWith(group.ID(gid)),
		entity.HasEntityTypeWith(entitytype.IsLocation(false)),
	).Count(ctx)
	if err != nil {
		return false, err
	}
	return n == 0, nil
}

// RunExport is invoked by the pubsub subscriber when an export job message is
// received. It transitions the row through running → completed/failed and
// uploads the artifact to blob storage.
func (s *ExportService) RunExport(ctx context.Context, exportID, gid uuid.UUID) {
	ctx, span := otel.Tracer("services").Start(ctx, "ExportService.RunExport")
	defer span.End()

	exp, err := s.repos.Exports.Get(ctx, gid, exportID)
	if err != nil {
		log.Err(err).Stringer("export_id", exportID).Stringer("gid", gid).Msg("export job: row not found or wrong group")
		return
	}
	if exp.Status != "pending" {
		log.Warn().Stringer("export_id", exportID).Str("status", exp.Status).Msg("export job: not pending, skipping")
		return
	}

	if err := s.repos.Exports.SetRunning(ctx, exportID); err != nil {
		log.Err(err).Msg("export job: failed to mark running")
		return
	}
	s.publishMutation(gid)

	artifactPath, sizeBytes, err := s.buildArtifact(ctx, exportID, gid)
	if err != nil {
		log.Err(err).Stringer("export_id", exportID).Msg("export job: failed")
		_ = s.repos.Exports.SetFailed(ctx, exportID, err.Error())
		s.publishMutation(gid)
		return
	}

	if err := s.repos.Exports.SetCompleted(ctx, exportID, artifactPath, sizeBytes); err != nil {
		log.Err(err).Msg("export job: failed to mark completed")
	}
	s.publishMutation(gid)
}

// buildArtifact does the actual zip generation: dump every group-scoped
// table to JSON, copy attachment blobs, write manifest, upload to blob
// storage. Returns the blob key and total size.
func (s *ExportService) buildArtifact(ctx context.Context, exportID, gid uuid.UUID) (string, int64, error) {
	tmp, err := os.CreateTemp("", fmt.Sprintf("homebox-export-%s-*.zip", exportID))
	if err != nil {
		return "", 0, fmt.Errorf("create temp file: %w", err)
	}
	tmpPath := tmp.Name()
	defer func() {
		_ = tmp.Close()
		_ = os.Remove(tmpPath)
	}()

	zw := zip.NewWriter(tmp)

	counts := make(map[string]int)
	dbSql := s.db.Sql()
	for i, spec := range exportTables {
		rows, err := dumpTable(ctx, dbSql, s.dialect, spec, gid)
		if err != nil {
			_ = zw.Close()
			return "", 0, fmt.Errorf("dump %s: %w", spec.name, err)
		}
		counts[spec.name] = len(rows)

		w, err := zw.Create(spec.name + ".json")
		if err != nil {
			_ = zw.Close()
			return "", 0, fmt.Errorf("zip create %s.json: %w", spec.name, err)
		}
		enc := json.NewEncoder(w)
		if err := enc.Encode(rows); err != nil {
			_ = zw.Close()
			return "", 0, fmt.Errorf("zip encode %s.json: %w", spec.name, err)
		}

		// Coarse-grained progress: 0..80% spans the table dumps, 80..95% the
		// attachment copies, 95..100% the upload.
		pct := int(float64(i+1) / float64(len(exportTables)) * 80)
		_ = s.repos.Exports.SetProgress(ctx, exportID, pct)
	}

	// Copy attachment blobs into the zip.
	if err := s.copyAttachmentBlobs(ctx, zw, gid); err != nil {
		_ = zw.Close()
		return "", 0, fmt.Errorf("copy attachments: %w", err)
	}
	_ = s.repos.Exports.SetProgress(ctx, exportID, 95)

	// Manifest last so we know the counts.
	mf := Manifest{
		SchemaVersion: ExportSchemaVersion,
		ExportedAt:    time.Now().UTC(),
		GroupID:       gid,
		Counts:        counts,
	}
	mw, err := zw.Create(manifestFile)
	if err != nil {
		_ = zw.Close()
		return "", 0, fmt.Errorf("zip create manifest: %w", err)
	}
	if err := json.NewEncoder(mw).Encode(mf); err != nil {
		_ = zw.Close()
		return "", 0, fmt.Errorf("zip encode manifest: %w", err)
	}

	if err := zw.Close(); err != nil {
		return "", 0, fmt.Errorf("zip close: %w", err)
	}

	// Upload to blob storage.
	if _, err := tmp.Seek(0, io.SeekStart); err != nil {
		return "", 0, fmt.Errorf("seek temp: %w", err)
	}
	stat, err := tmp.Stat()
	if err != nil {
		return "", 0, fmt.Errorf("stat temp: %w", err)
	}
	size := stat.Size()

	artifactPath := fmt.Sprintf("%s/exports/%s.zip", gid.String(), exportID.String())
	bucket, err := blob.OpenBucket(ctx, s.repos.Attachments.GetConnString())
	if err != nil {
		return "", 0, fmt.Errorf("open bucket: %w", err)
	}
	defer func() { _ = bucket.Close() }()

	bw, err := bucket.NewWriter(ctx, s.repos.Attachments.GetFullPath(artifactPath), &blob.WriterOptions{
		ContentType: "application/zip",
	})
	if err != nil {
		return "", 0, fmt.Errorf("blob writer: %w", err)
	}
	if _, err := io.Copy(bw, tmp); err != nil {
		_ = bw.Close()
		return "", 0, fmt.Errorf("blob copy: %w", err)
	}
	if err := bw.Close(); err != nil {
		return "", 0, fmt.Errorf("blob close: %w", err)
	}

	return artifactPath, size, nil
}

// copyAttachmentBlobs streams every attachment blob in the group — including
// thumbnail rows — into the zip under attachments/{attachment_id}. Lookup on
// the import side uses the file's stem (the attachment UUID) via the id map.
//
// Reuses the attachments tableSpec scope so the row dump and the blob copy
// can never disagree about which attachments belong to the group.
func (s *ExportService) copyAttachmentBlobs(ctx context.Context, zw *zip.Writer, gid uuid.UUID) error {
	var spec tableSpec
	for _, t := range exportTables {
		if t.name == "attachments" {
			spec = t
			break
		}
	}

	q := "SELECT id, path FROM attachments WHERE " + rebindPlaceholders(spec.scope, s.dialect)
	args := make([]any, 0, strings.Count(spec.scope, "?"))
	for i := 0; i < cap(args); i++ {
		args = append(args, gid.String())
	}
	rows, err := s.db.Sql().QueryContext(ctx, q, args...)
	if err != nil {
		return err
	}
	type attRef struct{ id, path string }
	var refs []attRef
	for rows.Next() {
		var id, path string
		if err := rows.Scan(&id, &path); err != nil {
			_ = rows.Close()
			return err
		}
		if path == "" {
			continue
		}
		refs = append(refs, attRef{id: id, path: path})
	}
	if err := rows.Close(); err != nil {
		return err
	}

	bucket, err := blob.OpenBucket(ctx, s.repos.Attachments.GetConnString())
	if err != nil {
		return err
	}
	defer func() { _ = bucket.Close() }()

	for _, ref := range refs {
		r, err := bucket.NewReader(ctx, s.repos.Attachments.GetFullPath(ref.path), nil)
		if err != nil {
			// Don't fail the whole export for one missing blob; just skip it.
			// On import the attachment row will exist but the blob won't —
			// same end state as a thumbnail-generation failure today.
			log.Warn().Err(err).Str("path", ref.path).Msg("export: attachment blob missing, skipping")
			continue
		}
		w, err := zw.Create(attachmentsDir + ref.id)
		if err != nil {
			_ = r.Close()
			return err
		}
		if _, err := io.Copy(w, r); err != nil {
			_ = r.Close()
			return err
		}
		_ = r.Close()
	}
	return nil
}

// publishExportJob sends a message on the export topic.
func (s *ExportService) publishExportJob(ctx context.Context, gid, exportID uuid.UUID) error {
	conn, err := utils.GenerateSubPubConn(s.pubSubConn, TopicCollectionExport)
	if err != nil {
		return err
	}
	topic, err := pubsub.OpenTopic(ctx, conn)
	if err != nil {
		return err
	}
	defer func() { _ = topic.Shutdown(ctx) }()
	return topic.Send(ctx, &pubsub.Message{
		Body: []byte("collection_export:" + exportID.String()),
		Metadata: map[string]string{
			"group_id":  gid.String(),
			"export_id": exportID.String(),
		},
	})
}

// publishImportJob sends a message on the import topic. The worker reads the
// staged upload from blob storage at uploadKey, unzips, restores into the
// group identified by gid, then deletes the staged upload.
func (s *ExportService) publishImportJob(ctx context.Context, gid, userID uuid.UUID, uploadKey string) error {
	conn, err := utils.GenerateSubPubConn(s.pubSubConn, TopicCollectionImport)
	if err != nil {
		return err
	}
	topic, err := pubsub.OpenTopic(ctx, conn)
	if err != nil {
		return err
	}
	defer func() { _ = topic.Shutdown(ctx) }()
	return topic.Send(ctx, &pubsub.Message{
		Body: []byte("collection_import:" + gid.String()),
		Metadata: map[string]string{
			"group_id":   gid.String(),
			"user_id":    userID.String(),
			"upload_key": uploadKey,
		},
	})
}

func (s *ExportService) publishMutation(gid uuid.UUID) {
	if s.bus != nil {
		s.bus.Publish(eventbus.EventExportMutation, eventbus.GroupMutationEvent{GID: gid})
	}
}

// dumpTable runs SELECT * for spec.scope and returns each row as a JSON-
// friendly map. UUIDs and JSON-blob columns come back from sqlite as []byte;
// we coerce to string here so json.Marshal does the right thing.
//
// Scope clauses may contain multiple ? placeholders (e.g. for an OR-of-
// subqueries). Each placeholder is filled with the same gid — none of the
// existing scopes need to vary by placeholder.
func dumpTable(ctx context.Context, db *sql.DB, dialect string, spec tableSpec, gid uuid.UUID) ([]map[string]any, error) {
	q := "SELECT * FROM " + spec.name
	var args []any
	if spec.scope != "" {
		q += " WHERE " + rebindPlaceholders(spec.scope, dialect)
		for i := 0; i < strings.Count(spec.scope, "?"); i++ {
			args = append(args, gid.String())
		}
	}
	rows, err := db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	out := make([]map[string]any, 0)
	for rows.Next() {
		vals := make([]any, len(cols))
		ptrs := make([]any, len(cols))
		for i := range vals {
			ptrs[i] = &vals[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			return nil, err
		}
		row := make(map[string]any, len(cols))
		for i, col := range cols {
			row[col] = normalizeScan(vals[i])
		}
		out = append(out, row)
	}
	return out, rows.Err()
}

// normalizeScan converts driver-returned values into JSON-marshallable
// shapes. The two big ones: []byte (UUIDs and JSON blobs in sqlite) becomes
// string, and time.Time stays as time.Time so json.Marshal renders RFC3339.
func normalizeScan(v any) any {
	switch x := v.(type) {
	case []byte:
		return string(x)
	default:
		return v
	}
}

// rebindPlaceholders rewrites "?" to "$1", "$2", … for postgres. SQLite uses
// "?" natively. Assumes scope clauses use a single placeholder per occurrence.
func rebindPlaceholders(s, dialect string) string {
	if dialect != "postgres" {
		return s
	}
	var b strings.Builder
	n := 0
	for _, ch := range s {
		if ch == '?' {
			n++
			fmt.Fprintf(&b, "$%d", n)
			continue
		}
		b.WriteRune(ch)
	}
	return b.String()
}

// =============================================================================
// Import path
// =============================================================================

// RunImport is invoked by the pubsub subscriber when an import job message
// is received. It reads the staged upload, validates the manifest, asserts
// the destination group is empty, and replays every row.
func (s *ExportService) RunImport(ctx context.Context, gid, userID uuid.UUID, uploadKey string) {
	ctx, span := otel.Tracer("services").Start(ctx, "ExportService.RunImport")
	defer span.End()

	// Hard scope check: refuse anything that doesn't live under the caller's
	// group prefix. Defence in depth — the handler already enforced this.
	prefix := gid.String() + "/imports/"
	if !strings.HasPrefix(uploadKey, prefix) {
		log.Error().Str("upload_key", uploadKey).Stringer("gid", gid).Msg("import job: upload key outside group prefix, refusing")
		return
	}

	if err := s.runImport(ctx, gid, userID, uploadKey); err != nil {
		log.Err(err).Stringer("gid", gid).Msg("import job: failed")
	}

	// Cleanup the staging blob whether the import succeeded or not — keeping
	// it around just lets a second delivery race against the populated DB.
	if err := s.deleteUpload(ctx, uploadKey); err != nil {
		log.Warn().Err(err).Str("upload_key", uploadKey).Msg("import job: failed to clean staging upload")
	}

	s.publishImportFinished(gid)
}

func (s *ExportService) runImport(ctx context.Context, gid, userID uuid.UUID, uploadKey string) error {
	// Precondition: no items (non-location entities) in this group. Default
	// seeded locations/tags/entity_types are fine; we wipe them below before
	// restoring.
	ready, err := s.IsGroupReadyForImport(ctx, gid)
	if err != nil {
		return fmt.Errorf("import precondition: %w", err)
	}
	if !ready {
		return errors.New("import requires a collection with no items")
	}

	// Stream the upload to a temp file so we can use archive/zip's seek API.
	bucket, err := blob.OpenBucket(ctx, s.repos.Attachments.GetConnString())
	if err != nil {
		return fmt.Errorf("open bucket: %w", err)
	}
	defer func() { _ = bucket.Close() }()

	r, err := bucket.NewReader(ctx, s.repos.Attachments.GetFullPath(uploadKey), nil)
	if err != nil {
		return fmt.Errorf("open upload: %w", err)
	}
	defer func() { _ = r.Close() }()

	tmp, err := os.CreateTemp("", "homebox-import-*.zip")
	if err != nil {
		return fmt.Errorf("create temp: %w", err)
	}
	tmpPath := tmp.Name()
	defer func() {
		_ = tmp.Close()
		_ = os.Remove(tmpPath)
	}()
	size, err := io.Copy(tmp, r)
	if err != nil {
		return fmt.Errorf("download upload: %w", err)
	}

	zr, err := zip.NewReader(tmp, size)
	if err != nil {
		return fmt.Errorf("open zip: %w", err)
	}

	mf, err := readManifest(zr)
	if err != nil {
		return fmt.Errorf("read manifest: %w", err)
	}
	if mf.SchemaVersion != ExportSchemaVersion {
		return fmt.Errorf("unsupported schema version %d (this server expects %d)", mf.SchemaVersion, ExportSchemaVersion)
	}

	// Wipe the seeded defaults (locations, tags, entity_types, notifiers,
	// etc.) so the imported collection isn't mixed with the auto-created
	// starter content. The empty-group precondition above guarantees this is
	// safe — there are no user-created items to lose.
	dbSqlWipe := s.db.Sql()
	if err := wipeGroup(ctx, dbSqlWipe, s.dialect, gid); err != nil {
		return fmt.Errorf("wipe before import: %w", err)
	}

	// idMap[table][oldID] = newID. Every PK is regenerated on import so that
	// re-importing the same export never collides with itself, and so that
	// importing into a server that already saw this data also works.
	idMap := make(map[string]map[string]string)
	rememberID := func(table, oldID, newID string) {
		if _, ok := idMap[table]; !ok {
			idMap[table] = make(map[string]string)
		}
		idMap[table][oldID] = newID
	}

	// remapFK looks up an old FK value in the id map for its target table
	// and returns the substituted new value, or the original if unknown
	// (which will surface as a FK violation on insert — better to fail loud
	// than silently null it out).
	remapFK := func(target string, v any) any {
		if v == nil {
			return nil
		}
		s := fmt.Sprint(v)
		if s == "" {
			return nil
		}
		if mapping, ok := idMap[target]; ok {
			if newID, found := mapping[s]; found {
				return newID
			}
		}
		return v
	}

	type deferredUpdate struct {
		table       string
		col         string
		newID       string
		oldFKValue  string
		targetTable string
	}
	var deferred []deferredUpdate

	dbSql := dbSqlWipe
	for _, spec := range exportTables {
		rows, err := readTableJSON(zr, spec.name+".json")
		if err != nil {
			return fmt.Errorf("read %s.json: %w", spec.name, err)
		}
		for _, row := range rows {
			// 1. PK remap (skipped for junction tables with pkCol == "").
			var newID string
			if spec.pkCol != "" {
				if v, ok := row[spec.pkCol]; ok && v != nil {
					old := fmt.Sprint(v)
					newID = uuid.NewString()
					row[spec.pkCol] = newID
					rememberID(spec.name, old, newID)
				}
			}
			// 2. Group remap.
			for _, col := range spec.groupCols {
				if _, ok := row[col]; ok {
					row[col] = gid.String()
				}
			}
			// 3. User remap (notifiers).
			for _, col := range spec.userCols {
				if _, ok := row[col]; ok {
					row[col] = userID.String()
				}
			}
			// 4. Immediate FK remap.
			for col, target := range spec.fkCols {
				if v, ok := row[col]; ok {
					row[col] = remapFK(target, v)
				}
			}
			// 5. Deferred FK: stash old value, null on insert, patch later.
			for col, target := range spec.deferCols {
				if v, ok := row[col]; ok && v != nil && v != "" {
					if newID != "" {
						deferred = append(deferred, deferredUpdate{
							table:       spec.name,
							col:         col,
							newID:       newID,
							oldFKValue:  fmt.Sprint(v),
							targetTable: target,
						})
					}
					row[col] = nil
				}
			}
			// 6. Blob path rewrite. Attachment paths are
			// "{group_id}/documents/{hash}"; the source gid prefix needs to
			// become the destination gid so the row points at where we will
			// actually upload the blob, and so cascade-cleanup on group
			// delete sweeps it correctly.
			if spec.name == "attachments" {
				if v, ok := row["path"]; ok {
					if str, ok := v.(string); ok && str != "" {
						row["path"] = rewriteBlobPath(str, mf.GroupID, gid)
					}
				}
			}
			if err := insertRow(ctx, dbSql, s.dialect, spec.name, row); err != nil {
				return fmt.Errorf("insert %s: %w", spec.name, err)
			}
		}
	}

	// Apply deferred updates (self-referential and forward-circular FKs).
	for _, d := range deferred {
		newFK := remapFK(d.targetTable, d.oldFKValue)
		q := fmt.Sprintf("UPDATE %s SET %s = %s WHERE id = %s",
			d.table, d.col, placeholder(s.dialect, 1), placeholder(s.dialect, 2))
		if _, err := dbSql.ExecContext(ctx, q, newFK, d.newID); err != nil {
			return fmt.Errorf("deferred update %s.%s: %w", d.table, d.col, err)
		}
	}

	// Restore attachment blobs. The zip names them attachments/{old_uuid};
	// look up the new attachment row through the id map.
	if err := s.restoreAttachmentBlobs(ctx, zr, idMap["attachments"]); err != nil {
		return fmt.Errorf("restore attachments: %w", err)
	}

	// Notify the frontend that lots of things just appeared.
	if s.bus != nil {
		s.bus.Publish(eventbus.EventEntityMutation, eventbus.GroupMutationEvent{GID: gid})
		s.bus.Publish(eventbus.EventTagMutation, eventbus.GroupMutationEvent{GID: gid})
	}
	return nil
}

// restoreAttachmentBlobs iterates attachments/* in the zip and writes each
// file to blob storage at the path recorded on the matching attachment row.
// Filenames in the zip use the source-side attachment UUID; idMap translates
// to the new UUID assigned during the row import.
func (s *ExportService) restoreAttachmentBlobs(ctx context.Context, zr *zip.Reader, idMap map[string]string) error {
	bucket, err := blob.OpenBucket(ctx, s.repos.Attachments.GetConnString())
	if err != nil {
		return err
	}
	defer func() { _ = bucket.Close() }()

	for _, f := range zr.File {
		if !strings.HasPrefix(f.Name, attachmentsDir) || f.FileInfo().IsDir() {
			continue
		}
		oldIDStr := strings.TrimPrefix(f.Name, attachmentsDir)
		newIDStr, ok := idMap[oldIDStr]
		if !ok {
			log.Warn().Str("name", f.Name).Msg("import: no attachment row matches blob, skipping")
			continue
		}
		id, err := uuid.Parse(newIDStr)
		if err != nil {
			log.Warn().Str("name", f.Name).Msg("import: remapped attachment id is not a uuid")
			continue
		}
		att, err := s.db.Attachment.Get(ctx, id)
		if err != nil {
			log.Warn().Err(err).Stringer("attachment_id", id).Msg("import: attachment row missing for blob")
			continue
		}
		zf, err := f.Open()
		if err != nil {
			return err
		}
		w, err := bucket.NewWriter(ctx, s.repos.Attachments.GetFullPath(att.Path), &blob.WriterOptions{
			ContentType: att.MimeType,
		})
		if err != nil {
			_ = zf.Close()
			return err
		}
		if _, err := io.Copy(w, zf); err != nil {
			_ = w.Close()
			_ = zf.Close()
			return err
		}
		if err := w.Close(); err != nil {
			_ = zf.Close()
			return err
		}
		_ = zf.Close()
	}
	return nil
}

// deleteUpload removes the staged import zip from blob storage.
func (s *ExportService) deleteUpload(ctx context.Context, uploadKey string) error {
	bucket, err := blob.OpenBucket(ctx, s.repos.Attachments.GetConnString())
	if err != nil {
		return err
	}
	defer func() { _ = bucket.Close() }()
	return bucket.Delete(ctx, s.repos.Attachments.GetFullPath(uploadKey))
}

func (s *ExportService) publishImportFinished(gid uuid.UUID) {
	if s.bus != nil {
		s.bus.Publish(eventbus.EventImportMutation, eventbus.GroupMutationEvent{GID: gid})
	}
}

// readManifest pulls and parses manifest.json out of the zip.
func readManifest(zr *zip.Reader) (Manifest, error) {
	var mf Manifest
	for _, f := range zr.File {
		if f.Name != manifestFile {
			continue
		}
		r, err := f.Open()
		if err != nil {
			return mf, err
		}
		defer func() { _ = r.Close() }()
		return mf, json.NewDecoder(r).Decode(&mf)
	}
	return mf, errors.New("manifest.json missing from zip")
}

// readTableJSON loads a single table file from the zip, tolerating its
// absence (returns an empty slice — exports may legitimately omit a table
// with zero rows in future versions).
func readTableJSON(zr *zip.Reader, name string) ([]map[string]any, error) {
	for _, f := range zr.File {
		if f.Name != name {
			continue
		}
		r, err := f.Open()
		if err != nil {
			return nil, err
		}
		defer func() { _ = r.Close() }()
		var out []map[string]any
		if err := json.NewDecoder(r).Decode(&out); err != nil {
			return nil, err
		}
		return out, nil
	}
	return nil, nil
}

// insertRow builds and runs an INSERT for one row's worth of column-value
// pairs. Self-maintaining: every JSON key becomes a column.
func insertRow(ctx context.Context, db *sql.DB, dialect, table string, row map[string]any) error {
	if len(row) == 0 {
		return nil
	}
	cols := make([]string, 0, len(row))
	for k := range row {
		cols = append(cols, k)
	}
	// Stable column order so generated SQL is deterministic in tests/logs.
	sortStrings(cols)

	args := make([]any, 0, len(cols))
	placeholders := make([]string, 0, len(cols))
	for i, c := range cols {
		args = append(args, row[c])
		placeholders = append(placeholders, placeholder(dialect, i+1))
	}

	q := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
		quoteIdent(dialect, table),
		joinQuoted(dialect, cols),
		strings.Join(placeholders, ", "),
	)
	_, err := db.ExecContext(ctx, q, args...)
	return err
}

// placeholder returns the dialect-specific positional placeholder.
func placeholder(dialect string, n int) string {
	if dialect == "postgres" {
		return fmt.Sprintf("$%d", n)
	}
	return "?"
}

// quoteIdent quotes an identifier. Both supported dialects accept double
// quotes around identifiers — including sqlite for reserved words like
// "primary" on the attachments table.
func quoteIdent(_ string, ident string) string {
	return `"` + ident + `"`
}

func joinQuoted(dialect string, cols []string) string {
	out := make([]string, len(cols))
	for i, c := range cols {
		out[i] = quoteIdent(dialect, c)
	}
	return strings.Join(out, ", ")
}

// rewriteBlobPath swaps the leading "{srcGid}/" segment of an attachment's
// blob key for "{dstGid}/". Anything else (including paths without that
// prefix) is returned unchanged so we never mangle data that happens to
// already point at the destination, or paths from a future scheme that
// doesn't lead with the gid.
func rewriteBlobPath(path string, srcGid, dstGid uuid.UUID) string {
	prefix := srcGid.String() + "/"
	if !strings.HasPrefix(path, prefix) {
		return path
	}
	return dstGid.String() + "/" + strings.TrimPrefix(path, prefix)
}

// wipeGroup deletes every group-scoped row in the export table list, in
// reverse dependency order. Used before an import so the seeded
// defaults don't pollute the restored collection.
//
// Reusing exportTables means new tables are wiped automatically once they're
// added to the export schema — no separate list to keep in sync.
func wipeGroup(ctx context.Context, db *sql.DB, dialect string, gid uuid.UUID) error {
	for i := len(exportTables) - 1; i >= 0; i-- {
		spec := exportTables[i]
		if spec.scope == "" {
			continue
		}
		q := "DELETE FROM " + quoteIdent(dialect, spec.name) +
			" WHERE " + rebindPlaceholders(spec.scope, dialect)
		args := make([]any, 0, strings.Count(spec.scope, "?"))
		for j := 0; j < cap(args); j++ {
			args = append(args, gid.String())
		}
		if _, err := db.ExecContext(ctx, q, args...); err != nil {
			return fmt.Errorf("wipe %s: %w", spec.name, err)
		}
	}
	return nil
}

// sortStrings is a tiny inlined sort to keep the file dependency-light.
func sortStrings(s []string) {
	for i := 1; i < len(s); i++ {
		for j := i; j > 0 && s[j-1] > s[j]; j-- {
			s[j], s[j-1] = s[j-1], s[j]
		}
	}
}
