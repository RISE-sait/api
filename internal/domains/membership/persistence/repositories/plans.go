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
	"log"
	"net/http"
)

type PlansRepository struct {
	Queries *db.Queries
}

func NewMembershipPlansRepository(container *di.Container) *PlansRepository {
	return &PlansRepository{
		Queries: container.Queries.MembershipDb,
	}
}

func (r *PlansRepository) CreateMembershipPlan(c context.Context, membershipPlan values.PlanCreateValues) *errLib.CommonError {

	dbParams := db.CreateMembershipPlanParams{
		MembershipID:     membershipPlan.MembershipID,
		Name:             membershipPlan.Name,
		Price:            membershipPlan.Price,
		PaymentFrequency: db.PaymentFrequency(membershipPlan.PaymentFrequency),
		AmtPeriods:       membershipPlan.AmtPeriods,
		AutoRenew:        membershipPlan.IsAutoRenew,
		JoiningFee:       membershipPlan.JoiningFees,
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
			MembershipID:     dbPlan.MembershipID,
			Name:             dbPlan.Name,
			Price:            dbPlan.Price,
			PaymentFrequency: string(dbPlan.PaymentFrequency),
			AmtPeriods:       dbPlan.AmtPeriods,
			IsAutoRenew:      dbPlan.AutoRenew,
			JoiningFees:      dbPlan.JoiningFee,
		},
		CreatedAt: dbPlan.CreatedAt,
		UpdatedAt: dbPlan.UpdatedAt,
	}

	return plan, nil
}

func (r *PlansRepository) GetMembershipPlanPaymentFrequencies() []string {
	dbFreqs := db.AllPaymentFrequencyValues()

	var freq []string

	for _, dbFreq := range dbFreqs {
		freq = append(freq, string(dbFreq))
	}

	return freq
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
				MembershipID:     dbPlan.MembershipID,
				Name:             dbPlan.Name,
				Price:            dbPlan.Price,
				PaymentFrequency: string(dbPlan.PaymentFrequency),
				AmtPeriods:       dbPlan.AmtPeriods,
				JoiningFees:      dbPlan.JoiningFee,
				IsAutoRenew:      dbPlan.AutoRenew,
			},
			CreatedAt: dbPlan.CreatedAt,
			UpdatedAt: dbPlan.UpdatedAt,
		}

		plans[i] = plan
	}

	return plans, nil
}

func (r *PlansRepository) UpdateMembershipPlan(c context.Context, plan values.PlanUpdateValues) *errLib.CommonError {

	dbMembershipParams := db.UpdateMembershipPlanParams{
		Name:             plan.Name,
		Price:            plan.Price,
		PaymentFrequency: db.PaymentFrequency(plan.PaymentFrequency),
		AmtPeriods:       plan.AmtPeriods,
		MembershipID:     plan.MembershipID,
		ID:               plan.ID,
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
