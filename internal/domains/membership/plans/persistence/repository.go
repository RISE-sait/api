package persistence

import (
	"api/cmd/server/di"
	entity "api/internal/domains/membership/plans/entities"
	db "api/internal/domains/membership/plans/persistence/sqlc/generated"
	"api/internal/domains/membership/plans/values"
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
		Queries: container.Queries.MembershipPlanDb,
	}
}

func (r *MembershipPlansRepository) CreateMembershipPlan(c context.Context, membershipPlan *values.MembershipPlanCreate) *errLib.CommonError {

	dbParams := db.CreateMembershipPlanParams{
		Name:  membershipPlan.Name,
		Price: membershipPlan.Price,
		PaymentFrequency: db.NullPaymentFrequency{
			PaymentFrequency: db.PaymentFrequency(membershipPlan.PaymentFrequency),
			Valid:            true,
		},
		AmtPeriods: sql.NullInt32{
			Int32: int32(membershipPlan.AmtPeriods),
			Valid: true,
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

func (r *MembershipPlansRepository) GetMembershipPlansByMembershipId(ctx context.Context, id uuid.UUID) ([]entity.MembershipPlan, *errLib.CommonError) {
	dbPlans, err := r.Queries.GetMembershipPlansByMembershipId(ctx, id)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errLib.New("No membership plans found", http.StatusNotFound)
		}
		return nil, errLib.New("Internal server error", http.StatusInternalServerError)
	}

	plans := make([]entity.MembershipPlan, len(dbPlans))
	for i, dbPlan := range dbPlans {
		plans[i] = entity.MembershipPlan{
			ID:               dbPlan.ID,
			Name:             dbPlan.Name,
			MembershipID:     dbPlan.MembershipID,
			Price:            dbPlan.Price,
			PaymentFrequency: string(dbPlan.PaymentFrequency.PaymentFrequency),
			AmtPeriods:       int(dbPlan.AmtPeriods.Int32),
		}
	}

	return plans, nil
}

func (r *MembershipPlansRepository) UpdateMembershipPlan(c context.Context, plan *values.MembershipPlanUpdate) *errLib.CommonError {

	dbMembershipParams := db.UpdateMembershipPlanParams{
		Name:  plan.Name,
		Price: plan.Price,
		PaymentFrequency: db.NullPaymentFrequency{
			PaymentFrequency: db.PaymentFrequency(plan.PaymentFrequency),
			Valid:            true,
		},
		AmtPeriods: sql.NullInt32{
			Int32: int32(plan.AmtPeriods),
			Valid: true,
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
