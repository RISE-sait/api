package values

import "github.com/google/uuid"

type Customer struct {
	CustomerID uuid.UUID
	FirstName  *string
	LastName   *string
	Email      string
	Phone      string
}
