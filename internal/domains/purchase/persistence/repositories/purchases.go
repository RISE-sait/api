package customer_membership

import (
	"api/internal/di"
	db "api/internal/domains/purchase/persistence/sqlc/generated"
	"api/internal/domains/purchase/values"
	errLib "api/internal/libs/errors"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"net/http"
)

type Repository struct {
	Queries *db.Queries
}

func NewRepository(container *di.Container) *Repository {
	return &Repository{
		Queries: container.Queries.PurchasesDb,
	}
}

func (r *Repository) GetJoiningFees(ctx context.Context, planID uuid.UUID) (decimal.Decimal, *errLib.CommonError) {

	if joiningFees, err := r.Queries.GetMembershipPlanJoiningFee(ctx, planID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return decimal.NewFromInt(0), nil
		}
		return decimal.Decimal{}, errLib.New(fmt.Sprintf("error getting joining fee for membership plan: %v", joiningFees), http.StatusBadRequest)
	} else {
		return joiningFees, nil
	}
}

func (r *Repository) Checkout(ctx context.Context, details values.MembershipPlanPurchaseInfo) *errLib.CommonError {

	if joiningFees, err := r.Queries.GetMembershipPlanJoiningFee(ctx, details.CustomerId); err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return errLib.New(fmt.Sprintf("error getting joining fee for membership plan: %v", joiningFees), http.StatusBadRequest)
		}
	}

	//dbParams := db.CreateCustomerMembershipPlanParams{
	//	CustomerID: details.CustomerId,
	//	MembershipPlanID: uuid.NullUUID{
	//		UUID:  details.MembershipPlanId,
	//		Valid: true,
	//	},
	//	Status:    db.MembershipStatus(details.Status),
	//	StartDate: details.StartDate,
	//}
	//
	//if details.RenewalDate != nil {
	//	dbParams.RenewalDate = sql.NullTime{
	//		Time:  *details.RenewalDate,
	//		Valid: true,
	//	}
	//}
	//
	//if err := r.Queries.CreateCustomerMembershipPlan(c, dbParams); err != nil {
	//	log.Println("error creating customer-membership-plan", err)
	//	return errLib.New("Internal server error", http.StatusInternalServerError)
	//}

	return nil
}
