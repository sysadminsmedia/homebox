package repo

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/samber/lo"
	"github.com/sysadminsmedia/homebox/backend/internal/core/services/reporting/eventbus"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/group"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/predicate"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/tag"
)

// TagRepository provides data access operations for tag entities.
// It supports hierarchical tag structures with parent-child relationships
// and enforces multi-tenant isolation through group membership.
type TagRepository struct {
	db  *ent.Client
	bus *eventbus.EventBus
}

type (
	// TagCreate represents the input data required to create a new tag.
	// Tags can optionally have a parent to form hierarchical structures
	// with a maximum depth of 5 levels.
	TagCreate struct {
		Name        string    `json:"name"        validate:"required,min=1,max=255"`
		ParentID    uuid.UUID `json:"parentId"    extensions:"x-nullable"`
		Description string    `json:"description" validate:"max=1000"`
		Color       string    `json:"color"`
		Icon        string    `json:"icon"        validate:"max=255"`
	}

	// TagUpdate represents the input data for updating an existing tag.
	// All fields can be modified, including moving the tag to a different
	// parent (while maintaining hierarchy constraints and preventing cycles).
	TagUpdate struct {
		ID          uuid.UUID `json:"id"`
		ParentID    uuid.UUID `json:"parentId"    extensions:"x-nullable"`
		Name        string    `json:"name"        validate:"required,min=1,max=255"`
		Description string    `json:"description" validate:"max=1000"`
		Color       string    `json:"color"`
		Icon        string    `json:"icon"        validate:"max=255"`
	}

	// TagSummary provides a lightweight representation of a tag without
	// its relationships. Used in lists and as nested references.
	TagSummary struct {
		ID          uuid.UUID `json:"id"`
		ParentID    uuid.UUID `json:"parentId"    extensions:"x-nullable"`
		Name        string    `json:"name"`
		Description string    `json:"description"`
		Color       string    `json:"color"`
		Icon        string    `json:"icon"`
		CreatedAt   time.Time `json:"createdAt"`
		UpdatedAt   time.Time `json:"updatedAt"`
	}

	// TagOut represents a complete tag with its parent and children relationships.
	// The Parent field is nil for root-level tags. Children is always initialized
	// (empty slice if the tag has no children).
	TagOut struct {
		TagSummary
		Parent   *TagSummary  `json:"parent,omitempty" extensions:"x-nullable"`
		Children []TagSummary `json:"children"`
	}
)

func mapTagSummary(tag *ent.Tag) TagSummary {
	return TagSummary{
		ID:          tag.ID,
		ParentID:    lo.Ternary(tag.Edges.Parent != nil, tag.Edges.Parent.ID, uuid.Nil),
		Name:        tag.Name,
		Description: tag.Description,
		Color:       tag.Color,
		Icon:        tag.Icon,
		CreatedAt:   tag.CreatedAt,
		UpdatedAt:   tag.UpdatedAt,
	}
}

var (
	mapTagOutErr = mapTErrFunc(mapTagOut)
	mapTagsOut   = mapTEachErrFunc(mapTagSummary)
)

func mapTagOut(tag *ent.Tag) TagOut {
	parent := lo.TernaryF(
		tag.Edges.Parent != nil,
		func() *TagSummary {
			p := mapTagSummary(tag.Edges.Parent)
			return &p
		},
		func() *TagSummary { return nil },
	)

	children := lo.Map(tag.Edges.Children, func(c *ent.Tag, _ int) TagSummary {
		summary := mapTagSummary(c)
		summary.ParentID = tag.ID
		return summary
	})

	return TagOut{
		TagSummary: mapTagSummary(tag),
		Parent:     parent,
		Children:   children,
	}
}

func (r *TagRepository) publishMutationEvent(gid uuid.UUID) {
	if r.bus != nil {
		r.bus.Publish(eventbus.EventTagMutation, eventbus.GroupMutationEvent{GID: gid})
	}
}

func (r *TagRepository) getOne(ctx context.Context, where ...predicate.Tag) (TagOut, error) {
	return mapTagOutErr(r.db.Tag.Query().
		Where(where...).
		WithGroup().
		WithParent().
		WithChildren().
		Only(ctx),
	)
}

// GetOne retrieves a single tag by ID, ensuring it belongs to the specified group.
// Returns the tag with its parent and children relationships fully populated.
// Returns an error if the tag doesn't exist or doesn't belong to the group.
func (r *TagRepository) GetOne(ctx context.Context, gid uuid.UUID, id uuid.UUID) (TagOut, error) {
	return r.getOne(ctx, tag.ID(id), tag.HasGroupWith(group.ID(gid)))
}

// GetOneByGroup retrieves a single tag by ID with group validation.
// This is an alias for GetOne, maintained for API consistency with other repositories.
func (r *TagRepository) GetOneByGroup(ctx context.Context, gid, id uuid.UUID) (TagOut, error) {
	return r.getOne(ctx, tag.ID(id), tag.HasGroupWith(group.ID(gid)))
}

// GetAll retrieves all tags belonging to the specified group, ordered by name.
// Tags are returned as summaries (without parent/children relationships loaded).
// Parent edges are loaded to populate ParentID fields in the summaries.
func (r *TagRepository) GetAll(ctx context.Context, groupID uuid.UUID) ([]TagSummary, error) {
	return mapTagsOut(r.db.Tag.Query().
		Where(tag.HasGroupWith(group.ID(groupID))).
		Order(ent.Asc(tag.FieldName)).
		WithGroup().
		WithParent().
		All(ctx),
	)
}

// getSubtreeDepth calculates the maximum depth of the subtree rooted at the given tag ID.
// Uses a recursive CTE to traverse the entire subtree and find the deepest level.
// Returns 1 for a tag with no children, and increases by 1 for each level.
func (r *TagRepository) getSubtreeDepth(ctx context.Context, id uuid.UUID) (int, error) {
	query := `
		WITH RECURSIVE tag_tree(id, depth) AS (
			SELECT id, 1 as depth
			FROM tags
			WHERE id = $1
			UNION ALL
			SELECT t.id, tt.depth + 1
			FROM tags t
			JOIN tag_tree tt ON t.tag_children = tt.id
		)
		SELECT MAX(depth) FROM tag_tree;
	`
	// Since we want the depth of the subtree *relative to the root*,
	// the query above calculates depth from root (1) downwards.
	// The MAX(depth) will be the height of the tree rooted at 'id'.

	rows, err := r.db.Sql().QueryContext(ctx, query, id)
	if err != nil {
		return 0, err
	}
	defer func() { _ = rows.Close() }()

	if rows.Next() {
		var maxDepth int
		if err := rows.Scan(&maxDepth); err != nil {
			return 0, err
		}
		return maxDepth, nil
	}
	return 0, nil
}

// checkDepth calculates how many levels deep the given parent tag is from the root.
// Uses a recursive CTE to traverse up the tree from the parent to the root.
// Returns 0 for root-level tags (parentID is uuid.Nil).
// Returns the number of levels from root to the given parent tag.
func (r *TagRepository) checkDepth(ctx context.Context, parentID uuid.UUID) (int, error) {
	if parentID == uuid.Nil {
		return 0, nil
	}

	query := `
		WITH RECURSIVE tag_parents(id, parent_id, depth) AS (
			SELECT id, tag_children, 1
			FROM tags
			WHERE id = $1
			UNION ALL
			SELECT t.id, t.tag_children, tp.depth + 1
			FROM tags t
			JOIN tag_parents tp ON t.id = tp.parent_id
		)
		SELECT MAX(depth) FROM tag_parents;
	`
	rows, err := r.db.Sql().QueryContext(ctx, query, parentID)
	if err != nil {
		return 0, err
	}
	defer func() { _ = rows.Close() }()

	if rows.Next() {
		var depth int
		if err := rows.Scan(&depth); err != nil {
			return 0, err
		}
		return depth, nil
	}

	return 0, nil
}

// checkCycle checks if setting movingID's parent to proposedParentID would create a cycle.
// Returns true if proposedParentID is a descendant of movingID (or if they are the same tag).
// Uses a recursive CTE to traverse all descendants of movingID.
// This prevents circular parent-child relationships in the tag hierarchy.
func (r *TagRepository) checkCycle(ctx context.Context, movingID, proposedParentID uuid.UUID) (bool, error) {
	if movingID == proposedParentID {
		return true, nil
	}

	query := `
		WITH RECURSIVE ancestors(id, parent_id) AS (
			SELECT id, tag_children
			FROM tags
			WHERE id = $1
			UNION ALL
			SELECT t.id, t.tag_children
			FROM tags t
			JOIN ancestors a ON t.id = a.parent_id
		)
		SELECT 1 FROM ancestors WHERE id = $2 LIMIT 1;
	`

	rows, err := r.db.Sql().QueryContext(ctx, query, proposedParentID, movingID)
	if err != nil {
		return false, err
	}
	defer func() { _ = rows.Close() }()

	if rows.Next() {
		return true, nil
	}
	return false, nil
}

// Create creates a new tag in the specified group.
// If ParentID is provided, validates that:
//   - The parent tag exists and belongs to the same group
//   - Adding this tag would not exceed the maximum depth of 5 levels
//
// Returns the created tag with all relationships fully populated.
// Publishes a tag mutation event on successful creation.
func (r *TagRepository) Create(ctx context.Context, groupID uuid.UUID, data TagCreate) (TagOut, error) {
	if data.ParentID != uuid.Nil {
		// Verify parent tag belongs to the same group
		parentTag, err := r.db.Tag.Query().
			Where(tag.ID(data.ParentID)).
			WithGroup().
			Only(ctx)
		if err != nil {
			if ent.IsNotFound(err) {
				return TagOut{}, fmt.Errorf("parent tag not found or does not belong to this group")
			}
			return TagOut{}, err
		}
		if parentTag.Edges.Group == nil || parentTag.Edges.Group.ID != groupID {
			return TagOut{}, fmt.Errorf("parent tag not found or does not belong to this group")
		}

		parentDepth, err := r.checkDepth(ctx, data.ParentID)
		if err != nil {
			return TagOut{}, err
		}
		// New item has depth 1.
		if parentDepth+1 > 5 {
			return TagOut{}, fmt.Errorf("max depth of 5 exceeded")
		}
	}

	q := r.db.Tag.Create().
		SetName(data.Name).
		SetDescription(data.Description).
		SetColor(data.Color).
		SetIcon(data.Icon).
		SetGroupID(groupID)

	if data.ParentID != uuid.Nil {
		q.SetParentID(data.ParentID)
	}

	createdTag, err := q.Save(ctx)
	if err != nil {
		return TagOut{}, err
	}

	// Re-fetch the tag to get fully populated edges (Parent, Children)
	freshTag, err := r.getOne(ctx, tag.ID(createdTag.ID), tag.HasGroupWith(group.ID(groupID)))
	if err != nil {
		return TagOut{}, err
	}

	r.publishMutationEvent(groupID)
	return freshTag, nil
}

func (r *TagRepository) update(ctx context.Context, groupID uuid.UUID, data TagUpdate, where ...predicate.Tag) (int, error) {
	if len(where) == 0 {
		panic("empty where not supported empty")
	}

	if data.ParentID != uuid.Nil {
		// Verify parent tag belongs to the same group
		parentTag, err := r.db.Tag.Query().
			Where(tag.ID(data.ParentID)).
			WithGroup().
			Only(ctx)
		if err != nil {
			if ent.IsNotFound(err) {
				return 0, fmt.Errorf("parent tag not found or does not belong to this group")
			}
			return 0, err
		}
		if parentTag.Edges.Group == nil || parentTag.Edges.Group.ID != groupID {
			return 0, fmt.Errorf("parent tag not found or does not belong to this group")
		}

		// 1. Check Cycle using CTE
		isCycle, err := r.checkCycle(ctx, data.ID, data.ParentID)
		if err != nil {
			return 0, err
		}
		if isCycle {
			return 0, fmt.Errorf("cycle detected")
		}

		// 2. Check Depth using CTEs
		// Depth of the new parent (how far down from root is the parent?)
		parentDepth, err := r.checkDepth(ctx, data.ParentID)
		if err != nil {
			return 0, err
		}

		// Depth of the subtree moving (how tall is the tree we are moving?)
		mySubtreeDepth, err := r.getSubtreeDepth(ctx, data.ID)
		if err != nil {
			return 0, err
		}

		if parentDepth+mySubtreeDepth > 5 {
			return 0, fmt.Errorf("max depth of 5 exceeded")
		}
	}

	q := r.db.Tag.Update().
		Where(where...).
		SetName(data.Name).
		SetDescription(data.Description).
		SetColor(data.Color).
		SetIcon(data.Icon)

	if data.ParentID != uuid.Nil {
		q.SetParentID(data.ParentID)
	} else {
		q.ClearParent()
	}

	return q.Save(ctx)
}

// UpdateByGroup updates an existing tag within the specified group.
// Validates that the tag exists and belongs to the group before updating.
// If ParentID is changed, additionally validates:
//   - The new parent tag exists and belongs to the same group
//   - The change would not create a cycle in the hierarchy
//   - The resulting tree would not exceed the maximum depth of 5 levels
//
// Returns an error if the tag is not found, belongs to a different group,
// or if the update would violate hierarchy constraints.
// Publishes a tag mutation event on successful update.
func (r *TagRepository) UpdateByGroup(ctx context.Context, gid uuid.UUID, data TagUpdate) (TagOut, error) {
	affected, err := r.update(ctx, gid, data, tag.ID(data.ID), tag.HasGroupWith(group.ID(gid)))
	if err != nil {
		return TagOut{}, err
	}

	if affected == 0 {
		return TagOut{}, fmt.Errorf("tag not found or does not belong to group")
	}

	r.publishMutationEvent(gid)
	return r.GetOne(ctx, gid, data.ID)
}

// delete removes the tag from the database. This should only be used when
// the tag's ownership is already confirmed/validated.
func (r *TagRepository) delete(ctx context.Context, id uuid.UUID) error {
	return r.db.Tag.DeleteOneID(id).Exec(ctx)
}

// DeleteByGroup deletes a tag from the specified group.
// Only deletes the tag if it exists and belongs to the group.
// Note: Child tags are not automatically deleted - they become root-level tags
// if their parent is deleted (depending on database cascade settings).
// Publishes a tag mutation event on successful deletion.
func (r *TagRepository) DeleteByGroup(ctx context.Context, gid, id uuid.UUID) error {
	_, err := r.db.Tag.Delete().
		Where(
			tag.ID(id),
			tag.HasGroupWith(group.ID(gid)),
		).Exec(ctx)
	if err != nil {
		return err
	}

	r.publishMutationEvent(gid)

	return nil
}
