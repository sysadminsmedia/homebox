package repo

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sysadminsmedia/homebox/backend/internal/core/services/reporting/eventbus"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/group"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/label"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/predicate"
)

type LabelRepository struct {
	db  *ent.Client
	bus *eventbus.EventBus
}

type (
	LabelCreate struct {
		Name        string    `json:"name"        validate:"required,min=1,max=255"`
		Description string    `json:"description" validate:"max=1000"`
		Color       string    `json:"color"`
		ParentID    uuid.UUID `json:"parentId"    extensions:"x-nullable"`
	}

	LabelUpdate struct {
		ID          uuid.UUID `json:"id"`
		Name        string    `json:"name"        validate:"required,min=1,max=255"`
		Description string    `json:"description" validate:"max=1000"`
		Color       string    `json:"color"`
		ParentID    uuid.UUID `json:"parentId"    extensions:"x-nullable"`
	}

	LabelSummary struct {
		ID          uuid.UUID `json:"id"`
		Name        string    `json:"name"`
		Description string    `json:"description"`
		Color       string    `json:"color"`
		CreatedAt   time.Time `json:"createdAt"`
		UpdatedAt   time.Time `json:"updatedAt"`
	}

	LabelOut struct {
		Parent *LabelSummary `json:"parent,omitempty"`
		LabelSummary
		Children []LabelSummary `json:"children"`
	}
)

func mapLabelSummary(label *ent.Label) LabelSummary {
	return LabelSummary{
		ID:          label.ID,
		Name:        label.Name,
		Description: label.Description,
		Color:       label.Color,
		CreatedAt:   label.CreatedAt,
		UpdatedAt:   label.UpdatedAt,
	}
}

var (
	mapLabelOutErr = mapTErrFunc(mapLabelOut)
	mapLabelsOut   = mapTEachErrFunc(mapLabelSummary)
)

func mapLabelOut(label *ent.Label) LabelOut {
	var parent *LabelSummary
	if label.Edges.Parent != nil {
		p := mapLabelSummary(label.Edges.Parent)
		parent = &p
	}

	children := make([]LabelSummary, 0, len(label.Edges.Children))
	for _, c := range label.Edges.Children {
		children = append(children, mapLabelSummary(c))
	}

	return LabelOut{
		Parent:   parent,
		Children: children,
		LabelSummary: mapLabelSummary(label),
	}
}

func (r *LabelRepository) publishMutationEvent(gid uuid.UUID) {
	if r.bus != nil {
		r.bus.Publish(eventbus.EventLabelMutation, eventbus.GroupMutationEvent{GID: gid})
	}
}

const maxLabelDepth = 20

// validateParentNotCircular checks if setting parentID as parent of labelID would create a circular reference
func (r *LabelRepository) validateParentNotCircular(ctx context.Context, labelID, parentID uuid.UUID) error {
	// Check if label is being set as its own parent
	if labelID == parentID {
		return fmt.Errorf("label cannot be its own parent")
	}

	// Check if parent exists and get its parent chain
	depth := 0
	currentID := parentID

	for currentID != uuid.Nil && depth < maxLabelDepth {
		l, err := r.db.Label.Query().
			Where(label.ID(currentID)).
			WithParent().
			Only(ctx)
		if err != nil {
			if ent.IsNotFound(err) {
				return fmt.Errorf("parent label not found")
			}
			return err
		}

		// Check if we've reached the label we're trying to set a parent for (circular reference)
		if l.ID == labelID {
			return fmt.Errorf("circular reference detected in label hierarchy")
		}

		depth++
		if depth >= maxLabelDepth {
			return fmt.Errorf("label hierarchy exceeds maximum depth of 20")
		}

		if l.Edges.Parent != nil {
			currentID = l.Edges.Parent.ID
		} else {
			currentID = uuid.Nil
		}
	}

	return nil
}

func (r *LabelRepository) getOne(ctx context.Context, where ...predicate.Label) (LabelOut, error) {
	return mapLabelOutErr(r.db.Label.Query().
		Where(where...).
		WithGroup().
		WithParent().
		WithChildren(func(lq *ent.LabelQuery) {
			lq.Order(label.ByName())
		}).
		Only(ctx),
	)
}

func (r *LabelRepository) GetOne(ctx context.Context, id uuid.UUID) (LabelOut, error) {
	return r.getOne(ctx, label.ID(id))
}

func (r *LabelRepository) GetOneByGroup(ctx context.Context, gid, ld uuid.UUID) (LabelOut, error) {
	return r.getOne(ctx, label.ID(ld), label.HasGroupWith(group.ID(gid)))
}

func (r *LabelRepository) GetAll(ctx context.Context, groupID uuid.UUID) ([]LabelSummary, error) {
	return mapLabelsOut(r.db.Label.Query().
		Where(label.HasGroupWith(group.ID(groupID))).
		Order(ent.Asc(label.FieldName)).
		WithGroup().
		All(ctx),
	)
}

func (r *LabelRepository) Create(ctx context.Context, groupID uuid.UUID, data LabelCreate) (LabelOut, error) {
	// Validate parent if provided
	if data.ParentID != uuid.Nil {
		// Check that parent exists and belongs to same group
		parent, err := r.db.Label.Query().
			Where(label.ID(data.ParentID), label.HasGroupWith(group.ID(groupID))).
			Only(ctx)
		if err != nil {
			if ent.IsNotFound(err) {
				return LabelOut{}, fmt.Errorf("parent label not found or does not belong to the same group")
			}
			return LabelOut{}, err
		}

		// Validate parent hierarchy depth
		depth := 0
		currentID := parent.ID
		for currentID != uuid.Nil && depth < maxLabelDepth {
			l, err := r.db.Label.Query().
				Where(label.ID(currentID)).
				WithParent().
				Only(ctx)
			if err != nil {
				return LabelOut{}, err
			}

			depth++
			if depth >= maxLabelDepth {
				return LabelOut{}, fmt.Errorf("label hierarchy exceeds maximum depth of 20")
			}

			if l.Edges.Parent != nil {
				currentID = l.Edges.Parent.ID
			} else {
				currentID = uuid.Nil
			}
		}
	}

	q := r.db.Label.Create().
		SetName(data.Name).
		SetDescription(data.Description).
		SetColor(data.Color).
		SetGroupID(groupID)

	if data.ParentID != uuid.Nil {
		q.SetParentID(data.ParentID)
	}

	label, err := q.Save(ctx)
	if err != nil {
		return LabelOut{}, err
	}

	label.Edges.Group = &ent.Group{ID: groupID} // bootstrap group ID
	r.publishMutationEvent(groupID)
	return mapLabelOut(label), nil
}

func (r *LabelRepository) update(ctx context.Context, data LabelUpdate, where ...predicate.Label) (int, error) {
	if len(where) == 0 {
		panic("empty where not supported empty")
	}

	q := r.db.Label.Update().
		Where(where...).
		SetName(data.Name).
		SetDescription(data.Description).
		SetColor(data.Color)

	if data.ParentID != uuid.Nil {
		q.SetParentID(data.ParentID)
	} else {
		q.ClearParent()
	}

	return q.Save(ctx)
}

func (r *LabelRepository) UpdateByGroup(ctx context.Context, gid uuid.UUID, data LabelUpdate) (LabelOut, error) {
	// Validate parent if provided
	if data.ParentID != uuid.Nil {
		// Check that parent exists and belongs to same group
		parent, err := r.db.Label.Query().
			Where(label.ID(data.ParentID), label.HasGroupWith(group.ID(gid))).
			Only(ctx)
		if err != nil {
			if ent.IsNotFound(err) {
				return LabelOut{}, fmt.Errorf("parent label not found or does not belong to the same group")
			}
			return LabelOut{}, err
		}

		// Validate no circular reference
		if err := r.validateParentNotCircular(ctx, data.ID, parent.ID); err != nil {
			return LabelOut{}, err
		}
	}

	_, err := r.update(ctx, data, label.ID(data.ID), label.HasGroupWith(group.ID(gid)))
	if err != nil {
		return LabelOut{}, err
	}

	r.publishMutationEvent(gid)
	return r.GetOne(ctx, data.ID)
}

// delete removes the label from the database. This should only be used when
// the label's ownership is already confirmed/validated.
func (r *LabelRepository) delete(ctx context.Context, id uuid.UUID) error {
	return r.db.Label.DeleteOneID(id).Exec(ctx)
}

func (r *LabelRepository) DeleteByGroup(ctx context.Context, gid, id uuid.UUID) error {
	_, err := r.db.Label.Delete().
		Where(
			label.ID(id),
			label.HasGroupWith(group.ID(gid)),
		).Exec(ctx)
	if err != nil {
		return err
	}

	r.publishMutationEvent(gid)

	return nil
}
