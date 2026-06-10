package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/sysadminsmedia/homebox/backend/internal/data/authz"
	"github.com/sysadminsmedia/homebox/backend/internal/data/authzrules"
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
		// Tenant-wide permissions applied to the membership created when the
		// invitation is accepted. Defaults to the full-access wildcard so
		// invites behave like they did before the permission system existed
		// and keep covering permissions added to the catalog later.
		field.Strings("permissions").
			Default(authz.FullAccess()),
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

// Policy of the GroupInvitationToken: mutations require the matching permission and are
// pinned to the viewer's tenant; reads are filtered by Interceptors.
func (GroupInvitationToken) Policy() ent.Policy {
	return authzrules.NewPolicy(authzrules.GroupInvitationTokenMutationRule())
}

// Interceptors of the GroupInvitationToken scope every read to the request viewer.
func (GroupInvitationToken) Interceptors() []ent.Interceptor {
	return []ent.Interceptor{authzrules.FilterGroupInvitationToken()}
}
