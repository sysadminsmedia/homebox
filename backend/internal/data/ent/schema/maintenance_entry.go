package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
	"github.com/sysadminsmedia/homebox/backend/internal/data/authzrules"
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

// Edges of the MaintenanceEntry.
func (MaintenanceEntry) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("entity", Entity.Type).
			Field("entity_id").
			Ref("maintenance_entries").
			Required().
			Unique(),
	}
}

// Policy of the MaintenanceEntry: mutations require the matching permission and are
// pinned to the viewer's tenant; reads are filtered by Interceptors.
func (MaintenanceEntry) Policy() ent.Policy {
	return authzrules.NewPolicy(authzrules.MaintenanceEntryMutationRule())
}

// Interceptors of the MaintenanceEntry scope every read to the request viewer.
func (MaintenanceEntry) Interceptors() []ent.Interceptor {
	return []ent.Interceptor{authzrules.FilterMaintenanceEntry()}
}
