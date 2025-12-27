package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"entgo.io/ent/schema/mixin"
	"github.com/google/uuid"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/schema/mixins"
)

// User holds the schema definition for the User entity.
type User struct {
	ent.Schema
}

func (User) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.BaseMixin{},
	}
}

// Fields of the User.
func (User) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").
			MaxLen(255).
			NotEmpty(),
		field.String("email").
			MaxLen(255).
			NotEmpty().
			Unique(),
		field.String("password").
			MaxLen(255).
			Nillable().
			Optional().
			Sensitive(),
		field.Bool("is_superuser").
			Default(false),
		field.Bool("superuser").
			Default(false),
		field.Enum("role").
			Default("user").
			Values("user", "owner"),
		field.Time("activated_on").
			Optional(),
		// OIDC identity mapping fields (issuer + subject)
		field.String("oidc_issuer").
			Optional().
			Nillable(),
		field.String("oidc_subject").
			Optional().
			Nillable(),
		// default_group_id is the user's primary tenant/group
		field.UUID("default_group_id", uuid.UUID{}).
			Optional().
			Nillable(),
	}
}

func (User) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("oidc_issuer", "oidc_subject").Unique(),
	}
}

// Edges of the User.
func (User) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("groups", Group.Type),
		edge.To("auth_tokens", AuthTokens.Type).
			Annotations(entsql.Annotation{
				OnDelete: entsql.Cascade,
			}),
		edge.To("notifiers", Notifier.Type).
			Annotations(entsql.Annotation{
				OnDelete: entsql.Cascade,
			}),
	}
}

// UserMixin when embedded in an ent.Schema, adds a reference to
// the Group entity.
type UserMixin struct {
	ref   string
	field string
	mixin.Schema
}

func (g UserMixin) Fields() []ent.Field {
	if g.field != "" {
		return []ent.Field{
			field.UUID(g.field, uuid.UUID{}),
		}
	}

	return nil
}

func (g UserMixin) Edges() []ent.Edge {
	e := edge.From("user", User.Type).
		Ref(g.ref).
		Unique().
		Required()

	if g.field != "" {
		e = e.Field(g.field)
	}

	return []ent.Edge{e}
}
