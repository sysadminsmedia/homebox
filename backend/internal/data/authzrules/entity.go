package authzrules

import (
	"context"

	"github.com/google/uuid"
	"github.com/sysadminsmedia/homebox/backend/internal/data/authz"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/accessgrant"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/entity"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/group"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/predicate"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/privacy"
)

// GrantTarget matches access grants addressed to the viewer, either directly
// or through one of their permission groups.
func GrantTarget(v *authz.Viewer) predicate.AccessGrant {
	if len(v.PermGroupIDs) == 0 {
		return accessgrant.UserID(v.UserID)
	}
	return accessgrant.Or(
		accessgrant.UserID(v.UserID),
		accessgrant.PermissionGroupIDIn(v.PermGroupIDs...),
	)
}

// grantWith matches entities carrying a grant for the viewer with all the
// given action predicates.
func grantWith(v *authz.Viewer, actions ...predicate.AccessGrant) predicate.Entity {
	preds := make([]predicate.AccessGrant, 0, len(actions)+1)
	preds = append(preds, GrantTarget(v))
	preds = append(preds, actions...)
	return entity.HasAccessGrantsWith(preds...)
}

// EntityReadable matches entities the viewer may read: anything in their
// tenant when they hold entity:read, plus individually granted rows. Grants
// are tenant-pinned by construction (they live in the entity's tenant), so a
// grant can never leak rows across tenants.
func EntityReadable(v *authz.Viewer) predicate.Entity {
	granted := grantWith(v, accessgrant.CanRead(true))
	if v.Has(authz.PermEntityRead) {
		return entity.Or(
			entity.HasGroupWith(group.ID(v.TenantID)),
			granted,
		)
	}
	return granted
}

// entityActionable matches entities the viewer may apply a write action to:
// tenant-wide when they hold perm, plus rows granted with the action.
func entityActionable(v *authz.Viewer, perm authz.Permission, actions ...predicate.AccessGrant) predicate.Entity {
	granted := grantWith(v, actions...)
	if v.Has(perm) {
		return entity.Or(
			entity.HasGroupWith(group.ID(v.TenantID)),
			granted,
		)
	}
	return granted
}

// EntityUpdatable matches entities the viewer may update.
func EntityUpdatable(v *authz.Viewer) predicate.Entity {
	return entityActionable(v, authz.PermEntityUpdate, accessgrant.CanUpdate(true))
}

// EntityDeletable matches entities the viewer may delete.
func EntityDeletable(v *authz.Viewer) predicate.Entity {
	return entityActionable(v, authz.PermEntityDelete, accessgrant.CanDelete(true))
}

// EntityAttachmentsWritable matches entities whose attachments the viewer may
// create/modify/delete: tenant-wide entity:update, or a row grant with
// can_update or can_attachments.
func EntityAttachmentsWritable(v *authz.Viewer) predicate.Entity {
	return entityActionable(v, authz.PermEntityUpdate,
		accessgrant.Or(accessgrant.CanUpdate(true), accessgrant.CanAttachments(true)))
}

// EntityChildrenWritable matches entities whose child rows (maintenance
// entries, custom fields) the viewer may mutate.
func EntityChildrenWritable(v *authz.Viewer) predicate.Entity {
	return entityActionable(v, authz.PermEntityUpdate, accessgrant.CanUpdate(true))
}

// EntityMutationRule authorizes entity create/update/delete.
func EntityMutationRule() privacy.MutationRule {
	return privacy.EntityMutationRuleFunc(func(ctx context.Context, m *ent.EntityMutation) error {
		v, err := viewerFor(ctx)
		if v == nil {
			return err
		}

		switch {
		case m.Op().Is(ent.OpCreate):
			if !v.Has(authz.PermEntityCreate) {
				return privacy.Denyf("authz: missing %s", authz.PermEntityCreate)
			}
			return allowIfTenantCreate(m, v)

		case m.Op().Is(ent.OpDelete | ent.OpDeleteOne):
			return allowEntityWrite(m, v, EntityDeletable(v))

		default: // update
			return allowEntityWrite(m, v, EntityUpdatable(v))
		}
	})
}

// allowEntityWrite allows an entity update/delete and pins the mutation to
// rows the viewer may write. With the tenant permission the pin is the
// tenant-or-grant predicate; without it the pin is grant-covered rows only.
// Rows outside the pin match nothing, so cross-tenant or ungranted writes
// become no-ops that surface as NotFoundError (404, no existence leak).
func allowEntityWrite(m *ent.EntityMutation, v *authz.Viewer, writable predicate.Entity) error {
	if changesTenant(m, v) {
		return privacy.Denyf("authz: cross-tenant write")
	}
	m.Where(writable)
	return privacy.Allow
}

// changesTenant reports whether the mutation sets the entity's tenant to a
// group other than the viewer's.
func changesTenant(m *ent.EntityMutation, v *authz.Viewer) bool {
	gid, ok := m.GroupID()
	return ok && gid != v.TenantID
}

// allowIfTenantCreate allows a create whose group (when set on the mutation)
// is the viewer's tenant. The required-edge validation rejects creates with
// no group at all.
func allowIfTenantCreate(m interface{ GroupID() (uuid.UUID, bool) }, v *authz.Viewer) error {
	if gid, ok := m.GroupID(); ok && gid != v.TenantID {
		return privacy.Denyf("authz: cross-tenant create")
	}
	return privacy.Allow
}

// childMutation is the common surface of mutations on entity-owned child
// resources (attachments, maintenance entries, fields).
type childMutation interface {
	Op() ent.Op
	EntityID() (uuid.UUID, bool)
	Client() *ent.Client
}

// checkChildParent authorizes the parent side of a child-resource mutation:
// when the mutation sets the entity edge (create, or re-parenting update),
// the target parent must be writable by the viewer. The actual rows touched
// by update/delete are pinned by the caller via m.Where(HasEntityWith(...)),
// so this is the only check that needs a query.
//
// Returns privacy.Allow on success so it can be returned directly.
func checkChildParent(ctx context.Context, m childMutation, writable predicate.Entity) error {
	eid, ok := m.EntityID()
	if !ok {
		// Create without a parent fails required-edge validation later;
		// update/delete without re-parenting is fully covered by the pin.
		return privacy.Allow
	}

	// System context: the predicate itself encodes what the viewer may
	// write, and the surrounding policy already validated the viewer.
	exists, err := m.Client().Entity.Query().
		Where(entity.ID(eid), writable).
		Exist(authz.NewSystemContext(ctx))
	if err != nil {
		return privacy.Denyf("authz: checking parent entity: %v", err)
	}
	if !exists {
		return privacy.Denyf("authz: parent entity not writable")
	}
	return privacy.Allow
}
