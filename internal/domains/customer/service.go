package customer

import (
	"api/internal/di"
	"api/internal/domains/customer/persistence"
	"api/internal/domains/customer/values"
	errLib "api/internal/libs/errors"
	"context"

	"github.com/google/uuid"
)

type CustomersService struct {
	Repo *persistence.CustomersRepository
}

func NewCustomersService(container *di.Container) *CustomersService {
	return &CustomersService{Repo: persistence.NewCustomersRepository(container)}
}

func (s *CustomersService) GetCustomers(ctx context.Context, eventId *uuid.UUID) ([]values.CustomerWithDetails, *errLib.CommonError) {
	return s.Repo.GetCustomers(ctx, eventId)
}
