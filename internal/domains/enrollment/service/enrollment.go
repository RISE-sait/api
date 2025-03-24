package enrollment

import (
	enrollmentRepo "api/internal/domains/enrollment/persistence"
	"api/internal/domains/enrollment/values"
	errLib "api/internal/libs/errors"
	"context"
	"net/http"

	"github.com/google/uuid"
)

type Service struct {
	EnrollmentRepository *enrollmentRepo.Repository
}

func NewEnrollmentService(enrollmentRepo *enrollmentRepo.Repository) *Service {
	return &Service{
		EnrollmentRepository: enrollmentRepo,
	}
}

func (s *Service) EnrollCustomer(ctx context.Context, details values.EnrollmentCreateDetails) (values.EnrollmentReadDetails, *errLib.CommonError) {

	var readDetails values.EnrollmentReadDetails

	isFull, err := s.EnrollmentRepository.GetEventIsFull(ctx, details.EventId)

	if err != nil {
		return readDetails, err
	}

	if *isFull {
		return readDetails, errLib.New("Event is full", http.StatusConflict)
	}

	return s.EnrollmentRepository.EnrollCustomer(ctx, details)

}

func (s *Service) GetEnrollments(ctx context.Context, eventId, customerId uuid.UUID) ([]values.EnrollmentReadDetails, *errLib.CommonError) {
	return s.EnrollmentRepository.GetEnrollments(ctx, eventId, customerId)
}

func (s *Service) UnEnrollCustomer(ctx context.Context, id uuid.UUID) *errLib.CommonError {
	return s.EnrollmentRepository.UnEnrollCustomer(ctx, id)
}
