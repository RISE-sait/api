package customer

import (
	"api/internal/di"
	entity "api/internal/domains/customer/entities"
	"api/internal/domains/customer/persistence"
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

func (s *CustomersService) GetCustomers(ctx context.Context, id uuid.UUID) ([]entity.Customer, *errLib.CommonError) {
	return s.Repo.GetCustomersByEventId(ctx, id)
}
