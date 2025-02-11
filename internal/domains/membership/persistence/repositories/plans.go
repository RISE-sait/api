package membership

import (
	"api/internal/di"
	db "api/internal/domains/membership/persistence/sqlc/generated"
	values "api/internal/domains/membership/values/plans"
	errLib "api/internal/libs/errors"
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"

	"github.com/google/uuid"
)

type MembershipPlansRepository struct {
	Queries *db.Queries
}

func NewMembershipPlansRepository(container *di.Container) *MembershipPlansRepository {
	return &MembershipPlansRepository{
		Queries: container.Queries.MembershipDb,
	}
}

func (r *MembershipPlansRepository) CreateMembershipPlan(c context.Context, membershipPlan *values.MembershipPlanDetails) *errLib.CommonError {

	var periods int32

	if membershipPlan.AmtPeriods != nil {
		periods = int32(*membershipPlan.AmtPeriods)
	}

	dbParams := db.CreateMembershipPlanParams{
		Name:  membershipPlan.Name,
		Price: int32(membershipPlan.Price),
		PaymentFrequency: db.NullPaymentFrequency{
			PaymentFrequency: db.PaymentFrequency(membershipPlan.PaymentFrequency),
			Valid:            true,
		},
		AmtPeriods: sql.NullInt32{
			Int32: periods,
			Valid: membershipPlan.AmtPeriods != nil,
		},
		MembershipID: membershipPlan.MembershipID,
	}
	row, err := r.Queries.CreateMembershipPlan(c, dbParams)

	if err != nil {
		log.Printf("Failed to create plan: %+v. Error: %v", membershipPlan, err.Error())
		return errLib.New("Internal server error", http.StatusInternalServerError)
	}

	if row == 0 {
		return errLib.New("Membership not found", http.StatusNotFound)
	}

	return nil
}

func (r *MembershipPlansRepository) GetMembershipPlansByMembershipId(ctx context.Context, id uuid.UUID) ([]values.MembershipPlanAllFields, *errLib.CommonError) {
	dbPlans, err := r.Queries.GetMembershipPlansByMembershipId(ctx, id)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errLib.New("No membership plans found", http.StatusNotFound)
		}
		return nil, errLib.New("Internal server error", http.StatusInternalServerError)
	}

	plans := make([]values.MembershipPlanAllFields, len(dbPlans))

	for i, dbPlan := range dbPlans {

		var periods *int

		if dbPlan.AmtPeriods.Valid {
			amtPeriods := int(dbPlan.AmtPeriods.Int32)
			periods = &amtPeriods
		}

		plans[i] = values.MembershipPlanAllFields{
			ID: dbPlan.ID,
			MembershipPlanDetails: values.MembershipPlanDetails{
				Name:             dbPlan.Name,
				MembershipID:     dbPlan.MembershipID,
				Price:            int64(dbPlan.Price),
				PaymentFrequency: string(dbPlan.PaymentFrequency.PaymentFrequency),
				AmtPeriods:       periods,
			},
		}
	}

	return plans, nil
}

func (r *MembershipPlansRepository) UpdateMembershipPlan(c context.Context, plan *values.MembershipPlanAllFields) *errLib.CommonError {

	var periods int32

	if plan.AmtPeriods != nil {
		periods = int32(*plan.AmtPeriods)
	}

	dbMembershipParams := db.UpdateMembershipPlanParams{
		Name:  plan.Name,
		Price: int32(plan.Price),
		PaymentFrequency: db.NullPaymentFrequency{
			PaymentFrequency: db.PaymentFrequency(plan.PaymentFrequency),
			Valid:            true,
		},
		AmtPeriods: sql.NullInt32{
			Int32: periods,
			Valid: plan.AmtPeriods != nil,
		},
		MembershipID: plan.MembershipID,
		ID:           plan.ID,
	}

	row, err := r.Queries.UpdateMembershipPlan(c, dbMembershipParams)

	if err != nil {
		log.Printf("Failed to update plan: %+v. Error: %v", plan, err.Error())
		return errLib.New("Internal server error", http.StatusInternalServerError)
	}

	if row == 0 {
		return errLib.New("Membership not found", http.StatusNotFound)
	}
	return nil
}

func (r *MembershipPlansRepository) DeleteMembershipPlan(c context.Context, membershipId, planId uuid.UUID) *errLib.CommonError {

	plan := db.DeleteMembershipPlanParams{
		MembershipID: membershipId,
		ID:           planId,
	}

	row, err := r.Queries.DeleteMembershipPlan(c, plan)

	if err != nil {
		log.Printf("Failed to delete plan with membership ID: %s and plan ID: %s. Error: %v", membershipId, planId, err.Error())
		return errLib.New("Internal server error", http.StatusInternalServerError)
	}

	if row == 0 {
		return errLib.New("Membership plan not found", http.StatusNotFound)
	}

	return nil
}
