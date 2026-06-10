package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/sysadminsmedia/homebox/backend/internal/data/authzrules"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/schema/mixins"
)

// EntityType holds the schema definition for the EntityType entity.
type EntityType struct {
	ent.Schema
}

func (EntityType) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.BaseMixin{},
		mixins.DetailsMixin{},
		GroupMixin{ref: "entity_types"},
	}
}

// Fields of the EntityType.
func (EntityType) Fields() []ent.Field {
	return []ent.Field{
		field.Bool("is_location").
			Default(false),
		field.String("icon").
			MaxLen(255).
			Optional(),
	}
}

// Edges of the EntityType.
func (EntityType) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("entities", Entity.Type).
			Annotations(entsql.Annotation{
				OnDelete: entsql.Restrict,
			}),
		edge.To("default_template", EntityTemplate.Type).
			Unique(),
	}
}

// Policy of the EntityType: mutations require the matching permission and are
// pinned to the viewer's tenant; reads are filtered by Interceptors.
func (EntityType) Policy() ent.Policy {
	return authzrules.NewPolicy(authzrules.EntityTypeMutationRule())
}

// Interceptors of the EntityType scope every read to the request viewer.
func (EntityType) Interceptors() []ent.Interceptor {
	return []ent.Interceptor{authzrules.FilterEntityType()}
}
