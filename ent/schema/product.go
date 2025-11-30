package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

// Product holds the schema definition for the Product entity.
type Product struct {
	ent.Schema
}

// Fields of the Product.
func (Product) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.String("barcode").
			Unique().
			Nillable().
			Optional(),
		field.String("name").
			NotEmpty(),
		field.String("brand").
			Nillable().
			Optional(),
		field.Float("calories").
			Min(0),
		field.Float("protein").
			Min(0),
		field.Float("fat").
			Min(0),
		field.Float("saturated_fat").
			Min(0),
		field.Float("carbohydrates").
			Min(0),
		field.Float("fiber").
			Min(0),
		field.Float("sugars").
			Min(0),
		field.Float("sodium").
			Min(0),
		field.String("created_by").
			Nillable().
			Optional(),
	}
}

// Edges of the Product.
func (Product) Edges() []ent.Edge {
	return nil
}
