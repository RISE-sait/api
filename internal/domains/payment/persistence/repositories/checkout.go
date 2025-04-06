package payment

import (
	"api/internal/di"
	db "api/internal/domains/payment/persistence/sqlc/generated"
	values "api/internal/domains/payment/values"
	errLib "api/internal/libs/errors"
	contextUtils "api/utils/context"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"log"
	"net/http"
)

type CheckoutRepository struct {
	Queries *db.Queries
}

func NewCheckoutRepository(container *di.Container) *CheckoutRepository {
	return &CheckoutRepository{
		Queries: container.Queries.PurchasesDb,
	}
}

func (r *CheckoutRepository) GetMembershipPlanJoiningRequirement(ctx context.Context, planID uuid.UUID) (values.MembershipPlanJoiningRequirement, *errLib.CommonError) {

	if requirements, err := r.Queries.GetMembershipPlanJoiningRequirements(ctx, planID); err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return values.MembershipPlanJoiningRequirement{}, errLib.New(fmt.Sprintf("error getting joining requirement for membership plan: %v", planID), http.StatusBadRequest)
		} else {
			return values.MembershipPlanJoiningRequirement{}, errLib.New(fmt.Sprintf("membership plan not found for plan id: %v", planID), http.StatusBadRequest)
		}
	} else {
		response := values.MembershipPlanJoiningRequirement{
			ID:            requirements.ID,
			Name:          requirements.Name,
			StripePriceID: requirements.StripePriceID,
			MembershipID:  requirements.MembershipID,
			CreatedAt:     requirements.CreatedAt,
			UpdatedAt:     requirements.UpdatedAt,
		}

		if requirements.StripeJoiningFeeID.Valid {
			response.StripeJoiningFeeID = requirements.StripeJoiningFeeID.String
		}

		if requirements.AmtPeriods.Valid {
			response.AmtPeriods = &requirements.AmtPeriods.Int32
		}

		return response, nil
	}
}

func (r *CheckoutRepository) GetProgramRegistrationPriceIdForCustomer(ctx context.Context, programID uuid.UUID) (string, *errLib.CommonError) {

	customerID, ctxErr := contextUtils.GetUserID(ctx)
	if ctxErr != nil {
		return "", ctxErr
	}

	program, err := r.Queries.GetProgram(ctx, programID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", errLib.New(fmt.Sprintf("program not found: %v", programID), http.StatusNotFound)
		}
		log.Printf("Error getting program: %v", err)
		return "", errLib.New("failed to get program", http.StatusInternalServerError)
	}

	if status, err := r.Queries.GetProgramCapacityStatus(ctx, programID); err != nil {
		log.Printf("Error getting program capacity status: %v", err)
		return "", errLib.New("failed to get program stauts", http.StatusInternalServerError)
	} else {

		if !status.Capacity.Valid {
			return "", errLib.New(fmt.Sprintf("program has no capacity: %v", program.Name), http.StatusBadRequest)
		}

		if status.EnrolledCount >= int64(status.Capacity.Int32) {
			return "", errLib.New(fmt.Sprintf("program is full: %v", program.Name), http.StatusBadRequest)
		}
	}

	if isCustomerExist, err := r.Queries.IsCustomerExist(ctx, customerID); err != nil {
		log.Printf("Error getting customer: %v", err)
		return "", errLib.New("failed to get customer", http.StatusInternalServerError)
	} else if !isCustomerExist {
		return "", errLib.New(fmt.Sprintf("customer not found: %v", customerID), http.StatusNotFound)
	}

	priceID, err := r.Queries.GetProgramRegistrationPriceIdForCustomer(ctx, db.GetProgramRegistrationPriceIdForCustomerParams{
		CustomerID: customerID,
		ProgramID:  programID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", errLib.New(fmt.Sprintf("customer is not eligible for program registration: %v", program.Name), http.StatusBadRequest)
		}
		log.Printf("Error getting GetProgramRegisterPricesForCustomer: %v", err)
		return "", errLib.New("failed to check get registration prices for customer", http.StatusInternalServerError)
	}

	return priceID, nil
}
