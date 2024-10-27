package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/schema/mixins"
)

// AuthTokens holds the schema definition for the AuthTokens entity.
type AuthTokens struct {
	ent.Schema
}

func (AuthTokens) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.BaseMixin{},
	}
}

// Fields of the AuthTokens.
func (AuthTokens) Fields() []ent.Field {
	return []ent.Field{
		field.String("token").
			Unique().MaxLen(255),
		field.Time("expires_at").
			Default(func() time.Time { return time.Now().Add(time.Hour * 24 * 7) }),
	}
}

// Edges of the AuthTokens.
func (AuthTokens) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("auth_tokens").
			Unique(),
		edge.To("roles", AuthRoles.Type).
			Unique().
			Annotations(entsql.Annotation{
				OnDelete: entsql.Cascade,
			}),
	}
}

func (AuthTokens) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("token"),
	}
}
