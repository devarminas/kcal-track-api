package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

// UserLog holds the schema definition for the UserLog entity.
type UserLog struct {
	ent.Schema
}

// Fields of the UserLog.
func (UserLog) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),

		field.Float("amount").
			Min(0),

		field.Time("date").
			Default(time.Now),

		field.String("user_id").
			Nillable().
			Optional(),
		field.UUID("product_id", uuid.UUID{}),
	}
}

// Edges of the UserLog.
func (UserLog) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("product", Product.Type).
			Ref("user_logs").
			Field("product_id").
			Unique().
			Required(),
	}
}
