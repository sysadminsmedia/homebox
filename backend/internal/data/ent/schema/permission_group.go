package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/sysadminsmedia/homebox/backend/internal/data/authzrules"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/schema/mixins"
)

// PermissionGroup is a named, tenant-scoped bundle of permissions. Users who
// are members of a permission group hold all of its permissions within the
// owning tenant (Group). Not to be confused with Group, which is the tenant.
type PermissionGroup struct {
	ent.Schema
}

func (PermissionGroup) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.BaseMixin{},
		mixins.DetailsMixin{},
		GroupMixin{ref: "permission_groups", field: "group_id"},
	}
}

func (PermissionGroup) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("name", "group_id").
			Unique(),
	}
}

// Fields of the PermissionGroup.
func (PermissionGroup) Fields() []ent.Field {
	return []ent.Field{
		// Tenant-wide permissions (authz.Permission strings) held by every
		// member of this permission group.
		field.Strings("permissions").
			Default([]string{}),
	}
}

// Edges of the PermissionGroup.
func (PermissionGroup) Edges() []ent.Edge {
	return []ent.Edge{
		// Members. Users must also be members of the owning tenant; that
		// invariant is enforced in the service layer.
		edge.To("users", User.Type),
	}
}

// Policy of the PermissionGroup: mutations require the matching permission and are
// pinned to the viewer's tenant; reads are filtered by Interceptors.
func (PermissionGroup) Policy() ent.Policy {
	return authzrules.NewPolicy(authzrules.PermissionGroupMutationRule())
}

// Interceptors of the PermissionGroup scope every read to the request viewer.
func (PermissionGroup) Interceptors() []ent.Interceptor {
	return []ent.Interceptor{authzrules.FilterPermissionGroup()}
}
