package schema

import "entgo.io/ent"

// Migration holds the schema definition for the Migration entity.
type Migration struct {
	ent.Schema
}

// Fields of the Migration.
func (Migration) Fields() []ent.Field {
	return nil
}

// Edges of the Migration.
func (Migration) Edges() []ent.Edge {
	return nil
}
