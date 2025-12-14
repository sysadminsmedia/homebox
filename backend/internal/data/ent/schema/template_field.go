package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/schema/mixins"
)

// TemplateField holds the schema definition for the TemplateField entity.
// Template fields define custom fields that will be added to items created from a template.
type TemplateField struct {
	ent.Schema
}

func (TemplateField) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.BaseMixin{},
		mixins.DetailsMixin{},
	}
}

// Fields of the TemplateField.
func (TemplateField) Fields() []ent.Field {
	return []ent.Field{
		field.Enum("type").
			Values("text"),
		field.String("text_value").
			MaxLen(500).
			Optional(),
	}
}

// Edges of the TemplateField.
func (TemplateField) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("item_template", ItemTemplate.Type).
			Ref("fields").
			Unique(),
	}
}
