package event

import (
	eventRepo "api/internal/domains/event/persistence/repository"
	errLib "api/internal/libs/errors"
	"context"
	"net/http"

	"github.com/google/uuid"
)

type CustomerEnrollmentService struct {
	EnrollmentRepository *eventRepo.CustomerEnrollmentRepository
}

func NewEnrollmentService(enrollmentRepo *eventRepo.CustomerEnrollmentRepository) *CustomerEnrollmentService {
	return &CustomerEnrollmentService{
		EnrollmentRepository: enrollmentRepo,
	}
}

func (s *CustomerEnrollmentService) EnrollCustomer(ctx context.Context, eventID, customerID uuid.UUID) *errLib.CommonError {

	if isFull, err := s.EnrollmentRepository.GetEventIsFull(ctx, eventID); err != nil {
		return err
	} else if isFull {
		return errLib.New("Event is full", http.StatusBadRequest)
	}

	return s.EnrollmentRepository.EnrollCustomer(ctx, eventID, customerID)

}

func (s *CustomerEnrollmentService) UnEnrollCustomer(ctx context.Context, eventID, customerID uuid.UUID) *errLib.CommonError {
	return s.EnrollmentRepository.UnEnrollCustomer(ctx, eventID, customerID)
}
