package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/schema/mixins"
)

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

func (EntityType) Fields() []ent.Field {
	return []ent.Field{
		field.String("icon").
			MaxLen(255).
			Optional(),
		field.String("color").
			MaxLen(255).
			Optional(),
		field.Bool("location_type").
			Default(false),
	}
}

func (EntityType) Indexes() []ent.Index {
	return []ent.Index{
		// Unique index on the "title" field.
		index.Fields("name"),
		index.Fields("location_type"),
	}
}

func (EntityType) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("entities", Entity.Type),
	}
}
