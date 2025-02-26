package customer_membership

import (
	"api/internal/di"
	db "api/internal/domains/purchase/persistence/sqlc/generated"
	"api/internal/domains/purchase/values"
	errLib "api/internal/libs/errors"
	"context"
	"database/sql"
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

func (r *Repository) Purchase(c context.Context, details *values.MembershipPlanPurchaseInfo) *errLib.CommonError {

	dbParams := db.PurchaseMembershipParams{
		CustomerID:       details.CustomerId,
		MembershipPlanID: details.MembershipPlanId,
		Status:           db.MembershipStatus(details.Status),
		StartDate:        sql.NullTime{Time: details.StartDate, Valid: true},
	}

	if details.RenewalDate != nil {
		dbParams.RenewalDate = sql.NullTime{
			Time:  *details.RenewalDate,
			Valid: true,
		}
	}

	row, err := r.Queries.PurchaseMembership(c, dbParams)

	if err != nil {
		return errLib.New("Internal server error", http.StatusInternalServerError)
	}

	if row == 0 {
		return errLib.New("Purchase failed", http.StatusInternalServerError)
	}

	return nil
}
