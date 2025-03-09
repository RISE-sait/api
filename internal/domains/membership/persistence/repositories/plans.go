package membership

import (
	"api/internal/di"
	db "api/internal/domains/membership/persistence/sqlc/generated"
	values "api/internal/domains/membership/values"
	errLib "api/internal/libs/errors"
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"

	"github.com/google/uuid"
)

type PlansRepository struct {
	Queries *db.Queries
}

func NewMembershipPlansRepository(container *di.Container) *PlansRepository {
	return &PlansRepository{
		Queries: container.Queries.MembershipDb,
	}
}

func (r *PlansRepository) CreateMembershipPlan(c context.Context, membershipPlan *values.PlanCreateValues) *errLib.CommonError {

	var periods int32

	if membershipPlan.AmtPeriods != nil {
		periods = int32(*membershipPlan.AmtPeriods)
	}

	dbParams := db.CreateMembershipPlanParams{
		Name:  membershipPlan.Name,
		Price: int32(membershipPlan.Price),
		PaymentFrequency: db.NullPaymentFrequency{
			PaymentFrequency: db.PaymentFrequency(*membershipPlan.PaymentFrequency),
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

func (r *PlansRepository) GetMembershipPlanById(ctx context.Context, id uuid.UUID) (*values.PlanReadValues, *errLib.CommonError) {
	dbPlan, err := r.Queries.GetMembershipPlanById(ctx, id)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errLib.New("Plan not found", http.StatusNotFound)
		}
		return nil, errLib.New("Internal server error", http.StatusInternalServerError)
	}

	plan := values.PlanReadValues{
		ID:           dbPlan.ID,
		Name:         dbPlan.Name,
		MembershipID: dbPlan.MembershipID,
		Price:        int64(dbPlan.Price),
	}

	if dbPlan.AmtPeriods.Valid {
		amtPeriods := int(dbPlan.AmtPeriods.Int32)
		plan.AmtPeriods = &amtPeriods
	}

	if dbPlan.PaymentFrequency.Valid {
		freq := string(dbPlan.PaymentFrequency.PaymentFrequency)
		plan.PaymentFrequency = &freq
	}

	return &plan, nil
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
			ID:           dbPlan.ID,
			Name:         dbPlan.Name,
			MembershipID: dbPlan.MembershipID,
			Price:        int64(dbPlan.Price),
		}

		if dbPlan.AmtPeriods.Valid {
			amtPeriods := int(dbPlan.AmtPeriods.Int32)
			plan.AmtPeriods = &amtPeriods
		}

		if dbPlan.PaymentFrequency.Valid {
			freq := string(dbPlan.PaymentFrequency.PaymentFrequency)
			plan.PaymentFrequency = &freq
		}

		plans[i] = plan
	}

	return plans, nil
}

func (r *PlansRepository) UpdateMembershipPlan(c context.Context, plan *values.PlanUpdateValues) *errLib.CommonError {

	var periods int32

	if plan.AmtPeriods != nil {
		periods = int32(*plan.AmtPeriods)
	}

	dbMembershipParams := db.UpdateMembershipPlanParams{
		Name:  plan.Name,
		Price: int32(plan.Price),
		PaymentFrequency: db.NullPaymentFrequency{
			PaymentFrequency: db.PaymentFrequency(*plan.PaymentFrequency),
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
