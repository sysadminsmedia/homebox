package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/sysadminsmedia/homebox/backend/internal/data/authzrules"
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
			Values("text", "number", "boolean", "time"),
		field.String("text_value").
			MaxLen(500).
			Optional(),
		field.Int("number_value").
			Optional(),
		field.Bool("boolean_value").
			Default(false),
		field.Time("time_value").
			Default(time.Now),
	}
}

// Edges of the TemplateField.
func (TemplateField) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("entity_template", EntityTemplate.Type).
			Ref("fields").
			Unique(),
	}
}

// Policy of the TemplateField: mutations require the matching permission and are
// pinned to the viewer's tenant; reads are filtered by Interceptors.
func (TemplateField) Policy() ent.Policy {
	return authzrules.NewPolicy(authzrules.TemplateFieldMutationRule())
}

// Interceptors of the TemplateField scope every read to the request viewer.
func (TemplateField) Interceptors() []ent.Interceptor {
	return []ent.Interceptor{authzrules.FilterTemplateField()}
}
