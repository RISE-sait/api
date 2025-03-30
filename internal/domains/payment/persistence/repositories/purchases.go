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
	"log"
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

	if requirements, err := r.Queries.GetMembershipPlanJoiningRequirements(ctx, planID); err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return values.MembershipPlanJoiningRequirement{}, errLib.New(fmt.Sprintf("error getting joining requirement for membership plan: %v", planID), http.StatusBadRequest)
		} else {
			return values.MembershipPlanJoiningRequirement{}, errLib.New(fmt.Sprintf("membership plan not found for plan id: %v", planID), http.StatusBadRequest)
		}
	} else {
		response := values.MembershipPlanJoiningRequirement{
			ID:               requirements.ID,
			Name:             requirements.Name,
			Price:            requirements.Price,
			JoiningFee:       requirements.JoiningFee,
			AutoRenew:        requirements.AutoRenew,
			MembershipID:     requirements.MembershipID,
			PaymentFrequency: string(requirements.PaymentFrequency),
			CreatedAt:        requirements.CreatedAt,
			UpdatedAt:        requirements.UpdatedAt,
		}

		if requirements.PaymentFrequency == db.PaymentFrequencyOnce {
			response.IsOneTimePayment = true
		}

		if requirements.AmtPeriods.Valid {
			response.AmtPeriods = &requirements.AmtPeriods.Int32
		}

		return response, nil
	}
}

func (r *Repository) GetProgramRegistrationInfoByCustomer(ctx context.Context, customerID, programID uuid.UUID) (values.ProgramRegistrationInfo, *errLib.CommonError) {

	var response values.ProgramRegistrationInfo

	program, err := r.Queries.GetProgram(ctx, programID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return response, errLib.New(fmt.Sprintf("program not found: %v", programID), http.StatusNotFound)
		}
		log.Printf("Error getting program: %v", err)
		return response, errLib.New("failed to get program", http.StatusInternalServerError)
	}

	response.ProgramName = program.Name

	if isCustomerExist, err := r.Queries.IsCustomerExist(ctx, customerID); err != nil {
		log.Printf("Error getting customer: %v", err)
		return response, errLib.New("failed to get customer", http.StatusInternalServerError)
	} else if !isCustomerExist {
		return response, errLib.New(fmt.Sprintf("customer not found: %v", customerID), http.StatusNotFound)
	}

	prices, err := r.Queries.GetProgramRegisterPricesForCustomer(ctx, db.GetProgramRegisterPricesForCustomerParams{
		CustomerID: customerID,
		ProgramID:  programID,
	})
	if err != nil {
		log.Printf("Error getting GetProgramRegisterPricesForCustomer: %v", err)
		return response, errLib.New("failed to check get registration prices for customer", http.StatusInternalServerError)
	}

	memberProgramPrice := prices.MemberProgramPrice
	memberDropInPrice := prices.MemberDropInPrice

	nonMemberProgramPrice := prices.NonMemberProgramPrice
	nonMemberDropInPrice := prices.NonMemberDropInPrice

	// check if the customer is a member, and if it's paid by program or drop-in
	switch {
	case memberProgramPrice.Valid:
		response.Price = memberProgramPrice
	case memberDropInPrice.Valid:
		response.Price = memberDropInPrice
	case nonMemberProgramPrice.Valid:
		response.Price = nonMemberProgramPrice
	case nonMemberDropInPrice.Valid:
		response.Price = nonMemberDropInPrice
	}

	if response.Price.Valid {
		return response, nil
	}

	return response, errLib.New(fmt.Sprintf("customer is ineligible for program registration: %v", program.Name), http.StatusBadRequest)
}
