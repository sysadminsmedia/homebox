package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/schema/mixins"
)

// ItemTemplate holds the schema definition for the ItemTemplate entity.
type ItemTemplate struct {
	ent.Schema
}

func (ItemTemplate) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.BaseMixin{},
		mixins.DetailsMixin{},
		GroupMixin{ref: "item_templates"},
	}
}

func (ItemTemplate) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("name"),
	}
}

// Fields of the ItemTemplate.
func (ItemTemplate) Fields() []ent.Field {
	return []ent.Field{
		// Notes for the template (instructions, hints, etc.)
		field.String("notes").
			MaxLen(1000).
			Optional(),

		// ------------------------------------
		// Default values for item fields
		field.Int("default_quantity").
			Default(1),
		field.Bool("default_insured").
			Default(false),

		// ------------------------------------
		// Default item identification
		field.String("default_manufacturer").
			MaxLen(255).
			Optional(),

		// ------------------------------------
		// Default warranty settings
		field.Bool("default_lifetime_warranty").
			Default(false),
		field.Text("default_warranty_details").
			MaxLen(1000).
			Optional(),

		// ------------------------------------
		// Template metadata
		field.Bool("include_warranty_fields").
			Default(false).
			Comment("Whether to include warranty fields in items created from this template"),
		field.Bool("include_purchase_fields").
			Default(false).
			Comment("Whether to include purchase fields in items created from this template"),
		field.Bool("include_sold_fields").
			Default(false).
			Comment("Whether to include sold fields in items created from this template"),
	}
}

// Edges of the ItemTemplate.
func (ItemTemplate) Edges() []ent.Edge {
	owned := func(s string, t any) ent.Edge {
		return edge.To(s, t).
			Annotations(entsql.Annotation{
				OnDelete: entsql.Cascade,
			})
	}

	return []ent.Edge{
		owned("fields", TemplateField.Type),
	}
}
