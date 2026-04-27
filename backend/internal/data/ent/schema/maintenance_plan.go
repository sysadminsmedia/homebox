package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/schema/mixins"
)

type MaintenanceIntervalUnit string

const (
	MaintenanceIntervalUnitHour  MaintenanceIntervalUnit = "hour"
	MaintenanceIntervalUnitDay   MaintenanceIntervalUnit = "day"
	MaintenanceIntervalUnitWeek  MaintenanceIntervalUnit = "week"
	MaintenanceIntervalUnitMonth MaintenanceIntervalUnit = "month"
	MaintenanceIntervalUnitYear  MaintenanceIntervalUnit = "year"
)

type MaintenancePlan struct {
	ent.Schema
}

func (MaintenancePlan) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.BaseMixin{},
	}
}

func (MaintenancePlan) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("entity_id", uuid.UUID{}),
		field.String("name").
			MaxLen(255).
			NotEmpty(),
		field.String("description").
			MaxLen(2500).
			Optional(),
		field.Int("interval_value").
			Positive(),
		field.Enum("interval_unit").
			Values(
				string(MaintenanceIntervalUnitHour),
				string(MaintenanceIntervalUnitDay),
				string(MaintenanceIntervalUnitWeek),
				string(MaintenanceIntervalUnitMonth),
				string(MaintenanceIntervalUnitYear),
			),
		field.Bool("active").
			Default(true),
		field.Time("last_completed_at").
			Optional().
			Nillable(),
		field.Time("next_due_at").
			Optional().
			Nillable(),
	}
}

func (MaintenancePlan) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("entity", Entity.Type).
			Field("entity_id").
			Ref("maintenance_plans").
			Required().
			Unique(),
		edge.To("maintenance_entries", MaintenanceEntry.Type).
			Annotations(entsql.Annotation{
				OnDelete: entsql.Cascade,
			}),
	}
}
