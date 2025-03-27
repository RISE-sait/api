package payment

import (
	"api/internal/di"
	db "api/internal/domains/payment/persistence/sqlc/generated"
	"api/internal/domains/payment/values"
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

func (r *Repository) GetMembershipPlanJoiningRequirement(ctx context.Context, planID uuid.UUID) (values.MembershipPlanJoiningRequirement, *errLib.CommonError) {

	if planInfo, err := r.Queries.GetMembershipPlanJoiningRequirements(ctx, planID); err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return values.MembershipPlanJoiningRequirement{}, errLib.New(fmt.Sprintf("error getting joining requirement for membership plan: %v", planID), http.StatusBadRequest)
		} else {
			return values.MembershipPlanJoiningRequirement{}, errLib.New(fmt.Sprintf("membership plan not found for plan id: %v", planID), http.StatusBadRequest)
		}
	} else {
		response := values.MembershipPlanJoiningRequirement{
			ID:               planInfo.ID,
			Name:             planInfo.Name,
			Price:            planInfo.Price,
			JoiningFee:       planInfo.JoiningFee,
			AutoRenew:        planInfo.AutoRenew,
			MembershipID:     planInfo.MembershipID,
			PaymentFrequency: string(planInfo.PaymentFrequency),
			CreatedAt:        planInfo.CreatedAt,
			UpdatedAt:        planInfo.UpdatedAt,
		}

		if planInfo.PaymentFrequency == db.PaymentFrequencyOnce {
			response.IsOneTimePayment = true
		}

		if planInfo.AmtPeriods.Valid {
			response.AmtPeriods = &planInfo.AmtPeriods.Int32
		}

		return response, nil
	}
}

func (r *Repository) GetProgramRegistrationInfoByCustomer(ctx context.Context, customerID, programID uuid.UUID) (values.ProgramRegistrationInfo, *errLib.CommonError) {

	var response values.ProgramRegistrationInfo

	info, _ := r.Queries.GetProgramRegisterInfoForCustomer(ctx, db.GetProgramRegisterInfoForCustomerParams{
		CustomerID: customerID,
		ProgramID:  programID,
	})

	if !info.CustomerExists {
		return response, errLib.New(fmt.Sprintf("customer not found: %v", customerID), http.StatusNotFound)
	}

	if !info.ProgramExists {
		return response, errLib.New(fmt.Sprintf("program not found: %v", programID), http.StatusNotFound)
	}

	if !info.CustomerHasActiveMembership {
		return response, errLib.New("customer does not have an active membership", http.StatusBadRequest)
	}

	if !info.IsEligible.Valid {
		return values.ProgramRegistrationInfo{
			EligibleRequirement: nil,
		}, errLib.New("program registration requirement not found", http.StatusNotFound)
	}

	if !info.IsEligible.Bool {
		return values.ProgramRegistrationInfo{
			EligibleRequirement: nil,
		}, errLib.New("customer is not eligible to join program", http.StatusBadRequest)
	}

	if !info.PricePerBooking.Valid {
		return response, errLib.New("price per booking not found", http.StatusNotFound)
	}

	response = values.ProgramRegistrationInfo{
		ProgramName: info.ProgramName,
		EligibleRequirement: &struct {
			PricePerBooking decimal.Decimal
		}{
			PricePerBooking: info.PricePerBooking.Decimal,
		},
	}

	return response, nil
}
