package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
	"github.com/google/uuid"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/schema/mixins"
)

// Group holds the schema definition for the Group entity.
type Group struct {
	ent.Schema
}

func (Group) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.BaseMixin{},
	}
}

// Fields of the Home.
func (Group) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").
			MaxLen(255).
			NotEmpty(),
		field.String("currency").
			Default("usd"),
	}
}

// Edges of the Home.
func (Group) Edges() []ent.Edge {
	owned := func(name string, t any) ent.Edge {
		return edge.To(name, t).
			Annotations(entsql.Annotation{
				OnDelete: entsql.Cascade,
			})
	}

	return []ent.Edge{
		// Use edge.From + Ref("groups") to model M:M between users and groups via junction table
		edge.From("users", User.Type).Ref("groups"),
		owned("locations", Location.Type),
		owned("items", Item.Type),
		owned("tags", Tag.Type),
		owned("invitation_tokens", GroupInvitationToken.Type),
		owned("notifiers", Notifier.Type),
		owned("item_templates", ItemTemplate.Type),
		// $scaffold_edge
	}
}

// GroupMixin when embedded in an ent.Schema, adds a reference to
// the Group entity.
type GroupMixin struct {
	ref   string
	field string
	mixin.Schema
}

func (g GroupMixin) Fields() []ent.Field {
	if g.field != "" {
		return []ent.Field{
			field.UUID(g.field, uuid.UUID{}),
		}
	}

	return nil
}

func (g GroupMixin) Edges() []ent.Edge {
	e := edge.From("group", Group.Type).
		Ref(g.ref).
		Unique().
		Required()

	if g.field != "" {
		e = e.Field(g.field)
	}

	return []ent.Edge{e}
}
