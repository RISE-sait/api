package enrollment

import (
	"api/internal/domains/enrollment/entity"
	"api/internal/domains/enrollment/values"
	errLib "api/internal/libs/errors"
	"context"
	"github.com/google/uuid"
)

type RepositoryInterface interface {
	EnrollCustomer(c context.Context, input values.EnrollmentDetails) (*entity.Enrollment, *errLib.CommonError)
	GetEnrollments(c context.Context, eventId, customerId uuid.UUID) ([]entity.Enrollment, *errLib.CommonError)
	UnEnrollCustomer(c context.Context, id uuid.UUID) *errLib.CommonError
	GetEventIsFull(c context.Context, eventId uuid.UUID) (*bool, *errLib.CommonError)
}
