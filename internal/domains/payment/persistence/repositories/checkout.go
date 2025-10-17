package payment

import (
	"api/internal/di"
	"api/internal/domains/event/service"
	eventDb "api/internal/domains/event/persistence/sqlc/generated"
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
	eventQueries   *eventDb.Queries
	programService *program.Service
	eventService   *service.Service
}

func NewCheckoutRepository(container *di.Container) *CheckoutRepository {
	return &CheckoutRepository{
		paymentQueries: container.Queries.PurchasesDb,
		eventQueries:   container.Queries.EventDb,
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
			JoiningFee:    int(requirements.JoiningFee),
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

// CheckCustomerHasActiveMembership checks if customer already has an active membership for the given plan
func (r *CheckoutRepository) CheckCustomerHasActiveMembership(ctx context.Context, customerID uuid.UUID, membershipPlanID uuid.UUID) (bool, *errLib.CommonError) {
	count, err := r.paymentQueries.CheckCustomerActiveMembership(ctx, db.CheckCustomerActiveMembershipParams{
		CustomerID:       customerID,
		MembershipPlanID: membershipPlanID,
	})
	if err != nil {
		log.Printf("Error checking customer active membership: %v", err)
		return false, errLib.New("failed to check existing membership", http.StatusInternalServerError)
	}

	return count > 0, nil
}

// CheckCustomerHasEventMembershipAccess checks if customer has an active membership for ANY of the event's required memberships
func (r *CheckoutRepository) CheckCustomerHasEventMembershipAccess(ctx context.Context, eventID uuid.UUID, customerID uuid.UUID) (bool, *errLib.CommonError) {
	// Use the optimized SQL query that checks via junction table join
	hasAccess, err := r.eventQueries.CheckCustomerHasEventMembershipAccess(ctx, eventDb.CheckCustomerHasEventMembershipAccessParams{
		EventID:    eventID,
		CustomerID: customerID,
	})
	if err != nil {
		log.Printf("Error checking customer event membership access: %v", err)
		return false, errLib.New("failed to check event membership access", http.StatusInternalServerError)
	}

	return hasAccess, nil
}
