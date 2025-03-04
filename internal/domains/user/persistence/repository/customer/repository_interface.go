package customer

import (
	values "api/internal/domains/user/values/customer"
	"api/internal/domains/user/values/user"
	errLib "api/internal/libs/errors"
	"context"
)

type RepositoryInterface interface {
	UpdateStats(ctx context.Context, valuesToUpdate values.StatsUpdateValue) *errLib.CommonError
	GetCustomers(ctx context.Context, hubspotIds []string) ([]user.ReadValue, *errLib.CommonError)
}
