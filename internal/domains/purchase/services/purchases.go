package purchase

import (
	"api/internal/di"
	membership "api/internal/domains/membership/persistence/repositories"
	repository "api/internal/domains/purchase/persistence/repositories"
	errLib "api/internal/libs/errors"
	"api/internal/services/square"
	"context"
	"github.com/google/uuid"
	squareClient "github.com/square/square-go-sdk/client"
)

type Service struct {
	PurchaseRepo        *repository.Repository
	MembershipPlansRepo *membership.PlansRepository
	SquareClient        *squareClient.Client
}

func NewPurchaseService(container *di.Container) *Service {
	return &Service{
		PurchaseRepo:        repository.NewRepository(container),
		MembershipPlansRepo: membership.NewMembershipPlansRepository(container),
		SquareClient:        container.SquareClient,
	}
}

func (s *Service) Checkout(ctx context.Context, membershipPlanID, userID uuid.UUID) (string, *errLib.CommonError) {

	joiningFees, err := s.PurchaseRepo.GetJoiningFees(ctx, membershipPlanID)

	if err != nil {
		return "", err
	}

	if link, err := square.GetPaymentLink(s.SquareClient, ctx, userID, membershipPlanID.String(), 1, joiningFees); err != nil {
		return "", err
	} else {
		return link, nil
	}

	//plan, err := s.MembershipPlansRepo.GetMembershipPlanById(ctx, details.MembershipPlanId)
	//
	//if err != nil {
	//	return err
	//}
	//
	//amtPeriods := plan.AmtPeriods
	//frequency := plan.PaymentFrequency
	//
	//startDate := details.StartDate
	//
	//var renewalDate *time.Time
	//
	//if amtPeriods != nil {
	//	switch frequency {
	//	case string(db.PaymentFrequencyDay):
	//		renewalDateTemp := startDate.AddDate(0, 0, int(*amtPeriods))
	//		renewalDate = &renewalDateTemp
	//	case string(db.PaymentFrequencyWeek):
	//		renewalDateTemp := startDate.AddDate(0, 0, int(*amtPeriods*7))
	//		renewalDate = &renewalDateTemp
	//	case string(db.PaymentFrequencyMonth):
	//		renewalDateTemp := startDate.AddDate(0, int(*amtPeriods), 0)
	//		renewalDate = &renewalDateTemp
	//	default:
	//		return errLib.New("Invalid payment frequency", http.StatusBadRequest)
	//	}
	//} else {
	//	// If amtPeriods is nil, renewalDate remains nil
	//	renewalDate = nil
	//}
	//
	//details.RenewalDate = renewalDate
	//
	//return s.PurchaseRepo.Purchase(ctx, details)

}

//func (s *Service) ProcessSquareWebhook(ctx context.Context, event dto.SquareWebhookEventDto) *errLib.CommonError {
//	if event.Type != "payment.created" && event.Type != "payment.updated" {
//		return nil // Ignore other event types
//	}
//
//	var paymentData struct {
//		Payment dto.SquarePaymentDto `json:"payment"`
//	}
//
//	if err := json.Unmarshal(event.Data.Object, &paymentData); err != nil {
//		return errLib.New("failed to parse payment data", http.StatusInternalServerError)
//	}
//
//	payment := paymentData.Payment
//
//	// Process payment based on status
//	switch payment.Status {
//	case "APPROVED":
//		// Handle successful payment logic
//	case "FAILED":
//		// Handle failed payment logic
//	default:
//		// Ignore other statuses
//	}
//
//	return nil
//}
