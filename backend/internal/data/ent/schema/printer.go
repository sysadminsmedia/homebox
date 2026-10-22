package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/schema/mixins"
)

// Printer holds the schema definition for the Printer entity.
type Printer struct {
	ent.Schema
}

func (Printer) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.BaseMixin{},
		mixins.DetailsMixin{},
		GroupMixin{ref: "printers"},
	}
}

func (Printer) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("name"),
		index.Fields("is_default"),
		index.Fields("printer_type"),
	}
}

// Fields of the Printer.
func (Printer) Fields() []ent.Field {
	return []ent.Field{
		// Printer type: ipp, cups, brother_raster
		field.Enum("printer_type").
			Values("ipp", "cups", "brother_raster").
			Default("ipp").
			Comment("Type of printer connection"),

		// Connection address (IPP URI or CUPS printer name)
		field.String("address").
			MaxLen(512).
			NotEmpty().
			Comment("IPP URI (ipp://host:port/path) or CUPS printer name"),

		// Default printer flag
		field.Bool("is_default").
			Default(false).
			Comment("Whether this is the default label printer"),

		// Label dimensions for validation
		field.Float("label_width_mm").
			Optional().
			Positive().
			Comment("Expected label width in mm for validation"),
		field.Float("label_height_mm").
			Optional().
			Positive().
			Comment("Expected label height in mm for validation"),

		// Print quality
		field.Int("dpi").
			Default(300).
			Min(72).
			Max(1200).
			Comment("Printer DPI for optimal rendering"),

		// Media type identifier (optional, for IPP)
		field.String("media_type").
			Optional().
			MaxLen(100).
			Comment("Media type identifier for IPP (e.g., 'labels')"),

		// Cached status
		field.Enum("status").
			Values("online", "offline", "unknown").
			Default("unknown").
			Comment("Cached printer status"),

		// Last status check
		field.Time("last_status_check").
			Optional().
			Nillable().
			Comment("When status was last verified"),
	}
}

// Edges of the Printer.
func (Printer) Edges() []ent.Edge {
	return nil
}
