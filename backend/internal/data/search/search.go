// Package search provides the pluggable free-text search abstraction used by
// the entity repository.
//
// A search Engine translates a user-supplied query string into an ent
// predicate that selects the matching entities. The default engine
// (DriverDatabase) performs tokenized, case- and accent-insensitive matching
// directly in the database and works on both SQLite and PostgreSQL with no
// extra infrastructure. The Meilisearch engine (DriverMeilisearch) delegates
// matching to an external Meilisearch instance for typo-tolerant,
// relevance-ranked search.
//
// To add a new engine (e.g. Elasticsearch):
//
//  1. Implement the Engine interface. An external engine typically queries
//     its own index scoped to the group ID and returns
//     entity.IDIn(matchedIDs...) as the predicate, which preserves the
//     repository's filtering, pagination, and eager-loading behavior. See
//     MeilisearchEngine for the reference implementation.
//  2. Keep the engine's index up to date by subscribing to entity mutations
//     (the repositories publish events on the event bus).
//  3. Register a new driver constant and construction case in NewEngine, and
//     document the driver value for HBOX_SEARCH_DRIVER.
package search

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/sysadminsmedia/homebox/backend/internal/core/services/reporting/eventbus"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/predicate"
	"github.com/sysadminsmedia/homebox/backend/internal/sys/config"
)

// Supported search drivers.
const (
	DriverDatabase    = "database"
	DriverMeilisearch = "meilisearch"
)

// Engine translates free-text queries into entity predicates.
type Engine interface {
	// Predicate returns an ent predicate selecting the entities within the
	// given group that match the free-text query. A nil predicate (with nil
	// error) means the query has no usable terms and no search filter should
	// be applied.
	//
	// The caller is responsible for all non-search filtering (group, type,
	// tags, pagination, ...); implementations must only express the text
	// match itself.
	Predicate(ctx context.Context, gid uuid.UUID, query string) (predicate.Entity, error)
}

// TagFacet is a tag value present on a group's entities together with how many
// of them carry it, e.g. {"Electronics", 12}. The json tags decode a
// Meilisearch facet hit; the database engine populates the fields directly.
type TagFacet struct {
	Name  string `json:"value"`
	Count int    `json:"count"`
}

// FieldFacet is one value of a custom field together with the number of a
// group's entities that carry that value, e.g. {"Clean", 12}.
type FieldFacet struct {
	Value string `json:"value"`
	Count int    `json:"count"`
}

// Faceter is an optional capability for engines that can enumerate the values
// available for filtering — tag names and per-custom-field values, each with
// the number of matching entities. It backs the search UI's filter sidebar
// (filter by tag, by "Special Field = Clean", ...). Both the database and
// Meilisearch engines implement it; callers type-assert for it:
//
//	if f, ok := engine.(search.Faceter); ok {
//		tags, err := f.SearchTags(ctx, gid, "")
//	}
//
// Like Predicate, these methods only enumerate facet values; applying a chosen
// filter to the result set remains the repository's job.
type Faceter interface {
	// SearchTags returns the tag names used within a group ranked by entity
	// count, optionally narrowed to those matching query (a case-insensitive
	// substring of the tag name). An empty query returns the most-used tags.
	SearchTags(ctx context.Context, gid uuid.UUID, query string) ([]TagFacet, error)

	// FieldFacets returns every custom field present on a group's entities
	// mapped to its value distribution (value -> entity count). It is the
	// discovery call: which fields can be filtered on and what values each has.
	FieldFacets(ctx context.Context, gid uuid.UUID) (map[string][]FieldFacet, error)

	// SearchFieldValues returns the distinct values of a single custom field
	// within a group ranked by entity count, optionally narrowed to those whose
	// value matches query (a case-insensitive substring).
	SearchFieldValues(ctx context.Context, gid uuid.UUID, field, query string) ([]FieldFacet, error)
}

// Both engines provide the faceting capability.
var (
	_ Faceter = (*DatabaseEngine)(nil)
	_ Faceter = (*MeilisearchEngine)(nil)
)

// NewEngine constructs the search engine selected by cfg.Driver. An empty
// driver selects the database engine. The event bus may be nil, in which case
// external engines fall back to startup-only index builds.
func NewEngine(cfg config.SearchConf, db *ent.Client, bus *eventbus.EventBus) (Engine, error) {
	switch strings.ToLower(strings.TrimSpace(cfg.Driver)) {
	case "", DriverDatabase:
		return NewDatabaseEngine(db), nil
	case DriverMeilisearch:
		return NewMeilisearchEngine(cfg.Meilisearch, db, bus)
	default:
		return nil, fmt.Errorf("unsupported search driver: %q (supported: %s, %s)", cfg.Driver, DriverDatabase, DriverMeilisearch)
	}
}
