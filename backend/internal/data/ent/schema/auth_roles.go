package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/sysadminsmedia/homebox/backend/internal/data/authzrules"
)

// AuthRoles holds the schema definition for the AuthRoles entity.
type AuthRoles struct {
	ent.Schema
}

// Fields of the AuthRoles.
func (AuthRoles) Fields() []ent.Field {
	return []ent.Field{
		field.Enum("role").
			Default("user").
			Values(
				"admin",       // can do everything - currently unused
				"user",        // default login role
				"attachments", // Read Attachments
			),
	}
}

// Edges of the AuthRoles.
func (AuthRoles) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("token", AuthTokens.Type).
			Ref("roles").
			Unique(),
	}
}

// Policy of the AuthRoles: mutations require the matching permission and are
// pinned to the viewer's tenant; reads are filtered by Interceptors.
func (AuthRoles) Policy() ent.Policy {
	return authzrules.NewPolicy()
}

// Interceptors of the AuthRoles scope every read to the request viewer.
func (AuthRoles) Interceptors() []ent.Interceptor {
	return []ent.Interceptor{authzrules.FilterAuthRoles()}
}
