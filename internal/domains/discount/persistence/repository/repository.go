package discount

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"

	"api/internal/di"
	db "api/internal/domains/discount/persistence/sqlc/generated"
	values "api/internal/domains/discount/values"
	errLib "api/internal/libs/errors"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type Repository struct {
	Queries *db.Queries
	Tx      *sql.Tx
}

func NewRepository(container *di.Container) *Repository {
	return &Repository{Queries: container.Queries.DiscountDb}
}

func (r *Repository) GetTx() *sql.Tx { return r.Tx }

func (r *Repository) WithTx(tx *sql.Tx) *Repository {
	return &Repository{Queries: r.Queries.WithTx(tx), Tx: tx}
}

func mapDbDiscount(d db.Discount) values.ReadValues {
	val := values.ReadValues{
		ID: d.ID,
		CreateValues: values.CreateValues{
			Name:            d.Name,
			Description:     d.Description.String,
			DiscountPercent: int(d.DiscountPercent),
			IsUseUnlimited:  d.IsUseUnlimited,
			UsePerClient:    int(d.UsePerClient.Int32),
			IsActive:        d.IsActive,
			ValidFrom:       d.ValidFrom,
			ValidTo:         d.ValidTo.Time,
		},
		CreatedAt: d.CreatedAt,
		UpdatedAt: d.UpdatedAt,
	}
	return val
}

func (r *Repository) Create(ctx context.Context, details values.CreateValues) (values.ReadValues, *errLib.CommonError) {
	params := db.CreateDiscountParams{
		Name:            details.Name,
		Description:     sql.NullString{String: details.Description, Valid: details.Description != ""},
		DiscountPercent: int32(details.DiscountPercent),
		IsUseUnlimited:  details.IsUseUnlimited,
		UsePerClient:    sql.NullInt32{Int32: int32(details.UsePerClient), Valid: details.UsePerClient > 0},
		IsActive:        details.IsActive,
		ValidFrom:       details.ValidFrom,
		ValidTo:         sql.NullTime{Time: details.ValidTo, Valid: !details.ValidTo.IsZero()},
	}
	d, err := r.Queries.CreateDiscount(ctx, params)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			return values.ReadValues{}, errLib.New("Discount already exists", http.StatusConflict)
		}
		log.Printf("Failed to create discount: %v", err)
		return values.ReadValues{}, errLib.New("Internal server error", http.StatusInternalServerError)
	}
	return mapDbDiscount(d), nil
}

func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (values.ReadValues, *errLib.CommonError) {
	d, err := r.Queries.GetDiscountById(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return values.ReadValues{}, errLib.New("Discount not found", http.StatusNotFound)
		}
		log.Printf("Failed to get discount: %v", err)
		return values.ReadValues{}, errLib.New("Internal server error", http.StatusInternalServerError)
	}
	return mapDbDiscount(d), nil
}

func (r *Repository) List(ctx context.Context) ([]values.ReadValues, *errLib.CommonError) {
	discounts, err := r.Queries.ListDiscounts(ctx)
	if err != nil {
		log.Printf("Failed to list discounts: %v", err)
		return nil, errLib.New("Internal server error", http.StatusInternalServerError)
	}
	res := make([]values.ReadValues, len(discounts))
	for i, d := range discounts {
		res[i] = mapDbDiscount(d)
	}
	return res, nil
}

func (r *Repository) Update(ctx context.Context, details values.UpdateValues) (values.ReadValues, *errLib.CommonError) {
	params := db.UpdateDiscountParams{
		ID:              details.ID,
		Name:            details.Name,
		Description:     sql.NullString{String: details.Description, Valid: details.Description != ""},
		DiscountPercent: int32(details.DiscountPercent),
		IsUseUnlimited:  details.IsUseUnlimited,
		UsePerClient:    sql.NullInt32{Int32: int32(details.UsePerClient), Valid: details.UsePerClient > 0},
		IsActive:        details.IsActive,
		ValidFrom:       details.ValidFrom,
		ValidTo:         sql.NullTime{Time: details.ValidTo, Valid: !details.ValidTo.IsZero()},
	}
	d, err := r.Queries.UpdateDiscount(ctx, params)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			return values.ReadValues{}, errLib.New("Discount already exists", http.StatusConflict)
		}
		log.Printf("Failed to update discount: %v", err)
		return values.ReadValues{}, errLib.New("Internal server error", http.StatusInternalServerError)
	}
	return mapDbDiscount(d), nil
}

func (r *Repository) Delete(ctx context.Context, id uuid.UUID) *errLib.CommonError {
	rows, err := r.Queries.DeleteDiscount(ctx, id)
	if err != nil {
		log.Printf("Failed to delete discount: %v", err)
		return errLib.New("Internal server error", http.StatusInternalServerError)
	}
	if rows == 0 {
		return errLib.New("Discount not found", http.StatusNotFound)
	}
	return nil
}

func (r *Repository) GetByNameActive(ctx context.Context, name string) (values.ReadValues, *errLib.CommonError) {
	d, err := r.Queries.GetDiscountByNameActive(ctx, name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return values.ReadValues{}, errLib.New("Discount not found", http.StatusNotFound)
		}
		return values.ReadValues{}, errLib.New("Internal server error", http.StatusInternalServerError)
	}
	return mapDbDiscount(d), nil
}

func (r *Repository) GetUsageCount(ctx context.Context, customerID, discountID uuid.UUID) (int32, *errLib.CommonError) {
	count, err := r.Queries.GetUsageCount(ctx, db.GetUsageCountParams{CustomerID: customerID, DiscountID: discountID})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, nil
		}
		log.Printf("Failed to get usage count: %v", err)
		return 0, errLib.New("Internal server error", http.StatusInternalServerError)
	}
	return count, nil
}

func (r *Repository) IncrementUsage(ctx context.Context, customerID, discountID uuid.UUID) *errLib.CommonError {
	_, err := r.Queries.IncrementUsage(ctx, db.IncrementUsageParams{CustomerID: customerID, DiscountID: discountID})
	if err != nil {
		log.Printf("Failed to increment usage: %v", err)
		return errLib.New("Internal server error", http.StatusInternalServerError)
	}
	return nil
}

func (r *Repository) GetRestrictedPlans(ctx context.Context, discountID uuid.UUID) ([]uuid.UUID, *errLib.CommonError) {
	plans, err := r.Queries.GetRestrictedPlans(ctx, discountID)
	if err != nil {
		log.Printf("Failed to get restricted plans: %v", err)
		return nil, errLib.New("Internal server error", http.StatusInternalServerError)
	}
	return plans, nil
}
