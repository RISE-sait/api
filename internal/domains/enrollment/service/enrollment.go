package enrollment_service

import (
	"api/internal/domains/enrollment/entity"
	enrollmentRepo "api/internal/domains/enrollment/persistence/repository/enrollment"
	eventCapacityRepo "api/internal/domains/enrollment/persistence/repository/event_capacity"
	"api/internal/domains/enrollment/values"
	errLib "api/internal/libs/errors"
	"context"
	"net/http"

	"github.com/google/uuid"
)

type EnrollmentService struct {
	EnrollmentRepository    *enrollmentRepo.Repository
	EventCapacityRepository *eventCapacityRepo.Repository
}

func NewEnrollmentService(enrollmentRepo *enrollmentRepo.Repository, eventCapacityRepo *eventCapacityRepo.Repository) *EnrollmentService {
	return &EnrollmentService{
		EnrollmentRepository:    enrollmentRepo,
		EventCapacityRepository: eventCapacityRepo,
	}
}

func (s *EnrollmentService) EnrollCustomer(ctx context.Context, details values.EnrollmentDetails) (*entity.Enrollment, *errLib.CommonError) {

	isFull, err := s.EventCapacityRepository.GetEventIsFull(ctx, details.EventId)

	if err != nil {
		return nil, err
	}

	if *isFull {
		return nil, errLib.New("Event is full", http.StatusConflict)
	}

	return s.EnrollmentRepository.EnrollCustomer(ctx, details)

}

func (s *EnrollmentService) GetEnrollments(ctx context.Context, eventId, customerId *uuid.UUID) ([]entity.Enrollment, *errLib.CommonError) {
	return s.EnrollmentRepository.GetEnrollments(ctx, eventId, customerId)
}

func (s *EnrollmentService) UnEnrollCustomer(ctx context.Context, id uuid.UUID) *errLib.CommonError {
	return s.EnrollmentRepository.UnEnrollCustomer(ctx, id)
}
