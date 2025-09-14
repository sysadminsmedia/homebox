package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/schema/mixins"
)

type MaintenanceEntry struct {
	ent.Schema
}

// Mixin for the MaintenanceEntry.
func (MaintenanceEntry) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.BaseMixin{},
	}
}

// Fields of the EntityField.
func (MaintenanceEntry) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("entity_id", uuid.UUID{}),
		field.Time("date").
			Optional(),
		field.Time("scheduled_date").
			Optional(),
		field.String("name").
			MaxLen(255).
			NotEmpty(),
		field.String("description").
			MaxLen(2500).
			Optional(),
		field.Float("cost").
			Default(0.0),
	}
}

// Edges of the EntityField.
func (MaintenanceEntry) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("entity", Entity.Type).
			Field("entity_id").
			Ref("maintenance_entries").
			Required().
			Unique(),
	}
}
