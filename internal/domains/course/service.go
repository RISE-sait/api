package course

import (
	persistence "api/internal/domains/course/persistence"
	"api/internal/domains/course/values"
	errLib "api/internal/libs/errors"
	"context"

	"github.com/google/uuid"
)

type CourseService struct {
	Repo persistence.CourseRepositoryInterface
}

func NewCourseService(repo persistence.CourseRepositoryInterface) *CourseService {
	return &CourseService{Repo: repo}
}

func (s *CourseService) CreateCourse(ctx context.Context, input *values.CourseDetails) (*values.CourseAllFields, *errLib.CommonError) {

	course, err := s.Repo.CreateCourse(ctx, input)

	return course, err
}

func (s *CourseService) GetCourseById(ctx context.Context, id uuid.UUID) (*values.CourseAllFields, *errLib.CommonError) {
	return s.Repo.GetCourseById(ctx, id)
}

func (s *CourseService) GetCourses(ctx context.Context, name, description *string) ([]values.CourseAllFields, *errLib.CommonError) {

	return s.Repo.GetCourses(ctx, name, description)

}

func (s *CourseService) UpdateCourse(ctx context.Context, input *values.CourseAllFields) *errLib.CommonError {

	return s.Repo.UpdateCourse(ctx, input)
}

func (s *CourseService) DeleteCourse(ctx context.Context, id uuid.UUID) *errLib.CommonError {
	return s.Repo.DeleteCourse(ctx, id)
}
