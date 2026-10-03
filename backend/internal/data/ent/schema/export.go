package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"

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
		// kind distinguishes server-produced export artifacts from
		// user-uploaded import zips. The whole row lifecycle (status,
		// progress, error) applies identically to both flavors — only the
		// terminal action differs ("download" vs "restore"). Keeping them
		// in one table avoids duplicating the entire job-tracking schema.
		field.Enum("kind").
			Values("export", "import").
			Default("export"),
		field.Enum("status").
			Values("pending", "running", "completed", "failed").
			Default("pending"),
		field.Int("progress").
			Default(0),
		// artifact_path is the blob key this row points at: for kind=export
		// it's the server-produced zip; for kind=import it's the upload
		// staged at "{gid}/imports/{uuid}.zip" before the worker restores
		// it.
		field.String("artifact_path").
			Optional(),
		field.Int64("size_bytes").
			Default(0),
		field.String("error").
			MaxLen(1000).
			Optional(),
		// origin separates on-demand exports (subject to the 7-day sweep)
		// from scheduled backups (pruned by the destination's retention).
		field.Enum("origin").
			Values("manual", "scheduled").
			Default("manual"),
		// destination_id points at the backup destination holding the
		// artifact; nil means the primary storage. A plain column rather than
		// an edge so deleting a destination never cascades into history.
		field.UUID("destination_id", uuid.UUID{}).
			Optional().
			Nillable(),
	}
}

func (Export) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("group_id"),
		index.Fields("group_id", "status"),
	}
}
