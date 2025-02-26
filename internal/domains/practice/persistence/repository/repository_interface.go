package repository

import (
	"api/internal/domains/practice/entity"
	"api/internal/domains/practice/values"
	errLib "api/internal/libs/errors"
	"context"
	"github.com/google/uuid"
)

type PracticeRepositoryInterface interface {
	Create(ctx context.Context, input *values.PracticeDetails) (*entity.Practice, *errLib.CommonError)
	GetPracticeByName(ctx context.Context, name string) (*entity.Practice, *errLib.CommonError)
	List(ctx context.Context) ([]entity.Practice, *errLib.CommonError)
	Update(ctx context.Context, input *entity.Practice) *errLib.CommonError
	Delete(ctx context.Context, id uuid.UUID) *errLib.CommonError
}
