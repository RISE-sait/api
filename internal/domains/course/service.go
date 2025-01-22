package course

import (
	"api/internal/domains/course/dto"
	infra "api/internal/domains/course/infra"
	"api/internal/libs/errors"
	"api/internal/libs/validators"
	"context"
	"github.com/google/uuid"
)

type Service struct {
	Repo *infra.Repository
}

func NewService(repo *infra.Repository) *Service {
	return &Service{Repo: repo}
}

func (s *Service) CreateCourse(ctx context.Context, body dto.CreateCourseRequestBody) *errLib.CommonError {

	if err := validators.ValidateDto(body); err != nil {
		return err
	}
	if err := s.Repo.CreateCourse(ctx, &body); err != nil {
		return err
	}

	return nil
}

func (s *Service) GetCourseById(ctx context.Context, id uuid.UUID) (*dto.CourseResponse, *errLib.CommonError) {
	course, err := s.Repo.GetCourseById(ctx, id)
	if err != nil {
		return nil, err
	}

	return &dto.CourseResponse{
		ID:   course.ID,
		Name: course.Name,
	}, nil
}

func (s *Service) GetAllCourses(ctx context.Context) (*[]dto.CourseResponse, *errLib.CommonError) {
	courses, err := s.Repo.GetAllCourses(ctx, "")
	if err != nil {
		return nil, err
	}

	var result []dto.CourseResponse
	for _, course := range courses {
		result = append(result, dto.CourseResponse{
			ID:   course.ID,
			Name: course.Name,
		})
	}

	return &result, nil
}

func (s *Service) UpdateCourse(ctx context.Context, body dto.UpdateCourseRequest) *errLib.CommonError {

	if err := validators.ValidateDto(body); err != nil {
		return err
	}

	if err := s.Repo.UpdateCourse(ctx, &body); err != nil {
		return err
	}

	return nil
}

func (s *Service) DeleteCourse(ctx context.Context, id uuid.UUID) *errLib.CommonError {
	if err := s.Repo.DeleteCourse(ctx, id); err != nil {
		return err
	}

	return nil
}
