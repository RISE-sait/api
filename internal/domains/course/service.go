package course

import (
	"api/cmd/server/di"
	entity "api/internal/domains/course/entities"
	"api/internal/domains/course/persistence"
	"api/internal/domains/course/values"
	errLib "api/internal/libs/errors"
	"context"

	"github.com/google/uuid"
)

type CourseService struct {
	Repo *persistence.CourseRepository
}

func NewCourseService(container *di.Container) *CourseService {
	return &CourseService{Repo: persistence.NewCourseRepository(container)}
}

func (s *CourseService) CreateCourse(ctx context.Context, input *values.CourseCreate) *errLib.CommonError {

	return s.Repo.CreateCourse(ctx, input)
}

func (s *CourseService) GetCourseById(ctx context.Context, id uuid.UUID) (*entity.Course, *errLib.CommonError) {
	return s.Repo.GetCourseById(ctx, id)
}

func (s *CourseService) GetAllCourses(ctx context.Context) ([]entity.Course, *errLib.CommonError) {

	return s.Repo.GetAllCourses(ctx, "")

}

func (s *CourseService) UpdateCourse(ctx context.Context, input *values.CourseUpdate) *errLib.CommonError {

	return s.Repo.UpdateCourse(ctx, input)
}

func (s *CourseService) DeleteCourse(ctx context.Context, id uuid.UUID) *errLib.CommonError {
	return s.Repo.DeleteCourse(ctx, id)
}
