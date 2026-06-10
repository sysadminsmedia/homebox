package schema

import (
	"context"
	"errors"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
	"github.com/sysadminsmedia/homebox/backend/internal/data/authzrules"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/schema/mixins"
)

// AccessGrant is a row-level ACL entry on a single Entity. It targets exactly
// one of a user or a permission group and grants specific actions on that
// entity (and, via can_attachments, its attachments). Grants let a user
// without tenant-wide entity permissions access one specific item.
type AccessGrant struct {
	ent.Schema
}

func (AccessGrant) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.BaseMixin{},
		// Tenant scoping. Always the same tenant as the target entity; kept
		// denormalized for cheap tenant filters and cascade on tenant delete.
		GroupMixin{ref: "access_grants", field: "group_id"},
	}
}

func (AccessGrant) Indexes() []ent.Index {
	return []ent.Index{
		// One grant per (entity, target). NULLs are distinct in both sqlite
		// and postgres, so the two partial targets don't collide.
		index.Fields("entity_id", "user_id").
			Unique(),
		index.Fields("entity_id", "permission_group_id").
			Unique(),
		index.Fields("user_id"),
		index.Fields("permission_group_id"),
	}
}

// Fields of the AccessGrant.
func (AccessGrant) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("entity_id", uuid.UUID{}),
		field.UUID("user_id", uuid.UUID{}).
			Optional().
			Nillable(),
		field.UUID("permission_group_id", uuid.UUID{}).
			Optional().
			Nillable(),
		field.Bool("can_read").
			Default(false),
		field.Bool("can_update").
			Default(false),
		field.Bool("can_delete").
			Default(false),
		field.Bool("can_attachments").
			Default(false),
	}
}

// Edges of the AccessGrant.
func (AccessGrant) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("entity", Entity.Type).
			Ref("access_grants").
			Field("entity_id").
			Unique().
			Required(),
		edge.To("user", User.Type).
			Field("user_id").
			Unique().
			Annotations(entsql.Annotation{OnDelete: entsql.Cascade}),
		edge.To("permission_group", PermissionGroup.Type).
			Field("permission_group_id").
			Unique().
			Annotations(entsql.Annotation{OnDelete: entsql.Cascade}),
	}
}

// Hooks of the AccessGrant. Uses only the generic mutation interface so the
// schema package does not depend on generated code.
func (AccessGrant) Hooks() []ent.Hook {
	return []ent.Hook{
		func(next ent.Mutator) ent.Mutator {
			return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
				if m.Op().Is(ent.OpCreate) {
					if hasUUIDField(m, "user_id") == hasUUIDField(m, "permission_group_id") {
						return nil, errors.New("access grant must target exactly one of a user or a permission group")
					}
				}

				// update/delete/attachments imply read; normalize so query
				// predicates only ever need to check the specific action.
				for _, f := range []string{"can_update", "can_delete", "can_attachments"} {
					if v, ok := m.Field(f); ok {
						if b, _ := v.(bool); b {
							if err := m.SetField("can_read", true); err != nil {
								return nil, err
							}
							break
						}
					}
				}

				return next.Mutate(ctx, m)
			})
		},
	}
}

// hasUUIDField reports whether a non-nil UUID value is set for the field on
// this mutation.
func hasUUIDField(m ent.Mutation, name string) bool {
	v, ok := m.Field(name)
	if !ok || v == nil {
		return false
	}
	id, ok := v.(uuid.UUID)
	return ok && id != uuid.Nil
}

// Policy of the AccessGrant: mutations require the matching permission and are
// pinned to the viewer's tenant; reads are filtered by Interceptors.
func (AccessGrant) Policy() ent.Policy {
	return authzrules.NewPolicy(authzrules.AccessGrantMutationRule())
}

// Interceptors of the AccessGrant scope every read to the request viewer.
func (AccessGrant) Interceptors() []ent.Interceptor {
	return []ent.Interceptor{authzrules.FilterAccessGrant()}
}
