package practice

import (
	"api/internal/domains/practice/entity"
	"api/internal/domains/practice/persistence/repository"
	"api/internal/domains/practice/values"
	errLib "api/internal/libs/errors"
	"context"

	"github.com/google/uuid"
)

type Service struct {
	Repo repository.PracticeRepositoryInterface
}

func NewPracticeService(repo repository.PracticeRepositoryInterface) *Service {
	return &Service{Repo: repo}
}

func (s *Service) CreatePractice(ctx context.Context, input *values.PracticeDetails) (*entity.Practice, *errLib.CommonError) {

	practice, err := s.Repo.Create(ctx, input)

	return practice, err
}

func (s *Service) GetPracticeByName(ctx context.Context, name string) (*entity.Practice, *errLib.CommonError) {
	return s.Repo.GetPracticeByName(ctx, name)
}

func (s *Service) GetPractices(ctx context.Context, name, description *string) ([]entity.Practice, *errLib.CommonError) {

	return s.Repo.List(ctx)

}

func (s *Service) Update(ctx context.Context, input *entity.Practice) *errLib.CommonError {

	return s.Repo.Update(ctx, input)
}

func (s *Service) DeletePractice(ctx context.Context, id uuid.UUID) *errLib.CommonError {
	return s.Repo.Delete(ctx, id)
}
