package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/schema/mixins"
)

// EntityField holds the schema definition for the EntityField entity.
type EntityField struct {
	ent.Schema
}

func (EntityField) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.BaseMixin{},
		mixins.DetailsMixin{},
	}
}

// Fields of the EntityField.
func (EntityField) Fields() []ent.Field {
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

// Edges of the EntityField.
func (EntityField) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("entity", Entity.Type).
			Ref("fields").
			Unique(),
	}
}
