package membership_plan

import (
	"api/internal/domains/membership/plans/dto"
	db "api/internal/domains/membership/plans/infra/sqlc/generated"
	errLib "api/internal/libs/errors"
	"context"
	"net/http"

	"github.com/google/uuid"
)

type Repo struct {
	Queries *db.Queries
}

func (r *Repo) CreateMembershipPlan(c context.Context, membershipPlan *dto.CreateMembershipPlanRequest) *errLib.CommonError {

	dbParams := membershipPlan.ToDBParams()

	row, err := r.Queries.CreateMembershipPlan(c, *dbParams)

	if err != nil {
		return errLib.TranslateDBErrorToCommonError(err)
	}

	if row == 0 {
		return errLib.New("Membership plan not created", http.StatusInternalServerError)
	}

	return nil
}

func (r *Repo) GetMembershipPlansByMembershipId(ctx context.Context, id uuid.UUID) ([]db.MembershipPlan, *errLib.CommonError) {
	plans, err := r.Queries.GetMembershipPlansByMembershipId(ctx, id)

	if err != nil {
		return nil, errLib.TranslateDBErrorToCommonError(err)
	}

	return plans, nil
}

func (r *Repo) UpdateMembershipPlan(c context.Context, plan *dto.UpdateMembershipPlanRequest) *errLib.CommonError {

	dbMembershipParams := plan.ToDBParams()

	row, err := r.Queries.UpdateMembershipPlan(c, *dbMembershipParams)

	if err != nil {
		return errLib.TranslateDBErrorToCommonError(err)
	}

	if row == 0 {
		return errLib.New("Membership plan not found", http.StatusNotFound)
	}
	return nil
}

func (r *Repo) DeleteMembershipPlan(c context.Context, plan *db.DeleteMembershipPlanParams) *errLib.CommonError {
	row, err := r.Queries.DeleteMembershipPlan(c, *plan)

	if err != nil {
		return errLib.TranslateDBErrorToCommonError(err)
	}

	if row == 0 {
		return errLib.New("Membership plan not found", http.StatusNotFound)
	}

	return nil
}
