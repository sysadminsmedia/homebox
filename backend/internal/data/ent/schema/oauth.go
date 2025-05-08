package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/schema/mixins"
)

type OAuth struct {
	ent.Schema
}

func (OAuth) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.BaseMixin{},
	}
}

func (OAuth) Fields() []ent.Field {
	return []ent.Field{
		field.String("provider").
			NotEmpty(),
		field.String("sub").
			NotEmpty(),
	}
}

func (OAuth) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("oauth").
			Unique(),
	}
}

func (OAuth) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("provider", "sub"),
	}
}
