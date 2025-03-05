package course

import (
	values "api/internal/domains/course/values"
	errLib "api/internal/libs/errors"
	"context"
	"github.com/google/uuid"
)

type RepositoryInterface interface {
	CreateCourse(ctx context.Context, input values.CreateCourseDetails) (values.ReadDetails, *errLib.CommonError)
	GetCourseById(ctx context.Context, id uuid.UUID) (values.ReadDetails, *errLib.CommonError)
	GetCourses(ctx context.Context, name, description *string) ([]values.ReadDetails, *errLib.CommonError)
	UpdateCourse(ctx context.Context, input values.UpdateCourseDetails) *errLib.CommonError
	DeleteCourse(ctx context.Context, id uuid.UUID) *errLib.CommonError
}
