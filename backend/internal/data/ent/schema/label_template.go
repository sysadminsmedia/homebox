package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/schema/mixins"
)

// LabelTemplate holds the schema definition for the LabelTemplate entity.
type LabelTemplate struct {
	ent.Schema
}

func (LabelTemplate) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.BaseMixin{},
		mixins.DetailsMixin{},
		GroupMixin{ref: "label_templates"},
	}
}

func (LabelTemplate) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("name"),
		index.Fields("is_shared"),
		index.Fields("preset"),
	}
}

// Fields of the LabelTemplate.
func (LabelTemplate) Fields() []ent.Field {
	return []ent.Field{
		// Template dimensions (in mm)
		field.Float("width").
			Default(62.0).
			Comment("Label width in mm"),
		field.Float("height").
			Default(29.0).
			Comment("Label height in mm"),

		// Preset reference (optional - for standard label sizes)
		field.String("preset").
			Optional().
			MaxLen(50).
			Comment("Preset size key like 'brother_dk2205'"),

		// Sharing settings
		field.Bool("is_shared").
			Default(false).
			Comment("Whether template is shared with group"),

		// Template canvas data (JSON containing all elements for Fabric.js)
		field.JSON("canvas_data", map[string]interface{}{}).
			Optional().
			Comment("Fabric.js compatible canvas JSON"),

		// Output settings
		field.String("output_format").
			Default("png").
			Comment("Output format: png, pdf"),
		field.Int("dpi").
			Default(300).
			Comment("Output DPI for rendering"),

		// Brother printer media type (for direct printing)
		field.String("media_type").
			Optional().
			MaxLen(50).
			Comment("Brother media type like 'DK-22251' for direct printing"),

		// Owner tracking for private templates
		field.UUID("owner_id", uuid.UUID{}).
			Comment("User who created this template"),
	}
}

// Edges of the LabelTemplate.
func (LabelTemplate) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("owner", User.Type).
			Ref("label_templates").
			Field("owner_id").
			Unique().
			Required(),
	}
}
