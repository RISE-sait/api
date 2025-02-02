package facility

import (
	"api/cmd/server/di"
	"api/internal/domains/facility/persistence"
	"api/internal/domains/facility/values"
	errLib "api/internal/libs/errors"
	"context"

	"github.com/google/uuid"
)

type FacilityService struct {
	Repo *persistence.FacilityRepository
}

func NewFacilityService(container *di.Container) *FacilityService {
	return &FacilityService{Repo: persistence.NewFacilityRepository(container)}
}

func (s *FacilityService) CreateFacility(ctx context.Context, facility *values.FacilityDetails) *errLib.CommonError {

	return s.Repo.CreateFacility(ctx, facility)

}

func (s *FacilityService) GetFacilityById(ctx context.Context, id uuid.UUID) (*values.FacilityAllFields, *errLib.CommonError) {
	return s.Repo.GetFacility(ctx, id)

}

func (s *FacilityService) GetAllFacilities(ctx context.Context) ([]values.FacilityAllFields, *errLib.CommonError) {
	return s.Repo.GetAllFacilities(ctx, "")

}

func (s *FacilityService) UpdateFacility(ctx context.Context, facility *values.FacilityAllFields) *errLib.CommonError {

	return s.Repo.UpdateFacility(ctx, facility)

}

func (s *FacilityService) DeleteFacility(ctx context.Context, id uuid.UUID) *errLib.CommonError {
	return s.Repo.DeleteFacility(ctx, id)
}
