package customer

import (
	values "api/internal/domains/customer/values"
	errLib "api/internal/libs/errors"
	"context"
)

type RepositoryInterface interface {
	UpdateStats(ctx context.Context, valuesToUpdate values.StatsUpdateValue) *errLib.CommonError
	GetCustomers(ctx context.Context, hubspotIds []string) ([]values.ReadValue, *errLib.CommonError)
}
