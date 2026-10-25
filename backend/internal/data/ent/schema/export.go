package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/sysadminsmedia/homebox/backend/internal/data/authzrules"
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
	}
}

func (Export) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("group_id"),
		index.Fields("group_id", "status"),
	}
}

// Policy of the Export: mutations require the matching permission and are
// pinned to the viewer's tenant; reads are filtered by Interceptors.
func (Export) Policy() ent.Policy {
	return authzrules.NewPolicy(authzrules.ExportMutationRule())
}

// Interceptors of the Export scope every read to the request viewer.
func (Export) Interceptors() []ent.Interceptor {
	return []ent.Interceptor{authzrules.FilterExport()}
}
