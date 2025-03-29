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

	hasActiveMembership, err := r.Queries.GetCustomerHasActiveMembershipPlan(ctx, customerID)
	if err != nil {
		log.Printf("Error getting GetCustomerHasActiveMembershipPlan: %v", err)
		return response, errLib.New("failed to check if customer has active membership", http.StatusInternalServerError)
	}

	if !hasActiveMembership {

		// those who dont have active membership can only join the program as payg

		paygPrice, priceErr := r.Queries.GetPaygPrice(ctx, programID)
		if priceErr != nil {
			log.Printf("Error getting PAYG price: %v", priceErr)
			return response, errLib.New("failed to get PAYG pricing information", http.StatusInternalServerError)
		}

		if !paygPrice.Valid {
			return response, errLib.New("program doesn't support PAYG registration", http.StatusBadRequest)
		}

		response.Price = paygPrice

		return response, nil
	}

	// those who have active membership, check if they're eligible

	programRegistrationPrice, err := r.Queries.GetProgramRegisterInfoForCustomer(ctx, db.GetProgramRegisterInfoForCustomerParams{
		CustomerID: customerID,
		ProgramID:  programID,
	})

	if err != nil {
		log.Printf("Error getting customer eligibility for program: %v", err)
		return response, errLib.New("failed to get customer eligibility for program", http.StatusInternalServerError)
	}

	if !programRegistrationPrice.Valid {
		return response, errLib.New("Customer is ineligible to join program", http.StatusBadRequest)
	}

	response.Price = decimal.NullDecimal{
		Decimal: programRegistrationPrice.Decimal,
		Valid:   true,
	}

	return response, nil
}
