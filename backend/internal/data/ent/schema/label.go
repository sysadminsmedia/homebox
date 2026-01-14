package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
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

func (Label) Indexes() []ent.Index {
	return []ent.Index{
		// Index on name for sorting and searching
		index.Fields("name"),
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
		edge.To("items", Item.Type),
	}
}
