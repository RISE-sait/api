package purchase

import (
	"api/internal/di"
	db "api/internal/domains/identity/persistence/sqlc/generated"
	membership "api/internal/domains/membership/persistence/repositories"
	repository "api/internal/domains/purchase/persistence/repositories"
	"api/internal/domains/purchase/values"
	errLib "api/internal/libs/errors"
	"context"
	"net/http"
	"time"
)

type Service struct {
	PurchaseRepo        *repository.Repository
	MembershipPlansRepo *membership.PlansRepository
}

func NewPurchaseService(container *di.Container) *Service {
	return &Service{
		PurchaseRepo:        repository.NewRepository(container),
		MembershipPlansRepo: membership.NewMembershipPlansRepository(container)}
}

func (s *Service) Purchase(ctx context.Context, details values.MembershipPlanPurchaseInfo) *errLib.CommonError {

	plan, err := s.MembershipPlansRepo.GetMembershipPlanById(ctx, details.MembershipPlanId)

	if err != nil {
		return err
	}

	amtPeriods := plan.AmtPeriods
	frequency := plan.PaymentFrequency

	startDate := details.StartDate

	var renewalDate *time.Time

	if amtPeriods != nil {
		switch frequency {
		case string(db.PaymentFrequencyDay):
			renewalDateTemp := startDate.AddDate(0, 0, int(*amtPeriods))
			renewalDate = &renewalDateTemp
		case string(db.PaymentFrequencyWeek):
			renewalDateTemp := startDate.AddDate(0, 0, int(*amtPeriods*7))
			renewalDate = &renewalDateTemp
		case string(db.PaymentFrequencyMonth):
			renewalDateTemp := startDate.AddDate(0, int(*amtPeriods), 0)
			renewalDate = &renewalDateTemp
		default:
			return errLib.New("Invalid payment frequency", http.StatusBadRequest)
		}
	} else {
		// If amtPeriods is nil, renewalDate remains nil
		renewalDate = nil
	}

	details.RenewalDate = renewalDate

	return s.PurchaseRepo.Purchase(ctx, details)

}
