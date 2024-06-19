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

func (MaintenanceEntry) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.BaseMixin{},
	}
}

func (MaintenanceEntry) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("item_id", uuid.UUID{}),
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

// Edges of the ItemField.
func (MaintenanceEntry) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("item", Item.Type).
			Field("item_id").
			Ref("maintenance_entries").
			Required().
			Unique(),
	}
}
