package repositories

import (
	"api/internal/utils"
	db "api/sqlc"
	"context"
	"log"
	"net/http"

	"github.com/google/uuid"
)

type MembershipPlansRepository struct {
	Queries *db.Queries
}

func (r *MembershipPlansRepository) CreateMembershipPlan(c context.Context, membershipPlan *db.CreateMembershipPlanParams) *utils.HTTPError {
	row, err := r.Queries.CreateMembershipPlan(c, *membershipPlan)

	if err != nil {
		return utils.MapDatabaseError(err)
	}

	if row == 0 {
		log.Printf("Failed to create membership plan: %+v", *membershipPlan)
		return utils.CreateHTTPError("Internal Server Error", http.StatusInternalServerError)
	}

	return nil
}

func (r *MembershipPlansRepository) GetMembershipById(c context.Context, id uuid.UUID) (*db.Membership, *utils.HTTPError) {
	membership, err := r.Queries.GetMembershipById(c, id)

	if err != nil {
		return nil, utils.MapDatabaseError(err)
	}

	return &membership, nil
}

func (r *MembershipPlansRepository) GetMembershipPlanDetails(c context.Context, membershipId uuid.UUID, planId uuid.UUID) (*[]db.MembershipPlan, *utils.HTTPError) {

	plan := db.GetMembershipPlansParams{
		Column1: membershipId,
		Column2: planId,
	}

	memberships, err := r.Queries.GetMembershipPlans(c, plan)

	if err != nil {
		return nil, utils.MapDatabaseError(err)
	}

	return &memberships, nil
}

func (r *MembershipPlansRepository) UpdateMembershipPlan(c context.Context, plan *db.UpdateMembershipPlanParams) *utils.HTTPError {

	if isExist, _ := r.Queries.IsMembershipIDExist(c, plan.MembershipID); !isExist {
		return utils.CreateHTTPError("Membership ID not found", http.StatusNotFound)
	}

	row, err := r.Queries.UpdateMembershipPlan(c, *plan)

	if err != nil {
		return utils.MapDatabaseError(err)
	}

	if row == 0 {
		log.Printf("Failed to update membership plan: %+v", *plan)
		return utils.CreateHTTPError("Membership plan not found", http.StatusNotFound)
	}

	return nil
}

func (r *MembershipPlansRepository) DeleteMembershipPlan(c context.Context, plan *db.DeleteMembershipPlanParams) *utils.HTTPError {
	row, err := r.Queries.DeleteMembershipPlan(c, *plan)

	if err != nil {
		return utils.MapDatabaseError(err)
	}

	if row == 0 {
		log.Printf("Failed to delete membership plan: %+v", *plan)
		return utils.CreateHTTPError("Membership plan not found", http.StatusNotFound)
	}

	return nil
}
