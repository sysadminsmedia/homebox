package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/schema/mixins"
)

// Label holds the schema definition for the Label entity.
type Label struct {
	ent.Schema
}

func (Label) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.BaseMixin{},
		mixins.DetailsMixin{},
		GroupMixin{ref: "labels"},
	}
}

// Fields of the Label.
func (Label) Fields() []ent.Field {
	return []ent.Field{
		field.String("color").
			MaxLen(255).
			Optional(),
	}
}

// Edges of the Label.
func (Label) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("children", Label.Type).
			Annotations(entsql.Annotation{
				OnDelete: entsql.Cascade,
			}).
			From("parent").
			Unique(),
		edge.To("items", Item.Type),
	}
}
