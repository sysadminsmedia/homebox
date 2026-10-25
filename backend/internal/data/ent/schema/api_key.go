package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/sysadminsmedia/homebox/backend/internal/data/authzrules"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/schema/mixins"
)

// APIKey holds the schema definition for static, user-issued API keys that
// authenticate as the owning user.
type APIKey struct {
	ent.Schema
}

func (APIKey) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.BaseMixin{},
		UserMixin{
			ref:   "api_keys",
			field: "user_id",
		},
	}
}

func (APIKey) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").
			MaxLen(255).
			NotEmpty(),
		field.Bytes("token").
			Unique().
			Sensitive(),
		field.Time("expires_at").
			Optional().
			Nillable(),
		field.Time("last_used_at").
			Optional().
			Nillable(),
	}
}

func (APIKey) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("token"),
		index.Fields("user_id"),
	}
}

// Policy of the APIKey: mutations require the matching permission and are
// pinned to the viewer's tenant; reads are filtered by Interceptors.
func (APIKey) Policy() ent.Policy {
	return authzrules.NewPolicy(authzrules.APIKeyMutationRule())
}

// Interceptors of the APIKey scope every read to the request viewer.
func (APIKey) Interceptors() []ent.Interceptor {
	return []ent.Interceptor{authzrules.FilterAPIKey()}
}
