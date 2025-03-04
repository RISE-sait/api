package staff

import (
	values "api/internal/domains/user/values/staff"
	errLib "api/internal/libs/errors"
	"context"
	"github.com/google/uuid"
)

type RepositoryInterface interface {
	GetByID(c context.Context, id uuid.UUID) (values.ReadValues, *errLib.CommonError)
	List(ctx context.Context, role *string, hubspotIds []string) ([]values.ReadValues, *errLib.CommonError)
	Update(c context.Context, staffFields values.UpdateValues) (values.ReadValues, *errLib.CommonError)
	Delete(c context.Context, id uuid.UUID) *errLib.CommonError
}
