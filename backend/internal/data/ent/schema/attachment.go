package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/sysadminsmedia/homebox/backend/internal/data/authzrules"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/schema/mixins"
)

// Attachment holds the schema definition for the Attachment entity.
type Attachment struct {
	ent.Schema
}

func (Attachment) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.BaseMixin{},
	}
}

// Fields of the Attachment.
func (Attachment) Fields() []ent.Field {
	return []ent.Field{
		field.Enum("type").Values("photo", "manual", "warranty", "attachment", "receipt", "thumbnail").Default("attachment"),
		field.Bool("primary").Default(false),
		field.String("title").Default(""),
		field.String("path").Default(""),
		field.String("mime_type").Default("application/octet-stream"),
	}
}

// Edges of the Attachment.
func (Attachment) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("entity", Entity.Type).
			Ref("attachments").
			Unique(),
		edge.To("thumbnail", Attachment.Type).
			Unique(),
	}
}

// Policy of the Attachment: mutations require the matching permission and are
// pinned to the viewer's tenant; reads are filtered by Interceptors.
func (Attachment) Policy() ent.Policy {
	return authzrules.NewPolicy(authzrules.AttachmentMutationRule())
}

// Interceptors of the Attachment scope every read to the request viewer.
func (Attachment) Interceptors() []ent.Interceptor {
	return []ent.Interceptor{authzrules.FilterAttachment()}
}
