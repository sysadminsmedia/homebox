package repo

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/entity"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/group"
)

// LocationItemsResult includes exact server-side pagination totals and the
// number of items physically assigned directly to the selected location.
type LocationItemsResult struct {
	PaginationResult[EntitySummary]
	DirectTotal int `json:"directTotal"`
}

// ErrLocationRootNotLocation identifies an inventory item used in place of a location root.
var ErrLocationRootNotLocation = errors.New("select a location, not an inventory item")

// ErrInvalidLocationQuery identifies invalid location depth or pagination values.
var ErrInvalidLocationQuery = errors.New("invalid location depth or pagination parameters")

// MaxLocationDepth is the maximum descendant depth used for the "all" option.
const MaxLocationDepth = maxAncestorDepth

type locationItemRow struct {
	id   uuid.UUID
	path string
}

// A location has one physical containment chain. Explicit location overrides
// take precedence over an item's ordinary parent; the same COALESCE rule is
// used by the existing Homebox location tree. The CTE counts only location
// nodes as depth, so grouping items inside other items doesn't consume an
// additional location level. The visited UUID chain prevents cycles from
// reintroducing the root or duplicates while preserving each item's path.
// The two indexed parent columns support fast, override-aware traversal.
const locationDescendantCTE = `
WITH RECURSIVE location_scope(id, location_depth, path, parent_path, visited) AS (
    SELECT root.id, 0, CAST(root.name AS TEXT), CAST('' AS TEXT),
           '|' || CAST(root.id AS TEXT) || '|'
    FROM entities root
    JOIN entity_types rt ON rt.id = root.entity_type_entities
    WHERE root.id = $1 AND root.group_entities = $2 AND rt.is_location = true

    UNION ALL

    SELECT child.id,
           node.location_depth + CASE WHEN ct.is_location = true THEN 1 ELSE 0 END,
           node.path || ' / ' || child.name,
           node.path,
           node.visited || CAST(child.id AS TEXT) || '|'
    FROM entities child
    JOIN location_scope node
       ON (child.entity_location_entities = node.id
           OR (child.entity_location_entities IS NULL AND child.entity_children = node.id))
    LEFT JOIN entity_types ct ON ct.id = child.entity_type_entities
    WHERE child.group_entities = $2
      AND node.visited NOT LIKE '%|' || CAST(child.id AS TEXT) || '|%'
      AND node.location_depth + CASE WHEN ct.is_location = true THEN 1 ELSE 0 END <= $3
)
`

const locationDescendantWhere = `
 FROM location_scope node
 JOIN entities e ON e.id = node.id AND e.group_entities = $2
 LEFT JOIN entity_types et ON et.id = e.entity_type_entities
 WHERE node.id <> $1
   AND COALESCE(et.is_location, false) = false
   AND e.archived = false
`

// QueryLocationDescendants selects at the database level, not by loading the
// whole inventory into the browser. The recursive scope never crosses group
// boundaries; pagination occurs before hydrating the page of item cards.
// depth is 1, 2, 3, or MaxLocationDepth (for All). The visited UUID
// chain prevents cycles without limiting non-location containment depth.
func (r *EntityRepository) QueryLocationDescendants(
	ctx context.Context, gid, locationID uuid.UUID, depth, page, pageSize int,
) (LocationItemsResult, error) {
	if depth < 1 || depth > maxAncestorDepth || page < 1 || pageSize < 1 || pageSize > 96 {
		return LocationItemsResult{}, ErrInvalidLocationQuery
	}
	// The API accepts integer pages; reject an offset that would overflow on
	// either 32- or 64-bit servers rather than returning a wrapped page.
	if page-1 > int(^uint(0)>>1)/pageSize {
		return LocationItemsResult{}, fmt.Errorf("%w: location page offset is too large", ErrInvalidLocationQuery)
	}
	// Only load the root and its type. GetOneByGroup eagerly hydrates children,
	// attachments and other unrelated data, defeating bounded pagination.
	root, err := r.db.Entity.Query().
		Where(entity.ID(locationID), entity.HasGroupWith(group.ID(gid))).
		WithEntityType().
		Only(ctx)
	if err != nil {
		return LocationItemsResult{}, err
	}
	if root.Edges.EntityType == nil || !root.Edges.EntityType.IsLocation {
		return LocationItemsResult{}, ErrLocationRootNotLocation
	}

	result := LocationItemsResult{
		PaginationResult: PaginationResult[EntitySummary]{
			Page: page, PageSize: pageSize, Items: []EntitySummary{},
		},
	}

	// Count without applying pagination, including the count of items whose
	// effective parent is this location rather than a nested container.
	countQuery := locationDescendantCTE + `
SELECT COUNT(*),
       COALESCE(SUM(CASE WHEN COALESCE(e.entity_location_entities, e.entity_children) = $1
                         THEN 1 ELSE 0 END), 0)
` + locationDescendantWhere
	err = r.db.Sql().QueryRowContext(ctx, countQuery, locationID, gid, depth).
		Scan(&result.Total, &result.DirectTotal)
	if err != nil {
		return LocationItemsResult{}, err
	}
	if result.Total == 0 {
		return result, nil
	}

	// Stable name + UUID order ensures pagination cannot drop/duplicate rows.
	// parent_path describes the actual storage path and excludes the item name.
	pageQuery := locationDescendantCTE + `
SELECT node.id, node.parent_path
` + locationDescendantWhere + `
 ORDER BY LOWER(e.name), e.id LIMIT $4 OFFSET $5
`
	rows, err := r.db.Sql().QueryContext(ctx, pageQuery, locationID, gid, depth, pageSize, (page-1)*pageSize)
	if err != nil {
		return LocationItemsResult{}, err
	}
	entries := make([]locationItemRow, 0, pageSize)
	for rows.Next() {
		var entry locationItemRow
		if scanErr := rows.Scan(&entry.id, &entry.path); scanErr != nil {
			_ = rows.Close()
			return LocationItemsResult{}, scanErr
		}
		entries = append(entries, entry)
	}
	if scanErr := rows.Err(); scanErr != nil {
		_ = rows.Close()
		return LocationItemsResult{}, scanErr
	}
	if closeErr := rows.Close(); closeErr != nil {
		return LocationItemsResult{}, closeErr
	}
	if len(entries) == 0 {
		return result, nil
	}

	ids := make([]uuid.UUID, 0, len(entries))
	for _, entry := range entries {
		ids = append(ids, entry.id)
	}
	// Reuse the regular item projection (attachments, tags and parent), without
	// repeating costly per-item queries. The IDs are bounded by pageSize<=96.
	data, err := r.QueryByGroup(ctx, gid, EntityQuery{
		IDs: ids, Page: -1, PageSize: -1,
	})
	if err != nil {
		return LocationItemsResult{}, err
	}

	itemsByID := make(map[uuid.UUID]EntitySummary, len(data.Items))
	for _, item := range data.Items {
		itemsByID[item.ID] = item
	}
	for _, entry := range entries {
		if item, ok := itemsByID[entry.id]; ok {
			item.LocationPath = strings.TrimSpace(entry.path)
			result.Items = append(result.Items, item)
		}
	}
	return result, nil
}
