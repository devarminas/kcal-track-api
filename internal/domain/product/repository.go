package product

import (
	"context"
	"errors"
	"fmt"

	"devarminas/kcal-track-api/ent"

	"github.com/google/uuid"
)

// ErrNilProduct is returned when the repository is asked to persist a nil aggregate.
var ErrNilProduct = errors.New("product repository requires a product")

// Repository persists product aggregates inside a transaction.
type Repository interface {
	Save(ctx context.Context, prod *Product) (*ent.Product, error)
}

type repository struct {
	client *ent.Client
}

// NewRepository returns a repository backed by the supplied ent client.
func NewRepository(client *ent.Client) Repository {
	return &repository{client: client}
}

// Save inserts a product inside a transaction so the operation is atomic.
func (r *repository) Save(ctx context.Context, prod *Product) (*ent.Product, error) {
	if prod == nil {
		return nil, ErrNilProduct
	}

	tx, err := r.client.Tx(ctx)
	if err != nil {
		return nil, fmt.Errorf("product repository: begin tx: %w", err)
	}

	entProduct, err := r.saveWithTx(ctx, tx, prod)
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return nil, fmt.Errorf("%w (rollback failed: %v)", err, rbErr)
		}
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("product repository: commit tx: %w", err)
	}

	return entProduct, nil
}

func (r *repository) saveWithTx(ctx context.Context, tx *ent.Tx, prod *Product) (*ent.Product, error) {
	builder := tx.Product.Create().
		SetName(prod.Name).
		SetCalories(prod.Nutrition.Calories).
		SetProtein(prod.Nutrition.Protein).
		SetFat(prod.Nutrition.Fat).
		SetSaturatedFat(prod.Nutrition.SaturatedFat).
		SetCarbohydrates(prod.Nutrition.Carbohydrates).
		SetFiber(prod.Nutrition.Fiber).
		SetSugars(prod.Nutrition.Sugars).
		SetSodium(prod.Nutrition.Sodium)

	if !prod.ID.IsZero() {
		builder.SetID(uuid.UUID(prod.ID))
	}

	if prod.Barcode != nil {
		builder.SetNillableBarcode(prod.Barcode)
	}
	if prod.Brand != nil {
		builder.SetNillableBrand(prod.Brand)
	}
	if prod.CreatedBy != nil {
		builder.SetNillableCreatedBy(prod.CreatedBy)
	}

	return builder.Save(ctx)
}
