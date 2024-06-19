package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/schema/mixins"
)

type Notifier struct {
	ent.Schema
}

func (Notifier) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.BaseMixin{},
		GroupMixin{
			ref:   "notifiers",
			field: "group_id",
		},
		UserMixin{
			ref:   "notifiers",
			field: "user_id",
		},
	}
}

// Fields of the Notifier.
func (Notifier) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").
			MaxLen(255).
			NotEmpty(),
		field.String("url").
			Sensitive().
			MaxLen(2083). // supposed max length of URL
			NotEmpty(),
		field.Bool("is_active").
			Default(true),
	}
}

func (Notifier) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id"),
		index.Fields("user_id", "is_active"),
		index.Fields("group_id"),
		index.Fields("group_id", "is_active"),
	}
}
