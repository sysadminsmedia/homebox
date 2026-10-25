package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/sysadminsmedia/homebox/backend/internal/data/authzrules"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/schema/mixins"
)

// Tag holds the schema definition for the Tag entity.
type Tag struct {
	ent.Schema
}

func (Tag) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.BaseMixin{},
		mixins.DetailsMixin{},
		GroupMixin{ref: "tags"},
	}
}

// Fields of the Tag.
func (Tag) Fields() []ent.Field {
	return []ent.Field{
		field.String("color").
			MaxLen(255).
			Optional(),
		field.String("icon").
			MaxLen(255).
			Optional(),
	}
}

// Edges of the Tag.
func (Tag) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("entities", Entity.Type),
		edge.To("children", Tag.Type).
			From("parent").
			Unique(),
	}
}

// Policy of the Tag: mutations require the matching permission and are
// pinned to the viewer's tenant; reads are filtered by Interceptors.
func (Tag) Policy() ent.Policy {
	return authzrules.NewPolicy(authzrules.TagMutationRule())
}

// Interceptors of the Tag scope every read to the request viewer.
func (Tag) Interceptors() []ent.Interceptor {
	return []ent.Interceptor{authzrules.FilterTag()}
}
