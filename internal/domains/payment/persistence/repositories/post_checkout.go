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
	"time"
)

type PostCheckoutRepository struct {
	Queries *db.Queries
}

func NewPostCheckoutRepository(container *di.Container) *PostCheckoutRepository {
	return &PostCheckoutRepository{
		Queries: container.Queries.PurchasesDb,
	}
}

func (r *PostCheckoutRepository) EnrollCustomerInProgramEvents(ctx context.Context, customerID, programID uuid.UUID) *errLib.CommonError {

	if err := r.Queries.EnrollCustomerInProgramEvents(ctx, db.EnrollCustomerInProgramEventsParams{
		CustomerID: customerID,
		ProgramID:  programID,
	}); err != nil {
		return errLib.New(fmt.Sprintf("error enrolling customer in program events: %v", err), http.StatusBadRequest)
	}
	return nil
}

func (r *PostCheckoutRepository) GetProgramIdByStripePriceId(ctx context.Context, priceID string) (uuid.UUID, *errLib.CommonError) {

	if programID, err := r.Queries.GetProgramIdByStripePriceId(ctx, priceID); err != nil {
		return uuid.Nil, errLib.New(fmt.Sprintf("error enrolling customer in program events: %v", err), http.StatusBadRequest)
	} else {
		return programID, nil
	}
}

func (r *PostCheckoutRepository) EnrollCustomerInMembershipPlan(ctx context.Context, customerID, planID uuid.UUID, cancelAtDateTime time.Time) *errLib.CommonError {

	if err := r.Queries.EnrollCustomerInMembershipPlan(ctx, db.EnrollCustomerInMembershipPlanParams{
		CustomerID:       customerID,
		MembershipPlanID: planID,
		Status:           db.MembershipMembershipStatusActive,
		StartDate:        time.Now(),
		RenewalDate: sql.NullTime{
			Time:  cancelAtDateTime,
			Valid: !cancelAtDateTime.IsZero(),
		},
	}); err != nil {
		return errLib.New(fmt.Sprintf("error enrolling customer in membership plan: %v", err), http.StatusBadRequest)
	}
	return nil
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
