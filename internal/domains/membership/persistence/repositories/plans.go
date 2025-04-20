package membership

import (
	"api/internal/di"
	db "api/internal/domains/membership/persistence/sqlc/generated"
	values "api/internal/domains/membership/values"
	errLib "api/internal/libs/errors"
	"context"
	"database/sql"
	"errors"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"log"
	"net/http"
)

type PlansRepository struct {
	Queries *db.Queries
	Tx      *sql.Tx
}

func NewMembershipPlansRepository(container *di.Container) *PlansRepository {
	return &PlansRepository{
		Queries: container.Queries.MembershipDb,
	}
}

func (r *PlansRepository) GetTx() *sql.Tx {
	return r.Tx
}

func (r *PlansRepository) WithTx(tx *sql.Tx) *PlansRepository {
	return &PlansRepository{
		Queries: r.Queries.WithTx(tx),
		Tx:      tx,
	}
}

func (r *PlansRepository) CreateMembershipPlan(c context.Context, membershipPlan values.PlanCreateValues) *errLib.CommonError {

	dbParams := db.CreateMembershipPlanParams{
		MembershipID: membershipPlan.MembershipID,
		Name:         membershipPlan.Name,
		StripeJoiningFeeID: sql.NullString{
			String: membershipPlan.StripeJoiningFeesID,
			Valid:  membershipPlan.StripeJoiningFeesID != "",
		},
		StripePriceID: membershipPlan.StripePriceID,
	}

	if membershipPlan.AmtPeriods != nil {
		dbParams.AmtPeriods = sql.NullInt32{
			Int32: *membershipPlan.AmtPeriods,
			Valid: true,
		}
	}

	_, err := r.Queries.CreateMembershipPlan(c, dbParams)

	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			switch pqErr.Constraint {
			case "fk_membership":
				return errLib.New("Membership not found", http.StatusNotFound)
			case "unique_membership_name":
				return errLib.New("Membership plan with this name already exists", http.StatusConflict)
			case "unique_stripe_price_id":
				return errLib.New("Membership plan with this stripe price ID already exists", http.StatusConflict)
			}
		}
		log.Printf("Failed to create plan: %+v. Error: %v", membershipPlan, err.Error())
		return errLib.New("Internal server error", http.StatusInternalServerError)
	}

	return nil
}

func (r *PlansRepository) GetMembershipPlanById(ctx context.Context, id uuid.UUID) (values.PlanReadValues, *errLib.CommonError) {
	dbPlan, err := r.Queries.GetMembershipPlanById(ctx, id)

	var response values.PlanReadValues

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return response, errLib.New("Plan not found", http.StatusNotFound)
		}
		return response, errLib.New("Internal server error", http.StatusInternalServerError)
	}

	plan := values.PlanReadValues{
		ID: dbPlan.ID,
		PlanDetails: values.PlanDetails{
			MembershipID:        dbPlan.MembershipID,
			Name:                dbPlan.Name,
			StripeJoiningFeesID: dbPlan.StripeJoiningFeeID.String,
			StripePriceID:       dbPlan.StripePriceID,
		},
		CreatedAt: dbPlan.CreatedAt,
		UpdatedAt: dbPlan.UpdatedAt,
	}

	if dbPlan.AmtPeriods.Valid {
		plan.AmtPeriods = &dbPlan.AmtPeriods.Int32
	}

	return plan, nil
}

func (r *PlansRepository) GetMembershipPlans(ctx context.Context, membershipId uuid.UUID) ([]values.PlanReadValues, *errLib.CommonError) {

	dbPlans, err := r.Queries.GetMembershipPlans(ctx, membershipId)

	if err != nil {
		log.Printf("Failed to get plans: %+v. Error: %v", membershipId, err.Error())
		return nil, errLib.New("Internal server error", http.StatusInternalServerError)
	}

	plans := make([]values.PlanReadValues, len(dbPlans))

	for i, dbPlan := range dbPlans {

		plan := values.PlanReadValues{
			ID: dbPlan.ID,
			PlanDetails: values.PlanDetails{
				MembershipID:        dbPlan.MembershipID,
				Name:                dbPlan.Name,
				StripeJoiningFeesID: dbPlan.StripeJoiningFeeID.String,
				StripePriceID:       dbPlan.StripePriceID,
			},
			CreatedAt: dbPlan.CreatedAt,
			UpdatedAt: dbPlan.UpdatedAt,
		}

		if dbPlan.AmtPeriods.Valid {
			plan.AmtPeriods = &dbPlan.AmtPeriods.Int32
		}

		plans[i] = plan
	}

	return plans, nil
}

func (r *PlansRepository) UpdateMembershipPlan(c context.Context, plan values.PlanUpdateValues) *errLib.CommonError {

	dbMembershipParams := db.UpdateMembershipPlanParams{
		ID:           plan.ID,
		MembershipID: plan.MembershipID,
		Name:         plan.Name,
		StripeJoiningFeeID: sql.NullString{
			String: plan.StripeJoiningFeesID,
			Valid:  plan.StripeJoiningFeesID != "",
		},
		StripePriceID: plan.StripePriceID,
	}

	if plan.AmtPeriods != nil {
		dbMembershipParams.AmtPeriods = sql.NullInt32{
			Int32: *plan.AmtPeriods,
			Valid: true,
		}
	}

	_, err := r.Queries.UpdateMembershipPlan(c, dbMembershipParams)

	if err != nil {
		var pqErr *pq.Error

		if errors.As(err, &pqErr) {
			switch pqErr.Constraint {
			case "fk_membership":
				return errLib.New("Membership not found", http.StatusNotFound)
			}
		}
		log.Printf("Failed to update plan: %+v. Error: %v", plan, err.Error())
		return errLib.New("Internal server error", http.StatusInternalServerError)
	}

	return nil
}

func (r *PlansRepository) DeleteMembershipPlan(c context.Context, id uuid.UUID) *errLib.CommonError {

	row, err := r.Queries.DeleteMembershipPlan(c, id)

	if err != nil {
		log.Printf("Failed to delete plan with Id: %s. Error: %v", id, err.Error())
		return errLib.New("Internal server error", http.StatusInternalServerError)
	}

	if row == 0 {
		return errLib.New("Membership plan not found", http.StatusNotFound)
	}

	return nil
}
