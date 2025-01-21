package facility_types

import (
	dto "api/internal/shared/dto/facilityType"
	"api/internal/types"
	"context"

	"github.com/google/uuid"
)

type Service struct {
	Repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{Repo: repo}
}

func (s *Service) GetFacilityTypeByID(ctx context.Context, id uuid.UUID) (*dto.FacilityTypeResponse, *types.HTTPError) {
	facilityType, err := s.Repo.GetFacilityType(ctx, id)
	if err != nil {
		return nil, err
	}
	return dto.ToFacilityTypeResponse(facilityType), nil
}

func (s *Service) GetAllFacilityTypes(ctx context.Context) (*[]dto.FacilityTypeResponse, *types.HTTPError) {
	facilityTypes, err := s.Repo.GetAllFacilityTypes(ctx)
	if err != nil {
		return nil, err
	}
	var responses []dto.FacilityTypeResponse
	for _, facilityType := range *facilityTypes {
		responses = append(responses, *dto.ToFacilityTypeResponse(&facilityType))
	}

	return &responses, nil
}

func (s *Service) CreateFacilityType(ctx context.Context, name string) *types.HTTPError {
	err := s.Repo.CreateFacilityType(ctx, name)
	if err != nil {
		return err
	}
	return nil
}

func (s *Service) UpdateFacilityType(ctx context.Context, targetBody dto.UpdateFacilityTypeRequest) *types.HTTPError {
	params := targetBody.ToDBParams()

	err := s.Repo.UpdateFacilityType(ctx, params)
	if err != nil {
		return err
	}
	return nil
}

func (s *Service) DeleteFacilityType(ctx context.Context, id uuid.UUID) *types.HTTPError {
	err := s.Repo.DeleteFacilityType(ctx, id)
	if err != nil {
		return err
	}
	return nil
}
