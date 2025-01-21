package facilities

import (
	dto "api/internal/shared/dto/facility"
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

func (s *Service) CreateFacility(ctx context.Context, req *dto.CreateFacilityRequest) *types.HTTPError {
	params := req.ToDBParams()
	return s.Repo.CreateFacility(ctx, params)
}

func (s *Service) GetFacility(ctx context.Context, id uuid.UUID) (*dto.FacilityResponse, *types.HTTPError) {
	facility, err := s.Repo.GetFacility(ctx, id)
	if err != nil {
		return nil, err
	}

	return dto.ToFacilityResponse(facility), nil
}

func (s *Service) GetAllFacilities(ctx context.Context) (*[]dto.FacilityResponse, *types.HTTPError) {
	facilities, err := s.Repo.GetAllFacilities(ctx)
	if err != nil {
		return nil, err
	}

	// Map DB models to DTOs
	var responses []dto.FacilityResponse
	for _, facility := range *facilities {
		responses = append(responses, *dto.ToFacilityResponse(&facility))
	}

	return &responses, nil
}

func (s *Service) UpdateFacility(ctx context.Context, req *dto.UpdateFacilityRequest) *types.HTTPError {
	params := req.ToDBParams()
	return s.Repo.UpdateFacility(ctx, params)
}

func (s *Service) DeleteFacility(ctx context.Context, id uuid.UUID) *types.HTTPError {
	return s.Repo.DeleteFacility(ctx, id)
}
