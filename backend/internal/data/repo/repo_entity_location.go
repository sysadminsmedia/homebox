package repo

import (
	"context"

	"github.com/google/uuid"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/entity"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/entitytype"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/group"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/predicate"
	"github.com/sysadminsmedia/homebox/backend/internal/sys/validate"
)

// Two columns say where an entity lives. entity_children is containment; since
// 0.26 it was also the only source of location, which is why an item couldn't be
// a child of one item and stored somewhere else (#1688).
//
// entity_location_entities is a sparse override, only valid when the parent is a
// non-location entity. NULL means inherit from the parent chain, so an item in a
// location is still just parent=<location>, and the two never both claim one.

const locationField = "locationId"

// resolveLocationOverride normalizes a (parent, location) pair. Returns what to
// persist; a uuid.Nil override means clear it. subjectID is Nil on create.
//
//   - no location: parent as given, override cleared (the old behavior)
//   - location, no parent or same-location parent: location becomes the parent
//   - location + parent item: override stored, the #1688 case
//   - location + a different location as parent: rejected, since both claim to
//     say where this lives and picking one silently loses data
func resolveLocationOverride(
	ctx context.Context,
	c *ent.EntityClient,
	gid, subjectID, parentID, locationID uuid.UUID,
) (parent, override uuid.UUID, err error) {
	if locationID == uuid.Nil {
		return parentID, uuid.Nil, nil
	}

	if locationID == subjectID {
		return uuid.Nil, uuid.Nil, validate.NewFieldErrors(
			validate.NewFieldError(locationField, "an entity cannot be its own location"),
		)
	}

	// One query, so a cross-group UUID can't be probed by the difference between
	// "not a location" and "not found".
	isLocation, err := c.Query().
		Where(
			entity.ID(locationID),
			entity.HasGroupWith(group.ID(gid)),
			entity.HasEntityTypeWith(entitytype.IsLocation(true)),
		).
		Exist(ctx)
	if err != nil {
		return uuid.Nil, uuid.Nil, err
	}
	if !isLocation {
		return uuid.Nil, uuid.Nil, validate.NewFieldErrors(
			validate.NewFieldError(locationField, "location must reference an existing location in this group"),
		)
	}

	if parentID == uuid.Nil || parentID == locationID {
		return locationID, uuid.Nil, nil
	}

	parentIsLocation, err := c.Query().
		Where(
			entity.ID(parentID),
			entity.HasGroupWith(group.ID(gid)),
			entity.HasEntityTypeWith(entitytype.IsLocation(true)),
		).
		Exist(ctx)
	if err != nil {
		return uuid.Nil, uuid.Nil, err
	}
	if parentIsLocation {
		return uuid.Nil, uuid.Nil, validate.NewFieldErrors(
			validate.NewFieldError(locationField,
				"cannot set a location while the parent is a different location; move the entity by changing its parent instead"),
		)
	}

	return parentID, locationID, nil
}

// resolveEntityLocation returns e's override if set, else the nearest location
// ancestor. Needs the Location and Parent edges loaded.
func resolveEntityLocation(ctx context.Context, c *ent.EntityClient, e *ent.Entity) (*EntitySummary, error) {
	if e.Edges.Location != nil {
		s := mapEntitySummary(e.Edges.Location)
		return &s, nil
	}
	return nearestLocationAncestor(ctx, c, e.Edges.Parent)
}

// effectiveLocationIn is the query-side twin of resolveEntityLocation:
// entities parented directly to one of ids, plus entities elsewhere in the
// tree that name it explicitly.
func effectiveLocationIn(ids []uuid.UUID) predicate.Entity {
	return entity.Or(
		entity.And(
			entity.HasParentWith(entity.IDIn(ids...)),
			entity.Not(entity.HasLocation()),
		),
		entity.HasLocationWith(entity.IDIn(ids...)),
	)
}

// clearChildLocationOverrides drops direct children's overrides so they go back
// to inheriting. Grandchildren are left alone — their override is their own
// parent's toggle to manage.
func clearChildLocationOverrides(ctx context.Context, c *ent.EntityClient, gid, parentID uuid.UUID) (int, error) {
	return c.Update().
		Where(
			entity.HasParentWith(entity.ID(parentID)),
			entity.HasGroupWith(group.ID(gid)),
			entity.HasLocation(),
		).
		ClearLocation().
		Save(ctx)
}

// locationOrParentColumn is the effective-location expression for the raw SQL
// queries, where ent's predicate builder isn't available.
func locationOrParentColumn(table string) string {
	return "COALESCE(" + table + ".entity_location_entities, " + table + ".entity_children)"
}
