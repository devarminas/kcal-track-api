package product

import (
	"errors"
	"fmt"
	"strings"

	"devarminas/kcal-track-api/ent"

	"github.com/google/uuid"
)

// ProductID identifies a product in the domain.
type ProductID uuid.UUID

// NewProductID creates a new unique product identifier.
func NewProductID() ProductID {
	return ProductID(uuid.New())
}

// Product aggregates the data required to represent a nutrition product.
type Product struct {
	ID        ProductID
	Barcode   *string
	Name      string
	Brand     *string
	Nutrition Nutrition
	CreatedBy *string
}

// String returns the textual representation of the identifier.
func (id ProductID) String() string {
	return uuid.UUID(id).String()
}

// IsZero reports whether the identifier has not been assigned yet.
func (id ProductID) IsZero() bool {
	return uuid.UUID(id) == uuid.Nil
}

// Nutrition captures the macronutrient information for a product.
type Nutrition struct {
	Calories      float64
	Protein       float64
	Fat           float64
	SaturatedFat  float64
	Carbohydrates float64
	Fiber         float64
	Sugars        float64
	Sodium        float64
}

// ProductOption configures an optional value when creating a product.
type ProductOption func(*Product) error

var (
	ErrInvalidProductName    = errors.New("product requires a name")
	ErrInvalidNutritionValue = errors.New("nutrition value cannot be negative")
	ErrZeroProductID         = errors.New("product id cannot be zero")
	ErrNilEntProduct         = errors.New("ent product cannot be nil")
)

// NewProduct validates the provided information and returns a fresh aggregate.
func NewProduct(name string, nutrition Nutrition, opts ...ProductOption) (*Product, error) {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return nil, ErrInvalidProductName
	}

	if err := nutrition.Validate(); err != nil {
		return nil, err
	}

	prod := &Product{
		ID:        NewProductID(),
		Name:      trimmed,
		Nutrition: nutrition,
	}

	for _, opt := range opts {
		if err := opt(prod); err != nil {
			return nil, err
		}
	}

	return prod, nil
}

// Validate ensures nutritional values are non-negative.
func (n Nutrition) Validate() error {
	nutrients := map[string]float64{
		"calories":      n.Calories,
		"protein":       n.Protein,
		"fat":           n.Fat,
		"saturatedFat":  n.SaturatedFat,
		"carbohydrates": n.Carbohydrates,
		"fiber":         n.Fiber,
		"sugars":        n.Sugars,
		"sodium":        n.Sodium,
	}

	for name, value := range nutrients {
		if value < 0 {
			return fmt.Errorf("%s %w", name, ErrInvalidNutritionValue)
		}
	}

	return nil
}

// WithBarcode sets the optional barcode of the product.
func WithBarcode(barcode string) ProductOption {
	return func(p *Product) error {
		p.Barcode = optionalString(barcode)
		return nil
	}
}

// WithBrand sets the optional brand of the product.
func WithBrand(brand string) ProductOption {
	return func(p *Product) error {
		p.Brand = optionalString(brand)
		return nil
	}
}

// WithCreatedBy records the user that created the product.
func WithCreatedBy(createdBy string) ProductOption {
	return func(p *Product) error {
		p.CreatedBy = optionalString(createdBy)
		return nil
	}
}

// WithProductID overrides the identifier when hydrating an existing aggregate.
func WithProductID(id ProductID) ProductOption {
	return func(p *Product) error {
		if id.IsZero() {
			return ErrZeroProductID
		}
		p.ID = id
		return nil
	}
}

// FromEnt translates an ent Product into the domain aggregate.
func FromEnt(entProduct *ent.Product) (*Product, error) {
	if entProduct == nil {
		return nil, ErrNilEntProduct
	}

	nutrition := Nutrition{
		Calories:      entProduct.Calories,
		Protein:       entProduct.Protein,
		Fat:           entProduct.Fat,
		SaturatedFat:  entProduct.SaturatedFat,
		Carbohydrates: entProduct.Carbohydrates,
		Fiber:         entProduct.Fiber,
		Sugars:        entProduct.Sugars,
		Sodium:        entProduct.Sodium,
	}

	opts := []ProductOption{WithProductID(ProductID(entProduct.ID))}
	if entProduct.Barcode != nil {
		opts = append(opts, WithBarcode(*entProduct.Barcode))
	}
	if entProduct.Brand != nil {
		opts = append(opts, WithBrand(*entProduct.Brand))
	}
	if entProduct.OwnerID != nil {
		opts = append(opts, WithCreatedBy(*entProduct.OwnerID))
	}

	return NewProduct(entProduct.Name, nutrition, opts...)
}

func optionalString(value string) *string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}
