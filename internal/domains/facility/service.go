package facility

import (
	"api/internal/domains/facility/dto"
	repository "api/internal/domains/facility/infra/persistence"
	errLib "api/internal/libs/errors"
	"api/internal/libs/validators"
	"context"

	"github.com/google/uuid"
)

type Service struct {
	Repo *repository.FacilityRepository
}

func NewService(repo *repository.FacilityRepository) *Service {
	return &Service{Repo: repo}
}

func (s *Service) CreateFacility(ctx context.Context, body dto.CreateFacilityRequest) *errLib.CommonError {

	if err := validators.ValidateDto(&body); err != nil {
		return err
	}

	if err := s.Repo.CreateFacility(ctx, &body); err != nil {
		return err
	}

	return nil
}

func (s *Service) GetFacilityById(ctx context.Context, id uuid.UUID) (*dto.FacilityResponse, *errLib.CommonError) {
	facility, err := s.Repo.GetFacility(ctx, id)
	if err != nil {
		return nil, err
	}

	return &dto.FacilityResponse{
		ID:             facility.ID,
		Name:           facility.Name,
		Location:       facility.Location,
		FacilityTypeID: facility.FacilityTypeID,
	}, nil
}

func (s *Service) GetAllFacilities(ctx context.Context) (*[]dto.FacilityResponse, *errLib.CommonError) {
	facilities, err := s.Repo.GetAllFacilities(ctx)
	if err != nil {
		return nil, err
	}

	result := []dto.FacilityResponse{}
	for _, facility := range facilities {
		result = append(result, dto.FacilityResponse{
			ID:             facility.ID,
			Name:           facility.Name,
			Location:       facility.Location,
			FacilityTypeID: facility.FacilityTypeID,
		})
	}

	return &result, nil
}

func (s *Service) UpdateFacility(ctx context.Context, body dto.UpdateFacilityRequest) *errLib.CommonError {

	if err := validators.ValidateDto(&body); err != nil {
		return err
	}

	if err := s.Repo.UpdateFacility(ctx, &body); err != nil {
		return err
	}

	return nil
}

func (s *Service) DeleteFacility(ctx context.Context, id uuid.UUID) *errLib.CommonError {
	if err := s.Repo.DeleteFacility(ctx, id); err != nil {
		return err
	}

	return nil
}
