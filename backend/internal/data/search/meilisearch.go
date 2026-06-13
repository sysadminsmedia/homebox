package search

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/meilisearch/meilisearch-go"
	"github.com/rs/zerolog/log"
	"github.com/sysadminsmedia/homebox/backend/internal/core/services/reporting/eventbus"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/entity"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/group"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/predicate"
	"github.com/sysadminsmedia/homebox/backend/internal/sys/config"
)

const (
	// meiliTaskPollInterval is how often task completion is polled while
	// indexing.
	meiliTaskPollInterval = 50 * time.Millisecond

	// meiliReindexBatch is the number of entities loaded from the database
	// and pushed to Meilisearch per request during reindexing.
	meiliReindexBatch = 1000

	// meiliFieldFacetPrefix namespaces the per-field facet attributes
	// (field_facets.<field name>). Each custom field becomes its own facet
	// under this prefix so the search UI can offer an independent value filter
	// per field. See FieldFacets/SearchFieldValues.
	meiliFieldFacetPrefix = "field_facets"
)

// meiliReindexDebounce coalesces bursts of mutation events (e.g. a CSV
// import) into a single reindex per group. Variable so tests can shorten it.
var meiliReindexDebounce = 2 * time.Second

// meiliDocument is the shape of an entity stored in the Meilisearch index.
// It mirrors the surfaces the database engine searches: the entity columns in
// entityColumns plus tag names and custom field text values.
//
// Custom fields are stored twice, for two different jobs:
//   - Fields (an array) is searchable, so a field's value matches free-text
//     queries just like the database engine.
//   - FieldFacets (an object keyed by field name) is filterable, so each field
//     becomes its own facet — "Special Field" and "Color" are independent
//     filters rather than one undifferentiated bucket of values.
type meiliDocument struct {
	ID           string            `json:"id"`
	GroupID      string            `json:"group_id"`
	Name         string            `json:"name"`
	Description  string            `json:"description"`
	SerialNumber string            `json:"serial_number"`
	ModelNumber  string            `json:"model_number"`
	Manufacturer string            `json:"manufacturer"`
	Notes        string            `json:"notes"`
	PurchaseFrom string            `json:"purchase_from"`
	Tags         []string          `json:"tags"`
	Fields       []meiliField      `json:"fields"`
	FieldFacets  map[string]string `json:"field_facets"`
}

// meiliField is a custom field on an entity. The name is stored for
// inspectability but only the value is searchable, matching the database
// engine's behavior.
type meiliField struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// MeilisearchEngine implements Engine backed by an external Meilisearch
// instance, providing typo-tolerant ("fuzzy") relevance-ranked matching.
//
// Queries are sent to Meilisearch scoped to the group and the matching entity
// IDs are returned as an entity.IDIn predicate, which the repository then
// intersects with its own filters and pagination. Because of that
// intersection, documents that linger in the index after an entity is deleted
// can never surface in results — index maintenance only has to guarantee that
// *existing* entities are indexed, which keeps it simple:
//
//   - the full index is rebuilt (upserted) in the background at startup, and
//   - entity/tag mutation events trigger a debounced reindex of the affected
//     group, which also prunes that group's stale documents.
//
// Results are capped at MaxHits (HBOX_SEARCH_MEILISEARCH_MAX_HITS); a search
// that legitimately matches more entities than that is truncated.
type MeilisearchEngine struct {
	client  meilisearch.ServiceManager
	index   meilisearch.IndexManager
	db      *ent.Client
	maxHits int64

	mu      sync.Mutex
	pending map[uuid.UUID]struct{}
	timer   *time.Timer
}

// NewMeilisearchEngine connects to Meilisearch, ensures the index and its
// settings exist, subscribes to mutation events for incremental indexing, and
// kicks off a full reindex in the background. It fails fast when the instance
// is unreachable so a misconfiguration is caught at startup.
func NewMeilisearchEngine(cfg config.MeilisearchConf, db *ent.Client, bus *eventbus.EventBus) (*MeilisearchEngine, error) {
	client := meilisearch.New(cfg.Host, meilisearch.WithAPIKey(cfg.APIKey))

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if _, err := client.HealthWithContext(ctx); err != nil {
		return nil, fmt.Errorf("meilisearch is not reachable at %s: %w", cfg.Host, err)
	}

	e := &MeilisearchEngine{
		client:  client,
		index:   client.Index(cfg.Index),
		db:      db,
		maxHits: cfg.MaxHits,
		pending: map[uuid.UUID]struct{}{},
	}

	if err := e.ensureIndex(ctx, cfg.Index); err != nil {
		return nil, err
	}

	if bus != nil {
		onMutation := func(data any) {
			if event, ok := data.(eventbus.GroupMutationEvent); ok {
				e.scheduleReindex(event.GID)
			}
		}
		bus.Subscribe(eventbus.EventEntityMutation, onMutation)
		bus.Subscribe(eventbus.EventTagMutation, onMutation)
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
		defer cancel()
		if err := e.ReindexAll(ctx); err != nil {
			log.Error().Err(err).Msg("meilisearch: initial reindex failed; results may be incomplete until entities are modified or the server restarts")
		}
	}()

	return e, nil
}

// Predicate implements Engine.
func (e *MeilisearchEngine) Predicate(ctx context.Context, gid uuid.UUID, query string) (predicate.Entity, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, nil
	}

	resp, err := e.index.SearchWithContext(ctx, query, &meilisearch.SearchRequest{
		// uuid.String() emits only [0-9a-f-], so inlining it in the filter
		// expression is safe
		Filter:               fmt.Sprintf("group_id = %q", gid.String()),
		Limit:                e.maxHits,
		AttributesToRetrieve: []string{"id"},
		// require every term to match, mirroring the database engine's
		// AND-of-tokens semantics (typo tolerance still applies per term)
		MatchingStrategy: meilisearch.All,
	})
	if err != nil {
		return nil, fmt.Errorf("meilisearch query failed: %w", err)
	}

	ids := make([]uuid.UUID, 0, len(resp.Hits))
	for _, hit := range resp.Hits {
		var doc struct {
			ID string `json:"id"`
		}
		if err := hit.DecodeInto(&doc); err != nil {
			return nil, fmt.Errorf("meilisearch returned an undecodable hit: %w", err)
		}
		id, err := uuid.Parse(doc.ID)
		if err != nil {
			log.Warn().Str("id", doc.ID).Msg("meilisearch: skipping hit with non-uuid id")
			continue
		}
		ids = append(ids, id)
	}

	// entity.IDIn with no ids compiles to FALSE: no matches
	return entity.IDIn(ids...), nil
}

// SearchTags returns the tag values used within a group, ranked by how many
// entities carry each tag, optionally narrowed to those matching query (a
// prefix/substring of the tag name). It powers the tag filter in the search UI:
// an empty query lists a group's most-used tags, and a non-empty one
// autocompletes as the user types.
//
// This relies on tags being a filterable, facet-searched attribute (see
// ensureIndex). Unlike Predicate it does not return entities; the UI feeds the
// chosen tags back into its own tag filter.
func (e *MeilisearchEngine) SearchTags(ctx context.Context, gid uuid.UUID, query string) ([]TagFacet, error) {
	raw, err := e.index.FacetSearchWithContext(ctx, &meilisearch.FacetSearchRequest{
		FacetName:  "tags",
		FacetQuery: strings.TrimSpace(query),
		// uuid.String() emits only [0-9a-f-], so inlining it is safe
		Filter: fmt.Sprintf("group_id = %q", gid.String()),
	})
	if err != nil {
		return nil, fmt.Errorf("meilisearch facet search failed: %w", err)
	}

	var resp struct {
		FacetHits []TagFacet `json:"facetHits"`
	}
	if err := json.Unmarshal(*raw, &resp); err != nil {
		return nil, fmt.Errorf("meilisearch returned an undecodable facet response: %w", err)
	}
	return resp.FacetHits, nil
}

// FieldFacets returns every custom field present on a group's entities mapped
// to its value distribution (value -> entity count). It drives the search UI's
// filter sidebar: which fields can be filtered on and what values each one
// currently has. Values within a field are unordered.
//
// It reads the facet distribution of every field_facets.<name> attribute in a
// single request, so the UI need not know the field names in advance.
func (e *MeilisearchEngine) FieldFacets(ctx context.Context, gid uuid.UUID) (map[string][]FieldFacet, error) {
	resp, err := e.index.SearchWithContext(ctx, "", &meilisearch.SearchRequest{
		// uuid.String() emits only [0-9a-f-], so inlining it is safe
		Filter: fmt.Sprintf("group_id = %q", gid.String()),
		// "*" expands to every filterable attribute; we keep only the
		// field_facets.<name> ones below. Distribution is computed over the
		// whole filtered set, independent of the (unused) hit page.
		Facets:               []string{"*"},
		AttributesToRetrieve: []string{"id"},
	})
	if err != nil {
		return nil, fmt.Errorf("meilisearch field facet distribution failed: %w", err)
	}
	if len(resp.FacetDistribution) == 0 {
		return map[string][]FieldFacet{}, nil
	}

	var dist map[string]map[string]int
	if err := json.Unmarshal(resp.FacetDistribution, &dist); err != nil {
		return nil, fmt.Errorf("meilisearch returned an undecodable facet distribution: %w", err)
	}

	prefix := meiliFieldFacetPrefix + "."
	out := make(map[string][]FieldFacet)
	for attr, values := range dist {
		name, ok := strings.CutPrefix(attr, prefix)
		if !ok {
			continue // group_id, tags, the empty field_facets parent, ...
		}
		// "*" enumerates every field_facets.<name> in the whole index, so a
		// field used only by *other* groups shows up here with an empty
		// distribution (the group filter zeroes its counts). Skipping empties
		// is what scopes the result to fields this group actually uses.
		if len(values) == 0 {
			continue
		}
		facets := make([]FieldFacet, 0, len(values))
		for v, c := range values {
			facets = append(facets, FieldFacet{Value: v, Count: c})
		}
		out[name] = facets
	}
	return out, nil
}

// SearchFieldValues returns the distinct values of a single custom field within
// a group, ranked by how many entities carry each value and optionally narrowed
// to those whose value matches query (a prefix/substring). It powers a per-field
// value picker in the search UI, e.g. field "Special Field" -> [{"Clean",12}].
//
// Like SearchTags it only enumerates facet values; applying the chosen filter
// to the result set remains the repository's job.
func (e *MeilisearchEngine) SearchFieldValues(ctx context.Context, gid uuid.UUID, field, query string) ([]FieldFacet, error) {
	raw, err := e.index.FacetSearchWithContext(ctx, &meilisearch.FacetSearchRequest{
		FacetName:  meiliFieldFacetPrefix + "." + field,
		FacetQuery: strings.TrimSpace(query),
		// uuid.String() emits only [0-9a-f-], so inlining it is safe
		Filter: fmt.Sprintf("group_id = %q", gid.String()),
	})
	if err != nil {
		return nil, fmt.Errorf("meilisearch field facet search failed: %w", err)
	}

	var resp struct {
		FacetHits []FieldFacet `json:"facetHits"`
	}
	if err := json.Unmarshal(*raw, &resp); err != nil {
		return nil, fmt.Errorf("meilisearch returned an undecodable facet response: %w", err)
	}
	return resp.FacetHits, nil
}

// ensureIndex creates the index (ignoring "already exists") and applies the
// searchable/filterable attribute settings.
//
// tags is both searchable (so a tag name matches in free-text queries) and
// filterable. Filterability is what makes it a facet: it lets the index be
// queried for the tag values present in a group and narrowed by tag via the
// facet-search endpoint (see SearchTags). The forthcoming "e-commerce" search
// UI builds its tag filter from those facets, so facetSearch is enabled here.
func (e *MeilisearchEngine) ensureIndex(ctx context.Context, uid string) error {
	task, err := e.client.CreateIndexWithContext(ctx, &meilisearch.IndexConfig{Uid: uid, PrimaryKey: "id"})
	if err != nil {
		return fmt.Errorf("meilisearch create index: %w", err)
	}
	done, err := e.client.WaitForTaskWithContext(ctx, task.TaskUID, meiliTaskPollInterval)
	if err != nil {
		return fmt.Errorf("meilisearch create index: %w", err)
	}
	if done.Status == meilisearch.TaskStatusFailed && done.Error.Code != "index_already_exists" {
		return fmt.Errorf("meilisearch create index: %s", done.Error.Message)
	}

	task, err = e.index.UpdateSettingsWithContext(ctx, &meilisearch.Settings{
		SearchableAttributes: []string{
			"name", "description", "serial_number", "model_number",
			"manufacturer", "notes", "purchase_from", "tags", "fields.value",
		},
		// group_id scopes every query; tags and the per-field facets under
		// field_facets.* are faceted for the tag/field filter UI. Marking the
		// field_facets parent filterable makes every nested field_facets.<name>
		// a facet without having to enumerate field names up front.
		FilterableAttributes: []string{"group_id", "tags", meiliFieldFacetPrefix},
		// facet search is disabled by default in Meilisearch >= 1.12; enable it
		// so SearchTags/SearchFieldValues can resolve/autocomplete facet values.
		FacetSearch: true,
	})
	if err != nil {
		return fmt.Errorf("meilisearch update settings: %w", err)
	}
	if err := e.waitForTask(ctx, task, "update settings"); err != nil {
		return err
	}
	return nil
}

// ReindexAll rebuilds the documents for every entity in the database. Existing
// documents are upserted in place, so search keeps working while it runs.
func (e *MeilisearchEngine) ReindexAll(ctx context.Context) error {
	return e.reindex(ctx, nil)
}

// ReindexGroup rebuilds the documents for a single group and prunes documents
// for entities that no longer exist in it.
func (e *MeilisearchEngine) ReindexGroup(ctx context.Context, gid uuid.UUID) error {
	return e.reindex(ctx, &gid)
}

// reindex upserts documents for all entities (gid == nil) or one group's
// entities, then deletes that scope's documents that no longer correspond to
// a database row.
func (e *MeilisearchEngine) reindex(ctx context.Context, gid *uuid.UUID) error {
	q := e.db.Entity.Query().
		WithGroup().
		WithTag().
		WithFields().
		Order(ent.Asc(entity.FieldID))
	if gid != nil {
		q = q.Where(entity.HasGroupWith(group.ID(*gid)))
	}

	indexed := make(map[string]struct{})
	for offset := 0; ; offset += meiliReindexBatch {
		entities, err := q.Clone().Offset(offset).Limit(meiliReindexBatch).All(ctx)
		if err != nil {
			return fmt.Errorf("meilisearch reindex: loading entities: %w", err)
		}
		if len(entities) == 0 {
			break
		}

		docs := make([]meiliDocument, 0, len(entities))
		for _, row := range entities {
			doc := buildMeiliDocument(row)
			docs = append(docs, doc)
			indexed[doc.ID] = struct{}{}
		}

		task, err := e.index.AddDocumentsWithContext(ctx, docs, nil)
		if err != nil {
			return fmt.Errorf("meilisearch reindex: adding documents: %w", err)
		}
		if err := e.waitForTask(ctx, task, "add documents"); err != nil {
			return err
		}

		if len(entities) < meiliReindexBatch {
			break
		}
	}

	return e.pruneStale(ctx, gid, indexed)
}

// pruneStale removes documents within the reindexed scope whose entity no
// longer exists. Stale documents are harmless for correctness (the predicate
// is intersected with the live database), so pruning is best-effort hygiene.
func (e *MeilisearchEngine) pruneStale(ctx context.Context, gid *uuid.UUID, indexed map[string]struct{}) error {
	queryFields := []string{"id"}
	dq := &meilisearch.DocumentsQuery{Fields: queryFields, Limit: meiliReindexBatch}
	if gid != nil {
		dq.Filter = fmt.Sprintf("group_id = %q", gid.String())
	}

	var stale []string
	for offset := int64(0); ; offset += meiliReindexBatch {
		dq.Offset = offset
		var page meilisearch.DocumentsResult
		if err := e.index.GetDocumentsWithContext(ctx, dq, &page); err != nil {
			return fmt.Errorf("meilisearch reindex: listing documents: %w", err)
		}
		for _, hit := range page.Results {
			var doc struct {
				ID string `json:"id"`
			}
			if err := hit.DecodeInto(&doc); err != nil {
				continue
			}
			if _, ok := indexed[doc.ID]; !ok {
				stale = append(stale, doc.ID)
			}
		}
		if int64(len(page.Results)) < meiliReindexBatch {
			break
		}
	}

	if len(stale) == 0 {
		return nil
	}

	task, err := e.index.DeleteDocumentsWithContext(ctx, stale, nil)
	if err != nil {
		return fmt.Errorf("meilisearch reindex: deleting stale documents: %w", err)
	}
	return e.waitForTask(ctx, task, "delete stale documents")
}

// scheduleReindex queues a group for reindexing, coalescing rapid mutation
// bursts into one pass per group.
func (e *MeilisearchEngine) scheduleReindex(gid uuid.UUID) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.pending[gid] = struct{}{}
	if e.timer == nil {
		e.timer = time.AfterFunc(meiliReindexDebounce, e.flushPending)
	}
}

func (e *MeilisearchEngine) flushPending() {
	e.mu.Lock()
	gids := make([]uuid.UUID, 0, len(e.pending))
	for gid := range e.pending {
		gids = append(gids, gid)
	}
	e.pending = map[uuid.UUID]struct{}{}
	e.timer = nil
	e.mu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	for _, gid := range gids {
		if err := e.ReindexGroup(ctx, gid); err != nil {
			log.Error().Err(err).Str("group_id", gid.String()).Msg("meilisearch: group reindex failed")
		}
	}
}

func (e *MeilisearchEngine) waitForTask(ctx context.Context, task *meilisearch.TaskInfo, op string) error {
	done, err := e.client.WaitForTaskWithContext(ctx, task.TaskUID, meiliTaskPollInterval)
	if err != nil {
		return fmt.Errorf("meilisearch %s: %w", op, err)
	}
	if done.Status == meilisearch.TaskStatusFailed {
		return fmt.Errorf("meilisearch %s: %s", op, done.Error.Message)
	}
	return nil
}

func buildMeiliDocument(e *ent.Entity) meiliDocument {
	doc := meiliDocument{
		ID:           e.ID.String(),
		Name:         e.Name,
		Description:  e.Description,
		SerialNumber: e.SerialNumber,
		ModelNumber:  e.ModelNumber,
		Manufacturer: e.Manufacturer,
		Notes:        e.Notes,
		PurchaseFrom: e.PurchaseFrom,
		// empty slices/maps (not nil) so documents serialize as []/{} not null
		Tags:        []string{},
		Fields:      []meiliField{},
		FieldFacets: map[string]string{},
	}
	if e.Edges.Group != nil {
		doc.GroupID = e.Edges.Group.ID.String()
	}
	for _, t := range e.Edges.Tag {
		doc.Tags = append(doc.Tags, t.Name)
	}
	for _, f := range e.Edges.Fields {
		if f.TextValue != "" {
			doc.Fields = append(doc.Fields, meiliField{Name: f.Name, Value: f.TextValue})
			// last value wins if a field name repeats on one entity; a facet
			// only needs one value per (entity, field) anyway
			doc.FieldFacets[f.Name] = f.TextValue
		}
	}
	return doc
}
