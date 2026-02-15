package repo

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sysadminsmedia/homebox/backend/internal/core/services/reporting/eventbus"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/group"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/predicate"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/tag"
)

type TagRepository struct {
	db  *ent.Client
	bus *eventbus.EventBus
}

type (
	TagCreate struct {
		Name        string    `json:"name"        validate:"required,min=1,max=255"`
		ParentID    uuid.UUID `json:"parentId"    extensions:"x-nullable"`
		Description string    `json:"description" validate:"max=1000"`
		Color       string    `json:"color"`
		Icon        string    `json:"icon"        validate:"max=255"`
	}

	TagUpdate struct {
		ID          uuid.UUID `json:"id"`
		ParentID    uuid.UUID `json:"parentId"    extensions:"x-nullable"`
		Name        string    `json:"name"        validate:"required,min=1,max=255"`
		Description string    `json:"description" validate:"max=1000"`
		Color       string    `json:"color"`
		Icon        string    `json:"icon"        validate:"max=255"`
	}

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

	TagOut struct {
		TagSummary
		Parent   *TagSummary  `json:"parent,omitempty" extensions:"x-nullable"`
		Children []TagSummary `json:"children"`
	}
)

func mapTagSummary(tag *ent.Tag) TagSummary {
	var parentID uuid.UUID
	if tag.Edges.Parent != nil {
		parentID = tag.Edges.Parent.ID
	}

	return TagSummary{
		ID:          tag.ID,
		ParentID:    parentID,
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
	var parent *TagSummary
	if tag.Edges.Parent != nil {
		p := mapTagSummary(tag.Edges.Parent)
		parent = &p
	}

	children := make([]TagSummary, 0)
	if tag.Edges.Children != nil {
		for _, c := range tag.Edges.Children {
			summary := mapTagSummary(c)
			summary.ParentID = tag.ID
			children = append(children, summary)
		}
	}

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

func (r *TagRepository) GetOne(ctx context.Context, gid uuid.UUID, id uuid.UUID) (TagOut, error) {
	return r.getOne(ctx, tag.ID(id), tag.HasGroupWith(group.ID(gid)))
}

func (r *TagRepository) GetOneByGroup(ctx context.Context, gid, id uuid.UUID) (TagOut, error) {
	return r.getOne(ctx, tag.ID(id), tag.HasGroupWith(group.ID(gid)))
}

func (r *TagRepository) GetAll(ctx context.Context, groupID uuid.UUID) ([]TagSummary, error) {
	return mapTagsOut(r.db.Tag.Query().
		Where(tag.HasGroupWith(group.ID(groupID))).
		Order(ent.Asc(tag.FieldName)).
		WithGroup().
		WithParent().
		All(ctx),
	)
}

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

// checkCycle checks if targetID is an ancestor of startID.
// This is used when moving startID to be a child of targetID.
// If targetID is already a child/descendant of startID, then making startID a child of targetID creates a cycle.
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

func (r *TagRepository) Create(ctx context.Context, groupID uuid.UUID, data TagCreate) (TagOut, error) {
	if data.ParentID != uuid.Nil {
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

func (r *TagRepository) update(ctx context.Context, data TagUpdate, where ...predicate.Tag) (int, error) {
	if len(where) == 0 {
		panic("empty where not supported empty")
	}

	if data.ParentID != uuid.Nil {
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

func (r *TagRepository) UpdateByGroup(ctx context.Context, gid uuid.UUID, data TagUpdate) (TagOut, error) {
	affected, err := r.update(ctx, data, tag.ID(data.ID), tag.HasGroupWith(group.ID(gid)))
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
