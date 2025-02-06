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

type FacilityTypesService struct {
	Repo *repository.FacilityTypesRepository
}

func NewFacilityTypesService(container *di.Container) *FacilityTypesService {
	return &FacilityTypesService{Repo: repository.NewFacilityTypesRepository(container)}
}

func (s *FacilityTypesService) CreateFacility(ctx context.Context, name string) *errLib.CommonError {

	return s.Repo.CreateFacilityType(ctx, name)

}

func (s *FacilityTypesService) GetFacilityTypeById(ctx context.Context, id uuid.UUID) (*string, *errLib.CommonError) {
	return s.Repo.GetFacilityType(ctx, id)

}

func (s *FacilityTypesService) GetAllFacilityTypes(ctx context.Context) ([]entity.FacilityType, *errLib.CommonError) {
	return s.Repo.GetAllFacilityTypes(ctx, "")

}

func (s *FacilityTypesService) UpdateFacilityType(ctx context.Context, facility *values.FacilityType) *errLib.CommonError {

	return s.Repo.UpdateFacilityType(ctx, facility)

}

func (s *FacilityTypesService) DeleteFacilityType(ctx context.Context, id uuid.UUID) *errLib.CommonError {
	return s.Repo.DeleteFacilityType(ctx, id)
}
