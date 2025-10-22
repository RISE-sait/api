package discount

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"

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
	var discountAmount *float64
	if d.DiscountAmount.Valid {
		// Parse string to float64
		if amount, err := strconv.ParseFloat(d.DiscountAmount.String, 64); err == nil {
			discountAmount = &amount
		}
	}

	var durationMonths *int
	if d.DurationMonths.Valid {
		months := int(d.DurationMonths.Int32)
		durationMonths = &months
	}

	var maxRedemptions *int
	if d.MaxRedemptions.Valid {
		max := int(d.MaxRedemptions.Int32)
		maxRedemptions = &max
	}

	var stripeCouponID *string
	if d.StripeCouponID.Valid {
		stripeCouponID = &d.StripeCouponID.String
	}

	val := values.ReadValues{
		ID: d.ID,
		CreateValues: values.CreateValues{
			Name:            d.Name,
			Description:     d.Description.String,
			DiscountPercent: int(d.DiscountPercent),
			DiscountAmount:  discountAmount,
			DiscountType:    values.DiscountType(d.DiscountType),
			IsUseUnlimited:  d.IsUseUnlimited,
			UsePerClient:    int(d.UsePerClient.Int32),
			IsActive:        d.IsActive,
			ValidFrom:       d.ValidFrom,
			ValidTo:         d.ValidTo.Time,
			DurationType:    values.DurationType(d.DurationType),
			DurationMonths:  durationMonths,
			AppliesTo:       values.AppliesTo(d.AppliesTo),
			MaxRedemptions:  maxRedemptions,
			StripeCouponID:  stripeCouponID,
		},
		TimesRedeemed: int(d.TimesRedeemed),
		CreatedAt:     d.CreatedAt,
		UpdatedAt:     d.UpdatedAt,
	}
	return val
}

func (r *Repository) Create(ctx context.Context, details values.CreateValues) (values.ReadValues, *errLib.CommonError) {
	var discountAmount sql.NullString
	if details.DiscountAmount != nil {
		// Store as string to maintain precision
		discountAmount = sql.NullString{String: fmt.Sprintf("%.2f", *details.DiscountAmount), Valid: true}
	}

	var durationMonths sql.NullInt32
	if details.DurationMonths != nil {
		durationMonths = sql.NullInt32{Int32: int32(*details.DurationMonths), Valid: true}
	}

	var maxRedemptions sql.NullInt32
	if details.MaxRedemptions != nil {
		maxRedemptions = sql.NullInt32{Int32: int32(*details.MaxRedemptions), Valid: true}
	}

	var stripeCouponID sql.NullString
	if details.StripeCouponID != nil {
		stripeCouponID = sql.NullString{String: *details.StripeCouponID, Valid: true}
	}

	params := db.CreateDiscountParams{
		Name:            details.Name,
		Description:     sql.NullString{String: details.Description, Valid: details.Description != ""},
		DiscountPercent: int32(details.DiscountPercent),
		DiscountAmount:  discountAmount,
		DiscountType:    db.DiscountType(details.DiscountType),
		IsUseUnlimited:  details.IsUseUnlimited,
		UsePerClient:    sql.NullInt32{Int32: int32(details.UsePerClient), Valid: details.UsePerClient > 0},
		IsActive:        details.IsActive,
		ValidFrom:       details.ValidFrom,
		ValidTo:         sql.NullTime{Time: details.ValidTo, Valid: !details.ValidTo.IsZero()},
		DurationType:    db.DiscountDurationType(details.DurationType),
		DurationMonths:  durationMonths,
		AppliesTo:       db.DiscountAppliesTo(details.AppliesTo),
		MaxRedemptions:  maxRedemptions,
		StripeCouponID:  stripeCouponID,
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
	var discountAmount sql.NullString
	if details.DiscountAmount != nil {
		// Store as string to maintain precision
		discountAmount = sql.NullString{String: fmt.Sprintf("%.2f", *details.DiscountAmount), Valid: true}
	}

	var durationMonths sql.NullInt32
	if details.DurationMonths != nil {
		durationMonths = sql.NullInt32{Int32: int32(*details.DurationMonths), Valid: true}
	}

	var maxRedemptions sql.NullInt32
	if details.MaxRedemptions != nil {
		maxRedemptions = sql.NullInt32{Int32: int32(*details.MaxRedemptions), Valid: true}
	}

	var stripeCouponID sql.NullString
	if details.StripeCouponID != nil {
		stripeCouponID = sql.NullString{String: *details.StripeCouponID, Valid: true}
	}

	params := db.UpdateDiscountParams{
		ID:              details.ID,
		Name:            details.Name,
		Description:     sql.NullString{String: details.Description, Valid: details.Description != ""},
		DiscountPercent: int32(details.DiscountPercent),
		DiscountAmount:  discountAmount,
		DiscountType:    db.DiscountType(details.DiscountType),
		IsUseUnlimited:  details.IsUseUnlimited,
		UsePerClient:    sql.NullInt32{Int32: int32(details.UsePerClient), Valid: details.UsePerClient > 0},
		IsActive:        details.IsActive,
		ValidFrom:       details.ValidFrom,
		ValidTo:         sql.NullTime{Time: details.ValidTo, Valid: !details.ValidTo.IsZero()},
		DurationType:    db.DiscountDurationType(details.DurationType),
		DurationMonths:  durationMonths,
		AppliesTo:       db.DiscountAppliesTo(details.AppliesTo),
		MaxRedemptions:  maxRedemptions,
		StripeCouponID:  stripeCouponID,
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

func (r *Repository) GetByName(ctx context.Context, name string) (values.ReadValues, *errLib.CommonError) {
	d, err := r.Queries.GetDiscountByName(ctx, name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return values.ReadValues{}, errLib.New("Discount not found", http.StatusNotFound)
		}
		return values.ReadValues{}, errLib.New("Internal server error", http.StatusInternalServerError)
	}
	return mapDbDiscount(d), nil
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
