package schema

import (
	"entgo.io/ent"

	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/schema/mixins"
)

type {{ .Scaffold.model }} struct {
	ent.Schema
}

func ({{ .Scaffold.model }}) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.BaseMixin{},
		{{- if .Scaffold.by_group }}
		GroupMixin{ref: "{{ snakecase .Scaffold.model  }}s"},
		{{- end }}
	}
}

// Fields of the {{ .Scaffold.model }}.
func ({{ .Scaffold.model }}) Fields() []ent.Field {
	return []ent.Field{
		// field.String("name").
	}
}

// Edges of the {{ .Scaffold.model }}.
func ({{ .Scaffold.model }}) Edges() []ent.Edge {
	return []ent.Edge{
		// edge.From("group", Group.Type).
	}
}

func ({{ .Scaffold.model }}) Indexes() []ent.Index {
	return []ent.Index{
		// index.Fields("token"),
	}
}