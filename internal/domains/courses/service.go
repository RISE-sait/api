package courses

import (
	dto "api/internal/shared/dto/course"
	"api/internal/types"
	"context"

	"github.com/google/uuid"
)

type Service struct {
	Repo *Repository
}

func NewCourseService(repo *Repository) *Service {
	return &Service{Repo: repo}
}

func (s *Service) CreateCourse(ctx context.Context, body dto.CreateCourseRequestBody) *types.HTTPError {
	// Perform any additional business logic validation if needed
	params := body.ToDBParams()
	return s.Repo.CreateCourse(ctx, params)
}

func (s *Service) GetCourseById(ctx context.Context, id uuid.UUID) (*dto.CourseResponse, *types.HTTPError) {
	course, err := s.Repo.GetCourseById(ctx, id)
	if err != nil {
		return nil, err
	}
	return dto.ToCourseResponse(course), nil
}

func (s *Service) GetAllCourses(ctx context.Context) (*[]dto.CourseResponse, *types.HTTPError) {
	courses, err := s.Repo.GetAllCourses(ctx, "")

	if err != nil {
		return nil, err
	}

	var courseResponses []dto.CourseResponse
	for _, course := range *courses {
		courseResponses = append(courseResponses, *dto.ToCourseResponse(&course))
	}

	// Respond with the list of courses
	return &courseResponses, nil
}

func (s *Service) UpdateCourse(ctx context.Context, body dto.UpdateCourseRequest) *types.HTTPError {
	params := body.ToDBParams()
	return s.Repo.UpdateCourse(ctx, params)
}

func (s *Service) DeleteCourse(ctx context.Context, id uuid.UUID) *types.HTTPError {
	return s.Repo.DeleteCourse(ctx, id)
}
