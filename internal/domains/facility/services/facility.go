package service

import (
	"api/internal/di"
	entity "api/internal/domains/facility/entities"
	repository "api/internal/domains/facility/persistence/repositories"
	"api/internal/domains/facility/values"
	errLib "api/internal/libs/errors"
	"context"

	"github.com/google/uuid"
)

type FacilityService struct {
	Repo *repository.FacilityRepository
}

func NewFacilityService(container *di.Container) *FacilityService {
	return &FacilityService{Repo: repository.NewFacilityRepository(container)}
}

func (s *FacilityService) CreateFacility(ctx context.Context, facility *values.FacilityDetails) (*entity.Facility, *errLib.CommonError) {

	return s.Repo.CreateFacility(ctx, facility)

}

func (s *FacilityService) GetFacilityById(ctx context.Context, id uuid.UUID) (*entity.Facility, *errLib.CommonError) {
	return s.Repo.GetFacility(ctx, id)

}

func (s *FacilityService) GetFacilities(ctx context.Context, name string) ([]entity.Facility, *errLib.CommonError) {
	return s.Repo.GetFacilities(ctx, name)

}

func (s *FacilityService) UpdateFacility(ctx context.Context, facility *entity.Facility) *errLib.CommonError {

	return s.Repo.UpdateFacility(ctx, facility)

}

func (s *FacilityService) DeleteFacility(ctx context.Context, id uuid.UUID) *errLib.CommonError {
	return s.Repo.DeleteFacility(ctx, id)
}
