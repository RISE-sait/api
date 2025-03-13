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

	amtPeriods := int(plan.AmtPeriods)
	frequency := plan.PaymentFrequency

	startDate := details.StartDate

	var renewalDate time.Time

	switch frequency {
	case string(db.PaymentFrequencyDay):
		renewalDate = startDate.AddDate(0, 0, amtPeriods)
	case string(db.PaymentFrequencyWeek):
		renewalDate = startDate.AddDate(0, 0, amtPeriods*7)
	case string(db.PaymentFrequencyMonth):
		renewalDate = startDate.AddDate(0, amtPeriods, 0)
	default:
		return errLib.New("Invalid payment frequency", http.StatusBadRequest)
	}

	details.RenewalDate = &renewalDate

	return s.PurchaseRepo.Purchase(ctx, details)

}
