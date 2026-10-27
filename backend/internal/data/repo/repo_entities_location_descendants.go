package repo

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

// LocationItemsResult includes exact server-side pagination totals and the
// number of items physically assigned directly to the selected location.
type LocationItemsResult struct {
	PaginationResult[EntitySummary]
	DirectTotal int `json:"directTotal"`
}

type locationItemRow struct {
	id   uuid.UUID
	path string
}

// A location has one physical containment chain. Explicit location overrides
// take precedence over an item's ordinary parent; the same COALESCE rule is
// used by the existing Homebox location tree. The CTE counts only location
// nodes as depth, so grouping items inside other items doesn't consume an
// additional location level.
const locationDescendantCTE = `
WITH RECURSIVE location_scope(id, location_depth, hops, path, parent_path) AS (
    SELECT root.id, 0, 0, root.name, ''
    FROM entities root
    JOIN entity_types rt ON rt.id = root.entity_type_entities
    WHERE root.id = $1 AND root.group_entities = $2 AND rt.is_location = true

    UNION ALL

    SELECT child.id,
           node.location_depth + CASE WHEN ct.is_location = true THEN 1 ELSE 0 END,
           node.hops + 1,
           node.path || ' / ' || child.name,
           node.path
    FROM entities child
    JOIN location_scope node
       ON COALESCE(child.entity_location_entities, child.entity_children) = node.id
    LEFT JOIN entity_types ct ON ct.id = child.entity_type_entities
    WHERE child.group_entities = $2
      AND node.hops < 64
      AND node.location_depth + CASE WHEN ct.is_location = true THEN 1 ELSE 0 END <= $3
)
`

const locationDescendantWhere = `
 FROM location_scope node
 JOIN entities e ON e.id = node.id AND e.group_entities = $2
 LEFT JOIN entity_types et ON et.id = e.entity_type_entities
 WHERE node.hops > 0
   AND COALESCE(et.is_location, false) = false
   AND e.archived = false
`

// QueryLocationDescendants selects at the database level, not by loading the
// whole inventory into the browser. The recursive scope never crosses group
// boundaries; pagination occurs before hydrating the page of item cards.
// depth is 1, 2, 3, or 64 (for All). A hop guard prevents corrupt cycles.
func (r *EntityRepository) QueryLocationDescendants(
	ctx context.Context, gid, locationID uuid.UUID, depth, page, pageSize int,
) (LocationItemsResult, error) {
	if depth < 1 || depth > maxAncestorDepth || page < 1 || pageSize < 1 || pageSize > 96 {
		return LocationItemsResult{}, fmt.Errorf("invalid location depth or pagination parameters")
	}
	// The API accepts integer pages; reject an offset that would overflow on
	// either 32- or 64-bit servers rather than returning a wrapped page.
	if page-1 > int(^uint(0)>>1)/pageSize {
		return LocationItemsResult{}, fmt.Errorf("location page offset is too large")
	}
	root, err := r.GetOneByGroup(ctx, gid, locationID)
	if err != nil {
		return LocationItemsResult{}, err
	}
	if root.EntityType == nil || !root.EntityType.IsLocation {
		return LocationItemsResult{}, fmt.Errorf("select a location, not an inventory item")
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
