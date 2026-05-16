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

// PasswordResetTokens holds single-use reset tokens for the forgot-password flow.
// The schema mirrors AuthTokens; the only addition is `used_at`, which is set
// when the token is consumed so a replay reuses neither the row nor a new one.
type PasswordResetTokens struct {
	ent.Schema
}

func (PasswordResetTokens) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.BaseMixin{},
	}
}

func (PasswordResetTokens) Fields() []ent.Field {
	return []ent.Field{
		field.Bytes("token").
			Unique(),
		field.Time("expires_at").
			Default(func() time.Time { return time.Now().Add(time.Hour) }),
		field.Time("used_at").
			Optional().
			Nillable(),
	}
}

func (PasswordResetTokens) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("password_reset_tokens").
			Unique().
			Required().
			Annotations(entsql.Annotation{
				OnDelete: entsql.Cascade,
			}),
	}
}

func (PasswordResetTokens) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("token"),
	}
}
