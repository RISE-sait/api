package payment

import (
	"api/internal/di"
	enrollment "api/internal/domains/enrollment/service"
	repository "api/internal/domains/payment/persistence/repositories"
	errLib "api/internal/libs/errors"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	identityRepo "api/internal/domains/identity/persistence/repository/user"
	membershipRepo "api/internal/domains/membership/persistence/repositories"
	"api/utils/email"

	"github.com/google/uuid"
)

// SquareWebhookService handles incoming Square webhook events.
type SquareWebhookService struct {
	PostCheckoutRepository *repository.PostCheckoutRepository
	EnrollmentService      *enrollment.CustomerEnrollmentService
	UserRepo               *identityRepo.UsersRepository
	PlansRepo              *membershipRepo.PlansRepository
}

func NewSquareWebhookService(container *di.Container) *SquareWebhookService {
	return &SquareWebhookService{
		PostCheckoutRepository: repository.NewPostCheckoutRepository(container),
		EnrollmentService:      enrollment.NewCustomerEnrollmentService(container),
		UserRepo:               identityRepo.NewUserRepository(container),
		PlansRepo:              membershipRepo.NewMembershipPlansRepository(container),
	}
}

type squareEvent struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

// squareSubscription represents the subset of the Square Subscription object we
// care about for membership enrollment.
type squareSubscription struct {
	ID         string `json:"id"`
	CustomerID string `json:"customer_id"`
	PlanID     string `json:"plan_id"`
	Status     string `json:"status"`
}

// Handle processes a Square webhook payload.
func (s *SquareWebhookService) Handle(payload []byte) *errLib.CommonError {
	var evt squareEvent
	if err := json.Unmarshal(payload, &evt); err != nil {
		return errLib.New("failed to parse webhook", http.StatusBadRequest)
	}

	switch evt.Type {
	case "subscription.created", "subscription.updated":
		return s.handleSubscriptionEvent(evt.Data)
	default:
		// Unhandled event types are ignored
		log.Println("Unhandled Square event type", evt.Type)
		return nil
	}
}

func (s *SquareWebhookService) handleSubscriptionEvent(data json.RawMessage) *errLib.CommonError {
	var sub squareSubscription
	if err := json.Unmarshal(data, &sub); err != nil {
		return errLib.New("invalid subscription payload", http.StatusBadRequest)
	}

	if sub.Status != "ACTIVE" && sub.Status != "PENDING" {
		// Only process active/pending subscriptions
		return nil
	}

	customerID, err := uuid.Parse(sub.CustomerID)
	if err != nil {
		return errLib.New("invalid customer id", http.StatusBadRequest)
	}

	planID, periods, err := s.PostCheckoutRepository.GetMembershipPlanByStripePriceID(context.Background(), sub.PlanID)
	if err != nil {
		if commonErr, ok := err.(*errLib.CommonError); ok {
			return commonErr
		}
		return errLib.New(err.Error(), http.StatusInternalServerError)
	}

	cancelAt := time.Time{}
	if periods != nil {
		cancelAt = time.Now().AddDate(0, int(*periods), 0)
	}

	if err := s.EnrollmentService.EnrollCustomerInMembershipPlan(context.Background(), customerID, planID, cancelAt); err != nil {
		return err
	}

	s.sendMembershipPurchaseEmail(customerID, planID)
	return nil
}

func (s *SquareWebhookService) sendMembershipPurchaseEmail(userID, planID uuid.UUID) {
	userInfo, err := s.UserRepo.GetUserInfo(context.Background(), "", userID)
	if err != nil || userInfo.Email == nil {
		return
	}

	plan, pErr := s.PlansRepo.GetMembershipPlanById(context.Background(), planID)
	if pErr != nil {
		return
	}

	email.SendMembershipPurchaseEmail(*userInfo.Email, userInfo.FirstName, plan.Name)
}
