package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/schema/mixins"
)

// Export holds the schema definition for the Export entity. An Export row
// tracks a collection-archive job: its lifecycle status and, on completion,
// the blob storage key for the produced zip artifact.
type Export struct {
	ent.Schema
}

func (Export) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.BaseMixin{},
		GroupMixin{
			ref:   "exports",
			field: "group_id",
		},
	}
}

func (Export) Fields() []ent.Field {
	return []ent.Field{
		field.Enum("status").
			Values("pending", "running", "completed", "failed").
			Default("pending"),
		field.Int("progress").
			Default(0),
		field.String("artifact_path").
			Optional(),
		field.Int64("size_bytes").
			Default(0),
		field.String("error").
			MaxLen(1000).
			Optional(),
	}
}

func (Export) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("group_id"),
		index.Fields("group_id", "status"),
	}
}
