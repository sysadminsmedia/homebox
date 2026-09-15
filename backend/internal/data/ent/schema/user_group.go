package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

// UserGroup is the through entity for the User<->Group M:M relation. It carries
// the per-membership role so that "owner" is scoped to a single group rather
// than being a global flag on the user.
type UserGroup struct {
	ent.Schema
}

func (UserGroup) Annotations() []schema.Annotation {
	return []schema.Annotation{
		field.ID("user_id", "group_id"),
		entsql.Annotation{Table: "user_groups"},
	}
}

func (UserGroup) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("user_id", uuid.UUID{}),
		field.UUID("group_id", uuid.UUID{}),
		field.Enum("role").
			Values("user", "owner").
			Default("user"),
	}
}

func (UserGroup) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("user", User.Type).
			Field("user_id").
			Unique().
			Required().
			Annotations(entsql.Annotation{OnDelete: entsql.Cascade}),
		edge.To("group", Group.Type).
			Field("group_id").
			Unique().
			Required().
			Annotations(entsql.Annotation{OnDelete: entsql.Cascade}),
	}
}
