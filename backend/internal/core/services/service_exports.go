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
	"path"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel"
	"gocloud.dev/blob"
	"gocloud.dev/pubsub"

	"github.com/sysadminsmedia/homebox/backend/internal/core/services/reporting/eventbus"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/entity"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/entitytemplate"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/entitytype"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/group"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/notifier"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/tag"
	"github.com/sysadminsmedia/homebox/backend/internal/data/repo"
	"github.com/sysadminsmedia/homebox/backend/internal/sys/config"
	"github.com/sysadminsmedia/homebox/backend/pkgs/utils"
)

// ExportSchemaVersion is the on-disk version of the export zip layout.
// Bump this when manifest/file shapes change in incompatible ways and import
// can no longer round-trip an older export.
const ExportSchemaVersion = 1

// entitiesTable is the on-disk name of the entities table. Hoisted out of
// the exportTables literal so the same string isn't repeated across every
// FK/scope reference that points back at it.
const entitiesTable = "entities"

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
		deferCols: map[string]string{"entity_template_location": entitiesTable},
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
		name:      entitiesTable,
		scope:     "group_entities = ?",
		pkCol:     "id",
		groupCols: []string{"group_entities"},
		fkCols:    map[string]string{"entity_type_entities": "entity_types"},
		deferCols: map[string]string{"entity_children": entitiesTable},
	},
	{
		name:   "entity_fields",
		scope:  "entity_fields IN (SELECT id FROM entities WHERE group_entities = ?)",
		pkCol:  "id",
		fkCols: map[string]string{"entity_fields": entitiesTable},
	},
	{
		name:   "maintenance_entries",
		scope:  "entity_id IN (SELECT id FROM entities WHERE group_entities = ?)",
		pkCol:  "id",
		fkCols: map[string]string{"entity_id": entitiesTable},
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
		fkCols:    map[string]string{"entity_attachments": entitiesTable},
		deferCols: map[string]string{"attachment_thumbnail": "attachments"},
	},
	{
		name:   "tag_entities",
		scope:  "tag_id IN (SELECT id FROM tags WHERE group_tags = ?)",
		fkCols: map[string]string{"tag_id": "tags", "entity_id": entitiesTable},
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

	// topics caches the publisher topic per topic name so it is opened once
	// and reused for the lifetime of the process. Publishers must never call
	// Shutdown on these: the default mem:// driver returns a shared singleton
	// per URL, so shutting it down after one send permanently breaks every
	// later publish until restart (#1592).
	topicsMu sync.Mutex
	topics   map[string]*pubsub.Topic
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
		_ = s.repos.Exports.SetFailed(ctx, gid, out.ID, "failed to enqueue: "+err.Error())
		return out, err
	}

	s.publishMutation(gid)
	return out, nil
}

// EnqueueImport creates a tracked import row pointing at the zip already
// staged at uploadKey and publishes a job for the worker to pick up. The
// returned row carries the ID the frontend can poll for progress.
// uploadKey must live under "{gid}/imports/" — the worker re-validates
// this before reading.
func (s *ExportService) EnqueueImport(ctx context.Context, gid uuid.UUID, userID uuid.UUID, uploadKey string, sizeBytes int64) (repo.ExportOut, error) {
	ctx, span := otel.Tracer("services").Start(ctx, "ExportService.EnqueueImport")
	defer span.End()

	row, err := s.repos.Exports.CreateImport(ctx, gid, uploadKey, sizeBytes)
	if err != nil {
		return row, err
	}

	if err := s.publishImportJob(ctx, gid, userID, row.ID); err != nil {
		// Mark the row failed so the user sees what happened instead of a
		// permanently-pending entry. Best-effort: if the SetFailed also
		// fails we still return the publish error to the caller.
		_ = s.repos.Exports.SetFailed(ctx, gid, row.ID, "failed to enqueue: "+err.Error())
		return row, err
	}
	return row, nil
}

// IsGroupReadyForImport returns true when gid contains no user-created data
// across any table that wipeGroup will delete. Default locations, tags, and
// the lazily-created "Item"/"Location" entity_types from registration are
// tolerated — the import wipes them before restoring. Any extra rows beyond
// those seed baselines, or any presence in tables that aren't seeded
// (entity_templates, notifiers), blocks the import so a one-click restore
// can't silently destroy work.
//
// The seed-baseline counts are coarse: a user who deletes some default tags
// and then adds the same number of custom tags would pass with a false
// negative. Acceptable trade-off versus adding a per-row "is_seed" flag.
//
// Tables not checked explicitly are covered transitively: template_fields
// require templates; entity_fields/attachments/maintenance_entries/tag_entities
// require entities or tags.
func (s *ExportService) IsGroupReadyForImport(ctx context.Context, gid uuid.UUID) (bool, error) {
	items, err := s.db.Entity.Query().Where(
		entity.HasGroupWith(group.ID(gid)),
		entity.HasEntityTypeWith(entitytype.IsLocation(false)),
	).Count(ctx)
	if err != nil {
		return false, err
	}
	if items > 0 {
		return false, nil
	}

	locations, err := s.db.Entity.Query().Where(
		entity.HasGroupWith(group.ID(gid)),
		entity.HasEntityTypeWith(entitytype.IsLocation(true)),
	).Count(ctx)
	if err != nil {
		return false, err
	}
	if locations > len(defaultLocations()) {
		return false, nil
	}

	tags, err := s.db.Tag.Query().Where(tag.HasGroupWith(group.ID(gid))).Count(ctx)
	if err != nil {
		return false, err
	}
	if tags > len(defaultTags()) {
		return false, nil
	}

	// Entity types are lazily created with names "Item" and "Location" the
	// first time GetDefault is called for each. Anything beyond those two
	// implies a user-customized type.
	const defaultEntityTypeCount = 2
	entityTypes, err := s.db.EntityType.Query().Where(entitytype.HasGroupWith(group.ID(gid))).Count(ctx)
	if err != nil {
		return false, err
	}
	if entityTypes > defaultEntityTypeCount {
		return false, nil
	}

	templates, err := s.db.EntityTemplate.Query().Where(entitytemplate.HasGroupWith(group.ID(gid))).Count(ctx)
	if err != nil {
		return false, err
	}
	if templates > 0 {
		return false, nil
	}

	notifiers, err := s.db.Notifier.Query().Where(notifier.HasGroupWith(group.ID(gid))).Count(ctx)
	if err != nil {
		return false, err
	}
	if notifiers > 0 {
		return false, nil
	}

	return true, nil
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

	if err := s.repos.Exports.SetRunning(ctx, gid, exportID); err != nil {
		log.Err(err).Msg("export job: failed to mark running")
		return
	}
	s.publishMutation(gid)

	artifactPath, sizeBytes, err := s.buildArtifact(ctx, exportID, gid)
	if err != nil {
		log.Err(err).Stringer("export_id", exportID).Msg("export job: failed")
		_ = s.repos.Exports.SetFailed(ctx, gid, exportID, err.Error())
		s.publishMutation(gid)
		return
	}

	if err := s.repos.Exports.SetCompleted(ctx, gid, exportID, artifactPath, sizeBytes); err != nil {
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
		_ = s.repos.Exports.SetProgress(ctx, gid, exportID, pct)
	}

	// Copy attachment blobs into the zip.
	if err := s.copyAttachmentBlobs(ctx, zw, gid); err != nil {
		_ = zw.Close()
		return "", 0, fmt.Errorf("copy attachments: %w", err)
	}
	_ = s.repos.Exports.SetProgress(ctx, gid, exportID, 95)

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
	defer func() { _ = rows.Close() }()
	type attRef struct{ id, path string }
	var refs []attRef
	for rows.Next() {
		var id, path string
		if err := rows.Scan(&id, &path); err != nil {
			return err
		}
		if path == "" {
			continue
		}
		refs = append(refs, attRef{id: id, path: path})
	}
	if err := rows.Err(); err != nil {
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

// openTopic returns the long-lived publisher topic for name, opening and
// caching it on first use. The topic is reused across publishes and stays
// open for the process lifetime — see the topics field doc for why a
// per-publish Shutdown must not be reintroduced here.
func (s *ExportService) openTopic(ctx context.Context, name string) (*pubsub.Topic, error) {
	s.topicsMu.Lock()
	defer s.topicsMu.Unlock()
	if topic, ok := s.topics[name]; ok {
		return topic, nil
	}
	conn, err := utils.GenerateSubPubConn(s.pubSubConn, name)
	if err != nil {
		return nil, err
	}
	topic, err := pubsub.OpenTopic(ctx, conn)
	if err != nil {
		return nil, err
	}
	if s.topics == nil {
		s.topics = make(map[string]*pubsub.Topic)
	}
	s.topics[name] = topic
	return topic, nil
}

// Shutdown flushes and closes every cached publisher topic. It is called
// once from the app's graceful-shutdown path; publishes after Shutdown
// would reopen topics, so it must only run at service teardown.
func (s *ExportService) Shutdown(ctx context.Context) error {
	s.topicsMu.Lock()
	defer s.topicsMu.Unlock()
	var errs []error
	for name, topic := range s.topics {
		if err := topic.Shutdown(ctx); err != nil {
			errs = append(errs, fmt.Errorf("shutdown topic %q: %w", name, err))
		}
	}
	s.topics = nil
	return errors.Join(errs...)
}

// publishExportJob sends a message on the export topic.
func (s *ExportService) publishExportJob(ctx context.Context, gid, exportID uuid.UUID) error {
	topic, err := s.openTopic(ctx, TopicCollectionExport)
	if err != nil {
		return err
	}
	return topic.Send(ctx, &pubsub.Message{
		Body: []byte("collection_export:" + exportID.String()),
		Metadata: map[string]string{
			"group_id":  gid.String(),
			"export_id": exportID.String(),
		},
	})
}

// publishImportJob sends a message on the import topic. The worker loads
// the tracked import row by importID, reads the staged upload from blob
// storage at the row's artifact_path, unzips, restores into the group
// identified by gid, then deletes the staged upload.
func (s *ExportService) publishImportJob(ctx context.Context, gid, userID, importID uuid.UUID) error {
	topic, err := s.openTopic(ctx, TopicCollectionImport)
	if err != nil {
		return err
	}
	return topic.Send(ctx, &pubsub.Message{
		Body: []byte("collection_import:" + gid.String()),
		Metadata: map[string]string{
			"group_id":  gid.String(),
			"user_id":   userID.String(),
			"import_id": importID.String(),
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
// is received. It loads the tracked import row, validates the staged
// upload, asserts the destination group is empty, and replays every row.
// Status/progress on the row drives the polling UI on the frontend.
func (s *ExportService) RunImport(ctx context.Context, gid, userID, importID uuid.UUID) {
	ctx, span := otel.Tracer("services").Start(ctx, "ExportService.RunImport")
	defer span.End()

	row, err := s.repos.Exports.Get(ctx, gid, importID)
	if err != nil {
		log.Err(err).Stringer("import_id", importID).Stringer("gid", gid).Msg("import job: row not found or wrong group")
		return
	}
	if row.Kind != "import" {
		log.Error().Stringer("import_id", importID).Str("kind", row.Kind).Msg("import job: row is not an import, refusing")
		return
	}
	if row.Status != "pending" {
		log.Warn().Stringer("import_id", importID).Str("status", row.Status).Msg("import job: not pending, skipping")
		return
	}
	uploadKey := row.ArtifactPath

	// Hard scope check: refuse anything that doesn't live under the caller's
	// group prefix. Defence in depth — the handler already enforced this.
	prefix := gid.String() + "/imports/"
	if !strings.HasPrefix(uploadKey, prefix) {
		log.Error().Str("upload_key", uploadKey).Stringer("gid", gid).Msg("import job: upload key outside group prefix, refusing")
		_ = s.repos.Exports.SetFailed(ctx, gid, importID, "upload outside group prefix")
		s.publishImportFinished(gid)
		return
	}

	if err := s.repos.Exports.SetRunning(ctx, gid, importID); err != nil {
		log.Err(err).Stringer("import_id", importID).Msg("import job: failed to mark running")
		return
	}
	s.publishImportFinished(gid)

	if err := s.runImport(ctx, gid, userID, importID, uploadKey); err != nil {
		log.Err(err).Stringer("gid", gid).Msg("import job: failed")
		_ = s.repos.Exports.SetFailed(ctx, gid, importID, err.Error())
	} else {
		// On success the upload zip has been fully restored; keep the row
		// size_bytes (set when the upload was staged) and just flip status.
		if err := s.repos.Exports.SetCompleted(ctx, gid, importID, uploadKey, row.SizeBytes); err != nil {
			log.Err(err).Stringer("import_id", importID).Msg("import job: failed to mark completed")
		}
	}

	// Cleanup the staging blob whether the import succeeded or not — keeping
	// it around just lets a second delivery race against the populated DB.
	if err := s.deleteUpload(ctx, uploadKey); err != nil {
		log.Warn().Err(err).Str("upload_key", uploadKey).Msg("import job: failed to clean staging upload")
	}

	s.publishImportFinished(gid)
}

func (s *ExportService) runImport(ctx context.Context, gid, userID, importID uuid.UUID, uploadKey string) error {
	// setProgress is best-effort: a failed status update is logged but never
	// aborts the import itself — progress is observability, not correctness.
	setProgress := func(pct int) {
		if err := s.repos.Exports.SetProgress(ctx, gid, importID, pct); err != nil {
			log.Warn().Err(err).Stringer("import_id", importID).Int("pct", pct).Msg("import job: failed to update progress")
		}
		s.publishImportFinished(gid)
	}

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

	if err := enforceZipUncompressedLimit(zr, size); err != nil {
		return err
	}

	mf, err := readManifest(zr)
	if err != nil {
		return fmt.Errorf("read manifest: %w", err)
	}
	if mf.SchemaVersion != ExportSchemaVersion {
		return fmt.Errorf("unsupported schema version %d (this server expects %d)", mf.SchemaVersion, ExportSchemaVersion)
	}
	// Progress budget: 0–5% download + manifest, ~5–80% reserved for the DB
	// phase (reported once after commit because intermediate setProgress
	// calls would deadlock on SQLite — the write tx holds the single
	// writer lock and ent's pool can't take it), 80–95% per-file blob
	// restore, 95–100% finalization.
	setProgress(5)

	// All DB work — the seed wipe, every row insert, and the deferred FK
	// patches — runs in a single tx so the group never sits in a half-imported
	// state. If anything below fails, the deferred Rollback unwinds the wipe
	// too. Blob uploads and bus notifications run only after Commit because
	// (a) blobs are not transactional, and (b) restoreAttachmentBlobs needs to
	// look up rows via the ent client, which uses its own pool and would not
	// see uncommitted writes under Postgres READ COMMITTED.
	tx, err := s.db.Sql().BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin import tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	// Wipe the seeded defaults (locations, tags, entity_types, notifiers,
	// etc.) so the imported collection isn't mixed with the auto-created
	// starter content. The empty-group precondition above guarantees this is
	// safe — there are no user-created items to lose.
	if err := wipeGroup(ctx, tx, s.dialect, gid); err != nil {
		return fmt.Errorf("wipe before import: %w", err)
	}

	idMap, err := s.replayImportRows(ctx, tx, zr, gid, userID, mf.GroupID)
	if err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit import: %w", err)
	}
	setProgress(80)

	// Restore attachment blobs. The zip names them attachments/{old_uuid};
	// look up the new attachment row through the id map. Must run post-commit
	// because the lookup goes through the ent client, which uses a different
	// connection than our tx.
	blobProgress := func(done, total int) {
		if total <= 0 {
			return
		}
		setProgress(80 + int(float64(done)/float64(total)*15))
	}
	if err := s.restoreAttachmentBlobs(ctx, zr, idMap["attachments"], blobProgress); err != nil {
		// Compensating cleanup. The tx is already committed, so a partial blob
		// restore leaves rows pointing at blobs that don't exist on disk and —
		// because IsGroupReadyForImport rejects non-empty groups — blocks any
		// retry. Wipe the freshly-imported rows so the group goes back to its
		// pre-import (empty) state. Successfully uploaded blobs are left on
		// disk; on retry the same content hashes will write to the same paths.
		if werr := wipeGroup(ctx, s.db.Sql(), s.dialect, gid); werr != nil {
			log.Err(werr).Stringer("gid", gid).Msg("import job: blob restore failed and rollback wipe also failed; group left in partially imported state")
		}
		return fmt.Errorf("restore attachments: %w", err)
	}
	setProgress(95)

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
// to the new UUID assigned during the row import. The optional onProgress
// callback is invoked after each blob is written so the import row's
// progress field stays current during what can be the slowest phase of a
// restore.
func (s *ExportService) restoreAttachmentBlobs(ctx context.Context, zr *zip.Reader, idMap map[string]string, onProgress func(done, total int)) error {
	bucket, err := blob.OpenBucket(ctx, s.repos.Attachments.GetConnString())
	if err != nil {
		return err
	}
	defer func() { _ = bucket.Close() }()

	// Pre-count blob entries so onProgress can report a meaningful ratio.
	total := 0
	for _, f := range zr.File {
		if strings.HasPrefix(f.Name, attachmentsDir) && !f.FileInfo().IsDir() {
			total++
		}
	}
	done := 0

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
		done++
		if onProgress != nil {
			onProgress(done, total)
		}
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

// sqlExecer is the minimal interface used by the import path so the same
// helpers work against a *sql.DB (auto-commit) and a *sql.Tx (transactional
// import). Both stdlib types implement ExecContext with this signature.
type sqlExecer interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

// insertRow builds and runs an INSERT for one row's worth of column-value
// pairs. Self-maintaining: every JSON key becomes a column.
func insertRow(ctx context.Context, db sqlExecer, dialect, table string, row map[string]any) error {
	if len(row) == 0 {
		return nil
	}
	// Reject any attacker-shaped identifiers before they reach the SQL
	// builder. Column names flow from JSON keys in an attacker-controlled
	// zip; quoteIdent also escapes embedded quotes, but rejecting up front
	// gives a clear error and keeps the SQL we generate trivial to audit.
	if !isValidSQLIdent(table) {
		return fmt.Errorf("invalid table identifier %q", table)
	}
	cols := make([]string, 0, len(row))
	for k := range row {
		if !isValidSQLIdent(k) {
			return fmt.Errorf("invalid column identifier %q on table %q", k, table)
		}
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
// "primary" on the attachments table. Any embedded double-quote is escaped
// per the SQL standard (and shared dialect behavior) by doubling it, so a
// stray quote can never close the identifier and inject SQL. Callers should
// still validate identifiers via isValidSQLIdent for attacker-supplied input;
// this escape is defence-in-depth, not the primary gate.
func quoteIdent(_ string, ident string) string {
	return `"` + strings.ReplaceAll(ident, `"`, `""`) + `"`
}

// isValidSQLIdent returns true if s is a syntactically conservative SQL
// identifier: an ASCII letter or underscore followed by letters, digits, or
// underscores. The import path runs JSON map keys through this before they
// are interpolated as column names, so a hostile export zip cannot smuggle
// SQL into a table name or column list. dumpTable populates these keys from
// rows.Columns(), which only ever returns plain identifiers, so every legit
// key satisfies this check.
func isValidSQLIdent(s string) bool {
	if s == "" {
		return false
	}
	for i, r := range s {
		switch {
		case r >= 'a' && r <= 'z',
			r >= 'A' && r <= 'Z',
			r == '_':
			// always allowed
		case (r >= '0' && r <= '9') && i > 0:
			// digits allowed anywhere except the first character
		default:
			return false
		}
	}
	return true
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

// enforceZipUncompressedLimit rejects zip bombs before any member is opened.
// Legitimate exports compress ~3-10x (JSON tables compress well, attachment
// binaries barely at all); 100x the compressed upload is a generous ceiling
// that still flags any pathological expansion ratio. Both per-entry and
// cumulative caps are checked since either alone is bypassable. The declared
// uncompressed size in the central directory is what attackers control, but
// typical bombs declare accurate-but-tiny per-entry sizes that sum to a huge
// total — the cumulative check is what stops them.
func enforceZipUncompressedLimit(zr *zip.Reader, uploadSize int64) error {
	const maxZipExpansionRatio = 100
	maxUncompressed := uint64(uploadSize) * maxZipExpansionRatio
	var total uint64
	for _, f := range zr.File {
		if f.UncompressedSize64 > maxUncompressed {
			return fmt.Errorf("import rejected: zip entry %q declares uncompressed size %d, exceeds limit %d", f.Name, f.UncompressedSize64, maxUncompressed)
		}
		if f.UncompressedSize64 > maxUncompressed-total {
			return fmt.Errorf("import rejected: zip cumulative uncompressed size exceeds limit %d", maxUncompressed)
		}
		total += f.UncompressedSize64
	}
	return nil
}

// replayImportRows reads each table file from the zip, regenerates every PK,
// remaps group/user/FK columns, rewrites attachment blob paths from the source
// gid prefix to the destination, and inserts the row into tx. Self-referential
// and forward-circular FKs are stashed and patched in a second pass so the
// first INSERT can succeed before the referenced row exists. Returns
// idMap[table][oldID]=newID so the post-commit blob restore can resolve
// attachment file names back to the just-inserted rows.
func (s *ExportService) replayImportRows(ctx context.Context, tx *sql.Tx, zr *zip.Reader, gid, userID, srcGroupID uuid.UUID) (map[string]map[string]string, error) {
	idMap := make(map[string]map[string]string)
	rememberID := func(table, oldID, newID string) {
		if _, ok := idMap[table]; !ok {
			idMap[table] = make(map[string]string)
		}
		idMap[table][oldID] = newID
	}

	// remapFK substitutes an old FK value with its remapped new value, or
	// returns the original if unknown (which surfaces as a FK violation on
	// insert — better to fail loud than silently null it out).
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
		table, col, newID, oldFKValue, targetTable string
	}
	var deferred []deferredUpdate

	for _, spec := range exportTables {
		rows, err := readTableJSON(zr, spec.name+".json")
		if err != nil {
			return nil, fmt.Errorf("read %s.json: %w", spec.name, err)
		}
		for _, row := range rows {
			newID, err := remapImportRow(row, spec, gid, userID, srcGroupID, remapFK, rememberID)
			if err != nil {
				return nil, err
			}
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
			if err := insertRow(ctx, tx, s.dialect, spec.name, row); err != nil {
				return nil, fmt.Errorf("insert %s: %w", spec.name, err)
			}
		}
	}

	// Apply deferred updates (self-referential and forward-circular FKs).
	for _, d := range deferred {
		newFK := remapFK(d.targetTable, d.oldFKValue)
		q := fmt.Sprintf("UPDATE %s SET %s = %s WHERE id = %s",
			d.table, d.col, placeholder(s.dialect, 1), placeholder(s.dialect, 2))
		if _, err := tx.ExecContext(ctx, q, newFK, d.newID); err != nil {
			return nil, fmt.Errorf("deferred update %s.%s: %w", d.table, d.col, err)
		}
	}

	return idMap, nil
}

// remapImportRow rewrites a single row in place: regenerates its PK, swaps
// group/user/FK columns, and validates+rewrites attachment blob paths from
// the source gid prefix to the destination gid. Returns the new PK (empty
// for junction tables with no pkCol) so the caller can record deferred FK
// updates against it.
func remapImportRow(
	row map[string]any,
	spec tableSpec,
	gid, userID, srcGroupID uuid.UUID,
	remapFK func(target string, v any) any,
	rememberID func(table, oldID, newID string),
) (string, error) {
	var newID string
	if spec.pkCol != "" {
		if v, ok := row[spec.pkCol]; ok && v != nil {
			old := fmt.Sprint(v)
			newID = uuid.NewString()
			row[spec.pkCol] = newID
			rememberID(spec.name, old, newID)
		}
	}
	for _, col := range spec.groupCols {
		if _, ok := row[col]; ok {
			row[col] = gid.String()
		}
	}
	for _, col := range spec.userCols {
		if _, ok := row[col]; ok {
			row[col] = userID.String()
		}
	}
	for col, target := range spec.fkCols {
		if v, ok := row[col]; ok {
			row[col] = remapFK(target, v)
		}
	}
	// Attachment paths are "{group_id}/documents/{hash}"; rewrite the source
	// gid prefix to the destination so the row points at where we will
	// actually upload the blob and so cascade-cleanup on group delete sweeps
	// it correctly.
	//
	// The zip is attacker-controlled (an admin imports a file they
	// uploaded). Without validation, a crafted path like
	// "{srcGid}/documents/../../etc/foo" would survive rewriteBlobPath
	// (it only swaps the gid prefix) and reach the blob writer; the
	// fileblob backend doesn't resolve ".." segments. Validate the
	// source shape strictly, then re-validate the result.
	if spec.name == "attachments" {
		if err := rewriteAttachmentPath(row, srcGroupID, gid); err != nil {
			return "", err
		}
	}
	return newID, nil
}

// rewriteAttachmentPath validates the attachment row's path column, swaps
// the source gid prefix for the destination gid, and re-validates the
// result. Mutates row in place.
func rewriteAttachmentPath(row map[string]any, srcGroupID, dstGroupID uuid.UUID) error {
	v, ok := row["path"]
	if !ok {
		return fmt.Errorf("attachment row missing path column")
	}
	str, ok := v.(string)
	if !ok || str == "" {
		return fmt.Errorf("attachment row has empty/non-string path")
	}
	cleanPath := path.Clean(str)
	srcPrefix := srcGroupID.String() + "/documents/"
	if !strings.HasPrefix(cleanPath, srcPrefix) {
		return fmt.Errorf("attachment path %q does not live under source group's documents prefix", str)
	}
	newPath := rewriteBlobPath(cleanPath, srcGroupID, dstGroupID)
	dstPrefix := dstGroupID.String() + "/documents/"
	if !strings.HasPrefix(newPath, dstPrefix) {
		return fmt.Errorf("rewritten attachment path %q escapes destination group's documents prefix", newPath)
	}
	row["path"] = newPath
	return nil
}

// wipeGroup deletes every group-scoped row in the export table list, in
// reverse dependency order. Used before an import so the seeded
// defaults don't pollute the restored collection.
//
// Reusing exportTables means new tables are wiped automatically once they're
// added to the export schema — no separate list to keep in sync.
func wipeGroup(ctx context.Context, db sqlExecer, dialect string, gid uuid.UUID) error {
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
