package payment

import (
	"api/internal/di"
	"api/internal/domains/event/service"
	db "api/internal/domains/payment/persistence/sqlc/generated"
	values "api/internal/domains/payment/values"
	"api/internal/domains/program"
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
	paymentQueries *db.Queries
	programService *program.Service
	eventService   *service.Service
}

func NewCheckoutRepository(container *di.Container) *CheckoutRepository {
	return &CheckoutRepository{
		paymentQueries: container.Queries.PurchasesDb,
		programService: program.NewProgramService(container),
		eventService:   service.NewEventService(container),
	}
}

func (r *CheckoutRepository) GetMembershipPlanJoiningRequirement(ctx context.Context, planID uuid.UUID) (values.MembershipPlanJoiningRequirement, *errLib.CommonError) {

	if requirements, err := r.paymentQueries.GetMembershipPlanJoiningRequirements(ctx, planID); err != nil {
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

func (r *CheckoutRepository) GetRegistrationPriceIdForCustomerByProgramID(ctx context.Context, programID uuid.UUID) (isPayPerEvent bool, priceID string, error *errLib.CommonError) {

	customerID, ctxErr := contextUtils.GetUserID(ctx)
	if ctxErr != nil {
		return false, "", ctxErr
	}

	retrievedProgram, getProgramErr := r.programService.GetProgram(ctx, programID)
	if getProgramErr != nil {
		return false, "", getProgramErr
	}

	if isCustomerExist, err := r.paymentQueries.IsCustomerExist(ctx, customerID); err != nil {
		log.Printf("Error getting customer: %v", err)
		return false, "", errLib.New("failed to get customer", http.StatusInternalServerError)
	} else if !isCustomerExist {
		return false, "", errLib.New(fmt.Sprintf("customer not found: %v", customerID), http.StatusNotFound)
	}

	registrationInfo, err := r.paymentQueries.GetRegistrationPriceIdForCustomer(ctx, db.GetRegistrationPriceIdForCustomerParams{
		CustomerID: customerID,
		ProgramID:  programID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, "", errLib.New(fmt.Sprintf("customer is not eligible for retrievedProgram registration: %v", retrievedProgram.Name), http.StatusForbidden)
		}
		log.Printf("Error getting GetProgramRegisterPricesForCustomer: %v", err)
		return false, "", errLib.New("failed to check get registration prices for customer", http.StatusInternalServerError)
	}

	return registrationInfo.PayPerEvent, registrationInfo.StripePriceID, nil
}

func (r *CheckoutRepository) GetProgramIDOfEvent(ctx context.Context, eventID uuid.UUID) (uuid.UUID, *errLib.CommonError) {

	retrievedEvent, err := r.eventService.GetEvent(ctx, eventID)

	if err != nil {
		if err.HTTPCode == http.StatusNotFound {
			return uuid.Nil, errLib.New(fmt.Sprintf("retrievedEvent not found: %v", eventID), http.StatusNotFound)
		}
		log.Printf("Error getting program of the retrievedEvent: %v", err)
		return uuid.Nil, errLib.New("failed to get program of the retrievedEvent", http.StatusInternalServerError)
	}

	return retrievedEvent.Program.ID, nil
}
