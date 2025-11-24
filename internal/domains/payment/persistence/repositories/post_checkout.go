package payment

import (
	"api/internal/di"
	db "api/internal/domains/payment/persistence/sqlc/generated"
	errLib "api/internal/libs/errors"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"net/http"
)

type PostCheckoutRepository struct {
	Queries *db.Queries
}

func NewPostCheckoutRepository(container *di.Container) *PostCheckoutRepository {
	return &PostCheckoutRepository{
		Queries: container.Queries.PurchasesDb,
	}
}

func (r *PostCheckoutRepository) GetProgramIdByStripePriceId(ctx context.Context, priceID string) (uuid.UUID, *errLib.CommonError) {

	programID, err := r.Queries.GetProgramIdByStripePriceId(ctx, priceID)

	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return uuid.Nil, errLib.New(fmt.Sprintf("error getting program id of price id: %v", err), http.StatusBadRequest)
		}
	}
	return programID, nil
}

func (r *PostCheckoutRepository) GetEventIdByStripePriceId(ctx context.Context, priceID string) (uuid.UUID, *errLib.CommonError) {

	eventID, err := r.Queries.GetEventIdByStripePriceId(ctx, priceID)

	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {

			return uuid.Nil, errLib.New(fmt.Sprintf("error getting event id of price id: %v", err), http.StatusBadRequest)
		}
	}
	return eventID, nil
}

func (r *PostCheckoutRepository) GetMembershipPlanByStripePriceID(ctx context.Context, id string) (planID uuid.UUID, amtPeriods *int32, error *errLib.CommonError) {
	membershipPlan, err := r.Queries.GetMembershipPlanByStripePriceId(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return uuid.Nil, nil, errLib.New("membership plan not found", http.StatusNotFound)
		}
		return uuid.Nil, nil, errLib.New(fmt.Sprintf("error getting membership plan by stripe price ID: %v", err), http.StatusBadRequest)
	}

	if membershipPlan.ID == uuid.Nil {
		return uuid.Nil, nil, errLib.New("membership plan not found", http.StatusNotFound)
	}

	if !membershipPlan.AmtPeriods.Valid {
		return membershipPlan.ID, nil, nil
	}

	periods := membershipPlan.AmtPeriods.Int32

	return membershipPlan.ID, &periods, nil
}

func (r *PostCheckoutRepository) GetMembershipPlanAmtPeriods(ctx context.Context, planID uuid.UUID) (*int32, *errLib.CommonError) {
	amtPeriods, err := r.Queries.GetMembershipPlanAmtPeriods(ctx, planID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errLib.New("membership plan not found", http.StatusNotFound)
		}
		return nil, errLib.New(fmt.Sprintf("error getting membership plan amt_periods: %v", err), http.StatusBadRequest)
	}

	if !amtPeriods.Valid {
		return nil, nil
	}

	periods := amtPeriods.Int32
	return &periods, nil
}
