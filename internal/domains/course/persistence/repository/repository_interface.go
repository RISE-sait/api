package course

import (
	entity "api/internal/domains/course/entity"
	values "api/internal/domains/course/values"
	errLib "api/internal/libs/errors"
	"context"
	"github.com/google/uuid"
)

type RepositoryInterface interface {
	CreateCourse(ctx context.Context, input *values.Details) (*entity.Course, *errLib.CommonError)
	GetCourseById(ctx context.Context, id uuid.UUID) (*entity.Course, *errLib.CommonError)
	GetCourses(ctx context.Context, name, description *string) ([]entity.Course, *errLib.CommonError)
	UpdateCourse(ctx context.Context, input *entity.Course) (*entity.Course, *errLib.CommonError)
	DeleteCourse(ctx context.Context, id uuid.UUID) *errLib.CommonError
}
