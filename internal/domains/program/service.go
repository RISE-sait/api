package program

import (
	"api/internal/di"
	repo "api/internal/domains/program/persistence"
	"api/internal/domains/program/values"
	errLib "api/internal/libs/errors"
	"context"
	"github.com/google/uuid"
	"log"
)

type Service struct {
	repo *repo.Repository
}

func NewProgramService(container *di.Container) *Service {

	log.Println("Creating new program service")
	return &Service{
		repo: repo.NewProgramRepository(container),
	}
}

func (s *Service) GetProgram(ctx context.Context, programID uuid.UUID) (values.GetProgramValues, *errLib.CommonError) {

	return s.repo.GetProgramByID(ctx, programID)
}

func (s *Service) GetPrograms(ctx context.Context, programType string) ([]values.GetProgramValues, *errLib.CommonError) {

	return s.repo.List(ctx, programType)
}

func (s *Service) GetProgramLevels() []string {

	return s.repo.GetProgramLevels()
}

func (s *Service) CreateProgram(ctx context.Context, details values.CreateProgramValues) *errLib.CommonError {
	return s.repo.Create(ctx, details)
}

func (s *Service) UpdateProgram(ctx context.Context, details values.UpdateProgramValues) *errLib.CommonError {
	return s.repo.Update(ctx, details)
}

func (s *Service) DeleteProgram(ctx context.Context, id uuid.UUID) *errLib.CommonError {
	return s.repo.Delete(ctx, id)
}
