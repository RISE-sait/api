package payment

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"api/internal/di"
	discountService "api/internal/domains/discount/service"
	enrollment "api/internal/domains/enrollment/service"
	membership "api/internal/domains/membership/persistence/repositories"
	repository "api/internal/domains/payment/persistence/repositories"
	errLib "api/internal/libs/errors"
	contextUtils "api/utils/context"

	"github.com/google/uuid"
	squareClient "github.com/square/square-go-sdk/client"
)

type Service struct {
	CheckoutRepo        *repository.CheckoutRepository
	MembershipPlansRepo *membership.PlansRepository
	SquareClient        *squareClient.Client // deprecated
	SquareServiceURL    string
	DiscountService     *discountService.Service
	EnrollmentService   *enrollment.CustomerEnrollmentService
}

func NewPurchaseService(container *di.Container) *Service {
	return &Service{
		CheckoutRepo:        repository.NewCheckoutRepository(container),
		MembershipPlansRepo: membership.NewMembershipPlansRepository(container),
		DiscountService:     discountService.NewService(container),
		EnrollmentService:   enrollment.NewCustomerEnrollmentService(container),
		SquareClient:        container.SquareClient,
		SquareServiceURL:    os.Getenv("SQUARE_SERVICE_URL"),
	}
}

func (s *Service) postToSquare(endpoint string, payload interface{}) (string, *errLib.CommonError) {
	if s.SquareServiceURL == "" {
		return "", errLib.New("square service url not configured", http.StatusInternalServerError)
	}
	b, _ := json.Marshal(payload)
	resp, err := http.Post(s.SquareServiceURL+endpoint, "application/json", bytes.NewBuffer(b))
	if err != nil {
		return "", errLib.New(fmt.Sprintf("square service request failed: %v", err), http.StatusInternalServerError)
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return "", errLib.New(string(data), resp.StatusCode)
	}
	type response struct {
		CheckoutLink string `json:"checkout_page_url"`
	}
	var r response
	json.Unmarshal(data, &r)
	if r.CheckoutLink == "" {
		return string(data), nil
	}
	return r.CheckoutLink, nil
}

func (s *Service) CheckoutMembershipPlan(ctx context.Context, membershipPlanID uuid.UUID, discountCode *string) (string, *errLib.CommonError) {
	requirements, err := s.CheckoutRepo.GetMembershipPlanJoiningRequirement(ctx, membershipPlanID)
	if err != nil {
		return "", err
	}

	userID, ctxErr := contextUtils.GetUserID(ctx)
	if ctxErr != nil {
		return "", ctxErr
	}

	payload := map[string]interface{}{
		"plan_id":     requirements.StripePriceID,
		"customer_id": userID.String(),
	}

	return s.postToSquare("/checkout/subscription", payload)
}

func (s *Service) CheckoutProgram(ctx context.Context, programID uuid.UUID) (string, *errLib.CommonError) {
	customerID, ctxErr := contextUtils.GetUserID(ctx)
	if ctxErr != nil {
		return "", ctxErr
	}

	isPayPerEvent, priceID, err := s.CheckoutRepo.GetRegistrationPriceIdForCustomerByProgramID(ctx, programID)
	if err != nil {
		return "", err
	}

	if isPayPerEvent {
		return "", errLib.New("program is not pay-per-event", http.StatusBadRequest)
	}

	if err := s.EnrollmentService.ReserveSeatInProgram(ctx, programID, customerID); err != nil {
		log.Println("Failed to reserve seat in program:", err)
		return "", err
	}

	payload := map[string]interface{}{
		"amount": priceID, // TODO: replace with actual price/plan structure from Python service
	}
	return s.postToSquare("/checkout/payment", payload)
}

func (s *Service) CheckoutEvent(ctx context.Context, eventID uuid.UUID) (string, *errLib.CommonError) {
	customerID, ctxErr := contextUtils.GetUserID(ctx)
	if ctxErr != nil {
		return "", ctxErr
	}

	programID, err := s.CheckoutRepo.GetProgramIDOfEvent(ctx, eventID)
	if err != nil {
		return "", err
	}

	isPayPerEvent, priceID, err := s.CheckoutRepo.GetRegistrationPriceIdForCustomerByProgramID(ctx, programID)
	if err != nil {
		return "", err
	}

	if !isPayPerEvent {
		return "", errLib.New("event is pay-per-program", http.StatusBadRequest)
	}

	if err := s.EnrollmentService.ReserveSeatInEvent(ctx, eventID, customerID); err != nil {
		return "", err
	}

	payload := map[string]interface{}{
		"plan_id": priceID,
	}
	return s.postToSquare("/checkout/payment", payload)
}
