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
