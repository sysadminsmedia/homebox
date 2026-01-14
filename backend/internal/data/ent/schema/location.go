package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/index"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/schema/mixins"
)

// Location holds the schema definition for the Location entity.
type Location struct {
	ent.Schema
}

func (Location) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.BaseMixin{},
		mixins.DetailsMixin{},
		GroupMixin{ref: "locations"},
	}
}

func (Location) Indexes() []ent.Index {
	return []ent.Index{
		// Index on name for sorting and searching
		index.Fields("name"),
	}
}

// Fields of the Location.
func (Location) Fields() []ent.Field {
	return nil
}

// Edges of the Location.
func (Location) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("children", Location.Type).
			From("parent").
			Unique(),
		edge.To("items", Item.Type).
			Annotations(entsql.Annotation{
				OnDelete: entsql.Cascade,
			}),
	}
}
