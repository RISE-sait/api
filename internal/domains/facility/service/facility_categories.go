package service

import (
	"api/internal/di"
	entity "api/internal/domains/facility/entity"
	repository "api/internal/domains/facility/persistence/repositories"
	errLib "api/internal/libs/errors"
	"context"

	"github.com/google/uuid"
)

type FacilityCategoriesService struct {
	Repo *repository.FacilityTypesRepository
}

func NewFacilityCategoriesService(container *di.Container) *FacilityCategoriesService {
	return &FacilityCategoriesService{Repo: repository.NewFacilityTypesRepository(container)}
}

func (s *FacilityCategoriesService) Create(ctx context.Context, name string) *errLib.CommonError {

	return s.Repo.CreateFacilityType(ctx, name)

}

func (s *FacilityCategoriesService) GetById(ctx context.Context, id uuid.UUID) (*string, *errLib.CommonError) {
	return s.Repo.GetFacilityType(ctx, id)

}

func (s *FacilityCategoriesService) List(ctx context.Context) ([]entity.Category, *errLib.CommonError) {
	return s.Repo.GetAllFacilityTypes(ctx, "")

}

func (s *FacilityCategoriesService) Update(ctx context.Context, facility *entity.Category) *errLib.CommonError {

	return s.Repo.UpdateFacilityType(ctx, facility)

}

func (s *FacilityCategoriesService) Delete(ctx context.Context, id uuid.UUID) *errLib.CommonError {
	return s.Repo.DeleteFacilityType(ctx, id)
}
