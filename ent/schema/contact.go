package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

// Contact holds the schema definition for the Contact entity.
type Contact struct {
	ent.Schema
}

// Fields of the Contact.
func (Contact) Fields() []ent.Field {
	return []ent.Field{
		field.String("email").
			NotEmpty(),
		field.String("link").
			NotEmpty(),
		field.String("type").
			NotEmpty(),
		field.String("message").
			NotEmpty(),
		field.Time("created_at").
			Default(time.Now).Immutable(),
	}
}

// Edges of the Contact.
func (Contact) Edges() []ent.Edge {
	return nil
}
