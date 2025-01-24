package course

import (
	entity "api/internal/domains/course/entities"
	"api/internal/domains/course/infra/persistence"
	errLib "api/internal/libs/errors"
	"context"

	"github.com/google/uuid"
)

type CourseService struct {
	Repo *persistence.CourseRepository
}

func NewCourseService(repo *persistence.CourseRepository) *CourseService {
	return &CourseService{Repo: repo}
}

func (s *CourseService) CreateCourse(ctx context.Context, course *entity.Course) *errLib.CommonError {

	return s.Repo.CreateCourse(ctx, course)
}

func (s *CourseService) GetCourseById(ctx context.Context, id uuid.UUID) (*entity.Course, *errLib.CommonError) {
	return s.Repo.GetCourseById(ctx, id)
}

func (s *CourseService) GetAllCourses(ctx context.Context) ([]entity.Course, *errLib.CommonError) {
	return s.Repo.GetAllCourses(ctx, "")

}

func (s *CourseService) UpdateCourse(ctx context.Context, course *entity.Course) *errLib.CommonError {

	return s.Repo.UpdateCourse(ctx, course)
}

func (s *CourseService) DeleteCourse(ctx context.Context, id uuid.UUID) *errLib.CommonError {
	return s.Repo.DeleteCourse(ctx, id)
}
