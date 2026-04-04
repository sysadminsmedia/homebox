package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/schema/mixins"
)

// Entity holds the schema definition for the Entity entity.
type Entity struct {
	ent.Schema
}

func (Entity) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.BaseMixin{},
		mixins.DetailsMixin{},
		GroupMixin{ref: "entities"},
	}
}

func (Entity) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("name"),
		index.Fields("manufacturer"),
		index.Fields("model_number"),
		index.Fields("serial_number"),
		index.Fields("archived"),
		index.Fields("asset_id"),
	}
}

// Fields of the Entity.
func (Entity) Fields() []ent.Field {
	return []ent.Field{
		field.String("import_ref").
			Optional().
			MaxLen(100),
		field.String("notes").
			MaxLen(1000).
			Optional(),
		field.Float("quantity").
			Default(1),
		field.Bool("insured").
			Default(false),
		field.Bool("archived").
			Default(false),
		field.Int("asset_id").
			Default(0),
		field.Bool("sync_child_entity_locations").
			Default(false),

		// ------------------------------------
		// item identification
		field.String("serial_number").
			MaxLen(255).
			Optional(),
		field.String("model_number").
			MaxLen(255).
			Optional(),
		field.String("manufacturer").
			MaxLen(255).
			Optional(),

		// ------------------------------------
		// Item Warranty
		field.Bool("lifetime_warranty").
			Default(false),
		field.Time("warranty_expires").
			Optional(),
		field.Text("warranty_details").
			MaxLen(1000).
			Optional(),

		// ------------------------------------
		// item purchase
		field.Time("purchase_time").
			Optional(),
		field.String("purchase_from").
			Optional(),
		field.Float("purchase_price").
			Default(0),

		// ------------------------------------
		// Sold Details
		field.Time("sold_time").
			Optional(),
		field.String("sold_to").
			Optional(),
		field.Float("sold_price").
			Default(0),
		field.String("sold_notes").
			MaxLen(1000).
			Optional(),
	}
}

// Edges of the Entity.
func (Entity) Edges() []ent.Edge {
	owned := func(s string, t any) ent.Edge {
		return edge.To(s, t).
			Annotations(entsql.Annotation{
				OnDelete: entsql.Cascade,
			})
	}

	return []ent.Edge{
		edge.To("children", Entity.Type).
			From("parent").
			Unique(),
		edge.From("tag", Tag.Type).
			Ref("entities"),
		edge.From("entity_type", EntityType.Type).
			Ref("entities").
			Unique().
			Required(),
		owned("fields", EntityField.Type),
		owned("maintenance_entries", MaintenanceEntry.Type),
		owned("attachments", Attachment.Type),
	}
}
