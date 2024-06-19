package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/schema/mixins"
)

// GroupInvitationToken holds the schema definition for the GroupInvitationToken entity.
type GroupInvitationToken struct {
	ent.Schema
}

func (GroupInvitationToken) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.BaseMixin{},
	}
}

// Fields of the GroupInvitationToken.
func (GroupInvitationToken) Fields() []ent.Field {
	return []ent.Field{
		field.Bytes("token").
			Unique(),
		field.Time("expires_at").
			Default(func() time.Time { return time.Now().Add(time.Hour * 24 * 7) }),
		field.Int("uses").
			Default(0),
	}
}

// Edges of the GroupInvitationToken.
func (GroupInvitationToken) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("group", Group.Type).
			Ref("invitation_tokens").
			Unique(),
	}
}
