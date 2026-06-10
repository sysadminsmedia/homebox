package repo

import (
	"context"

	"github.com/google/uuid"
	"github.com/sysadminsmedia/homebox/backend/internal/data/authz"
	"github.com/sysadminsmedia/homebox/backend/internal/data/authzrules"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/entity"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/entitytemplate"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/entitytype"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/group"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/predicate"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/tag"
)

// dedupNonNil returns the unique non-zero IDs in ids, preserving order. It is
// used by the assert* helpers so callers don't need to filter inputs before
// asking whether a set of UUIDs is reachable inside a tenant group.
func dedupNonNil(ids []uuid.UUID) []uuid.UUID {
	if len(ids) == 0 {
		return nil
	}
	seen := make(map[uuid.UUID]struct{}, len(ids))
	out := make([]uuid.UUID, 0, len(ids))
	for _, id := range ids {
		if id == uuid.Nil {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out
}

// assertEntityInGroup returns an ent.NotFoundError when id is set but does not
// resolve to an entity inside gid. A zero id is a no-op so callers can pass
// optional ParentID / DefaultLocationID values unconditionally. Returning the
// not-found error (rather than a distinct sentinel) preserves existing 404
// handling at the API edge and avoids leaking cross-tenant existence.
func assertEntityInGroup(ctx context.Context, c *ent.EntityClient, gid, id uuid.UUID) error {
	if id == uuid.Nil {
		return nil
	}
	exists, err := c.Query().
		Where(entity.ID(id), entity.HasGroupWith(group.ID(gid))).
		Exist(ctx)
	if err != nil {
		return err
	}
	if !exists {
		return &ent.NotFoundError{}
	}
	return nil
}

// assertEntityTypeInGroup is the EntityType analog of assertEntityInGroup.
func assertEntityTypeInGroup(ctx context.Context, c *ent.EntityTypeClient, gid, id uuid.UUID) error {
	if id == uuid.Nil {
		return nil
	}
	exists, err := c.Query().
		Where(entitytype.ID(id), entitytype.HasGroupWith(group.ID(gid))).
		Exist(ctx)
	if err != nil {
		return err
	}
	if !exists {
		return &ent.NotFoundError{}
	}
	return nil
}

// assertEntityTemplateInGroup is the EntityTemplate analog. id may be the zero
// UUID (no-op).
func assertEntityTemplateInGroup(ctx context.Context, c *ent.EntityTemplateClient, gid, id uuid.UUID) error {
	if id == uuid.Nil {
		return nil
	}
	exists, err := c.Query().
		Where(entitytemplate.ID(id), entitytemplate.HasGroupWith(group.ID(gid))).
		Exist(ctx)
	if err != nil {
		return err
	}
	if !exists {
		return &ent.NotFoundError{}
	}
	return nil
}

// assertTagsInGroup verifies that every supplied tag id belongs to gid.
// Duplicate and zero ids are tolerated. The count comparison after Where(IDIn)
// catches both "tag does not exist" and "tag exists in another group".
func assertTagsInGroup(ctx context.Context, c *ent.TagClient, gid uuid.UUID, ids []uuid.UUID) error {
	cleaned := dedupNonNil(ids)
	if len(cleaned) == 0 {
		return nil
	}
	count, err := c.Query().
		Where(tag.IDIn(cleaned...), tag.HasGroupWith(group.ID(gid))).
		Count(ctx)
	if err != nil {
		return err
	}
	if count != len(cleaned) {
		return &ent.NotFoundError{}
	}
	return nil
}

// assertEntityActionable verifies, before a mutation with side effects runs,
// that the request viewer may apply the given write action to the entity. It
// mirrors the predicate the ent privacy layer pins onto the mutation, so the
// mutation itself remains the enforcement; this pre-flight exists to surface
// a clean NotFoundError (404, no existence leak) instead of a silent no-op,
// and to stop side effects (attachment blob deletion, child updates) from
// running when the row write would match nothing. System contexts skip it.
func assertEntityActionable(ctx context.Context, c *ent.EntityClient, id uuid.UUID, writable func(*authz.Viewer) predicate.Entity) error {
	if authz.IsSystem(ctx) {
		return nil
	}
	v := authz.FromContext(ctx)
	if v == nil {
		// No viewer: the privacy layer denies the mutation itself.
		return nil
	}
	exists, err := c.Query().
		Where(entity.ID(id), writable(v)).
		Exist(authz.NewSystemContext(ctx))
	if err != nil {
		return err
	}
	if !exists {
		return &ent.NotFoundError{}
	}
	return nil
}

// assertEntityUpdatable: the viewer may update the entity (tenant permission
// or row grant).
func assertEntityUpdatable(ctx context.Context, c *ent.EntityClient, id uuid.UUID) error {
	return assertEntityActionable(ctx, c, id, authzrules.EntityUpdatable)
}

// assertEntityDeletable: the viewer may delete the entity.
func assertEntityDeletable(ctx context.Context, c *ent.EntityClient, id uuid.UUID) error {
	return assertEntityActionable(ctx, c, id, authzrules.EntityDeletable)
}
