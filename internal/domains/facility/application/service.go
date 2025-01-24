package facility

import (
	entity "api/internal/domains/facility/entities"
	"api/internal/domains/facility/infra/persistence"
	errLib "api/internal/libs/errors"
	"context"

	"github.com/google/uuid"
)

type FacilityService struct {
	Repo *persistence.FacilityRepository
}

func NewFacilityManager(service *persistence.FacilityRepository) *FacilityService {
	return &FacilityService{Repo: service}
}

func (s *FacilityService) CreateFacility(ctx context.Context, facility *entity.Facility) *errLib.CommonError {

	return s.Repo.CreateFacility(ctx, facility)

}

func (s *FacilityService) GetFacilityById(ctx context.Context, id uuid.UUID) (*entity.Facility, *errLib.CommonError) {
	return s.Repo.GetFacility(ctx, id)

}

func (s *FacilityService) GetAllFacilities(ctx context.Context) ([]entity.Facility, *errLib.CommonError) {
	return s.Repo.GetAllFacilities(ctx, "")

}

func (s *FacilityService) UpdateFacility(ctx context.Context, facility *entity.Facility) *errLib.CommonError {

	return s.Repo.UpdateFacility(ctx, facility)

}

func (s *FacilityService) DeleteFacility(ctx context.Context, id uuid.UUID) *errLib.CommonError {
	return s.Repo.DeleteFacility(ctx, id)
}
