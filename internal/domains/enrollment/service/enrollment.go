package enrollment

import (
	"api/internal/di"
	repo "api/internal/domains/enrollment/persistence/repository"
	errLib "api/internal/libs/errors"
	"context"
	"github.com/google/uuid"
	"time"
)

type ICustomerEnrollmentService interface {
	EnrollCustomerInProgramEvents(
		ctx context.Context,
		customerID uuid.UUID,
		programID uuid.UUID,
	) *errLib.CommonError

	EnrollCustomerInMembershipPlan(
		ctx context.Context,
		customerID uuid.UUID,
		planID uuid.UUID,
		cancelAtDateTime time.Time,
	) *errLib.CommonError
}

var _ ICustomerEnrollmentService = (*CustomerEnrollmentService)(nil)

type CustomerEnrollmentService struct {
	EnrollmentRepository *repo.CustomerEnrollmentRepository
}

func NewCustomerEnrollmentService(container *di.Container) *CustomerEnrollmentService {
	_repo := repo.NewEnrollmentRepository(container.DB)
	return &CustomerEnrollmentService{
		EnrollmentRepository: _repo,
	}
}

func (s *CustomerEnrollmentService) EnrollCustomerInProgramEvents(ctx context.Context, customerID, programID uuid.UUID) *errLib.CommonError {

	return s.EnrollmentRepository.EnrollCustomerInProgramEvents(ctx, customerID, programID)
}

func (s *CustomerEnrollmentService) EnrollCustomerInMembershipPlan(ctx context.Context, customerID, planID uuid.UUID, cancelAtDateTime time.Time) *errLib.CommonError {
	return s.EnrollmentRepository.EnrollCustomerInMembershipPlan(ctx, customerID, planID, cancelAtDateTime)
}

func (s *CustomerEnrollmentService) UnEnrollCustomer(ctx context.Context, eventID, customerID uuid.UUID) *errLib.CommonError {
	return s.EnrollmentRepository.UnEnrollCustomer(ctx, eventID, customerID)
}
